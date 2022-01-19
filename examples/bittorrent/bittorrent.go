package examples

// Note: This is deprecated, the actual implementation can be found here: https://github.com/netsys-lab/bittorrent-over-scion

/*
// This is a Bittorrent specific abstraction which will later be put
// into Bittorrent code. It should only present how the smp API
// could be integrated into Bittorrent.
type BittorrentPeer struct {
	Addr        string
	PieceBitmap []byte // Represents which pieces the peer has to download
}

// This represents a high level abstraction of a Bittorrent specific socket that connects to several Peers
// It provides functions for applying a new peer set, requesting and fetching pieces,
// and performing peer lookup
type BittorrentMultipathSock struct {
	Peers                   []BittorrentPeer
	TrackerUrl              string
	PeerSocks               []smp.MPPeerSock
	PathSelectionProperties []string // TODO: Design a real struct for this, string is only dummy

}

func NewBittorrentMultipathSock(trackerUrl string) *BittorrentMultipathSock {
	return &BittorrentMultipathSock{
		TrackerUrl: trackerUrl,
	}
}

// This would need to return the peers to Bittorrent
// So it has knowledge about all requesting peers
func (btMpSock BittorrentMultipathSock) PeerLokup() []BittorrentPeer {
	btMpSock.Peers = []BittorrentPeer{
		{Addr: "Peer1", PieceBitmap: make([]byte, 0)},
		{Addr: "Peer2", PieceBitmap: make([]byte, 0)},
		{Addr: "Peer3", PieceBitmap: make([]byte, 0)},
	}
	return btMpSock.Peers
}

// TODO: Do we want to pass a particular peer here, or should the socket decide
// mgartner: I think the question from which peer a piece is requested is performed
// by Bittorrent itself, but BittorrentMultipathSock decides over which path(s)
// Furthermore, BittorrentMultipathSock could simply pass the message to the respective
// MPPeerSock
func (btMpSock BittorrentMultipathSock) RequestPiece(pieceIndex int, peerIndex int) error {
	// Send Request Message to a particular peer
	pieceMsg := make([]byte, 1200)
	// Encode pieceIndex here into pieceMsg
	_, err := btMpSock.PeerSocks[peerIndex].Write(pieceMsg)
	return err
}

//
func (btMpSock BittorrentMultipathSock) ReceivePiece(pieceIndex int, peerIndex int, result []byte) error {
	// Download the actual piece from the peer passed to RequestPiece
	// Here we will call Read on the respective MPPeerSock
	// Of course, here happens more magic than simply read a packet
	// But this will be covered later
	_, err := btMpSock.PeerSocks[peerIndex].Read(result)
	return err
}

func (btMpSock BittorrentMultipathSock) SendPiece(pieceIndex int, peerIndex int, result []byte) error {
	// Download the actual piece from the peer passed to RequestPiece
	// Here we will call Read on the respective MPPeerSock
	// Of course, here happens more magic than simply read a packet
	// But this will be covered later
	_, err := btMpSock.PeerSocks[peerIndex].Write(result)
	return err
}

// Note: This main is only pointing out the flow how Bittorrent can
// use the smp library, but should not be a complete working example
func main() {
	peers := []string{"peer1", "peer2", "peer3"} // Later real addresses

	btmp := NewBittorrentMultipathSock("tracker")
	btmp.PeerLokup()

	pieces := []int{0, 1, 2, 3, 5, 6}
	peerIndex := 0
	go func() {
		for _, p := range pieces {
			piece := make([]byte, 1200)
			err := btmp.RequestPiece(p, peerIndex)
			if err != nil {
				log.Fatal("Failed to connect MPPeerSock", err)
				os.Exit(1)
			}

			err = btmp.ReceivePiece(p, peerIndex, piece)
			if err != nil {
				log.Fatal("Failed to connect MPPeerSock", err)
				os.Exit(1)
			}

			if peerIndex == len(peers)-1 {
				peerIndex = 0
			}
		}
	}()

	for _, p := range pieces {
		piece := make([]byte, 1200)
		err := btmp.SendPiece(p, peerIndex, piece)
		if err != nil {
			log.Fatal("Failed to connect MPPeerSock", err)
			os.Exit(1)
		}

		if peerIndex == len(peers)-1 {
			peerIndex = 0
		}
	}
}
*/
