// ABOUTME: Integration between PVI and PVX for automatic module installation
// ABOUTME: Provides functionality for PVX to install required modules automatically

package pvi

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"tamarou.com/pvm/internal/cli/progress"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/modules"
	"tamarou.com/pvm/internal/perl"
)

// pvxProgressTracker implements progress.Tracker for PVX integration
type pvxProgressTracker struct {
	writer  io.Writer
	verbose bool
	status  *progress.Status
	running bool
}

func (t *pvxProgressTracker) Start(operation string, total int) {
	t.status = &progress.Status{
		Operation: operation,
		Total:     total,
		Current:   0,
		Message:   "Starting",
	}
	t.running = true
	if t.verbose && t.writer != nil {
		fmt.Fprintf(t.writer, "Starting %s for %d items\n", operation, total)
	}
}

func (t *pvxProgressTracker) Update(current int, message string) {
	if t.status != nil {
		t.status.Current = current
		t.status.Message = message
	}
	if t.verbose && t.writer != nil {
		fmt.Fprintf(t.writer, "[%d] %s\n", current, message)
	}
}

func (t *pvxProgressTracker) Finish(result *progress.Result) {
	t.running = false
	if t.verbose && t.writer != nil && result != nil {
		if result.Success {
			fmt.Fprintf(t.writer, "Completed %s successfully\n", result.Operation)
		} else {
			fmt.Fprintf(t.writer, "Failed %s: %s\n", result.Operation, result.Message)
		}
	}
}

func (t *pvxProgressTracker) SetTotal(total int) {
	if t.status != nil {
		t.status.Total = total
	}
}

func (t *pvxProgressTracker) SetMessage(message string) {
	if t.status != nil {
		t.status.Message = message
	}
}

func (t *pvxProgressTracker) IsRunning() bool {
	return t.running
}

func (t *pvxProgressTracker) GetProgress() *progress.Status {
	if t.status == nil {
		return &progress.Status{
			Operation: "unknown",
			Total:     0,
			Current:   0,
			Message:   "Not started",
		}
	}
	return t.status
}

// pvxParallelProgressTracker implements modules.ParallelProgressTracker for PVX integration
type pvxParallelProgressTracker struct {
	writer  io.Writer
	verbose bool
}

func (t *pvxParallelProgressTracker) StartParallel(operations []string) {
	if t.verbose && t.writer != nil {
		fmt.Fprintf(t.writer, "Starting parallel operations: %v\n", operations)
	}
}

func (t *pvxParallelProgressTracker) UpdateOperation(id string, status modules.OperationStatus, message string) {
	if t.verbose && t.writer != nil {
		fmt.Fprintf(t.writer, "[%s] %s\n", id, message)
	}
}

func (t *pvxParallelProgressTracker) FinishParallel(results []*modules.OperationResult) {
	if t.verbose && t.writer != nil {
		fmt.Fprintf(t.writer, "Completed parallel operations: %d results\n", len(results))
	}
}

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

	// Create provider using builder pattern (extracted package)
	source := "metacpan" // Default to MetaCPAN
	if cfg.PVI != nil && cfg.PVI.MetadataSource != "" {
		source = cfg.PVI.MetadataSource
	}

	providerResult, err := NewProviderBuilder().
		WithConfig(cfg).
		WithSource(source).
		WithResolver().
		Build()
	if err != nil {
		return result, errors.NewModuleError(
			"PVI-904",
			"Failed to create CPAN provider",
			err,
		)
	}

	// Use the extracted modules system for installation

	// Create progress tracker that outputs to PVX's writer
	var tracker progress.Tracker
	if options.OutputWriter != nil {
		tracker = &pvxProgressTracker{
			writer:  options.OutputWriter,
			verbose: options.Verbose,
		}
	} else {
		tracker = progress.NewNullTracker()
	}

	// Create logger for installer
	logger := log.NewLogger(log.LevelInfo, os.Stderr, "PVX-Integration")

	// Create unified installer (extracted package)
	installer := modules.NewInstaller(
		providerResult.Provider,
		providerResult.Resolver,
		tracker,
		logger,
	)

	// Create parallel progress tracker
	var parallelTracker modules.ParallelProgressTracker
	if options.OutputWriter != nil {
		parallelTracker = &pvxParallelProgressTracker{
			writer:  options.OutputWriter,
			verbose: options.Verbose,
		}
	} else {
		parallelTracker = &pvxParallelProgressTracker{
			writer:  nil,
			verbose: false,
		}
	}

	// Create parallel coordinator for batch installation
	coordinator := modules.NewParallelCoordinator(
		installer,
		2, // Use only 2 workers for PVX to avoid overwhelming the system
		parallelTracker,
	)

	// Set up install options using extracted modules types
	installOptions := modules.InstallOptions{
		PerlPath:          perlPath,
		InstallDir:        options.InstallDir,
		VersionConstraint: "", // No version constraint for PVX
		Force:             false,
		RunTests:          !options.SkipTests,
		SkipDependencies:  false, // Install dependencies for PVX
		Verbose:           options.Verbose,
		Cleanup:           true,
		Parallel:          len(options.RequiredModules) > 1,
		Workers:           2,
		Context:           context.Background(),
	}

	// Install modules using parallel coordinator
	installResults, err := coordinator.InstallModules(context.Background(), options.RequiredModules, installOptions)
	if err != nil {
		return result, errors.NewModuleError(
			"PVI-905",
			fmt.Sprintf("Failed to install modules for PVX: %v", err),
			err,
		)
	}

	// Convert unified results to PVX results
	for _, installResult := range installResults {
		if installResult.Success {
			result.InstalledModules = append(result.InstalledModules, installResult.ModuleName)
			if options.Verbose && options.OutputWriter != nil {
				fmt.Fprintf(options.OutputWriter, "Successfully installed module %s v%s\n",
					installResult.ModuleName, installResult.Version)
			}
		} else {
			result.FailedModules = append(result.FailedModules, installResult.ModuleName)
			// Combine all errors into a single error message
			errorMsg := strings.Join(installResult.Errors, "; ")
			if errorMsg == "" {
				errorMsg = "installation failed"
			}
			result.Errors = append(result.Errors, fmt.Errorf("%s", errorMsg))
			if options.Verbose && options.OutputWriter != nil {
				fmt.Fprintf(options.OutputWriter, "Failed to install module %s: %s\n",
					installResult.ModuleName, errorMsg)
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

// Helper functions for PVX integration

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
