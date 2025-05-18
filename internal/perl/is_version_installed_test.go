// ABOUTME: Tests for error paths in IsVersionInstalled function
// ABOUTME: Provides comprehensive error path testing for IsVersionInstalled

package perl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsVersionInstalledErrorPaths tests all error paths in IsVersionInstalled
func TestIsVersionInstalledErrorPaths(t *testing.T) {
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
		isInstalled, err := IsVersionInstalled("invalid-version")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid version format")
		assert.False(t, isInstalled)
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
		isInstalled, err := IsVersionInstalled("5.38.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock LoadRegistry error")
		assert.False(t, isInstalled)
	})
}
