// ABOUTME: Documentation caching system with HTTP conditional requests and TTL management
// ABOUTME: Supports disk-based caching with ETag/Last-Modified headers for efficient updates

package docs

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tamarou.com/pvm/internal/errors"
)

// DocsCache manages persistent caching for external documentation
type DocsCache struct {
	cacheDir   string
	ttl        time.Duration
	locks      map[string]*sync.Mutex
	globalLock sync.Mutex
}

// docsCacheEntry represents a cached documentation entry
type docsCacheEntry struct {
	Metadata docsCacheMetadata `json:"metadata"`
	Data     interface{}       `json:"data"`
}

// docsCacheMetadata contains metadata about cached documentation
type docsCacheMetadata struct {
	Key          string    `json:"key"`
	Timestamp    time.Time `json:"timestamp"`
	Expires      time.Time `json:"expires"`
	ETag         string    `json:"etag,omitempty"`
	LastModified string    `json:"last_modified,omitempty"`
	ContentHash  string    `json:"content_hash,omitempty"`
	URL          string    `json:"url,omitempty"`
	ContentType  string    `json:"content_type,omitempty"`
	Size         int64     `json:"size,omitempty"`
}

// NewDocsCache creates a new documentation cache instance
func NewDocsCache(cacheDir string, ttl time.Duration) *DocsCache {
	return &DocsCache{
		cacheDir: cacheDir,
		ttl:      ttl,
		locks:    make(map[string]*sync.Mutex),
	}
}

// Get retrieves a cached item and unmarshals it into the result
func (c *DocsCache) Get(key string, result interface{}) bool {
	mutex := c.getMutex(key)
	mutex.Lock()
	defer mutex.Unlock()

	entry, exists := c.getEntry(key)
	if !exists {
		return false
	}

	// Check if entry is expired
	if c.ttl > 0 && time.Now().After(entry.Metadata.Expires) {
		c.deleteEntry(key)
		return false
	}

	// Unmarshal data into result
	data, err := json.Marshal(entry.Data)
	if err != nil {
		return false
	}

	if err := json.Unmarshal(data, result); err != nil {
		return false
	}

	return true
}

// Set stores an item in the cache with the specified TTL
func (c *DocsCache) Set(key string, data interface{}, source string) error {
	return c.SetWithHTTPHeaders(key, data, source, nil, "")
}

// SetWithHTTPHeaders stores an item with HTTP headers for conditional requests
func (c *DocsCache) SetWithHTTPHeaders(key string, data interface{}, source string, headers http.Header, url string) error {
	mutex := c.getMutex(key)
	mutex.Lock()
	defer mutex.Unlock()

	// Calculate content hash
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return errors.NewDocumentationError("DOCS-001", "Failed to marshal data for caching", err)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(dataBytes))

	// Create metadata
	metadata := docsCacheMetadata{
		Key:         key,
		Timestamp:   time.Now(),
		ContentHash: hash,
		URL:         url,
		Size:        int64(len(dataBytes)),
	}

	// Set expiration
	switch {
	case c.ttl > 0:
		metadata.Expires = metadata.Timestamp.Add(c.ttl)
	case c.ttl == -1:
		// Never expire
		metadata.Expires = time.Time{}
	default:
		// TTL = 0, expire immediately (for testing)
		metadata.Expires = metadata.Timestamp.Add(-time.Hour)
	}

	// Extract HTTP headers if provided
	if headers != nil {
		metadata.ETag = headers.Get("ETag")
		metadata.LastModified = headers.Get("Last-Modified")
		metadata.ContentType = headers.Get("Content-Type")
	}

	// Create cache entry
	entry := docsCacheEntry{
		Metadata: metadata,
		Data:     data,
	}

	// Store to disk
	return c.setEntry(key, entry)
}

// IsExpired checks if a cached item is expired
func (c *DocsCache) IsExpired(key string) bool {
	mutex := c.getMutex(key)
	mutex.Lock()
	defer mutex.Unlock()

	entry, exists := c.getEntry(key)
	if !exists {
		return true
	}

	return c.ttl > 0 && time.Now().After(entry.Metadata.Expires)
}

// GetConditionalHeaders returns headers for HTTP conditional requests
func (c *DocsCache) GetConditionalHeaders(key string) http.Header {
	mutex := c.getMutex(key)
	mutex.Lock()
	defer mutex.Unlock()

	entry, exists := c.getEntry(key)
	if !exists {
		return nil
	}

	headers := make(http.Header)
	if entry.Metadata.ETag != "" {
		headers.Set("If-None-Match", entry.Metadata.ETag)
	}
	if entry.Metadata.LastModified != "" {
		headers.Set("If-Modified-Since", entry.Metadata.LastModified)
	}

	return headers
}

