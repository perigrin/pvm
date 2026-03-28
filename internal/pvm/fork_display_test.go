// ABOUTME: Tests for fork-aware display in versions and current commands
// ABOUTME: Verifies fork display names appear correctly in pvm versions and pvm current output

package pvm

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/perl"
)

// TestVersionsCommand_ShowsForkDisplayName verifies that pvm versions shows
// fork display names (e.g. "mycompany/myfork-5.40.2") and remote source in parentheses.
func TestVersionsCommand_ShowsForkDisplayName(t *testing.T) {
	cli.ResetGlobalState()
	originalGetInstalledVersions := perl.GetInstalledVersions
	defer func() {
		perl.GetInstalledVersions = originalGetInstalledVersions
		cli.ResetGlobalState()
	}()

	perl.GetInstalledVersions = func() ([]perl.VersionInfo, error) {
		return []perl.VersionInfo{
			{
				Version:     "myfork-5.40.2",
				InstallPath: "/opt/perl/mycompany/myfork-5.40.2",
				InstallTime: time.Now(),
				Source:      "pvm",
				Remote:      "mycompany",
				ForkName:    "myfork",
				BaseVersion: "5.40.2",
			},
			{
				Version:     "5.38.0",
				InstallPath: "/opt/perl/5.38.0",
				InstallTime: time.Now(),
				Source:      "pvm",
			},
		}, nil
	}

	cmd := newVersionsCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "mycompany/myfork-5.40.2", "fork display name should appear")
	assert.Contains(t, output, "(mycompany)", "remote source should appear in parentheses for fork")
	assert.Contains(t, output, "5.38.0", "stock version should appear")
}

// TestVersionsCommand_StockVersionNoRemoteAnnotation verifies that stock versions show
// no remote annotation while fork versions show their remote.
func TestVersionsCommand_StockVersionNoRemoteAnnotation(t *testing.T) {
	cli.ResetGlobalState()
	originalGetInstalledVersions := perl.GetInstalledVersions
	defer func() {
		perl.GetInstalledVersions = originalGetInstalledVersions
		cli.ResetGlobalState()
	}()

	perl.GetInstalledVersions = func() ([]perl.VersionInfo, error) {
		return []perl.VersionInfo{
			{
				Version:     "5.40.2",
				InstallPath: "/opt/perl/5.40.2",
				InstallTime: time.Now(),
				Source:      "pvm",
			},
		}, nil
	}

	cmd := newVersionsCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "5.40.2", "stock version should appear")
	assert.NotContains(t, output, "(origin)", "stock versions should not show (origin) annotation")
}

// TestUninstallCommand_AcceptsForkDisplayName verifies that the uninstall command
// accepts a fork display name (containing "/") and resolves it correctly.
func TestUninstallCommand_AcceptsForkDisplayName(t *testing.T) {
	originalFindByDisplayName := perl.FindByDisplayName
	originalUninstallVersionByDisplayName := perl.UninstallVersionByDisplayName
	defer func() {
		perl.FindByDisplayName = originalFindByDisplayName
		perl.UninstallVersionByDisplayName = originalUninstallVersionByDisplayName
	}()

	findCalled := false
	uninstallCalled := false

	perl.FindByDisplayName = func(name string) *perl.VersionInfo {
		findCalled = true
		assert.Equal(t, "mycompany/myfork-5.40.2", name)
		return &perl.VersionInfo{
			Version:     "myfork-5.40.2",
			InstallPath: "/opt/perl/mycompany/myfork-5.40.2",
			Source:      "pvm",
			Remote:      "mycompany",
			ForkName:    "myfork",
			BaseVersion: "5.40.2",
		}
	}

	perl.UninstallVersionByDisplayName = func(displayName string) error {
		uninstallCalled = true
		assert.Equal(t, "mycompany/myfork-5.40.2", displayName)
		return nil
	}

	cmd := newUninstallCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--force", "mycompany/myfork-5.40.2"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.True(t, findCalled, "FindByDisplayName should be called for fork identifiers")
	assert.True(t, uninstallCalled, "UninstallVersionByDisplayName should be called")
}

