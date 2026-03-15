// ABOUTME: Tests for error paths in UninstallVersion function
// ABOUTME: Provides comprehensive error path testing for UninstallVersion

package perl

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestUninstallVersionErrorPaths tests all error paths in UninstallVersion
func TestUninstallVersionErrorPaths(t *testing.T) {
	// Save original functions to restore later
	originalLoadRegistry := LoadRegistry
	originalSaveRegistry := SaveRegistry
	originalParseVersion := ParseVersion
	originalOsRemoveAll := osRemoveAll
	defer func() {
		LoadRegistry = originalLoadRegistry
		SaveRegistry = originalSaveRegistry
		ParseVersion = originalParseVersion
		osRemoveAll = originalOsRemoveAll
	}()

	// Test 1: ParseVersion fails
	t.Run("ParseVersionFails", func(t *testing.T) {
		// Mock ParseVersion to fail
		ParseVersion = func(version string) (PerlVersion, error) {
			return PerlVersion{}, errors.New("mock ParseVersion error")
		}

		// Call function - should return error
		err := UninstallVersion("invalid-version")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid version format")
	})

	// Test 2: LoadRegistry fails
	t.Run("LoadRegistryFails", func(t *testing.T) {
		// Mock ParseVersion to succeed
		ParseVersion = func(version string) (PerlVersion, error) {
			return PerlVersion{Major: 5, Minor: 38, Patch: 0}, nil
		}

		// Mock LoadRegistry to fail
		LoadRegistry = func() (*VersionRegistry, error) {
			return nil, errors.New("mock LoadRegistry error")
		}

		// Call function - should return error
		err := UninstallVersion("5.38.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock LoadRegistry error")
	})

	// Test 3: Version not found
	t.Run("VersionNotFound", func(t *testing.T) {
		// Mock ParseVersion to succeed
		ParseVersion = func(version string) (PerlVersion, error) {
			return PerlVersion{Major: 5, Minor: 38, Patch: 0}, nil
		}

		// Mock LoadRegistry to return empty registry
		LoadRegistry = func() (*VersionRegistry, error) {
			return &VersionRegistry{
				Versions: make(map[string]VersionInfo),
			}, nil
		}

		// Call function - should return error
		err := UninstallVersion("5.38.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not installed")
	})

	// Test 4: RemoveAll fails
	t.Run("RemoveAllFails", func(t *testing.T) {
		// Mock ParseVersion to succeed
		ParseVersion = func(version string) (PerlVersion, error) {
			return PerlVersion{Major: 5, Minor: 38, Patch: 0}, nil
		}

		// Mock LoadRegistry to return registry with our test version
		LoadRegistry = func() (*VersionRegistry, error) {
			return &VersionRegistry{
				Versions: map[string]VersionInfo{
					"5.38.0": {
						Version:     "5.38.0",
						InstallPath: "/opt/perl/5.38.0",
						InstallTime: time.Now(),
						Source:      "pvm", // Not "system", to test removal
					},
				},
			}, nil
		}

		// Mock SaveRegistry to succeed
		SaveRegistry = func(r *VersionRegistry) error {
			return nil
		}

		// Mock osRemoveAll to fail
		osRemoveAll = func(path string) error {
			return errors.New("mock RemoveAll error")
		}

		// Call function - should return error
		err := UninstallVersion("5.38.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to remove installation directory")
	})

	// Test 5: SaveRegistry fails
	t.Run("SaveRegistryFails", func(t *testing.T) {
		// Mock ParseVersion to succeed
		ParseVersion = func(version string) (PerlVersion, error) {
			return PerlVersion{Major: 5, Minor: 38, Patch: 0}, nil
		}

		// Mock LoadRegistry to return registry with our test version
		LoadRegistry = func() (*VersionRegistry, error) {
			return &VersionRegistry{
				Versions: map[string]VersionInfo{
					"5.38.0": {
						Version:     "5.38.0",
						InstallPath: "/opt/perl/5.38.0",
						InstallTime: time.Now(),
						Source:      "system", // Use "system" to skip RemoveAll
					},
				},
			}, nil
		}

		// Mock osRemoveAll to succeed (shouldn't be called anyway)
		osRemoveAll = func(path string) error {
			return nil
		}

		// Mock SaveRegistry to fail
		SaveRegistry = func(r *VersionRegistry) error {
			return errors.New("mock SaveRegistry error")
		}

		// Call function - should return error
		err := UninstallVersion("5.38.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock SaveRegistry error")
	})
}
