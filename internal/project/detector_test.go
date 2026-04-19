// ABOUTME: Tests for Perl project root detection in the project package
// ABOUTME: Exercises marker-file discovery (cpanfile, dist.ini, Makefile.PL, etc.) and detection cache invalidation

package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectProject(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Clear cache before each test
	ClearDetectionCache()

	tests := []struct {
		name           string
		setupFunc      func(string) string // Returns the directory to test from
		expectedResult func(*testing.T, *ProjectContext)
	}{
		{
			name: "NoProjectMarkers",
			setupFunc: func(root string) string {
				// Create a subdirectory without any project markers
				subDir := filepath.Join(root, "no-project")
				err := os.MkdirAll(subDir, 0755)
				require.NoError(t, err)
				return subDir
			},
			expectedResult: func(t *testing.T, ctx *ProjectContext) {
				assert.False(t, ctx.IsProject)
				assert.Empty(t, ctx.RootDir)
				assert.Empty(t, ctx.PerlVersion)
				assert.False(t, ctx.HasCpanfile)
				assert.Empty(t, ctx.ConfigFile)
			},
		},
		{
			name: "PerlVersionProject",
			setupFunc: func(root string) string {
				projectDir := filepath.Join(root, "perl-version-project")
				err := os.MkdirAll(projectDir, 0755)
				require.NoError(t, err)

				// Create .perl-version file
				perlVersionFile := filepath.Join(projectDir, ".perl-version")
				err = os.WriteFile(perlVersionFile, []byte("5.38.0\n"), 0644)
				require.NoError(t, err)

				// Test from subdirectory
				subDir := filepath.Join(projectDir, "lib", "My")
				err = os.MkdirAll(subDir, 0755)
				require.NoError(t, err)

				return subDir
			},
			expectedResult: func(t *testing.T, ctx *ProjectContext) {
				assert.True(t, ctx.IsProject)
				assert.Contains(t, ctx.RootDir, "perl-version-project")
				assert.Equal(t, "5.38.0", ctx.PerlVersion)
				assert.Equal(t, ".perl-version", ctx.DetectionInfo)
				assert.Contains(t, ctx.LocalLibDir, filepath.Join("perl-version-project", "local"))
			},
		},
		{
			name: "CpanfileProject",
			setupFunc: func(root string) string {
				projectDir := filepath.Join(root, "cpanfile-project")
				err := os.MkdirAll(projectDir, 0755)
				require.NoError(t, err)

				// Create cpanfile
				cpanfile := filepath.Join(projectDir, "cpanfile")
				err = os.WriteFile(cpanfile, []byte("requires 'DBI';\n"), 0644)
				require.NoError(t, err)

				return projectDir
			},
			expectedResult: func(t *testing.T, ctx *ProjectContext) {
				assert.True(t, ctx.IsProject)
				assert.Contains(t, ctx.RootDir, "cpanfile-project")
				assert.True(t, ctx.HasCpanfile)
				assert.Equal(t, "cpanfile", ctx.DetectionInfo)
			},
		},
		{
			name: "PvmTomlProject",
			setupFunc: func(root string) string {
				projectDir := filepath.Join(root, "pvm-toml-project")
				err := os.MkdirAll(projectDir, 0755)
				require.NoError(t, err)

				// Create pvm.toml
				pvmToml := filepath.Join(projectDir, "pvm.toml")
				err = os.WriteFile(pvmToml, []byte("[project]\nname = \"test\"\n"), 0644)
				require.NoError(t, err)

				return projectDir
			},
			expectedResult: func(t *testing.T, ctx *ProjectContext) {
				assert.True(t, ctx.IsProject)
				assert.Contains(t, ctx.RootDir, "pvm-toml-project")
				assert.Contains(t, ctx.ConfigFile, "pvm.toml")
				assert.Equal(t, "pvm.toml", ctx.DetectionInfo)
			},
		},
		{
			name: "GitProject",
			setupFunc: func(root string) string {
				projectDir := filepath.Join(root, "git-project")
				err := os.MkdirAll(projectDir, 0755)
				require.NoError(t, err)

				// Create .git directory
				gitDir := filepath.Join(projectDir, ".git")
				err = os.MkdirAll(gitDir, 0755)
				require.NoError(t, err)

				return projectDir
			},
			expectedResult: func(t *testing.T, ctx *ProjectContext) {
				assert.True(t, ctx.IsProject)
				assert.Contains(t, ctx.RootDir, "git-project")
				assert.Equal(t, ".git", ctx.DetectionInfo)
			},
		},
		{
			name: "ComplexProject",
			setupFunc: func(root string) string {
				projectDir := filepath.Join(root, "complex-project")
				err := os.MkdirAll(projectDir, 0755)
				require.NoError(t, err)

				// Create multiple project markers
				perlVersionFile := filepath.Join(projectDir, ".perl-version")
				err = os.WriteFile(perlVersionFile, []byte("5.40.0"), 0644)
				require.NoError(t, err)

				cpanfile := filepath.Join(projectDir, "cpanfile")
				err = os.WriteFile(cpanfile, []byte("requires 'Moose';\n"), 0644)
				require.NoError(t, err)

				pvmToml := filepath.Join(projectDir, "pvm.toml")
				err = os.WriteFile(pvmToml, []byte("[project]\nname = \"complex\"\n"), 0644)
				require.NoError(t, err)

				// Test from deeply nested directory
				deepDir := filepath.Join(projectDir, "lib", "My", "Module")
				err = os.MkdirAll(deepDir, 0755)
				require.NoError(t, err)

				return deepDir
			},
			expectedResult: func(t *testing.T, ctx *ProjectContext) {
				assert.True(t, ctx.IsProject)
				assert.Contains(t, ctx.RootDir, "complex-project")
				assert.Equal(t, "5.40.0", ctx.PerlVersion)
				assert.True(t, ctx.HasCpanfile)
				assert.Contains(t, ctx.ConfigFile, "pvm.toml")
				// Should detect by .perl-version first (highest priority)
				assert.Equal(t, ".perl-version", ctx.DetectionInfo)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear cache before each subtest
			ClearDetectionCache()

			testDir := tt.setupFunc(tmpDir)
			ctx, err := DetectProject(testDir)
			require.NoError(t, err)
			tt.expectedResult(t, ctx)
		})
	}
}

