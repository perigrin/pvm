// ABOUTME: Unit tests for binary installation functionality
// ABOUTME: Tests binary download, extraction, and installation logic

package perl

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInstallFromBinary(t *testing.T) {
	// Skip test if we can't create temp directories
	tmpDir, err := os.MkdirTemp("", "pvm-binary-install-test")
	if err != nil {
		t.Skip("Cannot create temp directory:", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		options     *BinaryInstallOptions
		expectError bool
		errorMsg    string
	}{
		{
			name: "invalid version",
			options: &BinaryInstallOptions{
				Version:    "invalid-version",
				Platform:   "linux-amd64",
				InstallDir: filepath.Join(tmpDir, "test-install-invalid"),
				Context:    context.Background(),
			},
			expectError: true,
			errorMsg:    "Invalid version format",
		},
		{
			name: "empty version",
			options: &BinaryInstallOptions{
				Version:    "",
				Platform:   "linux-amd64",
				InstallDir: filepath.Join(tmpDir, "test-install-empty"),
				Context:    context.Background(),
			},
			expectError: true,
			errorMsg:    "Invalid version format",
		},
		{
			name:        "nil options",
			options:     nil,
			expectError: true,
			errorMsg:    "Invalid version format", // Should fail on version validation after setting defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InstallFromBinary(tt.options)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errorMsg)) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
				}
			}
		})
	}
}

func TestExtractTarGzArchive(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pvm-extract-test")
	if err != nil {
		t.Skip("Cannot create temp directory:", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test tar.gz archive
	archivePath := filepath.Join(tmpDir, "test.tar.gz")
	err = createTestTarGz(archivePath)
	if err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}

	// Extract the archive
	extractDir := filepath.Join(tmpDir, "extract")
	size, err := extractTarGzArchive(archivePath, extractDir)
	if err != nil {
		t.Fatalf("Failed to extract archive: %v", err)
	}

	// Verify extraction
	if size <= 0 {
		t.Error("Expected positive extraction size")
	}

	// Check that files were extracted
	testFile := filepath.Join(extractDir, "bin", "perl")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Expected test file not found after extraction")
	}
}

func TestExtractZipArchive(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pvm-extract-zip-test")
	if err != nil {
		t.Skip("Cannot create temp directory:", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test ZIP archive
	archivePath := filepath.Join(tmpDir, "test.zip")
	err = createTestZip(archivePath)
	if err != nil {
		t.Fatalf("Failed to create test ZIP archive: %v", err)
	}

	// Extract the archive
	extractDir := filepath.Join(tmpDir, "extract")
	size, err := extractZipArchive(archivePath, extractDir)
	if err != nil {
		t.Fatalf("Failed to extract ZIP archive: %v", err)
	}

	// Verify extraction
	if size <= 0 {
		t.Error("Expected positive extraction size")
	}

	// Check that files were extracted
	testFile := filepath.Join(extractDir, "bin", "perl.exe")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Expected test file not found after extraction")
	}
}

func TestVerifyBinaryInstallation(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pvm-verify-test")
	if err != nil {
		t.Skip("Cannot create temp directory:", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		setupFunc   func(string) error
		expectError bool
	}{
		{
			name: "valid installation",
			setupFunc: func(installDir string) error {
				binDir := filepath.Join(installDir, "bin")
				err := os.MkdirAll(binDir, 0755)
				if err != nil {
					return err
				}

				perlPath := filepath.Join(binDir, "perl")
				return os.WriteFile(perlPath, []byte("#!/usr/bin/perl\n"), 0755)
			},
			expectError: false,
		},
		{
			name: "missing perl binary",
			setupFunc: func(installDir string) error {
				return os.MkdirAll(installDir, 0755)
			},
			expectError: true,
		},
		{
			name: "non-executable perl binary",
			setupFunc: func(installDir string) error {
				binDir := filepath.Join(installDir, "bin")
				err := os.MkdirAll(binDir, 0755)
				if err != nil {
					return err
				}

				perlPath := filepath.Join(binDir, "perl")
				return os.WriteFile(perlPath, []byte("#!/usr/bin/perl\n"), 0644) // Not executable
			},
			expectError: true,
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

			// Test verification
			err = verifyBinaryInstallation(installDir, "5.38.0")

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSetBinaryPermissions(t *testing.T) {
	// Skip on Windows as it doesn't have Unix permissions
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pvm-permissions-test")
	if err != nil {
		t.Skip("Cannot create temp directory:", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test directory structure
	binDir := filepath.Join(tmpDir, "bin")
	libDir := filepath.Join(tmpDir, "lib")
	err = os.MkdirAll(binDir, 0700) // Wrong permissions initially
	if err != nil {
		t.Fatalf("Failed to create bin directory: %v", err)
	}
	err = os.MkdirAll(libDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	// Create test files
	perlBinary := filepath.Join(binDir, "perl")
	err = os.WriteFile(perlBinary, []byte("#!/usr/bin/perl\n"), 0600) // Wrong permissions
	if err != nil {
		t.Fatalf("Failed to create perl binary: %v", err)
	}

	libFile := filepath.Join(libDir, "test.pm")
	err = os.WriteFile(libFile, []byte("package Test;\n"), 0600) // Wrong permissions
	if err != nil {
		t.Fatalf("Failed to create lib file: %v", err)
	}

	// Set correct permissions
	err = setBinaryPermissions(tmpDir)
	if err != nil {
		t.Fatalf("Failed to set permissions: %v", err)
	}

	// Verify permissions
	info, err := os.Stat(binDir)
	if err != nil {
		t.Fatalf("Failed to stat bin directory: %v", err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("Expected bin directory permissions 0755, got %o", info.Mode().Perm())
	}

	info, err = os.Stat(perlBinary)
	if err != nil {
		t.Fatalf("Failed to stat perl binary: %v", err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("Expected perl binary permissions 0755, got %o", info.Mode().Perm())
	}

	info, err = os.Stat(libFile)
	if err != nil {
		t.Fatalf("Failed to stat lib file: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("Expected lib file permissions 0644, got %o", info.Mode().Perm())
	}
}

// Helper function to create a test tar.gz archive
func createTestTarGz(archivePath string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add a bin directory
	err = tarWriter.WriteHeader(&tar.Header{
		Name:     "bin/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
		ModTime:  time.Now(),
	})
	if err != nil {
		return err
	}

	// Add a perl binary
	perlContent := "#!/usr/bin/perl\nprint \"Hello, World!\\n\";\n"
	err = tarWriter.WriteHeader(&tar.Header{
		Name:     "bin/perl",
		Mode:     0755,
		Size:     int64(len(perlContent)),
		Typeflag: tar.TypeReg,
		ModTime:  time.Now(),
	})
	if err != nil {
		return err
	}

	_, err = tarWriter.Write([]byte(perlContent))
	if err != nil {
		return err
	}

	return nil
}

// Helper function to create a test ZIP archive
func createTestZip(archivePath string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// Add a perl binary
	perlWriter, err := zipWriter.Create("bin/perl.exe")
	if err != nil {
		return err
	}

	perlContent := "@echo off\nperl.exe %*\n"
	_, err = io.WriteString(perlWriter, perlContent)
	if err != nil {
		return err
	}

	return nil
}
