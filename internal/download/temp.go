// ABOUTME: Temporary file management and cleanup for download operations
// ABOUTME: Handles secure temporary files with proper permissions and atomic operations

package download

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tamarou.com/pvm/internal/diskspace"
)

// TempFileManager manages temporary files for downloads
type TempFileManager struct {
	tempDir string
	prefix  string
}

// NewTempFileManager creates a new temporary file manager
func NewTempFileManager() *TempFileManager {
	return &TempFileManager{
		tempDir: os.TempDir(),
		prefix:  "pvm-download-",
	}
}

// NewTempFileManagerWithDir creates a temporary file manager with custom directory
func NewTempFileManagerWithDir(tempDir string) *TempFileManager {
	return &TempFileManager{
		tempDir: tempDir,
		prefix:  "pvm-download-",
	}
}

// CreateTempFile creates a temporary file for downloading
func (tfm *TempFileManager) CreateTempFile(baseName string) (string, error) {
	// Ensure temp directory exists
	if err := os.MkdirAll(tfm.tempDir, 0755); err != nil {
		return "", fmt.Errorf("creating temp directory: %w", err)
	}

	// Generate temp file name
	timestamp := time.Now().Unix()
	tempName := fmt.Sprintf("%s%s-%d.tmp", tfm.prefix, baseName, timestamp)
	tempPath := filepath.Join(tfm.tempDir, tempName)

	// Create the file with secure permissions
	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	file.Close()

	return tempPath, nil
}

// AtomicMove moves a file from temporary location to final destination atomically
func (tfm *TempFileManager) AtomicMove(tempPath, finalPath string) error {
	// Ensure destination directory exists
	destDir := filepath.Dir(finalPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	// Check if source file exists
	if _, err := os.Stat(tempPath); err != nil {
		return fmt.Errorf("temp file not found: %w", err)
	}

	// Try atomic rename first (works if on same filesystem)
	err := os.Rename(tempPath, finalPath)
	if err == nil {
		return nil
	}

	// If rename fails (cross-filesystem), fall back to copy and delete
	return tfm.copyAndDelete(tempPath, finalPath)
}

// copyAndDelete copies a file and deletes the source (fallback for cross-filesystem moves)
func (tfm *TempFileManager) copyAndDelete(srcPath, dstPath string) error {
	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("getting source file info: %w", err)
	}

	// Create destination file
	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy data
	_, err = srcFile.WriteTo(dstFile)
	if err != nil {
		// Clean up partial destination file
		os.Remove(dstPath)
		return fmt.Errorf("copying file data: %w", err)
	}

	// Sync to ensure data is written
	if err := dstFile.Sync(); err != nil {
		os.Remove(dstPath)
		return fmt.Errorf("syncing destination file: %w", err)
	}

	// Remove source file
	if err := os.Remove(srcPath); err != nil {
		// Don't fail here - destination is complete
		// Just log the issue (in a real implementation)
	}

	return nil
}

// CleanupTempFile removes a temporary file
func (tfm *TempFileManager) CleanupTempFile(tempPath string) error {
	if tempPath == "" {
		return nil
	}

	err := os.Remove(tempPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing temp file: %w", err)
	}

	return nil
}

// CleanupOldTempFiles removes old temporary files from the temp directory
func (tfm *TempFileManager) CleanupOldTempFiles(maxAge time.Duration) error {
	entries, err := os.ReadDir(tfm.tempDir)
	if err != nil {
		return fmt.Errorf("reading temp directory: %w", err)
	}

	cutoff := time.Now().Add(-maxAge)

	for _, entry := range entries {
		// Only clean up our temp files
		if !strings.HasPrefix(entry.Name(), tfm.prefix) {
			continue
		}

		filePath := filepath.Join(tfm.tempDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		// Remove if older than cutoff
		if info.ModTime().Before(cutoff) {
			os.Remove(filePath)
		}
	}

	return nil
}

// ValidateDiskSpace checks if there's enough disk space for a download
func (tfm *TempFileManager) ValidateDiskSpace(requiredBytes int64) error {
	// Get available space in temp directory
	available, err := tfm.getAvailableSpace(tfm.tempDir)
	if err != nil {
		return fmt.Errorf("checking available disk space: %w", err)
	}

	// Add some buffer (10% extra)
	requiredWithBuffer := requiredBytes + (requiredBytes / 10)

	if available < requiredWithBuffer {
		return fmt.Errorf("insufficient disk space: need %d bytes, have %d bytes available",
			requiredWithBuffer, available)
	}

	return nil
}

// getAvailableSpace gets available disk space for a directory
func (tfm *TempFileManager) getAvailableSpace(dir string) (int64, error) {
	spaceInfo, err := diskspace.GetSpaceInfo(dir)
	if err != nil {
		return 0, fmt.Errorf("getting disk space information: %w", err)
	}

	return spaceInfo.Available, nil
}

// SecureDelete securely deletes a file by overwriting it before removal
func (tfm *TempFileManager) SecureDelete(filePath string) error {
	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Already gone
		}
		return fmt.Errorf("getting file info: %w", err)
	}

	// Only securely delete regular files
	if !info.Mode().IsRegular() {
		return os.Remove(filePath)
	}

	// Open file for writing
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("opening file for secure delete: %w", err)
	}
	defer file.Close()

	// Overwrite with zeros
	zeros := make([]byte, 4096)
	size := info.Size()
	for written := int64(0); written < size; {
		n := len(zeros)
		if size-written < int64(n) {
			n = int(size - written)
		}

		_, err := file.Write(zeros[:n])
		if err != nil {
			break // Best effort
		}
		written += int64(n)
	}

	// Sync to ensure data is written
	file.Sync()
	file.Close()

	// Finally remove the file
	return os.Remove(filePath)
}

// TempDownload represents a download in progress with temporary file management
type TempDownload struct {
	manager   *TempFileManager
	tempPath  string
	finalPath string
	cleanedUp bool
}

// NewTempDownload creates a new temporary download
func (tfm *TempFileManager) NewTempDownload(finalPath string) (*TempDownload, error) {
	baseName := filepath.Base(finalPath)
	tempPath, err := tfm.CreateTempFile(baseName)
	if err != nil {
		return nil, err
	}

	return &TempDownload{
		manager:   tfm,
		tempPath:  tempPath,
		finalPath: finalPath,
	}, nil
}

// TempPath returns the temporary file path
func (td *TempDownload) TempPath() string {
	return td.tempPath
}

// FinalPath returns the final destination path
func (td *TempDownload) FinalPath() string {
	return td.finalPath
}

// Commit moves the temporary file to its final location
func (td *TempDownload) Commit() error {
	if td.cleanedUp {
		return fmt.Errorf("download already cleaned up")
	}

	err := td.manager.AtomicMove(td.tempPath, td.finalPath)
	if err != nil {
		return err
	}

	td.cleanedUp = true
	return nil
}

// Cleanup removes the temporary file
func (td *TempDownload) Cleanup() error {
	if td.cleanedUp {
		return nil
	}

	err := td.manager.CleanupTempFile(td.tempPath)
	td.cleanedUp = true
	return err
}

// SetExecutable makes the temporary file executable (Unix only)
func (td *TempDownload) SetExecutable() error {
	return os.Chmod(td.tempPath, 0755)
}
