// ABOUTME: Perl installation registry for tracking installed versions
// ABOUTME: Provides functions to register, list, and uninstall Perl versions

package perl

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/xdg"
)

// Wrappers for easier mocking in tests
var osRemoveAll = os.RemoveAll
var detectSystemPerl = DetectSystemPerl
var osStat = os.Stat
var ioutilReadFile = os.ReadFile
var ioutilWriteFile = os.WriteFile
var osRename = os.Rename

// Registry error codes
const (
	ErrRegistryIOFailed   = "601" // Failed to read/write registry file
	ErrVersionNotFound    = "602" // Requested version not found in registry
	ErrVersionExists      = "603" // Version already exists in registry
	ErrInvalidRegistry    = "604" // Invalid registry data format
	ErrUninstallFailed    = "605" // Failed to uninstall version
	ErrVersionInvalid     = "606" // Invalid version name
	ErrRegistryLockFailed = "607" // Failed to acquire registry lock
)

// registryFileName is the name of the registry file
const registryFileName = "registry.json"

// generateUUID generates a simple UUID for registry entries
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// VersionInfo contains metadata about an installed Perl version
type VersionInfo struct {
	// Version string in normalized format (X.Y.Z)
	Version string `json:"version"`

	// InstallPath is the absolute path to the installation directory
	InstallPath string `json:"install_path"`

	// InstallTime is when the version was installed
	InstallTime time.Time `json:"install_time"`

	// Source is where the version came from: "pvm", "plenv", "perlbrew", "system"
	Source string `json:"source"`

	// BuildOptions contains the options used to build this version (if built by pvm)
	BuildOptions *BuildOptions `json:"build_options,omitempty"`
}

// VersionRegistry holds information about all installed Perl versions
type VersionRegistry struct {
	// Map of UUID -> VersionInfo
	Versions map[string]VersionInfo `json:"versions"`
}

// loadRegistryFunc loads the version registry from disk
func loadRegistryFunc() (*VersionRegistry, error) {
	// Initialize empty registry
	registry := &VersionRegistry{
		Versions: make(map[string]VersionInfo),
	}

	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return registry, errors.NewSystemError("001",
			"Failed to determine XDG directories", err)
	}

	// Ensure data directory exists
	err = dirs.EnsureDirs()
	if err != nil {
		return registry, errors.NewSystemError("002",
			"Failed to create required directories", err)
	}

	// Path to registry file
	registryPath := filepath.Join(dirs.DataDir, registryFileName)

	// Check if registry file exists
	if _, err := osStat(registryPath); os.IsNotExist(err) {
		// Registry file doesn't exist yet, return empty registry
		return registry, nil
	}

	// Read registry file
	data, err := ioutilReadFile(registryPath)
	if err != nil {
		return registry, errors.NewVersionError(
			ErrRegistryIOFailed,
			"Failed to read registry file",
			err).
			WithLocation(registryPath)
	}

	// Parse JSON
	if len(data) > 0 {
		err = json.Unmarshal(data, registry)
		if err != nil {
			return registry, errors.NewVersionError(
				ErrInvalidRegistry,
				"Failed to parse registry file",
				err).
				WithLocation(registryPath)
		}
	}

	return registry, nil
}

// LoadRegistry is a variable that points to loadRegistryFunc,
// allowing it to be replaced in tests
var LoadRegistry = loadRegistryFunc

// saveRegistryFunc saves the version registry to disk
func saveRegistryFunc(registry *VersionRegistry) error {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return errors.NewSystemError("001",
			"Failed to determine XDG directories", err)
	}

	// Ensure data directory exists
	err = dirs.EnsureDirs()
	if err != nil {
		return errors.NewSystemError("002",
			"Failed to create required directories", err)
	}

	// Path to registry file
	registryPath := filepath.Join(dirs.DataDir, registryFileName)

	// Convert registry to JSON
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return errors.NewVersionError(
			ErrRegistryIOFailed,
			"Failed to encode registry to JSON",
			err)
	}

	// Write to temporary file first to ensure atomic update
	tempPath := registryPath + ".tmp"
	err = ioutilWriteFile(tempPath, data, 0644)
	if err != nil {
		return errors.NewVersionError(
			ErrRegistryIOFailed,
			"Failed to write registry file",
			err).
			WithLocation(tempPath)
	}

	// Rename temporary file to final path (atomic operation)
	err = osRename(tempPath, registryPath)
	if err != nil {
		return errors.NewVersionError(
			ErrRegistryIOFailed,
			"Failed to update registry file",
			err).
			WithLocation(registryPath)
	}

	return nil
}

// SaveRegistry is a variable that points to saveRegistryFunc,
// allowing it to be replaced in tests
var SaveRegistry = saveRegistryFunc

