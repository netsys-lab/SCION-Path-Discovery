package examples

import (
	"github.com/netsys-lab/scion-multipath-lib/smp"
)

// 0.0.2
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
	Connections             []smp.MonitoredConn //
	PathSelectionProperties []string            // TODO: Design a real struct for this, string is only dummy

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
	}
	return btMpSock.Peers
}

// TODO: Do we want to pass a particular peer here, or should the socket decide
// mgartner: I think the question from which peer a piece is requested is performed
// by Bittorrent itself, but BittorrentMultipathSock decides over which path(s)
// Furthermore, BittorrentMultipathSock could simply pass the message to the respective
// MPPeerSock
func (btMpSock BittorrentMultipathSock) RequestPiece(pieceIndex int, peer BittorrentPeer) {
	// Send Request Message to a particular peer
}

//
func (btMpSock BittorrentMultipathSock) ReceivePiece(pieceIndex int, result []byte) {
	// Download the actual piece from the peer passed to RequestPiece
}
