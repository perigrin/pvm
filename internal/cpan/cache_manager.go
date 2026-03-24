// ABOUTME: Cache management functionality for CPAN operations
// ABOUTME: Provides cache validation, cleanup, and statistics operations

package cpan

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"tamarou.com/pvm/internal/errors"
)

// CacheManager provides cache management operations
type CacheManager struct {
	cacheDir string
	logger   *log.Logger
}

// CacheStats contains statistics about cache usage
type CacheStats struct {
	TotalFiles     int           `json:"total_files"`
	TotalSize      int64         `json:"total_size"`
	ExpiredFiles   int           `json:"expired_files"`
	ExpiredSize    int64         `json:"expired_size"`
	ValidFiles     int           `json:"valid_files"`
	ValidSize      int64         `json:"valid_size"`
	OldestEntry    time.Time     `json:"oldest_entry"`
	NewestEntry    time.Time     `json:"newest_entry"`
	AverageAge     time.Duration `json:"average_age"`
	CacheDirectory string        `json:"cache_directory"`
}

// NewCacheManager creates a new cache manager instance
func NewCacheManager(cacheDir string, logger *log.Logger) (*CacheManager, error) {
	if logger == nil {
		logger = log.New(os.Stderr, "[CacheManager] ", log.LstdFlags)
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, errors.NewSystemError("201", "Failed to create cache directory", err)
	}

	return &CacheManager{
		cacheDir: cacheDir,
		logger:   logger,
	}, nil
}

// ValidateCache checks the integrity and health of the cache
func (cm *CacheManager) ValidateCache() error {
	cm.logger.Printf("Validating cache directory: %s", cm.cacheDir)

	// Check if cache directory exists and is accessible
	info, err := os.Stat(cm.cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewSystemError("202", "Cache directory does not exist", err)
		}
		return errors.NewSystemError("203", "Cannot access cache directory", err)
	}

	if !info.IsDir() {
		return errors.NewSystemError("204", "Cache path is not a directory", nil)
	}

	// Check directory permissions
	if info.Mode().Perm()&0200 == 0 {
		return errors.NewSystemError("205", "Cache directory is not writable", nil)
	}

	// Validate cache files
	files, err := os.ReadDir(cm.cacheDir)
	if err != nil {
		return errors.NewSystemError("206", "Failed to read cache directory", err)
	}

	invalidFiles := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(cm.cacheDir, file.Name())
		if err := cm.validateCacheFile(filePath); err != nil {
			cm.logger.Printf("Invalid cache file %s: %v", file.Name(), err)
			invalidFiles++
		}
	}

	if invalidFiles > 0 {
		cm.logger.Printf("Found %d invalid cache files", invalidFiles)
	}

	cm.logger.Printf("Cache validation completed")
	return nil
}

// CleanupCache removes expired cache entries and optionally old entries
func (cm *CacheManager) CleanupCache(olderThan time.Duration) error {
	cm.logger.Printf("Starting cache cleanup (removing entries older than %v)", olderThan)

	files, err := os.ReadDir(cm.cacheDir)
	if err != nil {
		return errors.NewSystemError("207", "Failed to read cache directory for cleanup", err)
	}

	removedCount := 0
	totalSize := int64(0)

	cutoffTime := time.Now().Add(-olderThan)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(cm.cacheDir, file.Name())
		info, err := file.Info()
		if err != nil {
			cm.logger.Printf("Error getting file info for %s: %v", file.Name(), err)
			continue
		}

		// Check if file is older than cutoff or expired based on content
		shouldRemove := info.ModTime().Before(cutoffTime)

		// Also check if cache entry has expired based on metadata
		if !shouldRemove {
			if expired, err := cm.isCacheFileExpired(filePath); err == nil && expired {
				shouldRemove = true
			}
		}

		if shouldRemove {
			if err := os.Remove(filePath); err != nil {
				cm.logger.Printf("Error removing cache file %s: %v", file.Name(), err)
				continue
			}
			removedCount++
			totalSize += info.Size()
			cm.logger.Printf("Removed cache file: %s", file.Name())
		}
	}

	cm.logger.Printf("Cache cleanup completed: removed %d files (%d bytes)", removedCount, totalSize)
	return nil
}

