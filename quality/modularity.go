package quality

import "github.com/go-packs/leidengo/graph"

// Modularity implements the standard Newman-Girvan modularity quality function:
//
//	Q = (1/2m) * Σ_c [ Σ_in - γ * (Σ_tot)² / (2m) ]
//
// where γ is the resolution parameter (default 1.0).
// Higher γ → more, smaller communities. Lower γ → fewer, larger communities.
type Modularity struct {
	ResolutionParam float64 // γ, default 1.0
}

// NewModularity returns a Modularity with the given resolution (1.0 is standard).
func NewModularity(resolution float64) *Modularity {
	return &Modularity{ResolutionParam: resolution}
}

func (m *Modularity) Name() string { return "Modularity" }

func (m *Modularity) Resolution() float64 { return m.ResolutionParam }

// DeltaQuality computes the modularity gain of moving nodeID into destComm.
//
// Formula: ΔQ = [ k_i_in_dest - k_i_in_src ] / m  -  γ * k_i * [ K_dest - (K_src - k_i) ] / (2m²)
//
// Here we return the value scaled by 2m for efficiency in comparisons.
func (m *Modularity) DeltaQuality(g *graph.Graph, p *graph.Partition, nodeID, srcComm, destComm int, kiInSrc, kiInDest float64) float64 {
	twoM := 2.0 * g.TotalWeight()
	if twoM == 0 {
		return 0
	}
	if srcComm == destComm {
		return 0
	}

	ki := g.Degree(nodeID)
	kcDest := p.CommunityDegree(destComm)
	kcSrc := p.CommunityDegree(srcComm)

	termDest := 2.0*kiInDest - m.ResolutionParam*ki*kcDest/twoM*2.0
	termSrc := -2.0*kiInSrc + m.ResolutionParam*ki*(kcSrc-ki)/twoM*2.0
	
	return termDest + termSrc
}

// Quality computes the full modularity score Q ∈ [-0.5, 1.0].
func (m *Modularity) Quality(g *graph.Graph, p *graph.Partition) float64 {
	twoM := 2.0 * g.TotalWeight()
	if twoM == 0 {
		return 0
	}
	var q float64
	for _, commID := range p.UniqueCommunities() {
		internal := p.CommunityInternalWeight(commID)
		degree := p.CommunityDegree(commID)
		q += internal/twoM - m.ResolutionParam*(degree/twoM)*(degree/twoM)
	}
	return q
}
