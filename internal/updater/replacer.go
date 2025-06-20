// ABOUTME: Atomic binary replacement logic for PVM self-updater
// ABOUTME: Handles safe binary replacement with proper file operations and error handling

package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// BinaryReplacer handles atomic replacement of the PVM binary
type BinaryReplacer struct {
	backupManager *BackupManager
}

// NewBinaryReplacer creates a new binary replacer
func NewBinaryReplacer() *BinaryReplacer {
	return &BinaryReplacer{
		backupManager: NewBackupManager(),
	}
}

// ReplacementOptions configures binary replacement behavior
type ReplacementOptions struct {
	CurrentPath    string // Path to current PVM binary
	NewPath        string // Path to new binary file
	BackupEnabled  bool   // Whether to create backup before replacement
	ValidateBinary bool   // Whether to validate the new binary before replacement
	DryRun         bool   // Whether to simulate replacement without actual changes
}

// ReplacementResult contains information about the replacement operation
type ReplacementResult struct {
	Success       bool          // Whether replacement was successful
	BackupPath    string        // Path to backup file (if created)
	Duration      time.Duration // Time taken for replacement
	ValidationOK  bool          // Whether binary validation passed
	SimulatedOnly bool          // Whether this was a dry run
}

// ReplaceBinary performs atomic binary replacement
func (r *BinaryReplacer) ReplaceBinary(opts *ReplacementOptions) (*ReplacementResult, error) {
	startTime := time.Now()

	result := &ReplacementResult{
		SimulatedOnly: opts.DryRun,
	}

	if opts == nil {
		return result, fmt.Errorf("replacement options cannot be nil")
	}

	// Validate inputs
	if err := r.validateReplacementOptions(opts); err != nil {
		return result, fmt.Errorf("invalid options: %w", err)
	}

	// Check if running process would be affected
	if err := r.checkRunningProcess(opts.CurrentPath); err != nil {
		return result, fmt.Errorf("running process check failed: %w", err)
	}

	// Validate new binary if requested
	if opts.ValidateBinary {
		if err := r.validateNewBinary(opts.NewPath); err != nil {
			return result, fmt.Errorf("binary validation failed: %w", err)
		}
		result.ValidationOK = true
	}

	// Create backup if enabled
	if opts.BackupEnabled {
		backupPath, err := r.backupManager.CreateBackup(opts.CurrentPath)
		if err != nil {
			return result, fmt.Errorf("backup creation failed: %w", err)
		}
		result.BackupPath = backupPath
	}

	// Perform replacement (skip if dry run)
	if !opts.DryRun {
		if err := r.performAtomicReplacement(opts.CurrentPath, opts.NewPath); err != nil {
			// If replacement fails and we have a backup, attempt rollback
			if result.BackupPath != "" {
				rollbackErr := r.performRollback(opts.CurrentPath, result.BackupPath)
				if rollbackErr != nil {
					return result, fmt.Errorf("replacement failed and rollback failed: %w (original error: %v)", rollbackErr, err)
				}
			}
			return result, fmt.Errorf("atomic replacement failed: %w", err)
		}
	}

	result.Success = true
	result.Duration = time.Since(startTime)
	return result, nil
}

// validateReplacementOptions validates the replacement options
func (r *BinaryReplacer) validateReplacementOptions(opts *ReplacementOptions) error {
	if opts.CurrentPath == "" {
		return fmt.Errorf("current binary path cannot be empty")
	}

	if opts.NewPath == "" {
		return fmt.Errorf("new binary path cannot be empty")
	}

	// Check if current binary exists
	if _, err := os.Stat(opts.CurrentPath); os.IsNotExist(err) {
		return fmt.Errorf("current binary not found: %s", opts.CurrentPath)
	}

	// Check if new binary exists
	if _, err := os.Stat(opts.NewPath); os.IsNotExist(err) {
		return fmt.Errorf("new binary not found: %s", opts.NewPath)
	}

	return nil
}

