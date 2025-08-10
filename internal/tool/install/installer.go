// ABOUTME: Core installation logic for global tool installation
// ABOUTME: Provides tool installation using existing PVM local-lib system with isolation

package install

import (
	"context"
	"fmt"
	"time"

	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/pm/deps"
	"tamarou.com/pvm/internal/pm/modules"
	"tamarou.com/pvm/internal/tool"
)

const (
	// Installer error codes
	ErrInstallFailed  = "TOOL-INSTALL-001"
	ErrInstallExists  = "TOOL-INSTALL-002"
	ErrInstallInvalid = "TOOL-INSTALL-003"
	ErrInstallEnv     = "TOOL-INSTALL-004"
	ErrInstallResolve = "TOOL-INSTALL-005"
)

// InstallOptions contains options for installing a global tool
type InstallOptions struct {
	// Tool name to install
	ToolName string

	// Module name (if different from tool name)
	ModuleName string

	// Version constraint
	VersionConstraint string

	// Perl executable path
	PerlPath string

	// Force installation even if tool exists
	Force bool

	// Run tests during installation
	RunTests bool

	// Skip dependency installation
	SkipDependencies bool

	// Additional build arguments
	BuildArgs []string

	// Verbose output
	Verbose bool

	// Context for cancellation
	Context context.Context

	// Progress callback
	ProgressCallback func(stage string, progress float64, message string)
}

// InstallResult contains information about the installation result
type InstallResult struct {
	// Tool name
	ToolName string

	// Module name
	ModuleName string

	// Installed version
	Version string

	// Installation path
	InstallPath string

	// Whether installation was successful
	Success bool

	// Dependencies that were installed
	Dependencies []string

	// Installation duration
	Duration time.Duration

	// Warning messages
	Warnings []string

	// Error messages
	Errors []string
}

// ToolInstaller handles global tool installation
type ToolInstaller struct {
	storage      *ToolStorage
	toolMapping  *tool.ToolMapping
	cpanProvider cpan.Provider
	depResolver  deps.DependencyResolver
}

// NewToolInstaller creates a new tool installer
func NewToolInstaller() (*ToolInstaller, error) {
	storage, err := NewToolStorage()
	if err != nil {
		return nil, errors.NewSystemError(ErrInstallEnv,
			"Failed to initialize tool storage", err)
	}

	toolMapping := tool.NewToolMapping()

	return &ToolInstaller{
		storage:     storage,
		toolMapping: toolMapping,
	}, nil
}

// SetCPANProvider sets the CPAN provider for metadata and downloads
func (ti *ToolInstaller) SetCPANProvider(provider cpan.Provider) {
	ti.cpanProvider = provider
}

// SetDependencyResolver sets the dependency resolver
func (ti *ToolInstaller) SetDependencyResolver(resolver deps.DependencyResolver) {
	ti.depResolver = resolver
}

