package leiden

import (
	"container/list"
	"math/rand"

	"github.com/go-packs/leidengo/graph"
	"github.com/go-packs/leidengo/quality"
	"github.com/go-packs/leidengo/utils"
)

// localMovingPhase performs the greedy local moving step of the Leiden algorithm.
// Nodes are processed in a random order using a stable queue:
// after a node is moved, all its neighbors that are in a different community
// are re-added to the queue (if not already present).
//
// Returns true if any node was moved (partition changed).
func localMovingPhase(
	g *graph.Graph,
	p *graph.Partition,
	qf quality.QualityFunction,
	rng *rand.Rand,
	subset []int, // nil means all nodes
) bool {
	nodes := g.Nodes()
	if subset != nil {
		nodes = subset
	}
	nodes = utils.ShuffleInts(nodes, rng)

	subsetM := subsetMap(subset)

	// inQueue tracks whether a node is currently in the queue to avoid duplicates.
	inQueue := make([]bool, g.NodeCount())
	queue := list.New()
	for _, n := range nodes {
		queue.PushBack(n)
		inQueue[n] = true
	}

	changed := false

	for queue.Len() > 0 {
		element := queue.Front()
		nodeID := element.Value.(int)
		queue.Remove(element)
		inQueue[nodeID] = false

		bestComm, bestDelta := bestCommunityForNode(g, p, nodeID, qf, subsetM)
		currentComm := p.CommunityOf(nodeID)

		if bestComm != currentComm && bestDelta > 1e-12 {
			changed = true
			p.MoveNode(nodeID, bestComm)

			// Re-enqueue neighbors in different communities
			for neighbor := range g.Neighbors(nodeID) {
				if !inQueue[neighbor] && p.CommunityOf(neighbor) != bestComm {
					if subsetM == nil || subsetM[neighbor] {
						queue.PushBack(neighbor)
						inQueue[neighbor] = true
					}
				}
			}
		}
	}

	return changed
}

// bestCommunityForNode returns the community giving the highest positive delta Q
// among all communities adjacent to nodeID. Returns current community and 0.0 if
// no improvement is found.
func bestCommunityForNode(
	g *graph.Graph,
	p *graph.Partition,
	nodeID int,
	qf quality.QualityFunction,
	subsetM map[int]bool,
) (bestComm int, bestDelta float64) {
	currentComm := p.CommunityOf(nodeID)
	bestComm = currentComm
	bestDelta = 0.0

	// Collect distinct neighboring communities (excluding the node's own community)
	// and filter by subset if necessary.
	neighborWeights := p.NeighborCommunityWeights(nodeID)
	kiInSrc := neighborWeights[currentComm]

	for commID, kiInDest := range neighborWeights {
		if commID == currentComm {
			continue
		}
		
		// If subsetM is provided, we only consider communities that contain at least one node from the subset.
		if subsetM != nil {
			inSubset := false
			for neighbor := range g.Neighbors(nodeID) {
				if p.NodeCommunity[neighbor] == commID && subsetM[neighbor] {
					inSubset = true
					break
				}
			}
			if !inSubset {
				continue
			}
		}

		delta := qf.DeltaQuality(g, p, nodeID, currentComm, commID, kiInSrc, kiInDest)
		if delta > bestDelta {
			bestDelta = delta
			bestComm = commID
		}
	}
	return bestComm, bestDelta
}

// subsetMap builds a fast O(1) lookup from a subset slice.
func subsetMap(subset []int) map[int]bool {
	if subset == nil {
		return nil
	}
	m := make(map[int]bool, len(subset))
	for _, n := range subset {
		m[n] = true
	}
	return m
}
