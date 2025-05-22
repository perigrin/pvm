// ABOUTME: Tests for enhanced error formatting functionality
// ABOUTME: Covers context lines, suggestions, and various error scenarios

package psc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/typechecker"
)

func TestNewErrorFormatter(t *testing.T) {
	formatter := NewErrorFormatter()

	if formatter == nil {
		t.Fatal("NewErrorFormatter() returned nil")
	}

	if !formatter.showContext {
		t.Error("Expected showContext to be true by default")
	}

	if formatter.contextLines != 2 {
		t.Errorf("Expected contextLines to be 2, got %d", formatter.contextLines)
	}

	if !formatter.colorEnabled {
		t.Error("Expected colorEnabled to be true by default")
	}

	if formatter.sourceCodeCache == nil {
		t.Error("Expected sourceCodeCache to be initialized")
	}
}

func TestSetContextLines(t *testing.T) {
	formatter := NewErrorFormatter()
	formatter.SetContextLines(5)

	if formatter.contextLines != 5 {
		t.Errorf("Expected contextLines to be 5, got %d", formatter.contextLines)
	}
}

func TestSetColorEnabled(t *testing.T) {
	formatter := NewErrorFormatter()

	formatter.SetColorEnabled(false)
	if formatter.colorEnabled {
		t.Error("Expected colorEnabled to be false")
	}

	formatter.SetColorEnabled(true)
	if !formatter.colorEnabled {
		t.Error("Expected colorEnabled to be true")
	}
}

func TestGenerateSuggestion(t *testing.T) {
	formatter := NewErrorFormatter()

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "Int to Str mismatch",
			message:  "Type mismatch: expected Int, got Str",
			expected: "Consider using numeric conversion: int($value) or 0 + $value",
		},
		{
			name:     "Str to Int mismatch",
			message:  "Type mismatch: expected Str, got Int",
			expected: "Consider using string interpolation: \"$value\" or explicit conversion",
		},
		{
			name:     "General type mismatch",
			message:  "Type mismatch: expected Bool, got ArrayRef",
			expected: "Check that the assigned value matches the declared type",
		},
		{
			name:     "Undefined variable",
			message:  "Variable undefined or not found",
			expected: "Make sure the variable is declared before use, or check for typos",
		},
		{
			name:     "Type annotation error",
			message:  "Invalid type annotation syntax",
			expected: "Review the type annotation syntax: my TypeName $variable = value;",
		},
		{
			name:     "Assignment error",
			message:  "Assignment type compatibility error",
			expected: "Ensure the right side of the assignment is compatible with the declared type",
		},
		{
			name:     "Unknown error",
			message:  "Some other type error",
			expected: "Review the Typed Perl documentation for more information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &typechecker.TypeCheckError{Message: tt.message}
			suggestion := formatter.generateSuggestion(err)

			if suggestion != tt.expected {
				t.Errorf("Expected suggestion %q, got %q", tt.expected, suggestion)
			}
		})
	}
}

func TestFormatSeverity(t *testing.T) {
	tests := []struct {
		name         string
		severity     ErrorSeverity
		colorEnabled bool
		expectedText string
		hasColor     bool
	}{
		{
			name:         "Error without color",
			severity:     SeverityError,
			colorEnabled: false,
			expectedText: "error:",
			hasColor:     false,
		},
		{
			name:         "Error with color",
			severity:     SeverityError,
			colorEnabled: true,
			expectedText: "error:",
			hasColor:     true,
		},
		{
			name:         "Warning without color",
			severity:     SeverityWarning,
			colorEnabled: false,
			expectedText: "warning:",
			hasColor:     false,
		},
		{
			name:         "Warning with color",
			severity:     SeverityWarning,
			colorEnabled: true,
			expectedText: "warning:",
			hasColor:     true,
		},
		{
			name:         "Info without color",
			severity:     SeverityInfo,
			colorEnabled: false,
			expectedText: "info:",
			hasColor:     false,
		},
		{
			name:         "Info with color",
			severity:     SeverityInfo,
			colorEnabled: true,
			expectedText: "info:",
			hasColor:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewErrorFormatter()
			formatter.SetColorEnabled(tt.colorEnabled)

			result := formatter.formatSeverity(tt.severity)

			// Check if the text is contained
			if !strings.Contains(result, tt.expectedText) {
				t.Errorf("Expected result to contain %q, got %q", tt.expectedText, result)
			}

			// Check for ANSI color codes
			hasAnsiCodes := strings.Contains(result, "\033[")
			if tt.hasColor && !hasAnsiCodes {
				t.Error("Expected ANSI color codes but found none")
			}
			if !tt.hasColor && hasAnsiCodes {
				t.Error("Expected no ANSI color codes but found some")
			}
		})
	}
}

