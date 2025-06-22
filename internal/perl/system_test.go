// ABOUTME: Tests for system Perl detection functionality
// ABOUTME: Tests detection of installed Perl versions on the system

package perl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSystemPerl(t *testing.T) {
	// Create test Perl executable for controlled testing
	tempDir, err := os.MkdirTemp("", "perl-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a test perl executable with known version info
	testPerlPath := filepath.Join(tempDir, "perl")
	if err := createTestPerlExecutable(testPerlPath); err != nil {
		t.Fatalf("Failed to create test Perl executable: %v", err)
	}

	// Replace DetectSystemPerl with a version that uses our test executable
	originalDetectSystemPerl := DetectSystemPerl
	defer func() { DetectSystemPerl = originalDetectSystemPerl }()

	DetectSystemPerl = func() (*SystemPerl, error) {
		return extractPerlInfo(testPerlPath, true)
	}

	perl, err := DetectSystemPerl()
	if err != nil {
		t.Fatalf("Failed to detect system Perl: %v", err)
	}

	// Check that we got valid information from our test executable
	if perl.Path == "" {
		t.Error("Expected Perl path to be non-empty")
	}

	if perl.Version != "5.38.0" {
		t.Errorf("Expected version 5.38.0, got %s", perl.Version)
	}

	if perl.Architecture != "x86_64" {
		t.Errorf("Expected architecture x86_64, got %s", perl.Architecture)
	}

	if !perl.IsPrimary {
		t.Error("Expected IsPrimary to be true for system Perl")
	}

	t.Logf("Detected Perl: %s (version %s, %s)", perl.Path, perl.Version, perl.Architecture)
}

func TestDetectAllSystemPerls(t *testing.T) {
	// Create test Perl installations for controlled testing
	tempDir, err := os.MkdirTemp("", "perl-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create primary test perl executable
	testPerlPath1 := filepath.Join(tempDir, "perl1")
	if err := createTestPerlExecutable(testPerlPath1); err != nil {
		t.Fatalf("Failed to create test Perl executable 1: %v", err)
	}

	// Create secondary test perl executable
	testPerlPath2 := filepath.Join(tempDir, "perl2")
	if err := createTestPerlExecutable(testPerlPath2); err != nil {
		t.Fatalf("Failed to create test Perl executable 2: %v", err)
	}

	// Replace DetectSystemPerl with a version that uses our primary test executable
	originalDetectSystemPerl := DetectSystemPerl
	defer func() { DetectSystemPerl = originalDetectSystemPerl }()

	DetectSystemPerl = func() (*SystemPerl, error) {
		return extractPerlInfo(testPerlPath1, true)
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
	// Create test Perl executable for controlled testing
	tempDir, err := os.MkdirTemp("", "perl-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test perl executable
	testPerlPath := filepath.Join(tempDir, "perl")
	if err := createTestPerlExecutable(testPerlPath); err != nil {
		t.Fatalf("Failed to create test Perl executable: %v", err)
	}

	// Test with explicit path
	version, err := GetSystemPerlVersion(testPerlPath)
	if err != nil {
		t.Fatalf("Failed to get Perl version: %v", err)
	}

	if version == "" {
		t.Error("Expected version to be non-empty")
	}

	// Expected version from test executable
	expectedVersion := "5.38.0"
	if version != expectedVersion {
		t.Errorf("Expected version %q, got %q (len=%d)", expectedVersion, version, len(version))
	}

	t.Logf("Perl version: %s", version)
}

func TestTestPerlExecutable(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "perl-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test perl executable
	testPerlPath := filepath.Join(tempDir, "testperl")
	if err := createTestPerlExecutable(testPerlPath); err != nil {
		t.Fatalf("Failed to create test Perl executable: %v", err)
	}

	// Test extraction with test executable
	perl, err := extractPerlInfo(testPerlPath, false)
	if err != nil {
		t.Fatalf("Failed to extract info from test Perl: %v", err)
	}

	// Check extracted information
	if perl.Version != "5.38.0" {
		t.Errorf("Expected version 5.38.0, got %s", perl.Version)
	}

	if perl.Architecture != "x86_64" {
		t.Errorf("Expected architecture x86_64, got %s", perl.Architecture)
	}

	if perl.IsPrimary {
		t.Error("Expected IsPrimary to be false for test Perl")
	}
}

// createTestPerlExecutable creates a test Perl executable that responds to -v and -e flags
// This simulates a real Perl installation for testing PVM's detection logic
func createTestPerlExecutable(path string) error {
	// Create a shell script that mimics perl behavior for testing
	perlVersionOutput := `This is perl 5, version 38, subversion 0 (v5.38.0) built for x86_64-linux

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
	uname='linux test-system 6.2.0-26-generic #26~22.04.1-ubuntu smp ubuntu x86_64 gnulinux '
	config_args='-des -Dusedevel -Dprefix=/tmp/test-perl-5.38.0'
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

	// Create shell script that handles both -v and -e flags like real perl
	script := `#!/bin/sh

case "$1" in
  "-v")
    cat <<'EOT'
` + perlVersionOutput + `
EOT
    ;;
  "-e")
    # Handle -e flag for version extraction (perl -e 'print $^V')
    if [ "$2" = "print \$^V" ] || [ "$2" = 'print $^V' ]; then
      printf "v5.38.0"
    else
      echo "Test perl executable"
    fi
    ;;
  *)
    echo "Test perl executable - use -v for version info"
    ;;
esac
`

	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		return err
	}

	return nil
}
