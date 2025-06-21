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

func TestGetPlatformTriple(t *testing.T) {
	triple := GetPlatformTriple()

	// Should be in format "goos-goarch"
	if triple == "" {
		t.Error("Platform triple should not be empty")
	}

	// Should contain runtime.GOOS and runtime.GOARCH
	expectedTriple := runtime.GOOS + "-" + runtime.GOARCH
	if triple != expectedTriple {
		t.Errorf("Expected platform triple %s, got %s", expectedTriple, triple)
	}

	// Even if not a known platform, should still be properly formatted
	if len(triple) < 5 { // minimum: "os-arch" (at least 5 chars)
		t.Errorf("Platform triple seems malformed: %s", triple)
	}
}

func TestGetBinaryExtension(t *testing.T) {
	ext := GetBinaryExtension()

	if runtime.GOOS == "windows" {
		if ext != ".exe" {
			t.Errorf("Expected .exe extension on Windows, got %s", ext)
		}
	} else {
		if ext != "" {
			t.Errorf("Expected empty extension on Unix, got %s", ext)
		}
	}
}

func TestGetArchiveExtension(t *testing.T) {
	ext := GetArchiveExtension()

	if runtime.GOOS == "windows" {
		if ext != ".zip" {
			t.Errorf("Expected .zip extension on Windows, got %s", ext)
		}
	} else {
		if ext != ".tar.gz" {
			t.Errorf("Expected .tar.gz extension on Unix, got %s", ext)
		}
	}
}

func TestIsSupportedPlatform(t *testing.T) {
	// Test current platform
	isSupported := IsSupportedPlatform()

	// Define expected support based on the implementation
	expectedSupported := false
	switch runtime.GOOS {
	case "linux":
		expectedSupported = runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64"
	case "darwin":
		expectedSupported = runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64"
	case "windows":
		expectedSupported = runtime.GOARCH == "amd64"
	}

	if isSupported != expectedSupported {
		t.Errorf("Expected platform support %v for %s-%s, got %v",
			expectedSupported, runtime.GOOS, runtime.GOARCH, isSupported)
	}
}

func TestBinaryDistributionFunctions(t *testing.T) {
	// Test that all functions work together coherently
	triple := GetPlatformTriple()
	binExt := GetBinaryExtension()
	archExt := GetArchiveExtension()
	supported := IsSupportedPlatform()

	t.Logf("Platform: %s", triple)
	t.Logf("Binary extension: '%s'", binExt)
	t.Logf("Archive extension: '%s'", archExt)
	t.Logf("Supported: %v", supported)

	// Basic sanity checks
	if triple == "" {
		t.Error("Platform triple should not be empty")
	}

	// Binary extension should be either empty or .exe
	if binExt != "" && binExt != ".exe" {
		t.Errorf("Unexpected binary extension: %s", binExt)
	}

	// Archive extension should be either .tar.gz or .zip
	if archExt != ".tar.gz" && archExt != ".zip" {
		t.Errorf("Unexpected archive extension: %s", archExt)
	}
}

func TestPlatformTripleFormatting(t *testing.T) {
	testCases := []struct {
		goos     string
		goarch   string
		expected string
	}{
		{"linux", "amd64", "linux-amd64"},
		{"linux", "arm64", "linux-arm64"},
		{"darwin", "amd64", "darwin-amd64"},
		{"darwin", "arm64", "darwin-arm64"},
		{"windows", "amd64", "windows-amd64"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			// We can't change runtime.GOOS/GOARCH in tests, but we can verify the format
			// This test ensures our function produces the expected format
			if runtime.GOOS == tc.goos && runtime.GOARCH == tc.goarch {
				actual := GetPlatformTriple()
				if actual != tc.expected {
					t.Errorf("Expected %s, got %s", tc.expected, actual)
				}
			}
		})
	}
}
