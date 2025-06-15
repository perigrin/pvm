package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProjectAwareConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	ClearDetectionCache()

	tests := []struct {
		name         string
		setupFunc    func(string) string
		expectedPath func(string) string // function to generate expected path
		expectEmpty  bool
	}{
		{
			name: "NoProject",
			setupFunc: func(root string) string {
				subDir := filepath.Join(root, "no-project")
				err := os.MkdirAll(subDir, 0755)
				require.NoError(t, err)
				return subDir
			},
			expectEmpty: true,
		},
		{
			name: "ProjectWithPvmToml",
			setupFunc: func(root string) string {
				projectDir := filepath.Join(root, "pvm-project")
				err := os.MkdirAll(projectDir, 0755)
				require.NoError(t, err)

				pvmToml := filepath.Join(projectDir, "pvm.toml")
				err = os.WriteFile(pvmToml, []byte("[project]\nname = \"test\"\n"), 0644)
				require.NoError(t, err)

				return projectDir
			},
			expectedPath: func(root string) string {
				return filepath.Join(root, "pvm-project", "pvm.toml")
			},
		},
		{
			name: "ProjectWithPerlVersion",
			setupFunc: func(root string) string {
				projectDir := filepath.Join(root, "perl-project")
				err := os.MkdirAll(projectDir, 0755)
				require.NoError(t, err)

				perlVersionFile := filepath.Join(projectDir, ".perl-version")
				err = os.WriteFile(perlVersionFile, []byte("5.38.0"), 0644)
				require.NoError(t, err)

				return projectDir
			},
			expectedPath: func(root string) string {
				return filepath.Join(root, "perl-project", ".pvm", "pvm.toml")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ClearDetectionCache()
			testDir := tt.setupFunc(tmpDir)

			configPath, err := GetProjectAwareConfigPath(testDir)
			require.NoError(t, err)

			if tt.expectEmpty {
				assert.Empty(t, configPath)
			} else {
				expectedPath := tt.expectedPath(tmpDir)
				assert.Equal(t, expectedPath, configPath)
			}
		})
	}
}

func TestGetProjectAwareLibPath(t *testing.T) {
	tmpDir := t.TempDir()
	ClearDetectionCache()

	tests := []struct {
		name         string
		setupFunc    func(string) string
		expectedPath func(string) string
		expectEmpty  bool
	}{
		{
			name: "NoProject",
			setupFunc: func(root string) string {
				subDir := filepath.Join(root, "no-project")
				err := os.MkdirAll(subDir, 0755)
				require.NoError(t, err)
				return subDir
			},
			expectEmpty: true,
		},
		{
			name: "ProjectWithCpanfile",
			setupFunc: func(root string) string {
				projectDir := filepath.Join(root, "cpan-project")
				err := os.MkdirAll(projectDir, 0755)
				require.NoError(t, err)

				cpanfile := filepath.Join(projectDir, "cpanfile")
				err = os.WriteFile(cpanfile, []byte("requires 'DBI';\n"), 0644)
				require.NoError(t, err)

				subDir := filepath.Join(projectDir, "t")
				err = os.MkdirAll(subDir, 0755)
				require.NoError(t, err)

				return subDir
			},
			expectedPath: func(root string) string {
				return filepath.Join(root, "cpan-project", "lib")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ClearDetectionCache()
			testDir := tt.setupFunc(tmpDir)

			libPath, err := GetProjectAwareLibPath(testDir)
			require.NoError(t, err)

			if tt.expectEmpty {
				assert.Empty(t, libPath)
			} else {
				expectedPath := tt.expectedPath(tmpDir)
				assert.Equal(t, expectedPath, libPath)
			}
		})
	}
}

