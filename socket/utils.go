package socket

import "github.com/netsys-lab/scion-path-discovery/packets"

func ConnAlreadyOpen(conns []packets.UDPConn, id string) bool {
	for _, v := range conns {
		if v.GetId() == id {
			return true
		}
	}

	return false
}
