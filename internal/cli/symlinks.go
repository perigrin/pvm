// ABOUTME: Symlink management for multi-entry point CLI
// ABOUTME: Creates and manages symlinks for the different binary entry points

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// CreateSymlinks creates symlinks for all components based on the provided binary path
// Returns a map of component names to the paths of created symlinks
func CreateSymlinks(binaryPath string) (map[string]string, error) {
	// Get the directory of the binary
	dir := filepath.Dir(binaryPath)

	// Get the base name of the binary
	base := filepath.Base(binaryPath)

	// Create a map to store the symlink paths
	symlinks := make(map[string]string)

	// List of components to create symlinks for
	components := []string{
		ComponentPVM,
		ComponentPVX,
		ComponentPVI,
		ComponentPSC,
	}

	// Create symlinks for each component
	for _, component := range components {
		// Skip if the binary is already named after this component
		if base == component || base == component+".exe" {
			continue
		}

		// Create the symlink path
		var symlinkPath string
		if runtime.GOOS == "windows" {
			symlinkPath = filepath.Join(dir, component+".exe")
		} else {
			symlinkPath = filepath.Join(dir, component)
		}

		// Remove existing symlink if it exists
		os.Remove(symlinkPath)

		// Create the symlink
		var err error
		if runtime.GOOS == "windows" {
			// Windows requires different handling - we copy the binary instead
			// as creating symlinks requires admin privileges
			err = os.Link(binaryPath, symlinkPath)
			if err != nil {
				// If hard link fails, try to copy the file
				err = copyFile(binaryPath, symlinkPath)
			}
		} else {
			// On Unix, we can create a symlink
			err = os.Symlink(binaryPath, symlinkPath)
		}

		if err != nil {
			return symlinks, fmt.Errorf("failed to create symlink for %s: %v", component, err)
		}

		// Store the symlink path
		symlinks[component] = symlinkPath
	}

	return symlinks, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Read the source file
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Write to the destination file
	return os.WriteFile(dst, data, 0755)
}

// VerifySymlinks checks if all symlinks for the different components exist
// Returns a map of component names to symlink status (true if exists, false otherwise)
func VerifySymlinks(binaryPath string) map[string]bool {
	// Get the directory of the binary
	dir := filepath.Dir(binaryPath)

	// Create a map to store the symlink status
	status := make(map[string]bool)

	// List of components to check
	components := []string{
		ComponentPVM,
		ComponentPVX,
		ComponentPVI,
		ComponentPSC,
	}

	// Check each component
	for _, component := range components {
		var symlinkPath string
		if runtime.GOOS == "windows" {
			symlinkPath = filepath.Join(dir, component+".exe")
		} else {
			symlinkPath = filepath.Join(dir, component)
		}

		// Check if the symlink exists
		_, err := os.Stat(symlinkPath)
		status[component] = err == nil
	}

	return status
}
