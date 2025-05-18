// ABOUTME: Tests for Perl build and installation functionality
// ABOUTME: Provides tests for building Perl from source

package perl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/xdg"
)

// Create a mock function type for os.RemoveAll that we'll use for testing
var osRemoveAllFunc = os.RemoveAll

// Helper function to reset mocks after tests
func resetMocks() {
	osRemoveAllFunc = os.RemoveAll
	runCommandWithProgressFunc = doRunCommandWithProgress
	DownloadPerlSourceFunc = doDownloadPerlSource
	extractArchiveFunc = doExtractArchive
}

// Our init function is now much simpler since we're using function variables
func init() {
	// The variables runCommandWithProgressFunc, DownloadPerlSourceFunc, and extractArchiveFunc
	// are already defined in the production code and can be replaced directly
}

// Setup temporary directories for testing
func setupTestDirs(t *testing.T) (*xdg.Dirs, string) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "pvm-build-test-*")
	require.NoError(t, err)

	// Create XDG directories
	appDir := "pvm"
	dirs := &xdg.Dirs{
		// XDG standard directories
		ConfigHome: filepath.Join(tempDir, "config"),
		DataHome:   filepath.Join(tempDir, "data"),
		CacheHome:  filepath.Join(tempDir, "cache"),
		StateHome:  filepath.Join(tempDir, "state"),
		
		// Application-specific directories
		ConfigDir: filepath.Join(tempDir, "config", appDir),
		CacheDir:  filepath.Join(tempDir, "cache", appDir),
		DataDir:   filepath.Join(tempDir, "data", appDir),
		StateDir:  filepath.Join(tempDir, "state", appDir),
		
		// PVM-specific directories
		VersionsDir:        filepath.Join(tempDir, "data", appDir, "versions"),
		SourcesDir:         filepath.Join(tempDir, "cache", appDir, "sources"),
		ShimsDir:           filepath.Join(tempDir, "data", appDir, "shims"),
		TypeDefinitionsDir: filepath.Join(tempDir, "data", appDir, "type_definitions"),
		BuildDir:           filepath.Join(tempDir, "cache", appDir, "build"),
	}
	
	// Set the EnsureDirs function
	dirs.EnsureDirs = func() error {
		// Create all required directories
		dirsToCreate := []string{
			dirs.ConfigDir,
			dirs.CacheDir,
			dirs.DataDir,
			dirs.StateDir,
			dirs.VersionsDir,
			dirs.SourcesDir,
			dirs.ShimsDir,
			dirs.TypeDefinitionsDir,
			dirs.BuildDir,
		}
		
		for _, dir := range dirsToCreate {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
		
		return nil
	}

	// Create the directories
	err = dirs.EnsureDirs()
	require.NoError(t, err)

	// Replace the GetDirs function for testing
	originalGetDirs := xdg.GetDirs
	t.Cleanup(func() {
		xdg.GetDirs = originalGetDirs
		os.RemoveAll(tempDir)
		resetMocks()
	})

	xdg.GetDirs = func() (*xdg.Dirs, error) {
		return dirs, nil
	}

	return dirs, tempDir
}

// Create a mock source archive for testing
func createMockArchive(t *testing.T, dirs *xdg.Dirs, version string) string {
	// Determine the filename
	archiveFilename := fmt.Sprintf("perl-%s.tar.gz", version)
	archivePath := filepath.Join(dirs.SourcesDir, archiveFilename)

	// Create a minimal tar.gz file (just enough to be valid)
	f, err := os.Create(archivePath)
	require.NoError(t, err)
	defer f.Close()

	// Just create an empty file - we'll mock the extraction
	// In a more thorough test, we would create a real tar.gz with minimal content
	
	return archivePath
}

// Mock successful extraction for testing
func mockExtraction(t *testing.T, tempDir string) func(archivePath, destDir string, ctx context.Context) (string, error) {
	return func(archivePath, destDir string, ctx context.Context) (string, error) {
		// Create a mock extracted directory
		extractedDir := filepath.Join(destDir, "perl-5.36.0")
		err := os.MkdirAll(extractedDir, 0755)
		require.NoError(t, err)
		
		// Create a mock Configure script
		configureFile := filepath.Join(extractedDir, "Configure")
		err = os.WriteFile(configureFile, []byte("#!/bin/sh\necho 'Mock Configure'"), 0755)
		require.NoError(t, err)
		
		return extractedDir, nil
	}
}

