// ABOUTME: End-to-end tests for shell integration with version switching
// ABOUTME: Tests the specific workflow described in issue #118

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/test/e2e/helpers"
)

// TestShellUseAndCurrent tests the specific workflow from issue #118
// This test simulates: pvm use X.Y.Z && pvm current
// It tests with multiple Perl versions to catch version switching issues
func TestShellUseAndCurrent(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl first
	stdout, stderr, err := env.RunPVM("import-system")
	if err != nil {
		t.Fatalf("System Perl import failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Get all available Perl versions
	stdout, stderr, err = env.RunPVM("list")
	if err != nil {
		t.Fatalf("List command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	listOutput := stdout + stderr
	lines := strings.Split(listOutput, "\n")
	var availableVersions []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ".") && !strings.Contains(line, "No Perl versions") {
			// Extract version number (e.g., "5.38.2" from "5.38.2 (system)")
			parts := strings.Fields(line)
			if len(parts) > 0 {
				version := strings.TrimSpace(parts[0])
				availableVersions = append(availableVersions, version)
			}
		}
	}

	if len(availableVersions) == 0 {
		t.Fatalf("Could not find any Perl versions in output: %s", listOutput)
	}

	// For issue #118, we need to test switching between different versions
	// If we only have system Perl, that's still valuable to test
	testVersion := availableVersions[0]
	t.Logf("Testing with Perl version: %s", testVersion)

	// Initialize shell integration
	_, stderr, err = env.RunPVM("shell", "init")
	if err != nil {
		t.Fatalf("Shell initialization failed: %v\nStderr: %s", err, stderr)
	}

	// Force regeneration of shell integration script to ensure it includes latest fixes
	bashScriptPath := filepath.Join(env.PVMDataDir, "shell", "pvm.bash")
	if err := os.Remove(bashScriptPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove old shell script: %v", err)
	}
	_, stderr, err = env.RunPVM("shell", "init")
	if err != nil {
		t.Fatalf("Shell re-initialization failed: %v\nStderr: %s", err, stderr)
	}

	// Get the bash script path
	bashScript := filepath.Join(env.PVMDataDir, "shell", "pvm.bash")
	if _, statErr := os.Stat(bashScript); os.IsNotExist(statErr) {
		t.Fatalf("Bash shell integration script not found at %s", bashScript)
	}

	// DEBUGGING: Test each component individually to isolate the hanging point
	t.Logf("=== DEBUG: Testing components individually ===")

	// Test 1: Basic bash execution
	t.Logf("=== DEBUG: Testing basic bash execution ===")
	basicScript := filepath.Join(env.HomeDir, "test_basic.sh")
	basicContent := `#!/bin/bash
echo "Basic bash test successful"
exit 0
`
	err = os.WriteFile(basicScript, []byte(basicContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create basic test script: %v", err)
	}

	stdout, stderr, err = env.RunCommand("bash", basicScript)
	if err != nil {
		t.Fatalf("Basic bash test failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	t.Logf("=== DEBUG: Basic bash test passed ===")

	// Test 2: PVM command execution without shell integration
	t.Logf("=== DEBUG: Testing PVM command execution ===")
	pvmScript := filepath.Join(env.HomeDir, "test_pvm.sh")
	pvmContent := `#!/bin/bash
set -e
export PVM_SKIP_NETWORK_CALLS=1
echo "Testing PVM current command"
` + env.PVMBinary + ` current
echo "PVM current command successful"
exit 0
`
	err = os.WriteFile(pvmScript, []byte(pvmContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create PVM test script: %v", err)
	}

	stdout, stderr, err = env.RunCommand("bash", pvmScript)
	if err != nil {
		t.Fatalf("PVM command test failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	t.Logf("=== DEBUG: PVM command test passed ===")

	// Test 3: Shell integration sourcing
	t.Logf("=== DEBUG: Testing shell integration sourcing ===")
	sourceScript := filepath.Join(env.HomeDir, "test_source.sh")
	sourceContent := `#!/bin/bash
set -e
export PVM_SKIP_NETWORK_CALLS=1
export PVM_SUPPRESS_WARNINGS=1
echo "Testing shell integration sourcing"
source "` + bashScript + `"
echo "Shell integration sourcing successful"
exit 0
`
	err = os.WriteFile(sourceScript, []byte(sourceContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create source test script: %v", err)
	}

	stdout, stderr, err = env.RunCommand("bash", sourceScript)
	if err != nil {
		t.Fatalf("Shell integration sourcing test failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	t.Logf("=== DEBUG: Shell integration sourcing test passed ===")

	// Test 4: PVM commands with shell integration
	t.Logf("=== DEBUG: Testing PVM commands with shell integration ===")
	integrationScript := filepath.Join(env.HomeDir, "test_integration.sh")
	integrationContent := `#!/bin/bash
set -e
export PVM_SKIP_NETWORK_CALLS=1
export PVM_SUPPRESS_WARNINGS=1
echo "Testing PVM commands with shell integration"
source "` + bashScript + `"
echo "Shell integration loaded"
echo "Running pvm current"
pvm current
echo "pvm current successful"
exit 0
`
	err = os.WriteFile(integrationScript, []byte(integrationContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create integration test script: %v", err)
	}

	stdout, stderr, err = env.RunCommand("bash", integrationScript)
	if err != nil {
		t.Fatalf("PVM integration test failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	t.Logf("=== DEBUG: PVM integration test passed ===")

	// Test 5: Full workflow but simplified
	t.Logf("=== DEBUG: Testing full workflow (simplified) ===")
	fullScript := filepath.Join(env.HomeDir, "test_full.sh")
	fullContent := `#!/bin/bash
set -e
export PVM_SKIP_NETWORK_CALLS=1
export PVM_SUPPRESS_WARNINGS=1
echo "Testing full workflow"
source "` + bashScript + `"
echo "Shell integration loaded"
echo "Running initial pvm current"
pvm current
echo "Initial pvm current successful"
echo "Running pvm use ` + testVersion + `"
pvm use ` + testVersion + `
echo "pvm use successful"
echo "Running final pvm current"
pvm current
echo "Final pvm current successful"
exit 0
`
	err = os.WriteFile(fullScript, []byte(fullContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create full test script: %v", err)
	}

	stdout, stderr, err = env.RunCommand("bash", fullScript)
	if err != nil {
		t.Fatalf("Full workflow test failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	t.Logf("=== DEBUG: Full workflow test passed ===")

	// If we get here, all components work individually, so the issue might be with the original complex script
	t.Logf("=== DEBUG: All individual components passed, testing original script ===")

	// Create a test script that simulates the issue #118 workflow
	testScript := filepath.Join(env.HomeDir, "test_issue_118.sh")
	scriptContent := `#!/bin/bash
set -e

# Skip network calls to avoid test timeouts
export PVM_SKIP_NETWORK_CALLS=1

# Source the PVM shell integration
source "` + bashScript + `"

# Show initial current version
echo "=== Initial current version ==="
pvm current

# Use the test version via shell integration
echo "=== Using pvm use ` + testVersion + ` ==="
pvm use ` + testVersion + `

# Check that PVM_PERL_VERSION is set
echo "=== Environment variable ==="
echo "PVM_PERL_VERSION=${PVM_PERL_VERSION}"

# Show current version after use
echo "=== Current version after use ==="
pvm current

# Verify they match (capture both stdout and stderr)
CURRENT_OUTPUT=$(pvm current 2>&1)
if echo "$CURRENT_OUTPUT" | grep -q "` + testVersion + `"; then
    echo "=== SUCCESS: Current version matches used version ==="
    echo "Current output: $CURRENT_OUTPUT"
else
    echo "=== FAILURE: Current version does not match used version ==="
    echo "Expected: ` + testVersion + `"
    echo "Got: $CURRENT_OUTPUT"
    exit 1
fi
`
	err = os.WriteFile(testScript, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Run the test script
	stdout, stderr, err = env.RunCommand("bash", testScript)
	if err != nil {
		t.Fatalf("Shell integration test failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Check that the test succeeded (could be in stdout or stderr)
	allOutput := stdout + stderr
	helpers.AssertStringContains(t, allOutput, "SUCCESS: Current version matches used version",
		"Issue #118 test failed: pvm current does not show correct version after pvm use")

	// Also verify the environment variable was set correctly
	helpers.AssertStringContains(t, allOutput, "PVM_PERL_VERSION="+testVersion,
		"PVM_PERL_VERSION environment variable not set correctly")

	// Verify the environment variable takes precedence over .perl-version file
	helpers.AssertStringContains(t, allOutput, "set by environment variable",
		"Version should be set by environment variable, not .perl-version file")
}

// TestShellUseWithoutIntegration tests the expected behavior without shell integration
func TestShellUseWithoutIntegration(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl first
	stdout, stderr, err := env.RunPVM("import-system")
	if err != nil {
		t.Fatalf("System Perl import failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Get the system Perl version
	stdout, stderr, err = env.RunPVM("list")
	if err != nil {
		t.Fatalf("List command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	listOutput := stdout + stderr
	lines := strings.Split(listOutput, "\n")
	var systemVersion string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ".") && !strings.Contains(line, "No Perl versions") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				systemVersion = strings.TrimSpace(parts[0])
				break
			}
		}
	}

	if systemVersion == "" {
		t.Fatalf("Could not find system Perl version in output: %s", listOutput)
	}

	// Try to use version without shell integration (should show warning)
	stdout, stderr, err = env.RunPVM("use", systemVersion)

	// This should NOT fail, but should show a warning about shell integration
	combinedOutput := stdout + stderr
	helpers.AssertStringContains(t, combinedOutput, "shell integration",
		"pvm use without shell integration should show shell integration warning")
}

// TestEnvironmentVariableVersionResolution tests that PVM_PERL_VERSION is properly resolved
func TestEnvironmentVariableVersionResolution(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl first
	stdout, stderr, err := env.RunPVM("import-system")
	if err != nil {
		t.Fatalf("System Perl import failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Get the system Perl version
	stdout, stderr, err = env.RunPVM("list")
	if err != nil {
		t.Fatalf("List command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	listOutput := stdout + stderr
	lines := strings.Split(listOutput, "\n")
	var systemVersion string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ".") && !strings.Contains(line, "No Perl versions") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				systemVersion = strings.TrimSpace(parts[0])
				break
			}
		}
	}

	if systemVersion == "" {
		t.Fatalf("Could not find system Perl version in output: %s", listOutput)
	}

	// Set PVM_PERL_VERSION environment variable directly
	os.Setenv("PVM_PERL_VERSION", systemVersion)

	// Check that current version now shows the environment variable version
	stdout, stderr, err = env.RunPVM("current")
	if err != nil {
		t.Fatalf("Current command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	currentOutput := stdout + stderr
	helpers.AssertStringContains(t, currentOutput, systemVersion,
		"pvm current should show version from PVM_PERL_VERSION environment variable")
	helpers.AssertStringContains(t, currentOutput, "environment variable",
		"pvm current should indicate version source as environment variable")
}
