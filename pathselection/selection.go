package pathselection

import (
	"io"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/snet"
)

type PathAlternatives struct {
	HostAddr addr.HostAddr
	Paths    []snet.Path
}

type PathEnumerator interface {
	Enumerate(addr.HostAddr) PathAlternatives
}

type PathQuality struct {
	Timestamp time.Time
	RTT       time.Duration
	Bytes     int
	Duration  time.Duration
	Path      snet.Path
}

type QualityDB interface {
	GetPathHighBandwidth(addr.HostAddr) snet.Path
	GetPathLowLatency(addr.HostAddr) snet.Path
	//GetPathFunc takes as second argument a function that is
	//then called recursively over all PathQuality pairs, always
	//retaining the returned result as the first input for the
	//next call. The path associated with the last returned
	//PathQuality struct is then picked.
	GetPathFunc(addr.HostAddr, func(PathQuality, PathQuality) PathQuality) snet.Path
	//GetPathCustom takes as second argument a function that is
	//called with the PathQuality array of all the alternative
	//paths for the host address. The path associated with the
	//returned PathQuality is then returned
	GetPathCustom(addr.HostAddr, func([]PathQuality) PathQuality) snet.Path

	AddPathAlternatives(PathAlternatives) error
	AddPathQuality(PathQuality) error
}

type MeasuringReaderWriter interface {
	io.Reader
	io.Writer
	Measure(snet.Path) chan PathQuality
}
