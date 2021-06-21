package smp

import (
	"fmt"
	"sort"
	"time"

	"github.com/scionproto/scion/go/lib/snet"
)

type byLatency []snet.Path

func (paths byLatency) Len() int {
	return len(paths)
}

func (paths byLatency) Swap(i, j int) {
	paths[i], paths[j] = paths[j], paths[i]
}

func (paths byLatency) Less(i, j int) bool {
	return sumupLatencies(paths[i].Metadata().Latency) < sumupLatencies(paths[j].Metadata().Latency)
}

func sumupLatencies(latencies []time.Duration) (totalLatency time.Duration) {
	totalLatency = 0
	for _, latency := range latencies {
		totalLatency += latency
	}
	return totalLatency
}

// Select the (count) paths from given path array with the lowest total latencies
func selectLowestLatencies(count int, paths []snet.Path) (selectedPaths []snet.Path) {
	sort.Sort(byLatency(paths))
	selectedPaths = paths[0 : count-1]
	fmt.Println("Selected Paths, ", count, "paths with lowest total latencies: ", selectedPaths)
	return selectedPaths
}

// Select the paths from given path array with lowest total latency
func selectLowestLatency(paths []snet.Path) (selectedPath snet.Path) {
	for _, path := range paths {
		if selectedPath == nil || sumupLatencies(path.Metadata().Latency) < sumupLatencies(selectedPath.Metadata().Latency) {
			selectedPath = path
		}
	}
	fmt.Println("Selected Path with lowest total latency: ", sumupLatencies(selectedPath.Metadata().Latency), " ", selectedPath)
	return selectedPath
}