// checkRunningProcess checks if the current process might be affected by replacement
func (r *BinaryReplacer) checkRunningProcess(currentPath string) error {
	// Get the executable path of the current process
	execPath, err := os.Executable()
	if err != nil {
		// If we can't determine the executable path, proceed with caution
		return nil
	}

	// Resolve any symlinks
	resolvedExecPath, err := filepath.EvalSymlinks(execPath)
	if err == nil {
		execPath = resolvedExecPath
	}

	resolvedCurrentPath, err := filepath.EvalSymlinks(currentPath)
	if err == nil {
		currentPath = resolvedCurrentPath
	}

	// Check if we're trying to replace the currently running binary
	if execPath == currentPath {
		// On Windows, this would likely fail due to file locking
		if runtime.GOOS == "windows" {
			return fmt.Errorf("cannot replace currently running binary on Windows")
		}
		// On Unix-like systems, this is generally safe but warn the user
		// The replacement will work, but the current process will continue
		// running the old binary until it exits
	}

	return nil
}

// validateNewBinary performs basic validation of the new binary
func (r *BinaryReplacer) validateNewBinary(newPath string) error {
	// Check file permissions
	info, err := os.Stat(newPath)
	if err != nil {
		return fmt.Errorf("getting file info: %w", err)
	}

	// Ensure it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("new binary is not a regular file")
	}

	// Check if it's executable (on Unix-like systems)
	if runtime.GOOS != "windows" {
		if info.Mode()&0111 == 0 {
			return fmt.Errorf("new binary is not executable")
		}
	}

	// Basic size check (executable should be reasonably sized)
	minSize := int64(1024 * 1024)       // 1MB minimum
	maxSize := int64(100 * 1024 * 1024) // 100MB maximum

	if info.Size() < minSize {
		return fmt.Errorf("new binary is too small (%d bytes, minimum %d)", info.Size(), minSize)
	}

	if info.Size() > maxSize {
		return fmt.Errorf("new binary is too large (%d bytes, maximum %d)", info.Size(), maxSize)
	}

	return nil
}

// performAtomicReplacement performs the actual atomic replacement
func (r *BinaryReplacer) performAtomicReplacement(currentPath, newPath string) error {
	// Get directory of current binary
	currentDir := filepath.Dir(currentPath)
	currentName := filepath.Base(currentPath)

	// Create temporary name for atomic operation
	tempName := fmt.Sprintf(".%s.tmp.%d", currentName, time.Now().UnixNano())
	tempPath := filepath.Join(currentDir, tempName)

	// Copy new binary to temporary location in same directory
	// This ensures the rename operation will be atomic on most filesystems
	if err := r.copyFile(newPath, tempPath); err != nil {
		return fmt.Errorf("copying new binary to temp location: %w", err)
	}

	// Ensure cleanup of temp file if something goes wrong
	defer func() {
		if _, err := os.Stat(tempPath); err == nil {
			os.Remove(tempPath)
		}
	}()

	// Copy permissions from original file
	if err := r.copyPermissions(currentPath, tempPath); err != nil {
		return fmt.Errorf("copying permissions: %w", err)
	}

	// Perform atomic rename
	if err := os.Rename(tempPath, currentPath); err != nil {
		return fmt.Errorf("atomic rename failed: %w", err)
	}

	return nil
}

// performRollback rolls back to the backup binary
func (r *BinaryReplacer) performRollback(currentPath, backupPath string) error {
	// Use the same atomic replacement logic for rollback
	return r.performAtomicReplacement(currentPath, backupPath)
}

// copyFile copies a file from src to dst
func (r *BinaryReplacer) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy file contents
	_, err = srcFile.WriteTo(dstFile)
	if err != nil {
		return fmt.Errorf("copying file contents: %w", err)
	}

	// Ensure data is written to disk
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("syncing destination file: %w", err)
	}

	return nil
}

// copyPermissions copies file permissions and ownership from src to dst
func (r *BinaryReplacer) copyPermissions(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("getting source file info: %w", err)
	}

	// Copy file mode/permissions
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("setting file permissions: %w", err)
	}

	// On Unix-like systems, try to copy ownership
	if runtime.GOOS != "windows" {
		if err := r.copyOwnership(src, dst); err != nil {
			// Don't fail on ownership copy errors - it might not be possible
			// if running as non-root user
			_ = err
		}
	}

	return nil
}

