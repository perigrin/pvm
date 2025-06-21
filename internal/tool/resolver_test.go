// ABOUTME: Tests for CPAN resolver functionality and MetaCPAN integration
// ABOUTME: Includes tests for search strategies, caching, and fuzzy matching
package tool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCPANResolver(t *testing.T) {
	resolver := NewCPANResolver()

	assert.NotNil(t, resolver)
	assert.NotNil(t, resolver.client)
	assert.NotNil(t, resolver.cache)
	assert.NotNil(t, resolver.cacheTime)
	assert.Equal(t, "https://fastapi.metacpan.org/v1", resolver.baseURL)
	assert.Equal(t, 1*time.Hour, resolver.cacheTTL)
}

func TestSearchTool_EmptyToolName(t *testing.T) {
	resolver := NewCPANResolver()

	_, err := resolver.SearchTool("")
	assert.Error(t, err)

	var toolErr *ToolError
	require.ErrorAs(t, err, &toolErr)
	assert.Equal(t, ErrInvalidToolName, toolErr.Code)
}

func TestSearchTool_Cache(t *testing.T) {
	resolver := NewCPANResolver()

	// Create a test resolution
	testResolution := &ToolResolution{
		ToolName:   "test-tool",
		ModuleName: "App::TestTool",
		Source:     "cpan-test",
	}

	// Set in cache
	resolver.setCached("test-tool", testResolution)

	// Should return cached result
	result, err := resolver.SearchTool("test-tool")
	require.NoError(t, err)
	assert.Equal(t, testResolution.ToolName, result.ToolName)
	assert.Equal(t, testResolution.ModuleName, result.ModuleName)
	assert.Equal(t, testResolution.Source, result.Source)
}

func TestSearchTool_CacheExpiry(t *testing.T) {
	resolver := NewCPANResolver()
	resolver.cacheTTL = 10 * time.Millisecond // Very short TTL for testing

	// Create a test resolution
	testResolution := &ToolResolution{
		ToolName:   "test-tool",
		ModuleName: "App::TestTool",
		Source:     "cpan-test",
	}

	// Set in cache
	resolver.setCached("test-tool", testResolution)

	// Should return cached result immediately
	result := resolver.getCached("test-tool")
	assert.NotNil(t, result)

	// Wait for cache to expire
	time.Sleep(20 * time.Millisecond)

	// Should return nil after expiry
	result = resolver.getCached("test-tool")
	assert.Nil(t, result)
}

func TestCalculateSimilarity(t *testing.T) {
	resolver := NewCPANResolver()

	testCases := []struct {
		toolName   string
		moduleName string
		expected   float64
		threshold  float64
	}{
		{"ack", "App::Ack", 1.0, 0.01},          // Exact match (after App:: removal)
		{"ack", "App::Acknowledge", 0.27, 0.05}, // Partial match (calculated value)
		{"test", "App::Test", 1.0, 0.01},        // Exact match
		{"prove", "Test::Harness", 0.14, 0.05},  // Some similarity
		{"critic", "Perl::Critic", 1.0, 0.01},   // Exact match
		{"xyz", "App::Abc", 0.0, 0.01},          // No similarity
	}

	for _, tc := range testCases {
		t.Run(tc.toolName+"_vs_"+tc.moduleName, func(t *testing.T) {
			similarity := resolver.calculateSimilarity(tc.toolName, tc.moduleName)
			assert.InDelta(t, tc.expected, similarity, tc.threshold)
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	testCases := []struct {
		a        string
		b        string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"abc", "ac", 1},
		{"abc", "bc", 1},
		{"abc", "xyz", 3},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
	}

	for _, tc := range testCases {
		t.Run(tc.a+"_"+tc.b, func(t *testing.T) {
			distance := levenshteinDistance(tc.a, tc.b)
			assert.Equal(t, tc.expected, distance)
		})
	}
}

func TestMinMax(t *testing.T) {
	assert.Equal(t, 1, min(1, 2))
	assert.Equal(t, 1, min(2, 1))
	assert.Equal(t, 5, min(5, 5))

	assert.Equal(t, 2, max(1, 2))
	assert.Equal(t, 2, max(2, 1))
	assert.Equal(t, 5, max(5, 5))
}

func TestSetCachedGetCached(t *testing.T) {
	resolver := NewCPANResolver()

	testResolution := &ToolResolution{
		ToolName:   "test-tool",
		ModuleName: "App::TestTool",
		Source:     "cpan-test",
	}

	// Initially should return nil
	result := resolver.getCached("test-tool")
	assert.Nil(t, result)

	// Set in cache
	resolver.setCached("test-tool", testResolution)

	// Should now return the cached result
	result = resolver.getCached("test-tool")
	require.NotNil(t, result)
	assert.Equal(t, testResolution.ToolName, result.ToolName)
	assert.Equal(t, testResolution.ModuleName, result.ModuleName)
	assert.Equal(t, testResolution.Source, result.Source)

	// Cache time should be set
	cacheTime, exists := resolver.cacheTime["test-tool"]
	assert.True(t, exists)
	assert.True(t, time.Since(cacheTime) < time.Second)
}

// Note: The following tests would require a mock HTTP client or integration tests
// against the actual MetaCPAN API. For now, we're testing the core logic without
// external dependencies.

func TestSearchTool_NotFound(t *testing.T) {
	resolver := NewCPANResolver()

	// Override the search methods to return not found
	// This would require dependency injection or mocking for full testing
	_, err := resolver.SearchTool("definitely-nonexistent-tool-12345")
	assert.Error(t, err)

	var toolErr *ToolError
	require.ErrorAs(t, err, &toolErr)
	assert.Equal(t, ErrToolNotFound, toolErr.Code)
}

// Note: Full testing of the CPAN resolver would require either:
// 1. A test that uses actual MetaCPAN API (integration test)
// 2. HTTP client mocking with dependency injection
// For now, we test the core logic components separately
