// ABOUTME: Tests for semantic version parsing and comparison utilities
// ABOUTME: Comprehensive test coverage for all version operations

package version

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *SemanticVersion
		expectError bool
	}{
		{
			name:  "basic version",
			input: "1.2.3",
			expected: &SemanticVersion{
				Major: 1, Minor: 2, Patch: 3,
				Original: "1.2.3",
			},
		},
		{
			name:  "version with v prefix",
			input: "v1.2.3",
			expected: &SemanticVersion{
				Major: 1, Minor: 2, Patch: 3,
				Original: "v1.2.3",
			},
		},
		{
			name:  "version with prerelease",
			input: "1.2.3-alpha.1",
			expected: &SemanticVersion{
				Major: 1, Minor: 2, Patch: 3,
				Prerelease: "alpha.1",
				Original:   "1.2.3-alpha.1",
			},
		},
		{
			name:  "version with build metadata",
			input: "1.2.3+build.123",
			expected: &SemanticVersion{
				Major: 1, Minor: 2, Patch: 3,
				Build:    "build.123",
				Original: "1.2.3+build.123",
			},
		},
		{
			name:  "version with prerelease and build",
			input: "1.2.3-beta.2+build.456",
			expected: &SemanticVersion{
				Major: 1, Minor: 2, Patch: 3,
				Prerelease: "beta.2",
				Build:      "build.456",
				Original:   "1.2.3-beta.2+build.456",
			},
		},
		{
			name:  "major only",
			input: "1",
			expected: &SemanticVersion{
				Major: 1, Minor: 0, Patch: 0,
				Original: "1",
			},
		},
		{
			name:  "major.minor only",
			input: "1.2",
			expected: &SemanticVersion{
				Major: 1, Minor: 2, Patch: 0,
				Original: "1.2",
			},
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "invalid format",
			input:       "not.a.version",
			expectError: true,
		},
		{
			name:        "negative numbers",
			input:       "-1.2.3",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseVersion(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Major != tt.expected.Major {
				t.Errorf("Major: expected %d, got %d", tt.expected.Major, result.Major)
			}
			if result.Minor != tt.expected.Minor {
				t.Errorf("Minor: expected %d, got %d", tt.expected.Minor, result.Minor)
			}
			if result.Patch != tt.expected.Patch {
				t.Errorf("Patch: expected %d, got %d", tt.expected.Patch, result.Patch)
			}
			if result.Prerelease != tt.expected.Prerelease {
				t.Errorf("Prerelease: expected %s, got %s", tt.expected.Prerelease, result.Prerelease)
			}
			if result.Build != tt.expected.Build {
				t.Errorf("Build: expected %s, got %s", tt.expected.Build, result.Build)
			}
			if result.Original != tt.expected.Original {
				t.Errorf("Original: expected %s, got %s", tt.expected.Original, result.Original)
			}
		})
	}
}

