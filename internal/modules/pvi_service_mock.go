// ABOUTME: Mock PVI service implementation for comprehensive testing
// ABOUTME: Provides configurable mock responses for all PVI module operations

package modules

import (
	pviModules "tamarou.com/pvm/internal/pvi/modules"
)

// MockPVIService provides a configurable mock implementation of PVIModuleService
// This enables comprehensive unit testing of Manager coordination logic
type MockPVIService struct {
	// Mock responses for ListInstalled
	ListInstalledResult []*pviModules.InstalledModule
	ListInstalledError  error

	// Mock responses for CheckOutdated
	CheckOutdatedResult []*pviModules.OutdatedModuleInfo
	CheckOutdatedError  error

	// Mock responses for InstallModule
	InstallModuleResult *pviModules.ModuleInstallResult
	InstallModuleError  error

	// Mock responses for InstallModulesParallel
	InstallParallelResult *pviModules.ParallelInstallResult
	InstallParallelError  error

	// Mock responses for RemoveModule
	RemoveModuleResult *pviModules.RemoveModuleResult
	RemoveModuleError  error

	// Call tracking for verification
	ListInstalledCalls   []ListInstalledCall
	CheckOutdatedCalls   []CheckOutdatedCall
	InstallModuleCalls   []InstallModuleCall
	InstallParallelCalls []InstallParallelCall
	RemoveModuleCalls    []RemoveModuleCall
}

// Call tracking structures
type ListInstalledCall struct {
	Options *pviModules.ModuleListOptions
}

type CheckOutdatedCall struct {
	Options     *pviModules.CheckOutdatedOptions
	CheckLatest func(string) (string, error)
}

type InstallModuleCall struct {
	Options *pviModules.ModuleInstallOptions
}

type InstallParallelCall struct {
	Options *pviModules.ParallelInstallOptions
}

type RemoveModuleCall struct {
	Options *pviModules.RemoveModuleOptions
}

// NewMockPVIService creates a new mock PVI service
func NewMockPVIService() *MockPVIService {
	return &MockPVIService{
		ListInstalledCalls:   make([]ListInstalledCall, 0),
		CheckOutdatedCalls:   make([]CheckOutdatedCall, 0),
		InstallModuleCalls:   make([]InstallModuleCall, 0),
		InstallParallelCalls: make([]InstallParallelCall, 0),
		RemoveModuleCalls:    make([]RemoveModuleCall, 0),
	}
}

// ListInstalled mock implementation
func (m *MockPVIService) ListInstalled(options *pviModules.ModuleListOptions) ([]*pviModules.InstalledModule, error) {
	m.ListInstalledCalls = append(m.ListInstalledCalls, ListInstalledCall{
		Options: options,
	})
	return m.ListInstalledResult, m.ListInstalledError
}

// CheckOutdated mock implementation
func (m *MockPVIService) CheckOutdated(options *pviModules.CheckOutdatedOptions, checkLatest func(string) (string, error)) ([]*pviModules.OutdatedModuleInfo, error) {
	m.CheckOutdatedCalls = append(m.CheckOutdatedCalls, CheckOutdatedCall{
		Options:     options,
		CheckLatest: checkLatest,
	})
	return m.CheckOutdatedResult, m.CheckOutdatedError
}

// InstallModule mock implementation
func (m *MockPVIService) InstallModule(options *pviModules.ModuleInstallOptions) (*pviModules.ModuleInstallResult, error) {
	m.InstallModuleCalls = append(m.InstallModuleCalls, InstallModuleCall{
		Options: options,
	})
	return m.InstallModuleResult, m.InstallModuleError
}

// InstallModulesParallel mock implementation
func (m *MockPVIService) InstallModulesParallel(options *pviModules.ParallelInstallOptions) (*pviModules.ParallelInstallResult, error) {
	m.InstallParallelCalls = append(m.InstallParallelCalls, InstallParallelCall{
		Options: options,
	})
	return m.InstallParallelResult, m.InstallParallelError
}

// RemoveModule mock implementation
func (m *MockPVIService) RemoveModule(options *pviModules.RemoveModuleOptions) (*pviModules.RemoveModuleResult, error) {
	m.RemoveModuleCalls = append(m.RemoveModuleCalls, RemoveModuleCall{
		Options: options,
	})
	return m.RemoveModuleResult, m.RemoveModuleError
}

