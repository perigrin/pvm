// ABOUTME: XDG Base Directory support for the PVM Ecosystem
// ABOUTME: Provides functions to handle XDG-compliant directory paths

package xdg

import (
	"os"
	"path/filepath"
	"runtime"

	"tamarou.com/pvm/internal/errors"
)

const (
	// Application directory name
	AppDirName = "pvm"

	// Sub-directory names
	VersionsDir        = "versions"
	SourcesDir         = "sources"
	ShimsDir           = "shims"
	TypeDefinitionsDir = "type_definitions"
	BuildDir           = "build"
)

// Dirs holds the directories used by the PVM Ecosystem
type Dirs struct {
	// XDG standard directories
	ConfigHome string // For configuration files
	CacheHome  string // For non-essential data files
	DataHome   string // For data files
	StateHome  string // For state files
	BinHome    string // For executable files

	// Application-specific directories
	ConfigDir string // App-specific config directory
	CacheDir  string // App-specific cache directory
	DataDir   string // App-specific data directory
	StateDir  string // App-specific state directory
	BinDir    string // App-specific bin directory

	// PVM-specific directories
	VersionsDir        string // For installed Perl versions
	SourcesDir         string // For downloaded source archives
	ShimsDir           string // For shim executables
	TypeDefinitionsDir string // For type definitions
	BuildDir           string // For build cache

	// Function pointers for easier testing
	EnsureDirs func() error
}

// getDirsFunc is the actual implementation of GetDirs
func getDirsFunc() (*Dirs, error) {
	dirs := &Dirs{}

	// Read environment variables directly to support runtime changes
	dirs.ConfigHome = getXDGConfigHome()
	dirs.CacheHome = getXDGCacheHome()
	dirs.DataHome = getXDGDataHome()
	dirs.StateHome = getXDGStateHome()
	dirs.BinHome = getXDGBinHome()

	// Set application-specific directories
	dirs.ConfigDir = filepath.Join(dirs.ConfigHome, AppDirName)
	dirs.CacheDir = filepath.Join(dirs.CacheHome, AppDirName)
	dirs.DataDir = filepath.Join(dirs.DataHome, AppDirName)
	dirs.StateDir = filepath.Join(dirs.StateHome, AppDirName)
	dirs.BinDir = dirs.BinHome // Use BinHome directly for tool shims

	// Set PVM-specific directories
	dirs.VersionsDir = filepath.Join(dirs.DataDir, VersionsDir)
	dirs.SourcesDir = filepath.Join(dirs.CacheDir, SourcesDir)
	dirs.ShimsDir = filepath.Join(dirs.DataDir, ShimsDir)
	dirs.TypeDefinitionsDir = filepath.Join(dirs.DataDir, TypeDefinitionsDir)
	dirs.BuildDir = filepath.Join(dirs.CacheDir, BuildDir)

	// Set the EnsureDirs function
	dirs.EnsureDirs = func() error {
		return dirs.ensureDirsImpl()
	}

	return dirs, nil
}

// GetDirs is a variable that points to getDirsFunc,
// allowing it to be replaced in tests
var GetDirs = getDirsFunc

// ensureDirsImpl is the actual implementation of EnsureDirs
func (d *Dirs) ensureDirsImpl() error {
	// Create required directories
	dirsToCreate := []string{
		d.ConfigDir,
		d.CacheDir,
		d.DataDir,
		d.StateDir,
		d.BinDir,
		d.VersionsDir,
		d.SourcesDir,
		d.ShimsDir,
		d.TypeDefinitionsDir,
		d.BuildDir,
	}

	for _, dir := range dirsToCreate {
		if err := ensureDir(dir); err != nil {
			return errors.NewSystemError("001",
				"Failed to create directory", err).
				WithLocation(dir)
		}
	}

	return nil
}

// GetConfigFilePath returns the path to the configuration file
func (d *Dirs) GetConfigFilePath() string {
	return filepath.Join(d.ConfigDir, "pvm.toml")
}

// GetProjectConfigPath returns the path to the project's configuration file
func GetProjectConfigPath(projectDir string) string {
	return filepath.Join(projectDir, ".pvm", "pvm.toml")
}

// GetSystemConfigPath returns the path to the system-wide configuration file
func GetSystemConfigPath() string {
	if runtime.GOOS == "windows" {
		// On Windows, use ProgramData
		return filepath.Join(os.Getenv("ProgramData"), "pvm", "pvm.toml")
	}
	return "/etc/pvm/pvm.toml"
}

// Helper functions

// getXDGConfigHome returns XDG_CONFIG_HOME or default
func getXDGConfigHome() string {
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return configHome
	}
	return filepath.Join(getHomeDir(), ".config")
}

// getXDGCacheHome returns XDG_CACHE_HOME or default
func getXDGCacheHome() string {
	if cacheHome := os.Getenv("XDG_CACHE_HOME"); cacheHome != "" {
		return cacheHome
	}
	return filepath.Join(getHomeDir(), ".cache")
}

// getXDGDataHome returns XDG_DATA_HOME or default
func getXDGDataHome() string {
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return dataHome
	}
	return filepath.Join(getHomeDir(), ".local", "share")
}

// getXDGStateHome returns XDG_STATE_HOME or default
func getXDGStateHome() string {
	if stateHome := os.Getenv("XDG_STATE_HOME"); stateHome != "" {
		return stateHome
	}
	return filepath.Join(getHomeDir(), ".local", "state")
}

// getXDGBinHome returns XDG_BIN_HOME or default
func getXDGBinHome() string {
	if binHome := os.Getenv("XDG_BIN_HOME"); binHome != "" {
		return binHome
	}
	return filepath.Join(getHomeDir(), ".local", "bin")
}

// getHomeDir returns the user's home directory
func getHomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}
	return "."
}

// ensureDir ensures that the directory exists
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
