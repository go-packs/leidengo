package graph

// Partition holds the assignment of each node to a community.
// Community IDs are arbitrary non-negative integers; use FlattenCommunityIDs
// to normalise them to 0..k-1 after the algorithm converges.
type Partition struct {
	NodeCommunity     []int              // community ID for each node
	communityNodes    map[int][]int      // communityID -> list of node IDs
	nodePosition      []int              // index of nodeID in its communityNodes list
	communityDegree   map[int]float64    // communityID -> sum of member weighted degrees
	communityInternal map[int]float64    // communityID -> sum of internal edge weights (×2 for undirected)
	communityWeight   map[int]float64    // communityID -> sum of member node weights
	nextCommID        int
	g                 *Graph
}

// NewSingletonPartition creates a partition where each node is its own community.
func NewSingletonPartition(g *Graph) *Partition {
	n := g.NodeCount()
	p := &Partition{
		NodeCommunity:     make([]int, n),
		communityNodes:    make(map[int][]int, n),
		nodePosition:      make([]int, n),
		communityDegree:   make(map[int]float64, n),
		communityInternal: make(map[int]float64, n),
		communityWeight:   make(map[int]float64, n),
		nextCommID:        n,
		g:                 g,
	}
	for i := 0; i < n; i++ {
		p.NodeCommunity[i] = i
		p.communityNodes[i] = []int{i}
		p.nodePosition[i] = 0
		p.communityDegree[i] = g.Degree(i)
		p.communityInternal[i] = g.WeightBetween(i, i) // self-loop only
		p.communityWeight[i] = g.NodeWeight(i)
	}
	return p
}

// NumCommunities returns the number of distinct communities.
func (p *Partition) NumCommunities() int { return len(p.communityNodes) }

// CommunityOf returns the community ID of a node.
func (p *Partition) CommunityOf(nodeID int) int { return p.NodeCommunity[nodeID] }

// NodesInCommunity returns all node IDs belonging to a community.
func (p *Partition) NodesInCommunity(commID int) []int { return p.communityNodes[commID] }

// CommunityDegree returns the sum of weighted degrees of all nodes in a community.
func (p *Partition) CommunityDegree(commID int) float64 { return p.communityDegree[commID] }

// CommunityInternalWeight returns the sum of weights of edges internal to a community.
func (p *Partition) CommunityInternalWeight(commID int) float64 {
	return p.communityInternal[commID]
}

// CommunityWeight returns the sum of node weights in a community.
func (p *Partition) CommunityWeight(commID int) float64 {
	return p.communityWeight[commID]
}

// UniqueCommunities returns a slice of all distinct community IDs.
func (p *Partition) UniqueCommunities() []int {
	ids := make([]int, 0, len(p.communityNodes))
	for id := range p.communityNodes {
		ids = append(ids, id)
	}
	return ids
}

// WeightedEdgesToCommunity returns the total weight of edges from nodeID to community commID.
func (p *Partition) WeightedEdgesToCommunity(nodeID, commID int) float64 {
	var total float64
	for neighbor, w := range p.g.Neighbors(nodeID) {
		if p.NodeCommunity[neighbor] == commID {
			total += w
		}
	}
	return total
}

// MoveNode moves nodeID from its current community to destComm.
// Caller is responsible for ensuring destComm exists.
func (p *Partition) MoveNode(nodeID, destComm int) {
	srcComm := p.NodeCommunity[nodeID]
	if srcComm == destComm {
		return
	}

	// Update internal weights: edges from nodeID to srcComm are lost; edges to destComm are gained.
	for neighbor, w := range p.g.Neighbors(nodeID) {
		if neighbor == nodeID {
			continue
		}
		nc := p.NodeCommunity[neighbor]
		if nc == srcComm {
			p.communityInternal[srcComm] -= 2 * w
		} else if nc == destComm {
			p.communityInternal[destComm] += 2 * w
		}
	}
	// self-loop contribution
	selfW := p.g.WeightBetween(nodeID, nodeID)
	p.communityInternal[srcComm] -= selfW
	p.communityInternal[destComm] += selfW

	// Update degree sums
	deg := p.g.Degree(nodeID)
	p.communityDegree[srcComm] -= deg
	p.communityDegree[destComm] += deg

	// Update node weight sums
	nw := p.g.NodeWeight(nodeID)
	p.communityWeight[srcComm] -= nw
	p.communityWeight[destComm] += nw

	// Update membership lists
	p.removeMember(srcComm, nodeID)
	p.nodePosition[nodeID] = len(p.communityNodes[destComm])
	p.communityNodes[destComm] = append(p.communityNodes[destComm], nodeID)
	p.NodeCommunity[nodeID] = destComm
}

// AddSingletonCommunity creates a new community containing only nodeID.
// Returns the new community ID.
func (p *Partition) AddSingletonCommunity(nodeID int) int {
	commID := p.nextCommID
	p.nextCommID++
	p.communityNodes[commID] = nil // initialize
	p.communityDegree[commID] = 0
	p.communityInternal[commID] = 0
	p.communityWeight[commID] = 0
	p.MoveNode(nodeID, commID)
	return commID
}

func (p *Partition) removeMember(commID, nodeID int) {
	members := p.communityNodes[commID]
	pos := p.nodePosition[nodeID]
	lastIdx := len(members) - 1

	if pos != lastIdx {
		lastNode := members[lastIdx]
		members[pos] = lastNode
		p.nodePosition[lastNode] = pos
	}

	p.communityNodes[commID] = members[:lastIdx]

	if len(p.communityNodes[commID]) == 0 {
		delete(p.communityNodes, commID)
		delete(p.communityDegree, commID)
		delete(p.communityInternal, commID)
		delete(p.communityWeight, commID)
	}
}

// FlattenCommunityIDs returns a new []int mapping nodeID -> sequential community index 0..k-1.
func (p *Partition) FlattenCommunityIDs() []int {
	remap := make(map[int]int, len(p.communityNodes))
	idx := 0
	flat := make([]int, len(p.NodeCommunity))
	for nodeID, commID := range p.NodeCommunity {
		if _, ok := remap[commID]; !ok {
			remap[commID] = idx
			idx++
		}
		flat[nodeID] = remap[commID]
	}
	return flat
}

// Copy returns a deep copy of this partition.
func (p *Partition) Copy() *Partition {
	n := len(p.NodeCommunity)
	nc := make([]int, n)
	copy(nc, p.NodeCommunity)

	np := make([]int, n)
	copy(np, p.nodePosition)

	cn := make(map[int][]int, len(p.communityNodes))
	for k, v := range p.communityNodes {
		cp := make([]int, len(v))
		copy(cp, v)
		cn[k] = cp
	}
	cd := make(map[int]float64, len(p.communityDegree))
	for k, v := range p.communityDegree {
		cd[k] = v
	}
	ci := make(map[int]float64, len(p.communityInternal))
	for k, v := range p.communityInternal {
		ci[k] = v
	}
	cw := make(map[int]float64, len(p.communityWeight))
	for k, v := range p.communityWeight {
		cw[k] = v
	}
	return &Partition{
		NodeCommunity:     nc,
		communityNodes:    cn,
		nodePosition:      np,
		communityDegree:   cd,
		communityInternal: ci,
		communityWeight:   cw,
		nextCommID:        p.nextCommID,
		g:                 p.g,
	}
}
