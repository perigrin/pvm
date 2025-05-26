// ABOUTME: Tests for configuration file watcher functionality
// ABOUTME: Comprehensive test coverage for file system monitoring and event handling

package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigWatcher_Creation(t *testing.T) {
	opts := DefaultWatcherOptions()
	watcher, err := NewConfigWatcher(opts)
	if err != nil {
		t.Fatalf("Failed to create config watcher: %v", err)
	}

	if watcher == nil {
		t.Fatal("Expected non-nil watcher")
	}

	if watcher.debounceTime != opts.DebounceTime {
		t.Errorf("Expected debounce time %v, got %v", opts.DebounceTime, watcher.debounceTime)
	}

	if watcher.validateOnChange != opts.ValidateOnChange {
		t.Errorf("Expected validate on change %v, got %v", opts.ValidateOnChange, watcher.validateOnChange)
	}
}

func TestConfigWatcher_StartStop(t *testing.T) {
	// Set up HOME environment variable for tests
	testHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", oldHome)

	// Create temporary test directory
	testDir := t.TempDir()

	// Create a test config file
	configContent := `
[pvm]
default_perl = "5.38.0"
build_jobs = 4
`
	configPath := filepath.Join(testDir, "pvm.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Change to test directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	// Create .pvm directory
	pvmDir := filepath.Join(testDir, ".pvm")
	if err := os.MkdirAll(pvmDir, 0755); err != nil {
		t.Fatalf("Failed to create .pvm directory: %v", err)
	}

	// Move config to .pvm directory
	pvmConfigPath := filepath.Join(pvmDir, "pvm.toml")
	if err := os.Rename(configPath, pvmConfigPath); err != nil {
		t.Fatalf("Failed to move config to .pvm directory: %v", err)
	}

	opts := DefaultWatcherOptions()
	opts.DebounceTime = 100 * time.Millisecond

	watcher, err := NewConfigWatcher(opts)
	if err != nil {
		t.Fatalf("Failed to create config watcher: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test start
	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	if !watcher.IsRunning() {
		t.Error("Expected watcher to be running")
	}

	// Wait for start event
	select {
	case event := <-watcher.Events():
		if event.Type != EventWatcherStarted {
			t.Errorf("Expected watcher started event, got %v", event.Type)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for watcher started event")
	}

	// Test stop
	if err := watcher.Stop(); err != nil {
		t.Fatalf("Failed to stop watcher: %v", err)
	}

	if watcher.IsRunning() {
		t.Error("Expected watcher to be stopped")
	}
}

func TestConfigWatcher_FileChangeDetection(t *testing.T) {
	// Set up HOME environment variable for tests
	testHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", oldHome)

	// Create temporary test directory
	testDir := t.TempDir()

	// Create initial config file
	configContent := `
[pvm]
default_perl = "5.38.0"
build_jobs = 4
`

	// Create .pvm directory
	pvmDir := filepath.Join(testDir, ".pvm")
	if err := os.MkdirAll(pvmDir, 0755); err != nil {
		t.Fatalf("Failed to create .pvm directory: %v", err)
	}

	configPath := filepath.Join(pvmDir, "pvm.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Change to test directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	opts := DefaultWatcherOptions()
	opts.DebounceTime = 100 * time.Millisecond

	watcher, err := NewConfigWatcher(opts)
	if err != nil {
		t.Fatalf("Failed to create config watcher: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer watcher.Stop()

	// Wait for start event
	select {
	case <-watcher.Events():
		// Consume start event
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for watcher started event")
	}

	// Modify the config file
	newContent := `
[pvm]
default_perl = "5.40.0"
build_jobs = 8
`

	if err := os.WriteFile(configPath, []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to update config file: %v", err)
	}

	// Wait for config change event
	select {
	case event := <-watcher.Events():
		if event.Type != EventConfigChanged {
			t.Errorf("Expected config changed event, got %v", event.Type)
		}
		if event.Config == nil {
			t.Error("Expected config in change event")
		} else {
			if event.Config.PVM.DefaultPerl != "5.40.0" {
				t.Errorf("Expected default perl 5.40.0, got %s", event.Config.PVM.DefaultPerl)
			}
			if event.Config.PVM.BuildJobs != 8 {
				t.Errorf("Expected build jobs 8, got %d", event.Config.PVM.BuildJobs)
			}
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for config change event")
	}
}

func TestConfigWatcher_ValidationFailure(t *testing.T) {
	// Set up HOME environment variable for tests
	testHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", oldHome)

	// Create temporary test directory
	testDir := t.TempDir()

	// Create initial valid config file
	configContent := `
[pvm]
default_perl = "5.38.0"
build_jobs = 4
`

	// Create .pvm directory
	pvmDir := filepath.Join(testDir, ".pvm")
	if err := os.MkdirAll(pvmDir, 0755); err != nil {
		t.Fatalf("Failed to create .pvm directory: %v", err)
	}

	configPath := filepath.Join(pvmDir, "pvm.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Change to test directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	opts := DefaultWatcherOptions()
	opts.DebounceTime = 100 * time.Millisecond
	opts.ValidateOnChange = true

	watcher, err := NewConfigWatcher(opts)
	if err != nil {
		t.Fatalf("Failed to create config watcher: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer watcher.Stop()

	// Wait for start event
	select {
	case <-watcher.Events():
		// Consume start event
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for watcher started event")
	}

	// Create invalid config (negative build jobs)
	invalidContent := `
[pvm]
default_perl = "5.38.0"
build_jobs = -1
`

	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Wait for validation error event
	select {
	case event := <-watcher.Events():
		if event.Type != EventConfigValidation {
			t.Errorf("Expected config validation event, got %v", event.Type)
		}
		if event.Error == nil {
			t.Error("Expected error in validation event")
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for validation error event")
	}
}

func TestConfigWatcher_MultipleFiles(t *testing.T) {
	// Set up HOME environment variable for tests
	testHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", oldHome)

	// Create temporary test directory
	testDir := t.TempDir()

	// Create initial config files in different locations
	pvmDir := filepath.Join(testDir, ".pvm")
	if err := os.MkdirAll(pvmDir, 0755); err != nil {
		t.Fatalf("Failed to create .pvm directory: %v", err)
	}

	// Project config
	projectContent := `
[pvm]
default_perl = "5.38.0"
`
	projectConfigPath := filepath.Join(pvmDir, "pvm.toml")
	if err := os.WriteFile(projectConfigPath, []byte(projectContent), 0644); err != nil {
		t.Fatalf("Failed to create project config: %v", err)
	}

	// Change to test directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	opts := DefaultWatcherOptions()
	opts.DebounceTime = 100 * time.Millisecond

	watcher, err := NewConfigWatcher(opts)
	if err != nil {
		t.Fatalf("Failed to create config watcher: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer watcher.Stop()

	// Wait for start event
	select {
	case <-watcher.Events():
		// Consume start event
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for watcher started event")
	}

	// Update project config
	newProjectContent := `
[pvm]
default_perl = "5.40.0"
build_jobs = 8
`

	if err := os.WriteFile(projectConfigPath, []byte(newProjectContent), 0644); err != nil {
		t.Fatalf("Failed to update project config: %v", err)
	}

	// Wait for change event
	select {
	case event := <-watcher.Events():
		if event.Type != EventConfigChanged {
			t.Errorf("Expected config changed event, got %v", event.Type)
		}
		if event.Config == nil {
			t.Error("Expected config in change event")
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for config change event")
	}
}

func TestConfigWatcher_Debouncing(t *testing.T) {
	// Set up HOME environment variable for tests
	testHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", oldHome)

	// Create temporary test directory
	testDir := t.TempDir()

	// Create .pvm directory
	pvmDir := filepath.Join(testDir, ".pvm")
	if err := os.MkdirAll(pvmDir, 0755); err != nil {
		t.Fatalf("Failed to create .pvm directory: %v", err)
	}

	configPath := filepath.Join(pvmDir, "pvm.toml")
	initialContent := `
[pvm]
default_perl = "5.38.0"
build_jobs = 4
`
	if err := os.WriteFile(configPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Change to test directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	opts := DefaultWatcherOptions()
	opts.DebounceTime = 300 * time.Millisecond // Longer debounce for testing

	watcher, err := NewConfigWatcher(opts)
	if err != nil {
		t.Fatalf("Failed to create config watcher: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer watcher.Stop()

	// Wait for start event
	select {
	case <-watcher.Events():
		// Consume start event
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for watcher started event")
	}

	// Make rapid changes to the file
	for i := 0; i < 5; i++ {
		content := fmt.Sprintf(`
[pvm]
default_perl = "5.38.0"
build_jobs = %d
`, 4+i)
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to update config file: %v", err)
		}
		time.Sleep(50 * time.Millisecond) // Quick succession
	}

	// Should only get one change event due to debouncing
	eventCount := 0
	timeout := time.After(1 * time.Second)

	for {
		select {
		case event := <-watcher.Events():
			if event.Type == EventConfigChanged {
				eventCount++
				// Should be the last change (build_jobs = 8)
				if event.Config != nil && event.Config.PVM.BuildJobs != 8 {
					t.Errorf("Expected final build jobs value 8, got %d", event.Config.PVM.BuildJobs)
				}
			}
		case <-timeout:
			goto done
		}
	}

done:
	if eventCount != 1 {
		t.Errorf("Expected 1 config change event due to debouncing, got %d", eventCount)
	}
}

func TestConfigWatcher_ForceReload(t *testing.T) {
	// Set up HOME environment variable for tests
	testHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", oldHome)

	// Create temporary test directory
	testDir := t.TempDir()

	// Create .pvm directory
	pvmDir := filepath.Join(testDir, ".pvm")
	if err := os.MkdirAll(pvmDir, 0755); err != nil {
		t.Fatalf("Failed to create .pvm directory: %v", err)
	}

	configPath := filepath.Join(pvmDir, "pvm.toml")
	configContent := `
[pvm]
default_perl = "5.38.0"
build_jobs = 4
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Change to test directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	watcher, err := NewConfigWatcher(DefaultWatcherOptions())
	if err != nil {
		t.Fatalf("Failed to create config watcher: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer watcher.Stop()

	// Wait for start event
	select {
	case <-watcher.Events():
		// Consume start event
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for watcher started event")
	}

	// Force a reload
	watcher.ForceReload()

	// Should get a config change event
	select {
	case event := <-watcher.Events():
		if event.Type != EventConfigChanged {
			t.Errorf("Expected config changed event, got %v", event.Type)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for forced reload event")
	}
}
