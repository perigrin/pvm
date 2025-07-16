// ABOUTME: Tests for MetaCPAN provider development version support functionality
// ABOUTME: Validates the GetPerlCoreVersionsWithDev method and dev version filtering

package cpan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaCPANProvider_GetPerlCoreVersionsWithDev(t *testing.T) {
	tests := []struct {
		name           string
		includeDev     bool
		expectedStable []string
		expectedDev    []string
		shouldError    bool
	}{
		{
			name:           "stable versions only",
			includeDev:     false,
			expectedStable: []string{"5.40.0", "5.38.2"},
			expectedDev:    []string{},
			shouldError:    false,
		},
		{
			name:           "include development versions",
			includeDev:     true,
			expectedStable: []string{"5.40.0", "5.38.2"},
			expectedDev:    []string{"5.39.0", "5.37.12"},
			shouldError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server with mock MetaCPAN response
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/release/_search", r.URL.Path)
				assert.Equal(t, "distribution:perl", r.URL.Query().Get("q"))
				assert.Equal(t, "GET", r.Method)

				// Mock response with both stable and dev versions
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
									"version": "5.039000",
									"date": "2024-03-15T12:00:00Z"
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
									"version": "5.037012",
									"date": "2023-12-01T12:00:00Z"
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

			// Test the functionality
			versions, err := provider.GetPerlCoreVersionsWithDev(context.Background(), tt.includeDev)

			if tt.shouldError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, versions)

			// Verify stable versions are included
			for _, expectedVersion := range tt.expectedStable {
				assert.Contains(t, versions, expectedVersion, "Expected stable version %s not found", expectedVersion)
			}

			// Verify dev versions are included/excluded based on includeDev flag
			if tt.includeDev {
				for _, expectedVersion := range tt.expectedDev {
					assert.Contains(t, versions, expectedVersion, "Expected dev version %s not found", expectedVersion)
				}
			} else {
				for _, devVersion := range tt.expectedDev {
					assert.NotContains(t, versions, devVersion, "Dev version %s should not be included", devVersion)
				}
			}
		})
	}
}

func TestMetaCPANProvider_GetPerlCoreVersionsWithDev_NetworkDisabled(t *testing.T) {
	// Create provider with network disabled
	provider, err := NewMetaCPANProvider()
	require.NoError(t, err)
	provider.disableNetwork = true

	// Test should return error when network is disabled
	versions, err := provider.GetPerlCoreVersionsWithDev(context.Background(), false)

	assert.Error(t, err)
	assert.Nil(t, versions)
	assert.Contains(t, err.Error(), "Network access is disabled")
}

func TestMetaCPANProvider_GetPerlCoreVersionsWithDev_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create provider with test server URL
	provider, err := NewMetaCPANProvider()
	require.NoError(t, err)
	provider.baseURL = server.URL

	// Test should return error on HTTP error
	versions, err := provider.GetPerlCoreVersionsWithDev(context.Background(), false)

	assert.Error(t, err)
	assert.Nil(t, versions)
	assert.Contains(t, err.Error(), "HTTP error")
}

func TestMetaCPANProvider_GetPerlCoreVersionsWithDev_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	// Create provider with test server URL
	provider, err := NewMetaCPANProvider()
	require.NoError(t, err)
	provider.baseURL = server.URL

	// Test should return error on invalid JSON
	versions, err := provider.GetPerlCoreVersionsWithDev(context.Background(), false)

	assert.Error(t, err)
	assert.Nil(t, versions)
	assert.Contains(t, err.Error(), "Failed to parse JSON response")
}

func TestMetaCPANProvider_GetPerlCoreVersionsWithDev_EmptyResponse(t *testing.T) {
	// Create a test server that returns empty response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hits": {"hits": []}}`))
	}))
	defer server.Close()

	// Create provider with test server URL
	provider, err := NewMetaCPANProvider()
	require.NoError(t, err)
	provider.baseURL = server.URL

	// Test should return empty slice (not error)
	versions, err := provider.GetPerlCoreVersionsWithDev(context.Background(), false)

	assert.NoError(t, err)
	assert.NotNil(t, versions)
	assert.Empty(t, versions)
}

func TestMetaCPANProvider_GetPerlCoreVersionsWithDev_BackwardCompatibility(t *testing.T) {
	// Test that GetPerlCoreVersions still works and returns only stable versions
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

	// Create provider with test server URL
	provider, err := NewMetaCPANProvider()
	require.NoError(t, err)
	provider.baseURL = server.URL

	// Test the original method
	versions, err := provider.GetPerlCoreVersions(context.Background())

	require.NoError(t, err)
	require.NotNil(t, versions)

	// Should only contain stable versions
	assert.Contains(t, versions, "5.40.0")
	assert.NotContains(t, versions, "5.39.0") // Dev version should be filtered out
}
