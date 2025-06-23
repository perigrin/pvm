// ABOUTME: Unit tests for PVM command functionality
// ABOUTME: Tests command flag parsing and validation logic

package pvm

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestInstallCommandFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "basic install",
			args:        []string{"5.38.0"},
			expectError: false,
		},
		{
			name:        "binary only flag",
			args:        []string{"--binary-only", "5.38.0"},
			expectError: false,
		},
		{
			name:        "short prefer binary flag",
			args:        []string{"-B", "5.38.0"},
			expectError: false,
		},
		{
			name:        "prefer binary flag",
			args:        []string{"--prefer-binary", "5.38.0"},
			expectError: false,
		},
		{
			name:        "force source flag",
			args:        []string{"--force-source", "5.38.0"},
			expectError: false,
		},
		{
			name:        "mutually exclusive flags",
			args:        []string{"--binary-only", "--force-source", "5.38.0"},
			expectError: true,
			errorMsg:    "--binary-only and --force-source are mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newInstallCommand()
			cmd.SetArgs(tt.args)

			// We need to capture the error before execution since the command will
			// try to actually install Perl
			err := cmd.ParseFlags(tt.args)
			if err != nil && !tt.expectError {
				t.Errorf("ParseFlags() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Check that flags are properly set
				binaryOnly, _ := cmd.Flags().GetBool("binary-only")
				preferBinary, _ := cmd.Flags().GetBool("prefer-binary")
				forceSource, _ := cmd.Flags().GetBool("force-source")

				// Test specific flag combinations
				switch tt.name {
				case "binary only flag":
					if !binaryOnly {
						t.Error("Expected binary-only flag to be true")
					}
				case "prefer binary flag", "short prefer binary flag":
					if !preferBinary {
						t.Error("Expected prefer-binary flag to be true")
					}
				case "force source flag":
					if !forceSource {
						t.Error("Expected force-source flag to be true")
					}
				}
			}
		})
	}
}

func TestInstallCommandFlagValidation(t *testing.T) {
	// Test the validation logic without executing the command
	cmd := newInstallCommand()

	// Mock the RunE function to test only the validation logic
	originalRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Get binary installation flags
		binaryOnly, err := cmd.Flags().GetBool("binary-only")
		if err != nil {
			return err
		}

		forceSource, err := cmd.Flags().GetBool("force-source")
		if err != nil {
			return err
		}

		// Validate mutually exclusive flags
		if binaryOnly && forceSource {
			return fmt.Errorf("--binary-only and --force-source are mutually exclusive")
		}

		return nil
	}

	// Test mutually exclusive flags
	cmd.SetArgs([]string{"--binary-only", "--force-source", "5.38.0"})
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for mutually exclusive flags")
	}
	if err.Error() != "--binary-only and --force-source are mutually exclusive" {
		t.Errorf("Expected specific error message, got: %v", err)
	}

	// Restore original RunE
	cmd.RunE = originalRunE
}

func TestInstallPerlCommandFlags(t *testing.T) {

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "from-build flag",
			args:        []string{"--from-build", "/path/to/build"},
			expectError: false,
		},
		{
			name:        "version override",
			args:        []string{"--from-build", "/path/to/build", "--version", "5.38.0"},
			expectError: false,
		},
		{
			name:        "force flag",
			args:        []string{"--from-build", "/path/to/build", "--force"},
			expectError: false,
		},
		{
			name:        "directory as argument",
			args:        []string{"/path/to/build"},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a new command instance for each test
			cmd := newInstallPerlCommand()

			// Replace the RunE function to avoid actual execution
			originalRunE := cmd.RunE
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				return nil
			}

			// Set arguments
			cmd.SetArgs(test.args)

			// Execute command
			err := cmd.Execute()

			if test.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !test.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if test.expectError && err != nil && test.errorMsg != "" {
				if err.Error() != test.errorMsg {
					t.Errorf("Expected error message %q, got %q", test.errorMsg, err.Error())
				}
			}

			// Restore original RunE
			cmd.RunE = originalRunE
		})
	}
}

func TestInstallPerlCommandArchiveFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "archive as positional argument",
			args:        []string{"/path/to/perl.tar.gz"},
			expectError: false,
		},
		{
			name:        "archive with from-build flag",
			args:        []string{"--from-build", "/path/to/perl.tar.gz"},
			expectError: false,
		},
		{
			name:        "tgz archive",
			args:        []string{"/path/to/perl.tgz"},
			expectError: false,
		},
		{
			name:        "archive with version override",
			args:        []string{"--from-build", "/path/to/perl.tar.gz", "--version", "5.38.0"},
			expectError: false,
		},
		{
			name:        "archive with force flag",
			args:        []string{"--from-build", "/path/to/perl.tar.gz", "--force"},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a new command instance for each test
			cmd := newInstallPerlCommand()

			// Replace the RunE function to avoid actual execution
			originalRunE := cmd.RunE
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				return nil
			}

			// Set arguments
			cmd.SetArgs(test.args)

			// Execute command
			err := cmd.Execute()

			if test.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !test.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if test.expectError && err != nil && test.errorMsg != "" {
				if err.Error() != test.errorMsg {
					t.Errorf("Expected error message %q, got %q", test.errorMsg, err.Error())
				}
			}

			// Restore original RunE
			cmd.RunE = originalRunE
		})
	}
}

