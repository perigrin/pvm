// ABOUTME: Manages PATH directory integration for PVX shims
// ABOUTME: Handles adding shim directory to user's PATH and detecting conflicts

package shim

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PathManager handles PATH-related operations for shims
type PathManager struct {
	shimDir string
}

// NewPathManager creates a new PATH manager
func NewPathManager(shimDir string) *PathManager {
	return &PathManager{
		shimDir: shimDir,
	}
}

// IsInPath checks if the shim directory is in the user's PATH
func (p *PathManager) IsInPath() bool {
	pathEnv := os.Getenv("PATH")
	pathSeparator := ":"
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	}

	paths := strings.Split(pathEnv, pathSeparator)
	for _, path := range paths {
		if strings.TrimSpace(path) == p.shimDir {
			return true
		}
	}

	return false
}

// GetPathEntries returns all directories in the current PATH
func (p *PathManager) GetPathEntries() []string {
	pathEnv := os.Getenv("PATH")
	pathSeparator := ":"
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	}

	var paths []string
	for _, path := range strings.Split(pathEnv, pathSeparator) {
		if trimmed := strings.TrimSpace(path); trimmed != "" {
			paths = append(paths, trimmed)
		}
	}

	return paths
}

// FindConflicts finds executables in PATH that conflict with our shims
func (p *PathManager) FindConflicts(shimNames []string) (map[string][]string, error) {
	conflicts := make(map[string][]string)
	pathEntries := p.GetPathEntries()

	for _, shimName := range shimNames {
		var conflictPaths []string

		for _, pathEntry := range pathEntries {
			// Skip our own shim directory
			if pathEntry == p.shimDir {
				continue
			}

			// Check for executable files with the same name
			candidates := []string{shimName}
			if runtime.GOOS == "windows" {
				// On Windows, also check common executable extensions
				candidates = append(candidates, shimName+".exe", shimName+".bat", shimName+".cmd")
			}

			for _, candidate := range candidates {
				execPath := filepath.Join(pathEntry, candidate)
				if info, err := os.Stat(execPath); err == nil && !info.IsDir() {
					// Check if it's executable
					if isExecutable(execPath) {
						conflictPaths = append(conflictPaths, execPath)
					}
				}
			}
		}

		if len(conflictPaths) > 0 {
			conflicts[shimName] = conflictPaths
		}
	}

	return conflicts, nil
}

// GetPrecedence returns the precedence order of the shim directory relative to PATH
func (p *PathManager) GetPrecedence() (int, error) {
	pathEntries := p.GetPathEntries()
	for i, path := range pathEntries {
		if path == p.shimDir {
			return i, nil
		}
	}

	return -1, fmt.Errorf("shim directory not found in PATH")
}

// SuggestPathOrder suggests where the shim directory should be placed in PATH
func (p *PathManager) SuggestPathOrder() string {
	if runtime.GOOS == "windows" {
		return "Add to the beginning of PATH for highest precedence, or after system directories for safer operation"
	}
	return "Add to the beginning of PATH for highest precedence, or after /usr/local/bin for safer operation"
}

// GeneratePathString generates a PATH string with the shim directory included
func (p *PathManager) GeneratePathString(position PathPosition) string {
	pathEntries := p.GetPathEntries()
	pathSeparator := ":"
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	}

	// Remove shim directory if it already exists
	var filteredPaths []string
	for _, path := range pathEntries {
		if path != p.shimDir {
			filteredPaths = append(filteredPaths, path)
		}
	}

	// Add shim directory at the requested position
	switch position {
	case PathPositionFirst:
		filteredPaths = append([]string{p.shimDir}, filteredPaths...)
	case PathPositionLast:
		filteredPaths = append(filteredPaths, p.shimDir)
	case PathPositionAfterSystem:
		// Insert after common system directories
		insertIndex := 0
		systemDirs := []string{"/bin", "/usr/bin", "/usr/local/bin"}
		if runtime.GOOS == "windows" {
			systemDirs = []string{"C:\\Windows\\System32", "C:\\Windows"}
		}

		for i, path := range filteredPaths {
			for _, sysDir := range systemDirs {
				if strings.HasPrefix(path, sysDir) {
					insertIndex = i + 1
				}
			}
		}

		// Insert at the calculated position
		if insertIndex >= len(filteredPaths) {
			filteredPaths = append(filteredPaths, p.shimDir)
		} else {
			filteredPaths = append(filteredPaths[:insertIndex], append([]string{p.shimDir}, filteredPaths[insertIndex:]...)...)
		}
	}

	return strings.Join(filteredPaths, pathSeparator)
}

// PathPosition represents where to place the shim directory in PATH
type PathPosition int

const (
	PathPositionFirst PathPosition = iota
	PathPositionLast
	PathPositionAfterSystem
)

// isExecutable checks if a file is executable
func isExecutable(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	// On Windows, check file extension
	if runtime.GOOS == "windows" {
		ext := strings.ToLower(filepath.Ext(filePath))
		return ext == ".exe" || ext == ".bat" || ext == ".cmd" || ext == ".com"
	}

	// On Unix-like systems, check execute permission
	return info.Mode()&0111 != 0
}

// DetectShellType detects the current shell type
func DetectShellType() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		// Fallback for Windows
		if runtime.GOOS == "windows" {
			return "cmd"
		}
		return "bash"
	}

	// Extract shell name from path
	shellName := filepath.Base(shell)
	switch shellName {
	case "bash":
		return "bash"
	case "zsh":
		return "zsh"
	case "fish":
		return "fish"
	case "pwsh", "powershell":
		return "pwsh"
	default:
		return "bash" // Default fallback
	}
}

// GetShellConfigFile returns the configuration file for the given shell
func GetShellConfigFile(shell string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	switch shell {
	case "bash":
		// Try .bashrc first, then .bash_profile
		bashrc := filepath.Join(homeDir, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc, nil
		}
		return filepath.Join(homeDir, ".bash_profile"), nil
	case "zsh":
		return filepath.Join(homeDir, ".zshrc"), nil
	case "fish":
		configDir := filepath.Join(homeDir, ".config", "fish")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create fish config directory: %w", err)
		}
		return filepath.Join(configDir, "config.fish"), nil
	case "pwsh":
		if runtime.GOOS == "windows" {
			documentsDir := os.Getenv("USERPROFILE")
			if documentsDir == "" {
				return "", fmt.Errorf("USERPROFILE environment variable not set")
			}
			return filepath.Join(documentsDir, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"), nil
		}
		// PowerShell on Unix
		configDir := filepath.Join(homeDir, ".config", "powershell")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create PowerShell config directory: %w", err)
		}
		return filepath.Join(configDir, "Microsoft.PowerShell_profile.ps1"), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}
