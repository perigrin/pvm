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
	// Error prefix
	PrefixXDG = "XDG"

	// Default directory names
	DefaultConfigDirName = ".config"
	DefaultCacheDirName  = ".cache"
	DefaultDataDirName   = ".local/share"
	DefaultStateDirName  = ".local/state"

	// Windows specific directory names
	WindowsConfigDirName = "AppData\\Roaming"
	WindowsCacheDirName  = "AppData\\Local\\Cache"
	WindowsDataDirName   = "AppData\\Local"
	WindowsStateDirName  = "AppData\\Local\\State"

	// Darwin (macOS) specific directory names
	DarwinCacheDirName = "Library/Caches"
	DarwinDataDirName  = "Library/Application Support"

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

	// Application-specific directories
	ConfigDir string // App-specific config directory
	CacheDir  string // App-specific cache directory
	DataDir   string // App-specific data directory
	StateDir  string // App-specific state directory

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
	var err error

	// Get XDG directories with proper fallbacks
	dirs.ConfigHome, err = getConfigHome()
	if err != nil {
		return nil, err
	}

	dirs.CacheHome, err = getCacheHome()
	if err != nil {
		return nil, err
	}

	dirs.DataHome, err = getDataHome()
	if err != nil {
		return nil, err
	}

	dirs.StateHome, err = getStateHome()
	if err != nil {
		return nil, err
	}

	// Set application-specific directories
	dirs.ConfigDir = filepath.Join(dirs.ConfigHome, AppDirName)
	dirs.CacheDir = filepath.Join(dirs.CacheHome, AppDirName)
	dirs.DataDir = filepath.Join(dirs.DataHome, AppDirName)
	dirs.StateDir = filepath.Join(dirs.StateHome, AppDirName)

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

// getConfigHome returns the XDG_CONFIG_HOME directory
func getConfigHome() (string, error) {
	// Check XDG_CONFIG_HOME environment variable
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome != "" {
		return configHome, nil
	}

	// Fallback to platform-specific default
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.NewSystemError("002",
			"Failed to determine user home directory", err)
	}

	// Platform-specific fallbacks
	switch runtime.GOOS {
	case "windows":
		configHome = filepath.Join(home, WindowsConfigDirName)
	default:
		configHome = filepath.Join(home, DefaultConfigDirName)
	}

	return configHome, nil
}

// getCacheHome returns the XDG_CACHE_HOME directory
func getCacheHome() (string, error) {
	// Check XDG_CACHE_HOME environment variable
	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome != "" {
		return cacheHome, nil
	}

	// Fallback to platform-specific default
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.NewSystemError("003",
			"Failed to determine user home directory", err)
	}

	// Platform-specific fallbacks
	switch runtime.GOOS {
	case "windows":
		cacheHome = filepath.Join(home, WindowsCacheDirName)
	case "darwin":
		cacheHome = filepath.Join(home, DarwinCacheDirName)
	default:
		cacheHome = filepath.Join(home, DefaultCacheDirName)
	}

	return cacheHome, nil
}

// getDataHome returns the XDG_DATA_HOME directory
func getDataHome() (string, error) {
	// Check XDG_DATA_HOME environment variable
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome != "" {
		return dataHome, nil
	}

	// Fallback to platform-specific default
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.NewSystemError("004",
			"Failed to determine user home directory", err)
	}

	// Platform-specific fallbacks
	switch runtime.GOOS {
	case "windows":
		dataHome = filepath.Join(home, WindowsDataDirName)
	case "darwin":
		dataHome = filepath.Join(home, DarwinDataDirName)
	default:
		dataHome = filepath.Join(home, DefaultDataDirName)
	}

	return dataHome, nil
}

// getStateHome returns the XDG_STATE_HOME directory
func getStateHome() (string, error) {
	// Check XDG_STATE_HOME environment variable
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome != "" {
		return stateHome, nil
	}

	// Fallback to platform-specific default
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.NewSystemError("005",
			"Failed to determine user home directory", err)
	}

	// Platform-specific fallbacks
	switch runtime.GOOS {
	case "windows":
		stateHome = filepath.Join(home, WindowsStateDirName)
	default:
		stateHome = filepath.Join(home, DefaultStateDirName)
	}

	return stateHome, nil
}

// ensureDir ensures that the directory exists
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
