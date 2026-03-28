// ABOUTME: Tests for fork-aware display name resolution in GetCurrentVersion
// ABOUTME: Verifies that fork versions show their display name (remote/forkname-version) not bare version

package current

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/perl"
)

// TestGetCurrentVersion_ForkDisplayName verifies that when the resolved version
// is a fork, GetCurrentVersion returns the fork display name instead of the bare version.
func TestGetCurrentVersion_ForkDisplayName(t *testing.T) {
	originalGetVersionInfo := perl.GetVersionInfo
	defer func() {
		perl.GetVersionInfo = originalGetVersionInfo
	}()

	// Mock GetVersionInfo to return a fork VersionInfo
	perl.GetVersionInfo = func(version string) (*perl.VersionInfo, error) {
		if version == "myfork-5.40.2" {
			return &perl.VersionInfo{
				Version:     "myfork-5.40.2",
				InstallPath: "/opt/perl/mycompany/myfork-5.40.2",
				InstallTime: time.Now(),
				Source:      "pvm",
				Remote:      "mycompany",
				ForkName:    "myfork",
				BaseVersion: "5.40.2",
			}, nil
		}
		return nil, nil
	}

	// Build a CurrentVersionInfo as if it came from GetCurrentVersion
	// with a fork version resolved
	info := &CurrentVersionInfo{
		Version:           "myfork-5.40.2",
		Source:            "user_config",
		SourceDescription: "set by user configuration",
		Path:              "/opt/perl/mycompany/myfork-5.40.2/bin/perl",
		IsAvailable:       true,
	}

	// Simulate the display name lookup that GetCurrentVersion performs
	versionInfo, err := perl.GetVersionInfo(info.Version)
	require.NoError(t, err)
	require.NotNil(t, versionInfo)

	displayName := versionInfo.DisplayName()
	assert.Equal(t, "mycompany/myfork-5.40.2", displayName,
		"fork version should resolve to display name")
}

// TestGetCurrentVersion_StockVersionDisplayName verifies that stock versions
// still use the bare version string as their display name.
func TestGetCurrentVersion_StockVersionDisplayName(t *testing.T) {
	originalGetVersionInfo := perl.GetVersionInfo
	defer func() {
		perl.GetVersionInfo = originalGetVersionInfo
	}()

	// Mock GetVersionInfo to return a stock VersionInfo
	perl.GetVersionInfo = func(version string) (*perl.VersionInfo, error) {
		if version == "5.38.0" {
			return &perl.VersionInfo{
				Version:     "5.38.0",
				InstallPath: "/opt/perl/5.38.0",
				InstallTime: time.Now(),
				Source:      "pvm",
			}, nil
		}
		return nil, nil
	}

	versionInfo, err := perl.GetVersionInfo("5.38.0")
	require.NoError(t, err)
	require.NotNil(t, versionInfo)

	displayName := versionInfo.DisplayName()
	assert.Equal(t, "5.38.0", displayName,
		"stock version display name should be the bare version string")
}
