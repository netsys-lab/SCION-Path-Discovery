package socket

import (
	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/scionproto/scion/go/lib/snet"
)

type UnderlaySocket interface {
	Listen() error
	WaitForDialIn() (packets.UDPConn, *snet.UDPAddr, error)
	Dial(remote snet.UDPAddr, path snet.Path) (packets.UDPConn, error)
	DialAll(remote snet.UDPAddr, path []snet.Path) ([]packets.UDPConn, error)
	CloseAll() []error
	GetConnections() []packets.UDPConn
}
