// ABOUTME: Type definitions for version detection and comparison
// ABOUTME: Central location for version-related data structures

package version

import "time"

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	CurrentVersion *SemanticVersion
	LatestVersion  *SemanticVersion
	Release        *GitHubRelease
	UpdateNeeded   bool
	IsPrerelease   bool
}

// VersionCheckResult represents the result of a version check
type VersionCheckResult struct {
	CurrentVersion  string     `json:"current_version"`
	LatestVersion   string     `json:"latest_version"`
	UpdateAvailable bool       `json:"update_available"`
	IsPrerelease    bool       `json:"is_prerelease"`
	ReleaseURL      string     `json:"release_url,omitempty"`
	ReleaseNotes    string     `json:"release_notes,omitempty"`
	PublishedAt     *time.Time `json:"published_at,omitempty"`
	Error           string     `json:"error,omitempty"`
}

// CheckOptions configures version checking behavior
type CheckOptions struct {
	IncludePrerelease bool
	Repository        string // Format: "owner/repo"
	GitHubToken       string // Optional for higher rate limits
	Timeout           time.Duration
}

// DefaultCheckOptions returns default version check options
func DefaultCheckOptions() *CheckOptions {
	return &CheckOptions{
		IncludePrerelease: false,
		Repository:        "perigrin/pvm", // Default repository
		GitHubToken:       GitHubToken,    // Use build-time token if available
		Timeout:           30 * time.Second,
	}
}
