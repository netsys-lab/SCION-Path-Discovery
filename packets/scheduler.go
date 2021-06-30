package packets

import (
	"github.com/netsys-lab/scion-multipath-lib/peers"
)

// PacketSchedulers have a list of "selected" paths which comes
// from the path selection module. These paths are considered
// "optimal" regarding the path selection properties that
// the application provides
// The scheduler decides for each packet over which of the optimal
// paths it will be sent
type PacketScheduler interface {
	SetPathlevelPeers(peers []peers.PathlevelPeer)
	Write(data []byte, peer string)
	Read(data []byte, peer string)
}

// Implements a PacketScheduler that calculates weights out of
// PathQualities and sends packets depending on the weight of
// each alternative
type WeighedScheduler struct {
	packetGen     PacketGen
	packetHandler PacketHandler
	peers         []peers.PathlevelPeer
}
