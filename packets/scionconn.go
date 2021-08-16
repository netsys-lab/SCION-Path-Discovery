package packets

import (
	"net"

	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
)

var _ TransportConn = (*SCIONConn)(nil)

// var _ TransportConstructor = SCIONTransportConstructor

func SCIONTransportConstructor() TransportConn {
	return &SCIONConn{}
}

// This one extends a SCION connection to collect metrics for each connection
// Since a connection has always one path, the metrics are also path metrics
// 0.0.3: Collecting metrics for read and written bytes is better at a place
// where both information are available, so we put it here, not obsolete
type SCIONConn struct { // Former: MonitoredConn
	internalConn *snet.Conn
	path         *snet.Path
	peer         string
	state        int // See Connection States
	metrics      PacketMetrics
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

func (sc *SCIONConn) Connect(addr snet.UDPAddr, path *snet.Path) error {
	appnet.SetPath(&addr, *path)
	conn, err := appnet.DialAddr(&addr)
	sc.path = path
	if err != nil {
		return err
	}

	sc.internalConn = conn
	return nil

}

func (sc *SCIONConn) Listen(addr snet.UDPAddr) error {
	udpAddr := net.UDPAddr{
		IP:   addr.Host.IP,
		Port: addr.Host.Port,
	}
	conn, err := appnet.Listen(&udpAddr)
	if err != nil {
		return err
	}
	sc.internalConn = conn
	return nil
}

func (sc *SCIONConn) Close() error {
	return sc.internalConn.Close()
}

func (sc *SCIONConn) GetMetrics() *PacketMetrics {
	return &sc.metrics
}
