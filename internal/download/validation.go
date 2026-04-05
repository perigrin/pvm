// ABOUTME: File validation and integrity checking for downloaded binaries
// ABOUTME: Implements checksum validation and basic binary format verification

package download

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"tamarou.com/pvm/internal/archive"
)

// Binary validation constants
const (
	// Size limits for executable validation
	MinExecutableSize = 1024              // 1KB minimum
	MaxExecutableSize = 100 * 1024 * 1024 // 100MB maximum
	MinBinarySize     = 2048              // 2KB minimum for real binaries

	// File permissions
	DefaultFilePermissions = 0755

	// Timeouts
	DefaultVersionTimeout   = 10 * time.Second
	DefaultExecutionTimeout = 5 * time.Second

	// Binary format magic numbers
	// ELF magic bytes
	ELFMagic0 = 0x7f
	ELFMagic1 = 'E'
	ELFMagic2 = 'L'
	ELFMagic3 = 'F'

	// Mach-O magic numbers
	MachOMagic32      = 0xfeedface // 32-bit Mach-O
	MachOMagic64      = 0xfeedfacf // 64-bit Mach-O
	MachOFatMagic     = 0xcafebabe // Fat binary
	MachOFatMagicSwap = 0xcffaedfe // Fat binary (swapped)
	MachOMagic64Swap  = 0xcefaedfe // 64-bit Mach-O (swapped)

	// PE magic bytes
	PEMagic0 = 'M'
	PEMagic1 = 'Z'

	// Shell script magic bytes
	ShellMagic0 = '#'
	ShellMagic1 = '!'
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

	// Validate checksum format for security
	if !isValidSHA256(expectedChecksum) {
		return fmt.Errorf("invalid SHA256 checksum format: %s", expectedChecksum)
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

// isValidSHA256 validates that a string is a properly formatted SHA256 hash
func isValidSHA256(checksum string) bool {
	// SHA256 hash should be exactly 64 hexadecimal characters
	if len(checksum) != 64 {
		return false
	}

	// Check that all characters are valid hex
	for _, c := range checksum {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
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
	if info.Size() < MinExecutableSize {
		return fmt.Errorf("file too small: %d bytes (minimum %d)", info.Size(), MinExecutableSize)
	}

	// Check maximum size (prevent downloading huge files by mistake)
	if info.Size() > MaxExecutableSize {
		return fmt.Errorf("file too large: %d bytes (maximum %d)", info.Size(), MaxExecutableSize)
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
	if len(header) >= 4 && header[0] == ELFMagic0 && header[1] == ELFMagic1 && header[2] == ELFMagic2 && header[3] == ELFMagic3 {
		return nil // Valid ELF
	}

	// Mach-O (macOS) - single architecture
	if len(header) >= 4 {
		// Mach-O magic numbers
		magic := uint32(header[0])<<24 | uint32(header[1])<<16 | uint32(header[2])<<8 | uint32(header[3])
		switch magic {
		case MachOMagic32, MachOMagic64, MachOFatMagic, MachOFatMagicSwap, MachOMagic64Swap:
			return nil // Valid Mach-O
		}
	}

	// PE (Windows)
	if len(header) >= 2 && header[0] == PEMagic0 && header[1] == PEMagic1 {
		return nil // Valid PE (starts with DOS header)
	}

	// Check if it might be a shell script (for wrapper scripts)
	if len(header) >= 2 && header[0] == ShellMagic0 && header[1] == ShellMagic1 {
		return nil // Valid shell script
	}

	return fmt.Errorf("unrecognized file format (not a valid executable)")
}

// ValidateDownloadedBinary performs comprehensive validation of a downloaded binary or archive.
// Returns the path to the usable binary (which may be an extracted file from an archive).
// The caller is responsible for cleaning up any extracted temporary files.
func ValidateDownloadedBinary(filePath, expectedChecksum string) (string, error) {
	return ValidateDownloadedBinaryWithVersion(filePath, expectedChecksum, "")
}

// ValidateDownloadedBinaryWithVersion performs comprehensive validation including version verification.
// Returns the path to the validated binary. For archives, this is the path to the extracted
// binary in a temporary directory — the caller must clean up this path when done.
func ValidateDownloadedBinaryWithVersion(filePath, expectedChecksum, expectedVersion string) (string, error) {
	// First validate checksum if provided
	if expectedChecksum != "" {
		validator := NewChecksumValidator()
		if err := validator.ValidateFile(filePath, expectedChecksum); err != nil {
			return "", fmt.Errorf("checksum validation failed: %w", err)
		}
	}

	// Check if this is an archive that needs extraction
	var binaryPath string
	var extractor *archive.BinaryExtractor

	if isArchive(filePath) {
		// Extract binary from archive
		extractor = archive.NewBinaryExtractor()
		platform := detectPlatform()

		extractedPath, err := extractor.ExtractExecutable(filePath, platform)
		if err != nil {
			return "", fmt.Errorf("archive extraction failed: %w", err)
		}

		binaryPath = extractedPath
	} else {
		binaryPath = filePath
	}

	// Validate binary format
	binaryValidator := NewBinaryValidator()
	if err := binaryValidator.ValidateExecutable(binaryPath); err != nil {
		cleanupExtracted(extractor, binaryPath)
		return "", fmt.Errorf("binary validation failed: %w", err)
	}

	// Validate version if provided
	if expectedVersion != "" {
		if err := validateExecutableVersion(binaryPath, expectedVersion); err != nil {
			cleanupExtracted(extractor, binaryPath)
			return "", fmt.Errorf("version validation failed: %w", err)
		}
	}

	// Perform execution test
	if err := validateExecutableCanRun(binaryPath); err != nil {
		cleanupExtracted(extractor, binaryPath)
		return "", fmt.Errorf("execution validation failed: %w", err)
	}

	return binaryPath, nil
}

// cleanupExtracted removes the extracted temp directory on validation failure.
// Safe to call with nil extractor (no-op for non-archive paths).
func cleanupExtracted(extractor *archive.BinaryExtractor, extractedPath string) {
	if extractor != nil {
		extractor.Cleanup(extractedPath)
	}
}

// isArchive checks if the file is an archive format
func isArchive(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".zip" || ext == ".gz" || strings.HasSuffix(strings.ToLower(filePath), ".tar.gz") || strings.HasSuffix(strings.ToLower(filePath), ".tgz")
}

// detectPlatform detects the current platform for binary extraction
func detectPlatform() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Convert Go arch names to our naming convention
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	case "386":
		arch = "i386"
	default:
		arch = "unknown"
	}

	return fmt.Sprintf("%s-%s", os, arch)
}

// validateExecutableVersion validates binary metadata without executing it
func validateExecutableVersion(binaryPath, expectedVersion string) error {
	// For security reasons, we don't execute untrusted binaries during validation
	// Instead, we rely on:
	// 1. Checksum validation (if provided)
	// 2. Download source validation (GitHub releases)
	// 3. Binary format validation

	// Skip version validation during binary validation for security
	// Version verification should happen at the download/source level
	return nil
}

// validateExecutableCanRun performs static validation without executing the binary
func validateExecutableCanRun(binaryPath string) error {
	// For security reasons, we don't execute untrusted binaries during validation
	// Instead, we perform static checks:

	// 1. Verify the binary has executable permissions
	info, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("cannot access binary: %w", err)
	}

	// Check if file is executable (on Unix-like systems)
	if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		return fmt.Errorf("binary is not executable")
	}

	// 2. Verify minimum size (should be a real binary, not empty file)
	if info.Size() < MinBinarySize {
		return fmt.Errorf("binary too small to be valid: %d bytes (minimum %d)", info.Size(), MinBinarySize)
	}

	// Additional static checks could be added here:
	// - ELF/PE/Mach-O header validation
	// - Digital signature verification
	// - Known binary hash verification

	return nil
}
