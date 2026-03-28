// ABOUTME: Tests for BuildOptions.SourceDir — local directory source skipping download/extract
// ABOUTME: Verifies that when SourceDir is set, download and extract are not invoked

package perl

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildPerlSourceDirSkipsDownload verifies that when SourceDir is set,
// DownloadPerlSourceFunc is never called.
func TestBuildPerlSourceDirSkipsDownload(t *testing.T) {
	dirs, _ := setupTestDirs(t)

	version := "5.36.0"

	// Create a real source directory with a mock Configure script
	sourceDir := filepath.Join(t.TempDir(), "perl-src")
	createMockSourceDir(t, sourceDir)

	// Track whether download was called
	downloadCalled := false
	DownloadPerlSourceFunc = func(options *DownloadOptions) (*DownloadResult, error) {
		downloadCalled = true
		return nil, nil
	}

	// Track whether extract was called
	extractCalled := false
	extractArchiveFunc = func(archivePath, destDir string, ctx context.Context) (string, error) {
		extractCalled = true
		return "", nil
	}

	// Mock command execution to succeed silently
	runCommandWithProgressFunc = func(
		dir string,
		command string,
		args []string,
		ctx context.Context,
		progressCb func(line string),
	) error {
		return nil
	}

	options := &BuildOptions{
		Version:    version,
		SourceDir:  sourceDir,
		InstallDir: filepath.Join(dirs.VersionsDir, version),
		Context:    context.Background(),
	}

	result, err := BuildPerl(options)

	require.NoError(t, err)
	assert.Equal(t, version, result.Version)
	assert.False(t, downloadCalled, "SourceDir set: download must not be called")
	assert.False(t, extractCalled, "SourceDir set: extract must not be called")
}

// TestBuildPerlSourceDirUsedAsSrcDir verifies that Configure is run inside SourceDir.
func TestBuildPerlSourceDirUsedAsSrcDir(t *testing.T) {
	dirs, _ := setupTestDirs(t)

	version := "5.36.0"

	sourceDir := filepath.Join(t.TempDir(), "perl-fork-src")
	createMockSourceDir(t, sourceDir)

	var configureDir string
	runCommandWithProgressFunc = func(
		dir string,
		command string,
		args []string,
		ctx context.Context,
		progressCb func(line string),
	) error {
		if command == "./Configure" {
			configureDir = dir
		}
		return nil
	}

	options := &BuildOptions{
		Version:    version,
		SourceDir:  sourceDir,
		InstallDir: filepath.Join(dirs.VersionsDir, version),
		Context:    context.Background(),
	}

	result, err := BuildPerl(options)

	require.NoError(t, err)
	assert.Equal(t, version, result.Version)
	assert.Equal(t, sourceDir, configureDir,
		"Configure must run inside SourceDir when SourceDir is set")
}

// TestBuildPerlEmptySourceDirPreservesExistingBehavior verifies that leaving
// SourceDir empty still triggers the normal download/extract path.
func TestBuildPerlEmptySourceDirPreservesExistingBehavior(t *testing.T) {
	dirs, tempDir := setupTestDirs(t)

	version := "5.36.0"
	archivePath := createMockArchive(t, dirs, version)

	extractCalled := false
	extractArchiveFunc = mockExtraction(t, tempDir)
	_ = archivePath

	// Wrap the extraction mock to also set our flag
	realMock := extractArchiveFunc
	extractArchiveFunc = func(archivePath, destDir string, ctx context.Context) (string, error) {
		extractCalled = true
		return realMock(archivePath, destDir, ctx)
	}

	runCommandWithProgressFunc = func(
		dir string,
		command string,
		args []string,
		ctx context.Context,
		progressCb func(line string),
	) error {
		return nil
	}

	options := &BuildOptions{
		Version:    version,
		SourceFile: archivePath,
		InstallDir: filepath.Join(dirs.VersionsDir, version),
		Context:    context.Background(),
		// SourceDir intentionally left empty
	}

	result, err := BuildPerl(options)

	require.NoError(t, err)
	assert.Equal(t, version, result.Version)
	assert.True(t, extractCalled, "Empty SourceDir: extract must still be called")
}

// createMockSourceDir creates a minimal source directory with a Configure script,
// as would be produced by a git clone of a Perl fork.
func createMockSourceDir(t *testing.T, dir string) {
	t.Helper()
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "Configure"), []byte("#!/bin/sh\necho 'Mock Configure'"), 0755)
	require.NoError(t, err)
}
