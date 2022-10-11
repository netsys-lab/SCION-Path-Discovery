package smp

import (
	"time"

	"github.com/netsys-lab/scion-path-discovery/packets"
	lookup "github.com/netsys-lab/scion-path-discovery/pathlookup"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/netsys-lab/scion-path-discovery/socket"
	"github.com/scionproto/scion/go/lib/snet"
	log "github.com/sirupsen/logrus"
)

type PanSocketOptions struct {
	Transport                   string // "QUIC" | "SCION"
	PathSelectionResponsibility string // "CLIENT" | "SERVER" | "BOTH"
	MultiportMode               bool
}

var defaultPanSocketOptions = &PanSocketOptions{
	Transport:                   "SCION",
	PathSelectionResponsibility: "CLIENT",
}

type PanSocket struct {
	Peer              *snet.UDPAddr
	Local             string
	UnderlaySocket    socket.UnderlaySocket
	PathQualityDB     pathselection.PathQualityDatabase
	Mode              string
	Options           *MPSocketOptions
	MetricsInterval   time.Duration
	metricsTicker     *time.Ticker
	OnNewConnReceived chan packets.UDPConn
}

//
// Instantiates a new Multipath Peer Socket
// peer argument may be omitted for a socket waiting for an incoming connections
//
func NewPanSock(local string, peer *snet.UDPAddr, options *MPSocketOptions) *PanSocket {

	sock := &PanSocket{
		Peer:              peer,
		Local:             local,
		PathQualityDB:     pathselection.NewInMemoryPathQualityDatabase(),
		Options:           defaultSocketOptions,
		MetricsInterval:   1000 * time.Millisecond,
		OnNewConnReceived: make(chan packets.UDPConn, 16),
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

// Listen on the provided local address
// This call does not wait for incoming connections
// and shout be called for both, waiting and dialing sockets
//
func (mp *PanSocket) Listen() error {
	err := mp.UnderlaySocket.Listen()
	if err != nil {
		return err
	}

	conns := mp.UnderlaySocket.GetConnections()
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
func (mp *PanSocket) WaitForPeerConnect() (*snet.UDPAddr, error) {
	log.Debugf("Waiting for incoming connection")
	remote, err := mp.UnderlaySocket.WaitForDialIn()
	if err != nil {
		return nil, err
	}
	log.Debugf("Accepted connection from %s", remote.String())
	mp.Peer = remote
	log.Debugf("Done path selection")
	// wait until first signal on channel
	// selectedPathSet := <-mp.OnPathsetChange
	// time.Sleep(1 * time.Second)
	// dial all paths selected by user algorithm

	mp.collectMetrics()
	go func() {
		conns := mp.UnderlaySocket.GetConnections()
		mp.PathQualityDB.SetConnections(conns)
		for {
			log.Debug("CLIENT Waiting for new connections...")
			conn, err := mp.UnderlaySocket.WaitForIncomingConn()
			// New conn
			if conn == nil && err == nil {
				log.Warn("CLIENT Socket does not implement WaitForIncomingConn, stopping here...")
				return
			}
			if err != nil {
				log.Errorf("CLIENT Failed to wait for incoming connection %s", err.Error())
				return
			}
		}
	}()

	return remote, err
}

func (mp *PanSocket) collectMetrics() {
	mp.metricsTicker = time.NewTicker(mp.MetricsInterval)
	go func() {
		for {
			select {
			case <-mp.metricsTicker.C:
				mp.PathQualityDB.UpdateMetrics()
				break
				// case <-mp.metricsChan:
				// 	return
			}
		}

	}()

}

func (mp *PanSocket) GetAvailablePaths() ([]snet.Path, error) {
	return lookup.PathLookup(mp.Peer.String())
}

//
// Set Peer after instantiating the socket
// This does not connect automatically after changing the peer
//
func (mp *PanSocket) SetPeer(peer *snet.UDPAddr) {
	mp.Peer = peer
}

// Could call dialPath for all paths. However, not the connections over included
// should be idled or closed here
func (mp *PanSocket) DialAll(pathAlternatives *pathselection.PathSet, options *socket.ConnectOptions) error {
	opts := socket.DialOptions{}
	if options != nil {
		opts.SendAddrPacket = options.SendAddrPacket
	}
	conns, err := mp.UnderlaySocket.DialAll(*mp.Peer, pathAlternatives.Paths, opts)
	if err != nil {
		return err
	}

	log.Debugf("Dialed all to %s, got %d connections", mp.Peer.String(), len(conns))

	mp.PathQualityDB.SetConnections(conns)
	return nil
}

// A first approach could be to open connections over all
// Paths to later reduce time effort for switching paths
func (mp *PanSocket) Connect(pathAlternatives *pathselection.PathSet, options *socket.ConnectOptions) error {
	// TODO: Rethink default values here...
	opts := &socket.ConnectOptions{}
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
	err = mp.DialAll(pathAlternatives, opts)
	if err != nil {
		return err
	}
	if !opts.NoMetricsCollection {
		mp.collectMetrics()
	}

	return nil
}

func (mp *PanSocket) Disconnect() []error {
	errs := mp.UnderlaySocket.CloseAll()
	mp.metricsTicker.Stop()
	return errs
}
