// ABOUTME: Configuration loader for the PVM Ecosystem
// ABOUTME: Provides functionality to load configuration from multiple sources

package config

import (
	"os"
	"path/filepath"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/project"
	"tamarou.com/pvm/internal/xdg"
)

// Function variables to allow for easier testing
var (
	loadSystemConfig                   = defaultLoadSystemConfig
	loadProjectConfig                  = defaultLoadProjectConfig
	loadProjectConfigWithoutValidation = defaultLoadProjectConfigWithoutValidation
)

// LoadEffectiveConfig loads the configuration from all sources and merges them
// using the following precedence order (highest to lowest):
// 1. Project configuration (.pvm/pvm.toml)
// 2. User configuration ($XDG_CONFIG_HOME/pvm/pvm.toml)
// 3. System configuration (/etc/pvm/pvm.toml)
func LoadEffectiveConfig() (*Config, error) {
	return LoadEffectiveConfigWithOptions(true)
}

// LoadEffectiveConfigWithOptions loads config with validation control
func LoadEffectiveConfigWithOptions(validate bool) (*Config, error) {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return nil, errors.NewConfigError("101",
			"Failed to determine XDG directories", err)
	}

	// Ensure directories exist
	err = dirs.EnsureDirs()
	if err != nil {
		return nil, errors.NewConfigError("102",
			"Failed to create required directories", err)
	}

	// Load configurations
	systemConfig, _ := loadSystemConfig()
	userConfig, _ := loadUserConfig(dirs)

	var projectConfig *Config
	if validate {
		projectConfig, _ = loadProjectConfig()
	} else {
		projectConfig, _ = loadProjectConfigWithoutValidation()
	}

	// Merge configurations according to precedence
	// System (lowest) <- User <- Project (highest)

	// Create empty result config
	result := NewDefaultConfig()

	// Apply configs in order of precedence (low to high)
	// This ensures higher priority configs override lower priority ones
	if systemConfig != nil {
		if systemConfig.PVM != nil {
			mergePVMConfig(result.PVM, systemConfig.PVM)
		}
		if systemConfig.PVX != nil {
			mergePVXConfig(result.PVX, systemConfig.PVX)
		}
		if systemConfig.PVI != nil {
			mergePVIConfig(result.PVI, systemConfig.PVI)
		}
		if systemConfig.PSC != nil {
			mergePSCConfig(result.PSC, systemConfig.PSC)
		}
	}

	if userConfig != nil {
		if userConfig.PVM != nil {
			mergePVMConfig(result.PVM, userConfig.PVM)
		}
		if userConfig.PVX != nil {
			mergePVXConfig(result.PVX, userConfig.PVX)
		}
		if userConfig.PVI != nil {
			mergePVIConfig(result.PVI, userConfig.PVI)
		}
		if userConfig.PSC != nil {
			mergePSCConfig(result.PSC, userConfig.PSC)
		}
	}

	if projectConfig != nil {
		if projectConfig.PVM != nil {
			mergePVMConfig(result.PVM, projectConfig.PVM)
		}
		if projectConfig.PVX != nil {
			mergePVXConfig(result.PVX, projectConfig.PVX)
		}
		if projectConfig.PVI != nil {
			mergePVIConfig(result.PVI, projectConfig.PVI)
		}
		if projectConfig.PSC != nil {
			mergePSCConfig(result.PSC, projectConfig.PSC)
		}
	}

	return result, nil
}

// defaultLoadSystemConfig loads the system-wide configuration
func defaultLoadSystemConfig() (*Config, error) {
	// Get system config path
	systemConfigPath := xdg.GetSystemConfigPath()

	// Check if file exists
	if _, err := os.Stat(systemConfigPath); os.IsNotExist(err) {
		return nil, nil // No system config, return nil
	}

	// Parse configuration file
	return ParseFile(systemConfigPath)
}

// loadUserConfig loads the user's configuration
func loadUserConfig(dirs *xdg.Dirs) (*Config, error) {
	// Get user config path
	userConfigPath := dirs.GetConfigFilePath()

	// Check if file exists
	if _, err := os.Stat(userConfigPath); os.IsNotExist(err) {
		return nil, nil // No user config, return nil
	}

	// Parse configuration file
	return ParseFile(userConfigPath)
}

