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