func TestColorize(t *testing.T) {
	formatter := NewErrorFormatter()

	tests := []struct {
		name         string
		text         string
		color        string
		colorEnabled bool
		expectColor  bool
	}{
		{
			name:         "Red text with color enabled",
			text:         "error message",
			color:        "red",
			colorEnabled: true,
			expectColor:  true,
		},
		{
			name:         "Text with color disabled",
			text:         "error message",
			color:        "red",
			colorEnabled: false,
			expectColor:  false,
		},
		{
			name:         "Unknown color",
			text:         "error message",
			color:        "purple",
			colorEnabled: true,
			expectColor:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter.SetColorEnabled(tt.colorEnabled)
			result := formatter.colorize(tt.text, tt.color)

			// Check if original text is contained
			if !strings.Contains(result, tt.text) {
				t.Errorf("Expected result to contain original text %q", tt.text)
			}

			// Check for color codes
			hasColor := strings.Contains(result, "\033[")
			if tt.expectColor && !hasColor {
				t.Error("Expected color codes but found none")
			}
			if !tt.expectColor && hasColor {
				t.Error("Expected no color codes but found some")
			}
		})
	}
}

func TestGetContextLines(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.pl")

	content := `#!/usr/bin/perl
use strict;
use warnings;

my Int $count = "hello";
my Str $name = 42;
print "Done\n";`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	formatter := NewErrorFormatter()
	formatter.SetContextLines(2)

	tests := []struct {
		name        string
		errorLine   int
		expectLines int
		expectError bool
	}{
		{
			name:        "Valid line with context",
			errorLine:   5, // my Int $count = "hello";
			expectLines: 5, // 2 lines before + error line + 2 lines after
			expectError: false,
		},
		{
			name:        "First line",
			errorLine:   1,
			expectLines: 3, // error line + 2 lines after
			expectError: false,
		},
		{
			name:        "Last line",
			errorLine:   7,
			expectLines: 3, // 2 lines before + error line
			expectError: false,
		},
		{
			name:        "Invalid line number",
			errorLine:   0,
			expectLines: 0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := formatter.getContextLines(testFile, tt.errorLine)

			if tt.expectLines == 0 {
				if len(lines) != 0 {
					t.Errorf("Expected no lines, got %d", len(lines))
				}
				return
			}

			if len(lines) != tt.expectLines {
				t.Errorf("Expected %d lines, got %d", tt.expectLines, len(lines))
			}

			// Check that error line is marked
			if tt.errorLine > 0 {
				found := false
				for _, line := range lines {
					if strings.HasPrefix(line, ">> ") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected to find error line marker '>>'")
				}
			}
		})
	}
}

func TestFormatError(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.pl")

	content := `use strict;
my Int $count = "hello";
print "Done\n";`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	formatter := NewErrorFormatter()
	formatter.SetColorEnabled(false) // Disable color for easier testing

	typeErr := &typechecker.TypeCheckError{
		Message: "Type mismatch: expected Int, got Str",
		Path:    testFile,
		Line:    2,
		Column:  17,
	}

	result := formatter.FormatError(typeErr)

	// Check that the result contains expected components
	expectedComponents := []string{
		testFile + ":2:17:",
		"error:",
		"Type mismatch: expected Int, got Str",
		">> ",
		"help:",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(result, component) {
			t.Errorf("Expected result to contain %q, got:\n%s", component, result)
		}
	}
}

func TestFormatErrors(t *testing.T) {
	formatter := NewErrorFormatter()
	formatter.SetColorEnabled(false)

	// Test empty errors
	result := formatter.FormatErrors([]typechecker.TypeCheckError{})
	if result != "" {
		t.Errorf("Expected empty string for no errors, got %q", result)
	}

	// Test multiple errors
	errors := []typechecker.TypeCheckError{
		{
			Message: "First error",
			Path:    "test1.pl",
			Line:    1,
			Column:  1,
		},
		{
			Message: "Second error",
			Path:    "test2.pl",
			Line:    2,
			Column:  2,
		},
	}

	result = formatter.FormatErrors(errors)

	// Should contain both errors
	if !strings.Contains(result, "First error") {
		t.Error("Expected result to contain first error")
	}
	if !strings.Contains(result, "Second error") {
		t.Error("Expected result to contain second error")
	}

	// Should be separated by newlines
	if !strings.Contains(result, "\n") {
		t.Error("Expected errors to be separated by newlines")
	}
}

func TestSourceCodeCaching(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.pl")

	content := `use strict;
my Int $count = 42;
print "Done\n";`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	formatter := NewErrorFormatter()

	// First call should read from file
	lines1 := formatter.getCachedSourceLines(testFile)
	if lines1 == nil {
		t.Fatal("Expected to get source lines")
	}

	// Second call should use cache
	lines2 := formatter.getCachedSourceLines(testFile)
	if lines2 == nil {
		t.Fatal("Expected to get cached source lines")
	}

	// Should be the same instance
	if &lines1[0] != &lines2[0] {
		t.Error("Expected cached lines to be the same instance")
	}

	// Check cache contents
	if _, exists := formatter.sourceCodeCache[testFile]; !exists {
		t.Error("Expected file to be in cache")
	}
}

func TestNonExistentFile(t *testing.T) {
	formatter := NewErrorFormatter()

	lines := formatter.getCachedSourceLines("/non/existent/file.pl")
	if lines != nil {
		t.Error("Expected nil for non-existent file")
	}

	// Should not be in cache
	if _, exists := formatter.sourceCodeCache["/non/existent/file.pl"]; exists {
		t.Error("Expected non-existent file not to be in cache")
	}
}
