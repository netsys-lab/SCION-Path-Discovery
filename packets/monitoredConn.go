package packets

import "github.com/scionproto/scion/go/lib/snet"

// MonitoredConn This one extends a SCION connection to collect metrics for each connection
// Since a connection has always one path, the metrics are also path metrics
// This becomes obsolete if we collect the metrics inside the packet Gen/Handler
type MonitoredConn struct {
	InternalConn *snet.Conn
	Path         *snet.Path
	State        int // See Connection States
}

// This simply wraps conn.Read and will later collect metrics
func (mConn *MonitoredConn) Read(b []byte) (int, error) {
	n, err := mConn.InternalConn.Read(b)
	return n, err
}

// This simply wraps conn.Write and will later collect metrics
func (mConn *MonitoredConn) Write(b []byte) (int, error) {
	n, err := mConn.InternalConn.Write(b)
	return n, err
}
