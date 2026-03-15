// ABOUTME: Perl version string parsing and constraint handling
// ABOUTME: Provides functionality for parsing and comparing Perl version strings

package perl

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/errors"
)

// Version error codes
const (
	ErrInvalidVersion      = "101" // Invalid version string format
	ErrInvalidConstraint   = "102" // Invalid constraint string format
	ErrUnsatisfiable       = "103" // Unsatisfiable constraint
	ErrInvalidVersionAlias = "104" // Invalid version alias
)

// PerlVersion represents a parsed Perl version (major.minor.patch)
type PerlVersion struct {
	Major int
	Minor int
	Patch int
	Dev   bool // True for development versions (e.g., 5.39.0)
}

// Constraint represents a version constraint with an operator and version
type Constraint struct {
	Operator string // >, <, >=, <=, ==, =, ~>
	Version  PerlVersion
}

// ConstraintSet is a collection of constraints that must all be satisfied
type ConstraintSet []Constraint

// parseVersionFunc parses a version string into a PerlVersion struct
func parseVersionFunc(version string) (PerlVersion, error) {
	// Check for common formats and normalize
	version = strings.TrimSpace(version)

	// Remove "perl-" prefix if present
	version = strings.TrimPrefix(version, "perl-")

	// Remove "v" prefix if present
	version = strings.TrimPrefix(version, "v")

	// Define regex patterns for valid version formats
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`),         // 5.32.1
		regexp.MustCompile(`^(\d+)\.(\d+)$`),                // 5.32
		regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)-RC(\d+)$`), // 5.32.0-RC1 (treat as 5.32.0)
		regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)_(\d+)$`),   // 5.32.1_01 (treat as 5.32.1)
	}

	// Try each pattern
	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(version)
		if matches != nil {
			major, _ := strconv.Atoi(matches[1])
			minor, _ := strconv.Atoi(matches[2])

			// Default patch to 0 if not present
			patch := 0
			if len(matches) > 3 {
				patch, _ = strconv.Atoi(matches[3])
			}

			// Check if it's a development version
			// (odd-numbered minor versions are considered development releases)
			isDev := minor%2 != 0

			return PerlVersion{
				Major: major,
				Minor: minor,
				Patch: patch,
				Dev:   isDev,
			}, nil
		}
	}

	return PerlVersion{}, errors.NewVersionError(
		ErrInvalidVersion,
		fmt.Sprintf("Invalid version string format: %s", version),
		nil)
}

// String returns the string representation of the version
func (v PerlVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare compares this version with another version
// Returns: -1 if v < other, 0 if v == other, 1 if v > other
func (v PerlVersion) Compare(other PerlVersion) int {
	// Compare major version
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	// Compare minor version
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	// Compare patch version
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// All components are equal
	return 0
}

// IsStable returns true if the version is a stable release
func (v PerlVersion) IsStable() bool {
	return !v.Dev
}

// ParseConstraint parses a constraint string (e.g., ">=5.32.0")
func ParseConstraint(constraint string) (Constraint, error) {
	constraint = strings.TrimSpace(constraint)

	// Define pattern for constraint
	// Operators: >, <, >=, <=, ==, =, ~> (pessimistic version constraint, like ~> 5.32 means >= 5.32.0, < 5.33.0)
	pattern := regexp.MustCompile(`^([><]=?|==|=|~>)\s*(.+)$`)
	matches := pattern.FindStringSubmatch(constraint)

	if matches == nil {
		// If no operator is specified, assume "=" (exact match)
		version, err := ParseVersion(constraint)
		if err != nil {
			return Constraint{}, errors.NewVersionError(
				ErrInvalidConstraint,
				fmt.Sprintf("Invalid constraint string format: %s", constraint),
				err)
		}

		return Constraint{
			Operator: "=",
			Version:  version,
		}, nil
	}

	operator := matches[1]
	versionStr := matches[2]

	// Parse the version part
	version, err := ParseVersion(versionStr)
	if err != nil {
		return Constraint{}, errors.NewVersionError(
			ErrInvalidConstraint,
			fmt.Sprintf("Invalid version in constraint: %s", constraint),
			err)
	}

	return Constraint{
		Operator: operator,
		Version:  version,
	}, nil
}

// Satisfies checks if a version satisfies a constraint
func (c Constraint) Satisfies(version PerlVersion) bool {
	compare := version.Compare(c.Version)

	switch c.Operator {
	case ">":
		return compare > 0
	case ">=":
		return compare >= 0
	case "<":
		return compare < 0
	case "<=":
		return compare <= 0
	case "==", "=":
		return compare == 0
	case "~>":
		// Pessimistic version constraint, like ~> 5.32 means >= 5.32.0, < 5.33.0
		if compare < 0 {
			return false
		}

		// For ~> 5.32.1, we require < 5.33.0
		nextMinor := PerlVersion{
			Major: c.Version.Major,
			Minor: c.Version.Minor + 1,
			Patch: 0,
		}

		return version.Compare(nextMinor) < 0
	}

	// Unknown operator, always false
	return false
}

// String returns the string representation of the constraint
func (c Constraint) String() string {
	return fmt.Sprintf("%s %s", c.Operator, c.Version.String())
}

// ParseConstraintSet parses a comma-separated set of constraints
func ParseConstraintSet(constraints string) (ConstraintSet, error) {
	// Split constraints by comma
	constraintStrings := strings.Split(constraints, ",")
	result := make(ConstraintSet, 0, len(constraintStrings))

	for _, constraintStr := range constraintStrings {
		constraintStr = strings.TrimSpace(constraintStr)
		if constraintStr == "" {
			continue
		}

		constraint, err := ParseConstraint(constraintStr)
		if err != nil {
			return nil, err
		}

		result = append(result, constraint)
	}

	return result, nil
}

// Satisfies checks if a version satisfies all constraints in the set
func (cs ConstraintSet) Satisfies(version PerlVersion) bool {
	for _, constraint := range cs {
		if !constraint.Satisfies(version) {
			return false
		}
	}

	return true
}

// String returns the string representation of the constraint set
func (cs ConstraintSet) String() string {
	parts := make([]string, len(cs))
	for i, constraint := range cs {
		parts[i] = constraint.String()
	}

	return strings.Join(parts, ", ")
}

// ResolveVersionAlias resolves a version alias to an actual version
// Uses the provided map of aliases (from configuration)
func ResolveVersionAlias(alias string, aliases map[string]string) (string, error) {
	// If it's not an alias, return as is
	if !strings.HasPrefix(alias, "@") {
		return alias, nil
	}

	// Remove @ prefix
	aliasName := alias[1:]

	// Look up in the aliases map
	if version, ok := aliases[aliasName]; ok {
		return version, nil
	}

	// Special built-in aliases
	switch aliasName {
	case "stable", "latest-stable":
		return ResolveLatestStableVersion()
	case "dev", "latest-dev":
		return ResolveLatestDevVersion()
	case "latest":
		return ResolveLatestStableVersion() // Latest stable for install latest
	case "system":
		return GetSystemVersionString()
	}

	return "", errors.NewVersionError(
		ErrInvalidVersionAlias,
		fmt.Sprintf("Unknown version alias: %s", alias),
		nil)
}

// ResolveLatestStableVersion returns the latest stable Perl version
// Fetches from MetaCPAN API dynamically
func ResolveLatestStableVersion() (string, error) {
	return resolveLatestVersionWithDev(false)
}

// ResolveLatestDevVersion returns the latest development Perl version
func ResolveLatestDevVersion() (string, error) {
	return resolveLatestVersionWithDev(true)
}

// ResolveLatestVersion returns the latest Perl version (stable or dev)
func ResolveLatestVersion() (string, error) {
	return resolveLatestVersionWithDev(true)
}

// resolveLatestVersionWithDev fetches the latest version from MetaCPAN with dev support
func resolveLatestVersionWithDev(includeDev bool) (string, error) {
	// Create MetaCPAN provider
	provider, err := cpan.NewMetaCPANProvider()
	if err != nil {
		// Fallback to hardcoded values on error
		if includeDev {
			return "5.39.0", nil // Latest dev fallback
		}
		return "5.40.2", nil // Latest stable fallback
	}

	// Fetch versions with dev support
	ctx := context.Background()
	versions, err := provider.GetPerlCoreVersionsWithDev(ctx, includeDev)
	if err != nil {
		// Fallback to hardcoded values on error
		if includeDev {
			return "5.39.0", nil // Latest dev fallback
		}
		return "5.40.2", nil // Latest stable fallback
	}

	if len(versions) == 0 {
		// Fallback to hardcoded values if no versions found
		if includeDev {
			return "5.39.0", nil // Latest dev fallback
		}
		return "5.40.2", nil // Latest stable fallback
	}

	// Sort versions to find the latest
	sortedVersions := make([]PerlVersion, 0, len(versions))
	for _, versionStr := range versions {
		version, err := ParseVersion(versionStr)
		if err != nil {
			continue // Skip invalid versions
		}
		sortedVersions = append(sortedVersions, version)
	}

	if len(sortedVersions) == 0 {
		// Fallback to hardcoded values if no valid versions found
		if includeDev {
			return "5.39.0", nil // Latest dev fallback
		}
		return "5.40.2", nil // Latest stable fallback
	}

	// Sort to find the latest version
	sort.Slice(sortedVersions, func(i, j int) bool {
		return sortedVersions[i].Compare(sortedVersions[j]) > 0
	})

	// Filter by dev status if needed
	for _, version := range sortedVersions {
		if includeDev || !version.Dev {
			return version.String(), nil
		}
	}

	// Fallback to hardcoded values if no matching versions found
	if includeDev {
		return "5.39.0", nil // Latest dev fallback
	}
	return "5.40.2", nil // Latest stable fallback
}

// GetSystemVersionString returns the version of the system Perl as a string
func GetSystemVersionString() (string, error) {
	perl, err := DetectSystemPerl()
	if err != nil {
		return "", err
	}

	return perl.Version, nil
}

// ParseVersion is a variable that points to parseVersionFunc,
// allowing it to be replaced in tests
var ParseVersion = parseVersionFunc

// FindBestMatch finds the best version that satisfies the constraints
// Returns the highest matching version from the available versions
func FindBestMatch(constraint string, availableVersions []string) (string, error) {
	// Parse the constraint
	constraintSet, err := ParseConstraintSet(constraint)
	if err != nil {
		return "", err
	}

	// Track the highest matching version
	var bestMatch PerlVersion
	var bestMatchStr string
	bestMatchFound := false

	for _, versionStr := range availableVersions {
		version, err := ParseVersion(versionStr)
		if err != nil {
			// Skip invalid versions
			continue
		}

		if constraintSet.Satisfies(version) {
			if !bestMatchFound || version.Compare(bestMatch) > 0 {
				bestMatch = version
				bestMatchStr = versionStr
				bestMatchFound = true
			}
		}
	}

	if !bestMatchFound {
		return "", errors.NewVersionError(
			ErrUnsatisfiable,
			fmt.Sprintf("No available version satisfies constraints: %s", constraintSet.String()),
			nil)
	}

	return bestMatchStr, nil
}

// isVersionAvailableFunc checks if a specific version is available
func isVersionAvailableFunc(version string, availableVersions []string) bool {
	for _, availableVersion := range availableVersions {
		if version == availableVersion {
			return true
		}
	}

	return false
}

// IsVersionAvailable is a variable that points to isVersionAvailableFunc,
// allowing it to be replaced in tests
var IsVersionAvailable = isVersionAvailableFunc
