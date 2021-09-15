package packets

import (
	"github.com/scionproto/scion/go/lib/snet"
)

// TODO: Remove to configurable options
const (
	PACKET_SIZE = 1400
)

func ConnTypeToString(connType int) string {
	switch connType {
	case ConnectionTypes.Bidirectional:
		return "bidirectional"

	case ConnectionTypes.Outgoing:
		return "outgoing"

	case ConnectionTypes.Incoming:
		return "incoming"
	}

	return ""
}

var ConnectionTypes = newConnectionTypes()

func newConnectionTypes() *connectionTypes {
	return &connectionTypes{
		Incoming:      1,
		Outgoing:      2,
		Bidirectional: 3,
	}
}

type connectionTypes struct {
	Incoming      int
	Outgoing      int
	Bidirectional int
}

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
	GetType() int
}

type TransportConstructor func() UDPConn
