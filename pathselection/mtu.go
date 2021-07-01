package pathselection

import (
	"fmt"
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

// Select the (count) paths from given path array with the largest MTUs
func SelectLargestMTUs(count int, pathSet []PathQuality) (selectedPathSet []PathQuality) {
	lenPaths := len(pathSet)
	var pathsToReturn []PathQuality
	if lenPaths > 0 {
		sort.Sort(byMTU(pathSet))
		if lenPaths < count {
			pathsToReturn = pathSet[0:lenPaths]
		} else {
			pathsToReturn = pathSet[0:count]
		}
	}
	fmt.Println("Selected pathSet with largest MTUs:")
	for i, returnPath := range pathsToReturn {
		fmt.Printf("Path %d: %+v\n", i, returnPath)
		selectedPathSet = append(selectedPathSet, PathQuality{MTU: returnPath.Path.Metadata().MTU, Path: returnPath.Path})
	}
	return selectedPathSet
}

// Select the paths from given path array with largest MTU
func SelectLargestMTU(pathSet []PathQuality) (selectedPath PathQuality) {
	return SelectLargestMTUs(1, pathSet)[0]
}
