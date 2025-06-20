// ABOUTME: Main updater orchestration for PVM self-updater
// ABOUTME: Coordinates version checking, downloading, and atomic replacement

package updater

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tamarou.com/pvm/internal/download"
	"tamarou.com/pvm/internal/version"
)

// Updater orchestrates the complete update process
type Updater struct {
	versionChecker *version.GitHubClient
	downloader     *download.Downloader
	replacer       *BinaryReplacer
	rollbackMgr    *RollbackManager
}

// NewUpdater creates a new updater instance
func NewUpdater() *Updater {
	return &Updater{
		versionChecker: version.NewGitHubClient(),
		downloader:     download.NewDownloader(),
		replacer:       NewBinaryReplacer(),
		rollbackMgr:    NewRollbackManager(),
	}
}

// NewUpdaterWithToken creates a new updater with GitHub authentication
func NewUpdaterWithToken(token string) *Updater {
	return &Updater{
		versionChecker: version.NewGitHubClientWithToken(token),
		downloader:     download.NewDownloader(),
		replacer:       NewBinaryReplacer(),
		rollbackMgr:    NewRollbackManager(),
	}
}

// UpdateOptions configures the update process
type UpdateOptions struct {
	// Version selection
	TargetVersion     string // Specific version to update to (empty = latest)
	IncludePrerelease bool   // Whether to consider pre-release versions

	// Repository configuration
	Repository  string // GitHub repository (owner/repo format)
	GitHubToken string // Optional GitHub token for higher rate limits

	// Update behavior
	Force        bool // Skip version checks and force update
	DryRun       bool // Simulate update without making changes
	Backup       bool // Create backup before updating (default: true)
	AutoRollback bool // Automatically rollback on failure (default: true)

	// Installation method handling
	IgnoreInstallMethod bool // Ignore installation method and force self-update

	// Progress reporting
	ProgressCallback func(stage UpdateStage, message string, progress float64)

	// Context for cancellation
	Context context.Context
}

// DefaultUpdateOptions returns default update options
func DefaultUpdateOptions() *UpdateOptions {
	return &UpdateOptions{
		Repository:   "perigrin/pvm-dev",
		Backup:       true,
		AutoRollback: true,
		Context:      context.Background(),
	}
}

// UpdateStage represents different stages of the update process
type UpdateStage int

const (
	StageCheckingVersion UpdateStage = iota
	StageDetectingPlatform
	StageDownloading
	StageValidating
	StageCreatingBackup
	StageReplacing
	StageValidatingUpdate
	StageCleaningUp
	StageRollingBack
	StageDone
)

func (s UpdateStage) String() string {
	switch s {
	case StageCheckingVersion:
		return "Checking for updates"
	case StageDetectingPlatform:
		return "Detecting platform"
	case StageDownloading:
		return "Downloading update"
	case StageValidating:
		return "Validating download"
	case StageCreatingBackup:
		return "Creating backup"
	case StageReplacing:
		return "Installing update"
	case StageValidatingUpdate:
		return "Validating installation"
	case StageCleaningUp:
		return "Cleaning up"
	case StageRollingBack:
		return "Rolling back"
	case StageDone:
		return "Update complete"
	default:
		return "Unknown stage"
	}
}

// UpdateResult contains information about the update operation
type UpdateResult struct {
	Success           bool          `json:"success"`
	UpdatePerformed   bool          `json:"update_performed"`
	PreviousVersion   string        `json:"previous_version"`
	NewVersion        string        `json:"new_version"`
	Duration          time.Duration `json:"duration"`
	BackupCreated     bool          `json:"backup_created"`
	BackupPath        string        `json:"backup_path,omitempty"`
	RollbackPerformed bool          `json:"rollback_performed"`
	DryRun            bool          `json:"dry_run"`
	Message           string        `json:"message"`
}

