// ABOUTME: Perl binary cache management functionality
// ABOUTME: Provides local caching for downloaded Perl binary distributions with integrity validation

package perl

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/xdg"
)

// Cache error codes
const (
	ErrCacheInitFailed   = "601" // Failed to initialize cache
	ErrCacheOperFailed   = "602" // Cache operation failed
	ErrCacheCorrupted    = "603" // Cache is corrupted
	ErrCacheEntryInvalid = "604" // Cache entry is invalid
)

// Default cache configuration
const (
	DefaultCacheMaxAge  = 30 * 24 * time.Hour    // 30 days
	DefaultCacheMaxSize = 5 * 1024 * 1024 * 1024 // 5GB
	MetadataFileName    = "metadata.json"
)

// BinaryCacheEntry represents metadata for a cached binary
type BinaryCacheEntry struct {
	// Version of the Perl binary
	Version string `json:"version"`

	// Platform triple (e.g., "linux-amd64")
	Platform string `json:"platform"`

	// SHA-256 checksum of the binary file
	Checksum string `json:"checksum"`

	// Size of the binary file in bytes
	Size int64 `json:"size"`

	// Original download URL
	DownloadURL string `json:"download_url"`

	// When the binary was cached
	CachedAt time.Time `json:"cached_at"`

	// Last access time (for LRU cleanup)
	LastAccessed time.Time `json:"last_accessed"`
}

// BinaryCacheResult represents a cache hit result
type BinaryCacheResult struct {
	// Path to the cached binary file
	Path string

	// Cache entry metadata
	Entry *BinaryCacheEntry
}

// BinaryCache manages local caching of Perl binaries
type BinaryCache struct {
	// Root cache directory
	cacheDir string

	// Mutex for thread-safe operations
	mutex sync.RWMutex
}

// NewBinaryCache creates a new binary cache instance
func NewBinaryCache() (*BinaryCache, error) {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return nil, errors.NewSystemError(ErrCacheInitFailed,
			"Failed to determine XDG directories", err)
	}

	// Ensure directories exist
	err = dirs.EnsureDirs()
	if err != nil {
		return nil, errors.NewSystemError(ErrCacheInitFailed,
			"Failed to create required directories", err)
	}

	// Create binaries cache directory
	cacheDir := filepath.Join(dirs.CacheDir, "binaries")
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return nil, errors.NewSystemError(ErrCacheInitFailed,
			"Failed to create binaries cache directory", err)
	}

	return &BinaryCache{
		cacheDir: cacheDir,
	}, nil
}

// Get retrieves a binary from the cache if it exists
func (bc *BinaryCache) Get(version, platform string) (*BinaryCacheResult, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	// Get version/platform directory
	versionDir := filepath.Join(bc.cacheDir, version, platform)
	metadataPath := filepath.Join(versionDir, MetadataFileName)

	// Check if metadata exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		// Cache miss
		return nil, nil
	}

	// Read metadata
	metadataData, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, errors.NewSystemError(ErrCacheOperFailed,
			"Failed to read cache metadata", err).
			WithLocation(metadataPath)
	}

	var entry BinaryCacheEntry
	err = json.Unmarshal(metadataData, &entry)
	if err != nil {
		return nil, errors.NewSystemError(ErrCacheCorrupted,
			"Failed to parse cache metadata", err).
			WithLocation(metadataPath)
	}

	// Determine binary filename based on platform
	var binaryExt string
	if platform == "windows-amd64" {
		binaryExt = ".zip"
	} else {
		binaryExt = ".tar.gz"
	}
	binaryFilename := fmt.Sprintf("perl-%s-%s%s", version, platform, binaryExt)
	binaryPath := filepath.Join(versionDir, binaryFilename)

	// Check if binary file exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Binary file is missing, remove corrupt metadata
		_ = os.Remove(metadataPath)
		return nil, nil
	}

	// Update last accessed time
	entry.LastAccessed = time.Now()
	bc.updateMetadata(metadataPath, &entry)

	return &BinaryCacheResult{
		Path:  binaryPath,
		Entry: &entry,
	}, nil
}

// Put stores a binary in the cache
func (bc *BinaryCache) Put(sourcePath string, entry *BinaryCacheEntry) (string, error) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	// Create version/platform directory
	versionDir := filepath.Join(bc.cacheDir, entry.Version, entry.Platform)
	err := os.MkdirAll(versionDir, 0755)
	if err != nil {
		return "", errors.NewSystemError(ErrCacheOperFailed,
			"Failed to create cache directory", err).
			WithLocation(versionDir)
	}

	// Determine binary filename based on platform
	var binaryExt string
	if entry.Platform == "windows-amd64" {
		binaryExt = ".zip"
	} else {
		binaryExt = ".tar.gz"
	}
	binaryFilename := fmt.Sprintf("perl-%s-%s%s", entry.Version, entry.Platform, binaryExt)
	destPath := filepath.Join(versionDir, binaryFilename)

	// Copy source file to cache
	err = bc.copyFile(sourcePath, destPath)
	if err != nil {
		return "", errors.NewSystemError(ErrCacheOperFailed,
			"Failed to copy binary to cache", err).
			WithLocation(destPath)
	}

	// Update entry timestamps if not set
	if entry.CachedAt.IsZero() {
		entry.CachedAt = time.Now()
	}
	if entry.LastAccessed.IsZero() {
		entry.LastAccessed = time.Now()
	}

	// Write metadata
	metadataPath := filepath.Join(versionDir, MetadataFileName)
	err = bc.updateMetadata(metadataPath, entry)
	if err != nil {
		// Clean up binary file if metadata write fails
		_ = os.Remove(destPath)
		return "", err
	}

	return destPath, nil
}

