package socket

import (
	"bytes"
	"encoding/gob"
	"math/rand"

	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/path"
	log "github.com/sirupsen/logrus"
)

var _ packets.UDPConn = (*packets.QUICReliableConn)(nil)

var _ UnderlaySocket = (*QUICSocket)(nil)

type DialPacketQuic struct {
	Addr snet.UDPAddr
	// Path snet.Path
	NumPaths int
}

// TODO: extend this further. It may be useful to use more than
// one native UDP socket due to performance limitations
//type Socket interface {
//	net.Conn
//}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type QUICSocket struct {
	listenConns          []*packets.QUICReliableConn
	local                string
	localAddr            *snet.UDPAddr
	transportConstructor packets.TransportConstructor
	dialConns            []*packets.QUICReliableConn
	acceptedConns        chan []*packets.QUICReliableConn
	options              *SockOptions
	NoReturnPathConn     bool
}

func NewQUICSocket(local string, opts *SockOptions) *QUICSocket {
	s := QUICSocket{
		local:         local,
		listenConns:   make([]*packets.QUICReliableConn, 0),
		dialConns:     make([]*packets.QUICReliableConn, 0),
		acceptedConns: make(chan []*packets.QUICReliableConn, 0),
		options:       opts,
	}

	gob.Register(path.Path{})

	return &s
}

func (s *QUICSocket) Listen() error {
	lAddr, err := snet.ParseUDPAddr(s.local)
	if err != nil {
		return err
	}

	s.localAddr = lAddr
	conn := &packets.QUICReliableConn{
		NoReturnPathConn: s.NoReturnPathConn,
	}
	s.listenConns = append(s.listenConns, conn)
	err = conn.Listen(*s.localAddr)

	return err
}

