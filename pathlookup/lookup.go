package lookup

import (
	"context"

	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
)

// This wraps the usage of appnet to query paths
// May later be put into an interface or struct
func PathLookup(peer string) ([]snet.Path, error) {
	udpAddr, err := appnet.ResolveUDPAddr(peer)
	if err != nil {
		return nil, err
	}
	paths, err := appnet.DefNetwork().PathQuerier.Query(context.Background(), udpAddr.IA)
	if err != nil {
		return nil, err
	}

	return paths, nil
}
