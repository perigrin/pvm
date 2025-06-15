// ABOUTME: Tests for cross-platform system Perl detection and installation management
// ABOUTME: Provides comprehensive test coverage for SystemPerlManager functionality

package perl

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

func TestNewSystemPerlManager(t *testing.T) {
	manager := NewSystemPerlManager()
	if manager == nil {
		t.Fatal("NewSystemPerlManager() returned nil")
	}
	
	// Should have detected some distributions based on platform
	if len(manager.availableDistributions) == 0 && len(manager.preferredDistributions) == 0 {
		t.Log("No distributions detected - this is expected in test environments")
	}
}

func TestPerlDistributionString(t *testing.T) {
	testCases := []struct {
		dist     PerlDistribution
		expected string
	}{
		{DistributionSystem, "system"},
		{DistributionStrawberry, "strawberry"},
		{DistributionActivePerl, "activeperl"},
		{DistributionHomebrew, "homebrew"},
		{DistributionApt, "apt"},
		{DistributionYum, "yum"},
		{DistributionDnf, "dnf"},
		{DistributionPacman, "pacman"},
		{DistributionZypper, "zypper"},
		{DistributionPerlBrew, "perlbrew"},
		{DistributionPlenv, "plenv"},
		{PerlDistribution(999), "unknown"},
	}
	
	for _, tc := range testCases {
		result := tc.dist.String()
		if result != tc.expected {
			t.Errorf("Distribution %d.String() = %q, expected %q", tc.dist, result, tc.expected)
		}
	}
}

func TestDetectAvailableDistributions(t *testing.T) {
	manager := &SystemPerlManager{}
	manager.detectAvailableDistributions()
	
	// The results depend on what's actually installed on the system
	// We can't assert specific distributions, but we can verify the logic
	
	switch runtime.GOOS {
	case "windows":
		// On Windows, might have choco, scoop, or winget
		t.Logf("Windows: detected %d distributions", len(manager.availableDistributions))
	case "darwin":
		// On macOS, might have brew
		t.Logf("macOS: detected %d distributions", len(manager.availableDistributions))
	case "linux":
		// On Linux, might have various package managers
		t.Logf("Linux: detected %d distributions", len(manager.availableDistributions))
	}
	
	// Check for cross-platform tools
	if _, err := exec.LookPath("perlbrew"); err == nil {
		found := false
		for _, dist := range manager.availableDistributions {
			if dist == DistributionPerlBrew {
				found = true
				break
			}
		}
		if !found {
			t.Error("perlbrew is available but not detected")
		}
	}
}

func TestSetPreferredDistributions(t *testing.T) {
	manager := &SystemPerlManager{}
	manager.setPreferredDistributions()
	
	switch runtime.GOOS {
	case "windows":
		if len(manager.preferredDistributions) == 0 {
			t.Error("Windows should have preferred distributions")
		}
		if manager.preferredDistributions[0] != DistributionStrawberry {
			t.Error("Windows should prefer Strawberry Perl first")
		}
	case "darwin":
		if len(manager.preferredDistributions) == 0 {
			t.Error("macOS should have preferred distributions")
		}
		if manager.preferredDistributions[0] != DistributionHomebrew {
			t.Error("macOS should prefer Homebrew first")
		}
	case "linux":
		if len(manager.preferredDistributions) == 0 {
			t.Error("Linux should have preferred distributions")
		}
		// Linux preference order depends on what's available
	}
}

func TestHasCommand(t *testing.T) {
	manager := &SystemPerlManager{}
	
	// Test with a command that should exist
	if !manager.hasCommand("echo") {
		t.Error("hasCommand('echo') should return true")
	}
	
	// Test with a command that shouldn't exist
	if manager.hasCommand("definitely-not-a-real-command-12345") {
		t.Error("hasCommand() should return false for non-existent commands")
	}
}

func TestDetectOrInstallPerl(t *testing.T) {
	manager := NewSystemPerlManager()
	
	// Try to detect or install Perl
	perl, err := manager.DetectOrInstallPerl()
	
	// If system Perl is already available, this should succeed
	if existingPerl, existingErr := DetectSystemPerl(); existingErr == nil {
		if err != nil {
			t.Errorf("DetectOrInstallPerl() failed when system Perl exists: %v", err)
		}
		if perl == nil {
			t.Error("DetectOrInstallPerl() returned nil Perl when system Perl exists")
		}
		if perl.Path != existingPerl.Path {
			t.Errorf("DetectOrInstallPerl() returned different path: got %s, expected %s", perl.Path, existingPerl.Path)
		}
	} else {
		// System Perl not available - installation would be attempted
		// In test environment, this might fail due to lack of sudo/admin privileges
		t.Logf("System Perl not available, installation attempt result: err=%v", err)
	}
}