// TestBuildPerlBasic tests a basic successful build process
func TestBuildPerlBasic(t *testing.T) {
	// Setup test environment
	dirs, tempDir := setupTestDirs(t)
	
	// Save original extraction function and mock it
	originalExtractArchiveFunc := extractArchiveFunc
	t.Cleanup(func() {
		extractArchiveFunc = originalExtractArchiveFunc
	})
	extractArchiveFunc = mockExtraction(t, tempDir)
	
	// Create a mock source archive
	version := "5.36.0"
	archivePath := createMockArchive(t, dirs, version)
	
	// Mock the command execution
	runCommandWithProgressFunc = func(
		dir string,
		command string,
		args []string,
		ctx context.Context,
		progressCb func(line string),
	) error {
		// Report progress for testing the callback
		if progressCb != nil {
			progressCb(fmt.Sprintf("Mock progress for %s", command))
		}
		return nil
	}
	
	// Track progress calls
	var progressStages []BuildProgressStage
	var progressDetails []string
	progressCallback := func(stage BuildProgressStage, details string, progress float64) {
		progressStages = append(progressStages, stage)
		progressDetails = append(progressDetails, details)
	}
	
	// Create build options
	options := &BuildOptions{
		Version:          version,
		SourceFile:       archivePath,
		InstallDir:       filepath.Join(dirs.VersionsDir, version),
		BuildJobs:        2,
		RunTests:         true,
		CleanupBuildDir:  true,
		ProgressCallback: progressCallback,
		Context:          context.Background(),
	}
	
	// Run the build
	result, err := BuildPerl(options)
	
	// Verify results
	require.NoError(t, err)
	assert.Equal(t, version, result.Version)
	assert.Equal(t, filepath.Join(dirs.VersionsDir, version), result.InstallPath)
	
	// Verify progress reporting
	assert.Contains(t, progressStages, StageExtract)
	assert.Contains(t, progressStages, StageConfigure)
	assert.Contains(t, progressStages, StageCompile)
	assert.Contains(t, progressStages, StageTest)
	assert.Contains(t, progressStages, StageInstall)
	assert.Contains(t, progressStages, StageCleanup)
	assert.Contains(t, progressStages, StageDone)
}

// TestBuildPerlDownload tests the automatic download of source when not provided
func TestBuildPerlDownload(t *testing.T) {
	// Setup test environment
	dirs, tempDir := setupTestDirs(t)
	
	// Save original extraction function and mock it
	originalExtractArchiveFunc := extractArchiveFunc
	t.Cleanup(func() {
		extractArchiveFunc = originalExtractArchiveFunc
	})
	extractArchiveFunc = mockExtraction(t, tempDir)
	
	// Mock the source download
	version := "5.36.0"
	archivePath := filepath.Join(dirs.SourcesDir, fmt.Sprintf("perl-%s.tar.gz", version))
	
	DownloadPerlSourceFunc = func(options *DownloadOptions) (*DownloadResult, error) {
		assert.Equal(t, version, options.Version)
		
		// Create a mock source archive
		err := os.MkdirAll(filepath.Dir(archivePath), 0755)
		require.NoError(t, err)
		
		f, err := os.Create(archivePath)
		require.NoError(t, err)
		f.Close()
		
		return &DownloadResult{
			Version: version,
			Path:    archivePath,
		}, nil
	}
	
	// Mock the command execution
	runCommandWithProgressFunc = func(
		dir string,
		command string,
		args []string,
		ctx context.Context,
		progressCb func(line string),
	) error {
		if progressCb != nil {
			progressCb(fmt.Sprintf("Mock progress for %s", command))
		}
		return nil
	}
	
	// Create build options without source file
	options := &BuildOptions{
		Version:         version,
		InstallDir:      filepath.Join(dirs.VersionsDir, version),
		BuildJobs:       2,
		RunTests:        false,
		CleanupBuildDir: false,
		Context:         context.Background(),
	}
	
	// Run the build
	result, err := BuildPerl(options)
	
	// Verify results
	require.NoError(t, err)
	assert.Equal(t, version, result.Version)
	assert.NotZero(t, result.Duration)
}

// TestBuildPerlFailedConfigure tests handling of a failed Configure stage
func TestBuildPerlFailedConfigure(t *testing.T) {
	// Setup test environment
	dirs, tempDir := setupTestDirs(t)
	
	// Save original extraction function and mock it
	originalExtractArchiveFunc := extractArchiveFunc
	t.Cleanup(func() {
		extractArchiveFunc = originalExtractArchiveFunc
	})
	extractArchiveFunc = mockExtraction(t, tempDir)
	
	// Create a mock source archive
	version := "5.36.0"
	archivePath := createMockArchive(t, dirs, version)
	
	// Mock the command execution with a failure at Configure stage
	configureFailed := false
	runCommandWithProgressFunc = func(
		dir string,
		command string,
		args []string,
		ctx context.Context,
		progressCb func(line string),
	) error {
		if command == "./Configure" {
			configureFailed = true
			return fmt.Errorf("Mock Configure failure")
		}
		return nil
	}
	
	// Create build options
	options := &BuildOptions{
		Version:         version,
		SourceFile:      archivePath,
		InstallDir:      filepath.Join(dirs.VersionsDir, version),
		BuildJobs:       2,
		RunTests:        false,
		CleanupBuildDir: false,
		Context:         context.Background(),
	}
	
	// Run the build
	result, err := BuildPerl(options)
	
	// Verify results
	assert.Error(t, err)
	assert.True(t, configureFailed)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), ErrConfigureFailed)
}

