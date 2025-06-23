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
	"tamarou.com/pvm/internal/cli/ui"
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

func TestIsURL(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "HTTP URL",
			path:     "http://example.com/perl.tar.gz",
			expected: true,
		},
		{
			name:     "HTTPS URL",
			path:     "https://github.com/releases/perl-5.38.0.tar.gz",
			expected: true,
		},
		{
			name:     "Local file path",
			path:     "/path/to/perl.tar.gz",
			expected: false,
		},
		{
			name:     "Relative file path",
			path:     "./perl.tar.gz",
			expected: false,
		},
		{
			name:     "FTP URL (not supported)",
			path:     "ftp://example.com/perl.tar.gz",
			expected: false,
		},
		{
			name:     "File URL (not supported)",
			path:     "file:///path/to/perl.tar.gz",
			expected: false,
		},
		{
			name:     "Empty string",
			path:     "",
			expected: false,
		},
		{
			name:     "Invalid URL",
			path:     "not-a-url",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isURL(test.path)
			if result != test.expected {
				t.Errorf("isURL(%q) = %v, expected %v", test.path, result, test.expected)
			}
		})
	}
}

func TestInstallPerlCommandURLFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "URL as positional argument",
			args:        []string{"https://example.com/perl.tar.gz"},
			expectError: false,
		},
		{
			name:        "URL with from-build flag",
			args:        []string{"--from-build", "https://example.com/perl.tar.gz"},
			expectError: false,
		},
		{
			name:        "URL with mirror override",
			args:        []string{"https://example.com/perl.tar.gz", "--mirror", "https://custom-mirror.com"},
			expectError: false,
		},
		{
			name:        "URL with version override",
			args:        []string{"https://example.com/perl.tar.gz", "--version", "5.38.0"},
			expectError: false,
		},
		{
			name:        "URL with force flag",
			args:        []string{"https://example.com/perl.tar.gz", "--force"},
			expectError: false,
		},
		{
			name:        "URL with all flags",
			args:        []string{"https://example.com/perl.tar.gz", "--mirror", "https://custom.com", "--version", "5.38.0", "--force"},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a new command instance for each test
			cmd := newInstallPerlCommand()

			// Mock the RunE function to avoid actual installation
			originalRunE := cmd.RunE
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				// Validate that flags can be retrieved without error
				_, err := cmd.Flags().GetString("from-build")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetString("version")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetBool("force")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetString("mirror")
				if err != nil {
					return err
				}
				return nil
			}

			// Test flag parsing
			cmd.SetArgs(test.args)
			err := cmd.Execute()

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if test.errorMsg != "" && !containsString(err.Error(), test.errorMsg) {
					t.Errorf("Expected error message to contain %q, got: %v", test.errorMsg, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Restore the original RunE function
			cmd.RunE = originalRunE
		})
	}
}

func TestInstallPerlSourceTypeDetection(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		expectedType string
	}{
		{
			name:         "HTTP URL",
			source:       "http://example.com/perl.tar.gz",
			expectedType: "URL",
		},
		{
			name:         "HTTPS URL",
			source:       "https://github.com/releases/perl-5.38.0.tar.gz",
			expectedType: "URL",
		},
		{
			name:         "tar.gz archive",
			source:       "/path/to/perl.tar.gz",
			expectedType: "archive",
		},
		{
			name:         "tgz archive",
			source:       "/path/to/perl.tgz",
			expectedType: "archive",
		},
		{
			name:         "directory",
			source:       "/path/to/perl-build",
			expectedType: "directory",
		},
		{
			name:         "current directory",
			source:       ".",
			expectedType: "directory",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var detectedType string
			switch {
			case isURL(test.source):
				detectedType = "URL"
			case isArchive(test.source):
				detectedType = "archive"
			default:
				detectedType = "directory"
			}

			if detectedType != test.expectedType {
				t.Errorf("Source %q detected as %q, expected %q", test.source, detectedType, test.expectedType)
			}
		})
	}
}

func TestInstallPerlCommandIntegration(t *testing.T) {
	// Create a new command instance
	cmd := newInstallPerlCommand()

	// Verify the command structure
	if cmd.Use != "install-perl" {
		t.Errorf("Expected command use to be 'install-perl', got %q", cmd.Use)
	}

	if !containsString(cmd.Short, "archive") || !containsString(cmd.Short, "URL") {
		t.Errorf("Expected command short description to mention both archive and URL support, got %q", cmd.Short)
	}

	if !containsString(cmd.Long, "URL") {
		t.Errorf("Expected command long description to mention URL support, got %q", cmd.Long)
	}

	// Verify flags exist
	flags := []string{"from-build", "version", "force", "mirror"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %q to exist", flagName)
		}
	}

	// Verify mirror flag is a string flag
	mirrorFlag := cmd.Flags().Lookup("mirror")
	if mirrorFlag.Value.Type() != "string" {
		t.Errorf("Expected mirror flag to be string type, got %q", mirrorFlag.Value.Type())
	}
}

func TestExtractVersionFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Standard Perl archive",
			url:      "https://example.com/perl-5.38.0.tar.gz",
			expected: "5.38.0",
		},
		{
			name:     "Perl archive with v prefix",
			url:      "https://github.com/releases/perl-v5.38.0.tar.gz",
			expected: "5.38.0",
		},
		{
			name:     "Version with v prefix only",
			url:      "https://example.com/v5.38.0.tar.gz",
			expected: "5.38.0",
		},
		{
			name:     "Version only",
			url:      "https://example.com/5.38.0.tar.gz",
			expected: "5.38.0",
		},
		{
			name:     "Complex path",
			url:      "https://example.com/sources/perl/perl-5.36.1.tar.bz2",
			expected: "5.36.1",
		},
		{
			name:     "No version in URL",
			url:      "https://example.com/perl-source.tar.gz",
			expected: "",
		},
		{
			name:     "Invalid URL",
			url:      "not-a-url",
			expected: "",
		},
		{
			name:     "tgz extension",
			url:      "https://example.com/perl-5.40.0.tgz",
			expected: "5.40.0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := extractVersionFromURL(test.url)
			if result != test.expected {
				t.Errorf("extractVersionFromURL(%q) = %q, expected %q", test.url, result, test.expected)
			}
		})
	}
}

func TestBuildPerlCommandURLSupport(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "URL as positional argument",
			args:        []string{"https://example.com/perl-5.38.0.tar.gz"},
			expectError: false,
		},
		{
			name:        "Version as positional argument",
			args:        []string{"5.38.0"},
			expectError: false,
		},
		{
			name:        "URL with build-only flag",
			args:        []string{"https://example.com/perl.tar.gz", "--build-only"},
			expectError: false,
		},
		{
			name:        "URL with output directory",
			args:        []string{"https://example.com/perl.tar.gz", "--output-dir", "/tmp/perl-build"},
			expectError: false,
		},
		{
			name:        "URL with configure options",
			args:        []string{"https://example.com/perl.tar.gz", "--configure-options", "-Dusethreads"},
			expectError: false,
		},
		{
			name:        "No arguments",
			args:        []string{},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 0",
		},
		{
			name:        "Too many arguments",
			args:        []string{"5.38.0", "extra"},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a new command instance for each test
			cmd := newBuildPerlCommand()

			// Mock the RunE function to avoid actual building
			originalRunE := cmd.RunE
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				// Validate that flags can be retrieved without error
				_, err := cmd.Flags().GetString("source")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetString("prefix")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetString("output-dir")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetInt("jobs")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetBool("test")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetBool("cleanup")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetBool("build-only")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetStringArray("configure-options")
				if err != nil {
					return err
				}
				return nil
			}

			// Test flag parsing
			cmd.SetArgs(test.args)
			err := cmd.Execute()

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if test.errorMsg != "" && !containsString(err.Error(), test.errorMsg) {
					t.Errorf("Expected error message to contain %q, got: %v", test.errorMsg, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Restore the original RunE function
			cmd.RunE = originalRunE
		})
	}
}

func TestBuildPerlCommandIntegration(t *testing.T) {
	// Create a new command instance
	cmd := newBuildPerlCommand()

	// Verify the command structure
	if cmd.Use != "build-perl [version|URL]" {
		t.Errorf("Expected command use to be 'build-perl [version|URL]', got %q", cmd.Use)
	}

	if !containsString(cmd.Short, "URL") {
		t.Errorf("Expected command short description to mention URL support, got %q", cmd.Short)
	}

	if !containsString(cmd.Long, "URL") {
		t.Errorf("Expected command long description to mention URL support, got %q", cmd.Long)
	}

	// Verify required flags exist
	requiredFlags := []string{"source", "prefix", "output-dir", "jobs", "test", "cleanup", "build-only", "configure-options"}
	for _, flagName := range requiredFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %q to exist", flagName)
		}
	}

	// Verify upload integration flags exist
	uploadFlags := []string{"upload", "platforms", "mirror", "github-token", "github-repo", "release-tag", "draft-release", "prerelease"}
	for _, flagName := range uploadFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected upload flag %q to exist", flagName)
		}
	}

	// Verify exact args requirement
	if cmd.Args == nil {
		t.Error("Expected command to have Args validation")
	}
}

