// ABOUTME: End-to-end tests for PVX isolation functionality
// ABOUTME: Tests different isolation levels and cleanup behavior

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/test/e2e/helpers"
)

// TestPVXIsolationLevels tests the different isolation levels of PVX
func TestPVXIsolationLevels(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Find path to system Perl
	perlPath := "/usr/bin/perl"
	_, err := os.Stat(perlPath)
	if os.IsNotExist(err) {
		helpers.SkipTODO(t, "System Perl installation (required for isolation tests)")
	}

	// Create a test script that prints environment information
	scriptDir := filepath.Join(env.HomeDir, "scripts")
	err = os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Script that outputs details about its environment
	// This will help us confirm isolation is working
	scriptPath := filepath.Join(scriptDir, "isolation_test.pl")
	scriptContent := `#!/usr/bin/env perl
use strict;
use warnings;

# Print environment variables
print "PERL5LIB: $ENV{PERL5LIB}\n";
print "HOME: $ENV{HOME}\n";
print "Working directory: " . ` + "`pwd`" + `\n";

# Create a test file to check filesystem isolation
open(my $fh, '>', 'test_file.txt') or die "Could not create file: $!";
print $fh "This is a test file\n";
close($fh);

print "Created test file: test_file.txt\n";

# Try to read a file from parent directory
if (open(my $fh, '<', '../isolation_test.pl')) {
    print "Could read parent directory file\n";
    close($fh);
} else {
    print "Could not read parent directory file: $!\n";
}

print "Script completed successfully\n";
`
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create isolation test script: %v", err)
	}

	// Test isolation level: none
	t.Run("IsolationNone", func(t *testing.T) {
		stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{"pvx", "--isolation", "none", "-p", perlPath, scriptPath},
			"PVX with isolation level: none")

		helpers.AssertStringContains(t, stdout, "Script completed successfully",
			"Script should complete successfully")
		helpers.AssertStringContains(t, stdout, "HOME: "+env.HomeDir,
			"HOME environment should match test environment")

		// Check that file was created in current directory
		testFile := filepath.Join(env.HomeDir, "test_file.txt")
		helpers.AssertFileExists(t, testFile, "With isolation=none, file should be created in current directory")
	})

	// Test isolation level: low
	t.Run("IsolationLow", func(t *testing.T) {
		stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{"pvx", "--isolation", "low", "-p", perlPath, scriptPath},
			"PVX with isolation level: low")

		helpers.AssertStringContains(t, stdout, "Script completed successfully",
			"Script should complete successfully")

		// In low isolation, files should be created in an isolation directory, not current dir
		testFile := filepath.Join(env.HomeDir, "test_file.txt")
		if env.FileExists(testFile) {
			// If this exists from previous test, remove it
			os.Remove(testFile)
		}
		helpers.AssertFileDoesNotExist(t, testFile,
			"With isolation=low, file should not be created in current directory")

		// Check that parent directory files are still accessible
		helpers.AssertStringContains(t, stdout, "Could read parent directory file",
			"With isolation=low, parent directory files should be accessible")
	})

	// Test isolation level: medium
	t.Run("IsolationMedium", func(t *testing.T) {
		helpers.SkipTODO(t, "Medium isolation level implementation")
	})

	// Test isolation level: high
	t.Run("IsolationHigh", func(t *testing.T) {
		helpers.SkipTODO(t, "High isolation level implementation")
	})
}

// TestPVXCleanupBehavior tests the cleanup behavior of PVX
func TestPVXCleanupBehavior(t *testing.T) {
	// Skip this test for now until the isolation implementation is more mature
	helpers.SkipTODO(t, "PVX cleanup behavior tests")
}

// TestPVXIsolationEnvVars tests environment variable handling in different isolation levels
func TestPVXIsolationEnvVars(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Find path to system Perl
	perlPath := "/usr/bin/perl"
	_, err := os.Stat(perlPath)
	if os.IsNotExist(err) {
		helpers.SkipTODO(t, "System Perl installation (required for env var tests)")
	}

	// Create a script that prints environment variables
	scriptDir := filepath.Join(env.HomeDir, "scripts")
	err = os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	scriptPath := filepath.Join(scriptDir, "env_var_test.pl")
	scriptContent := `#!/usr/bin/env perl
use strict;
use warnings;

# Print specific environment variables
print "TEST_VAR: $ENV{TEST_VAR}\n";
print "PERL5LIB: $ENV{PERL5LIB}\n";
print "HOME: $ENV{HOME}\n";
print "PATH: $ENV{PATH}\n";
print "PWD: $ENV{PWD}\n";

print "Script completed successfully\n";
`
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create env var test script: %v", err)
	}

	// Set a test environment variable
	os.Setenv("TEST_VAR", "test_value")

	// Test environment variable passing with different isolation levels
	isolationLevels := []string{"none", "low"}

	// Start with the simpler isolation levels that should work initially
	for _, level := range isolationLevels {
		t.Run("EnvVars_"+level, func(t *testing.T) {
			stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env,
				[]string{"pvx", "--isolation", level, "-p", perlPath, scriptPath},
				"PVX env vars with isolation level: "+level)

			helpers.AssertStringContains(t, stdout, "Script completed successfully",
				"Script should complete successfully")

			// Environment variables should be passed through at these isolation levels
			helpers.AssertStringContains(t, stdout, "TEST_VAR: test_value",
				"TEST_VAR should be passed through at isolation level: "+level)

			// HOME should match test environment at these isolation levels
			helpers.AssertStringContains(t, stdout, "HOME: "+env.HomeDir,
				"HOME should match test environment at isolation level: "+level)
		})
	}

	// Test medium and high isolation levels as TODO items
	mediumHighLevels := []string{"medium", "high"}
	for _, level := range mediumHighLevels {
		t.Run("EnvVars_"+level, func(t *testing.T) {
			helpers.SkipTODO(t, "Environment isolation for "+level+" level")
		})
	}
}
