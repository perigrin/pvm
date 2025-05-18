// ABOUTME: Tests for Perl version string parsing and constraint handling
// ABOUTME: Verifies parsing and comparison of Perl version strings

package perl

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected PerlVersion
		isError  bool
	}{
		{"5.32.1", PerlVersion{5, 32, 1, false}, false},
		{"v5.32.1", PerlVersion{5, 32, 1, false}, false},
		{"5.32", PerlVersion{5, 32, 0, false}, false},
		{"perl-5.32.1", PerlVersion{5, 32, 1, false}, false},
		{"5.32.1-RC1", PerlVersion{5, 32, 1, false}, false},
		{"5.32.1_01", PerlVersion{5, 32, 1, false}, false},
		{"5.39.0", PerlVersion{5, 39, 0, true}, false},      // Dev version (odd minor)
		{"  5.32.1  ", PerlVersion{5, 32, 1, false}, false}, // Whitespace handling
		{"invalid", PerlVersion{}, true},
		{"5.32.x", PerlVersion{}, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			version, err := ParseVersion(test.input)

			if test.isError {
				if err == nil {
					t.Errorf("Expected error for input '%s', got nil", test.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input '%s': %v", test.input, err)
				}

				if version.Major != test.expected.Major ||
					version.Minor != test.expected.Minor ||
					version.Patch != test.expected.Patch ||
					version.Dev != test.expected.Dev {
					t.Errorf("For input '%s', expected %+v, got %+v",
						test.input, test.expected, version)
				}
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		version  PerlVersion
		expected string
	}{
		{PerlVersion{5, 32, 1, false}, "5.32.1"},
		{PerlVersion{5, 38, 0, false}, "5.38.0"},
		{PerlVersion{5, 39, 0, true}, "5.39.0"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.version.String()
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		v1       PerlVersion
		v2       PerlVersion
		expected int
	}{
		{PerlVersion{5, 32, 1, false}, PerlVersion{5, 32, 1, false}, 0},  // Equal
		{PerlVersion{5, 32, 1, false}, PerlVersion{5, 32, 2, false}, -1}, // v1 < v2
		{PerlVersion{5, 32, 2, false}, PerlVersion{5, 32, 1, false}, 1},  // v1 > v2
		{PerlVersion{5, 32, 1, false}, PerlVersion{5, 34, 0, false}, -1}, // Minor version different
		{PerlVersion{5, 34, 0, false}, PerlVersion{5, 32, 1, false}, 1},  // Minor version different
		{PerlVersion{5, 32, 1, false}, PerlVersion{6, 0, 0, false}, -1},  // Major version different
		{PerlVersion{6, 0, 0, false}, PerlVersion{5, 32, 1, false}, 1},   // Major version different
		// Dev status doesn't affect comparison
		{PerlVersion{5, 39, 0, true}, PerlVersion{5, 38, 0, false}, 1},
	}

	for i, test := range tests {
		t.Run(test.v1.String()+" vs "+test.v2.String(), func(t *testing.T) {
			result := test.v1.Compare(test.v2)
			if result != test.expected {
				t.Errorf("Test %d: Expected %d, got %d for %v vs %v",
					i, test.expected, result, test.v1, test.v2)
			}
		})
	}
}

func TestIsStable(t *testing.T) {
	tests := []struct {
		version  PerlVersion
		expected bool
	}{
		{PerlVersion{5, 32, 1, false}, true}, // Stable
		{PerlVersion{5, 38, 0, false}, true}, // Stable
		{PerlVersion{5, 39, 0, true}, false}, // Dev
		{PerlVersion{5, 37, 1, true}, false}, // Dev
	}

	for _, test := range tests {
		t.Run(test.version.String(), func(t *testing.T) {
			result := test.version.IsStable()
			if result != test.expected {
				t.Errorf("Expected %v, got %v for %v",
					test.expected, result, test.version)
			}
		})
	}
}

func TestParseConstraint(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		version  PerlVersion
		isError  bool
	}{
		{">= 5.32.1", ">=", PerlVersion{5, 32, 1, false}, false},
		{">5.32.1", ">", PerlVersion{5, 32, 1, false}, false},
		{"< 5.32.1", "<", PerlVersion{5, 32, 1, false}, false},
		{"<=5.32.1", "<=", PerlVersion{5, 32, 1, false}, false},
		{"=5.32.1", "=", PerlVersion{5, 32, 1, false}, false},
		{"==5.32.1", "==", PerlVersion{5, 32, 1, false}, false},
		{"~> 5.32", "~>", PerlVersion{5, 32, 0, false}, false},
		{"~>5.32.1", "~>", PerlVersion{5, 32, 1, false}, false},
		{"5.32.1", "=", PerlVersion{5, 32, 1, false}, false}, // Implicit =
		{"invalid", "", PerlVersion{}, true},
		{">= invalid", "", PerlVersion{}, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			constraint, err := ParseConstraint(test.input)

			if test.isError {
				if err == nil {
					t.Errorf("Expected error for input '%s', got nil", test.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input '%s': %v", test.input, err)
				}

				if constraint.Operator != test.operator {
					t.Errorf("For input '%s', expected operator '%s', got '%s'",
						test.input, test.operator, constraint.Operator)
				}

				if constraint.Version.Major != test.version.Major ||
					constraint.Version.Minor != test.version.Minor ||
					constraint.Version.Patch != test.version.Patch {
					t.Errorf("For input '%s', expected version %v, got %v",
						test.input, test.version, constraint.Version)
				}
			}
		})
	}
}

func TestConstraintSatisfies(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		satisfies  bool
	}{
		{">= 5.32.0", "5.32.0", true},
		{">= 5.32.0", "5.32.1", true},
		{">= 5.32.0", "5.31.0", false},
		{"> 5.32.0", "5.32.0", false},
		{"> 5.32.0", "5.32.1", true},
		{"< 5.32.0", "5.31.0", true},
		{"< 5.32.0", "5.32.0", false},
		{"<= 5.32.0", "5.32.0", true},
		{"<= 5.32.0", "5.32.1", false},
		{"= 5.32.0", "5.32.0", true},
		{"= 5.32.0", "5.32.1", false},
		{"~> 5.32", "5.32.0", true},
		{"~> 5.32", "5.32.1", true},
		{"~> 5.32", "5.33.0", false},
		{"~> 5.32.1", "5.32.1", true},
		{"~> 5.32.1", "5.32.2", true},
		{"~> 5.32.1", "5.33.0", false},
	}

	for _, test := range tests {
		t.Run(test.constraint+" vs "+test.version, func(t *testing.T) {
			constraint, err := ParseConstraint(test.constraint)
			if err != nil {
				t.Fatalf("Failed to parse constraint '%s': %v", test.constraint, err)
			}

			version, err := ParseVersion(test.version)
			if err != nil {
				t.Fatalf("Failed to parse version '%s': %v", test.version, err)
			}

			result := constraint.Satisfies(version)
			if result != test.satisfies {
				t.Errorf("Expected %v, got %v for %s vs %s",
					test.satisfies, result, test.constraint, test.version)
			}
		})
	}
}

func TestParseConstraintSet(t *testing.T) {
	tests := []struct {
		input   string
		count   int
		isError bool
	}{
		{">= 5.32.0, < 5.34.0", 2, false},
		{">= 5.32.0", 1, false},
		{"", 0, false},
		{",", 0, false},
		{">= 5.32.0, invalid", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			constraints, err := ParseConstraintSet(test.input)

			if test.isError {
				if err == nil {
					t.Errorf("Expected error for input '%s', got nil", test.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input '%s': %v", test.input, err)
				}

				if len(constraints) != test.count {
					t.Errorf("For input '%s', expected %d constraints, got %d",
						test.input, test.count, len(constraints))
				}
			}
		})
	}
}

func TestConstraintSetSatisfies(t *testing.T) {
	tests := []struct {
		constraints string
		version     string
		satisfies   bool
	}{
		{">= 5.32.0, < 5.34.0", "5.32.1", true},
		{">= 5.32.0, < 5.34.0", "5.34.0", false},
		{">= 5.32.0, < 5.34.0", "5.31.0", false},
		{"", "5.32.0", true}, // Empty set is always satisfied
	}

	for _, test := range tests {
		t.Run(test.constraints+" vs "+test.version, func(t *testing.T) {
			constraints, err := ParseConstraintSet(test.constraints)
			if err != nil {
				t.Fatalf("Failed to parse constraints '%s': %v", test.constraints, err)
			}

			version, err := ParseVersion(test.version)
			if err != nil {
				t.Fatalf("Failed to parse version '%s': %v", test.version, err)
			}

			result := constraints.Satisfies(version)
			if result != test.satisfies {
				t.Errorf("Expected %v, got %v for %s vs %s",
					test.satisfies, result, test.constraints, test.version)
			}
		})
	}
}

func TestFindBestMatch(t *testing.T) {
	availableVersions := []string{
		"5.30.0", "5.30.1", "5.30.2",
		"5.32.0", "5.32.1",
		"5.34.0",
	}

	tests := []struct {
		constraint string
		expected   string
		isError    bool
	}{
		{">= 5.32.0", "5.34.0", false},
		{">= 5.32.0, < 5.34.0", "5.32.1", false},
		{"= 5.30.1", "5.30.1", false},
		{">= 5.36.0", "", true}, // No matching version
	}

	for _, test := range tests {
		t.Run(test.constraint, func(t *testing.T) {
			result, err := FindBestMatch(test.constraint, availableVersions)

			if test.isError {
				if err == nil {
					t.Errorf("Expected error for constraint '%s', got nil", test.constraint)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for constraint '%s': %v", test.constraint, err)
				}

				if result != test.expected {
					t.Errorf("For constraint '%s', expected '%s', got '%s'",
						test.constraint, test.expected, result)
				}
			}
		})
	}
}

func TestIsVersionAvailable(t *testing.T) {
	availableVersions := []string{"5.30.0", "5.32.1", "5.34.0"}

	tests := []struct {
		version  string
		expected bool
	}{
		{"5.32.1", true},
		{"5.32.0", false},
		{"5.36.0", false},
	}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			result := IsVersionAvailable(test.version, availableVersions)
			if result != test.expected {
				t.Errorf("For version '%s', expected %v, got %v",
					test.version, test.expected, result)
			}
		})
	}
}

func TestResolveVersionAlias(t *testing.T) {
	aliases := map[string]string{
		"stable": "5.32.1",
		"legacy": "5.30.3",
	}

	tests := []struct {
		alias    string
		expected string
		isError  bool
	}{
		{"5.32.1", "5.32.1", false},  // Not an alias
		{"@stable", "5.32.1", false}, // User-defined alias
		{"@legacy", "5.30.3", false}, // User-defined alias
		{"@unknown", "", true},       // Unknown alias
	}

	for _, test := range tests {
		t.Run(test.alias, func(t *testing.T) {
			// Skip built-in aliases as they depend on hardcoded values or system state
			if test.alias == "@latest" || test.alias == "@latest-stable" ||
				test.alias == "@latest-dev" || test.alias == "@system" {
				t.Skip("Skipping built-in alias")
			}

			result, err := ResolveVersionAlias(test.alias, aliases)

			if test.isError {
				if err == nil {
					t.Errorf("Expected error for alias '%s', got nil", test.alias)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for alias '%s': %v", test.alias, err)
				}

				if result != test.expected {
					t.Errorf("For alias '%s', expected '%s', got '%s'",
						test.alias, test.expected, result)
				}
			}
		})
	}
}