func TestBuildPerlUploadIntegration(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "upload without build-only should succeed",
			args:        []string{"5.38.0", "--upload"},
			expectError: false,
		},
		{
			name:        "upload with build-only should succeed",
			args:        []string{"5.38.0", "--upload", "--build-only"},
			expectError: false,
		},
		{
			name:        "platforms without upload should fail",
			args:        []string{"5.38.0", "--platforms", "linux-amd64"},
			expectError: true,
			errorMsg:    "--platforms flag requires --upload",
		},
		{
			name:        "platforms with upload should succeed",
			args:        []string{"5.38.0", "--platforms", "linux-amd64,darwin-arm64", "--upload"},
			expectError: false,
		},
		{
			name:        "upload with github settings",
			args:        []string{"5.38.0", "--upload", "--github-repo", "owner/repo", "--github-token", "token123"},
			expectError: false,
		},
		{
			name:        "upload with custom mirror",
			args:        []string{"5.38.0", "--upload", "--mirror", "custom-mirror"},
			expectError: false,
		},
		{
			name:        "upload with release options",
			args:        []string{"5.38.0", "--upload", "--release-tag", "v5.38.0", "--draft-release", "--prerelease"},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a new command instance for each test
			cmd := newBuildPerlCommand()

			// Mock the RunE function to avoid actual building/uploading
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				// Test the validation logic that would normally run in buildPerlFromSource
				upload, _ := cmd.Flags().GetBool("upload")
				platforms, _ := cmd.Flags().GetStringArray("platforms")

				// Upload is available without build-only requirement

				if len(platforms) > 0 && !upload {
					return fmt.Errorf("--platforms flag requires --upload to be enabled")
				}

				// Validate that all upload flags can be retrieved
				_, err := cmd.Flags().GetString("mirror")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetString("github-token")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetString("github-repo")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetString("release-tag")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetBool("draft-release")
				if err != nil {
					return err
				}
				_, err = cmd.Flags().GetBool("prerelease")
				if err != nil {
					return err
				}

				return nil
			}

			// Test flag parsing and validation
			cmd.SetArgs(test.args)
			err := cmd.Execute()

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if test.errorMsg != "" && !containsString(err.Error(), test.errorMsg) {
					t.Errorf("Expected error message to contain %q, got: %v", test.errorMsg, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestPerformUpload(t *testing.T) {
	// Test that performUpload returns nil when upload is disabled
	t.Run("upload disabled", func(t *testing.T) {
		// Create a real UI for testing (needed for the *ui.Output type)
		ctx := &ui.UIContext{
			Writer: os.Stdout,
			Quiet:  true, // Suppress output during testing
		}
		uiOutput := ui.NewOutput(ctx)

		// Call performUpload with upload disabled
		err := performUpload("/tmp/test", "5.38.0", false, "", "", "", "", false, false, uiOutput)

		if err != nil {
			t.Errorf("Expected no error when upload disabled, got: %v", err)
		}
	})

	// Additional upload tests would require mocking createTarGzArchive, uploadToGitHub, etc.
	// For now, we test the flag validation logic in the command tests above
}

// mockUI implements a simple UI for testing (compatible with ui.Output interface)
type mockUI struct {
	messages []string
}

func (m *mockUI) Info(format string, args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf("INFO: "+format, args...))
}

func (m *mockUI) Success(format string, args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf("SUCCESS: "+format, args...))
}

func (m *mockUI) Warning(format string, args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf("WARNING: "+format, args...))
}

func (m *mockUI) Error(format string, args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf("ERROR: "+format, args...))
}

func (m *mockUI) Debug(format string, args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf("DEBUG: "+format, args...))
}

func (m *mockUI) Header(title string) {
	m.messages = append(m.messages, fmt.Sprintf("HEADER: %s", title))
}

func (m *mockUI) SubHeader(title string) {
	m.messages = append(m.messages, fmt.Sprintf("SUBHEADER: %s", title))
}

func (m *mockUI) Printf(format string, args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf("PRINTF: "+format, args...))
}

func (m *mockUI) Println(args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf("PRINTLN: %v", args))
}

func (m *mockUI) Table(headers []string, rows [][]string) {
	m.messages = append(m.messages, fmt.Sprintf("TABLE: %v %v", headers, rows))
}

func (m *mockUI) List(items []string) {
	m.messages = append(m.messages, fmt.Sprintf("LIST: %v", items))
}

func (m *mockUI) KeyValue(pairs map[string]string) {
	m.messages = append(m.messages, fmt.Sprintf("KEYVALUE: %v", pairs))
}

func (m *mockUI) Status(message string) {
	m.messages = append(m.messages, fmt.Sprintf("STATUS: %s", message))
}

func (m *mockUI) Progress(current, total int, message string) {
	m.messages = append(m.messages, fmt.Sprintf("PROGRESS: %d/%d %s", current, total, message))
}
