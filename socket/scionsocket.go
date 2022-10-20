package socket

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/netsec-ethz/scion-apps/pkg/pan"
	"github.com/netsys-lab/scion-path-discovery/packets"
	lookup "github.com/netsys-lab/scion-path-discovery/pathlookup"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/path"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"inet.af/netaddr"
)

type SCIONConn struct {
	internalConn pan.Conn
	peer         string
	remote       *snet.UDPAddr
	path         *snet.Path
	metrics      *packets.PathMetrics
	local        *snet.UDPAddr
	socketLocal  *snet.UDPAddr
	selector     *pathselection.FixedSelector
}

// This simply wraps conn.Read and will later collect metrics
func (qc *SCIONConn) Read(b []byte) (int, error) {
	n, err := qc.internalConn.Read(b)
	if err != nil {
		return n, err
	}
	m := qc.GetMetrics()
	m.ReadBytes += int64(n)
	m.ReadPackets++
	return n, err
}

// This simply wraps conn.Write and will later collect metrics
func (qc *SCIONConn) Write(b []byte) (int, error) {
	n, err := qc.internalConn.Write(b)

	m := qc.GetMetrics()
	m.WrittenBytes += int64(n)
	m.WrittenPackets++
	if err != nil {
		return n, err
	}
	return n, err
}

func (qc *SCIONConn) Close() error {
	if qc.internalConn == nil {
		return nil
	}
	err := qc.internalConn.Close()
	if err != nil {
		return err
	}

	return nil
}

func (qc *SCIONConn) GetMetrics() *packets.PathMetrics {
	// return qc.metrics
	return packets.GetMetricsDB().GetOrCreate(qc.socketLocal, qc.path)
}

func (qc *SCIONConn) GetPath() *snet.Path {
	return qc.path
}

func (qc *SCIONConn) SetPath(path *snet.Path) error {
	qc.path = path
	if qc.selector != nil {
		logrus.Error("SETPATHFROMSNET")
		qc.selector.SetPathFromSnet(*path)
	}
	// qc.metrics.Path = path
	return nil
}

func (qc *SCIONConn) GetRemote() *snet.UDPAddr {
	return qc.remote
}

func (qc *SCIONConn) LocalAddr() net.Addr {
	return qc.local
}

func (qc *SCIONConn) RemoteAddr() net.Addr {
	return qc.remote
}

func (qc *SCIONConn) SetDeadline(t time.Time) error {
	return qc.internalConn.SetDeadline(t)
}

func (qc *SCIONConn) SetReadDeadline(t time.Time) error {
	return qc.internalConn.SetReadDeadline(t)
}

func (qc *SCIONConn) SetWriteDeadline(t time.Time) error {
	return qc.internalConn.SetWriteDeadline(t)
}

var _ packets.UDPConn = (*SCIONConn)(nil)

var _ UnderlaySocket = (*SCIONSocket)(nil)

type SCIONSocket struct {
	local          string
	localAddr      *snet.UDPAddr
	conns          []*SCIONConn
	Conn           pan.Conn
	ConnectedPeers []RemotePeer
	listenConn     pan.ListenConn
}

func (s *SCIONSocket) GetMetrics() []*packets.PathMetrics {
	return packets.GetMetricsDB().GetBySocket(s.localAddr)
}

func (s *SCIONSocket) Local() *snet.UDPAddr {
	return s.localAddr
}

func (s *SCIONSocket) AggregateMetrics() *packets.PathMetrics {
	ms := packets.GetMetricsDB().GetBySocket(s.localAddr)

	sumBwMbitsRead := make([]int64, 0)
	sumBwMbitsWrite := make([]int64, 0)
	for i, m := range ms {
		bwMbits := make([]int64, 0)
		for j, b := range m.ReadBandwidth {
			val := int64(float64(b*8) / 1024 / 1024)
			bwMbits = append(bwMbits, val)
			if i == 0 {
				sumBwMbitsRead = append(sumBwMbitsRead, val)
			} else {
				// Avoid OoR errors
				if j < len(sumBwMbitsRead) {
					sumBwMbitsRead[j] += val
				}

			}
		}
		bwMbits = make([]int64, 0)
		for j, b := range m.WrittenBandwidth {
			val := int64(float64(b*8) / 1024 / 1024)
			bwMbits = append(bwMbits, val)
			if i == 0 {
				sumBwMbitsWrite = append(sumBwMbitsWrite, val)
			} else {
				// Avoid OoR errors
				if j < len(sumBwMbitsWrite) {
					sumBwMbitsWrite[j] += val
				}

			}
		}
		log.Error("[SCIONSocket] bwMbitsWritten: ", bwMbits, " for path", lookup.PathToString(*m.Path))
	}
	// TODO: Maybe add other fields
	pm := &packets.PathMetrics{
		ReadBandwidth:    sumBwMbitsRead,
		WrittenBandwidth: sumBwMbitsWrite,
	}
	return pm
}

