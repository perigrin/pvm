// ABOUTME: Version constraint handling for dependencies
// ABOUTME: Implements parsing and checking version constraints

package deps

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/errors"
)

// ConstraintOperator represents a comparison operator for version constraints
type ConstraintOperator int

const (
	OpEqual ConstraintOperator = iota
	OpNotEqual
	OpGreaterThan
	OpGreaterThanOrEqual
	OpLessThan
	OpLessThanOrEqual
)

// VersionConstraint represents a parsed version constraint
type VersionConstraint struct {
	// Raw is the original constraint string
	Raw string

	// Constraints is a list of individual constraints that must all be satisfied (AND logic)
	Constraints []SingleConstraint
}

// SingleConstraint represents a single version comparison
type SingleConstraint struct {
	// Operator is the comparison operator
	Operator ConstraintOperator

	// Version is the version to compare against
	Version string
}

// Regular expressions for parsing version constraints
var (
	// Matches the common operators: ==, !=, >, >=, <, <=
	// Note: Order matters - check two-char operators before single-char ones
	// The version part must start with a digit or 'v'
	operatorRegex = regexp.MustCompile(`^\s*(==|!=|>=|<=|>|<)\s*(v?\d.*)$`)

	// Matches simple version numbers without operators
	versionRegex = regexp.MustCompile(`^\s*v?(\d+(\.\d+)*).*$`)
)

// ParseVersionConstraint parses a version constraint string
func ParseVersionConstraint(constraint string) (*VersionConstraint, error) {
	result := &VersionConstraint{
		Raw:         constraint,
		Constraints: []SingleConstraint{},
	}

	if constraint == "" {
		// Empty constraint means no constraint (any version)
		return result, nil
	}

	// Split by comma for multiple constraints
	parts := strings.Split(constraint, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Try to match operator pattern (e.g., ">= 1.0", "< 2.0", "!= 1.5")
		if opMatch := operatorRegex.FindStringSubmatch(part); len(opMatch) >= 3 {
			result.Constraints = append(result.Constraints, SingleConstraint{
				Operator: operatorToEnum(opMatch[1]),
				Version:  normalizeVersion(opMatch[2]),
			})
			continue
		}

		// Try to match simple version number (e.g., "1.0")
		// In Perl/cpanfile, a bare version number means "this version or higher"
		if verMatch := versionRegex.FindStringSubmatch(part); len(verMatch) >= 2 {
			result.Constraints = append(result.Constraints, SingleConstraint{
				Operator: OpGreaterThanOrEqual, // Bare version means >= in Perl
				Version:  normalizeVersion(verMatch[0]),
			})
			continue
		}

		// If we can't parse this part, return an error
		return nil, errors.NewSystemError(
			ErrInvalidVersionPattern,
			fmt.Sprintf("Invalid version constraint pattern: %s", part),
			nil)
	}

	// If no constraints were parsed, but we had non-empty input, it's an error
	if len(result.Constraints) == 0 && constraint != "" {
		return nil, errors.NewSystemError(
			ErrInvalidVersionPattern,
			fmt.Sprintf("Invalid version constraint pattern: %s", constraint),
			nil)
	}

	return result, nil
}

// operatorToEnum converts a string operator to a ConstraintOperator
func operatorToEnum(op string) ConstraintOperator {
	switch op {
	case "==":
		return OpEqual
	case "!=":
		return OpNotEqual
	case ">":
		return OpGreaterThan
	case ">=":
		return OpGreaterThanOrEqual
	case "<":
		return OpLessThan
	case "<=":
		return OpLessThanOrEqual
	default:
		return OpEqual // Default to equal for unknown operators
	}
}

// normalizeVersion cleans up a version string
func normalizeVersion(version string) string {
	// Remove leading 'v' if present
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	return version
}

// CheckVersionConstraint checks if a version satisfies a constraint
func CheckVersionConstraint(version, constraint string) (bool, error) {
	// Parse the constraint
	parsedConstraint, err := ParseVersionConstraint(constraint)
	if err != nil {
		return false, err
	}

	// If there are no constraints, it's satisfied by any version
	if len(parsedConstraint.Constraints) == 0 {
		return true, nil
	}

	// Normalize the version
	normVersion := normalizeVersion(version)

	// All constraints must be satisfied (AND logic)
	for _, c := range parsedConstraint.Constraints {
		satisfied, err := checkSingleConstraint(normVersion, c)
		if err != nil {
			return false, err
		}
		if !satisfied {
			return false, nil
		}
	}

	return true, nil
}

// checkSingleConstraint checks if a version satisfies a single constraint
func checkSingleConstraint(version string, constraint SingleConstraint) (bool, error) {
	switch constraint.Operator {
	case OpEqual:
		return compareVersions(version, constraint.Version) == 0, nil
	case OpNotEqual:
		return compareVersions(version, constraint.Version) != 0, nil
	case OpGreaterThan:
		return compareVersions(version, constraint.Version) > 0, nil
	case OpGreaterThanOrEqual:
		return compareVersions(version, constraint.Version) >= 0, nil
	case OpLessThan:
		return compareVersions(version, constraint.Version) < 0, nil
	case OpLessThanOrEqual:
		return compareVersions(version, constraint.Version) <= 0, nil
	default:
		return false, errors.NewSystemError(
			ErrInvalidVersionPattern,
			fmt.Sprintf("Unknown constraint operator: %v", constraint.Operator),
			nil)
	}
}

// compareVersions compares two version strings
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	// Split the versions by dots
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Compare each part
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int
		if i < len(parts1) {
			num1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			num2, _ = strconv.Atoi(parts2[i])
		}

		if num1 > num2 {
			return 1
		} else if num1 < num2 {
			return -1
		}
	}

	// If we get here, the versions are equal up to the number of parts
	return 0
}
