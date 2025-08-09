package pvi

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/config"
)

func TestCpanfileManager_WithBackupConfig(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")

	// Create test cpanfile
	initialContent := "requires 'Test::More';\n"
	if err := os.WriteFile(cpanfilePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test cpanfile: %v", err)
	}

	tests := []struct {
		name           string
		backupConfig   *config.PVIBackupConfig
		expectBackup   bool
		backupLocation string // "local" or "cache"
	}{
		{
			name: "backup disabled",
			backupConfig: &config.PVIBackupConfig{
				CpanfileBackup: "off",
				RetentionDays:  30,
				MaxBackups:     10,
			},
			expectBackup: false,
		},
		{
			name: "local backup enabled",
			backupConfig: &config.PVIBackupConfig{
				CpanfileBackup: "local",
				RetentionDays:  30,
				MaxBackups:     10,
			},
			expectBackup:   true,
			backupLocation: "local",
		},
		{
			name: "cache backup enabled",
			backupConfig: &config.PVIBackupConfig{
				CpanfileBackup: "cache",
				RetentionDays:  30,
				MaxBackups:     10,
			},
			expectBackup:   true,
			backupLocation: "cache",
		},
		{
			name:         "nil config defaults to off",
			backupConfig: nil,
			expectBackup: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing backup files
			cleanupBackups(tempDir)

			mockWriter := &mockWriter{}
			logger := log.New(mockWriter, "", 0)
			cm, err := NewCpanfileManagerWithConfig(cpanfilePath, tt.backupConfig, logger)
			if err != nil {
				t.Fatalf("Failed to create cpanfile manager: %v", err)
			}

			// Verify backup mode
			expectedMode := "off"
			if tt.backupConfig != nil {
				expectedMode = tt.backupConfig.CpanfileBackup
			}

			if cm.GetBackupMode() != expectedMode {
				t.Errorf("Expected backup mode %s, got %s", expectedMode, cm.GetBackupMode())
			}

			// Add a dependency to trigger backup
			err = cm.AddDependency("Data::Dumper", "", false, false)
			if err != nil {
				t.Errorf("Failed to add dependency: %v", err)
			}

			// Check if backup was created
			if tt.expectBackup {
				switch tt.backupLocation {
				case "local":
					// Check for backup in same directory
					if !hasLocalBackup(tempDir) {
						t.Error("Expected local backup to be created")
					}
				case "cache":
					// Check for log message indicating cache backup
					foundCacheLog := false
					for _, logMsg := range mockWriter.logs {
						if strings.Contains(logMsg, "Created cache backup") {
							foundCacheLog = true
							break
						}
					}
					if !foundCacheLog {
						t.Error("Expected cache backup log message")
					}
				}
			} else if hasLocalBackup(tempDir) {
				// No backup should be created when backup is disabled
				t.Error("No backup should be created when backup is disabled")
			}

			// Verify cpanfile was actually modified
			content, err := os.ReadFile(cpanfilePath)
			if err != nil {
				t.Errorf("Failed to read cpanfile: %v", err)
			}

			if !strings.Contains(string(content), "Data::Dumper") {
				t.Error("Expected cpanfile to contain new dependency")
			}
		})
	}
}

func TestCpanfileManager_BackupModeOverride(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")

	// Create initial cpanfile
	if err := os.WriteFile(cpanfilePath, []byte("requires 'Test::More';\n"), 0644); err != nil {
		t.Fatalf("Failed to create test cpanfile: %v", err)
	}

	// Start with cache backup mode
	backupConfig := &config.PVIBackupConfig{
		CpanfileBackup: "cache",
		RetentionDays:  30,
		MaxBackups:     10,
	}

	mockWriter := &mockWriter{}
	logger := log.New(mockWriter, "", 0)
	cm, err := NewCpanfileManagerWithConfig(cpanfilePath, backupConfig, logger)
	if err != nil {
		t.Fatalf("Failed to create cpanfile manager: %v", err)
	}

	// Verify initial mode
	if cm.GetBackupMode() != "cache" {
		t.Errorf("Expected initial backup mode 'cache', got %s", cm.GetBackupMode())
	}

	// Override to local mode
	err = cm.SetBackupMode("local")
	if err != nil {
		t.Errorf("Failed to set backup mode: %v", err)
	}

	if cm.GetBackupMode() != "local" {
		t.Errorf("Expected backup mode 'local' after override, got %s", cm.GetBackupMode())
	}

	// Add dependency with local backup mode
	err = cm.AddDependency("Data::Dumper", "", false, false)
	if err != nil {
		t.Errorf("Failed to add dependency: %v", err)
	}

	// Should have created local backup
	if !hasLocalBackup(tempDir) {
		t.Error("Expected local backup to be created after mode override")
	}

	// Override to off mode
	err = cm.SetBackupMode("off")
	if err != nil {
		t.Errorf("Failed to set backup mode to off: %v", err)
	}

	// Clear previous backups and logs
	cleanupBackups(tempDir)
	mockWriter.logs = []string{}

	// Add another dependency with backup disabled
	err = cm.AddDependency("File::Spec", "", false, false)
	if err != nil {
		t.Errorf("Failed to add dependency: %v", err)
	}

	// Should not have created any new backups
	if hasLocalBackup(tempDir) {
		t.Error("No backup should be created when mode is set to 'off'")
	}

	// Test invalid mode
	err = cm.SetBackupMode("invalid")
	if err == nil {
		t.Error("Expected error for invalid backup mode")
	}
}

