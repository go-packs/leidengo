package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-packs/leidengo/graph"
	"github.com/go-packs/leidengo/leiden"
	"github.com/go-packs/leidengo/quality"
)

type Node struct {
	ID        int     `json:"id"`
	Community int     `json:"community"`
	Weight    float64 `json:"weight"`
}

type Link struct {
	Source int     `json:"source"`
	Target int     `json:"target"`
	Weight float64 `json:"weight"`
}

type GraphData struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

func main() {
	http.HandleFunc("/api/graph", handleGraph)
	http.HandleFunc("/api/custom", handleCustomGraph)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Println("Leiden Visualization Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

type CustomGraphRequest struct {
	Edges string  `json:"edges"` // format: "src tgt weight\nsrc2 tgt2 weight2"
	QF    string  `json:"qf"`
	Res   float64 `json:"res"`
}

func handleCustomGraph(w http.ResponseWriter, r *http.Request) {
	// Add basic CORS for robustness during local dev if needed
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CustomGraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse edges to find max node ID
	type edge struct {
		u, v int
		w    float64
	}
	var edges []edge
	maxID := -1

	lines := strings.Split(req.Edges, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		u, _ := strconv.Atoi(parts[0])
		v, _ := strconv.Atoi(parts[1])
		weight := 1.0
		if len(parts) >= 3 {
			weight, _ = strconv.ParseFloat(parts[2], 64)
		}
		edges = append(edges, edge{u, v, weight})
		if u > maxID {
			maxID = u
		}
		if v > maxID {
			maxID = v
		}
	}

	if maxID == -1 {
		http.Error(w, "No valid edges found", http.StatusBadRequest)
		return
	}

	g := graph.New(maxID + 1)
	for _, e := range edges {
		g.AddEdge(e.u, e.v, e.w)
	}

	var qf quality.QualityFunction
	if strings.ToLower(req.QF) == "cpm" {
		qf = quality.NewCPM(req.Res)
	} else {
		qf = quality.NewModularity(req.Res)
	}

	result := leiden.Run(g, leiden.Options{
		QualityFunc:   qf,
		NumIterations: -1,
		RandomSeed:    42,
	})

	data := GraphData{
		Nodes: make([]Node, g.NodeCount()),
		Links: []Link{},
	}

	for i := 0; i < g.NodeCount(); i++ {
		data.Nodes[i] = Node{
			ID:        i,
			Community: result.FlatCommunities[i],
			Weight:    g.NodeWeight(i),
		}
		for neighbor, weight := range g.Neighbors(i) {
			if i < neighbor {
				data.Links = append(data.Links, Link{
					Source: i,
					Target: neighbor,
					Weight: weight,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func handleGraph(w http.ResponseWriter, r *http.Request) {
	graphType := r.URL.Query().Get("type")
	qfType := r.URL.Query().Get("qf")
	resStr := r.URL.Query().Get("res")

	res, err := strconv.ParseFloat(resStr, 64)
	if err != nil {
		res = 1.0
	}

	var g *graph.Graph
	switch graphType {
	case "cliques":
		g = generateTwoCliques()
	case "ring":
		g = generateRing(20)
	case "clusters":
		g = generateClusters(3, 10)
	default:
		g = generateTwoCliques()
	}

	var qf quality.QualityFunction
	if strings.ToLower(qfType) == "cpm" {
		qf = quality.NewCPM(res)
	} else {
		qf = quality.NewModularity(res)
	}

	result := leiden.Run(g, leiden.Options{
		QualityFunc:   qf,
		NumIterations: -1,
		RandomSeed:    42,
	})

	data := GraphData{
		Nodes: make([]Node, g.NodeCount()),
		Links: []Link{},
	}

	for i := 0; i < g.NodeCount(); i++ {
		data.Nodes[i] = Node{
			ID:        i,
			Community: result.FlatCommunities[i],
			Weight:    g.NodeWeight(i),
		}
		for neighbor, weight := range g.Neighbors(i) {
			if i < neighbor { // avoid double links for visualization
				data.Links = append(data.Links, Link{
					Source: i,
					Target: neighbor,
					Weight: weight,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func generateTwoCliques() *graph.Graph {
	g := graph.New(20)
	// Clique A: 0-9
	for i := 0; i < 10; i++ {
		for j := i + 1; j < 10; j++ {
			g.AddEdge(i, j, 1.0)
		}
	}
	// Clique B: 10-19
	for i := 10; i < 20; i++ {
		for j := i + 1; j < 20; j++ {
			g.AddEdge(i, j, 1.0)
		}
	}
	// Bridge
	g.AddEdge(9, 10, 0.5)
	return g
}

func generateRing(n int) *graph.Graph {
	g := graph.New(n)
	for i := 0; i < n; i++ {
		g.AddEdge(i, (i+1)%n, 1.0)
	}
	return g
}

func generateClusters(k, nPerCluster int) *graph.Graph {
	total := k * nPerCluster
	g := graph.New(total)
	rng := rand.New(rand.NewSource(42))

	for c := 0; c < k; c++ {
		start := c * nPerCluster
		for i := 0; i < nPerCluster; i++ {
			for j := i + 1; j < nPerCluster; j++ {
				if rng.Float64() < 0.7 {
					g.AddEdge(start+i, start+j, 1.0)
				}
			}
		}
	}

	// Inter-cluster edges
	for i := 0; i < total; i++ {
		for j := i + 1; j < total; j++ {
			if g.WeightBetween(i, j) == 0 && rng.Float64() < 0.05 {
				g.AddEdge(i, j, 0.2)
			}
		}
	}
	return g
}
