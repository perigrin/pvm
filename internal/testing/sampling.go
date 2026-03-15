// ABOUTME: Test sampling utilities for performance optimization in short mode
// ABOUTME: Provides consistent random sampling for long-running test suites

package basetesting

import (
	"math/rand"
	"testing"
)

// SampleTest skips a test with given probability in short mode
// sampleRate should be between 0.0 and 1.0 (e.g., 0.1 for 10% sampling)
func SampleTest(t *testing.T, sampleRate float64) {
	if !testing.Short() {
		return // Run all tests in full mode
	}

	// Use deterministic seed based on test name for consistent sampling
	seed := int64(0)
	for _, c := range t.Name() {
		seed = seed*31 + int64(c)
	}

	r := rand.New(rand.NewSource(seed))
	if r.Float64() > sampleRate {
		t.Skipf("Skipping test in short mode (not selected in %.0f%% sample)", sampleRate*100)
	}
}

// SampleE2ETest is a convenience function for e2e tests with 10% sampling
func SampleE2ETest(t *testing.T) {
	SampleTest(t, 0.1)
}

// SampleTypeCheckerTest is a convenience function for typechecker tests with 10% sampling
func SampleTypeCheckerTest(t *testing.T) {
	SampleTest(t, 0.1)
}
