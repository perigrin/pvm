// ABOUTME: Tests for enhanced Perl build system
// ABOUTME: Validates error handling, dependency checking, and build optimizations

package perl

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"tamarou.com/pvm/internal/log"
)

func TestBuildManager_ValidateBuildOptions(t *testing.T) {
	logger := log.NewLogger(log.LevelDebug, os.Stderr, "test")
	bm, err := NewBuildManager(logger)
	if err != nil {
		t.Fatalf("Failed to create build manager: %v", err)
	}

	tests := []struct {
		name    string
		options *BuildOptions
		wantErr bool
	}{
		{
			name:    "nil options",
			options: nil,
			wantErr: true,
		},
		{
			name:    "empty version",
			options: &BuildOptions{},
			wantErr: true,
		},
		{
			name: "invalid version format",
			options: &BuildOptions{
				Version: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid options",
			options: &BuildOptions{
				Version: "5.38.2",
			},
			wantErr: false,
		},
		{
			name: "options with defaults",
			options: &BuildOptions{
				Version:   "5.38.2",
				BuildJobs: -1, // Should be set to CPU count
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bm.validateBuildOptions(tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBuildOptions() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check defaults are set
			if !tt.wantErr && tt.options != nil {
				if tt.options.Context == nil {
					t.Error("Context should be set to default")
				}
				if tt.options.BuildJobs <= 0 {
					t.Error("BuildJobs should be set to positive value")
				}
				if tt.options.TestTimeout <= 0 {
					t.Error("TestTimeout should be set")
				}
			}
		})
	}
}

func TestDependencyChecker(t *testing.T) {
	dc, err := NewDependencyChecker()
	if err != nil {
		t.Fatalf("Failed to create dependency checker: %v", err)
	}

	// Check dependencies
	info, err := dc.CheckBuildDependencies()
	if err != nil {
		t.Fatalf("Failed to check dependencies: %v", err)
	}

	// We should have some required dependencies detected
	if len(info.Required) == 0 && len(info.Missing) == 0 {
		t.Error("No dependencies detected")
	}

	// Check that basic tools are detected
	basicTools := []string{"make", "tar", "gzip"}
	for _, tool := range basicTools {
		found := false
		for _, dep := range info.Required {
			if dep == tool {
				found = true
				break
			}
		}
		for _, dep := range info.Missing {
			if dep == tool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Basic tool %s not checked", tool)
		}
	}

	// Verify install hints are provided for missing deps
	for _, dep := range info.Missing {
		if _, ok := info.InstallHint[dep]; !ok {
			t.Errorf("No install hint for missing dependency: %s", dep)
		}
	}
}

func TestChecksumDatabase(t *testing.T) {
	db, err := NewChecksumDatabase()
	if err != nil {
		t.Fatalf("Failed to create checksum database: %v", err)
	}

	// Test known checksums
	tests := []struct {
		version  string
		hasValue bool
	}{
		{"5.40.0", true},
		{"5.38.2", true},
		{"5.36.3", true},
		{"5.99.99", false}, // Non-existent version
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			checksum, err := db.GetChecksum(tt.version)
			if tt.hasValue {
				if err != nil {
					t.Errorf("Expected checksum for %s, got error: %v", tt.version, err)
				}
				if checksum == "" {
					t.Errorf("Expected non-empty checksum for %s", tt.version)
				}
			} else if err == nil {
				t.Errorf("Expected error for unknown version %s", tt.version)
			}
		})
	}

	// Test adding custom checksum
	customVersion := "5.99.0"
	customChecksum := "abcdef0123456789"

	db.AddChecksum(customVersion, customChecksum)

	retrieved, err := db.GetChecksum(customVersion)
	if err != nil {
		t.Errorf("Failed to retrieve custom checksum: %v", err)
	}
	if retrieved != customChecksum {
		t.Errorf("Custom checksum mismatch: got %s, want %s", retrieved, customChecksum)
	}
}

func TestBuildCache(t *testing.T) {
	// Create temporary cache directory
	tempDir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", tempDir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	cache, err := NewBuildCache()
	if err != nil {
		t.Fatalf("Failed to create build cache: %v", err)
	}

	// Test cache miss
	options := &BuildOptions{
		Version:          "5.38.2",
		ConfigureOptions: []string{"-Dusethreads"},
	}

	result, found := cache.GetCachedBuild("5.38.2", options)
	if found {
		t.Error("Expected cache miss for new build")
	}
	if result != nil {
		t.Error("Expected nil result for cache miss")
	}

	// Save to cache
	buildResult := &BuildResult{
		Version:     "5.38.2",
		InstallPath: "/tmp/perl-5.38.2",
		BuildPath:   "/tmp/build",
		Duration:    10 * time.Minute,
		Timestamp:   time.Now(),
	}

	configCache := map[string]string{
		"cc":     "gcc",
		"osname": "linux",
	}

	err = cache.SaveBuild("5.38.2", options, buildResult, configCache)
	if err != nil {
		t.Fatalf("Failed to save to cache: %v", err)
	}

	// Test cache hit
	// Create dummy install directory for verification
	os.MkdirAll(buildResult.InstallPath, 0755)
	defer os.RemoveAll(buildResult.InstallPath)

	cachedResult, found := cache.GetCachedBuild("5.38.2", options)
	if !found {
		t.Error("Expected cache hit")
	}
	if cachedResult == nil {
		t.Error("Expected non-nil cached result")
	} else if cachedResult.InstallPath != buildResult.InstallPath {
		t.Errorf("Cached install path mismatch: got %s, want %s",
			cachedResult.InstallPath, buildResult.InstallPath)
	}

	// Test config cache
	config, found := cache.GetConfigCache("5.38.2")
	if !found {
		t.Error("Expected config cache hit")
	}
	if config["cc"] != "gcc" {
		t.Errorf("Config cache mismatch: got %s, want gcc", config["cc"])
	}

	// Test cache expiration
	oldEntry := cache.entries[cache.calculateBuildHash("5.38.2", options)]
	oldEntry.Timestamp = time.Now().Add(-8 * 24 * time.Hour) // 8 days old

	_, found = cache.GetCachedBuild("5.38.2", options)
	if found {
		t.Error("Expected cache miss for expired entry")
	}
}

