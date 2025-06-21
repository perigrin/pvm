// ABOUTME: Unified module manager implementing ModuleManager interface
// ABOUTME: Provides module listing, searching, management operations across all PVM components

package modules

import (
	"context"
	"fmt"
	"time"

	"tamarou.com/pvm/internal/cli/progress"
	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/pvi/modules" // Import PVI modules for existing functionality
)

// Error codes for module management operations
const (
	ErrManagerListFailed     = "MOD-4501" // Failed to list modules
	ErrManagerSearchFailed   = "MOD-4502" // Failed to search modules
	ErrManagerRemoveFailed   = "MOD-4503" // Failed to remove modules
	ErrManagerUpdateFailed   = "MOD-4504" // Failed to update modules
	ErrManagerInstallFailed  = "MOD-4505" // Failed to install modules
	ErrManagerOutdatedFailed = "MOD-4506" // Failed to check outdated modules
)

// Manager provides a unified interface for module management operations
type Manager struct {
	provider cpan.Provider
	tracker  progress.Tracker
	logger   *log.Logger
}

// NewManager creates a new module manager
func NewManager(provider cpan.Provider, tracker progress.Tracker, logger *log.Logger) *Manager {
	return &Manager{
		provider: provider,
		tracker:  tracker,
		logger:   logger,
	}
}

// List returns modules matching the given filter
func (m *Manager) List(ctx context.Context, filter ModuleFilter) ([]*Module, error) {
	// Convert filter to PVI module list options
	listOptions := &modules.ModuleListOptions{
		PerlPath:    "", // Will be resolved by the listing function
		Pattern:     filter.Pattern,
		IncludeCore: filter.IncludeCore,
		Context:     ctx,
	}

	// Get installed modules using existing PVI functionality
	installedModules, err := modules.ListInstalledModules(listOptions)
	if err != nil {
		return nil, errors.NewSystemError(
			ErrManagerListFailed,
			fmt.Sprintf("Failed to list installed modules: %v", err),
			err)
	}

	// Convert PVI module format to unified Module format
	var result []*Module
	for _, installed := range installedModules {
		module := &Module{
			Name:             installed.Name,
			Version:          installed.Version,
			Description:      installed.Description,
			Path:             installed.Path,
			InstallationTime: installed.InstallationTime,
			CoreModule:       installed.CoreModule,
		}

		// Apply additional filtering
		if filter.LatestOnly {
			// For latest only filtering, we'd need to implement deduplication logic
			// For now, we include all since we're listing installed modules
		}

		result = append(result, module)
	}

	return result, nil
}

// SearchModules searches for available modules matching the query
func (m *Manager) SearchModules(ctx context.Context, query ModuleQuery) ([]*Module, error) {
	if m.tracker != nil {
		m.tracker.Start(fmt.Sprintf("Searching modules for '%s'", query.Query), 1)
		defer func() {
			if m.tracker.IsRunning() {
				result := &progress.Result{
					Operation: "search",
					Target:    query.Query,
					Success:   false,
				}
				m.tracker.Finish(result)
			}
		}()
	}

	// Use the provider to search modules
	startTime := time.Now()
	searchResults, err := m.provider.SearchModules(ctx, query.Query, query.Limit)
	duration := time.Since(startTime)

	if err != nil {
		return nil, errors.NewSystemError(
			ErrManagerSearchFailed,
			fmt.Sprintf("Failed to search modules for query '%s': %v", query.Query, err),
			err)
	}

	// Convert search results to unified Module format
	var result []*Module
	for _, searchResult := range searchResults.Results {
		module := &Module{
			Name:        searchResult.Name,
			Version:     searchResult.Version,
			Description: searchResult.Abstract,
			Author:      searchResult.Author,
		}
		result = append(result, module)
	}

	// Update progress tracker with success
	if m.tracker != nil && m.tracker.IsRunning() {
		progressResult := &progress.Result{
			Operation: "search",
			Target:    query.Query,
			Success:   true,
			Duration:  duration,
			Message:   fmt.Sprintf("Found %d modules matching '%s'", len(result), query.Query),
		}
		m.tracker.Finish(progressResult)
	}

	return result, nil
}

