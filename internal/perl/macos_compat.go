// ABOUTME: macOS version detection and compatibility utilities
// ABOUTME: Provides functions to detect macOS versions and determine Configure script patching needs

package perl

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// MacOSVersion represents a macOS version
type MacOSVersion struct {
	Major int
	Minor int
	Patch int
}

// String returns the version as a string (e.g., "15.6.0")
func (v MacOSVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// IsNewerThan returns true if this version is newer than the other version
func (v MacOSVersion) IsNewerThan(other MacOSVersion) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

// GetMacOSVersion detects the current macOS version using sw_vers
func GetMacOSVersion() (*MacOSVersion, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("not running on macOS")
	}

	// Try sw_vers first
	cmd := exec.Command("sw_vers", "-productVersion")
	output, err := cmd.Output()
	if err != nil {
		// If sw_vers fails, try using system_profiler as fallback
		return getMacOSVersionFallback()
	}

	version := strings.TrimSpace(string(output))
	return ParseMacOSVersion(version)
}

// getMacOSVersionFallback uses system_profiler as a fallback method
func getMacOSVersionFallback() (*MacOSVersion, error) {
	cmd := exec.Command("system_profiler", "SPSoftwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to detect macOS version: %w", err)
	}

	// Look for "System Version:" line
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "System Version:") {
			// Extract version from line like "System Version: macOS 15.6 (24G5033e)"
			re := regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				return ParseMacOSVersion(matches[1])
			}
		}
	}

	return nil, fmt.Errorf("could not parse macOS version from system_profiler")
}

// ParseMacOSVersion parses a version string like "15.6.0" into a MacOSVersion
func ParseMacOSVersion(version string) (*MacOSVersion, error) {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid macOS version format: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	var patch int
	if len(parts) >= 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", parts[2])
		}
	}

	return &MacOSVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// NeedsConfigurePatch determines if a Perl version needs Configure script patching on macOS
func NeedsConfigurePatch(perlVersion string) bool {
	if runtime.GOOS != "darwin" {
		return false
	}

	// Get macOS version
	macOSVer, err := GetMacOSVersion()
	if err != nil {
		// If we can't detect macOS version, assume we need patching for safety
		return true
	}

	// Parse Perl version
	parsedPerlVersion, err := ParseVersion(perlVersion)
	if err != nil {
		// If we can't parse Perl version, assume we need patching for safety
		return true
	}

	// macOS Big Sur (11.x) and later require patching for Perl versions before 5.32.0
	if macOSVer.Major >= 11 {
		// Perl versions before 5.32.0 need patching
		if parsedPerlVersion.Major < 5 {
			return true
		}
		if parsedPerlVersion.Major == 5 && parsedPerlVersion.Minor < 32 {
			return true
		}
	}

	// macOS Sequoia (15.x) and later may need patching for even newer Perl versions
	// depending on when the Configure script was last updated
	if macOSVer.Major >= 15 {
		// Even some newer Perl versions might need patching for Sequoia
		if parsedPerlVersion.Major < 5 {
			return true
		}
		if parsedPerlVersion.Major == 5 && parsedPerlVersion.Minor < 34 {
			return true
		}
	}

	return false
}

// GetConfigurePatchStrategy returns the appropriate patching strategy
type ConfigurePatchStrategy int

const (
	PatchStrategyNone ConfigurePatchStrategy = iota
	PatchStrategyDarwinHints
	PatchStrategyConfigureScript
	PatchStrategyEnvironmentOverride
)

// GetConfigurePatchStrategy determines the best patching strategy for the given versions
func GetConfigurePatchStrategy(perlVersion string) ConfigurePatchStrategy {
	if !NeedsConfigurePatch(perlVersion) {
		return PatchStrategyNone
	}

	parsedPerlVersion, err := ParseVersion(perlVersion)
	if err != nil {
		// Default to hints patching if we can't parse version
		return PatchStrategyDarwinHints
	}

	// For very old Perl versions, try environment override first
	if parsedPerlVersion.Major < 5 || (parsedPerlVersion.Major == 5 && parsedPerlVersion.Minor < 20) {
		return PatchStrategyEnvironmentOverride
	}

	// For Perl 5.20-5.31, patch darwin hints file
	if parsedPerlVersion.Major == 5 && parsedPerlVersion.Minor < 32 {
		return PatchStrategyDarwinHints
	}

	// For newer versions that still need patching, patch Configure script directly
	return PatchStrategyConfigureScript
}

// GetMacOSCompatibilityInfo returns human-readable info about macOS compatibility
func GetMacOSCompatibilityInfo(perlVersion string) string {
	if runtime.GOOS != "darwin" {
		return ""
	}

	macOSVer, err := GetMacOSVersion()
	if err != nil {
		return "Warning: Could not detect macOS version"
	}

	needsPatch := NeedsConfigurePatch(perlVersion)
	strategy := GetConfigurePatchStrategy(perlVersion)

	var info strings.Builder
	info.WriteString(fmt.Sprintf("macOS Version: %s\n", macOSVer.String()))

	if !needsPatch {
		info.WriteString("Configure Patching: Not required\n")
		return info.String()
	}

	info.WriteString("Configure Patching: Required\n")
	switch strategy {
	case PatchStrategyDarwinHints:
		info.WriteString("Patch Strategy: Darwin hints file modification\n")
	case PatchStrategyConfigureScript:
		info.WriteString("Patch Strategy: Configure script modification\n")
	case PatchStrategyEnvironmentOverride:
		info.WriteString("Patch Strategy: Environment variable override\n")
	default:
		info.WriteString("Patch Strategy: Unknown\n")
	}

	return info.String()
}
