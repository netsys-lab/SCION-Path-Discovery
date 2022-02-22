package lookup

import (
	"github.com/netsys-lab/scion-path-discovery/sutils"
	"github.com/scionproto/scion/go/lib/snet"
)

// This wraps the usage of appnet to query paths
// May later be put into an interface or struct
func PathLookup(peer string) ([]snet.Path, error) {
	udpAddr, err := sutils.ResolveUDPAddr(peer)
	if err != nil {
		return nil, err
	}
	paths, err := sutils.QueryPaths(udpAddr)
	if err != nil {
		return nil, err
	}

	return paths, nil
}