// FindOutdated finds installed modules that have newer versions available
func (m *Manager) FindOutdated(ctx context.Context, filter ModuleFilter) ([]*OutdatedModule, error) {
	if m.tracker != nil {
		m.tracker.Start("Checking for outdated modules", 1)
		defer func() {
			if m.tracker.IsRunning() {
				result := &progress.Result{
					Operation: "check_outdated",
					Target:    "modules",
					Success:   false,
				}
				m.tracker.Finish(result)
			}
		}()
	}

	// Create check options
	checkOptions := &modules.CheckOutdatedOptions{
		PerlPath:    "", // Will be resolved
		Pattern:     filter.Pattern,
		IncludeCore: filter.IncludeCore,
		Provider:    m.provider,
		Context:     ctx,
	}

	// Create version check function using our provider
	checkLatest := func(moduleName string) (string, error) {
		moduleInfo, err := m.provider.GetModuleInfo(ctx, moduleName)
		if err != nil {
			return "", err
		}
		return moduleInfo.Version, nil
	}

	// Use existing PVI functionality to check outdated modules
	startTime := time.Now()
	outdatedModules, err := modules.CheckOutdatedModules(checkOptions, checkLatest)
	duration := time.Since(startTime)

	if err != nil {
		return nil, errors.NewSystemError(
			ErrManagerOutdatedFailed,
			fmt.Sprintf("Failed to check outdated modules: %v", err),
			err)
	}

	// Convert to unified format
	var result []*OutdatedModule
	for _, outdated := range outdatedModules {
		result = append(result, &OutdatedModule{
			Name:           outdated.Name,
			CurrentVersion: outdated.InstalledVersion,
			LatestVersion:  outdated.LatestVersion,
			CoreModule:     false, // PVI OutdatedModuleInfo doesn't have CoreModule field
		})
	}

	// Update progress tracker with success
	if m.tracker != nil && m.tracker.IsRunning() {
		progressResult := &progress.Result{
			Operation: "check_outdated",
			Target:    "modules",
			Success:   true,
			Duration:  duration,
			Message:   fmt.Sprintf("Found %d outdated modules", len(result)),
		}
		m.tracker.Finish(progressResult)
	}

	return result, nil
}

// Install installs one or more modules with the given options
func (m *Manager) Install(ctx context.Context, moduleNames []string, opts InstallOptions) error {
	// Convert options to PVI install options
	installOptions := &modules.ModuleInstallOptions{
		PerlPath:          opts.PerlPath,
		InstallDir:        opts.InstallDir,
		VersionConstraint: opts.VersionConstraint,
		Force:             opts.Force,
		RunTests:          opts.RunTests,
		SkipDependencies:  opts.SkipDependencies,
		Verbose:           opts.Verbose,
		Cleanup:           opts.Cleanup,
		Provider:          m.provider,
		Context:           ctx,
	}

	// Install modules (use parallel if multiple modules)
	if len(moduleNames) > 1 && opts.Parallel {
		// Create individual module options for each module
		var moduleOptions []*modules.ModuleInstallOptions
		for _, moduleName := range moduleNames {
			moduleOpt := *installOptions // Copy base options
			moduleOpt.ModuleName = moduleName
			moduleOptions = append(moduleOptions, &moduleOpt)
		}

		parallelOptions := &modules.ParallelInstallOptions{
			Modules: moduleOptions,
			Workers: opts.Workers,
			Context: ctx,
		}

		_, err := modules.InstallModulesParallel(parallelOptions)
		if err != nil {
			return errors.NewSystemError(
				ErrManagerInstallFailed,
				fmt.Sprintf("Failed to install modules in parallel: %v", err),
				err)
		}
	} else {
		// Install modules sequentially
		for _, moduleName := range moduleNames {
			installOptions.ModuleName = moduleName
			_, err := modules.InstallModule(installOptions)
			if err != nil {
				return errors.NewSystemError(
					ErrManagerInstallFailed,
					fmt.Sprintf("Failed to install module '%s': %v", moduleName, err),
					err)
			}
		}
	}

	return nil
}

