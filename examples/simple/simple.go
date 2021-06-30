package main

import (
	"fmt"
	"log"
	"os"

	smp "github.com/netsys-lab/scion-multipath-lib/api"
	"github.com/netsys-lab/scion-multipath-lib/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
)

func customPathSelectAlg(paths []snet.Path) ([]snet.Path, error) {
	paths1 := pathselection.SelectShortestPaths(5, paths)
	pathsToReturn := pathselection.SelectLowestLatencies(3, paths1)
	// pathsToReturn := []snet.Path{smp.SelectLowestLatency(paths)}
	return pathsToReturn, nil
}

func main() {
	peers := []string{"18-ffaa:1:ef8,[127.0.0.1]:12345", "peer2", "peer3"} // Later real addresses
	local := "peer0"
	var parsedPeers []*snet.UDPAddr
	for _, peer := range peers {
		parsedPeer, _ := snet.ParseUDPAddr(peer)
		parsedPeers = append(parsedPeers, parsedPeer)
	}

	for _, peer := range parsedPeers {
		mpSock := smp.NewMPPeerSock(local, peer)
		err := mpSock.Connect(customPathSelectAlg)
		if err != nil {
			log.Fatal("Failed to connect MPPeerSock", err)
			os.Exit(1)
		}
		defer mpSock.Disconnect()

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