// RegisterVersion adds a new Perl version to the registry
func RegisterVersion(versionInfo VersionInfo) error {
	// Validate version
	parsedVersion, err := ParseVersion(versionInfo.Version)
	if err != nil {
		return errors.NewVersionError(
			ErrVersionInvalid,
			fmt.Sprintf("Invalid version format: %s", versionInfo.Version),
			err)
	}

	// Make sure we use the normalized version format
	versionInfo.Version = parsedVersion.String()

	// Load existing registry
	registry, err := LoadRegistry()
	if err != nil {
		return err
	}

	// Add to registry with new UUID (no duplicate checking)
	uuid := generateUUID()
	registry.Versions[uuid] = versionInfo

	// Save updated registry
	return SaveRegistry(registry)
}

// Variable for easier mocking in tests
var registerVersion = RegisterVersion

// RegisterVersionAfterBuild registers a version after successful build
func RegisterVersionAfterBuild(buildResult *BuildResult, source string) error {
	// Convert build result to version info
	versionInfo := VersionInfo{
		Version:      buildResult.Version,
		InstallPath:  buildResult.InstallPath,
		InstallTime:  time.Now(),
		Source:       source,
		BuildOptions: nil, // Could be added if needed
	}

	// Register the version
	return RegisterVersion(versionInfo)
}

// getInstalledVersionsFunc returns a list of all installed Perl versions
func getInstalledVersionsFunc() ([]VersionInfo, error) {
	// Load registry
	registry, err := LoadRegistry()
	if err != nil {
		return nil, err
	}

	// Extract values from map into a slice
	versions := make([]VersionInfo, 0, len(registry.Versions))
	for _, versionInfo := range registry.Versions {
		versions = append(versions, versionInfo)
	}

	// Sort versions by version number (descending), with string versions at the end
	sort.Slice(versions, func(i, j int) bool {
		versionI, errI := ParseVersion(versions[i].Version)
		versionJ, errJ := ParseVersion(versions[j].Version)

		// Both are valid semantic versions - compare numerically (newest first)
		if errI == nil && errJ == nil {
			return versionI.Compare(versionJ) > 0
		}

		// Version i is valid, j is string - i comes first
		if errI == nil && errJ != nil {
			return true
		}

		// Version j is valid, i is string - j comes first
		if errI != nil && errJ == nil {
			return false
		}

		// Both are string versions - sort alphabetically
		return versions[i].Version < versions[j].Version
	})

	return versions, nil
}

// GetInstalledVersions is a variable that points to getInstalledVersionsFunc,
// allowing it to be replaced in tests
var GetInstalledVersions = getInstalledVersionsFunc

// GetVersionInfo returns information about a specific installed version
func GetVersionInfo(version string) (*VersionInfo, error) {
	// Parse version to normalize format
	parsedVersion, err := ParseVersion(version)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrVersionInvalid,
			fmt.Sprintf("Invalid version format: %s", version),
			err)
	}
	normalizedVersion := parsedVersion.String()

	// Load registry
	registry, err := LoadRegistry()
	if err != nil {
		return nil, err
	}

	// Look up version by searching through all entries
	for _, versionInfo := range registry.Versions {
		if versionInfo.Version == normalizedVersion {
			return &versionInfo, nil
		}
	}

	return nil, errors.NewVersionError(
		ErrVersionNotFound,
		fmt.Sprintf("Version %s is not installed", normalizedVersion),
		nil)
}

// isVersionInstalledFunc checks if a specific version is installed
func isVersionInstalledFunc(version string) (bool, error) {
	// Parse version to normalize format
	parsedVersion, err := ParseVersion(version)
	if err != nil {
		return false, errors.NewVersionError(
			ErrVersionInvalid,
			fmt.Sprintf("Invalid version format: %s", version),
			err)
	}
	normalizedVersion := parsedVersion.String()

	// Load registry
	registry, err := LoadRegistry()
	if err != nil {
		return false, err
	}

	// Check if version exists in registry by searching through all entries
	for _, versionInfo := range registry.Versions {
		if versionInfo.Version == normalizedVersion {
			return true, nil
		}
	}
	return false, nil
}

// IsVersionInstalled is a variable that points to isVersionInstalledFunc,
// allowing it to be replaced in tests
var IsVersionInstalled = isVersionInstalledFunc

