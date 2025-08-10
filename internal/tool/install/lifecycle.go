// ABOUTME: Install, update, remove operations for global tool lifecycle
// ABOUTME: Provides high-level lifecycle management functions for tools

package install

import (
	"context"
	"fmt"
	"time"

	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/diskspace"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/pm/deps"
)

const (
	// Lifecycle error codes
	ErrLifecycleUpdate = "TOOL-LIFECYCLE-001"
	ErrLifecycleRemove = "TOOL-LIFECYCLE-002"
	ErrLifecycleList   = "TOOL-LIFECYCLE-003"
)

// UpdateOptions contains options for updating a tool
type UpdateOptions struct {
	// Tool name to update
	ToolName string

	// Target version (empty for latest)
	TargetVersion string

	// Force update even if same version
	Force bool

	// Run tests during update
	RunTests bool

	// Verbose output
	Verbose bool

	// Context for cancellation
	Context context.Context

	// Progress callback
	ProgressCallback func(stage string, progress float64, message string)
}

// UpdateResult contains information about the update operation
type UpdateResult struct {
	// Tool name
	ToolName string

	// Previous version
	PreviousVersion string

	// New version
	NewVersion string

	// Whether update was successful
	Success bool

	// Whether version actually changed
	VersionChanged bool

	// Update duration
	Duration time.Duration

	// Warning messages
	Warnings []string

	// Error messages
	Errors []string
}

// RemoveOptions contains options for removing a tool
type RemoveOptions struct {
	// Tool name to remove
	ToolName string

	// Remove dependencies if not used by other tools
	RemoveDependencies bool

	// Context for cancellation
	Context context.Context

	// Progress callback
	ProgressCallback func(stage string, progress float64, message string)
}

// RemoveResult contains information about the removal operation
type RemoveResult struct {
	// Tool name
	ToolName string

	// Whether removal was successful
	Success bool

	// Dependencies that were also removed
	RemovedDependencies []string

	// Removal duration
	Duration time.Duration

	// Warning messages
	Warnings []string

	// Error messages
	Errors []string
}

// ToolInfo contains information about an installed tool
type ToolInfo struct {
	// Tool identification
	ToolName   string
	ModuleName string
	Version    string

	// Installation information
	InstallDate  time.Time
	InstallPath  string
	Status       string
	LastVerified time.Time

	// Dependencies
	Dependencies []string

	// Size information
	DiskUsage int64
}

// LifecycleManager provides high-level tool lifecycle operations
type LifecycleManager struct {
	installer *ToolInstaller
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager() (*LifecycleManager, error) {
	installer, err := NewToolInstaller()
	if err != nil {
		return nil, err
	}

	return &LifecycleManager{
		installer: installer,
	}, nil
}

// SetCPANProvider sets the CPAN provider
func (lm *LifecycleManager) SetCPANProvider(provider cpan.Provider) {
	lm.installer.SetCPANProvider(provider)
}

// SetDependencyResolver sets the dependency resolver
func (lm *LifecycleManager) SetDependencyResolver(resolver deps.DependencyResolver) {
	lm.installer.SetDependencyResolver(resolver)
}

// InstallTool installs a new tool
func (lm *LifecycleManager) InstallTool(options *InstallOptions) (*InstallResult, error) {
	return lm.installer.InstallTool(options)
}

// UpdateTool updates an existing tool to a newer version
func (lm *LifecycleManager) UpdateTool(options *UpdateOptions) (*UpdateResult, error) {
	startTime := time.Now()

	// Initialize result
	result := &UpdateResult{
		ToolName: options.ToolName,
		Warnings: []string{},
		Errors:   []string{},
	}

	// Set default context
	if options.Context == nil {
		options.Context = context.Background()
	}

	// Report progress
	reportProgress := func(stage string, progress float64, message string) {
		if options.ProgressCallback != nil {
			options.ProgressCallback(stage, progress, message)
		}
	}

	reportProgress("check", 0.1, "Checking current installation")

	// Check if tool is installed
	if !lm.installer.IsToolInstalled(options.ToolName) {
		return result, errors.NewSystemError(ErrLifecycleUpdate,
			fmt.Sprintf("Tool %s is not installed", options.ToolName), nil)
	}

	// Get current version
	currentVersion, err := lm.installer.GetInstalledVersion(options.ToolName)
	if err != nil {
		return result, errors.NewSystemError(ErrLifecycleUpdate,
			fmt.Sprintf("Failed to get current version of %s", options.ToolName), err)
	}

	result.PreviousVersion = currentVersion

	// Check if we need to update
	if options.TargetVersion != "" && options.TargetVersion == currentVersion && !options.Force {
		result.NewVersion = currentVersion
		result.Success = true
		result.VersionChanged = false
		result.Duration = time.Since(startTime)
		return result, nil
	}

	reportProgress("update", 0.2, "Updating tool")

	// Create install options for update (force install over existing)
	installOptions := &InstallOptions{
		ToolName:          options.ToolName,
		VersionConstraint: options.TargetVersion,
		Force:             true, // Always force for updates
		RunTests:          options.RunTests,
		Verbose:           options.Verbose,
		Context:           options.Context,
		ProgressCallback: func(stage string, progress float64, message string) {
			// Map install progress to update progress (0.2 to 1.0)
			updateProgress := 0.2 + (progress * 0.8)
			reportProgress("update", updateProgress, message)
		},
	}

	// Perform the installation (which will overwrite existing)
	installResult, err := lm.installer.InstallTool(installOptions)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, errors.NewSystemError(ErrLifecycleUpdate,
			fmt.Sprintf("Failed to update tool %s", options.ToolName), err)
	}

	// Update result
	result.Success = installResult.Success
	result.NewVersion = installResult.Version
	result.VersionChanged = (result.PreviousVersion != result.NewVersion)
	result.Duration = time.Since(startTime)
	result.Warnings = append(result.Warnings, installResult.Warnings...)
	result.Errors = append(result.Errors, installResult.Errors...)

	log.Infof("Successfully updated tool %s from %s to %s",
		options.ToolName, result.PreviousVersion, result.NewVersion)

	return result, nil
}

