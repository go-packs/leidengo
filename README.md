# leidengo

A high-performance, pure-Go implementation of the **Leiden community detection algorithm** (Traag, Waltman & van Eck, 2019).

The Leiden algorithm improves on the popular Louvain algorithm by **guaranteeing that all communities in the final partition are internally well-connected** — a property Louvain cannot ensure.

---

## Features

*   **Fast & Efficient**: Optimized $O(1)$ community membership updates and queue-based local moving.
*   **Scalable**: Efficiently handles graphs with tens of thousands of nodes in milliseconds.
*   **Weighted Graphs**: Full support for undirected weighted edges and node weights.
*   **Hierarchical Aggregation**: Correctly preserves self-loops and internal weights across aggregation levels.
*   **Visualization**: Built-in web-based interactive visualizer using D3.js.

---

## Install

```bash
go get github.com/go-packs/leidengo
```

---

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/go-packs/leidengo/graph"
    "github.com/go-packs/leidengo/leiden"
    "github.com/go-packs/leidengo/quality"
)

func main() {
    // Build a graph
    g := graph.New(6)
    g.AddEdge(0, 1, 1.0)
    g.AddEdge(1, 2, 1.0)
    g.AddEdge(2, 0, 1.0) // triangle A
    g.AddEdge(3, 4, 1.0)
    g.AddEdge(4, 5, 1.0)
    g.AddEdge(5, 3, 1.0) // triangle B
    g.AddEdge(2, 3, 0.1) // weak bridge

    // Run with default options (Modularity γ=1.0, seed=42)
    result := leiden.Run(g, leiden.DefaultOptions())

    fmt.Println(result.FlatCommunities) // e.g. [0 0 0 1 1 1]
    fmt.Printf("Quality: %.4f\n", result.Quality)
}
```

---

## Visualization Tool

Explore how different parameters affect community detection in real-time.

```bash
cd viz
go run main.go
```
Open **`http://localhost:8080`** to:
*   Visualize preset graphs (Ring, Two Cliques, Random Clusters).
*   **Dynamic Input**: Paste your own edge list to visualize any network.
*   Interactively adjust **Resolution (γ)** and switch between **Modularity** and **CPM**.

---

## Package Overview

```
leidengo/
├── graph/
│   ├── graph.go       — Undirected weighted graph (adjacency map, degree, node weights)
│   └── partition.go   — Partition with O(1) membership updates and weight tracking
├── leiden/
│   ├── leiden.go      — Run(), RunN(), Options, Result — main entry points
│   ├── local_move.go  — Phase 1: greedy node reassignment with stable O(1) queue
│   ├── refine.go      — Phase 2: probabilistic well-connected sub-partition
│   └── aggregate.go   — Phase 3: hierarchical graph aggregation
├── quality/
│   ├── quality.go     — QualityFunction interface (optimized for single-pass neighbor collection)
│   ├── modularity.go  — Newman-Girvan Modularity
│   └── cpm.go         — Constant Potts Model (weighted for aggregation support)
└── viz/
    └── static/        — D3.js web visualizer
```

---

## API

### `leiden.Run(g, opts) Result`

Runs the full Leiden algorithm. Converges until no further quality improvement is possible.

```go
opts := leiden.Options{
    QualityFunc:   quality.NewModularity(1.0),
    NumIterations: -1,      // -1 = run until convergence
    RandomSeed:    42,
    InitialPartition: nil,
}
result := leiden.Run(g, opts)
```

### Quality Function Interface

Implement the `QualityFunction` interface for custom metrics. It is designed for maximum efficiency by accepting pre-calculated neighbor weights.

```go
type QualityFunction interface {
    DeltaQuality(g *Graph, p *Partition, node, src, dest int, kiInSrc, kiInDest float64) float64
    Quality(g *Graph, p *Partition) float64
    Resolution() float64
    Name() string
}
```

---

## Quality Functions

### Modularity
The standard Newman-Girvan modularity.
*   **Standard**: `quality.NewModularity(1.0)`
*   **Formula**: $Q = \frac{1}{2m} \sum_c [\Sigma_{in} - \gamma \frac{(\Sigma_{tot})^2}{2m}]$

### CPM (Constant Potts Model)
Resolution-limit-free quality function, essential for hierarchical clustering.
*   **Standard**: `quality.NewCPM(0.1)`
*   **Formula**: $H = \sum_c [e_c - \gamma \frac{w_c(w_c-1)}{2}]$ (where $w_c$ is the sum of node weights)

---

## Performance Optimizations

1.  **$O(1)$ Membership Management**: Community member removal and addition are true constant-time operations through position-indexed lists.
2.  **Stable Front-Removal Queue**: Phase 1 uses `container/list` to ensure $O(1)$ front removals and prevent slice reallocation overhead.
3.  **Single-Pass Weight Collection**: Neighboring community weights are collected in a single pass over a node's adjacency map before quality calculations.
4.  **Incremental Tracking**: `Partition` tracks community internal weights, degrees, and total node weights incrementally to avoid full graph scans.

---

## Benchmarks

| Graph | Nodes | Edges | Communities | Modularity | Time |
|-------|-------|-------|-------------|------------|------|
| Ring (8) | 8 | 9 | 2 | ~0.35 | <1ms |
| Two cliques | 10 | 21 | 2 | ~0.49 | <1ms |
| Karate Club | 34 | 78 | 4 | ~0.37 | <1ms |
| Grid 100×100 | 10,000 | 19,800 | ~100 | ~0.93 | ~45ms |

---

## Running Tests

```bash
go test ./...
```

---

## References

- Traag, V.A., Waltman, L. & van Eck, N.J. (2019). [From Louvain to Leiden: guaranteeing well-connected communities](https://www.nature.com/articles/s41598-019-41695-z). *Scientific Reports*, 9, 5233.
- Blondel, V.D. et al. (2008). Fast unfolding of communities in large networks. *Journal of Statistical Mechanics*.
