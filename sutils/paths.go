package sutils

import (
	"context"
	"errors"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/daemon"
	"github.com/scionproto/scion/go/lib/snet"
)

func SetPath(addr *snet.UDPAddr, path *snet.Path) {
	addr.Path = (*path).Path()
	addr.NextHop = (*path).UnderlayNextHop()
}

func queryPaths(sciond daemon.Connector, ctx context.Context, dst *snet.UDPAddr) ([]snet.Path, error) {
	flags := daemon.PathReqFlags{Refresh: false, Hidden: false}
	snetPaths, err := sciond.Paths(ctx, dst.IA, addr.IA{}, flags)
	return snetPaths, err
}

func setDefaultPath(sciond daemon.Connector, ctx context.Context, dst *snet.UDPAddr) error {
	paths, err := queryPaths(sciond, ctx, dst)
	if err != nil {
		return err
	}
	if len(paths) > 0 {
		dst.Path = paths[0].Path()
		dst.NextHop = paths[0].UnderlayNextHop()
		return nil
	}

	return errors.New("No path found")
}

func SetDefaultPath(addr *snet.UDPAddr) error {
	daemonConn, err := findSciond(context.Background())
	if err != nil {
		return err
	}
	return setDefaultPath(daemonConn, context.Background(), addr)
}

func QueryPaths(addr *snet.UDPAddr) ([]snet.Path, error) {
	daemonConn, err := findSciond(context.Background())
	if err != nil {
		return nil, err
	}

	return queryPaths(daemonConn, context.Background(), addr)
}
