// ABOUTME: Tool storage and isolation management for global tool installation
// ABOUTME: Manages isolated tool storage in ~/.local/share/pvm/tools/ with metadata

package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/xdg"
)

const (
	// Tool storage directory under XDG data directory
	ToolsDir = "tools"

	// Metadata file name for each tool
	MetadataFileName = "metadata.json"

	// Storage error codes
	ErrStorageInit     = "TOOL-STORAGE-001"
	ErrStorageCreate   = "TOOL-STORAGE-002"
	ErrStorageRead     = "TOOL-STORAGE-003"
	ErrStorageWrite    = "TOOL-STORAGE-004"
	ErrStorageDelete   = "TOOL-STORAGE-005"
	ErrMetadataInvalid = "TOOL-STORAGE-006"
)

// ToolMetadata contains information about an installed tool
type ToolMetadata struct {
	// Tool identification
	ToolName   string `json:"tool_name"`
	ModuleName string `json:"module_name"`
	Version    string `json:"version"`

	// Installation information
	InstallDate  time.Time `json:"install_date"`
	InstallPath  string    `json:"install_path"`
	LocalLibPath string    `json:"local_lib_path"`
	BinPath      string    `json:"bin_path"`

	// Dependencies
	Dependencies []string `json:"dependencies,omitempty"`

	// Build information
	PerlVersion string   `json:"perl_version,omitempty"`
	BuildArgs   []string `json:"build_args,omitempty"`

	// Status
	Status       string    `json:"status"`
	LastVerified time.Time `json:"last_verified,omitempty"`

	// Custom metadata
	CustomData map[string]string `json:"custom_data,omitempty"`
}

// ToolStorage manages the isolated storage for global tools
type ToolStorage struct {
	baseDir string
}

// NewToolStorage creates a new tool storage manager
func NewToolStorage() (*ToolStorage, error) {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return nil, errors.NewSystemError(ErrStorageInit,
			"Failed to get XDG directories", err)
	}

	// Create tools directory under XDG data directory
	toolsDir := filepath.Join(dirs.DataDir, ToolsDir)

	return &ToolStorage{
		baseDir: toolsDir,
	}, nil
}

// GetToolPath returns the isolated directory path for a tool
func (ts *ToolStorage) GetToolPath(toolName string) string {
	return filepath.Join(ts.baseDir, toolName)
}

// GetMetadataPath returns the metadata file path for a tool
func (ts *ToolStorage) GetMetadataPath(toolName string) string {
	return filepath.Join(ts.GetToolPath(toolName), MetadataFileName)
}

// CreateToolDirectory creates an isolated directory for a tool
func (ts *ToolStorage) CreateToolDirectory(toolName string) error {
	toolPath := ts.GetToolPath(toolName)

	if err := os.MkdirAll(toolPath, 0755); err != nil {
		return errors.NewSystemError(ErrStorageCreate,
			fmt.Sprintf("Failed to create tool directory for %s", toolName), err)
	}

	// Create subdirectories for local::lib structure
	subdirs := []string{
		filepath.Join(toolPath, "lib", "perl5"),
		filepath.Join(toolPath, "bin"),
		filepath.Join(toolPath, "man"),
		filepath.Join(toolPath, "share"),
	}

	for _, subdir := range subdirs {
		if err := os.MkdirAll(subdir, 0755); err != nil {
			return errors.NewSystemError(ErrStorageCreate,
				fmt.Sprintf("Failed to create subdirectory %s", subdir), err)
		}
	}

	return nil
}

// SaveMetadata saves tool metadata to the tool's directory
func (ts *ToolStorage) SaveMetadata(metadata *ToolMetadata) error {
	if metadata.ToolName == "" {
		return errors.NewSystemError(ErrMetadataInvalid,
			"Tool name cannot be empty", nil)
	}

	metadataPath := ts.GetMetadataPath(metadata.ToolName)

	// Ensure tool directory exists
	if err := ts.CreateToolDirectory(metadata.ToolName); err != nil {
		return err
	}

	// Marshal metadata to JSON
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return errors.NewSystemError(ErrStorageWrite,
			fmt.Sprintf("Failed to marshal metadata for %s", metadata.ToolName), err)
	}

	// Write metadata file
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return errors.NewSystemError(ErrStorageWrite,
			fmt.Sprintf("Failed to write metadata file for %s", metadata.ToolName), err)
	}

	return nil
}

