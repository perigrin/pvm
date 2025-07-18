// ABOUTME: Comprehensive tests for XDG_CACHE_HOME expansion in cache functionality
// ABOUTME: Tests fix for Issue #127 - ensures environment variables are properly expanded, not used literally

package cpan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXDGCacheHomeExpansion(t *testing.T) {
	// Save original environment
	oldCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if oldCacheHome != "" {
			_ = os.Setenv("XDG_CACHE_HOME", oldCacheHome)
		} else {
			_ = os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	t.Run("XDG_CACHE_HOME_Set", func(t *testing.T) {
		// Create temporary directory for XDG_CACHE_HOME
		tempCacheDir := t.TempDir()
		_ = os.Setenv("XDG_CACHE_HOME", tempCacheDir)

		// Test that $XDG_CACHE_HOME is expanded to the actual directory
		cache, err := NewCache("$XDG_CACHE_HOME/pvm-test", 24)
		require.NoError(t, err, "NewCache should not return an error")
		require.NotNil(t, cache, "Cache should not be nil")

		expectedDir := filepath.Join(tempCacheDir, "pvm-test")
		assert.Equal(t, expectedDir, cache.cacheDir, "Cache directory should expand XDG_CACHE_HOME")

		// Verify the directory was created
		_, err = os.Stat(expectedDir)
		assert.NoError(t, err, "Cache directory should exist after creation")

		// Most importantly, verify that the cache directory was expanded correctly
		// This test ensures that the environment variable is expanded and not used literally
		assert.Contains(t, cache.cacheDir, tempCacheDir, "Cache directory should be expanded and contain the temp directory path")
		assert.NotContains(t, cache.cacheDir, "$XDG_CACHE_HOME", "Cache directory should not contain literal '$XDG_CACHE_HOME'")
	})

	t.Run("XDG_CACHE_HOME_Unset", func(t *testing.T) {
		// Unset XDG_CACHE_HOME to test fallback behavior
		_ = os.Unsetenv("XDG_CACHE_HOME")

		// This should not expand since XDG_CACHE_HOME is not set
		cache, err := NewCache("$XDG_CACHE_HOME/pvm-test", 24)
		require.NoError(t, err, "NewCache should not return an error even when XDG_CACHE_HOME is unset")
		require.NotNil(t, cache, "Cache should not be nil")

		// When XDG_CACHE_HOME is unset, it should use the literal path
		expectedDir := "$XDG_CACHE_HOME/pvm-test"
		assert.Equal(t, expectedDir, cache.cacheDir, "Cache directory should remain literal when XDG_CACHE_HOME is unset")

		// Verify the literal directory was created
		_, err = os.Stat(expectedDir)
		assert.NoError(t, err, "Literal cache directory should exist when XDG_CACHE_HOME is unset")
	})

	t.Run("XDG_CACHE_HOME_With_Braces", func(t *testing.T) {
		// Test ${XDG_CACHE_HOME} format
		tempCacheDir := t.TempDir()
		_ = os.Setenv("XDG_CACHE_HOME", tempCacheDir)

		cache, err := NewCache("${XDG_CACHE_HOME}/pvm-test", 24)
		require.NoError(t, err, "NewCache should not return an error")
		require.NotNil(t, cache, "Cache should not be nil")

		expectedDir := filepath.Join(tempCacheDir, "pvm-test")
		assert.Equal(t, expectedDir, cache.cacheDir, "Cache directory should expand ${XDG_CACHE_HOME}")

		// Verify the directory was created
		_, err = os.Stat(expectedDir)
		assert.NoError(t, err, "Cache directory should exist after creation")
	})

	t.Run("XDG_CACHE_HOME_Embedded_In_Path", func(t *testing.T) {
		// Test embedded environment variable in path
		tempCacheDir := t.TempDir()
		tempPrefix := t.TempDir()
		_ = os.Setenv("XDG_CACHE_HOME", tempCacheDir)

		cache, err := NewCache(tempPrefix+"/$XDG_CACHE_HOME/pvm-test", 24)
		require.NoError(t, err, "NewCache should not return an error")
		require.NotNil(t, cache, "Cache should not be nil")

		expectedDir := filepath.Clean(filepath.Join(tempPrefix, tempCacheDir, "pvm-test"))
		actualDir := filepath.Clean(cache.cacheDir)
		assert.Equal(t, expectedDir, actualDir, "Cache directory should expand embedded XDG_CACHE_HOME")

		// Verify the directory was created
		_, err = os.Stat(actualDir)
		assert.NoError(t, err, "Cache directory should exist after creation")
	})

	t.Run("Cache_Operations_With_Expanded_Path", func(t *testing.T) {
		// Test that cache operations work correctly with expanded paths
		tempCacheDir := t.TempDir()
		_ = os.Setenv("XDG_CACHE_HOME", tempCacheDir)

		cache, err := NewCache("$XDG_CACHE_HOME/pvm-test", 24)
		require.NoError(t, err, "NewCache should not return an error")

		// Test cache operations
		testKey := "test-key"
		testData := map[string]string{"test": "data"}

		// Store data in cache
		err = cache.Set(testKey, testData, "test-source")
		require.NoError(t, err, "Cache.Set should not return an error")

		// Retrieve data from cache
		var retrievedData map[string]string
		found := cache.Get(testKey, &retrievedData)
		assert.True(t, found, "Cache.Get should find the stored data")
		assert.Equal(t, testData, retrievedData, "Retrieved data should match stored data")

		// Verify cache file was created in the expanded directory
		expectedDir := filepath.Join(tempCacheDir, "pvm-test")
		files, err := os.ReadDir(expectedDir)
		require.NoError(t, err, "Should be able to read cache directory")
		assert.Len(t, files, 1, "Cache directory should contain one file")
		assert.True(t, files[0].Name() != "", "Cache file should have a name")
	})

	t.Run("Multiple_Environment_Variables", func(t *testing.T) {
		// Test multiple environment variables in path
		tempCacheDir := t.TempDir()
		tempPrefix := t.TempDir()
		_ = os.Setenv("XDG_CACHE_HOME", tempCacheDir)
		_ = os.Setenv("TEST_PREFIX", tempPrefix)
		defer func() { _ = os.Unsetenv("TEST_PREFIX") }()

		cache, err := NewCache("$TEST_PREFIX/$XDG_CACHE_HOME/pvm-test", 24)
		require.NoError(t, err, "NewCache should not return an error")
		require.NotNil(t, cache, "Cache should not be nil")

		// Clean up the expected path by removing any double slashes
		expectedDir := filepath.Clean(filepath.Join(tempPrefix, tempCacheDir, "pvm-test"))
		actualDir := filepath.Clean(cache.cacheDir)
		assert.Equal(t, expectedDir, actualDir, "Cache directory should expand multiple environment variables")

		// Verify the directory was created
		_, err = os.Stat(actualDir)
		assert.NoError(t, err, "Cache directory should exist after creation")
	})
}

func TestEnvironmentVariableExpansion(t *testing.T) {
	// Test the expandEnvironmentVariables function directly
	oldCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if oldCacheHome != "" {
			_ = os.Setenv("XDG_CACHE_HOME", oldCacheHome)
		} else {
			_ = os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	t.Run("Simple_Variable", func(t *testing.T) {
		_ = os.Setenv("XDG_CACHE_HOME", "/tmp/test-cache")

		result := expandEnvironmentVariables("$XDG_CACHE_HOME")
		assert.Equal(t, "/tmp/test-cache", result, "Simple variable should be expanded")
	})

	t.Run("Variable_With_Braces", func(t *testing.T) {
		_ = os.Setenv("XDG_CACHE_HOME", "/tmp/test-cache")

		result := expandEnvironmentVariables("${XDG_CACHE_HOME}")
		assert.Equal(t, "/tmp/test-cache", result, "Variable with braces should be expanded")
	})

	t.Run("Variable_In_Path", func(t *testing.T) {
		_ = os.Setenv("XDG_CACHE_HOME", "/tmp/test-cache")

		result := expandEnvironmentVariables("$XDG_CACHE_HOME/subdir")
		assert.Equal(t, "/tmp/test-cache/subdir", result, "Variable in path should be expanded")
	})

	t.Run("Variable_With_Braces_In_Path", func(t *testing.T) {
		_ = os.Setenv("XDG_CACHE_HOME", "/tmp/test-cache")

		result := expandEnvironmentVariables("${XDG_CACHE_HOME}/subdir")
		assert.Equal(t, "/tmp/test-cache/subdir", result, "Variable with braces in path should be expanded")
	})

	t.Run("Nonexistent_Variable", func(t *testing.T) {
		_ = os.Unsetenv("XDG_CACHE_HOME")

		result := expandEnvironmentVariables("$XDG_CACHE_HOME")
		assert.Equal(t, "$XDG_CACHE_HOME", result, "Nonexistent variable should remain unchanged")
	})

	t.Run("Empty_String", func(t *testing.T) {
		result := expandEnvironmentVariables("")
		assert.Equal(t, "", result, "Empty string should remain empty")
	})

	t.Run("No_Variables", func(t *testing.T) {
		result := expandEnvironmentVariables("/path/without/variables")
		assert.Equal(t, "/path/without/variables", result, "Path without variables should remain unchanged")
	})

	t.Run("Multiple_Variables", func(t *testing.T) {
		_ = os.Setenv("XDG_CACHE_HOME", "/tmp/cache")
		_ = os.Setenv("TEST_VAR", "test")
		defer func() { _ = os.Unsetenv("TEST_VAR") }()

		result := expandEnvironmentVariables("$XDG_CACHE_HOME/$TEST_VAR/subdir")
		assert.Equal(t, "/tmp/cache/test/subdir", result, "Multiple variables should be expanded")
	})

	t.Run("Mixed_Formats", func(t *testing.T) {
		_ = os.Setenv("XDG_CACHE_HOME", "/tmp/cache")
		_ = os.Setenv("TEST_VAR", "test")
		defer func() { _ = os.Unsetenv("TEST_VAR") }()

		result := expandEnvironmentVariables("$XDG_CACHE_HOME/${TEST_VAR}/subdir")
		assert.Equal(t, "/tmp/cache/test/subdir", result, "Mixed variable formats should be expanded")
	})
}
