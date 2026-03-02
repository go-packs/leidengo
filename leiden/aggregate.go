package leiden

import "github.com/go-packs/leidengo/graph"

// aggregateResult holds the aggregated graph and mappings between levels.
type aggregateResult struct {
	// aggregated is the new collapsed graph; each node represents one refined community.
	aggregated *graph.Graph

	// nodeToComm is the original graph's NodeCommunity slice from the refined partition.
	nodeToComm []int

	// commToNewNode maps refined community ID → new super-node ID in aggregated graph.
	commToNewNode map[int]int
}

// aggregateGraph collapses g according to the refined partition.
// Each community becomes one super-node; edge weights between super-nodes are
// the sum of all cross-community edge weights in the original graph.
func aggregateGraph(g *graph.Graph, refined *graph.Partition) aggregateResult {
	// Assign sequential IDs 0..k-1 to communities.
	commToNewNode := make(map[int]int)
	idx := 0
	for _, commID := range refined.UniqueCommunities() {
		commToNewNode[commID] = idx
		idx++
	}

	agg := graph.New(idx)

	// Set super-node weights to the sum of member node weights.
	for commID, newNode := range commToNewNode {
		var w float64
		for _, nodeID := range refined.NodesInCommunity(commID) {
			w += g.NodeWeight(nodeID)
		}
		agg.SetNodeWeight(newNode, w)
	}

	// Accumulate edge weights between super-nodes.
	// We track (lo, hi) pairs of super-node IDs to avoid double-adding undirected edges.
	type edgeKey struct{ a, b int }
	accumulated := make(map[edgeKey]float64)

	for nodeID := 0; nodeID < g.NodeCount(); nodeID++ {
		srcComm := refined.CommunityOf(nodeID)
		newSrc := commToNewNode[srcComm]

		for neighbor, w := range g.Neighbors(nodeID) {
			tgtComm := refined.CommunityOf(neighbor)
			newTgt := commToNewNode[tgtComm]

			if newSrc == newTgt {
				// Internal (self-loop on the super-node).
				// Each undirected internal edge appears twice (once per endpoint),
				// so we halve here to avoid double-counting.
				accumulated[edgeKey{newSrc, newSrc}] += w / 2.0
			} else {
				lo, hi := newSrc, newTgt
				if lo > hi {
					lo, hi = hi, lo
				}
				// Each undirected cross-edge appears twice; halve.
				accumulated[edgeKey{lo, hi}] += w / 2.0
			}
		}
	}

	for key, w := range accumulated {
		_ = agg.AddEdge(key.a, key.b, w)
	}

	return aggregateResult{
		aggregated:    agg,
		nodeToComm:    refined.NodeCommunity,
		commToNewNode: commToNewNode,
	}
}

// liftPartition maps a partition defined on the aggregated graph back to original node IDs.
//
// Walk: original node
//   → refined community (via refined.CommunityOf)
//   → super-node ID (via commToNewNode)
//   → top-level community (via partOnAgg.CommunityOf)
func liftPartition(
	originalG *graph.Graph,
	refined *graph.Partition,
	partOnAgg *graph.Partition,
	commToNewNode map[int]int,
) *graph.Partition {
	n := originalG.NodeCount()
	nc := make([]int, n)
	for nodeID := 0; nodeID < n; nodeID++ {
		refinedComm := refined.CommunityOf(nodeID)
		superNode := commToNewNode[refinedComm]
		nc[nodeID] = partOnAgg.CommunityOf(superNode)
	}
	return partitionFromNodeCommunity(originalG, nc)
}

// partitionFromNodeCommunity constructs a Partition from a raw NodeCommunity slice.
func partitionFromNodeCommunity(g *graph.Graph, nc []int) *graph.Partition {
	p := graph.NewSingletonPartition(g)
	// firstOccurrence maps nc-community-ID → the representative node's singleton community in p.
	firstOccurrence := make(map[int]int)
	for nodeID, commID := range nc {
		if rep, exists := firstOccurrence[commID]; !exists {
			firstOccurrence[commID] = nodeID // keep this node's singleton community as-is
		} else {
			// Move this node into the representative's community.
			p.MoveNode(nodeID, p.CommunityOf(rep))
		}
	}
	return p
}
