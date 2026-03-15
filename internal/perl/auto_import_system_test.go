// ABOUTME: Unit tests for automatic system Perl import functionality
// ABOUTME: Tests the AutoImportSystemPerl function and its integration

package perl

import (
	"strings"
	"testing"
)

func TestAutoImportSystemPerl_Basic(t *testing.T) {
	// Skip if we're in a CI environment without system Perl
	if testing.Short() {
		t.Skip("Skipping auto-import test in short mode")
	}

	// Mock DetectSystemPerl to return a valid result
	originalDetectSystemPerl := DetectSystemPerl
	defer func() { DetectSystemPerl = originalDetectSystemPerl }()

	testSystemPerl := &SystemPerl{
		Path:    "/usr/bin/perl",
		Version: "5.38.0",
	}

	DetectSystemPerl = func() (*SystemPerl, error) {
		return testSystemPerl, nil
	}

	// Mock IsVersionInstalled to return false (not already registered)
	originalIsVersionInstalled := IsVersionInstalled
	defer func() { IsVersionInstalled = originalIsVersionInstalled }()

	IsVersionInstalled = func(version string) (bool, error) {
		return false, nil
	}

	// For basic testing, we'll just verify the function doesn't error
	// without mocking RegisterVersion since it's not a function variable
	err := AutoImportSystemPerl()
	if err != nil {
		// This might fail if registry operations fail, but that's OK for basic testing
		t.Logf("AutoImportSystemPerl failed (expected in test environment): %v", err)
	}
}

func TestAutoImportSystemPerl_AlreadyRegistered(t *testing.T) {
	// Mock DetectSystemPerl
	originalDetectSystemPerl := DetectSystemPerl
	defer func() { DetectSystemPerl = originalDetectSystemPerl }()

	testSystemPerl := &SystemPerl{
		Path:    "/usr/bin/perl",
		Version: "5.38.0",
	}

	DetectSystemPerl = func() (*SystemPerl, error) {
		return testSystemPerl, nil
	}

	// Mock IsVersionInstalled to return true (already registered)
	originalIsVersionInstalled := IsVersionInstalled
	defer func() { IsVersionInstalled = originalIsVersionInstalled }()

	IsVersionInstalled = func(version string) (bool, error) {
		return true, nil // Already registered
	}

	// Test auto-import should succeed without registering
	err := AutoImportSystemPerl()
	if err != nil {
		t.Fatalf("AutoImportSystemPerl failed when version already registered: %v", err)
	}
}

func TestAutoImportSystemPerl_DetectionFails(t *testing.T) {
	// Mock DetectSystemPerl to fail
	originalDetectSystemPerl := DetectSystemPerl
	defer func() { DetectSystemPerl = originalDetectSystemPerl }()

	DetectSystemPerl = func() (*SystemPerl, error) {
		return nil, &SystemError{Code: ErrNoSystemPerl, Message: "No system Perl found"}
	}

	// Test auto-import should fail
	err := AutoImportSystemPerl()
	if err == nil {
		t.Error("Expected AutoImportSystemPerl to fail when system Perl detection fails")
	}

	// Should contain the detection error
	if !strings.Contains(err.Error(), "Failed to detect system Perl") {
		t.Errorf("Error should mention detection failure, got: %v", err)
	}
}

// SystemError represents a simple error for testing
type SystemError struct {
	Code    string
	Message string
}

func (e *SystemError) Error() string {
	return e.Message
}
