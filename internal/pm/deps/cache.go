// ABOUTME: Cache for dependency resolution
// ABOUTME: Provides caching of dependency resolution results

package deps

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"tamarou.com/pvm/internal/errors"
)

// CacheEntry represents a cached dependency tree
type CacheEntry struct {
	// ModuleName is the name of the resolved module
	ModuleName string

	// CreatedAt is when the cache entry was created
	CreatedAt time.Time

	// Options holds the original resolution options (serialized)
	Options string

	// Result holds the resolution result (serialized)
	Result string
}

// DependencyCache provides caching for resolved dependencies
type DependencyCache struct {
	// CacheDir is the directory where cache files are stored
	CacheDir string

	// TTL is the time-to-live for cached entries in hours
	TTL int
}

// NewDependencyCache creates a new dependency cache
func NewDependencyCache(cacheDir string, ttl int) (*DependencyCache, error) {
	// Create cache directory if it doesn't exist
	if cacheDir != "" {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return nil, errors.NewSystemError(
				"PVI-3010",
				"Failed to create dependency cache directory",
				err)
		}
	}

	return &DependencyCache{
		CacheDir: cacheDir,
		TTL:      ttl,
	}, nil
}

// Get retrieves a cached dependency tree
func (c *DependencyCache) Get(moduleName string, optionsKey string) (*DependencyResolutionResult, bool) {
	if c.CacheDir == "" || c.TTL <= 0 {
		return nil, false
	}

	// Generate cache key
	cacheKey := generateCacheKey(moduleName, optionsKey)
	cacheFile := filepath.Join(c.CacheDir, cacheKey+".json")

	// Check if cache file exists
	info, err := os.Stat(cacheFile)
	if os.IsNotExist(err) {
		return nil, false
	}

	// Check if cache entry is expired
	if time.Since(info.ModTime()).Hours() > float64(c.TTL) {
		_ = os.Remove(cacheFile) // Clean up expired entry
		return nil, false
	}

	// Read and parse cache file
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	// Deserialize result
	var result DependencyResolutionResult
	if err := json.Unmarshal([]byte(entry.Result), &result); err != nil {
		return nil, false
	}

	return &result, true
}

// Set stores a dependency tree in the cache
func (c *DependencyCache) Set(moduleName string, optionsKey string, result *DependencyResolutionResult) error {
	if c.CacheDir == "" || c.TTL <= 0 || result == nil {
		return nil // Caching disabled or no result to cache
	}

	// Serialize result
	resultData, err := json.Marshal(result)
	if err != nil {
		return errors.NewSystemError(
			"PVI-3011",
			"Failed to serialize dependency result for caching",
			err)
	}

	// Create cache entry
	entry := CacheEntry{
		ModuleName: moduleName,
		CreatedAt:  time.Now(),
		Options:    optionsKey,
		Result:     string(resultData),
	}

	// Serialize entry
	data, err := json.Marshal(entry)
	if err != nil {
		return errors.NewSystemError(
			"PVI-3012",
			"Failed to serialize cache entry",
			err)
	}

	// Generate cache key
	cacheKey := generateCacheKey(moduleName, optionsKey)
	cacheFile := filepath.Join(c.CacheDir, cacheKey+".json")

	// Write to file
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return errors.NewSystemError(
			"PVI-3013",
			"Failed to write dependency cache file",
			err)
	}

	return nil
}

// Clear removes all cached entries
func (c *DependencyCache) Clear() error {
	if c.CacheDir == "" {
		return nil // Nothing to clear
	}

	// Read cache directory
	entries, err := os.ReadDir(c.CacheDir)
	if err != nil {
		return errors.NewSystemError(
			"PVI-3014",
			"Failed to read dependency cache directory",
			err)
	}

	// Remove each cache file
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}

		// Only remove JSON files
		if filepath.Ext(entry.Name()) == ".json" {
			if err := os.Remove(filepath.Join(c.CacheDir, entry.Name())); err != nil {
				// Continue even if one file fails to be removed
				continue
			}
		}
	}

	return nil
}

// CleanExpired removes expired cache entries
func (c *DependencyCache) CleanExpired() error {
	if c.CacheDir == "" || c.TTL <= 0 {
		return nil // Nothing to clean
	}

	// Read cache directory
	entries, err := os.ReadDir(c.CacheDir)
	if err != nil {
		return errors.NewSystemError(
			"PVI-3015",
			"Failed to read dependency cache directory",
			err)
	}

	// Check each cache file
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}

		// Only check JSON files
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		// Get file info to check modification time
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Remove if expired
		if time.Since(info.ModTime()).Hours() > float64(c.TTL) {
			_ = os.Remove(filepath.Join(c.CacheDir, entry.Name()))
		}
	}

	return nil
}

// generateCacheKey creates a unique key for a module and options
func generateCacheKey(moduleName, optionsKey string) string {
	// Simple implementation - in a real system you might want to hash the combination
	return moduleName + "-" + optionsKey
}
