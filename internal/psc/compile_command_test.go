// ABOUTME: Tests for PSC compile command functionality and target validation
// ABOUTME: Validates compilation targets, flag parsing, and integration with compiler registry

package psc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileCommand(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "psc_compile_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a typed Perl test file
	typedContent := `#!/usr/bin/perl
use v5.36;
use strict;
use warnings;

my Int $count = 42;
my Str $name = "test";
my ArrayRef[Int] @items = (1, 2, 3);
my HashRef[Str] %config = (debug => "yes", verbose => "no");

sub process_item(Int $item) -> Int {
    return $item * 2;
}

print "Count: $count\n";
`

	typedFile := filepath.Join(tempDir, "typed.pl")
	err = os.WriteFile(typedFile, []byte(typedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create typed test file: %v", err)
	}

	// Create an untyped Perl test file for inference testing
	untypedContent := `#!/usr/bin/perl
use strict;
use warnings;

my $count = 42;
my $name = "test";
my @items = (1, 2, 3);

sub process_item {
    my $item = shift;
    return $item * 2;
}

print "Count: $count\n";
`

	untypedFile := filepath.Join(tempDir, "untyped.pl")
	err = os.WriteFile(untypedFile, []byte(untypedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create untyped test file: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		inputFile   string
		expectError bool
		checkOutput func(t *testing.T, output string)
	}{
		{
			name:      "Clean target - strip annotations",
			args:      []string{"--target=clean", typedFile},
			inputFile: typedFile,
			checkOutput: func(t *testing.T, output string) {
				// Check for any version pragma (dynamic from PVM, not hard-coded)
				// Use regex for more precise matching
				hasVersionPragma := strings.Contains(output, "use v") && strings.Contains(output, ";")
				if !hasVersionPragma {
					t.Errorf("Expected Perl version pragma in clean output. Output: %q", output)
				}
				// Clean output should not contain type annotations
				if strings.Contains(output, "Int $count") {
					t.Error("Clean output should not contain type annotations")
				}
				// Check for clean variable declarations
				if !strings.Contains(output, "my $count") {
					t.Errorf("Clean output should contain variable declarations without types. Output: %q", output)
				}
			},
		},
		{
			name:      "Typed target - preserve annotations",
			args:      []string{"--target=typed", typedFile},
			inputFile: typedFile,
			checkOutput: func(t *testing.T, output string) {
				// Check for any version pragma (dynamic from PVM)
				if !strings.Contains(output, "use v") || !strings.Contains(output, ";") {
					t.Error("Expected Perl version pragma in typed output")
				}
				// Typed output should preserve type annotations
				if !strings.Contains(output, "Int $count") {
					t.Error("Typed output should preserve type annotations")
				}
			},
		},
		{
			name:      "Inferred target - add annotations",
			args:      []string{"--target=inferred", untypedFile},
			inputFile: untypedFile,
			checkOutput: func(t *testing.T, output string) {
				// Check for any version pragma (dynamic from PVM)
				if !strings.Contains(output, "use v") || !strings.Contains(output, ";") {
					t.Error("Expected Perl version pragma in inferred output")
				}
				// Inferred output should have added some annotations
				if output == "" {
					t.Error("Expected non-empty inferred output")
				}
			},
		},
		{
			name:      "Inferred with verbose style",
			args:      []string{"--target=inferred", "--style=verbose", untypedFile},
			inputFile: untypedFile,
			checkOutput: func(t *testing.T, output string) {
				if output == "" {
					t.Error("Expected non-empty verbose inferred output")
				}
			},
		},
		{
			name:      "Compile with output file",
			args:      []string{"--target=clean", "--output=" + filepath.Join(tempDir, "clean_output.pl"), typedFile},
			inputFile: typedFile,
			checkOutput: func(t *testing.T, output string) {
				// Check that output file was created
				outputPath := filepath.Join(tempDir, "clean_output.pl")
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Error("Expected output file to be created")
				}

				// Read and validate output file content
				content, err := os.ReadFile(outputPath)
				if err != nil {
					t.Errorf("Failed to read output file: %v", err)
					return
				}

				// Check for any version pragma (dynamic from PVM)
				contentStr := string(content)
				if !strings.Contains(contentStr, "use v") || !strings.Contains(contentStr, ";") {
					t.Error("Expected Perl version pragma in output file")
				}
			},
		},
		{
			name:        "Invalid target",
			args:        []string{"--target=invalid", typedFile},
			inputFile:   typedFile,
			expectError: true,
		},
		{
			name:        "Invalid confidence value",
			args:        []string{"--target=inferred", "--confidence=1.5", untypedFile},
			inputFile:   untypedFile,
			expectError: true,
		},
		{
			name:        "Invalid style",
			args:        []string{"--target=inferred", "--style=invalid", untypedFile},
			inputFile:   untypedFile,
			expectError: true,
		},
		{
			name:        "In-place and output together",
			args:        []string{"--in-place", "--output=test.pl", typedFile},
			inputFile:   typedFile,
			expectError: true,
		},
		{
			name:        "Nonexistent file",
			args:        []string{"/nonexistent/file.pl"},
			expectError: true,
		},
		{
			name:        "No file argument",
			args:        []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create command and set args
			cmd := newCompileCommand()
			cmd.SetArgs(tt.args)

			// Capture output
			output := &strings.Builder{}
			cmd.SetOut(output)
			cmd.SetErr(output)

			// Execute command
			err := cmd.Execute()

			if tt.expectError {
				if err == nil {
					t.Error("Expected command to fail, but it succeeded")
				}
				return
			}

			if err != nil {
				t.Errorf("Command failed unexpectedly: %v", err)
				return
			}

			// Run output checks if provided
			if tt.checkOutput != nil {
				tt.checkOutput(t, output.String())
			}
		})
	}
}