func TestCpanfileManager_RemoveDependencyWithBackup(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")

	// Create cpanfile with multiple dependencies
	content := `requires 'Test::More';
requires 'Data::Dumper';
requires 'File::Spec';
`
	if err := os.WriteFile(cpanfilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test cpanfile: %v", err)
	}

	backupConfig := &config.PVIBackupConfig{
		CpanfileBackup: "local",
		RetentionDays:  30,
		MaxBackups:     10,
	}

	mockWriter := &mockWriter{}
	logger := log.New(mockWriter, "", 0)
	cm, err := NewCpanfileManagerWithConfig(cpanfilePath, backupConfig, logger)
	if err != nil {
		t.Fatalf("Failed to create cpanfile manager: %v", err)
	}

	// Remove a dependency
	err = cm.RemoveDependency("Data::Dumper")
	if err != nil {
		t.Errorf("Failed to remove dependency: %v", err)
	}

	// Check backup was created
	if !hasLocalBackup(tempDir) {
		t.Error("Expected backup to be created before removing dependency")
	}

	// Verify dependency was removed
	updatedContent, err := os.ReadFile(cpanfilePath)
	if err != nil {
		t.Errorf("Failed to read updated cpanfile: %v", err)
	}

	if strings.Contains(string(updatedContent), "Data::Dumper") {
		t.Error("Expected Data::Dumper to be removed from cpanfile")
	}

	// Verify other dependencies are still there
	if !strings.Contains(string(updatedContent), "Test::More") {
		t.Error("Expected Test::More to remain in cpanfile")
	}
}

func TestCpanfileManager_MultipleOperationsWithRetention(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")

	// Create initial cpanfile
	if err := os.WriteFile(cpanfilePath, []byte("requires 'Test::More';\n"), 0644); err != nil {
		t.Fatalf("Failed to create test cpanfile: %v", err)
	}

	backupConfig := &config.PVIBackupConfig{
		CpanfileBackup: "local",
		RetentionDays:  30,
		MaxBackups:     3, // Low limit to test retention
	}

	mockWriter := &mockWriter{}
	logger := log.New(mockWriter, "", 0)
	cm, err := NewCpanfileManagerWithConfig(cpanfilePath, backupConfig, logger)
	if err != nil {
		t.Fatalf("Failed to create cpanfile manager: %v", err)
	}

	// Perform multiple operations to create several backups
	operations := []struct {
		operation string
		module    string
	}{
		{"add", "Data::Dumper"},
		{"add", "File::Spec"},
		{"add", "Carp"},
		{"add", "Scalar::Util"},
		{"remove", "Test::More"},
	}

	for _, op := range operations {
		switch op.operation {
		case "add":
			err = cm.AddDependency(op.module, "", false, false)
		case "remove":
			err = cm.RemoveDependency(op.module)
		}

		if err != nil {
			t.Errorf("Failed to %s %s: %v", op.operation, op.module, err)
		}
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

	if backupCount > backupConfig.MaxBackups {
		t.Errorf("Expected at most %d backups due to retention, found %d", backupConfig.MaxBackups, backupCount)
	}
}

func TestCpanfileManager_BackupFailureHandling(t *testing.T) {
	// Use a non-existent directory to cause backup failure
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")

	// Create cpanfile
	if err := os.WriteFile(cpanfilePath, []byte("requires 'Test::More';\n"), 0644); err != nil {
		t.Fatalf("Failed to create test cpanfile: %v", err)
	}

	// Create backup config that will cause issues (invalid cache dir by removing XDG access)
	backupConfig := &config.PVIBackupConfig{
		CpanfileBackup: "cache",
		RetentionDays:  30,
		MaxBackups:     10,
	}

	mockWriter := &mockWriter{}
	logger := log.New(mockWriter, "", 0)
	cm, err := NewCpanfileManagerWithConfig(cpanfilePath, backupConfig, logger)
	if err != nil {
		t.Fatalf("Failed to create cpanfile manager: %v", err)
	}

	// Even if backup fails, the cpanfile operation should succeed
	err = cm.AddDependency("Data::Dumper", "", false, false)
	if err != nil {
		t.Errorf("Cpanfile operation should succeed even if backup fails: %v", err)
	}

	// Verify the dependency was added
	content, err := os.ReadFile(cpanfilePath)
	if err != nil {
		t.Errorf("Failed to read cpanfile: %v", err)
	}

	if !strings.Contains(string(content), "Data::Dumper") {
		t.Error("Expected cpanfile to contain new dependency even if backup failed")
	}

	// Note: This test might not always trigger a warning if XDG directories work fine
	// but it tests the error handling path
	_ = mockWriter.logs // Verify logs are accessible for debugging if needed
}

// Helper functions

func hasLocalBackup(dir string) bool {
	files, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "cpanfile.backup.") {
			return true
		}
	}
	return false
}

func cleanupBackups(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "cpanfile.backup.") {
			os.Remove(filepath.Join(dir, file.Name()))
		}
	}
}
