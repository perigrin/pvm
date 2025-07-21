// ABOUTME: Tests for binary Perl test helpers
// ABOUTME: Validates that binary Perl installation works correctly in test environments

package helpers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureBinaryPerl(t *testing.T) {
	// Test that we can get a binary Perl installation
	perlPath := EnsureBinaryPerl(t, DefaultTestPerlVersion)

	// Verify the perl executable exists
	if _, err := os.Stat(perlPath); err != nil {
		t.Fatalf("Binary Perl executable not found at %s: %v", perlPath, err)
	}

	// Verify it's executable
	if stat, err := os.Stat(perlPath); err != nil || stat.IsDir() {
		t.Fatalf("Binary Perl path is not a valid executable: %s", perlPath)
	}

	t.Logf("Successfully ensured binary Perl at: %s", perlPath)
}

func TestSetupTestPerlEnvironment(t *testing.T) {
	// Test that we can set up a complete test environment
	env := SetupTestPerlEnvironment(t, DefaultTestPerlVersion)

	// Verify the environment is properly configured
	if env.Version != DefaultTestPerlVersion {
		t.Errorf("Expected version %s, got %s", DefaultTestPerlVersion, env.Version)
	}

	if env.PerlPath == "" {
		t.Error("PerlPath should not be empty")
	}

	if env.InstallDir == "" {
		t.Error("InstallDir should not be empty")
	}

	if env.BinDir == "" {
		t.Error("BinDir should not be empty")
	}

	// Verify the directory structure makes sense
	expectedBinDir := filepath.Dir(env.PerlPath)
	if env.BinDir != expectedBinDir {
		t.Errorf("BinDir %s doesn't match Perl path parent %s", env.BinDir, expectedBinDir)
	}

	expectedInstallDir := filepath.Dir(env.BinDir)
	if env.InstallDir != expectedInstallDir {
		t.Errorf("InstallDir %s doesn't match expected %s", env.InstallDir, expectedInstallDir)
	}

	// Verify environment variables were set
	if testVersion := os.Getenv("PVM_TEST_PERL_VERSION"); testVersion != DefaultTestPerlVersion {
		t.Errorf("PVM_TEST_PERL_VERSION not set correctly: got %s, expected %s", testVersion, DefaultTestPerlVersion)
	}

	if testPath := os.Getenv("PVM_TEST_PERL_PATH"); testPath != env.PerlPath {
		t.Errorf("PVM_TEST_PERL_PATH not set correctly: got %s, expected %s", testPath, env.PerlPath)
	}

	t.Logf("Successfully set up test environment for Perl %s", env.Version)
}

func TestGetBinaryPerlPath(t *testing.T) {
	// First ensure we have binary Perl installed
	EnsureBinaryPerl(t, DefaultTestPerlVersion)

	// Now test that getBinaryPerlPath can find it
	perlPath := getBinaryPerlPath(DefaultTestPerlVersion)

	if perlPath == "" {
		t.Fatal("getBinaryPerlPath returned empty string for installed version")
	}

	// Verify it exists
	if _, err := os.Stat(perlPath); err != nil {
		t.Fatalf("Perl executable not found at returned path %s: %v", perlPath, err)
	}

	t.Logf("Found binary Perl at: %s", perlPath)
}

func TestValidatePerlInstallation(t *testing.T) {
	perlPath := EnsureBinaryPerl(t, DefaultTestPerlVersion)

	// This should not fail for a properly installed binary Perl
	ValidatePerlInstallation(t, perlPath)

	t.Logf("Successfully validated Perl installation at: %s", perlPath)
}
