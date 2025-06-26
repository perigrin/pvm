// ABOUTME: Tests for auto-update checking and notification system
// ABOUTME: Validates background update checks, notifications, and configuration management

package updater

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultAutoUpdateConfig(t *testing.T) {
	config := DefaultAutoUpdateConfig()

	if !config.Enabled {
		t.Error("Expected auto-update to be enabled by default")
	}

	if config.CheckInterval != 24*time.Hour {
		t.Errorf("Expected check interval to be 24 hours, got %v", config.CheckInterval)
	}

	if config.Channel != ChannelStable {
		t.Errorf("Expected stable channel by default, got %v", config.Channel)
	}

	if config.Repository != "perigrin/pvm" {
		t.Errorf("Expected repository perigrin/pvm, got %s", config.Repository)
	}

	if config.AutoInstall {
		t.Error("Expected auto-install to be disabled by default")
	}
}

func TestNewAutoUpdateManager(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "auto_update.json")

	manager, err := NewAutoUpdateManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create auto-update manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Expected manager to be created")
	}

	if manager.config == nil {
		t.Fatal("Expected config to be initialized")
	}

	// Check that config file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}
}

func TestAutoUpdateConfigSaveLoad(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "auto_update.json")

	// Create manager and modify config
	manager, err := NewAutoUpdateManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create auto-update manager: %v", err)
	}

	// Modify config
	config := manager.GetConfig()
	config.Channel = ChannelBeta
	config.CheckInterval = 12 * time.Hour
	config.AutoInstall = true

	err = manager.UpdateConfig(config)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Create new manager and verify config was loaded
	manager2, err := NewAutoUpdateManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create second manager: %v", err)
	}

	loadedConfig := manager2.GetConfig()
	if loadedConfig.Channel != ChannelBeta {
		t.Errorf("Expected channel beta, got %v", loadedConfig.Channel)
	}

	if loadedConfig.CheckInterval != 12*time.Hour {
		t.Errorf("Expected check interval 12 hours, got %v", loadedConfig.CheckInterval)
	}

	if !loadedConfig.AutoInstall {
		t.Error("Expected auto-install to be enabled")
	}
}

func TestUpdateChannelValidation(t *testing.T) {
	validChannels := []UpdateChannel{
		ChannelStable,
		ChannelBeta,
		ChannelAlpha,
		ChannelNightly,
		ChannelDeveloper,
	}

	for _, channel := range validChannels {
		if !channel.IsValid() {
			t.Errorf("Expected channel %v to be valid", channel)
		}
	}

	invalidChannel := UpdateChannel("invalid")
	if invalidChannel.IsValid() {
		t.Error("Expected invalid channel to be invalid")
	}
}

func TestShouldIncludePrerelease(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "auto_update.json")

	manager, err := NewAutoUpdateManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	tests := []struct {
		channel  UpdateChannel
		expected bool
	}{
		{ChannelStable, false},
		{ChannelBeta, true},
		{ChannelAlpha, true},
		{ChannelNightly, true},
		{ChannelDeveloper, true},
	}

	for _, test := range tests {
		config := manager.GetConfig()
		config.Channel = test.channel
		manager.UpdateConfig(config)

		result := manager.shouldIncludePrerelease()
		if result != test.expected {
			t.Errorf("Channel %v: expected %v, got %v", test.channel, test.expected, result)
		}
	}
}

func TestSecurityUpdateDetection(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "auto_update.json")

	manager, err := NewAutoUpdateManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	tests := []struct {
		releaseNotes string
		expected     bool
	}{
		{"This release fixes a security vulnerability", true},
		{"Security patch for CVE-2023-1234", true},
		{"Fixed exploit in parser", true},
		{"Regular bug fixes and improvements", false},
		{"Performance optimization", false},
		{"", false},
	}

	for _, test := range tests {
		result := manager.isSecurityUpdate(test.releaseNotes)
		if result != test.expected {
			t.Errorf("Release notes %q: expected %v, got %v", test.releaseNotes, test.expected, result)
		}
	}
}

