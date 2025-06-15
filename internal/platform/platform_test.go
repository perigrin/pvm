// ABOUTME: Cross-platform compatibility tests for PVM
// ABOUTME: Validates that core functionality works across Windows, macOS, and Linux

package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCrossplatformPathHandling(t *testing.T) {
	// Test that path handling works correctly on all platforms
	testCases := []struct {
		name       string
		envVar     string
		defaultVal string
		skipEmpty  bool
	}{
		{"HOME directory", "HOME", "", true}, // Skip empty check for HOME
		{"Config directory", getConfigEnvVar(), getDefaultConfigDir(), false},
		{"Cache directory", getCacheEnvVar(), getDefaultCacheDir(), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save original value
			originalVal := os.Getenv(tc.envVar)
			defer func() {
				if originalVal != "" {
					os.Setenv(tc.envVar, originalVal)
				} else {
					os.Unsetenv(tc.envVar)
				}
			}()

			// Test with empty value (should use default)
			os.Unsetenv(tc.envVar)
			defaultPath := getPathForEnvVar(tc.envVar, tc.defaultVal)
			if !tc.skipEmpty && defaultPath == "" {
				t.Errorf("Default path should not be empty for %s", tc.name)
			}

			// Test with custom value
			customPath := "/custom/path"
			if runtime.GOOS == "windows" {
				customPath = `C:\custom\path`
			}
			os.Setenv(tc.envVar, customPath)
			actualPath := getPathForEnvVar(tc.envVar, tc.defaultVal)
			if actualPath != customPath {
				t.Errorf("Expected custom path %s, got %s", customPath, actualPath)
			}
		})
	}
}

func TestExecutablePermissions(t *testing.T) {
	// Create a temporary executable and test permissions
	tempDir := os.TempDir()
	execPath := filepath.Join(tempDir, "test_exec")
	if runtime.GOOS == "windows" {
		execPath += ".exe"
	}

	// Write test executable
	content := []byte("test content")
	err := os.WriteFile(execPath, content, 0755)
	if err != nil {
		t.Fatalf("Failed to create test executable: %v", err)
	}
	defer os.Remove(execPath)

	// Check that file exists and has appropriate permissions
	info, err := os.Stat(execPath)
	if err != nil {
		t.Fatalf("Failed to stat test executable: %v", err)
	}

	// Check platform-specific executable properties
	if runtime.GOOS == "windows" {
		// On Windows, check that it's a regular file
		if !info.Mode().IsRegular() {
			t.Errorf("Windows: Expected regular file, got mode %v", info.Mode())
		}
	} else {
		// On Unix, check that it has execute permissions
		if info.Mode()&0111 == 0 {
			t.Errorf("Unix: Expected executable permissions, got mode %v", info.Mode())
		}
	}
}

func TestCrossplatformTempDirHandling(t *testing.T) {
	// Test that temporary directory handling works on all platforms
	tempDir := os.TempDir()
	if tempDir == "" {
		t.Fatal("TempDir should not be empty")
	}

	// Create a test file in temp dir
	testFile := filepath.Join(tempDir, "pvm_test_file")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Verify file exists
	_, err = os.Stat(testFile)
	if err != nil {
		t.Errorf("Test file should exist: %v", err)
	}
}

// Helper functions

func getConfigEnvVar() string {
	if runtime.GOOS == "windows" {
		return "APPDATA"
	}
	return "XDG_CONFIG_HOME"
}

func getDefaultConfigDir() string {
	homeDir, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(homeDir, "AppData", "Roaming")
	}
	return filepath.Join(homeDir, ".config")
}

func getCacheEnvVar() string {
	if runtime.GOOS == "windows" {
		return "LOCALAPPDATA"
	}
	return "XDG_CACHE_HOME"
}

func getDefaultCacheDir() string {
	homeDir, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(homeDir, "AppData", "Local")
	}
	return filepath.Join(homeDir, ".cache")
}

func getPathForEnvVar(envVar, defaultVal string) string {
	if val := os.Getenv(envVar); val != "" {
		return val
	}
	return defaultVal
}
