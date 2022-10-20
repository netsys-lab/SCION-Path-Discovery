package smp

import (
	"sort"
	"strings"

	"github.com/netsys-lab/scion-path-discovery/packets"
	lookup "github.com/netsys-lab/scion-path-discovery/pathlookup"
	"github.com/netsys-lab/scion-path-discovery/pathselection"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/sirupsen/logrus"
)

func NewDisjointPathSelectionSocket(remote *PanSocket, numConns, numExploreConns int) *DisjointPathselection {
	return &DisjointPathselection{
		remote:          remote,
		NumExploreConns: numExploreConns,
		NumConns:        numConns,
		metricsMap:      make(map[string]*packets.PathMetrics),
	}
}

func numPathsConflict(path1, path2 snet.Path) int {
	path1Interfaces := path1.Metadata().Interfaces
	path2Interfaces := path2.Metadata().Interfaces
	conflicts := 0
	for i, intP1 := range path1Interfaces {
		for j, intP2 := range path2Interfaces {
			if i == 0 && j == 0 {
				continue
			}
			if i == (len(path1Interfaces)-1) && j == (len(path2Interfaces)-1) {
				continue
			}
			if intP1.IA.Equal(intP2.IA) && intP1.ID == intP2.ID {
				conflicts++
			}
		}
	}
	return conflicts
}

func (dj *DisjointPathselection) GetPathConflictEntries() ([]PathWrap, error) {
	// Put all paths in one list and sort them according to number of hops
	allPaths := make([]PathWrap, 0)
	paths, err := lookup.PathLookup(dj.remote.Peer.String())
	if err != nil {
		return nil, err
	}

	for _, pp := range paths {
		pw := PathWrap{
			Address: *dj.remote.Peer,
			Path:    pp,
		}
		allPaths = append(allPaths, pw)
	}

	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i].Path.Metadata().Interfaces) < len(allPaths[j].Path.Metadata().Interfaces)
	})

	// Add conflicts
	for i, path := range allPaths {
		for _, path2 := range allPaths[i:] {
			curConflicts := numPathsConflict(path.Path, path2.Path)
			path.NumConflicts += curConflicts
		}
	}

	sort.Slice(allPaths, func(i, j int) bool {
		return allPaths[i].NumConflicts < allPaths[j].NumConflicts
	})

	logrus.Debug("[DisjointPathselection] Updated PathConflictEntries to remote ", dj.remote.Peer.String(), " got ", len(allPaths), " entries")

	return allPaths, nil
}

type PathWrap struct {
	Address      snet.UDPAddr
	Path         snet.Path
	NumConflicts int
}

type DisjointPathselection struct {
	remote                   *PanSocket
	NumExploreConns          int
	NumConns                 int
	metricsMap               map[string]*packets.PathMetrics // Indicates the performance of the particular pathset -> id = path1|path2|path3 etc
	latestBestWriteBandwidth int64
	numUpdates               int64
	latestPathSet            []snet.Path
	currentPathSet           []snet.Path
}

// Perm calls f with each permutation of a.
func Perm(a []string, f func([]string)) {
	perm(a, f, 0)
}

// Permute the values at index i to len(a)-1.
func perm(a []string, f func([]string), i int) {
	if i > len(a) {
		f(a)
		return
	}
	perm(a, f, i+1)
	for j := i + 1; j < len(a); j++ {
		a[i], a[j] = a[j], a[i]
		perm(a, f, i+1)
		a[i], a[j] = a[j], a[i]
	}
}

