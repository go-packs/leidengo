# leidengo

A pure-Go implementation of the **Leiden community detection algorithm** (Traag, Waltman & van Eck, 2019).

The Leiden algorithm improves on the popular Louvain algorithm by **guaranteeing that all communities in the final partition are internally well-connected** — a property Louvain cannot ensure.

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
    fmt.Printf("Communities: %d\n", result.NumCommunities())
}
```

---

## Package Overview

```
leidengo/
├── graph/
│   ├── graph.go       — Undirected weighted graph (adjacency map, degree, total weight)
│   └── partition.go   — Partition with O(1) incremental weight updates on node moves
├── leiden/
│   ├── leiden.go      — Run(), RunN(), Options, Result — main entry points
│   ├── local_move.go  — Phase 1: greedy queue-based node reassignment
│   ├── refine.go      — Phase 2: probabilistic well-connected sub-partition
│   └── aggregate.go   — Phase 3: graph aggregation + partition lifting
├── quality/
│   ├── quality.go     — QualityFunction interface
│   ├── modularity.go  — Newman-Girvan Modularity (with resolution γ)
│   └── cpm.go         — Constant Potts Model (resolution-limit-free)
├── utils/
│   └── random.go      — Seeded RNG, shuffle helpers
└── examples/
    └── main.go        — Ring graph, two cliques, Karate Club, CPM demo
```

---

## API

### `leiden.Run(g, opts) Result`

Runs the full Leiden algorithm. Converges until no improvement or `NumIterations` is reached.

```go
opts := leiden.Options{
    QualityFunc:   quality.NewModularity(1.0), // or quality.NewCPM(0.3)
    NumIterations: -1,      // -1 = run until convergence
    RandomSeed:    42,      // -1 = non-deterministic
    InitialPartition: nil,  // nil = singleton start
}
result := leiden.Run(g, opts)
```

**Result fields:**

| Field | Type | Description |
|-------|------|-------------|
| `FlatCommunities` | `[]int` | `FlatCommunities[nodeID]` = community index 0..k-1 |
| `Quality` | `float64` | Final quality score (modularity or CPM) |
| `Iterations` | `int` | Number of Leiden iterations performed |
| `Partition` | `*graph.Partition` | Full partition object |

### `leiden.RunN(g, qf, n, seed) Result`

Runs the algorithm `n` times with different seeds and returns the best result by quality score.

```go
best := leiden.RunN(g, quality.NewModularity(1.0), 10, 0)
```

---

## Quality Functions

### Modularity

```go
m := quality.NewModularity(1.0) // γ=1.0 is standard
```

The standard Newman-Girvan modularity. `γ > 1` → more, smaller communities; `γ < 1` → fewer, larger.

**Formula:** `Q = (1/2m) Σ_c [Σ_in - γ·(Σ_tot)²/(2m)]`

### CPM (Constant Potts Model)

```go
c := quality.NewCPM(0.1) // γ is the minimum internal density
```

Resolution-limit-free quality function. Every community in the optimal partition has internal edge density > γ.

**Formula:** `H = Σ_c [e_c - γ·n_c·(n_c-1)/2]`

### Custom Quality Function

Implement the `quality.QualityFunction` interface:

```go
type QualityFunction interface {
    DeltaQuality(g *graph.Graph, p *graph.Partition, nodeID, destComm int) float64
    Quality(g *graph.Graph, p *graph.Partition) float64
    Name() string
}
```

---

## Algorithm Phases

The Leiden algorithm iterates three phases until convergence:

**1. Local Moving**
Nodes are processed in random order via a stable queue. Each node is greedily moved to the neighbouring community that maximises `ΔQ`. Neighbours of moved nodes are re-enqueued for re-evaluation.

**2. Refinement** *(Leiden's key innovation)*
Within each community, a fresh singleton partition is created and nodes are probabilistically merged. A node can only merge into a sub-community `C'` if:
```
w(node, C') >= θ · k_i · |C'|   (well-connectedness condition)
```
Merges are sampled with probability `∝ exp(ΔQ / θ)`.

**3. Aggregation**
Each refined community is collapsed into a single super-node. Edge weights between super-nodes are the sum of all connecting edges. The algorithm continues on the aggregated graph.

---

## Benchmarks

| Graph | Nodes | Edges | Communities | Modularity | Time |
|-------|-------|-------|-------------|------------|------|
| Ring (8) | 8 | 9 | 2 | ~0.35 | <1ms |
| Two cliques | 10 | 21 | 2 | ~0.49 | <1ms |
| Karate Club | 34 | 78 | 4 | ~0.37 | <1ms |
| Grid 100×100 | 10,000 | 19,800 | ~100 | ~0.93 | ~50ms |

---

## Running Tests

```bash
go test ./...
```

```bash
go test ./leiden/... -v -run TestTwoCliquesRecovery
```

---

## Running the Example

```bash
cd examples
go run main.go
```

---

## References

- Traag, V.A., Waltman, L. & van Eck, N.J. (2019). [From Louvain to Leiden: guaranteeing well-connected communities](https://www.nature.com/articles/s41598-019-41695-z). *Scientific Reports*, 9, 5233.
- Blondel, V.D. et al. (2008). Fast unfolding of communities in large networks. *Journal of Statistical Mechanics*.
- Zachary, W.W. (1977). An information flow model for conflict and fission in small groups. *Journal of Anthropological Research*.
