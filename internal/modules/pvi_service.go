// ABOUTME: PVI module service interface for dependency injection
// ABOUTME: Abstracts PVI module operations to enable comprehensive testing and decoupling

package modules

import (
	pviModules "tamarou.com/pvm/internal/pvi/modules"
)

// PVIModuleService provides an interface for PVI module operations
// This abstraction enables dependency injection and comprehensive testing
// by decoupling the unified Manager from direct PVI function calls
type PVIModuleService interface {
	// ListInstalled lists all installed Perl modules
	ListInstalled(options *pviModules.ModuleListOptions) ([]*pviModules.InstalledModule, error)

	// CheckOutdated finds installed modules that have newer versions available
	CheckOutdated(options *pviModules.CheckOutdatedOptions, checkLatest func(string) (string, error)) ([]*pviModules.OutdatedModuleInfo, error)

	// InstallModule installs a single Perl module and its dependencies
	InstallModule(options *pviModules.ModuleInstallOptions) (*pviModules.ModuleInstallResult, error)

	// InstallModulesParallel installs multiple modules in parallel
	InstallModulesParallel(options *pviModules.ParallelInstallOptions) (*pviModules.ParallelInstallResult, error)

	// RemoveModule uninstalls a Perl module
	RemoveModule(options *pviModules.RemoveModuleOptions) (*pviModules.RemoveModuleResult, error)
}

// RealPVIService provides the production implementation that delegates to actual PVI functions
type RealPVIService struct{}

// NewRealPVIService creates a new production PVI service implementation
func NewRealPVIService() *RealPVIService {
	return &RealPVIService{}
}

// ListInstalled delegates to the PVI ListInstalledModules function
func (r *RealPVIService) ListInstalled(options *pviModules.ModuleListOptions) ([]*pviModules.InstalledModule, error) {
	return pviModules.ListInstalledModules(options)
}

// CheckOutdated delegates to the PVI CheckOutdatedModules function
func (r *RealPVIService) CheckOutdated(options *pviModules.CheckOutdatedOptions, checkLatest func(string) (string, error)) ([]*pviModules.OutdatedModuleInfo, error) {
	return pviModules.CheckOutdatedModules(options, checkLatest)
}

// InstallModule delegates to the PVI InstallModule function
func (r *RealPVIService) InstallModule(options *pviModules.ModuleInstallOptions) (*pviModules.ModuleInstallResult, error) {
	return pviModules.InstallModule(options)
}

// InstallModulesParallel delegates to the PVI InstallModulesParallel function
func (r *RealPVIService) InstallModulesParallel(options *pviModules.ParallelInstallOptions) (*pviModules.ParallelInstallResult, error) {
	return pviModules.InstallModulesParallel(options)
}

// RemoveModule delegates to the PVI RemoveModule function
func (r *RealPVIService) RemoveModule(options *pviModules.RemoveModuleOptions) (*pviModules.RemoveModuleResult, error) {
	return pviModules.RemoveModule(options)
}
