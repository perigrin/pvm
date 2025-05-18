// ABOUTME: End-to-end tests for PVM version management
// ABOUTME: Tests installation, listing, and switching of Perl versions

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/test/e2e/helpers"
)

// Test basic version command functionality
func TestVersionCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test version command
	stdout := helpers.AssertPVMSucceeds(t, env, []string{"version"}, "PVM version command failed")
	helpers.AssertStringContains(t, stdout, "pvm", "Version output does not contain version information")
}

// Test importing system Perl
func TestImportSystemPerl(t *testing.T) {
	// Check if perl is installed
	_, err := os.Stat("/usr/bin/perl")
	if os.IsNotExist(err) {
		t.Skip("System Perl not found, skipping test")
	}

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl
	stdout := helpers.AssertPVMSucceeds(t, env, []string{"import-system"}, "Failed to import system Perl")
	helpers.AssertStringContains(t, stdout, "Imported system Perl", "Import output does not indicate success")

	// Check that system Perl is now listed
	stdout = helpers.AssertPVMSucceeds(t, env, []string{"list"}, "Failed to list Perl versions")
	helpers.AssertStringContains(t, stdout, "system", "System Perl not listed after import")
}

// Test version switching functionality
func TestVersionSwitching(t *testing.T) {
	// This test requires a real Perl version, so we'll import system Perl
	_, err := os.Stat("/usr/bin/perl")
	if os.IsNotExist(err) {
		t.Skip("System Perl not found, skipping test")
	}

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl first
	helpers.AssertPVMSucceeds(t, env, []string{"import-system"}, "Failed to import system Perl")

	// Get the system Perl version
	stdout := helpers.AssertPVMSucceeds(t, env, []string{"list"}, "Failed to list Perl versions")
	lines := strings.Split(stdout, "\n")
	var systemVersion string
	for _, line := range lines {
		if strings.Contains(line, "system") {
			// Extract version from line like "* 5.34.1 (system)"
			parts := strings.Fields(line)
			if len(parts) > 1 {
				systemVersion = parts[1]
				break
			}
		}
	}

	if systemVersion == "" {
		t.Fatal("Could not determine system Perl version")
	}

	// Test local version setting
	helpers.AssertPVMSucceeds(t, env, []string{"local", systemVersion}, "Failed to set local Perl version")

	// Verify .perl-version file was created
	dotPerlVersionPath := filepath.Join(env.HomeDir, ".perl-version")
	helpers.AssertFileExists(t, dotPerlVersionPath, "No .perl-version file created")
	helpers.AssertPerlVersionFile(t, dotPerlVersionPath, systemVersion, "Wrong version in .perl-version file")

	// Test global version setting
	helpers.AssertPVMSucceeds(t, env, []string{"global", systemVersion}, "Failed to set global Perl version")

	// Verify config file was updated
	configFile := filepath.Join(env.PVMConfigDir, "pvm.toml")
	helpers.AssertFileExists(t, configFile, "No config file created")
	helpers.AssertFileContains(t, configFile, fmt.Sprintf("default_perl = \"%s\"", systemVersion),
		"Config file does not contain correct default Perl")

	// Test 'use' command
	stdout = helpers.AssertPVMSucceeds(t, env, []string{"use", systemVersion}, "Failed to use Perl version")
	helpers.AssertStringContains(t, stdout, systemVersion, "Use command output does not contain version")
}

// TestInstallPerl tests installing a Perl version
// This test is slow and should be skipped with -short flag
func TestInstallPerl(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Perl installation test in short mode")
	}

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Install a small/fast-to-build Perl version
	// Using 5.12.5 as an example of a smaller, faster build
	version := "5.12.5"

	// This could take several minutes
	t.Logf("Installing Perl %s (this may take a while)...", version)
	stdout := helpers.AssertPVMSucceeds(t, env, []string{"install", version},
		fmt.Sprintf("Failed to install Perl %s", version))

	helpers.AssertStringContains(t, stdout, "Successfully installed", "Install output does not indicate success")

	// Verify installation
	stdout = helpers.AssertPVMSucceeds(t, env, []string{"list"}, "Failed to list Perl versions")
	helpers.AssertStringContains(t, stdout, version, "Installed version not listed")

	// Verify binary exists
	perlBin := filepath.Join(env.PVMDataDir, "versions", version, "bin", "perl")
	helpers.AssertFileExists(t, perlBin, "Perl binary not found at expected location")

	// Test using the installed version
	helpers.AssertPVMSucceeds(t, env, []string{"use", version}, "Failed to use installed Perl version")
}

// TestUninstallPerl tests uninstalling a Perl version
// This test depends on TestInstallPerl and is also slow
func TestUninstallPerl(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Perl uninstallation test in short mode")
	}

	// This depends on TestInstallPerl
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl to have something we can uninstall
	helpers.AssertPVMSucceeds(t, env, []string{"import-system"}, "Failed to import system Perl")

	// Get the system Perl version
	stdout := helpers.AssertPVMSucceeds(t, env, []string{"list"}, "Failed to list Perl versions")
	lines := strings.Split(stdout, "\n")
	var systemVersion string
	for _, line := range lines {
		if strings.Contains(line, "system") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				systemVersion = parts[1]
				break
			}
		}
	}

	if systemVersion == "" {
		t.Fatal("Could not determine system Perl version")
	}

	// Uninstall it
	stdout = helpers.AssertPVMSucceeds(t, env, []string{"uninstall", systemVersion},
		"Failed to uninstall Perl version")
	helpers.AssertStringContains(t, stdout, "uninstalled", "Uninstall output does not indicate success")

	// Verify it's gone from the list
	stdout = helpers.AssertPVMSucceeds(t, env, []string{"list"}, "Failed to list Perl versions")
	helpers.AssertStringDoesNotContain(t, stdout, systemVersion, "Uninstalled version still listed")
}
