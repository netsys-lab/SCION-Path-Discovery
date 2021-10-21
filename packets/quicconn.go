package packets

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"net"
	"sync"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/netsec-ethz/scion-apps/pkg/appnet/appquic"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/spath"

	log "github.com/sirupsen/logrus"
)

var _ UDPConn = (*QUICReliableConn)(nil)

func QUICConnConstructor() UDPConn {
	return &QUICReliableConn{}
}

func (qc *QUICReliableConn) GetType() int {
	return ConnectionTypes.Bidirectional
}

type returnPath struct {
	path    *spath.Path
	nextHop *net.UDPAddr
}

type returnPathConn struct {
	net.PacketConn
	mutex sync.RWMutex
	path  *returnPath
}

func Listen(listen *net.UDPAddr) (*returnPathConn, error) {
	sconn, err := appnet.Listen(listen)
	if err != nil {
		return nil, err
	}
	return newReturnPathConn(sconn), nil
}

func newReturnPathConn(conn *snet.Conn) *returnPathConn {
	return &returnPathConn{
		PacketConn: conn,
	}
}

func (c *returnPathConn) ReadFrom(p []byte) (int, net.Addr, error) {
	n, addr, err := c.PacketConn.ReadFrom(p)
	for _, ok := err.(*snet.OpError); err != nil && ok; {
		n, addr, err = c.PacketConn.ReadFrom(p)
	}
	if err == nil {
		if saddr, ok := addr.(*snet.UDPAddr); ok && c.path == nil {
			log.Debugf("Setting return path")
			c.mutex.Lock()
			defer c.mutex.Unlock()
			c.path = &returnPath{path: &saddr.Path, nextHop: saddr.NextHop}
			// saddr.Path = nil // hide it,
			saddr.NextHop = nil
		}
	}
	return n, addr, err
}

func (c *returnPathConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	if saddr, ok := addr.(*snet.UDPAddr); ok && c.path != nil { // XXX && saddr.IA = localIA
		c.mutex.RLock()
		defer c.mutex.RUnlock()

		retPath := c.path
		addr = &snet.UDPAddr{
			IA:      saddr.IA,
			Host:    saddr.Host,
			Path:    *retPath.path,
			NextHop: retPath.nextHop,
		}
	}
	return c.PacketConn.WriteTo(p, addr)
}

// TODO: Implement SCION/QUIC here
type QUICReliableConn struct { // Former: MonitoredConn
	BasicConn
	internalConn quic.Stream
	listener     quic.Listener
	session      quic.Session
	path         *snet.Path
	peer         string
	remote       *snet.UDPAddr
	state        int // See Connection States
	metrics      PathMetrics
	local        *snet.UDPAddr
	Ready        chan bool
	closed       bool
	id           string
}

// This simply wraps conn.Read and will later collect metrics
func (qc *QUICReliableConn) Read(b []byte) (int, error) {
	if qc.internalConn == nil {
		<-qc.Ready
	}
	// log.Warnf("Ready")
	n, err := qc.internalConn.Read(b)
	if err != nil {
		return n, err
	}
	qc.metrics.ReadBytes += int64(n)
	qc.metrics.ReadPackets++
	return n, err
}

func (qc *QUICReliableConn) Dial(addr snet.UDPAddr, path *snet.Path) error {
	qc.state = ConnectionStates.Pending
	sconn, err := appnet.Listen(nil)
	if err != nil {
		return err
	}

	if addr.Path.IsEmpty() {
		appnet.SetPath(&addr, *path)
	}

	qc.peer = addr.String()
	qc.path = path

	host := appnet.MangleSCIONAddr(qc.peer)
	log.Debugf("Dialing to %s and host %s", addr.String(), qc.peer)

	session, err := quic.Dial(sconn, &addr, host, &tls.Config{
		Certificates:       appquic.GetDummyTLSCerts(),
		NextProtos:         []string{"scion-filetransfer"},
		InsecureSkipVerify: true,
	}, &quic.Config{
		KeepAlive: true,
	})

	if err != nil {
		return err
	}

	log.Debugf("Opening Stream to %s", addr.String())

	qc.session = session
	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		return err
	}

	qc.remote = &addr
	log.Debugf("Opened Stream to %s", addr.String())
	qc.state = ConnectionStates.Open
	qc.internalConn = stream
	log.Debugf("Setting stream %p to conn %p", &qc.internalConn, qc)
	return nil
}

