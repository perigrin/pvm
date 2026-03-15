// ABOUTME: Backup creation and management for PVM binary updates
// ABOUTME: Handles backup creation, validation, and cleanup with proper metadata

package updater

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// BackupManager handles backup operations for binary updates
type BackupManager struct {
	backupDir string
}

// NewBackupManager creates a new backup manager
func NewBackupManager() *BackupManager {
	return &BackupManager{}
}

// NewBackupManagerWithDir creates a backup manager with a specific backup directory
func NewBackupManagerWithDir(backupDir string) *BackupManager {
	return &BackupManager{
		backupDir: backupDir,
	}
}

// BackupMetadata contains information about a backup
type BackupMetadata struct {
	OriginalPath string    `json:"original_path"`
	BackupPath   string    `json:"backup_path"`
	CreatedAt    time.Time `json:"created_at"`
	OriginalSize int64     `json:"original_size"`
	Checksum     string    `json:"checksum"`
	Version      string    `json:"version,omitempty"`
}

// CreateBackup creates a backup of the specified file
func (bm *BackupManager) CreateBackup(filePath string) (string, error) {
	// Ensure backup directory exists
	backupDir, err := bm.getBackupDir()
	if err != nil {
		return "", fmt.Errorf("getting backup directory: %w", err)
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("creating backup directory: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("getting file info: %w", err)
	}

	// Generate backup filename with timestamp
	fileName := filepath.Base(filePath)
	timestamp := time.Now().Format("20060102-150405")
	backupFileName := fmt.Sprintf("%s.backup.%s", fileName, timestamp)
	backupPath := filepath.Join(backupDir, backupFileName)

	// Copy file to backup location
	if err := bm.copyFile(filePath, backupPath); err != nil {
		return "", fmt.Errorf("copying file to backup: %w", err)
	}

	// Calculate checksum for verification
	checksum, err := bm.calculateChecksum(filePath)
	if err != nil {
		// Don't fail backup creation if checksum fails
		checksum = ""
	}

	// Create metadata file
	metadata := &BackupMetadata{
		OriginalPath: filePath,
		BackupPath:   backupPath,
		CreatedAt:    time.Now(),
		OriginalSize: fileInfo.Size(),
		Checksum:     checksum,
	}

	metadataPath := backupPath + ".meta"
	if err := bm.saveMetadata(metadata, metadataPath); err != nil {
		// Don't fail backup creation if metadata save fails
		_ = err
	}

	return backupPath, nil
}

// ValidateBackup validates a backup against its metadata
func (bm *BackupManager) ValidateBackup(backupPath string) error {
	// Check if backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupPath)
	}

	// Load metadata
	metadataPath := backupPath + ".meta"
	metadata, err := bm.loadMetadata(metadataPath)
	if err != nil {
		// If no metadata, do basic validation
		return bm.basicBackupValidation(backupPath)
	}

	// Validate file size
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		return fmt.Errorf("getting backup file info: %w", err)
	}

	if fileInfo.Size() != metadata.OriginalSize {
		return fmt.Errorf("backup file size mismatch: expected %d, got %d",
			metadata.OriginalSize, fileInfo.Size())
	}

	// Validate checksum if available
	if metadata.Checksum != "" {
		actualChecksum, err := bm.calculateChecksum(backupPath)
		if err != nil {
			return fmt.Errorf("calculating backup checksum: %w", err)
		}

		if actualChecksum != metadata.Checksum {
			return fmt.Errorf("backup checksum mismatch")
		}
	}

	return nil
}

