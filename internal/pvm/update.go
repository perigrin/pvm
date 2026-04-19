// ABOUTME: PVM self-updater command implementation
// ABOUTME: Handles the complete update workflow with user interface and progress reporting

package pvm

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/updater"
)

// executeUpdateCommand implements the update command functionality
func executeUpdateCommand(cmd *cobra.Command, args []string) error {
	// Create UI instance for enhanced output
	uiOutput := ui.NewDefaultOutput()

	// Load configuration
	cfg, err := config.LoadEffectiveConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get update configuration with defaults
	updateCfg := cfg.PVM.Update
	if updateCfg == nil {
		// Use default config if no update config exists
		updateCfg = config.NewDefaultConfig().PVM.Update
	}

	// Get flags (flags override configuration)
	targetVersion, _ := cmd.Flags().GetString("version")
	checkOnly, _ := cmd.Flags().GetBool("check")
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	noBackup, _ := cmd.Flags().GetBool("no-backup")
	noRollback, _ := cmd.Flags().GetBool("no-rollback")
	prerelease, _ := cmd.Flags().GetBool("prerelease")
	token, _ := cmd.Flags().GetString("token")
	ignoreInstallMethod, _ := cmd.Flags().GetBool("ignore-install-method")

	// Determine effective GitHub token (flag overrides config)
	effectiveToken := token
	if effectiveToken == "" {
		effectiveToken = updateCfg.GitHubToken
	}

	// Create updater
	var updaterInstance *updater.Updater
	if effectiveToken != "" {
		updaterInstance = updater.NewUpdaterWithToken(effectiveToken)
	} else {
		updaterInstance = updater.NewUpdater()
	}

	// Determine effective prerelease setting (flag overrides config)
	effectivePrerelease := prerelease
	if !cmd.Flags().Changed("prerelease") {
		effectivePrerelease = updateCfg.CheckPrerelease
	}

	// Determine effective backup setting (flag overrides config)
	effectiveBackup := !noBackup
	if !cmd.Flags().Changed("no-backup") {
		effectiveBackup = updateCfg.BackupEnabled
	}

	// Determine effective auto-rollback setting (flag overrides config)
	effectiveAutoRollback := !noRollback
	if !cmd.Flags().Changed("no-rollback") {
		effectiveAutoRollback = updateCfg.AutoRollbackEnabled
	}

	// Configure update options with config defaults and flag overrides
	opts := &updater.UpdateOptions{
		TargetVersion:       targetVersion,
		IncludePrerelease:   effectivePrerelease,
		Repository:          updateCfg.Repository,
		GitHubToken:         effectiveToken,
		Force:               force,
		DryRun:              dryRun,
		Backup:              effectiveBackup,
		AutoRollback:        effectiveAutoRollback,
		IgnoreInstallMethod: ignoreInstallMethod,
		Context:             cmd.Context(),
	}

	// Set up progress callback. The download/replace stages fire on every
	// HTTP read chunk (potentially 100+ per second), so only the \r-based
	// progress bar renders inside those stages — printing `message` with
	// a trailing \n on every tick would push the bar onto a new line each
	// time and produce the "scrolling bars" bug. For other stages the
	// message is printed once as a status line.
	var lastStage updater.UpdateStage
	opts.ProgressCallback = func(stage updater.UpdateStage, message string, progress float64) {
		barStage := stage == updater.StageDownloading || stage == updater.StageReplacing

		// Print stage header on transitions. Leading \n clears any
		// in-progress bar from the previous stage.
		if stage != lastStage {
			cmd.Printf("\n=== %s ===\n", stage.String())
			lastStage = stage
		}

		if barStage {
			// Bar-owned stage: let showProgressBar own the line via \r.
			// Don't print the per-tick message — that would newline and
			// the bar would scroll instead of overwrite.
			if progress > 0 && progress < 1.0 {
				showProgressBar(progress, 40)
			}
			return
		}

		// Non-bar stage: one status line per event.
		if message != "" {
			if progress > 0 && progress < 1.0 {
				cmd.Printf("%s (%.1f%%)\n", message, progress*100)
			} else {
				cmd.Println(message)
			}
		}
	}

	// Check for updates first
	cmd.Println("Checking for updates...")
	updateInfo, err := updaterInstance.CheckForUpdates(opts)
	if err != nil {
		// Provide a helpful hint if the error looks like an auth or rate limit problem
		errMsg := err.Error()
		if strings.Contains(errMsg, "returned 401") || strings.Contains(errMsg, "returned 403") || strings.Contains(errMsg, "rate limit") {
			cmd.Println("Hint: A GitHub token may be required. Configure one with:")
			cmd.Println("  pvm update --token YOUR_TOKEN")
			cmd.Println("  export GITHUB_TOKEN=YOUR_TOKEN")
		}
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	// Display current and available version info
	cmd.Printf("Current version: %s\n", updateInfo.CurrentVersion.String())
	cmd.Printf("Latest version:  %s\n", updateInfo.LatestVersion.String())

	if updateInfo.IsPrerelease {
		cmd.Println("Note: Latest version is a pre-release")
	}

	// If only checking, show result and exit
	if checkOnly {
		if updateInfo.UpdateNeeded {
			cmd.Printf("Update available: %s → %s\n",
				updateInfo.CurrentVersion.String(),
				updateInfo.LatestVersion.String())
			if updateInfo.Release != nil && updateInfo.Release.Body != "" {
				uiOutput.SubHeader("Release Notes")
				uiOutput.GlowMarkdown(updateInfo.Release.Body)
			}
		} else {
			cmd.Println("PVM is up to date")
		}
		return nil
	}

	// Check if update is needed
	if !updateInfo.UpdateNeeded && !force {
		cmd.Println("PVM is already up to date")
		return nil
	}

	// Show what will be updated
	if dryRun {
		cmd.Printf("\nDry run: Would update from %s to %s\n",
			updateInfo.CurrentVersion.String(),
			updateInfo.LatestVersion.String())
	} else {
		cmd.Printf("\nUpdating from %s to %s...\n",
			updateInfo.CurrentVersion.String(),
			updateInfo.LatestVersion.String())
	}

	// Show release notes if available
	if updateInfo.Release != nil && updateInfo.Release.Body != "" && !dryRun {
		uiOutput.SubHeader("Release Notes")
		uiOutput.GlowMarkdown(updateInfo.Release.Body)
		cmd.Println()
	}

	// Perform the update
	result, err := updaterInstance.PerformUpdate(opts)
	if err != nil {
		// Show detailed error information
		cmd.Printf("\nUpdate failed: %v\n", err)

		if result != nil && result.RollbackPerformed {
			cmd.Println("Automatic rollback was performed")
		}

		if result != nil && result.BackupCreated && result.BackupPath != "" {
			cmd.Printf("Backup created at: %s\n", result.BackupPath)
			cmd.Println("You can manually rollback using: pvm update --rollback")
		}

		return err
	}

	// Show success result
	cmd.Printf("\n%s\n", result.Message)

	if result.UpdatePerformed {
		cmd.Printf("Update completed in %s\n", result.Duration.Round(time.Millisecond))

		if result.BackupCreated && result.BackupPath != "" {
			cmd.Printf("Backup saved to: %s\n", result.BackupPath)
		}

		// Suggest restarting shell if needed
		cmd.Println("\nNote: You may need to restart your shell or run 'hash -r' to use the updated version")
	}

	return nil
}

// showProgressBar displays a simple progress bar
func showProgressBar(progress float64, width int) {
	filled := int(float64(width) * progress)

	fmt.Print("\r[")
	for i := 0; i < width; i++ {
		switch {
		case i < filled:
			fmt.Print("=")
		case i == filled:
			fmt.Print(">")
		default:
			fmt.Print(" ")
		}
	}
	fmt.Printf("] %.1f%%", progress*100)

	if progress >= 1.0 {
		fmt.Println()
	}
}

// newRollbackCommand creates a rollback command (can be added later)
func newRollbackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback to a previous version",
		Long:  "Rollback PVM to a previous version from backup",
		RunE:  executeRollbackCommand,
	}

	cmd.Flags().String("backup", "", "Specific backup to restore from")
	cmd.Flags().Bool("list", false, "List available backups")
	cmd.Flags().Bool("dry-run", false, "Show what would be restored without making changes")

	return cmd
}

