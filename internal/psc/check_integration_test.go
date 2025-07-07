// ABOUTME: End-to-end integration tests for PSC check command
// ABOUTME: Tests complete workflow from command line to error output

package psc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/typechecker"
)

// TestCheckCommandWorkflow tests the complete check command workflow
func TestCheckCommandWorkflow(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name         string
		filename     string
		content      string
		expectErrors int
		expectOutput []string
	}{
		{
			name:     "Valid typed Perl code",
			filename: "valid.pl",
			content: `use strict;
my Int $count = 42;
my Str $name = "hello";
print "Count: $count, Name: $name\n";`,
			expectErrors: 0,
			expectOutput: []string{},
		},
		{
			name:     "Type mismatch error",
			filename: "mismatch.pl",
			content: `use strict;
my Int $count = "hello";
print "Count: $count\n";`,
			expectErrors: 1,
			expectOutput: []string{
				"error:",
				"is not compatible with",
				"help:",
			},
		},
		{
			name:     "Multiple type errors",
			filename: "multiple.pl",
			content: `use strict;
my Int $count = "hello";
my Str $name = [];
my Bool $flag = "not_bool";`,
			expectErrors: 3,
			expectOutput: []string{
				"error:",
				"is not compatible with",
			},
		},
		{
			name:     "Mixed typed and untyped code",
			filename: "mixed.pl",
			content: `use strict;
my Int $typed_var = 42;
my $untyped_var = "hello";
my Str $another_typed = $untyped_var;`,
			expectErrors: 0, // Should handle gracefully with type inference
			expectOutput: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, tt.filename)
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test the complete check workflow
			errorCount, err := checkFile(ui.NewDefaultOutput(), testFile, false, false, false, false)
			if err != nil {
				t.Fatalf("checkFile failed: %v", err)
			}

			if errorCount != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, errorCount)
			}

			// Test that the file can be type checked without errors
			tc, err := typechecker.NewTypeCheck()
			if err != nil {
				t.Fatalf("Failed to create typechecker: %v", err)
			}

			result, err := tc.CheckFile(testFile)
			if err != nil {
				t.Fatalf("TypeChecker.CheckFile failed: %v", err)
			}

			if len(result.Errors) != tt.expectErrors {
				t.Errorf("TypeChecker found %d errors, expected %d", len(result.Errors), tt.expectErrors)
			}

			// Test error formatting if errors are expected
			if tt.expectErrors > 0 {
				formatter := NewErrorFormatter()
				formatter.SetColorEnabled(false) // Disable color for testing
				output := formatter.FormatErrors(result.Errors)

				for _, expectedOutput := range tt.expectOutput {
					if !strings.Contains(output, expectedOutput) {
						t.Errorf("Expected output to contain %q, got:\n%s", expectedOutput, output)
					}
				}
			}
		})
	}
}

