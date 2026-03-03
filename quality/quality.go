// Package quality provides quality functions for community detection.
package quality

import "github.com/go-packs/leidengo/graph"

// QualityFunction is the interface implemented by all quality metrics.
// Methods operate at the node-move granularity for efficiency.
type QualityFunction interface {
	// DeltaQuality returns the change in quality when nodeID is moved to destComm.
	// kiInDest and kiInSrc are pre-calculated edge weights from nodeID to destComm and srcComm.
	// A positive value means the move improves quality.
	DeltaQuality(g *graph.Graph, p *graph.Partition, nodeID, srcComm, destComm int, kiInSrc, kiInDest float64) float64

	// Quality computes the total quality of the partition.
	Quality(g *graph.Graph, p *graph.Partition) float64

	// Resolution returns the resolution parameter of the quality function.
	Resolution() float64

	// Name returns the name of the quality function.
	Name() string
}
