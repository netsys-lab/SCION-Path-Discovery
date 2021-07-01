package pathselection

import (
	"fmt"
	"sort"
)

type byHopCount []PathQuality

func (pathSet byHopCount) Len() int {
	return len(pathSet)
}

func (pathSet byHopCount) Swap(i, j int) {
	pathSet[i].Path, pathSet[j].Path = pathSet[j].Path, pathSet[i].Path
}

func (pathSet byHopCount) Less(i, j int) bool {
	return len(pathSet[i].Path.Metadata().Interfaces) < len(pathSet[j].Path.Metadata().Interfaces)
}

// Select the (count) shortest paths from given path array
func SelectShortestPaths(count int, pathSet []PathQuality) (selectedPathSet []PathQuality) {
	lenPaths := len(pathSet)
	var pathsToReturn []PathQuality
	if lenPaths > 0 {
		sort.Sort(byHopCount(pathSet))
		if lenPaths < count {
			pathsToReturn = pathSet[0:lenPaths]
		} else {
			pathsToReturn = pathSet[0:count]
		}
	}
	fmt.Println("Selected shortest paths:")
	for i, returnPath := range pathsToReturn {
		fmt.Printf("Path %d: %+v\n", i, returnPath)
		selectedPathSet = append(selectedPathSet, PathQuality{Hopcount: len(returnPath.Path.Metadata().Interfaces), Path: returnPath.Path})
	}
	return selectedPathSet
}

// Select the shortest paths from given path array
func SelectShortestPath(pathSet []PathQuality) (selectedPath PathQuality) {
	return SelectShortestPaths(1, pathSet)[0]
}
