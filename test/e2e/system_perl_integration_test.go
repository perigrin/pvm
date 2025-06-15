// ABOUTME: Integration tests for SystemPerlManager functionality
// ABOUTME: Tests cross-platform system Perl detection and management

package e2e

import (
	"os"
	"testing"

	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/test/e2e/helpers"
)

// TestSystemPerlManagerDetection tests that SystemPerlManager can detect system Perl
func TestSystemPerlManagerDetection(t *testing.T) {
	manager := perl.NewSystemPerlManager()

	// Test detection
	systemPerl, err := manager.DetectOrInstallPerl()
	if err != nil {
		// If detection fails and we're in CI, that's expected
		if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
			t.Skipf("SystemPerl detection failed in CI environment: %v", err)
		}
		// Otherwise, this might be a real issue
		t.Errorf("SystemPerlManager.DetectOrInstallPerl() failed: %v", err)
		return
	}

	if systemPerl == nil {
		t.Fatal("SystemPerlManager returned nil SystemPerl")
	}

	// Validate the detected Perl
	err = manager.ValidateInstallation(systemPerl)
	if err != nil {
		t.Errorf("System Perl validation failed: %v", err)
	}

	t.Logf("Successfully detected system Perl: %s at %s", systemPerl.Version, systemPerl.Path)
}

// TestSystemPerlManagerPlatformDistributions tests platform-specific distribution detection
func TestSystemPerlManagerPlatformDistributions(t *testing.T) {
	manager := perl.NewSystemPerlManager()

	availableDistributions := manager.GetAvailableDistributions()
	preferredDistributions := manager.GetPreferredDistributions()

	t.Logf("Available distributions: %d", len(availableDistributions))
	for i, dist := range availableDistributions {
		t.Logf("  %d. %s", i+1, dist.String())
	}

	t.Logf("Preferred distributions: %d", len(preferredDistributions))
	for i, dist := range preferredDistributions {
		t.Logf("  %d. %s", i+1, dist.String())
	}

	// Should have at least some preferences based on platform
	if len(preferredDistributions) == 0 {
		t.Error("No preferred distributions detected for current platform")
	}
}

// TestSystemPerlHelperIntegration tests the updated E2E helper functions
func TestSystemPerlHelperIntegration(t *testing.T) {
	// Test the basic detection
	if !helpers.HasSystemPerl() {
		t.Skip("System Perl not available for helper integration test")
	}

	perlPath := helpers.FindSystemPerl()
	if perlPath == "" {
		t.Error("FindSystemPerl() returned empty path when HasSystemPerl() returned true")
	}

	t.Logf("Helper found system Perl at: %s", perlPath)
}

// TestSystemPerlEnforcement tests the strict enforcement of system Perl
func TestSystemPerlEnforcement(t *testing.T) {
	// Skip in short mode to avoid long installation times
	if testing.Short() {
		t.Skip("Skipping enforcement test in short mode")
	}

	// Skip in CI to avoid permission issues
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping enforcement test in CI environment")
	}

	// Test the EnsureSystemPerl function
	systemPerl := helpers.EnsureSystemPerl(t)
	if systemPerl == nil {
		t.Fatal("EnsureSystemPerl returned nil")
	}

	t.Logf("EnsureSystemPerl succeeded: %s at %s", systemPerl.Version, systemPerl.Path)
}

// TestSystemPerlValidation tests the validation functionality
func TestSystemPerlValidation(t *testing.T) {
	// Get system Perl using the manager
	manager := perl.NewSystemPerlManager()
	systemPerl, err := manager.DetectOrInstallPerl()
	if err != nil {
		t.Skipf("System Perl not available: %v", err)
	}

	// Test validation
	err = manager.ValidateInstallation(systemPerl)
	if err != nil {
		t.Errorf("System Perl validation failed: %v", err)
	}

	// Test validation with invalid Perl
	invalidPerl := &perl.SystemPerl{
		Path:    "/definitely/not/a/real/perl/path",
		Version: "5.999.999",
	}

	err = manager.ValidateInstallation(invalidPerl)
	if err == nil {
		t.Error("Validation should fail for invalid Perl path")
	}
}

// TestSystemPerlUpdates tests the update checking functionality
func TestSystemPerlUpdates(t *testing.T) {
	manager := perl.NewSystemPerlManager()

	hasUpdates, err := manager.CheckForUpdates()
	if err != nil {
		t.Errorf("CheckForUpdates() failed: %v", err)
	}

	// For now, this should always return false (not implemented)
	if hasUpdates {
		t.Error("CheckForUpdates() should return false (not yet implemented)")
	}

	t.Log("Update checking works as expected (placeholder implementation)")
}
