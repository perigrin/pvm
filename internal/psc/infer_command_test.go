// ABOUTME: Tests for PSC infer command functionality and integration
// ABOUTME: Validates command flag parsing, file processing, and error handling

package psc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	basetesting "tamarou.com/pvm/internal/testing"
)

func TestInferCommand(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "psc_infer_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple test Perl file
	testContent := `#!/usr/bin/perl
use strict;
use warnings;

my $count = 42;
my $name = "test";
my @items = (1, 2, 3);
my %config = (debug => 1, verbose => 0);

sub process_item {
    my $item = shift;
    return $item * 2;
}

print "Count: $count\n";
`

	testFile := filepath.Join(tempDir, "test.pl")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		expectError bool
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "Basic inference to stdout",
			args: []string{testFile},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "use v") {
					t.Error("Expected Perl version pragma in output")
				}
			},
		},
		{
			name: "Inference with verbose style",
			args: []string{"--style=verbose", testFile},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "use v") {
					t.Error("Expected Perl version pragma in verbose output")
				}
			},
		},
		{
			name: "Inference with compact style",
			args: []string{"--style=compact", testFile},
			checkOutput: func(t *testing.T, output string) {
				if output == "" {
					t.Error("Expected non-empty output for compact style")
				}
			},
		},
		{
			name: "Inference with comments style",
			args: []string{"--style=comments", testFile},
			checkOutput: func(t *testing.T, output string) {
				if output == "" {
					t.Error("Expected non-empty output for comments style")
				}
			},
		},
		{
			name: "Inference with output file",
			args: []string{"--output=" + filepath.Join(tempDir, "output.pl"), testFile},
			checkOutput: func(t *testing.T, output string) {
				// Check that output file was created
				outputPath := filepath.Join(tempDir, "output.pl")
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Error("Expected output file to be created")
				}

				// Read and validate output file content
				content, err := os.ReadFile(outputPath)
				if err != nil {
					t.Errorf("Failed to read output file: %v", err)
					return
				}

				if !strings.Contains(string(content), "use v") {
					t.Error("Expected Perl version pragma in output file")
				}
			},
		},
		{
			name:        "Invalid style",
			args:        []string{"--style=invalid", testFile},
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
			cmd := newInferCommand()
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

func TestInferCommandFlagParsing(t *testing.T) {
	cmd := newInferCommand()

	// Test default flag values
	defaults := map[string]interface{}{
		"output":              "",
		"style":               "inline",
		"preserve-comments":   true,
		"preserve-formatting": true,
		"progress":            false,
		"verbose":             false,
	}

	for flagName, expectedDefault := range defaults {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Flag %s not found", flagName)
			continue
		}

		actualDefault := flag.DefValue
		expectedStr := ""
		switch v := expectedDefault.(type) {
		case string:
			expectedStr = v
		case bool:
			if v {
				expectedStr = "true"
			} else {
				expectedStr = "false"
			}
		}

		if actualDefault != expectedStr {
			t.Errorf("Flag %s: expected default %v, got %v", flagName, expectedStr, actualDefault)
		}
	}
}

func TestInferCommandHelp(t *testing.T) {
	cmd := newInferCommand()

	// Check that help text contains key information
	helpText := cmd.Long

	expectedSections := []string{
		"Type inference is deterministic",
		"Output formats:",
		"Examples:",
		"inline",
		"verbose",
		"compact",
		"comments",
	}

	for _, section := range expectedSections {
		if !strings.Contains(helpText, section) {
			t.Errorf("Help text missing expected section: %s", section)
		}
	}
}

func TestInferCommandIntegration(t *testing.T) {
	basetesting.SkipUnlessIntegration(t, "PSC infer command integration test")

	// Create a more complex test file to verify end-to-end functionality
	tempDir, err := os.MkdirTemp("", "psc_infer_integration")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	complexContent := `#!/usr/bin/perl
use strict;
use warnings;

# Test various types and constructs
my $number = 42;
my $string = "hello world";
my $float = 3.14;
my @array = (1, 2, 3, 4, 5);
my %hash = (
    name => "test",
    value => 100,
    active => 1,
);

sub calculate {
    my ($a, $b) = @_;
    return $a + $b;
}

sub process_data {
    my $data = shift;
    if (ref($data) eq 'ARRAY') {
        return scalar(@$data);
    }
    return 0;
}

my $result = calculate($number, 10);
my $count = process_data(\@array);

print "Result: $result, Count: $count\n";
`

	testFile := filepath.Join(tempDir, "complex.pl")
	err = os.WriteFile(testFile, []byte(complexContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create complex test file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "inferred.pl")

	// Run inference command
	cmd := newInferCommand()
	cmd.SetArgs([]string{
		"--output=" + outputFile,
		"--style=verbose",
		"--verbose",
		testFile,
	})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Inference command failed: %v", err)
	}

	// Verify output file was created and has reasonable content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(content)

	// Basic validation checks
	if !strings.Contains(outputStr, "use v") {
		t.Error("Output missing Perl version pragma")
	}

	if len(outputStr) < len(complexContent) {
		t.Error("Output seems too short - may have lost content")
	}

	// Should contain some of the original code structure
	if !strings.Contains(outputStr, "calculate") {
		t.Error("Output missing function names")
	}

	if !strings.Contains(outputStr, "process_data") {
		t.Error("Output missing function definitions")
	}
}
