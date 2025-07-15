// ABOUTME: Test execution mode helpers for conditional test execution
// ABOUTME: Provides environment-based test filtering instead of permanent skips

package basetesting

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestMode represents different test execution modes
type TestMode int

const (
	ModeUnit TestMode = iota
	ModeIntegration
	ModePerformance
	ModeStress
	ModeFull
)

// GetTestMode determines the current test execution mode from environment
func GetTestMode() TestMode {
	mode := strings.ToLower(os.Getenv("PVM_TEST_MODE"))

	switch mode {
	case "unit":
		return ModeUnit
	case "integration":
		return ModeIntegration
	case "performance":
		return ModePerformance
	case "stress":
		return ModeStress
	case "full":
		return ModeFull
	default:
		// Default behavior: run unit and integration tests
		return ModeIntegration
	}
}

// ShouldRunPerformanceTests checks if performance tests should execute
func ShouldRunPerformanceTests() bool {
	mode := GetTestMode()
	return mode == ModePerformance || mode == ModeFull
}

// ShouldRunStressTests checks if stress tests should execute
func ShouldRunStressTests() bool {
	mode := GetTestMode()
	return mode == ModeStress || mode == ModeFull
}

// ShouldRunIntegrationTests checks if integration tests should execute
func ShouldRunIntegrationTests() bool {
	mode := GetTestMode()
	return mode != ModeUnit
}

// ShouldRunLongRunningTests checks if long-running tests should execute
func ShouldRunLongRunningTests() bool {
	mode := GetTestMode()
	return mode == ModePerformance || mode == ModeStress || mode == ModeFull
}

// ShouldMockExternalAPIs checks if external API calls should be mocked
func ShouldMockExternalAPIs() bool {
	// Check for explicit environment variable to enable mocking
	if mock := os.Getenv("PVM_MOCK_EXTERNAL_APIS"); mock != "" {
		return strings.ToLower(mock) == "true" || mock == "1"
	}

	// Mock external APIs by default in CI environments
	if os.Getenv("CI") != "" {
		return true
	}

	// Don't mock in unit-only mode (mocking is for integration tests)
	mode := GetTestMode()
	return mode != ModeUnit
}

// SkipUnlessPerformance skips test unless performance mode is enabled
func SkipUnlessPerformance(t *testing.T, reason string) {
	if !ShouldRunPerformanceTests() {
		t.Skipf("Performance test skipped (run 'make test-performance' or 'make test-full' to run): %s", reason)
	}
}

// SkipUnlessStress skips test unless stress mode is enabled
func SkipUnlessStress(t *testing.T, reason string) {
	if !ShouldRunStressTests() {
		t.Skipf("Stress test skipped (run 'make test-stress' or 'make test-full' to run): %s", reason)
	}
}

// SkipUnlessIntegration skips test unless integration mode is enabled
func SkipUnlessIntegration(t *testing.T, reason string) {
	if !ShouldRunIntegrationTests() {
		t.Skipf("Integration test skipped (set PVM_TEST_MODE=integration to run): %s", reason)
	}
}

// SkipUnlessLongRunning skips test unless long-running tests are enabled
func SkipUnlessLongRunning(t *testing.T, reason string) {
	if !ShouldRunLongRunningTests() {
		t.Skipf("Long-running test skipped (run 'make test-full' to run): %s", reason)
	}
}

// GetSampleRate returns the test sampling rate from environment (0.0-1.0)
func GetSampleRate() float64 {
	rate := os.Getenv("PVM_TEST_SAMPLE_RATE")
	if rate == "" {
		return 1.0 // Run all tests by default
	}

	parsed, err := strconv.ParseFloat(rate, 64)
	if err != nil || parsed < 0 || parsed > 1 {
		return 1.0
	}

	return parsed
}

// Documentation for test modes:
//
// Environment Variables:
//   PVM_TEST_MODE=unit         - Run only unit tests (fastest)
//   PVM_TEST_MODE=integration  - Run unit + integration tests (default)
//   PVM_TEST_MODE=performance  - Run performance tests + unit/integration
//   PVM_TEST_MODE=stress       - Run stress tests + unit/integration
//   PVM_TEST_MODE=full         - Run all tests including performance + stress
//   PVM_TEST_SAMPLE_RATE=0.1   - Run only 10% of tests (for quick sampling)
//   PVM_MOCK_EXTERNAL_APIS=true - Mock external API calls to avoid rate limiting
//
// Usage Examples:
//   make test                           # Default: integration tests (~1.4 min)
//   make test-performance               # Include performance tests
//   make test-stress                    # Include stress tests
//   make test-full                      # Run everything (~3.8 min)
//   PVM_TEST_SAMPLE_RATE=0.1 make test  # Quick 10% sample
//   PVM_MOCK_EXTERNAL_APIS=true make test # Mock external APIs (auto-enabled in CI)
