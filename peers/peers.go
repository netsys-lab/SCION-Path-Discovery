package peers

import (
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
)

// A peer is basically a string containing the SCION address
// Although this lib could collect per peer information, it's not the core idea
// of a multipath library. We work only on path-level not on peer-level, because
// we assume that applications mostly need a connection interface to a particular peer,
// not a connection interface to all available peers
type Peer string

// Pathlevel peers contain the original peer address, path information and also quality information
type PathlevelPeer struct {
	Peer        string
	PeerAddr    snet.UDPAddr
	PathQuality pathselection.PathQuality
}
