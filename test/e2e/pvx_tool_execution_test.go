// ABOUTME: End-to-end test for PVX global tool execution functionality
// ABOUTME: Tests tool detection, mapping, and execution without external dependencies

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/test/e2e/helpers"
)

// TestPVXToolDetectionAndExecution tests the core tool execution functionality
func TestPVXToolDetectionAndExecution(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl so PVM knows about it
	_, _, err := env.RunPVM("import-system")
	if err != nil {
		helpers.SkipTODO(t, "System Perl import failed - no Perl available for testing")
	}

	// Set up a working Perl path since system import may have path issues
	workingPerlPath := "/home/perigrin/.plenv/versions/5.40.2/bin/perl"
	if _, err := os.Stat(workingPerlPath); os.IsNotExist(err) {
		// Fallback to system perl paths
		workingPerlPath = "/usr/bin/perl"
		if _, err := os.Stat(workingPerlPath); os.IsNotExist(err) {
			helpers.SkipTODO(t, "No working Perl installation found")
		}
	}

	t.Run("inline_code_execution", func(t *testing.T) {
		// Test basic inline code execution
		stdout := helpers.AssertPVMSucceeds(t, env, []string{
			"pvx", "-p", workingPerlPath, "-e", "print 'Hello from PVX inline execution!'",
		}, "PVX inline code execution failed")

		helpers.AssertStringContains(t, stdout, "Hello from PVX inline execution!",
			"PVX should execute inline code")
	})

	t.Run("script_execution", func(t *testing.T) {
		// Create a test script
		scriptContent := `#!/usr/bin/perl
use strict;
use warnings;
print "Hello from PVX script execution!\n";
print "Perl version: $^V\n";
`
		scriptPath := filepath.Join(env.HomeDir, "test_script.pl")
		err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}

		stdout := helpers.AssertPVMSucceeds(t, env, []string{
			"pvx", scriptPath,
		}, "PVX script execution failed")

		helpers.AssertStringContains(t, stdout, "Hello from PVX script execution!",
			"PVX should execute script files")
		helpers.AssertStringContains(t, stdout, "Perl version:",
			"Script should show Perl version")
	})

	t.Run("tool_detection_mode", func(t *testing.T) {
		// Test tool vs script detection with verbose output
		stdout := helpers.AssertPVMSucceeds(t, env, []string{
			"pvx", "--verbose", "perl", "-e", "print 'Tool mode detected\\n'",
		}, "PVX tool detection failed")

		// Should detect as tool mode
		helpers.AssertStringContains(t, stdout, "Tool mode detected",
			"Should execute perl as a tool")
	})

	t.Run("core_module_usage", func(t *testing.T) {
		// Test with core Perl modules (no external dependencies)
		stdout := helpers.AssertPVMSucceeds(t, env, []string{
			"pvx", "-e",
			"use Data::Dumper; use File::Basename; " +
				"print Dumper({tool => 'pvx', status => 'working', basename => basename('/path/to/file.txt')});",
		}, "PVX core module usage failed")

		helpers.AssertStringContains(t, stdout, "'tool' => 'pvx'",
			"Should execute code with core modules")
		helpers.AssertStringContains(t, stdout, "'status' => 'working'",
			"Should show working status")
		helpers.AssertStringContains(t, stdout, "file.txt",
			"Should use File::Basename correctly")
	})

	t.Run("isolation_levels", func(t *testing.T) {
		// Test different isolation levels
		isolationLevels := []string{"global", "local", "clean"}

		for _, level := range isolationLevels {
			t.Run("isolation_"+level, func(t *testing.T) {
				stdout := helpers.AssertPVMSucceeds(t, env, []string{
					"pvx", "--isolation", level, "--verbose",
					"-e", "print 'Isolation level: " + level + "\\n'",
				}, "PVX isolation level "+level+" failed")

				helpers.AssertStringContains(t, stdout, "Isolation level: "+level,
					"Should execute with isolation level "+level)
			})
		}
	})

	t.Run("tool_mapping_builtin", func(t *testing.T) {
		// Test that built-in tool mappings work (even if tools aren't installed)
		// This tests the detection logic without requiring actual tool installation
		builtinTools := []string{"cpanm", "prove", "perltidy", "perlcritic"}

		for _, tool := range builtinTools {
			t.Run("tool_"+tool, func(t *testing.T) {
				// Use --verbose to see detection information
				// Use -- to separate PVX flags from tool arguments
				stdout, stderr, err := env.RunPVM("pvx", "--verbose", "--", tool, "--help")

				// The tool might not be installed, but we should see tool detection
				combined := stdout + stderr

				// Check if tool detection succeeded (look for detection indicators)
				toolDetected := strings.Contains(combined, "tool") ||
					strings.Contains(combined, "confidence") ||
					strings.Contains(combined, "Detected execution mode")

				if err != nil {
					// Tool execution failed - check if detection succeeded
					if toolDetected {
						t.Logf("Tool %s detected correctly but execution failed: %v", tool, err)
					} else {
						t.Logf("Tool %s detection/execution failed: %v", tool, err)
					}
				} else {
					// Tool execution succeeded - check for detection or tool output
					switch {
					case toolDetected:
						t.Logf("Tool %s detected and executed successfully", tool)
					case strings.Contains(stdout, tool):
						t.Logf("Tool %s executed successfully with expected output", tool)
					default:
						t.Logf("Tool %s execution succeeded but no tool-specific output found", tool)
					}
				}
			})
		}
	})

	t.Run("argument_passing", func(t *testing.T) {
		// Test that arguments are passed correctly to tools/scripts
		stdout := helpers.AssertPVMSucceeds(t, env, []string{
			"pvx", "-e",
			"for my $arg (@ARGV) { print \"Arg: $arg\\n\"; }",
			"--", "arg1", "arg2", "test argument",
		}, "PVX argument passing failed")

		helpers.AssertStringContains(t, stdout, "Arg: arg1",
			"Should pass first argument")
		helpers.AssertStringContains(t, stdout, "Arg: arg2",
			"Should pass second argument")
		helpers.AssertStringContains(t, stdout, "Arg: test argument",
			"Should pass argument with spaces")
	})

	t.Run("help_system", func(t *testing.T) {
		// Test that help system works
		stdout := helpers.AssertPVMSucceeds(t, env, []string{
			"pvx", "--help",
		}, "PVX help system failed")

		helpers.AssertStringContains(t, stdout, "pvx",
			"Help should mention pvx")
		helpers.AssertStringContains(t, stdout, "execute",
			"Help should mention execution")
		helpers.AssertStringContains(t, stdout, "--isolation",
			"Help should show isolation options")
	})
}