// TestCheckCommandRecursive tests recursive directory checking
func TestCheckCommandRecursive(t *testing.T) {
	tempDir := t.TempDir()

	// Create a directory structure with multiple Perl files
	libDir := filepath.Join(tempDir, "lib")
	err := os.MkdirAll(libDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	// Create test files
	files := map[string]string{
		"script.pl": `use strict;
my Int $count = 42;
print "Count: $count\n";`,
		"lib/Module.pm": `package Module;
use strict;
my Str $name = "module";
1;`,
		"lib/Error.pm": `package Error;
use strict;
my Int $bad = "error";
1;`,
		"not_perl.txt": "This is not a Perl file",
	}

	for filename, content := range files {
		fullPath := filepath.Join(tempDir, filename)
		dir := filepath.Dir(fullPath)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	// Test recursive checking
	totalFiles, totalErrors, err := checkDirectory(ui.NewDefaultOutput(), tempDir, false, false, false)
	if err != nil {
		t.Fatalf("checkDirectory failed: %v", err)
	}

	// Should check 3 Perl files (.pl and .pm files)
	if totalFiles != 3 {
		t.Errorf("Expected to check 3 files, checked %d", totalFiles)
	}

	// Should find 1 error in lib/Error.pm
	if totalErrors != 1 {
		t.Errorf("Expected 1 error, found %d", totalErrors)
	}
}

// TestCheckCommandErrorHandling tests various error conditions
func TestCheckCommandErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() string
		expectError bool
	}{
		{
			name: "Non-existent file",
			setup: func() string {
				return "/non/existent/file.pl"
			},
			expectError: true,
		},
		{
			name: "Directory without recursive flag",
			setup: func() string {
				tempDir := t.TempDir()
				return tempDir
			},
			expectError: false, // Should skip directory, not error
		},
		{
			name: "Empty file",
			setup: func() string {
				tempDir := t.TempDir()
				emptyFile := filepath.Join(tempDir, "empty.pl")
				os.WriteFile(emptyFile, []byte(""), 0644)
				return emptyFile
			},
			expectError: false, // Empty file should not error
		},
		{
			name: "Invalid Perl syntax",
			setup: func() string {
				tempDir := t.TempDir()
				invalidFile := filepath.Join(tempDir, "invalid.pl")
				os.WriteFile(invalidFile, []byte("my $var = unclosed string"), 0644)
				return invalidFile
			},
			expectError: false, // Syntax errors are reported, not fatal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()

			_, err := checkFile(ui.NewDefaultOutput(), filePath, false, false, false, false)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestPerformanceWithLargeFile removed - synthetic stress tests are premature
// TODO: Replace with real-world project testing when grammar is more complete

// TestStrictMode tests strict mode behavior
func TestStrictMode(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file with type errors
	errorFile := filepath.Join(tempDir, "errors.pl")
	content := `use strict;
my Int $count = "hello";
my Str $name = [];`

	err := os.WriteFile(errorFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test non-strict mode (should not exit with error)
	errorCount, err := checkFile(ui.NewDefaultOutput(), errorFile, false, false, false, false)
	if err != nil {
		t.Errorf("Non-strict mode should not return error: %v", err)
	}

	if errorCount == 0 {
		t.Error("Expected to find type errors")
	}

	// Test strict mode - we can't directly test os.Exit, but we can verify
	// that the error count is properly returned for strict mode logic
	errorCountStrict, err := checkFile(ui.NewDefaultOutput(), errorFile, true, false, false, false)
	if err != nil {
		t.Errorf("Strict mode should not return error from checkFile: %v", err)
	}

	if errorCountStrict != errorCount {
		t.Errorf("Strict mode should return same error count: expected %d, got %d",
			errorCount, errorCountStrict)
	}
}

// TestVerboseOutput tests verbose mode output
func TestVerboseOutput(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "verbose_test.pl")
	content := `use strict;
my Int $count = 42;
my Str $name = "hello";`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test verbose mode - we can't easily capture stdout in unit tests,
	// but we can verify the function completes without error
	errorCount, err := checkFile(ui.NewDefaultOutput(), testFile, false, true, false, false)
	if err != nil {
		t.Errorf("Verbose mode failed: %v", err)
	}

	if errorCount != 0 {
		t.Errorf("Expected no errors in valid file, got %d", errorCount)
	}
}

// TestFileExtensionFiltering tests that only Perl files are checked
func TestFileExtensionFiltering(t *testing.T) {
	tempDir := t.TempDir()

	// Create files with different extensions
	files := map[string]string{
		"script.pl":    "my Int $x = 42;",
		"module.pm":    "my Int $x = 42;",
		"test.t":       "my Int $x = 42;",
		"readme.txt":   "This is a text file",
		"config.json":  `{"key": "value"}`,
		"script.py":    "x = 42",
		"no_extension": "my Int $x = 42;",
	}

	for filename, content := range files {
		fullPath := filepath.Join(tempDir, filename)
		err := os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	// Test recursive checking to see which files are processed
	totalFiles, totalErrors, err := checkDirectory(ui.NewDefaultOutput(), tempDir, false, false, false)
	if err != nil {
		t.Fatalf("checkDirectory failed: %v", err)
	}

	// Should only check .pl, .pm, and .t files (3 files)
	expectedFiles := 3
	if totalFiles != expectedFiles {
		t.Errorf("Expected to check %d Perl files, checked %d", expectedFiles, totalFiles)
	}

	if totalErrors != 0 {
		t.Errorf("Expected no errors in valid Perl files, found %d", totalErrors)
	}
}