// This simply wraps conn.Write and will later collect metrics
func (qc *QUICReliableConn) Write(b []byte) (int, error) {
	// log.Debugf("Writing on stream %p of conn %p", &qc.internalConn, qc)
	n, err := qc.internalConn.Write(b)
	qc.metrics.WrittenBytes += int64(n)
	qc.metrics.WrittenPackets++
	if err != nil {
		return n, err
	}
	return n, err
}

func (qc *QUICReliableConn) WriteStream(b []byte) (int, error) {
	bts := make([]byte, 8)
	binary.BigEndian.PutUint64(bts, uint64(len(b)))
	n, err := qc.Write(bts)
	if err != nil {
		return n, err
	}

	n, err = qc.Write(b)
	return n, err

}

func (qc *QUICReliableConn) ReadStream(b []byte) (int, error) {
	bts := make([]byte, 8)
	n, err := qc.Read(bts)
	if err != nil {
		return n, err
	}
	len := binary.BigEndian.Uint64(bts)
	buf := make([]byte, 9000)
	b = make([]byte, len)
	var i uint64
	for i < len {
		n, err := qc.Read(buf)
		if err != nil {
			return int(i), err
		}
		copy(b[i:int(i)+n], buf)
		i += uint64(n)
	}

	return int(i), err

}

func (qc *QUICReliableConn) Close() error {
	qc.state = ConnectionStates.Closed
	return qc.internalConn.Close()
}

func (qc *QUICReliableConn) AcceptStream() (quic.Stream, error) {
	log.Debugf("Accepting on quic %s", qc.listener.Addr())
	session, err := qc.listener.Accept(context.Background())
	if err != nil {
		return nil, err
	}
	log.Debugf("Got session on quic %s", qc.listener.Addr())

	stream, err := session.AcceptStream(context.Background())
	log.Debugf("ASÖLKD on quic %s", qc.listener.Addr())
	if err != nil {
		return nil, err
	}

	log.Debugf("Accepted on quic %s with %p", qc.listener.Addr(), qc)

	// qc.internalConn = stream

	return stream, nil
}

func (qc *QUICReliableConn) Listen(addr snet.UDPAddr) error {
	qc.Ready = make(chan bool, 0)
	udpAddr := net.UDPAddr{
		IP:   addr.Host.IP,
		Port: addr.Host.Port,
	}
	qc.local = &addr
	sconn, err := Listen(&udpAddr) // appnet.Listen(&udpAddr)
	if err != nil {
		return err
	}
	listener, err := quic.Listen(sconn, &tls.Config{
		Certificates: appquic.GetDummyTLSCerts(),
		NextProtos:   []string{"scion-filetransfer"},
	}, &quic.Config{
		KeepAlive: true,
	})

	if err != nil {
		return err
	}

	log.Debugf("Listen on quic %s wtih scion %s", listener.Addr(), sconn.LocalAddr())

	qc.listener = listener
	qc.state = ConnectionStates.Pending
	return nil
}

func (qc *QUICReliableConn) GetMetrics() *PathMetrics {
	return &qc.metrics
}

func (qc *QUICReliableConn) GetInternalConn() quic.Stream {
	return qc.internalConn
}

func (qc *QUICReliableConn) GetPath() *snet.Path {
	return qc.path
}
func (qc *QUICReliableConn) GetRemote() *snet.UDPAddr {
	return qc.remote
}

func (qc *QUICReliableConn) SetPath(path *snet.Path) {
	qc.path = path
}
func (qc *QUICReliableConn) SetRemote(remote *snet.UDPAddr) {
	qc.remote = remote
}
func (qc *QUICReliableConn) SetLocal(local snet.UDPAddr) {
	qc.local = &local
}

func (qc *QUICReliableConn) SetStream(stream quic.Stream) {
	qc.internalConn = stream
	qc.state = ConnectionStates.Open
}

func (qc *QUICReliableConn) LocalAddr() net.Addr {
	return qc.local
}

func (qc *QUICReliableConn) RemoteAddr() net.Addr {
	return qc.remote
}

func (qc *QUICReliableConn) SetDeadline(t time.Time) error {
	return qc.internalConn.SetDeadline(t)
}

func (qc *QUICReliableConn) SetReadDeadline(t time.Time) error {
	return qc.internalConn.SetReadDeadline(t)
}

func (qc *QUICReliableConn) SetWriteDeadline(t time.Time) error {
	return qc.internalConn.SetWriteDeadline(t)
}

func (qc *QUICReliableConn) GetId() string {
	return qc.id
}

func (qc *QUICReliableConn) SetId(id string) {
	qc.id = id
}
