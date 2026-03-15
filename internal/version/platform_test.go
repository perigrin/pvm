// ABOUTME: Tests for platform detection and binary selection functionality
// ABOUTME: Validates cross-platform asset filtering and compatibility checking

package version

import (
	"runtime"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	platform := DetectPlatform()

	if platform == nil {
		t.Fatal("DetectPlatform returned nil")
	}

	if platform.OS != runtime.GOOS {
		t.Errorf("Expected OS %s, got %s", runtime.GOOS, platform.OS)
	}

	if platform.Architecture != runtime.GOARCH {
		t.Errorf("Expected architecture %s, got %s", runtime.GOARCH, platform.Architecture)
	}

	// Check Windows extension
	if runtime.GOOS == "windows" {
		if platform.Extension != ".exe" {
			t.Errorf("Expected .exe extension on Windows, got %s", platform.Extension)
		}
	} else {
		if platform.Extension != "" {
			t.Errorf("Expected empty extension on non-Windows, got %s", platform.Extension)
		}
	}
}

func TestPlatformString(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		expected string
	}{
		{
			name:     "Linux AMD64",
			platform: Platform{OS: "linux", Architecture: "amd64"},
			expected: "linux-amd64",
		},
		{
			name:     "Darwin ARM64",
			platform: Platform{OS: "darwin", Architecture: "arm64"},
			expected: "darwin-arm64",
		},
		{
			name:     "Windows AMD64",
			platform: Platform{OS: "windows", Architecture: "amd64", Extension: ".exe"},
			expected: "windows-amd64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.platform.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPlatformAssetPattern(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		expected string
	}{
		{
			name:     "Linux AMD64",
			platform: Platform{OS: "linux", Architecture: "amd64"},
			expected: "pvm-linux-amd64",
		},
		{
			name:     "Darwin ARM64",
			platform: Platform{OS: "darwin", Architecture: "arm64"},
			expected: "pvm-darwin-arm64",
		},
		{
			name:     "Windows AMD64",
			platform: Platform{OS: "windows", Architecture: "amd64", Extension: ".exe"},
			expected: "pvm-windows-amd64.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.platform.AssetPattern()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPlatformIsSupported(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		expected bool
	}{
		{
			name:     "Linux AMD64 - supported",
			platform: Platform{OS: "linux", Architecture: "amd64"},
			expected: true,
		},
		{
			name:     "Linux ARM64 - supported",
			platform: Platform{OS: "linux", Architecture: "arm64"},
			expected: true,
		},
		{
			name:     "Darwin AMD64 - supported",
			platform: Platform{OS: "darwin", Architecture: "amd64"},
			expected: true,
		},
		{
			name:     "Darwin ARM64 - supported",
			platform: Platform{OS: "darwin", Architecture: "arm64"},
			expected: true,
		},
		{
			name:     "Windows AMD64 - supported",
			platform: Platform{OS: "windows", Architecture: "amd64"},
			expected: true,
		},
		{
			name:     "Linux 386 - unsupported",
			platform: Platform{OS: "linux", Architecture: "386"},
			expected: false,
		},
		{
			name:     "FreeBSD AMD64 - unsupported",
			platform: Platform{OS: "freebsd", Architecture: "amd64"},
			expected: false,
		},
		{
			name:     "Windows ARM64 - unsupported",
			platform: Platform{OS: "windows", Architecture: "arm64"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.platform.IsSupported()
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestInstallationMethodString(t *testing.T) {
	tests := []struct {
		name     string
		method   InstallationMethod
		expected string
	}{
		{"Binary", InstallationBinary, "binary"},
		{"Homebrew", InstallationHomebrew, "homebrew"},
		{"APT", InstallationAPT, "apt"},
		{"Yum", InstallationYum, "yum"},
		{"Pacman", InstallationPacman, "pacman"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestInstallationMethodCanSelfUpdate(t *testing.T) {
	tests := []struct {
		name     string
		method   InstallationMethod
		expected bool
	}{
		{"Binary can self-update", InstallationBinary, true},
		{"Homebrew cannot self-update", InstallationHomebrew, false},
		{"APT cannot self-update", InstallationAPT, false},
		{"Yum cannot self-update", InstallationYum, false},
		{"Pacman cannot self-update", InstallationPacman, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method.CanSelfUpdate()
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestFilterAssets(t *testing.T) {
	assets := []GitHubAsset{
		{Name: "pvm-linux-amd64", Size: 1000},
		{Name: "pvm-linux-arm64", Size: 1100},
		{Name: "pvm-darwin-amd64", Size: 1200},
		{Name: "pvm-darwin-arm64", Size: 1300},
		{Name: "pvm-windows-amd64.exe", Size: 1400},
		{Name: "checksums.txt", Size: 100},
		{Name: "README.md", Size: 200},
	}

	tests := []struct {
		name     string
		platform Platform
		expected int
	}{
		{
			name:     "Linux AMD64",
			platform: Platform{OS: "linux", Architecture: "amd64"},
			expected: 1,
		},
		{
			name:     "Darwin ARM64",
			platform: Platform{OS: "darwin", Architecture: "arm64"},
			expected: 1,
		},
		{
			name:     "Windows AMD64",
			platform: Platform{OS: "windows", Architecture: "amd64", Extension: ".exe"},
			expected: 1,
		},
		{
			name:     "Unsupported platform",
			platform: Platform{OS: "freebsd", Architecture: "amd64"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterAssets(assets, &tt.platform)
			if len(result) != tt.expected {
				t.Errorf("Expected %d assets, got %d", tt.expected, len(result))
			}

			// Verify correct asset was selected
			if tt.expected == 1 && len(result) > 0 {
				expectedPattern := tt.platform.AssetPattern()
				if !contains(result[0].Name, expectedPattern) {
					t.Errorf("Expected asset name to contain %s, got %s", expectedPattern, result[0].Name)
				}
			}
		})
	}
}

func TestIsAssetMatch(t *testing.T) {
	tests := []struct {
		name      string
		assetName string
		pattern   string
		expected  bool
	}{
		{
			name:      "Simple exact match",
			assetName: "pvm-darwin-arm64",
			pattern:   "pvm-darwin-arm64",
			expected:  true,
		},
		{
			name:      "Versioned asset name",
			assetName: "pvm-1.0.0-rc30-darwin-arm64.tar.gz",
			pattern:   "pvm-darwin-arm64",
			expected:  true,
		},
		{
			name:      "Versioned asset name - Linux",
			assetName: "pvm-1.0.0-rc30-linux-amd64.tar.gz",
			pattern:   "pvm-linux-amd64",
			expected:  true,
		},
		{
			name:      "Versioned asset name - Windows",
			assetName: "pvm-1.0.0-rc30-windows-amd64.exe.zip",
			pattern:   "pvm-windows-amd64",
			expected:  true,
		},
		{
			name:      "Complex version number",
			assetName: "pvm-2.1.0-beta.1-darwin-arm64.tar.gz",
			pattern:   "pvm-darwin-arm64",
			expected:  true,
		},
		{
			name:      "Non-matching OS",
			assetName: "pvm-1.0.0-rc30-linux-amd64.tar.gz",
			pattern:   "pvm-darwin-arm64",
			expected:  false,
		},
		{
			name:      "Non-matching architecture",
			assetName: "pvm-1.0.0-rc30-darwin-amd64.tar.gz",
			pattern:   "pvm-darwin-arm64",
			expected:  false,
		},
		{
			name:      "Wrong tool name",
			assetName: "pvx-1.0.0-rc30-darwin-arm64.tar.gz",
			pattern:   "pvm-darwin-arm64",
			expected:  false,
		},
		{
			name:      "Invalid pattern",
			assetName: "some-file.tar.gz",
			pattern:   "invalid-pattern",
			expected:  false,
		},
		{
			name:      "Asset with extra suffixes",
			assetName: "pvm-1.0.0-rc30-darwin-arm64-signed.tar.gz",
			pattern:   "pvm-darwin-arm64",
			expected:  true,
		},
		{
			name:      "Asset with build metadata",
			assetName: "pvm-1.0.0-rc30+build.123-darwin-arm64.tar.gz",
			pattern:   "pvm-darwin-arm64",
			expected:  true,
		},
		{
			name:      "Substring false positive",
			assetName: "some-pvm-darwin-arm64-tool",
			pattern:   "pvm-darwin-arm64",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAssetMatch(tt.assetName, tt.pattern)
			if result != tt.expected {
				t.Errorf("isAssetMatch(%q, %q) = %v, expected %v",
					tt.assetName, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestFilterAssetsWithVersionedNames(t *testing.T) {
	// Test with realistic versioned asset names like those in actual GitHub releases
	assets := []GitHubAsset{
		{Name: "pvm-1.0.0-rc30-linux-amd64.tar.gz", Size: 1000},
		{Name: "pvm-1.0.0-rc30-linux-arm64.tar.gz", Size: 1100},
		{Name: "pvm-1.0.0-rc30-darwin-amd64.tar.gz", Size: 1200},
		{Name: "pvm-1.0.0-rc30-darwin-arm64.tar.gz", Size: 1300},
		{Name: "pvm-1.0.0-rc30-windows-amd64.exe.zip", Size: 1400},
		{Name: "checksums.txt", Size: 100},
		{Name: "README.md", Size: 200},
	}

	tests := []struct {
		name     string
		platform Platform
		expected int
	}{
		{
			name:     "Linux AMD64 with versioned assets",
			platform: Platform{OS: "linux", Architecture: "amd64"},
			expected: 1,
		},
		{
			name:     "Darwin ARM64 with versioned assets",
			platform: Platform{OS: "darwin", Architecture: "arm64"},
			expected: 1,
		},
		{
			name:     "Windows AMD64 with versioned assets",
			platform: Platform{OS: "windows", Architecture: "amd64", Extension: ".exe"},
			expected: 1,
		},
		{
			name:     "Unsupported platform with versioned assets",
			platform: Platform{OS: "freebsd", Architecture: "amd64"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterAssets(assets, &tt.platform)
			if len(result) != tt.expected {
				t.Errorf("Expected %d assets, got %d", tt.expected, len(result))
			}

			// Verify correct asset was selected
			if tt.expected == 1 && len(result) > 0 {
				expectedPattern := tt.platform.AssetPattern()
				if !isAssetMatch(result[0].Name, expectedPattern) {
					t.Errorf("Expected asset name %s to match pattern %s", result[0].Name, expectedPattern)
				}
			}
		})
	}
}

func TestSelectBestAsset(t *testing.T) {
	assets := []GitHubAsset{
		{Name: "pvm-linux-amd64", Size: 1000},
		{Name: "pvm-darwin-amd64", Size: 1200},
		{Name: "pvm-windows-amd64.exe", Size: 1400},
	}

	tests := []struct {
		name        string
		platform    Platform
		expectError bool
	}{
		{
			name:        "Linux AMD64 - should find asset",
			platform:    Platform{OS: "linux", Architecture: "amd64"},
			expectError: false,
		},
		{
			name:        "Unsupported platform - should error",
			platform:    Platform{OS: "freebsd", Architecture: "amd64"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SelectBestAsset(assets, &tt.platform)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if result != nil {
					t.Error("Expected nil result when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected asset but got nil")
				} else {
					expectedPattern := tt.platform.AssetPattern()
					if !contains(result.Name, expectedPattern) {
						t.Errorf("Expected asset name to contain %s, got %s", expectedPattern, result.Name)
					}
				}
			}
		})
	}
}

func TestGetUpdateAsset(t *testing.T) {
	release := &GitHubRelease{
		TagName: "v1.0.0",
		Assets: []GitHubAsset{
			{Name: "pvm-linux-amd64", Size: 1000},
			{Name: "pvm-darwin-amd64", Size: 1200},
			{Name: "pvm-darwin-arm64", Size: 1300}, // Add ARM64 macOS support
			{Name: "pvm-windows-amd64.exe", Size: 1400},
		},
	}

	tests := []struct {
		name        string
		release     *GitHubRelease
		platform    *Platform
		expectError bool
	}{
		{
			name:        "Valid release and supported platform",
			release:     release,
			platform:    &Platform{OS: "linux", Architecture: "amd64"},
			expectError: false,
		},
		{
			name:        "Valid release, unsupported platform",
			release:     release,
			platform:    &Platform{OS: "freebsd", Architecture: "amd64"},
			expectError: true,
		},
		{
			name:        "Nil release",
			release:     nil,
			platform:    &Platform{OS: "linux", Architecture: "amd64"},
			expectError: true,
		},
		{
			name:        "Nil platform (should use DetectPlatform)",
			release:     release,
			platform:    nil,
			expectError: false, // Should work with detected platform
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetUpdateAsset(tt.release, tt.platform)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if result != nil {
					t.Error("Expected nil result when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected asset but got nil")
				}
			}
		})
	}
}

func TestValidateBinaryCompatibility(t *testing.T) {
	platform := &Platform{OS: "linux", Architecture: "amd64"}

	tests := []struct {
		name        string
		asset       *GitHubAsset
		platform    *Platform
		expectError bool
	}{
		{
			name:        "Compatible asset",
			asset:       &GitHubAsset{Name: "pvm-linux-amd64", Size: 1000},
			platform:    platform,
			expectError: false,
		},
		{
			name:        "Incompatible asset",
			asset:       &GitHubAsset{Name: "pvm-windows-amd64.exe", Size: 1000},
			platform:    platform,
			expectError: true,
		},
		{
			name:        "Nil asset",
			asset:       nil,
			platform:    platform,
			expectError: true,
		},
		{
			name:        "Nil platform",
			asset:       &GitHubAsset{Name: "pvm-linux-amd64", Size: 1000},
			platform:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBinaryCompatibility(tt.asset, tt.platform)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Note: contains function is defined in github_test.go
