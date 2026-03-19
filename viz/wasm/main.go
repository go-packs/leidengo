//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/go-packs/leidengo/graph"
	"github.com/go-packs/leidengo/leiden"
	"github.com/go-packs/leidengo/quality"
)

func runLeiden(this js.Value, args []js.Value) any {
	if len(args) < 3 {
		return "Error: missing arguments"
	}

	edgeList := args[0].String()
	qfType := args[1].String()
	res := args[2].Float()

	// Parse edges
	lines := strings.Split(edgeList, "\n")
	type edge struct {
		u, v int
		w    float64
	}
	var edges []edge
	maxID := -1

	for _, line := range lines {
		parts := strings.Fields(strings.TrimSpace(line))
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
		return "Error: no valid edges"
	}

	g := graph.New(maxID + 1)
	for _, e := range edges {
		g.AddEdge(e.u, e.v, e.w)
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

	// Format for D3
	type Node struct {
		ID        int     `json:"id" js:"id"`
		Community int     `json:"community" js:"community"`
		Weight    float64 `json:"weight" js:"weight"`
	}
	type Link struct {
		Source int     `json:"source" js:"source"`
		Target int     `json:"target" js:"target"`
		Weight float64 `json:"weight" js:"weight"`
	}

	nodes := make([]Node, g.NodeCount())
	links := []Link{}

	for i := 0; i < g.NodeCount(); i++ {
		nodes[i] = Node{ID: i, Community: result.FlatCommunities[i], Weight: g.NodeWeight(i)}
		for neighbor, weight := range g.Neighbors(i) {
			if i < neighbor {
				links = append(links, Link{Source: i, Target: neighbor, Weight: weight})
			}
		}
	}

	output, _ := json.Marshal(map[string]any{
		"nodes": nodes,
		"links": links,
	})
	return string(output)
}

func main() {
	js.Global().Set("runLeidenWasm", js.FuncOf(runLeiden))
	fmt.Println("Leiden Wasm Loaded!")
	select {} // Keep Go running
}
