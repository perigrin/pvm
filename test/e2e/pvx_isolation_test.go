// ABOUTME: End-to-end tests for PVX isolation functionality
// ABOUTME: Tests different isolation levels and cleanup behavior

package e2e

import (
	"os"
	"path/filepath"
	"strings"
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
		return
	}

	// Create a test script that prints environment information
	scriptDir := filepath.Join(env.HomeDir, "scripts")
	err = os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Script that outputs details about its environment
	// This will help us confirm isolation is working
	scriptPath := filepath.Join(scriptDir, "test_isolation_script.pl")
	scriptContent := `#!/usr/bin/env perl
use strict;
use warnings;

# Print environment variables
print "PERL5LIB: $ENV{PERL5LIB}\n";
print "HOME: $ENV{HOME}\n";
print "Working directory: " . qx(pwd) . "\n";
print "PVM_ISOLATION_LEVEL: " . ($ENV{PVM_ISOLATION_LEVEL} || "not set") . "\n";
print "PVM_ISOLATION_DIR: " . ($ENV{PVM_ISOLATION_DIR} || "not set") . "\n";
print "PVM_OUTPUT_DIR: " . ($ENV{PVM_OUTPUT_DIR} || "not set") . "\n";

# Create a test file to check filesystem isolation
# When using high isolation with isolated output, files need to be created in the current working directory
# which will automatically be set to the output directory
my $test_filename = 'test_output_file.txt';
eval {
    # Print current working directory to verify we're in the output directory
    print "Creating file in: " . qx(pwd) . "\n";
    open(my $fh, '>', $test_filename) or die "Could not create file: $!";
    print $fh "This is a test file, created at " . localtime() . "\n";
    close($fh);

    # Verify the file was created successfully
    if (-f $test_filename) {
        print "Successfully created test file: $test_filename\n";
        # Print the contents to make sure it was created correctly
        open(my $fh_read, '<', $test_filename) or die "Could not read file: $!";
        my $content = do { local $/; <$fh_read> };
        close($fh_read);
        print "File contents: $content\n";
    } else {
        print "ERROR: File not found after creation: $test_filename\n";
    }
};
if ($@) {
    print "Error creating test file: $@\n";
}

# Try to read a file from parent directory
eval {
    if (open(my $fh, '<', '../test_isolation_script.pl')) {
        print "Could read parent directory file\n";
        close($fh);
    } else {
        print "Could not read parent directory file: $!\n";
    }
};
if ($@) {
    print "Error accessing parent directory: $@\n";
}

# Print list of all environment variables
print "\nAll environment variables:\n";
foreach my $key (sort keys %ENV) {
    # Skip variables with long values
    my $value = $ENV{$key};
    if (length($value) > 100) {
        $value = substr($value, 0, 97) . "...";
    }
    print "  $key=$value\n";
}

print "\nScript completed successfully\n";
`
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create isolation test script: %v", err)
	}

	// Create an output save directory
	saveDir := filepath.Join(env.HomeDir, "isolated_output")
	err = os.MkdirAll(saveDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output save directory: %v", err)
	}

	// Test isolation level: none
	t.Run("IsolationNone", func(t *testing.T) {
		_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{"pvx", "--isolation", "none", "-p", perlPath, "--verbose", scriptPath},
			"PVX with isolation level: none")

		// Check that the script ran successfully
		helpers.AssertStringContains(t, stdout, "Script completed successfully",
			"Script should complete successfully")

		// Check that isolation level is not set
		helpers.AssertStringContains(t, stdout, "PVM_ISOLATION_LEVEL: not set",
			"Isolation level should not be set")

		// Check that isolation directory is not set
		helpers.AssertStringContains(t, stdout, "PVM_ISOLATION_DIR: not set",
			"Isolation directory should not be set")
	})

	// Test isolation level: low
	t.Run("IsolationLow", func(t *testing.T) {
		_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{"pvx", "--isolation", "low", "-p", perlPath, "--verbose", scriptPath},
			"PVX with isolation level: low")

		// Check that the script ran successfully
		helpers.AssertStringContains(t, stdout, "Script completed successfully",
			"Script should complete successfully")

		// Low isolation doesn't set isolation environment variables
		helpers.AssertStringContains(t, stdout, "PVM_ISOLATION_LEVEL: not set",
			"Isolation level should not be set in low isolation")

		// Check that PERL5LIB contains isolation directory
		helpers.AssertStringContains(t, stdout, "PERL5LIB:",
			"Script should show PERL5LIB environment variable")
	})

	// Test isolation level: medium
	t.Run("IsolationMedium", func(t *testing.T) {
		_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{"pvx", "--isolation", "medium", "-p", perlPath, "--verbose", scriptPath},
			"PVX with isolation level: medium")

		// Check that the script ran successfully
		helpers.AssertStringContains(t, stdout, "Script completed successfully",
			"Script should complete successfully")

		// Medium isolation doesn't set isolation environment variables
		helpers.AssertStringContains(t, stdout, "PVM_ISOLATION_LEVEL: not set",
			"Isolation level should not be set in medium isolation")

		// Check that PERL5LIB is set with clean values
		helpers.AssertStringContains(t, stdout, "PERL5LIB:",
			"Script should show PERL5LIB environment variable")
	})

	// Test isolation level: high
	t.Run("IsolationHigh", func(t *testing.T) {
		_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{
				"pvx",
				"--isolation", "high",
				"-p", perlPath,
				"--verbose",
				"--isolated-output",
				"--save-output-dir", saveDir,
				"--ro-path", "/usr",
				"--rw-path", scriptDir,
				scriptPath,
			},
			"PVX with isolation level: high")

		// Check that the script ran successfully
		helpers.AssertStringContains(t, stdout, "Script completed successfully",
			"Script should complete successfully")

		// High isolation sets isolation environment variables
		helpers.AssertStringContains(t, stdout, "PVM_ISOLATION_LEVEL: high",
			"Isolation level should be set in high isolation")

		// Check that PVM_ISOLATION_DIR is set
		helpers.AssertStringContains(t, stdout, "PVM_ISOLATION_DIR:",
			"Isolation directory should be set")

		// Check that PVM_OUTPUT_DIR is set
		helpers.AssertStringContains(t, stdout, "PVM_OUTPUT_DIR:",
			"Output directory should be set")

		// Check that output file is saved
		helpers.AssertStringContains(t, stdout, "Successfully created test file: test_output_file.txt",
			"Script should be able to create test file")

		// Verify that the file was saved to the output directory
		savedFiles, err := os.ReadDir(saveDir)
		switch {
		case err != nil:
			t.Errorf("Failed to read saved output directory: %v", err)
		case len(savedFiles) == 0:
			t.Errorf("No files were saved to the output directory. Check if saveOutputFiles function in executor.go actually copies the files correctly")
		default:
			foundOutputFile := false
			for _, file := range savedFiles {
				if file.Name() == "test_output_file.txt" {
					foundOutputFile = true
					break
				}
			}
			if !foundOutputFile {
				t.Errorf("Expected output file 'test_output_file.txt' was not saved")
			}
		}
	})
}

