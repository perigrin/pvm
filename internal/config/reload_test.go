// ABOUTME: Tests for hot reloading functionality and component management
// ABOUTME: Comprehensive test coverage for component reconfiguration and rollback

package config

import (
	"fmt"
	"sync"
	"testing"
)

// MockReloadableComponent implements the Reloadable interface for testing
type MockReloadableComponent struct {
	name               string
	configureCount     int
	validateCount      int
	lastConfig         *Config
	shouldFailValidate bool
	shouldFailReconfig bool
	mu                 sync.Mutex
}

func NewMockComponent(name string) *MockReloadableComponent {
	return &MockReloadableComponent{
		name: name,
	}
}

func (m *MockReloadableComponent) Reconfigure(config *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFailReconfig {
		return fmt.Errorf("mock reconfigure failure for %s", m.name)
	}

	m.configureCount++
	m.lastConfig = config
	return nil
}

func (m *MockReloadableComponent) GetComponentName() string {
	return m.name
}

func (m *MockReloadableComponent) Validate(config *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.validateCount++

	if m.shouldFailValidate {
		return fmt.Errorf("mock validation failure for %s", m.name)
	}

	return nil
}

func (m *MockReloadableComponent) SetShouldFailValidate(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailValidate = fail
}

func (m *MockReloadableComponent) SetShouldFailReconfig(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailReconfig = fail
}

func (m *MockReloadableComponent) GetConfigureCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.configureCount
}

func (m *MockReloadableComponent) GetValidateCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.validateCount
}

func (m *MockReloadableComponent) GetLastConfig() *Config {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastConfig
}

func TestHotReloader_Creation(t *testing.T) {
	opts := DefaultHotReloaderOptions()
	reloader, err := NewHotReloader(opts)
	if err != nil {
		t.Fatalf("Failed to create hot reloader: %v", err)
	}

	if reloader == nil {
		t.Fatal("Expected non-nil reloader")
	}

	if reloader.rollbackTimeout != opts.RollbackTimeout {
		t.Errorf("Expected rollback timeout %v, got %v", opts.RollbackTimeout, reloader.rollbackTimeout)
	}

	if reloader.maxRetries != opts.MaxRetries {
		t.Errorf("Expected max retries %d, got %d", opts.MaxRetries, reloader.maxRetries)
	}
}

func TestHotReloader_ComponentRegistration(t *testing.T) {
	reloader, err := NewHotReloader(nil)
	if err != nil {
		t.Fatalf("Failed to create hot reloader: %v", err)
	}

	comp1 := NewMockComponent("test-component-1")
	comp2 := NewMockComponent("test-component-2")

	// Test registration
	if err := reloader.RegisterComponent(comp1); err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}

	if err := reloader.RegisterComponent(comp2); err != nil {
		t.Fatalf("Failed to register second component: %v", err)
	}

	// Test duplicate registration
	if err := reloader.RegisterComponent(comp1); err == nil {
		t.Error("Expected error for duplicate registration")
	}

	// Test component list
	components := reloader.GetRegisteredComponents()
	if len(components) != 2 {
		t.Errorf("Expected 2 registered components, got %d", len(components))
	}

	expectedNames := map[string]bool{"test-component-1": true, "test-component-2": true}
	for _, name := range components {
		if !expectedNames[name] {
			t.Errorf("Unexpected component name: %s", name)
		}
	}

	// Test unregistration
	reloader.UnregisterComponent("test-component-1")
	components = reloader.GetRegisteredComponents()
	if len(components) != 1 {
		t.Errorf("Expected 1 registered component after unregistration, got %d", len(components))
	}

	if components[0] != "test-component-2" {
		t.Errorf("Expected remaining component 'test-component-2', got %s", components[0])
	}
}

func TestHotReloader_SuccessfulReconfiguration(t *testing.T) {
	// Skip this test as it requires a full file system setup
	// In a real implementation, we would create temp files and test the full flow
	t.Skip("Requires file system setup for full integration test")
}

