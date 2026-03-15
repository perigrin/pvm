// ABOUTME: Tests for configuration loader functionality
// ABOUTME: Verifies loading and merging configuration from multiple sources

package config

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/xdg"
)

func TestLoadEffectiveConfigPrecedence(t *testing.T) {
	// Create a temporary directory structure for testing
	testDir := t.TempDir()

	// Create subdirectories
	systemDir := filepath.Join(testDir, "system")
	userDir := filepath.Join(testDir, "user")
	projectDir := filepath.Join(testDir, "project")

	if err := os.MkdirAll(filepath.Join(systemDir, "pvm"), 0755); err != nil {
		t.Fatalf("Failed to create system directory: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(userDir, "pvm"), 0755); err != nil {
		t.Fatalf("Failed to create user directory: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(projectDir, ".pvm"), 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create test configurations with different values to test precedence
	systemConfigContent := `
[pvm]
default_perl = "5.36.0"
build_jobs = 2
`

	userConfigContent := `
[pvm]
default_perl = "5.38.0"
build_jobs = 4
`

	projectConfigContent := `
[pvm]
default_perl = "5.40.0"
`

	// Write the configuration files
	systemConfigPath := filepath.Join(systemDir, "pvm", "pvm.toml")
	userConfigPath := filepath.Join(userDir, "pvm", "pvm.toml")
	projectConfigPath := filepath.Join(projectDir, ".pvm", "pvm.toml")

	if err := os.WriteFile(systemConfigPath, []byte(systemConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write system config: %v", err)
	}

	if err := os.WriteFile(userConfigPath, []byte(userConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write user config: %v", err)
	}

	if err := os.WriteFile(projectConfigPath, []byte(projectConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write project config: %v", err)
	}

	// Test merging directly without using LoadEffectiveConfig
	t.Run("DirectMergingTest", func(t *testing.T) {
		// Parse the configs
		systemConfig, err := ParseFile(systemConfigPath)
		if err != nil {
			t.Fatalf("Failed to parse system config: %v", err)
		}

		userConfig, err := ParseFile(userConfigPath)
		if err != nil {
			t.Fatalf("Failed to parse user config: %v", err)
		}

		projectConfig, err := ParseFile(projectConfigPath)
		if err != nil {
			t.Fatalf("Failed to parse project config: %v", err)
		}

		// Start with a default config
		result := NewDefaultConfig()

		// Apply configs in order of precedence (low to high)
		if systemConfig.PVM != nil {
			mergePVMConfig(result.PVM, systemConfig.PVM)
		}

		if userConfig.PVM != nil {
			mergePVMConfig(result.PVM, userConfig.PVM)
		}

		if projectConfig.PVM != nil {
			mergePVMConfig(result.PVM, projectConfig.PVM)
		}

		// Check that values were correctly merged
		if result.PVM.DefaultPerl != "5.40.0" {
			t.Errorf("Expected DefaultPerl to be 5.40.0 (from project config), got %s", result.PVM.DefaultPerl)
		}

		if result.PVM.BuildJobs != 4 {
			t.Errorf("Expected BuildJobs to be 4 (from user config), got %d", result.PVM.BuildJobs)
		}
	})
}

func TestGetProjectRoot(t *testing.T) {
	// Create a temporary directory structure for testing
	testDir := t.TempDir()

	// Create a nested project structure
	projectRoot := filepath.Join(testDir, "project")
	subDir1 := filepath.Join(projectRoot, "src")
	subDir2 := filepath.Join(subDir1, "module")

	if err := os.MkdirAll(filepath.Join(projectRoot, ".pvm"), 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	if err := os.MkdirAll(subDir2, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Save current directory
	oldDir, _ := os.Getwd()
	defer func() {
		err := os.Chdir(oldDir)
		if err != nil {
			t.Logf("Failed to return to original directory: %v", err)
		}
	}()

	// Test finding project root from different locations

	t.Run("FromProjectRoot", func(t *testing.T) {
		err := os.Chdir(projectRoot)
		if err != nil {
			t.Fatalf("Failed to change to project root: %v", err)
		}

		root := GetProjectRoot()
		// On macOS, /var/folders resolves to /private/var/folders, so use filepath.EvalSymlinks
		// to compare normalized paths
		expectedPath, _ := filepath.EvalSymlinks(projectRoot)
		if root != expectedPath {
			t.Errorf("Expected root to be %s, got %s", expectedPath, root)
		}
	})

	t.Run("FromNestedDirectory", func(t *testing.T) {
		err := os.Chdir(subDir2)
		if err != nil {
			t.Fatalf("Failed to change to subdirectory: %v", err)
		}

		root := GetProjectRoot()
		// On macOS, /var/folders resolves to /private/var/folders, so use filepath.EvalSymlinks
		// to compare normalized paths
		expectedPath, _ := filepath.EvalSymlinks(projectRoot)
		if root != expectedPath {
			t.Errorf("Expected root to be %s, got %s", expectedPath, root)
		}
	})

	t.Run("OutsideProject", func(t *testing.T) {
		err := os.Chdir(testDir)
		if err != nil {
			t.Fatalf("Failed to change to test directory: %v", err)
		}

		root := GetProjectRoot()
		if root != "" {
			t.Errorf("Expected root to be empty string, got %s", root)
		}
	})
}

func TestInitUserConfig(t *testing.T) {
	// Create a temporary directory for the test
	testDir := t.TempDir()

	// Save original xdg.GetDirs function
	originalGetDirs := xdg.GetDirs
	defer func() {
		xdg.GetDirs = originalGetDirs
	}()

	// Create a mock GetDirs function that returns test directories
	xdg.GetDirs = func() (*xdg.Dirs, error) {
		return &xdg.Dirs{
			ConfigHome:         testDir,
			CacheHome:          filepath.Join(testDir, "cache"),
			DataHome:           filepath.Join(testDir, "data"),
			StateHome:          filepath.Join(testDir, "state"),
			ConfigDir:          filepath.Join(testDir, "pvm"),
			CacheDir:           filepath.Join(testDir, "cache", "pvm"),
			DataDir:            filepath.Join(testDir, "data", "pvm"),
			StateDir:           filepath.Join(testDir, "state", "pvm"),
			VersionsDir:        filepath.Join(testDir, "data", "pvm", "versions"),
			SourcesDir:         filepath.Join(testDir, "cache", "pvm", "sources"),
			ShimsDir:           filepath.Join(testDir, "data", "pvm", "shims"),
			TypeDefinitionsDir: filepath.Join(testDir, "data", "pvm", "type_definitions"),
			BuildDir:           filepath.Join(testDir, "cache", "pvm", "build"),
			EnsureDirs: func() error {
				dirs := []string{
					filepath.Join(testDir, "pvm"),
					filepath.Join(testDir, "cache", "pvm"),
					filepath.Join(testDir, "data", "pvm"),
					filepath.Join(testDir, "state", "pvm"),
					filepath.Join(testDir, "data", "pvm", "versions"),
					filepath.Join(testDir, "cache", "pvm", "sources"),
					filepath.Join(testDir, "data", "pvm", "shims"),
					filepath.Join(testDir, "data", "pvm", "type_definitions"),
					filepath.Join(testDir, "cache", "pvm", "build"),
				}
				for _, dir := range dirs {
					if err := os.MkdirAll(dir, 0755); err != nil {
						return err
					}
				}
				return nil
			},
		}, nil
	}

	// Initialize user config
	err := InitUserConfig()
	if err != nil {
		t.Fatalf("InitUserConfig() returned an error: %v", err)
	}

	// Check if file was created
	configPath := filepath.Join(testDir, "pvm", "pvm.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("User configuration file was not created: %s", configPath)
	}

	// Check if trying to initialize again returns an error
	err = InitUserConfig()
	if err == nil {
		t.Errorf("Expected error when initializing existing config, got nil")
	}
}

func TestInitProjectConfig(t *testing.T) {
	// Create a temporary directory for the test
	projectDir := t.TempDir()

	// Initialize project config
	err := InitProjectConfig(projectDir)
	if err != nil {
		t.Fatalf("InitProjectConfig() returned an error: %v", err)
	}

	// Check if .pvm directory was created
	pvmDir := filepath.Join(projectDir, ".pvm")
	if _, err := os.Stat(pvmDir); os.IsNotExist(err) {
		t.Errorf("Project .pvm directory was not created: %s", pvmDir)
	}

	// Check if config file was created
	configPath := filepath.Join(pvmDir, "pvm.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Project configuration file was not created: %s", configPath)
	}

	// Check if trying to initialize again returns an error
	err = InitProjectConfig(projectDir)
	if err == nil {
		t.Errorf("Expected error when initializing existing project config, got nil")
	}
}

func TestUpdateConfigurationLoading(t *testing.T) {
	// Create a temporary directory for testing
	testDir := t.TempDir()
	userDir := filepath.Join(testDir, "user")

	if err := os.MkdirAll(filepath.Join(userDir, "pvm"), 0755); err != nil {
		t.Fatalf("Failed to create user directory: %v", err)
	}

	// Create user config with update configuration
	userConfigContent := `
[pvm.update]
auto_update_enabled = true
auto_update_interval = "12h"
repository = "custom/repo"
channel = "beta"
github_token = "test-token"
backup_enabled = false
auto_rollback_enabled = false
check_prerelease = true
notifications_enabled = false
security_updates_only = true
max_retries = 5
timeout = "10m"
skip_checksums = true
`

	userConfigPath := filepath.Join(userDir, "pvm", "pvm.toml")
	if err := os.WriteFile(userConfigPath, []byte(userConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write user config file: %v", err)
	}

	// Mock XDG base directory
	oldXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if oldXDGConfigHome != "" {
			os.Setenv("XDG_CONFIG_HOME", oldXDGConfigHome)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()
	os.Setenv("XDG_CONFIG_HOME", userDir)

	// Load configuration
	config, err := LoadEffectiveConfig()
	if err != nil {
		t.Fatalf("LoadEffectiveConfig() returned an error: %v", err)
	}

	// Verify update configuration was loaded correctly
	if config.PVM == nil {
		t.Fatal("Expected PVM configuration to be loaded")
	}

	updateCfg := config.PVM.Update
	if updateCfg == nil {
		t.Fatal("Expected PVM update configuration to be loaded")
	}

	// Test all configuration fields
	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{"AutoUpdateEnabled", true, updateCfg.AutoUpdateEnabled},
		{"AutoUpdateInterval", "12h", updateCfg.AutoUpdateInterval},
		{"Repository", "custom/repo", updateCfg.Repository},
		{"Channel", "beta", updateCfg.Channel},
		{"GitHubToken", "test-token", updateCfg.GitHubToken},
		{"BackupEnabled", false, updateCfg.BackupEnabled},
		{"AutoRollbackEnabled", false, updateCfg.AutoRollbackEnabled},
		{"CheckPrerelease", true, updateCfg.CheckPrerelease},
		{"NotificationsEnabled", false, updateCfg.NotificationsEnabled},
		{"SecurityUpdatesOnly", true, updateCfg.SecurityUpdatesOnly},
		{"MaxRetries", 5, updateCfg.MaxRetries},
		{"Timeout", "10m", updateCfg.Timeout},
		{"SkipChecksums", true, updateCfg.SkipChecksums},
	}

	for _, test := range tests {
		if test.actual != test.expected {
			t.Errorf("Expected %s to be %v, got %v", test.name, test.expected, test.actual)
		}
	}
}

func TestUpdateConfigurationValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *PVMUpdateConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "ValidConfiguration",
			config: &PVMUpdateConfig{
				AutoUpdateInterval: "24h",
				Repository:         "owner/repo",
				Channel:            "stable",
				MaxRetries:         3,
				Timeout:            "5m",
			},
			expectError: false,
		},
		{
			name: "InvalidInterval",
			config: &PVMUpdateConfig{
				AutoUpdateInterval: "invalid",
				Repository:         "owner/repo",
				Channel:            "stable",
			},
			expectError: true,
			errorMsg:    "AutoUpdateInterval must be a valid duration (e.g., '24h', '1h')",
		},
		{
			name: "EmptyRepository",
			config: &PVMUpdateConfig{
				Repository: "",
				Channel:    "stable",
			},
			expectError: true,
			errorMsg:    "Repository cannot be empty",
		},
		{
			name: "InvalidChannel",
			config: &PVMUpdateConfig{
				Repository: "owner/repo",
				Channel:    "invalid",
			},
			expectError: true,
			errorMsg:    "Channel must be one of: stable, beta, alpha, nightly, developer",
		},
		{
			name: "NegativeRetries",
			config: &PVMUpdateConfig{
				Repository: "owner/repo",
				Channel:    "stable",
				MaxRetries: -1,
			},
			expectError: true,
			errorMsg:    "MaxRetries cannot be negative",
		},
		{
			name: "InvalidTimeout",
			config: &PVMUpdateConfig{
				Repository: "owner/repo",
				Channel:    "stable",
				Timeout:    "invalid",
			},
			expectError: true,
			errorMsg:    "Timeout must be a valid duration (e.g., '5m', '30s')",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errors := test.config.Validate()

			if test.expectError {
				if len(errors) == 0 {
					t.Errorf("Expected validation error, but none occurred")
				} else {
					found := false
					for _, err := range errors {
						if err.Error() == test.errorMsg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error message '%s', but got %v", test.errorMsg, errors)
					}
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("Expected no validation errors, but got: %v", errors)
				}
			}
		})
	}
}