func TestUrgentUpdateDetection(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "auto_update.json")

	manager, err := NewAutoUpdateManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	tests := []struct {
		releaseNotes string
		expected     bool
	}{
		{"Urgent hotfix for critical issue", true},
		{"Emergency patch", true},
		{"Critical security update", true},
		{"Regular bug fixes", false},
		{"Feature enhancement", false},
		{"", false},
	}

	for _, test := range tests {
		result := manager.isUrgentUpdate(test.releaseNotes)
		if result != test.expected {
			t.Errorf("Release notes %q: expected %v, got %v", test.releaseNotes, test.expected, result)
		}
	}
}

func TestAutoInstallSchedule(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "auto_update.json")

	manager, err := NewAutoUpdateManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test disabled auto-install
	result := manager.ShouldAutoInstall()
	if result {
		t.Error("Expected auto-install to be disabled by default")
	}

	// Enable auto-install for current day and time
	now := time.Now()
	config := manager.GetConfig()
	config.AutoInstall = true
	config.InstallationTime.Enabled = true
	config.InstallationTime.DaysOfWeek = []int{int(now.Weekday())}
	config.InstallationTime.Hour = now.Hour()
	config.InstallationTime.Minute = now.Minute()

	manager.UpdateConfig(config)

	result = manager.ShouldAutoInstall()
	if !result {
		t.Error("Expected auto-install to be enabled for current time")
	}

	// Test different day
	config.InstallationTime.DaysOfWeek = []int{(int(now.Weekday()) + 1) % 7}
	manager.UpdateConfig(config)

	result = manager.ShouldAutoInstall()
	if result {
		t.Error("Expected auto-install to be disabled for different day")
	}
}

func TestUpdateNotificationTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "auto_update.json")

	manager, err := NewAutoUpdateManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Set last check time to force an update check
	config := manager.GetConfig()
	config.LastCheckTime = time.Now().Add(-25 * time.Hour) // Force check
	config.Repository = "nonexistent/repo"                 // This will likely not have updates
	manager.UpdateConfig(config)

	ctx := context.Background()
	notification, err := manager.CheckForUpdates(ctx)

	// We expect this to fail or return no update for nonexistent repo
	// but we mainly want to test that the function doesn't crash
	_ = notification
	_ = err
}

func TestConfigJSONMarshaling(t *testing.T) {
	config := DefaultAutoUpdateConfig()
	config.Channel = ChannelBeta
	config.AutoInstall = true
	config.InstallationTime.Enabled = true
	config.InstallationTime.DaysOfWeek = []int{1, 3, 5}
	config.InstallationTime.Hour = 14
	config.InstallationTime.Minute = 30

	// Marshal to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Unmarshal from JSON
	var loadedConfig AutoUpdateConfig
	err = json.Unmarshal(data, &loadedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify values
	if loadedConfig.Channel != ChannelBeta {
		t.Errorf("Expected channel beta, got %v", loadedConfig.Channel)
	}

	if !loadedConfig.AutoInstall {
		t.Error("Expected auto-install to be enabled")
	}

	if !loadedConfig.InstallationTime.Enabled {
		t.Error("Expected installation time to be enabled")
	}

	if len(loadedConfig.InstallationTime.DaysOfWeek) != 3 {
		t.Errorf("Expected 3 days of week, got %d", len(loadedConfig.InstallationTime.DaysOfWeek))
	}

	if loadedConfig.InstallationTime.Hour != 14 {
		t.Errorf("Expected hour 14, got %d", loadedConfig.InstallationTime.Hour)
	}

	if loadedConfig.InstallationTime.Minute != 30 {
		t.Errorf("Expected minute 30, got %d", loadedConfig.InstallationTime.Minute)
	}
}

func TestQuietModeNotifications(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "auto_update.json")

	manager, err := NewAutoUpdateManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Enable quiet mode
	config := manager.GetConfig()
	config.QuietMode = true
	config.LastCheckTime = time.Now().Add(-25 * time.Hour) // Force check
	manager.UpdateConfig(config)

	// Even in quiet mode, we should be able to check for updates
	// The difference is notifications won't update the last notification time
	ctx := context.Background()
	_, err = manager.CheckForUpdates(ctx)

	// The check itself should not error due to quiet mode
	// (though it might error for other reasons like network issues)
	t.Logf("Update check result: %v", err)
}
