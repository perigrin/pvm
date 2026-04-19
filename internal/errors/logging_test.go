// ABOUTME: Tests for structured error logging integration in the errors package
// ABOUTME: Verifies that LogError routes PVM Error types and standard errors to the correct log levels and formats

package errors

import (
	"bytes"
	stdErrors "errors"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/log"
)

func TestLogError(t *testing.T) {
	// Create a buffer to capture log output
	buf := new(bytes.Buffer)

	// Create a test logger (not used directly, but configures output)
	_ = log.NewLogger(log.LevelDebug, buf, "TEST")

	// Set as global logger
	log.SetGlobalOutput(buf)
	log.SetGlobalLevel(log.LevelDebug)
	log.SetGlobalComponent("TEST")

	// Test logging different error types

	// 1. Standard error
	stdErr := stdErrors.New("standard error")
	LogError(stdErr)

	// 2. Structured user input error
	userErr := NewUserInputError(PrefixPVM, "001", "user input error", nil)
	LogError(userErr)

	// 3. Structured system error
	sysErr := NewSystemError("002", "system error", nil)
	LogError(sysErr)

	// Check the output
	output := buf.String()

	// Verify standard error was logged
	if !strings.Contains(output, "standard error") {
		t.Error("Expected standard error to be logged")
	}

	// Verify user input error was logged at warning level
	if !strings.Contains(output, "[WARNING]") && !strings.Contains(output, "user input error") {
		t.Error("Expected user input error to be logged at warning level")
	}

	// Verify system error was logged at error level
	if !strings.Contains(output, "[ERROR]") && !strings.Contains(output, "system error") {
		t.Error("Expected system error to be logged at error level")
	}
}

func TestLogErrorWithLocation(t *testing.T) {
	// Create a buffer to capture log output
	buf := new(bytes.Buffer)

	// Set the logger
	log.SetGlobalOutput(buf)
	log.SetGlobalLevel(log.LevelDebug)

	// Create a structured error without a location
	err := New(PrefixPVM, CategoryVersion, "001", "Error without location", nil)

	// Log with location
	LogErrorWithLocation(err, "test_file.go:123")

	// Check the output
	output := buf.String()

	// Verify location was added
	if !strings.Contains(output, "test_file.go:123") {
		t.Error("Expected location to be added to error")
	}
}

func TestLogDebug(t *testing.T) {
	// Create a buffer to capture log output
	buf := new(bytes.Buffer)

	// Set the logger
	log.SetGlobalOutput(buf)
	log.SetGlobalLevel(log.LevelDebug)

	// Create a structured error
	err := New(PrefixPVM, CategoryVersion, "001", "Debug error", nil)

	// Log at debug level
	LogDebug(err)

	// Check the output
	output := buf.String()

	// Verify error was logged at debug level
	if !strings.Contains(output, "[DEBUG]") || !strings.Contains(output, "Debug error") {
		t.Error("Expected error to be logged at debug level")
	}
}
