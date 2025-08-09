package pvi

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/config"
)

// mockWriter implements io.Writer and captures log messages for testing
type mockWriter struct {
	logs []string
}

func (w *mockWriter) Write(p []byte) (n int, err error) {
	w.logs = append(w.logs, string(p))
	return len(p), nil
}

func TestNewBackupManager(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.PVIBackupConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &config.PVIBackupConfig{
				CpanfileBackup: "local",
				RetentionDays:  30,
				MaxBackups:     10,
			},
			expectError: false,
		},
		{
			name:        "nil config uses defaults",
			config:      nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := &mockWriter{}
			logger := log.New(mockWriter, "", 0)
			bm, err := NewBackupManager(tt.config, logger)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if bm == nil {
				t.Error("Expected backup manager but got nil")
			}

			// Test default config when nil is provided
			if tt.config == nil {
				if bm.GetBackupMode() != "off" {
					t.Errorf("Expected default backup mode 'off', got %s", bm.GetBackupMode())
				}
			}
		})
	}
}

func TestBackupManager_SetGetBackupMode(t *testing.T) {
	mockWriter := &mockWriter{}
	logger := log.New(mockWriter, "", 0)
	bm, err := NewBackupManager(nil, logger)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	tests := []struct {
		mode        string
		expectError bool
	}{
		{"off", false},
		{"local", false},
		{"cache", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			err := bm.SetBackupMode(tt.mode)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if bm.GetBackupMode() != tt.mode {
				t.Errorf("Expected mode %s, got %s", tt.mode, bm.GetBackupMode())
			}
		})
	}
}

func TestBackupManager_BackupCpanfile_Off(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")

	// Create test cpanfile
	content := "requires 'Test::More';\n"
	if err := os.WriteFile(cpanfilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test cpanfile: %v", err)
	}

	config := &config.PVIBackupConfig{
		CpanfileBackup: "off",
		RetentionDays:  30,
		MaxBackups:     10,
	}

	mockWriter := &mockWriter{}
	logger := log.New(mockWriter, "", 0)
	bm, err := NewBackupManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	// Backup should be skipped
	err = bm.BackupCpanfile(cpanfilePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check no backup was created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "cpanfile.backup.") {
			t.Errorf("Backup file should not be created in 'off' mode, but found: %s", file.Name())
		}
	}
}

func TestBackupManager_BackupCpanfile_Local(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")

	// Create test cpanfile
	content := "requires 'Test::More';\n"
	if err := os.WriteFile(cpanfilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test cpanfile: %v", err)
	}

	config := &config.PVIBackupConfig{
		CpanfileBackup: "local",
		RetentionDays:  30,
		MaxBackups:     10,
	}

	mockWriter := &mockWriter{}
	logger := log.New(mockWriter, "", 0)
	bm, err := NewBackupManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	// Create backup
	err = bm.BackupCpanfile(cpanfilePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check backup was created in local directory
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	backupFound := false
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "cpanfile.backup.") {
			backupFound = true

			// Verify backup content matches original
			backupPath := filepath.Join(tempDir, file.Name())
			backupContent, err := os.ReadFile(backupPath)
			if err != nil {
				t.Errorf("Failed to read backup file: %v", err)
			} else if string(backupContent) != content {
				t.Errorf("Backup content doesn't match original. Expected: %s, Got: %s", content, string(backupContent))
			}
			break
		}
	}

	if !backupFound {
		t.Error("Expected backup file to be created in local mode")
	}

	// Verify logger was called
	if len(mockWriter.logs) == 0 {
		t.Error("Expected logger to be used")
	}
}

func TestBackupManager_BackupCpanfile_Cache(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "project", "cpanfile")

	// Create project directory and cpanfile
	if err := os.MkdirAll(filepath.Dir(cpanfilePath), 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	content := "requires 'Data::Dumper';\n"
	if err := os.WriteFile(cpanfilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test cpanfile: %v", err)
	}

	config := &config.PVIBackupConfig{
		CpanfileBackup: "cache",
		RetentionDays:  30,
		MaxBackups:     10,
	}

	mockWriter := &mockWriter{}
	logger := log.New(mockWriter, "", 0)
	bm, err := NewBackupManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	// Create backup
	err = bm.BackupCpanfile(cpanfilePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify no backup in project directory
	projectFiles, err := os.ReadDir(filepath.Dir(cpanfilePath))
	if err != nil {
		t.Fatalf("Failed to read project directory: %v", err)
	}

	for _, file := range projectFiles {
		if strings.HasPrefix(file.Name(), "cpanfile.backup.") {
			t.Errorf("Backup file should not be in project directory for cache mode, but found: %s", file.Name())
		}
	}

	// Verify logger was called for cache creation
	if len(mockWriter.logs) == 0 {
		t.Error("Expected logger to be used")
	}

	foundCacheMessage := false
	for _, logMsg := range mockWriter.logs {
		if strings.Contains(logMsg, "Created cache backup") {
			foundCacheMessage = true
			break
		}
	}

	if !foundCacheMessage {
		t.Error("Expected cache backup creation log message")
	}
}

func TestBackupManager_BackupCpanfile_NonexistentFile(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "nonexistent.cpanfile")

	config := &config.PVIBackupConfig{
		CpanfileBackup: "local",
		RetentionDays:  30,
		MaxBackups:     10,
	}

	mockWriter := &mockWriter{}
	logger := log.New(mockWriter, "", 0)
	bm, err := NewBackupManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	// Should not error for nonexistent file
	err = bm.BackupCpanfile(cpanfilePath)
	if err != nil {
		t.Errorf("Unexpected error for nonexistent file: %v", err)
	}
}

