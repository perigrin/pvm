// ABOUTME: Tests for error paths in GetVersionInfo function
// ABOUTME: Provides comprehensive error path testing for GetVersionInfo

package perl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetVersionInfoErrorPaths tests all error paths in GetVersionInfo
func TestGetVersionInfoErrorPaths(t *testing.T) {
	// Save original functions to restore later
	originalLoadRegistry := LoadRegistry
	originalParseVersion := ParseVersion
	defer func() {
		LoadRegistry = originalLoadRegistry
		ParseVersion = originalParseVersion
	}()

	// Test 1: ParseVersion fails
	t.Run("ParseVersionFails", func(t *testing.T) {
		// Mock ParseVersion to fail
		ParseVersion = func(version string) (PerlVersion, error) {
			return PerlVersion{}, errors.New("mock ParseVersion error")
		}

		// Call function - should return error
		info, err := GetVersionInfo("5.38.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid version format")
		assert.Nil(t, info)
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
		info, err := GetVersionInfo("5.38.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock LoadRegistry error")
		assert.Nil(t, info)
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
		info, err := GetVersionInfo("5.38.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not installed")
		assert.Nil(t, info)
	})
}