// defaultLoadProjectConfig loads the project-level configuration
func defaultLoadProjectConfig() (*Config, error) {
	// Start from current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, errors.NewConfigError("103",
			"Failed to determine current directory", err)
	}

	// Check for .pvm/pvm.toml in current directory and parents
	config, err := findProjectConfig(currentDir)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// defaultLoadProjectConfigWithoutValidation loads project config without validation
func defaultLoadProjectConfigWithoutValidation() (*Config, error) {
	// Start from current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, errors.NewConfigError("103",
			"Failed to determine current directory", err)
	}

	// Check for .pvm/pvm.toml in current directory and parents
	config, err := findProjectConfigWithoutValidation(currentDir)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// findProjectConfigWithoutValidation recursively searches for .pvm/pvm.toml without validation
func findProjectConfigWithoutValidation(dir string) (*Config, error) {
	// Check for .pvm/pvm.toml
	projectConfigPath := xdg.GetProjectConfigPath(dir)
	if _, err := os.Stat(projectConfigPath); err == nil {
		// Found project config, parse it without validation
		return ParseFileWithoutValidation(projectConfigPath)
	}

	// Check if we've reached the filesystem root
	parentDir := filepath.Dir(dir)
	if parentDir == dir {
		// Reached filesystem root, no project config found
		return nil, nil
	}

	// Continue searching in parent directory
	return findProjectConfigWithoutValidation(parentDir)
}

// findProjectConfig recursively searches for .pvm/pvm.toml in the given directory
// and its parent directories, stopping at the filesystem root
func findProjectConfig(dir string) (*Config, error) {
	// Check for .pvm/pvm.toml
	projectConfigPath := xdg.GetProjectConfigPath(dir)
	if _, err := os.Stat(projectConfigPath); err == nil {
		// Found project config, parse it
		return ParseFile(projectConfigPath)
	}

	// Check if we've reached the filesystem root
	parentDir := filepath.Dir(dir)
	if parentDir == dir {
		// Reached the root, no project config found
		return nil, nil
	}

	// Recursively check the parent directory
	return findProjectConfig(parentDir)
}

// getProjectRootFunc returns the root directory of the project containing a .pvm directory
// Returns an empty string if no project root is found
func getProjectRootFunc() string {
	// Start from current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	return findProjectRoot(currentDir)
}

// GetProjectRoot is a variable that points to getProjectRootFunc,
// allowing it to be replaced in tests
var GetProjectRoot = getProjectRootFunc

// findProjectRoot recursively searches for a .pvm directory in the given directory

// SaveUserConfig saves the configuration to the user's configuration file
func SaveUserConfig(cfg *Config) error {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return errors.NewConfigError("111",
			"Failed to determine XDG directories", err)
	}

	// Ensure directories exist
	err = dirs.EnsureDirs()
	if err != nil {
		return errors.NewConfigError("112",
			"Failed to create required directories", err)
	}

	// Create user config path
	configPath := filepath.Join(dirs.ConfigDir, "pvm.toml")

	// Save the configuration
	return SaveToFile(cfg, configPath)
}

// and its parent directories, stopping at the filesystem root
func findProjectRoot(dir string) string {
	// Check for .pvm directory
	pvmDir := filepath.Join(dir, ".pvm")
	if _, err := os.Stat(pvmDir); err == nil {
		// Found project root
		return dir
	}

	// Check if we've reached the filesystem root
	parentDir := filepath.Dir(dir)
	if parentDir == dir {
		// Reached the root, no project root found
		return ""
	}

	// Recursively check the parent directory
	return findProjectRoot(parentDir)
}

