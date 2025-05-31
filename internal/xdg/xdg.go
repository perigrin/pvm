// ABOUTME: XDG Base Directory support for the PVM Ecosystem
// ABOUTME: Provides functions to handle XDG-compliant directory paths

package xdg

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/adrg/xdg"
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

	// Use the proven xdg library for directory detection
	dirs.ConfigHome = xdg.ConfigHome
	dirs.CacheHome = xdg.CacheHome
	dirs.DataHome = xdg.DataHome
	dirs.StateHome = xdg.StateHome

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

// ensureDir ensures that the directory exists
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
