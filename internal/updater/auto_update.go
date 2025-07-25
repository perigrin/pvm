// ABOUTME: Auto-update checking and notification system for PVM
// ABOUTME: Handles background update checks, notifications, and update channel management

package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tamarou.com/pvm/internal/version"
)

// AutoUpdateManager handles automatic update checking and notifications
type AutoUpdateManager struct {
	versionChecker *version.GitHubClient
	configPath     string
	config         *AutoUpdateConfig
}

// AutoUpdateConfig stores auto-update configuration
type AutoUpdateConfig struct {
	Enabled           bool                `json:"enabled"`
	CheckInterval     time.Duration       `json:"check_interval"`
	Channel           UpdateChannel       `json:"channel"`
	LastCheckTime     time.Time           `json:"last_check_time"`
	LastNotification  time.Time           `json:"last_notification"`
	NotificationDelay time.Duration       `json:"notification_delay"`
	Repository        string              `json:"repository"`
	GitHubToken       string              `json:"github_token,omitempty"`
	QuietMode         bool                `json:"quiet_mode"`
	AutoInstall       bool                `json:"auto_install"`
	InstallationTime  AutoInstallSchedule `json:"installation_time"`
	// testClient is used for dependency injection in tests (not serialized)
	testClient version.GitHubClientInterface `json:"-"`
}

// UpdateChannel represents different update channels
type UpdateChannel string

const (
	ChannelStable    UpdateChannel = "stable"
	ChannelBeta      UpdateChannel = "beta"
	ChannelAlpha     UpdateChannel = "alpha"
	ChannelNightly   UpdateChannel = "nightly"
	ChannelDeveloper UpdateChannel = "developer"
)

// AutoInstallSchedule defines when auto-installation should occur
type AutoInstallSchedule struct {
	Enabled    bool  `json:"enabled"`
	DaysOfWeek []int `json:"days_of_week"` // 0=Sunday, 1=Monday, etc.
	Hour       int   `json:"hour"`         // 0-23
	Minute     int   `json:"minute"`       // 0-59
}

// UpdateNotification represents an update notification
type UpdateNotification struct {
	Available      bool          `json:"available"`
	CurrentVersion string        `json:"current_version"`
	LatestVersion  string        `json:"latest_version"`
	Channel        UpdateChannel `json:"channel"`
	ReleaseNotes   string        `json:"release_notes"`
	Urgent         bool          `json:"urgent"`
	SecurityUpdate bool          `json:"security_update"`
	DownloadSize   int64         `json:"download_size"`
	Timestamp      time.Time     `json:"timestamp"`
}

// DefaultAutoUpdateConfig returns default auto-update configuration
func DefaultAutoUpdateConfig() *AutoUpdateConfig {
	return &AutoUpdateConfig{
		Enabled:           true,
		CheckInterval:     24 * time.Hour, // Check daily
		Channel:           ChannelStable,
		NotificationDelay: 4 * time.Hour, // Wait 4 hours between notifications
		Repository:        "perigrin/pvm",
		QuietMode:         false,
		AutoInstall:       false,
		InstallationTime: AutoInstallSchedule{
			Enabled:    false,
			DaysOfWeek: []int{0}, // Sunday
			Hour:       2,        // 2 AM
			Minute:     0,
		},
	}
}

// NewAutoUpdateManager creates a new auto-update manager
func NewAutoUpdateManager(configPath string) (*AutoUpdateManager, error) {
	manager := &AutoUpdateManager{
		versionChecker: version.NewGitHubClient(),
		configPath:     configPath,
	}

	// Load or create config
	config, err := manager.loadConfig()
	if err != nil {
		// If config doesn't exist, create default
		config = DefaultAutoUpdateConfig()
		if err := manager.saveConfig(config); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}

	manager.config = config
	return manager, nil
}

// NewAutoUpdateManagerWithToken creates a new auto-update manager with GitHub token
func NewAutoUpdateManagerWithToken(configPath, token string) (*AutoUpdateManager, error) {
	manager, err := NewAutoUpdateManager(configPath)
	if err != nil {
		return nil, err
	}

	manager.versionChecker = version.NewGitHubClientWithToken(token)
	manager.config.GitHubToken = token
	return manager, nil
}