func (dj *DisjointPathselection) GetNextProbingPathset() (pathselection.PathSet, error) {
	logrus.Debug("[DisjointPathselection] GetNextProbingPathSet called")
	alreadyCheckedPathsets := make([]string, 0)
	for k, _ := range dj.metricsMap {
		alreadyCheckedPathsets = append(alreadyCheckedPathsets, k)
	}

	conflictPaths, err := dj.GetPathConflictEntries()
	if err != nil {
		return pathselection.PathSet{}, err
	}

	psId := ""
	defaultPsId := ""

	fixedPaths := dj.NumConns - dj.NumExploreConns
	for i := 0; i < fixedPaths; i++ {
		defaultPsId += lookup.PathToString(conflictPaths[i].Path) + "|"
	}

	matchingPsId := ""
	remainingPaths := conflictPaths[fixedPaths:]
	remainginPathIds := make([]string, 0)
	for _, p := range remainingPaths {
		remainginPathIds = append(remainginPathIds, lookup.PathToString(p.Path))
	}

	matchingPs := pathselection.PathSet{}
	Perm(remainginPathIds, func(s []string) {
		psId = defaultPsId
		for i, v := range s {
			if i < (dj.NumConns - fixedPaths) {
				psId = psId + v + "|"
			}

		}

		parts := strings.Split(psId, "|")
		isPathNewForAlreadyChecked := make([]bool, 0)
		for _, p := range alreadyCheckedPathsets {
			newP := false // One path is new
			for _, v := range parts {
				if !strings.Contains(p, v) {
					newP = true
					break
				}
			}
			isPathNewForAlreadyChecked = append(isPathNewForAlreadyChecked, newP)
		}

		pathCombinationWasUsedBefore := false
		// If at least one entry is false, then this one is not new
		for _, newP := range isPathNewForAlreadyChecked {
			if !newP {
				pathCombinationWasUsedBefore = true
				break
			}
		}

		if !pathCombinationWasUsedBefore && matchingPsId == "" {
			matchingPsId = psId
			logrus.Debug("[DisjointPathselection] Found new Pathset to evaluate: ", psId)
			paths := make([]snet.Path, 0)

			for k := 0; k < fixedPaths; k++ {
				paths = append(paths, conflictPaths[k].Path)
			}

			for i, pId := range s {
				if i < (dj.NumConns - fixedPaths) {
					for _, p := range conflictPaths {
						if lookup.PathToString(p.Path) == pId {
							paths = append(paths, p.Path)
							break
						}
					}
				}
			}

			matchingPs = pathselection.WrapPathset(paths)
		}
	})

	logrus.Error(matchingPs)

	if matchingPsId != "" {
		return matchingPs, nil
	}

	logrus.Debug("[DisjointPathselection] No new Pathset found")

	return pathselection.PathSet{}, nil
}

func (dj *DisjointPathselection) InitialPathset() (pathselection.PathSet, error) {
	// Here we have our new path set, from which we start
	dj.latestPathSet = dj.currentPathSet
	// Explore new
	ps, err := dj.GetNextProbingPathset()
	if err != nil {
		return ps, err
	}
	logrus.Debug("[DisjointPathSelection] Initial paths: ", ps.Paths)
	return ps, nil
}

func (dj *DisjointPathselection) UpdatePathSelection() (bool, error) {
	if dj.remote == nil {
		return false, nil
	}
	pathSet := dj.remote.GetCurrentPathset()
	currentId := ""
	for _, p := range pathSet.Paths {
		currentId += lookup.PathToString(p.SnetPath) + "|"
	}

	logrus.Trace("Looking up id ", currentId)

	newMetrics := dj.remote.AggregateMetrics()
	dj.metricsMap[currentId] = newMetrics

	logrus.Debug("[DisjointPathselection] UpdatePathSelection called")
	dj.numUpdates++

	// Compare to best, to make socket re-dial to improve performance
	if dj.numUpdates%5 == 0 {
		logrus.Error("[DisjointPathselection] Comparing old bw ", dj.latestBestWriteBandwidth, " to ", newMetrics.LastAverageWriteBandwidth(5))
		// TODO: This is not working properly here...
		if newMetrics.LastAverageWriteBandwidth(5) > dj.latestBestWriteBandwidth {
			logrus.Debug("[DisjointPathselection] Got better pathset, reconnecting")
			dj.latestBestWriteBandwidth = newMetrics.LastAverageWriteBandwidth(5)

			// Here we have our new path set, from which we start
			dj.latestPathSet = dj.currentPathSet
		}

		// Explore new
		ps, err := dj.GetNextProbingPathset()
		if err != nil {
			return false, err
		}

		logrus.Debug("[DisjointPathselection] Got new pathset, applying paths...")
		paths := pathselection.UnwrapPathset(ps)

		if len(ps.Paths) == 0 {
			paths = dj.latestPathSet
		}
		logrus.Warn(paths)
		conns := dj.remote.UnderlaySocket.GetConnections()
		if len(paths) < len(conns) {
			logrus.Warn("[DisjointPathSelection] Invalid pathset found...")
			return false, nil
		}
		// Here the number of connections won't change, so we have the same number of connections
		// as paths

		for i, c := range conns {
			c.SetPath(&paths[i])
		}
		return false, nil

	}
	// }
	return false, nil
}