// TestBuildPerlPlatformOptions tests platform-specific configure options
func TestBuildPerlPlatformOptions(t *testing.T) {
	// Get platform-specific configure options
	options, err := getPlatformConfigureOptions()
	
	// Verify results
	require.NoError(t, err)
	assert.NotEmpty(t, options)
	
	// Check that platform-specific options are present
	switch runtime.GOOS {
	case "darwin":
		assert.Contains(t, options, "-Duseshrplib")
		assert.Contains(t, options, "-Dusedtrace")
	case "linux":
		assert.Contains(t, options, "-Duseshrplib")
		assert.Contains(t, options, "-Duselargefiles")
	case "windows":
		assert.Contains(t, options, "-Duseshrplib")
		assert.Contains(t, options, "-Duseithreads")
	}
}

// TestBuildPerlCancellation tests cancellation of a build in progress
func TestBuildPerlCancellation(t *testing.T) {
	// Setup test environment
	dirs, tempDir := setupTestDirs(t)
	
	// Save original extraction function and mock it
	originalExtractArchiveFunc := extractArchiveFunc
	t.Cleanup(func() {
		extractArchiveFunc = originalExtractArchiveFunc
	})
	extractArchiveFunc = mockExtraction(t, tempDir)
	
	// Create a mock source archive
	version := "5.36.0"
	archivePath := createMockArchive(t, dirs, version)
	
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	
	// Cancel after a short delay
	cancelled := false
	configureStarted := false
	
	runCommandWithProgressFunc = func(
		dir string,
		command string,
		args []string,
		ctx context.Context,
		progressCb func(line string),
	) error {
		if command == "./Configure" {
			configureStarted = true
			// Signal to cancel
			go func() {
				time.Sleep(50 * time.Millisecond)
				cancelled = true
				cancel()
			}()
			
			// Simulate the command running
			time.Sleep(100 * time.Millisecond)
			
			// Check if context was cancelled
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		}
		return nil
	}
	
	// Create build options
	options := &BuildOptions{
		Version:         version,
		SourceFile:      archivePath,
		InstallDir:      filepath.Join(dirs.VersionsDir, version),
		BuildJobs:       2,
		RunTests:        false,
		CleanupBuildDir: false,
		Context:         ctx,
	}
	
	// Run the build
	result, err := BuildPerl(options)
	
	// Verify results
	assert.Error(t, err)
	assert.True(t, cancelled)
	assert.True(t, configureStarted)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.Canceled)
}

// TestBuildPerlWithDefaults tests using default values for options
func TestBuildPerlWithDefaults(t *testing.T) {
	// Setup test environment
	dirs, tempDir := setupTestDirs(t)
	
	// Save original extraction function and mock it
	originalExtractArchiveFunc := extractArchiveFunc
	t.Cleanup(func() {
		extractArchiveFunc = originalExtractArchiveFunc
	})
	extractArchiveFunc = mockExtraction(t, tempDir)
	
	// Mock the source download
	version := "5.36.0"
	archivePath := filepath.Join(dirs.SourcesDir, fmt.Sprintf("perl-%s.tar.gz", version))
	
	DownloadPerlSourceFunc = func(options *DownloadOptions) (*DownloadResult, error) {
		// Create a mock source archive
		err := os.MkdirAll(filepath.Dir(archivePath), 0755)
		require.NoError(t, err)
		
		f, err := os.Create(archivePath)
		require.NoError(t, err)
		f.Close()
		
		return &DownloadResult{
			Version: version,
			Path:    archivePath,
		}, nil
	}
	
	// Mock the command execution
	runCommandWithProgressFunc = func(
		dir string,
		command string,
		args []string,
		ctx context.Context,
		progressCb func(line string),
	) error {
		return nil
	}
	
	// Create build options with minimal configuration
	options := &BuildOptions{
		Version: version,
	}
	
	// Run the build
	result, err := BuildPerl(options)
	
	// Verify results
	require.NoError(t, err)
	assert.Equal(t, version, result.Version)
	assert.Equal(t, filepath.Join(dirs.VersionsDir, version), result.InstallPath)
	assert.NotEmpty(t, result.BuildPath)
	assert.NotZero(t, result.Duration)
}