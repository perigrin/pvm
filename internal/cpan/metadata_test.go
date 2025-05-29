// ABOUTME: Tests for CPAN metadata retrieval
// ABOUTME: Verifies module information and search functionality

package cpan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetaCPANGetModuleInfo tests retrieving metadata for a module from MetaCPAN
func TestMetaCPANGetModuleInfo(t *testing.T) {
	// Create a test server that returns a mock module response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request  
		assert.Equal(t, "/module/Test::Module", r.URL.Path, "URL path should match the module name")
		assert.Equal(t, "GET", r.Method, "HTTP method should be GET")
		assert.Equal(t, defaultUserAgent, r.Header.Get("User-Agent"), "User-Agent header should be set")
		assert.Equal(t, "application/json", r.Header.Get("Accept"), "Accept header should be set")

		// Return a mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"name": "Test-Module.pm",
			"version": "1.0.0",
			"author": "TESTUSER",
			"abstract": "Test module for unit testing",
			"description": "This is a test module used for unit testing",
			"distribution": "Test-Module",
			"dist_version": "1.0.0",
			"download_url": "https://cpan.metacpan.org/authors/id/T/TE/TESTUSER/Test-Module-1.0.0.tar.gz",
			"date": "2023-01-01T00:00:00Z",
			"status": "latest",
			"module": [
				{
					"name": "Test::Module",
					"version": "1.0.0",
					"indexed": true
				}
			],
			"metadata": {
				"license": ["perl_5"]
			},
			"resources": {
				"repository": {
					"url": "https://github.com/testuser/Test-Module",
					"web": "https://github.com/testuser/Test-Module"
				},
				"bugtracker": {
					"web": "https://github.com/testuser/Test-Module/issues"
				},
				"homepage": "https://metacpan.org/pod/Test::Module"
			},
			"dependency": [
				{
					"phase": "runtime",
					"module": "Exporter",
					"requires": "5.57"
				},
				{
					"phase": "runtime",
					"module": "Carp",
					"requires": "0"
				}
			],
			"stats": {
				"pass": 50,
				"fail": 10,
				"na": 0,
				"unknown": 0
			},
			"stats_recent": {
				"downloads_30d": 1000
			}
		}`))
	}))
	defer server.Close()

	// Create a provider that uses the test server
	provider, err := NewMetaCPANProvider()
	require.NoError(t, err, "NewMetaCPANProvider should not return an error")
	require.NotNil(t, provider, "Provider should not be nil")

	// Directly set the baseURL on the provider to ensure it's used
	provider.baseURL = server.URL

	// Get module info
	moduleInfo, err := provider.GetModuleInfo(context.Background(), "Test::Module")
	require.NoError(t, err, "GetModuleInfo should not return an error")
	require.NotNil(t, moduleInfo, "ModuleInfo should not be nil")

	// Check module info fields
	assert.Equal(t, "Test::Module", moduleInfo.Name, "Module name should match")
	assert.Equal(t, "1.0.0", moduleInfo.Version, "Module version should match")
	assert.Equal(t, "TESTUSER", moduleInfo.Author, "Module author should match")
	assert.Equal(t, "Test module for unit testing", moduleInfo.Abstract, "Module abstract should match")
	assert.Equal(t, "This is a test module used for unit testing", moduleInfo.Description, "Module description should match")
	assert.Equal(t, "Test-Module", moduleInfo.Distribution, "Module distribution should match")
	assert.Equal(t, "1.0.0", moduleInfo.DistributionVersion, "Module distribution version should match")
	assert.Equal(t, "https://github.com/testuser/Test-Module", moduleInfo.Repository, "Module repository should match")
	assert.Equal(t, "https://metacpan.org/pod/Test::Module", moduleInfo.Homepage, "Module homepage should match")
	assert.Equal(t, "https://github.com/testuser/Test-Module/issues", moduleInfo.Bugtracker, "Module bugtracker should match")
	assert.Equal(t, "perl_5", moduleInfo.License, "Module license should match")
	assert.Equal(t, "https://metacpan.org/pod/Test::Module", moduleInfo.Documentation, "Module documentation should match")
	assert.Equal(t, 1000, moduleInfo.Downloads, "Module downloads should match")
	assert.InDelta(t, 4.17, moduleInfo.Rating, 0.01, "Module rating should match")
	assert.True(t, moduleInfo.HasTests, "Module should have tests")

	// Check release date
	expectedDate, err := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
	require.NoError(t, err, "Failed to parse expected date")
	assert.Equal(t, expectedDate, moduleInfo.ReleaseDate, "Module release date should match")

	// Check dependencies
	require.Len(t, moduleInfo.Dependencies, 2, "Module should have 2 dependencies")
	assert.Equal(t, "Exporter", moduleInfo.Dependencies[0].Name, "First dependency name should match")
	assert.Equal(t, "5.57", moduleInfo.Dependencies[0].Version, "First dependency version should match")
	assert.Equal(t, "runtime", moduleInfo.Dependencies[0].Phase, "First dependency phase should match")
	assert.Equal(t, "Carp", moduleInfo.Dependencies[1].Name, "Second dependency name should match")
}