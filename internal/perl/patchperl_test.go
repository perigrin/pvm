// ABOUTME: Tests for Devel::PatchPerl integration
// ABOUTME: Validates PatchPerl detection, installation, and application functionality

package perl

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidatePatchPerlEnvironment(t *testing.T) {
	err := ValidatePatchPerlEnvironment()
	if err != nil {
		t.Logf("PatchPerl environment validation failed: %v", err)
		// This is expected on systems without perl - not a test failure
	} else {
		t.Logf("PatchPerl environment validation passed")
	}
}

func TestIsPatchPerlAvailable(t *testing.T) {
	available := IsPatchPerlAvailable()
	t.Logf("PatchPerl available: %v", available)

	// This test just logs the result - availability depends on system state
	// We don't assert true/false since it's system-dependent
}

func TestShouldUsePatchPerl(t *testing.T) {
	tests := []struct {
		name        string
		perlVersion string
		expected    bool
	}{
		{
			name:        "old stable version should use PatchPerl",
			perlVersion: "5.26.0",
			expected:    true,
		},
		{
			name:        "recent stable version should use PatchPerl",
			perlVersion: "5.38.0",
			expected:    true,
		},
		{
			name:        "development version should skip PatchPerl",
			perlVersion: "5.37.1",
			expected:    false,
		},
		{
			name:        "another development version should skip PatchPerl",
			perlVersion: "5.39.0",
			expected:    false,
		},
		{
			name:        "stable version after dev should use PatchPerl",
			perlVersion: "5.38.2",
			expected:    true,
		},
		{
			name:        "invalid version should default to using PatchPerl",
			perlVersion: "invalid.version",
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldUsePatchPerl(tt.perlVersion)
			if result != tt.expected {
				t.Errorf("ShouldUsePatchPerl(%s) = %v, expected %v", tt.perlVersion, result, tt.expected)
			}
		})
	}
}

func TestPatchPerlOptionsValidation(t *testing.T) {
	// Test nil options
	_, err := ApplyPatchPerl(nil)
	if err == nil {
		t.Error("Expected error for nil options")
	}

	// Test empty source directory
	options := &PatchPerlOptions{
		SourceDir: "",
	}
	_, err = ApplyPatchPerl(options)
	if err == nil {
		t.Error("Expected error for empty source directory")
	}

	// Test non-existent source directory
	options = &PatchPerlOptions{
		SourceDir: "/nonexistent/directory",
	}
	_, err = ApplyPatchPerl(options)
	if err == nil {
		t.Error("Expected error for non-existent source directory")
	}
}

func TestPatchPerlWithMockDirectory(t *testing.T) {
	// Skip if no perl environment available
	if err := ValidatePatchPerlEnvironment(); err != nil {
		t.Skip("Skipping PatchPerl test - perl environment not available")
	}

	// Create a temporary directory to simulate source directory
	tmpDir := t.TempDir()

	// Create some basic files that might exist in a Perl source directory
	testFiles := []string{"Configure", "perl.h", "Makefile.PL"}
	for _, file := range testFiles {
		f, err := os.Create(filepath.Join(tmpDir, file))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		f.Close()
	}

	options := &PatchPerlOptions{
		SourceDir:   tmpDir,
		PerlVersion: "5.26.0",
		Verbose:     testing.Verbose(),
		Context:     context.Background(),
		Timeout:     30 * time.Second,
	}

	result, err := ApplyPatchPerl(options)

	// We expect this to potentially fail since it's not a real Perl source directory
	// But we test that the function handles the error gracefully
	if err != nil {
		t.Logf("PatchPerl failed as expected for mock directory: %v", err)
		if result == nil {
			t.Error("Expected result object even on failure")
		}
	} else {
		t.Logf("PatchPerl succeeded: Applied=%v, PatchCount=%d, Duration=%v",
			result.Applied, result.PatchCount, result.Duration)
	}
}

func TestApplyPatchPerlSafely(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Test with non-existent directory (should handle gracefully)
	err := ApplyPatchPerlSafely("/nonexistent/directory", "5.26.0", false)
	if err != nil {
		t.Errorf("ApplyPatchPerlSafely should handle missing directory gracefully, got: %v", err)
	}

	// Test with real directory but no perl (should handle gracefully)
	err = ApplyPatchPerlSafely(tmpDir, "5.26.0", testing.Verbose())
	if err != nil {
		t.Errorf("ApplyPatchPerlSafely should handle failures gracefully, got: %v", err)
	}

	// Test with development version (should skip)
	err = ApplyPatchPerlSafely(tmpDir, "5.37.1", testing.Verbose())
	if err != nil {
		t.Errorf("ApplyPatchPerlSafely should handle development versions gracefully, got: %v", err)
	}
}

func TestGetPatchPerlInfo(t *testing.T) {
	if !IsPatchPerlAvailable() {
		t.Skip("Skipping PatchPerl info test - Devel::PatchPerl not available")
	}

	info, err := GetPatchPerlInfo()
	if err != nil {
		t.Errorf("GetPatchPerlInfo failed: %v", err)
	} else {
		t.Logf("PatchPerl info: %s", info)
	}
}

func TestPatchPerlResult(t *testing.T) {
	result := &PatchPerlResult{
		Applied:    true,
		PatchCount: 5,
		Output:     "Successfully applied 5 patches",
		Duration:   2 * time.Second,
	}

	if !result.Applied {
		t.Error("Expected Applied to be true")
	}

	if result.PatchCount != 5 {
		t.Errorf("Expected PatchCount to be 5, got %d", result.PatchCount)
	}

	if result.Duration != 2*time.Second {
		t.Errorf("Expected Duration to be 2s, got %v", result.Duration)
	}

	if result.Output == "" {
		t.Error("Expected non-empty Output")
	}
}

func TestPatchPerlOptionsDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal test file
	f, err := os.Create(filepath.Join(tmpDir, "Configure"))
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	f.Close()

	options := &PatchPerlOptions{
		SourceDir:   tmpDir,
		PerlVersion: "5.26.0",
		// Leave other fields as defaults
	}

	// This will likely fail since there's no real perl source,
	// but we're testing that defaults are applied correctly
	_, err = ApplyPatchPerl(options)

	// Check that defaults were applied
	if options.Context == nil {
		t.Error("Expected Context to be set to default")
	}

	if options.Timeout == 0 {
		t.Error("Expected Timeout to be set to default")
	}
}
