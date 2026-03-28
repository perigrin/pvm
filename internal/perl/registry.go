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

// getRegistryPath returns the path to the registry file
// Checks PVM_PERL_REGISTRY environment variable first, falls back to XDG standard location
func getRegistryPath() (string, error) {
	// Check if custom registry path is set via environment variable
	if customPath := os.Getenv("PVM_PERL_REGISTRY"); customPath != "" {
		return customPath, nil
	}

	// Fall back to XDG standard location
	dirs, err := xdg.GetDirs()
	if err != nil {
		return "", errors.NewSystemError("001",
			"Failed to determine XDG directories", err)
	}

	// Ensure data directory exists
	err = dirs.EnsureDirs()
	if err != nil {
		return "", errors.NewSystemError("002",
			"Failed to create required directories", err)
	}

	return filepath.Join(dirs.DataDir, registryFileName), nil
}

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

	// Remote is the fork remote name (e.g. "mycompany"); empty for stock origin installs
	Remote string `json:"remote,omitempty"`

	// ForkName is the name of the fork (e.g. "myfork"); empty when not a named fork
	ForkName string `json:"fork_name,omitempty"`

	// BaseVersion is the upstream Perl version the fork is based on; empty for stock installs
	BaseVersion string `json:"base_version,omitempty"`
}

