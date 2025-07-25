// ABOUTME: Perl detection helpers for end-to-end tests
// ABOUTME: Provides portable Perl detection for test environments

package helpers

import (
	"os/exec"
	"runtime"
	"testing"

	"tamarou.com/pvm/internal/perl"
)

// FindSystemPerl finds the system Perl executable path
// Returns empty string if not found
func FindSystemPerl() string {
	// Use the internal perl package's detection
	systemPerl, err := perl.DetectSystemPerl()
	if err != nil {
		return ""
	}
	return systemPerl.Path
}

// HasSystemPerl checks if a system Perl is available
func HasSystemPerl() bool {
	return FindSystemPerl() != ""
}

// Note: SkipIfNoSystemPerl and EnsureSystemPerl have been removed.

// DefaultTestPerlVersion is the Perl version used for testing (read from .perl-version)
const DefaultTestPerlVersion = "5.40.2"

// TestPerlEnvironment is a dummy struct for backward compatibility
type TestPerlEnvironment struct {
	Version string
}

// SetupTestPerlEnvironment is now a no-op since PVM shell integration handles everything
// Kept for backward compatibility during cleanup transition
func SetupTestPerlEnvironment(t *testing.T, version string) *TestPerlEnvironment {
	t.Helper()
	// No-op: PVM shell integration automatically handles Perl environment setup
	t.Logf("Using PVM shell integration for Perl %s environment", version)
	return &TestPerlEnvironment{Version: version}
}

// EnsureBinaryPerl is now a no-op since PVM shell integration handles everything
// Returns a placeholder path for backward compatibility during cleanup transition
func EnsureBinaryPerl(t *testing.T, version string) string {
	t.Helper()
	// No-op: PVM shell integration automatically handles Perl availability
	t.Logf("Using PVM shell integration for Perl %s (no manual binary management needed)", version)
	return "/usr/bin/perl" // Placeholder - PVM will use correct version automatically
}

// HasTreeSitter checks if the tree-sitter library is available
func HasTreeSitter() bool {
	// Tree-sitter is built into the binary now, so we just need to check
	// if the tree-sitter-typed-perl directory exists and is properly built

	// For now, we'll use a simple check - try to find the tree-sitter CLI
	// In production, PSC would use the embedded tree-sitter library
	_, err := exec.LookPath("tree-sitter")
	return err == nil
}

// SkipIfNoTreeSitter skips the test if tree-sitter is not available
func SkipIfNoTreeSitter(t *testing.T) {
	if !HasTreeSitter() {
		t.Skip("Tree-sitter not available, skipping test")
	}
}

// GetPerlCommand returns the appropriate perl command for the platform
func GetPerlCommand() string {
	if runtime.GOOS == "windows" {
		return "perl.exe"
	}
	return "perl"
}
