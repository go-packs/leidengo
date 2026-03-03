package quality

import "github.com/go-packs/leidengo/graph"

// CPM implements the Constant Potts Model quality function:
//
//	H = Σ_c [ e_c - γ * w_c*(w_c-1)/2 ]
//
// where e_c is the number of internal edges in community c,
// w_c is the sum of node weights, and γ is the resolution parameter.
//
// CPM has the resolution-limit-free property: the optimal resolution is
// independent of the size of the graph.
type CPM struct {
	ResolutionParam float64 // γ — sets the minimum density of communities
}

// NewCPM returns a CPM quality function with the given resolution.
// Typical values: 0.01 (coarse) to 0.5 (fine).
func NewCPM(resolution float64) *CPM {
	return &CPM{ResolutionParam: resolution}
}

func (c *CPM) Name() string { return "CPM" }

func (c *CPM) Resolution() float64 { return c.ResolutionParam }

// DeltaQuality returns the CPM gain of moving nodeID into destComm.
//
// ΔH = ΔH(remove from src) + ΔH(add to dest)
// ΔH(add to C) = kiIn - γ * wi * wc
// ΔH(remove from C) = -kiIn + γ * wi * (wc - wi)
func (c *CPM) DeltaQuality(g *graph.Graph, p *graph.Partition, nodeID, srcComm, destComm int, kiInSrc, kiInDest float64) float64 {
	if srcComm == destComm {
		return 0
	}

	wi := g.NodeWeight(nodeID)
	wcDest := p.CommunityWeight(destComm)
	wcSrc := p.CommunityWeight(srcComm)

	termDest := kiInDest - c.ResolutionParam*wi*wcDest
	termSrc := -kiInSrc + c.ResolutionParam*wi*(wcSrc-wi)

	return termDest + termSrc
}

// Quality computes the total CPM quality of the partition.
func (c *CPM) Quality(g *graph.Graph, p *graph.Partition) float64 {
	var h float64
	for _, commID := range p.UniqueCommunities() {
		internal := p.CommunityInternalWeight(commID)
		wc := p.CommunityWeight(commID)
		h += internal - c.ResolutionParam*wc*(wc-1)/2.0
	}
	return h
}
