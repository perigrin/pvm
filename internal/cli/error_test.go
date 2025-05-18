package cli

import (
	"errors"
	"strings"
	"testing"
)

func TestNewError(t *testing.T) {
	prefix := PrefixPVM
	category := CategoryConfig
	code := "001"
	message := "Test error"
	innerErr := errors.New("inner error")
	
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

func TestErrorWithChain(t *testing.T) {
	innerErr := errors.New("inner error")
	err := NewError(PrefixPVM, CategoryConfig, "001", "Test error", innerErr)
	
	// Test WithDetail
	err = err.WithDetail("Error details")
	if err.Detail != "Error details" {
		t.Errorf("Expected detail 'Error details', got %s", err.Detail)
	}
	
	// Test WithLocation
	err = err.WithLocation("config.go:123")
	if err.Location != "config.go:123" {
		t.Errorf("Expected location 'config.go:123', got %s", err.Location)
	}
	
	// Test WithHint
	err = err.WithHint("Check your config file")
	if err.Hint != "Check your config file" {
		t.Errorf("Expected hint 'Check your config file', got %s", err.Hint)
	}
}

func TestErrorOutput(t *testing.T) {
	innerErr := errors.New("inner error")
	err := NewError(PrefixPVM, CategoryConfig, "001", "Test error", innerErr)
	err = err.WithDetail("Error details").WithLocation("config.go:123").WithHint("Check your config file")
	
	// Test error string
	errStr := err.Error()
	
	mustContain := []string{
		"PVM-001: Test error",
		"Detail: Error details",
		"Location: config.go:123",
		"Hint: Check your config file",
	}
	
	for _, s := range mustContain {
		if !strings.Contains(errStr, s) {
			t.Errorf("Expected error string to contain '%s', but it doesn't: %s", s, errStr)
		}
	}
	
	// In non-verbose mode, inner error should not be included
	if strings.Contains(errStr, "inner error") {
		t.Errorf("Error string should not contain inner error in non-verbose mode: %s", errStr)
	}
	
	// Set verbose mode and check again
	Verbose = true
	errStr = err.Error()
	if !strings.Contains(errStr, "inner error") {
		t.Errorf("Error string should contain inner error in verbose mode: %s", errStr)
	}
	
	// Reset verbose mode
	Verbose = false
}

func TestErrorUnwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := NewError(PrefixPVM, CategoryConfig, "001", "Test error", innerErr)
	
	unwrapped := errors.Unwrap(err)
	if unwrapped != innerErr {
		t.Errorf("Expected unwrapped error to be %v, got %v", innerErr, unwrapped)
	}
}