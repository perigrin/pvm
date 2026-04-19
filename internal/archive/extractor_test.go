// ABOUTME: Tests for archive extraction functionality
// ABOUTME: Covers tar.gz and zip extraction with security validation

package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinaryExtractor_ExtractExecutable(t *testing.T) {
	tests := []struct {
		name          string
		createArchive func(t *testing.T) string
		platform      string
		expectError   bool
		errorMatch    string
	}{
		{
			name: "valid tar.gz with pvm binary",
			createArchive: func(t *testing.T) string {
				return createTestTarGz(t, "pvm", createMockBinary())
			},
			platform:    "linux-amd64",
			expectError: false,
		},
		{
			name: "valid zip with pvm.exe binary",
			createArchive: func(t *testing.T) string {
				return createTestZip(t, "pvm.exe", createMockBinary())
			},
			platform:    "windows-amd64",
			expectError: false,
		},
		{
			name: "tar.gz with platform-named binary (fallback)",
			createArchive: func(t *testing.T) string {
				return createTestTarGz(t, "pvm-darwin-arm64", createMockBinary())
			},
			platform:    "darwin-arm64",
			expectError: false,
		},
		{
			name: "archive without executable",
			createArchive: func(t *testing.T) string {
				return createTestTarGz(t, "README.txt", []byte("documentation"))
			},
			platform:    "linux-amd64",
			expectError: true,
			errorMatch:  "no executable found",
		},
		{
			name: "unsupported archive format",
			createArchive: func(t *testing.T) string {
				tempFile, err := os.CreateTemp("", "test-*.rar")
				require.NoError(t, err)
				tempFile.WriteString("fake rar content")
				tempFile.Close()
				return tempFile.Name()
			},
			platform:    "linux-amd64",
			expectError: true,
			errorMatch:  "unsupported archive format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archivePath := tt.createArchive(t)
			defer os.Remove(archivePath)

			extractor := NewBinaryExtractor()
			extractedPath, err := extractor.ExtractExecutable(archivePath, tt.platform)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMatch != "" {
					assert.Contains(t, err.Error(), tt.errorMatch)
				}
			} else {
				require.NoError(t, err)

				// Verify extracted file exists
				_, err := os.Stat(extractedPath)
				assert.NoError(t, err)

				// Verify the extracted file has a recognized name
				filename := filepath.Base(extractedPath)
				if strings.HasPrefix(tt.platform, "windows") {
					assert.True(t, filename == "pvm.exe" || filename == "pvm-"+tt.platform+".exe",
						"expected pvm.exe or pvm-%s.exe, got %s", tt.platform, filename)
				} else {
					assert.True(t, filename == "pvm" || filename == "pvm-"+tt.platform,
						"expected pvm or pvm-%s, got %s", tt.platform, filename)
				}

				// Cleanup
				err = extractor.Cleanup(extractedPath)
				assert.NoError(t, err)
			}
		})
	}
}