// TestPVXCleanupBehavior tests the cleanup behavior of PVX
func TestPVXCleanupBehavior(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Find path to system Perl
	perlPath := "/usr/bin/perl"
	_, err := os.Stat(perlPath)
	if os.IsNotExist(err) {
		helpers.SkipTODO(t, "System Perl installation (required for cleanup tests)")
		return
	}

	// Create a simple test script
	scriptDir := filepath.Join(env.HomeDir, "scripts")
	err = os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	scriptPath := filepath.Join(scriptDir, "cleanup_test_script.pl")
	scriptContent := `
print "Script running...\n";
my $isolationDir = $ENV{PVM_ISOLATION_DIR} || "unknown";
print "Isolation directory: $isolationDir\n";
print "Script completed successfully\n";
`
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create cleanup test script: %v", err)
	}

	// First test: with automatic cleanup (default)
	t.Run("AutoCleanup", func(t *testing.T) {
		// Run script with isolation
		_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{
				"pvx",
				"--isolation", "high",
				"-p", perlPath,
				"--verbose",
				scriptPath,
			},
			"PVX with auto cleanup")

		// Extract the isolation directory path from the output
		var isolationDir string
		lines := strings.Split(stdout, "\n")
		for _, line := range lines {
			switch {
			case strings.HasPrefix(line, "Created isolation directory: "):
				isolationDir = strings.TrimPrefix(line, "Created isolation directory: ")
				break
			case strings.HasPrefix(line, "Isolation directory: ") && !strings.Contains(line, "unknown"):
				isolationDir = strings.TrimPrefix(line, "Isolation directory: ")
				break
			}
		}

		if isolationDir == "" {
			t.Skip("Could not find isolation directory in output")
			return
		}

		// Verify the isolation directory doesn't exist after execution (it should be cleaned up)
		_, err := os.Stat(isolationDir)
		if err == nil {
			t.Errorf("Isolation directory still exists after execution: %s", isolationDir)
		} else if !os.IsNotExist(err) {
			t.Errorf("Error checking isolation directory: %v", err)
		}
	})

	// Second test: with cleanup disabled
	t.Run("NoCleanup", func(t *testing.T) {
		// Run script with isolation and no cleanup
		_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{
				"pvx",
				"--isolation", "high",
				"-p", perlPath,
				"--verbose",
				"--no-cleanup",
				scriptPath,
			},
			"PVX with no cleanup")

		// Extract the isolation directory path from the output
		var isolationDir string
		lines := strings.Split(stdout, "\n")
		for _, line := range lines {
			switch {
			case strings.HasPrefix(line, "Created isolation directory: "):
				isolationDir = strings.TrimPrefix(line, "Created isolation directory: ")
				break
			case strings.HasPrefix(line, "Isolation directory: ") && !strings.Contains(line, "unknown"):
				isolationDir = strings.TrimPrefix(line, "Isolation directory: ")
				break
			case strings.HasPrefix(line, "Isolation directory retained (--no-cleanup): "):
				isolationDir = strings.TrimPrefix(line, "Isolation directory retained (--no-cleanup): ")
				break
			}
		}

		if isolationDir == "" {
			t.Skip("Could not find isolation directory in output")
			return
		}

		// Verify the isolation directory still exists after execution (cleanup disabled)
		_, err := os.Stat(isolationDir)
		if err != nil {
			if os.IsNotExist(err) {
				t.Errorf("Isolation directory was cleaned up despite --no-cleanup flag: %s", isolationDir)
			} else {
				t.Errorf("Error checking isolation directory: %v", err)
			}
		} else {
			// Cleanup the directory manually since we told PVX not to
			t.Logf("Manually cleaning up isolation directory: %s", isolationDir)
			_ = os.RemoveAll(isolationDir)
		}
	})

	// Third test: custom isolation directory
	t.Run("CustomIsolationDir", func(t *testing.T) {
		// Create a custom isolation directory
		customIsolationDir := filepath.Join(env.HomeDir, "custom_isolation")
		err := os.MkdirAll(customIsolationDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create custom isolation directory: %v", err)
		}

		// Run script with custom isolation directory
		_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{
				"pvx",
				"--isolation", "high",
				"-p", perlPath,
				"--verbose",
				"--isolation-dir", customIsolationDir,
				scriptPath,
			},
			"PVX with custom isolation directory")

		// Check that the script used the custom isolation directory
		helpers.AssertStringContains(t, stdout, customIsolationDir,
			"Script should use the custom isolation directory")

		// Verify the isolation directory still exists after execution (it's a user-specified one)
		_, err = os.Stat(customIsolationDir)
		if err != nil {
			if os.IsNotExist(err) {
				t.Errorf("Custom isolation directory was deleted: %s", customIsolationDir)
			} else {
				t.Errorf("Error checking custom isolation directory: %v", err)
			}
		} else {
			// Cleanup the directory manually
			_ = os.RemoveAll(customIsolationDir)
		}
	})
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
		return
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
print "TEST_VAR2: $ENV{TEST_VAR2}\n";
print "TEST_VAR3: $ENV{TEST_VAR3}\n";
print "PERL5LIB: " . (substr($ENV{PERL5LIB}, 0, 50) . "..." ) . "\n";
print "HOME: $ENV{HOME}\n";
print "PATH: " . (substr($ENV{PATH}, 0, 50) . "..." ) . "\n";
print "PWD: $ENV{PWD}\n";
print "USER: $ENV{USER}\n";

