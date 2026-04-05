// ABOUTME: Advanced error recovery for PVM update failures
// ABOUTME: Provides comprehensive recovery strategies for various failure scenarios

package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tamarou.com/pvm/internal/download"
	"tamarou.com/pvm/internal/version"
)

// RecoveryManager handles advanced error recovery scenarios
type RecoveryManager struct {
	rollbackManager *RollbackManager
	backupManager   *BackupManager
	downloader      *download.Downloader
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager() *RecoveryManager {
	return &RecoveryManager{
		rollbackManager: NewRollbackManager(),
		backupManager:   NewBackupManager(),
		downloader:      download.NewDownloader(),
	}
}

// RecoveryScenario represents different types of recovery scenarios
type RecoveryScenario int

const (
	ScenarioCorruptedBinary RecoveryScenario = iota
	ScenarioPartialUpdate
	ScenarioPermissionDenied
	ScenarioFileSystemFull
	ScenarioNetworkFailure
	ScenarioChecksumMismatch
	ScenarioIncompatibleBinary
	ScenarioUnknownFailure
)

// RecoveryStrategy defines how to handle each scenario
type RecoveryStrategy struct {
	Scenario    RecoveryScenario
	Description string
	Priority    int // Lower = higher priority
	Automatic   bool
	Action      func(*RecoveryContext) error
}

// RecoveryContext contains information needed for recovery
type RecoveryContext struct {
	Scenario       RecoveryScenario
	TargetPath     string
	FailedVersion  string
	PreviousBackup string
	ErrorDetails   error
	AttemptCount   int
	MaxAttempts    int
}

// RecoveryResult contains the outcome of recovery attempts
type RecoveryResult struct {
	Success          bool
	Strategy         string
	Duration         time.Duration
	RecoveredPath    string
	AttemptsUsed     int
	RemainingBackups int
	Message          string
}

// PerformRecovery attempts to recover from update failures
func (rm *RecoveryManager) PerformRecovery(ctx *RecoveryContext) (*RecoveryResult, error) {
	startTime := time.Now()

	result := &RecoveryResult{
		AttemptsUsed: ctx.AttemptCount,
	}

	// Get available recovery strategies for this scenario
	strategies := rm.getRecoveryStrategies(ctx.Scenario)

	// Sort strategies by priority
	sort.Slice(strategies, func(i, j int) bool {
		return strategies[i].Priority < strategies[j].Priority
	})

	// Try each strategy in order
	for _, strategy := range strategies {
		if ctx.AttemptCount >= ctx.MaxAttempts {
			break
		}

		ctx.AttemptCount++
		result.AttemptsUsed++

		fmt.Printf("Attempting recovery strategy: %s\n", strategy.Description)

		if err := strategy.Action(ctx); err != nil {
			fmt.Printf("Recovery strategy failed: %v\n", err)
			continue
		}

		// Strategy succeeded
		result.Success = true
		result.Strategy = strategy.Description
		result.Duration = time.Since(startTime)
		result.RecoveredPath = ctx.TargetPath
		result.Message = fmt.Sprintf("Successfully recovered using: %s", strategy.Description)

		// Check remaining backups
		backups, _ := rm.backupManager.ListBackups()
		result.RemainingBackups = len(backups)

		return result, nil
	}

	// All strategies failed
	result.Duration = time.Since(startTime)
	result.Message = "All recovery strategies failed"

	return result, fmt.Errorf("recovery failed after %d attempts", result.AttemptsUsed)
}

// getRecoveryStrategies returns applicable strategies for a scenario
func (rm *RecoveryManager) getRecoveryStrategies(scenario RecoveryScenario) []*RecoveryStrategy {
	strategies := []*RecoveryStrategy{
		{
			Scenario:    ScenarioCorruptedBinary,
			Description: "Rollback to most recent backup",
			Priority:    1,
			Automatic:   true,
			Action:      rm.rollbackToLatestBackup,
		},
		{
			Scenario:    ScenarioCorruptedBinary,
			Description: "Rollback to stable backup",
			Priority:    2,
			Automatic:   true,
			Action:      rm.rollbackToStableBackup,
		},
		{
			Scenario:    ScenarioPartialUpdate,
			Description: "Complete interrupted update",
			Priority:    1,
			Automatic:   true,
			Action:      rm.completeInterruptedUpdate,
		},
		{
			Scenario:    ScenarioPartialUpdate,
			Description: "Clean rollback and retry",
			Priority:    2,
			Automatic:   false,
			Action:      rm.cleanRollbackAndRetry,
		},
		{
			Scenario:    ScenarioPermissionDenied,
			Description: "Fix permissions and retry",
			Priority:    1,
			Automatic:   true,
			Action:      rm.fixPermissionsAndRetry,
		},
		{
			Scenario:    ScenarioFileSystemFull,
			Description: "Clean temporary files and retry",
			Priority:    1,
			Automatic:   true,
			Action:      rm.cleanTempFilesAndRetry,
		},
		{
			Scenario:    ScenarioChecksumMismatch,
			Description: "Re-download and verify",
			Priority:    1,
			Automatic:   true,
			Action:      rm.redownloadAndVerify,
		},
		{
			Scenario:    ScenarioIncompatibleBinary,
			Description: "Download previous compatible version",
			Priority:    1,
			Automatic:   false,
			Action:      rm.downloadCompatibleVersion,
		},
		{
			Scenario:    ScenarioUnknownFailure,
			Description: "Emergency rollback to last known good",
			Priority:    1,
			Automatic:   true,
			Action:      rm.emergencyRollback,
		},
	}

	// Filter strategies for the specific scenario
	var applicable []*RecoveryStrategy
	for _, strategy := range strategies {
		if strategy.Scenario == scenario {
			applicable = append(applicable, strategy)
		}
	}

	return applicable
}

// Recovery action implementations

func (rm *RecoveryManager) rollbackToLatestBackup(ctx *RecoveryContext) error {
	opts := &RollbackOptions{
		TargetPath:     ctx.TargetPath,
		ValidateBackup: true,
		DryRun:         false,
	}

	result, err := rm.rollbackManager.PerformRollback(opts)
	if err != nil {
		return fmt.Errorf("latest backup rollback failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("latest backup rollback was not successful")
	}

	return nil
}

func (rm *RecoveryManager) rollbackToStableBackup(ctx *RecoveryContext) error {
	// Find the oldest backup (more likely to be stable)
	backups, err := rm.backupManager.ListBackups()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}

	if len(backups) == 0 {
		return fmt.Errorf("no backups available for stable rollback")
	}

	// Sort by creation time (oldest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.Before(backups[j].CreatedAt)
	})

	stableBackup := backups[0]

	opts := &RollbackOptions{
		TargetPath:     ctx.TargetPath,
		BackupPath:     stableBackup.BackupPath,
		ValidateBackup: true,
		DryRun:         false,
	}

	result, err := rm.rollbackManager.PerformRollback(opts)
	if err != nil {
		return fmt.Errorf("stable backup rollback failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("stable backup rollback was not successful")
	}

	return nil
}