func TestIsInProject(t *testing.T) {
	tmpDir := t.TempDir()
	ClearDetectionCache()

	// Create project
	projectDir := filepath.Join(tmpDir, "test-project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	perlVersionFile := filepath.Join(projectDir, ".perl-version")
	err = os.WriteFile(perlVersionFile, []byte("5.38.0"), 0644)
	require.NoError(t, err)

	// Create non-project directory
	nonProjectDir := filepath.Join(tmpDir, "not-project")
	err = os.MkdirAll(nonProjectDir, 0755)
	require.NoError(t, err)

	// Test project directory
	inProject, err := IsInProject(projectDir)
	require.NoError(t, err)
	assert.True(t, inProject)

	// Test non-project directory
	inProject, err = IsInProject(nonProjectDir)
	require.NoError(t, err)
	assert.False(t, inProject)

	// Test subdirectory of project
	subDir := filepath.Join(projectDir, "lib", "My")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	inProject, err = IsInProject(subDir)
	require.NoError(t, err)
	assert.True(t, inProject)
}

func TestGetProjectRoot(t *testing.T) {
	tmpDir := t.TempDir()
	ClearDetectionCache()

	// Create project
	projectDir := filepath.Join(tmpDir, "root-test")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	gitDir := filepath.Join(projectDir, ".git")
	err = os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	// Create deep subdirectory
	deepDir := filepath.Join(projectDir, "lib", "My", "Deep", "Module")
	err = os.MkdirAll(deepDir, 0755)
	require.NoError(t, err)

	// Test getting root from deep directory
	root, err := GetProjectRoot(deepDir)
	require.NoError(t, err)
	assert.Equal(t, projectDir, root)

	// Test non-project directory
	nonProjectDir := filepath.Join(tmpDir, "not-project")
	err = os.MkdirAll(nonProjectDir, 0755)
	require.NoError(t, err)

	root, err = GetProjectRoot(nonProjectDir)
	require.NoError(t, err)
	assert.Empty(t, root)
}

func TestGetProjectPerlVersion(t *testing.T) {
	tmpDir := t.TempDir()
	ClearDetectionCache()

	tests := []struct {
		name            string
		setupFunc       func(string) string
		expectedVersion string
	}{
		{
			name: "ProjectWithPerlVersion",
			setupFunc: func(root string) string {
				projectDir := filepath.Join(root, "version-project")
				err := os.MkdirAll(projectDir, 0755)
				require.NoError(t, err)

				perlVersionFile := filepath.Join(projectDir, ".perl-version")
				err = os.WriteFile(perlVersionFile, []byte("5.40.0"), 0644)
				require.NoError(t, err)

				return projectDir
			},
			expectedVersion: "5.40.0",
		},
		{
			name: "ProjectWithoutPerlVersion",
			setupFunc: func(root string) string {
				projectDir := filepath.Join(root, "no-version-project")
				err := os.MkdirAll(projectDir, 0755)
				require.NoError(t, err)

				cpanfile := filepath.Join(projectDir, "cpanfile")
				err = os.WriteFile(cpanfile, []byte("requires 'strict';\n"), 0644)
				require.NoError(t, err)

				return projectDir
			},
			expectedVersion: "",
		},
		{
			name: "NoProject",
			setupFunc: func(root string) string {
				noProjectDir := filepath.Join(root, "no-project")
				err := os.MkdirAll(noProjectDir, 0755)
				require.NoError(t, err)
				return noProjectDir
			},
			expectedVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ClearDetectionCache()
			testDir := tt.setupFunc(tmpDir)

			version, err := GetProjectPerlVersion(testDir)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedVersion, version)
		})
	}
}

func TestConfigIntegrationErrorHandling(t *testing.T) {
	// Test with invalid directory
	_, err := GetProjectAwareConfigPath("/nonexistent/directory")
	assert.Error(t, err)

	_, err = GetProjectAwareLibPath("/nonexistent/directory")
	assert.Error(t, err)

	_, err = IsInProject("/nonexistent/directory")
	assert.Error(t, err)

	_, err = GetProjectRoot("/nonexistent/directory")
	assert.Error(t, err)

	_, err = GetProjectPerlVersion("/nonexistent/directory")
	assert.Error(t, err)
}
