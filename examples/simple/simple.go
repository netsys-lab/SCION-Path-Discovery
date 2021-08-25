package main

import (
	"fmt"
	smp "github.com/netsys-lab/scion-path-discovery/api"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
	"log"
)

//CurrentSelection this struct can stay in here, users could add more fields
type CurrentSelection struct {
	PathSet pathselection.PathSet
}

// NewFullPathSet contain all initially available paths
func NewFullPathSet(addr *snet.UDPAddr) (CurrentSelection, error) {
	pathSet, err := pathselection.QueryPaths(addr)
	return CurrentSelection{PathSet: pathSet}, err
}

//CustomPathSelectAlg this is where the user actually wants to implement its logic in
func (currSel *CurrentSelection) CustomPathSelectAlg() {
	currSel.PathSet.GetPathLargeMTU(3)
}



func main() {
	pathselection.InitHashMap()
	peers := []string{"18-ffaa:1:ef8,[127.0.0.1]:12345"} // Later real addresses
	local := "peer0"
	for _, peer := range peers {
		parsedAddr, _ := snet.ParseUDPAddr(peer)
		currentSelection, err := NewFullPathSet(parsedAddr)
		if err != nil {
			return
		}

		//example for DB Query
		//db, _ := pathselection.GetPathSet(parsedAddr)

		mpSock := smp.NewMPPeerSock(local, parsedAddr)
		currentSelection.CustomPathSelectAlg()
		err = mpSock.Connect(currentSelection.PathSet)
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
			fmt.Printf("Read %d bytes of data from %s\n", n, mpSock.Local)
		}(mpSock)

		data := make([]byte, 1200)
		n, err := mpSock.Write(data)
		if err != nil {
			log.Fatal("Failed to connect MPPeerSock", err)
		}
		fmt.Printf("Wrote %d bytes of data to %s\n", n, mpSock.Peer)
	}
}