func TestDetectProjectCaching(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Clear cache
	ClearDetectionCache()

	// Create a project
	projectDir := filepath.Join(tmpDir, "cache-test")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	perlVersionFile := filepath.Join(projectDir, ".perl-version")
	err = os.WriteFile(perlVersionFile, []byte("5.38.0"), 0644)
	require.NoError(t, err)

	// First detection should populate cache
	ctx1, err := DetectProject(projectDir)
	require.NoError(t, err)
	assert.True(t, ctx1.IsProject)

	// Remove the project marker file
	err = os.Remove(perlVersionFile)
	require.NoError(t, err)

	// Second detection should return cached result
	ctx2, err := DetectProject(projectDir)
	require.NoError(t, err)
	assert.True(t, ctx2.IsProject) // Should still be true due to caching

	// Clear cache and detect again
	ClearDetectionCache()
	ctx3, err := DetectProject(projectDir)
	require.NoError(t, err)
	assert.False(t, ctx3.IsProject) // Should now be false as file was removed
}

func TestDetectProjectPriorityOrder(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Clear cache
	ClearDetectionCache()

	// Create project with all markers
	projectDir := filepath.Join(tmpDir, "priority-test")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create markers in reverse priority order
	gitDir := filepath.Join(projectDir, ".git")
	err = os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	pvmToml := filepath.Join(projectDir, "pvm.toml")
	err = os.WriteFile(pvmToml, []byte("[project]\nname = \"test\"\n"), 0644)
	require.NoError(t, err)

	cpanfile := filepath.Join(projectDir, "cpanfile")
	err = os.WriteFile(cpanfile, []byte("requires 'Test::More';\n"), 0644)
	require.NoError(t, err)

	perlVersionFile := filepath.Join(projectDir, ".perl-version")
	err = os.WriteFile(perlVersionFile, []byte("5.38.0"), 0644)
	require.NoError(t, err)

	// Detection should use .perl-version (highest priority)
	ctx, err := DetectProject(projectDir)
	require.NoError(t, err)
	assert.Equal(t, ".perl-version", ctx.DetectionInfo)
	assert.Equal(t, "5.38.0", ctx.PerlVersion)
	assert.True(t, ctx.HasCpanfile)
	assert.Contains(t, ctx.ConfigFile, "pvm.toml")
}

func TestGetCurrentProject(t *testing.T) {
	// This test verifies the function works but doesn't test specific behavior
	// since it depends on the actual working directory
	ctx, err := GetCurrentProject()
	require.NoError(t, err)
	assert.NotNil(t, ctx)
	// The result depends on whether we're in a project or not
}

