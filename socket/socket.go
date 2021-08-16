package socket

import (
	"github.com/netsys-lab/scion-path-discovery/packets"
	"github.com/scionproto/scion/go/lib/snet"
)

type UnderlaySocket interface {
	Listen() error
	Accept() (packets.TransportConn, error)
	AcceptAll() (*snet.UDPAddr, []packets.TransportConn, error)
	Dial(remote snet.UDPAddr, path snet.Path) (packets.TransportConn, error)
	DialAll(remote snet.UDPAddr, path []snet.Path) ([]packets.TransportConn, error)
	CloseAll() []error
	GetConnections() []packets.TransportConn
}
