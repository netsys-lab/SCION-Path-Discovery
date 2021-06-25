package main

import (
	"fmt"
	"log"
	"os"

	smp "github.com/netsys-lab/scion-multipath-lib/api"
	"github.com/netsys-lab/scion-multipath-lib/smp"
	"github.com/scionproto/scion/go/lib/snet"
)

func customPathSelectAlg(paths []snet.Path) ([]snet.Path, error) {
	paths1 := smp.SelectShortestPaths(5, paths)
	pathsToReturn := smp.SelectLowestLatencies(3, paths1)
	// pathsToReturn := []snet.Path{smp.SelectLowestLatency(paths)}
	return pathsToReturn, nil
}

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
