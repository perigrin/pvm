// ABOUTME: Configuration resolution helpers for the PVM Ecosystem
// ABOUTME: Provides utility functions for resolving configuration values from multiple sources

package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"tamarou.com/pvm/internal/project"
)

// InstallOptionsBase contains common install option fields
type InstallOptionsBase struct {
	ModuleName        string
	VersionConstraint string
	PerlPath          string
	InstallDir        string
	RunTests          bool
	Force             bool
	Cleanup           bool
	Verbose           bool
	SkipDependencies  bool
	Context           context.Context
}

// PerlPathResolver is a function type for resolving Perl paths
type PerlPathResolver func() (string, error)

// ResolvePerlPath resolves the Perl interpreter path from flag value or current system
func ResolvePerlPath(flagValue string, resolver PerlPathResolver) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	if resolver == nil {
		return "", fmt.Errorf("no perl path resolver provided")
	}

	perlPath, err := resolver()
	if err != nil {
		return "", fmt.Errorf("failed to get current Perl path: %w", err)
	}

	return perlPath, nil
}

// ResolveInstallDirectory resolves the installation directory based on project context and flags
func ResolveInstallDirectory(flagValue string, moduleArgs []string) (string, error) {
	// If explicitly specified via flag, use it
	if flagValue != "" {
		return flagValue, nil
	}

	// Try to detect project context
	projectCtx, err := project.GetCurrentProject()
	if err != nil {
		// If not in a project or can't detect, return empty to use default
		return "", nil
	}

	// If in a project and no explicit directory specified, use project lib directory
	if projectCtx.IsProject {
		return projectCtx.LocalLibDir, nil
	}

	return "", nil
}

// ValidateProjectContext validates that we are in a valid project directory
func ValidateProjectContext() (*project.ProjectContext, error) {
	projectCtx, err := project.GetCurrentProject()
	if err != nil {
		return nil, fmt.Errorf("failed to detect project context: %w", err)
	}

	if !projectCtx.IsProject {
		return nil, fmt.Errorf("not in a workspace directory. Use 'pvm workspace init' to create a workspace")
	}

	return projectCtx, nil
}

// ResolveModulesFromArgs resolves module names from command arguments or cpanfile
// This function accepts a cpanfile reader to avoid import cycles
func ResolveModulesFromArgs(args []string, includeDev bool, cpanfileReader func(string, bool) ([]string, error)) ([]string, error) {
	// If modules specified as arguments, use them
	if len(args) > 0 {
		return args, nil
	}

	// No modules specified, try to read from cpanfile
	projectCtx, err := ValidateProjectContext()
	if err != nil {
		return nil, fmt.Errorf("no modules specified and %w", err)
	}

	// Check for cpanfile
	cpanfilePath := filepath.Join(projectCtx.RootDir, "cpanfile")
	if _, err := os.Stat(cpanfilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no modules specified and no cpanfile found in project directory")
	}

	// Use the provided reader function to get modules
	if cpanfileReader == nil {
		return nil, fmt.Errorf("no cpanfile reader provided")
	}

	moduleNames, err := cpanfileReader(cpanfilePath, includeDev)
	if err != nil {
		return nil, fmt.Errorf("failed to read cpanfile: %w", err)
	}

	if len(moduleNames) == 0 {
		return nil, fmt.Errorf("no modules found in cpanfile")
	}

	return moduleNames, nil
}

// ResolveStringValue resolves a string configuration value with precedence:
// 1. Flag value (if provided and not empty)
// 2. Config value (if not empty)
// 3. Default value
func ResolveStringValue(flagValue, configValue, defaultValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if configValue != "" {
		return configValue
	}
	return defaultValue
}

// ResolveBoolValue resolves a boolean configuration value with precedence:
// 1. Flag value (if explicitly set)
// 2. Config value
// 3. Default value
func ResolveBoolValue(flagSet bool, flagValue bool, configValue bool, defaultValue bool) bool {
	if flagSet {
		return flagValue
	}
	// For booleans, we can't distinguish between explicit false and unset,
	// so we use configValue if available, otherwise defaultValue
	return configValue
}

// ResolveStringSlice resolves a string slice configuration value with precedence:
// 1. Flag value (if provided and not empty)
// 2. Config value (if not empty)
// 3. Default value
func ResolveStringSlice(flagValue, configValue, defaultValue []string) []string {
	if len(flagValue) > 0 {
		return flagValue
	}
	if len(configValue) > 0 {
		return configValue
	}
	return defaultValue
}

// ValidateConfiguration performs basic validation on the effective configuration
func ValidateConfiguration(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Use the existing Validate method from types.go
	validationErrors := cfg.Validate()
	if len(validationErrors) > 0 {
		// Return the first validation error
		return validationErrors[0]
	}

	return nil
}

// GetEffectiveConfiguration loads and validates the effective configuration
func GetEffectiveConfiguration() (*Config, error) {
	cfg, err := LoadEffectiveConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	err = ValidateConfiguration(cfg)
	if err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}
