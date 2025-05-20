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
	OpRange // Represents a version range (e.g., ">= 1.0, < 2.0")
	OpExact // Exact version match with no operator (e.g., "1.2.3")
)

// VersionConstraint represents a parsed version constraint
type VersionConstraint struct {
	// Raw is the original constraint string
	Raw string

	// Operator is the comparison operator
	Operator ConstraintOperator

	// Version is the version to compare against
	Version string

	// UpperVersion is used for range constraints (e.g., the "2.0" in ">= 1.0, < 2.0")
	UpperVersion string

	// UpperOperator is used for range constraints (e.g., the "<" in ">= 1.0, < 2.0")
	UpperOperator ConstraintOperator
}

// Regular expressions for parsing version constraints
var (
	// Matches the common operators: ==, !=, >, >=, <, <=
	operatorRegex = regexp.MustCompile(`^\s*(==|!=|>=|>|<=|<)\s*(.+)$`)

	// Matches version ranges like ">= 1.0, < 2.0"
	rangeRegex = regexp.MustCompile(`^\s*(>=|>)\s*(.+?)\s*,\s*(<=|<)\s*(.+)\s*$`)

	// Matches simple version numbers without operators
	versionRegex = regexp.MustCompile(`^\s*v?(\d+(\.\d+)*).*$`)
)

// ParseVersionConstraint parses a version constraint string
func ParseVersionConstraint(constraint string) (*VersionConstraint, error) {
	if constraint == "" {
		// Empty constraint means no constraint
		return &VersionConstraint{
			Raw:      "",
			Operator: OpEqual,
			Version:  "",
		}, nil
	}

	// Try to match range pattern first (e.g., ">= 1.0, < 2.0")
	if rangeMatch := rangeRegex.FindStringSubmatch(constraint); len(rangeMatch) >= 5 {
		return &VersionConstraint{
			Raw:           constraint,
			Operator:      OpRange, // Use the special range operator
			Version:       normalizeVersion(rangeMatch[2]),
			UpperOperator: operatorToEnum(rangeMatch[3]),
			UpperVersion:  normalizeVersion(rangeMatch[4]),
		}, nil
	}

	// Try to match operator pattern (e.g., ">= 1.0")
	if opMatch := operatorRegex.FindStringSubmatch(constraint); len(opMatch) >= 3 {
		return &VersionConstraint{
			Raw:      constraint,
			Operator: operatorToEnum(opMatch[1]),
			Version:  normalizeVersion(opMatch[2]),
		}, nil
	}

	// Try to match simple version number (e.g., "1.0")
	if verMatch := versionRegex.FindStringSubmatch(constraint); len(verMatch) >= 2 {
		return &VersionConstraint{
			Raw:      constraint,
			Operator: OpExact,
			Version:  normalizeVersion(verMatch[0]),
		}, nil
	}

	// If constraint is not valid (e.g., "invalid"), let's try to handle it
	if versionRegex.MatchString(constraint) {
		return &VersionConstraint{
			Raw:      constraint,
			Operator: OpExact,
			Version:  normalizeVersion(constraint),
		}, nil
	}

	// If we get here, the constraint couldn't be parsed
	return nil, errors.NewSystemError(
		ErrInvalidVersionPattern,
		fmt.Sprintf("Invalid version constraint pattern: %s", constraint),
		nil)
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

	// If the constraint is empty, it's satisfied by any version
	if parsedConstraint.Raw == "" {
		return true, nil
	}

	// Normalize the version
	normVersion := normalizeVersion(version)

	// Check based on operator
	switch parsedConstraint.Operator {
	case OpEqual, OpExact:
		return compareVersions(normVersion, parsedConstraint.Version) == 0, nil
	case OpNotEqual:
		return compareVersions(normVersion, parsedConstraint.Version) != 0, nil
	case OpGreaterThan:
		return compareVersions(normVersion, parsedConstraint.Version) > 0, nil
	case OpGreaterThanOrEqual:
		return compareVersions(normVersion, parsedConstraint.Version) >= 0, nil
	case OpLessThan:
		return compareVersions(normVersion, parsedConstraint.Version) < 0, nil
	case OpLessThanOrEqual:
		return compareVersions(normVersion, parsedConstraint.Version) <= 0, nil
	case OpRange:
		// For a range, both conditions must be true
		lowerSatisfied := false
		upperSatisfied := false

		// Check lower bound
		switch parsedConstraint.Operator {
		case OpGreaterThan:
			lowerSatisfied = compareVersions(normVersion, parsedConstraint.Version) > 0
		case OpGreaterThanOrEqual:
			lowerSatisfied = compareVersions(normVersion, parsedConstraint.Version) >= 0
		}

		// Check upper bound
		switch parsedConstraint.UpperOperator {
		case OpLessThan:
			upperSatisfied = compareVersions(normVersion, parsedConstraint.UpperVersion) < 0
		case OpLessThanOrEqual:
			upperSatisfied = compareVersions(normVersion, parsedConstraint.UpperVersion) <= 0
		}

		return lowerSatisfied && upperSatisfied, nil
	default:
		return false, errors.NewSystemError(
			ErrInvalidVersionPattern,
			fmt.Sprintf("Unknown constraint operator: %v", parsedConstraint.Operator),
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
