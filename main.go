package main

import (
	"fmt"
	"net"
)

// Pathselection/Multipath library draft 0.0.1
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
// not part of the libary. We could also move this part into examples folder

// Connection States, need to be redefined/extended
const (
	CON_IDLE        = 0
	CON_ACTIVE      = 1
	CON_CLOSED      = 2
	CON_HANDSHAKING = 3
)

// This represents a multipath socket that can handle 1-n paths.
// Each socket is bound to a specific peer
// TODO: One socket that handles multiple peers? This could be done by a wrapper
// that handles multiple MPPeerSocks
type MPPeerSock struct {
	Peer                    string
	OnPathsetChange         chan []string // TODO: Design a real struct for this, string is only dummy
	Pathset                 []string      // TODO: Design a real struct for this, string is only dummy
	Connections             []MonitoredConn
	PathSelectionProperties []string // TODO: Design a real struct for this, string is only dummy
}

// This one extends a SCION connection to collect metrics for each connection
// Since a connection has always one path, the metrics are also path metrics
type MonitoredConn struct {
	internalCon net.Conn // Is later SCION conn, or with TAPS a connection independently of the network/transport
	Path        string   // string is only a dummy here, needs to be a real path interface
	State       int      // See Connection States
}

// This simply wraps con.Read and will later collect metrics
func (mCon MonitoredConn) Read(b []byte) (int, error) {
	n, err := mCon.internalCon.Read(b)
	return n, err
}

// This simply wraps con.Write and will later collect metrics
func (mCon MonitoredConn) Write(b []byte) (int, error) {
	n, err := mCon.internalCon.Write(b)
	return n, err
}

func newMonitoredConn(path string) (*MonitoredConn, error) {

	// Here need to be done some SCION or TAPS stuff
	conn, err := net.Dial("scion", path)
	if err != nil {
		return nil, err
	}
	return &MonitoredConn{
		Path:        path,
		internalCon: conn,
		State:       CON_HANDSHAKING,
	}, nil
}

func newMPSock(peer string) *MPPeerSock {
	return &MPPeerSock{
		Peer:            peer,
		OnPathsetChange: make(chan []string),
	}
}

func (mp MPPeerSock) closeConn(conn MonitoredConn) {
	conn.internalCon.Close()
}

// A first approach could be to open connections over all
// Paths to later reduce time effort for switching paths
func (mp MPPeerSock) connect() ([]MonitoredConn, error) {

	go func() {
		// Do some operations on the metrics here
		// and then maybe fire pathset change event
		mp.OnPathsetChange <- []string{"Path1", "Path2", "Path3"}
	}()
	return []MonitoredConn{}, nil
}

// TODO: Close all connections gracefully...
func (mp MPPeerSock) disconnect() error {
	return nil
}

// This one should "activate" the connection over the respective path
// or create one if its not there yet
func (mp MPPeerSock) dialPath(path string) (*MonitoredConn, error) {
	return newMonitoredConn(path)
}

// Could call dialPath for all paths. However, not the connections over included
// should be idled or closed here
func (mp MPPeerSock) dialAll(path []string) ([]MonitoredConn, error) {
	return []MonitoredConn{}, nil
}

func main() {

	peers := []string{"peer1", "peer2", "peer3"} // Later real addresses
	manualSelection := false

	for _, peer := range peers {
		mpSock := newMPSock(peer)

		// TODO: We could remove the return of the connections for
		// the connect and dial methods since the socket keeps
		// them, but maybe its easier for applications, espacially for dialPath
		_, err := mpSock.connect()
		if err != nil {
			return
		}
		defer mpSock.disconnect()

		// Maybe we find a cooler approach here...
		// The basic idea is to have a function that can react to pathSetChanges
		// and e.g. always use all paths that are returned from the socket
		// or only a subset
		go func() {
			for {
				// Note: This is probably the most important line in this draft.
				// Before the lib fires this event, a lot of magic needs to be done
				// First idea was to use a channel to wait for an event
				// Maybe we can find a more elegant way instead of this for loop to wait for the
				// OnPathsetChange message
				newPaths := <-mpSock.OnPathsetChange

				// Manual Selection is only for show purposes...
				if manualSelection {
					// Here we could close the last connection (maybe when they are sorted somehow)
					if len(mpSock.Connections) > 0 {
						mpSock.closeConn(mpSock.Connections[len(mpSock.Connections)-1])
					}
					// And use the first path returned to open a new one.
					// This example is not intended to make sense, but to show
					// how interacting with the socket could work
					if len(newPaths) > 0 {
						con, err := mpSock.dialPath(newPaths[0])
						if err != nil {
							return
						}

						if con.State == CON_HANDSHAKING {
							fmt.Printf("Connection for path %s is now handshaking", con.Path)
						}
					}
				} else {
					// This dials connections over all new paths and closes the old ones
					// This could be wrapped in a "MPSock" struct that does also packet scheduling
					// But we will see later...
					_, err = mpSock.dialAll(newPaths)
				}

				// The socket keeps always an up to date list of all connections
				for _, con := range mpSock.Connections {
					if con.State == CON_ACTIVE {
						fmt.Printf("Connection for path %s is active", con.Path)
					}

				}

			}

		}()
	}
}
