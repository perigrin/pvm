// ABOUTME: Legacy tool integration for plenv and perlbrew
// ABOUTME: Provides functionality to detect and import from existing Perl version managers

package perl

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"tamarou.com/pvm/internal/errors"
)

// userHomeDir is a variable that wraps os.UserHomeDir, making it testable
var userHomeDir = os.UserHomeDir

// Legacy tool error codes
const (
	ErrLegacyToolNotFound    = "201" // Legacy tool installation not found
	ErrLegacyVersionNotFound = "202" // Specific version not found in legacy tool
	ErrLegacyImportFailed    = "203" // Failed to import from legacy tool
)

// LegacyToolType represents the type of legacy Perl version manager
type LegacyToolType string

const (
	Plenv    LegacyToolType = "plenv"
	Perlbrew LegacyToolType = "perlbrew"
)

// LegacyPerl represents a Perl installation from a legacy tool
type LegacyPerl struct {
	// Path to the Perl installation
	Path string

	// Version string
	Version string

	// Source tool (plenv or perlbrew)
	Source LegacyToolType

	// Is this the tool's current/default version?
	IsDefault bool
}

// DetectPlenv checks for plenv installations and returns their paths
func DetectPlenv() ([]LegacyPerl, error) {
	var installations []LegacyPerl

	// Check if plenv is installed
	homeDir, err := userHomeDir()
	if err != nil {
		return nil, errors.NewVersionError(
			ErrLegacyToolNotFound,
			"Failed to determine user home directory",
			err)
	}

	plenvDir := filepath.Join(homeDir, ".plenv")
	versionsDir := filepath.Join(plenvDir, "versions")

	// Check if plenv versions directory exists
	info, err := os.Stat(versionsDir)
	if os.IsNotExist(err) || !info.IsDir() {
		return nil, errors.NewVersionError(
			ErrLegacyToolNotFound,
			fmt.Sprintf("No plenv installation found at %s", versionsDir),
			nil)
	}

	// Get default/global version if available
	defaultVersion := ""
	globalVersionFile := filepath.Join(plenvDir, "version")
	if data, err := os.ReadFile(globalVersionFile); err == nil {
		defaultVersion = strings.TrimSpace(string(data))
	}

	// Read all version directories
	err = filepath.WalkDir(versionsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root versions directory
		if path == versionsDir {
			return nil
		}

		// Only process directories directly under versions/
		if d.IsDir() && filepath.Dir(path) == versionsDir {
			// Extract version from directory name
			version := filepath.Base(path)

			// Check if this is a valid Perl installation by looking for the perl binary
			perlBin := filepath.Join(path, "bin", "perl")
			if _, err := os.Stat(perlBin); err == nil {
				isDefault := version == defaultVersion
				installations = append(installations, LegacyPerl{
					Path:      path,
					Version:   version,
					Source:    Plenv,
					IsDefault: isDefault,
				})
			}

			// Skip subdirectories
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, errors.NewVersionError(
			ErrLegacyImportFailed,
			"Failed to scan plenv installations",
			err)
	}

	if len(installations) == 0 {
		return nil, errors.NewVersionError(
			ErrLegacyVersionNotFound,
			fmt.Sprintf("No valid Perl installations found in plenv (%s)", versionsDir),
			nil)
	}

	return installations, nil
}

// readPerlVersionFileFunc reads a .perl-version file from the specified directory
func readPerlVersionFileFunc(dir string) (string, error) {
	versionFile := filepath.Join(dir, ".perl-version")

	// Check if the file exists
	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		return "", errors.NewVersionError(
			ErrLegacyVersionNotFound,
			fmt.Sprintf("No .perl-version file found in %s", dir),
			nil)
	}

	// Read the file content
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return "", errors.NewVersionError(
			ErrLegacyImportFailed,
			fmt.Sprintf("Failed to read .perl-version file: %s", versionFile),
			err)
	}

	// Trim whitespace and return
	version := strings.TrimSpace(string(data))
	if version == "" {
		return "", errors.NewVersionError(
			ErrLegacyVersionNotFound,
			fmt.Sprintf("Empty .perl-version file: %s", versionFile),
			nil)
	}

	return version, nil
}

// ReadPerlVersionFile is a variable that points to readPerlVersionFileFunc,
// allowing it to be replaced in tests
var ReadPerlVersionFile = readPerlVersionFileFunc