func TestReadPerlVersion(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"SimpleVersion", "5.38.0", "5.38.0"},
		{"VersionWithNewline", "5.38.0\n", "5.38.0"},
		{"VersionWithWhitespace", "  5.38.0  \n", "5.38.0"},
		{"EmptyFile", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := filepath.Join(tmpDir, "test-perl-version")
			err := os.WriteFile(file, []byte(tt.content), 0644)
			require.NoError(t, err)

			version, err := readPerlVersion(file)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, version)
		})
	}

	// Test non-existent file
	version, err := readPerlVersion(filepath.Join(tmpDir, "nonexistent"))
	assert.Error(t, err)
	assert.Empty(t, version)
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test-file")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Test existing file
	assert.True(t, fileExists(testFile))

	// Test non-existent file
	assert.False(t, fileExists(filepath.Join(tmpDir, "nonexistent")))

	// Test directory
	assert.True(t, fileExists(tmpDir))
}

func TestEnrichProjectContext(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a project directory
	projectDir := filepath.Join(tmpDir, "enrich-test")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create various project files
	perlVersionFile := filepath.Join(projectDir, ".perl-version")
	err = os.WriteFile(perlVersionFile, []byte("5.38.0"), 0644)
	require.NoError(t, err)

	cpanfile := filepath.Join(projectDir, "cpanfile")
	err = os.WriteFile(cpanfile, []byte("requires 'DBI';\n"), 0644)
	require.NoError(t, err)

	pvmToml := filepath.Join(projectDir, "pvm.toml")
	err = os.WriteFile(pvmToml, []byte("[project]\nname = \"test\"\n"), 0644)
	require.NoError(t, err)

	// Test enriching context that was detected by git (lowest priority)
	ctx := &ProjectContext{
		IsProject:     true,
		RootDir:       projectDir,
		DetectionInfo: ".git",
	}

	enrichProjectContext(ctx)

	assert.Equal(t, "5.38.0", ctx.PerlVersion)
	assert.True(t, ctx.HasCpanfile)
	assert.Contains(t, ctx.ConfigFile, "pvm.toml")
}

func TestDetectProjectErrorHandling(t *testing.T) {
	// Test with invalid directory
	ctx, err := DetectProject("/nonexistent/directory/path")
	assert.Error(t, err)
	assert.Nil(t, ctx)
}

func BenchmarkDetectProject(b *testing.B) {
	tmpDir := b.TempDir()

	// Create a project structure
	projectDir := filepath.Join(tmpDir, "benchmark-project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(b, err)

	perlVersionFile := filepath.Join(projectDir, ".perl-version")
	err = os.WriteFile(perlVersionFile, []byte("5.38.0"), 0644)
	require.NoError(b, err)

	// Create a deep directory structure
	deepDir := filepath.Join(projectDir, "lib", "My", "Deep", "Module", "Structure")
	err = os.MkdirAll(deepDir, 0755)
	require.NoError(b, err)

	// Clear cache
	ClearDetectionCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DetectProject(deepDir)
		require.NoError(b, err)
	}
}

func BenchmarkDetectProjectCached(b *testing.B) {
	tmpDir := b.TempDir()

	// Create a project structure
	projectDir := filepath.Join(tmpDir, "benchmark-project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(b, err)

	perlVersionFile := filepath.Join(projectDir, ".perl-version")
	err = os.WriteFile(perlVersionFile, []byte("5.38.0"), 0644)
	require.NoError(b, err)

	// Create a deep directory structure
	deepDir := filepath.Join(projectDir, "lib", "My", "Deep", "Module", "Structure")
	err = os.MkdirAll(deepDir, 0755)
	require.NoError(b, err)

	// Clear cache and do one detection to populate cache
	ClearDetectionCache()
	_, err = DetectProject(deepDir)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DetectProject(deepDir)
		require.NoError(b, err)
	}
}