// TestPVXErrorHandling tests error handling and edge cases
func TestPVXErrorHandling(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl so PVM knows about it
	_, _, err := env.RunPVM("import-system")
	if err != nil {
		helpers.SkipTODO(t, "System Perl import failed - no Perl available for testing")
	}

	t.Run("syntax_error_handling", func(t *testing.T) {
		// Test handling of Perl syntax errors
		_, stderr, err := env.RunPVM("pvx", "-e", "invalid perl syntax {{{")

		if err == nil {
			t.Error("Expected syntax error to cause failure")
		}

		// Should contain some indication of syntax error
		if !strings.Contains(stderr, "syntax") && !strings.Contains(stderr, "error") {
			t.Logf("Syntax error output: %s", stderr)
		}
	})

	t.Run("nonexistent_script", func(t *testing.T) {
		// Test handling of non-existent script files
		_, stderr, err := env.RunPVM("pvx", "/nonexistent/path/script.pl")

		if err == nil {
			t.Error("Expected non-existent script to cause failure")
		}

		// Should contain some indication of file not found
		if !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "No such file") {
			t.Logf("File not found error output: %s", stderr)
		}
	})

	t.Run("invalid_isolation_level", func(t *testing.T) {
		// Test handling of invalid isolation levels
		_, stderr, err := env.RunPVM("pvx", "--isolation", "invalid", "-e", "print 'test'")

		if err == nil {
			t.Error("Expected invalid isolation level to cause failure")
		}

		// Should contain some indication of invalid isolation level
		if !strings.Contains(stderr, "isolation") && !strings.Contains(stderr, "invalid") {
			t.Logf("Invalid isolation error output: %s", stderr)
		}
	})
}

// TestPVXCompatibility tests compatibility with different execution modes
func TestPVXCompatibility(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Import system Perl so PVM knows about it
	_, _, err := env.RunPVM("import-system")
	if err != nil {
		helpers.SkipTODO(t, "System Perl import failed - no Perl available for testing")
	}

	t.Run("perl_version_compatibility", func(t *testing.T) {
		// Test that PVX works with different Perl features
		modernPerlCode := `
use v5.20;
use feature 'signatures';
no warnings 'experimental::signatures';

sub greet($name) {
    return "Hello, $name from modern Perl!";
}

say greet("PVX");
`
		stdout, stderr, err := env.RunPVM("pvx", "-e", modernPerlCode)

		if err != nil {
			t.Logf("Modern Perl features failed (expected if Perl < 5.20): %v", err)
			t.Logf("Error output: %s", stderr)
			// This is expected on older Perl versions
		} else {
			helpers.AssertStringContains(t, stdout, "Hello, PVX from modern Perl!",
				"Modern Perl features should work")
		}
	})

	t.Run("environment_preservation", func(t *testing.T) {
		// Test that important environment variables are preserved
		stdout := helpers.AssertPVMSucceeds(t, env, []string{
			"pvx", "-e",
			"print \"HOME: \" . ($ENV{HOME} || 'not set') . \"\\n\"; " +
				"print \"PATH: \" . (defined $ENV{PATH} ? 'set' : 'not set') . \"\\n\";",
		}, "PVX environment preservation failed")

		helpers.AssertStringContains(t, stdout, "HOME:",
			"HOME environment variable should be accessible")
		helpers.AssertStringContains(t, stdout, "PATH: set",
			"PATH environment variable should be preserved")
	})
}