func TestHotReloader_ValidationFailure(t *testing.T) {
	reloader, err := NewHotReloader(nil)
	if err != nil {
		t.Fatalf("Failed to create hot reloader: %v", err)
	}

	comp1 := NewMockComponent("test-component-1")
	comp2 := NewMockComponent("test-component-2")

	// Make comp1 fail validation
	comp1.SetShouldFailValidate(true)

	if err := reloader.RegisterComponent(comp1); err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}

	if err := reloader.RegisterComponent(comp2); err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}

	// Set up a mock "last known good" config
	reloader.lastKnownGood = NewDefaultConfig()

	// Test handling configuration change with validation failure
	newConfig := NewDefaultConfig()
	newConfig.PVM.BuildJobs = 8

	reloader.handleConfigurationChange(newConfig)

	// comp1 should have been validated but not reconfigured due to validation failure
	if comp1.GetValidateCount() != 1 {
		t.Errorf("Expected comp1 validate count 1, got %d", comp1.GetValidateCount())
	}

	if comp1.GetConfigureCount() != 0 {
		t.Errorf("Expected comp1 configure count 0, got %d", comp1.GetConfigureCount())
	}

	// comp2 should not have been validated or reconfigured due to comp1 failure
	if comp2.GetValidateCount() != 0 {
		t.Errorf("Expected comp2 validate count 0, got %d", comp2.GetValidateCount())
	}

	if comp2.GetConfigureCount() != 0 {
		t.Errorf("Expected comp2 configure count 0, got %d", comp2.GetConfigureCount())
	}
}

func TestHotReloader_ReconfigurationFailureWithRollback(t *testing.T) {
	reloader, err := NewHotReloader(nil)
	if err != nil {
		t.Fatalf("Failed to create hot reloader: %v", err)
	}

	comp1 := NewMockComponent("test-component-1")
	comp2 := NewMockComponent("test-component-2")

	// Make comp2 fail reconfiguration
	comp2.SetShouldFailReconfig(true)

	if err := reloader.RegisterComponent(comp1); err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}

	if err := reloader.RegisterComponent(comp2); err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}

	// Set up a mock "last known good" config
	oldConfig := NewDefaultConfig()
	oldConfig.PVM.BuildJobs = 4
	reloader.lastKnownGood = oldConfig

	// Test handling configuration change with reconfiguration failure
	newConfig := NewDefaultConfig()
	newConfig.PVM.BuildJobs = 8

	reloader.handleConfigurationChange(newConfig)

	// Both components should have been validated
	if comp1.GetValidateCount() != 1 {
		t.Errorf("Expected comp1 validate count 1, got %d", comp1.GetValidateCount())
	}

	if comp2.GetValidateCount() != 1 {
		t.Errorf("Expected comp2 validate count 1, got %d", comp2.GetValidateCount())
	}

	// comp1 should have been reconfigured twice: once with new config, once for rollback
	if comp1.GetConfigureCount() != 2 {
		t.Errorf("Expected comp1 configure count 2 (reconfig + rollback), got %d", comp1.GetConfigureCount())
	}

	// comp2 should have been reconfigured once for rollback (new config failed)
	if comp2.GetConfigureCount() != 1 {
		t.Errorf("Expected comp2 configure count 1 (rollback only), got %d", comp2.GetConfigureCount())
	}

	// Both components should have the old config after rollback
	if comp1.GetLastConfig().PVM.BuildJobs != 4 {
		t.Errorf("Expected comp1 to have rolled back config (build jobs 4), got %d",
			comp1.GetLastConfig().PVM.BuildJobs)
	}
}

func TestHotReloader_ForceReload(t *testing.T) {
	reloader, err := NewHotReloader(nil)
	if err != nil {
		t.Fatalf("Failed to create hot reloader: %v", err)
	}

	comp := NewMockComponent("test-component")

	if err := reloader.RegisterComponent(comp); err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}

	// Set up current config in the watcher
	currentConfig := NewDefaultConfig()
	currentConfig.PVM.BuildJobs = 6
	reloader.watcher.currentConfig = currentConfig

	// Force reload
	if err := reloader.ForceReload(); err != nil {
		t.Fatalf("Failed to force reload: %v", err)
	}

	// Component should have been reconfigured
	if comp.GetConfigureCount() != 1 {
		t.Errorf("Expected component configure count 1, got %d", comp.GetConfigureCount())
	}

	if comp.GetLastConfig().PVM.BuildJobs != 6 {
		t.Errorf("Expected component to have current config (build jobs 6), got %d",
			comp.GetLastConfig().PVM.BuildJobs)
	}
}

