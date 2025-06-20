// ABOUTME: Rollback functionality for failed PVM updates
// ABOUTME: Handles automatic and manual rollback operations with validation

package updater

import (
	"fmt"
	"os"
	"time"
)

// RollbackManager handles rollback operations
type RollbackManager struct {
	backupManager *BackupManager
	replacer      *BinaryReplacer
}

// NewRollbackManager creates a new rollback manager
func NewRollbackManager() *RollbackManager {
	return &RollbackManager{
		backupManager: NewBackupManager(),
		replacer:      NewBinaryReplacer(),
	}
}

// RollbackOptions configures rollback behavior
type RollbackOptions struct {
	TargetPath     string // Path to restore the binary to
	BackupPath     string // Specific backup to restore from (optional)
	ValidateBackup bool   // Whether to validate backup before rollback
	DryRun         bool   // Whether to simulate rollback without actual changes
}

// RollbackResult contains information about the rollback operation
type RollbackResult struct {
	Success          bool          // Whether rollback was successful
	BackupPath       string        // Path to backup that was used
	Duration         time.Duration // Time taken for rollback
	ValidationPassed bool          // Whether backup validation passed
	SimulatedOnly    bool          // Whether this was a dry run
}

// PerformRollback performs a rollback operation
func (rm *RollbackManager) PerformRollback(opts *RollbackOptions) (*RollbackResult, error) {
	startTime := time.Now()

	result := &RollbackResult{
		SimulatedOnly: opts.DryRun,
	}

	if opts == nil {
		return result, fmt.Errorf("rollback options cannot be nil")
	}

	// Validate options
	if err := rm.validateRollbackOptions(opts); err != nil {
		return result, fmt.Errorf("invalid options: %w", err)
	}

	// Determine which backup to use
	backupPath := opts.BackupPath
	if backupPath == "" {
		// Find the most recent backup
		latestBackup, err := rm.findLatestBackup()
		if err != nil {
			return result, fmt.Errorf("finding latest backup: %w", err)
		}
		if latestBackup == nil {
			return result, fmt.Errorf("no backups available for rollback")
		}
		backupPath = latestBackup.BackupPath
	}

	result.BackupPath = backupPath

	// Validate backup if requested
	if opts.ValidateBackup {
		if err := rm.backupManager.ValidateBackup(backupPath); err != nil {
			return result, fmt.Errorf("backup validation failed: %w", err)
		}
		result.ValidationPassed = true
	}

	// Perform rollback (skip if dry run)
	if !opts.DryRun {
		if err := rm.performRollbackOperation(opts.TargetPath, backupPath); err != nil {
			return result, fmt.Errorf("rollback operation failed: %w", err)
		}
	}

	result.Success = true
	result.Duration = time.Since(startTime)
	return result, nil
}

// FindAvailableBackups returns a list of available backups for rollback
func (rm *RollbackManager) FindAvailableBackups() ([]*BackupMetadata, error) {
	return rm.backupManager.ListBackups()
}

