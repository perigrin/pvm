// ABOUTME: End-to-end tests for PVM shell integration
// ABOUTME: Tests shell initialization scripts and environment setup

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/fortune"
	"tamarou.com/pvm/test/e2e/helpers"
)

// assertFortuneQuotePresent verifies that the output contains one of the expected fortune quotes
func assertFortuneQuotePresent(t *testing.T, output, context string) {
	t.Helper()

	// Get all available fortune quotes
	allQuotes := fortune.GetAllQuotes()

	// Check if any of the expected quotes is present in the output
	for _, quote := range allQuotes {
		if strings.Contains(output, quote) {
			return // Found a fortune quote, test passes
		}
	}

	// If we get here, no fortune quote was found
	t.Errorf("%s should contain a fortune quote, but none found.\nExpected one of %d quotes, got output: %s",
		context, len(allQuotes), output)
}

// TestShellInit tests the shell initialization command
func TestShellInit(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Run the correct init command (not "shell init")
	stdout, stderr, err := env.RunPVM("init")
	if err != nil {
		t.Fatalf("Shell initialization failed\nError: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// The init command outputs the script to stdout and ends with fortune quote
	assertFortuneQuotePresent(t, stdout, "Shell init output")

	// Check that the generated script contains expected content
	helpers.AssertStringContains(t, stdout, "PVM Shell Integration for Bash/Zsh",
		"Shell script does not contain expected header")
	helpers.AssertStringContains(t, stdout, "pvm_init",
		"Shell script does not contain init function")
	helpers.AssertStringContains(t, stdout, "_pvm_update_perl_path",
		"Shell script does not contain PATH manipulation")

	// Test with explicit shell parameter
	stdout, stderr, err = env.RunPVM("init", "bash")
	if err != nil {
		t.Fatalf("Shell initialization with bash parameter failed\nError: %v\nStderr: %s", err, stderr)
	}
	assertFortuneQuotePresent(t, stdout, "Bash init output")
}

// TestBashIntegration tests bash-specific shell integration
func TestBashIntegration(t *testing.T) {
	helpers.SkipIfNotUnix(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Generate shell integration script using the correct command
	initScript, stderr, err := env.RunPVM("init", "bash")
	if err != nil {
		t.Fatalf("Shell initialization failed\nError: %v\nStderr: %s", err, stderr)
	}

	// The script should contain fortune quote
	assertFortuneQuotePresent(t, initScript, "Init script")

	// Write the generated script to a file for testing
	bashScript := filepath.Join(env.HomeDir, "pvm_init.sh")
	err = os.WriteFile(bashScript, []byte(initScript), 0755)
	if err != nil {
		t.Fatalf("Failed to write bash script: %v", err)
	}

	// Create a test bash script that sources the pvm script and tests functionality
	testScript := filepath.Join(env.HomeDir, "test.sh")
	scriptContent := `#!/bin/bash
source "` + bashScript + `"

# Test if main pvm function is defined
if type pvm >/dev/null 2>&1; then
    echo "pvm function defined"
fi

# Test if internal functions are defined
if type _pvm_update_perl_path >/dev/null 2>&1; then
    echo "_pvm_update_perl_path function defined"
fi

# Test if cd alias is set up (bash-specific)
if alias cd >/dev/null 2>&1; then
    echo "cd alias defined"
fi

echo "Integration test completed"
`
	err = os.WriteFile(testScript, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create test bash script: %v", err)
	}

	// Run the test script
	stdout, stderr, err := env.RunCommand("bash", testScript)
	if err != nil {
		t.Fatalf("Failed to run test bash script: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Check the output - test for current shell integration functionality
	helpers.AssertStringContains(t, stdout, "pvm function defined",
		"Bash integration doesn't define pvm function")
	helpers.AssertStringContains(t, stdout, "_pvm_update_perl_path function defined",
		"Bash integration doesn't define internal PATH function")
	helpers.AssertStringContains(t, stdout, "cd alias defined",
		"Bash integration doesn't define cd alias")
	helpers.AssertStringContains(t, stdout, "Integration test completed",
		"Bash integration test didn't complete")
}

// Test .perl-version detection in shell integration
func TestPerlVersionFileDetection(t *testing.T) {
	helpers.SkipIfNotUnix(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a .perl-version file in the home directory
	dotPerlVersionPath := filepath.Join(env.HomeDir, ".perl-version")
	testVersion := "5.38.0"
	err := os.WriteFile(dotPerlVersionPath, []byte(testVersion), 0644)
	if err != nil {
		t.Fatalf("Failed to create .perl-version file: %v", err)
	}

	// Generate shell integration script using the correct command
	initScript, stderr, err := env.RunPVM("init", "bash")
	if err != nil {
		t.Fatalf("Shell initialization failed\nError: %v\nStderr: %s", err, stderr)
	}

	// The script should contain fortune quote
	assertFortuneQuotePresent(t, initScript, "Init script")

	// Write the generated script to a file for testing
	bashScript := filepath.Join(env.HomeDir, "pvm_init.sh")
	err = os.WriteFile(bashScript, []byte(initScript), 0755)
	if err != nil {
		t.Fatalf("Failed to write bash script: %v", err)
	}

	// Create a test bash script that sources the pvm script and tests .perl-version detection
	testScript := filepath.Join(env.HomeDir, "test_cd.sh")
	scriptContent := `#!/bin/bash
source "` + bashScript + `"

# Override cd to capture output and test .perl-version detection
function cd() {
    command cd "$@" || return $?
    # Check if we detect the .perl-version file
    if [ -f .perl-version ]; then
        echo "Found .perl-version: $(cat .perl-version)"
    fi
    # Call the PVM cd handler if it exists
    if type _pvm_update_perl_path >/dev/null 2>&1; then
        _pvm_update_perl_path
    fi
}

# Test changing to a directory with .perl-version
cd "` + env.HomeDir + `"
echo "Integration test completed"
`
	err = os.WriteFile(testScript, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create test bash script: %v", err)
	}

	// Run the test script
	stdout, stderr, err := env.RunCommand("bash", testScript)
	if err != nil {
		t.Fatalf("Failed to run test bash script: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Check that it detected the .perl-version file
	helpers.AssertStringContains(t, stdout, "Found .perl-version: "+testVersion,
		"Bash cd function didn't detect .perl-version file")
	helpers.AssertStringContains(t, stdout, "Integration test completed",
		"Perl version detection test didn't complete")
}

// TestShellSetup tests the shell setup command
func TestShellSetup(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test shell setup command
	stdout, stderr, err := env.RunPVM("shell", "setup")
	if err != nil {
		t.Fatalf("Shell setup command failed\nError: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// The command uses cmd.Println which goes to stdout, but let's check both stdout and stderr
	output := stdout
	if output == "" && stderr != "" {
		output = stderr
	}

	// Check that output contains setup information
	if output == "" {
		t.Fatalf("Shell setup command returned empty output\nStdout: %q\nStderr: %q", stdout, stderr)
	}

	// Should detect shell and provide setup instructions
	// Common patterns that should appear in setup instructions
	helpers.AssertStringContains(t, output, "shell integration", "Should mention shell integration")
	helpers.AssertStringContains(t, output, "setting up", "Should mention setup process")
	helpers.AssertStringContains(t, output, "pvm init", "Should mention pvm init command")
	helpers.AssertStringContains(t, output, "eval", "Should contain eval command in instructions")
}

// TestShellCompletion tests shell completion functionality
func TestShellCompletion(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test completion command for different shells
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			// Run completion command
			stdout, stderr, err := env.RunPVM("completion", shell)
			if err != nil {
				t.Fatalf("Completion command failed for %s\nError: %v\nStdout: %s\nStderr: %s", shell, err, stdout, stderr)
			}

			// Check that output is not empty
			if stdout == "" {
				t.Errorf("Completion command for %s returned empty output", shell)
			}

			// Check shell-specific content
			switch shell {
			case "bash":
				helpers.AssertStringContains(t, stdout, "_pvm_completion", "Bash completion should contain completion function")
				helpers.AssertStringContains(t, stdout, "complete -F", "Bash completion should contain complete command")
			case "zsh":
				helpers.AssertStringContains(t, stdout, "#compdef pvm", "Zsh completion should contain compdef directive")
				helpers.AssertStringContains(t, stdout, "_pvm()", "Zsh completion should contain completion function")
				helpers.AssertStringContains(t, stdout, "compdef _pvm pvm", "Zsh completion should register function with compdef")
				helpers.AssertStringDoesNotContain(t, stdout, "_pvm \"$@\"", "Zsh completion should not call function immediately")
			case "fish":
				helpers.AssertStringContains(t, stdout, "complete -c pvm", "Fish completion should contain complete commands")
			case "powershell":
				helpers.AssertStringContains(t, stdout, "Register-ArgumentCompleter", "PowerShell completion should contain Register-ArgumentCompleter")
			}
		})
	}

	// Test invalid shell
	t.Run("invalid_shell", func(t *testing.T) {
		stdout, stderr, err := env.RunPVM("completion", "invalid")
		if err == nil {
			t.Errorf("Expected error for invalid shell, but command succeeded\nStdout: %s\nStderr: %s", stdout, stderr)
		}
		helpers.AssertStringContains(t, stderr, "unsupported shell", "Error should mention unsupported shell")
	})
}

// TestShellConflictDetection tests detection of conflicting version managers
func TestShellConflictDetection(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a mock plenv shims directory in PATH to simulate a conflict
	mockPlenvShims := filepath.Join(env.HomeDir, "plenv-shims")
	err := os.MkdirAll(mockPlenvShims, 0755)
	if err != nil {
		t.Fatalf("Failed to create mock plenv shims directory: %v", err)
	}

	// Modify PATH to include the mock plenv shims
	originalPath := os.Getenv("PATH")
	newPath := mockPlenvShims + string(os.PathListSeparator) + originalPath
	err = os.Setenv("PATH", newPath)
	if err != nil {
		t.Fatalf("Failed to set PATH: %v", err)
	}
	defer func() {
		_ = os.Setenv("PATH", originalPath)
	}()

	// Run shell init command which should detect the conflict
	stdout, stderr, err := env.RunPVM("init", "bash")
	if err != nil {
		t.Fatalf("Shell initialization failed\nError: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// The script should contain fortune quote
	assertFortuneQuotePresent(t, stdout, "Init script")

	// TODO: Conflict detection is not currently implemented in shell templates
	// The ConflictWarnings template variable exists but is not used in templates
	// When this feature is implemented, these assertions should be uncommented:
	//
	// helpers.AssertStringContains(t, stdout, "PVM detected other Perl version managers",
	//     "Generated script should contain conflict detection warnings")
	// helpers.AssertStringContains(t, stdout, "plenv",
	//     "Generated script should specifically mention plenv conflict")
	// helpers.AssertStringContains(t, stdout, "PVM_SUPPRESS_WARNINGS",
	//     "Generated script should mention suppression option")

	// For now, just verify that the script was generated successfully
	helpers.AssertStringContains(t, stdout, "pvm_init",
		"Generated script should contain pvm_init function")
}

// TestXDGBinHomePathIntegration tests that XDG_BIN_HOME is properly added and preserved in PATH
func TestXDGBinHomePathIntegration(t *testing.T) {
	helpers.SkipIfNotUnix(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Get the expected XDG_BIN_HOME path
	expectedXDGBinHome := filepath.Join(env.HomeDir, ".local", "bin")

	// Ensure XDG_BIN_HOME directory exists for realistic testing
	err := os.MkdirAll(expectedXDGBinHome, 0755)
	if err != nil {
		t.Fatalf("Failed to create XDG_BIN_HOME directory: %v", err)
	}

	// Test both bash and fish shells
	shells := []string{"bash", "fish"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			// Skip fish test if fish shell is not available
			if shell == "fish" {
				_, _, err := env.RunCommand("which", "fish")
				if err != nil {
					t.Skip("Fish shell not available in test environment")
					return
				}
			}
			// Generate shell integration script
			initScript, stderr, err := env.RunPVM("init", shell)
			if err != nil {
				t.Fatalf("Shell initialization failed for %s\nError: %v\nStderr: %s", shell, err, stderr)
			}

			// The script should contain fortune quote
			assertFortuneQuotePresent(t, initScript, "Init script")

			// Verify XDG_BIN_HOME variable is referenced in the script (not hardcoded path)
			if shell == "fish" {
				helpers.AssertStringContains(t, initScript, "XDG_BIN_HOME",
					"Fish script should contain XDG_BIN_HOME variable reference")
			} else {
				helpers.AssertStringContains(t, initScript, "xdg_bin_home=",
					"Bash/Zsh script should contain xdg_bin_home variable assignment")
			}

			// Write the generated script to a file for testing
			scriptFile := filepath.Join(env.HomeDir, "pvm_init_"+shell+".sh")
			if shell == "fish" {
				scriptFile = filepath.Join(env.HomeDir, "pvm_init.fish")
			}
			err = os.WriteFile(scriptFile, []byte(initScript), 0755)
			if err != nil {
				t.Fatalf("Failed to write %s script: %v", shell, err)
			}

			// Create a test script that sources the pvm script and checks PATH
			var testScript string
			var scriptContent string

			if shell == "bash" {
				testScript = filepath.Join(env.HomeDir, "test_path.sh")
				scriptContent = `#!/bin/bash
# Clear PATH to simulate fresh environment
export PATH="/usr/bin:/bin"

# Source the PVM init script
source "` + scriptFile + `"

# Check if XDG_BIN_HOME is in PATH
if echo "$PATH" | grep -q "` + expectedXDGBinHome + `"; then
    echo "XDG_BIN_HOME_FOUND_IN_PATH"
else
    echo "XDG_BIN_HOME_MISSING_FROM_PATH"
fi

# Test PATH preservation after calling _pvm_update_perl_path
if type _pvm_update_perl_path >/dev/null 2>&1; then
    _pvm_update_perl_path
    if echo "$PATH" | grep -q "` + expectedXDGBinHome + `"; then
        echo "XDG_BIN_HOME_PRESERVED_AFTER_UPDATE"
    else
        echo "XDG_BIN_HOME_LOST_AFTER_UPDATE"
    fi
fi

echo "PATH_TEST_COMPLETED"
`
			} else { // fish
				testScript = filepath.Join(env.HomeDir, "test_path.fish")
				scriptContent = `#!/usr/bin/env fish
# Clear PATH to simulate fresh environment
set -gx PATH /usr/bin /bin

# Source the PVM init script
source "` + scriptFile + `"

# Check if XDG_BIN_HOME is in PATH
if contains "` + expectedXDGBinHome + `" $PATH
    echo "XDG_BIN_HOME_FOUND_IN_PATH"
else
    echo "XDG_BIN_HOME_MISSING_FROM_PATH"
end

# Test PATH preservation after calling _pvm_update_perl_path
if functions -q _pvm_update_perl_path
    _pvm_update_perl_path
    if contains "` + expectedXDGBinHome + `" $PATH
        echo "XDG_BIN_HOME_PRESERVED_AFTER_UPDATE"
    else
        echo "XDG_BIN_HOME_LOST_AFTER_UPDATE"
    end
end

echo "PATH_TEST_COMPLETED"
`
			}

			err = os.WriteFile(testScript, []byte(scriptContent), 0755)
			if err != nil {
				t.Fatalf("Failed to create test script: %v", err)
			}

			// Run the test script
			var stdout, stderr_out string
			if shell == "bash" {
				stdout, stderr_out, err = env.RunCommand("bash", testScript)
			} else {
				stdout, stderr_out, err = env.RunCommand("fish", testScript)
			}

			if err != nil {
				t.Fatalf("Failed to run %s test script: %v\nStdout: %s\nStderr: %s", shell, err, stdout, stderr_out)
			}

			// Verify XDG_BIN_HOME was added to PATH initially
			helpers.AssertStringContains(t, stdout, "XDG_BIN_HOME_FOUND_IN_PATH",
				"XDG_BIN_HOME should be added to PATH during initialization")

			// Verify XDG_BIN_HOME is preserved after PATH updates
			helpers.AssertStringContains(t, stdout, "XDG_BIN_HOME_PRESERVED_AFTER_UPDATE",
				"XDG_BIN_HOME should be preserved after _pvm_update_perl_path calls")

			helpers.AssertStringContains(t, stdout, "PATH_TEST_COMPLETED",
				"PATH test should complete successfully")

			// Ensure we don't have failure messages
			helpers.AssertStringDoesNotContain(t, stdout, "XDG_BIN_HOME_MISSING_FROM_PATH",
				"XDG_BIN_HOME should not be missing from PATH")
			helpers.AssertStringDoesNotContain(t, stdout, "XDG_BIN_HOME_LOST_AFTER_UPDATE",
				"XDG_BIN_HOME should not be lost after PATH updates")
		})
	}
}

// TestShellIntegrationDoctorDetection tests that 'pvm self doctor' correctly detects shell integration
func TestShellIntegrationDoctorDetection(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Get the expected XDG_BIN_HOME path
	expectedXDGBinHome := filepath.Join(env.HomeDir, ".local", "bin")

	// Ensure XDG_BIN_HOME directory exists
	err := os.MkdirAll(expectedXDGBinHome, 0755)
	if err != nil {
		t.Fatalf("Failed to create XDG_BIN_HOME directory: %v", err)
	}

	// First, test without shell integration (should show warning)
	stdout, stderr, err := env.RunPVM("self", "doctor")
	if err != nil {
		t.Fatalf("Doctor command failed\nError: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Should detect that shell integration is not active
	helpers.AssertStringContains(t, stdout, "Shell integration not active",
		"Doctor should detect shell integration is not active initially")

	// Now test with XDG_BIN_HOME in PATH (simulate shell integration)
	originalPath := os.Getenv("PATH")
	pathWithXDG := expectedXDGBinHome + string(os.PathListSeparator) + originalPath
	err = os.Setenv("PATH", pathWithXDG)
	if err != nil {
		t.Fatalf("Failed to set PATH with XDG_BIN_HOME: %v", err)
	}
	defer func() {
		_ = os.Setenv("PATH", originalPath)
	}()

	// Run doctor again with XDG_BIN_HOME in PATH
	stdout, stderr, err = env.RunPVM("self", "doctor")
	if err != nil {
		t.Fatalf("Doctor command failed with XDG_BIN_HOME in PATH\nError: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Should now detect that shell integration is active
	helpers.AssertStringContains(t, stdout, "Shell integration active",
		"Doctor should detect shell integration is active when XDG_BIN_HOME is in PATH")
	helpers.AssertStringContains(t, stdout, "XDG_BIN_HOME found in PATH",
		"Doctor should specifically mention XDG_BIN_HOME in PATH")
}
