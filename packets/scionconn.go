package packets

import (
	"net"

	optimizedconn "github.com/johannwagner/scion-optimized-connection/pkg"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
)

var _ UDPConn = (*SCIONConn)(nil)

// var _ TransportConstructor = SCIONTransportConstructor

func SCIONTransportConstructor() UDPConn {
	return &SCIONConn{}
}

func (sc *SCIONConn) GetType() int {
	return sc.connType
}

// This one extends a SCION connection to collect metrics for each connection
// Since a connection has always one path, the metrics are also path metrics
// 0.0.3: Collecting metrics for read and written bytes is better at a place
// where both information are available, so we put it here, not obsolete
type SCIONConn struct { // Former: MonitoredConn
	internalConn *optimizedconn.OptimizedSCIONConn
	path         *snet.Path
	peer         string
	state        int // See Connection States
	metrics      PathMetrics
	remote       *snet.UDPAddr
	local        *net.UDPAddr
	connType     int
}

// This simply wraps conn.Read and will later collect metrics
func (sc *SCIONConn) ReadStream(b []byte) (int, error) {
	n, err := sc.internalConn.Read(b)
	if err != nil {
		return n, err
	}
	sc.metrics.ReadBytes += int64(n)
	sc.metrics.ReadPackets++
	return n, err
}

// This simply wraps conn.Write and will later collect metrics
func (sc *SCIONConn) WriteStream(b []byte) (int, error) {
	n, err := sc.internalConn.Write(b)
	sc.metrics.WrittenBytes += int64(n)
	sc.metrics.WrittenPackets++
	if err != nil {
		return n, err
	}
	return n, err
}

// This simply wraps conn.Read and will later collect metrics
func (sc *SCIONConn) Read(b []byte) (int, error) {
	n, err := sc.internalConn.Read(b)
	if err != nil {
		return n, err
	}
	sc.metrics.ReadBytes += int64(n)
	sc.metrics.ReadPackets++
	return n, err
}

// This simply wraps conn.Write and will later collect metrics
func (sc *SCIONConn) Write(b []byte) (int, error) {
	n, err := sc.internalConn.Write(b)
	sc.metrics.WrittenBytes += int64(n)
	sc.metrics.WrittenPackets++
	if err != nil {
		return n, err
	}
	return n, err
}

func (sc *SCIONConn) Dial(addr snet.UDPAddr, path *snet.Path) error {
	appnet.SetPath(&addr, *path)
	conn, err := optimizedconn.Dial(sc.local, &addr)
	if err != nil {
		return err
	}
	sc.remote = &addr
	sc.path = path
	sc.internalConn = conn
	sc.connType = ConnectionTypes.Outgoing
	return nil

}

func (sc *SCIONConn) Listen(addr snet.UDPAddr) error {
	udpAddr := net.UDPAddr{
		IP:   addr.Host.IP,
		Port: addr.Host.Port,
	}
	conn, err := optimizedconn.Listen(&udpAddr)
	if err != nil {
		return err
	}
	sc.internalConn = conn
	sc.local = &udpAddr
	sc.connType = ConnectionTypes.Incoming
	return nil
}

func (sc *SCIONConn) SetLocal(addr snet.UDPAddr) {
	sc.local = &net.UDPAddr{
		IP: addr.Host.IP,
	}
}

func (sc *SCIONConn) Close() error {
	return sc.internalConn.Close()
}

func (sc *SCIONConn) GetMetrics() *PathMetrics {
	return &sc.metrics
}

func (sc *SCIONConn) GetPath() *snet.Path {
	return sc.path
}
func (sc *SCIONConn) GetRemote() *snet.UDPAddr {
	return sc.remote
}

func (sc *SCIONConn) SetPath(path *snet.Path) {
	sc.path = path
}
func (sc *SCIONConn) SetRemote(remote *snet.UDPAddr) {
	sc.remote = remote
}
