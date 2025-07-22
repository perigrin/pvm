// ABOUTME: Unified module installer implementing ModuleInstaller interface
// ABOUTME: Provides standardized module installation across all PVM components

package modules

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
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
			"Not in a workspace directory. Use 'pvm workspace init' to create a workspace",
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
	// Use default perl if no path specified
	if perlPath == "" {
		perlPath = "perl"
	}

	// Method 1: Try Perl command execution (most reliable)
	if version, err := i.getVersionFromPerl(module, perlPath); err == nil && version != "" {
		return version, nil
	}

	// Method 2: Try file parsing as fallback
	if version, err := i.getVersionFromFile(module, perlPath); err == nil && version != "" {
		return version, nil
	}

	// Return "undef" for unknown versions - this is not an error condition
	// Many Perl modules legitimately have no version or return undef
	return "undef", nil
}

// validateModuleName checks if a module name is safe to use
func (i *Installer) validateModuleName(module string) error {
	if module == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	// Valid Perl module names: word chars, double colons, length limit
	modulePattern := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(?:::[A-Za-z_][A-Za-z0-9_]*)*$`)
	if !modulePattern.MatchString(module) {
		return fmt.Errorf("invalid module name format: %s", module)
	}

	if len(module) > 255 {
		return fmt.Errorf("module name too long: %d characters", len(module))
	}

	return nil
}

// getVersionFromPerl executes a Perl command to get module version
func (i *Installer) getVersionFromPerl(module, perlPath string) (string, error) {
	// Validate module name to prevent injection attacks
	if err := i.validateModuleName(module); err != nil {
		return "", err
	}

	// Use safer Perl execution with -M flag instead of string interpolation
	// This approach avoids injection vulnerabilities
	script := `
		use strict;
		use warnings;
		my $module = $ARGV[0];
		eval "require $module";
		if ($@) {
			print "undef";
		} else {
			no strict 'refs';
			my $version = ${$module . '::VERSION'};
			if (defined $version) {
				print "$version";
			} else {
				print "undef";
			}
		}
	`

	// Execute the Perl script with module name as argument
	cmd := exec.Command(perlPath, "-e", script, module)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to execute Perl for module %s", module)
	}

	version := strings.TrimSpace(stdout.String())

	// Clean up version string - remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Don't return "undef" as a version string, return empty instead
	if version == "undef" || version == "" {
		return "", fmt.Errorf("module %s has no version", module)
	}

	return version, nil
}

// getVersionFromFile tries to extract version from module file content
func (i *Installer) getVersionFromFile(module, perlPath string) (string, error) {
	// First, try to find the module file
	modulePath, err := i.findModuleFile(module, perlPath)
	if err != nil {
		return "", err
	}

	// Read the file content
	content, err := os.ReadFile(modulePath)
	if err != nil {
		return "", fmt.Errorf("failed to read module file %s: %w", modulePath, err)
	}

	return i.extractVersionFromContent(string(content))
}

// findModuleFile locates the file path for a given module
func (i *Installer) findModuleFile(module, perlPath string) (string, error) {
	// Validate module name to prevent injection attacks
	if err := i.validateModuleName(module); err != nil {
		return "", err
	}

	// Use Perl to find the module file location with safe argument passing
	script := `
		use strict;
		use warnings;
		my $module = $ARGV[0];
		$module =~ s/::/\//g;
		$module .= ".pm";

		foreach my $inc (@INC) {
			my $path = "$inc/$module";
			if (-f $path) {
				print $path;
				last;
			}
		}
	`

	cmd := exec.Command(perlPath, "-e", script, module)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to find module file for %s", module)
	}

	path := strings.TrimSpace(stdout.String())
	if path == "" {
		return "", fmt.Errorf("module file not found for %s", module)
	}

	return path, nil
}

// extractVersionFromContent extracts version using regex patterns
func (i *Installer) extractVersionFromContent(content string) (string, error) {
	// Pattern 1: Standard $VERSION = "value" or $VERSION = value
	versionRe := regexp.MustCompile(`(?m)^\s*(?:our\s+)?\$VERSION\s*=\s*['"]([^'"]+)['"]`)
	if matches := versionRe.FindStringSubmatch(content); len(matches) > 1 {
		return matches[1], nil
	}

	// Pattern 2: $VERSION = value without quotes
	versionRe2 := regexp.MustCompile(`(?m)^\s*(?:our\s+)?\$VERSION\s*=\s*([0-9]+(?:\.[0-9]+)*)`)
	if matches := versionRe2.FindStringSubmatch(content); len(matches) > 1 {
		return matches[1], nil
	}

	// Pattern 3: version pragma - our $VERSION = version->declare("value")
	useVersionRe := regexp.MustCompile(`(?m)^\s*(?:our\s+)?\$VERSION\s*=\s*version->declare\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	if matches := useVersionRe.FindStringSubmatch(content); len(matches) > 1 {
		return matches[1], nil
	}

	// Pattern 4: version->parse("value")
	parseVersionRe := regexp.MustCompile(`(?m)^\s*(?:our\s+)?\$VERSION\s*=\s*version->parse\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	if matches := parseVersionRe.FindStringSubmatch(content); len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("no version found in module content")
}
