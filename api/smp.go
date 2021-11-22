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

type MPSocketOptions struct {
	Transport                   string // "QUIC" | "SCION"
	PathSelectionResponsibility string // "CLIENT" | "SERVER" | "BOTH"
	MultiportMode               bool
}

var defaultSocketOptions = &MPSocketOptions{
	Transport:                   "SCION",
	PathSelectionResponsibility: "BOTH",
}

// MPPeerSock This represents a multipath socket that can handle 1-n paths.
// Each socket is bound to a specific peer
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
	Mode                    string
	Options                 *MPSocketOptions
	MetricsInterval         time.Duration
	selection               pathselection.CustomPathSelection
	MetricsInterval         time.Duration
}

//
// Instantiates a new Multipath Peer Socket
// peer argument may be omitted for a socket waiting for an incoming connections
//
func NewMPPeerSock(local string, peer *snet.UDPAddr, options *MPSocketOptions) *MPPeerSock {

	sock := &MPPeerSock{
		Peer:                peer,
		Local:               local,
		OnPathsetChange:     make(chan pathselection.PathSet),
		PacketScheduler:     &packets.SampleFirstPathScheduler{},
		PathQualityDB:       pathselection.NewInMemoryPathQualityDatabase(),
		OnConnectionsChange: make(chan []packets.UDPConn),
		Options:             defaultSocketOptions,
		MetricsInterval:     100 * time.Millisecond,
	}

	if options != nil {
		sock.Options = options
	}

	socketOptions := &socket.SockOptions{}
	socketOptions.MultiportMode = sock.Options.MultiportMode
	socketOptions.PathSelectionResponsibility = sock.Options.PathSelectionResponsibility

	switch sock.Options.Transport {
	case "SCION":
		sock.UnderlaySocket = socket.NewSCIONSocket(local)
		break
	case "QUIC":
		sock.UnderlaySocket = socket.NewQUICSocket(local, socketOptions)
		break
	}

	return sock
}

//
// Set Mode after intantiating the socket
//
func (mp *MPPeerSock) SetMode(mode string) {
	mp.Mode = mode
}

//
// Set Peer after instantiating the socket
// This does not connect automatically after changing the peer
//
func (mp *MPPeerSock) SetPeer(peer *snet.UDPAddr) {
	mp.Peer = peer
}

//
// Listen on the provided local address
// This call does not wait for incoming connections
// and shout be called for both, waiting and dialing sockets
//
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

//
// This method waits until a remote MPPeerSock calls connect to this
// socket's local address
// A pathselection may be passed, which lets the socket dialing back to its remote
// (e.g. for server-side path selection)
// Since the MPPeerSock waits for only one incoming connection to determine a new peer
// it starts waiting for other connections (if no selection passed) and fires the
// OnConnectionsChange event for each new incoming connection
//
func (mp *MPPeerSock) WaitForPeerConnect(sel pathselection.CustomPathSelection) (*snet.UDPAddr, error) {
	log.Debugf("Waiting for incoming connection")
	remote, err := mp.UnderlaySocket.WaitForDialIn()
	if err != nil {
		return nil, err
	}
	log.Debugf("Accepted connection from %s", remote.String())
	mp.Peer = remote
	mp.selection = sel
	// Start selection process -> will update DB
	mp.StartPathSelection(sel, sel == nil)
	log.Infof("Done path selection")
	// wait until first signal on channel
	// selectedPathSet := <-mp.OnPathsetChange
	// time.Sleep(1 * time.Second)
	// dial all paths selected by user algorithm
	if sel != nil {
		err = mp.DialAll(mp.SelectedPathSet, &socket.ConnectOptions{
			SendAddrPacket: false,
		})
		mp.collectMetrics()
	} else {
		mp.collectMetrics()
		go func() {
			conns := mp.UnderlaySocket.GetConnections()
			mp.PacketScheduler.SetConnections(conns)
			mp.PathQualityDB.SetConnections(conns)
			mp.connectionSetChange(conns)
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

				conns := mp.UnderlaySocket.GetConnections()
				mp.PacketScheduler.SetConnections(conns)
				mp.PathQualityDB.SetConnections(conns)
				mp.connectionSetChange(conns)
			}
		}()
	}

	return remote, err
}

func (mp *MPPeerSock) collectMetrics() {

	ticker := time.NewTicker(mp.MetricsInterval)
	go func() {
		<-ticker.C
		mp.PathQualityDB.UpdateMetrics()
	}()

}