func TestCompileCommandFlagParsing(t *testing.T) {
	cmd := newCompileCommand()

	// Test default flag values
	defaults := map[string]string{
		"target":              "typed",
		"output":              "",
		"in-place":            "false",
		"preserve-comments":   "true",
		"preserve-formatting": "true",
		"style":               "inline",
		"confidence":          "0.7",
		"include-uncertain":   "false",
		"progress":            "false",
		"verbose":             "false",
		"strict":              "false",
	}

	for flagName, expectedDefault := range defaults {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Flag %s not found", flagName)
			continue
		}

		if flag.DefValue != expectedDefault {
			t.Errorf("Flag %s: expected default %s, got %s", flagName, expectedDefault, flag.DefValue)
		}
	}
}

func TestCompileCommandHelp(t *testing.T) {
	cmd := newCompileCommand()

	// Check that help text contains key information
	helpText := cmd.Long

	expectedSections := []string{
		"Compilation targets:",
		"clean",
		"typed",
		"inferred",
		"Output options:",
		"Examples:",
		"--target=clean",
		"--target=inferred",
	}

	for _, section := range expectedSections {
		if !strings.Contains(helpText, section) {
			t.Errorf("Help text missing expected section: %s", section)
		}
	}
}

func TestCompileCommandInPlace(t *testing.T) {
	// Test in-place compilation functionality
	tempDir, err := os.MkdirTemp("", "psc_compile_inplace")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file with type annotations
	originalContent := `#!/usr/bin/perl
use v5.36;

my Int $test = 42;
print "Test: $test\n";
`

	testFile := filepath.Join(tempDir, "inplace_test.pl")
	err = os.WriteFile(testFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run in-place compilation to clean target
	cmd := newCompileCommand()
	cmd.SetArgs([]string{"--target=clean", "--in-place", testFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("In-place compilation failed: %v", err)
	}

	// Read the modified file
	modifiedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	modifiedStr := string(modifiedContent)

	// Check that type annotations were removed
	if strings.Contains(modifiedStr, "Int $test") {
		t.Error("In-place clean compilation should have removed type annotations")
	}

	if !strings.Contains(modifiedStr, "my $test") {
		t.Error("In-place compilation should preserve variable declarations")
	}

	// Check that original content structure is preserved (with dynamic version from PVM)
	if !strings.Contains(modifiedStr, "use v") || !strings.Contains(modifiedStr, ";") {
		t.Error("In-place compilation should preserve version pragma")
	}

	if !strings.Contains(modifiedStr, "print") {
		t.Error("In-place compilation should preserve print statements")
	}
}

func TestCompileCommandTargetValidation(t *testing.T) {
	// Test that all expected compilation targets are supported
	tempDir, err := os.MkdirTemp("", "psc_compile_targets")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testContent := `#!/usr/bin/perl
my $test = 42;
`

	testFile := filepath.Join(tempDir, "test.pl")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validTargets := []string{"clean", "typed", "inferred"}

	for _, target := range validTargets {
		t.Run("target_"+target, func(t *testing.T) {
			cmd := newCompileCommand()
			cmd.SetArgs([]string{"--target=" + target, testFile})

			output := &strings.Builder{}
			cmd.SetOut(output)
			cmd.SetErr(output)

			err := cmd.Execute()
			if err != nil {
				t.Errorf("Target %s should be valid but failed: %v", target, err)
			}
		})
	}
}
