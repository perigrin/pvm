// ABOUTME: Tests for upload-binary command functionality
// ABOUTME: Validates command flag parsing and archive creation logic

package pvm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUploadBinaryCommand(t *testing.T) {
	cmd := newUploadBinaryCommand()

	assert.Equal(t, "upload-binary [directory]", cmd.Use)
	assert.Equal(t, "Upload binary archives to GitHub releases and custom mirrors", cmd.Short)
	assert.NotNil(t, cmd.RunE)

	// Check that flags are defined
	flags := []string{
		"version", "platform", "mirror", "archive",
		"github-token", "github-repo", "release-tag", "draft-release", "prerelease",
		"create-archive", "output-archive", "compression",
		"verify-upload", "force", "max-retries", "timeout",
	}

	for _, flag := range flags {
		assert.NotNil(t, cmd.Flags().Lookup(flag), "Flag %s should be defined", flag)
	}
}

func TestDetectVersionFromArchiveName(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"perl-5.38.0-linux-amd64.tar.gz", "5.38.0"},
		{"perl-5.40.0-darwin-arm64.tgz", "5.40.0"},
		{"perl-5.36.2.tar.gz", "5.36.2"},
		{"some-other-file.tar.gz", ""},
		{"perl-5.38.0.tar.gz", "5.38.0"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := detectVersionFromArchiveName(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectVersionFromDirectory(t *testing.T) {
	t.Run("FromDirectoryName", func(t *testing.T) {
		// Create a temporary directory with version in name
		tempDir := t.TempDir()
		versionDir := filepath.Join(tempDir, "perl-5.38.0-build")
		err := os.MkdirAll(versionDir, 0755)
		require.NoError(t, err)

		version, err := detectVersionFromDirectory(versionDir)
		require.NoError(t, err)
		assert.Equal(t, "5.38.0", version)
	})

	t.Run("NoVersionFound", func(t *testing.T) {
		tempDir := t.TempDir()
		_, err := detectVersionFromDirectory(tempDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not detect version")
	})
}

func TestCreateTarGzArchive(t *testing.T) {
	// Create a temporary directory with some files
	sourceDir := t.TempDir()

	// Create test files
	testFile := filepath.Join(sourceDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	subDir := filepath.Join(sourceDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	subFile := filepath.Join(subDir, "sub.txt")
	err = os.WriteFile(subFile, []byte("sub content"), 0644)
	require.NoError(t, err)

	// Create archive
	outputPath := filepath.Join(t.TempDir(), "test.tar.gz")
	err = createTarGzArchive(sourceDir, outputPath)
	require.NoError(t, err)

	// Verify archive was created
	_, err = os.Stat(outputPath)
	assert.NoError(t, err)

	// Verify archive is not empty
	info, err := os.Stat(outputPath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestUploadBinaryCommand_FlagValidation(t *testing.T) {
	t.Run("BothArchiveAndCreateArchive", func(t *testing.T) {
		cmd := newUploadBinaryCommand()
		cmd.SetArgs([]string{"test-dir", "--archive", "test.tar.gz", "--create-archive"})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot specify both --archive and --create-archive")
	})

	t.Run("NeitherArchiveNorCreateArchive", func(t *testing.T) {
		cmd := newUploadBinaryCommand()
		cmd.SetArgs([]string{"test-dir", "--create-archive=false", "--archive="})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must specify either --archive or --create-archive")
	})
}

func TestUploadBinaryCommand_GitHubValidation(t *testing.T) {
	cmd := newUploadBinaryCommand()
	tempDir := t.TempDir()

	// Create a test file to avoid source directory validation error
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Provide version to avoid version detection error
	cmd.SetArgs([]string{tempDir, "--version", "5.38.0", "--github-repo", "owner/repo"})
	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub token required")
}

// Mock command for testing command flags without execution
func createMockUploadCommand() *cobra.Command {
	cmd := newUploadBinaryCommand()
	// Replace RunE with a mock that just validates flags
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Just return nil for flag validation tests
		return nil
	}
	return cmd
}

func TestUploadBinaryCommand_FlagParsing(t *testing.T) {
	cmd := createMockUploadCommand()

	// Test flag parsing
	cmd.SetArgs([]string{
		"test-dir",
		"--version", "5.38.0",
		"--platform", "linux-amd64",
		"--github-repo", "owner/repo",
		"--github-token", "ghp_test",
		"--release-tag", "v5.38.0",
		"--draft-release",
		"--max-retries", "5",
		"--timeout", "15m",
	})

	err := cmd.Execute()
	assert.NoError(t, err)

	// Verify flags were parsed correctly
	version, _ := cmd.Flags().GetString("version")
	assert.Equal(t, "5.38.0", version)

	platform, _ := cmd.Flags().GetString("platform")
	assert.Equal(t, "linux-amd64", platform)

	githubRepo, _ := cmd.Flags().GetString("github-repo")
	assert.Equal(t, "owner/repo", githubRepo)

	draftRelease, _ := cmd.Flags().GetBool("draft-release")
	assert.True(t, draftRelease)

	maxRetries, _ := cmd.Flags().GetInt("max-retries")
	assert.Equal(t, 5, maxRetries)

	timeout, _ := cmd.Flags().GetString("timeout")
	assert.Equal(t, "15m", timeout)
}
