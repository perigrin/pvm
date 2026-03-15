// ABOUTME: Auto-update command for PVM - manages automatic update checking and configuration
// ABOUTME: Provides user interface for configuring background update checks and notifications

package pvm

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/updater"
	"tamarou.com/pvm/internal/xdg"
)

// executeAutoUpdateCommand implements the auto-update command functionality
func executeAutoUpdateCommand(cmd *cobra.Command, args []string) error {
	// Get configuration path
	dirs, err := xdg.GetDirs()
	if err != nil {
		return fmt.Errorf("failed to get directories: %w", err)
	}
	configPath := filepath.Join(dirs.ConfigDir, "auto_update.json")

	// Create auto-update manager
	token, _ := cmd.Flags().GetString("token")
	var manager *updater.AutoUpdateManager
	if token != "" {
		manager, err = updater.NewAutoUpdateManagerWithToken(configPath, token)
	} else {
		manager, err = updater.NewAutoUpdateManager(configPath)
	}
	if err != nil {
		return fmt.Errorf("failed to create auto-update manager: %w", err)
	}

	// Handle subcommands
	if len(args) == 0 {
		return showAutoUpdateStatus(cmd, manager)
	}

	switch args[0] {
	case "enable":
		return enableAutoUpdate(cmd, manager)
	case "disable":
		return disableAutoUpdate(cmd, manager)
	case "config":
		return configureAutoUpdate(cmd, manager, args[1:])
	case "check":
		return checkForUpdatesNow(cmd, manager)
	case "status":
		return showAutoUpdateStatus(cmd, manager)
	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
}

// showAutoUpdateStatus displays current auto-update configuration and status
func showAutoUpdateStatus(cmd *cobra.Command, manager *updater.AutoUpdateManager) error {
	config := manager.GetConfig()

	cmd.Println("Auto-Update Status:")
	cmd.Printf("  Enabled: %v\n", config.Enabled)
	cmd.Printf("  Channel: %s\n", config.Channel)
	cmd.Printf("  Check Interval: %s\n", config.CheckInterval)
	cmd.Printf("  Repository: %s\n", config.Repository)
	cmd.Printf("  Quiet Mode: %v\n", config.QuietMode)
	cmd.Printf("  Auto Install: %v\n", config.AutoInstall)

	if !config.LastCheckTime.IsZero() {
		cmd.Printf("  Last Check: %s (%s ago)\n",
			config.LastCheckTime.Format("2006-01-02 15:04:05"),
			time.Since(config.LastCheckTime).Round(time.Minute))
	} else {
		cmd.Println("  Last Check: Never")
	}

	if !config.LastNotification.IsZero() {
		cmd.Printf("  Last Notification: %s (%s ago)\n",
			config.LastNotification.Format("2006-01-02 15:04:05"),
			time.Since(config.LastNotification).Round(time.Minute))
	}

	if config.InstallationTime.Enabled {
		cmd.Println("  Auto Installation Schedule:")
		cmd.Printf("    Enabled: %v\n", config.InstallationTime.Enabled)
		cmd.Printf("    Days: %s\n", formatDaysOfWeek(config.InstallationTime.DaysOfWeek))
		cmd.Printf("    Time: %02d:%02d\n", config.InstallationTime.Hour, config.InstallationTime.Minute)
	}

	// Show next check time
	if config.Enabled && !config.LastCheckTime.IsZero() {
		nextCheck := config.LastCheckTime.Add(config.CheckInterval)
		if nextCheck.After(time.Now()) {
			cmd.Printf("  Next Check: %s (in %s)\n",
				nextCheck.Format("2006-01-02 15:04:05"),
				time.Until(nextCheck).Round(time.Minute))
		} else {
			cmd.Println("  Next Check: Now (overdue)")
		}
	}

	return nil
}

// enableAutoUpdate enables automatic update checking
func enableAutoUpdate(cmd *cobra.Command, manager *updater.AutoUpdateManager) error {
	config := manager.GetConfig()
	config.Enabled = true

	if err := manager.UpdateConfig(config); err != nil {
		return fmt.Errorf("failed to enable auto-update: %w", err)
	}

	cmd.Println("Auto-update checking enabled")
	return nil
}

// disableAutoUpdate disables automatic update checking
func disableAutoUpdate(cmd *cobra.Command, manager *updater.AutoUpdateManager) error {
	config := manager.GetConfig()
	config.Enabled = false

	if err := manager.UpdateConfig(config); err != nil {
		return fmt.Errorf("failed to disable auto-update: %w", err)
	}

	cmd.Println("Auto-update checking disabled")
	return nil
}

// configureAutoUpdate handles configuration changes
func configureAutoUpdate(cmd *cobra.Command, manager *updater.AutoUpdateManager, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("config requires key and value: pvm auto-update config <key> <value>")
	}

	key := args[0]
	value := args[1]

	config := manager.GetConfig()

	switch key {
	case "channel":
		channel := updater.UpdateChannel(value)
		if !channel.IsValid() {
			return fmt.Errorf("invalid channel: %s (valid: stable, beta, alpha, nightly, developer)", value)
		}
		config.Channel = channel
		cmd.Printf("Update channel set to: %s\n", channel)

	case "interval":
		duration, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid interval format: %s (examples: 24h, 12h, 1h30m)", value)
		}
		if duration < time.Hour {
			return fmt.Errorf("interval cannot be less than 1 hour")
		}
		config.CheckInterval = duration
		cmd.Printf("Check interval set to: %s\n", duration)

	case "quiet":
		quiet, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s (use true or false)", value)
		}
		config.QuietMode = quiet
		cmd.Printf("Quiet mode set to: %v\n", quiet)

	case "auto-install":
		autoInstall, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s (use true or false)", value)
		}
		config.AutoInstall = autoInstall
		cmd.Printf("Auto-install set to: %v\n", autoInstall)

	case "repository":
		config.Repository = value
		cmd.Printf("Repository set to: %s\n", value)

	case "install-schedule":
		if len(args) < 4 {
			return fmt.Errorf("install-schedule requires: days hour minute (e.g., 'Sun,Mon' 2 30)")
		}

		// Parse days of week
		daysStr := args[1]
		days, err := parseDaysOfWeek(daysStr)
		if err != nil {
			return fmt.Errorf("invalid days format: %s", err)
		}

		// Parse hour
		hour, err := strconv.Atoi(args[2])
		if err != nil || hour < 0 || hour > 23 {
			return fmt.Errorf("invalid hour: %s (must be 0-23)", args[2])
		}

		// Parse minute
		minute, err := strconv.Atoi(args[3])
		if err != nil || minute < 0 || minute > 59 {
			return fmt.Errorf("invalid minute: %s (must be 0-59)", args[3])
		}

		config.InstallationTime.Enabled = true
		config.InstallationTime.DaysOfWeek = days
		config.InstallationTime.Hour = hour
		config.InstallationTime.Minute = minute

		cmd.Printf("Installation schedule set to: %s at %02d:%02d\n",
			formatDaysOfWeek(days), hour, minute)

	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	if err := manager.UpdateConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// checkForUpdatesNow performs an immediate update check
