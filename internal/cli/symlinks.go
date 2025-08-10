// ABOUTME: Symlink management for multi-entry point CLI
// ABOUTME: Creates and manages symlinks for the different binary entry points

package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"tamarou.com/pvm/internal/platform"
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
		ComponentPM,
		ComponentPSC,
	}

	// Create symlinks for each component
	for _, component := range components {
		// Skip if the binary is already named after this component
		if base == component || base == component+".exe" {
			continue
		}

		// Create the symlink path
		symlinkPath := filepath.Join(dir, platform.ExecutableName(component))

		// Remove existing symlink if it exists
		_ = os.Remove(symlinkPath) // Ignoring error as it's OK if the file doesn't exist

		// Create the link (symlink on Unix, hard link/copy on Windows)
		err := platform.CreateLink(binaryPath, symlinkPath)

		if err != nil {
			return symlinks, fmt.Errorf("failed to create symlink for %s: %v", component, err)
		}

		// Store the symlink path
		symlinks[component] = symlinkPath
	}

	return symlinks, nil
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
		ComponentPM,
		ComponentPSC,
	}

	// Check each component
	for _, component := range components {
		symlinkPath := filepath.Join(dir, platform.ExecutableName(component))

		// Check if the symlink exists
		_, err := os.Stat(symlinkPath)
		status[component] = err == nil
	}

	return status
}