func TestBackupManager_BackupCleanup_MaxBackups(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")

	// Create test cpanfile
	content := "requires 'Test::More';\n"
	if err := os.WriteFile(cpanfilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test cpanfile: %v", err)
	}

	config := &config.PVIBackupConfig{
		CpanfileBackup: "local",
		RetentionDays:  365, // Long retention to test max backups limit
		MaxBackups:     3,   // Limit to 3 backups
	}

	mockWriter := &mockWriter{}
	logger := log.New(mockWriter, "", 0)
	bm, err := NewBackupManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	// Create 5 backups
	for i := 0; i < 5; i++ {
		err = bm.BackupCpanfile(cpanfilePath)
		if err != nil {
			t.Errorf("Failed to create backup %d: %v", i+1, err)
		}

		// Sleep briefly to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Count backup files
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	backupCount := 0
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "cpanfile.backup.") {
			backupCount++
		}
	}

	if backupCount > config.MaxBackups {
		t.Errorf("Expected at most %d backups, found %d", config.MaxBackups, backupCount)
	}

	// Verify cleanup log messages
	hasCleanupLog := false
	for _, logMsg := range mockWriter.logs {
		if strings.Contains(logMsg, "Removed excess backup") {
			hasCleanupLog = true
			break
		}
	}

	if !hasCleanupLog {
		t.Error("Expected cleanup log message for excess backups")
	}
}

func TestBackupManager_BackupCleanup_RetentionDays(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")

	// Create test cpanfile
	content := "requires 'Test::More';\n"
	if err := os.WriteFile(cpanfilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test cpanfile: %v", err)
	}

	// Create an old backup file manually
	oldBackupPath := filepath.Join(tempDir, "cpanfile.backup.20200101120000")
	if err := os.WriteFile(oldBackupPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create old backup: %v", err)
	}

	// Set the modification time to be old
	oldTime := time.Now().AddDate(0, 0, -40) // 40 days ago
	if err := os.Chtimes(oldBackupPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old backup time: %v", err)
	}

	config := &config.PVIBackupConfig{
		CpanfileBackup: "local",
		RetentionDays:  30, // 30 day retention
		MaxBackups:     10,
	}

	mockWriter := &mockWriter{}
	logger := log.New(mockWriter, "", 0)
	bm, err := NewBackupManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	// Create a new backup, which should trigger cleanup
	err = bm.BackupCpanfile(cpanfilePath)
	if err != nil {
		t.Errorf("Failed to create backup: %v", err)
	}

	// Verify old backup was removed
	if _, err := os.Stat(oldBackupPath); !os.IsNotExist(err) {
		t.Error("Expected old backup to be removed due to retention policy")
	}

	// Verify cleanup log message
	hasRetentionLog := false
	for _, logMsg := range mockWriter.logs {
		if strings.Contains(logMsg, "Removed old backup") {
			hasRetentionLog = true
			break
		}
	}

	if !hasRetentionLog {
		t.Error("Expected retention cleanup log message")
	}
}

func TestBackupManager_ProjectHash(t *testing.T) {
	config := &config.PVIBackupConfig{
		CpanfileBackup: "cache",
		RetentionDays:  30,
		MaxBackups:     10,
	}

	mockWriter := &mockWriter{}
	logger := log.New(mockWriter, "", 0)
	bm, err := NewBackupManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	// Test that different paths produce different hashes
	path1 := "/home/user/project1/cpanfile"
	path2 := "/home/user/project2/cpanfile"

	hash1 := bm.getProjectHash(path1)
	hash2 := bm.getProjectHash(path2)

	if hash1 == hash2 {
		t.Error("Expected different hashes for different project paths")
	}

	if len(hash1) != 12 || len(hash2) != 12 {
		t.Error("Expected hash to be 12 characters long")
	}

	// Test that same path produces same hash
	hash1Again := bm.getProjectHash(path1)
	if hash1 != hash1Again {
		t.Error("Expected same hash for same project path")
	}
}
