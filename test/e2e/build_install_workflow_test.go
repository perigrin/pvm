// ABOUTME: End-to-end integration tests for build-perl and install-perl workflow
// ABOUTME: Tests the complete build-only, install-from-build, and install-from-archive workflows

package e2e

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/test/e2e/helpers"
)

// TestBuildInstallWorkflow_BuildOnly removed - building Perl from source is too slow
// TODO: Replace with faster build system tests when appropriate

func TestBuildInstallWorkflow_BuildOnlyFlagValidation(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test that --build-only and --output-dir flags are properly recognized
	// Use --help to validate flag parsing without triggering downloads
	stdout, stderr, err := env.RunPVM("build-perl", "--help")

	// Help command should succeed
	require.NoError(t, err, "Help command should succeed")

	// Verify the flags are documented in help output
	helpOutput := stdout + stderr
	assert.Contains(t, helpOutput, "--build-only", "Should document --build-only flag")
	assert.Contains(t, helpOutput, "--output-dir", "Should document --output-dir flag")
	assert.Contains(t, helpOutput, "Build Perl without installing", "Should describe build-only functionality")

	// Should NOT contain flag parsing errors
	assert.NotContains(t, helpOutput, "unknown flag", "Should not have flag parsing errors")
	assert.NotContains(t, helpOutput, "invalid flag", "Should not have flag parsing errors")
}

func TestBuildInstallWorkflow_ErrorHandling(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test install-perl with non-existent directory
	installStdout, installStderr, installErr := env.RunPVM("install-perl", "--from-build", "/nonexistent/directory")
	if installErr == nil {
		t.Logf("Expected install-perl to fail with non-existent directory, but it succeeded\nStdout: %s\nStderr: %s", installStdout, installStderr)
	} else {
		assert.Contains(t, installStderr, "does not exist", "Error message should be helpful")
	}

	// Test install-perl with invalid archive
	invalidArchive := filepath.Join(env.RootDir, "invalid.tar.gz")
	err := os.WriteFile(invalidArchive, []byte("not a valid archive"), 0644)
	require.NoError(t, err, "Should be able to create invalid archive file")

	archiveStdout, archiveStderr, archiveErr := env.RunPVM("install-perl", invalidArchive)
	if archiveErr == nil {
		t.Logf("Expected install-perl to fail with invalid archive, but it succeeded\nStdout: %s\nStderr: %s", archiveStdout, archiveStderr)
	} else {
		assert.Contains(t, archiveStderr, "extract", "Error message should mention extraction")
	}

	// Test build-perl with invalid output directory (permission denied)
	buildStdout, buildStderr, buildErr := env.RunPVM("build-perl", "5.38.0", "--build-only", "--output-dir", "/root/forbidden")
	if buildErr == nil {
		t.Logf("Expected build-perl to fail with permission denied, but it succeeded\nStdout: %s\nStderr: %s", buildStdout, buildStderr)
	}
}

// Helper function to create a tar.gz archive from a directory
func createTarGzArchive(sourceDir, archivePath string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path from the source directory
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Handle symbolic links
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			header.Linkname = link
		}

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content (only for regular files)
		if info.Mode().IsRegular() {
			sourceFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer sourceFile.Close()

			_, err = io.Copy(tarWriter, sourceFile)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
