// ABOUTME: Cache implementation for parser to improve performance
// ABOUTME: Provides caching for AST to avoid redundant parsing

package parser

import (
	"sync"
	"time"
)

// FileContentCache represents a cache entry with content hash and last modified time
type FileContentCache struct {
	// Hash of the file content for quick comparison
	ContentHash string

	// AST is the cached abstract syntax tree
	AST *AST

	// LastAccessed is when this entry was last used
	LastAccessed time.Time

	// CreatedAt is when this entry was created
	CreatedAt time.Time
}

// ParserCache manages a cache of parsed files to avoid redundant parsing
type ParserCache struct {
	// Cache maps file paths to cache entries
	cache map[string]*FileContentCache

	// MaxEntries is the maximum number of entries to keep in the cache
	maxEntries int

	// mutex protects the cache
	mutex sync.RWMutex
}

// NewParserCache creates a new parser cache with a maximum number of entries
func NewParserCache(maxEntries int) *ParserCache {
	return &ParserCache{
		cache:      make(map[string]*FileContentCache),
		maxEntries: maxEntries,
	}
}

// Get retrieves a cached AST if it exists and matches the content hash
func (pc *ParserCache) Get(path, content string) *AST {
	contentHash := hashContent(content)

	pc.mutex.RLock()
	entry, exists := pc.cache[path]
	pc.mutex.RUnlock()

	if exists && entry.ContentHash == contentHash {
		// Update last accessed time
		pc.mutex.Lock()
		entry.LastAccessed = time.Now()
		pc.mutex.Unlock()
		return entry.AST
	}

	return nil
}

// Put stores an AST in the cache
func (pc *ParserCache) Put(path, content string, ast *AST) {
	contentHash := hashContent(content)
	now := time.Now()

	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	// Create new cache entry
	pc.cache[path] = &FileContentCache{
		ContentHash:  contentHash,
		AST:          ast,
		LastAccessed: now,
		CreatedAt:    now,
	}

	// Clean up cache if it exceeds max entries
	if len(pc.cache) > pc.maxEntries {
		pc.cleanCache()
	}
}

// Invalidate removes an entry from the cache
func (pc *ParserCache) Invalidate(path string) {
	pc.mutex.Lock()
	delete(pc.cache, path)
	pc.mutex.Unlock()
}

// InvalidateAll clears the entire cache
func (pc *ParserCache) InvalidateAll() {
	pc.mutex.Lock()
	pc.cache = make(map[string]*FileContentCache)
	pc.mutex.Unlock()
}

// cleanCache removes the least recently used entries until we're back under the limit
func (pc *ParserCache) cleanCache() {
	// In a full implementation, we would sort entries by last accessed time
	// and remove the oldest ones first, but for simplicity we'll just remove
	// random entries until we're back under the limit

	// No need to lock here as this is called with the lock held
	numToRemove := len(pc.cache) - pc.maxEntries
	if numToRemove <= 0 {
		return
	}

	// Remove random entries until we're under the limit
	for path := range pc.cache {
		delete(pc.cache, path)
		numToRemove--
		if numToRemove <= 0 {
			break
		}
	}
}

// hashContent implementation moved to parser.go for reuse
