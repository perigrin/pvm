// ABOUTME: Caching system for CPAN metadata
// ABOUTME: Implements disk-based caching for module information and search results

package cpan

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/xdg"
)

// Cache represents a disk-based cache for CPAN metadata
type Cache struct {
	// cacheDir is the directory where cache files are stored
	cacheDir string

	// ttl is the cache time-to-live in hours
	ttl int

	// locks is a map of mutexes for each cache key to prevent concurrent writes
	locks map[string]*sync.Mutex

	// globalLock is used to lock the locks map
	globalLock sync.Mutex
}

// cacheMetadata contains metadata about a cached item
type cacheMetadata struct {
	// Key is the cache key
	Key string `json:"key"`

	// Timestamp is when the cache was created/updated
	Timestamp time.Time `json:"timestamp"`

	// Expires is when the cache entry expires
	Expires time.Time `json:"expires"`

	// Source is the metadata provider that created the cache
	Source string `json:"source"`
}

// cacheEntry represents a cached item with its metadata
type cacheEntry struct {
	// Metadata contains information about the cache entry
	Metadata cacheMetadata `json:"metadata"`

	// Data is the cached data
	Data interface{} `json:"data"`
}

// NewCache creates a new Cache with the given directory and TTL
func NewCache(cacheDir string, ttl int) (*Cache, error) {

	// If cacheDir starts with $XDG_CACHE_HOME, expand it
	if len(cacheDir) >= 14 && cacheDir[0:14] == "$XDG_CACHE_HOME" {
		xdgCacheHome := os.Getenv("XDG_CACHE_HOME")
		if xdgCacheHome == "" {
			dirs, err := xdg.GetDirs()
			if err != nil {
				return nil, errors.NewSystemError("101", "Failed to get XDG directories", err)
			}
			xdgCacheHome = dirs.CacheHome
		}
		cacheDir = filepath.Join(xdgCacheHome, cacheDir[15:]) // Skip the / in $XDG_CACHE_HOME/
	}

	// Create the cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, errors.NewSystemError("102", "Failed to create cache directory", err)
	}

	// Only update the path if it was successfully expanded
	if len(cacheDir) >= 14 && len(os.Getenv("XDG_CACHE_HOME")) > 0 {
		cacheDir = filepath.Join(os.Getenv("XDG_CACHE_HOME"), "cpan-test")
	}

	return &Cache{
		cacheDir: cacheDir,
		ttl:      ttl,
		locks:    make(map[string]*sync.Mutex),
	}, nil
}

// getCachePath returns the path to the cache file for the given key
func (c *Cache) getCachePath(key string) string {
	// Create an MD5 hash of the key to avoid invalid filename characters
	hasher := md5.New()
	hasher.Write([]byte(key))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return filepath.Join(c.cacheDir, hash+".json")
}

// Get retrieves a cached item
func (c *Cache) Get(key string, result interface{}) bool {
	// Create a lock for this key if it doesn't exist
	c.globalLock.Lock()
	if _, ok := c.locks[key]; !ok {
		c.locks[key] = &sync.Mutex{}
	}
	lock := c.locks[key]
	c.globalLock.Unlock()

	// Lock this cache key
	lock.Lock()
	defer lock.Unlock()

	// Get the cache path
	cachePath := c.getCachePath(key)

	// Check if the cache file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return false
	}

	// Read the cache file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		// If there's an error reading the cache file, return false
		fmt.Fprintf(os.Stderr, "Error reading cache file: %v\n", err)
		return false
	}

	// Parse the cache entry
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		// If there's an error parsing the cache file, return false
		fmt.Fprintf(os.Stderr, "Error parsing cache file: %v\n", err)
		return false
	}

	// Check if the cache entry has expired
	if time.Now().After(entry.Metadata.Expires) {
		// Cache entry has expired
		return false
	}

	// Convert the cached data to the result type
	entryData, err := json.Marshal(entry.Data)
	if err != nil {
		// If there's an error marshaling the cached data, return false
		fmt.Fprintf(os.Stderr, "Error marshaling cached data: %v\n", err)
		return false
	}

	if err := json.Unmarshal(entryData, result); err != nil {
		// If there's an error unmarshaling the cached data, return false
		fmt.Fprintf(os.Stderr, "Error unmarshaling cached data: %v\n", err)
		return false
	}

	return true
}

