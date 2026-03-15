// ABOUTME: Semantic version parsing and comparison utilities
// ABOUTME: Handles version string normalization and constraint matching

package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SemanticVersion represents a semantic version
type SemanticVersion struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
	Original   string
}

var (
	// semverRegex matches semantic version strings
	semverRegex = regexp.MustCompile(`^v?(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([0-9A-Za-z\-\.]+))?(?:\+([0-9A-Za-z\-\.]+))?$`)
)

// ParseVersion parses a version string into a SemanticVersion
func ParseVersion(versionStr string) (*SemanticVersion, error) {
	if versionStr == "" {
		return nil, fmt.Errorf("empty version string")
	}

	matches := semverRegex.FindStringSubmatch(versionStr)
	if matches == nil {
		return nil, fmt.Errorf("invalid semantic version: %s", versionStr)
	}

	var err error
	sv := &SemanticVersion{Original: versionStr}

	// Major version (required)
	if sv.Major, err = strconv.Atoi(matches[1]); err != nil {
		return nil, fmt.Errorf("invalid major version: %s", matches[1])
	}

	// Minor version (optional)
	if matches[2] != "" {
		if sv.Minor, err = strconv.Atoi(matches[2]); err != nil {
			return nil, fmt.Errorf("invalid minor version: %s", matches[2])
		}
	}

	// Patch version (optional)
	if matches[3] != "" {
		if sv.Patch, err = strconv.Atoi(matches[3]); err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", matches[3])
		}
	}

	// Prerelease (optional)
	if matches[4] != "" {
		sv.Prerelease = matches[4]
	}

	// Build metadata (optional)
	if matches[5] != "" {
		sv.Build = matches[5]
	}

	return sv, nil
}

// String returns the string representation of the version
func (sv *SemanticVersion) String() string {
	version := fmt.Sprintf("%d.%d.%d", sv.Major, sv.Minor, sv.Patch)

	if sv.Prerelease != "" {
		version += "-" + sv.Prerelease
	}

	if sv.Build != "" {
		version += "+" + sv.Build
	}

	return version
}

// Compare compares two semantic versions
// Returns: -1 if sv < other, 0 if sv == other, 1 if sv > other
func (sv *SemanticVersion) Compare(other *SemanticVersion) int {
	// Compare major version
	if sv.Major != other.Major {
		if sv.Major < other.Major {
			return -1
		}
		return 1
	}

	// Compare minor version
	if sv.Minor != other.Minor {
		if sv.Minor < other.Minor {
			return -1
		}
		return 1
	}

	// Compare patch version
	if sv.Patch != other.Patch {
		if sv.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// Compare prerelease versions
	return comparePrerelease(sv.Prerelease, other.Prerelease)
}

// comparePrerelease compares prerelease versions
func comparePrerelease(a, b string) int {
	// If both have no prerelease, they're equal
	if a == "" && b == "" {
		return 0
	}

	// Version without prerelease > version with prerelease
	if a == "" && b != "" {
		return 1
	}
	if a != "" && b == "" {
		return -1
	}

	// Both have prereleases - compare lexically
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aPart, bPart string

		if i < len(aParts) {
			aPart = aParts[i]
		}
		if i < len(bParts) {
			bPart = bParts[i]
		}

		// If one part is missing, the other is greater
		if aPart == "" && bPart != "" {
			return -1
		}
		if aPart != "" && bPart == "" {
			return 1
		}

		// Try to compare as numbers first
		aNum, aIsNum := parseIntSafe(aPart)
		bNum, bIsNum := parseIntSafe(bPart)

		if aIsNum && bIsNum {
			if aNum != bNum {
				if aNum < bNum {
					return -1
				}
				return 1
			}
		} else {
			// String comparison
			if aPart != bPart {
				if aPart < bPart {
					return -1
				}
				return 1
			}
		}
	}

	return 0
}

// parseIntSafe safely parses an integer, returning false if not an integer
func parseIntSafe(s string) (int, bool) {
	i, err := strconv.Atoi(s)
	return i, err == nil
}

// IsNewer returns true if sv is newer than other
func (sv *SemanticVersion) IsNewer(other *SemanticVersion) bool {
	return sv.Compare(other) > 0
}

// IsOlder returns true if sv is older than other
func (sv *SemanticVersion) IsOlder(other *SemanticVersion) bool {
	return sv.Compare(other) < 0
}

// IsEqual returns true if sv equals other
func (sv *SemanticVersion) IsEqual(other *SemanticVersion) bool {
	return sv.Compare(other) == 0
}

// IsPrerelease returns true if this is a prerelease version
func (sv *SemanticVersion) IsPrerelease() bool {
	return sv.Prerelease != ""
}

// NormalizeVersionString normalizes a version string (adds v prefix if missing)
func NormalizeVersionString(version string) string {
	if version == "" {
		return ""
	}

	if !strings.HasPrefix(version, "v") {
		return "v" + version
	}
	return version
}

// StripVersionPrefix removes the v prefix from a version string
func StripVersionPrefix(version string) string {
	return strings.TrimPrefix(version, "v")
}

// CompareVersionStrings compares two version strings
func CompareVersionStrings(a, b string) (int, error) {
	versionA, err := ParseVersion(a)
	if err != nil {
		return 0, fmt.Errorf("parsing version %s: %w", a, err)
	}

	versionB, err := ParseVersion(b)
	if err != nil {
		return 0, fmt.Errorf("parsing version %s: %w", b, err)
	}

	return versionA.Compare(versionB), nil
}

// IsVersionNewer returns true if version a is newer than version b
func IsVersionNewer(a, b string) (bool, error) {
	result, err := CompareVersionStrings(a, b)
	if err != nil {
		return false, err
	}
	return result > 0, nil
}
