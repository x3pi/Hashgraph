package consensus

import (
	"fmt"
	"sort"

	"github.com/x3pi/Hashgraph/event"
	"github.com/x3pi/Hashgraph/node"
)

// CalculateGlobalConsensusTime calculates a global consensus timestamp for events.
// This simplified version uses the median of all known event creation timestamps
// from events that are considered valid across the network.
// It then updates the ConsensusTime for all events on all nodes.
func CalculateGlobalConsensusTime(nodes []*node.Node) {
	var allTimestamps []int64
	uniqueEvents := make(map[string]event.Event) // To collect unique events based on hash

	// Collect all unique, valid events from all nodes
	// We use the creation timestamp (event.Timestamp) for this simplified median calculation.
	for _, n := range nodes {
		for _, e := range n.Events {
			if e.Valid { // Only consider events marked as valid by nodes
				if _, exists := uniqueEvents[e.Hash]; !exists {
					uniqueEvents[e.Hash] = e
					allTimestamps = append(allTimestamps, e.Timestamp)
				}
			}
		}
	}

	if len(allTimestamps) == 0 {
		fmt.Println("🕒 Không có giao dịch hợp lệ nào để tính thời gian đồng thuận toàn cục.")
		return
	}

	sort.Slice(allTimestamps, func(i, j int) bool {
		return allTimestamps[i] < allTimestamps[j]
	})

	// Calculate median timestamp
	var medianTime int64
	if len(allTimestamps)%2 == 1 {
		medianTime = allTimestamps[len(allTimestamps)/2]
	} else {
		// For even number, average of middle two, or just take the lower one for simplicity
		medianTime = allTimestamps[len(allTimestamps)/2-1]
		// Or: (allTimestamps[len(allTimestamps)/2-1] + allTimestamps[len(allTimestamps)/2]) / 2
	}

	fmt.Printf("\n🕒 Thời gian đồng thuận toàn cục được tính là: %d\n", medianTime)

	// Apply this consensus time to all corresponding events on all nodes
	for _, n := range nodes {
		updatedNodeEvents := make(map[string]event.Event)
		changed := false
		for hash, e := range n.Events {
			// Apply consensus time if the event is one of the unique valid events
			// and its consensus time hasn't been set or needs updating.
			if _, isGloballyConsidered := uniqueEvents[e.Hash]; isGloballyConsidered {
				if e.ConsensusTime != medianTime {
					e.ConsensusTime = medianTime
					changed = true
				}
			}
			updatedNodeEvents[hash] = e
		}
		n.Events = updatedNodeEvents
		if changed {
			// fmt.Printf("🕒 Node %d đã cập nhật thời gian đồng thuận cho các giao dịch.\n", n.ID)
		}
	}
}
