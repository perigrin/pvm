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
	"sort"
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

// DirectoryCandidate represents a potential root directory
type DirectoryCandidate struct {
	Path         string
	HasBuildFile bool
	BuildFile    string
	Priority     int
}

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

	// Keep track of potential root directories
	seenDirs := make(map[string]bool)
	allEntries := make(map[string]tar.Header)

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

		// Store all entries for later analysis
		allEntries[name] = *header

		// Track all top-level directories
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
			// Sanitize symlink target for security
			linkTarget := sanitizePath(header.Linkname)
			if linkTarget == "" {
				log.Warnf("Skipping unsafe symlink target: %s", header.Linkname)
				continue
			}

			// Ensure symlink target is relative (prevent absolute path attacks)
			if filepath.IsAbs(linkTarget) {
				log.Warnf("Skipping absolute symlink target: %s", header.Linkname)
				continue
			}

			// Create parent directory if doesn't exist
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return nil, errors.NewSystemError(
					ErrExtractionFailed,
					"Failed to create parent directory for symlink",
					err).
					WithLocation(filepath.Dir(target))
			}

			// Create symlink with sanitized target
			if err := os.Symlink(linkTarget, target); err != nil {
				return nil, errors.NewSystemError(
					ErrExtractionFailed,
					"Failed to create symlink",
					err).
					WithLocation(target)
			}
		}
	}

	// Determine the best root directory using improved logic
	rootDir, err := selectBestRootDirectory(targetDir, seenDirs, allEntries)
	if err != nil {
		return nil, errors.NewSystemError(
			ErrExtractionFailed,
			"Failed to determine root directory",
			err).WithLocation(targetDir)
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
	// Reject Unix-style absolute paths before converting separators,
	// since filepath.IsAbs on Windows won't detect "/etc/passwd" as absolute
	if strings.HasPrefix(path, "/") {
		return ""
	}

	// Convert to platform-specific path separator
	path = filepath.FromSlash(path)

	// Reject absolute paths for security (catches Windows-style C:\ and UNC \\)
	if filepath.IsAbs(path) {
		return ""
	}

	// Handle ../
	if strings.Contains(path, "..") {
		return ""
	}

	// Clean the path to handle any remaining irregularities
	path = filepath.Clean(path)

	// Double-check - reject any path that's still absolute after cleaning
	if filepath.IsAbs(path) {
		return ""
	}

	return path
}

// selectBestRootDirectory analyzes potential root directories and selects the best one
func selectBestRootDirectory(targetDir string, seenDirs map[string]bool, allEntries map[string]tar.Header) (string, error) {
	if len(seenDirs) == 0 {
		return "", fmt.Errorf("no directories found in archive")
	}

	// Build system files in priority order
	buildFiles := []string{
		"Build.PL",    // Module::Build (highest priority)
		"Makefile.PL", // ExtUtils::MakeMaker
		"Makefile",    // Pre-built Makefile (lowest priority)
	}

	// Create candidates for each potential root directory
	var candidates []DirectoryCandidate
	for dirName := range seenDirs {
		candidate := DirectoryCandidate{
			Path:     dirName,
			Priority: 0, // Default priority
		}

		// Check if this directory contains build system files
		// Use forward slashes for lookup since tar archive entries use forward slashes
		for i, buildFile := range buildFiles {
			buildPath := dirName + "/" + buildFile
			if _, exists := allEntries[buildPath]; exists {
				candidate.HasBuildFile = true
				candidate.BuildFile = buildFile
				// Higher priority for preferred build systems (Build.PL = 300, Makefile.PL = 200, Makefile = 100)
				candidate.Priority = 300 - (i * 100)
				break
			}
		}

		candidates = append(candidates, candidate)
	}

	// Sort candidates by priority (highest first), then alphabetically for deterministic behavior
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Priority != candidates[j].Priority {
			return candidates[i].Priority > candidates[j].Priority
		}
		return candidates[i].Path < candidates[j].Path
	})

	// Select the best candidate
	selected := candidates[0]

	// Log the selection decision for debugging
	log.Debugf("Root directory selection for %s:", targetDir)
	log.Debugf("  Found %d potential directories:", len(candidates))
	for _, candidate := range candidates {
		if candidate.HasBuildFile {
			log.Debugf("    %s (build file: %s, priority: %d)", candidate.Path, candidate.BuildFile, candidate.Priority)
		} else {
			log.Debugf("    %s (no build file, priority: %d)", candidate.Path, candidate.Priority)
		}
	}
	log.Debugf("  Selected: %s", selected.Path)

	// Warn if we're selecting a directory without build files
	if !selected.HasBuildFile {
		log.Warnf("Selected root directory '%s' does not contain build system files (Build.PL, Makefile.PL, Makefile)", selected.Path)
		log.Warnf("This may cause build failures - consider checking archive structure")
	}

	return selected.Path, nil
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