func TestDetectProjectInDirectory(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Clear cache before each test
	ClearDetectionCache()

	tests := []struct {
		name           string
		setupFunc      func(string) (string, string) // Returns (testDir, parentDir)
		expectedResult func(*testing.T, *ProjectContext)
	}{
		{
			name: "NoProjectMarkersInDirectory",
			setupFunc: func(root string) (string, string) {
				// Create parent directory with project marker
				parentDir := filepath.Join(root, "parent-project")
				err := os.MkdirAll(parentDir, 0755)
				require.NoError(t, err)

				// Create .perl-version in parent
				perlVersionFile := filepath.Join(parentDir, ".perl-version")
				err = os.WriteFile(perlVersionFile, []byte("5.38.0\n"), 0644)
				require.NoError(t, err)

				// Create subdirectory without markers
				subDir := filepath.Join(parentDir, "empty-subdir")
				err = os.MkdirAll(subDir, 0755)
				require.NoError(t, err)

				return subDir, parentDir
			},
			expectedResult: func(t *testing.T, ctx *ProjectContext) {
				// Should NOT find project since we only check the specific directory
				assert.False(t, ctx.IsProject)
				assert.Empty(t, ctx.RootDir)
				assert.Empty(t, ctx.PerlVersion)
				assert.False(t, ctx.HasCpanfile)
				assert.Empty(t, ctx.ConfigFile)
				assert.Empty(t, ctx.DetectionInfo)
			},
		},
		{
			name: "PerlVersionInTargetDirectory",
			setupFunc: func(root string) (string, string) {
				// Create parent without markers
				parentDir := filepath.Join(root, "parent-no-project")
				err := os.MkdirAll(parentDir, 0755)
				require.NoError(t, err)

				// Create target directory with .perl-version
				targetDir := filepath.Join(parentDir, "target-project")
				err = os.MkdirAll(targetDir, 0755)
				require.NoError(t, err)

				perlVersionFile := filepath.Join(targetDir, ".perl-version")
				err = os.WriteFile(perlVersionFile, []byte("5.40.0\n"), 0644)
				require.NoError(t, err)

				return targetDir, parentDir
			},
			expectedResult: func(t *testing.T, ctx *ProjectContext) {
				assert.True(t, ctx.IsProject)
				assert.Contains(t, ctx.RootDir, "target-project")
				assert.Equal(t, "5.40.0", ctx.PerlVersion)
				assert.Equal(t, ".perl-version", ctx.DetectionInfo)
				assert.Contains(t, ctx.LocalLibDir, filepath.Join("target-project", "local"))
			},
		},
		{
			name: "CpanfileInTargetDirectory",
			setupFunc: func(root string) (string, string) {
				parentDir := filepath.Join(root, "parent")
				targetDir := filepath.Join(parentDir, "cpanfile-target")
				err := os.MkdirAll(targetDir, 0755)
				require.NoError(t, err)

				// Create cpanfile in target
				cpanfile := filepath.Join(targetDir, "cpanfile")
				err = os.WriteFile(cpanfile, []byte("requires 'Moose';\n"), 0644)
				require.NoError(t, err)

				return targetDir, parentDir
			},
			expectedResult: func(t *testing.T, ctx *ProjectContext) {
				assert.True(t, ctx.IsProject)
				assert.Contains(t, ctx.RootDir, "cpanfile-target")
				assert.True(t, ctx.HasCpanfile)
				assert.Equal(t, "cpanfile", ctx.DetectionInfo)
			},
		},
		{
			name: "PvmTomlInTargetDirectory",
			setupFunc: func(root string) (string, string) {
				parentDir := filepath.Join(root, "parent")
				targetDir := filepath.Join(parentDir, "pvm-target")
				err := os.MkdirAll(targetDir, 0755)
				require.NoError(t, err)

				// Create pvm.toml in target
				pvmToml := filepath.Join(targetDir, "pvm.toml")
				err = os.WriteFile(pvmToml, []byte("[project]\nname = \"test\"\n"), 0644)
				require.NoError(t, err)

				return targetDir, parentDir
			},
			expectedResult: func(t *testing.T, ctx *ProjectContext) {
				assert.True(t, ctx.IsProject)
				assert.Contains(t, ctx.RootDir, "pvm-target")
				assert.Contains(t, ctx.ConfigFile, "pvm.toml")
				assert.Equal(t, "pvm.toml", ctx.DetectionInfo)
			},
		},
		{
			name: "GitInTargetDirectory",
			setupFunc: func(root string) (string, string) {
				parentDir := filepath.Join(root, "parent")
				targetDir := filepath.Join(parentDir, "git-target")
				err := os.MkdirAll(targetDir, 0755)
				require.NoError(t, err)

				// Create .git directory in target
				gitDir := filepath.Join(targetDir, ".git")
				err = os.MkdirAll(gitDir, 0755)
				require.NoError(t, err)

				return targetDir, parentDir
			},
			expectedResult: func(t *testing.T, ctx *ProjectContext) {
				assert.True(t, ctx.IsProject)
				assert.Contains(t, ctx.RootDir, "git-target")
				assert.Equal(t, ".git", ctx.DetectionInfo)
			},
		},
		{
			name: "PriorityOrderInTargetDirectory",
			setupFunc: func(root string) (string, string) {
				parentDir := filepath.Join(root, "parent")
				targetDir := filepath.Join(parentDir, "priority-target")
				err := os.MkdirAll(targetDir, 0755)
				require.NoError(t, err)

				// Create multiple markers in target directory (.perl-version should have priority)
				perlVersionFile := filepath.Join(targetDir, ".perl-version")
				err = os.WriteFile(perlVersionFile, []byte("5.38.0\n"), 0644)
				require.NoError(t, err)

				cpanfile := filepath.Join(targetDir, "cpanfile")
				err = os.WriteFile(cpanfile, []byte("requires 'DBI';\n"), 0644)
				require.NoError(t, err)

				gitDir := filepath.Join(targetDir, ".git")
				err = os.MkdirAll(gitDir, 0755)
				require.NoError(t, err)

				return targetDir, parentDir
			},
			expectedResult: func(t *testing.T, ctx *ProjectContext) {
				assert.True(t, ctx.IsProject)
				assert.Contains(t, ctx.RootDir, "priority-target")
				assert.Equal(t, ".perl-version", ctx.DetectionInfo) // Should detect .perl-version first
				assert.Equal(t, "5.38.0", ctx.PerlVersion)
				assert.True(t, ctx.HasCpanfile) // Should enrich with other project info
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear cache for each test
			ClearDetectionCache()

			testDir, parentDir := tt.setupFunc(tmpDir)

			// Verify parent setup if needed
			_ = parentDir // Suppress unused variable warning

			// Test DetectProjectInDirectory
			ctx, err := DetectProjectInDirectory(testDir)
			require.NoError(t, err)
			require.NotNil(t, ctx)

			tt.expectedResult(t, ctx)
		})
	}
}

