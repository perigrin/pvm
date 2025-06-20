// ABOUTME: Unified module installer implementing ModuleInstaller interface
// ABOUTME: Provides standardized module installation across all PVM components

package modules

import (
	"context"
	"fmt"
	"time"

	"tamarou.com/pvm/internal/cli/progress"
	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/project"
	"tamarou.com/pvm/internal/pvi/deps"
	"tamarou.com/pvm/internal/pvi/modules" // Import PVI modules for existing functionality
)

// Error codes for module installation
const (
	ErrInstallerSetupFailed    = "MOD-4601" // Failed to setup installation environment
	ErrInstallerInstallFailed  = "MOD-4602" // Installation failed
	ErrInstallerInvalidOptions = "MOD-4603" // Invalid installation options
	ErrInstallerPerlNotFound   = "MOD-4604" // Perl interpreter not found
	ErrInstallerProjectFailed  = "MOD-4605" // Project context detection failed
)

// Installer provides unified module installation capabilities
type Installer struct {
	provider cpan.Provider
	resolver deps.DependencyResolver
	tracker  progress.Tracker
	logger   *log.Logger
}

// NewInstaller creates a module installer
func NewInstaller(provider cpan.Provider, resolver deps.DependencyResolver, tracker progress.Tracker, logger *log.Logger) *Installer {
	return &Installer{
		provider: provider,
		resolver: resolver,
		tracker:  tracker,
		logger:   logger,
	}
}

// InstallModule installs a single module with the given options
func (i *Installer) InstallModule(ctx context.Context, module string, opts InstallOptions) (*InstallResult, error) {
	if i.tracker != nil {
		i.tracker.Start(fmt.Sprintf("Installing %s", module), 1)
		defer func() {
			if i.tracker.IsRunning() {
				result := &progress.Result{
					Operation: "install",
					Target:    module,
					Success:   false,
					Duration:  0,
				}
				i.tracker.Finish(result)
			}
		}()
	}

	// Convert to PVI module install options
	pviOptions, err := i.convertInstallOptions(module, opts)
	if err != nil {
		return nil, errors.NewSystemError(
			ErrInstallerInvalidOptions,
			fmt.Sprintf("Failed to convert install options for %s", module),
			err)
	}

	// Use existing PVI installation functionality
	startTime := time.Now()
	installResult, err := modules.InstallModule(pviOptions)
	duration := time.Since(startTime)

	result := &InstallResult{
		ModuleName: module,
		Duration:   duration,
		Success:    err == nil,
	}

	// Extract information from PVI install result if available
	if installResult != nil {
		result.Version = installResult.Version
		// Convert dependency results to strings
		for _, dep := range installResult.Dependencies {
			result.Dependencies = append(result.Dependencies, dep.ModuleName)
		}
		result.Warnings = installResult.Warnings
		result.Path = installResult.InstallPath
	}

	if err != nil {
		result.Errors = []string{err.Error()}
		return result, errors.NewSystemError(
			ErrInstallerInstallFailed,
			fmt.Sprintf("Failed to install module %s", module),
			err)
	}

	// Try to get version information
	if version, verErr := i.getInstalledVersion(module, opts.PerlPath); verErr == nil {
		result.Version = version
	}

	// Update progress tracker
	if i.tracker != nil && i.tracker.IsRunning() {
		progressResult := &progress.Result{
			Operation: "install",
			Target:    module,
			Success:   true,
			Duration:  duration,
			Message:   fmt.Sprintf("Successfully installed %s", module),
		}
		i.tracker.Finish(progressResult)
	}

	return result, nil
}

// InstallBatch installs multiple modules, potentially in parallel
func (i *Installer) InstallBatch(ctx context.Context, modules []string, opts InstallOptions) ([]*InstallResult, error) {
	if len(modules) == 0 {
		return []*InstallResult{}, nil
	}

	if i.tracker != nil {
		i.tracker.Start(fmt.Sprintf("Installing %d modules", len(modules)), len(modules))
	}

	results := make([]*InstallResult, 0, len(modules))

	for idx, module := range modules {
		if i.tracker != nil {
			i.tracker.Update(idx+1, fmt.Sprintf("Installing %s", module))
		}

		result, err := i.InstallModule(ctx, module, opts)
		if result != nil {
			results = append(results, result)
		}

		// Continue with other modules even if one fails
		if err != nil && i.logger != nil {
			i.logger.Errorf("Failed to install module %s: %v", module, err)
		}
	}

	if i.tracker != nil && i.tracker.IsRunning() {
		successful := 0
		for _, result := range results {
			if result.Success {
				successful++
			}
		}

		progressResult := &progress.Result{
			Operation: "batch_install",
			Target:    fmt.Sprintf("%d modules", len(modules)),
			Success:   successful == len(modules),
			Message:   fmt.Sprintf("Installed %d/%d modules successfully", successful, len(modules)),
		}
		i.tracker.Finish(progressResult)
	}

	return results, nil
}

