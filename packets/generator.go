package packets

import "github.com/scionproto/scion/go/lib/snet"

// Information required to create SCION common and address headers
// Path headers are then created using the path provided in the Generate func
type PacketGenMeta struct {
	LocalAddr  snet.UDPAddr
	RemoteAddr snet.UDPAddr
}

// The Packetgenerator needs previously presented meta
// Furthermore, it collects metrics under the hood that can
// be queried using GetMetrics
// This is probably done by the MPPeerSock itself, who can then
// insert the metrics into the QualityDatabase
// To avoid having one Generator per path, the path can be passed
// individually for each packet
type PacketGen interface {
	SetMeta(metrics PacketGenMeta)
	GetMetrics() PacketMetrics
	// GetMetrics() PacketMetrics
	// Creating a SCION packet out of the provided payload
	// returns the full SCION packet in a new byte slice
	// Metrics are collected here, because all SCION header information
	// are available
	Generate(payload []byte, path snet.Path) ([]byte, error)
}
