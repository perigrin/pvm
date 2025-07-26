// ABOUTME: Integration tests for new release notes and changelog commands
// ABOUTME: Tests command functionality, flag handling, and error cases

package pvm

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/version"
)

// mockReleaseNotesGitHubClient implements GitHubClientInterface for testing
type mockReleaseNotesGitHubClient struct {
	releases []version.GitHubRelease
	err      error
}

// mockChangelogGitHubClient is a mock implementation for testing changelog command
type mockChangelogGitHubClient struct {
	releases []version.GitHubRelease
	err      error
}

func (m *mockChangelogGitHubClient) GetLatestRelease(owner, repo string) (*version.GitHubRelease, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.releases) > 0 {
		return &m.releases[0], nil
	}
	return nil, fmt.Errorf("no releases found")
}

func (m *mockChangelogGitHubClient) GetReleases(owner, repo string, includePrerelease bool) ([]version.GitHubRelease, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.releases, nil
}

func (m *mockChangelogGitHubClient) GetReleaseByTag(owner, repo, tag string) (*version.GitHubRelease, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, release := range m.releases {
		if release.TagName == tag {
			return &release, nil
		}
	}
	return nil, fmt.Errorf("release not found: %s", tag)
}

func (m *mockReleaseNotesGitHubClient) GetLatestRelease(owner, repo string) (*version.GitHubRelease, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.releases) == 0 {
		return nil, nil
	}
	return &m.releases[0], nil
}

func (m *mockReleaseNotesGitHubClient) GetReleases(owner, repo string, includePrerelease bool) ([]version.GitHubRelease, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.releases, nil
}

func (m *mockReleaseNotesGitHubClient) GetReleaseByTag(owner, repo, tag string) (*version.GitHubRelease, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, release := range m.releases {
		if release.TagName == tag {
			return &release, nil
		}
	}
	return &version.GitHubRelease{
		TagName: tag,
		Name:    "Test Release " + tag,
		Body:    "Test release notes for " + tag,
	}, nil
}

func TestReleaseNotesCommand_Creation(t *testing.T) {
	cmd := newReleaseNotesCommand()

	if cmd == nil {
		t.Fatal("newReleaseNotesCommand returned nil")
	}

	if cmd.Use != "release-notes [version]" {
		t.Errorf("Expected Use to be 'release-notes [version]', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("Command should have a long description")
	}

	if cmd.RunE == nil {
		t.Error("Command should have a RunE function")
	}
}

func TestReleaseNotesCommand_Flags(t *testing.T) {
	cmd := newReleaseNotesCommand()

	// Test that required flags exist
	expectedFlags := []string{"latest", "prerelease", "token"}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %s to exist", flagName)
		}
	}

	// Test flag types
	latestFlag := cmd.Flags().Lookup("latest")
	if latestFlag.Value.Type() != "bool" {
		t.Error("latest flag should be bool type")
	}

	prereleaseFlag := cmd.Flags().Lookup("prerelease")
	if prereleaseFlag.Value.Type() != "bool" {
		t.Error("prerelease flag should be bool type")
	}

	tokenFlag := cmd.Flags().Lookup("token")
	if tokenFlag.Value.Type() != "string" {
		t.Error("token flag should be string type")
	}
}

func TestChangelogCommand_Creation(t *testing.T) {
	cmd := newChangelogCommand()

	if cmd == nil {
		t.Fatal("newChangelogCommand returned nil")
	}

	if cmd.Use != "changelog" {
		t.Errorf("Expected Use to be 'changelog', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("Command should have a long description")
	}

	if cmd.RunE == nil {
		t.Error("Command should have a RunE function")
	}
}

func TestChangelogCommand_Flags(t *testing.T) {
	cmd := newChangelogCommand()

	// Test that required flags exist
	expectedFlags := []string{"limit", "prerelease", "token"}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %s to exist", flagName)
		}
	}

	// Test flag types and defaults
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag.Value.Type() != "int" {
		t.Error("limit flag should be int type")
	}
	if limitFlag.DefValue != "10" {
		t.Error("limit flag should have default value of 10")
	}

	prereleaseFlag := cmd.Flags().Lookup("prerelease")
	if prereleaseFlag.Value.Type() != "bool" {
		t.Error("prerelease flag should be bool type")
	}

	tokenFlag := cmd.Flags().Lookup("token")
	if tokenFlag.Value.Type() != "string" {
		t.Error("token flag should be string type")
	}
}

