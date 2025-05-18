// ABOUTME: End-to-end tests for PVM shim functionality
// ABOUTME: Tests shim creation, execution, and rehashing

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/test/e2e/helpers"
)

// TestShimCreation tests the creation of shims
func TestShimCreation(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl first
	_, err := os.Stat("/usr/bin/perl")
	if os.IsNotExist(err) {
		t.Skip("System Perl not found, skipping test")
	}

	// Import system Perl
	helpers.AssertPVMSucceeds(t, env, []string{"import-system"}, "Failed to import system Perl")

	// Run rehash command to create shims
	stdout := helpers.AssertPVMSucceeds(t, env, []string{"rehash"}, "Failed to rehash shims")
	helpers.AssertStringContains(t, stdout, "Rehashed", "Rehash output does not indicate success")

	// Check that perl shim was created
	perlShimPath := filepath.Join(env.PVMShimsDir, "perl")
	helpers.AssertFileExists(t, perlShimPath, "Perl shim not created")

	// Check other common Perl tool shims
	commonTools := []string{"perldoc", "cpan"}
	for _, tool := range commonTools {
		shimPath := filepath.Join(env.PVMShimsDir, tool)
		if _, err := os.Stat(shimPath); err == nil {
			// Not all tools will be available in all Perl installations,
			// so we just log which ones we found
			t.Logf("Found shim for %s", tool)
		}
	}
}

// TestShimExecutable tests that shims are executable
func TestShimExecutable(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl first
	_, err := os.Stat("/usr/bin/perl")
	if os.IsNotExist(err) {
		t.Skip("System Perl not found, skipping test")
	}

	// Import system Perl
	helpers.AssertPVMSucceeds(t, env, []string{"import-system"}, "Failed to import system Perl")

	// Run rehash command to create shims
	helpers.AssertPVMSucceeds(t, env, []string{"rehash"}, "Failed to rehash shims")

	// Check that perl shim is executable
	perlShimPath := filepath.Join(env.PVMShimsDir, "perl")
	fileInfo, err := os.Stat(perlShimPath)
	if err != nil {
		t.Fatalf("Failed to get stats for perl shim: %v", err)
	}

	if fileInfo.Mode()&0111 == 0 {
		t.Error("Perl shim is not executable")
	}

	// Try to execute the perl shim
	stdout, stderr, err := env.RunCommand(perlShimPath, "-v")
	if err != nil {
		t.Errorf("Failed to execute perl shim: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	helpers.AssertStringContains(t, stdout, "This is perl", "Perl shim execution does not show version info")
}

// TestShimPathPriority tests that shims take priority in PATH
func TestShimPathPriority(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl first
	_, err := os.Stat("/usr/bin/perl")
	if os.IsNotExist(err) {
		t.Skip("System Perl not found, skipping test")
	}

	// Import system Perl
	helpers.AssertPVMSucceeds(t, env, []string{"import-system"}, "Failed to import system Perl")

	// Run rehash command to create shims
	helpers.AssertPVMSucceeds(t, env, []string{"rehash"}, "Failed to rehash shims")

	// Initialize shell integration
	helpers.AssertPVMSucceeds(t, env, []string{"shell", "init"}, "Failed to initialize shell")

	// Create a test script to check which perl is found in PATH
	testScript := filepath.Join(env.HomeDir, "test_which.sh")
	scriptContent := `#!/bin/bash
source "` + filepath.Join(env.PVMDataDir, "shell", "pvm.bash") + `"
which perl
`
	err = os.WriteFile(testScript, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Run the test script
	stdout, stderr, err := env.RunCommand("bash", testScript)
	if err != nil {
		t.Fatalf("Failed to run test script: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Check that the shim is found first
	helpers.AssertStringContains(t, stdout, env.PVMShimsDir,
		"Shim is not found first in PATH")
}

// TestRehashCommand tests the rehash command
func TestRehashCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl first
	_, err := os.Stat("/usr/bin/perl")
	if os.IsNotExist(err) {
		t.Skip("System Perl not found, skipping test")
	}

	// Import system Perl
	helpers.AssertPVMSucceeds(t, env, []string{"import-system"}, "Failed to import system Perl")

	// Get initial state before rehash
	initialFiles, err := os.ReadDir(env.PVMShimsDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to read shims directory: %v", err)
	}
	initialCount := len(initialFiles)

	// Run rehash command
	stdout := helpers.AssertPVMSucceeds(t, env, []string{"rehash"}, "Failed to rehash shims")
	helpers.AssertStringContains(t, stdout, "Rehashed", "Rehash output does not indicate success")

	// Check that shims were created
	newFiles, err := os.ReadDir(env.PVMShimsDir)
	if err != nil {
		t.Fatalf("Failed to read shims directory after rehash: %v", err)
	}

	if len(newFiles) <= initialCount {
		t.Errorf("No new shims created by rehash (before: %d, after: %d)",
			initialCount, len(newFiles))
	}

	t.Logf("Rehash created %d shims", len(newFiles)-initialCount)

	// Run rehash again and verify it doesn't create duplicate shims
	helpers.AssertPVMSucceeds(t, env, []string{"rehash"}, "Failed to rehash shims a second time")

	finalFiles, err := os.ReadDir(env.PVMShimsDir)
	if err != nil {
		t.Fatalf("Failed to read shims directory after second rehash: %v", err)
	}

	if len(finalFiles) != len(newFiles) {
		t.Errorf("Second rehash changed the number of shims (before: %d, after: %d)",
			len(newFiles), len(finalFiles))
	}
}

// TestShimVersionResolution tests that shims use the correct Perl version
func TestShimVersionResolution(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl first
	_, err := os.Stat("/usr/bin/perl")
	if os.IsNotExist(err) {
		t.Skip("System Perl not found, skipping test")
	}

	// Import system Perl
	helpers.AssertPVMSucceeds(t, env, []string{"import-system"}, "Failed to import system Perl")

	// Run rehash command to create shims
	helpers.AssertPVMSucceeds(t, env, []string{"rehash"}, "Failed to rehash shims")

	// Get the system Perl path
	systemPerlPath, stderr, err := env.RunCommand("which", "perl")
	if err != nil {
		t.Fatalf("Failed to find system perl: %v\nStderr: %s", err, stderr)
	}
	systemPerlPath = strings.TrimSpace(systemPerlPath)

	// Create a .perl-version file
	dotPerlVersionPath := filepath.Join(env.HomeDir, ".perl-version")

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

	err = os.WriteFile(dotPerlVersionPath, []byte(systemVersion), 0644)
	if err != nil {
		t.Fatalf("Failed to create .perl-version file: %v", err)
	}

	// Initialize shell integration
	helpers.AssertPVMSucceeds(t, env, []string{"shell", "init"}, "Failed to initialize shell")

	// Create a test script to check which Perl is executed through the shim
	testScript := filepath.Join(env.HomeDir, "test_shim.sh")
	scriptContent := `#!/bin/bash
source "` + filepath.Join(env.PVMDataDir, "shell", "pvm.bash") + `"
# Use the shim to get the perl path
perl -e 'print $^X, "\n"'
`
	err = os.WriteFile(testScript, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Run the test script
	var shimStdout, shimStderr string
	shimStdout, shimStderr, err = env.RunCommand("bash", testScript)
	if err != nil {
		t.Fatalf("Failed to run test script: %v\nStdout: %s\nStderr: %s", err, shimStdout, shimStderr)
	}

	// The shim should be using the real Perl binary, not itself
	if strings.Contains(shimStdout, env.PVMShimsDir) {
		t.Errorf("Shim is not resolving to the real Perl binary: %s", shimStdout)
	}
}
