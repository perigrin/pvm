// ABOUTME: Tests for XDG Base Directory support
// ABOUTME: Verifies XDG directory paths and creation functionality

package xdg

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetDirs(t *testing.T) {
	// Save current environment variables to restore later
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldCacheHome := os.Getenv("XDG_CACHE_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	oldStateHome := os.Getenv("XDG_STATE_HOME")

	// Restore environment variables after the test
	defer func() {
		_ = os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		_ = os.Setenv("XDG_CACHE_HOME", oldCacheHome)
		_ = os.Setenv("XDG_DATA_HOME", oldDataHome)
		_ = os.Setenv("XDG_STATE_HOME", oldStateHome)
	}()

	t.Run("ExplicitEnvironmentVariables", func(t *testing.T) {
		// Set test environment variables
		testDir := t.TempDir()
		testConfigDir := filepath.Join(testDir, "config")
		testCacheDir := filepath.Join(testDir, "cache")
		testDataDir := filepath.Join(testDir, "data")
		testStateDir := filepath.Join(testDir, "state")

		_ = os.Setenv("XDG_CONFIG_HOME", testConfigDir)
		_ = os.Setenv("XDG_CACHE_HOME", testCacheDir)
		_ = os.Setenv("XDG_DATA_HOME", testDataDir)
		_ = os.Setenv("XDG_STATE_HOME", testStateDir)

		dirs, err := GetDirs()
		if err != nil {
			t.Fatalf("GetDirs() returned an error: %v", err)
		}

		// Verify XDG base directories
		if dirs.ConfigHome != testConfigDir {
			t.Errorf("Expected ConfigHome to be %s, got %s", testConfigDir, dirs.ConfigHome)
		}
		if dirs.CacheHome != testCacheDir {
			t.Errorf("Expected CacheHome to be %s, got %s", testCacheDir, dirs.CacheHome)
		}
		if dirs.DataHome != testDataDir {
			t.Errorf("Expected DataHome to be %s, got %s", testDataDir, dirs.DataHome)
		}
		if dirs.StateHome != testStateDir {
			t.Errorf("Expected StateHome to be %s, got %s", testStateDir, dirs.StateHome)
		}

		// Verify application-specific directories
		expectedConfigDir := filepath.Join(testConfigDir, AppDirName)
		if dirs.ConfigDir != expectedConfigDir {
			t.Errorf("Expected ConfigDir to be %s, got %s", expectedConfigDir, dirs.ConfigDir)
		}

		expectedCacheDir := filepath.Join(testCacheDir, AppDirName)
		if dirs.CacheDir != expectedCacheDir {
			t.Errorf("Expected CacheDir to be %s, got %s", expectedCacheDir, dirs.CacheDir)
		}

		expectedDataDir := filepath.Join(testDataDir, AppDirName)
		if dirs.DataDir != expectedDataDir {
			t.Errorf("Expected DataDir to be %s, got %s", expectedDataDir, dirs.DataDir)
		}

		// Verify PVM-specific directories
		expectedVersionsDir := filepath.Join(expectedDataDir, VersionsDir)
		if dirs.VersionsDir != expectedVersionsDir {
			t.Errorf("Expected VersionsDir to be %s, got %s", expectedVersionsDir, dirs.VersionsDir)
		}

		expectedSourcesDir := filepath.Join(expectedCacheDir, SourcesDir)
		if dirs.SourcesDir != expectedSourcesDir {
			t.Errorf("Expected SourcesDir to be %s, got %s", expectedSourcesDir, dirs.SourcesDir)
		}
	})

	t.Run("DefaultFallbacks", func(t *testing.T) {
		// Clear environment variables to test defaults
		_ = os.Unsetenv("XDG_CONFIG_HOME")
		_ = os.Unsetenv("XDG_CACHE_HOME")
		_ = os.Unsetenv("XDG_DATA_HOME")
		_ = os.Unsetenv("XDG_STATE_HOME")

		dirs, err := GetDirs()
		if err != nil {
			t.Fatalf("GetDirs() returned an error: %v", err)
		}

		// Verify that we got valid non-empty directories
		// The github.com/adrg/xdg library handles platform-specific defaults
		if dirs.ConfigHome == "" {
			t.Error("Expected non-empty ConfigHome")
		}

		if dirs.CacheHome == "" {
			t.Error("Expected non-empty CacheHome")
		}

		if dirs.DataHome == "" {
			t.Error("Expected non-empty DataHome")
		}

		if dirs.StateHome == "" {
			t.Error("Expected non-empty StateHome")
		}

		// Verify that directories are absolute paths
		if !filepath.IsAbs(dirs.ConfigHome) {
			t.Errorf("ConfigHome should be absolute path, got %s", dirs.ConfigHome)
		}
		if !filepath.IsAbs(dirs.CacheHome) {
			t.Errorf("CacheHome should be absolute path, got %s", dirs.CacheHome)
		}
		if !filepath.IsAbs(dirs.DataHome) {
			t.Errorf("DataHome should be absolute path, got %s", dirs.DataHome)
		}
		if !filepath.IsAbs(dirs.StateHome) {
			t.Errorf("StateHome should be absolute path, got %s", dirs.StateHome)
		}
	})
}

func TestEnsureDirs(t *testing.T) {
	// Create a temporary directory for the test
	testDir := t.TempDir()

	// Set environment variables to use the test directory
	_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(testDir, "config"))
	_ = os.Setenv("XDG_CACHE_HOME", filepath.Join(testDir, "cache"))
	_ = os.Setenv("XDG_DATA_HOME", filepath.Join(testDir, "data"))
	_ = os.Setenv("XDG_STATE_HOME", filepath.Join(testDir, "state"))

	// Get directories
	dirs, err := GetDirs()
	if err != nil {
		t.Fatalf("GetDirs() returned an error: %v", err)
	}

	// Ensure directories exist
	err = dirs.EnsureDirs()
	if err != nil {
		t.Fatalf("EnsureDirs() returned an error: %v", err)
	}

	// Check if directories were created
	dirsToCheck := []string{
		dirs.ConfigDir,
		dirs.CacheDir,
		dirs.DataDir,
		dirs.StateDir,
		dirs.VersionsDir,
		dirs.SourcesDir,
		dirs.ShimsDir,
		dirs.TypeDefinitionsDir,
		dirs.BuildDir,
	}

	for _, dir := range dirsToCheck {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory %s was not created", dir)
		}
	}
}

