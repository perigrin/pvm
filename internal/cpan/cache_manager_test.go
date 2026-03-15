// ABOUTME: Tests for cache management functionality
// ABOUTME: Validates cache operations, cleanup, and statistics

package cpan

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCacheManager(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.New(os.Stderr, "[TestCache] ", log.LstdFlags)

	cm, err := NewCacheManager(tempDir, logger)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	if cm.cacheDir != tempDir {
		t.Errorf("Expected cache dir %s, got %s", tempDir, cm.cacheDir)
	}

	// Check that directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}
}

func TestCacheManagerValidateCache(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.New(os.Stderr, "[TestCache] ", log.LstdFlags)

	cm, err := NewCacheManager(tempDir, logger)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	// Test with empty cache directory
	err = cm.ValidateCache()
	if err != nil {
		t.Errorf("ValidateCache failed on empty directory: %v", err)
	}

	// Create a valid cache file
	cache, err := NewCache(tempDir, 24)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	err = cache.Set("test-key", map[string]string{"test": "data"}, "test")
	if err != nil {
		t.Fatalf("Failed to set cache entry: %v", err)
	}

	// Validate should succeed
	err = cm.ValidateCache()
	if err != nil {
		t.Errorf("ValidateCache failed with valid cache file: %v", err)
	}

	// Create an invalid file
	invalidFile := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	// Validation should still succeed (it logs warnings but doesn't fail)
	err = cm.ValidateCache()
	if err != nil {
		t.Errorf("ValidateCache failed with invalid file present: %v", err)
	}
}

func TestCacheManagerGetCacheStats(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.New(os.Stderr, "[TestCache] ", log.LstdFlags)

	cm, err := NewCacheManager(tempDir, logger)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	// Test with empty cache
	stats, err := cm.GetCacheStats()
	if err != nil {
		t.Fatalf("GetCacheStats failed: %v", err)
	}

	if stats.TotalFiles != 0 {
		t.Errorf("Expected 0 files in empty cache, got %d", stats.TotalFiles)
	}

	if stats.CacheDirectory != tempDir {
		t.Errorf("Expected cache directory %s, got %s", tempDir, stats.CacheDirectory)
	}

	// Add some cache entries
	cache, err := NewCache(tempDir, 24)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Add valid entries
	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("test-key-%d", i)
		data := map[string]string{"test": fmt.Sprintf("data-%d", i)}
		err = cache.Set(key, data, "test")
		if err != nil {
			t.Fatalf("Failed to set cache entry %d: %v", i, err)
		}
	}

	// Add expired entry
	expiredCache, err := NewCache(tempDir, 0) // 0 TTL = immediately expired
	if err != nil {
		t.Fatalf("Failed to create expired cache: %v", err)
	}
	err = expiredCache.Set("expired-key", map[string]string{"test": "expired"}, "test")
	if err != nil {
		t.Fatalf("Failed to set expired cache entry: %v", err)
	}

	// Get updated stats
	stats, err = cm.GetCacheStats()
	if err != nil {
		t.Fatalf("GetCacheStats failed after adding entries: %v", err)
	}

	if stats.TotalFiles != 4 {
		t.Errorf("Expected 4 files in cache, got %d", stats.TotalFiles)
	}

	if stats.ValidFiles != 3 {
		t.Errorf("Expected 3 valid files, got %d", stats.ValidFiles)
	}

	if stats.ExpiredFiles != 1 {
		t.Errorf("Expected 1 expired file, got %d", stats.ExpiredFiles)
	}

	if stats.TotalSize <= 0 {
		t.Error("Expected positive total size")
	}

	if stats.AverageAge <= 0 {
		t.Error("Expected positive average age")
	}
}

func TestCacheManagerCleanupCache(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.New(os.Stderr, "[TestCache] ", log.LstdFlags)

	cm, err := NewCacheManager(tempDir, logger)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	// Create cache entries with different ages
	cache, err := NewCache(tempDir, 24)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Recent entry
	err = cache.Set("recent-key", map[string]string{"test": "recent"}, "test")
	if err != nil {
		t.Fatalf("Failed to set recent cache entry: %v", err)
	}

	// Create an old file by modifying its timestamp
	oldCache, err := NewCache(tempDir, 24)
	if err != nil {
		t.Fatalf("Failed to create old cache: %v", err)
	}
	err = oldCache.Set("old-key", map[string]string{"test": "old"}, "test")
	if err != nil {
		t.Fatalf("Failed to set old cache entry: %v", err)
	}

	// Manually modify the file timestamp to make it old
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read cache directory: %v", err)
	}

	if len(files) >= 2 {
		// Make one file old
		oldTime := time.Now().Add(-25 * time.Hour)
		oldFilePath := filepath.Join(tempDir, files[1].Name())
		err = os.Chtimes(oldFilePath, oldTime, oldTime)
		if err != nil {
			t.Fatalf("Failed to change file time: %v", err)
		}
	}

	// Test cleanup with 24-hour cutoff
	err = cm.CleanupCache(24 * time.Hour)
	if err != nil {
		t.Fatalf("CleanupCache failed: %v", err)
	}

	// Check that old file was removed
	filesAfter, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read cache directory after cleanup: %v", err)
	}

	if len(filesAfter) >= len(files) {
		t.Error("Expected some files to be removed during cleanup")
	}
}

