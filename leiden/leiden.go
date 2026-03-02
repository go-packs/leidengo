// Package leiden implements the Leiden community detection algorithm.
//
// The Leiden algorithm (Traag, Waltman & van Eck, 2019) improves on the Louvain
// algorithm by guaranteeing that all communities in the final partition are
// internally well-connected. It alternates three phases:
//
//  1. Local Moving   — greedily move nodes to neighbouring communities to maximise quality.
//  2. Refinement     — sub-partition each community, ensuring well-connectedness.
//  3. Aggregation    — collapse each refined community into a super-node and repeat.
//
// Usage:
//
//	g := graph.New(4)
//	g.AddEdge(0, 1, 1.0)
//	g.AddEdge(1, 2, 1.0)
//	g.AddEdge(2, 3, 1.0)
//	g.AddEdge(3, 0, 1.0)
//
//	partition := leiden.Run(g, leiden.DefaultOptions())
//	fmt.Println(partition.FlattenCommunityIDs()) // e.g. [0 0 1 1]
package leiden

import (
	"math/rand"

	"github.com/go-packs/leidengo/graph"
	"github.com/go-packs/leidengo/quality"
	"github.com/go-packs/leidengo/utils"
)

// Options controls the behaviour of the Leiden algorithm.
type Options struct {
	// QualityFunc is the objective function to maximise.
	// Defaults to Modularity with resolution 1.0 if nil.
	QualityFunc quality.QualityFunction

	// NumIterations is the maximum number of Leiden iterations (local move + refine + aggregate).
	// Set to -1 to run until convergence (no further improvement).
	NumIterations int

	// RandomSeed for reproducibility. Set to -1 for non-deterministic behaviour.
	RandomSeed int64

	// InitialPartition is an optional starting partition.
	// If nil, every node starts in its own singleton community.
	InitialPartition *graph.Partition
}

// DefaultOptions returns sensible defaults: Modularity γ=1.0, converge until stable, seed=42.
func DefaultOptions() Options {
	return Options{
		QualityFunc:   quality.NewModularity(1.0),
		NumIterations: -1,
		RandomSeed:    42,
	}
}

// Result holds the output of a Leiden run.
type Result struct {
	// Partition is the final community assignment on the original graph.
	Partition *graph.Partition

	// FlatCommunities is the normalised community array: FlatCommunities[nodeID] = communityIndex.
	FlatCommunities []int

	// Quality is the final quality score of the partition.
	Quality float64

	// Iterations is the number of Leiden iterations performed.
	Iterations int
}

// Run executes the Leiden algorithm on graph g with the given options.
func Run(g *graph.Graph, opts Options) Result {
	if opts.QualityFunc == nil {
		opts.QualityFunc = quality.NewModularity(1.0)
	}

	rng := utils.NewRNG(opts.RandomSeed)

	// Working copies — we aggregate the graph each iteration, so we keep the original.
	currentG := g
	var currentP *graph.Partition
	if opts.InitialPartition != nil {
		currentP = opts.InitialPartition.Copy()
	} else {
		currentP = graph.NewSingletonPartition(currentG)
	}

	// originalPartition tracks the community of each original node across aggregations.
	// After aggregation, originalPartition[originalNode] = community on the *current* graph.
	// We update it each iteration via liftPartition.
	originalPartition := currentP

	iterations := 0

	for {
		if opts.NumIterations >= 0 && iterations >= opts.NumIterations {
			break
		}

		// Phase 1: Local Moving
		changed := localMovingPhase(currentG, currentP, opts.QualityFunc, rng, nil)
		if !changed {
			break // Converged
		}

		// Phase 2: Refinement — produces a finer partition
		refined := refinementPhase(currentG, currentP, opts.QualityFunc, rng)

		// Phase 3: Aggregation — collapse refined communities into super-nodes
		aggResult := aggregateGraph(currentG, refined)

		// Create an initial partition on the aggregated graph based on currentP.
		// Each super-node's community is the coarse community its original members had.
		partOnAgg := buildAggPartition(aggResult, currentP)

		// Track the mapping back to original nodes by lifting through all aggregations.
		originalPartition = liftPartition(g, refined, partOnAgg, aggResult.commToNewNode)

		// Step into the aggregated graph for the next iteration.
		currentG = aggResult.aggregated
		currentP = partOnAgg

		iterations++

		// If the aggregated graph has as many nodes as communities, we're done.
		if currentG.NodeCount() == currentP.NumCommunities() {
			break
		}
	}

	// One final local moving pass on the last level for clean-up.
	localMovingPhase(currentG, currentP, opts.QualityFunc, rng, nil)

	// Compute final quality on the original graph.
	finalQ := opts.QualityFunc.Quality(g, originalPartition)

	return Result{
		Partition:       originalPartition,
		FlatCommunities: originalPartition.FlattenCommunityIDs(),
		Quality:         finalQ,
		Iterations:      iterations,
	}
}

// buildAggPartition builds a Partition on the aggregated graph where each
// super-node's initial community matches the coarse community of its source nodes.
func buildAggPartition(agg aggregateResult, coarseP *graph.Partition) *graph.Partition {
	n := agg.aggregated.NodeCount()
	nc := make([]int, n)

	for origComm, newNode := range agg.commToNewNode {
		// Find any representative original node to get the coarse community.
		// (All nodes in origComm share the same coarse community.)
		_ = origComm
		nc[newNode] = newNode // start as singleton
	}

	// Map each original-graph node → refined comm → super-node → coarse community.
	// Build a mapping: superNodeID → coarse community ID.
	superToCoarse := make(map[int]int, n)
	for nodeID, refinedComm := range agg.nodeToComm {
		superNode := agg.commToNewNode[refinedComm]
		coarseComm := coarseP.CommunityOf(nodeID)
		superToCoarse[superNode] = coarseComm
	}

	for superNode, coarseComm := range superToCoarse {
		nc[superNode] = coarseComm
	}

	return partitionFromNodeCommunity(agg.aggregated, nc)
}

// RunWithSeed is a convenience wrapper for Run with an explicit seed.
func RunWithSeed(g *graph.Graph, qf quality.QualityFunction, seed int64) Result {
	return Run(g, Options{
		QualityFunc:   qf,
		NumIterations: -1,
		RandomSeed:    seed,
	})
}

// RunN runs the algorithm n times with different seeds and returns the best result by quality.
func RunN(g *graph.Graph, qf quality.QualityFunction, n int, baseSeed int64) Result {
	rng := rand.New(rand.NewSource(baseSeed))
	var best Result
	for i := 0; i < n; i++ {
		seed := rng.Int63()
		r := Run(g, Options{
			QualityFunc:   qf,
			NumIterations: -1,
			RandomSeed:    seed,
		})
		if i == 0 || r.Quality > best.Quality {
			best = r
		}
	}
	return best
}
