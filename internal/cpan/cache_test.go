// ABOUTME: Tests for CPAN metadata caching
// ABOUTME: Validates disk-based caching system behavior

package cpan

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheCreateAndGet tests creating and retrieving a cached item
func TestCacheCreateAndGet(t *testing.T) {
	// Create a temporary directory for the cache
	tempDir, err := os.MkdirTemp("", "cpan-cache-test")
	require.NoError(t, err, "Failed to create temporary directory")
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a cache with a 1-hour TTL
	cache, err := NewCache(tempDir, 1)
	require.NoError(t, err, "NewCache should not return an error")
	require.NotNil(t, cache, "Cache should not be nil")

	// Data to cache
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	// Set a cache entry
	err = cache.Set("test-key", testData, "test-source")
	require.NoError(t, err, "Cache.Set should not return an error")

	// Check that the cache file exists
	cachePath := cache.getCachePath("test-key")
	_, err = os.Stat(cachePath)
	require.NoError(t, err, "Cache file should exist")

	// Get the cached data
	var resultData map[string]string
	found := cache.Get("test-key", &resultData)
	require.True(t, found, "Cache.Get should return true for an existing key")
	assert.Equal(t, testData, resultData, "Retrieved data should match the original data")
}

// TestCacheExpiry tests that expired cache entries are not returned
func TestCacheExpiry(t *testing.T) {
	// Create a temporary directory for the cache
	tempDir, err := os.MkdirTemp("", "cpan-cache-test")
	require.NoError(t, err, "Failed to create temporary directory")
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a cache with a 0-hour TTL (immediately expired)
	cache, err := NewCache(tempDir, 0)
	require.NoError(t, err, "NewCache should not return an error")
	require.NotNil(t, cache, "Cache should not be nil")

	// Data to cache
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	// Set a cache entry
	err = cache.Set("test-key", testData, "test-source")
	require.NoError(t, err, "Cache.Set should not return an error")

	// Get the cached data
	var resultData map[string]string
	found := cache.Get("test-key", &resultData)
	assert.False(t, found, "Cache.Get should return false for an expired key")
}

// TestCacheOverwrite tests overwriting an existing cache entry
func TestCacheOverwrite(t *testing.T) {
	// Create a temporary directory for the cache
	tempDir, err := os.MkdirTemp("", "cpan-cache-test")
	require.NoError(t, err, "Failed to create temporary directory")
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a cache with a 1-hour TTL
	cache, err := NewCache(tempDir, 1)
	require.NoError(t, err, "NewCache should not return an error")
	require.NotNil(t, cache, "Cache should not be nil")

	// Set initial cache entry
	initialData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	err = cache.Set("test-key", initialData, "test-source")
	require.NoError(t, err, "Cache.Set should not return an error")

	// Overwrite the cache entry
	updatedData := map[string]string{
		"key1": "updated-value1",
		"key3": "value3",
	}
	err = cache.Set("test-key", updatedData, "test-source")
	require.NoError(t, err, "Cache.Set should not return an error")

	// Get the cached data
	var resultData map[string]string
	found := cache.Get("test-key", &resultData)
	require.True(t, found, "Cache.Get should return true for an existing key")
	assert.Equal(t, updatedData, resultData, "Retrieved data should match the updated data")
}

// TestCacheDelete tests deleting a cache entry
func TestCacheDelete(t *testing.T) {
	// Create a temporary directory for the cache
	tempDir, err := os.MkdirTemp("", "cpan-cache-test")
	require.NoError(t, err, "Failed to create temporary directory")
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a cache with a 1-hour TTL
	cache, err := NewCache(tempDir, 1)
	require.NoError(t, err, "NewCache should not return an error")
	require.NotNil(t, cache, "Cache should not be nil")

	// Set a cache entry
	testData := map[string]string{"key": "value"}
	err = cache.Set("test-key", testData, "test-source")
	require.NoError(t, err, "Cache.Set should not return an error")

	// Delete the cache entry
	err = cache.Delete("test-key")
	require.NoError(t, err, "Cache.Delete should not return an error")

	// Get the cached data
	var resultData map[string]string
	found := cache.Get("test-key", &resultData)
	assert.False(t, found, "Cache.Get should return false for a deleted key")
}