func (rm *RecoveryManager) completeInterruptedUpdate(ctx *RecoveryContext) error {
	// Look for partially downloaded files
	tempDir := os.TempDir()
	pattern := filepath.Join(tempDir, "pvm-update-*")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("searching for partial downloads: %w", err)
	}

	for _, match := range matches {
		// Try to resume download or clean up
		if info, err := os.Stat(match); err == nil {
			if time.Since(info.ModTime()) > 24*time.Hour {
				// Old file, clean it up
				os.Remove(match)
			}
		}
	}

	// Check if update can be completed
	// This would involve checking for partial state and attempting to complete
	// For now, we'll indicate that manual intervention is needed
	return fmt.Errorf("partial update state requires manual cleanup")
}

func (rm *RecoveryManager) cleanRollbackAndRetry(ctx *RecoveryContext) error {
	// First, perform a clean rollback
	if err := rm.rollbackToLatestBackup(ctx); err != nil {
		return fmt.Errorf("clean rollback failed: %w", err)
	}

	// Clean up any temporary files
	if err := rm.cleanupTemporaryFiles(); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to clean temporary files: %v\n", err)
	}

	// Reset state for retry
	return fmt.Errorf("rollback completed, manual retry required")
}

func (rm *RecoveryManager) fixPermissionsAndRetry(ctx *RecoveryContext) error {
	// Check and fix permissions on target directory
	targetDir := filepath.Dir(ctx.TargetPath)

	// Check if we can write to the target directory
	info, err := os.Stat(targetDir)
	if err != nil {
		return fmt.Errorf("cannot access target directory: %w", err)
	}

	// Check write permissions
	if info.Mode()&0200 == 0 {
		// Try to fix permissions (this might fail if we don't have permission)
		if err := os.Chmod(targetDir, info.Mode()|0200); err != nil {
			return fmt.Errorf("cannot fix directory permissions: %w", err)
		}
	}

	// Check target file permissions if it exists
	if info, err := os.Stat(ctx.TargetPath); err == nil {
		if info.Mode()&0200 == 0 {
			if err := os.Chmod(ctx.TargetPath, info.Mode()|0700); err != nil {
				return fmt.Errorf("cannot fix file permissions: %w", err)
			}
		}
	}

	return nil
}

func (rm *RecoveryManager) cleanTempFilesAndRetry(ctx *RecoveryContext) error {
	if err := rm.cleanupTemporaryFiles(); err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	// Check available disk space
	targetDir := filepath.Dir(ctx.TargetPath)
	if err := rm.checkDiskSpace(targetDir); err != nil {
		return fmt.Errorf("insufficient disk space after cleanup: %w", err)
	}

	return nil
}

