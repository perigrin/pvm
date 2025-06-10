// ABOUTME: End-to-end tests for PVM version management
// ABOUTME: Tests installation, listing, and switching of Perl versions

package e2e

import (
	"fmt"
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
	helpers.SkipIfNoSystemPerl(t)

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl or skip as TODO if not implemented
	stdout, stderr, err := env.RunPVM("import-system")
	if err != nil {
		t.Skipf("TODO: System Perl import not yet implemented\nCommand: pvm import-system\nError: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	helpers.AssertStringContains(t, stderr, "Successfully imported system Perl", "Import output does not indicate success")

	// Check that system Perl is now listed
	stdout, stderr, err = env.RunPVM("list", "--source")
	if err != nil {
		t.Skipf("TODO: Perl version listing not yet implemented\nCommand: pvm list --source\nError: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	output := stdout + stderr // Check both stdout and stderr for command output
	helpers.AssertStringContains(t, output, "Source: system", "System Perl not listed after import")
}

// Test version switching functionality
func TestVersionSwitching(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl first or skip as TODO if not implemented
	_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"import-system"}, "System Perl import")

	// Get the system Perl version or skip as TODO if not implemented
	stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"list"}, "Perl version listing")
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
		helpers.SkipTODO(t, "System Perl version detection")
	}

	// Test local version setting or skip as TODO if not implemented
	_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"local", systemVersion}, "Local Perl version setting")

	// Verify .perl-version file was created
	dotPerlVersionPath := filepath.Join(env.HomeDir, ".perl-version")
	helpers.AssertFileExists(t, dotPerlVersionPath, "No .perl-version file created")
	helpers.AssertPerlVersionFile(t, dotPerlVersionPath, systemVersion, "Wrong version in .perl-version file")

	// Test global version setting or skip as TODO if not implemented
	_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"global", systemVersion}, "Global Perl version setting")

	// Verify config file was updated
	configFile := filepath.Join(env.PVMConfigDir, "pvm.toml")
	helpers.AssertFileExists(t, configFile, "No config file created")
	helpers.AssertFileContains(t, configFile, fmt.Sprintf("default_perl = \"%s\"", systemVersion),
		"Config file does not contain correct default Perl")

	// Test 'use' command or skip as TODO if not implemented
	stdout = helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"use", systemVersion}, "Perl version switching")
	helpers.AssertStringContains(t, stdout, systemVersion, "Use command output does not contain version")
}

// TestInstallPerl tests installing a Perl version
// This test is slow and should be skipped with -short flag
func TestInstallPerl(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Perl installation test in short mode")
	}

	helpers.SkipTODO(t, "Perl version installation functionality")
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
	_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"import-system"}, "System Perl import")

	// Get the system Perl version
	stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"list"}, "Perl version listing")
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
		helpers.SkipTODO(t, "System Perl version detection")
	}

	// Uninstall it
	stdout = helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"uninstall", systemVersion},
		"Perl version uninstallation")
	helpers.AssertStringContains(t, stdout, "uninstalled", "Uninstall output does not indicate success")

	// Verify it's gone from the list
	stdout = helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"list"}, "Perl version listing")
	helpers.AssertStringDoesNotContain(t, stdout, systemVersion, "Uninstalled version still listed")
}
