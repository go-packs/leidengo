// Package utils provides utility functions for the Leiden algorithm.
package utils

import "math/rand"

// ShuffleInts returns a copy of slice s in random order using rng.
func ShuffleInts(s []int, rng *rand.Rand) []int {
	out := make([]int, len(s))
	copy(out, s)
	rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// NewRNG creates a seeded random source. Use seed=-1 for a non-deterministic run.
func NewRNG(seed int64) *rand.Rand {
	if seed < 0 {
		return rand.New(rand.NewSource(rand.Int63()))
	}
	return rand.New(rand.NewSource(seed))
}

// SampleSubset returns a random subset of size k from slice s (without replacement).
func SampleSubset(s []int, k int, rng *rand.Rand) []int {
	if k >= len(s) {
		return ShuffleInts(s, rng)
	}
	perm := rng.Perm(len(s))
	out := make([]int, k)
	for i := 0; i < k; i++ {
		out[i] = s[perm[i]]
	}
	return out
}