func TestBinaryExtractor_ValidatePath(t *testing.T) {
	extractor := NewBinaryExtractor()

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "valid relative path",
			path:        "bin/pvm",
			expectError: false,
		},
		{
			name:        "path traversal with ..",
			path:        "../../../etc/passwd",
			expectError: true,
		},
		{
			name:        "absolute path",
			path:        "/usr/bin/pvm",
			expectError: true,
		},
		{
			name:        "nested relative path",
			path:        "some/nested/path/file",
			expectError: false,
		},
		{
			name:        "hidden directory traversal",
			path:        "bin/../../../etc/passwd",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := extractor.validatePath(tt.path)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBinaryExtractor_FindMainExecutable(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "extractor-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name         string
		files        map[string][]byte // filename -> content
		platform     string
		expectError  bool
		errorMatch   string
		expectedFile string
	}{
		{
			name: "linux platform with pvm binary",
			files: map[string][]byte{
				"pvm":        createMockBinary(),
				"README.txt": []byte("documentation"),
			},
			platform:     "linux-amd64",
			expectError:  false,
			expectedFile: "pvm",
		},
		{
			name: "windows platform with pvm.exe",
			files: map[string][]byte{
				"pvm.exe":    createMockBinary(),
				"README.txt": []byte("documentation"),
			},
			platform:     "windows-amd64",
			expectError:  false,
			expectedFile: "pvm.exe",
		},
		{
			name: "platform-named binary (pvm-darwin-arm64)",
			files: map[string][]byte{
				"pvm-darwin-arm64": createMockBinary(),
				"README.txt":       []byte("documentation"),
			},
			platform:     "darwin-arm64",
			expectError:  false,
			expectedFile: "pvm-darwin-arm64",
		},
		{
			name: "platform-named binary (pvm-linux-amd64)",
			files: map[string][]byte{
				"pvm-linux-amd64": createMockBinary(),
			},
			platform:     "linux-amd64",
			expectError:  false,
			expectedFile: "pvm-linux-amd64",
		},
		{
			name: "no executable found",
			files: map[string][]byte{
				"README.txt": []byte("documentation"),
				"config.yml": []byte("config"),
			},
			platform:    "linux-amd64",
			expectError: true,
			errorMatch:  "no executable found",
		},
		{
			name: "multiple executables",
			files: map[string][]byte{
				"pvm":     createMockBinary(),
				"bin/pvm": createMockBinary(),
			},
			platform:    "linux-amd64",
			expectError: true,
			errorMatch:  "multiple executables found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test files
			var createdFiles []string
			for filename, content := range tt.files {
				fullPath := filepath.Join(tempDir, filename)

				// Create directory if needed
				dir := filepath.Dir(fullPath)
				if dir != tempDir {
					err := os.MkdirAll(dir, 0755)
					require.NoError(t, err)
				}

				// Create file
				err := os.WriteFile(fullPath, content, 0755)
				require.NoError(t, err)
				createdFiles = append(createdFiles, fullPath)
			}

			extractor := NewBinaryExtractor()
			executablePath, err := extractor.findMainExecutable(createdFiles, tt.platform)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMatch != "" {
					assert.Contains(t, err.Error(), tt.errorMatch)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedFile, filepath.Base(executablePath))
			}
		})
	}
}

