// ABOUTME: Integration between PVI and PVX for automatic module installation
// ABOUTME: Provides functionality for PVX to install required modules automatically

package pvi

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/pvi/deps"
	"tamarou.com/pvm/internal/pvi/modules"
)

// PVXIntegrationOptions contains options for automatic module installation
type PVXIntegrationOptions struct {
	// PerlVersion specifies which Perl version to install modules for
	PerlVersion string

	// RequiredModules is the list of modules to install
	RequiredModules []string

	// InstallDir is the directory to install modules to (optional)
	InstallDir string

	// Verbose enables verbose output
	Verbose bool

	// MaxRetries is the maximum number of retry attempts for failed installations
	MaxRetries int

	// SkipTests whether to skip running tests during installation
	SkipTests bool

	// Output writer for status messages (optional, defaults to discarding output)
	OutputWriter io.Writer
}

// PVXIntegrationResult contains the result of automatic module installation
type PVXIntegrationResult struct {
	// InstalledModules contains successfully installed modules
	InstalledModules []string

	// FailedModules contains modules that failed to install
	FailedModules []string

	// SkippedModules contains modules that were already installed
	SkippedModules []string

	// Errors contains any errors encountered during installation
	Errors []error
}

// InstallModulesForPVX automatically installs required modules for PVX execution
func InstallModulesForPVX(options *PVXIntegrationOptions) (*PVXIntegrationResult, error) {
	if options == nil {
		return nil, errors.NewModuleError(
			"PVI-901",
			"No installation options provided",
			nil,
		)
	}

	result := &PVXIntegrationResult{
		InstalledModules: []string{},
		FailedModules:    []string{},
		SkippedModules:   []string{},
		Errors:           []error{},
	}

	if len(options.RequiredModules) == 0 {
		if options.Verbose && options.OutputWriter != nil {
			fmt.Fprintln(options.OutputWriter, "No modules required for installation")
		}
		return result, nil
	}

	if options.Verbose && options.OutputWriter != nil {
		fmt.Fprintf(options.OutputWriter, "Installing %d required modules for Perl %s\n",
			len(options.RequiredModules), options.PerlVersion)
	}

	// Get configuration
	cfg, err := config.LoadEffectiveConfig()
	if err != nil {
		return result, errors.NewModuleError(
			"PVI-902",
			"Failed to load configuration",
			err,
		)
	}

	// Resolve Perl executable
	perlPath, err := resolvePerlExecutable(options.PerlVersion)
	if err != nil {
		return result, errors.NewModuleError(
			"PVI-903",
			fmt.Sprintf("Failed to resolve Perl executable for version %s", options.PerlVersion),
			err,
		)
	}

	// Set up CPAN provider
	source := "metacpan" // Default to MetaCPAN
	if cfg.PVI != nil && cfg.PVI.MetadataSource != "" {
		source = cfg.PVI.MetadataSource
	}

	var providerOptions []cpan.ProviderOption
	if cfg.PVI != nil {
		if cfg.PVI.DefaultMirror != "" {
			providerOptions = append(providerOptions, cpan.WithBaseURL(cfg.PVI.DefaultMirror))
		}
		if cfg.PVI.CacheDir != "" && cfg.PVI.CacheTTL > 0 {
			providerOptions = append(providerOptions, cpan.WithCache(cfg.PVI.CacheDir, cfg.PVI.CacheTTL))
		}
	}

	provider, err := cpan.NewProvider(source, providerOptions...)
	if err != nil {
		return result, errors.NewModuleError(
			"PVI-904",
			"Failed to create CPAN provider",
			err,
		)
	}

	// Set up dependency resolver
	resolver := deps.NewDependencyResolver()

	// Process each required module
	for _, moduleName := range options.RequiredModules {
		if options.Verbose {
			if options.OutputWriter != nil {
				fmt.Fprintf(options.OutputWriter, "Processing module: %s\n", moduleName)
			}
		}

		// Check if module is already installed
		installed, err := isModuleInstalled(moduleName, perlPath)
		if err != nil {
			if options.Verbose {
				if options.OutputWriter != nil {
					fmt.Fprintf(options.OutputWriter, "Warning: Could not check if %s is installed: %v\n", moduleName, err)
				}
			}
			// Continue with installation attempt
		} else if installed {
			if options.Verbose {
				if options.OutputWriter != nil {
					fmt.Fprintf(options.OutputWriter, "Module %s is already installed, skipping\n", moduleName)
				}
			}
			result.SkippedModules = append(result.SkippedModules, moduleName)
			continue
		}

		// Attempt to install the module
		err = installSingleModule(moduleName, perlPath, options, provider, resolver)
		if err != nil {
			result.FailedModules = append(result.FailedModules, moduleName)
			result.Errors = append(result.Errors, err)

			if options.Verbose {
				if options.OutputWriter != nil {
					fmt.Fprintf(options.OutputWriter, "Failed to install module %s: %v\n", moduleName, err)
				}
			}
		} else {
			result.InstalledModules = append(result.InstalledModules, moduleName)

			if options.Verbose {
				if options.OutputWriter != nil {
					fmt.Fprintf(options.OutputWriter, "Successfully installed module %s\n", moduleName)
				}
			}
		}
	}

	// Summary
	if options.Verbose {
		if options.OutputWriter != nil {
			fmt.Fprintf(options.OutputWriter, "Module installation complete:\n")
			fmt.Fprintf(options.OutputWriter, "  Installed: %d modules\n", len(result.InstalledModules))
			fmt.Fprintf(options.OutputWriter, "  Skipped: %d modules\n", len(result.SkippedModules))
			fmt.Fprintf(options.OutputWriter, "  Failed: %d modules\n", len(result.FailedModules))
		}
	}

	return result, nil
}