// copyOwnership copies file ownership from src to dst (Unix-like systems only)
func (r *BinaryReplacer) copyOwnership(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// This is platform-specific and might not work in all cases
	// We'll implement a basic version that works on most Unix-like systems
	return r.chownFile(dst, srcInfo)
}

// GetCurrentBinaryPath attempts to determine the path of the currently running PVM binary
func GetCurrentBinaryPath() (string, error) {
	// Get the executable path
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("getting executable path: %w", err)
	}

	// Resolve symlinks to get the actual binary path
	resolvedPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		// If we can't resolve symlinks, use the original path
		return execPath, nil
	}

	return resolvedPath, nil
}

// DetectInstallationMethod determines how PVM was installed
func DetectInstallationMethod(binaryPath string) (InstallationMethod, error) {
	// Check for common package manager installation paths
	lowerPath := strings.ToLower(binaryPath)

	// Check for Homebrew (macOS and Linux)
	if strings.Contains(lowerPath, "/homebrew/") ||
		strings.Contains(lowerPath, "/usr/local/") ||
		strings.Contains(lowerPath, "/opt/homebrew/") {
		return InstallationHomebrew, nil
	}

	// Check for system package managers (Linux)
	if strings.HasPrefix(lowerPath, "/usr/bin/") ||
		strings.HasPrefix(lowerPath, "/usr/local/bin/") {
		// Could be APT, YUM, or other system package manager
		// For now, we'll classify as binary installation
		return InstallationBinary, nil
	}

	// Check for Windows package managers
	if runtime.GOOS == "windows" {
		if strings.Contains(lowerPath, "chocolatey") {
			return InstallationBinary, nil // Treat as binary for now
		}
		if strings.Contains(lowerPath, "scoop") {
			return InstallationBinary, nil // Treat as binary for now
		}
	}

	// Default to binary installation
	return InstallationBinary, nil
}

// InstallationMethod represents how PVM was installed
type InstallationMethod int

const (
	InstallationBinary InstallationMethod = iota
	InstallationHomebrew
	InstallationAPT
	InstallationYum
	InstallationPacman
	InstallationChocolatey
	InstallationScoop
)

func (i InstallationMethod) String() string {
	switch i {
	case InstallationBinary:
		return "binary"
	case InstallationHomebrew:
		return "homebrew"
	case InstallationAPT:
		return "apt"
	case InstallationYum:
		return "yum"
	case InstallationPacman:
		return "pacman"
	case InstallationChocolatey:
		return "chocolatey"
	case InstallationScoop:
		return "scoop"
	default:
		return "unknown"
	}
}

// CanSelfUpdate returns whether PVM can self-update given the installation method
func (i InstallationMethod) CanSelfUpdate() bool {
	switch i {
	case InstallationBinary:
		return true
	case InstallationHomebrew:
		return false // Should use 'brew upgrade pvm' instead
	case InstallationAPT, InstallationYum, InstallationPacman:
		return false // Should use system package manager
	case InstallationChocolatey, InstallationScoop:
		return false // Should use respective package manager
	default:
		return false
	}
}

// GetUpdateInstructions returns instructions for updating via the installation method
func (i InstallationMethod) GetUpdateInstructions() string {
	switch i {
	case InstallationBinary:
		return "Use 'pvm update' to update to the latest version"
	case InstallationHomebrew:
		return "Use 'brew upgrade pvm' to update via Homebrew"
	case InstallationAPT:
		return "Use 'sudo apt update && sudo apt upgrade pvm' to update via APT"
	case InstallationYum:
		return "Use 'sudo yum update pvm' to update via YUM"
	case InstallationPacman:
		return "Use 'sudo pacman -Syu pvm' to update via Pacman"
	case InstallationChocolatey:
		return "Use 'choco upgrade pvm' to update via Chocolatey"
	case InstallationScoop:
		return "Use 'scoop update pvm' to update via Scoop"
	default:
		return "Update method unknown for this installation"
	}
}
