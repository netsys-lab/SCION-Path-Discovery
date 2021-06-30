package pathselection

import (
	"fmt"
	"sort"

	"github.com/scionproto/scion/go/lib/snet"
)

type byHopCount []snet.Path

func (paths byHopCount) Len() int {
	return len(paths)
}

func (paths byHopCount) Swap(i, j int) {
	paths[i], paths[j] = paths[j], paths[i]
}

func (paths byHopCount) Less(i, j int) bool {
	return len(paths[i].Metadata().Interfaces) < len(paths[j].Metadata().Interfaces)
}

// Select the (count) shortest paths from given path array
func SelectShortestPaths(count int, paths []snet.Path) (selectedPaths []snet.Path) {
	lenPaths := len(paths)
	if lenPaths > 0 {
		sort.Sort(byHopCount(paths))
		if lenPaths < count {
			selectedPaths = paths[0:lenPaths]
		} else {
			selectedPaths = paths[0:count]
		}
	}
	fmt.Println("Selected shortest paths:")
	for i, path := range selectedPaths {
		fmt.Printf("Path %d: %+v\n", i, path)
	}
	return selectedPaths
}

// Select the shortest paths from given path array
func SelectShortestPath(paths []snet.Path) snet.Path {
	return SelectShortestPaths(1, paths)[0]
}
