package packets

import (
	"github.com/netsys-lab/scion-path-discovery/peers"
	"github.com/scionproto/scion/go/lib/snet"
)

// PacketSchedulers have a list of "selected" paths which comes
// from the path selection module. These paths are considered
// "optimal" regarding the path selection properties that
// the application provides
// The scheduler decides for each packet over which of the optimal
// paths it will be sent
type PacketScheduler interface {
	SetPathlevelPeers(peers []peers.PathlevelPeer)
	SetPathQualities([]snet.Path) error
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	WriteStream([]byte) (int, error)
	ReadStream([]byte) (int, error)
	ConnectPeers() error
}

// Implements a PacketScheduler that calculates weights out of
// PathQualities and sends packets depending on the weight of
// each alternative
type WeighedScheduler struct {
	peers       []peers.PathlevelPeer
	connections []TransportConn
}
