package socket

import "net"

// TODO: extend this further. It may be useful to use more than
// one native UDP socket due to performance limitations
type Socket interface {
	net.Conn
}
