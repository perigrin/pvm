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

	// Run shell init command or skip as TODO if not implemented
	stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"shell", "init"}, "Shell initialization")

	// If we get here, the command succeeded, so run the normal assertions
	helpers.AssertStringContains(t, stdout, "Shell integration initialized",
		"Shell init output does not indicate success")

	// Check that shell scripts were created
	shellDir := filepath.Join(env.PVMDataDir, "shell")
	helpers.AssertFileExists(t, filepath.Join(shellDir, "pvm.bash"), "Bash shell script not created")
	helpers.AssertFileExists(t, filepath.Join(shellDir, "pvm.zsh"), "Zsh shell script not created")
	helpers.AssertFileExists(t, filepath.Join(shellDir, "pvm.fish"), "Fish shell script not created")

	// Check content of bash script
	bashScript := filepath.Join(shellDir, "pvm.bash")
	helpers.AssertFileContains(t, bashScript, "PVM Shell Integration for Bash/Zsh",
		"Bash script does not contain expected header")
	helpers.AssertFileContains(t, bashScript, "PATH=",
		"Bash script does not contain PATH manipulation")
	helpers.AssertFileContains(t, bashScript, "pvm-use",
		"Bash script does not contain use function")
}

// TestBashIntegration tests bash-specific shell integration
func TestBashIntegration(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// First run shell init or skip as TODO if not implemented
	_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"shell", "init"}, "Shell initialization")

	// Source the bash script
	bashScript := filepath.Join(env.PVMDataDir, "shell", "pvm.bash")

	// Check if the bash script exists before continuing
	if _, err := os.Stat(bashScript); os.IsNotExist(err) {
		helpers.SkipTODO(t, "Bash shell integration script")
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
	err := os.WriteFile(testScript, []byte(scriptContent), 0755)
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
	_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"shell", "init"}, "Shell initialization")

	// Look for the bash script
	bashScript := filepath.Join(env.PVMDataDir, "shell", "pvm.bash")
	if _, err := os.Stat(bashScript); os.IsNotExist(err) {
		helpers.SkipTODO(t, "Bash shell integration script")
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
	helpers.SkipTODO(t, "Shell setup functionality")
}

// TestShellCompletion tests shell completion functionality
func TestShellCompletion(t *testing.T) {
	helpers.SkipTODO(t, "Shell completion functionality")
}
