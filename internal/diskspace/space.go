// ABOUTME: Cross-platform disk space utilities for PVM
// ABOUTME: Provides APIs for checking available space and calculating directory sizes
package diskspace

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// SpaceInfo represents disk space information for a filesystem
type SpaceInfo struct {
	Total     int64 // Total space in bytes
	Free      int64 // Free space in bytes
	Available int64 // Available space to current user in bytes
}

// GetSpaceInfo returns disk space information for the filesystem containing the given path
func GetSpaceInfo(path string) (*SpaceInfo, error) {
	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving absolute path: %w", err)
	}

	originalPath := absPath

	// Find existing directory in the path hierarchy
	for {
		if _, err := os.Stat(absPath); err == nil {
			break
		}
		parent := filepath.Dir(absPath)
		if parent == absPath {
			return nil, fmt.Errorf("no existing directory found in path hierarchy for %s", originalPath)
		}
		absPath = parent
	}

	return getSpaceInfo(absPath)
}

// CalculateDirectorySize calculates the total size of a directory and all its contents
func CalculateDirectorySize(dirPath string) (int64, error) {
	// Check if directory exists first
	if _, err := os.Stat(dirPath); err != nil {
		return 0, fmt.Errorf("accessing directory %s: %w", dirPath, err)
	}

	var totalSize int64

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip files/directories we can't access
			return nil
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				// Skip files we can't stat
				return nil
			}
			totalSize += info.Size()
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("calculating directory size for %s: %w", dirPath, err)
	}

	return totalSize, nil
}
