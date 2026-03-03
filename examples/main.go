// Command leidengo-example demonstrates the Leiden algorithm on benchmark graphs.
package main

import (
	"fmt"

	"github.com/go-packs/leidengo/graph"
	"github.com/go-packs/leidengo/leiden"
	"github.com/go-packs/leidengo/quality"
)

func main() {
	fmt.Println("=== Leiden Algorithm — leidengo ===")
	fmt.Println()

	fmt.Println("--- Example 1: Simple ring graph ---")
	runRingGraph()

	fmt.Println()
	fmt.Println("--- Example 2: Two cliques connected by a bridge ---")
	runTwoCliques()

	fmt.Println()
	fmt.Println("--- Example 3: Zachary Karate Club (34 nodes) ---")
	runKarateClub()

	fmt.Println()
	fmt.Println("--- Example 4: CPM quality function ---")
	runCPM()
}

// runRingGraph detects communities in a ring of 8 nodes.
func runRingGraph() {
	g := graph.New(8)
	edges := [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}, {4, 5}, {5, 6}, {6, 7}, {7, 4}, {3, 4}}
	for _, e := range edges {
		_ = g.AddEdge(e[0], e[1], 1.0)
	}

	result := leiden.Run(g, leiden.DefaultOptions())
	printResult(result)
}

// runTwoCliques creates two dense cliques with a single bridge edge.
// The algorithm should recover the two cliques as communities.
func runTwoCliques() {
	// Clique A: nodes 0-4, Clique B: nodes 5-9
	g := graph.New(10)
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	for i := 5; i < 10; i++ {
		for j := i + 1; j < 10; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	_ = g.AddEdge(4, 5, 0.1) // weak bridge

	result := leiden.Run(g, leiden.DefaultOptions())
	printResult(result)

	// Verify: nodes 0-4 should be in one community, 5-9 in another.
	flat := result.FlatCommunities
	commA := flat[0]
	for i := 1; i < 5; i++ {
		if flat[i] != commA {
			fmt.Printf("  ⚠ Node %d unexpectedly not in clique A community\n", i)
		}
	}
	commB := flat[5]
	for i := 6; i < 10; i++ {
		if flat[i] != commB {
			fmt.Printf("  ⚠ Node %d unexpectedly not in clique B community\n", i)
		}
	}
	if commA != commB {
		fmt.Println("  ✓ Two cliques correctly separated into distinct communities")
	}
}

// runKarateClub runs Leiden on Zachary's Karate Club graph (1977).
// The classic result has 2 or 4 communities depending on resolution.
func runKarateClub() {
	g := karateClubGraph()

	opts := leiden.DefaultOptions()
	opts.QualityFunc = quality.NewModularity(1.0)

	result := leiden.Run(g, opts)
	printResult(result)

	fmt.Printf("  Classic split: instructor (node 0) vs president (node 33)\n")
	fmt.Printf("  Node  0 community: %d\n", result.FlatCommunities[0])
	fmt.Printf("  Node 33 community: %d\n", result.FlatCommunities[33])
}

// runCPM demonstrates the CPM quality function.
func runCPM() {
	g := graph.New(10)
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	for i := 5; i < 10; i++ {
		for j := i + 1; j < 10; j++ {
			_ = g.AddEdge(i, j, 1.0)
		}
	}
	_ = g.AddEdge(2, 7, 1.0)

	opts := leiden.Options{
		QualityFunc:   quality.NewCPM(0.1),
		NumIterations: -1,
		RandomSeed:    42,
	}
	result := leiden.Run(g, opts)
	fmt.Printf("  Quality function: CPM (γ=0.1)\n")
	printResult(result)
}

func printResult(r leiden.Result) {
	fmt.Printf("  Communities found : %d\n", countUnique(r.FlatCommunities))
	fmt.Printf("  Quality score     : %.6f\n", r.Quality)
	fmt.Printf("  Iterations        : %d\n", r.Iterations)
	fmt.Printf("  Assignment        : %v\n", r.FlatCommunities)
}

func countUnique(s []int) int {
	seen := make(map[int]bool)
	for _, v := range s {
		seen[v] = true
	}
	return len(seen)
}

// karateClubGraph builds Zachary's Karate Club network (34 nodes, 78 edges).
func karateClubGraph() *graph.Graph {
	g := graph.New(34)
	edges := [][2]int{
		{0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {0, 8}, {0, 10},
		{0, 11}, {0, 12}, {0, 13}, {0, 17}, {0, 19}, {0, 21}, {0, 31},
		{1, 2}, {1, 3}, {1, 7}, {1, 13}, {1, 17}, {1, 19}, {1, 21}, {1, 30},
		{2, 3}, {2, 7}, {2, 8}, {2, 9}, {2, 13}, {2, 27}, {2, 28}, {2, 32},
		{3, 7}, {3, 12}, {3, 13},
		{4, 6}, {4, 10},
		{5, 6}, {5, 10}, {5, 16},
		{6, 16},
		{8, 30}, {8, 32}, {8, 33},
		{9, 33},
		{13, 33},
		{14, 32}, {14, 33},
		{15, 32}, {15, 33},
		{18, 32}, {18, 33},
		{19, 33},
		{20, 32}, {20, 33},
		{22, 32}, {22, 33},
		{23, 25}, {23, 27}, {23, 29}, {23, 32}, {23, 33},
		{24, 25}, {24, 27}, {24, 31},
		{25, 31},
		{26, 29}, {26, 33},
		{27, 33},
		{28, 31}, {28, 33},
		{29, 32}, {29, 33},
		{30, 32}, {30, 33},
		{31, 32}, {31, 33},
		{32, 33},
	}
	for _, e := range edges {
		_ = g.AddEdge(e[0], e[1], 1.0)
	}
	return g
}