// SetupInstallationEnvironment sets up the installation environment with common configuration
// Note: This function is a placeholder that should be implemented by the caller
// using the appropriate provider builder to avoid import cycles
func SetupInstallationEnvironment(provider cpan.Provider, resolver deps.DependencyResolver) *InstallationEnvironment {
	return &InstallationEnvironment{
		Provider: provider,
		Resolver: resolver,
	}
}

// RequireProjectContext ensures we're in a valid project context
func RequireProjectContext() (*project.ProjectContext, error) {
	projectCtx, err := project.GetCurrentProject()
	if err != nil {
		return nil, errors.NewSystemError(
			ErrInstallerProjectFailed,
			"Failed to detect project context",
			err)
	}

	if !projectCtx.IsProject {
		return nil, errors.NewSystemError(
			ErrInstallerProjectFailed,
			"Not in a project directory. Use 'pvm project init' to create a project",
			nil)
	}

	return projectCtx, nil
}

// ResolvePerlPath resolves the Perl interpreter path with fallback logic
func ResolvePerlPath(perlPath string) (string, error) {
	if perlPath != "" {
		return perlPath, nil
	}

	resolved, err := perl.GetCurrentPerlPath()
	if err != nil {
		return "", errors.NewSystemError(
			ErrInstallerPerlNotFound,
			"Failed to resolve Perl interpreter path",
			err)
	}

	return resolved, nil
}

// CreateInstallOptions creates standardized install options from command parameters
func CreateInstallOptions(module, version, perlPath, installDir string, skipTests, force, verbose, skipDeps bool, provider cpan.Provider, resolver deps.DependencyResolver, progressCallback InstallProgressCallback, ctx context.Context) *ModuleInstallOptions {
	return &ModuleInstallOptions{
		ModuleName:         module,
		VersionConstraint:  version,
		PerlPath:           perlPath,
		InstallDir:         installDir,
		RunTests:           !skipTests,
		Force:              force,
		Cleanup:            true,
		Verbose:            verbose,
		SkipDependencies:   skipDeps,
		Provider:           provider,
		DependencyResolver: resolver,
		ProgressCallback:   progressCallback,
		Context:            ctx,
	}
}

// InstallationEnvironment contains the setup environment for installation
type InstallationEnvironment struct {
	// Provider is the configured CPAN provider
	Provider cpan.Provider

	// Resolver is the configured dependency resolver
	Resolver deps.DependencyResolver
}

// convertInstallOptions converts unified install options to PVI-specific options
func (i *Installer) convertInstallOptions(module string, opts InstallOptions) (*modules.ModuleInstallOptions, error) {
	perlPath, err := ResolvePerlPath(opts.PerlPath)
	if err != nil {
		return nil, err
	}

	// Create progress callback if tracker is available
	var progressCallback modules.InstallProgressCallback
	if i.tracker != nil {
		tracker := i.tracker // Capture for use in callback
		progressCallback = func(stage modules.InstallProgressStage, moduleName string, details string, progress float64) {
			if tracker.IsRunning() {
				message := fmt.Sprintf("%s: %s", stage.String(), details)
				// Convert progress percentage to current/total for progress tracker
				current := int(progress * 100)
				tracker.Update(current, message)
			}
		}
	}

	return &modules.ModuleInstallOptions{
		ModuleName:         module,
		VersionConstraint:  opts.VersionConstraint,
		PerlPath:           perlPath,
		InstallDir:         opts.InstallDir,
		RunTests:           opts.RunTests,
		Force:              opts.Force,
		Cleanup:            opts.Cleanup,
		Verbose:            opts.Verbose,
		SkipDependencies:   opts.SkipDependencies,
		Provider:           i.provider,
		DependencyResolver: i.resolver,
		ProgressCallback:   progressCallback,
		Context:            opts.Context,
	}, nil
}

// getInstalledVersion tries to get the installed version of a module
func (i *Installer) getInstalledVersion(module, perlPath string) (string, error) {
	// This is a simplified version - in practice, we'd query the installed modules
	// For now, return empty version to avoid complex implementation
	return "", fmt.Errorf("version detection not implemented")
}
