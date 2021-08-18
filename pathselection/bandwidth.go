package pathselection

import (
	"github.com/scionproto/scion/go/lib/snet"
)

type byBandwidth []PathQuality

func (pathSet byBandwidth) Len() int {
	return len(pathSet)
}

func (pathSet byBandwidth) Swap(i, j int) {
	pathSet[i].Path, pathSet[j].Path = pathSet[j].Path, pathSet[i].Path
}

func (pathSet byBandwidth) Less(i, j int) bool {
	// todo
	return false
}

// GetPathHighBandwidth Select the shortest paths from given path array
func (pathSet *PathSet) GetPathHighBandwidth() snet.Path {
	return SelectPaths(1, pathSet.Paths)[0]
}
