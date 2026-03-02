package graph

import (
	"reflect"
	"sort"
	"testing"
)

func TestPartition_AddSingletonCommunity(t *testing.T) {
	g := New(3)
	_ = g.AddEdge(0, 0, 1.0) // self-loop
	_ = g.AddEdge(0, 1, 2.0)

	p := NewSingletonPartition(g)
	origComm0 := p.CommunityOf(0)
	
	// Create a new singleton community for node 0
	newCommID := p.AddSingletonCommunity(0)

	if p.CommunityOf(0) != newCommID {
		t.Errorf("expected node 0 in community %d, got %d", newCommID, p.CommunityOf(0))
	}
	if p.NumCommunities() != 3 {
		t.Errorf("expected 3 communities, got %d", p.NumCommunities())
	}
	if _, exists := p.communityNodes[origComm0]; exists {
		t.Errorf("original community %d should have been deleted", origComm0)
	}
	
	if p.CommunityInternalWeight(newCommID) != 1.0 {
		t.Errorf("expected internal weight 1.0 (self-loop), got %f", p.CommunityInternalWeight(newCommID))
	}
	if p.CommunityDegree(newCommID) != 3.0 { // self-loop (1.0) + edge to 1 (2.0)
		t.Errorf("expected community degree 3.0, got %f", p.CommunityDegree(newCommID))
	}
}

func TestPartition_UniqueCommunities(t *testing.T) {
	g := New(5)
	p := NewSingletonPartition(g)
	
	// Merge 0 and 1
	p.MoveNode(1, p.CommunityOf(0))
	// Merge 2 and 3
	p.MoveNode(3, p.CommunityOf(2))
	
	unique := p.UniqueCommunities()
	if len(unique) != 3 {
		t.Fatalf("expected 3 unique communities, got %d", len(unique))
	}
	
	sort.Ints(unique)
	expected := []int{p.CommunityOf(0), p.CommunityOf(2), p.CommunityOf(4)}
	sort.Ints(expected)
	
	if !reflect.DeepEqual(unique, expected) {
		t.Errorf("expected communities %v, got %v", expected, unique)
	}
}

func TestPartition_CommunityWeight(t *testing.T) {
	g := New(3)
	g.SetNodeWeight(0, 10.0)
	g.SetNodeWeight(1, 20.0)
	g.SetNodeWeight(2, 30.0)

	p := NewSingletonPartition(g)
	if p.CommunityWeight(p.CommunityOf(0)) != 10.0 {
		t.Errorf("expected weight 10.0, got %f", p.CommunityWeight(p.CommunityOf(0)))
	}

	// Merge 1 into 0
	p.MoveNode(1, p.CommunityOf(0))
	if p.CommunityWeight(p.CommunityOf(0)) != 30.0 {
		t.Errorf("expected weight 30.0 after merge, got %f", p.CommunityWeight(p.CommunityOf(0)))
	}
}

func TestPartition_Copy(t *testing.T) {
	g := New(3)
	_ = g.AddEdge(0, 1, 1.0)
	p := NewSingletonPartition(g)
	p.MoveNode(1, p.CommunityOf(0))
	
	p2 := p.Copy()
	
	if p2.NumCommunities() != p.NumCommunities() {
		t.Errorf("copy NumCommunities mismatch: %d vs %d", p2.NumCommunities(), p.NumCommunities())
	}
	
	if !reflect.DeepEqual(p2.NodeCommunity, p.NodeCommunity) {
		t.Error("copy NodeCommunity mismatch")
	}

	if !reflect.DeepEqual(p2.nodePosition, p.nodePosition) {
		t.Error("copy nodePosition mismatch")
	}

	if !reflect.DeepEqual(p2.communityWeight, p.communityWeight) {
		t.Error("copy communityWeight mismatch")
	}
	
	// Modify original, copy should remain unchanged
	p.MoveNode(1, p.AddSingletonCommunity(1))
	
	if p2.CommunityOf(1) == p.CommunityOf(1) {
		t.Error("copy should not be affected by changes to original")
	}
	if p2.NumCommunities() != 2 {
		t.Errorf("expected copy to still have 2 communities, got %d", p2.NumCommunities())
	}
}

func TestPartition_MoveNode_EdgeCases(t *testing.T) {
	g := New(2)
	_ = g.AddEdge(0, 1, 1.0)
	_ = g.AddEdge(0, 0, 0.5) // self-loop
	
	p := NewSingletonPartition(g)
	comm0 := p.CommunityOf(0)
	
	// Move to same community
	p.MoveNode(0, comm0)
	if p.CommunityOf(0) != comm0 {
		t.Error("MoveNode to same community failed")
	}
	if p.CommunityInternalWeight(comm0) != 0.5 {
		t.Errorf("expected internal weight 0.5, got %f", p.CommunityInternalWeight(comm0))
	}
	
	// Move node with self-loop
	comm1 := p.CommunityOf(1)
	p.MoveNode(0, comm1)
	
	if p.CommunityOf(0) != comm1 {
		t.Error("failed to move node 0 to community 1")
	}
	// Internal weight of comm1 should be:
	// 0.0 (initial) + 0.5 (node 0 self-loop) + 2 * 1.0 (edge 0-1) = 2.5
	if p.CommunityInternalWeight(comm1) != 2.5 {
		t.Errorf("expected internal weight 2.5, got %f", p.CommunityInternalWeight(comm1))
	}
	
	if _, exists := p.communityNodes[comm0]; exists {
		t.Error("community 0 should be empty and deleted")
	}
}

func TestPartition_NodesInCommunity(t *testing.T) {
	g := New(3)
	p := NewSingletonPartition(g)
	p.MoveNode(1, p.CommunityOf(0))
	
	nodes := p.NodesInCommunity(p.CommunityOf(0))
	sort.Ints(nodes)
	expected := []int{0, 1}
	if !reflect.DeepEqual(nodes, expected) {
		t.Errorf("expected nodes %v, got %v", expected, nodes)
	}
}
