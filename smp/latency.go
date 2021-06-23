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
		if latency > 0 {
			totalLatency += latency
		}
	}
	return totalLatency
}

// Select the (count) paths from given path array with the lowest total latencies
func SelectLowestLatencies(count int, paths []snet.Path) (selectedPaths []snet.Path) {
	lenPaths := len(paths)
	if lenPaths > 0 {
		sort.Sort(byLatency(paths))
		if lenPaths < count {
			selectedPaths = paths[0:lenPaths]
		} else {
			selectedPaths = paths[0:count]
		}
	}
	fmt.Println("Selected paths with lowest total latencies:")
	for i, path := range selectedPaths {
		fmt.Printf("Path %d: %+v\n", i, path)
	}
	return selectedPaths
}

// Select the paths from given path array with lowest total latency
func SelectLowestLatency(paths []snet.Path) snet.Path {
	return SelectLowestLatencies(1, paths)[0]
}