// InstallModules installs one or more modules and returns detailed results
func (m *Manager) InstallModules(ctx context.Context, moduleNames []string, opts InstallOptions) ([]*InstallResult, error) {
	var results []*InstallResult

	// Convert options to PVI install options
	installOptions := &modules.ModuleInstallOptions{
		PerlPath:          opts.PerlPath,
		InstallDir:        opts.InstallDir,
		VersionConstraint: opts.VersionConstraint,
		Force:             opts.Force,
		RunTests:          opts.RunTests,
		SkipDependencies:  opts.SkipDependencies,
		Verbose:           opts.Verbose,
		Cleanup:           opts.Cleanup,
		Provider:          m.provider,
		Context:           ctx,
	}

	// Install modules (use parallel if multiple modules)
	if len(moduleNames) > 1 && opts.Parallel {
		// Create individual module options for each module
		var moduleOptions []*modules.ModuleInstallOptions
		for _, moduleName := range moduleNames {
			moduleOpt := *installOptions // Copy base options
			moduleOpt.ModuleName = moduleName
			moduleOptions = append(moduleOptions, &moduleOpt)
		}

		parallelOptions := &modules.ParallelInstallOptions{
			Modules: moduleOptions,
			Workers: opts.Workers,
			Context: ctx,
		}

		parallelResults, err := modules.InstallModulesParallel(parallelOptions)
		if err != nil {
			return nil, errors.NewSystemError(
				ErrManagerInstallFailed,
				fmt.Sprintf("Failed to install modules in parallel: %v", err),
				err)
		}

		// Convert parallel results to unified format
		for _, result := range parallelResults.Results {
			var deps []string
			for _, dep := range result.Dependencies {
				deps = append(deps, dep.ModuleName)
			}

			results = append(results, &InstallResult{
				ModuleName:   result.ModuleName,
				Version:      result.Version,
				Success:      result.Success,
				Duration:     time.Duration(0), // Duration not available in this result
				Dependencies: deps,
				Warnings:     result.Warnings,
				Errors:       result.Errors,
				Path:         result.InstallPath,
			})
		}
	} else {
		// Install modules sequentially
		for _, moduleName := range moduleNames {
			installOptions.ModuleName = moduleName
			result, err := modules.InstallModule(installOptions)
			if err != nil {
				// Create a failed result entry
				results = append(results, &InstallResult{
					ModuleName: moduleName,
					Success:    false,
					Errors:     []string{err.Error()},
				})
				continue
			}

			// Convert result to unified format
			var deps []string
			for _, dep := range result.Dependencies {
				deps = append(deps, dep.ModuleName)
			}

			results = append(results, &InstallResult{
				ModuleName:   result.ModuleName,
				Version:      result.Version,
				Success:      result.Success,
				Duration:     time.Duration(0), // Duration not available in this result
				Dependencies: deps,
				Warnings:     result.Warnings,
				Errors:       result.Errors,
				Path:         result.InstallPath,
			})
		}
	}

	return results, nil
}

// Remove uninstalls the specified modules
func (m *Manager) Remove(ctx context.Context, moduleNames []string) error {
	for _, moduleName := range moduleNames {
		removeOptions := &modules.RemoveModuleOptions{
			ModuleName: moduleName,
			Context:    ctx,
		}

		_, err := modules.RemoveModule(removeOptions)
		if err != nil {
			return errors.NewSystemError(
				ErrManagerRemoveFailed,
				fmt.Sprintf("Failed to remove module '%s': %v", moduleName, err),
				err)
		}
	}

	return nil
}

// Update updates the specified modules to latest versions
func (m *Manager) Update(ctx context.Context, moduleNames []string) error {
	// For each module, find the latest version and install it
	for _, moduleName := range moduleNames {
		// Get latest version info
		moduleInfo, err := m.provider.GetModuleInfo(ctx, moduleName)
		if err != nil {
			return errors.NewSystemError(
				ErrManagerUpdateFailed,
				fmt.Sprintf("Failed to get latest version for module '%s': %v", moduleName, err),
				err)
		}

		// Install the latest version (this effectively updates it)
		installOptions := &modules.ModuleInstallOptions{
			ModuleName:        moduleName,
			VersionConstraint: moduleInfo.Version,
			Provider:          m.provider,
			Context:           ctx,
			Force:             true, // Force to overwrite existing version
		}

		_, err = modules.InstallModule(installOptions)
		if err != nil {
			return errors.NewSystemError(
				ErrManagerUpdateFailed,
				fmt.Sprintf("Failed to update module '%s' to version '%s': %v", moduleName, moduleInfo.Version, err),
				err)
		}
	}

	return nil
}