// DisplayName returns the human-readable name for this version.
// For fork installs it returns "<remote>/<forkname>-<version>" or "<remote>/<version>".
// For stock (origin) installs it returns the bare version string.
func (v VersionInfo) DisplayName() string {
	if v.Remote == "" || v.Remote == "origin" {
		return v.Version
	}
	if v.ForkName == "" {
		return v.Remote + "/" + v.BaseVersion
	}
	return v.Remote + "/" + v.ForkName + "-" + v.BaseVersion
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

	// Get registry file path
	registryPath, err := getRegistryPath()
	if err != nil {
		return registry, err
	}

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
	// Get registry file path
	registryPath, err := getRegistryPath()
	if err != nil {
		return err
	}

	// If using a custom registry path (not XDG), ensure the directory exists
	if customPath := os.Getenv("PVM_PERL_REGISTRY"); customPath != "" {
		registryDir := filepath.Dir(registryPath)
		if err := os.MkdirAll(registryDir, 0755); err != nil {
			return errors.NewSystemError("002",
				"Failed to create registry directory", err)
		}
	}

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

// getVersionInfoFunc returns information about a specific installed version
func getVersionInfoFunc(version string) (*VersionInfo, error) {
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

	// Look up version by searching through all entries (registry is UUID-keyed)
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

// GetVersionInfo is a variable that points to getVersionInfoFunc,
// allowing it to be replaced in tests
var GetVersionInfo = getVersionInfoFunc

// findByDisplayNameFunc looks up a VersionInfo by its display name.
// For stock installs the display name is the bare version (e.g. "5.38.0").
// For fork installs it is "<remote>/<forkname>-<version>" or "<remote>/<version>".
// Returns nil when no matching entry is found.
func findByDisplayNameFunc(name string) *VersionInfo {
	registry, err := LoadRegistry()
	if err != nil {
		return nil
	}

	for _, info := range registry.Versions {
		if info.DisplayName() == name {
			// Return a copy so callers cannot mutate registry internals
			v := info
			return &v
		}
	}
	return nil
}

// FindByDisplayName is a variable that points to findByDisplayNameFunc,
// allowing it to be replaced in tests
var FindByDisplayName = findByDisplayNameFunc

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

// uninstallVersionFunc removes a Perl version from the system
func uninstallVersionFunc(version string) error {
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

	// Find the registry key for this version (registry is UUID-keyed, not version-keyed)
	registryKey := ""
	var versionInfo VersionInfo
	for key, info := range registry.Versions {
		if info.Version == normalizedVersion {
			registryKey = key
			versionInfo = info
			break
		}
	}

	if registryKey == "" {
		return errors.NewVersionError(
			ErrVersionNotFound,
			fmt.Sprintf("Version %s is not installed", normalizedVersion),
			nil)
	}

	// If this is a system Perl, don't remove it, just unregister
	if versionInfo.Source == "system" {
		delete(registry.Versions, registryKey)
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
	delete(registry.Versions, registryKey)

	// Save updated registry
	return SaveRegistry(registry)
}

// UninstallVersion is a variable that points to uninstallVersionFunc,
// allowing it to be replaced in tests
var UninstallVersion = uninstallVersionFunc

// uninstallVersionByDisplayNameFunc removes a Perl version identified by its display name.
// For stock installs the display name is the bare version (e.g. "5.38.0").
// For fork installs it is "<remote>/<forkname>-<version>" or "<remote>/<version>".
// The clone cache is NOT removed, since other versions may share it.
func uninstallVersionByDisplayNameFunc(displayName string) error {
	// Load registry
	registry, err := LoadRegistry()
	if err != nil {
		return err
	}

	// Find the registry entry by display name
	registryKey := ""
	var versionInfo VersionInfo
	for key, info := range registry.Versions {
		if info.DisplayName() == displayName {
			registryKey = key
			versionInfo = info
			break
		}
	}

	if registryKey == "" {
		return errors.NewVersionError(
			ErrVersionNotFound,
			fmt.Sprintf("Version %s is not installed", displayName),
			nil)
	}

	// If this is a system Perl, don't remove it, just unregister
	if versionInfo.Source == "system" {
		delete(registry.Versions, registryKey)
		return SaveRegistry(registry)
	}

	// Remove installation directory
	err = osRemoveAll(versionInfo.InstallPath)
	if err != nil {
		return errors.NewVersionError(
			ErrUninstallFailed,
			fmt.Sprintf("Failed to remove installation directory for %s", displayName),
			err).
			WithLocation(versionInfo.InstallPath)
	}

	// Remove from registry
	delete(registry.Versions, registryKey)

	// Save updated registry
	return SaveRegistry(registry)
}

// UninstallVersionByDisplayName is a variable that points to uninstallVersionByDisplayNameFunc,
// allowing it to be replaced in tests
var UninstallVersionByDisplayName = uninstallVersionByDisplayNameFunc

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

	// Process each entry with a two-level scan.
	// - If an entry contains bin/perl directly it is a stock install.
	// - If an entry is a directory with no bin/perl it is treated as a remote
	//   namespace and its subdirectories are scanned for fork installs.
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		entryName := entry.Name()
		entryPath := filepath.Join(versionsDir, entryName)

		// Check if perl binary exists directly inside this entry (stock install)
		perlBinary := filepath.Join(entryPath, "bin", "perl")
		if _, err := os.Stat(perlBinary); err == nil {
			// Stock install found
			info, err := entry.Info()
			if err != nil {
				continue
			}
			uuid := generateUUID()
			newRegistry.Versions[uuid] = VersionInfo{
				Version:     entryName,
				InstallPath: entryPath,
				InstallTime: info.ModTime(),
				Source:      "pvm", // Installations in PVM versions directory are from PVM
			}
			continue
		}

		// No bin/perl at this level — treat as a remote namespace and scan subdirs
		remoteName := entryName
		subEntries, err := os.ReadDir(entryPath)
		if err != nil {
			continue
		}
		for _, subEntry := range subEntries {
			if !subEntry.IsDir() {
				continue
			}
			subName := subEntry.Name()
			subPath := filepath.Join(entryPath, subName)

			// Check for perl binary in the subdirectory
			subPerlBinary := filepath.Join(subPath, "bin", "perl")
			if _, err := os.Stat(subPerlBinary); os.IsNotExist(err) {
				continue
			}

			// Reconstruct fork metadata from the directory name.
			// Fork dirs follow the pattern "<forkname>-<version>" or just "<version>".
			var forkName, baseVersion string
			if idx := strings.LastIndex(subName, "-"); idx != -1 {
				// Potential "<forkname>-<version>" — verify the version portion parses
				candidate := subName[idx+1:]
				if _, err := ParseVersion(candidate); err == nil {
					forkName = subName[:idx]
					baseVersion = candidate
				}
			}
			// If forkName is still empty, subName itself should be a bare version
			if forkName == "" {
				if _, err := ParseVersion(subName); err == nil {
					baseVersion = subName
				}
			}
			// Skip entries where we could not determine a valid base version
			if baseVersion == "" {
				continue
			}

			info, err := subEntry.Info()
			if err != nil {
				continue
			}
			uuid := generateUUID()
			newRegistry.Versions[uuid] = VersionInfo{
				Version:     subName,
				InstallPath: subPath,
				InstallTime: info.ModTime(),
				Source:      "pvm",
				Remote:      remoteName,
				ForkName:    forkName,
				BaseVersion: baseVersion,
			}
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
