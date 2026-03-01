package graph

import (
	"testing"
)

func TestGraphBasics(t *testing.T) {
	g := New(4)
	if g.NodeCount() != 4 {
		t.Fatalf("expected 4 nodes, got %d", g.NodeCount())
	}

	if err := g.AddEdge(0, 1, 2.0); err != nil {
		t.Fatal(err)
	}
	if err := g.AddEdge(1, 2, 1.0); err != nil {
		t.Fatal(err)
	}
	if err := g.AddEdge(2, 3, 3.0); err != nil {
		t.Fatal(err)
	}

	if g.TotalWeight() != 6.0 {
		t.Errorf("expected total weight 6.0, got %f", g.TotalWeight())
	}

	if g.Degree(1) != 3.0 { // edges to 0 (w=2) and 2 (w=1)
		t.Errorf("expected degree(1)=3.0, got %f", g.Degree(1))
	}

	if g.WeightBetween(0, 1) != 2.0 {
		t.Errorf("expected w(0,1)=2.0, got %f", g.WeightBetween(0, 1))
	}
	if g.WeightBetween(1, 0) != 2.0 {
		t.Errorf("expected w(1,0)=2.0, got %f", g.WeightBetween(1, 0))
	}
	if g.WeightBetween(0, 2) != 0.0 {
		t.Errorf("expected w(0,2)=0.0, got %f", g.WeightBetween(0, 2))
	}
}

func TestInvalidEdge(t *testing.T) {
	g := New(3)
	if err := g.AddEdge(0, 5, 1.0); err == nil {
		t.Error("expected error for out-of-range node")
	}
	if err := g.AddEdge(0, 1, -1.0); err == nil {
		t.Error("expected error for non-positive weight")
	}
}

func TestPartitionSingleton(t *testing.T) {
	g := New(4)
	_ = g.AddEdge(0, 1, 1.0)
	_ = g.AddEdge(2, 3, 1.0)

	p := NewSingletonPartition(g)
	if p.NumCommunities() != 4 {
		t.Errorf("expected 4 singleton communities, got %d", p.NumCommunities())
	}
	for i := 0; i < 4; i++ {
		if p.CommunityOf(i) != i {
			t.Errorf("node %d should be in community %d, got %d", i, i, p.CommunityOf(i))
		}
	}
}

func TestPartitionMoveNode(t *testing.T) {
	g := New(4)
	_ = g.AddEdge(0, 1, 1.0)
	_ = g.AddEdge(0, 2, 1.0)
	_ = g.AddEdge(1, 2, 1.0)
	_ = g.AddEdge(2, 3, 0.5)

	p := NewSingletonPartition(g)

	// Move node 1 into node 0's community
	p.MoveNode(1, p.CommunityOf(0))
	if p.CommunityOf(1) != p.CommunityOf(0) {
		t.Error("node 1 should be in same community as node 0 after move")
	}
	if p.NumCommunities() != 3 {
		t.Errorf("expected 3 communities after merge, got %d", p.NumCommunities())
	}

	// Internal weight of the merged community should reflect edge (0,1)
	comm := p.CommunityOf(0)
	if p.CommunityInternalWeight(comm) != 2.0 { // edge (0,1) counted twice for undirected
		t.Errorf("expected internal weight 2.0, got %f", p.CommunityInternalWeight(comm))
	}
}

func TestFlattenCommunityIDs(t *testing.T) {
	g := New(4)
	_ = g.AddEdge(0, 1, 1.0)
	_ = g.AddEdge(2, 3, 1.0)

	p := NewSingletonPartition(g)
	p.MoveNode(1, p.CommunityOf(0))
	p.MoveNode(3, p.CommunityOf(2))

	flat := p.FlattenCommunityIDs()
	if flat[0] != flat[1] {
		t.Error("nodes 0 and 1 should be in the same community")
	}
	if flat[2] != flat[3] {
		t.Error("nodes 2 and 3 should be in the same community")
	}
	if flat[0] == flat[2] {
		t.Error("communities {0,1} and {2,3} should be distinct")
	}
}

func TestWeightedEdgesToCommunity(t *testing.T) {
	g := New(4)
	_ = g.AddEdge(0, 1, 2.0)
	_ = g.AddEdge(0, 2, 3.0)
	_ = g.AddEdge(0, 3, 1.0)

	p := NewSingletonPartition(g)
	// Move 1 and 2 into the same community
	p.MoveNode(2, p.CommunityOf(1))

	comm12 := p.CommunityOf(1)
	w := p.WeightedEdgesToCommunity(0, comm12)
	if w != 5.0 { // 2.0 + 3.0
		t.Errorf("expected 5.0 weighted edges from 0 to comm{1,2}, got %f", w)
	}
}
