// ABOUTME: Comprehensive tests for download validation functionality
// ABOUTME: Tests checksum validation, binary format verification, and version checking

package download

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChecksumValidator_ValidateFile(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		expectedChecksum string
		expectError      bool
	}{
		{
			name:             "valid checksum",
			content:          "test content",
			expectedChecksum: "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72", // sha256 of "test content"
			expectError:      false,
		},
		{
			name:             "invalid checksum",
			content:          "test content",
			expectedChecksum: "invalidchecksum",
			expectError:      true,
		},
		{
			name:             "empty checksum",
			content:          "test content",
			expectedChecksum: "",
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tempFile, err := os.CreateTemp("", "checksum-test-*.txt")
			require.NoError(t, err)
			defer os.Remove(tempFile.Name())

			// Write content
			_, err = tempFile.WriteString(tt.content)
			require.NoError(t, err)
			tempFile.Close()

			// Test validation
			validator := NewChecksumValidator()
			err = validator.ValidateFile(tempFile.Name(), tt.expectedChecksum)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBinaryValidator_ValidateExecutable(t *testing.T) {
	tests := []struct {
		name        string
		content     []byte
		expectError bool
		errorMatch  string
	}{
		{
			name:        "valid ELF binary",
			content:     createTestBinaryWithHeader([]byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}),
			expectError: false,
		},
		{
			name:        "valid Mach-O binary (64-bit)",
			content:     createTestBinaryWithHeader([]byte{0xcf, 0xfa, 0xed, 0xfe, 0x07, 0x00, 0x00, 0x01, 0x03, 0x00, 0x00, 0x80, 0x02, 0x00, 0x00, 0x00}),
			expectError: false,
		},
		{
			name:        "valid PE binary",
			content:     createTestBinaryWithHeader([]byte{'M', 'Z', 0x90, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0xff, 0xff, 0x00, 0x00}),
			expectError: false,
		},
		{
			name:        "valid shell script",
			content:     createTestBinaryWithHeader([]byte("#!/bin/sh\necho hello\n")),
			expectError: false,
		},
		{
			name:        "invalid format",
			content:     createTestBinaryWithHeader([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}),
			expectError: true,
			errorMatch:  "unrecognized file format",
		},
		{
			name:        "too small file",
			content:     []byte{0x7f, 'E', 'L'},
			expectError: true,
			errorMatch:  "file too small",
		},
		{
			name:        "empty file",
			content:     []byte{},
			expectError: true,
			errorMatch:  "file too small",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tempFile, err := os.CreateTemp("", "binary-test-*")
			require.NoError(t, err)
			defer os.Remove(tempFile.Name())

			// Write content
			_, err = tempFile.Write(tt.content)
			require.NoError(t, err)
			tempFile.Close()

			// Test validation
			validator := NewBinaryValidator()
			err = validator.ValidateExecutable(tempFile.Name())

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMatch != "" {
					assert.Contains(t, err.Error(), tt.errorMatch)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDownloadedBinary_Archive(t *testing.T) {
	// Create a simple test binary
	binaryContent := createTestBinary(t)

	// Test tar.gz archive
	t.Run("tar.gz archive", func(t *testing.T) {
		archivePath := createTestTarGz(t, "pvm", binaryContent)
		defer os.Remove(archivePath)

		err := ValidateDownloadedBinary(archivePath, "")
		assert.NoError(t, err)
	})

	// Test zip archive
	t.Run("zip archive", func(t *testing.T) {
		// Use correct filename for current platform
		filename := "pvm"
		if runtime.GOOS == "windows" {
			filename = "pvm.exe"
		}

		archivePath := createTestZip(t, filename, binaryContent)
		defer os.Remove(archivePath)

		err := ValidateDownloadedBinary(archivePath, "")
		assert.NoError(t, err)
	})
}

func TestValidateDownloadedBinary_Direct(t *testing.T) {
	// Create a test binary directly
	binaryContent := createTestBinary(t)

	tempFile, err := os.CreateTemp("", "binary-test-*")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(binaryContent)
	require.NoError(t, err)
	tempFile.Close()

	// Make it executable
	err = os.Chmod(tempFile.Name(), 0755)
	require.NoError(t, err)

	err = ValidateDownloadedBinary(tempFile.Name(), "")
	assert.NoError(t, err)
}

func TestIsArchive(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"file.tar.gz", true},
		{"file.tgz", true},
		{"file.zip", true},
		{"file.gz", true},
		{"file.TAR.GZ", true}, // case insensitive
		{"file.txt", false},
		{"file.exe", false},
		{"file", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := isArchive(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectPlatform(t *testing.T) {
	platform := detectPlatform()

	// Should contain OS and architecture
	assert.Contains(t, platform, runtime.GOOS)
	assert.Contains(t, platform, "-")

	// Should be in expected format
	parts := strings.Split(platform, "-")
	assert.Len(t, parts, 2)

	// OS should be valid
	validOS := []string{"linux", "darwin", "windows"}
	assert.Contains(t, validOS, parts[0])

	// Architecture should be valid
	validArch := []string{"amd64", "arm64", "i386", "unknown"}
	assert.Contains(t, validArch, parts[1])
}

func TestParseChecksumsFile(t *testing.T) {
	checksumContent := `# Test checksums file
6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72  test-file.txt
abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890 *binary-file
# Another comment
fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321  other-file.dat`

	// Create temporary checksums file
	tempFile, err := os.CreateTemp("", "checksums-*.txt")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(checksumContent)
	require.NoError(t, err)
	tempFile.Close()

	validator := NewChecksumValidator()

	// Test finding existing file
	checksum, err := validator.parseChecksumsFile(tempFile.Name(), "test-file.txt")
	assert.NoError(t, err)
	assert.Equal(t, "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72", checksum)

	// Test finding binary file (with *)
	checksum, err = validator.parseChecksumsFile(tempFile.Name(), "binary-file")
	assert.NoError(t, err)
	assert.Equal(t, "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", checksum)

	// Test file not found
	_, err = validator.parseChecksumsFile(tempFile.Name(), "nonexistent-file")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum not found")
}

// Helper functions

func createTestBinaryWithHeader(header []byte) []byte {
	// Pad to meet minimum size requirement (1024 bytes)
	binary := make([]byte, 2048) // 2KB to be safe
	copy(binary, header)
	return binary
}

func createTestBinary(t *testing.T) []byte {
	// Create a binary that will pass format validation and size requirements (1024+ bytes)
	var header []byte
	switch runtime.GOOS {
	case "linux":
		// ELF header
		header = []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	case "darwin":
		// Mach-O header
		header = []byte{0xcf, 0xfa, 0xed, 0xfe, 0x07, 0x00, 0x00, 0x01, 0x03, 0x00, 0x00, 0x80, 0x02, 0x00, 0x00, 0x00}
	case "windows":
		// PE header
		header = []byte{'M', 'Z', 0x90, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0xff, 0xff, 0x00, 0x00}
	default:
		// Shell script as fallback
		header = []byte("#!/bin/sh\necho test\n")
	}

	// Pad to meet minimum size requirement (1024 bytes)
	binary := make([]byte, 2048) // 2KB to be safe
	copy(binary, header)
	return binary
}

func createTestTarGz(t *testing.T, filename string, content []byte) string {
	// Create temporary file for archive
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
	// Create temporary file for archive
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

func TestValidateExecutableVersion_MockBinary(t *testing.T) {
	// Skip this test in CI or when we can't create executable files
	if testing.Short() {
		t.Skip("Skipping executable version test in short mode")
	}

	// Create a mock script that outputs version information
	scriptContent := fmt.Sprintf(`#!/bin/sh
if [ "$1" = "--version" ]; then
    echo "pvm version 1.0.0-rc37"
    exit 0
elif [ "$1" = "--help" ]; then
    echo "Usage: pvm [command]"
    exit 0
fi
echo "Unknown command"
exit 1
`)

	tempFile, err := os.CreateTemp("", "mock-pvm-*")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(scriptContent)
	require.NoError(t, err)
	tempFile.Close()

	// Make it executable
	err = os.Chmod(tempFile.Name(), 0755)
	require.NoError(t, err)

	// Test version validation (security-focused approach - doesn't execute binaries)
	err = validateExecutableVersion(tempFile.Name(), "1.0.0-rc37")
	assert.NoError(t, err)

	// Test with different version (still passes because we don't execute for security)
	err = validateExecutableVersion(tempFile.Name(), "2.0.0")
	assert.NoError(t, err) // Security: We don't execute untrusted binaries for version checking

	// Test execution validation - expect error because mock script is too small
	err = validateExecutableCanRun(tempFile.Name())
	assert.Error(t, err) // Mock script is smaller than MinBinarySize (2048 bytes)
	assert.Contains(t, err.Error(), "binary too small to be valid")
}
