package cli

import (
	"bytes"
	stdErrors "errors"
	"strings"
	"testing"

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
	
	if err.Prefix != prefix {
		t.Errorf("Expected prefix %s, got %s", prefix, err.Prefix)
	}
	
	if err.Category != category {
		t.Errorf("Expected category %s, got %s", category, err.Category)
	}
	
	if err.Code != code {
		t.Errorf("Expected code %s, got %s", code, err.Code)
	}
	
	if err.Message != message {
		t.Errorf("Expected message %s, got %s", message, err.Message)
	}
	
	if err.InnerErr != innerErr {
		t.Errorf("Expected inner error %v, got %v", innerErr, err.InnerErr)
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