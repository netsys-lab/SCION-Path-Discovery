package socket

import (
	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/scionproto/scion/go/lib/snet"
)

type DialOptions struct {
	SendAddrPacket bool
	NumPaths       int
}

type UnderlaySocket interface {
	Listen() error
	WaitForDialIn() (*snet.UDPAddr, error)
	WaitForIncomingConn() (packets.UDPConn, error)
	Dial(remote snet.UDPAddr, path snet.Path, options DialOptions) (packets.UDPConn, error)
	DialAll(remote snet.UDPAddr, path []snet.Path, options DialOptions) ([]packets.UDPConn, error)
	CloseAll() []error
	GetConnections() []packets.UDPConn
}
