// ABOUTME: Hot reloading functionality for configuration changes
// ABOUTME: Provides validation and rollback mechanisms for safe configuration updates

package config

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Reloadable interface defines components that can be reconfigured at runtime
type Reloadable interface {
	// Reconfigure applies new configuration to the component
	// Returns an error if the new configuration cannot be applied
	Reconfigure(config *Config) error

	// GetComponentName returns the name of the component for logging/error reporting
	GetComponentName() string

	// Validate validates that the new configuration is compatible with the component
	// This is called before Reconfigure to check if the change can be applied
	Validate(config *Config) error
}

// HotReloader manages hot reloading of configuration for registered components
type HotReloader struct {
	mu         sync.RWMutex
	components map[string]Reloadable
	watcher    *ConfigWatcher
	eventSub   chan WatcherEvent
	stopChan   chan struct{}
	running    bool

	// Configuration backup for rollback
	lastKnownGood *Config

	// Rollback settings
	rollbackTimeout time.Duration
	maxRetries      int
}

// HotReloaderOptions configures the HotReloader behavior
type HotReloaderOptions struct {
	// RollbackTimeout specifies how long to wait before rolling back a failed config
	RollbackTimeout time.Duration

	// MaxRetries specifies the maximum number of retry attempts for failed reconfigurations
	MaxRetries int

	// WatcherOptions configures the underlying file watcher
	WatcherOptions *WatcherOptions
}

// DefaultHotReloaderOptions returns default options for the hot reloader
func DefaultHotReloaderOptions() *HotReloaderOptions {
	return &HotReloaderOptions{
		RollbackTimeout: 30 * time.Second,
		MaxRetries:      3,
		WatcherOptions:  DefaultWatcherOptions(),
	}
}

// NewHotReloader creates a new hot reloader for configuration changes
func NewHotReloader(opts *HotReloaderOptions) (*HotReloader, error) {
	if opts == nil {
		opts = DefaultHotReloaderOptions()
	}

	watcher, err := NewConfigWatcher(opts.WatcherOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create config watcher: %w", err)
	}

	return &HotReloader{
		components:      make(map[string]Reloadable),
		watcher:         watcher,
		eventSub:        make(chan WatcherEvent, 100),
		stopChan:        make(chan struct{}),
		rollbackTimeout: opts.RollbackTimeout,
		maxRetries:      opts.MaxRetries,
	}, nil
}

// RegisterComponent registers a component for hot reloading
func (hr *HotReloader) RegisterComponent(component Reloadable) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	name := component.GetComponentName()
	if _, exists := hr.components[name]; exists {
		return fmt.Errorf("component %s is already registered", name)
	}

	hr.components[name] = component
	return nil
}

// UnregisterComponent removes a component from hot reloading
func (hr *HotReloader) UnregisterComponent(componentName string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	delete(hr.components, componentName)
}

// Start begins the hot reloading service
func (hr *HotReloader) Start(ctx context.Context) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if hr.running {
		return fmt.Errorf("hot reloader is already running")
	}

	// Start the configuration watcher
	if err := hr.watcher.Start(ctx); err != nil {
		return fmt.Errorf("failed to start config watcher: %w", err)
	}

	// Store initial configuration as last known good
	hr.lastKnownGood = hr.watcher.GetCurrentConfig()

	hr.running = true

	// Start the event processing loop
	go hr.eventLoop(ctx)

	return nil
}

// Stop stops the hot reloading service
func (hr *HotReloader) Stop() error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if !hr.running {
		return nil
	}

	hr.running = false
	close(hr.stopChan)

	return hr.watcher.Stop()
}

// eventLoop processes configuration change events
func (hr *HotReloader) eventLoop(ctx context.Context) {
	events := hr.watcher.Events()

	for {
		select {
		case <-ctx.Done():
			return
		case <-hr.stopChan:
			return
		case event := <-events:
			hr.handleEvent(event)
		}
	}
}

// handleEvent processes a single configuration change event
func (hr *HotReloader) handleEvent(event WatcherEvent) {
	switch event.Type {
	case EventConfigChanged:
		hr.handleConfigurationChange(event.Config)
	case EventConfigError:
		// Log error but don't attempt rollback for watcher errors
		fmt.Printf("Configuration watcher error: %v\n", event.Error)
	case EventConfigValidation:
		// Configuration validation failed, rollback if needed
		fmt.Printf("Configuration validation failed: %v\n", event.Error)
		hr.rollbackConfiguration()
	case EventWatcherStarted:
		fmt.Println("Configuration watcher started")
	case EventWatcherStopped:
		fmt.Println("Configuration watcher stopped")
	}
}

