// ABOUTME: Tests for Perl build and installation functionality
// ABOUTME: Provides tests for building Perl from source

package perl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
		_ = os.RemoveAll(tempDir)
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
	defer func() { _ = f.Close() }()

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
		_ = f.Close()

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
		_ = f.Close()

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

// TestBuildPerlRelocatableOptions tests that relocatable configure options are included
func TestBuildPerlRelocatableOptions(t *testing.T) {
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

	// Capture configure options
	var capturedConfigureArgs []string
	runCommandWithProgressFunc = func(
		dir string,
		command string,
		args []string,
		ctx context.Context,
		progressCb func(line string),
	) error {
		if command == "./Configure" {
			capturedConfigureArgs = args
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
	require.NoError(t, err)
	assert.Equal(t, version, result.Version)

	// Verify relocatable configure option is included
	assert.Contains(t, capturedConfigureArgs, "-Duserelocatableinc",
		"Configure options should include -Duserelocatableinc for relocatable builds")

	// Verify other expected default options are still present
	assert.Contains(t, capturedConfigureArgs, "-des")
	assert.Contains(t, capturedConfigureArgs, "-Dusethreads")

	// Verify prefix option is present
	prefixOption := fmt.Sprintf("-Dprefix=%s", filepath.Join(dirs.VersionsDir, version))
	assert.Contains(t, capturedConfigureArgs, prefixOption)
}

// TestBuildPerlCustomOutputDirectory tests building Perl with a custom output directory
func TestBuildPerlCustomOutputDirectory(t *testing.T) {
	// Setup test environment
	dirs, tempDir := setupTestDirs(t)

	// Create a custom output directory
	customOutputDir := filepath.Join(tempDir, "custom-perl-output")
	err := os.MkdirAll(customOutputDir, 0755)
	require.NoError(t, err)

	// Save original extraction function and mock it
	originalExtractArchiveFunc := extractArchiveFunc
	t.Cleanup(func() {
		extractArchiveFunc = originalExtractArchiveFunc
	})
	extractArchiveFunc = mockExtraction(t, tempDir)

	// Create a mock source archive
	version := "5.36.0"
	archivePath := createMockArchive(t, dirs, version)

	// Capture configure arguments to verify custom prefix
	var capturedConfigureArgs []string
	runCommandWithProgressFunc = func(
		dir string,
		command string,
		args []string,
		ctx context.Context,
		progressCb func(line string),
	) error {
		// Capture configure command arguments
		if command == "./Configure" {
			capturedConfigureArgs = args
		}

		// Report progress for testing the callback
		if progressCb != nil {
			progressCb(fmt.Sprintf("Mock progress for %s", command))
		}
		return nil
	}

	// Create build options with custom output directory
	options := &BuildOptions{
		Version:         version,
		SourceFile:      archivePath,
		InstallDir:      customOutputDir, // This is what gets set when --output-dir is used
		BuildJobs:       1,
		RunTests:        false,
		CleanupBuildDir: true,
		Context:         context.Background(),
	}

	// Run the build
	result, err := BuildPerl(options)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, version, result.Version)
	assert.Equal(t, customOutputDir, result.InstallPath)

	// Verify that the custom output directory was used as the prefix
	prefixOption := fmt.Sprintf("-Dprefix=%s", customOutputDir)
	assert.Contains(t, capturedConfigureArgs, prefixOption,
		"Configure options should include custom output directory as prefix")

	// Verify other expected default options are still present
	assert.Contains(t, capturedConfigureArgs, "-des")
	assert.Contains(t, capturedConfigureArgs, "-Dusethreads")

	// Verify the custom output directory exists and was used
	assert.DirExists(t, customOutputDir)
}

// TestBuildPerlBuildOnlyMode tests build-only mode that skips installation
func TestBuildPerlBuildOnlyMode(t *testing.T) {
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
	outputDir := filepath.Join(tempDir, "build-output", version)

	// Track which commands were called
	var calledCommands []string
	runCommandWithProgressFunc = func(
		dir string,
		command string,
		args []string,
		ctx context.Context,
		progressCb func(line string),
	) error {
		// Record command with args for better tracking
		cmdWithArgs := command
		if len(args) > 0 {
			cmdWithArgs = fmt.Sprintf("%s %s", command, strings.Join(args, " "))
		}
		calledCommands = append(calledCommands, cmdWithArgs)

		// Report progress for testing the callback
		if progressCb != nil {
			progressCb(fmt.Sprintf("Mock progress for %s", command))
		}
		return nil
	}

	// Track progress calls to verify install stage is NOT called
	var progressStages []BuildProgressStage
	var progressDetails []string
	progressCallback := func(stage BuildProgressStage, details string, progress float64) {
		progressStages = append(progressStages, stage)
		progressDetails = append(progressDetails, details)
	}

	// Create build options with BuildOnly enabled
	options := &BuildOptions{
		Version:          version,
		SourceFile:       archivePath,
		InstallDir:       outputDir,
		BuildJobs:        2,
		RunTests:         true,
		CleanupBuildDir:  true,
		BuildOnly:        true, // This is the key flag being tested
		ProgressCallback: progressCallback,
		Context:          context.Background(),
	}

	// Run the build
	result, err := BuildPerl(options)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, version, result.Version)
	assert.Equal(t, outputDir, result.InstallPath)

	// Verify that make install was NOT called
	foundMakeInstall := false
	for _, cmd := range calledCommands {
		if strings.Contains(cmd, "make install") {
			foundMakeInstall = true
			break
		}
	}
	assert.False(t, foundMakeInstall, "Build-only mode should not run 'make install'")

	// Verify expected commands were called (Configure, make, make test)
	foundConfigure := false
	foundMake := false
	foundMakeTest := false

	for _, cmd := range calledCommands {
		if strings.HasPrefix(cmd, "./Configure") {
			foundConfigure = true
		}
		if strings.HasPrefix(cmd, "make -j") {
			foundMake = true
		}
		if strings.HasPrefix(cmd, "make test") {
			foundMakeTest = true
		}
	}

	assert.True(t, foundConfigure, "Should run Configure command")
	assert.True(t, foundMake, "Should run make command")
	assert.True(t, foundMakeTest, "Should run make test command")

	// Verify progress reporting - install stage should NOT be present
	assert.Contains(t, progressStages, StageExtract)
	assert.Contains(t, progressStages, StageConfigure)
	assert.Contains(t, progressStages, StageCompile)
	assert.Contains(t, progressStages, StageTest)
	assert.NotContains(t, progressStages, StageInstall,
		"Build-only mode should not report install stage")
	assert.Contains(t, progressStages, StageCleanup)
	assert.Contains(t, progressStages, StageDone)

	// Verify the "Build completed" message appears instead of "Installation completed"
	foundBuildCompleted := false
	for _, detail := range progressDetails {
		if detail == "Build completed" {
			foundBuildCompleted = true
			break
		}
	}
	assert.True(t, foundBuildCompleted, "Build-only mode should report 'Build completed'")
}
