package config

import (
	"testing"
)

func TestPVIBackupConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *PVIBackupConfig
		expectError bool
		errorCount  int
	}{
		{
			name: "valid config with off mode",
			config: &PVIBackupConfig{
				CpanfileBackup: "off",
				RetentionDays:  30,
				MaxBackups:     10,
			},
			expectError: false,
		},
		{
			name: "valid config with local mode",
			config: &PVIBackupConfig{
				CpanfileBackup: "local",
				RetentionDays:  30,
				MaxBackups:     10,
			},
			expectError: false,
		},
		{
			name: "valid config with cache mode",
			config: &PVIBackupConfig{
				CpanfileBackup: "cache",
				RetentionDays:  30,
				MaxBackups:     10,
			},
			expectError: false,
		},
		{
			name: "valid config with zero retention",
			config: &PVIBackupConfig{
				CpanfileBackup: "local",
				RetentionDays:  0, // Zero should be valid (no retention)
				MaxBackups:     10,
			},
			expectError: false,
		},
		{
			name: "valid config with zero max backups",
			config: &PVIBackupConfig{
				CpanfileBackup: "local",
				RetentionDays:  30,
				MaxBackups:     0, // Zero should be valid (no limit)
			},
			expectError: false,
		},
		{
			name: "invalid backup mode",
			config: &PVIBackupConfig{
				CpanfileBackup: "invalid",
				RetentionDays:  30,
				MaxBackups:     10,
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "empty backup mode",
			config: &PVIBackupConfig{
				CpanfileBackup: "",
				RetentionDays:  30,
				MaxBackups:     10,
			},
			expectError: false, // Empty mode should be valid (might default)
		},
		{
			name: "negative retention days",
			config: &PVIBackupConfig{
				CpanfileBackup: "local",
				RetentionDays:  -1,
				MaxBackups:     10,
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "negative max backups",
			config: &PVIBackupConfig{
				CpanfileBackup: "local",
				RetentionDays:  30,
				MaxBackups:     -1,
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "multiple validation errors",
			config: &PVIBackupConfig{
				CpanfileBackup: "invalid",
				RetentionDays:  -1,
				MaxBackups:     -5,
			},
			expectError: true,
			errorCount:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Validate()

			if tt.expectError {
				if len(errors) == 0 {
					t.Error("Expected validation errors but got none")
				}

				if tt.errorCount > 0 && len(errors) != tt.errorCount {
					t.Errorf("Expected %d validation errors, got %d", tt.errorCount, len(errors))
				}
			} else if len(errors) > 0 {
				t.Errorf("Unexpected validation errors: %v", errors)
			}
		})
	}
}

func TestPVIConfig_BackupValidation(t *testing.T) {
	// Test that PVIConfig properly validates its Backup field
	config := &PVIConfig{
		PreferredInstaller: "cpanm",
		DefaultMirror:      "https://cpan.metacpan.org",
		Backup: &PVIBackupConfig{
			CpanfileBackup: "invalid",
			RetentionDays:  -1,
			MaxBackups:     -1,
		},
	}

	errors := config.Validate()

	// Should include backup validation errors
	if len(errors) == 0 {
		t.Error("Expected validation errors from invalid backup config")
	}

	// Check that we get the expected backup validation errors
	expectedErrors := []string{
		"CpanfileBackup must be one of: off, local, cache",
		"RetentionDays cannot be negative",
		"MaxBackups cannot be negative",
	}

	for _, expectedError := range expectedErrors {
		found := false
		for _, err := range errors {
			if err.Error() == expectedError {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected validation error: %s", expectedError)
		}
	}
}

func TestDefaultConfig_BackupDefaults(t *testing.T) {
	// Test that the default configuration includes proper backup defaults
	cfg := NewDefaultConfig()

	if cfg.PVI == nil {
		t.Fatal("Expected PVI config to be initialized")
	}

	if cfg.PVI.Backup == nil {
		t.Fatal("Expected PVI Backup config to be initialized")
	}

	backup := cfg.PVI.Backup

	// Check default values
	if backup.CpanfileBackup != "off" {
		t.Errorf("Expected default backup mode 'off', got %s", backup.CpanfileBackup)
	}

	if backup.RetentionDays != 30 {
		t.Errorf("Expected default retention days 30, got %d", backup.RetentionDays)
	}

	if backup.MaxBackups != 10 {
		t.Errorf("Expected default max backups 10, got %d", backup.MaxBackups)
	}

	// Verify the config validates correctly
	errors := cfg.Validate()
	if len(errors) > 0 {
		t.Errorf("Default configuration should validate without errors, got: %v", errors)
	}
}

func TestConfig_BackupConfigIntegration(t *testing.T) {
	// Test full config validation including backup config
	cfg := &Config{
		PVI: &PVIConfig{
			PreferredInstaller: "cpanm",
			DefaultMirror:      "https://cpan.metacpan.org",
			MetadataSource:     "metacpan",
			CacheTTL:           24,
			Backup: &PVIBackupConfig{
				CpanfileBackup: "local",
				RetentionDays:  30,
				MaxBackups:     10,
			},
		},
	}

	errors := cfg.Validate()
	if len(errors) > 0 {
		t.Errorf("Valid config should not have validation errors, got: %v", errors)
	}

	// Test with invalid backup config
	cfg.PVI.Backup.CpanfileBackup = "invalid"
	cfg.PVI.Backup.RetentionDays = -1

	errors = cfg.Validate()
	if len(errors) == 0 {
		t.Error("Expected validation errors for invalid backup config")
	}

	// Should contain backup-specific errors
	hasBackupError := false
	for _, err := range errors {
		if err.Error() == "CpanfileBackup must be one of: off, local, cache" ||
			err.Error() == "RetentionDays cannot be negative" {
			hasBackupError = true
			break
		}
	}

	if !hasBackupError {
		t.Error("Expected backup-specific validation errors")
	}
}
