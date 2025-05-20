// ABOUTME: Module extraction functionality
// ABOUTME: Functions for extracting CPAN module archives

package modules

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// Error codes for extraction operations
const (
	ErrExtractionFailed = "PVI-4101" // Failed to extract module archive
	ErrBadArchiveFormat = "PVI-4102" // Unsupported or corrupted archive format
	ErrBuildDirFailed   = "PVI-4103" // Failed to create build directory
)

// ExtractionResult contains information about the extracted module
type ExtractionResult struct {
	// Path to the extracted directory
	ExtractedDir string

	// Module name
	ModuleName string

	// Original archive path
	ArchivePath string

	// Distribution name from CPAN
	Distribution string

	// Root directory of the extraction
	RootDir string
}

// ExtractModuleArchiveFunc is the function type for extracting a module archive
type ExtractModuleArchiveFunc func(archivePath, targetDir string, ctx context.Context) (*ExtractionResult, error)

// ExtractModuleArchive is a variable that holds the module extraction function
// It can be replaced in tests
var ExtractModuleArchive ExtractModuleArchiveFunc = extractModuleArchive

// extractModuleArchive is the actual implementation of the module extraction function
func extractModuleArchive(archivePath, targetDir string, ctx context.Context) (*ExtractionResult, error) {
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, errors.NewSystemError(
			ErrBuildDirFailed,
			"Failed to create target directory for extraction",
			err).
			WithLocation(targetDir)
	}

	// Open the archive file
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, errors.NewSystemError(
			ErrExtractionFailed,
			"Failed to open archive file",
			err).
			WithLocation(archivePath)
	}
	defer func() { _ = file.Close() }()

	// Create gzip reader for .tar.gz files
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, errors.NewSystemError(
			ErrBadArchiveFormat,
			"Failed to create gzip reader",
			err).
			WithLocation(archivePath)
	}
	defer func() { _ = gzr.Close() }()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Keep track of the root directory
	var rootDir string
	seenDirs := make(map[string]bool)

	// Extract each file
	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, errors.NewSystemError(
				ErrExtractionFailed,
				"Extraction cancelled",
				ctx.Err())
		default:
			// Continue processing
		}

		// Get next header
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, errors.NewSystemError(
				ErrExtractionFailed,
				"Failed to read archive header",
				err).
				WithLocation(archivePath)
		}

		// Normalize the path to avoid directory traversal attacks
		name := sanitizePath(header.Name)
		if name == "" {
			continue // Skip if path is unsafe
		}

		// Construct the target path
		target := filepath.Join(targetDir, name)

		// Check for directory traversal attacks - redundant but a good safety measure
		if !strings.HasPrefix(target, targetDir) {
			log.Warnf("Skipping potentially unsafe path: %s", header.Name)
			continue
		}

		// Track directories - detect root directory
		if filepath.Dir(name) == "." && header.Typeflag == tar.TypeDir {
			rootDir = name
		}

		// Remember all top-level directories we see
		dirName := strings.Split(name, string(os.PathSeparator))[0]
		if dirName != "" && !seenDirs[dirName] {
			seenDirs[dirName] = true
		}

		// Process based on file type
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, 0755); err != nil {
				return nil, errors.NewSystemError(
					ErrExtractionFailed,
					"Failed to create directory",
					err).
					WithLocation(target)
			}

		case tar.TypeReg:
			// Create parent directory if doesn't exist
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return nil, errors.NewSystemError(
					ErrExtractionFailed,
					"Failed to create parent directory",
					err).
					WithLocation(filepath.Dir(target))
			}

			// Create file
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return nil, errors.NewSystemError(
					ErrExtractionFailed,
					"Failed to create file",
					err).
					WithLocation(target)
			}

			// Copy content
			if _, err := io.Copy(f, tr); err != nil {
				_ = f.Close()
				return nil, errors.NewSystemError(
					ErrExtractionFailed,
					"Failed to write file content",
					err).
					WithLocation(target)
			}

			// Close file
			if err := f.Close(); err != nil {
				return nil, errors.NewSystemError(
					ErrExtractionFailed,
					"Failed to close file",
					err).
					WithLocation(target)
			}

		case tar.TypeSymlink:
			// Create parent directory if doesn't exist
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return nil, errors.NewSystemError(
					ErrExtractionFailed,
					"Failed to create parent directory for symlink",
					err).
					WithLocation(filepath.Dir(target))
			}

			// Create symlink
			if err := os.Symlink(header.Linkname, target); err != nil {
				return nil, errors.NewSystemError(
					ErrExtractionFailed,
					"Failed to create symlink",
					err).
					WithLocation(target)
			}
		}
	}

	// If we couldn't determine the root directory, use the first one we saw
	if rootDir == "" && len(seenDirs) > 0 {
		for dir := range seenDirs {
			rootDir = dir
			break
		}
	}

	// Get the extracted directory path
	extractedDir := filepath.Join(targetDir, rootDir)

	// Try to infer module and distribution names from the archive path
	baseName := filepath.Base(archivePath)
	baseName = strings.TrimSuffix(baseName, ".tar.gz")

	// Parse distribution name from the archive path
	distName := baseName
	dashIndex := strings.LastIndex(baseName, "-")
	if dashIndex > 0 {
		// Try to separate distribution name from version
		distName = baseName[:dashIndex]
	}

	// Create result
	result := &ExtractionResult{
		ExtractedDir: extractedDir,
		ModuleName:   "", // Will be determined later from metadata
		ArchivePath:  archivePath,
		Distribution: distName,
		RootDir:      rootDir,
	}

	return result, nil
}

// sanitizePath normalizes a file path from an archive
func sanitizePath(path string) string {
	// Convert to platform-specific path separator
	path = filepath.FromSlash(path)

	// Remove any root component
	path = filepath.Join(filepath.SplitList(path)...)

	// Handle Absolute paths
	if filepath.IsAbs(path) {
		path = path[1:]
	}

	// Handle ../
	if strings.Contains(path, "..") {
		return ""
	}

	return path
}

// DetectBuildSystem determines the build system used by a module
func DetectBuildSystem(dir string) (string, error) {
	// Check for various build files in order of preference
	buildFiles := []string{
		"Build.PL",    // Module::Build
		"Makefile.PL", // ExtUtils::MakeMaker
		"Makefile",    // Pre-built Makefile
	}

	for _, file := range buildFiles {
		path := filepath.Join(dir, file)
		if _, err := os.Stat(path); err == nil {
			return file, nil
		}
	}

	return "", fmt.Errorf("no supported build system found in %s", dir)
}
