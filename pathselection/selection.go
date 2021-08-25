package pathselection

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"io"
	"strings"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/snet"
)

var PathSetDB []PathSet
var hashMap map[string] int

type PathSet struct {
	Address snet.UDPAddr
	Paths   []PathQuality
}

func asSha256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func calcAddrHash(addr *snet.UDPAddr) string {
	var partHash strings.Builder
	h1 := asSha256(addr.IA.I.String())
	h2 := asSha256(addr.IA.A.String())
	partHash.WriteString(h1)
	partHash.WriteString(h2)
	return asSha256(partHash.String())
}

func GetPathSet(addr *snet.UDPAddr) (PathSet, error) {
	hash := calcAddrHash(addr)
	index, contained := hashMap[hash]
	if contained {
		return PathSetDB[index], nil
	} else {
		return PathSet{}, errors.New("404")
	}
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

func InitHashMap() {
	hashMap = make(map[string]int)
}

func QueryPaths(addr *snet.UDPAddr) (PathSet, error) {
	paths, err := appnet.DefNetwork().PathQuerier.Query(context.Background(), addr.IA)
	if err != nil {
		return PathSet{}, err
	}
	var pathQualities []PathQuality
	for _, path := range paths {
		pathEntry := PathQuality{Path: path}
		pathQualities = append(pathQualities, pathEntry)
	}
	tmpPathSet := PathSet{Address: *addr, Paths: pathQualities}

	if i, contained := hashMap[calcAddrHash(addr)]; contained {
		//update PathSetDB entry if already existing
		PathSetDB[i] = tmpPathSet
	} else {
		PathSetDB = append(PathSetDB, tmpPathSet)
		hash := calcAddrHash(addr)
		hashMap[hash] = len(PathSetDB) - 1
	}
	return tmpPathSet, nil
}
