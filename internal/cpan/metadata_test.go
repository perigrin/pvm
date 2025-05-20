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
		assert.Equal(t, "/module/Test/Module", r.URL.Path, "URL path should match the module name")
		assert.Equal(t, "GET", r.Method, "HTTP method should be GET")
		assert.Equal(t, defaultUserAgent, r.Header.Get("User-Agent"), "User-Agent header should be set")
		assert.Equal(t, "application/json", r.Header.Get("Accept"), "Accept header should be set")

		// Return a mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"name": "Test::Module",
			"version": "1.0.0",
			"author": "TESTUSER",
			"abstract": "Test module for unit testing",
			"description": "This is a test module used for unit testing",
			"distribution": "Test-Module",
			"dist_version": "1.0.0",
			"download_url": "https://cpan.metacpan.org/authors/id/T/TE/TESTUSER/Test-Module-1.0.0.tar.gz",
			"date": "2023-01-01T00:00:00Z",
			"status": "latest",
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
	assert.Equal(t, "0", moduleInfo.Dependencies[1].Version, "Second dependency version should match")
	assert.Equal(t, "runtime", moduleInfo.Dependencies[1].Phase, "Second dependency phase should match")
}

// TestMetaCPANSearchModules tests searching for modules using MetaCPAN
func TestMetaCPANSearchModules(t *testing.T) {
	// Create a test server that returns a mock search response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request
		assert.Equal(t, "/search/module", r.URL.Path, "URL path should be correct")
		assert.Equal(t, "test", r.URL.Query().Get("q"), "Query parameter should be set")
		assert.Equal(t, "20", r.URL.Query().Get("size"), "Size parameter should be set")
		assert.Equal(t, "GET", r.Method, "HTTP method should be GET")
		assert.Equal(t, defaultUserAgent, r.Header.Get("User-Agent"), "User-Agent header should be set")
		assert.Equal(t, "application/json", r.Header.Get("Accept"), "Accept header should be set")

		// Return a mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"total": 2,
			"hits": [
				{
					"_score": 1.5,
					"_source": {
						"name": "Test::Module1",
						"version": "1.0.0",
						"distribution": "Test-Module1",
						"author": "TESTUSER",
						"abstract": "First test module",
						"date": "2023-01-01T00:00:00Z",
						"status": "latest",
						"documentation": "https://metacpan.org/pod/Test::Module1"
					}
				},
				{
					"_score": 1.2,
					"_source": {
						"name": "Test::Module2",
						"version": "2.0.0",
						"distribution": "Test-Module2",
						"author": "TESTUSER",
						"abstract": "Second test module",
						"date": "2022-01-01T00:00:00Z",
						"status": "latest",
						"documentation": "https://metacpan.org/pod/Test::Module2"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	// Create a provider that uses the test server
	provider, err := NewMetaCPANProvider()
	require.NoError(t, err, "NewMetaCPANProvider should not return an error")
	require.NotNil(t, provider, "Provider should not be nil")

	// Directly set the baseURL on the provider to ensure it's used
	provider.baseURL = server.URL

	// Search for modules
	results, err := provider.SearchModules(context.Background(), "test", 20)
	require.NoError(t, err, "SearchModules should not return an error")
	require.NotNil(t, results, "Results should not be nil")

	// Check search results
	assert.Equal(t, "test", results.Query, "Query should match")
	assert.Equal(t, 2, results.Total, "Total should match")
	assert.Equal(t, "metacpan", results.Source, "Source should match")
	require.Len(t, results.Results, 2, "Should have 2 results")

	// Check first result
	assert.Equal(t, "Test::Module1", results.Results[0].Name, "First result name should match")
	assert.Equal(t, "1.0.0", results.Results[0].Version, "First result version should match")
	assert.Equal(t, "Test-Module1", results.Results[0].Distribution, "First result distribution should match")
	assert.Equal(t, "TESTUSER", results.Results[0].Author, "First result author should match")
	assert.Equal(t, "First test module", results.Results[0].Abstract, "First result abstract should match")
	assert.Equal(t, 1.5, results.Results[0].Score, "First result score should match")
	assert.True(t, results.Results[0].IsLatest, "First result should be latest")
	assert.True(t, results.Results[0].HasDocumentation, "First result should have documentation")

	// Check first result release date
	expectedDate1, err := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
	require.NoError(t, err, "Failed to parse expected date")
	assert.Equal(t, expectedDate1, results.Results[0].ReleaseDate, "First result release date should match")

	// Check second result
	assert.Equal(t, "Test::Module2", results.Results[1].Name, "Second result name should match")
	assert.Equal(t, "2.0.0", results.Results[1].Version, "Second result version should match")
	assert.Equal(t, "Test-Module2", results.Results[1].Distribution, "Second result distribution should match")
	assert.Equal(t, "TESTUSER", results.Results[1].Author, "Second result author should match")
	assert.Equal(t, "Second test module", results.Results[1].Abstract, "Second result abstract should match")
	assert.Equal(t, 1.2, results.Results[1].Score, "Second result score should match")
	assert.True(t, results.Results[1].IsLatest, "Second result should be latest")
	assert.True(t, results.Results[1].HasDocumentation, "Second result should have documentation")

	// Check second result release date
	expectedDate2, err := time.Parse(time.RFC3339, "2022-01-01T00:00:00Z")
	require.NoError(t, err, "Failed to parse expected date")
	assert.Equal(t, expectedDate2, results.Results[1].ReleaseDate, "Second result release date should match")
}

// TestDisabledNetwork tests that operations fail when network is disabled
func TestDisabledNetwork(t *testing.T) {
	// Create a provider with network disabled
	provider, err := NewMetaCPANProvider()
	require.NoError(t, err, "NewMetaCPANProvider should not return an error")
	require.NotNil(t, provider, "Provider should not be nil")

	// Directly set the disableNetwork flag to true
	provider.disableNetwork = true

	// Try to get module info
	moduleInfo, err := provider.GetModuleInfo(context.Background(), "Test::Module")
	assert.Error(t, err, "GetModuleInfo should return an error when network is disabled")
	assert.Nil(t, moduleInfo, "ModuleInfo should be nil when network is disabled")

	// Try to search for modules
	results, err := provider.SearchModules(context.Background(), "test", 20)
	assert.Error(t, err, "SearchModules should return an error when network is disabled")
	assert.Nil(t, results, "Results should be nil when network is disabled")
}
