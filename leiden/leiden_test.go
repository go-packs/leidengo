package leiden

import (
	"testing"

	"github.com/go-packs/leidengo/graph"
	"github.com/go-packs/leidengo/quality"
)

// TestTwoCliquesRecovery verifies that Leiden correctly identifies two dense cliques
// connected by a weak bridge edge.
func TestTwoCliquesRecovery(t *testing.T) {
	g := graph.New(10)
	// Clique A: 0-4
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	// Clique B: 5-9
	for i := 5; i < 10; i++ {
		for j := i + 1; j < 10; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	_ = g.AddEdge(4, 5, 0.01) // very weak bridge

	result := Run(g, DefaultOptions())
	flat := result.FlatCommunities

	// Nodes 0-4 should be in the same community
	commA := flat[0]
	for i := 1; i < 5; i++ {
		if flat[i] != commA {
			t.Errorf("node %d should be in community %d (clique A), got %d", i, commA, flat[i])
		}
	}

	// Nodes 5-9 should be in the same community
	commB := flat[5]
	for i := 6; i < 10; i++ {
		if flat[i] != commB {
			t.Errorf("node %d should be in community %d (clique B), got %d", i, commB, flat[i])
		}
	}

	// The two communities should be distinct
	if commA == commB {
		t.Error("clique A and clique B should be in different communities")
	}
}

func TestTwoCliques_SmallBridge(t *testing.T) {
	// 4-node clique
	g := graph.New(4)
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}

	// Resolution 0: should definitely merge all into 1 community
	lowRes := Run(g, Options{QualityFunc: quality.NewModularity(0.0), RandomSeed: 42})
	if lowRes.Partition.NumCommunities() != 1 {
		t.Errorf("expected 1 community at resolution 0, got %d", lowRes.Partition.NumCommunities())
	}

	// Extremely high resolution: should keep them as singletons
	highRes := Run(g, Options{QualityFunc: quality.NewModularity(100.0), RandomSeed: 42})
	if highRes.Partition.NumCommunities() < 4 {
		t.Errorf("expected 4 communities at high resolution, got %d", highRes.Partition.NumCommunities())
	}
}

func TestRunN_BetterQuality(t *testing.T) {
	g := graph.New(10)
	// Create a random-ish graph where different seeds might matter
	for i := 0; i < 10; i++ {
		_ = g.AddEdge(i, (i+1)%10, 1.0)
		_ = g.AddEdge(i, (i+2)%10, 0.5)
	}

	qf := quality.NewModularity(1.0)
	
	// Run once
	r1 := Run(g, Options{QualityFunc: qf, RandomSeed: 1})
	
	// Run N times
	rN := RunN(g, qf, 20, 42)
	
	if rN.Quality < r1.Quality {
		t.Errorf("RunN should find a result at least as good as a single run: %f vs %f", rN.Quality, r1.Quality)
	}
}

// TestSingleNodeGraph handles a trivial edge case.
func TestSingleNodeGraph(t *testing.T) {
	g := graph.New(1)
	result := Run(g, DefaultOptions())
	if len(result.FlatCommunities) != 1 {
		t.Errorf("expected 1 community assignment, got %d", len(result.FlatCommunities))
	}
}

// TestDisconnectedGraph — two completely separate components.
func TestDisconnectedGraph(t *testing.T) {
	g := graph.New(6)
	_ = g.AddEdge(0, 1, 1.0)
	_ = g.AddEdge(1, 2, 1.0)
	_ = g.AddEdge(2, 0, 1.0)
	_ = g.AddEdge(3, 4, 1.0)
	_ = g.AddEdge(4, 5, 1.0)
	_ = g.AddEdge(5, 3, 1.0)

	result := Run(g, DefaultOptions())
	flat := result.FlatCommunities

	commABC := flat[0]
	if flat[1] != commABC || flat[2] != commABC {
		t.Error("nodes 0,1,2 (triangle) should be in the same community")
	}

	commDEF := flat[3]
	if flat[4] != commDEF || flat[5] != commDEF {
		t.Error("nodes 3,4,5 (triangle) should be in the same community")
	}

	if commABC == commDEF {
		t.Error("the two triangles should be in different communities")
	}
}

