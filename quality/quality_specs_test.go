package quality

import (
	"testing"

	"github.com/go-packs/leidengo/graph"
)

func TestCPM_WeightedNodes(t *testing.T) {
	// In aggregated graphs, a super-node might represent multiple original nodes.
	// Let node 0 represent 2 original nodes, and node 1 represent 3.
	g := graph.New(2)
	g.SetNodeWeight(0, 2.0)
	g.SetNodeWeight(1, 3.0)
	_ = g.AddEdge(0, 1, 10.0)

	p := graph.NewSingletonPartition(g)
	cpm := NewCPM(0.5)

	// Moving node 1 into node 0's community
	comm0 := p.CommunityOf(0)
	comm1 := p.CommunityOf(1)
	
	nw := p.NeighborCommunityWeights(1)
	kiInSrc := nw[comm1]
	kiInDest := nw[comm0]

	delta := cpm.DeltaQuality(g, p, 1, comm1, comm0, kiInSrc, kiInDest)

	// Fixed implementation: (kiInDest - γ*wi*wcDest) - (kiInSrc - γ*wi*(wcSrc - wi))
	// wi = 3.0
	// Dest (comm0): kiInDest = 10.0, wcDest = 2.0, γ = 0.5 => 10.0 - 0.5 * 3.0 * 2.0 = 10.0 - 3.0 = 7.0
	// Src (comm1): kiInSrc = 0.0 (self-loop), wcSrc = 3.0, γ = 0.5 => -(0.0 - 0.5 * 3.0 * (3.0 - 3.0)) = 0.0
	// delta = 7.0 - 0.0 = 7.0
	expected := 7.0
	if !approxEqual(delta, expected, 1e-9) {
		t.Errorf("expected delta %f, got %f", expected, delta)
	}
}

func TestQuality_EmptyGraph(t *testing.T) {
	g := graph.New(0)
	p := graph.NewSingletonPartition(g)

	m := NewModularity(1.0)
	if m.Quality(g, p) != 0 {
		t.Errorf("expected 0 quality for empty graph, got %f", m.Quality(g, p))
	}

	cpm := NewCPM(0.1)
	if cpm.Quality(g, p) != 0 {
		t.Errorf("expected 0 quality for empty graph, got %f", cpm.Quality(g, p))
	}
}

func TestQuality_SelfLoops(t *testing.T) {
	g := graph.New(1)
	_ = g.AddEdge(0, 0, 1.0)
	p := graph.NewSingletonPartition(g)

	m := NewModularity(1.0)
	q := m.Quality(g, p)

	// 2m = 2 * 1.0 = 2.0
	// internal = 1.0 (self-loop weight)
	// degree = 1.0
	// Q = 1.0/2.0 - 1.0 * (1.0/2.0)^2 = 0.5 - 0.25 = 0.25
	if !approxEqual(q, 0.25, 1e-9) {
		t.Errorf("expected modularity 0.25 for single self-loop, got %f", q)
	}

	cpm := NewCPM(0.1)
	h := cpm.Quality(g, p)
	// Internal = 1.0
	// nc = 1
	// H = 1.0 - 0.1 * 1 * (1-1)/2 = 1.0
	if !approxEqual(h, 1.0, 1e-9) {
		t.Errorf("expected CPM quality 1.0 for single self-loop, got %f", h)
	}
}

func TestModularity_HighResolution(t *testing.T) {
	// High resolution should discourage large communities.
	g := graph.New(2)
	_ = g.AddEdge(0, 1, 1.0)
	p := graph.NewSingletonPartition(g)

	m := NewModularity(10.0) // Very high resolution
	
	comm0 := p.CommunityOf(0)
	comm1 := p.CommunityOf(1)
	nw := p.NeighborCommunityWeights(1)
	kiInSrc := nw[comm1]
	kiInDest := nw[comm0]

	delta := m.DeltaQuality(g, p, 1, comm1, comm0, kiInSrc, kiInDest)

	if delta >= 0 {
		t.Errorf("high resolution should result in negative delta for merge, got %f", delta)
	}
}