func (rm *RecoveryManager) redownloadAndVerify(ctx *RecoveryContext) error {
	// This would involve re-downloading the binary and verifying checksums
	// For now, we'll indicate that a fresh download is needed
	return fmt.Errorf("re-download required, checksum verification failed")
}

func (rm *RecoveryManager) downloadCompatibleVersion(ctx *RecoveryContext) error {
	// Find a compatible version for the current system
	updateInfo, err := rm.FindCompatibleVersion()
	if err != nil {
		return fmt.Errorf("failed to find compatible version: %w", err)
	}

	// Get the platform for asset selection
	platform := version.DetectPlatform()
	asset, err := version.GetUpdateAsset(updateInfo.Release, platform)
	if err != nil {
		return fmt.Errorf("failed to get compatible asset: %w", err)
	}

	// Create download options
	downloadOpts := &download.DownloadOptions{
		URL:             asset.BrowserDownloadURL,
		DestinationPath: ctx.TargetPath + ".recovery",
		Resume:          false, // Fresh download for recovery
		ProgressCallback: func(total, transferred int64, done bool) {
			if total > 0 && !done {
				percent := (transferred * 100) / total
				fmt.Printf("Downloading compatible version: %d%%\r", percent)
			}
		},
	}

	// Download the compatible version
	fmt.Printf("Downloading compatible version %s for platform %s...\n",
		updateInfo.LatestVersion.String(), platform.String())

	result, err := rm.downloader.Download(downloadOpts)
	if err != nil {
		return fmt.Errorf("failed to download compatible version: %w", err)
	}

	// Validate the downloaded binary
	validatedPath, err := download.ValidateDownloadedBinary(result.Path, "")
	if err != nil {
		// Clean up failed download
		os.Remove(result.Path)
		return fmt.Errorf("downloaded binary validation failed: %w", err)
	}

	// Clean up extraction temp directory when done (no-op for direct binaries)
	if validatedPath != result.Path {
		defer os.RemoveAll(filepath.Dir(validatedPath))
	}

	// Atomically replace the target binary using the validated binary path
	if err := os.Rename(validatedPath, ctx.TargetPath); err != nil {
		// Clean up on failure
		os.Remove(validatedPath)
		return fmt.Errorf("failed to install compatible version: %w", err)
	}

	// Make executable on Unix systems
	if platform.OS != "windows" {
		if err := os.Chmod(ctx.TargetPath, 0755); err != nil {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	fmt.Printf("Successfully installed compatible version %s\n", updateInfo.LatestVersion.String())
	return nil
}

func (rm *RecoveryManager) emergencyRollback(ctx *RecoveryContext) error {
	// Last resort: rollback to any available backup
	backups, err := rm.backupManager.ListBackups()
	if err != nil {
		return fmt.Errorf("emergency rollback: no backups available: %w", err)
	}

	if len(backups) == 0 {
		return fmt.Errorf("emergency rollback: no backups found")
	}

	// Try each backup until one works
	for _, backup := range backups {
		opts := &RollbackOptions{
			TargetPath:     ctx.TargetPath,
			BackupPath:     backup.BackupPath,
			ValidateBackup: false, // Skip validation in emergency
			DryRun:         false,
		}

		if result, err := rm.rollbackManager.PerformRollback(opts); err == nil && result.Success {
			return nil
		}
	}

	return fmt.Errorf("emergency rollback: all backups failed")
}

// Helper methods

func (rm *RecoveryManager) cleanupTemporaryFiles() error {
	tempDir := os.TempDir()
	pattern := filepath.Join(tempDir, "pvm-*")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("finding temp files: %w", err)
	}

	for _, match := range matches {
		if err := os.RemoveAll(match); err != nil {
			// Log but continue with other files
			fmt.Printf("Failed to remove temp file %s: %v\n", match, err)
		}
	}

	return nil
}

func (rm *RecoveryManager) checkDiskSpace(dir string) error {
	// This is a simplified check - in practice, we'd use syscalls to check available space
	tempFile, err := os.CreateTemp(dir, "pvm-space-check-*")
	if err != nil {
		return fmt.Errorf("disk space check failed: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Try to write a small amount of data
	testData := make([]byte, 1024*1024) // 1MB
	if _, err := tempFile.Write(testData); err != nil {
		return fmt.Errorf("insufficient disk space: %w", err)
	}

	return nil
}

// DiagnoseFailure attempts to determine the failure scenario
func (rm *RecoveryManager) DiagnoseFailure(targetPath string, err error) RecoveryScenario {
	if err == nil {
		return ScenarioUnknownFailure
	}

	errorMsg := err.Error()

	// Pattern matching on error messages
	switch {
	case containsAnyString(errorMsg, "permission denied", "access denied"):
		return ScenarioPermissionDenied
	case containsAnyString(errorMsg, "no space", "disk full", "file too large"):
		return ScenarioFileSystemFull
	case containsAnyString(errorMsg, "checksum", "hash", "integrity"):
		return ScenarioChecksumMismatch
	case containsAnyString(errorMsg, "network", "connection", "timeout", "dns"):
		return ScenarioNetworkFailure
	case containsAnyString(errorMsg, "exec format", "cannot execute", "incompatible"):
		return ScenarioIncompatibleBinary
	case containsAnyString(errorMsg, "interrupted", "partial", "incomplete"):
		return ScenarioPartialUpdate
	}

	// Check if binary exists and is corrupted
	if info, statErr := os.Stat(targetPath); statErr == nil {
		if info.Size() == 0 || info.Mode()&0111 == 0 {
			return ScenarioCorruptedBinary
		}
	}

	return ScenarioUnknownFailure
}

// containsAnyString checks if message contains any of the given patterns using standard library
func containsAnyString(message string, patterns ...string) bool {
	for _, pattern := range patterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}
	return false
}

// CreateRecoveryContext creates a recovery context from an update failure
func (rm *RecoveryManager) CreateRecoveryContext(targetPath, failedVersion string, err error) *RecoveryContext {
	scenario := rm.DiagnoseFailure(targetPath, err)

	// Find the most recent backup for context
	var previousBackup string
	if backups, listErr := rm.backupManager.ListBackups(); listErr == nil && len(backups) > 0 {
		// Sort by creation time (newest first)
		sort.Slice(backups, func(i, j int) bool {
			return backups[i].CreatedAt.After(backups[j].CreatedAt)
		})
		previousBackup = backups[0].BackupPath
	}

	return &RecoveryContext{
		Scenario:       scenario,
		TargetPath:     targetPath,
		FailedVersion:  failedVersion,
		PreviousBackup: previousBackup,
		ErrorDetails:   err,
		AttemptCount:   0,
		MaxAttempts:    3, // Allow up to 3 recovery attempts to balance reliability vs. speed
	}
}

// FindCompatibleVersion finds and returns a compatible version for the current system
func (rm *RecoveryManager) FindCompatibleVersion() (*version.UpdateInfo, error) {
	// Detect current platform
	platform := version.DetectPlatform()
	if !platform.IsSupported() {
		return nil, fmt.Errorf("platform %s is not supported for updates", platform.String())
	}

	// Create GitHub client for querying releases
	client := version.NewGitHubClient()

	// Get available releases (including prereleases for recovery)
	releases, err := client.GetReleases("perigrin", "pvm", true)
	if err != nil {
		return nil, fmt.Errorf("failed to query available releases: %w", err)
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found in repository")
	}

	// Filter releases to find compatible versions
	var compatibleReleases []version.GitHubRelease
	for _, release := range releases {
		// Skip draft releases
		if release.Draft {
			continue
		}

		// Check if release has assets compatible with current platform
		if asset, err := version.GetUpdateAsset(&release, platform); err == nil && asset != nil {
			// Validate binary compatibility
			if err := version.ValidateBinaryCompatibility(asset, platform); err == nil {
				compatibleReleases = append(compatibleReleases, release)
			}
		}
	}

	if len(compatibleReleases) == 0 {
		return nil, fmt.Errorf("no compatible versions found for platform %s", platform.String())
	}

	// Select best compatible version (prefer stable over prereleases)
	var selectedRelease *version.GitHubRelease

	// First, try to find the latest stable release
	for i := range compatibleReleases {
		release := &compatibleReleases[i]
		if !release.Prerelease {
			if selectedRelease == nil || release.CreatedAt.After(selectedRelease.CreatedAt) {
				selectedRelease = release
			}
		}
	}

	// If no stable release found, select latest prerelease
	if selectedRelease == nil {
		for i := range compatibleReleases {
			release := &compatibleReleases[i]
			if selectedRelease == nil || release.CreatedAt.After(selectedRelease.CreatedAt) {
				selectedRelease = release
			}
		}
	}

	if selectedRelease == nil {
		return nil, fmt.Errorf("no suitable compatible version found")
	}

	// Parse version information
	currentVersion, err := version.ParseVersion("0.0.0") // Use dummy version for recovery
	if err != nil {
		return nil, fmt.Errorf("failed to parse current version: %w", err)
	}

	compatibleVersion, err := version.ParseVersion(selectedRelease.TagName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse compatible version %s: %w", selectedRelease.TagName, err)
	}

	// Create UpdateInfo for the compatible version
	updateInfo := &version.UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  compatibleVersion,
		Release:        selectedRelease,
		UpdateNeeded:   true,
		IsPrerelease:   selectedRelease.Prerelease,
	}

	return updateInfo, nil
}
