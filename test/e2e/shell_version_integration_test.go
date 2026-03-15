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
	helpers.SkipIfNotUnix(t)
	// Use binary Perl for reliable testing
	helpers.SetupTestPerlEnvironment(t, helpers.DefaultTestPerlVersion)

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

	// Generate shell integration script (no longer creates files, outputs to stdout)
	// Generate shell integration script using the correct command
	initScript, stderr, err := env.RunPVM("init", "bash")
	if err != nil {
		t.Fatalf("Shell re-initialization failed: %v\nStderr: %s", err, stderr)
	}

	// Write the generated script to a file for testing
	bashScript := filepath.Join(env.HomeDir, "pvm_shell_integration.sh")
	err = os.WriteFile(bashScript, []byte(initScript), 0755)
	if err != nil {
		t.Fatalf("Failed to write shell integration script: %v", err)
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

	// Test 4: PVM commands with shell integration (bypass shell function)
	t.Logf("=== DEBUG: Testing PVM commands with shell integration ===")
	integrationScript := filepath.Join(env.HomeDir, "test_integration.sh")
	integrationContent := `#!/bin/bash
set -e
export PVM_SKIP_NETWORK_CALLS=1
export PVM_SUPPRESS_WARNINGS=1
echo "Testing PVM commands with shell integration"
source "` + bashScript + `"
echo "Shell integration loaded"
echo "Running pvm current directly via binary"
# Use the binary directly instead of shell function to avoid recursion
"$(_pvm_executable)" current
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

	// Test 5: Test shell function vs binary directly to isolate the issue
	t.Logf("=== DEBUG: Testing shell function vs binary directly ===")
	directScript := filepath.Join(env.HomeDir, "test_direct.sh")
	directContent := `#!/bin/bash
set -e
export PVM_SKIP_NETWORK_CALLS=1
export PVM_SUPPRESS_WARNINGS=1
echo "Testing shell function vs binary directly"
source "` + bashScript + `"
echo "Shell integration loaded"

echo "Test 1: Running pvm current via shell function"
timeout 30 bash -c 'pvm current' || echo "Shell function timed out"

echo "Test 2: Running pvm current directly via binary"
timeout 30 bash -c '"$(_pvm_executable)" current' || echo "Binary call timed out"

echo "Test completed"
exit 0
`
	err = os.WriteFile(directScript, []byte(directContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create direct test script: %v", err)
	}

	stdout, stderr, err = env.RunCommand("bash", directScript)
	if err != nil {
		t.Fatalf("Direct test failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	t.Logf("=== DEBUG: Direct test passed ===")

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
	helpers.AssertStringContains(t, allOutput, "set by PVM_PERL_VERSION",
		"Version should be set by PVM_PERL_VERSION environment variable, not .perl-version file")
}

// TestShellUseWithoutIntegration tests the expected behavior without shell integration
func TestShellUseWithoutIntegration(t *testing.T) {
	// Use binary Perl for reliable testing
	helpers.SetupTestPerlEnvironment(t, helpers.DefaultTestPerlVersion)

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
	stdout, stderr, _ = env.RunPVM("perl", "use", systemVersion)

	// This should NOT fail, but should show a warning about shell integration
	combinedOutput := stdout + stderr
	helpers.AssertStringContains(t, combinedOutput, "shell integration",
		"pvm use without shell integration should show shell integration warning")
}

// TestEnvironmentVariableVersionResolution tests that PVM_PERL_VERSION is properly resolved
func TestEnvironmentVariableVersionResolution(t *testing.T) {
	// Use binary Perl for reliable testing
	helpers.SetupTestPerlEnvironment(t, helpers.DefaultTestPerlVersion)

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
	helpers.AssertStringContains(t, currentOutput, "set by PVM_PERL_VERSION",
		"pvm current should indicate version source as PVM_PERL_VERSION environment variable")
}

// TestLibrarySpecificUse tests the new library-specific syntax for pvm use
func TestLibrarySpecificUse(t *testing.T) {
	helpers.ForceBashDetection(t)
	// Use binary Perl for reliable testing
	helpers.SetupTestPerlEnvironment(t, helpers.DefaultTestPerlVersion)

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl first
	stdout, stderr, err := env.RunPVM("import-system")
	if err != nil {
		t.Fatalf("System Perl import failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Get system Perl version
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
		t.Fatalf("Could not find system Perl version")
	}

	// Create a test library environment
	testLibraryName := "testlib"
	stdout, stderr, err = env.RunPVM("pvx", "--name", testLibraryName, "--isolation", "local", "-e", "print 'test'")
	if err != nil {
		t.Fatalf("Failed to create test library environment: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Test 1: Test sh-use with library syntax
	t.Run("ShUseWithLibrary", func(t *testing.T) {
		versionLibrarySpec := systemVersion + "@" + testLibraryName
		stdout, stderr, err := env.RunPVM("sh-use", versionLibrarySpec)
		if err != nil {
			t.Fatalf("sh-use with library failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		shellOutput := stdout + stderr
		helpers.AssertStringContains(t, shellOutput, "export PVM_PERL_VERSION='"+systemVersion+"'",
			"sh-use should set PVM_PERL_VERSION")
		helpers.AssertStringContains(t, shellOutput, "export PVM_PERL_LIBRARY='"+testLibraryName+"'",
			"sh-use should set PVM_PERL_LIBRARY")
		helpers.AssertStringContains(t, shellOutput, "export PVM_PERL_VERSION_FULL='"+versionLibrarySpec+"'",
			"sh-use should set PVM_PERL_VERSION_FULL")
		helpers.AssertStringContains(t, shellOutput, "Using Perl "+versionLibrarySpec,
			"sh-use should show version@library in output")
	})

	// Test 2: Test sh-use with non-existent library
	t.Run("ShUseWithNonExistentLibrary", func(t *testing.T) {
		versionLibrarySpec := systemVersion + "@nonexistent"
		stdout, stderr, err := env.RunPVM("sh-use", versionLibrarySpec)
		if err == nil {
			t.Fatalf("sh-use with non-existent library should fail")
		}

		errorOutput := stdout + stderr
		helpers.AssertStringContains(t, errorOutput, "does not exist",
			"sh-use should error for non-existent library")
		helpers.AssertStringContains(t, errorOutput, "pvm pvx --name nonexistent",
			"sh-use should suggest how to create the library")
	})

	// Test 3: Test system@library syntax
	t.Run("SystemWithLibrary", func(t *testing.T) {
		versionLibrarySpec := "system@" + testLibraryName
		stdout, stderr, err := env.RunPVM("sh-use", versionLibrarySpec)
		if err != nil {
			t.Fatalf("sh-use system@library failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		shellOutput := stdout + stderr
		helpers.AssertStringContains(t, shellOutput, "unset PVM_PERL_VERSION",
			"sh-use system@library should unset PVM_PERL_VERSION")
		helpers.AssertStringContains(t, shellOutput, "unset PVM_PERL_LIBRARY",
			"sh-use system@library should unset PVM_PERL_LIBRARY")
		helpers.AssertStringContains(t, shellOutput, "Using system Perl with library '"+testLibraryName+"'",
			"sh-use should show system with library message")
	})

	// Test 4: Test clearing library (version without @library)
	t.Run("ClearLibrary", func(t *testing.T) {
		stdout, stderr, err := env.RunPVM("sh-use", systemVersion)
		if err != nil {
			t.Fatalf("sh-use without library failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		shellOutput := stdout + stderr
		helpers.AssertStringContains(t, shellOutput, "export PVM_PERL_VERSION='"+systemVersion+"'",
			"sh-use should set PVM_PERL_VERSION")
		helpers.AssertStringContains(t, shellOutput, "unset PVM_PERL_LIBRARY",
			"sh-use without library should unset PVM_PERL_LIBRARY")
		helpers.AssertStringContains(t, shellOutput, "export PVM_PERL_VERSION_FULL='"+systemVersion+"'",
			"sh-use should set PVM_PERL_VERSION_FULL to just version")
	})
}

// TestLibraryEnvironmentResolution tests that library environments are properly resolved
func TestLibraryEnvironmentResolution(t *testing.T) {
	// Use binary Perl for reliable testing
	helpers.SetupTestPerlEnvironment(t, helpers.DefaultTestPerlVersion)

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl first
	stdout, stderr, err := env.RunPVM("import-system")
	if err != nil {
		t.Fatalf("System Perl import failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Get system Perl version
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
		t.Fatalf("Could not find system Perl version")
	}

	// Create a test library environment
	testLibraryName := "resolvertest"
	stdout, stderr, err = env.RunPVM("pvx", "--name", testLibraryName, "--isolation", "local", "-e", "print 'test'")
	if err != nil {
		t.Fatalf("Failed to create test library environment: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Set environment variables as if pvm use was called
	os.Setenv("PVM_PERL_VERSION", systemVersion)
	os.Setenv("PVM_PERL_LIBRARY", testLibraryName)

	// Test that current version resolution includes library information
	stdout, stderr, err = env.RunPVM("current", "-v")
	if err != nil {
		t.Fatalf("Current command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	currentOutput := stdout + stderr
	helpers.AssertStringContains(t, currentOutput, systemVersion,
		"pvm current should show the version")
	helpers.AssertStringContains(t, currentOutput, "PVM_PERL_VERSION",
		"pvm current should indicate version source as PVM_PERL_VERSION")
}

// TestLibrarySpecificUseSecurity tests security aspects of library-specific use syntax
func TestLibrarySpecificUseSecurity(t *testing.T) {
	// Use binary Perl for reliable testing
	helpers.SetupTestPerlEnvironment(t, helpers.DefaultTestPerlVersion)

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Get system version for testing
	stdout, stderr, err := env.RunPVM("import-system")
	if err != nil {
		t.Fatalf("System Perl import failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

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
		t.Fatalf("Could not find system Perl version")
	}

	// Test security validation for malicious library names
	securityTestCases := []struct {
		name          string
		versionSpec   string
		shouldFail    bool
		errorContains string
	}{
		{
			name:          "PathTraversalAttack",
			versionSpec:   systemVersion + "@../../../etc",
			shouldFail:    true,
			errorContains: "invalid path characters",
		},
		{
			name:          "ShellInjectionAttack",
			versionSpec:   systemVersion + "@lib'; rm -rf /; echo 'hack",
			shouldFail:    true,
			errorContains: "invalid path characters",
		},
		{
			name:          "MultipleAtSymbols",
			versionSpec:   systemVersion + "@lib1@lib2",
			shouldFail:    true,
			errorContains: "only one @ symbol allowed",
		},
		{
			name:          "EmptyLibraryAfterAt",
			versionSpec:   systemVersion + "@",
			shouldFail:    true,
			errorContains: "library name cannot be empty after @",
		},
		{
			name:          "BacktickInjection",
			versionSpec:   systemVersion + "@lib`id`",
			shouldFail:    true,
			errorContains: "alphanumeric characters",
		},
		{
			name:          "DollarSignInjection",
			versionSpec:   systemVersion + "@lib$(whoami)",
			shouldFail:    true,
			errorContains: "alphanumeric characters",
		},
	}

	for _, tc := range securityTestCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := env.RunPVM("sh-use", tc.versionSpec)

			if tc.shouldFail && err == nil {
				t.Fatalf("Expected security validation to fail for %s", tc.versionSpec)
			}

			if tc.shouldFail && tc.errorContains != "" {
				errorOutput := stdout + stderr
				if !strings.Contains(errorOutput, tc.errorContains) {
					t.Errorf("Error should contain '%s' but got: %s", tc.errorContains, errorOutput)
				}
			}
		})
	}
}
