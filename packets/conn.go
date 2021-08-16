package packets

import (
	"github.com/scionproto/scion/go/lib/snet"
)

// TODO: Remove to configurable options
const (
	PACKET_SIZE = 1400
)

type TransportConn interface {
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	Listen(snet.UDPAddr) error
	// Accept() (*peers.PathlevelPeer, error)

	Connect(snet.UDPAddr, *snet.Path) error
	Close() error
	// GetPeer() *peers.PathlevelPeer
	GetMetrics() *PacketMetrics
}

type ReliableTransportConn interface {
	TransportConn
	WriteStream([]byte) (int, error)
	ReadStream([]byte) (int, error)
}

type TransportConstructor func() TransportConn
