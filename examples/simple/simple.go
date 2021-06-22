package main

import (
	"fmt"
	"log"
	"os"

	smp "github.com/netsys-lab/scion-multipath-lib/api"
)

func main() {
	peers := []string{"peer1", "peer2", "peer3"} // Later real addresses
	local := "peer0"
	for _, peer := range peers {
		mpSock := smp.NewMPPeerSock(local, peer)
		err := mpSock.Connect()
		if err != nil {
			log.Fatal("Failed to connect MPPeerSock", err)
			os.Exit(1)
		}

		go func(mpSock *smp.MPPeerSock) {
			buf := make([]byte, 1200)
			n, err := mpSock.Read(buf)
			if err != nil {
				log.Fatal("Failed to connect MPPeerSock", err)
				os.Exit(1)
			}
			fmt.Printf("Read %d bytes of data from %s", n, mpSock.Local)
		}(mpSock)

		data := make([]byte, 1200)
		n, err := mpSock.Write(data)
		if err != nil {
			log.Fatal("Failed to connect MPPeerSock", err)
			os.Exit(1)
		}
		fmt.Printf("Wrote %d bytes of data to %s", n, mpSock.Peer)
	}
}

/*
Deprecated: Covers the old routine
package main

import (
	"fmt"

	smp "github.com/netsys-lab/scion-multipath-lib/api"
)

func main() {
	peers := []string{"peer1", "peer2", "peer3"} // Later real addresses
	manualSelection := false

	for _, peer := range peers {
		mpSock := smp.NewMPPeerSock(peer)

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
*/
