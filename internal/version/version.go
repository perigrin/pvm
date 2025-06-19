// ABOUTME: Contains version information for the PVM ecosystem
// ABOUTME: Used across all components to provide consistent versioning

package version

import (
	"fmt"
	"runtime"
	"strings"
)

// Version is the current version of the PVM ecosystem
// This can be overridden at build time with ldflags
var Version = "0.1.0"

// BuildTime is set at build time via ldflags
var BuildTime = "unknown"

// CommitHash is set at build time via ldflags
var CommitHash = "unknown"

// GetVersion returns the current version string
func GetVersion() string {
	return Version
}

// GetFullVersion returns version with build information
func GetFullVersion() string {
	return fmt.Sprintf("%s (built %s from %s)", Version, BuildTime, CommitHash)
}

// ComponentVersion returns a formatted version string for a specific component
func ComponentVersion(component string) string {
	return component + " " + Version
}

// GetBuildInfo returns detailed build information
func GetBuildInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"buildTime":  BuildTime,
		"commitHash": CommitHash,
		"goVersion":  runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	}
}

// GetVersionDisplay returns formatted version information as a string
// This is used by CLI commands to display version info with UI styling
func GetVersionDisplay(component string) string {
	return fmt.Sprintf("%s %s\nBuild Time: %s\nCommit: %s\nGo Version: %s\nOS/Arch: %s/%s",
		component, Version, BuildTime, CommitHash, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// GetCurrentVersion returns the current version as a SemanticVersion
func GetCurrentVersion() (*SemanticVersion, error) {
	return ParseVersion(Version)
}

// CheckForUpdates checks for available updates using the GitHub API
func CheckForUpdates(opts *CheckOptions) (*VersionCheckResult, error) {
	if opts == nil {
		opts = DefaultCheckOptions()
	}

	result := &VersionCheckResult{
		CurrentVersion: Version,
	}

	// Parse current version
	currentVer, err := ParseVersion(Version)
	if err != nil {
		result.Error = fmt.Sprintf("invalid current version: %v", err)
		return result, err
	}

	// Parse repository
	parts := strings.Split(opts.Repository, "/")
	if len(parts) != 2 {
		result.Error = "invalid repository format, expected owner/repo"
		return result, fmt.Errorf("invalid repository format: %s", opts.Repository)
	}
	owner, repo := parts[0], parts[1]

	// Create GitHub client
	var client *GitHubClient
	if opts.GitHubToken != "" {
		client = NewGitHubClientWithToken(opts.GitHubToken)
	} else {
		client = NewGitHubClient()
	}

	// Get latest release
	var release *GitHubRelease
	if opts.IncludePrerelease {
		releases, err := client.GetReleases(owner, repo, true)
		if err != nil {
			result.Error = fmt.Sprintf("failed to fetch releases: %v", err)
			return result, err
		}
		if len(releases) > 0 {
			release = &releases[0] // First release is the latest
		}
	} else {
		var err error
		release, err = client.GetLatestRelease(owner, repo)
		if err != nil {
			result.Error = fmt.Sprintf("failed to fetch latest release: %v", err)
			return result, err
		}
	}

	if release == nil {
		result.Error = "no releases found"
		return result, fmt.Errorf("no releases found")
	}

	// Parse latest version
	latestVer, err := ParseVersion(release.TagName)
	if err != nil {
		result.Error = fmt.Sprintf("invalid latest version: %v", err)
		return result, err
	}

	// Fill in result
	result.LatestVersion = release.TagName
	result.UpdateAvailable = latestVer.IsNewer(currentVer)
	result.IsPrerelease = latestVer.IsPrerelease()
	result.ReleaseURL = release.HTMLURL
	result.ReleaseNotes = release.Body
	result.PublishedAt = release.PublishedAt

	return result, nil
}

// GetUpdateInfo returns structured update information
func GetUpdateInfo(opts *CheckOptions) (*UpdateInfo, error) {
	if opts == nil {
		opts = DefaultCheckOptions()
	}

	// Get current version
	currentVer, err := GetCurrentVersion()
	if err != nil {
		return nil, fmt.Errorf("getting current version: %w", err)
	}

	// Check for updates
	result, err := CheckForUpdates(opts)
	if err != nil {
		return nil, fmt.Errorf("checking for updates: %w", err)
	}

	// Parse latest version
	latestVer, err := ParseVersion(result.LatestVersion)
	if err != nil {
		return nil, fmt.Errorf("parsing latest version: %w", err)
	}

	// Parse repository to get release info
	parts := strings.Split(opts.Repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s", opts.Repository)
	}
	owner, repo := parts[0], parts[1]

	// Create GitHub client
	var client *GitHubClient
	if opts.GitHubToken != "" {
		client = NewGitHubClientWithToken(opts.GitHubToken)
	} else {
		client = NewGitHubClient()
	}

	// Get release details
	release, err := client.GetReleaseByTag(owner, repo, result.LatestVersion)
	if err != nil {
		return nil, fmt.Errorf("getting release details: %w", err)
	}

	return &UpdateInfo{
		CurrentVersion: currentVer,
		LatestVersion:  latestVer,
		Release:        release,
		UpdateNeeded:   result.UpdateAvailable,
		IsPrerelease:   result.IsPrerelease,
	}, nil
}