// isModuleInstalled checks if a module is already installed
func isModuleInstalled(moduleName, perlPath string) (bool, error) {
	// Use the module manager to check installation status
	listOptions := &modules.ModuleListOptions{
		PerlPath:    perlPath,
		Pattern:     moduleName,
		IncludeCore: true,
		Context:     context.Background(),
	}

	installedModules, err := modules.ListInstalledModules(listOptions)
	if err != nil {
		return false, err
	}

	// Check if the module appears in the installed list
	for _, mod := range installedModules {
		if normalizeModuleName(mod.Name) == normalizeModuleName(moduleName) {
			return true, nil
		}
	}

	return false, nil
}

// installSingleModule installs a single module with retry logic
func installSingleModule(moduleName, perlPath string, options *PVXIntegrationOptions, provider cpan.Provider, resolver deps.DependencyResolver) error {
	maxRetries := options.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3 // Default retry count
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if options.Verbose && attempt > 1 {
			if options.OutputWriter != nil {
				fmt.Fprintf(options.OutputWriter, "Retry attempt %d/%d for module %s\n", attempt, maxRetries, moduleName)
			}
		}

		// Attempt installation
		installOptions := &modules.ModuleInstallOptions{
			ModuleName:         moduleName,
			PerlPath:           perlPath,
			RunTests:           !options.SkipTests,
			Force:              false,
			Cleanup:            true,
			Verbose:            options.Verbose,
			SkipDependencies:   false,
			Provider:           provider,
			DependencyResolver: resolver,
			Context:            context.Background(),
		}

		if options.InstallDir != "" {
			installOptions.InstallDir = options.InstallDir
		}

		// Add progress callback if verbose
		if options.Verbose {
			installOptions.ProgressCallback = func(stage modules.InstallProgressStage, mod string, details string, progress float64) {
				if details != "" {
					if options.OutputWriter != nil {
						fmt.Fprintf(options.OutputWriter, "  %s: %s\n", stage.String(), details)
					}
				}
			}
		}

		installResult, err := modules.InstallModule(installOptions)
		if err == nil && installResult.Success {
			// Success
			return nil
		}

		if err != nil {
			lastErr = err
		} else {
			lastErr = errors.NewModuleError(
				"PVI-904",
				fmt.Sprintf("Module installation failed: %v", installResult.Errors),
				nil,
			)
		}

		// If this is not the last attempt, wait before retrying
		if attempt < maxRetries {
			if options.Verbose {
				if options.OutputWriter != nil {
					fmt.Fprintf(options.OutputWriter, "Installation failed, will retry: %v\n", lastErr)
				}
			}
			time.Sleep(time.Second * 2) // Brief delay before retry
		}
	}

	// All attempts failed
	return errors.NewModuleError(
		"PVI-904",
		fmt.Sprintf("Failed to install module %s after %d attempts", moduleName, maxRetries),
		lastErr,
	)
}

// normalizeModuleName normalizes module names for comparison
func normalizeModuleName(name string) string {
	// Convert :: to - for comparison (common in some contexts)
	// and make case-insensitive
	return strings.ToLower(strings.ReplaceAll(name, "::", "-"))
}

// resolvePerlExecutable resolves the Perl executable for a given version
func resolvePerlExecutable(perlVersion string) (string, error) {
	if perlVersion == "" {
		// Use system Perl
		systemPerl, err := perl.DetectSystemPerl()
		if err != nil {
			return "", err
		}
		return systemPerl.Path, nil
	}

	// Resolve specific version
	options := &perl.ResolutionOptions{
		ExplicitVersion: perlVersion,
	}

	resolved, err := perl.ResolveVersion(options)
	if err != nil {
		return "", err
	}

	return resolved.Path, nil
}