// GetCacheStats returns detailed statistics about the cache
func (cm *CacheManager) GetCacheStats() (*CacheStats, error) {
	files, err := os.ReadDir(cm.cacheDir)
	if err != nil {
		return nil, errors.NewSystemError("208", "Failed to read cache directory for stats", err)
	}

	stats := &CacheStats{
		CacheDirectory: cm.cacheDir,
		OldestEntry:    time.Now(),
		NewestEntry:    time.Time{},
	}

	totalAge := time.Duration(0)
	now := time.Now()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(cm.cacheDir, file.Name())
		info, err := file.Info()
		if err != nil {
			continue
		}

		stats.TotalFiles++
		stats.TotalSize += info.Size()

		// Track age statistics
		age := now.Sub(info.ModTime())
		totalAge += age

		if info.ModTime().Before(stats.OldestEntry) {
			stats.OldestEntry = info.ModTime()
		}
		if info.ModTime().After(stats.NewestEntry) {
			stats.NewestEntry = info.ModTime()
		}

		// Check if file is expired
		if expired, err := cm.isCacheFileExpired(filePath); err == nil && expired {
			stats.ExpiredFiles++
			stats.ExpiredSize += info.Size()
		} else {
			stats.ValidFiles++
			stats.ValidSize += info.Size()
		}
	}

	if stats.TotalFiles > 0 {
		stats.AverageAge = totalAge / time.Duration(stats.TotalFiles)
	}

	return stats, nil
}

// PurgeExpiredEntries removes all expired cache entries
func (cm *CacheManager) PurgeExpiredEntries() error {
	cm.logger.Printf("Purging expired cache entries")

	files, err := os.ReadDir(cm.cacheDir)
	if err != nil {
		return errors.NewSystemError("209", "Failed to read cache directory for purging", err)
	}

	removedCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(cm.cacheDir, file.Name())
		if expired, err := cm.isCacheFileExpired(filePath); err == nil && expired {
			if err := os.Remove(filePath); err != nil {
				cm.logger.Printf("Error removing expired cache file %s: %v", file.Name(), err)
				continue
			}
			removedCount++
			cm.logger.Printf("Removed expired cache file: %s", file.Name())
		}
	}

	cm.logger.Printf("Purged %d expired cache entries", removedCount)
	return nil
}

// OptimizeCache performs cache optimization operations
func (cm *CacheManager) OptimizeCache() error {
	cm.logger.Printf("Starting cache optimization")

	// First, remove expired entries
	if err := cm.PurgeExpiredEntries(); err != nil {
		return err
	}

	// Validate remaining entries
	if err := cm.ValidateCache(); err != nil {
		return err
	}

	cm.logger.Printf("Cache optimization completed")
	return nil
}

// validateCacheFile checks if a cache file is valid JSON and has required structure
func (cm *CacheManager) validateCacheFile(filePath string) error {
	// Try to parse as cache entry to validate structure

	// Get the cache key from the filename (reverse the MD5 hash process isn't practical,
	// so we'll just check if the file can be read as a valid cache entry)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	// Try to parse as JSON
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	// Validate required fields
	if entry.Metadata.Key == "" {
		return fmt.Errorf("missing cache key")
	}
	if entry.Metadata.Timestamp.IsZero() {
		return fmt.Errorf("missing timestamp")
	}

	return nil
}

// isCacheFileExpired checks if a cache file has expired based on its metadata
func (cm *CacheManager) isCacheFileExpired(filePath string) (bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return false, err
	}

	// Use !Before instead of After so that entries expiring at exactly now
	// are considered expired (matters on Windows where timer resolution is coarse)
	return !time.Now().Before(entry.Metadata.Expires), nil
}
