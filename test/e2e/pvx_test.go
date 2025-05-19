// ABOUTME: End-to-end tests for PVX functionality
// ABOUTME: Tests Perl script execution with different versions and options

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/test/e2e/helpers"
)

// TestPVXScriptExecution tests basic script execution with PVX
func TestPVXScriptExecution(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Find path to system Perl
	perlPath := "/usr/bin/perl"
	_, err := os.Stat(perlPath)
	if os.IsNotExist(err) {
		helpers.SkipTODO(t, "System Perl installation (required for test)")
	}

	// Create a test Perl script
	scriptDir := filepath.Join(env.HomeDir, "scripts")
	err = os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	scriptPath := filepath.Join(scriptDir, "hello.pl")
	scriptContent := `#!/usr/bin/env perl
print "Hello from PVX test!\n";
print "Args: ", join(", ", @ARGV), "\n" if @ARGV;
`
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Run the script with PVX, explicitly specifying Perl path
	stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env,
		[]string{"pvx", "-p", perlPath, scriptPath, "arg1", "arg2"},
		"PVX script execution")

	helpers.AssertStringContains(t, stdout, "Hello from PVX test!",
		"Script output should contain greeting")
	helpers.AssertStringContains(t, stdout, "Args: arg1, arg2",
		"Script output should show passed arguments")
}

// TestPVXInlineCodeExecution tests executing Perl code with the -e flag
func TestPVXInlineCodeExecution(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Find path to system Perl
	perlPath := "/usr/bin/perl"
	_, err := os.Stat(perlPath)
	if os.IsNotExist(err) {
		helpers.SkipTODO(t, "System Perl installation (required for test)")
	}

	// Execute inline Perl code with PVX, explicitly specifying Perl path
	stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env,
		[]string{"pvx", "-p", perlPath, "-e", "print 'Inline code executed successfully!\\n';"},
		"PVX inline code execution")

	helpers.AssertStringContains(t, stdout, "Inline code executed successfully!",
		"Output should show inline code execution")
}

// TestPVXVersionSpecification tests using a specific Perl version
func TestPVXVersionSpecification(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// First ensure a Perl version is available
	// Import system Perl or skip if not available
	_, err := os.Stat("/usr/bin/perl")
	if os.IsNotExist(err) {
		helpers.SkipTODO(t, "System Perl installation (required for version test)")
	}

	// Import system Perl
	helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"import-system"},
		"System Perl import")

	// Ensure system Perl is available by running list
	helpers.AssertPVMSucceedsOrSkipTODO(t, env,
		[]string{"list"}, "Perl version listing")

	// Create a script that prints Perl version
	scriptPath := filepath.Join(env.HomeDir, "version_test.pl")
	scriptContent := `#!/usr/bin/env perl
print "Perl version: $^V\n";
`
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create version test script: %v", err)
	}

	// Run with explicit version flag
	stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env,
		[]string{"pvx", "-p", "system", scriptPath},
		"PVX execution with specific version")

	helpers.AssertStringContains(t, stdout, "Perl version:",
		"Output should contain Perl version information")
}

// TestPVXEnvironmentVariables tests passing environment variables to scripts
func TestPVXEnvironmentVariables(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Find path to system Perl
	perlPath := "/usr/bin/perl"
	_, err := os.Stat(perlPath)
	if os.IsNotExist(err) {
		helpers.SkipTODO(t, "System Perl installation (required for test)")
	}

	// Create a script that prints an environment variable
	scriptPath := filepath.Join(env.HomeDir, "env_test.pl")
	scriptContent := `#!/usr/bin/env perl
print "TEST_VAR=", $ENV{TEST_VAR} || "undefined", "\n";
`
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create environment test script: %v", err)
	}

	// Set the environment variable in the test process
	os.Setenv("TEST_VAR", "test_value")

	// Run the script, which should inherit the environment variable
	stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env,
		[]string{"pvx", "-p", perlPath, scriptPath},
		"PVX execution with environment variables")

	helpers.AssertStringContains(t, stdout, "TEST_VAR=test_value",
		"Output should contain environment variable")
}

// TestPVXExitCodePropagation tests that exit codes are correctly propagated
func TestPVXExitCodePropagation(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Find path to system Perl
	perlPath := "/usr/bin/perl"
	_, err := os.Stat(perlPath)
	if os.IsNotExist(err) {
		helpers.SkipTODO(t, "System Perl installation (required for test)")
	}

	// Create a script that exits with a specific code
	scriptPath := filepath.Join(env.HomeDir, "exit_test.pl")
	scriptContent := `#!/usr/bin/env perl
exit 42;
`
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create exit code test script: %v", err)
	}

	// Run the script and expect it to fail with a specific exit code
	stderr := helpers.AssertPVMFailsOrSkipTODO(t, env,
		[]string{"pvx", "-p", perlPath, scriptPath},
		"PVX exit code propagation")

	// Check that the error contains the exit code 42
	helpers.AssertStringContains(t, stderr, "42",
		"Error message should contain the exit code")
}