// Validate checks if a cached binary file matches its expected checksum
func (bc *BinaryCache) Validate(binaryPath, expectedChecksum string) (bool, error) {
	// Calculate checksum of the cached file
	checksum, err := bc.calculateChecksum(binaryPath)
	if err != nil {
		return false, errors.NewSystemError(ErrCacheOperFailed,
			"Failed to calculate checksum for validation", err).
			WithLocation(binaryPath)
	}

	return checksum == expectedChecksum, nil
}

// List returns all cached binary entries
func (bc *BinaryCache) List() ([]*BinaryCacheEntry, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return bc.listEntries()
}

// listEntries is an internal helper that doesn't acquire locks
func (bc *BinaryCache) listEntries() ([]*BinaryCacheEntry, error) {
	var entries []*BinaryCacheEntry

	// Walk through cache directory structure
	err := filepath.Walk(bc.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for metadata files
		if info.Name() == MetadataFileName {
			// Read metadata
			metadataData, err := os.ReadFile(path)
			if err != nil {
				// Skip corrupted metadata files
				return nil
			}

			var entry BinaryCacheEntry
			err = json.Unmarshal(metadataData, &entry)
			if err != nil {
				// Skip corrupted metadata files
				return nil
			}

			entries = append(entries, &entry)
		}

		return nil
	})

	if err != nil {
		return nil, errors.NewSystemError(ErrCacheOperFailed,
			"Failed to list cache entries", err)
	}

	return entries, nil
}

// CleanByAge removes cache entries older than the specified duration
func (bc *BinaryCache) CleanByAge(maxAge time.Duration) (int, error) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	entries, err := bc.listEntries()
	if err != nil {
		return 0, err
	}

	cutoffTime := time.Now().Add(-maxAge)
	cleanedCount := 0

	for _, entry := range entries {
		if entry.CachedAt.Before(cutoffTime) {
			err := bc.removeEntry(entry)
			if err != nil {
				// Log error but continue cleaning other entries
				continue
			}
			cleanedCount++
		}
	}

	return cleanedCount, nil
}

// CleanBySize removes oldest entries to keep total cache size under the limit
func (bc *BinaryCache) CleanBySize(maxSize int64) (int, error) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	entries, err := bc.listEntries()
	if err != nil {
		return 0, err
	}

	// Calculate total size
	var totalSize int64
	for _, entry := range entries {
		totalSize += entry.Size
	}

	// If under the limit, no cleanup needed
	if totalSize <= maxSize {
		return 0, nil
	}

	// Sort entries by last accessed time (LRU - least recently used first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LastAccessed.Before(entries[j].LastAccessed)
	})

	cleanedCount := 0
	for _, entry := range entries {
		if totalSize <= maxSize {
			break
		}

		err := bc.removeEntry(entry)
		if err != nil {
			// Log error but continue cleaning other entries
			continue
		}

		totalSize -= entry.Size
		cleanedCount++
	}

	return cleanedCount, nil
}

// Clean performs both age-based and size-based cleanup
func (bc *BinaryCache) Clean(maxAge time.Duration, maxSize int64) (int, int, error) {
	// First clean by age
	cleanedByAge, err := bc.CleanByAge(maxAge)
	if err != nil {
		return 0, 0, err
	}

	// Then clean by size
	cleanedBySize, err := bc.CleanBySize(maxSize)
	if err != nil {
		return cleanedByAge, 0, err
	}

	return cleanedByAge, cleanedBySize, nil
}

// removeEntry removes a cache entry and its associated files
func (bc *BinaryCache) removeEntry(entry *BinaryCacheEntry) error {
	// Get version/platform directory
	versionDir := filepath.Join(bc.cacheDir, entry.Version, entry.Platform)

	// Remove the entire version/platform directory
	err := os.RemoveAll(versionDir)
	if err != nil {
		return errors.NewSystemError(ErrCacheOperFailed,
			"Failed to remove cache entry", err).
			WithLocation(versionDir)
	}

	// Try to remove parent version directory if it's empty
	parentDir := filepath.Join(bc.cacheDir, entry.Version)
	entries, err := os.ReadDir(parentDir)
	if err == nil && len(entries) == 0 {
		_ = os.Remove(parentDir)
	}

	return nil
}

// updateMetadata writes cache entry metadata to disk
func (bc *BinaryCache) updateMetadata(metadataPath string, entry *BinaryCacheEntry) error {
	metadataData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return errors.NewSystemError(ErrCacheOperFailed,
			"Failed to serialize cache metadata", err)
	}

	err = os.WriteFile(metadataPath, metadataData, 0644)
	if err != nil {
		return errors.NewSystemError(ErrCacheOperFailed,
			"Failed to write cache metadata", err).
			WithLocation(metadataPath)
	}

	return nil
}

// copyFile copies a file from source to destination
func (bc *BinaryCache) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// calculateChecksum calculates SHA-256 checksum of a file
func (bc *BinaryCache) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