// CheckForUpdates performs an update check and returns notification if available
func (m *AutoUpdateManager) CheckForUpdates(ctx context.Context) (*UpdateNotification, error) {
	if !m.config.Enabled {
		return nil, nil
	}

	// Check if enough time has passed since last check
	if time.Since(m.config.LastCheckTime) < m.config.CheckInterval {
		return nil, nil
	}

	// Perform version check
	checkOpts := &version.CheckOptions{
		IncludePrerelease: m.shouldIncludePrerelease(),
		Repository:        m.config.Repository,
		GitHubToken:       m.config.GitHubToken,
		Timeout:           30 * time.Second,
		Client:            m.config.testClient, // For dependency injection in tests
	}

	updateInfo, err := version.GetUpdateInfo(checkOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	// Update last check time
	m.config.LastCheckTime = time.Now()
	if err := m.saveConfig(m.config); err != nil {
		// Log error but don't fail the check
		_ = err
	}

	// Check if update is needed
	if !updateInfo.UpdateNeeded {
		return &UpdateNotification{
			Available:      false,
			CurrentVersion: updateInfo.CurrentVersion.String(),
			Channel:        m.config.Channel,
			Timestamp:      time.Now(),
		}, nil
	}

	// Check if we should show notification (not too frequent)
	if !m.config.QuietMode && time.Since(m.config.LastNotification) < m.config.NotificationDelay {
		return nil, nil
	}

	// Create notification
	notification := &UpdateNotification{
		Available:      true,
		CurrentVersion: updateInfo.CurrentVersion.String(),
		LatestVersion:  updateInfo.LatestVersion.String(),
		Channel:        m.config.Channel,
		Timestamp:      time.Now(),
	}

	// Add release notes if available
	if updateInfo.Release != nil {
		notification.ReleaseNotes = updateInfo.Release.Body
		notification.SecurityUpdate = m.isSecurityUpdate(updateInfo.Release.Body)
		notification.Urgent = m.isUrgentUpdate(updateInfo.Release.Body)

		// Try to get download size
		if asset, err := version.GetUpdateAsset(updateInfo.Release, version.DetectPlatform()); err == nil {
			notification.DownloadSize = int64(asset.Size)
		}
	}

	// Update last notification time
	if !m.config.QuietMode {
		m.config.LastNotification = time.Now()
		if err := m.saveConfig(m.config); err != nil {
			// Log error but don't fail
			_ = err
		}
	}

	return notification, nil
}

// ShouldAutoInstall checks if an auto-installation should be performed
func (m *AutoUpdateManager) ShouldAutoInstall() bool {
	if !m.config.AutoInstall || !m.config.InstallationTime.Enabled {
		return false
	}

	now := time.Now()
	schedule := m.config.InstallationTime

	// Check if current day of week matches
	currentDay := int(now.Weekday())
	dayMatches := false
	for _, day := range schedule.DaysOfWeek {
		if day == currentDay {
			dayMatches = true
			break
		}
	}

	if !dayMatches {
		return false
	}

	// Check if current time matches (within 1 hour window)
	currentHour := now.Hour()
	return currentHour == schedule.Hour
}

// GetConfig returns the current configuration
func (m *AutoUpdateManager) GetConfig() *AutoUpdateConfig {
	return m.config
}

// UpdateConfig updates the configuration
func (m *AutoUpdateManager) UpdateConfig(config *AutoUpdateConfig) error {
	m.config = config
	return m.saveConfig(config)
}

// shouldIncludePrerelease determines if pre-release versions should be included
func (m *AutoUpdateManager) shouldIncludePrerelease() bool {
	switch m.config.Channel {
	case ChannelStable:
		return false
	case ChannelBeta, ChannelAlpha, ChannelNightly, ChannelDeveloper:
		return true
	default:
		return false
	}
}

// isSecurityUpdate checks if the release notes indicate a security update
func (m *AutoUpdateManager) isSecurityUpdate(releaseNotes string) bool {
	// Simple keyword matching - could be enhanced with more sophisticated detection
	keywords := []string{"security", "vulnerability", "CVE-", "exploit", "patch"}
	lowerNotes := strings.ToLower(releaseNotes)

	for _, keyword := range keywords {
		if strings.Contains(lowerNotes, keyword) {
			return true
		}
	}
	return false
}

// isUrgentUpdate checks if the release notes indicate an urgent update
func (m *AutoUpdateManager) isUrgentUpdate(releaseNotes string) bool {
	keywords := []string{"urgent", "critical", "emergency", "hotfix", "immediate"}
	lowerNotes := strings.ToLower(releaseNotes)

	for _, keyword := range keywords {
		if strings.Contains(lowerNotes, keyword) {
			return true
		}
	}
	return false
}

// loadConfig loads configuration from file
func (m *AutoUpdateManager) loadConfig() (*AutoUpdateConfig, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, err
	}

	var config AutoUpdateConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// saveConfig saves configuration to file
func (m *AutoUpdateManager) saveConfig(config *AutoUpdateConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SetTestClient sets a test client for dependency injection (testing only)
func (m *AutoUpdateManager) SetTestClient(client version.GitHubClientInterface) {
	m.config.testClient = client
}

// String methods for enums
func (c UpdateChannel) String() string {
	return string(c)
}

func (c UpdateChannel) IsValid() bool {
	switch c {
	case ChannelStable, ChannelBeta, ChannelAlpha, ChannelNightly, ChannelDeveloper:
		return true
	default:
		return false
	}
}
