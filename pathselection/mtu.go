package pathselection

import (
	"fmt"
	"sort"

	"github.com/scionproto/scion/go/lib/snet"
)

type byMTU []snet.Path

func (paths byMTU) Len() int {
	return len(paths)
}

func (paths byMTU) Swap(i, j int) {
	paths[i], paths[j] = paths[j], paths[i]
}

func (paths byMTU) Less(i, j int) bool {
	// switched so that lager MTUs are at index 0
	return paths[i].Metadata().MTU > paths[j].Metadata().MTU
}

// Select the (count) paths from given path array with the largest MTUs
func SelectLargestMTUs(count int, paths []snet.Path) (selectedPaths []snet.Path) {
	lenPaths := len(paths)
	if lenPaths > 0 {
		sort.Sort(byMTU(paths))
		if lenPaths < count {
			selectedPaths = paths[0:lenPaths]
		} else {
			selectedPaths = paths[0:count]
		}
	}
	fmt.Println("Selected paths with largest MTUs:")
	for i, path := range selectedPaths {
		fmt.Printf("Path %d: %+v\n", i, path)
	}
	return selectedPaths
}

// Select the paths from given path array with largest MTU
func SelectLargestMTU(paths []snet.Path) snet.Path {
	return SelectShortestPaths(1, paths)[0]
}