// executeRollbackCommand implements rollback functionality
func executeRollbackCommand(cmd *cobra.Command, args []string) error {
	backupPath, _ := cmd.Flags().GetString("backup")
	listBackups, _ := cmd.Flags().GetBool("list")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	rollbackMgr := updater.NewRollbackManager()

	// List available backups
	if listBackups {
		backups, err := rollbackMgr.FindAvailableBackups()
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		if len(backups) == 0 {
			cmd.Println("No backups available")
			return nil
		}

		cmd.Println("Available backups:")
		for i, backup := range backups {
			cmd.Printf("%d. %s (created: %s)\n",
				i+1,
				backup.BackupPath,
				backup.CreatedAt.Format("2006-01-02 15:04:05"))
		}
		return nil
	}

	// Get current binary path
	currentPath, err := updater.GetCurrentBinaryPath()
	if err != nil {
		return fmt.Errorf("failed to detect current binary: %w", err)
	}

	// Check rollback status
	status, err := rollbackMgr.GetRollbackStatus(currentPath)
	if err != nil {
		return fmt.Errorf("failed to check rollback status: %w", err)
	}

	if !status.RollbackPossible {
		return fmt.Errorf("rollback not possible: %s", status.Reason)
	}

	// Show rollback info
	cmd.Printf("Target: %s\n", currentPath)
	cmd.Printf("Available backups: %d\n", status.AvailableBackups)

	if status.LatestBackupPath != "" {
		cmd.Printf("Latest backup: %s (created: %s)\n",
			status.LatestBackupPath,
			status.LatestBackupDate.Format("2006-01-02 15:04:05"))
	}

	// Perform rollback
	if dryRun {
		cmd.Println("\nDry run: Would rollback to latest backup")
		return nil
	}

	cmd.Println("\nPerforming rollback...")

	// Use specified backup or latest
	targetBackup := backupPath
	if targetBackup == "" {
		targetBackup = status.LatestBackupPath
	}

	updaterInstance := updater.NewUpdater()
	err = updaterInstance.Rollback(currentPath, targetBackup)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	cmd.Println("Rollback completed successfully")
	return nil
}

// newUpdateCheckCommand creates a command that only checks for updates
func newUpdateCheckCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-update",
		Short: "Check for available updates",
		Long:  "Check if a newer version of PVM is available without installing it",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set the check flag and call the main update command
			cmd.Flags().Set("check", "true")
			return executeUpdateCommand(cmd, args)
		},
	}

	// Add relevant flags
	cmd.Flags().Bool("prerelease", false, "Include pre-release versions")
	cmd.Flags().String("token", "", "GitHub token for higher API rate limits")

	return cmd
}