func checkForUpdatesNow(cmd *cobra.Command, manager *updater.AutoUpdateManager) error {
	cmd.Println("Checking for updates...")

	// Force check by temporarily enabling and resetting last check time
	config := manager.GetConfig()
	originalEnabled := config.Enabled
	originalLastCheck := config.LastCheckTime

	config.Enabled = true
	config.LastCheckTime = time.Time{} // Reset to force check
	manager.UpdateConfig(config)

	notification, err := manager.CheckForUpdates(cmd.Context())
	if err != nil {
		// Restore original config
		config.Enabled = originalEnabled
		config.LastCheckTime = originalLastCheck
		manager.UpdateConfig(config)
		return fmt.Errorf("update check failed: %w", err)
	}

	// Restore original enabled state but keep the updated last check time
	config = manager.GetConfig()
	config.Enabled = originalEnabled
	manager.UpdateConfig(config)

	if notification == nil {
		cmd.Println("No updates available or check was skipped")
		return nil
	}

	if !notification.Available {
		cmd.Printf("No updates available. Current version: %s\n", notification.CurrentVersion)
		return nil
	}

	cmd.Printf("Update available: %s → %s\n", notification.CurrentVersion, notification.LatestVersion)
	cmd.Printf("Channel: %s\n", notification.Channel)

	if notification.SecurityUpdate {
		cmd.Println("⚠️  This is a SECURITY UPDATE")
	}

	if notification.Urgent {
		cmd.Println("🚨 This is an URGENT UPDATE")
	}

	if notification.DownloadSize > 0 {
		cmd.Printf("Download size: %s\n", formatBytes(notification.DownloadSize))
	}

	if notification.ReleaseNotes != "" {
		cmd.Println("\nRelease notes:")
		cmd.Println(notification.ReleaseNotes)
	}

	cmd.Println("\nTo install this update, run: pvm update")

	return nil
}

// formatDaysOfWeek converts day numbers to readable format
func formatDaysOfWeek(days []int) string {
	dayNames := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	names := make([]string, len(days))

	for i, day := range days {
		if day >= 0 && day <= 6 {
			names[i] = dayNames[day]
		} else {
			names[i] = fmt.Sprintf("Day%d", day)
		}
	}

	return strings.Join(names, ", ")
}

// parseDaysOfWeek parses day names into day numbers
func parseDaysOfWeek(daysStr string) ([]int, error) {
	dayMap := map[string]int{
		"sun": 0, "sunday": 0,
		"mon": 1, "monday": 1,
		"tue": 2, "tuesday": 2,
		"wed": 3, "wednesday": 3,
		"thu": 4, "thursday": 4,
		"fri": 5, "friday": 5,
		"sat": 6, "saturday": 6,
	}

	parts := strings.Split(strings.ToLower(daysStr), ",")
	days := make([]int, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if day, ok := dayMap[part]; ok {
			days[i] = day
		} else {
			return nil, fmt.Errorf("unknown day: %s", part)
		}
	}

	return days, nil
}

// formatBytes formats byte size in human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
