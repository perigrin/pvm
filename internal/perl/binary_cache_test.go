// ABOUTME: Tests for Perl binary cache management functionality
// ABOUTME: Ensures proper binary caching, validation, and cleanup

package perl

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tamarou.com/pvm/internal/xdg"
)

// Setup for testing binary cache
type binaryCacheTestEnv struct {
	// Temporary directories
	tempDir   string
	cacheDir  string
	binaryDir string

	// Cleanup functions
	cleanup []func()
}

// Setup binary cache test environment
func setupBinaryCacheTest(t *testing.T) *binaryCacheTestEnv {
	env := &binaryCacheTestEnv{
		cleanup: []func(){},
	}

	// Create temporary directory for cache
	tempDir, err := os.MkdirTemp("", "pvm-binary-cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	env.tempDir = tempDir
	env.cleanup = append(env.cleanup, func() { _ = os.RemoveAll(tempDir) })

	// Create cache and binary directories
	env.cacheDir = filepath.Join(tempDir, "cache")
	env.binaryDir = filepath.Join(env.cacheDir, "binaries")
	err = os.MkdirAll(env.binaryDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create binary directory: %v", err)
	}

	// Mock for xdg.GetDirs
	originalGetDirs := xdg.GetDirs
	env.cleanup = append(env.cleanup, func() { xdg.GetDirs = originalGetDirs })

	xdg.GetDirs = func() (*xdg.Dirs, error) {
		dirs := &xdg.Dirs{
			CacheDir: env.cacheDir,
		}

		// Mock EnsureDirs method
		dirs.EnsureDirs = func() error {
			return nil
		}

		return dirs, nil
	}

	return env
}

// Cleanup binary cache test environment
func (env *binaryCacheTestEnv) cleanup_() {
	// Run cleanup functions in reverse order
	for i := len(env.cleanup) - 1; i >= 0; i-- {
		env.cleanup[i]()
	}
}

// Test BinaryCache creation
func TestNewBinaryCache(t *testing.T) {
	env := setupBinaryCacheTest(t)
	defer env.cleanup_()

	cache, err := NewBinaryCache()
	if err != nil {
		t.Fatalf("Failed to create binary cache: %v", err)
	}

	if cache == nil {
		t.Fatal("Expected non-nil cache")
	}

	// Verify cache directory was created
	if _, err := os.Stat(cache.cacheDir); os.IsNotExist(err) {
		t.Errorf("Cache directory was not created: %s", cache.cacheDir)
	}
}

// Test binary cache put and get operations
func TestBinaryCachePutGet(t *testing.T) {
	env := setupBinaryCacheTest(t)
	defer env.cleanup_()

	cache, err := NewBinaryCache()
	if err != nil {
		t.Fatalf("Failed to create binary cache: %v", err)
	}

	// Create test binary content
	testContent := []byte("test binary content for caching")
	testChecksum := fmt.Sprintf("%x", sha256.Sum256(testContent))

	// Create a temporary file with test content
	tempFile, err := os.CreateTemp("", "test-binary-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(testContent)
	if err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}
	tempFile.Close()

	// Test putting binary in cache
	version := "5.38.0"
	platform := "linux-amd64"

	entry := &BinaryCacheEntry{
		Version:     version,
		Platform:    platform,
		Checksum:    testChecksum,
		Size:        int64(len(testContent)),
		DownloadURL: "https://example.com/test.tar.gz",
		CachedAt:    time.Now(),
	}

	cachedPath, err := cache.Put(tempFile.Name(), entry)
	if err != nil {
		t.Fatalf("Failed to put binary in cache: %v", err)
	}

	// Verify cached file exists
	if _, err := os.Stat(cachedPath); os.IsNotExist(err) {
		t.Errorf("Cached file does not exist: %s", cachedPath)
	}

	// Test getting binary from cache
	result, err := cache.Get(version, platform)
	if err != nil {
		t.Fatalf("Failed to get binary from cache: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil cache result")
	}

	if result.Path != cachedPath {
		t.Errorf("Expected cached path %s, got %s", cachedPath, result.Path)
	}

	if result.Entry.Version != version {
		t.Errorf("Expected version %s, got %s", version, result.Entry.Version)
	}

	if result.Entry.Platform != platform {
		t.Errorf("Expected platform %s, got %s", platform, result.Entry.Platform)
	}

	if result.Entry.Checksum != testChecksum {
		t.Errorf("Expected checksum %s, got %s", testChecksum, result.Entry.Checksum)
	}
}

// Test cache miss scenario
func TestBinaryCacheMiss(t *testing.T) {
	env := setupBinaryCacheTest(t)
	defer env.cleanup_()

	cache, err := NewBinaryCache()
	if err != nil {
		t.Fatalf("Failed to create binary cache: %v", err)
	}

	// Try to get a binary that doesn't exist in cache
	result, err := cache.Get("5.38.0", "linux-amd64")
	if err != nil {
		t.Fatalf("Unexpected error on cache miss: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result for cache miss, got %+v", result)
	}
}

// Test cache validation
func TestBinaryCacheValidation(t *testing.T) {
	env := setupBinaryCacheTest(t)
	defer env.cleanup_()

	cache, err := NewBinaryCache()
	if err != nil {
		t.Fatalf("Failed to create binary cache: %v", err)
	}

	// Create test binary with known content and checksum
	testContent := []byte("test binary for validation")
	correctChecksum := fmt.Sprintf("%x", sha256.Sum256(testContent))
	wrongChecksum := "incorrect_checksum"

	tempFile, err := os.CreateTemp("", "test-binary-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(testContent)
	if err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}
	tempFile.Close()

	version := "5.38.0"
	platform := "linux-amd64"

	// Test with correct checksum
	entry := &BinaryCacheEntry{
		Version:     version,
		Platform:    platform,
		Checksum:    correctChecksum,
		Size:        int64(len(testContent)),
		DownloadURL: "https://example.com/test.tar.gz",
		CachedAt:    time.Now(),
	}

	cachedPath, err := cache.Put(tempFile.Name(), entry)
	if err != nil {
		t.Fatalf("Failed to put binary in cache: %v", err)
	}

	// Validate with correct checksum should succeed
	valid, err := cache.Validate(cachedPath, correctChecksum)
	if err != nil {
		t.Fatalf("Failed to validate cached binary: %v", err)
	}

	if !valid {
		t.Errorf("Expected validation to succeed with correct checksum")
	}

	// Validate with wrong checksum should fail
	valid, err = cache.Validate(cachedPath, wrongChecksum)
	if err != nil {
		t.Fatalf("Failed to validate cached binary: %v", err)
	}

	if valid {
		t.Errorf("Expected validation to fail with wrong checksum")
	}
}

// Test cache listing
func TestBinaryCacheList(t *testing.T) {
	env := setupBinaryCacheTest(t)
	defer env.cleanup_()

	cache, err := NewBinaryCache()
	if err != nil {
		t.Fatalf("Failed to create binary cache: %v", err)
	}

	// Initially should be empty
	entries, err := cache.List()
	if err != nil {
		t.Fatalf("Failed to list cache entries: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected empty cache, got %d entries", len(entries))
	}

	// Add some test entries
	testCases := []struct {
		version  string
		platform string
	}{
		{"5.38.0", "linux-amd64"},
		{"5.38.0", "darwin-arm64"},
		{"5.36.0", "linux-amd64"},
	}

	for _, tc := range testCases {
		testContent := []byte(fmt.Sprintf("test content for %s-%s", tc.version, tc.platform))
		testChecksum := fmt.Sprintf("%x", sha256.Sum256(testContent))

		tempFile, err := os.CreateTemp("", "test-binary-*")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		_, err = tempFile.Write(testContent)
		if err != nil {
			t.Fatalf("Failed to write test content: %v", err)
		}
		tempFile.Close()

		entry := &BinaryCacheEntry{
			Version:     tc.version,
			Platform:    tc.platform,
			Checksum:    testChecksum,
			Size:        int64(len(testContent)),
			DownloadURL: "https://example.com/test.tar.gz",
			CachedAt:    time.Now(),
		}

		_, err = cache.Put(tempFile.Name(), entry)
		if err != nil {
			t.Fatalf("Failed to put binary in cache: %v", err)
		}
	}

	// List entries again
	entries, err = cache.List()
	if err != nil {
		t.Fatalf("Failed to list cache entries: %v", err)
	}

	if len(entries) != len(testCases) {
		t.Errorf("Expected %d cache entries, got %d", len(testCases), len(entries))
	}

	// Verify all expected entries are present
	for _, tc := range testCases {
		found := false
		for _, entry := range entries {
			if entry.Version == tc.version && entry.Platform == tc.platform {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected cache entry for %s-%s not found", tc.version, tc.platform)
		}
	}
}

// Test cache cleanup by age
func TestBinaryCacheCleanupByAge(t *testing.T) {
	env := setupBinaryCacheTest(t)
	defer env.cleanup_()

	cache, err := NewBinaryCache()
	if err != nil {
		t.Fatalf("Failed to create binary cache: %v", err)
	}

	// Create old and new cache entries
	oldTime := time.Now().Add(-35 * 24 * time.Hour) // 35 days ago
	newTime := time.Now().Add(-5 * 24 * time.Hour)  // 5 days ago

	testCases := []struct {
		version      string
		platform     string
		cachedAt     time.Time
		shouldRemain bool
	}{
		{"5.38.0", "linux-amd64", oldTime, false}, // Should be cleaned up
		{"5.38.0", "darwin-arm64", newTime, true}, // Should remain
		{"5.36.0", "linux-amd64", oldTime, false}, // Should be cleaned up
	}

	for _, tc := range testCases {
		testContent := []byte(fmt.Sprintf("test content for %s-%s", tc.version, tc.platform))
		testChecksum := fmt.Sprintf("%x", sha256.Sum256(testContent))

		tempFile, err := os.CreateTemp("", "test-binary-*")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		_, err = tempFile.Write(testContent)
		if err != nil {
			t.Fatalf("Failed to write test content: %v", err)
		}
		tempFile.Close()

		entry := &BinaryCacheEntry{
			Version:     tc.version,
			Platform:    tc.platform,
			Checksum:    testChecksum,
			Size:        int64(len(testContent)),
			DownloadURL: "https://example.com/test.tar.gz",
			CachedAt:    tc.cachedAt,
		}

		_, err = cache.Put(tempFile.Name(), entry)
		if err != nil {
			t.Fatalf("Failed to put binary in cache: %v", err)
		}
	}

	// Perform cleanup by age (30 days)
	cleaned, err := cache.CleanByAge(30 * 24 * time.Hour)
	if err != nil {
		t.Fatalf("Failed to clean cache by age: %v", err)
	}

	// Should have cleaned 2 old entries
	expectedCleaned := 2
	if cleaned != expectedCleaned {
		t.Errorf("Expected %d entries cleaned, got %d", expectedCleaned, cleaned)
	}

	// Verify remaining entries
	entries, err := cache.List()
	if err != nil {
		t.Fatalf("Failed to list cache entries: %v", err)
	}

	expectedRemaining := 1
	if len(entries) != expectedRemaining {
		t.Errorf("Expected %d remaining entries, got %d", expectedRemaining, len(entries))
	}

	// Verify the remaining entry is the one that should remain
	if len(entries) > 0 {
		entry := entries[0]
		if entry.Version != "5.38.0" || entry.Platform != "darwin-arm64" {
			t.Errorf("Wrong entry remained: %s-%s", entry.Version, entry.Platform)
		}
	}
}

// Test cache cleanup by size
func TestBinaryCacheCleanupBySize(t *testing.T) {
	env := setupBinaryCacheTest(t)
	defer env.cleanup_()

	cache, err := NewBinaryCache()
	if err != nil {
		t.Fatalf("Failed to create binary cache: %v", err)
	}

	// Create cache entries with different sizes
	testCases := []struct {
		version     string
		platform    string
		contentSize int
	}{
		{"5.38.0", "linux-amd64", 1000},  // 1KB
		{"5.38.0", "darwin-arm64", 2000}, // 2KB
		{"5.36.0", "linux-amd64", 3000},  // 3KB
	}

	for _, tc := range testCases {
		testContent := make([]byte, tc.contentSize)
		for i := range testContent {
			testContent[i] = byte(i % 256)
		}
		testChecksum := fmt.Sprintf("%x", sha256.Sum256(testContent))

		tempFile, err := os.CreateTemp("", "test-binary-*")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		_, err = tempFile.Write(testContent)
		if err != nil {
			t.Fatalf("Failed to write test content: %v", err)
		}
		tempFile.Close()

		entry := &BinaryCacheEntry{
			Version:     tc.version,
			Platform:    tc.platform,
			Checksum:    testChecksum,
			Size:        int64(len(testContent)),
			DownloadURL: "https://example.com/test.tar.gz",
			CachedAt:    time.Now(),
		}

		_, err = cache.Put(tempFile.Name(), entry)
		if err != nil {
			t.Fatalf("Failed to put binary in cache: %v", err)
		}
	}

	// Perform cleanup by size (keep only 4KB total)
	maxSize := int64(4000) // 4KB
	cleaned, err := cache.CleanBySize(maxSize)
	if err != nil {
		t.Fatalf("Failed to clean cache by size: %v", err)
	}

	// Should have cleaned at least 1 entry to get under the limit
	if cleaned < 1 {
		t.Errorf("Expected at least 1 entry cleaned, got %d", cleaned)
	}

	// Verify total size is under the limit
	entries, err := cache.List()
	if err != nil {
		t.Fatalf("Failed to list cache entries: %v", err)
	}

	totalSize := int64(0)
	for _, entry := range entries {
		totalSize += entry.Size
	}

	if totalSize > maxSize {
		t.Errorf("Total cache size %d exceeds limit %d", totalSize, maxSize)
	}
}

// Test cache corruption handling
func TestBinaryCacheCorruption(t *testing.T) {
	env := setupBinaryCacheTest(t)
	defer env.cleanup_()

	cache, err := NewBinaryCache()
	if err != nil {
		t.Fatalf("Failed to create binary cache: %v", err)
	}

	// Create a cache entry
	testContent := []byte("test binary content")
	testChecksum := fmt.Sprintf("%x", sha256.Sum256(testContent))

	tempFile, err := os.CreateTemp("", "test-binary-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(testContent)
	if err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}
	tempFile.Close()

	version := "5.38.0"
	platform := "linux-amd64"

	entry := &BinaryCacheEntry{
		Version:     version,
		Platform:    platform,
		Checksum:    testChecksum,
		Size:        int64(len(testContent)),
		DownloadURL: "https://example.com/test.tar.gz",
		CachedAt:    time.Now(),
	}

	cachedPath, err := cache.Put(tempFile.Name(), entry)
	if err != nil {
		t.Fatalf("Failed to put binary in cache: %v", err)
	}

	// Corrupt the cached file
	corruptContent := []byte("corrupted content")
	err = os.WriteFile(cachedPath, corruptContent, 0644)
	if err != nil {
		t.Fatalf("Failed to corrupt cached file: %v", err)
	}

	// Try to get the corrupted entry - should detect corruption
	result, err := cache.Get(version, platform)
	if err != nil {
		t.Fatalf("Unexpected error getting corrupted entry: %v", err)
	}

	// The cache should return the entry but validation should fail
	if result == nil {
		t.Fatal("Expected cache entry even if corrupted")
	}

	// Validate should fail
	valid, err := cache.Validate(result.Path, testChecksum)
	if err != nil {
		t.Fatalf("Failed to validate corrupted entry: %v", err)
	}

	if valid {
		t.Errorf("Expected validation to fail for corrupted file")
	}
}
