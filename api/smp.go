package smp

import (
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
	// "github.com/netsys-lab/scion-multipath-lib/peers"
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

// MPPeerSock This represents a multipath socket that can handle 1-n paths.
// Each socket is bound to a specific peer
// TODO: One socket that handles multiple peers? This could be done by a wrapper
// that handles multiple MPPeerSocks
type MPPeerSock struct {
	Peer                    *snet.UDPAddr
	OnPathsetChange         chan pathselection.PathSet
	Connections             []packets.MonitoredConn
	FullPathset             []snet.Path
	SelectedPathset         []snet.Path
	Connections             []packets.QUICReliableConn
	PathSelectionProperties []string // TODO: Design a real struct for this, string is only dummy
	PacketScheduler         packets.PacketScheduler
	Local                   string
}

func NewMPPeerSock(local string, peer *snet.UDPAddr) *MPPeerSock {
	return &MPPeerSock{
		Peer:            peer,
		Local:           local,
		OnPathsetChange: make(chan pathselection.PathSet),
	}
}

func (mp MPPeerSock) StartPathSelection() {
	// We could put a timer here.
	// Every X seconds we collect metrics from the packetScheduler
	// and provide them for path selection
	// Furthermore, a first pathset should be defined
	go func() {
		mp.OnPathsetChange <- pathselection.PathSet{}
	}()

	// Determine Pathlevelpeers
	// mp.PacketScheduler.SetPathlevelPeers()
}

//
// Added in 0.0.2
//

// Read from the peer over a specific path
// Here the socket could decide from which path to read or we have to read from all
func (mp MPPeerSock) Read(b []byte) (int, error) {
	return mp.PacketScheduler.Read(b)
}

// Write to the peer over a specific path
// Here the socket could decide over which path to write
func (mp MPPeerSock) Write(b []byte) (int, error) {
	return mp.PacketScheduler.Write(b)
}

func NewMonitoredConn(snetUDPAddr snet.UDPAddr, path *snet.Path) (*packets.QUICReliableConn, error) {
	appnet.SetPath(&snetUDPAddr, *path)
	conn, err := appnet.DialAddr(&snetUDPAddr)
	if err != nil {
		return nil, err
	}
	return &packets.QUICReliableConn{
		Path:         path,
		InternalConn: conn,
		State:        CONN_HANDSHAKING,
	}, nil
}

//TODO: can be removed?
//func NewMPSock(peer *snet.UDPAddr) *MPPeerSock {
//	return &MPPeerSock{
//		Peer:            peer,
//		OnPathsetChange: make(chan pathselection.PathSet),
//	}
//}

func CloseConn(conn packets.QUICReliableConn) error {
	return conn.InternalConn.Close()
}

// Connect A first approach could be to open connections over all
// Paths to later reduce time effort for switching paths
func (mp *MPPeerSock) Connect(pathSetWrapper pathselection.CustomPathSelection) error {
	selectedPathSet, err := pathSetWrapper.CustomPathSelectAlg(pathSetWrapper.GetPathSet())
	err = mp.DialAll(selectedPathSet)
	if err != nil {
		return err
	}
	//mp.OnPathsetChange <- selectedPathSet
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

// DialPath This one should "activate" the connection over the respective path
// or create one if its not there yet
func (mp *MPPeerSock) DialPath(path *snet.Path) (*packets.QUICReliableConn, error) {
	// copy mp.Peer to not interfere with other connections
	connection, err := NewMonitoredConn(*mp.Peer, path)
	if err != nil {
		return nil, err
	}
	return connection, nil
}

// DialAll Could call dialPath for all paths. However, not the connections over included
// should be idled or closed here
func (mp *MPPeerSock) DialAll(pathAlternatives *pathselection.PathSet) error {
	for _, p := range pathAlternatives.Paths {
		path := &p.Path
		connection, err := mp.DialPath(path)
		if err != nil {
			return err
		}
		mp.Connections = append(mp.Connections, *connection)
	}
	return nil
}

//
// Added in 0.0.3
//

// Read from the peer over a specific path
// Here the socket could decide from which path to read or we have to read from all
func (mp MPPeerSock) ReadStream(b []byte) (int, error) {
	return mp.PacketScheduler.ReadStream(b)
}

// Write to the peer over a specific path
// Here the socket could decide over which path to write
func (mp MPPeerSock) WriteStream(b []byte) (int, error) {
	return mp.PacketScheduler.WriteStream(b)
}
