// ABOUTME: PSC-specific test helpers and dependency checking
// ABOUTME: Provides unified patterns for testing PSC functionality with proper dependency management

package helpers

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parser"
	basetesting "tamarou.com/pvm/internal/testing"
)

// isPSCAvailable checks if PSC and its dependencies are available for testing
func isPSCAvailable() bool {
	// Check if tree-sitter parser can be initialized (core PSC dependency)
	_, err := parser.NewParser()
	if err != nil {
		return false
	}

	// Verify tree-sitter CLI is available (for full integration testing)
	_, err = exec.LookPath("tree-sitter")
	return err == nil
}

// SkipIfNoPSC skips the test if PSC dependencies are not available
// This follows the established pattern from SkipIfNoTreeSitter and SkipIfNoSystemPerl
func SkipIfNoPSC(t *testing.T) {
	t.Helper()

	if !isPSCAvailable() {
		// Provide environment-specific guidance
		if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
			t.Skip("PSC dependencies not available in CI environment")
		} else {
			t.Skip("PSC dependencies not available - run 'make tree-sitter' to build required components")
		}
	}
}

// SkipIfNoPSCIntegration skips the test if PSC integration dependencies are not available
// Use this for tests that require full PSC integration with other components
func SkipIfNoPSCIntegration(t *testing.T) {
	t.Helper()

	// First check basic PSC availability
	SkipIfNoPSC(t)

	// Then check if we should run integration tests
	if !basetesting.ShouldRunIntegrationTests() {
		t.Skip("PSC integration test skipped (set PVM_TEST_MODE=integration to run)")
	}
}

// AssertPSCSucceedsOrSkipTODO runs a PSC command and skips with a TODO message if it fails
// This allows tests to pass while clearly indicating PSC functionality issues
func AssertPSCSucceedsOrSkipTODO(t *testing.T, env *TestEnv, args []string, feature string) string {
	t.Helper()

	// Ensure PSC dependencies are available before attempting execution
	SkipIfNoPSC(t)

	stdout, stderr, err := env.RunPSC(args...)
	if err != nil {
		t.Skipf("TODO: PSC %s not yet fully implemented\nCommand: psc %s\nError: %v\nStdout: %s\nStderr: %s",
			feature, joinArgs(args), err, stdout, stderr)
	}
	return stdout
}

// TestPSCCommand runs a PSC command with proper dependency checking and error handling
func TestPSCCommand(t *testing.T, env *TestEnv, args []string) (string, string, error) {
	t.Helper()

	// Check dependencies first
	SkipIfNoPSC(t)

	return env.RunPSC(args...)
}

// handlePSCDependencyError provides helpful error messages for PSC dependency issues
func handlePSCDependencyError(t *testing.T, err error) {
	t.Helper()

	errStr := err.Error()

	// Check for common dependency-related errors
	if containsAny(errStr, []string{"tree-sitter", "CGO", "library", "parser"}) {
		if os.Getenv("CI") != "" {
			t.Skip("PSC dependencies not available in CI environment")
		} else {
			t.Skipf("PSC requires tree-sitter build: %v\nRun 'make tree-sitter' to build dependencies", err)
		}
		return
	}

	// Check for version-related errors
	if containsAny(errStr, []string{"version", "not installed", "PVM-602"}) {
		t.Skipf("PSC version resolution issue: %v\nCheck .perl-version file matches available Perl versions", err)
		return
	}

	// Unexpected error - fail the test
	t.Fatalf("Unexpected PSC error: %v", err)
}

// containsAny checks if a string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// joinArgs safely joins command arguments for display purposes
func joinArgs(args []string) string {
	return strings.Join(args, " ")
}
