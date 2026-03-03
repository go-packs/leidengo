package leiden

import (
	"math"
	"math/rand"

	"github.com/go-packs/leidengo/graph"
	"github.com/go-packs/leidengo/quality"
	"github.com/go-packs/leidengo/utils"
)

const (
	// theta is the "temperature" for probabilistic merging. Lower = more greedy.
	theta = 0.01
)

type communityCandidate struct {
	commID int
	delta  float64
}

// refinementPhase refines a partition by attempting sub-community merges within
// each community of the coarse partition.
//
// Leiden's key innovation: only well-connected sub-communities are considered for merging,
// guaranteeing that every final community is internally well-connected.
func refinementPhase(
	g *graph.Graph,
	coarsePartition *graph.Partition,
	qf quality.QualityFunction,
	rng *rand.Rand,
) *graph.Partition {
	// Start fresh: every node in its own singleton community.
	refined := graph.NewSingletonPartition(g)

	for _, commID := range coarsePartition.UniqueCommunities() {
		nodesInComm := coarsePartition.NodesInCommunity(commID)
		if len(nodesInComm) <= 1 {
			continue
		}
		refineSubset(g, refined, coarsePartition, nodesInComm, qf, rng)
	}

	return refined
}

// refineSubset merges singleton nodes into neighbouring sub-communities within a subset.
//
// Algorithm (Traag et al. 2019 §3.3):
//  1. Shuffle nodes in the subset.
//  2. For each singleton node, collect eligible neighbouring sub-communities:
//     a sub-community C' is eligible if w(node, C') >= θ * k_i * |C'|.
//  3. Merge probabilistically: P(C') ∝ exp(ΔQ(node→C') / θ).
func refineSubset(
	g *graph.Graph,
	refined *graph.Partition,
	coarse *graph.Partition,
	subset []int,
	qf quality.QualityFunction,
	rng *rand.Rand,
) {
	subsetM := subsetMap(subset)
	shuffled := utils.ShuffleInts(subset, rng)

	for _, nodeID := range shuffled {
		// Only merge nodes that are still singletons in the refined partition.
		currentComm := refined.CommunityOf(nodeID)
		if len(refined.NodesInCommunity(currentComm)) > 1 {
			continue
		}

		ki := g.Degree(nodeID)

		// Aggregate edge weight to each neighbouring refined community (within subset).
		neighborWeights := refined.NeighborCommunityWeights(nodeID)
		kiInSrc := neighborWeights[currentComm]

		// Filter by well-connectedness condition and positive delta Q.
		gamma := qf.Resolution()
		var candidates []communityCandidate
		for nc, wToComm := range neighborWeights {
			if nc == currentComm {
				continue
			}
			
			// Must be within the same coarse community
			inSubset := false
			for neighbor := range g.Neighbors(nodeID) {
				if refined.NodeCommunity[neighbor] == nc && subsetM[neighbor] {
					inSubset = true
					break
				}
			}
			if !inSubset {
				continue
			}

			wc := refined.CommunityWeight(nc)
			if wToComm < gamma*ki*wc {
				continue // would create a poorly connected community
			}
			delta := qf.DeltaQuality(g, refined, nodeID, currentComm, nc, kiInSrc, wToComm)
			if delta >= 0 {
				candidates = append(candidates, communityCandidate{nc, delta})
			}
		}

		if len(candidates) == 0 {
			continue
		}

		chosen := probabilisticChoice(candidates, rng)
		if chosen >= 0 {
			refined.MoveNode(nodeID, chosen)
		}
	}
}

// probabilisticChoice selects a community using softmax over delta Q values.
func probabilisticChoice(candidates []communityCandidate, rng *rand.Rand) int {
	if len(candidates) == 0 {
		return -1
	}
	if len(candidates) == 1 {
		return candidates[0].commID
	}

	weights := make([]float64, len(candidates))
	var total float64
	for i, c := range candidates {
		w := math.Exp(c.delta / theta)
		weights[i] = w
		total += w
	}
	if total == 0 {
		return -1
	}

	r := rng.Float64() * total
	var cumulative float64
	for i, w := range weights {
		cumulative += w
		if r <= cumulative {
			return candidates[i].commID
		}
	}
	return candidates[len(candidates)-1].commID
}
