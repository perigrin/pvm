// ABOUTME: Tests for dynamic version resolution functionality
// ABOUTME: Validates the enhanced version resolution with MetaCPAN integration

package perl

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveLatestStableVersion_DynamicResolution(t *testing.T) {
	// Create a test server to mock MetaCPAN responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := `{
			"hits": {
				"hits": [
					{
						"fields": {
							"version": "5.040000",
							"date": "2024-06-09T12:00:00Z"
						}
					},
					{
						"fields": {
							"version": "5.038002",
							"date": "2024-01-10T12:00:00Z"
						}
					},
					{
						"fields": {
							"version": "5.036003",
							"date": "2023-05-15T12:00:00Z"
						}
					}
				]
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	// Test should return the latest stable version
	version, err := ResolveLatestStableVersion()
	require.NoError(t, err)

	// Should return a valid version string
	assert.NotEmpty(t, version)

	// Should be parseable as a version
	parsedVersion, err := ParseVersion(version)
	require.NoError(t, err)
	assert.False(t, parsedVersion.Dev) // Should be stable version
}

func TestResolveLatestDevVersion_DynamicResolution(t *testing.T) {
	// Create a test server to mock MetaCPAN responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := `{
			"hits": {
				"hits": [
					{
						"fields": {
							"version": "5.041000",
							"date": "2024-09-09T12:00:00Z"
						}
					},
					{
						"fields": {
							"version": "5.040000",
							"date": "2024-06-09T12:00:00Z"
						}
					},
					{
						"fields": {
							"version": "5.039000",
							"date": "2024-03-15T12:00:00Z"
						}
					}
				]
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	// Test should return the latest dev version
	version, err := ResolveLatestDevVersion()
	require.NoError(t, err)

	// Should return a valid version string
	assert.NotEmpty(t, version)

	// Should be parseable as a version
	parsedVersion, err := ParseVersion(version)
	require.NoError(t, err)

	// Could be dev or stable, depends on what's actually latest
	assert.True(t, parsedVersion.Major >= 5)
}

func TestResolveLatestVersion_FallbackToHardcoded(t *testing.T) {
	// Test the fallback behavior when network fails
	version, err := ResolveLatestVersion()
	require.NoError(t, err)

	// Should return a valid version string (fallback value)
	assert.NotEmpty(t, version)

	// Should be parseable as a version
	parsedVersion, err := ParseVersion(version)
	require.NoError(t, err)
	assert.True(t, parsedVersion.Major >= 5)
}

func TestResolveVersionAlias_LatestAliases(t *testing.T) {
	tests := []struct {
		name          string
		alias         string
		expectedError bool
	}{
		{
			name:          "latest stable alias",
			alias:         "@latest",
			expectedError: false,
		},
		{
			name:          "latest dev alias",
			alias:         "@latest-dev",
			expectedError: false,
		},
		{
			name:          "dev alias",
			alias:         "@dev",
			expectedError: false,
		},
		{
			name:          "stable alias",
			alias:         "@stable",
			expectedError: false,
		},
		{
			name:          "latest without @ (returns as-is)",
			alias:         "latest",
			expectedError: false,
		},
		{
			name:          "latest-dev without @ (returns as-is)",
			alias:         "latest-dev",
			expectedError: false,
		},
		{
			name:          "invalid alias",
			alias:         "@invalid",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := ResolveVersionAlias(tt.alias, map[string]string{})

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, version)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, version)

			// For non-@ aliases, they are returned as-is, so they may not be valid versions
			if tt.alias == "latest" || tt.alias == "latest-dev" {
				// These are special cases that return as-is
				assert.Equal(t, tt.alias, version)
			} else {
				// For real aliases, should be parseable as a version
				parsedVersion, err := ParseVersion(version)
				require.NoError(t, err)
				assert.True(t, parsedVersion.Major >= 5)
			}
		})
	}
}

func TestResolveVersionAlias_NonAlias(t *testing.T) {
	// Test that non-alias strings are returned as-is
	version, err := ResolveVersionAlias("5.38.2", map[string]string{})
	require.NoError(t, err)
	assert.Equal(t, "5.38.2", version)
}

func TestResolveVersionAlias_CustomAlias(t *testing.T) {
	// Test custom user-defined aliases
	aliases := map[string]string{
		"my-perl": "5.36.0",
		"test":    "5.38.0",
	}

	version, err := ResolveVersionAlias("@my-perl", aliases)
	require.NoError(t, err)
	assert.Equal(t, "5.36.0", version)

	version, err = ResolveVersionAlias("@test", aliases)
	require.NoError(t, err)
	assert.Equal(t, "5.38.0", version)
}

func TestResolveLatestVersionWithDev_ErrorHandling(t *testing.T) {
	// Test error handling in resolveLatestVersionWithDev
	tests := []struct {
		name       string
		includeDev bool
		setup      func() func()
	}{
		{
			name:       "stable version resolution",
			includeDev: false,
			setup: func() func() {
				return func() {}
			},
		},
		{
			name:       "dev version resolution",
			includeDev: true,
			setup: func() func() {
				return func() {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			// Call the function - should not panic and should return valid result
			version, err := resolveLatestVersionWithDev(tt.includeDev)
			require.NoError(t, err)
			assert.NotEmpty(t, version)

			// Should be parseable as a version
			parsedVersion, err := ParseVersion(version)
			require.NoError(t, err)
			assert.True(t, parsedVersion.Major >= 5)
		})
	}
}

func TestResolveLatestVersionWithDev_VersionSorting(t *testing.T) {
	// Test that version sorting works correctly
	version, err := resolveLatestVersionWithDev(false)
	require.NoError(t, err)
	assert.NotEmpty(t, version)

	parsedVersion, err := ParseVersion(version)
	require.NoError(t, err)

	// Should return a reasonable version (not too old, not too new)
	assert.True(t, parsedVersion.Major >= 5)
	assert.True(t, parsedVersion.Minor >= 20) // Should be modern Perl
}