// InitUserConfig initializes the user configuration file with default values
func InitUserConfig() error {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return errors.NewConfigError("104",
			"Failed to determine XDG directories", err)
	}

	// Ensure directories exist
	err = dirs.EnsureDirs()
	if err != nil {
		return errors.NewConfigError("105",
			"Failed to create required directories", err)
	}

	// Get user config path
	userConfigPath := dirs.GetConfigFilePath()

	// Check if file exists
	if _, err := os.Stat(userConfigPath); err == nil {
		return errors.NewConfigError("106",
			"User configuration file already exists", nil).
			WithLocation(userConfigPath)
	}

	// Create new config with defaults
	config := NewDefaultConfig()

	// Save configuration
	return SaveToFile(config, userConfigPath)
}

// InitProjectConfig initializes a project configuration file with default values
func InitProjectConfig(projectDir string) error {
	// Create .pvm directory if it doesn't exist
	pvmDir := filepath.Join(projectDir, ".pvm")
	if err := os.MkdirAll(pvmDir, 0755); err != nil {
		return errors.NewConfigError("107",
			"Failed to create project configuration directory", err).
			WithLocation(pvmDir)
	}

	// Get project config path
	projectConfigPath := xdg.GetProjectConfigPath(projectDir)

	// Check if file exists
	if _, err := os.Stat(projectConfigPath); err == nil {
		return errors.NewConfigError("108",
			"Project configuration file already exists", nil).
			WithLocation(projectConfigPath)
	}

	// Create new config with defaults
	config := NewDefaultConfig()

	// Save configuration
	return SaveToFile(config, projectConfigPath)
}

// LoadEffectiveConfigWithProjectContext loads configuration using the new project detector
func LoadEffectiveConfigWithProjectContext() (*Config, error) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.NewConfigError("109",
			"Failed to get current working directory", err)
	}

	// Use project detector to find project-aware config
	return LoadEffectiveConfigForDirectory(wd)
}

// LoadEffectiveConfigForDirectory loads configuration for a specific directory using project detection
func LoadEffectiveConfigForDirectory(dir string) (*Config, error) {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return nil, errors.NewConfigError("101",
			"Failed to determine XDG directories", err)
	}

	// Ensure directories exist
	err = dirs.EnsureDirs()
	if err != nil {
		return nil, errors.NewConfigError("102",
			"Failed to create required directories", err)
	}

	// Load configurations
	systemConfig, _ := loadSystemConfig()
	userConfig, _ := loadUserConfig(dirs)

	// Use project detector for project config
	var projectConfig *Config
	configPath, err := project.GetProjectAwareConfigPath(dir)
	if err == nil && configPath != "" {
		projectConfig, _ = ParseFile(configPath)
	}

	// Create empty result config
	result := NewDefaultConfig()

	// Apply configs in order of precedence (low to high)
	if systemConfig != nil {
		if systemConfig.PVM != nil {
			mergePVMConfig(result.PVM, systemConfig.PVM)
		}
		if systemConfig.PVX != nil {
			mergePVXConfig(result.PVX, systemConfig.PVX)
		}
		if systemConfig.PVI != nil {
			mergePVIConfig(result.PVI, systemConfig.PVI)
		}
		if systemConfig.PSC != nil {
			mergePSCConfig(result.PSC, systemConfig.PSC)
		}
	}

	if userConfig != nil {
		if userConfig.PVM != nil {
			mergePVMConfig(result.PVM, userConfig.PVM)
		}
		if userConfig.PVX != nil {
			mergePVXConfig(result.PVX, userConfig.PVX)
		}
		if userConfig.PVI != nil {
			mergePVIConfig(result.PVI, userConfig.PVI)
		}
		if userConfig.PSC != nil {
			mergePSCConfig(result.PSC, userConfig.PSC)
		}
	}

	if projectConfig != nil {
		if projectConfig.PVM != nil {
			mergePVMConfig(result.PVM, projectConfig.PVM)
		}
		if projectConfig.PVX != nil {
			mergePVXConfig(result.PVX, projectConfig.PVX)
		}
		if projectConfig.PVI != nil {
			mergePVIConfig(result.PVI, projectConfig.PVI)
		}
		if projectConfig.PSC != nil {
			mergePSCConfig(result.PSC, projectConfig.PSC)
		}
	}

	return result, nil
}
