// ABOUTME: Tests for system Perl detection functionality
// ABOUTME: Tests detection of installed Perl versions on the system

package perl

import (
	"testing"
)

func TestDetectSystemPerl(t *testing.T) {
	perl, err := DetectSystemPerl()
	if err != nil {
		t.Fatalf("Failed to detect system Perl: %v", err)
	}

	// Check that we got valid information
	if perl.Path == "" {
		t.Error("Expected Perl path to be non-empty")
	}

	if perl.Version == "" {
		t.Error("Expected Perl version to be non-empty")
	}

	if perl.Architecture == "" {
		t.Error("Expected Perl architecture to be non-empty")
	}

	if !perl.IsPrimary {
		t.Error("Expected IsPrimary to be true for system Perl")
	}

	t.Logf("Detected Perl: %s (version %s, %s)", perl.Path, perl.Version, perl.Architecture)
}

func TestDetectAllSystemPerls(t *testing.T) {
	perls, err := DetectAllSystemPerls()
	if err != nil {
		t.Logf("Warning: %v", err)
	}

	if len(perls) == 0 {
		t.Error("Expected to find at least one Perl installation")
	}

	// Check that at least one Perl is marked as primary
	foundPrimary := false
	for _, perl := range perls {
		if perl.IsPrimary {
			foundPrimary = true
			break
		}
	}

	if !foundPrimary && len(perls) > 0 {
		t.Error("Expected at least one Perl to be marked as primary")
	}

	// Log all detected Perls
	t.Logf("Detected %d Perl installations:", len(perls))
	for i, perl := range perls {
		t.Logf("  %d. %s (version %s, %s, primary: %v)",
			i+1, perl.Path, perl.Version, perl.Architecture, perl.IsPrimary)
	}
}

func TestGetSystemPerlVersion(t *testing.T) {
	// Test with empty path (should find perl in PATH)
	version, err := GetSystemPerlVersion("")
	if err != nil {
		t.Fatalf("Failed to get Perl version with empty path: %v", err)
	}

	if version == "" {
		t.Error("Expected version to be non-empty")
	}

	t.Logf("Perl version: %s", version)
}
