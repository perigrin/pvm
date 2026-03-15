// ABOUTME: Cross-platform file permissions and ownership handling for PVM updates
// ABOUTME: Provides platform-specific implementations for permission preservation

package updater

import (
	"fmt"
	"os"
	"runtime"
)

// PreserveExtendedAttributes preserves extended attributes and metadata
func PreserveExtendedAttributes(src, dst string) error {
	switch runtime.GOOS {
	case "darwin":
		return preserveMacOSAttributes(src, dst)
	case "linux":
		return preserveLinuxAttributes(src, dst)
	case "windows":
		return preserveWindowsAttributes(src, dst)
	default:
		// For other platforms, just preserve basic permissions
		return preserveBasicPermissions(src, dst)
	}
}

// preserveBasicPermissions preserves just file mode/permissions
func preserveBasicPermissions(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("getting source file info: %w", err)
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// preserveMacOSAttributes preserves macOS-specific attributes
func preserveMacOSAttributes(src, dst string) error {
	// First preserve basic permissions
	if err := preserveBasicPermissions(src, dst); err != nil {
		return err
	}

	// TODO: Preserve extended attributes, quarantine flags, etc.
	// This would require platform-specific code using syscalls or cgo
	// For now, we'll just preserve basic permissions

	return nil
}

// preserveLinuxAttributes preserves Linux-specific attributes
func preserveLinuxAttributes(src, dst string) error {
	// First preserve basic permissions
	if err := preserveBasicPermissions(src, dst); err != nil {
		return err
	}

	// TODO: Preserve extended attributes, ACLs, etc.
	// This would require platform-specific code using syscalls
	// For now, we'll just preserve basic permissions

	return nil
}

// preserveWindowsAttributes preserves Windows-specific attributes
func preserveWindowsAttributes(src, dst string) error {
	// First preserve basic permissions
	if err := preserveBasicPermissions(src, dst); err != nil {
		return err
	}

	// TODO: Preserve Windows file attributes, ACLs, etc.
	// This would require Windows-specific code
	// For now, we'll just preserve basic permissions

	return nil
}

// EnsureExecutable ensures the file is executable on the current platform
func EnsureExecutable(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("getting file info: %w", err)
	}

	switch runtime.GOOS {
	case "windows":
		// On Windows, executability is determined by file extension
		// .exe files should already be executable
		return nil
	default:
		// On Unix-like systems, ensure executable bit is set
		mode := info.Mode()
		if mode&0111 == 0 {
			// Add executable permissions for owner, group, and others
			newMode := mode | 0111
			return os.Chmod(filePath, newMode)
		}
		return nil
	}
}

// CheckWritePermissions checks if we have permission to write to the target path
func CheckWritePermissions(targetPath string) error {
	// Check if the file exists
	if info, err := os.Stat(targetPath); err == nil {
		// File exists - check if we can write to it
		if info.Mode()&0200 == 0 {
			return fmt.Errorf("no write permission for file: %s", targetPath)
		}

		// Try to open for writing to be sure
		file, err := os.OpenFile(targetPath, os.O_WRONLY, 0)
		if err != nil {
			return fmt.Errorf("cannot open file for writing: %w", err)
		}
		file.Close()

		return nil
	} else if os.IsNotExist(err) {
		// File doesn't exist - check if we can write to the directory
		dir := targetPath
		if info, err := os.Stat(targetPath); err == nil && !info.IsDir() {
			dir = targetPath[:len(targetPath)-len(info.Name())-1]
		}

		// Check directory write permissions
		dirInfo, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("target directory does not exist: %s", dir)
		}

		if dirInfo.Mode()&0200 == 0 {
			return fmt.Errorf("no write permission for directory: %s", dir)
		}

		return nil
	} else {
		return fmt.Errorf("error checking target path: %w", err)
	}
}

// IsRunningWithSufficientPrivileges checks if we have sufficient privileges for update
func IsRunningWithSufficientPrivileges(targetPath string) (bool, error) {
	// Check write permissions to target
	if err := CheckWritePermissions(targetPath); err != nil {
		return false, err
	}

	// Additional privilege checks could go here:
	// - Check if running as administrator on Windows
	// - Check specific capabilities on Linux
	// - Check code signing permissions on macOS

	return true, nil
}

// RequireElevation determines if elevation is required for the update
func RequireElevation(targetPath string) (bool, string) {
	// Check current permissions
	sufficient, err := IsRunningWithSufficientPrivileges(targetPath)
	if err != nil {
		return true, fmt.Sprintf("Permission check failed: %v", err)
	}

	if sufficient {
		return false, ""
	}

	// Provide platform-specific elevation instructions
	switch runtime.GOOS {
	case "windows":
		return true, "Run as Administrator to update PVM"
	case "darwin":
		return true, "Use 'sudo pvm update' to update PVM with administrator privileges"
	default:
		return true, "Use 'sudo pvm update' to update PVM with root privileges"
	}
}