// CheckForUpdates checks if an update is available
func (u *Updater) CheckForUpdates(opts *UpdateOptions) (*version.UpdateInfo, error) {
	if opts == nil {
		opts = DefaultUpdateOptions()
	}

	// Create version check options
	checkOpts := &version.CheckOptions{
		IncludePrerelease: opts.IncludePrerelease,
		Repository:        opts.Repository,
		GitHubToken:       opts.GitHubToken,
		Timeout:           30 * time.Second,
	}

	return version.GetUpdateInfo(checkOpts)
}

// PerformUpdate performs the complete update process
func (u *Updater) PerformUpdate(opts *UpdateOptions) (*UpdateResult, error) {
	startTime := time.Now()

	result := &UpdateResult{
		DryRun: opts.DryRun,
	}

	if opts == nil {
		opts = DefaultUpdateOptions()
	}

	if opts.Context == nil {
		opts.Context = context.Background()
	}

	// Report progress
	reportProgress := func(stage UpdateStage, message string, progress float64) {
		if opts.ProgressCallback != nil {
			opts.ProgressCallback(stage, message, progress)
		}
	}

	// Get current binary path
	reportProgress(StageCheckingVersion, "Detecting current installation", 0.0)
	currentPath, err := GetCurrentBinaryPath()
	if err != nil {
		result.Message = fmt.Sprintf("Failed to detect current binary: %v", err)
		return result, err
	}

	// Check installation method
	if !opts.IgnoreInstallMethod {
		installMethod, err := DetectInstallationMethod(currentPath)
		if err != nil {
			result.Message = fmt.Sprintf("Failed to detect installation method: %v", err)
			return result, err
		}

		if !installMethod.CanSelfUpdate() {
			result.Message = fmt.Sprintf("Cannot self-update: %s. %s",
				installMethod.String(), installMethod.GetUpdateInstructions())
			return result, fmt.Errorf("self-update not supported for %s installations", installMethod.String())
		}
	}

	// Check for updates
	reportProgress(StageCheckingVersion, "Checking for updates", 0.1)
	updateInfo, err := u.CheckForUpdates(opts)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to check for updates: %v", err)
		return result, err
	}

	result.PreviousVersion = updateInfo.CurrentVersion.String()

	// Check if update is needed
	if !updateInfo.UpdateNeeded && !opts.Force {
		result.Message = fmt.Sprintf("Already up to date (version %s)", updateInfo.CurrentVersion.String())
		result.Success = true
		result.Duration = time.Since(startTime)
		return result, nil
	}

	result.NewVersion = updateInfo.LatestVersion.String()

	// Detect platform
	reportProgress(StageDetectingPlatform, "Detecting platform", 0.2)
	platform := version.DetectPlatform()
	if !platform.IsSupported() {
		result.Message = fmt.Sprintf("Platform %s is not supported for updates", platform.String())
		return result, fmt.Errorf("unsupported platform: %s", platform.String())
	}

	// Get update asset
	asset, err := version.GetUpdateAsset(updateInfo.Release, platform)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to find update asset: %v", err)
		return result, err
	}

	// Download update
	reportProgress(StageDownloading, fmt.Sprintf("Downloading %s", asset.Name), 0.3)
	tempDir, err := os.MkdirTemp("", "pvm-update-*")
	if err != nil {
		result.Message = fmt.Sprintf("Failed to create temp directory: %v", err)
		return result, err
	}
	defer os.RemoveAll(tempDir)

	downloadPath := filepath.Join(tempDir, asset.Name)
	downloadOpts := &download.DownloadOptions{
		URL:             asset.BrowserDownloadURL,
		DestinationPath: downloadPath,
		Context:         opts.Context,
		ProgressCallback: func(total, transferred int64, done bool) {
			if total > 0 {
				progress := 0.3 + (0.3 * float64(transferred) / float64(total))
				reportProgress(StageDownloading, fmt.Sprintf("Downloaded %d/%d bytes", transferred, total), progress)
			}
		},
	}

	downloadResult, err := u.downloader.Download(downloadOpts)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to download update: %v", err)
		return result, err
	}

	// Validate downloaded binary
	reportProgress(StageValidating, "Validating download", 0.6)
	if err := download.ValidateDownloadedBinary(downloadResult.Path, ""); err != nil {
		result.Message = fmt.Sprintf("Downloaded binary validation failed: %v", err)
		return result, err
	}

	// Perform replacement (or simulate if dry run)
	if opts.DryRun {
		result.Message = fmt.Sprintf("Dry run: would update from %s to %s",
			result.PreviousVersion, result.NewVersion)
		result.Success = true
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// Create backup
	var backupPath string
	if opts.Backup {
		reportProgress(StageCreatingBackup, "Creating backup", 0.7)
		backupMgr := NewBackupManager()
		backupPath, err = backupMgr.CreateBackup(currentPath)
		if err != nil {
			result.Message = fmt.Sprintf("Failed to create backup: %v", err)
			return result, err
		}
		result.BackupCreated = true
		result.BackupPath = backupPath
	}

	// Perform atomic replacement
	reportProgress(StageReplacing, "Installing update", 0.8)
	replaceOpts := &ReplacementOptions{
		CurrentPath:    currentPath,
		NewPath:        downloadResult.Path,
		BackupEnabled:  false, // We already created backup above
		ValidateBinary: true,
		DryRun:         false,
	}

	replaceResult, err := u.replacer.ReplaceBinary(replaceOpts)
	if err != nil {
		// Auto-rollback if enabled
		if opts.AutoRollback && backupPath != "" {
			reportProgress(StageRollingBack, "Rolling back due to failure", 0.9)
			rollbackErr := u.rollbackMgr.AutoRollback(currentPath)
			if rollbackErr != nil {
				result.Message = fmt.Sprintf("Update failed and rollback failed: %v (original error: %v)", rollbackErr, err)
			} else {
				result.RollbackPerformed = true
				result.Message = fmt.Sprintf("Update failed, rolled back to previous version: %v", err)
			}
		} else {
			result.Message = fmt.Sprintf("Update failed: %v", err)
		}
		return result, err
	}

	if !replaceResult.Success {
		result.Message = "Binary replacement did not complete successfully"
		return result, fmt.Errorf("replacement failed")
	}

	// Validate update
	reportProgress(StageValidatingUpdate, "Validating installation", 0.9)
	if err := u.rollbackMgr.ValidatePostRollback(currentPath); err != nil {
		result.Message = fmt.Sprintf("Update validation failed: %v", err)

		// Auto-rollback if enabled
		if opts.AutoRollback && backupPath != "" {
			reportProgress(StageRollingBack, "Rolling back due to validation failure", 0.95)
			rollbackErr := u.rollbackMgr.AutoRollback(currentPath)
			if rollbackErr != nil {
				result.Message = fmt.Sprintf("Update validation failed and rollback failed: %v (original error: %v)", rollbackErr, err)
			} else {
				result.RollbackPerformed = true
				result.Message = fmt.Sprintf("Update validation failed, rolled back to previous version: %v", err)
			}
		}
		return result, err
	}

	// Cleanup
	reportProgress(StageCleaningUp, "Cleaning up", 0.95)
	// Temporary files will be cleaned up by deferred os.RemoveAll

	reportProgress(StageDone, "Update complete", 1.0)
	result.Success = true
	result.UpdatePerformed = true
	result.Duration = time.Since(startTime)
	result.Message = fmt.Sprintf("Successfully updated from %s to %s",
		result.PreviousVersion, result.NewVersion)

	return result, nil
}

// Rollback performs a manual rollback to a previous version
func (u *Updater) Rollback(targetPath string, backupPath string) error {
	opts := &RollbackOptions{
		TargetPath:     targetPath,
		BackupPath:     backupPath,
		ValidateBackup: true,
		DryRun:         false,
	}

	result, err := u.rollbackMgr.PerformRollback(opts)
	if err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("rollback operation did not complete successfully")
	}

	return nil
}
