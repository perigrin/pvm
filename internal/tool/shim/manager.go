// ABOUTME: Core shim management functionality for PVX global tool execution
// ABOUTME: Handles creation, updating, and removal of executable shims for installed tools

package shim

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"tamarou.com/pvm/internal/tool"
	"tamarou.com/pvm/internal/xdg"
)

// Manager handles shim operations for global tools
type Manager struct {
	shimDir  string
	platform string
}

// NewManager creates a new shim manager
func NewManager() (*Manager, error) {
	shimDir, err := getShimDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to get shim directory: %w", err)
	}

	return &Manager{
		shimDir:  shimDir,
		platform: runtime.GOOS,
	}, nil
}

// getShimDirectory returns the directory where shims should be stored
func getShimDirectory() (string, error) {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return "", fmt.Errorf("failed to get XDG directories: %w", err)
	}

	// Ensure the shim directory exists
	if err := dirs.EnsureDirs(); err != nil {
		return "", fmt.Errorf("failed to create shim directory: %w", err)
	}

	return dirs.ShimsDir, nil
}

// CreateShim creates an executable shim for the given tool
func (m *Manager) CreateShim(toolName string, info *tool.ToolInfo) error {
	if toolName == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if info == nil {
		return fmt.Errorf("tool info cannot be nil")
	}

	shimPath := m.getShimPath(toolName)
	shimContent, err := m.generateShimContent(toolName, info)
	if err != nil {
		return fmt.Errorf("failed to generate shim content: %w", err)
	}

	// Write shim file
	if err := os.WriteFile(shimPath, []byte(shimContent), 0755); err != nil {
		return fmt.Errorf("failed to write shim file: %w", err)
	}

	return nil
}

// UpdateShim updates an existing shim for the given tool
func (m *Manager) UpdateShim(toolName string, info *tool.ToolInfo) error {
	// For now, updating is the same as creating
	return m.CreateShim(toolName, info)
}

// RemoveShim removes the shim for the given tool
func (m *Manager) RemoveShim(toolName string) error {
	if toolName == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	shimPath := m.getShimPath(toolName)
	if err := os.Remove(shimPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove shim: %w", err)
	}

	return nil
}

// ListShims returns a list of all installed shims
func (m *Manager) ListShims() ([]string, error) {
	entries, err := os.ReadDir(m.shimDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read shim directory: %w", err)
	}

	var shims []string
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			// Remove platform-specific extensions
			if m.platform == "windows" && strings.HasSuffix(name, ".bat") {
				name = strings.TrimSuffix(name, ".bat")
			}
			shims = append(shims, name)
		}
	}

	return shims, nil
}

// ShimExists checks if a shim exists for the given tool
func (m *Manager) ShimExists(toolName string) bool {
	shimPath := m.getShimPath(toolName)
	_, err := os.Stat(shimPath)
	return err == nil
}

// GetShimDirectory returns the directory where shims are stored
func (m *Manager) GetShimDirectory() string {
	return m.shimDir
}

// getShimPath returns the full path to the shim for the given tool
func (m *Manager) getShimPath(toolName string) string {
	filename := toolName
	if m.platform == "windows" {
		filename += ".bat"
	}
	return filepath.Join(m.shimDir, filename)
}

// HasPathConflict checks if the tool name conflicts with existing system commands
func (m *Manager) HasPathConflict(toolName string) (bool, string, error) {
	// Check if the tool exists in PATH
	execPath, err := exec.LookPath(toolName)
	if err != nil {
		// No conflict if not found in PATH
		return false, "", nil
	}

	// Check if the found executable is one of our shims
	shimPath := m.getShimPath(toolName)
	if execPath == shimPath {
		// It's our own shim, no conflict
		return false, "", nil
	}

	return true, execPath, nil
}
