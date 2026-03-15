// ABOUTME: Tests for MetaCPAN provider RC (Release Candidate) filtering functionality
// ABOUTME: Validates that RC versions are properly excluded from stable version lists

package cpan

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaCPANProvider_GetPerlCoreVersions_ExcludesRC(t *testing.T) {
	// Test that release candidates are excluded from stable versions
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify we're requesting maturity and authorized fields
		assert.Contains(t, r.URL.Query().Get("fields"), "maturity")
		assert.Contains(t, r.URL.Query().Get("fields"), "authorized")

		mockResponse := `{
			"hits": {
				"hits": [
					{
						"fields": {
							"version": "5.040003",
							"date": "2025-07-21T20:16:11",
							"maturity": "developer",
							"authorized": true
						}
					},
					{
						"fields": {
							"version": "5.040002",
							"date": "2025-04-13T12:00:00",
							"maturity": "released",
							"authorized": true
						}
					},
					{
						"fields": {
							"version": "5.040001",
							"date": "2025-01-18T12:00:00",
							"maturity": "released",
							"authorized": true
						}
					},
					{
						"fields": {
							"version": "5.040000",
							"date": "2024-06-09T12:00:00",
							"maturity": "released",
							"authorized": true
						}
					},
					{
						"fields": {
							"version": "5.039010",
							"date": "2024-05-15T12:00:00",
							"maturity": "developer",
							"authorized": true
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

	// Create provider with test server URL
	provider, err := NewMetaCPANProvider()
	require.NoError(t, err)
	provider.baseURL = server.URL

	// Test stable versions only (should exclude RC/developer releases)
	versions, err := provider.GetPerlCoreVersions(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, versions)

	// Should include only released versions
	assert.Contains(t, versions, "5.40.2", "Should include 5.40.2 (released)")
	assert.Contains(t, versions, "5.40.1", "Should include 5.40.1 (released)")
	assert.Contains(t, versions, "5.40.0", "Should include 5.40.0 (released)")

	// Should exclude developer/RC versions
	assert.NotContains(t, versions, "5.40.3", "Should exclude 5.40.3 (developer/RC)")
	assert.NotContains(t, versions, "5.39.10", "Should exclude 5.39.10 (development version)")

	// When including dev versions, both should be present
	devVersions, err := provider.GetPerlCoreVersionsWithDev(context.Background(), true)
	assert.NoError(t, err)
	assert.Contains(t, devVersions, "5.40.3", "Dev mode should include 5.40.3 RC")
	assert.Contains(t, devVersions, "5.39.10", "Dev mode should include 5.39.10")
}

func TestMetaCPANProvider_GetPerlCoreVersions_MaturityValues(t *testing.T) {
	// Test various maturity values to ensure proper filtering
	testCases := []struct {
		name             string
		maturity         string
		version          string
		shouldBeIncluded bool
	}{
		{"released version", "released", "5.40.0", true},
		{"developer version", "developer", "5.40.3", false},
		{"empty maturity", "", "5.40.4", false}, // Conservative: exclude if not explicitly "released"
		{"unknown maturity", "unknown", "5.40.5", false},
		{"RC maturity", "rc", "5.40.6", false},
		{"testing maturity", "testing", "5.40.7", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mockResponse := `{
					"hits": {
						"hits": [
							{
								"fields": {
									"version": "` + convertToMetaCPANVersion(tc.version) + `",
									"date": "2024-06-09T12:00:00",
									"maturity": "` + tc.maturity + `",
									"authorized": true
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

			provider, err := NewMetaCPANProvider()
			require.NoError(t, err)
			provider.baseURL = server.URL

			versions, err := provider.GetPerlCoreVersions(context.Background())
			assert.NoError(t, err)

			if tc.shouldBeIncluded {
				assert.Contains(t, versions, tc.version, "%s with maturity '%s' should be included", tc.version, tc.maturity)
			} else {
				assert.NotContains(t, versions, tc.version, "%s with maturity '%s' should be excluded", tc.version, tc.maturity)
			}
		})
	}
}

func TestMetaCPANProvider_GetPerlCoreVersions_ExcludesUnauthorized(t *testing.T) {
	// Test that unauthorized releases are excluded completely
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := `{
			"hits": {
				"hits": [
					{
						"fields": {
							"version": "5.040002",
							"date": "2025-04-13T12:00:00",
							"maturity": "released",
							"authorized": true
						}
					},
					{
						"fields": {
							"version": "5.040001",
							"date": "2025-01-18T12:00:00",
							"maturity": "released",
							"authorized": false
						}
					},
					{
						"fields": {
							"version": "5.040000",
							"date": "2024-06-09T12:00:00",
							"maturity": "released",
							"authorized": true
						}
					},
					{
						"fields": {
							"version": "5.039000",
							"date": "2024-05-15T12:00:00",
							"maturity": "released",
							"authorized": false
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

	provider, err := NewMetaCPANProvider()
	require.NoError(t, err)
	provider.baseURL = server.URL

	// Test stable versions (should exclude unauthorized releases)
	versions, err := provider.GetPerlCoreVersions(context.Background())
	assert.NoError(t, err)

	// Should include only authorized releases
	assert.Contains(t, versions, "5.40.2", "Should include 5.40.2 (authorized)")
	assert.Contains(t, versions, "5.40.0", "Should include 5.40.0 (authorized)")

	// Should exclude unauthorized releases
	assert.NotContains(t, versions, "5.40.1", "Should exclude 5.40.1 (unauthorized)")
	assert.NotContains(t, versions, "5.39.0", "Should exclude 5.39.0 (unauthorized)")

	// Even with dev versions, unauthorized releases should be excluded
	devVersions, err := provider.GetPerlCoreVersionsWithDev(context.Background(), true)
	assert.NoError(t, err)
	assert.NotContains(t, devVersions, "5.40.1", "Dev mode should still exclude 5.40.1 (unauthorized)")
	assert.NotContains(t, devVersions, "5.39.0", "Dev mode should still exclude 5.39.0 (unauthorized)")
}

// Helper function to convert standard version to MetaCPAN format for testing
func convertToMetaCPANVersion(version string) string {
	// Simple conversion for test purposes
	// 5.40.3 -> 5.040003
	var major, minor, patch int
	if _, err := fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch); err == nil {
		return fmt.Sprintf("%d.%03d%03d", major, minor, patch)
	}
	return version
}