func TestSemanticVersionString(t *testing.T) {
	tests := []struct {
		name     string
		version  *SemanticVersion
		expected string
	}{
		{
			name:     "basic version",
			version:  &SemanticVersion{Major: 1, Minor: 2, Patch: 3},
			expected: "1.2.3",
		},
		{
			name:     "with prerelease",
			version:  &SemanticVersion{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1"},
			expected: "1.2.3-alpha.1",
		},
		{
			name:     "with build",
			version:  &SemanticVersion{Major: 1, Minor: 2, Patch: 3, Build: "build.123"},
			expected: "1.2.3+build.123",
		},
		{
			name:     "with prerelease and build",
			version:  &SemanticVersion{Major: 1, Minor: 2, Patch: 3, Prerelease: "beta.2", Build: "build.456"},
			expected: "1.2.3-beta.2+build.456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSemanticVersionCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"equal versions", "1.2.3", "1.2.3", 0},
		{"major version diff", "2.0.0", "1.9.9", 1},
		{"minor version diff", "1.3.0", "1.2.9", 1},
		{"patch version diff", "1.2.4", "1.2.3", 1},
		{"prerelease vs release", "1.0.0", "1.0.0-alpha", 1},
		{"prerelease comparison", "1.0.0-alpha.2", "1.0.0-alpha.1", 1},
		{"prerelease vs prerelease", "1.0.0-beta", "1.0.0-alpha", 1},
		{"complex prerelease", "1.0.0-alpha.1.2", "1.0.0-alpha.1.1", 1},
		{"reverse major", "1.0.0", "2.0.0", -1},
		{"reverse minor", "1.1.0", "1.2.0", -1},
		{"reverse patch", "1.0.1", "1.0.2", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			va, err := ParseVersion(tt.a)
			if err != nil {
				t.Fatalf("error parsing version %s: %v", tt.a, err)
			}

			vb, err := ParseVersion(tt.b)
			if err != nil {
				t.Fatalf("error parsing version %s: %v", tt.b, err)
			}

			result := va.Compare(vb)
			if result != tt.expected {
				t.Errorf("Compare(%s, %s): expected %d, got %d", tt.a, tt.b, tt.expected, result)
			}

			// Test convenience methods
			switch {
			case tt.expected > 0:
				if !va.IsNewer(vb) {
					t.Errorf("IsNewer(%s, %s): expected true", tt.a, tt.b)
				}
				if va.IsOlder(vb) {
					t.Errorf("IsOlder(%s, %s): expected false", tt.a, tt.b)
				}
				if va.IsEqual(vb) {
					t.Errorf("IsEqual(%s, %s): expected false", tt.a, tt.b)
				}
			case tt.expected < 0:
				if va.IsNewer(vb) {
					t.Errorf("IsNewer(%s, %s): expected false", tt.a, tt.b)
				}
				if !va.IsOlder(vb) {
					t.Errorf("IsOlder(%s, %s): expected true", tt.a, tt.b)
				}
				if va.IsEqual(vb) {
					t.Errorf("IsEqual(%s, %s): expected false", tt.a, tt.b)
				}
			default:
				if va.IsNewer(vb) {
					t.Errorf("IsNewer(%s, %s): expected false", tt.a, tt.b)
				}
				if va.IsOlder(vb) {
					t.Errorf("IsOlder(%s, %s): expected false", tt.a, tt.b)
				}
				if !va.IsEqual(vb) {
					t.Errorf("IsEqual(%s, %s): expected true", tt.a, tt.b)
				}
			}
		})
	}
}

func TestIsPrerelease(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"release version", "1.2.3", false},
		{"alpha prerelease", "1.2.3-alpha", true},
		{"beta prerelease", "1.2.3-beta.1", true},
		{"rc prerelease", "1.2.3-rc.1", true},
		{"with build metadata", "1.2.3+build", false},
		{"prerelease with build", "1.2.3-alpha+build", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ParseVersion(tt.version)
			if err != nil {
				t.Fatalf("error parsing version %s: %v", tt.version, err)
			}

			result := v.IsPrerelease()
			if result != tt.expected {
				t.Errorf("IsPrerelease(%s): expected %t, got %t", tt.version, tt.expected, result)
			}
		})
	}
}

func TestNormalizeVersionString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"version without v", "1.2.3", "v1.2.3"},
		{"version with v", "v1.2.3", "v1.2.3"},
		{"prerelease without v", "1.2.3-alpha", "v1.2.3-alpha"},
		{"prerelease with v", "v1.2.3-alpha", "v1.2.3-alpha"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeVersionString(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeVersionString(%s): expected %s, got %s", tt.input, tt.expected, result)
			}
		})
	}
}

func TestStripVersionPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"version with v", "v1.2.3", "1.2.3"},
		{"version without v", "1.2.3", "1.2.3"},
		{"empty string", "", ""},
		{"just v", "v", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripVersionPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("StripVersionPrefix(%s): expected %s, got %s", tt.input, tt.expected, result)
			}
		})
	}
}

func TestCompareVersionStrings(t *testing.T) {
	tests := []struct {
		name        string
		a           string
		b           string
		expected    int
		expectError bool
	}{
		{"equal versions", "1.2.3", "1.2.3", 0, false},
		{"a newer", "1.2.4", "1.2.3", 1, false},
		{"b newer", "1.2.3", "1.2.4", -1, false},
		{"invalid a", "invalid", "1.2.3", 0, true},
		{"invalid b", "1.2.3", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompareVersionStrings(tt.a, tt.b)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("CompareVersionStrings(%s, %s): expected %d, got %d", tt.a, tt.b, tt.expected, result)
			}
		})
	}
}

func TestIsVersionNewer(t *testing.T) {
	tests := []struct {
		name        string
		a           string
		b           string
		expected    bool
		expectError bool
	}{
		{"a newer", "1.2.4", "1.2.3", true, false},
		{"b newer", "1.2.3", "1.2.4", false, false},
		{"equal", "1.2.3", "1.2.3", false, false},
		{"invalid a", "invalid", "1.2.3", false, true},
		{"invalid b", "1.2.3", "invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsVersionNewer(tt.a, tt.b)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("IsVersionNewer(%s, %s): expected %t, got %t", tt.a, tt.b, tt.expected, result)
			}
		})
	}
}