// CheckModuleAvailability checks if modules are available for installation
func CheckModuleAvailability(modules []string, perlVersion string) (*ModuleAvailabilityResult, error) {
	result := &ModuleAvailabilityResult{
		AvailableModules:   []string{},
		UnavailableModules: []string{},
		Errors:             []error{},
	}

	if len(modules) == 0 {
		return result, nil
	}

	// Get configuration for CPAN access
	cfg, err := config.LoadEffectiveConfig()
	if err != nil {
		return result, errors.NewModuleError(
			"PVI-905",
			"Failed to load configuration for module availability check",
			err,
		)
	}

	// Set up CPAN provider
	source := "metacpan" // Default to MetaCPAN
	if cfg.PVI != nil && cfg.PVI.MetadataSource != "" {
		source = cfg.PVI.MetadataSource
	}

	var providerOptions []cpan.ProviderOption
	if cfg.PVI != nil {
		if cfg.PVI.DefaultMirror != "" {
			providerOptions = append(providerOptions, cpan.WithBaseURL(cfg.PVI.DefaultMirror))
		}
		if cfg.PVI.CacheDir != "" && cfg.PVI.CacheTTL > 0 {
			providerOptions = append(providerOptions, cpan.WithCache(cfg.PVI.CacheDir, cfg.PVI.CacheTTL))
		}
	}

	provider, err := cpan.NewProvider(source, providerOptions...)
	if err != nil {
		return result, errors.NewModuleError(
			"PVI-906",
			"Failed to create CPAN provider for availability check",
			err,
		)
	}

	// Check each module
	ctx := context.Background()
	for _, moduleName := range modules {
		available, err := checkSingleModuleAvailability(provider, moduleName, ctx)
		switch {
		case err != nil:
			result.Errors = append(result.Errors, err)
			result.UnavailableModules = append(result.UnavailableModules, moduleName)
		case available:
			result.AvailableModules = append(result.AvailableModules, moduleName)
		default:
			result.UnavailableModules = append(result.UnavailableModules, moduleName)
		}
	}

	return result, nil
}

// ModuleAvailabilityResult contains the result of module availability checking
type ModuleAvailabilityResult struct {
	// AvailableModules contains modules that are available for installation
	AvailableModules []string

	// UnavailableModules contains modules that are not available
	UnavailableModules []string

	// Errors contains any errors encountered during availability checking
	Errors []error
}

// checkSingleModuleAvailability checks if a single module is available
func checkSingleModuleAvailability(provider cpan.Provider, moduleName string, ctx context.Context) (bool, error) {
	// Use the search functionality to check if module exists
	searchResult, err := provider.SearchModules(ctx, moduleName, 1)
	if err != nil {
		return false, err
	}

	// Check if we found an exact match
	for _, mod := range searchResult.Results {
		if normalizeModuleName(mod.Name) == normalizeModuleName(moduleName) {
			return true, nil
		}
	}

	return false, nil
}

// GetRequiredModulesForScript analyzes a Perl script to determine required modules
// This is a helper function for PVX to determine what modules might be needed
func GetRequiredModulesForScript(scriptPath string) ([]string, error) {
	// This is a simplified implementation
	// In a full implementation, this would parse the Perl script to extract
	// 'use' and 'require' statements

	// For now, return an empty list
	// This functionality could be enhanced later with actual Perl parsing
	return []string{}, nil
}

// CreateModuleEnvironment creates an environment with the required modules installed
// This is used by PVX to set up isolated environments with specific modules
func CreateModuleEnvironment(perlVersion string, modules []string, targetDir string) (*ModuleEnvironmentResult, error) {
	result := &ModuleEnvironmentResult{
		EnvironmentDir:   targetDir,
		InstalledModules: []string{},
		Errors:           []error{},
	}

	if len(modules) == 0 {
		return result, nil
	}

	// Install modules to the target directory
	options := &PVXIntegrationOptions{
		PerlVersion:     perlVersion,
		RequiredModules: modules,
		InstallDir:      targetDir,
		Verbose:         false,
		SkipTests:       true, // Skip tests for isolated environments
	}

	installResult, err := InstallModulesForPVX(options)
	if err != nil {
		return result, err
	}

	result.InstalledModules = installResult.InstalledModules
	result.Errors = installResult.Errors

	return result, nil
}

// ModuleEnvironmentResult contains the result of creating a module environment
type ModuleEnvironmentResult struct {
	// EnvironmentDir is the directory where modules were installed
	EnvironmentDir string

	// InstalledModules contains successfully installed modules
	InstalledModules []string

	// Errors contains any errors encountered
	Errors []error
}
