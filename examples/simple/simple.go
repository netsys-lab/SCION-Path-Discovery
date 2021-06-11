package main

import (
	"fmt"

	"github.com/netsys-lab/scion-multipath-lib/smp"

	"github.com/scionproto/scion/go/lib/snet"
)

func main() {

	peer1, nil := snet.ParseUDPAddr("19-ffaa:1:e9e,[127.0.0.1]:12345")
	peers := []*snet.UDPAddr{peer1} //, ""}
	manualSelection := false

	for _, peer := range peers {
		mpSock := smp.NewMPSock(peer)

		// TODO: We could remove the return of the connections for
		// the connect and dial methods since the socket keeps
		// them, but maybe its easier for applications, especially for dialPath
		_, err := mpSock.Connect()
		if err != nil {
			return
		}
		defer mpSock.Disconnect()

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
						mpSock.CloseConn(mpSock.Connections[len(mpSock.Connections)-1])
					}
					// And use the first path returned to open a new one.
					// This example is not intended to make sense, but to show
					// how interacting with the socket could work
					if len(newPaths) > 0 {
						conn, err := mpSock.DialPath(newPaths[0])
						if err != nil {
							return
						}

						if conn.State == smp.CONN_HANDSHAKING {
							fmt.Printf("Connection for path %s is now handshaking", conn.Path)
						}
					}
				} else {
					// This dials connections over all new paths and closes the old ones
					// This could be wrapped in a "MPSock" struct that does also packet scheduling
					// But we will see later...
					_, err = mpSock.DialAll(newPaths)
				}

				// The socket keeps always an up to date list of all connections
				for _, conn := range mpSock.Connections {
					if conn.State == smp.CONN_ACTIVE {
						fmt.Printf("Connection for path %s is active", conn.Path)
					}
				}
			}
		}()
	}
}