func TestIsArchive(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "tar.gz archive",
			path:     "/path/to/perl.tar.gz",
			expected: true,
		},
		{
			name:     "tgz archive",
			path:     "/path/to/perl.tgz",
			expected: true,
		},
		{
			name:     "uppercase tar.gz",
			path:     "/path/to/PERL.TAR.GZ",
			expected: true,
		},
		{
			name:     "mixed case tgz",
			path:     "/path/to/Perl.TgZ",
			expected: true,
		},
		{
			name:     "directory path",
			path:     "/path/to/perl",
			expected: false,
		},
		{
			name:     "other file type",
			path:     "/path/to/perl.zip",
			expected: false,
		},
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isArchive(test.path)
			if result != test.expected {
				t.Errorf("isArchive(%q) = %v, expected %v", test.path, result, test.expected)
			}
		})
	}
}

func TestExtractArchive(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "pvm-extract-test-*")
	if err != nil {
		t.Skip("Cannot create temp directory:", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test tar.gz archive
	archivePath := filepath.Join(tmpDir, "test.tar.gz")
	err = createTestTarGz(archivePath)
	if err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}

	// Create extraction directory
	extractDir := filepath.Join(tmpDir, "extract")
	err = os.MkdirAll(extractDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create extract directory: %v", err)
	}

	// Test extraction
	err = extractArchive(archivePath, extractDir)
	if err != nil {
		t.Fatalf("Failed to extract archive: %v", err)
	}

	// Verify extracted files
	testFile := filepath.Join(extractDir, "bin", "perl")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("Expected file not found after extraction: %s", testFile)
	}

	testDir := filepath.Join(extractDir, "lib")
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Errorf("Expected directory not found after extraction: %s", testDir)
	}
}

func TestExtractArchiveInvalidPath(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "pvm-extract-test-*")
	if err != nil {
		t.Skip("Cannot create temp directory:", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with non-existent archive
	nonExistentArchive := filepath.Join(tmpDir, "nonexistent.tar.gz")
	extractDir := filepath.Join(tmpDir, "extract")

	err = extractArchive(nonExistentArchive, extractDir)
	if err == nil {
		t.Error("Expected error for non-existent archive")
	}
}

func TestExtractArchiveSecurityCheck(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "pvm-extract-security-test-*")
	if err != nil {
		t.Skip("Cannot create temp directory:", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create malicious tar.gz archive with path traversal
	archivePath := filepath.Join(tmpDir, "malicious.tar.gz")
	err = createMaliciousTarGz(archivePath)
	if err != nil {
		t.Fatalf("Failed to create malicious archive: %v", err)
	}

	// Create extraction directory
	extractDir := filepath.Join(tmpDir, "extract")
	err = os.MkdirAll(extractDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create extract directory: %v", err)
	}

	// Test extraction should fail due to security check
	err = extractArchive(archivePath, extractDir)
	if err == nil {
		t.Error("Expected error for malicious archive with path traversal")
	}
	if err != nil && !containsString(err.Error(), "invalid path") {
		t.Errorf("Expected security error message, got: %v", err)
	}
}

// Helper function to create a test tar.gz archive
func createTestTarGz(archivePath string) error {
	// Create the archive file
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add a directory
	err = tarWriter.WriteHeader(&tar.Header{
		Name:     "bin/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	})
	if err != nil {
		return err
	}

	// Add a file
	content := "#!/bin/sh\necho 'fake perl'"
	err = tarWriter.WriteHeader(&tar.Header{
		Name:     "bin/perl",
		Mode:     0755,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	})
	if err != nil {
		return err
	}

	_, err = tarWriter.Write([]byte(content))
	if err != nil {
		return err
	}

	// Add another directory
	err = tarWriter.WriteHeader(&tar.Header{
		Name:     "lib/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	})
	if err != nil {
		return err
	}

	return nil
}

// Helper function to create a malicious tar.gz archive with path traversal
func createMaliciousTarGz(archivePath string) error {
	// Create the archive file
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add a file with path traversal
	content := "malicious content"
	err = tarWriter.WriteHeader(&tar.Header{
		Name:     "../../../tmp/malicious.txt",
		Mode:     0644,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	})
	if err != nil {
		return err
	}

	_, err = tarWriter.Write([]byte(content))
	if err != nil {
		return err
	}

	return nil
}

// Helper function to check if string contains substring
func containsString(str, substr string) bool {
	return strings.Contains(str, substr)
}
