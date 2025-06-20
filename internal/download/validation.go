// ABOUTME: File validation and integrity checking for downloaded binaries
// ABOUTME: Implements checksum validation and basic binary format verification

package download

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
)

// ChecksumValidator handles checksum validation
type ChecksumValidator struct {
	downloader *Downloader
}

// NewChecksumValidator creates a new checksum validator
func NewChecksumValidator() *ChecksumValidator {
	return &ChecksumValidator{
		downloader: NewDownloader(),
	}
}

// ValidateFile validates a file against an expected checksum
func (v *ChecksumValidator) ValidateFile(filePath, expectedChecksum string) error {
	if expectedChecksum == "" {
		return fmt.Errorf("expected checksum cannot be empty")
	}

	actualChecksum, err := v.calculateSHA256(filePath)
	if err != nil {
		return fmt.Errorf("calculating checksum: %w", err)
	}

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// DownloadAndValidateChecksums downloads a checksums file and validates against it
func (v *ChecksumValidator) DownloadAndValidateChecksums(checksumURL, filePath string) error {
	// Download checksums file to temporary location
	tempFile, err := os.CreateTemp("", "pvm-checksums-*.txt")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Download checksums
	opts := &DownloadOptions{
		URL:             checksumURL,
		DestinationPath: tempFile.Name(),
	}

	_, err = v.downloader.Download(opts)
	if err != nil {
		return fmt.Errorf("downloading checksums: %w", err)
	}

	// Parse checksums file and find our file
	expectedChecksum, err := v.parseChecksumsFile(tempFile.Name(), filePath)
	if err != nil {
		return fmt.Errorf("parsing checksums file: %w", err)
	}

	// Validate the file
	return v.ValidateFile(filePath, expectedChecksum)
}

// parseChecksumsFile parses a checksums file and returns the checksum for the given file
func (v *ChecksumValidator) parseChecksumsFile(checksumsPath, targetFile string) (string, error) {
	file, err := os.Open(checksumsPath)
	if err != nil {
		return "", fmt.Errorf("opening checksums file: %w", err)
	}
	defer file.Close()

	// Get the base name of the target file for matching
	targetBaseName := strings.TrimSuffix(targetFile, ".tmp")
	if idx := strings.LastIndex(targetBaseName, "/"); idx != -1 {
		targetBaseName = targetBaseName[idx+1:]
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse line: "checksum  filename" or "checksum *filename"
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		checksum := parts[0]
		filename := parts[1]

		// Remove binary mode indicator (*) if present
		if strings.HasPrefix(filename, "*") {
			filename = filename[1:]
		}

		// Check if this is our file
		if filename == targetBaseName || strings.HasSuffix(filename, targetBaseName) {
			return checksum, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading checksums file: %w", err)
	}

	return "", fmt.Errorf("checksum not found for file %s", targetBaseName)
}

// calculateSHA256 calculates the SHA256 checksum of a file
func (v *ChecksumValidator) calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// BinaryValidator validates binary file format and basic properties
type BinaryValidator struct{}

// NewBinaryValidator creates a new binary validator
func NewBinaryValidator() *BinaryValidator {
	return &BinaryValidator{}
}

// ValidateExecutable performs basic validation of an executable file
func (v *BinaryValidator) ValidateExecutable(filePath string) error {
	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file")
	}

	// Check minimum size (executables should be reasonably sized)
	minSize := int64(1024) // 1KB minimum
	if info.Size() < minSize {
		return fmt.Errorf("file too small: %d bytes (minimum %d)", info.Size(), minSize)
	}

	// Check maximum size (prevent downloading huge files by mistake)
	maxSize := int64(100 * 1024 * 1024) // 100MB maximum
	if info.Size() > maxSize {
		return fmt.Errorf("file too large: %d bytes (maximum %d)", info.Size(), maxSize)
	}

	// Read file header to validate format
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening file for validation: %w", err)
	}
	defer file.Close()

	// Read first few bytes to check file format
	header := make([]byte, 16)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return fmt.Errorf("reading file header: %w", err)
	}

	// Validate based on file format
	return v.validateFileFormat(header[:n], filePath)
}

// validateFileFormat validates the file format based on magic bytes
func (v *BinaryValidator) validateFileFormat(header []byte, filePath string) error {
	if len(header) < 4 {
		return fmt.Errorf("file header too short")
	}

	// Check for common executable formats

	// ELF (Linux/Unix)
	if len(header) >= 4 && header[0] == 0x7f && header[1] == 'E' && header[2] == 'L' && header[3] == 'F' {
		return nil // Valid ELF
	}

	// Mach-O (macOS) - single architecture
	if len(header) >= 4 {
		// Mach-O magic numbers
		magic := uint32(header[0])<<24 | uint32(header[1])<<16 | uint32(header[2])<<8 | uint32(header[3])
		switch magic {
		case 0xfeedface, 0xfeedfacf, 0xcafebabe, 0xcffaedfe, 0xcefaedfe:
			return nil // Valid Mach-O
		}
	}

	// PE (Windows)
	if len(header) >= 2 && header[0] == 'M' && header[1] == 'Z' {
		return nil // Valid PE (starts with DOS header)
	}

	// Check if it might be a shell script (for wrapper scripts)
	if len(header) >= 2 && header[0] == '#' && header[1] == '!' {
		return nil // Valid shell script
	}

	return fmt.Errorf("unrecognized file format (not a valid executable)")
}

// ValidateDownloadedBinary performs comprehensive validation of a downloaded binary
func ValidateDownloadedBinary(filePath, expectedChecksum string) error {
	// First validate checksum if provided
	if expectedChecksum != "" {
		validator := NewChecksumValidator()
		if err := validator.ValidateFile(filePath, expectedChecksum); err != nil {
			return fmt.Errorf("checksum validation failed: %w", err)
		}
	}

	// Then validate binary format
	binaryValidator := NewBinaryValidator()
	if err := binaryValidator.ValidateExecutable(filePath); err != nil {
		return fmt.Errorf("binary validation failed: %w", err)
	}

	return nil
}
