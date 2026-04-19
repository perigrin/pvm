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

	if err := applyExecutablePermissions(targetPath, os.FileMode(mode)); err != nil {
		return err
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

	if err := applyExecutablePermissions(targetPath, file.FileInfo().Mode()); err != nil {
		return err
	}

	return nil
}

// applyExecutablePermissions sets the mode on an extracted file, masking
// POSIX file-type bits (S_IFREG 0100000) out of tar header modes and ZIP
// FileInfo().Mode() before passing to os.Chmod. The target binary's
// executable bit is guaranteed separately by the updater's call to
// platform.MakeExecutable in internal/updater/updater.go — doing it here
// too would over-chmod every archive entry (man pages, completions) and
// silently mark them executable.
func applyExecutablePermissions(targetPath string, raw os.FileMode) error {
	perm := raw & os.ModePerm
	if err := os.Chmod(targetPath, perm); err != nil {
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
	// Define expected executable names by platform.
	// Prefer the canonical name (pvm / pvm.exe) but also accept
	// platform-suffixed names (pvm-linux-amd64, pvm-darwin-arm64)
	// for backwards compatibility with older release archives.
	var expectedNames []string
	if strings.HasPrefix(platform, "windows") {
		expectedNames = []string{"pvm.exe", "pvm-" + platform + ".exe"}
	} else {
		expectedNames = []string{"pvm", "pvm-" + platform}
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

		// Check if this matches an expected executable name. We match
		// by name only — mode bits are unreliable on cross-platform
		// tarballs (a packaging step occasionally strips +x). When a
		// matching candidate is found without the exec bit, re-apply
		// 0755 so downstream validation succeeds. This is scoped to the
		// single binary we care about; other archive entries are left
		// with whatever mode the archive specified.
		for _, expected := range expectedNames {
			if fileName != expected {
				continue
			}
			if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
				if err := os.Chmod(file, 0755); err != nil {
					return "", fmt.Errorf("setting executable bit on %s: %w", file, err)
				}
			}
			candidates = append(candidates, file)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no executable found in archive (expected: %v)", expectedNames)
	}

	if len(candidates) > 1 {
		// If candidates have different base names (e.g. "pvm" and "pvm-linux-amd64"),
		// prefer the canonical name (first in expectedNames). If they have the same
		// base name in different directories, that is genuinely ambiguous.
		canonicalName := expectedNames[0]
		var canonicalCandidates []string
		for _, c := range candidates {
			if filepath.Base(c) == canonicalName {
				canonicalCandidates = append(canonicalCandidates, c)
			}
		}
		if len(canonicalCandidates) == 1 {
			return canonicalCandidates[0], nil
		}
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