// DetectPerlbrew checks for perlbrew installations and returns their paths
func DetectPerlbrew() ([]LegacyPerl, error) {
	var installations []LegacyPerl

	// Check if perlbrew is installed
	homeDir, err := userHomeDir()
	if err != nil {
		return nil, errors.NewVersionError(
			ErrLegacyToolNotFound,
			"Failed to determine user home directory",
			err)
	}

	// Common perlbrew locations
	perlbrewDirs := []string{
		filepath.Join(homeDir, "perl5", "perlbrew"),
		filepath.Join(homeDir, ".perlbrew"),
	}

	var perlbrewDir string
	for _, dir := range perlbrewDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			perlbrewDir = dir
			break
		}
	}

	if perlbrewDir == "" {
		return nil, errors.NewVersionError(
			ErrLegacyToolNotFound,
			"No perlbrew installation found",
			nil)
	}

	perlsDir := filepath.Join(perlbrewDir, "perls")
	if info, err := os.Stat(perlsDir); os.IsNotExist(err) || !info.IsDir() {
		return nil, errors.NewVersionError(
			ErrLegacyToolNotFound,
			fmt.Sprintf("No perlbrew perls directory found at %s", perlsDir),
			nil)
	}

	// Get current version if available
	defaultVersion := ""
	currentVersionFile := filepath.Join(perlbrewDir, "CURRENT")
	if data, err := os.ReadFile(currentVersionFile); err == nil {
		defaultVersion = strings.TrimSpace(string(data))
	}

	// Read the perls directory
	err = filepath.WalkDir(perlsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root perls directory
		if path == perlsDir {
			return nil
		}

		// Only process directories directly under perls/
		if d.IsDir() && filepath.Dir(path) == perlsDir {
			// Extract version name from directory name
			versionName := filepath.Base(path)

			// Check if this is a valid Perl installation by looking for the perl binary
			perlBin := filepath.Join(path, "bin", "perl")
			if _, err := os.Stat(perlBin); err == nil {
				isDefault := versionName == defaultVersion

				// Clean up the version name (perlbrew often uses names like perl-5.32.1)
				cleanVersion := strings.TrimPrefix(versionName, "perl-")

				installations = append(installations, LegacyPerl{
					Path:      path,
					Version:   cleanVersion,
					Source:    Perlbrew,
					IsDefault: isDefault,
				})
			}

			// Skip subdirectories
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, errors.NewVersionError(
			ErrLegacyImportFailed,
			"Failed to scan perlbrew installations",
			err)
	}

	if len(installations) == 0 {
		return nil, errors.NewVersionError(
			ErrLegacyVersionNotFound,
			fmt.Sprintf("No valid Perl installations found in perlbrew (%s)", perlsDir),
			nil)
	}

	// Read aliases
	aliasFile := filepath.Join(perlbrewDir, "aliases")
	if _, err := os.Stat(aliasFile); err == nil {
		// Parse aliases file if it exists
		file, err := os.Open(aliasFile)
		if err == nil {
			defer func() { _ = file.Close() }()

			// Each line should be in the format: alias_name=actual_name
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "=") {
					parts := strings.SplitN(line, "=", 2)
					aliasName := strings.TrimSpace(parts[0])
					targetName := strings.TrimSpace(parts[1])

					// Potential enhancement: Process aliases and add to installations list
					// Currently just logging for informational purposes
					_ = aliasName
					_ = targetName
				}
			}
		}
	}

	return installations, nil
}

// GetPerlbrewAliases returns a map of perlbrew aliases
func GetPerlbrewAliases() (map[string]string, error) {
	aliases := make(map[string]string)

	// Check if perlbrew is installed
	homeDir, err := userHomeDir()
	if err != nil {
		return nil, errors.NewVersionError(
			ErrLegacyToolNotFound,
			"Failed to determine user home directory",
			err)
	}

	// Common perlbrew locations
	perlbrewDirs := []string{
		filepath.Join(homeDir, "perl5", "perlbrew"),
		filepath.Join(homeDir, ".perlbrew"),
	}

	var perlbrewDir string
	for _, dir := range perlbrewDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			perlbrewDir = dir
			break
		}
	}

	if perlbrewDir == "" {
		return nil, errors.NewVersionError(
			ErrLegacyToolNotFound,
			"No perlbrew installation found",
			nil)
	}

	// Read aliases file
	aliasFile := filepath.Join(perlbrewDir, "aliases")
	if _, err := os.Stat(aliasFile); err != nil {
		return aliases, nil // Return empty map if no aliases file
	}

	file, err := os.Open(aliasFile)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrLegacyImportFailed,
			"Failed to open perlbrew aliases file",
			err)
	}
	defer func() { _ = file.Close() }()

	// Each line should be in the format: alias_name=actual_name
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			aliasName := strings.TrimSpace(parts[0])
			targetName := strings.TrimSpace(parts[1])

			aliases[aliasName] = targetName
		}
	}

	return aliases, nil
}

// findDotPerlVersionFilesFunc searches for .perl-version files starting from the given directory
// and moving up through parent directories until found or reaching the root directory
func findDotPerlVersionFilesFunc(startDir string) ([]string, error) {
	var files []string

	// Ensure we have an absolute path
	absDir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrLegacyImportFailed,
			"Failed to get absolute path",
			err)
	}

	// Iterate up from the starting directory to the root
	currentDir := absDir
	for {
		versionFile := filepath.Join(currentDir, ".perl-version")
		if _, err := os.Stat(versionFile); err == nil {
			files = append(files, versionFile)
		}

		// Move up to parent directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// We've reached the root directory
			break
		}
		currentDir = parentDir
	}

	return files, nil
}

// FindDotPerlVersionFiles is a variable that points to findDotPerlVersionFilesFunc,
// allowing it to be replaced in tests
var FindDotPerlVersionFiles = findDotPerlVersionFilesFunc

// ImportFromLegacyTool imports Perl installations from a legacy tool
func ImportFromLegacyTool(toolType LegacyToolType) ([]LegacyPerl, error) {
	switch toolType {
	case Plenv:
		return DetectPlenv()
	case Perlbrew:
		return DetectPerlbrew()
	default:
		return nil, errors.NewVersionError(
			ErrLegacyToolNotFound,
			fmt.Sprintf("Unsupported legacy tool: %s", toolType),
			nil)
	}
}
