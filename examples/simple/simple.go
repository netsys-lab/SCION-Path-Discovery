package main

import (
	"fmt"
	"log"

	smp "github.com/netsys-lab/scion-path-discovery/api"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
)

func customPathSelectAlg(pathSet pathselection.PathSet) (pathsToReturn pathselection.PathSet, err error) {
	//shortestPathSubSet := pathselection.SelectShortestPaths(5, pathSet.Paths)
	//fastestPathSubSet := pathselection.SelectLowestLatencies(3, shortestPathSubSet)

	shortestPathSubSet := pathSet.GetPathLowLatency()
	pathsToReturn.Paths = append(pathsToReturn.Paths, pathselection.PathQuality{Path: shortestPathSubSet})
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
		}
		defer mpSock.Disconnect()

		go func(mpSock *smp.MPPeerSock) {
			buf := make([]byte, 1200)
			n, err := mpSock.Read(buf)
			if err != nil {
				log.Fatal("Failed to connect MPPeerSock", err)
			}
			fmt.Printf("Read %d bytes of data from %s", n, mpSock.Local)
		}(mpSock)

		data := make([]byte, 1200)
		n, err := mpSock.Write(data)
		if err != nil {
			log.Fatal("Failed to connect MPPeerSock", err)
		}
		fmt.Printf("Wrote %d bytes of data to %s", n, mpSock.Peer)
	}
}
