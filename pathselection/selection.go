package pathselection

import (
	"context"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"io"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/snet"
)

var PathSetDB []PathSet

type PathSet struct {
	Address snet.UDPAddr
	Paths   []PathQuality
}

func GetPathSet(udpAddr snet.UDPAddr) (PathSet, error) {
	i := 0
	return PathSetDB[i], nil
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
	GetPathHighBandwidth(number int) PathSet
	GetPathLowLatency(number int) PathSet
	GetPathLargeMTU(number int) PathSet
	GetPathSmallHopCount(number int) PathSet

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
}

type MeasuringReaderWriter interface {
	io.Reader
	io.Writer
	Measure(snet.Path) chan PathQuality
}

func NewPathSet() QualityDB {
	//return &PathSet{}
	return nil
}

func SelectPaths(count int, pathSet *PathSet) (newPathSet *PathSet) {
	lenPaths := len(pathSet.Paths)
	numPathsToReturn := 0
	if lenPaths > 0 {
		if lenPaths < count {
			numPathsToReturn = lenPaths
		} else {
			numPathsToReturn = count
		}
		for i := 0; i < numPathsToReturn; i++ {
			newPathSet.Paths = append(newPathSet.Paths, pathSet.Paths[i])
		}
	}
	return newPathSet
}

type CustomPathSelection interface {
	CustomPathSelectAlg()
}

func QueryPaths(addr *snet.UDPAddr) PathSet {
	paths, _ := appnet.DefNetwork().PathQuerier.Query(context.Background(), addr.IA)
	var pathQualities []PathQuality
	for _, path := range paths {
		pathEntry := PathQuality{Path: path}
		pathQualities = append(pathQualities, pathEntry)
	}
	tmpPathSet := PathSet{Address: *addr, Paths: pathQualities}
	PathSetDB = append(PathSetDB, tmpPathSet)
	return tmpPathSet
}