// InstallTool installs a global tool with isolation
func (ti *ToolInstaller) InstallTool(options *InstallOptions) (*InstallResult, error) {
	startTime := time.Now()

	// Validate options
	if err := ti.validateInstallOptions(options); err != nil {
		return nil, err
	}

	// Initialize result
	result := &InstallResult{
		ToolName:     options.ToolName,
		Dependencies: []string{},
		Warnings:     []string{},
		Errors:       []string{},
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

	// Check if tool already exists
	if ti.storage.ToolExists(options.ToolName) && !options.Force {
		return result, errors.NewSystemError(ErrInstallExists,
			fmt.Sprintf("Tool %s is already installed (use --force to reinstall)", options.ToolName), nil)
	}

	reportProgress("resolve", 0.1, "Resolving tool to module")

	// Resolve tool name to module name
	moduleName := options.ModuleName
	if moduleName == "" {
		resolution, err := ti.toolMapping.ResolveToolToModule(options.ToolName)
		if err != nil {
			return result, errors.NewSystemError(ErrInstallResolve,
				fmt.Sprintf("Failed to resolve tool %s to CPAN module", options.ToolName), err)
		}
		moduleName = resolution.ModuleName
		log.Infof("Resolved tool %s to module %s", options.ToolName, moduleName)
	}

	result.ModuleName = moduleName

	reportProgress("setup", 0.2, "Setting up isolated environment")

	// Create tool directory
	if err := ti.storage.CreateToolDirectory(options.ToolName); err != nil {
		return result, errors.NewSystemError(ErrInstallFailed,
			fmt.Sprintf("Failed to create tool directory for %s", options.ToolName), err)
	}

	// Get installation paths
	toolPath := ti.storage.GetToolPath(options.ToolName)
	localLibPath := ti.storage.GetToolLocalLibPath(options.ToolName)
	binPath := ti.storage.GetToolBinPath(options.ToolName)

	// Get Perl path
	perlPath := options.PerlPath
	if perlPath == "" {
		var err error
		perlPath, err = perl.GetCurrentPerlPath()
		if err != nil {
			return result, errors.NewSystemError(ErrInstallEnv,
				"Failed to determine Perl path", err)
		}
	}

	reportProgress("install", 0.3, "Installing module")

	// Create module installation options
	moduleOptions := &modules.ModuleInstallOptions{
		ModuleName:         moduleName,
		VersionConstraint:  options.VersionConstraint,
		PerlPath:           perlPath,
		InstallDir:         toolPath,
		RunTests:           options.RunTests,
		Force:              options.Force,
		Cleanup:            true,
		Verbose:            options.Verbose,
		SkipDependencies:   options.SkipDependencies,
		BuildArgs:          options.BuildArgs,
		Provider:           ti.cpanProvider,
		DependencyResolver: ti.depResolver,
		Context:            options.Context,
		ForceGlobal:        true, // Always install globally for tools
		ProgressCallback: func(stage modules.InstallProgressStage, moduleName string, details string, progress float64) {
			// Map module installation stages to tool installation progress
			stageProgress := 0.3 + (progress * 0.6) // Map 0-1 to 0.3-0.9
			reportProgress("install", stageProgress, fmt.Sprintf("%s: %s", stage.String(), details))
		},
	}

	// Install the module
	moduleResult, err := modules.InstallModule(moduleOptions)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, errors.NewSystemError(ErrInstallFailed,
			fmt.Sprintf("Failed to install module %s for tool %s", moduleName, options.ToolName), err)
	}

	// Check if installation was successful
	if !moduleResult.Success {
		result.Errors = append(result.Errors, moduleResult.Errors...)
		result.Warnings = append(result.Warnings, moduleResult.Warnings...)
		return result, errors.NewSystemError(ErrInstallFailed,
			fmt.Sprintf("Module installation failed for tool %s", options.ToolName), nil)
	}

	result.Version = moduleResult.Version
	result.InstallPath = toolPath
	result.Warnings = append(result.Warnings, moduleResult.Warnings...)

	// Collect dependencies
	for _, dep := range moduleResult.Dependencies {
		if dep.Success {
			result.Dependencies = append(result.Dependencies, dep.ModuleName)
		}
	}

	reportProgress("metadata", 0.9, "Saving installation metadata")

	// Create and save metadata
	metadata := &ToolMetadata{
		ToolName:     options.ToolName,
		ModuleName:   moduleName,
		Version:      result.Version,
		InstallDate:  time.Now(),
		InstallPath:  toolPath,
		LocalLibPath: localLibPath,
		BinPath:      binPath,
		Dependencies: result.Dependencies,
		PerlVersion:  perlPath, // Store the Perl path used
		BuildArgs:    options.BuildArgs,
		Status:       "installed",
		LastVerified: time.Now(),
		CustomData:   make(map[string]string),
	}

	if err := ti.storage.SaveMetadata(metadata); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to save metadata: %v", err))
	}

	result.Success = true
	result.Duration = time.Since(startTime)

	reportProgress("complete", 1.0, "Tool installation completed")

	log.Infof("Successfully installed tool %s (module: %s, version: %s) in %v",
		options.ToolName, moduleName, result.Version, result.Duration)

	return result, nil
}

// validateInstallOptions validates the installation options
func (ti *ToolInstaller) validateInstallOptions(options *InstallOptions) error {
	if options == nil {
		return errors.NewSystemError(ErrInstallInvalid,
			"Installation options cannot be nil", nil)
	}

	if options.ToolName == "" {
		return errors.NewSystemError(ErrInstallInvalid,
			"Tool name cannot be empty", nil)
	}

	// Validate tool name format
	detector := tool.NewDetector()
	if err := detector.ValidateToolName(options.ToolName); err != nil {
		return errors.NewSystemError(ErrInstallInvalid,
			fmt.Sprintf("Invalid tool name: %v", err), err)
	}

	return nil
}

// GetToolStorage returns the storage manager for direct access
func (ti *ToolInstaller) GetToolStorage() *ToolStorage {
	return ti.storage
}

// IsToolInstalled checks if a tool is installed
func (ti *ToolInstaller) IsToolInstalled(toolName string) bool {
	return ti.storage.ToolExists(toolName)
}

// GetInstalledVersion returns the installed version of a tool
func (ti *ToolInstaller) GetInstalledVersion(toolName string) (string, error) {
	metadata, err := ti.storage.LoadMetadata(toolName)
	if err != nil {
		return "", err
	}
	return metadata.Version, nil
}