// RemoveTool removes an installed tool
func (lm *LifecycleManager) RemoveTool(options *RemoveOptions) (*RemoveResult, error) {
	startTime := time.Now()

	// Initialize result
	result := &RemoveResult{
		ToolName:            options.ToolName,
		RemovedDependencies: []string{},
		Warnings:            []string{},
		Errors:              []string{},
	}

	// Set default context
	if options.Context == nil {
		options.Context = context.Background()
	}

	// Report progress
	reportProgress := func(stage string, progress float64, message string) {
		if options.ProgressCallback != nil {
			options.ProgressCallback(stage, progress, message)
		}
	}

	reportProgress("check", 0.1, "Checking tool installation")

	// Check if tool is installed
	if !lm.installer.IsToolInstalled(options.ToolName) {
		return result, errors.NewSystemError(ErrLifecycleRemove,
			fmt.Sprintf("Tool %s is not installed", options.ToolName), nil)
	}

	reportProgress("remove", 0.5, "Removing tool")

	// Remove the tool
	storage := lm.installer.GetToolStorage()
	if err := storage.RemoveTool(options.ToolName); err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, errors.NewSystemError(ErrLifecycleRemove,
			fmt.Sprintf("Failed to remove tool %s", options.ToolName), err)
	}

	result.Success = true
	result.Duration = time.Since(startTime)

	reportProgress("complete", 1.0, "Tool removal completed")

	log.Infof("Successfully removed tool %s", options.ToolName)

	return result, nil
}

// ListTools returns a list of all installed tools
func (lm *LifecycleManager) ListTools() ([]*ToolInfo, error) {
	storage := lm.installer.GetToolStorage()

	metadataList, err := storage.ListTools()
	if err != nil {
		return nil, errors.NewSystemError(ErrLifecycleList,
			"Failed to list installed tools", err)
	}

	var tools []*ToolInfo
	for _, metadata := range metadataList {
		toolInfo := &ToolInfo{
			ToolName:     metadata.ToolName,
			ModuleName:   metadata.ModuleName,
			Version:      metadata.Version,
			InstallDate:  metadata.InstallDate,
			InstallPath:  metadata.InstallPath,
			Status:       metadata.Status,
			LastVerified: metadata.LastVerified,
			Dependencies: metadata.Dependencies,
		}

		// Calculate disk usage (optional, might be expensive)
		if metadata.InstallPath != "" {
			if diskUsage, err := diskspace.CalculateDirectorySize(metadata.InstallPath); err == nil {
				toolInfo.DiskUsage = diskUsage
			} else {
				// Log warning but don't fail the listing operation
				log.Warnf("Failed to calculate disk usage for tool %s at %s: %v",
					metadata.ToolName, metadata.InstallPath, err)
				toolInfo.DiskUsage = 0
			}
		}

		tools = append(tools, toolInfo)
	}

	return tools, nil
}

// GetToolInfo returns detailed information about a specific tool
func (lm *LifecycleManager) GetToolInfo(toolName string) (*ToolInfo, error) {
	storage := lm.installer.GetToolStorage()

	metadata, err := storage.LoadMetadata(toolName)
	if err != nil {
		return nil, err
	}

	return &ToolInfo{
		ToolName:     metadata.ToolName,
		ModuleName:   metadata.ModuleName,
		Version:      metadata.Version,
		InstallDate:  metadata.InstallDate,
		InstallPath:  metadata.InstallPath,
		Status:       metadata.Status,
		LastVerified: metadata.LastVerified,
		Dependencies: metadata.Dependencies,
	}, nil
}

// VerifyTool checks if a tool installation is valid and complete
func (lm *LifecycleManager) VerifyTool(toolName string) error {
	storage := lm.installer.GetToolStorage()
	return storage.ValidateToolInstallation(toolName)
}

// CleanupTools removes orphaned tools and invalid installations
func (lm *LifecycleManager) CleanupTools() error {
	storage := lm.installer.GetToolStorage()
	return storage.CleanupOrphanedTools()
}