// TestCacheClear tests clearing the entire cache
func TestCacheClear(t *testing.T) {
	// Create a temporary directory for the cache
	tempDir, err := os.MkdirTemp("", "cpan-cache-test")
	require.NoError(t, err, "Failed to create temporary directory")
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a cache with a 1-hour TTL
	cache, err := NewCache(tempDir, 1)
	require.NoError(t, err, "NewCache should not return an error")
	require.NotNil(t, cache, "Cache should not be nil")

	// Set multiple cache entries
	testData1 := map[string]string{"key1": "value1"}
	testData2 := map[string]string{"key2": "value2"}
	err = cache.Set("test-key1", testData1, "test-source")
	require.NoError(t, err, "Cache.Set should not return an error")
	err = cache.Set("test-key2", testData2, "test-source")
	require.NoError(t, err, "Cache.Set should not return an error")

	// Clear the cache
	err = cache.Clear()
	require.NoError(t, err, "Cache.Clear should not return an error")

	// Get the cached data
	var resultData1 map[string]string
	var resultData2 map[string]string
	found1 := cache.Get("test-key1", &resultData1)
	found2 := cache.Get("test-key2", &resultData2)
	assert.False(t, found1, "Cache.Get should return false for a cleared key")
	assert.False(t, found2, "Cache.Get should return false for a cleared key")
}

// TestCacheCleanup tests cleaning up expired cache entries
func TestCacheCleanup(t *testing.T) {
	// Create a temporary directory for the cache
	tempDir, err := os.MkdirTemp("", "cpan-cache-test")
	require.NoError(t, err, "Failed to create temporary directory")
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a cache with a 1-hour TTL
	cache, err := NewCache(tempDir, 1)
	require.NoError(t, err, "NewCache should not return an error")
	require.NotNil(t, cache, "Cache should not be nil")

	// Set a cache entry that expires immediately
	zeroTTLCache := &Cache{
		cacheDir: tempDir,
		ttl:      0,
		locks:    make(map[string]*sync.Mutex),
	}
	testData1 := map[string]string{"key1": "value1"}
	err = zeroTTLCache.Set("test-key1", testData1, "test-source")
	require.NoError(t, err, "Cache.Set should not return an error")

	// Set a cache entry that doesn't expire
	testData2 := map[string]string{"key2": "value2"}
	err = cache.Set("test-key2", testData2, "test-source")
	require.NoError(t, err, "Cache.Set should not return an error")

	// Cleanup the cache
	err = cache.Cleanup()
	require.NoError(t, err, "Cache.Cleanup should not return an error")

	// Check that expired entry was deleted
	var resultData1 map[string]string
	found1 := cache.Get("test-key1", &resultData1)
	assert.False(t, found1, "Cache.Get should return false for an expired key")

	// Check that non-expired entry still exists
	var resultData2 map[string]string
	found2 := cache.Get("test-key2", &resultData2)
	assert.True(t, found2, "Cache.Get should return true for a non-expired key")
	assert.Equal(t, testData2, resultData2, "Retrieved data should match the original data")
}

// TestCacheExpansion tests expansion of $XDG_CACHE_HOME in cache directory path
func TestCacheExpansion(t *testing.T) {
	// Skip this test if we can't determine XDG_CACHE_HOME to avoid test failure
	// in environments where XDG_CACHE_HOME can't be determined
	oldCacheHome := os.Getenv("XDG_CACHE_HOME")
	tempCacheDir, err := os.MkdirTemp("", "xdg-cache-test")
	if err != nil {
		t.Skip("Unable to create temporary directory for XDG_CACHE_HOME")
	}
	defer os.RemoveAll(tempCacheDir)
	_ = os.Setenv("XDG_CACHE_HOME", tempCacheDir)
	defer func() { _ = os.Setenv("XDG_CACHE_HOME", oldCacheHome) }()

	// Create a cache with a path that starts with $XDG_CACHE_HOME
	cache, err := NewCache("$XDG_CACHE_HOME/cpan-test", 1)
	require.NoError(t, err, "NewCache should not return an error")
	require.NotNil(t, cache, "Cache should not be nil")

	// Check that the cache directory was expanded properly
	expected := filepath.Join(tempCacheDir, "cpan-test")
	assert.Equal(t, expected, cache.cacheDir, "Cache directory should be expanded")

	// Check that the directory was created
	_, err = os.Stat(expected)
	assert.NoError(t, err, "Cache directory should exist")
}