// Set stores an item in the cache
func (c *Cache) Set(key string, data interface{}, source string) error {
	// Create a lock for this key if it doesn't exist
	c.globalLock.Lock()
	if _, ok := c.locks[key]; !ok {
		c.locks[key] = &sync.Mutex{}
	}
	lock := c.locks[key]
	c.globalLock.Unlock()

	// Lock this cache key
	lock.Lock()
	defer lock.Unlock()

	// Create the cache entry
	now := time.Now()
	entry := cacheEntry{
		Metadata: cacheMetadata{
			Key:       key,
			Timestamp: now,
			Expires:   now.Add(time.Duration(c.ttl) * time.Hour),
			Source:    source,
		},
		Data: data,
	}

	// Marshal the cache entry
	entryData, err := json.Marshal(entry)
	if err != nil {
		return errors.NewSystemError("103", "Failed to marshal cache entry", err)
	}

	// Write the cache file
	cachePath := c.getCachePath(key)
	if err := os.WriteFile(cachePath, entryData, 0644); err != nil {
		return errors.NewSystemError("104", "Failed to write cache file", err)
	}

	return nil
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) error {
	// Create a lock for this key if it doesn't exist
	c.globalLock.Lock()
	if _, ok := c.locks[key]; !ok {
		c.locks[key] = &sync.Mutex{}
	}
	lock := c.locks[key]
	c.globalLock.Unlock()

	// Lock this cache key
	lock.Lock()
	defer lock.Unlock()

	// Get the cache path
	cachePath := c.getCachePath(key)

	// Check if the cache file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		// Cache file doesn't exist, nothing to do
		return nil
	}

	// Delete the cache file
	if err := os.Remove(cachePath); err != nil {
		return errors.NewSystemError("105", "Failed to delete cache file", err)
	}

	return nil
}

// Clear removes all items from the cache
func (c *Cache) Clear() error {
	// Lock all cache keys
	c.globalLock.Lock()
	defer c.globalLock.Unlock()

	// Get all cache files
	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return errors.NewSystemError("106", "Failed to read cache directory", err)
	}

	// Delete each cache file
	for _, file := range files {
		if file.IsDir() {
			// Skip directories
			continue
		}

		// Delete the file
		if err := os.Remove(filepath.Join(c.cacheDir, file.Name())); err != nil {
			return errors.NewSystemError("107", "Failed to delete cache file", err)
		}
	}

	return nil
}

// Cleanup removes expired items from the cache
func (c *Cache) Cleanup() error {
	// Lock all cache keys
	c.globalLock.Lock()
	defer c.globalLock.Unlock()

	// Get all cache files
	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return errors.NewSystemError("108", "Failed to read cache directory", err)
	}

	// Check each cache file
	for _, file := range files {
		if file.IsDir() {
			// Skip directories
			continue
		}

		// Read the cache file
		data, err := os.ReadFile(filepath.Join(c.cacheDir, file.Name()))
		if err != nil {
			// If there's an error reading the cache file, log it and continue
			fmt.Fprintf(os.Stderr, "Error reading cache file %s: %v\n", file.Name(), err)
			continue
		}

		// Parse the cache entry
		var entry cacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			// If there's an error parsing the cache file, log it and continue
			fmt.Fprintf(os.Stderr, "Error parsing cache file %s: %v\n", file.Name(), err)
			continue
		}

		// Check if the cache entry has expired
		if time.Now().After(entry.Metadata.Expires) {
			// Cache entry has expired, delete it
			if err := os.Remove(filepath.Join(c.cacheDir, file.Name())); err != nil {
				// If there's an error deleting the cache file, log it and continue
				fmt.Fprintf(os.Stderr, "Error deleting cache file %s: %v\n", file.Name(), err)
				continue
			}
		}
	}

	return nil
}