//
// Performs the first pathselection run and if noPeriodicPathSelection is false, also starts the cyclic pathselection
//
func (mp *MPPeerSock) StartPathSelection(sel pathselection.CustomPathSelection, noPeriodicPathSelection bool) {
	// Every X seconds we collect metrics from the underlaySocket and its connections
	// and provide them for path selection
	// So in a timer call underlaysocket.GetConnections
	// And write the measured metrics in the QualityDB
	// Then you could invoke this the path selection algorithm
	// And if this returns another pathset then currently active,
	// one could invoke this event here...
	// To connect over the new pathset, call mpSock.DialAll(pathset)

	if sel == nil {
		return
	}

	if !noPeriodicPathSelection {
		ticker := time.NewTicker(5 * time.Second)
		go func() {
			for range ticker.C {
				if mp.Peer != nil {
					mp.pathSelection(sel)
					mp.DialAll(mp.SelectedPathSet, &socket.ConnectOptions{
						SendAddrPacket:      true,
						DontWaitForIncoming: true,
					})
				}

			}
		}()
	}

	mp.pathSelection(sel)
}

func (mp *MPPeerSock) ForcePathSelection() {
	mp.pathSelection(mp.selection)
}

//
//  Actual pathselection implementation
//
func (mp *MPPeerSock) pathSelection(sel pathselection.CustomPathSelection) {
	mp.PathQualityDB.UpdatePathQualities(mp.Peer, mp.MetricsInterval)
	// update DB / collect metrics
	pathSet, err := mp.PathQualityDB.GetPathSet(mp.Peer)
	if err != nil {
		log.Errorf("Failed to get current pathset %s", err)
		return
	}
	selectedPathSet, err := sel.CustomPathSelectAlg(&pathSet)
	if err != nil {
		log.Errorf("Failed to get call customPathSelection %s", err)
		return
	}
	mp.SelectedPathSet = selectedPathSet
	mp.pathSetChange(*selectedPathSet)
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

// A first approach could be to open connections over all
// Paths to later reduce time effort for switching paths
func (mp *MPPeerSock) Connect(pathSetWrapper pathselection.CustomPathSelection, options *socket.ConnectOptions) error {
	// TODO: Rethink default values here...
	opts := &socket.ConnectOptions{}
	if options == nil {
		opts.SendAddrPacket = true
	} else {
		opts = options
	}
	var err error
	mp.selection = pathSetWrapper
	mp.StartPathSelection(pathSetWrapper, opts.NoPeriodicPathSelection)
	/*selectedPathSet, err := mp.PathQualityDB.GetPathSet(mp.Peer)
	if err != nil {
		return err
	}*/
	err = mp.DialAll(mp.SelectedPathSet, opts)
	if err != nil {
		return err
	}
	mp.collectMetrics()
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

// Could call dialPath for all paths. However, not the connections over included
// should be idled or closed here
func (mp *MPPeerSock) DialAll(pathAlternatives *pathselection.PathSet, options *socket.ConnectOptions) error {
	opts := socket.DialOptions{}
	if options != nil {
		opts.SendAddrPacket = options.SendAddrPacket
	}
	conns, err := mp.UnderlaySocket.DialAll(*mp.Peer, pathAlternatives.Paths, opts)
	if err != nil {
		return err
	}

	log.Debugf("Dialed all to %s, got %d connections", mp.Peer.String(), len(conns))

	if options == nil || !options.DontWaitForIncoming {
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
	}

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

type MPListenerOptions struct {
	Transport string // "QUIC" | "SCION"
}

var defaultListenerOptions = &MPListenerOptions{
	Transport: "SCION",
}

// Waits for multiple incoming MPPeerSock connections
// Since the MPPeerSock itself is bound to a particular
// peer, it can only wait for one incoming connection
// Therefor, the MPListener can be used to wait for multiple
// Incoming peer connections
type MPListener struct {
	local   string
	socket  socket.UnderlaySocket
	options *MPListenerOptions
}

// Instantiates a new MPListener
func NewMPListener(local string, options *MPListenerOptions) *MPListener {
	listener := &MPListener{
		options: defaultListenerOptions,
		local:   local,
	}
	if options != nil {
		listener.options = options
	}

	switch listener.options.Transport {
	case "SCION":
		listener.socket = socket.NewSCIONSocket(local)
		break
	case "QUIC":
		// No explicit path selection here, all done by later created MPPeerSocks
		listener.socket = socket.NewQUICSocket(local, &socket.SockOptions{
			PathSelectionResponsibility: "CLIENT",
		})
		break
	}
	return listener
}

// Needs to be called to listen before waiting can be started
func (l *MPListener) Listen() error {
	return l.socket.Listen()
}

// Waits for new incoming MPPeerSocks
// Should be called in a loop
// Using the returned addr, a new MPPeerSock can be instantiated
// That dials back to the incoming socket
func (l *MPListener) WaitForMPPeerSockConnect() (*snet.UDPAddr, error) {
	return l.socket.WaitForDialIn()
}