func NewSCIONSocket(local string) *SCIONSocket {
	s := SCIONSocket{
		local:          local,
		conns:          make([]*SCIONConn, 0),
		ConnectedPeers: make([]RemotePeer, 0),
	}

	gob.Register(path.Path{})

	return &s
}

func (s *SCIONSocket) Listen() error {
	logrus.Debug("[SCIONSocket] Listening on ", s.local, " ...")
	lAddr, err := snet.ParseUDPAddr(s.local)
	s.localAddr = lAddr
	if err != nil {
		return err
	}

	ipP := pan.IPPortValue{}
	shortAddr := fmt.Sprintf("%s:%d", lAddr.Host.IP, lAddr.Host.Port)
	ipP.Set(shortAddr)
	listener, err := pan.ListenUDP(context.Background(), ipP.Get(), nil)
	if err != nil {
		return err
	}

	s.listenConn = listener
	logrus.Debug("[SCIONSocket] Listen on ", s.local, " successful")
	return err
}

// TODO: This needs to be done for each incoming conn
func (s *SCIONSocket) WaitForIncomingConn(lAddr snet.UDPAddr) (packets.UDPConn, error) {
	ipP := pan.IPPortValue{}
	shortAddr := fmt.Sprintf("%s:%d", lAddr.Host.IP, lAddr.Host.Port)
	ipP.Set(shortAddr)

	logrus.Debug("[SCIONSocket] Waiting for Incoming Conn, new Listener on ", lAddr.String())
	listener, err := pan.ListenUDP(context.Background(), ipP.Get(), nil)
	if err != nil {
		return nil, err
	}

	logrus.Debug("[SCIONSocket] Reading handshake on ", lAddr.String())

	bts := make([]byte, packets.PACKET_SIZE)
	_, panRemote, panPath, err := listener.ReadFromVia(bts)
	if err != nil {
		return nil, err
	}

	sel := pathselection.FixedSelector{
		FixedPath: panPath,
	}
	err = listener.Close()
	conn, err := pan.DialUDP(context.Background(), ipP.Get(), panRemote, nil, &sel)

	p := DialPacket{}
	network := bytes.NewBuffer(bts) // Stand-in for a network connection
	dec := gob.NewDecoder(network)
	err = dec.Decode(&p)
	if err != nil {
		return nil, err
	}

	logrus.Debug("[SCIONSocket] Got handshake from remote ", p.Addr.String(), " over path ", lookup.PathToString(p.Path))

	// Send reply
	ret := DialPacket{}
	ret.Addr = lAddr
	ret.Path = p.Path

	var network2 bytes.Buffer
	enc := gob.NewEncoder(&network2)

	err = enc.Encode(ret)
	conn.Write(network2.Bytes())
	logrus.Debug("[SCIONSocket] Answer handshake to ", p.Addr.String())

	quicConn := &SCIONConn{
		internalConn: conn,
		path:         &p.Path,
		remote:       &p.Addr,
		metrics:      packets.GetMetricsDB().GetOrCreate(s.localAddr, &p.Path),
		local:        &lAddr,
		socketLocal:  s.localAddr,
	}

	s.conns = append(s.conns, quicConn)
	logrus.Debug("[SCIONSocket] Added new Conn: ", s.local, " to ", p.Addr.String())

	return quicConn, nil
}

