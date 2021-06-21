package smp

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
	return paths[i].Metadata().MTU < paths[j].Metadata().MTU
}

// Select the (count) paths from given path array with the largest MTUs
func selectLargestMTUs(count int, paths []snet.Path) (selectedPaths []snet.Path) {
	sort.Sort(byMTU(paths))
	selectedPaths = paths[len(paths)-count : len(paths)-1]
	fmt.Println("Selected Paths, ", count, "paths with largest MTUs: ", selectedPaths)
	return selectedPaths
}

// Select the paths from given path array with largest MTU
func selectLargestMTU(paths []snet.Path) (selectedPath snet.Path) {
	for _, path := range paths {
		if selectedPath == nil || path.Metadata().MTU > selectedPath.Metadata().MTU {
			selectedPath = path
		}
	}
	fmt.Println("Selected Path with MTU: ", selectedPath.Metadata().MTU, " ", selectedPath)
	return selectedPath
}
