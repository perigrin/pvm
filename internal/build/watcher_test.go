// ABOUTME: Tests for the continuous build system with file watching
// ABOUTME: Covers file change detection, build triggering, and event processing

package build

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"tamarou.com/pvm/internal/project"
)

func TestBuildWatcher_NewBuildWatcher(t *testing.T) {
	// Create temporary project directory
	tmpDir := t.TempDir()
	
	// Create project context
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   tmpDir,
	}
	
	config := DefaultWatcherConfig()
	watcher, err := NewBuildWatcher(projectCtx, config)
	
	if err != nil {
		t.Fatalf("Failed to create build watcher: %v", err)
	}
	
	if watcher == nil {
		t.Fatal("Build watcher is nil")
	}
	
	if watcher.projectContext != projectCtx {
		t.Error("Project context not set correctly")
	}
	
	if len(watcher.watchDirs) != len(config.WatchDirs) {
		t.Errorf("Watch directories not set correctly: got %d, want %d", len(watcher.watchDirs), len(config.WatchDirs))
	}
	
	if watcher.debounceDelay != config.DebounceDelay {
		t.Errorf("Debounce delay not set correctly: got %v, want %v", watcher.debounceDelay, config.DebounceDelay)
	}
}

func TestBuildWatcher_DefaultConfig(t *testing.T) {
	config := DefaultWatcherConfig()
	
	if config == nil {
		t.Fatal("Default config is nil")
	}
	
	expectedDirs := []string{"lib", "script", "t"}
	if len(config.WatchDirs) != len(expectedDirs) {
		t.Errorf("Default watch directories: got %v, want %v", config.WatchDirs, expectedDirs)
	}
	
	if config.DebounceDelay != 500*time.Millisecond {
		t.Errorf("Default debounce delay: got %v, want %v", config.DebounceDelay, 500*time.Millisecond)
	}
	
	if !config.EnableTypeCheck {
		t.Error("Type checking should be enabled by default")
	}
	
	if !config.EnableInline {
		t.Error("Inline builds should be enabled by default")
	}
	
	if config.EnableDist {
		t.Error("Distribution builds should be disabled by default")
	}
}

func TestBuildWatcher_FileFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   tmpDir,
	}
	
	config := DefaultWatcherConfig()
	watcher, err := NewBuildWatcher(projectCtx, config)
	if err != nil {
		t.Fatalf("Failed to create build watcher: %v", err)
	}
	
	testCases := []struct {
		path     string
		expected bool
		desc     string
	}{
		{"lib/MyModule.pm", true, "Perl module should be watched"},
		{"script/myscript.pl", true, "Perl script should be watched"},
		{"t/test.t", true, "Test file should be watched"},
		{"lib/MyModule.pmc", false, "Compiled module should be ignored"},
		{"build/lib/MyModule.pm", false, "Build directory should be ignored"},
		{"local/lib/perl5/Module.pm", false, "Local lib should be ignored"},
		{".git/config", false, "Git files should be ignored"},
		{"lib/temp.tmp", false, "Temporary files should be ignored"},
		{"lib/backup.bak", false, "Backup files should be ignored"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			fullPath := filepath.Join(tmpDir, tc.path)
			result := watcher.shouldWatchFile(fullPath)
			if result != tc.expected {
				t.Errorf("shouldWatchFile(%s): got %v, want %v", tc.path, result, tc.expected)
			}
		})
	}
}

