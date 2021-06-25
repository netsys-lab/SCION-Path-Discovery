package smp

import (
	"context"
	"fmt"

	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/netsys-lab/scion-multipath-lib/packets"
	"github.com/netsys-lab/scion-multipath-lib/peers"
)

// Pathselection/Multipath library draft 0.0.2
//
// These code fragments aim to provide an initial design for a "multipath library" that provides
// a generic interface for path monitoring to gather information and path selection based on these information
// Although the library must be designed to be integrated into any kind of application, this first draft is tailored
// to be integrated into Bittorrent.

// Designing and implementing a multipath library in SCION emerges the following two core problems (along with many others):
// 1) Path selection: Out of a potential huge set of paths, which of these should the library use
// 2) Packet scheduling: Which packet is sent over which path by the library

// Since Bittorrent provides its own logic for packet scheduling (not on packet, but on chunk level),
// this is not yet covered by this draft. However, it can be implemented on top of the interfaces and structs
// defined here.

// This draft covers the following design idea of a multipath library: To not implement a dedicated
// packet scheduler at the moment (which may be of course useful later), the library provides an API
// that provides an "optimal" set of paths to the applications and a socket that provides connections
// over the respective paths. The application can read and write data over the provided connections.
// Furthermore, the connections collect metrics under the hood which are then used for potential
// changes to the optimal path set.
// If the path set changes, an event will be emitted to the application, which can then react to the
// new set of optimal paths. This is my first idea that can be integrated into Bittorrent without re-implementing
// packet scheduling

// Note: The main func here is an example of an app using the library, so the code in there should
// not part of the library. We could also move this part into examples folder

// Connection States, need to be redefined/extended
const (
	CONN_IDLE        = 0
	CONN_ACTIVE      = 1
	CONN_CLOSED      = 2
	CONN_HANDSHAKING = 3
)

// This represents a multipath socket that can handle 1-n paths.
// Each socket is bound to a specific peer
// TODO: One socket that handles multiple peers? This could be done by a wrapper
// that handles multiple MPPeerSocks
type MPPeerSock struct {
	Peer                    *snet.UDPAddr
	OnPathsetChange         chan []snet.Path
	FullPathset             []snet.Path
	SelectedPathset         []snet.Path
	Connections             []MonitoredConn
	PathSelectionProperties []string // TODO: Design a real struct for this, string is only dummy
	PacketScheduler         packets.PacketScheduler
	Local                   string
}

func NewMPPeerSock(local, peer string) *MPPeerSock {
	return &MPPeerSock{
		Peer:            peer,
		Local:           local,
		OnPathsetChange: make(chan []string),
	}
}

func (mp MPPeerSock) StartPathSelection() {
	// We could put a timer here.
	// Every X seconds we collect metrics from the packetScheduler
	// and provide them for path selection
	// Furthermore, a first pathset should be defined
	go func() {
		mp.OnPathsetChange <- []string{"Path1", "Path2", "Path3"}
	}()

	// Determine Pathlevelpeers
	// mp.PacketScheduler.SetPathlevelPeers()
}

// A first approach could be to open connections over all
// Paths to later reduce time effort for switching paths
func (mp MPPeerSock) Connect() error {
	mp.StartPathSelection()
	return nil
}

// TODO: Close all connections gracefully...
func (mp MPPeerSock) Disconnect() error {
	return nil
}

//
// Added in 0.0.2
//

// Read from the peer over a specific path
// Here the socket could decide from which path to read or we have to read from all
func (mp MPPeerSock) Read(b []byte) (int, error) {
	return 0, nil
}

// Write to the peer over a specific path
// Here the socket could decide over which path to write
func (mp MPPeerSock) Write(b []byte) (int, error) {
	return 0, nil
}

// DEPRECATED
// This one extends a SCION connection to collect metrics for each connection
// Since a connection has always one path, the metrics are also path metrics
// This becomes obselete if we collect the metrics inside the packet Gen/Handler
type MonitoredConn struct {
	internalConn *snet.Conn
	Path         *snet.Path
	State        int // See Connection States
	// todo append metrices
}

type selAlg func([]snet.Path) ([]snet.Path, error)

// This simply wraps conn.Read and will later collect metrics
func (mConn *MonitoredConn) Read(b []byte) (int, error) {
	n, err := mConn.internalConn.Read(b)
	return n, err
}

// This simply wraps conn.Write and will later collect metrics
func (mConn *MonitoredConn) Write(b []byte) (int, error) {
	n, err := mConn.internalConn.Write(b)
	return n, err
}

func NewMonitoredConn(snetUDPAddr snet.UDPAddr, path *snet.Path) (*MonitoredConn, error) {
	appnet.SetPath(&snetUDPAddr, *path)
	conn, err := appnet.DialAddr(&snetUDPAddr)
	if err != nil {
		return nil, err
	}
	return &MonitoredConn{
		Path:         path,
		internalConn: conn,
		State:        CONN_HANDSHAKING,
	}, nil
}

func NewMPSock(peer *snet.UDPAddr) *MPPeerSock {
	return &MPPeerSock{
		Peer:            peer,
		OnPathsetChange: make(chan []snet.Path),
	}
}

func CloseConn(conn MonitoredConn) error {
	return conn.internalConn.Close()
}

// A first approach could be to open connections over all
// Paths to later reduce time effort for switching paths
func (mp *MPPeerSock) Connect(customPathSelection selAlg) error {
	var err error
	snetUDPAddr := mp.Peer
	mp.FullPathset, err = appnet.DefNetwork().PathQuerier.Query(context.Background(), snetUDPAddr.IA)
	if err != nil {
		return err
	}
	for i, path := range mp.FullPathset {
		fmt.Printf("Path %d: %+v\n", i, path)
	}
	mp.SelectedPathset, err = customPathSelection(mp.FullPathset)
	// err = mp.DialAll()
	// if err != nil {
	// 	return nil, err
	// }
	// mp.Connections[0].Write([]byte("Hello World!\n"))
	mp.OnPathsetChange <- mp.SelectedPathset
	return nil
}

func (mp *MPPeerSock) Disconnect() []error {
	var errs []error
	for _, conn := range mp.Connections {
		err := CloseConn(conn)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// This one should "activate" the connection over the respective path
// or create one if its not there yet
func (mp *MPPeerSock) DialPath(path *snet.Path) (*MonitoredConn, error) {
	// copy mp.Peer to not interfere with other connections
	connection, err := NewMonitoredConn(*mp.Peer, path)
	if err != nil {
		return nil, err
	}
	return connection, nil
}

// Could call dialPath for all paths. However, not the connections over included
// should be idled or closed here
func (mp *MPPeerSock) DialAll() error {
	for _, p := range mp.SelectedPathset {
		connection, err := mp.DialPath(&p)
		if err != nil {
			return err
		}
		mp.Connections = append(mp.Connections, *connection)
	}
	return nil
}