func (s *SCIONSocket) WaitForDialIn() (*snet.UDPAddr, error) {
	logrus.Debug("[SCIONSocket] Accepting handshake on ", s.local)

	bts := make([]byte, packets.PACKET_SIZE)
	_, panRemote, panPath, err := s.listenConn.ReadFromVia(bts)
	if err != nil {
		return nil, err
	}

	sel := pathselection.FixedSelector{
		FixedPath: panPath,
	}
	err = s.listenConn.Close()
	ipP := pan.IPPortValue{}
	shortAddr := fmt.Sprintf("%s:%d", s.localAddr.Host.IP, s.localAddr.Host.Port)
	ipP.Set(shortAddr)
	conn, err := pan.DialUDP(context.Background(), ipP.Get(), panRemote, nil, &sel)
	s.Conn = conn

	p := HandshakePacket{}
	network := bytes.NewBuffer(bts) // Stand-in for a network connection
	dec := gob.NewDecoder(network)
	err = dec.Decode(&p)
	if err != nil {
		return nil, err
	}

	logrus.Debug("[SCIONSocket] Got handshake from ", p.Addr.String(), " for ports=", len(p.Ports))

	var wg sync.WaitGroup
	ret := HandshakePacket{}
	ret.Ports = make([]int, 0)
	for i := 0; i < p.NumPorts; i++ {
		wg.Add(1)
		ret.Ports = append(ret.Ports, s.localAddr.Host.Port+11*(i+1)+52*len(s.ConnectedPeers))
		go func(i int) {
			l := s.localAddr.Copy()
			l.Host.Port = l.Host.Port + 11*(i+1) + 52*len(s.ConnectedPeers)
			_, err := s.WaitForIncomingConn(*l)
			if err != nil {
				log.Error(err)
				return
			}
			logrus.Debugf("[SCIONSocket] Dialed In %d of %d on %s from remote %s", i+1, p.NumPorts, l.String(), p.Addr.String())
			wg.Done()
		}(i)
	}

	// Send reply

	ret.Addr = *s.localAddr
	ret.NumPorts = p.NumPorts
	// ret.Ports = p.Ports

	var network2 bytes.Buffer
	enc := gob.NewEncoder(&network2)

	// TODO: Reliable
	err = enc.Encode(ret)
	conn.Write(network2.Bytes())
	logrus.Debug("[SCIONSocket] Sending handshake response to ", p.Addr.String())

	wg.Wait()
	addr := p.Addr
	return &addr, nil
}

func (s *SCIONSocket) DialAll(remote snet.UDPAddr, path []pathselection.PathQuality, options DialOptions) ([]packets.UDPConn, error) {
	if options.NumPaths == 0 && len(path) > 0 {
		options.NumPaths = len(path)
	}

	panAddr, err := pan.ResolveUDPAddr(remote.String())
	if err != nil {
		return nil, err
	}

	logrus.Debug("[SCIONSocket] Dialing all to ", remote.String())

	selector := &pathselection.FixedSelector{}
	conn, err := pan.DialUDP(context.Background(), netaddr.IPPort{}, panAddr, nil, selector)
	if err != nil {
		return nil, err
	}

	logrus.Debug("[SCIONSocket] Opened base conn to ", remote.String())
	s.Conn = conn
	// Send handshake
	ret := HandshakePacket{}
	ret.Addr = *s.localAddr
	ret.NumPorts = options.NumPaths
	ret.Ports = make([]int, options.NumPaths)

	for i := 0; i < options.NumPaths; i++ {
		port := remote.Host.Port + (i+1)*11 + 52*len(s.ConnectedPeers) // TODO: Boundary check, better ranges
		ret.Ports = append(ret.Ports, port)
	}

	var network2 bytes.Buffer
	enc := gob.NewEncoder(&network2)

	err = enc.Encode(ret)
	if err != nil {
		return nil, err
	}

	// TODO: Reliable
	conn.Write(network2.Bytes())
	logrus.Debug("[SCIONSocket] Started handshake to ", remote.String(), " with ports=", len(ret.Ports))

	bts := make([]byte, packets.PACKET_SIZE)
	_, err = conn.Read(bts)
	if err != nil {
		return nil, err
	}

	ps := HandshakePacket{}
	network := bytes.NewBuffer(bts) // Stand-in for a network connection
	dec := gob.NewDecoder(network)
	err = dec.Decode(&ps)
	if err != nil {
		log.Error("From decode")
		return nil, err
	}
	logrus.Debug("[SCIONSocket] Completed handshake to ", remote.String(), " with ports=", len(ret.Ports))

	// TODO: Ports may change here...
	var wg sync.WaitGroup

	for i, p := range path {
		wg.Add(1)
		go func(i int, p snet.Path) {
			l := remote.Copy()

			l.Host.Port = ps.Ports[i]

			local := s.localAddr.Copy()
			local.Host.Port = local.Host.Port + (i+1)*11 + 52*len(s.ConnectedPeers)
			// Waitgroup here before sending back response
			_, err := s.Dial(*local, *l, p)
			if err != nil {
				log.Error(err)
				return
			}
			logrus.Debugf("[SCIONSocket] Dialed %d of %d on %s to remote %s", i, options.NumPaths, s.local, l.String())
			wg.Done()
		}(i, p.SnetPath)
	}
	wg.Wait()

	log.Warn("DIAL ALL Done")

	return s.GetConnections(), nil
}

