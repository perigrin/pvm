// ABOUTME: Unit tests for binary installation validation functionality
// ABOUTME: Tests validation logic for binary Perl installations

package perl

import (
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateBinaryInstallation(t *testing.T) {
	// Skip on Windows: these tests create fake perl executables with shell script
	// content that Windows cannot execute (invalid PE format)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping binary validation tests on Windows (cannot create fake executables)")
	}
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pvm-validate-test")
	if err != nil {
		t.Skip("Cannot create temp directory:", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		setupFunc   func(string) error
		expectValid bool
		expectError bool
	}{
		{
			name: "valid installation",
			setupFunc: func(installDir string) error {
				// Create proper directory structure
				dirs := []string{"bin", "lib", "man", "share"}
				for _, dir := range dirs {
					err := os.MkdirAll(filepath.Join(installDir, dir), 0755)
					if err != nil {
						return err
					}
				}

				// Create perl executable
				perlPath := filepath.Join(installDir, "bin", "perl")
				if runtime.GOOS == "windows" {
					perlPath = filepath.Join(installDir, "bin", "perl.exe")
				}
				return os.WriteFile(perlPath, []byte("#!/usr/bin/perl\n"), 0755)
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "missing installation directory",
			setupFunc: func(installDir string) error {
				// Don't create the directory
				return nil
			},
			expectValid: false,
			expectError: true,
		},
		{
			name: "missing perl executable",
			setupFunc: func(installDir string) error {
				// Create directory structure but no perl executable
				return os.MkdirAll(filepath.Join(installDir, "bin"), 0755)
			},
			expectValid: false,
			expectError: true,
		},
		{
			name: "incomplete installation",
			setupFunc: func(installDir string) error {
				// Create only bin directory with perl
				err := os.MkdirAll(filepath.Join(installDir, "bin"), 0755)
				if err != nil {
					return err
				}

				perlPath := filepath.Join(installDir, "bin", "perl")
				if runtime.GOOS == "windows" {
					perlPath = filepath.Join(installDir, "bin", "perl.exe")
				}
				return os.WriteFile(perlPath, []byte("#!/usr/bin/perl\n"), 0755)
			},
			expectValid: true, // Should pass due to having bin/perl, score will be ~0.52 (dir:0.4 + exec:0.4*0.7 + version:0.3*0.5)
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installDir := filepath.Join(tmpDir, tt.name)

			// Setup test environment
			err := tt.setupFunc(installDir)
			if err != nil {
				t.Fatalf("Failed to setup test: %v", err)
			}

			// Test validation
			valid, warnings, err := ValidateBinaryInstallation(installDir)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if valid != tt.expectValid {
					t.Errorf("Expected valid=%v, got %v", tt.expectValid, valid)
				}

				// Log warnings for debugging
				if len(warnings) > 0 {
					t.Logf("Warnings: %v", warnings)
				}
			}
		})
	}
}

