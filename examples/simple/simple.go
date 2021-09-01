package main

import (
	"fmt"
	smp "github.com/netsys-lab/scion-path-discovery/api"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
	"log"
)

//LastSelection users could add more fields
type LastSelection struct {
	pathSet pathselection.PathSet
}

//NewFullPathSet contains all initially available paths
func NewFullPathSet(addr *snet.UDPAddr) (LastSelection, error) {
	pathSet, err := pathselection.QueryPaths(addr)
	return LastSelection{pathSet: pathSet}, err
}

//CustomPathSelectAlg this is where the user actually wants to implement its logic in
func (lastSel *LastSelection) CustomPathSelectAlg(pathSet *pathselection.PathSet) (*pathselection.PathSet, error) {
	return pathSet.GetPathHighBandwidth(3), nil
}

//GetPathSet must be implemented
func (lastSel *LastSelection) GetPathSet() *pathselection.PathSet {
	return &lastSel.pathSet
}

func main() {
	pathselection.InitHashMap()
	peers := []string{"18-ffaa:1:ef8,[127.0.0.1]:12345"} // Later real addresses
	local := "peer0"
	for _, peer := range peers {
		parsedAddr, _ := snet.ParseUDPAddr(peer)
		lastSelection, err := NewFullPathSet(parsedAddr)
		if err != nil {
			return
		}
		//example for DB Query
		//pathSetOutOfDB, err := pathselection.GetPathSet(parsedAddr)
		mpSock := smp.NewMPPeerSock(local, parsedAddr)
		err = mpSock.Connect(&lastSelection)
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