// TestUninstallCommand_StockVersionStillWorks verifies that uninstalling a plain
// version string (no "/") still works as before.
func TestUninstallCommand_StockVersionStillWorks(t *testing.T) {
	originalGetVersionInfo := perl.GetVersionInfo
	originalUninstallVersion := perl.UninstallVersion
	defer func() {
		perl.GetVersionInfo = originalGetVersionInfo
		perl.UninstallVersion = originalUninstallVersion
	}()

	getVersionInfoCalled := false
	uninstallCalled := false

	perl.GetVersionInfo = func(version string) (*perl.VersionInfo, error) {
		getVersionInfoCalled = true
		assert.Equal(t, "5.38.0", version)
		return &perl.VersionInfo{
			Version:     "5.38.0",
			InstallPath: "/opt/perl/5.38.0",
			Source:      "pvm",
		}, nil
	}

	perl.UninstallVersion = func(version string) error {
		uninstallCalled = true
		return nil
	}

	cmd := newUninstallCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--force", "5.38.0"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.True(t, getVersionInfoCalled, "GetVersionInfo should be called for stock versions")
	assert.True(t, uninstallCalled, "UninstallVersion should be called for stock versions")
}

// TestVersionsInfoDisplayName tests the DisplayName method for various VersionInfo configurations.
func TestVersionsInfoDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		info        perl.VersionInfo
		wantDisplay string
	}{
		{
			name:        "stock version has bare display name",
			info:        perl.VersionInfo{Version: "5.40.2", Source: "pvm"},
			wantDisplay: "5.40.2",
		},
		{
			name:        "origin remote has bare display name",
			info:        perl.VersionInfo{Version: "5.40.2", Remote: "origin", Source: "pvm"},
			wantDisplay: "5.40.2",
		},
		{
			name: "fork with name shows remote/forkname-version",
			info: perl.VersionInfo{
				Version: "myfork-5.40.2", Remote: "mycompany",
				ForkName: "myfork", BaseVersion: "5.40.2", Source: "pvm",
			},
			wantDisplay: "mycompany/myfork-5.40.2",
		},
		{
			name: "fork without name shows remote/version",
			info: perl.VersionInfo{
				Version: "5.40.2", Remote: "mycompany",
				ForkName: "", BaseVersion: "5.40.2", Source: "pvm",
			},
			wantDisplay: "mycompany/5.40.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantDisplay, tt.info.DisplayName())
		})
	}
}

// TestVersionsOutput_ForkRemoteAnnotation tests that the text output of pvm versions
// includes a remote annotation in parentheses for fork versions.
func TestVersionsOutput_ForkRemoteAnnotation(t *testing.T) {
	cli.ResetGlobalState()
	originalGetInstalledVersions := perl.GetInstalledVersions
	defer func() {
		perl.GetInstalledVersions = originalGetInstalledVersions
		cli.ResetGlobalState()
	}()

	perl.GetInstalledVersions = func() ([]perl.VersionInfo, error) {
		return []perl.VersionInfo{
			{
				Version:     "myfork-5.40.2",
				InstallPath: "/opt/perl/mycompany/myfork-5.40.2",
				InstallTime: time.Now(),
				Source:      "pvm",
				Remote:      "mycompany",
				ForkName:    "myfork",
				BaseVersion: "5.40.2",
			},
		}, nil
	}

	cmd := newVersionsCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// The fork version should appear with its display name and remote in parens
	lines := strings.Split(output, "\n")
	foundForkLine := false
	for _, line := range lines {
		if strings.Contains(line, "mycompany/myfork-5.40.2") {
			foundForkLine = true
			assert.Contains(t, line, "(mycompany)",
				"fork version line should include remote in parentheses")
		}
	}
	assert.True(t, foundForkLine, "should have a line with the fork display name")
}
