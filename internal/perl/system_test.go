// ABOUTME: Tests for system Perl detection functionality
// ABOUTME: Tests detection of installed Perl versions on the system

package perl

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDetectSystemPerl(t *testing.T) {
	// Skip test if perl is not installed
	_, err := exec.LookPath("perl")
	if err != nil {
		t.Skip("perl not found in PATH, skipping test")
	}

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
	// Skip test if perl is not installed
	_, err := exec.LookPath("perl")
	if err != nil {
		t.Skip("perl not found in PATH, skipping test")
	}

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
	// Skip test if perl is not installed
	perlPath, err := exec.LookPath("perl")
	if err != nil {
		t.Skip("perl not found in PATH, skipping test")
	}

	// Test with explicit path
	version, err := GetSystemPerlVersion(perlPath)
	if err != nil {
		t.Fatalf("Failed to get Perl version: %v", err)
	}

	if version == "" {
		t.Error("Expected version to be non-empty")
	}

	// Test with empty path (should find in PATH)
	version2, err := GetSystemPerlVersion("")
	if err != nil {
		t.Fatalf("Failed to get Perl version with empty path: %v", err)
	}

	if version2 == "" {
		t.Error("Expected version2 to be non-empty")
	}

	// Both versions should match
	if version != version2 {
		t.Errorf("Version mismatch: %s vs %s", version, version2)
	}

	t.Logf("Perl version: %s", version)
}

func TestMockPerlScript(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "perl-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a mock perl script
	mockPerlPath := filepath.Join(tempDir, "mockperl")
	if err := createMockPerlScript(mockPerlPath); err != nil {
		t.Fatalf("Failed to create mock Perl script: %v", err)
	}

	// Make it executable
	if err := os.Chmod(mockPerlPath, 0755); err != nil {
		t.Fatalf("Failed to make mock Perl script executable: %v", err)
	}

	// Test extraction with mock script
	perl, err := extractPerlInfo(mockPerlPath, false)
	if err != nil {
		t.Fatalf("Failed to extract info from mock Perl: %v", err)
	}

	// Check extracted information
	if perl.Version != "5.38.0" {
		t.Errorf("Expected version 5.38.0, got %s", perl.Version)
	}

	if perl.Architecture != "x86_64" {
		t.Errorf("Expected architecture x86_64, got %s", perl.Architecture)
	}

	if perl.IsPrimary {
		t.Error("Expected IsPrimary to be false for mock Perl")
	}
}

// Helper to create a mock perl script
func createMockPerlScript(path string) error {
	// Mock output simulating 'perl -v'
	mockOutput := `This is perl 5, version 38, subversion 0 (v5.38.0) built for x86_64-linux

	Copyright 1987-2023, Larry Wall

	Perl may be copied only under the terms of either the Artistic License or the
	GNU General Public License, which may be found in the Perl 5 source kit.

	Complete documentation for Perl, including FAQ lists, should be found on
	this system using "man perl" or "perldoc perl".  If you have access to the
	Internet, point your browser at https://www.perl.org/, the Perl Home Page.

	Summary of my perl5 (revision 5 version 38 subversion 0) configuration:
	  Platform: x86_64
		osname=linux
		osvers=6.2.0-26-generic
		archname=x86_64-linux
		uname='linux laptop 6.2.0-26-generic #26~22.04.1-ubuntu smp ubuntu x86_64 gnulinux '
		config_args='-des -Dusedevel -Dprefix=/home/user/perl-5.38.0'
		hint=recommended
		useposix=true
		d_sigaction=define
		useithreads=define
		usemultiplicity=define
		use64bitint=define
		use64bitall=define
		uselongdouble=undef
		usemymalloc=n
		default_inc_excludes_dot=define
	`

	// Create the script
	script := []byte("#!/bin/sh\n\nif [ \"$1\" = \"-v\" ]; then\n  cat <<'EOT'\n" + mockOutput + "\nEOT\nelif [ \"$1\" = \"-e\" ]; then\n  echo \"v5.38.0\"\nfi\n")

	return os.WriteFile(path, script, 0644)
}
