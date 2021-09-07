package pathselection

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/netsys-lab/scion-path-discovery/packets"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/snet"
)

type PathSet struct {
	Address snet.UDPAddr
	Paths   []PathQuality
}

type PathEnumerator interface {
	Enumerate(addr.HostAddr) PathSet
}

type PathQuality struct {
	packets.PathMetrics
	Timestamp time.Time
	HopCount  int
	MTU       uint16
	Latency   time.Duration
	RTT       time.Duration
	Bytes     int
	Duration  time.Duration
	Path      snet.Path
}

type SelecteablePathSet interface {
	GetPathHighBandwidth(number int) PathSet
	GetPathLowLatency(number int) PathSet
	GetPathLargeMTU(number int) PathSet
	GetPathSmallHopCount(number int) PathSet
}

type PathQualityDatabase interface {
	GetPathSet(addr *snet.UDPAddr) (PathSet, error)
	SetListenConnections([]packets.UDPConn)
	SetDialConnections([]packets.UDPConn)
	UpdatePathQualities(addr *snet.UDPAddr) error
	UpdateMetrics()

	// TODO: Rethink those...
	//GetPathFunc takes as second argument a function that is
	//then called recursively over all PathQuality pairs, always
	//retaining the returned result as the first input for the
	//next call. The path associated with the last returned
	//PathQuality struct is then picked.
	// GetPathFunc(addr.HostAddr, func(PathQuality, PathQuality) PathQuality) snet.Path
	//GetPathCustom takes as second argument a function that is
	//called with the PathQuality array of all the alternative
	//paths for the host address. The path associated with the
	//returned PathQuality is then returned
	// GetPathCustom(addr.HostAddr, func([]PathQuality) PathQuality) snet.Path
}

type InMemoryPathQualityDatabase struct {
	pathSetDB   []PathSet
	hashMap     map[string]int
	listenConns []packets.UDPConn
	dialConns   []packets.UDPConn
}

func (db *InMemoryPathQualityDatabase) SetListenConnections(conns []packets.UDPConn) {
	db.listenConns = conns
}

func (db *InMemoryPathQualityDatabase) SetDialConnections(conns []packets.UDPConn) {
	db.dialConns = conns
}

func (db *InMemoryPathQualityDatabase) UpdateMetrics() {
	// TODO: Do listen Cons have paths?
	for _, v := range db.listenConns {
		pathQuality, err := db.getPathQuality(v.GetRemote(), v.GetPath())
		if err != nil {
			log.Fatal(err)
		}

		connMetrics := v.GetMetrics()
		pathQuality.ReadBytes += connMetrics.ReadBytes
		pathQuality.WrittenBytes += connMetrics.WrittenBytes
		pathQuality.ReadPackets += connMetrics.ReadPackets
		pathQuality.WrittenPackets += connMetrics.WrittenPackets
	}

	for _, v := range db.dialConns {
		pathQuality, err := db.getPathQuality(v.GetRemote(), v.GetPath())
		if err != nil {
			log.Fatal(err)
		}

		connMetrics := v.GetMetrics()
		pathQuality.ReadBytes += connMetrics.ReadBytes
		pathQuality.WrittenBytes += connMetrics.WrittenBytes
		pathQuality.ReadPackets += connMetrics.ReadPackets
		pathQuality.WrittenPackets += connMetrics.WrittenPackets
	}
}

func (db *InMemoryPathQualityDatabase) getPathQuality(addr *snet.UDPAddr, path *snet.Path) (*PathQuality, error) {
	var pathQuality *PathQuality
	pathSet, err := db.GetPathSet(addr)
	if err != nil {
		return nil, err
	}

	for _, v := range pathSet.Paths {
		if bytes.Compare(v.Path.Path().Raw, (*path).Path().Raw) == 0 {
			pathQuality = &v
		}
	}

	if pathQuality == nil {
		return nil, errors.New(fmt.Sprintf("No PathQuality found for path"))
	}

	return pathQuality, nil

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

func (db *InMemoryPathQualityDatabase) GetPathSet(addr *snet.UDPAddr) (PathSet, error) {
	hash := calcAddrHash(addr)
	index, contained := db.hashMap[hash]
	if contained {
		return db.pathSetDB[index], nil
	} else {
		return PathSet{}, errors.New("404")
	}
}

/*
type MeasuringReaderWriter interface {
	io.Reader
	io.Writer
	Measure(snet.Path) chan PathQuality
}
*/
//TODO: can be removed?
//func NewPathSet() QualityDB {
//	//return &PathSet{}
//	return nil
//}

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
	CustomPathSelectAlg(*PathSet) (*PathSet, error)
}

func NewInMemoryPathQualityDatabase() *InMemoryPathQualityDatabase {
	return &InMemoryPathQualityDatabase{
		hashMap: make(map[string]int),
	}
}

func (db *InMemoryPathQualityDatabase) UpdatePathQualities(addr *snet.UDPAddr) error {
	paths, err := appnet.DefNetwork().PathQuerier.Query(context.Background(), addr.IA)
	if err != nil {
		return err
	}
	var pathQualities []PathQuality
	for _, path := range paths {
		pathEntry := PathQuality{Path: path}
		pathQualities = append(pathQualities, pathEntry)
	}
	tmpPathSet := PathSet{Address: *addr, Paths: pathQualities}

	if i, contained := db.hashMap[calcAddrHash(addr)]; contained {
		//update PathSetDB entry if already existing
		db.pathSetDB[i] = tmpPathSet
	} else {
		db.pathSetDB = append(db.pathSetDB, tmpPathSet)
		hash := calcAddrHash(addr)
		db.hashMap[hash] = len(db.pathSetDB) - 1
	}
	return nil
}

func UnwrapPathset(pathset PathSet) []snet.Path {
	paths := make([]snet.Path, 0)
	for _, p := range pathset.Paths {
		paths = append(paths, p.Path)
	}

	return paths
}