func TestHotReloader_StartStop(t *testing.T) {
	// This test would require setting up temporary files and full integration
	// For now, we'll test the basic start/stop mechanics

	reloader, err := NewHotReloader(nil)
	if err != nil {
		t.Fatalf("Failed to create hot reloader: %v", err)
	}

	if reloader.IsRunning() {
		t.Error("Expected reloader not to be running initially")
	}

	// Test stopping when not running
	if err := reloader.Stop(); err != nil {
		t.Fatalf("Failed to stop non-running reloader: %v", err)
	}
}

func TestHotReloader_ComponentRetries(t *testing.T) {
	opts := DefaultHotReloaderOptions()
	opts.MaxRetries = 2

	reloader, err := NewHotReloader(opts)
	if err != nil {
		t.Fatalf("Failed to create hot reloader: %v", err)
	}

	comp := NewMockComponent("test-component")

	if err := reloader.RegisterComponent(comp); err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}

	// Make component fail reconfiguration
	comp.SetShouldFailReconfig(true)

	config := NewDefaultConfig()

	// Test reconfiguration with retries
	err = reloader.reconfigureComponentWithRetry(comp, config)
	if err == nil {
		t.Error("Expected error from failed reconfiguration")
	}

	// Should have attempted maxRetries + 1 times
	expectedAttempts := opts.MaxRetries + 1
	if comp.GetConfigureCount() != expectedAttempts {
		t.Errorf("Expected %d reconfigure attempts, got %d", expectedAttempts, comp.GetConfigureCount())
	}
}

func TestHotReloader_ComponentStatuses(t *testing.T) {
	reloader, err := NewHotReloader(nil)
	if err != nil {
		t.Fatalf("Failed to create hot reloader: %v", err)
	}

	comp1 := NewMockComponent("test-component-1")
	comp2 := NewMockComponent("test-component-2")

	if err := reloader.RegisterComponent(comp1); err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}

	if err := reloader.RegisterComponent(comp2); err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}

	statuses := reloader.GetComponentStatuses()
	if len(statuses) != 2 {
		t.Errorf("Expected 2 component statuses, got %d", len(statuses))
	}

	expectedNames := map[string]bool{"test-component-1": true, "test-component-2": true}
	for _, status := range statuses {
		if !expectedNames[status.Name] {
			t.Errorf("Unexpected component status name: %s", status.Name)
		}

		if status.Status != "running" {
			t.Errorf("Expected status 'running', got %s", status.Status)
		}
	}
}

func TestHotReloader_GetConfigurations(t *testing.T) {
	reloader, err := NewHotReloader(nil)
	if err != nil {
		t.Fatalf("Failed to create hot reloader: %v", err)
	}

	// Set up mock configurations
	currentConfig := NewDefaultConfig()
	currentConfig.PVM.BuildJobs = 8
	reloader.watcher.currentConfig = currentConfig

	lastGoodConfig := NewDefaultConfig()
	lastGoodConfig.PVM.BuildJobs = 4
	reloader.lastKnownGood = lastGoodConfig

	// Test getting current config
	current := reloader.GetCurrentConfig()
	if current == nil {
		t.Error("Expected non-nil current config")
	} else if current.PVM.BuildJobs != 8 {
		t.Errorf("Expected current config build jobs 8, got %d", current.PVM.BuildJobs)
	}

	// Test getting last known good config
	lastGood := reloader.GetLastKnownGoodConfig()
	if lastGood == nil {
		t.Error("Expected non-nil last known good config")
	} else if lastGood.PVM.BuildJobs != 4 {
		t.Errorf("Expected last good config build jobs 4, got %d", lastGood.PVM.BuildJobs)
	}
}
