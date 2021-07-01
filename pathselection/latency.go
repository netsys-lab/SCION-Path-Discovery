package pathselection

import (
	"fmt"
	"sort"
	"time"
)

type byLatency []PathQuality

func (pathSet byLatency) Len() int {
	return len(pathSet)
}

func (pathSet byLatency) Swap(i, j int) {
	pathSet[i].Path, pathSet[j].Path = pathSet[j].Path, pathSet[i].Path
}

func (pathSet byLatency) Less(i, j int) bool {
	return sumupLatencies(pathSet[i].Path.Metadata().Latency) < sumupLatencies(pathSet[j].Path.Metadata().Latency)
}

func sumupLatencies(latencies []time.Duration) (totalLatency time.Duration) {
	totalLatency = 0
	for _, latency := range latencies {
		if latency > 0 {
			totalLatency += latency
		}
	}
	return totalLatency
}

// Select the (count) paths from given path array with the lowest total latencies
func SelectLowestLatencies(count int, pathSet []PathQuality) (selectedPathSet []PathQuality) {
	lenPaths := len(pathSet)
	var pathsToReturn []PathQuality
	if lenPaths > 0 {
		sort.Sort(byLatency(pathSet))
		if lenPaths < count {
			pathsToReturn = pathSet[0:lenPaths]
		} else {
			pathsToReturn = pathSet[0:count]
		}
	}
	fmt.Println("Selected pathSet with lowest total latencies:")
	for i, returnPath := range pathsToReturn {
		fmt.Printf("Path %d: %+v\n", i, returnPath)
		selectedPathSet = append(selectedPathSet, PathQuality{Latency: sumupLatencies(returnPath.Path.Metadata().Latency), Path: returnPath.Path})
	}
	return selectedPathSet
}

// Select the paths from given path array with lowest total latency
func SelectLowestLatency(pathSet []PathQuality) (selectedPath PathQuality) {
	return SelectLowestLatencies(1, pathSet)[0]
}
