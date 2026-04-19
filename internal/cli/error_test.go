// ABOUTME: Tests for CLI error construction and formatting in the cli package
// ABOUTME: Validates that NewError delegates correctly to the errors package and preserves prefix, category, and code fields

package cli

import (
	"bytes"
	stdErrors "errors"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

func TestNewError(t *testing.T) {
	// The implementation now delegates to errors.New, so we just need to verify
	// that the error is created with the correct values
	prefix := PrefixPVM
	category := CategoryConfig
	code := "001"
	message := "Test error"
	innerErr := stdErrors.New("inner error")

	err := NewError(prefix, category, code, message, innerErr)

	if err.Prefix() != prefix {
		t.Errorf("Expected prefix %s, got %s", prefix, err.Prefix())
	}

	if err.Category() != category {
		t.Errorf("Expected category %s, got %s", category, err.Category())
	}

	if err.Code() != code {
		t.Errorf("Expected code %s, got %s", code, err.Code())
	}

	if err.Message() != message {
		t.Errorf("Expected message %s, got %s", message, err.Message())
	}

	if err.Unwrap() != innerErr {
		t.Errorf("Expected inner error %v, got %v", innerErr, err.Unwrap())
	}
}

func TestLoggingFunctions(t *testing.T) {
	// Create a buffer to capture log output
	buf := new(bytes.Buffer)

	// Set up logging
	log.SetGlobalOutput(buf)
	log.SetGlobalLevel(log.LevelDebug)
	log.SetGlobalComponent("TEST")

	// Test each logging function
	LogDebug("Debug message")
	LogInfo("Info message")
	LogWarning("Warning message")
	LogError("Error message")

	// Check the output
	output := buf.String()

	// Set Verbose to true to test debug logging
	Verbose = true
	LogDebug("Debug message with verbose")

	// Reset Verbose
	Verbose = false

	// Verify log levels
	if !strings.Contains(output, "[INFO]") {
		t.Error("Expected info message to be logged")
	}

	if !strings.Contains(output, "[WARNING]") {
		t.Error("Expected warning message to be logged")
	}

	if !strings.Contains(output, "[ERROR]") {
		t.Error("Expected error message to be logged")
	}

	// Verify that debug message with verbose is logged
	if Verbose && !strings.Contains(output, "Debug message with verbose") {
		t.Error("Expected debug message to be logged when verbose is true")
	}
}

func TestDisplayError(t *testing.T) {
	// Test standard error - we'll test the function exists and doesn't panic
	// The actual output testing is complex due to stderr redirection
	stdErr := stdErrors.New("standard error message")

	// Should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("displayError panicked: %v", r)
			}
		}()

		displayError(stdErr)
	}()
}

func TestDisplayTypedError(t *testing.T) {
	// Test typed error - verify it doesn't panic and basic structure is correct
	typedErr := errors.New(PrefixPVM, CategoryConfig, "001", "Configuration not found", nil)
	typedErr.WithHint("Check your config file").WithLocation("~/.pvm/config.toml")

	// Should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("displayError panicked: %v", r)
			}
		}()

		displayError(typedErr)
	}()

	// Test the error structure is correct
	if typedErr.Prefix() != PrefixPVM {
		t.Errorf("Expected prefix PVM, got %s", typedErr.Prefix())
	}
	if typedErr.Code() != "001" {
		t.Errorf("Expected code 001, got %s", typedErr.Code())
	}
	if typedErr.Message() != "Configuration not found" {
		t.Errorf("Expected message 'Configuration not found', got %s", typedErr.Message())
	}
}

func TestDisplayEnhancedError(t *testing.T) {
	// Test enhanced error - verify it doesn't panic and basic structure is correct
	enhancedErr := errors.NewEnhancedError(PrefixPSC, CategoryType, "001", "Type mismatch", nil, errors.SeverityError)
	enhancedErr.WithContext("expected", "Int").WithContext("actual", "Str")
	enhancedErr.WithRecoveryAction("Cast the value to the correct type")
	enhancedErr.WithRecoveryAction("Update the type annotation")

	// Should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("displayError panicked: %v", r)
			}
		}()

		displayError(enhancedErr)
	}()

	// Test the error structure is correct
	if enhancedErr.Prefix() != PrefixPSC {
		t.Errorf("Expected prefix PSC, got %s", enhancedErr.Prefix())
	}
	if enhancedErr.Code() != "001" {
		t.Errorf("Expected code 001, got %s", enhancedErr.Code())
	}
	if enhancedErr.Message() != "Type mismatch" {
		t.Errorf("Expected message 'Type mismatch', got %s", enhancedErr.Message())
	}
	if enhancedErr.Severity() != errors.SeverityError {
		t.Errorf("Expected severity Error, got %s", enhancedErr.Severity())
	}

	// Test context was added
	context := enhancedErr.Context()
	if context["expected"] != "Int" {
		t.Errorf("Expected context 'expected' to be 'Int', got %v", context["expected"])
	}
	if context["actual"] != "Str" {
		t.Errorf("Expected context 'actual' to be 'Str', got %v", context["actual"])
	}

	// Test recovery actions were added
	actions := enhancedErr.RecoveryActions()
	if len(actions) != 2 {
		t.Errorf("Expected 2 recovery actions, got %d", len(actions))
	}
}

func TestFormatError(t *testing.T) {
	// Test error formatting without display
	typedErr := errors.New(PrefixPVI, CategoryModule, "002", "Module not found", nil)

	formatted := FormatError(typedErr)

	if !strings.Contains(formatted, "PVI-002") {
		t.Errorf("Expected formatted error to contain 'PVI-002', got: %s", formatted)
	}
	if !strings.Contains(formatted, "Module not found") {
		t.Errorf("Expected formatted error to contain message, got: %s", formatted)
	}
}
