package smp

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
func selectShortestPaths(count int, paths []snet.Path) (selectedPaths []snet.Path) {
	sort.Sort(byHopCount(paths))
	selectedPaths = paths[0 : count-1]
	fmt.Println("Selected Paths, ", count, "shortest Paths: ", selectedPaths)
	return selectedPaths
}

// Select the shortest paths from given path array
func selectShortestPath(paths []snet.Path) (selectedPath snet.Path) {
	for _, path := range paths {
		if selectedPath == nil || len(path.Metadata().Interfaces) < len(selectedPath.Metadata().Interfaces) {
			selectedPath = path
		}
	}
	fmt.Println("Selected Path with hop count: ", len(selectedPath.Metadata().Interfaces), " ", selectedPath)
	return selectedPath
}
