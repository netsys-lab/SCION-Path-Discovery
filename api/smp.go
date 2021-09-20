package smp

import (
	"time"

	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/netsys-lab/scion-path-discovery/socket"
	"github.com/scionproto/scion/go/lib/snet"
	log "github.com/sirupsen/logrus"
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
// TODO: Make fields private that should be private...
type MPPeerSock struct {
	Peer                    *snet.UDPAddr
	OnPathsetChange         chan pathselection.PathSet
	OnConnectionsChange     chan []packets.UDPConn
	PathSelectionProperties []string // TODO: Design a real struct for this, string is only dummy
	PacketScheduler         packets.PacketScheduler
	Local                   string
	UnderlaySocket          socket.UnderlaySocket
	TransportConstructor    packets.TransportConstructor
	PathQualityDB           pathselection.PathQualityDatabase
	SelectedPathSet         *pathselection.PathSet
}

func NewMPPeerSock(local string, peer *snet.UDPAddr) *MPPeerSock {
	return &MPPeerSock{
		Peer:                 peer,
		Local:                local,
		OnPathsetChange:      make(chan pathselection.PathSet),
		TransportConstructor: packets.QUICConnConstructor,
		UnderlaySocket:       socket.NewQUICSocket(local),
		PacketScheduler:      &packets.SampleFirstPathScheduler{},
		PathQualityDB:        pathselection.NewInMemoryPathQualityDatabase(),
		OnConnectionsChange:  make(chan []packets.UDPConn),
	}
}

func (mp *MPPeerSock) SetPeer(peer *snet.UDPAddr) {
	mp.Peer = peer
}

func (mp *MPPeerSock) Listen() error {
	err := mp.UnderlaySocket.Listen()
	if err != nil {
		return err
	}

	conns := mp.UnderlaySocket.GetConnections()
	mp.PacketScheduler.SetConnections(conns)
	mp.PathQualityDB.SetConnections(conns)
	log.Debugf("Listening on %s", mp.Local)
	return nil
}

func (mp *MPPeerSock) WaitForPeerConnect(pathSetWrapper pathselection.CustomPathSelection) (*snet.UDPAddr, error) {
	log.Debugf("Waiting for incoming connection")
	remote, err := mp.UnderlaySocket.WaitForDialIn()
	if err != nil {
		return nil, err
	}
	log.Debugf("Accepted connection from %s", remote.String())
	mp.Peer = remote

	// Start selection process -> will update DB
	mp.StartPathSelection(pathSetWrapper)
	// wait until first signal on channel
	// selectedPathSet := <-mp.OnPathsetChange
	// time.Sleep(1 * time.Second)
	// dial all paths selected by user algorithm
	err = mp.DialAll(mp.SelectedPathSet, &ConnectOptions{
		SendAddrPacket: false,
	})

	return remote, err
}

/*
func (mp *MPPeerSock) WaitForPeerConnectBack() (*snet.UDPAddr, error) {
	log.Debugf("Waiting for incoming connection")
	_, err := mp.UnderlaySocket.WaitForDialIn()
	if err != nil {
		return nil, err
	}
	log.Debugf("Accepted connection back %s", mp.Peer)
	// mp.Peer = remote

	return nil, err
}
*/

func (mp *MPPeerSock) StartPathSelection(pathSetWrapper pathselection.CustomPathSelection) {
	// DONE!
	// TODO: Nico/Karola: Implement metrics collection and path alg invocation
	// We could put a timer here.
	// Every X seconds we collect metrics from the underlaySocket and its connections
	// and provide them for path selection
	// So in a timer call underlaysocket.GetConnections
	// And write the measured metrics in the QualityDB
	// Then you could invoke this the path selection algorithm
	// And if this returns another pathset then currently active,
	// one could invoke this event here...
	// To connect over the new pathset, call mpSock.DialAll(pathset)

	// TODO: 10 seconds
	ticker := time.NewTicker(100 * time.Second)

	go func() {
		for range ticker.C {
			if mp.Peer != nil {
				mp.pathSelection(pathSetWrapper)
				mp.DialAll(mp.SelectedPathSet, nil)
			}

		}
	}()

	mp.pathSelection(pathSetWrapper)

	// Determine Pathlevelpeers
	// mp.PacketScheduler.SetPathlevelPeers()
}

func (mp *MPPeerSock) pathSelection(pathSetWrapper pathselection.CustomPathSelection) {
	mp.PathQualityDB.UpdatePathQualities(mp.Peer)
	mp.PathQualityDB.UpdateMetrics()
	// update DB / collect metrics
	pathSet, err := mp.PathQualityDB.GetPathSet(mp.Peer)
	if err != nil {
		return
	}
	// TODO: Error handling
	selectedPathSet, err := pathSetWrapper.CustomPathSelectAlg(&pathSet)
	mp.SelectedPathSet = selectedPathSet
	// mp.DialAll(selectedPathSet, &ConnectOptions{})
	mp.pathSetChange(*selectedPathSet)
	// mp.OnPathsetChange <- *selectedPathSet
}

//
// Added in 0.0.2
//

// Read from the peer over a specific path
// Here the socket could decide from which path to read or we have to read from all
func (mp *MPPeerSock) Read(b []byte) (int, error) {
	return mp.PacketScheduler.Read(b)
}

// Write to the peer over a specific path
// Here the socket could decide over which path to write
func (mp *MPPeerSock) Write(b []byte) (int, error) {
	return mp.PacketScheduler.Write(b)
}

type ConnectOptions struct {
	SendAddrPacket bool
}

// A first approach could be to open connections over all
// Paths to later reduce time effort for switching paths
func (mp *MPPeerSock) Connect(pathSetWrapper pathselection.CustomPathSelection, options *ConnectOptions) error {
	mp.StartPathSelection(pathSetWrapper)
	// TODO: Rethink default values here...
	opts := &ConnectOptions{}
	if options == nil {
		opts.SendAddrPacket = true
	} else {
		opts = options
	}
	var err error

	/*selectedPathSet, err := mp.PathQualityDB.GetPathSet(mp.Peer)
	if err != nil {
		return err
	}*/
	err = mp.DialAll(mp.SelectedPathSet, opts)
	if err != nil {
		return err
	}
	return nil
}

func (mp *MPPeerSock) pathSetChange(selectedPathset pathselection.PathSet) {
	select {
	// TODO: Fixme
	case mp.OnPathsetChange <- selectedPathset:
	default:
	}
}

func (mp *MPPeerSock) connectionSetChange(conns []packets.UDPConn) {
	select {
	case mp.OnConnectionsChange <- conns:
	default:
	}
}

func (mp *MPPeerSock) Disconnect() []error {
	mp.PacketScheduler.SetConnections(make([]packets.UDPConn, 0))
	return mp.UnderlaySocket.CloseAll()
}

// DialPath This one should "activate" the connection over the respective path
// or create one if its not there yet
/*func (mp *MPPeerSock) DialPath(path *snet.Path) (*packets.QUICReliableConn, error) {
	// copy mp.Peer to not interfere with other connections
	connection, err := NewMonitoredConn(*mp.Peer, path)
	if err != nil {
		return nil, err
	}
	return connection, nil
}
*/
// Could call dialPath for all paths. However, not the connections over included
// should be idled or closed here
func (mp *MPPeerSock) DialAll(pathAlternatives *pathselection.PathSet, options *ConnectOptions) error {
	opts := socket.DialOptions{}
	if options != nil {
		opts.SendAddrPacket = options.SendAddrPacket
	}
	conns, err := mp.UnderlaySocket.DialAll(*mp.Peer, pathselection.UnwrapPathset(*pathAlternatives), opts)
	if err != nil {
		return err
	}

	log.Debugf("Dialed all to %s, got %d connections", mp.Peer.String(), len(conns))

	go func() {
		for {
			log.Debugf("Waiting for new connections...")
			conn, err := mp.UnderlaySocket.WaitForIncomingConn()
			if conn == nil && err == nil {
				log.Debugf("Socket does not implement WaitForIncomingConn, stopping here...")
				return
			}
			if err != nil {
				log.Errorf("Failed to wait for incoming connection %s", err.Error())
				return
			}

			conns = mp.UnderlaySocket.GetConnections()
			mp.PacketScheduler.SetConnections(conns)
			mp.PathQualityDB.SetConnections(conns)
			mp.connectionSetChange(conns)
		}
	}()

	mp.PacketScheduler.SetConnections(conns)
	mp.PathQualityDB.SetConnections(conns)
	mp.connectionSetChange(conns)
	return nil
}

//
// Added in 0.0.3 - WIP, not ready yet
//

// Read from the peer over a specific path
// Here the socket could decide from which path to read or we have to read from all
func (mp *MPPeerSock) ReadStream(b []byte) (int, error) {
	return mp.PacketScheduler.ReadStream(b)
}

// Write to the peer over a specific path
// Here the socket could decide over which path to write
func (mp *MPPeerSock) WriteStream(b []byte) (int, error) {
	return mp.PacketScheduler.WriteStream(b)
}
