// ABOUTME: Backup system for safe file modifications during auto-fix operations
// ABOUTME: Provides timestamped backups, rollback functionality, and cleanup management

package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/xdg"
)

// Backup represents a single file backup
type Backup struct {
	OriginalPath string    `json:"original_path"`
	BackupPath   string    `json:"backup_path"`
	Timestamp    time.Time `json:"timestamp"`
	Checksum     string    `json:"checksum,omitempty"`
}

// BackupSession represents a collection of backups created during a single operation
type BackupSession struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Operation string            `json:"operation"`
	Backups   []Backup          `json:"backups"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Manager handles backup operations
type Manager struct {
	backupDir string
}

// NewManager creates a new backup manager
func NewManager() (*Manager, error) {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return nil, errors.NewSystemError("001", "Failed to determine XDG directories", err)
	}

	backupDir := filepath.Join(dirs.DataDir, "backups")
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		return nil, errors.NewSystemError("002", "Failed to create backup directory", err).
			WithLocation(backupDir)
	}

	return &Manager{
		backupDir: backupDir,
	}, nil
}

// CreateSession creates a new backup session for a specific operation
func (m *Manager) CreateSession(operation string) *BackupSession {
	sessionID := fmt.Sprintf("%d-%s", time.Now().Unix(), operation)
	return &BackupSession{
		ID:        sessionID,
		Timestamp: time.Now(),
		Operation: operation,
		Backups:   make([]Backup, 0),
		Metadata:  make(map[string]string),
	}
}

// BackupFile creates a backup of a file within a session
func (m *Manager) BackupFile(session *BackupSession, filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, no backup needed
		return nil
	}

	// Generate backup filename
	timestamp := time.Now().Format("20060102-150405")
	filename := filepath.Base(filePath)
	backupName := fmt.Sprintf("%s.backup-%s-%s", filename, session.ID, timestamp)
	backupPath := filepath.Join(m.backupDir, backupName)

	// Copy file to backup location
	err := m.copyFile(filePath, backupPath)
	if err != nil {
		return errors.NewSystemError("003", "Failed to create backup", err).
			WithLocation(filePath).
			WithHint("Ensure you have write permissions to backup directory")
	}

	// Calculate checksum for integrity verification
	checksum, err := m.calculateChecksum(filePath)
	if err != nil {
		// Checksum is optional, log but don't fail
		checksum = ""
	}

	// Add to session
	backup := Backup{
		OriginalPath: filePath,
		BackupPath:   backupPath,
		Timestamp:    time.Now(),
		Checksum:     checksum,
	}

	session.Backups = append(session.Backups, backup)
	return nil
}

// RestoreFile restores a file from backup
func (m *Manager) RestoreFile(backup Backup) error {
	// Check if backup exists
	if _, err := os.Stat(backup.BackupPath); os.IsNotExist(err) {
		return errors.NewSystemError("004", "Backup file not found", err).
			WithLocation(backup.BackupPath)
	}

	// Verify backup integrity if checksum is available
	if backup.Checksum != "" {
		currentChecksum, err := m.calculateChecksum(backup.BackupPath)
		if err == nil && currentChecksum != backup.Checksum {
			return errors.NewSystemError("005", "Backup file integrity check failed", nil).
				WithLocation(backup.BackupPath).
				WithHint("Backup may be corrupted")
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(backup.OriginalPath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return errors.NewSystemError("006", "Failed to create directory for restore", err).
			WithLocation(dir)
	}

	// Copy backup to original location
	err = m.copyFile(backup.BackupPath, backup.OriginalPath)
	if err != nil {
		return errors.NewSystemError("007", "Failed to restore file from backup", err).
			WithLocation(backup.OriginalPath)
	}

	return nil
}

// RestoreSession restores all files from a backup session
func (m *Manager) RestoreSession(session *BackupSession) error {
	var errors []error

	// Restore files in reverse order (last backed up, first restored)
	for i := len(session.Backups) - 1; i >= 0; i-- {
		backup := session.Backups[i]
		if err := m.RestoreFile(backup); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to restore %d files: %v", len(errors), errors)
	}

	return nil
}

// ListSessions returns all backup sessions, newest first
func (m *Manager) ListSessions() ([]*BackupSession, error) {
	files, err := os.ReadDir(m.backupDir)
	if err != nil {
		return nil, errors.NewSystemError("008", "Failed to list backup directory", err).
			WithLocation(m.backupDir)
	}

	// Parse session information from backup files
	sessionMap := make(map[string]*BackupSession)

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".backup-") && !strings.Contains(file.Name(), ".backup-") {
			continue
		}

		// Parse backup filename: original.backup-sessionid-timestamp
		parts := strings.Split(file.Name(), ".backup-")
		if len(parts) < 2 {
			continue
		}

		sessionParts := strings.Split(parts[1], "-")
		if len(sessionParts) < 2 {
			continue
		}

		// Extract session ID (timestamp-operation)
		sessionID := sessionParts[0] + "-" + sessionParts[1]

		session, exists := sessionMap[sessionID]
		if !exists {
			// Parse timestamp from session ID
			timestampStr := sessionParts[0]
			timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				continue
			}

			operation := sessionParts[1]
			session = &BackupSession{
				ID:        sessionID,
				Timestamp: time.Unix(timestamp, 0),
				Operation: operation,
				Backups:   make([]Backup, 0),
			}
			sessionMap[sessionID] = session
		}

		// Add backup to session
		backupPath := filepath.Join(m.backupDir, file.Name())
		backup := Backup{
			BackupPath: backupPath,
			Timestamp:  session.Timestamp,
		}
		session.Backups = append(session.Backups, backup)
	}

	// Convert to slice and sort by timestamp (newest first)
	sessions := make([]*BackupSession, 0, len(sessionMap))
	for _, session := range sessionMap {
		sessions = append(sessions, session)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Timestamp.After(sessions[j].Timestamp)
	})

	return sessions, nil
}

// CleanupOldBackups removes backups older than the specified duration
func (m *Manager) CleanupOldBackups(olderThan time.Duration) error {
	sessions, err := m.ListSessions()
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-olderThan)
	var removedCount int

	for _, session := range sessions {
		if session.Timestamp.Before(cutoff) {
			for _, backup := range session.Backups {
				if err := os.Remove(backup.BackupPath); err == nil {
					removedCount++
				}
			}
		}
	}

	if removedCount > 0 {
		fmt.Printf("Cleaned up %d old backup files\n", removedCount)
	}

	return nil
}

// GetLatestSession returns the most recent backup session
func (m *Manager) GetLatestSession() (*BackupSession, error) {
	sessions, err := m.ListSessions()
	if err != nil {
		return nil, err
	}

	if len(sessions) == 0 {
		return nil, fmt.Errorf("no backup sessions found")
	}

	return sessions[0], nil
}

// Helper methods

// copyFile copies a file from src to dst
func (m *Manager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// calculateChecksum calculates a simple checksum for file integrity
func (m *Manager) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	// Simple checksum based on size and modification time
	return fmt.Sprintf("%d-%d", info.Size(), info.ModTime().Unix()), nil
}