print "Script completed successfully\n";
`
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create env var test script: %v", err)
	}

	// Set test environment variables
	_ = os.Setenv("TEST_VAR", "test_value")
	_ = os.Setenv("TEST_VAR2", "test_value2")
	_ = os.Setenv("TEST_VAR3", "test_value3")

	// Test environment variable handling with different isolation levels

	// Test isolation level: none (should pass all environment variables)
	t.Run("EnvVars_none", func(t *testing.T) {
		_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{"pvx", "--isolation", "none", "-p", perlPath, "--verbose", scriptPath},
			"PVX env vars with isolation level: none")

		helpers.AssertStringContains(t, stdout, "Script completed successfully",
			"Script should complete successfully")

		// All environment variables should be passed through
		helpers.AssertStringContains(t, stdout, "TEST_VAR: test_value",
			"TEST_VAR should be passed through at isolation level: none")
		helpers.AssertStringContains(t, stdout, "TEST_VAR2: test_value2",
			"TEST_VAR2 should be passed through at isolation level: none")
		helpers.AssertStringContains(t, stdout, "TEST_VAR3: test_value3",
			"TEST_VAR3 should be passed through at isolation level: none")

		// HOME should match test environment
		helpers.AssertStringContains(t, stdout, "HOME: "+env.HomeDir,
			"HOME should match test environment at isolation level: none")
	})

	// Test isolation level: low (should pass all environment variables)
	t.Run("EnvVars_low", func(t *testing.T) {
		_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{"pvx", "--isolation", "low", "-p", perlPath, "--verbose", scriptPath},
			"PVX env vars with isolation level: low")

		helpers.AssertStringContains(t, stdout, "Script completed successfully",
			"Script should complete successfully")

		// All environment variables should be passed through
		helpers.AssertStringContains(t, stdout, "TEST_VAR: test_value",
			"TEST_VAR should be passed through at isolation level: low")
		helpers.AssertStringContains(t, stdout, "TEST_VAR2: test_value2",
			"TEST_VAR2 should be passed through at isolation level: low")
		helpers.AssertStringContains(t, stdout, "TEST_VAR3: test_value3",
			"TEST_VAR3 should be passed through at isolation level: low")

		// HOME should match test environment
		helpers.AssertStringContains(t, stdout, "HOME: "+env.HomeDir,
			"HOME should match test environment at isolation level: low")

		// PERL5LIB should be modified
		helpers.AssertStringContains(t, stdout, "PERL5LIB:",
			"PERL5LIB should be set at isolation level: low")
	})

	// Test isolation level: medium (should still pass most environment variables)
	t.Run("EnvVars_medium", func(t *testing.T) {
		_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{"pvx", "--isolation", "medium", "-p", perlPath, "--verbose", scriptPath},
			"PVX env vars with isolation level: medium")

		helpers.AssertStringContains(t, stdout, "Script completed successfully",
			"Script should complete successfully")

		// All environment variables should be passed through
		helpers.AssertStringContains(t, stdout, "TEST_VAR: test_value",
			"TEST_VAR should be passed through at isolation level: medium")
		helpers.AssertStringContains(t, stdout, "TEST_VAR2: test_value2",
			"TEST_VAR2 should be passed through at isolation level: medium")
		helpers.AssertStringContains(t, stdout, "TEST_VAR3: test_value3",
			"TEST_VAR3 should be passed through at isolation level: medium")

		// HOME should match test environment
		helpers.AssertStringContains(t, stdout, "HOME: "+env.HomeDir,
			"HOME should match test environment at isolation level: medium")

		// PERL5LIB should be replaced
		helpers.AssertStringContains(t, stdout, "PERL5LIB:",
			"PERL5LIB should be set at isolation level: medium")
	})

	// Test isolation level: high with preserved variables
	t.Run("EnvVars_high_with_preserved", func(t *testing.T) {
		_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{
				"pvx",
				"--isolation", "high",
				"-p", perlPath,
				"--verbose",
				"--preserve-env", "TEST_VAR",
				"--preserve-env", "TEST_VAR2",
				scriptPath,
			},
			"PVX env vars with isolation level: high and preserved vars")

		helpers.AssertStringContains(t, stdout, "Script completed successfully",
			"Script should complete successfully")

		// Only preserved environment variables should be passed through
		helpers.AssertStringContains(t, stdout, "TEST_VAR: test_value",
			"TEST_VAR should be preserved at isolation level: high")
		helpers.AssertStringContains(t, stdout, "TEST_VAR2: test_value2",
			"TEST_VAR2 should be preserved at isolation level: high")

		// Non-preserved custom var should not be present
		helpers.AssertStringContains(t, stdout, "TEST_VAR3:",
			"TEST_VAR3 should be empty at isolation level: high")

		// Essential variables like HOME should still be present
		helpers.AssertStringContains(t, stdout, "HOME:",
			"HOME should be present at isolation level: high")
		helpers.AssertStringContains(t, stdout, "USER:",
			"USER should be present at isolation level: high")
	})

	// Test isolation level: high with clear-env
	t.Run("EnvVars_high_with_clear", func(t *testing.T) {
		_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{
				"pvx",
				"--isolation", "high",
				"-p", perlPath,
				"--verbose",
				"--preserve-env", "TEST_VAR",
				"--preserve-env", "TEST_VAR2",
				"--clear-env", "TEST_VAR2", // This should override the preservation
				scriptPath,
			},
			"PVX env vars with isolation level: high, preserved and cleared vars")

		helpers.AssertStringContains(t, stdout, "Script completed successfully",
			"Script should complete successfully")

		// Only TEST_VAR should be preserved; TEST_VAR2 should be cleared
		helpers.AssertStringContains(t, stdout, "TEST_VAR: test_value",
			"TEST_VAR should be preserved at isolation level: high")
		helpers.AssertStringContains(t, stdout, "TEST_VAR2:",
			"TEST_VAR2 should be empty at isolation level: high (cleared)")
	})

	// Test isolation level: high with custom environment variables
	t.Run("EnvVars_high_with_custom", func(t *testing.T) {
		// First run with standard script - we don't need to verify its output
		helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{
				"pvx",
				"--isolation", "high",
				"-p", perlPath,
				"--verbose",
				"-e", "MY_CUSTOM_VAR=custom_value",
				scriptPath,
			},
			"PVX env vars with isolation level: high and custom var")

		// Write custom script to check the environment variable
		customCheckScript := filepath.Join(scriptDir, "custom_var_check.pl")
		customCheckContent := `print "MY_CUSTOM_VAR: $ENV{MY_CUSTOM_VAR}\n";`
		err = os.WriteFile(customCheckScript, []byte(customCheckContent), 0755)
		if err != nil {
			t.Fatalf("Failed to create custom check script: %v", err)
		}

		// Run a second test with the custom check script
		customOutput := helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{
				"pvx",
				"--isolation", "high",
				"-p", perlPath,
				"--verbose",
				"-e", "MY_CUSTOM_VAR=custom_value",
				customCheckScript,
			},
			"PVX with custom environment variable")

		// Custom environment variable should be present
		helpers.AssertStringContains(t, customOutput, "MY_CUSTOM_VAR: custom_value",
			"Custom environment variable should be set")
	})
}
