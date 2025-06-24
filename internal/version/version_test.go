// ABOUTME: Tests for main version checking and update detection functionality
// ABOUTME: Includes tests for CheckForUpdates and GetUpdateInfo functions

package version

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGetVersion(t *testing.T) {
	if GetVersion() != Version {
		t.Errorf("GetVersion() = %s, want %s", GetVersion(), Version)
	}
}

func TestComponentVersion(t *testing.T) {
	component := "PVM"
	expected := "PVM " + Version
	if got := ComponentVersion(component); got != expected {
		t.Errorf("ComponentVersion(%s) = %s, want %s", component, got, expected)
	}
}

func TestGetCurrentVersion(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() { Version = originalVersion }()

	tests := []struct {
		name        string
		version     string
		expectError bool
	}{
		{"valid version", "1.2.3", false},
		{"version with v prefix", "v1.2.3", false},
		{"prerelease version", "1.2.3-alpha.1", false},
		{"invalid version", "not.a.version", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version

			result, err := GetCurrentVersion()

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

			if result == nil {
				t.Error("expected version but got nil")
				return
			}

			if result.String() != StripVersionPrefix(tt.version) {
				t.Errorf("expected version %s, got %s", StripVersionPrefix(tt.version), result.String())
			}
		})
	}
}

func TestDefaultCheckOptions(t *testing.T) {
	opts := DefaultCheckOptions()

	if opts == nil {
		t.Fatal("expected options but got nil")
	}

	if opts.IncludePrerelease {
		t.Error("expected IncludePrerelease to be false")
	}

	if opts.Repository != "perigrin/pvm" {
		t.Errorf("expected Repository to be perigrin/pvm, got %s", opts.Repository)
	}

	if opts.Timeout != 30*time.Second {
		t.Errorf("expected Timeout to be 30s, got %v", opts.Timeout)
	}
}

func TestCheckForUpdates_InvalidRepository(t *testing.T) {
	opts := &CheckOptions{
		Repository: "invalid-repo-format",
	}

	result, err := CheckForUpdates(opts)

	if err == nil {
		t.Fatal("expected error for invalid repository format")
	}

	if result == nil {
		t.Error("expected result even with error")
	}

	if result != nil && result.Error == "" {
		t.Error("expected error field to be populated")
	}
}

func TestCheckForUpdates_InvalidCurrentVersion(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() { Version = originalVersion }()
	Version = "invalid.version"

	opts := &CheckOptions{
		Repository: "owner/repo",
	}

	result, err := CheckForUpdates(opts)

	if err == nil {
		t.Fatal("expected error for invalid current version")
	}

	if result == nil {
		t.Error("expected result even with error")
	}

	if result != nil && result.Error == "" {
		t.Error("expected error field to be populated")
	}
}

func TestVersionCheckResult_JSONSerialization(t *testing.T) {
	publishedAt := time.Now()
	result := &VersionCheckResult{
		CurrentVersion:  "1.0.0",
		LatestVersion:   "1.1.0",
		UpdateAvailable: true,
		IsPrerelease:    false,
		ReleaseURL:      "https://github.com/owner/repo/releases/tag/v1.1.0",
		ReleaseNotes:    "Release notes",
		PublishedAt:     &publishedAt,
		Error:           "",
	}

	// Test JSON marshaling
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	// Test JSON unmarshaling
	var decoded VersionCheckResult
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify fields
	if decoded.CurrentVersion != result.CurrentVersion {
		t.Errorf("CurrentVersion mismatch: expected %s, got %s", result.CurrentVersion, decoded.CurrentVersion)
	}

	if decoded.UpdateAvailable != result.UpdateAvailable {
		t.Errorf("UpdateAvailable mismatch: expected %t, got %t", result.UpdateAvailable, decoded.UpdateAvailable)
	}
}

func TestUpdateInfo_Structure(t *testing.T) {
	currentVer, err := ParseVersion("1.0.0")
	if err != nil {
		t.Fatalf("failed to parse current version: %v", err)
	}

	latestVer, err := ParseVersion("1.1.0")
	if err != nil {
		t.Fatalf("failed to parse latest version: %v", err)
	}

	release := &GitHubRelease{
		TagName: "v1.1.0",
		Name:    "Release 1.1.0",
	}

	updateInfo := &UpdateInfo{
		CurrentVersion: currentVer,
		LatestVersion:  latestVer,
		Release:        release,
		UpdateNeeded:   true,
		IsPrerelease:   false,
	}

	if updateInfo.CurrentVersion.String() != "1.0.0" {
		t.Errorf("expected current version 1.0.0, got %s", updateInfo.CurrentVersion.String())
	}

	if updateInfo.LatestVersion.String() != "1.1.0" {
		t.Errorf("expected latest version 1.1.0, got %s", updateInfo.LatestVersion.String())
	}

	if !updateInfo.UpdateNeeded {
		t.Error("expected update to be needed")
	}

	if updateInfo.Release.TagName != "v1.1.0" {
		t.Errorf("expected release tag v1.1.0, got %s", updateInfo.Release.TagName)
	}
}
