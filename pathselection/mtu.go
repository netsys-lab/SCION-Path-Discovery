package pathselection

import (
	"github.com/scionproto/scion/go/lib/snet"
	"sort"
)

type byMTU []PathQuality

func (pathSet byMTU) Len() int {
	return len(pathSet)
}

func (pathSet byMTU) Swap(i, j int) {
	pathSet[i].Path, pathSet[j].Path = pathSet[j].Path, pathSet[i].Path
}

func (pathSet byMTU) Less(i, j int) bool {
	// switched so that lager MTUs are at index 0
	return pathSet[i].Path.Metadata().MTU > pathSet[j].Path.Metadata().MTU
}

// GetPathLargeMTU Select the paths from given path array with largest MTU
func (pathSet *PathSet) GetPathLargeMTU() snet.Path {
	sort.Sort(byMTU(pathSet.Paths))
	return SelectPaths(1, pathSet)[0]
}