func (s *QUICSocket) WaitForIncomingConn() (packets.UDPConn, error) {
	if s.options == nil || !s.options.MultiportMode {
		log.Debugf("Waiting for new connection")
		stream, err := s.listenConns[0].AcceptStream()
		if err != nil {
			log.Fatalf("QUIC Accept err %s", err.Error())
		}

		log.Debugf("Accepted new Stream on listen socket")

		bts := make([]byte, packets.PACKET_SIZE)
		_, err = stream.Read(bts)

		if s.listenConns[0].GetInternalConn() == nil {
			s.listenConns[0].SetStream(stream)
			select {
			case s.listenConns[0].Ready <- true:
			default:
			}

			return s.listenConns[0], nil
		} else {
			newConn := &packets.QUICReliableConn{}
			id := RandStringBytes(32)
			newConn.SetId(id)
			newConn.SetLocal(*s.localAddr)
			newConn.SetRemote(s.listenConns[0].GetRemote())
			newConn.SetStream(stream)
			s.listenConns = append(s.listenConns, newConn)

			_, err = stream.Read(bts)
			if err != nil {
				return nil, err
			}
			return newConn, nil
		}
	} else {
		addr := s.localAddr.Copy()
		addr.Host.Port = s.localAddr.Host.Port + len(s.listenConns)
		conn := &packets.QUICReliableConn{}
		err := conn.Listen(*addr)
		if err != nil {
			return nil, err
		}

		stream, err := conn.AcceptStream()
		if err != nil {
			return nil, err
		}

		id := RandStringBytes(32)
		conn.SetId(id)

		conn.SetStream(stream)
		s.listenConns = append(s.listenConns, conn)
		bts := make([]byte, packets.PACKET_SIZE)
		_, err = stream.Read(bts)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

func (s *QUICSocket) WaitForDialIn() (*snet.UDPAddr, error) {
	bts := make([]byte, packets.PACKET_SIZE)
	log.Debugf("Wait for Dial In")
	stream, err := s.listenConns[0].AcceptStream()
	if err != nil {
		return nil, err
	}

	s.listenConns[0].SetStream(stream)

	select {
	case s.listenConns[0].Ready <- true:
	default:
	}

	_, err = stream.Read(bts)
	if err != nil {
		return nil, err
	}
	p := DialPacketQuic{}
	network := bytes.NewBuffer(bts) // Stand-in for a network connection
	dec := gob.NewDecoder(network)
	err = dec.Decode(&p)
	if err != nil {
		return nil, err
	}

	s.listenConns[0].SetRemote(&p.Addr)
	log.Debugf("Waiting for %d more connections", p.NumPaths-1)

	for i := 1; i < p.NumPaths; i++ {
		_, err := s.WaitForIncomingConn()
		if err != nil {
			return nil, err
		}
		log.Debugf("Dialed In %d of %d", i, p.NumPaths)
	}

	addr := p.Addr
	return &addr, nil
}

func (s *QUICSocket) Dial(remote snet.UDPAddr, path snet.Path, options DialOptions, i int) (packets.UDPConn, error) {
	// appnet.SetPath(&remote, path)
	if s.options == nil || !s.options.MultiportMode {
		conn := &packets.QUICReliableConn{}
		conn.SetLocal(*s.localAddr)
		err := conn.Dial(remote, &path)
		if err != nil {
			return nil, err
		}

		if options.SendAddrPacket {
			var network bytes.Buffer
			enc := gob.NewEncoder(&network)
			p := DialPacketQuic{
				Addr:     *s.localAddr,
				NumPaths: options.NumPaths,
			}

			err := enc.Encode(p)
			conn.Write(network.Bytes())
			if err != nil {
				return nil, err
			}
		}

		s.dialConns = append(s.dialConns, conn)

		return conn, nil
	} else {
		conn := &packets.QUICReliableConn{}
		conn.SetLocal(*s.localAddr)
		rem := remote.Copy()
		rem.Host.Port = remote.Host.Port + i
		log.Debugf("Remote port is %d", rem.Host.Port)
		err := conn.Dial(*rem, &path)
		if err != nil {
			return nil, err
		}

		// log.Debugf("Sending addr packet %d for conn %p", options.SendAddrPacket, &conn)
		if options.SendAddrPacket {
			var network bytes.Buffer
			enc := gob.NewEncoder(&network)
			p := DialPacketQuic{
				Addr:     *s.localAddr,
				NumPaths: options.NumPaths,
			}

			err := enc.Encode(p)
			conn.Write(network.Bytes())
			if err != nil {
				return nil, err
			}
		}

		s.dialConns = append(s.dialConns, conn)

		return conn, nil
	}
}

func (s *QUICSocket) DialAll(remote snet.UDPAddr, path []pathselection.PathQuality, options DialOptions) ([]packets.UDPConn, error) {
	if options.NumPaths == 0 && len(path) > 0 {
		options.NumPaths = len(path)
	}

	conns := make([]packets.UDPConn, 0)
	for i, v := range path {
		// Check if conn over path is already open
		connOpen := false
		var openConn packets.UDPConn
		for _, c := range s.dialConns {
			if c.GetId() == v.Id {
				connOpen = true
				openConn = c
				break
			}
		}
		if connOpen {
			log.Debugf("Connection over path id %s already open, skipping", v.Id)
			conns = append(conns, openConn)
			continue
		}
		conn, err := s.Dial(remote, v.Path, options, i)
		if err != nil {
			return nil, err
		}
		conn.SetId(v.Id)
		conns = append(conns, conn)
	}

	select {
	case s.listenConns[0].Ready <- true:
	default:
		// s.listenConns[0]
	}
	dialConns := make([]*packets.QUICReliableConn, 0)
	for _, v := range conns {
		q, ok := v.(*packets.QUICReliableConn)
		if ok {
			dialConns = append(dialConns, q)
		}
	}

	for _, v := range s.dialConns {
		connFound := false
		for _, c := range dialConns {
			if c.GetId() == v.GetId() {
				connFound = true
			}
		}

		// Gracefully close connections that are not used anymore
		// Meaning we set them to closed, so that th
		if !connFound {
			err := v.MarkAsClosed()
			log.Debugf("Marking conn with id %s closed due to no further usage", v.GetId())
			if err != nil {
				return nil, err
			}
		}
	}

	s.dialConns = dialConns
	return conns, nil
}

func (s *QUICSocket) GetConnections() []packets.UDPConn {
	conns := make([]packets.UDPConn, 0)
	for _, v := range s.listenConns {
		conns = append(conns, v)
	}
	for _, v := range s.dialConns {
		conns = append(conns, v)
	}
	return conns
}

func (s *QUICSocket) GetDialConnections() []packets.UDPConn {
	conns := make([]packets.UDPConn, 0)
	for _, v := range s.dialConns {
		conns = append(conns, v)
	}
	return conns
}

func (s *QUICSocket) CloseAll() []error {
	errors := make([]error, 0)
	for _, con := range s.dialConns {
		err := con.Close()
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