func TestInstallSystemPerl(t *testing.T) {
	manager := NewSystemPerlManager()
	
	// Skip if no installation methods available
	if len(manager.availableDistributions) == 0 {
		t.Skip("No installation methods available in test environment")
	}
	
	// Skip if system Perl already exists (don't want to reinstall)
	if _, err := DetectSystemPerl(); err == nil {
		t.Skip("System Perl already exists, skipping installation test")
	}
	
	// Skip installation test in CI/automated environments
	if os.Getenv("CI") != "" || os.Getenv("AUTOMATED_TESTING") != "" {
		t.Skip("Skipping installation test in CI/automated environment")
	}
	
	// This test would actually try to install Perl, which requires admin privileges
	// and is not suitable for automated testing
	t.Skip("Installation test requires admin privileges and is not suitable for automated testing")
}

func TestValidateInstallation(t *testing.T) {
	manager := &SystemPerlManager{}
	
	// Test with nil
	err := manager.ValidateInstallation(nil)
	if err == nil {
		t.Error("ValidateInstallation(nil) should return error")
	}
	
	// Test with non-existent path
	badPerl := &SystemPerl{
		Path:    "/definitely/not/a/real/path/perl",
		Version: "5.34.0",
	}
	err = manager.ValidateInstallation(badPerl)
	if err == nil {
		t.Error("ValidateInstallation() should fail for non-existent path")
	}
	
	// Test with system Perl if available
	if systemPerl, err := DetectSystemPerl(); err == nil {
		err = manager.ValidateInstallation(systemPerl)
		if err != nil {
			t.Errorf("ValidateInstallation() failed for system Perl: %v", err)
		}
	}
}

func TestGetAvailableDistributions(t *testing.T) {
	manager := NewSystemPerlManager()
	distributions := manager.GetAvailableDistributions()
	
	// Should return the same as the internal field
	if len(distributions) != len(manager.availableDistributions) {
		t.Error("GetAvailableDistributions() length doesn't match internal field")
	}
	
	for i, dist := range distributions {
		if dist != manager.availableDistributions[i] {
			t.Errorf("GetAvailableDistributions()[%d] = %v, expected %v", i, dist, manager.availableDistributions[i])
		}
	}
}

func TestGetPreferredDistributions(t *testing.T) {
	manager := NewSystemPerlManager()
	distributions := manager.GetPreferredDistributions()
	
	// Should return the same as the internal field
	if len(distributions) != len(manager.preferredDistributions) {
		t.Error("GetPreferredDistributions() length doesn't match internal field")
	}
	
	for i, dist := range distributions {
		if dist != manager.preferredDistributions[i] {
			t.Errorf("GetPreferredDistributions()[%d] = %v, expected %v", i, dist, manager.preferredDistributions[i])
		}
	}
}

func TestCheckForUpdates(t *testing.T) {
	manager := &SystemPerlManager{}
	
	hasUpdates, err := manager.CheckForUpdates()
	if err != nil {
		t.Errorf("CheckForUpdates() returned error: %v", err)
	}
	
	// For now, this should always return false
	if hasUpdates {
		t.Error("CheckForUpdates() should return false (not implemented)")
	}
}

// Integration test that requires system Perl
func TestSystemPerlManagerIntegration(t *testing.T) {
	// Skip if no system Perl available
	if _, err := DetectSystemPerl(); err != nil {
		t.Skip("System Perl not available for integration test")
	}
	
	manager := NewSystemPerlManager()
	
	// Test full workflow
	perl, err := manager.DetectOrInstallPerl()
	if err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}
	
	if perl == nil {
		t.Fatal("Integration test returned nil Perl")
	}
	
	// Validate the installation
	err = manager.ValidateInstallation(perl)
	if err != nil {
		t.Errorf("Installation validation failed: %v", err)
	}
	
	t.Logf("Integration test successful: Perl %s at %s", perl.Version, perl.Path)
}