// GetMetadata returns metadata for a cached item
func (c *DocsCache) GetMetadata(key string) (*docsCacheMetadata, bool) {
	mutex := c.getMutex(key)
	mutex.Lock()
	defer mutex.Unlock()

	entry, exists := c.getEntry(key)
	if !exists {
		return nil, false
	}

	return &entry.Metadata, true
}

// Delete removes an item from the cache
func (c *DocsCache) Delete(key string) error {
	mutex := c.getMutex(key)
	mutex.Lock()
	defer mutex.Unlock()

	return c.deleteEntry(key)
}

// Clear removes all items from the cache
func (c *DocsCache) Clear() error {
	c.globalLock.Lock()
	defer c.globalLock.Unlock()

	if err := os.RemoveAll(c.cacheDir); err != nil {
		return errors.NewDocumentationError("DOCS-005", "Failed to clear documentation cache", err)
	}

	return nil
}

// ListKeys returns all cached keys
func (c *DocsCache) ListKeys() ([]string, error) {
	if _, err := os.Stat(c.cacheDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	var keys []string
	err := filepath.WalkDir(c.cacheDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Extract key from filename
		relPath, err := filepath.Rel(c.cacheDir, path)
		if err != nil {
			return err
		}

		key := strings.TrimSuffix(relPath, ".json")
		key = strings.ReplaceAll(key, string(filepath.Separator), "/")
		keys = append(keys, key)

		return nil
	})

	if err != nil {
		return nil, errors.NewDocumentationError("DOCS-005", "Failed to list cache keys", err)
	}

	return keys, nil
}

// getMutex returns a mutex for the given key
func (c *DocsCache) getMutex(key string) *sync.Mutex {
	c.globalLock.Lock()
	defer c.globalLock.Unlock()

	if mutex, exists := c.locks[key]; exists {
		return mutex
	}

	mutex := &sync.Mutex{}
	c.locks[key] = mutex
	return mutex
}

// getEntry retrieves a cache entry from disk
func (c *DocsCache) getEntry(key string) (docsCacheEntry, bool) {
	filePath := c.getFilePath(key)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return docsCacheEntry{}, false
	}

	var entry docsCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		// Cache corruption, remove the file
		os.Remove(filePath)
		return docsCacheEntry{}, false
	}

	return entry, true
}

// setEntry stores a cache entry to disk
func (c *DocsCache) setEntry(key string, entry docsCacheEntry) error {
	filePath := c.getFilePath(key)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return errors.NewDocumentationError("DOCS-005", "Failed to create cache directory", err)
	}

	// Marshal entry
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return errors.NewDocumentationError("DOCS-001", "Failed to marshal cache entry", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return errors.NewDocumentationError("DOCS-005", "Failed to write cache entry", err)
	}

	return nil
}

// deleteEntry removes a cache entry from disk
func (c *DocsCache) deleteEntry(key string) error {
	filePath := c.getFilePath(key)

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return errors.NewDocumentationError("DOCS-005", "Failed to delete cache entry", err)
	}

	return nil
}

// getFilePath returns the file path for a cache key
func (c *DocsCache) getFilePath(key string) string {
	// Sanitize key for filesystem
	safeKey := strings.ReplaceAll(key, "/", string(filepath.Separator))
	safeKey = strings.ReplaceAll(safeKey, ":", "_")
	safeKey = strings.ReplaceAll(safeKey, "?", "_")
	safeKey = strings.ReplaceAll(safeKey, "*", "_")
	safeKey = strings.ReplaceAll(safeKey, "<", "_")
	safeKey = strings.ReplaceAll(safeKey, ">", "_")
	safeKey = strings.ReplaceAll(safeKey, "|", "_")

	return filepath.Join(c.cacheDir, safeKey+".json")
}

// CleanupExpired removes expired entries from the cache
func (c *DocsCache) CleanupExpired() error {
	if c.ttl <= 0 {
		return nil // No cleanup needed for never-expire or immediate-expire cache
	}

	keys, err := c.ListKeys()
	if err != nil {
		return err
	}

	var errs []error
	for _, key := range keys {
		if c.IsExpired(key) {
			if err := c.Delete(key); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return errors.NewDocumentationError("DOCS-005", "Failed to cleanup some expired entries", errs[0])
	}

	return nil
}

// Stats returns cache statistics
func (c *DocsCache) Stats() (map[string]interface{}, error) {
	keys, err := c.ListKeys()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_entries": len(keys),
		"cache_dir":     c.cacheDir,
		"ttl_hours":     c.ttl.Hours(),
	}

	// Count expired entries
	expiredCount := 0
	totalSize := int64(0)

	for _, key := range keys {
		if c.IsExpired(key) {
			expiredCount++
		}

		if metadata, exists := c.GetMetadata(key); exists {
			totalSize += metadata.Size
		}
	}

	stats["expired_entries"] = expiredCount
	stats["total_size_bytes"] = totalSize

	return stats, nil
}