// UninstallVersion removes a Perl version from the system
func UninstallVersion(version string) error {
	// Parse version to normalize format
	parsedVersion, err := ParseVersion(version)
	if err != nil {
		return errors.NewVersionError(
			ErrVersionInvalid,
			fmt.Sprintf("Invalid version format: %s", version),
			err)
	}
	normalizedVersion := parsedVersion.String()

	// Load registry
	registry, err := LoadRegistry()
	if err != nil {
		return err
	}

	// Check if version exists
	versionInfo, exists := registry.Versions[normalizedVersion]
	if !exists {
		return errors.NewVersionError(
			ErrVersionNotFound,
			fmt.Sprintf("Version %s is not installed", normalizedVersion),
			nil)
	}

	// If this is a system Perl, don't remove it, just unregister
	if versionInfo.Source == "system" {
		delete(registry.Versions, normalizedVersion)
		return SaveRegistry(registry)
	}

	// Remove installation directory
	err = osRemoveAll(versionInfo.InstallPath)
	if err != nil {
		return errors.NewVersionError(
			ErrUninstallFailed,
			fmt.Sprintf("Failed to remove installation directory for version %s", normalizedVersion),
			err).
			WithLocation(versionInfo.InstallPath)
	}

	// Remove from registry
	delete(registry.Versions, normalizedVersion)

	// Save updated registry
	return SaveRegistry(registry)
}

// ImportSystemPerl imports the system Perl into the registry
func ImportSystemPerl() error {
	// Detect system Perl
	systemPerl, err := DetectSystemPerl()
	if err != nil {
		return err
	}

	// Create version info
	// For system perl, InstallPath should be the directory containing the perl executable
	installPath := filepath.Dir(systemPerl.Path)
	versionInfo := VersionInfo{
		Version:     systemPerl.Version,
		InstallPath: installPath,
		InstallTime: time.Now(),
		Source:      "system",
	}

	// Register it
	return registerVersion(versionInfo)
}

// GetAvailableVersions returns a list of all available versions
// This includes installed versions and potentially downloadable versions
func GetAvailableVersions() ([]string, error) {
	// Get installed versions
	installedVersions, err := GetInstalledVersions()
	if err != nil {
		return nil, err
	}

	// Convert to strings
	result := make([]string, len(installedVersions))
	for i, versionInfo := range installedVersions {
		result[i] = versionInfo.Version
	}

	// In the future, we could add other available versions from download sources
	// For now, just return the installed versions

	return result, nil
}

// RebuildRegistry scans the versions directory and rebuilds the registry from existing installations
func RebuildRegistry() error {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return errors.NewSystemError("001",
			"Failed to determine XDG directories", err)
	}

	// Get versions directory
	versionsDir := dirs.VersionsDir

	// Create a new registry
	newRegistry := &VersionRegistry{
		Versions: make(map[string]VersionInfo),
	}

	// Check if versions directory exists
	if _, err := os.Stat(versionsDir); os.IsNotExist(err) {
		// No versions directory, save empty registry
		return SaveRegistry(newRegistry)
	}

	// Scan versions directory
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return errors.NewSystemError(
			ErrRegistryIOFailed,
			fmt.Sprintf("Failed to read versions directory: %s", versionsDir),
			err)
	}

	// Process each entry
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		version := entry.Name()
		installPath := filepath.Join(versionsDir, version)

		// Check if perl binary exists
		perlBinary := filepath.Join(installPath, "bin", "perl")
		if _, err := os.Stat(perlBinary); os.IsNotExist(err) {
			continue
		}

		// Get directory info for install time
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Add to registry with UUID
		uuid := generateUUID()
		newRegistry.Versions[uuid] = VersionInfo{
			Version:     version,
			InstallPath: installPath,
			InstallTime: info.ModTime(),
			Source:      "pvm", // Installations in PVM versions directory are from PVM
		}
	}

	// Add system perl
	if systemPerl, err := detectSystemPerl(); err == nil {
		systemPath := filepath.Dir(systemPerl.Path)
		isPVMInstall := strings.Contains(systemPath, filepath.Join("pvm", "versions"))

		if !isPVMInstall {
			uuid := generateUUID()
			newRegistry.Versions[uuid] = VersionInfo{
				Version:     systemPerl.Version,
				InstallPath: systemPath,
				InstallTime: time.Now(),
				Source:      "system",
			}
		}
	}

	// Auto-detect and add plenv installations directly to registry
	if plenvInstallations, err := DetectPlenv(); err == nil {
		for _, inst := range plenvInstallations {
			uuid := generateUUID()
			newRegistry.Versions[uuid] = VersionInfo{
				Version:     inst.Version,
				InstallPath: inst.Path, // Use original path, not symlink
				InstallTime: inst.InstallTime,
				Source:      string(inst.Source), // "plenv"
			}
		}
	}

	// Auto-detect and add perlbrew installations directly to registry
	if perlbrewInstallations, err := DetectPerlbrew(); err == nil {
		for _, inst := range perlbrewInstallations {
			uuid := generateUUID()
			newRegistry.Versions[uuid] = VersionInfo{
				Version:     inst.Version,
				InstallPath: inst.Path, // Use original path, not symlink
				InstallTime: inst.InstallTime,
				Source:      string(inst.Source), // "perlbrew"
			}
		}
	}

	// Save the new registry
	return SaveRegistry(newRegistry)
}