func TestValidateBinaryInstallationDetailed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping binary validation tests on Windows (cannot create fake executables)")
	}
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pvm-validate-detailed-test")
	if err != nil {
		t.Skip("Cannot create temp directory:", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a complete installation
	installDir := filepath.Join(tmpDir, "complete")
	err = createCompleteTestInstallation(installDir)
	if err != nil {
		t.Fatalf("Failed to create test installation: %v", err)
	}

	// Test detailed validation
	result, err := ValidateBinaryInstallationDetailed(installDir, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("Expected installation to be valid")
	}

	if result.CompletenessScore <= 0 {
		t.Error("Expected positive completeness score")
	}

	if result.PerlExecutable == "" {
		t.Error("Expected perl executable path to be set")
	}

	t.Logf("Completeness score: %.2f", result.CompletenessScore)
	t.Logf("Perl executable: %s", result.PerlExecutable)
	if len(result.Warnings) > 0 {
		t.Logf("Warnings: %v", result.Warnings)
	}
}

func TestValidateDirectoryStructure(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pvm-validate-dirs-test")
	if err != nil {
		t.Skip("Cannot create temp directory:", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name           string
		dirs           []string
		expectedScore  float64
		expectWarnings bool
	}{
		{
			name:           "complete structure",
			dirs:           []string{"bin", "lib", "man", "share"},
			expectedScore:  1.0, // 0.4 + 0.3 + 0.15 + 0.15 = 1.0
			expectWarnings: false,
		},
		{
			name:           "minimal structure",
			dirs:           []string{"bin", "lib"},
			expectedScore:  0.7,  // bin(0.4) + lib(0.3)
			expectWarnings: true, // Missing man and share
		},
		{
			name:           "bin only",
			dirs:           []string{"bin"},
			expectedScore:  0.4,
			expectWarnings: true,
		},
		{
			name:           "no directories",
			dirs:           []string{},
			expectedScore:  0.0,
			expectWarnings: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			// Create specified directories
			for _, dir := range tt.dirs {
				err := os.MkdirAll(filepath.Join(testDir, dir), 0755)
				if err != nil {
					t.Fatalf("Failed to create directory %s: %v", dir, err)
				}
			}

			// Test validation
			score, warnings := validateDirectoryStructure(testDir)

			// Use approximate comparison for floating point values
			if math.Abs(score-tt.expectedScore) > 0.01 {
				t.Errorf("Expected score %.2f, got %.2f", tt.expectedScore, score)
				t.Logf("Directories created: %v", tt.dirs)
			}

			hasWarnings := len(warnings) > 0
			if hasWarnings != tt.expectWarnings {
				t.Errorf("Expected warnings=%v, got %v warnings: %v", tt.expectWarnings, hasWarnings, warnings)
			}

			t.Logf("Score: %.1f, Warnings: %v", score, warnings)
		})
	}
}

func TestValidatePerlExecutable(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pvm-validate-exe-test")
	if err != nil {
		t.Skip("Cannot create temp directory:", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		setupFunc   func(string) error
		expectError bool
		minScore    float64
	}{
		{
			name: "valid executable",
			setupFunc: func(path string) error {
				return os.WriteFile(path, []byte("#!/usr/bin/perl\n"), 0755)
			},
			expectError: false,
			minScore:    1.0,
		},
		{
			name: "non-executable file",
			setupFunc: func(path string) error {
				err := os.WriteFile(path, []byte("#!/usr/bin/perl\n"), 0644)
				if err != nil {
					return err
				}
				// On Windows, this test doesn't apply
				if runtime.GOOS == "windows" {
					return os.Chmod(path, 0755) // Make it executable on Windows
				}
				return nil
			},
			expectError: runtime.GOOS != "windows", // Should fail on Unix, pass on Windows
			minScore:    0.0,
		},
		{
			name: "missing file",
			setupFunc: func(path string) error {
				// Don't create the file
				return nil
			},
			expectError: true,
			minScore:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perlPath := filepath.Join(tmpDir, tt.name)
			if runtime.GOOS == "windows" {
				perlPath += ".exe"
			}

			// Setup test
			err := tt.setupFunc(perlPath)
			if err != nil {
				t.Fatalf("Failed to setup test: %v", err)
			}

			// Test validation
			score, warnings, err := validatePerlExecutable(perlPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if score < tt.minScore {
					t.Errorf("Expected score >= %.1f, got %.1f", tt.minScore, score)
				}
			}

			t.Logf("Score: %.1f, Warnings: %v", score, warnings)
		})
	}
}

func TestExtractVersionFromOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name: "perl 5.38.0",
			output: `This is perl 5, version 38, subversion 0 (v5.38.0) built for x86_64-linux
Copyright 1987-2023, Larry Wall`,
			expected: "5.38.0",
		},
		{
			name: "perl with v prefix",
			output: `This is perl, v5.36.0 built for darwin-thread-multi
Copyright 1987-2022, Larry Wall`,
			expected: "5.36.0",
		},
		{
			name: "version in different format",
			output: `perl version 5.34.1 built for linux-gnu
This is free software`,
			expected: "5.34.1",
		},
		{
			name:     "no version found",
			output:   "Some other output without version",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersionFromOutput(tt.output)
			if result != tt.expected {
				t.Errorf("Expected version '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestValidateExpectedVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping binary validation tests on Windows (cannot create fake executables)")
	}
	// This test requires a real Perl installation to work properly
	// For now, we'll just test the error cases

	tmpDir, err := os.MkdirTemp("", "pvm-validate-version-test")
	if err != nil {
		t.Skip("Cannot create temp directory:", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with non-existent installation
	err = ValidateExpectedVersion(filepath.Join(tmpDir, "nonexistent"), "5.38.0")
	if err == nil {
		t.Error("Expected error for non-existent installation")
	}

	// Test with invalid expected version
	installDir := filepath.Join(tmpDir, "invalid-version")
	err = createCompleteTestInstallation(installDir)
	if err != nil {
		t.Fatalf("Failed to create test installation: %v", err)
	}

	err = ValidateExpectedVersion(installDir, "invalid.version")
	if err == nil {
		t.Error("Expected error for invalid expected version")
	}
	// Accept either version format error or validation failure (since our test perl doesn't return version)
	if !strings.Contains(err.Error(), "Invalid expected version format") &&
		!strings.Contains(err.Error(), "Could not detect Perl version") {
		t.Errorf("Expected version format or detection error, got: %v", err)
	}
}

// Helper function to create a complete test installation
func createCompleteTestInstallation(installDir string) error {
	// Create directory structure
	dirs := []string{"bin", "lib", "man", "share"}
	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(installDir, dir), 0755)
		if err != nil {
			return err
		}
	}

	// Create perl executable
	perlPath := filepath.Join(installDir, "bin", "perl")
	if runtime.GOOS == "windows" {
		perlPath = filepath.Join(installDir, "bin", "perl.exe")
	}

	return os.WriteFile(perlPath, []byte("#!/usr/bin/perl\n"), 0755)
}