func TestGetConfigFilePath(t *testing.T) {
	// Create a temporary directory for the test
	testDir := t.TempDir()
	_ = os.Setenv("XDG_CONFIG_HOME", testDir)

	// Get directories
	dirs, err := GetDirs()
	if err != nil {
		t.Fatalf("GetDirs() returned an error: %v", err)
	}

	configPath := dirs.GetConfigFilePath()
	expectedPath := filepath.Join(testDir, AppDirName, "pvm.toml")

	if configPath != expectedPath {
		t.Errorf("Expected config file path to be %s, got %s", expectedPath, configPath)
	}
}

func TestGetProjectConfigPath(t *testing.T) {
	// Create a test project directory
	projectDir := t.TempDir()

	configPath := GetProjectConfigPath(projectDir)
	expectedPath := filepath.Join(projectDir, ".pvm", "pvm.toml")

	if configPath != expectedPath {
		t.Errorf("Expected project config path to be %s, got %s", expectedPath, configPath)
	}
}

func TestGetSystemConfigPath(t *testing.T) {
	configPath := GetSystemConfigPath()
	var expectedPath string

	if runtime.GOOS == "windows" {
		programData := os.Getenv("ProgramData")
		if programData == "" {
			// Skip on Windows if ProgramData is not set
			t.Skip("ProgramData environment variable not set")
		}
		expectedPath = filepath.Join(programData, "pvm", "pvm.toml")
	} else {
		expectedPath = "/etc/pvm/pvm.toml"
	}

	if configPath != expectedPath {
		t.Errorf("Expected system config path to be %s, got %s", expectedPath, configPath)
	}
}