func TestCacheManagerPurgeExpiredEntries(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.New(os.Stderr, "[TestCache] ", log.LstdFlags)

	cm, err := NewCacheManager(tempDir, logger)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	// Create valid entries
	cache, err := NewCache(tempDir, 24)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	err = cache.Set("valid-key", map[string]string{"test": "valid"}, "test")
	if err != nil {
		t.Fatalf("Failed to set valid cache entry: %v", err)
	}

	// Create expired entry
	expiredCache, err := NewCache(tempDir, 0) // 0 TTL = immediately expired
	if err != nil {
		t.Fatalf("Failed to create expired cache: %v", err)
	}
	err = expiredCache.Set("expired-key", map[string]string{"test": "expired"}, "test")
	if err != nil {
		t.Fatalf("Failed to set expired cache entry: %v", err)
	}

	// Check initial file count
	filesBefore, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read cache directory: %v", err)
	}

	// Purge expired entries
	err = cm.PurgeExpiredEntries()
	if err != nil {
		t.Fatalf("PurgeExpiredEntries failed: %v", err)
	}

	// Check that expired entry was removed
	filesAfter, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read cache directory after purge: %v", err)
	}

	if len(filesAfter) >= len(filesBefore) {
		t.Error("Expected expired files to be removed")
	}

	// Verify valid entry still exists
	var validData map[string]string
	found := cache.Get("valid-key", &validData)
	if !found {
		t.Error("Valid cache entry was incorrectly removed")
	}
}

func TestCacheManagerOptimizeCache(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.New(os.Stderr, "[TestCache] ", log.LstdFlags)

	cm, err := NewCacheManager(tempDir, logger)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	// Create mixed cache entries
	cache, err := NewCache(tempDir, 24)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Valid entry
	err = cache.Set("valid-key", map[string]string{"test": "valid"}, "test")
	if err != nil {
		t.Fatalf("Failed to set valid cache entry: %v", err)
	}

	// Expired entry
	expiredCache, err := NewCache(tempDir, 0)
	if err != nil {
		t.Fatalf("Failed to create expired cache: %v", err)
	}
	err = expiredCache.Set("expired-key", map[string]string{"test": "expired"}, "test")
	if err != nil {
		t.Fatalf("Failed to set expired cache entry: %v", err)
	}

	// Invalid file
	invalidFile := filepath.Join(tempDir, "invalid.txt")
	err = os.WriteFile(invalidFile, []byte("not json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	// Optimize cache
	err = cm.OptimizeCache()
	if err != nil {
		t.Fatalf("OptimizeCache failed: %v", err)
	}

	// Verify valid entry still exists
	var validData map[string]string
	found := cache.Get("valid-key", &validData)
	if !found {
		t.Error("Valid cache entry was incorrectly removed during optimization")
	}

	// Get final stats
	stats, err := cm.GetCacheStats()
	if err != nil {
		t.Fatalf("GetCacheStats failed after optimization: %v", err)
	}

	// Should have fewer expired files after optimization
	if stats.ExpiredFiles > 0 {
		t.Errorf("Expected 0 expired files after optimization, got %d", stats.ExpiredFiles)
	}
}

func TestCacheManagerNonExistentDirectory(t *testing.T) {
	// Use a path guaranteed to fail on all platforms: a subdirectory of a regular
	// file. Creating a directory inside a file is impossible on Linux, macOS, and
	// Windows, unlike a Unix-rooted path such as "/this/path/does/not/exist" which
	// Windows may interpret relative to the current drive and create successfully.
	tempDir := t.TempDir()
	blockingFile := filepath.Join(tempDir, "not-a-dir")
	if err := os.WriteFile(blockingFile, []byte("block"), 0644); err != nil {
		t.Fatalf("Failed to create blocking file: %v", err)
	}
	nonExistentDir := filepath.Join(blockingFile, "subdir")
	logger := log.New(os.Stderr, "[TestCache] ", log.LstdFlags)

	cm, err := NewCacheManager(nonExistentDir, logger)
	if err == nil {
		t.Error("Expected error for non-existent directory path, but got none")
		// Clean up if it somehow succeeded
		os.RemoveAll(nonExistentDir)
	} else if cm != nil {
		// The cache manager should be nil when creation fails
		t.Error("Expected nil cache manager when creation fails")
	}
}

func TestCacheManagerWithNilLogger(t *testing.T) {
	tempDir := t.TempDir()

	cm, err := NewCacheManager(tempDir, nil)
	if err != nil {
		t.Fatalf("Failed to create cache manager with nil logger: %v", err)
	}

	if cm.logger == nil {
		t.Error("Expected default logger to be created when nil is passed")
	}

	// Should still work normally
	err = cm.ValidateCache()
	if err != nil {
		t.Errorf("ValidateCache failed with default logger: %v", err)
	}
}
