// ABOUTME: Cross-platform compatibility utilities for PVM
// ABOUTME: Provides unified interfaces for platform-specific operations

package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

// IsWindows returns true if running on Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// ExecutableExtension returns the platform-appropriate executable extension
func ExecutableExtension() string {
	if IsWindows() {
		return ".exe"
	}
	return ""
}

// ExecutableName returns the platform-appropriate executable name
func ExecutableName(base string) string {
	return base + ExecutableExtension()
}

// ScriptExtension returns the platform-appropriate script extension
func ScriptExtension() string {
	if IsWindows() {
		return ".bat"
	}
	return ""
}

// PathSeparator returns the platform-appropriate path separator
func PathSeparator() string {
	return string(os.PathSeparator)
}

// CanCreateSymlinks returns true if the platform supports symlink creation
// without elevated privileges
func CanCreateSymlinks() bool {
	// On Windows, symlinks require developer mode or admin privileges
	// On Unix systems, any user can create symlinks
	return !IsWindows()
}

// CreateLink creates either a symlink (Unix) or hard link/copy (Windows)
func CreateLink(target, link string) error {
	if IsWindows() {
		// Try hard link first, fall back to copy
		err := os.Link(target, link)
		if err != nil {
			// If hard link fails, copy the file
			return copyFile(target, link)
		}
		return nil
	}
	// On Unix, create a symlink
	return os.Symlink(target, link)
}

// copyFile copies a file from src to dst (used as Windows fallback)
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0755)
}

// HomeDir returns the user's home directory with cross-platform compatibility
func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to environment variables
		if IsWindows() {
			if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
				return userProfile
			}
			// Last resort for Windows
			return filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"))
		}
		// Fallback for Unix systems
		return os.Getenv("HOME")
	}
	return home
}

// ConfigDir returns the platform-appropriate configuration directory
func ConfigDir() string {
	if IsWindows() {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return appData
		}
		return filepath.Join(HomeDir(), "AppData", "Roaming")
	}

	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return xdgConfig
	}
	return filepath.Join(HomeDir(), ".config")
}

// CacheDir returns the platform-appropriate cache directory
func CacheDir() string {
	if IsWindows() {
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return localAppData
		}
		return filepath.Join(HomeDir(), "AppData", "Local")
	}

	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		return xdgCache
	}
	return filepath.Join(HomeDir(), ".cache")
}

// DataDir returns the platform-appropriate data directory
func DataDir() string {
	if IsWindows() {
		return ConfigDir() // On Windows, config and data are same
	}

	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		return xdgData
	}
	return filepath.Join(HomeDir(), ".local", "share")
}

// IsExecutable checks if a file is executable on the current platform
func IsExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if IsWindows() {
		// On Windows, check file extension
		ext := filepath.Ext(path)
		return ext == ".exe" || ext == ".bat" || ext == ".cmd"
	}

	// On Unix, check execute permission
	return info.Mode()&0111 != 0
}

// MakeExecutable makes a file executable on the current platform
func MakeExecutable(path string) error {
	if IsWindows() {
		// On Windows, no explicit permission changes needed
		return nil
	}

	// On Unix, set execute permission
	return os.Chmod(path, 0755)
}