// ValidateRollbackTarget validates that rollback can be performed to the target
func (rm *RollbackManager) ValidateRollbackTarget(targetPath string) error {
	// Check if target directory exists and is writable
	targetDir := targetPath
	if info, err := os.Stat(targetPath); err == nil && !info.IsDir() {
		// If target is a file, check its directory
		targetDir = targetPath[:len(targetPath)-len(info.Name())-1]
	}

	// Check if directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("target directory does not exist: %s", targetDir)
	}

	// Check if we can write to the target location
	// Try creating a temporary file to test write permissions
	tempFile, err := os.CreateTemp(targetDir, ".pvm-rollback-test-*")
	if err != nil {
		return fmt.Errorf("cannot write to target directory: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	return nil
}

// AutoRollback performs automatic rollback after a failed update
func (rm *RollbackManager) AutoRollback(targetPath string) error {
	opts := &RollbackOptions{
		TargetPath:     targetPath,
		ValidateBackup: true,
		DryRun:         false,
	}

	result, err := rm.PerformRollback(opts)
	if err != nil {
		return fmt.Errorf("auto-rollback failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("auto-rollback did not complete successfully")
	}

	return nil
}

// GetRollbackStatus checks if a rollback is needed or possible
func (rm *RollbackManager) GetRollbackStatus(targetPath string) (*RollbackStatus, error) {
	status := &RollbackStatus{
		TargetPath: targetPath,
	}

	// Check if target exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		status.TargetExists = false
		status.RollbackNeeded = true
		status.Reason = "target binary is missing"
	} else {
		status.TargetExists = true
	}

	// Find available backups
	backups, err := rm.backupManager.ListBackups()
	if err != nil {
		return status, fmt.Errorf("listing backups: %w", err)
	}

	status.AvailableBackups = len(backups)
	if len(backups) == 0 {
		status.RollbackPossible = false
		status.Reason = "no backups available"
		return status, nil
	}

	status.RollbackPossible = true

	// Find the most recent backup
	var latestBackup *BackupMetadata
	for _, backup := range backups {
		if latestBackup == nil || backup.CreatedAt.After(latestBackup.CreatedAt) {
			latestBackup = backup
		}
	}

	if latestBackup != nil {
		status.LatestBackupDate = latestBackup.CreatedAt
		status.LatestBackupPath = latestBackup.BackupPath
	}

	return status, nil
}

// RollbackStatus contains information about rollback availability
type RollbackStatus struct {
	TargetPath       string    `json:"target_path"`
	TargetExists     bool      `json:"target_exists"`
	RollbackNeeded   bool      `json:"rollback_needed"`
	RollbackPossible bool      `json:"rollback_possible"`
	AvailableBackups int       `json:"available_backups"`
	LatestBackupDate time.Time `json:"latest_backup_date,omitempty"`
	LatestBackupPath string    `json:"latest_backup_path,omitempty"`
	Reason           string    `json:"reason,omitempty"`
}

// validateRollbackOptions validates the rollback options
func (rm *RollbackManager) validateRollbackOptions(opts *RollbackOptions) error {
	if opts.TargetPath == "" {
		return fmt.Errorf("target path cannot be empty")
	}

	// If a specific backup is specified, check if it exists
	if opts.BackupPath != "" {
		if _, err := os.Stat(opts.BackupPath); os.IsNotExist(err) {
			return fmt.Errorf("specified backup not found: %s", opts.BackupPath)
		}
	}

	return nil
}

// findLatestBackup finds the most recent backup
func (rm *RollbackManager) findLatestBackup() (*BackupMetadata, error) {
	backups, err := rm.backupManager.ListBackups()
	if err != nil {
		return nil, fmt.Errorf("listing backups: %w", err)
	}

	if len(backups) == 0 {
		return nil, nil
	}

	// Find the most recent backup
	var latestBackup *BackupMetadata
	for _, backup := range backups {
		if latestBackup == nil || backup.CreatedAt.After(latestBackup.CreatedAt) {
			latestBackup = backup
		}
	}

	return latestBackup, nil
}

// performRollbackOperation performs the actual rollback
func (rm *RollbackManager) performRollbackOperation(targetPath, backupPath string) error {
	// Use the atomic replacement logic to restore from backup
	opts := &ReplacementOptions{
		CurrentPath:    targetPath,
		NewPath:        backupPath,
		BackupEnabled:  false, // Don't backup when rolling back
		ValidateBinary: true,  // Validate the backup before using it
		DryRun:         false,
	}

	result, err := rm.replacer.ReplaceBinary(opts)
	if err != nil {
		return fmt.Errorf("atomic rollback failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("rollback operation did not complete successfully")
	}

	return nil
}

// CleanupAfterRollback performs cleanup after a successful rollback
func (rm *RollbackManager) CleanupAfterRollback(usedBackupPath string) error {
	// Optionally remove the backup that was used for rollback
	// This is conservative - we might want to keep it for forensic purposes
	// For now, we'll leave the backup in place

	// Could add logic here to:
	// 1. Mark the backup as "used for rollback"
	// 2. Move it to a different location
	// 3. Update metadata

	return nil
}

// ValidatePostRollback validates the system after a rollback
func (rm *RollbackManager) ValidatePostRollback(targetPath string) error {
	// Check if target binary exists and is executable
	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("rollback target validation failed: %w", err)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("rollback target is not a regular file")
	}

	// On Unix-like systems, check if it's executable
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("rollback target is not executable")
	}

	// Could add more validation here:
	// 1. Try running the binary with --version
	// 2. Check binary format/magic bytes
	// 3. Verify digital signature

	return nil
}
