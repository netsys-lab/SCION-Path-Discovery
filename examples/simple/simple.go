package main

import (
	"fmt"
	"log"

	smp "github.com/netsys-lab/scion-path-discovery/api"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
)

//#####################################################################################################################################

//NewCurrentSelection
//TODO should be in selection.go so users do not have to implement it themselves
func NewCurrentSelection(pathSet pathselection.PathSet) (*pathselection.PathSet, error) {
	asdf := CurrentSelection{pathSet}
	qwer, nil := asdf.CustomPathSelectAlg()
	return qwer, nil
}

//CurrentSelection this struct can stay in here, user could add more fields
type CurrentSelection struct {
	PathSet pathselection.PathSet
}

//CustomPathSelectAlg this is where the user actually wants to implement its logic in
func (currSel CurrentSelection) CustomPathSelectAlg() (*pathselection.PathSet, error) {
	newPathSet := currSel.PathSet.GetPathLargeMTU(3)
	newPathSet.GetPathLargeMTU(3)
	return newPathSet, nil
}

//customPathSelectAlg legacy, only for line 64
func customPathSelectAlg(snet.UDPAddr, []pathselection.PathQuality) ([]pathselection.PathQuality, error) {
	return nil, nil
}

func main() {
	fullPathSet, err := pathselection.GetPathSet(snet.UDPAddr{})
	if err != nil {
		return
	}
	selectedPathSet, err := NewCurrentSelection(fullPathSet)

	// only so that selectedPathSet is used
	selectedPathSet.GetPathLargeMTU(5)

//#####################################################################################################################################




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
