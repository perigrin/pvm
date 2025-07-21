// ABOUTME: Binary Perl helpers for reliable test infrastructure
// ABOUTME: Leverages PVM's binary distributions to eliminate system Perl dependencies

package helpers

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/platform"
	"tamarou.com/pvm/internal/xdg"
)

const (
	// Default test Perl version - aligns with PVM's standard binary distribution
	DefaultTestPerlVersion = "5.40.0"

	// Binary installation timeout for tests
	BinaryInstallTimeout = 5 * time.Minute
)

// TestPerlEnvironment represents a configured Perl environment for testing
type TestPerlEnvironment struct {
	Version    string
	PerlPath   string
	InstallDir string
	BinDir     string
}

// EnsureBinaryPerl ensures a binary Perl installation is available for testing
// This is the primary function tests should use to get reliable Perl access
func EnsureBinaryPerl(t *testing.T, version string) string {
	t.Helper()

	if version == "" {
		version = DefaultTestPerlVersion
	}

	// Check if already installed
	if path := getBinaryPerlPath(version); path != "" {
		t.Logf("Using existing binary Perl %s at %s", version, path)
		return path
	}

	// Install binary Perl
	t.Logf("Installing binary Perl %s for testing...", version)

	ctx, cancel := context.WithTimeout(context.Background(), BinaryInstallTimeout)
	defer cancel()

	options := &perl.BinaryInstallOptions{
		Version:  version,
		Platform: platform.GetPlatformTriple(),
		Context:  ctx,
	}

	result, err := perl.InstallFromBinary(options)
	if err != nil {
		// Try to provide helpful error information
		t.Logf("Binary Perl installation failed: %v", err)
		t.Logf("Platform: %s", platform.GetPlatformTriple())

		// Check if binary is available for this platform
		available, checkErr := perl.CheckBinaryAvailability(version, platform.GetPlatformTriple())
		if checkErr != nil {
			t.Logf("Could not check binary availability: %v", checkErr)
		} else if !available {
			t.Logf("Binary Perl %s not available for platform %s", version, platform.GetPlatformTriple())
		}

		t.Skipf("Failed to install binary Perl %s: %v", version, err)
	}

	perlPath := filepath.Join(result.InstallPath, "bin", getBinaryPerlCommand())
	t.Logf("Successfully installed binary Perl %s at %s", version, perlPath)

	return perlPath
}

// GetBinaryPerlPath returns the path to an installed binary Perl, or empty string if not found
func getBinaryPerlPath(version string) string {
	// Use XDG data directory structure that PVM uses
	dirs, err := xdg.GetDirs()
	if err != nil {
		return ""
	}
	dataDir := dirs.DataHome

	installDir := filepath.Join(dataDir, "pvm", "versions", version)
	perlPath := filepath.Join(installDir, "bin", getBinaryPerlCommand())

	// Check if the perl executable exists and is executable
	if stat, err := os.Stat(perlPath); err == nil && !stat.IsDir() {
		return perlPath
	}

	return ""
}

// SetupTestPerlEnvironment creates a complete test environment with binary Perl
// This includes PATH setup and environment variable configuration
func SetupTestPerlEnvironment(t *testing.T, version string) *TestPerlEnvironment {
	t.Helper()

	perlPath := EnsureBinaryPerl(t, version)

	// Get the installation directory
	binDir := filepath.Dir(perlPath)
	installDir := filepath.Dir(binDir)

	env := &TestPerlEnvironment{
		Version:    version,
		PerlPath:   perlPath,
		InstallDir: installDir,
		BinDir:     binDir,
	}

	// Set up environment for this test
	t.Setenv("PVM_TEST_PERL_VERSION", version)
	t.Setenv("PVM_TEST_PERL_PATH", perlPath)

	// Add Perl bin directory to PATH for this test
	currentPath := os.Getenv("PATH")
	newPath := binDir + string(os.PathListSeparator) + currentPath
	t.Setenv("PATH", newPath)

	t.Logf("Test environment configured for Perl %s", version)
	t.Logf("  Perl executable: %s", perlPath)
	t.Logf("  Install directory: %s", installDir)

	return env
}

// EnsureDefaultTestPerl is a convenience function that ensures the default test Perl version
func EnsureDefaultTestPerl(t *testing.T) string {
	t.Helper()
	return EnsureBinaryPerl(t, DefaultTestPerlVersion)
}

// SkipIfBinaryPerlUnavailable skips the test if binary Perl cannot be installed
// This is more reliable than system Perl checks and provides better error messages
func SkipIfBinaryPerlUnavailable(t *testing.T, version string) {
	t.Helper()

	if version == "" {
		version = DefaultTestPerlVersion
	}

	// Quick check - if already installed, no need to skip
	if getBinaryPerlPath(version) != "" {
		return
	}

	// Check if binary is available for this platform
	available, err := perl.CheckBinaryAvailability(version, platform.GetPlatformTriple())
	if err != nil {
		t.Skipf("Cannot check binary Perl availability: %v", err)
	}

	if !available {
		t.Skipf("Binary Perl %s not available for platform %s", version, platform.GetPlatformTriple())
	}

	// Try a quick install attempt
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	options := &perl.BinaryInstallOptions{
		Version:  version,
		Platform: platform.GetPlatformTriple(),
		Context:  ctx,
	}

	_, err = perl.InstallFromBinary(options)
	if err != nil {
		t.Skipf("Binary Perl %s installation failed: %v", version, err)
	}
}

// getBinaryPerlCommand returns the appropriate perl executable name for the platform
func getBinaryPerlCommand() string {
	if runtime.GOOS == "windows" {
		return "perl.exe"
	}
	return "perl"
}

// ValidatePerlInstallation validates that a Perl installation is working correctly
func ValidatePerlInstallation(t *testing.T, perlPath string) {
	t.Helper()

	// Check that the executable exists
	if stat, err := os.Stat(perlPath); err != nil || stat.IsDir() {
		t.Fatalf("Perl executable not found or is directory: %s", perlPath)
	}

	// TODO: Could add version check here if needed
	// cmd := exec.Command(perlPath, "-v")
	// output, err := cmd.Output()
	// if err != nil {
	//     t.Fatalf("Perl executable not working: %v", err)
	// }
}

// CleanupBinaryPerl removes a test Perl installation (useful for cleanup in tests)
// This should be used sparingly as binary Perl installations are designed to be reused
func CleanupBinaryPerl(t *testing.T, version string) {
	t.Helper()

	path := getBinaryPerlPath(version)
	if path == "" {
		return // Nothing to clean up
	}

	// Get install directory (two levels up from perl executable)
	installDir := filepath.Dir(filepath.Dir(path))

	err := os.RemoveAll(installDir)
	if err != nil {
		t.Logf("Failed to cleanup binary Perl %s: %v", version, err)
	} else {
		t.Logf("Cleaned up binary Perl %s from %s", version, installDir)
	}
}
