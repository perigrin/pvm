// ABOUTME: Tests for error paths in GetAvailableVersions function
// ABOUTME: Provides comprehensive error path testing for GetAvailableVersions

package perl

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// testTime is a fixed time for testing
var testTime = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

// TestGetAvailableVersions tests all paths in GetAvailableVersions
func TestGetAvailableVersions(t *testing.T) {
	// Save original functions to restore later
	originalGetInstalledVersions := GetInstalledVersions
	defer func() {
		GetInstalledVersions = originalGetInstalledVersions
	}()

	// Test 1: GetInstalledVersions fails
	t.Run("GetInstalledVersionsFails", func(t *testing.T) {
		// Mock GetInstalledVersions to fail
		GetInstalledVersions = func() ([]VersionInfo, error) {
			return nil, errors.New("mock GetInstalledVersions error")
		}

		// Call function - should return error
		versions, err := GetAvailableVersions()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock GetInstalledVersions error")
		assert.Nil(t, versions)
	})

	// Test 2: Success path - returns installed versions
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Mock GetInstalledVersions to return some test versions
		expectedVersions := []VersionInfo{
			{
				Version:     "5.38.0",
				InstallPath: "/opt/perl/5.38.0",
				InstallTime: testTime,
				Source:      "pvm",
			},
			{
				Version:     "5.36.0",
				InstallPath: "/opt/perl/5.36.0",
				InstallTime: testTime,
				Source:      "pvm",
			},
		}

		GetInstalledVersions = func() ([]VersionInfo, error) {
			return expectedVersions, nil
		}

		// Call function - should succeed and return all versions
		versions, err := GetAvailableVersions()
		assert.NoError(t, err)
		assert.Len(t, versions, 2)
		assert.Contains(t, versions, "5.38.0")
		assert.Contains(t, versions, "5.36.0")
	})
}
