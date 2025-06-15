// ABOUTME: Perl detection helpers for end-to-end tests
// ABOUTME: Provides portable Perl detection for test environments

package helpers

import (
	"os"
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

// SkipIfNoSystemPerl skips the test if no system Perl is available
// Now attempts to install system Perl if not found (unless in CI environment)
func SkipIfNoSystemPerl(t *testing.T) {
	if HasSystemPerl() {
		return // System Perl already available
	}
	
	// Skip installation in CI/automated environments to avoid privilege issues
	if os.Getenv("CI") != "" || os.Getenv("AUTOMATED_TESTING") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("System Perl not available and running in CI environment")
	}
	
	// Skip installation if running tests with -short flag
	if testing.Short() {
		t.Skip("System Perl not available and running in short mode")
	}
	
	// Try to install system Perl using SystemPerlManager
	manager := perl.NewSystemPerlManager()
	systemPerl, err := manager.DetectOrInstallPerl()
	if err != nil {
		t.Skipf("System Perl not available and installation failed: %v", err)
	}
	
	// Validate the installation
	err = manager.ValidateInstallation(systemPerl)
	if err != nil {
		t.Skipf("System Perl installation validation failed: %v", err)
	}
	
	t.Logf("Successfully installed system Perl: %s at %s", systemPerl.Version, systemPerl.Path)
}

// EnsureSystemPerl ensures system Perl is available, installing it if necessary
// Returns the SystemPerl instance or fails the test
func EnsureSystemPerl(t *testing.T) *perl.SystemPerl {
	// Try to detect existing system Perl first
	if systemPerl, err := perl.DetectSystemPerl(); err == nil {
		return systemPerl
	}
	
	// Skip installation in CI/automated environments to avoid privilege issues
	if os.Getenv("CI") != "" || os.Getenv("AUTOMATED_TESTING") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Fatal("System Perl not available and running in CI environment")
	}
	
	// Skip installation if running tests with -short flag
	if testing.Short() {
		t.Fatal("System Perl not available and running in short mode")
	}
	
	// Try to install system Perl using SystemPerlManager
	manager := perl.NewSystemPerlManager()
	systemPerl, err := manager.DetectOrInstallPerl()
	if err != nil {
		t.Fatalf("System Perl not available and installation failed: %v", err)
	}
	
	// Validate the installation
	err = manager.ValidateInstallation(systemPerl)
	if err != nil {
		t.Fatalf("System Perl installation validation failed: %v", err)
	}
	
	t.Logf("Successfully ensured system Perl: %s at %s", systemPerl.Version, systemPerl.Path)
	return systemPerl
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
