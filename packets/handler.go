package packets

import "github.com/scionproto/scion/go/lib/snet"

// Same as for PacketGenerator, however we need to elaborate
// Which types of information are really required here
type PacketHandlerMeta struct {
	LocalAddr snet.UDPAddr
}

type PacketHandler interface {
	SetMeta(PacketGenMeta)
	GetMetrics() PacketMetrics
	// Handle works the other way round than generate
	// We put a full SCION packet (encoded) in the Handle func
	// and receive only the payload
	// Metrics are collected here, because all SCION header information
	// are available
	Handle([]byte) ([]byte, error)
}