// handleConfigurationChange processes a successful configuration change
func (hr *HotReloader) handleConfigurationChange(newConfig *Config) {
	hr.mu.Lock()
	components := make(map[string]Reloadable)
	for name, comp := range hr.components {
		components[name] = comp
	}
	hr.mu.Unlock()

	// Validate new configuration with all components - stop at first failure
	for name, component := range components {
		if err := component.Validate(newConfig); err != nil {
			// Validation failed, log error and stop immediately
			fmt.Printf("Validation error: component %s validation failed: %v\n", name, err)
			return
		}
	}

	// Apply configuration to all components
	var reconfigErrors []error
	successfulComponents := make([]string, 0)

	for name, component := range components {
		if err := hr.reconfigureComponentWithRetry(component, newConfig); err != nil {
			reconfigErrors = append(reconfigErrors,
				fmt.Errorf("component %s reconfiguration failed: %w", name, err))
		} else {
			successfulComponents = append(successfulComponents, name)
		}
	}

	if len(reconfigErrors) > 0 {
		// Some components failed, rollback all components for consistency
		for _, err := range reconfigErrors {
			fmt.Printf("Reconfiguration error: %v\n", err)
		}

		hr.rollbackConfiguration()
		return
	}

	// All components successfully reconfigured
	hr.mu.Lock()
	hr.lastKnownGood = newConfig
	hr.mu.Unlock()

	fmt.Printf("Configuration successfully reloaded for %d components\n", len(components))
}

// reconfigureComponentWithRetry attempts to reconfigure a component with retries
func (hr *HotReloader) reconfigureComponentWithRetry(component Reloadable, config *Config) error {
	var lastError error

	for attempt := 0; attempt <= hr.maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			time.Sleep(time.Duration(attempt) * time.Second)
			fmt.Printf("Retrying reconfiguration for component %s (attempt %d/%d)\n",
				component.GetComponentName(), attempt+1, hr.maxRetries+1)
		}

		if err := component.Reconfigure(config); err != nil {
			lastError = err
			continue
		}

		// Success
		return nil
	}

	return fmt.Errorf("failed after %d attempts: %w", hr.maxRetries+1, lastError)
}

// rollbackSuccessfulComponents rolls back components that were successfully reconfigured
func (hr *HotReloader) rollbackSuccessfulComponents(successfulComponents []string,
	allComponents map[string]Reloadable) {

	fmt.Printf("Rolling back %d successfully reconfigured components\n", len(successfulComponents))

	hr.mu.RLock()
	lastGood := hr.lastKnownGood
	hr.mu.RUnlock()

	if lastGood == nil {
		fmt.Println("No last known good configuration for rollback")
		return
	}

	for _, name := range successfulComponents {
		if component, exists := allComponents[name]; exists {
			if err := component.Reconfigure(lastGood); err != nil {
				fmt.Printf("Rollback failed for component %s: %v\n", name, err)
			} else {
				fmt.Printf("Successfully rolled back component %s\n", name)
			}
		}
	}
}

// rollbackConfiguration rolls back all components to the last known good configuration
func (hr *HotReloader) rollbackConfiguration() {
	hr.mu.Lock()
	lastGood := hr.lastKnownGood
	components := make(map[string]Reloadable)
	for name, comp := range hr.components {
		components[name] = comp
	}
	hr.mu.Unlock()

	if lastGood == nil {
		fmt.Println("No last known good configuration for rollback")
		return
	}

	fmt.Printf("Rolling back configuration for %d components\n", len(components))

	for name, component := range components {
		if err := component.Reconfigure(lastGood); err != nil {
			fmt.Printf("Rollback failed for component %s: %v\n", name, err)
		} else {
			fmt.Printf("Successfully rolled back component %s\n", name)
		}
	}
}

// ForceReload forces a reload of the current configuration to all components
func (hr *HotReloader) ForceReload() error {
	currentConfig := hr.watcher.GetCurrentConfig()
	if currentConfig == nil {
		return fmt.Errorf("no current configuration available")
	}

	hr.handleConfigurationChange(currentConfig)
	return nil
}

// GetCurrentConfig returns the current configuration
func (hr *HotReloader) GetCurrentConfig() *Config {
	return hr.watcher.GetCurrentConfig()
}

// GetLastKnownGoodConfig returns the last known good configuration
func (hr *HotReloader) GetLastKnownGoodConfig() *Config {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return hr.lastKnownGood
}

// IsRunning returns whether the hot reloader is currently running
func (hr *HotReloader) IsRunning() bool {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return hr.running
}

// GetRegisteredComponents returns the names of all registered components
func (hr *HotReloader) GetRegisteredComponents() []string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	names := make([]string, 0, len(hr.components))
	for name := range hr.components {
		names = append(names, name)
	}
	return names
}

// ComponentStatus represents the status of a component
type ComponentStatus struct {
	Name         string    `json:"name"`
	LastReconfig time.Time `json:"last_reconfig"`
	Status       string    `json:"status"`
	Error        string    `json:"error,omitempty"`
}

// GetComponentStatuses returns the status of all registered components
func (hr *HotReloader) GetComponentStatuses() []ComponentStatus {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	statuses := make([]ComponentStatus, 0, len(hr.components))
	for name := range hr.components {
		// For now, just return basic status
		// In a full implementation, components might track their own status
		statuses = append(statuses, ComponentStatus{
			Name:         name,
			LastReconfig: time.Now(), // Placeholder
			Status:       "running",
		})
	}
	return statuses
}
