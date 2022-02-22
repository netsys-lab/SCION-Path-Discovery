package sutils

import (
	"github.com/netsec-ethz/scion-apps/pkg/pan"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/snet"
	"inet.af/netaddr"
)

func ResolveUDPAddr(addr string) (*snet.UDPAddr, error) {
	laddr, err := pan.ResolveUDPAddr(addr)
	if err != nil {
		return nil, err
	}
	return PanToSnetUDPAddr(laddr), nil
}

func PanToSnetUDPAddr(a pan.UDPAddr) *snet.UDPAddr {
	return &snet.UDPAddr{
		IA:   addr.IA(a.IA),
		Host: netaddr.IPPortFrom(a.IP, a.Port).UDPAddr(),
	}
}
