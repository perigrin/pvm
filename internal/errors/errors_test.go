// ABOUTME: Tests for the structured Error type in the errors package
// ABOUTME: Validates error code formatting, category embedding, Is/As unwrapping, and hint text rendering

package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorFormatting(t *testing.T) {
	// Create a basic error
	err := New(PrefixPVM, CategoryVersion, "001", "Failed to detect version", nil)

	// Test basic error format
	errString := err.Error()

	// Verify error code format
	if !strings.Contains(errString, "PVM-001") {
		t.Errorf("Expected error code PVM-001, got: %s", errString)
	}

	// Verify message
	if !strings.Contains(errString, "Failed to detect version") {
		t.Errorf("Expected message 'Failed to detect version', got: %s", errString)
	}

	// Verify category
	if !strings.Contains(errString, CategoryVersion) {
		t.Errorf("Expected category '%s', got: %s", CategoryVersion, errString)
	}
}

func TestErrorChaining(t *testing.T) {
	// Create an inner error
	innerErr := errors.New("inner error")

	// Create an error that wraps the inner error
	err := New(PrefixCFG, CategoryConfig, "001", "Configuration error", innerErr)

	// Test error unwrapping
	unwrapped := errors.Unwrap(err)
	if unwrapped != innerErr {
		t.Errorf("Expected unwrapped error to be inner error, got: %v", unwrapped)
	}

	// Test error cause inclusion in the error string
	errString := err.Error()
	if !strings.Contains(errString, "inner error") {
		t.Errorf("Expected error string to contain inner error, got: %s", errString)
	}
}

func TestErrorAddContext(t *testing.T) {
	// Create a basic error
	err := New(PrefixPVM, CategoryVersion, "001", "Failed to detect version", nil)

	// Add context information
	err = err.WithDetail("Unable to parse version string")
	err = err.WithLocation("version.go:123")
	err = err.WithHint("Check version format in configuration")

	// Test context inclusion in error string
	errString := err.Error()

	// Verify detail
	if !strings.Contains(errString, "Detail: Unable to parse version string") {
		t.Errorf("Expected detail information, got: %s", errString)
	}

	// Verify location
	if !strings.Contains(errString, "Location: version.go:123") {
		t.Errorf("Expected location information, got: %s", errString)
	}

	// Verify hint
	if !strings.Contains(errString, "Hint: Check version format in configuration") {
		t.Errorf("Expected hint information, got: %s", errString)
	}
}

func TestErrorWrap(t *testing.T) {
	// Create an inner error
	innerErr := errors.New("inner error")

	// Wrap the error
	err := Wrap(innerErr, PrefixPVI, CategoryModule, "002", "Failed to install module")

	// Verify the wrapped error
	if err.Unwrap() != innerErr {
		t.Errorf("Expected inner error to be preserved, got: %v", err.Unwrap())
	}

	if err.Prefix() != PrefixPVI {
		t.Errorf("Expected prefix %s, got: %s", PrefixPVI, err.Prefix())
	}

	if err.Category() != CategoryModule {
		t.Errorf("Expected category %s, got: %s", CategoryModule, err.Category())
	}

	if err.Code() != "002" {
		t.Errorf("Expected code 002, got: %s", err.Code())
	}

	if err.Message() != "Failed to install module" {
		t.Errorf("Expected message 'Failed to install module', got: %s", err.Message())
	}
}

func TestErrorHelperFunctions(t *testing.T) {
	// Test helper functions for creating specific error types
	testCases := []struct {
		name     string
		err      *Error
		prefix   string
		category string
	}{
		{
			name:     "ConfigError",
			err:      NewConfigError("001", "Config error", nil),
			prefix:   PrefixCFG,
			category: CategoryConfig,
		},
		{
			name:     "VersionError",
			err:      NewVersionError("001", "Version error", nil),
			prefix:   PrefixPVM,
			category: CategoryVersion,
		},
		{
			name:     "ModuleError",
			err:      NewModuleError("001", "Module error", nil),
			prefix:   PrefixPVI,
			category: CategoryModule,
		},
		{
			name:     "ExecutionError",
			err:      NewExecutionError("001", "Execution error", nil),
			prefix:   PrefixPVX,
			category: CategoryExecution,
		},
		{
			name:     "TypeError",
			err:      NewTypeError("001", "Type error", nil),
			prefix:   PrefixPSC,
			category: CategoryType,
		},
		{
			name:     "SystemError",
			err:      NewSystemError("001", "System error", nil),
			prefix:   PrefixSYS,
			category: CategorySystem,
		},
		{
			name:     "UserInputError",
			err:      NewUserInputError(PrefixPVM, "001", "User input error", nil),
			prefix:   PrefixPVM,
			category: CategoryUserInput,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Prefix() != tc.prefix {
				t.Errorf("Expected prefix %s, got: %s", tc.prefix, tc.err.Prefix())
			}

			if tc.err.Category() != tc.category {
				t.Errorf("Expected category %s, got: %s", tc.category, tc.err.Category())
			}
		})
	}
}

func TestErrorIs(t *testing.T) {
	// Create errors with the same code and prefix but different messages
	err1 := New(PrefixPVM, CategoryVersion, "001", "Error one", nil)
	err2 := New(PrefixPVM, CategoryVersion, "001", "Error two", nil)

	// Create an error with a different code
	err3 := New(PrefixPVM, CategoryVersion, "002", "Error three", nil)

	// Test error matching
	if !errors.Is(err1, err2) {
		t.Error("Expected errors with same code and prefix to match")
	}

	if errors.Is(err1, err3) {
		t.Error("Expected errors with different codes not to match")
	}
}
