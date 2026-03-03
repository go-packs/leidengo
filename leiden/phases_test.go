package leiden

import (
	"math/rand"
	"testing"

	"github.com/go-packs/leidengo/graph"
	"github.com/go-packs/leidengo/quality"
)

func TestLocalMovingPhase(t *testing.T) {
	g := graph.New(4)
	_ = g.AddEdge(0, 1, 1.0)
	_ = g.AddEdge(2, 3, 1.0)
	// Initially all singletons. 0-1 and 2-3 should merge.
	p := graph.NewSingletonPartition(g)
	qf := quality.NewModularity(1.0)
	rng := rand.New(rand.NewSource(42))

	changed := localMovingPhase(g, p, qf, rng, nil)
	if !changed {
		t.Error("expected partition to change")
	}

	if p.CommunityOf(0) != p.CommunityOf(1) {
		t.Error("nodes 0 and 1 should be in the same community")
	}
	if p.CommunityOf(2) != p.CommunityOf(3) {
		t.Error("nodes 2 and 3 should be in the same community")
	}
	if p.CommunityOf(0) == p.CommunityOf(2) {
		t.Error("disconnected components should not merge")
	}
}

func TestRefinementPhase(t *testing.T) {
	g := graph.New(4)
	_ = g.AddEdge(0, 1, 1.0)
	_ = g.AddEdge(1, 2, 1.0)
	_ = g.AddEdge(2, 3, 1.0)
	
	// Create a coarse partition where all nodes are in one community
	p := graph.NewSingletonPartition(g)
	comm := p.CommunityOf(0)
	for i := 1; i < 4; i++ {
		p.MoveNode(i, comm)
	}

	qf := quality.NewModularity(1.0)
	rng := rand.New(rand.NewSource(42))

	refined := refinementPhase(g, p, qf, rng)
	
	// Refinement should produce some merges since they are well-connected
	if refined.NumCommunities() >= 4 {
		t.Errorf("refinement should have merged some nodes, got %d communities", refined.NumCommunities())
	}
}

func TestAggregateGraph(t *testing.T) {
	g := graph.New(4)
	_ = g.AddEdge(0, 1, 1.0) // internal to comm A
	_ = g.AddEdge(2, 3, 1.0) // internal to comm B
	_ = g.AddEdge(1, 2, 0.5) // between A and B
	
	p := graph.NewSingletonPartition(g)
	p.MoveNode(1, p.CommunityOf(0))
	p.MoveNode(3, p.CommunityOf(2))
	
	res := aggregateGraph(g, p)
	if res.aggregated.NodeCount() != 2 {
		t.Fatalf("expected 2 super-nodes, got %d", res.aggregated.NodeCount())
	}
	
	// Check super-node weights
	// Each node has weight 1.0 initially.
	for i := 0; i < 2; i++ {
		if res.aggregated.NodeWeight(i) != 2.0 {
			t.Errorf("expected super-node weight 2.0, got %f", res.aggregated.NodeWeight(i))
		}
	}
	
	// Check edge weight between super-nodes
	// Only edge (1,2) with weight 0.5 connects them.
	// In aggregated graph, this should be 0.5.
	w := res.aggregated.WeightBetween(0, 1)
	if w != 0.5 {
		t.Errorf("expected edge weight 0.5 between super-nodes, got %f", w)
	}
	
	// Check self-loops on super-nodes
	// Super-node 0 represents {0,1} which has internal edge (0,1) weight 1.0.
	// Super-node 1 represents {2,3} which has internal edge (2,3) weight 1.0.
	if res.aggregated.WeightBetween(0, 0) != 1.0 {
		t.Errorf("expected self-loop 1.0 on super-node 0, got %f", res.aggregated.WeightBetween(0, 0))
	}
}

func TestAggregateGraph_SelfLoops(t *testing.T) {
	// 1. Original self-loop should be preserved
	g1 := graph.New(1)
	_ = g1.AddEdge(0, 0, 10.0)
	p1 := graph.NewSingletonPartition(g1)
	res1 := aggregateGraph(g1, p1)
	if res1.aggregated.WeightBetween(0, 0) != 10.0 {
		t.Errorf("expected 10.0 self-loop weight, got %f", res1.aggregated.WeightBetween(0, 0))
	}

	// 2. Internal edge between distinct nodes should become a self-loop
	g2 := graph.New(2)
	_ = g2.AddEdge(0, 1, 5.0)
	p2 := graph.NewSingletonPartition(g2)
	p2.MoveNode(1, p2.CommunityOf(0))
	res2 := aggregateGraph(g2, p2)
	if res2.aggregated.WeightBetween(0, 0) != 5.0 {
		t.Errorf("expected 5.0 internal-to-selfloop weight, got %f", res2.aggregated.WeightBetween(0, 0))
	}

	// 3. Combined self-loop and internal edges
	g3 := graph.New(2)
	_ = g3.AddEdge(0, 0, 1.0)
	_ = g3.AddEdge(1, 1, 2.0)
	_ = g3.AddEdge(0, 1, 3.0)
	p3 := graph.NewSingletonPartition(g3)
	p3.MoveNode(1, p3.CommunityOf(0))
	res3 := aggregateGraph(g3, p3)
	// Expected super-node self-loop: 1.0 (self-0) + 2.0 (self-1) + 3.0 (edge 0-1) = 6.0
	if res3.aggregated.WeightBetween(0, 0) != 6.0 {
		t.Errorf("expected 6.0 combined self-loop weight, got %f", res3.aggregated.WeightBetween(0, 0))
	}
}
