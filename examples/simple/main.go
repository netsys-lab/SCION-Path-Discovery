package main

import (
	"flag"
	"log"
	"os"

	smp "github.com/netsys-lab/scion-path-discovery/api"
	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
)

//LastSelection users could add more fields
type LastSelection struct {
	lastSelectedPathSet pathselection.PathSet
}

//CustomPathSelectAlg this is where the user actually wants to implement its logic in
func (lastSel *LastSelection) CustomPathSelectAlg(pathSet *pathselection.PathSet) (*pathselection.PathSet, error) {
	return pathSet.GetPathHighBandwidth(3), nil
}

var localAddr *string = flag.String("l", "localhost:9999", "Set the local address")
var remoteAddr *string = flag.String("r", "18-ffaa:1:ef8,[127.0.0.1]:12345", "Set the remote address")
var isServer *bool = flag.Bool("s", false, "Run as Server (otherwise, client)")

func main() {
	// peers := []string{"peer1", "peer2", "peer3"} // Later real addresses
	flag.Parse()

	lastSelection := LastSelection{}
	mpSock := smp.NewMPPeerSock(*localAddr, nil)
	err := mpSock.Listen()
	if err != nil {
		log.Fatal("Failed to listen MPPeerSock", err)
		os.Exit(1)
	}

	if *isServer {
		remote, err := mpSock.WaitForPeerConnect(&lastSelection)
		if err != nil {
			log.Fatal("Failed to connect in-dialing peer", err)
			os.Exit(1)
		}
		log.Printf("Connected to %s", remote.String())
		bts := make([]byte, packets.PACKET_SIZE)
		n, err := mpSock.Read(bts)
		if err != nil {
			log.Fatalf("Failed to read bytes from peer %s, err: %v", remote.String(), err)
			os.Exit(1)
		}
		log.Printf("Read %d bytes from %s", n, remote.String())
	} else {
		peerAddr, err := snet.ParseUDPAddr(*remoteAddr)
		if err != nil {
			log.Fatalf("Failed to parse remote addr %s, err: %v", *remoteAddr, err)
			os.Exit(1)
		}
		mpSock.SetPeer(peerAddr)
		err = mpSock.Connect(&lastSelection, nil)
		if err != nil {
			log.Fatal("Failed to connect MPPeerSock", err)
			os.Exit(1)
		}
		bts := make([]byte, packets.PACKET_SIZE)
		n, err := mpSock.Write(bts)
		if err != nil {
			log.Fatalf("Failed to write bytes from peer %s, err: %v", *remoteAddr, err)
			os.Exit(1)
		}
		log.Printf("Wrote %d bytes to %s", n, *remoteAddr)
	}

	// mpSock.
	// mpSock.SetPeer(remote)
	// mpSock.Connect(customPathSelectAlg)
	defer mpSock.Disconnect()

}