func TestBinaryExtractor_ExtractTarGz_Security(t *testing.T) {
	// Test path traversal protection
	tempDir, err := os.MkdirTemp("", "security-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create malicious tar.gz with path traversal
	archivePath := createMaliciousTarGz(t)
	defer os.Remove(archivePath)

	extractor := NewBinaryExtractor()
	extractor.TempDir = tempDir

	_, err = extractor.extractTarGz(archivePath, tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid path")
}

func TestBinaryExtractor_CustomTempDir(t *testing.T) {
	customTempDir, err := os.MkdirTemp("", "custom-temp-*")
	require.NoError(t, err)
	defer os.RemoveAll(customTempDir)

	archivePath := createTestTarGz(t, "pvm", createMockBinary())
	defer os.Remove(archivePath)

	extractor := NewBinaryExtractor()
	extractor.TempDir = customTempDir

	extractedPath, err := extractor.ExtractExecutable(archivePath, "linux-amd64")
	require.NoError(t, err)

	// Verify extracted file is in custom temp directory
	assert.Contains(t, extractedPath, customTempDir)

	// Cleanup
	err = extractor.Cleanup(extractedPath)
	assert.NoError(t, err)
}

// Helper functions

func createMockBinary() []byte {
	// Create a binary that will pass format validation
	switch runtime.GOOS {
	case "linux":
		// ELF header - more complete to pass size validation
		elf := make([]byte, 2048) // 2KB minimum
		copy(elf, []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00})
		return elf
	case "darwin":
		// Mach-O header
		macho := make([]byte, 2048)
		copy(macho, []byte{0xcf, 0xfa, 0xed, 0xfe, 0x07, 0x00, 0x00, 0x01})
		return macho
	case "windows":
		// PE header
		pe := make([]byte, 2048)
		copy(pe, []byte{'M', 'Z', 0x90, 0x00, 0x03, 0x00, 0x00, 0x00})
		return pe
	default:
		// Shell script as fallback
		script := "#!/bin/sh\necho 'test binary'\n"
		padded := make([]byte, 2048)
		copy(padded, script)
		return padded
	}
}

func createTestTarGz(t *testing.T, filename string, content []byte) string {
	tempFile, err := os.CreateTemp("", "test-*.tar.gz")
	require.NoError(t, err)
	defer tempFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(tempFile)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add file to tar
	header := &tar.Header{
		Name: filename,
		Size: int64(len(content)),
		Mode: 0755,
	}

	err = tarWriter.WriteHeader(header)
	require.NoError(t, err)

	_, err = tarWriter.Write(content)
	require.NoError(t, err)

	return tempFile.Name()
}

func createTestZip(t *testing.T, filename string, content []byte) string {
	tempFile, err := os.CreateTemp("", "test-*.zip")
	require.NoError(t, err)
	defer tempFile.Close()

	// Create zip writer
	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	// Create file header with proper permissions
	header := &zip.FileHeader{
		Name:   filename,
		Method: zip.Deflate,
	}
	// Set executable permissions (0755)
	header.SetMode(0755)

	// Add file to zip with header
	writer, err := zipWriter.CreateHeader(header)
	require.NoError(t, err)

	_, err = writer.Write(content)
	require.NoError(t, err)

	return tempFile.Name()
}

func createMaliciousTarGz(t *testing.T) string {
	tempFile, err := os.CreateTemp("", "malicious-*.tar.gz")
	require.NoError(t, err)
	defer tempFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(tempFile)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add malicious file with path traversal
	content := []byte("evil")
	header := &tar.Header{
		Name: "../../../evil.txt",
		Size: int64(len(content)),
		Mode: 0644,
	}

	err = tarWriter.WriteHeader(header)
	require.NoError(t, err)

	_, err = tarWriter.Write(content)
	require.NoError(t, err)

	return tempFile.Name()
}

// createTestTarGzWithMode builds a tar.gz where the file header carries
// the supplied mode. Used to regression-test extraction when the archive
// author forgot to set the executable bit.
func createTestTarGzWithMode(t *testing.T, filename string, content []byte, mode int64) string {
	tempFile, err := os.CreateTemp("", "test-mode-*.tar.gz")
	require.NoError(t, err)
	defer tempFile.Close()

	gzWriter := gzip.NewWriter(tempFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	require.NoError(t, tarWriter.WriteHeader(&tar.Header{
		Name: filename,
		Size: int64(len(content)),
		Mode: mode,
	}))
	_, err = tarWriter.Write(content)
	require.NoError(t, err)

	return tempFile.Name()
}

// TestBinaryExtractor_ExtractNonExecutableTarEntry is the regression test
// for the bad-update bug that produced rc71's "binary validation failed:
// new binary is not executable" rollback. Cross-platform packaging can
// strip the executable bit; the extractor must defensively re-apply 0755
// so downstream validation doesn't reject a perfectly good binary.
func TestBinaryExtractor_ExtractNonExecutableTarEntry(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are POSIX-only")
	}

	archive := createTestTarGzWithMode(t, "pvm", createMockBinary(), 0o644)
	defer os.Remove(archive)

	extractor := NewBinaryExtractor()
	got, err := extractor.ExtractExecutable(archive, "linux-amd64")
	require.NoError(t, err, "extraction should succeed even when tar mode is 0644")
	defer os.RemoveAll(filepath.Dir(got))

	info, err := os.Stat(got)
	require.NoError(t, err)
	mode := info.Mode().Perm()
	assert.NotZerof(t, mode&0o111, "extracted binary must be executable, got mode %o", mode)
}