// Test helper methods for easy mock configuration

// WithListInstalledSuccess configures a successful ListInstalled response
func (m *MockPVIService) WithListInstalledSuccess(modules []*pviModules.InstalledModule) *MockPVIService {
	m.ListInstalledResult = modules
	m.ListInstalledError = nil
	return m
}

// WithListInstalledError configures a ListInstalled error response
func (m *MockPVIService) WithListInstalledError(err error) *MockPVIService {
	m.ListInstalledResult = nil
	m.ListInstalledError = err
	return m
}

// WithCheckOutdatedSuccess configures a successful CheckOutdated response
func (m *MockPVIService) WithCheckOutdatedSuccess(outdated []*pviModules.OutdatedModuleInfo) *MockPVIService {
	m.CheckOutdatedResult = outdated
	m.CheckOutdatedError = nil
	return m
}

// WithCheckOutdatedError configures a CheckOutdated error response
func (m *MockPVIService) WithCheckOutdatedError(err error) *MockPVIService {
	m.CheckOutdatedResult = nil
	m.CheckOutdatedError = err
	return m
}

// WithInstallModuleSuccess configures a successful InstallModule response
func (m *MockPVIService) WithInstallModuleSuccess(result *pviModules.ModuleInstallResult) *MockPVIService {
	m.InstallModuleResult = result
	m.InstallModuleError = nil
	return m
}

// WithInstallModuleError configures an InstallModule error response
func (m *MockPVIService) WithInstallModuleError(err error) *MockPVIService {
	m.InstallModuleResult = nil
	m.InstallModuleError = err
	return m
}

// WithInstallParallelSuccess configures a successful InstallModulesParallel response
func (m *MockPVIService) WithInstallParallelSuccess(result *pviModules.ParallelInstallResult) *MockPVIService {
	m.InstallParallelResult = result
	m.InstallParallelError = nil
	return m
}

// WithInstallParallelError configures an InstallModulesParallel error response
func (m *MockPVIService) WithInstallParallelError(err error) *MockPVIService {
	m.InstallParallelResult = nil
	m.InstallParallelError = err
	return m
}

// WithRemoveModuleSuccess configures a successful RemoveModule response
func (m *MockPVIService) WithRemoveModuleSuccess(result *pviModules.RemoveModuleResult) *MockPVIService {
	m.RemoveModuleResult = result
	m.RemoveModuleError = nil
	return m
}

// WithRemoveModuleError configures a RemoveModule error response
func (m *MockPVIService) WithRemoveModuleError(err error) *MockPVIService {
	m.RemoveModuleResult = nil
	m.RemoveModuleError = err
	return m
}

// Verification helper methods

// GetListInstalledCallCount returns the number of ListInstalled calls made
func (m *MockPVIService) GetListInstalledCallCount() int {
	return len(m.ListInstalledCalls)
}

// GetLastListInstalledCall returns the last ListInstalled call, or nil if none
func (m *MockPVIService) GetLastListInstalledCall() *ListInstalledCall {
	if len(m.ListInstalledCalls) == 0 {
		return nil
	}
	return &m.ListInstalledCalls[len(m.ListInstalledCalls)-1]
}

// GetInstallModuleCallCount returns the number of InstallModule calls made
func (m *MockPVIService) GetInstallModuleCallCount() int {
	return len(m.InstallModuleCalls)
}

// GetLastInstallModuleCall returns the last InstallModule call, or nil if none
func (m *MockPVIService) GetLastInstallModuleCall() *InstallModuleCall {
	if len(m.InstallModuleCalls) == 0 {
		return nil
	}
	return &m.InstallModuleCalls[len(m.InstallModuleCalls)-1]
}

// GetRemoveModuleCallCount returns the number of RemoveModule calls made
func (m *MockPVIService) GetRemoveModuleCallCount() int {
	return len(m.RemoveModuleCalls)
}

// GetLastRemoveModuleCall returns the last RemoveModule call, or nil if none
func (m *MockPVIService) GetLastRemoveModuleCall() *RemoveModuleCall {
	if len(m.RemoveModuleCalls) == 0 {
		return nil
	}
	return &m.RemoveModuleCalls[len(m.RemoveModuleCalls)-1]
}