func (s *SCIONSocket) Dial(local, remote snet.UDPAddr, path snet.Path) (packets.UDPConn, error) {
	panAddr, err := pan.ResolveUDPAddr(remote.String())
	if err != nil {
		return nil, err
	}

	ipP := pan.IPPortValue{}
	shortAddr := fmt.Sprintf("%s:%d", local.Host.IP, local.Host.Port)
	ipP.Set(shortAddr)

	// Set Pinging Selector with active probing on two paths
	selector := &pathselection.FixedSelector{}
	selector.SetPathFromSnet(path)

	logrus.Debug("[SCIONSocket] Dial new conn from ", local.String(), " to ", remote.String())
	session, err := pan.DialUDP(context.Background(), ipP.Get(), panAddr, nil, selector)
	if err != nil {
		return nil, err
	}

	logrus.Debug("[SCIONSocket] Dialed new conn from ", local.String(), " to ", remote.String(), " over path ", lookup.PathToString(path))

	// Send handshake
	ret := DialPacket{}
	ret.Addr = local
	ret.Path = path

	var network2 bytes.Buffer
	enc := gob.NewEncoder(&network2)

	err = enc.Encode(ret)
	if err != nil {
		return nil, err
	}

	quicConn := &SCIONConn{
		internalConn: session,
		path:         &path,
		remote:       &remote,
		metrics:      packets.GetMetricsDB().GetOrCreate(s.localAddr, &path),
		socketLocal:  s.localAddr,
		selector:     selector,
		// local:        session.LocalAddr(), // TODO: Local Addr
	}

	// For loop, deadline, write packet, read response
	for i := 0; i < 5; i++ {
		logrus.Debug("[SCIONSocket] Writing handshake from ", local.String(), " to ", remote.String())
		quicConn.Write(network2.Bytes())

		quicConn.SetReadDeadline(time.Now().Add(3 * time.Second))
		bts := make([]byte, packets.PACKET_SIZE)
		_, err := quicConn.Read(bts)
		logrus.Debug("[SCIONSocket] Read handshake response from ", local.String(), " to ", remote.String())
		if err != nil {
			i++
			continue
		}
		p := DialPacket{}
		network := bytes.NewBuffer(bts) // Stand-in for a network connection
		dec := gob.NewDecoder(network)
		err = dec.Decode(&p)
		if err != nil {
			return nil, err
		}
		break
	}

	logrus.Debug("[SCIONSocket] Dial complete from ", local.String(), " to ", remote.String())

	s.conns = append(s.conns, quicConn)
	return quicConn, nil
}

func (s *SCIONSocket) GetConnections() []packets.UDPConn {
	conns := make([]packets.UDPConn, 0)
	for _, c := range s.conns {
		conns = append(conns, c)
	}
	return conns
}

func (s *SCIONSocket) CloseAll() []error {
	errors := make([]error, 0)
	for _, con := range s.conns {
		err := con.Close()
		if err != nil {
			errors = append(errors, err)
		}
	}

	s.conns = make([]*SCIONConn, 0)
	s.ConnectedPeers = make([]RemotePeer, 0)
	return errors
}
