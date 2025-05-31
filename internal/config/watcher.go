// ABOUTME: File system watcher for configuration files
// ABOUTME: Provides real-time monitoring of configuration changes with validation

package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"tamarou.com/pvm/internal/xdg"
)

// WatcherEvent represents a configuration change event
type WatcherEvent struct {
	Type      WatcherEventType `json:"type"`
	Path      string           `json:"path"`
	Config    *Config          `json:"config,omitempty"`
	Error     error            `json:"error,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
}

// WatcherEventType represents the type of configuration change
type WatcherEventType string

const (
	// EventConfigChanged indicates configuration was successfully reloaded
	EventConfigChanged WatcherEventType = "config_changed"

	// EventConfigError indicates an error occurred during reload
	EventConfigError WatcherEventType = "config_error"

	// EventConfigValidation indicates validation failed during reload
	EventConfigValidation WatcherEventType = "config_validation"

	// EventWatcherStarted indicates the watcher has started
	EventWatcherStarted WatcherEventType = "watcher_started"

	// EventWatcherStopped indicates the watcher has stopped
	EventWatcherStopped WatcherEventType = "watcher_stopped"
)

// ConfigWatcher monitors configuration files for changes
type ConfigWatcher struct {
	mu            sync.RWMutex
	watcher       *fsnotify.Watcher
	currentConfig *Config
	eventChan     chan WatcherEvent
	stopChan      chan struct{}
	watchedPaths  []string
	running       bool

	// Configuration for the watcher
	debounceTime     time.Duration
	validateOnChange bool
}

// WatcherOptions configures the behavior of the ConfigWatcher
type WatcherOptions struct {
	// DebounceTime specifies how long to wait before processing changes
	// to avoid processing rapid successive changes
	DebounceTime time.Duration

	// ValidateOnChange specifies whether to validate configuration after reload
	ValidateOnChange bool

	// EventBufferSize specifies the size of the event channel buffer
	EventBufferSize int
}

// DefaultWatcherOptions returns default options for the watcher
func DefaultWatcherOptions() *WatcherOptions {
	return &WatcherOptions{
		DebounceTime:     500 * time.Millisecond,
		ValidateOnChange: true,
		EventBufferSize:  100,
	}
}

// NewConfigWatcher creates a new configuration file watcher
func NewConfigWatcher(opts *WatcherOptions) (*ConfigWatcher, error) {
	if opts == nil {
		opts = DefaultWatcherOptions()
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &ConfigWatcher{
		watcher:          fsWatcher,
		eventChan:        make(chan WatcherEvent, opts.EventBufferSize),
		stopChan:         make(chan struct{}),
		debounceTime:     opts.DebounceTime,
		validateOnChange: opts.ValidateOnChange,
	}, nil
}

// Start begins monitoring configuration files for changes
func (w *ConfigWatcher) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return fmt.Errorf("watcher is already running")
	}

	// Load initial configuration
	config, err := LoadEffectiveConfig()
	if err != nil {
		return fmt.Errorf("failed to load initial configuration: %w", err)
	}
	w.currentConfig = config

	// Determine configuration file paths to watch
	paths, err := w.getConfigurationPaths()
	if err != nil {
		return fmt.Errorf("failed to determine configuration paths: %w", err)
	}

	// Add paths to file watcher
	for _, path := range paths {
		// Watch the directory containing the config file, not the file itself
		// This handles cases where the file is replaced (common with editors)
		dir := filepath.Dir(path)
		if err := w.watcher.Add(dir); err != nil {
			// Log error but continue with other paths
			w.sendEvent(WatcherEvent{
				Type:      EventConfigError,
				Path:      path,
				Error:     fmt.Errorf("failed to watch directory %s: %w", dir, err),
				Timestamp: time.Now(),
			})
		}
	}
	w.watchedPaths = paths

	w.running = true

	// Send started event
	w.sendEvent(WatcherEvent{
		Type:      EventWatcherStarted,
		Config:    w.currentConfig,
		Timestamp: time.Now(),
	})

	// Start the main watch loop
	go w.watchLoop(ctx)

	return nil
}

// Stop stops the configuration watcher
func (w *ConfigWatcher) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return nil
	}

	w.running = false
	close(w.stopChan)

	err := w.watcher.Close()

	// Send stopped event
	w.sendEvent(WatcherEvent{
		Type:      EventWatcherStopped,
		Timestamp: time.Now(),
	})

	return err
}

// Events returns the channel for receiving watcher events
func (w *ConfigWatcher) Events() <-chan WatcherEvent {
	return w.eventChan
}

// GetCurrentConfig returns the current configuration
func (w *ConfigWatcher) GetCurrentConfig() *Config {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.currentConfig
}

// watchLoop is the main event processing loop
func (w *ConfigWatcher) watchLoop(ctx context.Context) {
	debouncer := make(map[string]*time.Timer)

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Check if this is a configuration file we care about
			if !w.isConfigurationFile(event.Name) {
				continue
			}

			// Debounce rapid changes
			if timer, exists := debouncer[event.Name]; exists {
				timer.Stop()
			}

			debouncer[event.Name] = time.AfterFunc(w.debounceTime, func() {
				w.handleConfigChange(event.Name)
				delete(debouncer, event.Name)
			})

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}

			w.sendEvent(WatcherEvent{
				Type:      EventConfigError,
				Error:     fmt.Errorf("file watcher error: %w", err),
				Timestamp: time.Now(),
			})
		}
	}
}

// handleConfigChange processes a configuration file change
func (w *ConfigWatcher) handleConfigChange(path string) {
	// Check if file still exists (might have been deleted)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File was deleted, reload configuration without this file
		w.reloadConfiguration()
		return
	}

	// Wait a moment for file operations to complete
	time.Sleep(50 * time.Millisecond)

	w.reloadConfiguration()
}

// reloadConfiguration reloads the configuration from all sources
func (w *ConfigWatcher) reloadConfiguration() {
	// Load new configuration without validation (we'll validate separately)
	newConfig, err := LoadEffectiveConfigWithOptions(false)
	if err != nil {
		w.sendEvent(WatcherEvent{
			Type:      EventConfigError,
			Error:     fmt.Errorf("failed to reload configuration: %w", err),
			Timestamp: time.Now(),
		})
		return
	}

	// Validate new configuration if enabled
	if w.validateOnChange {
		if errs := newConfig.Validate(); len(errs) > 0 {
			// Create a combined error from all validation errors
			var errorMsg string
			for i, err := range errs {
				if i > 0 {
					errorMsg += "; "
				}
				errorMsg += err.Error()
			}

			w.sendEvent(WatcherEvent{
				Type:      EventConfigValidation,
				Error:     fmt.Errorf("configuration validation failed: %s", errorMsg),
				Timestamp: time.Now(),
			})
			return
		}
	}

	// Update current configuration
	w.mu.Lock()
	w.currentConfig = newConfig
	w.mu.Unlock()

	// Send success event
	w.sendEvent(WatcherEvent{
		Type:      EventConfigChanged,
		Config:    newConfig,
		Timestamp: time.Now(),
	})
}

// getConfigurationPaths returns all possible configuration file paths
func (w *ConfigWatcher) getConfigurationPaths() ([]string, error) {
	var paths []string

	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return nil, fmt.Errorf("failed to get XDG directories: %w", err)
	}

	// System configuration
	systemPath := xdg.GetSystemConfigPath()
	if _, err := os.Stat(systemPath); err == nil {
		paths = append(paths, systemPath)
	}

	// User configuration
	userPath := dirs.GetConfigFilePath()
	if _, err := os.Stat(userPath); err == nil {
		paths = append(paths, userPath)
	}

	// Project configuration
	currentDir, err := os.Getwd()
	if err == nil {
		projectPath := w.findProjectConfigPath(currentDir)
		if projectPath != "" {
			paths = append(paths, projectPath)
		}
	}

	return paths, nil
}

// findProjectConfigPath finds the project configuration file path
func (w *ConfigWatcher) findProjectConfigPath(dir string) string {
	projectConfigPath := xdg.GetProjectConfigPath(dir)
	if _, err := os.Stat(projectConfigPath); err == nil {
		return projectConfigPath
	}

	// Check parent directories
	parentDir := filepath.Dir(dir)
	if parentDir == dir {
		return ""
	}

	return w.findProjectConfigPath(parentDir)
}

// isConfigurationFile checks if the given path is a configuration file we should monitor
func (w *ConfigWatcher) isConfigurationFile(path string) bool {
	fileName := filepath.Base(path)

	// Check for our configuration file names
	if fileName == "pvm.toml" {
		return true
	}

	// Check if it's in any of our watched paths
	for _, watchedPath := range w.watchedPaths {
		if path == watchedPath {
			return true
		}
	}

	return false
}

// sendEvent sends an event to the event channel
func (w *ConfigWatcher) sendEvent(event WatcherEvent) {
	select {
	case w.eventChan <- event:
	default:
		// Channel is full, drop the event to avoid blocking
		// In production, you might want to log this condition
	}
}

// IsRunning returns whether the watcher is currently running
func (w *ConfigWatcher) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// SetValidateOnChange enables or disables validation on configuration changes
func (w *ConfigWatcher) SetValidateOnChange(validate bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.validateOnChange = validate
}

// ForceReload forces a reload of the configuration
func (w *ConfigWatcher) ForceReload() {
	go w.reloadConfiguration()
}