// LoadMetadata loads tool metadata from the tool's directory
func (ts *ToolStorage) LoadMetadata(toolName string) (*ToolMetadata, error) {
	metadataPath := ts.GetMetadataPath(toolName)

	// Check if metadata file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return nil, errors.NewSystemError(ErrStorageRead,
			fmt.Sprintf("Tool %s not found", toolName), err)
	}

	// Read metadata file
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, errors.NewSystemError(ErrStorageRead,
			fmt.Sprintf("Failed to read metadata for %s", toolName), err)
	}

	// Unmarshal JSON
	var metadata ToolMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, errors.NewSystemError(ErrMetadataInvalid,
			fmt.Sprintf("Invalid metadata format for %s", toolName), err)
	}

	return &metadata, nil
}

// ToolExists checks if a tool is installed
func (ts *ToolStorage) ToolExists(toolName string) bool {
	metadataPath := ts.GetMetadataPath(toolName)
	_, err := os.Stat(metadataPath)
	return err == nil
}

// ListTools returns a list of all installed tools
func (ts *ToolStorage) ListTools() ([]*ToolMetadata, error) {
	// Check if tools directory exists
	if _, err := os.Stat(ts.baseDir); os.IsNotExist(err) {
		return []*ToolMetadata{}, nil
	}

	// Read tools directory
	entries, err := os.ReadDir(ts.baseDir)
	if err != nil {
		return nil, errors.NewSystemError(ErrStorageRead,
			"Failed to read tools directory", err)
	}

	var tools []*ToolMetadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		toolName := entry.Name()
		metadata, err := ts.LoadMetadata(toolName)
		if err != nil {
			// Skip tools with invalid metadata but don't fail
			continue
		}

		tools = append(tools, metadata)
	}

	return tools, nil
}

// RemoveTool removes a tool and all its files
func (ts *ToolStorage) RemoveTool(toolName string) error {
	toolPath := ts.GetToolPath(toolName)

	// Check if tool exists
	if !ts.ToolExists(toolName) {
		return errors.NewSystemError(ErrStorageRead,
			fmt.Sprintf("Tool %s not found", toolName), nil)
	}

	// Remove entire tool directory
	if err := os.RemoveAll(toolPath); err != nil {
		return errors.NewSystemError(ErrStorageDelete,
			fmt.Sprintf("Failed to remove tool %s", toolName), err)
	}

	return nil
}

// GetToolLocalLibPath returns the local::lib path for a tool
func (ts *ToolStorage) GetToolLocalLibPath(toolName string) string {
	return filepath.Join(ts.GetToolPath(toolName), "lib", "perl5")
}

// GetToolBinPath returns the bin path for a tool
func (ts *ToolStorage) GetToolBinPath(toolName string) string {
	return filepath.Join(ts.GetToolPath(toolName), "bin")
}

// CleanupOrphanedTools removes tools that are in inconsistent state
func (ts *ToolStorage) CleanupOrphanedTools() error {
	// Check if tools directory exists
	if _, err := os.Stat(ts.baseDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(ts.baseDir)
	if err != nil {
		return errors.NewSystemError(ErrStorageRead,
			"Failed to read tools directory", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		toolName := entry.Name()
		metadataPath := ts.GetMetadataPath(toolName)

		// Check if metadata file exists
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			// Remove orphaned tool directory
			toolPath := ts.GetToolPath(toolName)
			os.RemoveAll(toolPath)
		}
	}

	return nil
}

// ValidateToolInstallation checks if a tool installation is complete and valid
func (ts *ToolStorage) ValidateToolInstallation(toolName string) error {
	metadata, err := ts.LoadMetadata(toolName)
	if err != nil {
		return err
	}

	// Check required paths exist
	requiredPaths := []string{
		metadata.InstallPath,
		metadata.LocalLibPath,
		metadata.BinPath,
	}

	for _, path := range requiredPaths {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return errors.NewSystemError(ErrMetadataInvalid,
				fmt.Sprintf("Tool %s installation is incomplete: missing %s", toolName, path), err)
		}
	}

	return nil
}
