package quality

import (
	"math"
	"testing"

	"github.com/go-packs/leidengo/graph"
)

func TestModularityDeltaPositive(t *testing.T) {
	// Two triangles connected by one edge.
	// Merging node 0 into the triangle {1,2} should improve modularity.
	g := graph.New(6)
	_ = g.AddEdge(0, 1, 1.0)
	_ = g.AddEdge(0, 2, 1.0)
	_ = g.AddEdge(1, 2, 1.0)
	_ = g.AddEdge(3, 4, 1.0)
	_ = g.AddEdge(3, 5, 1.0)
	_ = g.AddEdge(4, 5, 1.0)
	_ = g.AddEdge(2, 3, 0.5)

	p := graph.NewSingletonPartition(g)
	p.MoveNode(1, p.CommunityOf(2)) // put 1 and 2 together

	m := NewModularity(1.0)
	comm12 := p.CommunityOf(2)
	comm0 := p.CommunityOf(0)
	nw := p.NeighborCommunityWeights(0)
	
	delta := m.DeltaQuality(g, p, 0, comm0, comm12, nw[comm0], nw[comm12])
	if delta <= 0 {
		t.Errorf("expected positive delta for moving 0 into {1,2}, got %f", delta)
	}
}

func TestModularityQualityRange(t *testing.T) {
	g := graph.New(4)
	_ = g.AddEdge(0, 1, 1.0)
	_ = g.AddEdge(1, 2, 1.0)
	_ = g.AddEdge(2, 3, 1.0)
	_ = g.AddEdge(3, 0, 1.0)

	p := graph.NewSingletonPartition(g)
	m := NewModularity(1.0)

	q := m.Quality(g, p)
	// Modularity of singleton partition:
	// internal = 0, degree = 2, 2m = 8
	// Q = 4 * (0/8 - (2/8)^2) = 4 * (-1/16) = -0.25
	if q >= 0 {
		t.Errorf("singleton partition should have negative modularity, got %f", q)
	}

	// All in one community
	p2 := graph.NewSingletonPartition(g)
	comm0 := p2.CommunityOf(0)
	p2.MoveNode(1, comm0)
	p2.MoveNode(2, comm0)
	p2.MoveNode(3, comm0)

	q2 := m.Quality(g, p2)
	// Q = 8/8 - (8/8)^2 = 0
	if !approxEqual(q2, 0, 1e-9) {
		t.Errorf("all-in-one community (cycle) should have 0 modularity, got %f", q2)
	}
}

func TestCPMDelta(t *testing.T) {
	g := graph.New(4)
	_ = g.AddEdge(0, 1, 1.0)
	_ = g.AddEdge(0, 2, 1.0)
	_ = g.AddEdge(0, 3, 1.0)
	_ = g.AddEdge(1, 2, 1.0)
	_ = g.AddEdge(1, 3, 1.0)
	_ = g.AddEdge(2, 3, 1.0)

	p := graph.NewSingletonPartition(g)
	p.MoveNode(1, p.CommunityOf(0))
	p.MoveNode(2, p.CommunityOf(0))

	cpm := NewCPM(0.1)
	comm0 := p.CommunityOf(0)
	comm3 := p.CommunityOf(3)
	nw := p.NeighborCommunityWeights(3)

	delta := cpm.DeltaQuality(g, p, 3, comm3, comm0, nw[comm3], nw[comm0])
	if delta <= 0 {
		t.Errorf("expected positive CPM delta for adding node 3 to clique {0,1,2}, got %f", delta)
	}
}

func TestModularityResolution(t *testing.T) {
	// Higher resolution should produce smaller communities (more negative delta for large communities).
	g := graph.New(6)
	for i := 0; i < 6; i++ {
		for j := i + 1; j < 6; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	p := graph.NewSingletonPartition(g)
	for i := 1; i < 6; i++ {
		p.MoveNode(i, p.CommunityOf(0))
	}

	mLow := NewModularity(0.5)
	mHigh := NewModularity(2.0)

	comm0 := p.CommunityOf(0)
	// Add a node to this community — lower resolution should give higher delta
	g2 := graph.New(7)
	for i := 0; i < 6; i++ {
		for j := i + 1; j < 6; j++ {
			_ = g2.AddEdge(i, j, 1.0)
		}
	}
	_ = g2.AddEdge(0, 6, 1.0)

	p2 := graph.NewSingletonPartition(g2)
	for i := 1; i < 6; i++ {
		p2.MoveNode(i, p2.CommunityOf(0))
	}

	comm6 := p2.CommunityOf(6)
	nw := p2.NeighborCommunityWeights(6)

	deltaLow := mLow.DeltaQuality(g2, p2, 6, comm6, comm0, nw[comm6], nw[comm0])
	deltaHigh := mHigh.DeltaQuality(g2, p2, 6, comm6, comm0, nw[comm6], nw[comm0])

	if deltaLow <= deltaHigh {
		t.Errorf("lower resolution should give higher delta: low=%f high=%f", deltaLow, deltaHigh)
	}
}

func approxEqual(a, b, eps float64) bool {
	return math.Abs(a-b) < eps
}