func TestDetectProjectInDirectoryVsDetectProject(t *testing.T) {
	// This test specifically validates the key difference between the two functions
	tmpDir := t.TempDir()
	ClearDetectionCache()

	// Setup: Create parent directory with .perl-version
	parentDir := filepath.Join(tmpDir, "parent-project")
	err := os.MkdirAll(parentDir, 0755)
	require.NoError(t, err)

	perlVersionFile := filepath.Join(parentDir, ".perl-version")
	err = os.WriteFile(perlVersionFile, []byte("5.38.0\n"), 0644)
	require.NoError(t, err)

	// Create empty subdirectory
	emptySubdir := filepath.Join(parentDir, "empty-subdir")
	err = os.MkdirAll(emptySubdir, 0755)
	require.NoError(t, err)

	// Test DetectProject - should find parent's .perl-version
	treeWalkingResult, err := DetectProject(emptySubdir)
	require.NoError(t, err)
	assert.True(t, treeWalkingResult.IsProject, "DetectProject should find parent directory markers")
	assert.Contains(t, treeWalkingResult.RootDir, "parent-project")
	assert.Equal(t, ".perl-version", treeWalkingResult.DetectionInfo)

	// Test DetectProjectInDirectory - should NOT find parent's .perl-version
	directoryOnlyResult, err := DetectProjectInDirectory(emptySubdir)
	require.NoError(t, err)
	assert.False(t, directoryOnlyResult.IsProject, "DetectProjectInDirectory should NOT find parent directory markers")
	assert.Empty(t, directoryOnlyResult.RootDir)
	assert.Empty(t, directoryOnlyResult.DetectionInfo)
}

func TestDetectProjectInDirectoryCaching(t *testing.T) {
	tmpDir := t.TempDir()
	ClearDetectionCache()

	// Create directory with .perl-version
	projectDir := filepath.Join(tmpDir, "cached-project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	perlVersionFile := filepath.Join(projectDir, ".perl-version")
	err = os.WriteFile(perlVersionFile, []byte("5.38.0\n"), 0644)
	require.NoError(t, err)

	// First call - should hit filesystem
	ctx1, err := DetectProjectInDirectory(projectDir)
	require.NoError(t, err)
	assert.True(t, ctx1.IsProject)

	// Second call - should hit cache
	ctx2, err := DetectProjectInDirectory(projectDir)
	require.NoError(t, err)
	assert.True(t, ctx2.IsProject)
	assert.Equal(t, ctx1.PerlVersion, ctx2.PerlVersion)

	// Verify cache separation from regular DetectProject
	ctx3, err := DetectProject(projectDir)
	require.NoError(t, err)
	assert.True(t, ctx3.IsProject)
	// Both should find the project but cache keys should be different
}
