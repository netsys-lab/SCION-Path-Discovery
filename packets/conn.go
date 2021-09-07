package packets

import (
	"github.com/scionproto/scion/go/lib/snet"
)

// TODO: Remove to configurable options
const (
	PACKET_SIZE = 1400
)

type UDPConn interface {
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	Listen(snet.UDPAddr) error
	Dial(snet.UDPAddr, *snet.Path) error
	Close() error
	GetMetrics() *PathMetrics
	GetPath() *snet.Path
	GetRemote() *snet.UDPAddr
	SetLocal(snet.UDPAddr)
	WriteStream([]byte) (int, error)
	ReadStream([]byte) (int, error)
}

type TransportConstructor func() UDPConn
