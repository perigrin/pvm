// ABOUTME: Automatic import functionality for plenv and perlbrew installations
// ABOUTME: Provides seamless migration during pvm init without user intervention

package perl

import (
	"fmt"
	"os"
	"path/filepath"

	"tamarou.com/pvm/internal/xdg"
)

// ImportResult represents the result of an import operation
type ImportResult struct {
	// Version that was imported
	Version string

	// Source tool (plenv or perlbrew)
	Source LegacyToolType

	// Path to the original installation
	OriginalPath string

	// Path to the symlink created by PVM
	SymlinkPath string

	// Whether this was set as default
	WasDefault bool
}

// AutoImportResults represents the results of automatic import during init
type AutoImportResults struct {
	// Plenv imports
	PlenvImports []ImportResult

	// Perlbrew imports
	PerlbrewImports []ImportResult

	// Total number of versions imported
	TotalImported int

	// Default version that was set (if any)
	DefaultVersion string
}

// AutoImportLegacyVersions automatically detects and imports legacy tool installations
// This is designed to be called during `pvm init` to provide seamless migration
func AutoImportLegacyVersions() (*AutoImportResults, error) {
	results := &AutoImportResults{}

	// Try to import from plenv
	plenvResults, err := importFromPlenv()
	if err != nil {
		// Log error but don't fail - plenv might not be installed
		// In a production version, we'd use a proper logger here
	} else {
		results.PlenvImports = plenvResults
		results.TotalImported += len(plenvResults)
	}

	// Try to import from perlbrew
	perlbrewResults, err := importFromPerlbrew()
	if err != nil {
		// Log error but don't fail - perlbrew might not be installed
		// In a production version, we'd use a proper logger here
	} else {
		results.PerlbrewImports = perlbrewResults
		results.TotalImported += len(perlbrewResults)
	}

	// Set default version if any was marked as default
	for _, result := range results.PlenvImports {
		if result.WasDefault {
			results.DefaultVersion = result.Version
			break
		}
	}
	if results.DefaultVersion == "" {
		for _, result := range results.PerlbrewImports {
			if result.WasDefault {
				results.DefaultVersion = result.Version
				break
			}
		}
	}

	return results, nil
}

// importFromPlenv detects and imports plenv installations
func importFromPlenv() ([]ImportResult, error) {
	var results []ImportResult

	// Detect plenv installations
	installations, err := DetectPlenv()
	if err != nil {
		// Return empty results if plenv is not found
		return results, nil
	}

	// Get XDG directories for PVM installations
	dirs, err := xdg.GetDirs()
	if err != nil {
		return nil, err
	}

	versionsDir := filepath.Join(dirs.DataDir, "versions")
	if err := os.MkdirAll(versionsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create versions directory: %w", err)
	}

	// Import each installation
	for _, inst := range installations {
		// Check if already imported
		isInstalled, err := IsVersionInstalled(inst.Version)
		if err != nil {
			continue // Skip on error
		}
		if isInstalled {
			continue // Already imported
		}

		// Create symlink
		symlinkPath := filepath.Join(versionsDir, inst.Version)
		err = os.Symlink(inst.Path, symlinkPath)
		if err != nil {
			continue // Skip on error
		}

		// Register in PVM registry
		versionInfo := VersionInfo{
			Version:     inst.Version,
			InstallPath: symlinkPath, // Use symlink path
			InstallTime: inst.InstallTime,
			Source:      string(inst.Source),
		}

		err = RegisterVersion(versionInfo)
		if err != nil {
			// Clean up symlink on registration failure
			os.Remove(symlinkPath)
			continue
		}

		results = append(results, ImportResult{
			Version:      inst.Version,
			Source:       inst.Source,
			OriginalPath: inst.Path,
			SymlinkPath:  symlinkPath,
			WasDefault:   inst.IsDefault,
		})
	}

	return results, nil
}

// importFromPerlbrew detects and imports perlbrew installations
func importFromPerlbrew() ([]ImportResult, error) {
	var results []ImportResult

	// Detect perlbrew installations
	installations, err := DetectPerlbrew()
	if err != nil {
		// Return empty results if perlbrew is not found
		return results, nil
	}

	// Get XDG directories for PVM installations
	dirs, err := xdg.GetDirs()
	if err != nil {
		return nil, err
	}

	versionsDir := filepath.Join(dirs.DataDir, "versions")
	if err := os.MkdirAll(versionsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create versions directory: %w", err)
	}

	// Import each installation
	for _, inst := range installations {
		// Check if already imported
		isInstalled, err := IsVersionInstalled(inst.Version)
		if err != nil {
			continue // Skip on error
		}
		if isInstalled {
			continue // Already imported
		}

		// Create symlink
		symlinkPath := filepath.Join(versionsDir, inst.Version)
		err = os.Symlink(inst.Path, symlinkPath)
		if err != nil {
			continue // Skip on error
		}

		// Register in PVM registry
		versionInfo := VersionInfo{
			Version:     inst.Version,
			InstallPath: symlinkPath, // Use symlink path
			InstallTime: inst.InstallTime,
			Source:      string(inst.Source),
		}

		err = RegisterVersion(versionInfo)
		if err != nil {
			// Clean up symlink on registration failure
			os.Remove(symlinkPath)
			continue
		}

		results = append(results, ImportResult{
			Version:      inst.Version,
			Source:       inst.Source,
			OriginalPath: inst.Path,
			SymlinkPath:  symlinkPath,
			WasDefault:   inst.IsDefault,
		})
	}

	return results, nil
}

// ShouldAutoImport checks if auto-import should be performed
// Returns true if this is the first time pvm init is being run
func ShouldAutoImport() bool {
	// Check if we already have versions registered
	versions, err := GetInstalledVersions()
	if err != nil {
		return true // If we can't check, assume first run
	}

	// If we have no versions installed, this is likely the first run
	return len(versions) == 0
}

// PrintAutoImportResults prints the results of auto-import to stderr
// This ensures the output doesn't interfere with shell eval
func PrintAutoImportResults(results *AutoImportResults) {
	if results.TotalImported == 0 {
		return // Nothing to report
	}

	// Print plenv imports
	if len(results.PlenvImports) > 0 {
		fmt.Fprintf(os.Stderr, "Detected plenv installation with %d versions - importing automatically...\n", len(results.PlenvImports))
		for _, result := range results.PlenvImports {
			defaultMark := ""
			if result.WasDefault {
				defaultMark = " (set as default)"
			}
			fmt.Fprintf(os.Stderr, "✓ Imported %s from plenv%s\n", result.Version, defaultMark)
		}
	}

	// Print perlbrew imports
	if len(results.PerlbrewImports) > 0 {
		fmt.Fprintf(os.Stderr, "Detected perlbrew installation with %d versions - importing automatically...\n", len(results.PerlbrewImports))
		for _, result := range results.PerlbrewImports {
			defaultMark := ""
			if result.WasDefault {
				defaultMark = " (set as default)"
			}
			fmt.Fprintf(os.Stderr, "✓ Imported %s from perlbrew%s\n", result.Version, defaultMark)
		}
	}
}