// TestModularityImprovement — quality should be higher with detected communities
// than with all nodes in a single community.
func TestModularityImprovement(t *testing.T) {
	g := graph.New(10)
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	for i := 5; i < 10; i++ {
		for j := i + 1; j < 10; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	_ = g.AddEdge(4, 5, 0.1)

	result := Run(g, DefaultOptions())

	// Compare against all-in-one partition
	allOne := graph.NewSingletonPartition(g)
	comm0 := allOne.CommunityOf(0)
	for i := 1; i < 10; i++ {
		allOne.MoveNode(i, comm0)
	}

	m := quality.NewModularity(1.0)
	allOneQ := m.Quality(g, allOne)

	if result.Quality <= allOneQ {
		t.Errorf("Leiden quality (%f) should exceed all-in-one quality (%f)", result.Quality, allOneQ)
	}
}

// TestReproducibility — same seed should produce same result.
func TestReproducibility(t *testing.T) {
	g := graph.New(20)
	for i := 0; i < 10; i++ {
		for j := i + 1; j < 10; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	for i := 10; i < 20; i++ {
		for j := i + 1; j < 20; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	_ = g.AddEdge(9, 10, 0.1)

	opts := DefaultOptions()
	opts.RandomSeed = 99

	r1 := Run(g, opts)
	r2 := Run(g, opts)

	for i, c := range r1.FlatCommunities {
		if c != r2.FlatCommunities[i] {
			t.Errorf("non-deterministic result at node %d: %d vs %d", i, c, r2.FlatCommunities[i])
		}
	}
}

// TestRunN returns a valid result.
func TestRunN(t *testing.T) {
	g := graph.New(6)
	_ = g.AddEdge(0, 1, 1.0)
	_ = g.AddEdge(1, 2, 1.0)
	_ = g.AddEdge(2, 0, 1.0)
	_ = g.AddEdge(3, 4, 1.0)
	_ = g.AddEdge(4, 5, 1.0)
	_ = g.AddEdge(5, 3, 1.0)

	result := RunN(g, quality.NewModularity(1.0), 5, 0)
	if result.Quality <= 0 {
		t.Errorf("expected positive quality from RunN, got %f", result.Quality)
	}
}

// TestCPMQualityFunction runs the algorithm with CPM quality.
func TestCPMQualityFunction(t *testing.T) {
	g := graph.New(8)
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	for i := 4; i < 8; i++ {
		for j := i + 1; j < 8; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	_ = g.AddEdge(3, 4, 0.1)

	opts := Options{
		QualityFunc:   quality.NewCPM(0.3),
		NumIterations: -1,
		RandomSeed:    42,
	}
	result := Run(g, opts)
	if len(result.FlatCommunities) != 8 {
		t.Errorf("expected assignment for 8 nodes, got %d", len(result.FlatCommunities))
	}
}

func TestOptions_InitialPartition(t *testing.T) {
	g := graph.New(6)
	for i := 0; i < 3; i++ {
		for j := i + 1; j < 3; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	for i := 3; i < 6; i++ {
		for j := i + 1; j < 6; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	
	// Start with a "perfect" partition
	initP := graph.NewSingletonPartition(g)
	c0 := initP.CommunityOf(0)
	initP.MoveNode(1, c0)
	initP.MoveNode(2, c0)
	c3 := initP.CommunityOf(3)
	initP.MoveNode(4, c3)
	initP.MoveNode(5, c3)

	opts := DefaultOptions()
	opts.InitialPartition = initP
	
	result := Run(g, opts)
	// Should stay at 2 communities
	if result.Partition.NumCommunities() != 2 {
		t.Errorf("expected 2 communities from perfect start, got %d", result.Partition.NumCommunities())
	}
}

func TestOptions_NumIterations(t *testing.T) {
	g := graph.New(10)
	for i := 0; i < 10; i++ {
		for j := i + 1; j < 10; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}

	// 1 iteration should still produce a valid partition
	opts := DefaultOptions()
	opts.NumIterations = 1
	result := Run(g, opts)
	if len(result.FlatCommunities) != 10 {
		t.Errorf("expected 10 nodes, got %d", len(result.FlatCommunities))
	}
}

func TestBuildAggPartition(t *testing.T) {
	g := graph.New(4)
	_ = g.AddEdge(0, 1, 1.0)
	_ = g.AddEdge(2, 3, 1.0)

	// Coarse: {0,1}, {2,3}
	coarse := graph.NewSingletonPartition(g)
	coarse.MoveNode(1, coarse.CommunityOf(0))
	coarse.MoveNode(3, coarse.CommunityOf(2))

	// Refined: {0}, {1}, {2}, {3} (all singletons)
	refined := graph.NewSingletonPartition(g)
	
	// commToNewNode maps refined comms to super-nodes: 0->0, 1->1, 2->2, 3->3
	commToNewNode := map[int]int{0: 0, 1: 1, 2: 2, 3: 3}
	aggRes := aggregateResult{
		aggregated:    graph.New(4),
		nodeToComm:    refined.NodeCommunity,
		commToNewNode: commToNewNode,
	}

	// Build partition on aggregated graph
	aggP := buildAggPartition(aggRes, coarse)
	
	// super-node 0 and 1 represent orig 0 and 1, which were in same coarse community
	if aggP.CommunityOf(0) != aggP.CommunityOf(1) {
		t.Error("super-nodes 0 and 1 should be in the same community")
	}
	// super-nodes 2 and 3 represent orig 2 and 3, which were in same coarse community
	if aggP.CommunityOf(2) != aggP.CommunityOf(3) {
		t.Error("super-nodes 2 and 3 should be in the same community")
	}
}

func TestLiftPartition_Logic(t *testing.T) {
	// Original graph with 4 nodes
	g := graph.New(4)
	_ = g.AddEdge(0, 1, 1.0)
	_ = g.AddEdge(2, 3, 1.0)

	// Refined partition: {0}, {1}, {2,3}
	refined := graph.NewSingletonPartition(g)
	refined.MoveNode(3, refined.CommunityOf(2))

	// Map refined communities to super-nodes: 0->0, 1->1, 2->2 (community of {2,3})
	commToNewNode := map[int]int{
		refined.CommunityOf(0): 0,
		refined.CommunityOf(1): 1,
		refined.CommunityOf(2): 2,
	}

	// Aggregated graph has 3 nodes
	agg := graph.New(3)
	// Partition on aggregated graph: {0,1}, {2}
	partOnAgg := graph.NewSingletonPartition(agg)
	partOnAgg.MoveNode(1, partOnAgg.CommunityOf(0))

	// Lifting should result in:
	// 0 -> super 0 -> comm {0,1} in partOnAgg
	// 1 -> super 1 -> comm {0,1} in partOnAgg
	// 2 -> super 2 -> comm {2} in partOnAgg
	// 3 -> super 2 -> comm {2} in partOnAgg
	// So nodes {0,1} are together, and {2,3} are together.
	lifted := liftPartition(g, refined, partOnAgg, commToNewNode)
	
	if lifted.CommunityOf(0) != lifted.CommunityOf(1) {
		t.Error("nodes 0 and 1 should be in the same community after lifting")
	}
	if lifted.CommunityOf(2) != lifted.CommunityOf(3) {
		t.Error("nodes 2 and 3 should be in the same community after lifting")
	}
	if lifted.CommunityOf(0) == lifted.CommunityOf(2) {
		t.Error("communities {0,1} and {2,3} should be distinct")
	}
}
