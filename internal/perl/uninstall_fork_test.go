// ABOUTME: Tests for fork-aware uninstall functionality in the registry
// ABOUTME: Verifies UninstallVersionByDisplayName and UUID-keyed registry uninstall

package perl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUninstallVersionByDisplayName_ForkVersion verifies that a fork version can be
// uninstalled by its display name (e.g. "mycompany/myfork-5.40.2").
func TestUninstallVersionByDisplayName_ForkVersion(t *testing.T) {
	originalLoadRegistry := LoadRegistry
	originalSaveRegistry := SaveRegistry
	originalOsRemoveAll := osRemoveAll
	defer func() {
		LoadRegistry = originalLoadRegistry
		SaveRegistry = originalSaveRegistry
		osRemoveAll = originalOsRemoveAll
	}()

	var savedRegistry *VersionRegistry
	removeAllCalled := false

	LoadRegistry = func() (*VersionRegistry, error) {
		return &VersionRegistry{
			Versions: map[string]VersionInfo{
				"uuid-fork-1": {
					Version:     "myfork-5.40.2",
					InstallPath: "/opt/perl/mycompany/myfork-5.40.2",
					InstallTime: time.Now(),
					Source:      "pvm",
					Remote:      "mycompany",
					ForkName:    "myfork",
					BaseVersion: "5.40.2",
				},
			},
		}, nil
	}

	SaveRegistry = func(r *VersionRegistry) error {
		savedRegistry = r
		return nil
	}

	osRemoveAll = func(path string) error {
		removeAllCalled = true
		assert.Equal(t, "/opt/perl/mycompany/myfork-5.40.2", path)
		return nil
	}

	err := UninstallVersionByDisplayName("mycompany/myfork-5.40.2")
	require.NoError(t, err)

	assert.True(t, removeAllCalled, "osRemoveAll should have been called")
	require.NotNil(t, savedRegistry)
	assert.Empty(t, savedRegistry.Versions, "registry should be empty after uninstall")
}

// TestUninstallVersionByDisplayName_StockVersion verifies that a stock version
// can be uninstalled by its bare version display name.
func TestUninstallVersionByDisplayName_StockVersion(t *testing.T) {
	originalLoadRegistry := LoadRegistry
	originalSaveRegistry := SaveRegistry
	originalOsRemoveAll := osRemoveAll
	defer func() {
		LoadRegistry = originalLoadRegistry
		SaveRegistry = originalSaveRegistry
		osRemoveAll = originalOsRemoveAll
	}()

	var savedRegistry *VersionRegistry

	LoadRegistry = func() (*VersionRegistry, error) {
		return &VersionRegistry{
			Versions: map[string]VersionInfo{
				"uuid-stock-1": {
					Version:     "5.40.2",
					InstallPath: "/opt/perl/5.40.2",
					InstallTime: time.Now(),
					Source:      "pvm",
				},
			},
		}, nil
	}

	SaveRegistry = func(r *VersionRegistry) error {
		savedRegistry = r
		return nil
	}

	osRemoveAll = func(path string) error {
		return nil
	}

	err := UninstallVersionByDisplayName("5.40.2")
	require.NoError(t, err)

	require.NotNil(t, savedRegistry)
	assert.Empty(t, savedRegistry.Versions)
}

// TestUninstallVersionByDisplayName_NotFound verifies an error is returned
// when no matching display name is found.
func TestUninstallVersionByDisplayName_NotFound(t *testing.T) {
	originalLoadRegistry := LoadRegistry
	defer func() {
		LoadRegistry = originalLoadRegistry
	}()

	LoadRegistry = func() (*VersionRegistry, error) {
		return &VersionRegistry{
			Versions: map[string]VersionInfo{},
		}, nil
	}

	err := UninstallVersionByDisplayName("mycompany/myfork-5.40.2")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

// TestUninstallVersionByDisplayName_SystemPerl verifies that system perl
// is only unregistered, not removed from disk.
func TestUninstallVersionByDisplayName_SystemPerl(t *testing.T) {
	originalLoadRegistry := LoadRegistry
	originalSaveRegistry := SaveRegistry
	originalOsRemoveAll := osRemoveAll
	defer func() {
		LoadRegistry = originalLoadRegistry
		SaveRegistry = originalSaveRegistry
		osRemoveAll = originalOsRemoveAll
	}()

	var savedRegistry *VersionRegistry
	removeAllCalled := false

	LoadRegistry = func() (*VersionRegistry, error) {
		return &VersionRegistry{
			Versions: map[string]VersionInfo{
				"uuid-sys-1": {
					Version:     "5.32.1",
					InstallPath: "/usr/bin",
					InstallTime: time.Now(),
					Source:      "system",
				},
			},
		}, nil
	}

	SaveRegistry = func(r *VersionRegistry) error {
		savedRegistry = r
		return nil
	}

	osRemoveAll = func(path string) error {
		removeAllCalled = true
		return nil
	}

	err := UninstallVersionByDisplayName("5.32.1")
	require.NoError(t, err)

	assert.False(t, removeAllCalled, "osRemoveAll should NOT be called for system perl")
	require.NotNil(t, savedRegistry)
	assert.Empty(t, savedRegistry.Versions)
}

// TestUninstallVersion_UUIDKeyedRegistry verifies that UninstallVersion works
// correctly with a UUID-keyed registry (the real registry format).
func TestUninstallVersion_UUIDKeyedRegistry(t *testing.T) {
	originalLoadRegistry := LoadRegistry
	originalSaveRegistry := SaveRegistry
	originalOsRemoveAll := osRemoveAll
	defer func() {
		LoadRegistry = originalLoadRegistry
		SaveRegistry = originalSaveRegistry
		osRemoveAll = originalOsRemoveAll
	}()

	var savedRegistry *VersionRegistry

	// Registry uses UUID keys (not version strings as keys)
	LoadRegistry = func() (*VersionRegistry, error) {
		return &VersionRegistry{
			Versions: map[string]VersionInfo{
				"some-random-uuid-1234": {
					Version:     "5.38.0",
					InstallPath: "/opt/perl/5.38.0",
					InstallTime: time.Now(),
					Source:      "pvm",
				},
			},
		}, nil
	}

	SaveRegistry = func(r *VersionRegistry) error {
		savedRegistry = r
		return nil
	}

	osRemoveAll = func(path string) error {
		return nil
	}

	err := UninstallVersion("5.38.0")
	require.NoError(t, err)

	require.NotNil(t, savedRegistry)
	assert.Empty(t, savedRegistry.Versions, "registry should be empty after uninstall")
}
