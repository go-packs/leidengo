// Package graph provides graph data structures for the Leiden algorithm.
package graph

import "fmt"

// Graph is an undirected, weighted graph.
type Graph struct {
	nodeCount   int
	adjacency   []map[int]float64 // nodeID -> neighborID -> weight
	nodeWeights []float64         // intrinsic node weights (default 1.0)
	totalWeight float64           // sum of all edge weights
}

// New creates an empty graph with n nodes (IDs 0..n-1).
func New(n int) *Graph {
	adj := make([]map[int]float64, n)
	for i := range adj {
		adj[i] = make(map[int]float64)
	}
	nw := make([]float64, n)
	for i := range nw {
		nw[i] = 1.0
	}
	return &Graph{
		nodeCount:   n,
		adjacency:   adj,
		nodeWeights: nw,
	}
}

// NodeCount returns the number of nodes.
func (g *Graph) NodeCount() int { return g.nodeCount }

// AddEdge adds an undirected weighted edge. Adding the same edge twice accumulates weight.
func (g *Graph) AddEdge(src, tgt int, weight float64) error {
	if src < 0 || src >= g.nodeCount || tgt < 0 || tgt >= g.nodeCount {
		return fmt.Errorf("node index out of range: src=%d tgt=%d n=%d", src, tgt, g.nodeCount)
	}
	if weight <= 0 {
		return fmt.Errorf("edge weight must be positive, got %f", weight)
	}
	if src == tgt {
		// self-loop: counts once in total weight
		g.adjacency[src][tgt] += weight
		g.totalWeight += weight
		return nil
	}
	g.adjacency[src][tgt] += weight
	g.adjacency[tgt][src] += weight
	g.totalWeight += weight // each undirected edge counted once
	return nil
}

// SetNodeWeight sets the intrinsic weight of node n.
func (g *Graph) SetNodeWeight(nodeID int, weight float64) {
	g.nodeWeights[nodeID] = weight
}

// NodeWeight returns the intrinsic weight of node n.
func (g *Graph) NodeWeight(nodeID int) float64 { return g.nodeWeights[nodeID] }

// Neighbors returns a map of neighbor -> edge weight for a node.
func (g *Graph) Neighbors(nodeID int) map[int]float64 { return g.adjacency[nodeID] }

// Degree returns the weighted degree (sum of edge weights) of a node.
func (g *Graph) Degree(nodeID int) float64 {
	var d float64
	for _, w := range g.adjacency[nodeID] {
		d += w
	}
	return d
}

// TotalWeight returns the total sum of all edge weights (each undirected edge counted once).
func (g *Graph) TotalWeight() float64 { return g.totalWeight }

// WeightBetween returns the edge weight between two nodes (0 if no edge).
func (g *Graph) WeightBetween(src, tgt int) float64 { return g.adjacency[src][tgt] }

// Nodes returns a slice of all node IDs.
func (g *Graph) Nodes() []int {
	nodes := make([]int, g.nodeCount)
	for i := range nodes {
		nodes[i] = i
	}
	return nodes
}