// ListBackups lists all available backups
func (bm *BackupManager) ListBackups() ([]*BackupMetadata, error) {
	backupDir, err := bm.getBackupDir()
	if err != nil {
		return nil, fmt.Errorf("getting backup directory: %w", err)
	}

	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return []*BackupMetadata{}, nil
	}

	// Read directory entries
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("reading backup directory: %w", err)
	}

	var backups []*BackupMetadata

	for _, entry := range entries {
		if entry.IsDir() || !entry.Type().IsRegular() {
			continue
		}

		// Skip metadata files
		if filepath.Ext(entry.Name()) == ".meta" {
			continue
		}

		// Try to load metadata for this backup
		backupPath := filepath.Join(backupDir, entry.Name())
		metadataPath := backupPath + ".meta"

		metadata, err := bm.loadMetadata(metadataPath)
		if err != nil {
			// Create basic metadata if none exists
			info, err := entry.Info()
			if err != nil {
				continue
			}

			metadata = &BackupMetadata{
				BackupPath:   backupPath,
				CreatedAt:    info.ModTime(),
				OriginalSize: info.Size(),
			}
		}

		backups = append(backups, metadata)
	}

	return backups, nil
}

// CleanupOldBackups removes backups older than the specified duration
func (bm *BackupManager) CleanupOldBackups(maxAge time.Duration) error {
	backups, err := bm.ListBackups()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}

	cutoff := time.Now().Add(-maxAge)
	var removedCount int

	for _, backup := range backups {
		if backup.CreatedAt.Before(cutoff) {
			if err := bm.RemoveBackup(backup.BackupPath); err != nil {
				// Log error but continue with other backups
				continue
			}
			removedCount++
		}
	}

	return nil
}

// RemoveBackup removes a specific backup and its metadata
func (bm *BackupManager) RemoveBackup(backupPath string) error {
	// Remove backup file
	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing backup file: %w", err)
	}

	// Remove metadata file
	metadataPath := backupPath + ".meta"
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		// Don't fail if metadata file doesn't exist
		_ = err
	}

	return nil
}

// RestoreFromBackup restores a file from backup
func (bm *BackupManager) RestoreFromBackup(backupPath, targetPath string) error {
	// Validate backup first
	if err := bm.ValidateBackup(backupPath); err != nil {
		return fmt.Errorf("backup validation failed: %w", err)
	}

	// Copy backup to target location
	if err := bm.copyFile(backupPath, targetPath); err != nil {
		return fmt.Errorf("restoring from backup: %w", err)
	}

	return nil
}

// getBackupDir returns the backup directory path
func (bm *BackupManager) getBackupDir() (string, error) {
	if bm.backupDir != "" {
		return bm.backupDir, nil
	}

	// Use XDG cache directory for backups
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		// Fallback to temp directory
		return filepath.Join(os.TempDir(), "pvm-backups"), nil
	}

	return filepath.Join(cacheDir, "pvm", "backups"), nil
}

// copyFile copies a file from src to dst
func (bm *BackupManager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer srcFile.Close()

	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy file contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("copying file contents: %w", err)
	}

	// Copy file permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("getting source file info: %w", err)
	}

	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("setting file permissions: %w", err)
	}

	return nil
}

// calculateChecksum calculates SHA256 checksum of a file
func (bm *BackupManager) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("calculating hash: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// saveMetadata saves backup metadata to a file
func (bm *BackupManager) saveMetadata(metadata *BackupMetadata, path string) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling metadata: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing metadata file: %w", err)
	}

	return nil
}

// loadMetadata loads backup metadata from a file
func (bm *BackupManager) loadMetadata(path string) (*BackupMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading metadata file: %w", err)
	}

	var metadata BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("unmarshaling metadata: %w", err)
	}

	return &metadata, nil
}

// basicBackupValidation performs basic validation when no metadata is available
func (bm *BackupManager) basicBackupValidation(backupPath string) error {
	// Check if file exists and is regular
	info, err := os.Stat(backupPath)
	if err != nil {
		return fmt.Errorf("getting backup file info: %w", err)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("backup is not a regular file")
	}

	// Check minimum size
	if info.Size() < 1024 {
		return fmt.Errorf("backup file too small: %d bytes", info.Size())
	}

	return nil
}
