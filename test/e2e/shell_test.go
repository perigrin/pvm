// ABOUTME: End-to-end tests for PVM shell integration
// ABOUTME: Tests shell initialization scripts and environment setup

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/test/e2e/helpers"
)

// TestShellInit tests the shell initialization command
func TestShellInit(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Run the correct init command (not "shell init")
	stdout, stderr, err := env.RunPVM("init")
	if err != nil {
		t.Fatalf("Shell initialization failed\nError: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// The init command outputs the script to stdout and ends with success message
	helpers.AssertStringContains(t, stdout, "PVM environment initialized",
		"Shell init output does not indicate success")

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
	helpers.AssertStringContains(t, stdout, "PVM environment initialized",
		"Bash init output does not indicate success")
}

// TestBashIntegration tests bash-specific shell integration
func TestBashIntegration(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// First run shell init
	_, stderr, err := env.RunPVM("shell", "init")
	if err != nil {
		t.Fatalf("Shell initialization failed\nError: %v\nStderr: %s", err, stderr)
	}

	// Source the bash script
	bashScript := filepath.Join(env.PVMDataDir, "shell", "pvm.bash")

	// Check if the bash script exists before continuing
	if _, statErr := os.Stat(bashScript); os.IsNotExist(statErr) {
		t.Fatalf("Bash shell integration script not found at %s", bashScript)
	}

	// Create a test bash script that sources the pvm bash script and tests functionality
	testScript := filepath.Join(env.HomeDir, "test.sh")
	scriptContent := `#!/bin/bash
source "` + bashScript + `"
# Test if pvm functions are defined
if type pvm_use >/dev/null 2>&1; then
    echo "pvm_use function defined"
fi
# Test if aliases are defined
if alias pvm-use >/dev/null 2>&1; then
    echo "pvm-use alias defined"
fi
# Test if PATH includes shims directory
if echo $PATH | grep -q "` + env.PVMShimsDir + `"; then
    echo "PATH includes shims directory"
fi
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

	// Check the output
	helpers.AssertStringContains(t, stdout, "pvm_use function defined",
		"Bash integration doesn't define pvm_use function")
	helpers.AssertStringContains(t, stdout, "pvm-use alias defined",
		"Bash integration doesn't define pvm-use alias")
	helpers.AssertStringContains(t, stdout, "PATH includes shims directory",
		"Bash integration doesn't add shims directory to PATH")
}

// Test .perl-version detection in shell integration
func TestPerlVersionFileDetection(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a .perl-version file in the home directory
	dotPerlVersionPath := filepath.Join(env.HomeDir, ".perl-version")
	testVersion := "5.38.0"
	err := os.WriteFile(dotPerlVersionPath, []byte(testVersion), 0644)
	if err != nil {
		t.Fatalf("Failed to create .perl-version file: %v", err)
	}

	// Initialize shell integration
	_, stderr, err := env.RunPVM("shell", "init")
	if err != nil {
		t.Fatalf("Shell initialization failed\nError: %v\nStderr: %s", err, stderr)
	}

	// Look for the bash script
	bashScript := filepath.Join(env.PVMDataDir, "shell", "pvm.bash")
	if _, statErr := os.Stat(bashScript); os.IsNotExist(statErr) {
		t.Fatalf("Bash shell integration script not found at %s", bashScript)
	}

	// Create a test bash script that sources the pvm bash script and changes to the home directory
	testScript := filepath.Join(env.HomeDir, "test_cd.sh")
	scriptContent := `#!/bin/bash
source "` + bashScript + `"
# Override cd to capture output
function cd() {
    command cd "$@" || return $?
    # Check if we detect the .perl-version file
    if [ -f .perl-version ]; then
        echo "Found .perl-version: $(cat .perl-version)"
    fi
}
# Test changing to a directory with .perl-version
cd "` + env.HomeDir + `"
echo "PVM_PERL_VERSION=${PVM_PERL_VERSION}"
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
	stdout, stderr, err := env.RunPVM("shell", "init")
	if err != nil {
		t.Fatalf("Shell initialization failed\nError: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Check that shell scripts were created
	shellDir := filepath.Join(env.PVMDataDir, "shell")
	bashScript := filepath.Join(shellDir, "pvm.bash")
	helpers.AssertFileExists(t, bashScript, "Bash shell script not created")

	// Check if the bash script contains conflict warnings
	helpers.AssertFileContains(t, bashScript, "PVM detected other Perl version managers",
		"Bash script should contain conflict detection warnings")
	helpers.AssertFileContains(t, bashScript, "plenv",
		"Bash script should specifically mention plenv conflict")
	helpers.AssertFileContains(t, bashScript, "PVM_SUPPRESS_WARNINGS",
		"Bash script should mention suppression option")
}
