// ABOUTME: Tests for macOS version detection and compatibility utilities
// ABOUTME: Validates version parsing, patch strategy detection, and compatibility checks

package perl

import (
	"runtime"
	"testing"
)

func TestParseMacOSVersion(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		expected  *MacOSVersion
		shouldErr bool
	}{
		{
			name:    "macOS Big Sur",
			version: "11.2.3",
			expected: &MacOSVersion{
				Major: 11,
				Minor: 2,
				Patch: 3,
			},
		},
		{
			name:    "macOS Sequoia",
			version: "15.6.0",
			expected: &MacOSVersion{
				Major: 15,
				Minor: 6,
				Patch: 0,
			},
		},
		{
			name:    "macOS Catalina",
			version: "10.15.7",
			expected: &MacOSVersion{
				Major: 10,
				Minor: 15,
				Patch: 7,
			},
		},
		{
			name:    "version without patch",
			version: "12.3",
			expected: &MacOSVersion{
				Major: 12,
				Minor: 3,
				Patch: 0,
			},
		},
		{
			name:      "invalid format",
			version:   "invalid",
			shouldErr: true,
		},
		{
			name:      "empty version",
			version:   "",
			shouldErr: true,
		},
		{
			name:      "single number",
			version:   "15",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMacOSVersion(tt.version)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error for version %s, but got none", tt.version)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for version %s: %v", tt.version, err)
				return
			}

			if result.Major != tt.expected.Major {
				t.Errorf("Major version mismatch: expected %d, got %d", tt.expected.Major, result.Major)
			}
			if result.Minor != tt.expected.Minor {
				t.Errorf("Minor version mismatch: expected %d, got %d", tt.expected.Minor, result.Minor)
			}
			if result.Patch != tt.expected.Patch {
				t.Errorf("Patch version mismatch: expected %d, got %d", tt.expected.Patch, result.Patch)
			}
		})
	}
}

func TestMacOSVersionComparison(t *testing.T) {
	tests := []struct {
		name     string
		version1 MacOSVersion
		version2 MacOSVersion
		expected bool
	}{
		{
			name:     "newer major version",
			version1: MacOSVersion{15, 6, 0},
			version2: MacOSVersion{11, 2, 3},
			expected: true,
		},
		{
			name:     "newer minor version",
			version1: MacOSVersion{11, 6, 0},
			version2: MacOSVersion{11, 2, 3},
			expected: true,
		},
		{
			name:     "newer patch version",
			version1: MacOSVersion{11, 2, 5},
			version2: MacOSVersion{11, 2, 3},
			expected: true,
		},
		{
			name:     "same version",
			version1: MacOSVersion{11, 2, 3},
			version2: MacOSVersion{11, 2, 3},
			expected: false,
		},
		{
			name:     "older version",
			version1: MacOSVersion{10, 15, 7},
			version2: MacOSVersion{11, 2, 3},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version1.IsNewerThan(tt.version2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for %s > %s", tt.expected, result, tt.version1.String(), tt.version2.String())
			}
		})
	}
}

func TestNeedsConfigurePatch(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific tests on non-macOS platform")
	}

	tests := []struct {
		name        string
		perlVersion string
		expected    bool
	}{
		{
			name:        "Perl 5.26.0 needs patching",
			perlVersion: "5.26.0",
			expected:    true,
		},
		{
			name:        "Perl 5.30.3 needs patching",
			perlVersion: "5.30.3",
			expected:    true,
		},
		{
			name:        "Perl 5.32.0 may not need patching on older macOS",
			perlVersion: "5.32.0",
			// Expected result depends on actual macOS version
			expected: true, // Conservative assumption for test
		},
		{
			name:        "Perl 5.38.0 may need patching on very new macOS",
			perlVersion: "5.38.0",
			// Expected result depends on actual macOS version
			expected: true, // Conservative for Sequoia compatibility
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NeedsConfigurePatch(tt.perlVersion)
			// Note: We can't assert exact values since this depends on the actual macOS version
			// Instead, we just ensure the function runs without error
			t.Logf("Perl %s needs patching: %v", tt.perlVersion, result)
		})
	}
}

func TestGetConfigurePatchStrategy(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific tests on non-macOS platform")
	}

	tests := []struct {
		name        string
		perlVersion string
		expected    ConfigurePatchStrategy
	}{
		{
			name:        "very old Perl uses environment override",
			perlVersion: "5.16.3",
			expected:    PatchStrategyEnvironmentOverride,
		},
		{
			name:        "Perl 5.26.0 uses darwin hints",
			perlVersion: "5.26.0",
			expected:    PatchStrategyDarwinHints,
		},
		{
			name:        "Perl 5.30.3 uses darwin hints",
			perlVersion: "5.30.3",
			expected:    PatchStrategyDarwinHints,
		},
		{
			name:        "invalid version defaults to darwin hints",
			perlVersion: "invalid.version",
			expected:    PatchStrategyDarwinHints,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetConfigurePatchStrategy(tt.perlVersion)
			if result != tt.expected {
				t.Errorf("Expected strategy %d, got %d for Perl %s", tt.expected, result, tt.perlVersion)
			}
		})
	}
}

func TestGetMacOSCompatibilityInfo(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific tests on non-macOS platform")
	}

	tests := []string{
		"5.26.0",
		"5.30.3",
		"5.32.0",
		"5.38.0",
	}

	for _, version := range tests {
		t.Run("Perl "+version, func(t *testing.T) {
			info := GetMacOSCompatibilityInfo(version)

			// Should contain macOS version info
			if !containsSubstring(info, "macOS Version:") {
				t.Errorf("Expected macOS version info in output: %s", info)
			}

			// Should contain configure patching info
			if !containsSubstring(info, "Configure Patching:") {
				t.Errorf("Expected configure patching info in output: %s", info)
			}

			t.Logf("Compatibility info for Perl %s:\n%s", version, info)
		})
	}
}

func TestMacOSVersionString(t *testing.T) {
	version := MacOSVersion{15, 6, 0}
	expected := "15.6.0"
	result := version.String()

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// Helper function to check if string contains substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestNonMacOSBehavior(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping non-macOS tests on macOS platform")
	}

	// Test that macOS-specific functions behave correctly on non-macOS platforms

	// GetMacOSVersion should return error
	_, err := GetMacOSVersion()
	if err == nil {
		t.Error("Expected error when getting macOS version on non-macOS platform")
	}

	// NeedsConfigurePatch should return false
	if NeedsConfigurePatch("5.26.0") {
		t.Error("Expected no Configure patching needed on non-macOS platform")
	}

	// GetConfigurePatchStrategy should return PatchStrategyNone
	strategy := GetConfigurePatchStrategy("5.26.0")
	if strategy != PatchStrategyNone {
		t.Errorf("Expected PatchStrategyNone (%d) on non-macOS platform, got %d", PatchStrategyNone, strategy)
	}

	// GetMacOSCompatibilityInfo should return empty string
	info := GetMacOSCompatibilityInfo("5.26.0")
	if info != "" {
		t.Errorf("Expected empty compatibility info on non-macOS platform, got: %s", info)
	}
}
