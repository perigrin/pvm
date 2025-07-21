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
// Use EnsureBinaryPerl() from perl_binary.go for reliable test infrastructure.

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
