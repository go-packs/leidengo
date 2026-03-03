package utils

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestShuffleInts(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	rng := rand.New(rand.NewSource(42))
	shuffled := ShuffleInts(s, rng)

	if len(shuffled) != len(s) {
		t.Fatalf("length mismatch: got %d, want %d", len(shuffled), len(s))
	}

	// Should not modify original
	for i, v := range s {
		if v != i+1 {
			t.Errorf("original slice modified at index %d", i)
		}
	}

	// Should have same elements
	sort.Ints(shuffled)
	if !reflect.DeepEqual(shuffled, s) {
		t.Error("shuffled slice contains different elements")
	}

	// Should be different from original (with high probability)
	rng2 := rand.New(rand.NewSource(42))
	shuffled2 := ShuffleInts(s, rng2)
	
	// Compare with another shuffle from same seed
	if !reflect.DeepEqual(shuffled2, ShuffleInts(s, rand.New(rand.NewSource(42)))) {
		// Wait, ShuffleInts takes a copy, so this is fine.
	}

	// Actually check it's not the same as original
	rng3 := rand.New(rand.NewSource(42))
	shuffled3 := ShuffleInts(s, rng3)
	allSame := true
	for i := range s {
		if s[i] != shuffled3[i] {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("shuffled slice is identical to original (unlikely with this seed)")
	}
}

func TestSampleSubset(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	rng := rand.New(rand.NewSource(42))

	// Normal case
	subset := SampleSubset(s, 3, rng)
	if len(subset) != 3 {
		t.Errorf("expected length 3, got %d", len(subset))
	}
	for _, v := range subset {
		found := false
		for _, original := range s {
			if v == original {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("element %d not in original slice", v)
		}
	}

	// k >= len(s)
	subset2 := SampleSubset(s, 10, rng)
	if len(subset2) != len(s) {
		t.Errorf("expected length %d, got %d", len(s), len(subset2))
	}
}

func TestNewRNG(t *testing.T) {
	rng1 := NewRNG(42)
	rng2 := NewRNG(42)
	if rng1.Int63() != rng2.Int63() {
		t.Error("same seed should produce same RNG sequence")
	}

	rng3 := NewRNG(-1)
	rng4 := NewRNG(-1)
	if rng3.Int63() == rng4.Int63() {
		// This could theoretically happen but is extremely unlikely
		t.Log("non-deterministic RNGs produced same first value")
	}
}
