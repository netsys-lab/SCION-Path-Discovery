package pathselection

import (
	"io"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/snet"
)

type PathSet struct {
	Address snet.UDPAddr
	Paths   []PathQuality
}

func (pathSet *PathSet) GetPathSet(udpAddr snet.UDPAddr) (PathSet, error) {
	panic("implement me")
}

func (pathSet *PathSet) AddPathAlternatives(set PathSet) error {
	panic("implement me")
}

func (pathSet *PathSet) AddPathQuality(quality PathQuality) error {
	panic("implement me")
}

func (pathSet *PathSet) GetPathFunc(hostAddr addr.HostAddr, f func(PathQuality, PathQuality) PathQuality) snet.Path {
	panic("implement me")
}

func (pathSet *PathSet) GetPathCustom(hostAddr addr.HostAddr, f func([]PathQuality) PathQuality) snet.Path {
	panic("implement me")
}



type PathEnumerator interface {
	Enumerate(addr.HostAddr) PathSet
}

type PathQuality struct {
	Timestamp time.Time
	HopCount  int
	MTU       uint16
	Latency   time.Duration
	RTT       time.Duration
	Bytes     int
	Duration  time.Duration
	Path      snet.Path
}

type QualityDB interface {
	GetPathHighBandwidth() snet.Path
	GetPathLowLatency() snet.Path
	GetPathLargeMTU() snet.Path
	GetPathSmallHopCount() snet.Path

	GetPathSet(snet.UDPAddr) (PathSet, error)

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

	AddPathAlternatives(PathSet) error
	AddPathQuality(PathQuality) error
}

type MeasuringReaderWriter interface {
	io.Reader
	io.Writer
	Measure(snet.Path) chan PathQuality
}

func NewPathSet() QualityDB {
	return &PathSet{}
	//return nil
}

func SelectPaths(count int, paths []PathQuality) (selectedPathSet []snet.Path) {
	lenPaths := len(paths)
	numPathsToReturn := 0
	if lenPaths > 0 {
		if lenPaths < count {
			numPathsToReturn = lenPaths
		} else {
			numPathsToReturn = count
		}
		for i := 0; i < numPathsToReturn; i++ {
			selectedPathSet = append(selectedPathSet, paths[i].Path)
		}
	}
	return selectedPathSet
}