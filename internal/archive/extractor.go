// ABOUTME: Archive extraction utilities for binary updates and installations
// ABOUTME: Provides secure extraction with path traversal protection and executable detection

package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// BinaryExtractor handles extraction of executable binaries from archives
type BinaryExtractor struct {
	TempDir string
}

// NewBinaryExtractor creates a new binary extractor
func NewBinaryExtractor() *BinaryExtractor {
	return &BinaryExtractor{}
}

// ExtractExecutable extracts the main executable from an archive
func (e *BinaryExtractor) ExtractExecutable(archivePath, platform string) (string, error) {
	// Create temporary directory for extraction
	tempDir, err := e.createTempDir()
	if err != nil {
		return "", fmt.Errorf("creating temp directory: %w", err)
	}

	// Extract archive based on format
	var extractedFiles []string
	switch {
	case strings.HasSuffix(archivePath, ".zip"):
		extractedFiles, err = e.extractZip(archivePath, tempDir)
	case strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz"):
		extractedFiles, err = e.extractTarGz(archivePath, tempDir)
	default:
		return "", fmt.Errorf("unsupported archive format: %s", archivePath)
	}

	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("extracting archive: %w", err)
	}

	// Find the main executable
	executablePath, err := e.findMainExecutable(extractedFiles, platform)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("finding executable: %w", err)
	}

	return executablePath, nil
}

// createTempDir creates a temporary directory for extraction
func (e *BinaryExtractor) createTempDir() (string, error) {
	if e.TempDir != "" {
		return os.MkdirTemp(e.TempDir, "pvm-extract-*")
	}
	return os.MkdirTemp("", "pvm-extract-*")
}

// extractTarGz extracts a tar.gz archive and returns list of extracted files
func (e *BinaryExtractor) extractTarGz(archivePath, destDir string) ([]string, error) {
	var extractedFiles []string

	// Open archive file
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("opening archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("creating gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar header: %w", err)
		}

		// Security: prevent path traversal
		if err := e.validatePath(header.Name); err != nil {
			return nil, fmt.Errorf("invalid path in archive: %w", err)
		}

		targetPath := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return nil, fmt.Errorf("creating directory %s: %w", targetPath, err)
			}

		case tar.TypeReg:
			// Extract regular file
			if err := e.extractTarFile(tarReader, targetPath, header.Mode); err != nil {
				return nil, fmt.Errorf("extracting file %s: %w", targetPath, err)
			}
			extractedFiles = append(extractedFiles, targetPath)
		}
	}

	return extractedFiles, nil
}

// extractZip extracts a ZIP archive and returns list of extracted files
func (e *BinaryExtractor) extractZip(archivePath, destDir string) ([]string, error) {
	var extractedFiles []string

	// Open ZIP file
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("opening zip archive: %w", err)
	}
	defer reader.Close()

	// Extract files
	for _, file := range reader.File {
		// Security: prevent path traversal
		if err := e.validatePath(file.Name); err != nil {
			return nil, fmt.Errorf("invalid path in archive: %w", err)
		}

		targetPath := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			// Create directory
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return nil, fmt.Errorf("creating directory %s: %w", targetPath, err)
			}
			continue
		}

		// Extract regular file
		if err := e.extractZipFile(file, targetPath); err != nil {
			return nil, fmt.Errorf("extracting file %s: %w", targetPath, err)
		}
		extractedFiles = append(extractedFiles, targetPath)
	}

	return extractedFiles, nil
}

// extractTarFile extracts a single file from tar archive
func (e *BinaryExtractor) extractTarFile(reader io.Reader, targetPath string, mode int64) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("creating parent directory: %w", err)
	}

	// Create file
	outFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer outFile.Close()

	// Copy content
	if _, err := io.Copy(outFile, reader); err != nil {
		return fmt.Errorf("copying file content: %w", err)
	}

	// Set permissions
	if err := os.Chmod(targetPath, os.FileMode(mode)); err != nil {
		return fmt.Errorf("setting file permissions: %w", err)
	}

	return nil
}

// extractZipFile extracts a single file from ZIP archive
func (e *BinaryExtractor) extractZipFile(file *zip.File, targetPath string) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("creating parent directory: %w", err)
	}

	// Open file in ZIP
	reader, err := file.Open()
	if err != nil {
		return fmt.Errorf("opening file in zip: %w", err)
	}
	defer reader.Close()

	// Create target file
	outFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer outFile.Close()

	// Copy content
	if _, err := io.Copy(outFile, reader); err != nil {
		return fmt.Errorf("copying file content: %w", err)
	}

	// Set permissions
	if err := os.Chmod(targetPath, file.FileInfo().Mode()); err != nil {
		return fmt.Errorf("setting file permissions: %w", err)
	}

	return nil
}

// validatePath checks for path traversal attacks
func (e *BinaryExtractor) validatePath(path string) error {
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains parent directory references: %s", path)
	}
	if strings.HasPrefix(path, "/") {
		return fmt.Errorf("path is absolute: %s", path)
	}
	return nil
}

// findMainExecutable finds the main executable from extracted files
func (e *BinaryExtractor) findMainExecutable(extractedFiles []string, platform string) (string, error) {
	// Define expected executable names by platform
	var expectedNames []string
	if strings.HasPrefix(platform, "windows") {
		expectedNames = []string{"pvm.exe"}
	} else {
		expectedNames = []string{"pvm"}
	}

	// Look for executable in common locations
	candidates := make([]string, 0)

	for _, file := range extractedFiles {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		// Skip non-regular files
		if !info.Mode().IsRegular() {
			continue
		}

		fileName := filepath.Base(file)

		// Check if this matches expected executable name
		for _, expected := range expectedNames {
			if fileName == expected {
				// Check if it's executable on Unix-like systems
				if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
					continue
				}
				candidates = append(candidates, file)
			}
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no executable found in archive (expected: %v)", expectedNames)
	}

	if len(candidates) > 1 {
		return "", fmt.Errorf("multiple executables found: %v", candidates)
	}

	return candidates[0], nil
}

// Cleanup removes the temporary extraction directory
func (e *BinaryExtractor) Cleanup(extractedPath string) error {
	// Find the temp directory (parent of extracted file)
	tempDir := extractedPath
	for {
		parent := filepath.Dir(tempDir)
		if parent == tempDir || !strings.Contains(filepath.Base(tempDir), "pvm-extract-") {
			break
		}
		tempDir = parent
	}

	if strings.Contains(filepath.Base(tempDir), "pvm-extract-") {
		return os.RemoveAll(tempDir)
	}

	return nil
}
