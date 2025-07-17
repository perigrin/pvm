// ABOUTME: Platform detection and binary selection for PVM updates
// ABOUTME: Maps current platform to GitHub release asset patterns for cross-platform updates

package version

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

// Platform represents the current system platform
type Platform struct {
	OS           string // Operating system (windows, darwin, linux)
	Architecture string // Architecture (amd64, arm64, etc.)
	Extension    string // Executable extension (.exe on Windows, empty elsewhere)
}

// debugLog prints debug messages only when PVM_DEBUG environment variable is set
func debugLog(format string, args ...interface{}) {
	if os.Getenv("PVM_DEBUG") != "" {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// DetectPlatform detects the current platform
func DetectPlatform() *Platform {
	platform := &Platform{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
	}

	// Add executable extension for Windows
	if platform.OS == "windows" {
		platform.Extension = ".exe"
	}

	debugLog("DetectPlatform: Detected platform %s", platform.String())
	return platform
}

// String returns a string representation of the platform
func (p *Platform) String() string {
	return fmt.Sprintf("%s-%s", p.OS, p.Architecture)
}

// AssetPattern returns the expected GitHub asset pattern for this platform
func (p *Platform) AssetPattern() string {
	// Convert GOOS/GOARCH to common GitHub release naming conventions
	os := p.OS
	arch := p.Architecture

	// Normalize OS names
	switch os {
	case "darwin":
		os = "darwin" // Keep as-is for macOS
	case "windows":
		os = "windows"
	case "linux":
		os = "linux"
	}

	// Normalize architecture names
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	case "386":
		arch = "386"
	}

	// Return pattern like "pvm-linux-amd64" or "pvm-windows-amd64.exe"
	pattern := fmt.Sprintf("pvm-%s-%s", os, arch)
	if p.Extension != "" {
		pattern += p.Extension
	}

	debugLog("AssetPattern: Generated pattern '%s' for platform %s", pattern, p.String())
	return pattern
}

// IsSupported returns true if this platform is supported for updates
func (p *Platform) IsSupported() bool {
	// Define supported platforms
	supportedCombinations := map[string][]string{
		"linux":   {"amd64", "arm64"},
		"darwin":  {"amd64", "arm64"},
		"windows": {"amd64"},
	}

	supportedArchs, osSupported := supportedCombinations[p.OS]
	if !osSupported {
		return false
	}

	for _, arch := range supportedArchs {
		if arch == p.Architecture {
			return true
		}
	}

	return false
}

// GetInstallationMethod detects how PVM was installed
func GetInstallationMethod() InstallationMethod {
	// Try to detect Homebrew installation
	if isHomebrewInstallation() {
		return InstallationHomebrew
	}

	// Check for other package managers (could be extended)
	// For now, assume binary installation if not Homebrew
	return InstallationBinary
}

// InstallationMethod represents how PVM was installed
type InstallationMethod int

const (
	InstallationBinary InstallationMethod = iota
	InstallationHomebrew
	InstallationAPT
	InstallationYum
	InstallationPacman
)

func (i InstallationMethod) String() string {
	switch i {
	case InstallationBinary:
		return "binary"
	case InstallationHomebrew:
		return "homebrew"
	case InstallationAPT:
		return "apt"
	case InstallationYum:
		return "yum"
	case InstallationPacman:
		return "pacman"
	default:
		return "unknown"
	}
}

// CanSelfUpdate returns true if PVM can self-update given the installation method
func (i InstallationMethod) CanSelfUpdate() bool {
	switch i {
	case InstallationBinary:
		return true
	case InstallationHomebrew:
		return false // Should use 'brew upgrade' instead
	default:
		return false // Package manager installations should use their package manager
	}
}

// isHomebrewInstallation tries to detect if PVM was installed via Homebrew
func isHomebrewInstallation() bool {
	// This is a simple heuristic - check if the executable path contains Homebrew paths
	// A more robust implementation could check for Homebrew metadata
	// For now, we'll use a simple path-based detection

	// Common Homebrew paths:
	// /usr/local/bin (Intel Macs)
	// /opt/homebrew/bin (Apple Silicon Macs)
	// /home/linuxbrew/.linuxbrew/bin (Linux)

	// This would need the actual executable path, which requires os.Executable()
	// For now, return false and implement this properly when needed
	return false
}

// FilterAssets filters GitHub release assets for the current platform
func FilterAssets(assets []GitHubAsset, platform *Platform) []GitHubAsset {
	pattern := platform.AssetPattern()
	var matches []GitHubAsset

	debugLog("FilterAssets: Looking for pattern '%s' in %d assets", pattern, len(assets))
	for i, asset := range assets {
		debugLog("FilterAssets: Asset %d: %s", i, asset.Name)
	}

	for _, asset := range assets {
		// Check if asset name matches our platform pattern using improved logic
		matched := isAssetMatch(asset.Name, pattern)
		debugLog("FilterAssets: '%s' matches pattern '%s': %t", asset.Name, pattern, matched)

		if matched {
			matches = append(matches, asset)
		}
	}

	debugLog("FilterAssets: Found %d matches for platform %s", len(matches), platform.String())
	return matches
}

// isAssetMatch checks if an asset name matches the platform pattern
// This handles versioned asset names like "pvm-1.0.0-rc30-darwin-arm64.tar.gz"
func isAssetMatch(assetName, pattern string) bool {
	// Extract the platform part from the pattern (everything after "pvm-")
	if !strings.HasPrefix(pattern, "pvm-") {
		return false
	}

	// Asset must also start with "pvm-"
	if !strings.HasPrefix(assetName, "pvm-") {
		return false
	}

	platformPart := pattern[4:] // Remove "pvm-" prefix

	// Try exact match first (for simple asset names like "pvm-darwin-arm64")
	if assetName == pattern {
		return true
	}

	// Handle versioned asset names
	// Pattern: pvm-darwin-arm64
	// Asset: pvm-1.0.0-rc30-darwin-arm64.tar.gz

	// Ensure the platform part is properly delimited
	// Look for pattern like "pvm-*-darwin-arm64" or "pvm-darwin-arm64"
	parts := strings.Split(assetName, "-")
	if len(parts) >= 3 {
		// Reconstruct platform from the last parts
		// For "pvm-1.0.0-rc30-darwin-arm64.tar.gz", parts would be:
		// ["pvm", "1.0.0", "rc30", "darwin", "arm64.tar.gz"]
		platformParts := strings.Split(platformPart, "-")
		if len(platformParts) == 2 { // e.g., ["darwin", "arm64"]
			os, arch := platformParts[0], platformParts[1]

			// Find the OS and check if arch follows
			for i, part := range parts {
				if part == os && i+1 < len(parts) {
					nextPart := parts[i+1]
					// Handle arch with extension (e.g., "arm64.tar.gz")
					if strings.HasPrefix(nextPart, arch) {
						return true
					}
				}
			}
		}
	}

	return false
}

// SelectBestAsset selects the best asset for the current platform
func SelectBestAsset(assets []GitHubAsset, platform *Platform) (*GitHubAsset, error) {
	filtered := FilterAssets(assets, platform)

	if len(filtered) == 0 {
		// Include available assets in error message for debugging
		var assetNames []string
		for _, asset := range assets {
			assetNames = append(assetNames, asset.Name)
		}
		return nil, fmt.Errorf("no assets found for platform %s (pattern: %s). Available assets: %v",
			platform.String(), platform.AssetPattern(), assetNames)
	}

	// For now, just return the first match
	// Could be enhanced to prefer certain naming conventions
	return &filtered[0], nil
}

// GetUpdateAsset gets the appropriate update asset from a GitHub release
func GetUpdateAsset(release *GitHubRelease, platform *Platform) (*GitHubAsset, error) {
	if release == nil {
		return nil, fmt.Errorf("release is nil")
	}

	if platform == nil {
		platform = DetectPlatform()
	}

	if !platform.IsSupported() {
		return nil, fmt.Errorf("platform %s is not supported for updates", platform.String())
	}

	asset, err := SelectBestAsset(release.Assets, platform)
	if err != nil {
		return nil, fmt.Errorf("selecting asset for platform %s: %w", platform.String(), err)
	}

	return asset, nil
}

// ValidateBinaryCompatibility performs basic validation of binary compatibility
func ValidateBinaryCompatibility(asset *GitHubAsset, platform *Platform) error {
	if asset == nil {
		return fmt.Errorf("asset is nil")
	}

	if platform == nil {
		return fmt.Errorf("platform is nil")
	}

	// Check if asset name matches expected pattern
	expected := platform.AssetPattern()
	if !strings.Contains(asset.Name, expected) {
		return fmt.Errorf("asset name %s doesn't match expected pattern %s", asset.Name, expected)
	}

	// Additional validations could be added here:
	// - File size checks
	// - Content-Type validation
	// - Digital signature verification

	return nil
}
