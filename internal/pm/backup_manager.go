// ABOUTME: Backup manager for cpanfile operations in PVI
// ABOUTME: Handles backup creation in different modes: off, local, and cache

package pm

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/xdg"
)

// BackupManager handles cpanfile backup operations
type BackupManager struct {
	config  *config.PMBackupConfig
	logger  *log.Logger
	xdgDirs *xdg.Dirs
}

// NewBackupManager creates a new backup manager with configuration
func NewBackupManager(backupConfig *config.PMBackupConfig, logger *log.Logger) (*BackupManager, error) {
	if backupConfig == nil {
		// Use default configuration if none provided
		backupConfig = &config.PMBackupConfig{
			CpanfileBackup: "off",
			RetentionDays:  30,
			MaxBackups:     10,
		}
	}

	xdgDirs, err := xdg.GetDirs()
	if err != nil {
		return nil, fmt.Errorf("failed to get XDG directories: %w", err)
	}

	return &BackupManager{
		config:  backupConfig,
		logger:  logger,
		xdgDirs: xdgDirs,
	}, nil
}

// BackupCpanfile creates a backup of the cpanfile if backup mode is enabled
func (bm *BackupManager) BackupCpanfile(cpanfilePath string) error {
	// Skip backup if mode is off
	if bm.config.CpanfileBackup == "off" {
		return nil
	}

	// Check if file exists
	if _, err := os.Stat(cpanfilePath); os.IsNotExist(err) {
		// No file to backup, this is normal for new cpanfiles
		return nil
	}

	switch bm.config.CpanfileBackup {
	case "local":
		return bm.createLocalBackup(cpanfilePath)
	case "cache":
		return bm.createCacheBackup(cpanfilePath)
	default:
		return fmt.Errorf("unsupported backup mode: %s", bm.config.CpanfileBackup)
	}
}

// createLocalBackup creates a backup in the same directory as the cpanfile
func (bm *BackupManager) createLocalBackup(cpanfilePath string) error {
	timestamp := time.Now().Format("20060102150405.000")
	backupPath := cpanfilePath + ".backup." + timestamp

	if err := copyFile(cpanfilePath, backupPath); err != nil {
		return fmt.Errorf("failed to create local backup: %w", err)
	}

	bm.logger.Printf("[PVI] Created local backup: %s", backupPath)

	// Clean up old backups
	return bm.cleanupLocalBackups(cpanfilePath)
}

// createCacheBackup creates a backup in the XDG cache directory
func (bm *BackupManager) createCacheBackup(cpanfilePath string) error {
	// Ensure cache directories exist
	if err := bm.xdgDirs.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create cache directories: %w", err)
	}

	// Create project-specific cache directory
	projectHash := bm.getProjectHash(cpanfilePath)
	cacheDir := filepath.Join(bm.xdgDirs.CacheDir, "cpanfile-backups", projectHash)

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache backup directory: %w", err)
	}

	// Create backup with timestamp
	timestamp := time.Now().Format("20060102150405.000")
	backupPath := filepath.Join(cacheDir, "cpanfile.backup."+timestamp)

	if err := copyFile(cpanfilePath, backupPath); err != nil {
		return fmt.Errorf("failed to create cache backup: %w", err)
	}

	bm.logger.Printf("[PVI] Created cache backup: %s", backupPath)

	// Clean up old backups
	return bm.cleanupCacheBackups(cacheDir)
}

// getProjectHash generates a hash for the project directory to avoid conflicts
func (bm *BackupManager) getProjectHash(cpanfilePath string) string {
	projectDir := filepath.Dir(filepath.Clean(cpanfilePath))
	hash := sha256.Sum256([]byte(projectDir))
	return fmt.Sprintf("%x", hash)[:12] // Use first 12 characters of hash
}

// cleanupLocalBackups removes old local backup files
func (bm *BackupManager) cleanupLocalBackups(cpanfilePath string) error {
	dir := filepath.Dir(cpanfilePath)
	baseName := filepath.Base(cpanfilePath)
	pattern := baseName + ".backup.*"

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory for cleanup: %w", err)
	}

	var backups []backupInfo
	for _, entry := range entries {
		if matched, _ := filepath.Match(pattern, entry.Name()); matched {
			fullPath := filepath.Join(dir, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue // Skip files we can't stat
			}
			backups = append(backups, backupInfo{
				path:    fullPath,
				modTime: info.ModTime(),
			})
		}
	}

	return bm.cleanupBackupList(backups)
}

// cleanupCacheBackups removes old cache backup files
func (bm *BackupManager) cleanupCacheBackups(cacheDir string) error {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory for cleanup: %w", err)
	}

	var backups []backupInfo
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "cpanfile.backup.") {
			fullPath := filepath.Join(cacheDir, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue // Skip files we can't stat
			}
			backups = append(backups, backupInfo{
				path:    fullPath,
				modTime: info.ModTime(),
			})
		}
	}

	return bm.cleanupBackupList(backups)
}

// backupInfo holds information about a backup file for cleanup purposes
type backupInfo struct {
	path    string
	modTime time.Time
}

// cleanupBackupList removes old backups based on retention policy
func (bm *BackupManager) cleanupBackupList(backups []backupInfo) error {
	if len(backups) == 0 {
		return nil
	}

	// Sort by modification time (oldest first)
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			if backups[i].modTime.After(backups[j].modTime) {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	// Remove backups older than retention days
	cutoffTime := time.Now().AddDate(0, 0, -bm.config.RetentionDays)
	var validBackups []backupInfo

	for _, backup := range backups {
		if backup.modTime.Before(cutoffTime) {
			if err := os.Remove(backup.path); err != nil {
				bm.logger.Printf("[PVI] Warning: failed to remove old backup %s: %v", backup.path, err)
			} else {
				bm.logger.Printf("[PVI] Removed old backup: %s", backup.path)
			}
		} else {
			validBackups = append(validBackups, backup)
		}
	}

	// Remove excess backups beyond MaxBackups
	if len(validBackups) > bm.config.MaxBackups {
		excess := len(validBackups) - bm.config.MaxBackups
		for i := 0; i < excess; i++ {
			if err := os.Remove(validBackups[i].path); err != nil {
				bm.logger.Printf("[PVI] Warning: failed to remove excess backup %s: %v", validBackups[i].path, err)
			} else {
				bm.logger.Printf("[PVI] Removed excess backup: %s", validBackups[i].path)
			}
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// GetBackupMode returns the current backup mode
func (bm *BackupManager) GetBackupMode() string {
	return bm.config.CpanfileBackup
}

// SetBackupMode temporarily overrides the backup mode (for CLI flag support)
func (bm *BackupManager) SetBackupMode(mode string) error {
	validModes := map[string]bool{
		"off":   true,
		"local": true,
		"cache": true,
	}

	if !validModes[mode] {
		return fmt.Errorf("invalid backup mode: %s (must be off, local, or cache)", mode)
	}

	bm.config.CpanfileBackup = mode
	return nil
}