func TestBuildWatcher_FileTypeDetection(t *testing.T) {
	tmpDir := t.TempDir()
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   tmpDir,
	}
	
	watcher, err := NewBuildWatcher(projectCtx, nil)
	if err != nil {
		t.Fatalf("Failed to create build watcher: %v", err)
	}
	
	testCases := []struct {
		path       string
		isSource   bool
		isTest     bool
		isConfig   bool
		desc       string
	}{
		{"lib/MyModule.pm", true, false, false, "Perl module is source"},
		{"script/myscript.pl", true, false, false, "Perl script is source"},
		{"t/test.t", false, true, false, "Test file is test"},
		{"pvm.toml", false, false, true, "PVM config is config"},
		{"cpanfile", false, false, true, "cpanfile is config"},
		{".perl-version", false, false, true, "perl-version is config"},
		{"README.md", false, false, false, "Documentation is none"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			fullPath := filepath.Join(tmpDir, tc.path)
			
			if watcher.isSourceFile(fullPath) != tc.isSource {
				t.Errorf("isSourceFile(%s): got %v, want %v", tc.path, watcher.isSourceFile(fullPath), tc.isSource)
			}
			
			if watcher.isTestFile(fullPath) != tc.isTest {
				t.Errorf("isTestFile(%s): got %v, want %v", tc.path, watcher.isTestFile(fullPath), tc.isTest)
			}
			
			if watcher.isConfigFile(fullPath) != tc.isConfig {
				t.Errorf("isConfigFile(%s): got %v, want %v", tc.path, watcher.isConfigFile(fullPath), tc.isConfig)
			}
		})
	}
}

func TestBuildWatcher_StartStop(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create lib directory for file discovery
	libDir := filepath.Join(tmpDir, "lib")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}
	
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   tmpDir,
	}
	
	config := DefaultWatcherConfig()
	config.DebounceDelay = 100 * time.Millisecond // Faster for testing
	
	watcher, err := NewBuildWatcher(projectCtx, config)
	if err != nil {
		t.Fatalf("Failed to create build watcher: %v", err)
	}
	
	// Test start
	err = watcher.Start()
	if err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	
	if !watcher.isWatching {
		t.Error("Watcher should be watching after start")
	}
	
	// Test double start (should fail)
	err = watcher.Start()
	if err == nil {
		t.Error("Expected error when starting already running watcher")
	}
	
	// Test stop
	err = watcher.Stop()
	if err != nil {
		t.Fatalf("Failed to stop watcher: %v", err)
	}
	
	if watcher.isWatching {
		t.Error("Watcher should not be watching after stop")
	}
	
	// Test double stop (should not fail)
	err = watcher.Stop()
	if err != nil {
		t.Errorf("Unexpected error when stopping already stopped watcher: %v", err)
	}
}

func TestBuildWatcher_FileInitialization(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create test files
	libDir := filepath.Join(tmpDir, "lib")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}
	
	testFile := filepath.Join(libDir, "TestModule.pm")
	if err := os.WriteFile(testFile, []byte("package TestModule; 1;"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   tmpDir,
	}
	
	watcher, err := NewBuildWatcher(projectCtx, nil)
	if err != nil {
		t.Fatalf("Failed to create build watcher: %v", err)
	}
	
	// Initialize file states
	err = watcher.initializeFileStates()
	if err != nil {
		t.Fatalf("Failed to initialize file states: %v", err)
	}
	
	// Check that the test file was tracked
	if _, exists := watcher.fileModTimes[testFile]; !exists {
		t.Errorf("Test file %s was not tracked during initialization", testFile)
	}
}

func TestBuildWatcher_TriggerBuild(t *testing.T) {
	tmpDir := t.TempDir()
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   tmpDir,
	}
	
	watcher, err := NewBuildWatcher(projectCtx, nil)
	if err != nil {
		t.Fatalf("Failed to create build watcher: %v", err)
	}
	
	// Start watcher
	err = watcher.Start()
	if err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer watcher.Stop()
	
	// Trigger a build
	watcher.TriggerBuild(BuildTypeTypeCheck, "manual trigger")
	
	// Wait for build result
	select {
	case result := <-watcher.Results():
		if result.Type != BuildTypeTypeCheck {
			t.Errorf("Build type: got %v, want %v", result.Type, BuildTypeTypeCheck)
		}
		// Note: The build might fail due to missing dependencies, but that's expected in test
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for build result")
	}
}

func TestBuildWatcher_BuildTypes(t *testing.T) {
	testCases := []struct {
		buildType BuildType
		expected  string
	}{
		{BuildTypeTypeCheck, "typecheck"},
		{BuildTypeInline, "inline"},
		{BuildTypeDistribution, "distribution"},
		{BuildTypeFull, "full"},
		{BuildType(999), "unknown"}, // Invalid type
	}
	
	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.buildType.String()
			if result != tc.expected {
				t.Errorf("BuildType.String(): got %s, want %s", result, tc.expected)
			}
		})
	}
}