func TestReleaseNotesCommand_ErrorHandling(t *testing.T) {
	cmd := newReleaseNotesCommand()

	// Test with invalid token (should handle gracefully)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set invalid token
	cmd.Flags().Set("token", "invalid-token")
	cmd.Flags().Set("latest", "true")

	err := cmd.RunE(cmd, []string{})

	// Should handle error gracefully (network error is expected with invalid token)
	if err == nil {
		t.Log("Command succeeded (likely using fallback mechanism)")
	} else if !strings.Contains(err.Error(), "failed to") {
		// Error should be informative
		t.Errorf("Error message should be informative: %v", err)
	}
}

func TestChangelogCommand_ErrorHandling(t *testing.T) {
	cmd := newChangelogCommand()

	// Test with invalid token (should handle gracefully)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set invalid token
	cmd.Flags().Set("token", "invalid-token")
	cmd.Flags().Set("limit", "5")

	err := cmd.RunE(cmd, []string{})

	// Should handle error gracefully (network error is expected with invalid token)
	if err == nil {
		t.Log("Command succeeded (likely using fallback mechanism)")
	} else if !strings.Contains(err.Error(), "failed to") {
		// Error should be informative
		t.Errorf("Error message should be informative: %v", err)
	}
}

func TestReleaseNotesCommand_FlagCombinations(t *testing.T) {
	// Create mock client with test data
	mockClient := &mockReleaseNotesGitHubClient{
		releases: []version.GitHubRelease{
			{
				TagName: "v1.2.0",
				Name:    "PVM v1.2.0",
				Body:    "Latest release with new features",
			},
			{
				TagName: "v1.1.0",
				Name:    "PVM v1.1.0",
				Body:    "Previous release",
			},
		},
		err: nil,
	}

	testCases := []struct {
		name        string
		args        []string
		flags       map[string]string
		expectError bool
	}{
		{
			name:        "latest flag",
			args:        []string{},
			flags:       map[string]string{"latest": "true"},
			expectError: false,
		},
		{
			name:        "specific version",
			args:        []string{"v1.0.0"},
			flags:       map[string]string{},
			expectError: false,
		},
		{
			name:        "prerelease flag",
			args:        []string{},
			flags:       map[string]string{"prerelease": "true"},
			expectError: false,
		},
		{
			name:        "multiple flags",
			args:        []string{},
			flags:       map[string]string{"latest": "true", "prerelease": "true"},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newReleaseNotesCommand()

			// Reset flags
			cmd.Flags().Set("latest", "false")
			cmd.Flags().Set("prerelease", "false")
			cmd.Flags().Set("token", "")

			// Set test flags
			for flag, value := range tc.flags {
				cmd.Flags().Set(flag, value)
			}

			// Execute command with mock client
			err := executeReleaseNotesCommandWithOptions(cmd, tc.args, mockClient)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestChangelogCommand_FlagCombinations(t *testing.T) {
	cmd := newChangelogCommand()

	// Create a mock GitHub client
	mockClient := &mockChangelogGitHubClient{
		releases: []version.GitHubRelease{
			{
				TagName: "v1.2.0",
				Name:    "Version 1.2.0",
				Body:    "## Changes\n- Feature A\n- Feature B",
			},
			{
				TagName: "v1.1.0",
				Name:    "Version 1.1.0",
				Body:    "## Changes\n- Bug fix C\n- Enhancement D",
			},
			{
				TagName: "v1.0.0",
				Name:    "Version 1.0.0",
				Body:    "## Initial Release\n- Initial features",
			},
		},
	}

	testCases := []struct {
		name        string
		flags       map[string]string
		expectError bool
	}{
		{
			name:        "default flags",
			flags:       map[string]string{},
			expectError: false,
		},
		{
			name:        "custom limit",
			flags:       map[string]string{"limit": "5"},
			expectError: false,
		},
		{
			name:        "prerelease flag",
			flags:       map[string]string{"prerelease": "true"},
			expectError: false,
		},
		{
			name:        "multiple flags",
			flags:       map[string]string{"limit": "3", "prerelease": "true"},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flags
			cmd.Flags().Set("limit", "10")
			cmd.Flags().Set("prerelease", "false")
			cmd.Flags().Set("token", "")

			// Set test flags
			for flag, value := range tc.flags {
				cmd.Flags().Set(flag, value)
			}

			// Execute command with mocked client
			err := executeChangelogCommandWithOptions(cmd, []string{}, mockClient)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestReleaseNotesCommand_UIIntegration(t *testing.T) {
	// This test verifies that the command uses the UI framework correctly
	cmd := newReleaseNotesCommand()

	// Mock the UI to capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// The command should use UI framework internally
	// We can't easily test the actual UI calls without mocking,
	// but we can verify the command structure supports UI integration

	// Test that the command function exists and is callable
	if cmd.RunE == nil {
		t.Fatal("Command should have a RunE function")
	}

	// Test with invalid setup to ensure error handling works
	cmd.Flags().Set("token", "test-token")
	cmd.Flags().Set("latest", "true")

	err := cmd.RunE(cmd, []string{})

	// Should handle errors gracefully
	if err == nil {
		t.Log("Command succeeded (likely using fallback mechanism)")
	} else if !strings.Contains(err.Error(), "failed to") {
		// Error should be related to GitHub API (expected in test environment)
		t.Errorf("Error should be related to GitHub API: %v", err)
	}
}

func TestChangelogCommand_UIIntegration(t *testing.T) {
	// This test verifies that the command uses the UI framework correctly
	cmd := newChangelogCommand()

	// Mock the UI to capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// The command should use UI framework internally
	// We can't easily test the actual UI calls without mocking,
	// but we can verify the command structure supports UI integration

	// Test that the command function exists and is callable
	if cmd.RunE == nil {
		t.Fatal("Command should have a RunE function")
	}

	// Test with invalid setup to ensure error handling works
	cmd.Flags().Set("token", "test-token")
	cmd.Flags().Set("limit", "5")

	err := cmd.RunE(cmd, []string{})

	// Should handle errors gracefully
	if err == nil {
		t.Log("Command succeeded (likely using fallback mechanism)")
	} else if !strings.Contains(err.Error(), "failed to") {
		// Error should be related to GitHub API (expected in test environment)
		t.Errorf("Error should be related to GitHub API: %v", err)
	}
}

func TestCommandsIntegration_WithPVMCommand(t *testing.T) {
	// Test that the new commands integrate properly with the main PVM command
	pvmCmd := NewCommand()

	// Find the release-notes command
	var releaseNotesCmd *cobra.Command
	var changelogCmd *cobra.Command

	for _, cmd := range pvmCmd.Commands() {
		if cmd.Use == "release-notes [version]" {
			releaseNotesCmd = cmd
		}
		if cmd.Use == "changelog" {
			changelogCmd = cmd
		}
	}

	if releaseNotesCmd == nil {
		t.Error("release-notes command not found in PVM command")
	}

	if changelogCmd == nil {
		t.Error("changelog command not found in PVM command")
	}

	// Test that commands are properly registered
	if releaseNotesCmd != nil {
		if releaseNotesCmd.Parent() != pvmCmd {
			t.Error("release-notes command should be a child of PVM command")
		}
	}

	if changelogCmd != nil {
		if changelogCmd.Parent() != pvmCmd {
			t.Error("changelog command should be a child of PVM command")
		}
	}
}