func TestArchiveExtractor(t *testing.T) {
	logger := log.NewLogger(log.LevelDebug, os.Stderr, "test")
	extractor := NewArchiveExtractor(logger)

	// This would test extraction with a small test archive
	// For now, just test the interface
	if extractor == nil {
		t.Error("Failed to create archive extractor")
	}
}

func TestCommandRunner(t *testing.T) {
	logger := log.NewLogger(log.LevelDebug, os.Stderr, "test")
	runner := NewCommandRunner(logger)

	// Test simple command
	output, err := runner.Run(".", "echo", []string{"hello"}, context.Background())
	if err != nil {
		t.Fatalf("Failed to run echo command: %v", err)
	}
	if !strings.Contains(output, "hello") {
		t.Errorf("Expected output to contain 'hello', got: %s", output)
	}

	// Test command with progress - use a command that definitely produces multiple lines
	var lines []string
	_, err = runner.RunWithProgress(
		".",
		"sh",
		[]string{"-c", "echo line1; echo line2; echo line3"},
		context.Background(),
		func(line string, isError bool) {
			lines = append(lines, line)
		},
	)
	if err != nil {
		t.Fatalf("Failed to run command with progress: %v", err)
	}
	if len(lines) == 0 {
		t.Errorf("Expected progress callbacks, got %d lines", len(lines))
	}
	t.Logf("Progress test captured %d lines: %v", len(lines), lines)

	// Test command timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err = runner.Run(".", "sleep", []string{"1"}, ctx)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestParallelExecutor(t *testing.T) {
	logger := log.NewLogger(log.LevelDebug, os.Stderr, "test")
	executor := NewParallelExecutor(2, logger)

	// Test successful parallel execution
	counter := 0
	mu := &sync.Mutex{}

	tasks := []Task{
		{
			Name: "task1",
			Fn: func() error {
				mu.Lock()
				counter++
				mu.Unlock()
				return nil
			},
		},
		{
			Name: "task2",
			Fn: func() error {
				mu.Lock()
				counter++
				mu.Unlock()
				return nil
			},
		},
		{
			Name: "task3",
			Fn: func() error {
				mu.Lock()
				counter++
				mu.Unlock()
				return nil
			},
		},
	}

	err := executor.Execute(context.Background(), tasks)
	if err != nil {
		t.Fatalf("Parallel execution failed: %v", err)
	}

	if counter != 3 {
		t.Errorf("Expected 3 tasks executed, got %d", counter)
	}

	// Test with error
	errorTasks := []Task{
		{
			Name: "success",
			Fn:   func() error { return nil },
		},
		{
			Name: "failure",
			Fn:   func() error { return fmt.Errorf("task error") },
		},
	}

	err = executor.Execute(context.Background(), errorTasks)
	if err == nil {
		t.Error("Expected error from failed task")
	}
}

func TestGetPlatformSpecificNotes(t *testing.T) {
	notes := GetPlatformSpecificNotes()
	if notes == "" {
		t.Error("Expected platform-specific notes")
	}

	// Check that notes contain platform name
	if !strings.Contains(notes, runtime.GOOS) &&
		!strings.Contains(notes, "macOS") &&
		!strings.Contains(notes, "Linux") &&
		!strings.Contains(notes, "Windows") {
		t.Error("Platform-specific notes don't mention platform")
	}
}

// Integration test for build process (requires mock implementation)
func TestBuildManagerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This would test the full build process with mocked components
	// For now, just ensure the build manager can be created
	logger := log.NewLogger(log.LevelDebug, os.Stderr, "test")
	bm, err := NewBuildManager(logger)
	if err != nil {
		t.Fatalf("Failed to create build manager: %v", err)
	}

	if bm.checksumsDB == nil {
		t.Error("Checksum database not initialized")
	}
	if bm.depChecker == nil {
		t.Error("Dependency checker not initialized")
	}
	if bm.buildCache == nil {
		t.Error("Build cache not initialized")
	}
}
