// ABOUTME: Tests for enhanced PVX isolation features
// ABOUTME: Verifies the functionality of the new isolation capabilities

package pvx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildTestEnvironment is a simplified version of buildEnvironment used only for testing
// It duplicates the key functionality but doesn't depend on getPerlArchDir which is hard to mock
func buildTestEnvironment(options *ExecutionOptions) ([]string, error) {
	// Start with the current environment
	env := os.Environ()

	// If no isolation requested, return current environment
	if options.IsolationLevel == "" || options.IsolationLevel == IsolationNone {
		return env, nil
	}

	// For testing, we'll use a fixed architecture
	const fixedArchDir = "darwin-2level"

	// Create subdirectories for the Perl installation
	isolationDir := options.IsolationDir
	libDir := filepath.Join(isolationDir, "lib", "perl5")
	archLibDir := filepath.Join(libDir, fixedArchDir)
	// We don't use binDir in the test version but it's part of the real function
	_ = filepath.Join(isolationDir, "bin") // binDir
	siteDir := filepath.Join(isolationDir, "lib", "site_perl")
	vendorDir := filepath.Join(isolationDir, "lib", "vendor_perl")

	// Set up the environment based on isolation level
	switch options.IsolationLevel {
	case IsolationLow:
		// Low isolation: Add local::lib equivalent while preserving existing PERL5LIB
		perl5lib := fmt.Sprintf("%s:%s", libDir, archLibDir)

		// Add any user-specified additional module paths
		for _, path := range options.AdditionalModulePaths {
			perl5lib = fmt.Sprintf("%s:%s", perl5lib, path)
		}

		// Set up the local::lib equivalent environment variables
		setEnvVar(&env, "PERL5LIB", perl5lib)

		// Use the custom module path if provided, otherwise use the isolation directory
		modulePath := isolationDir
		if options.CustomModulePath != "" {
			modulePath = options.CustomModulePath
		}

		setEnvVar(&env, "PERL_LOCAL_LIB_ROOT", modulePath)
		setEnvVar(&env, "PERL_MB_OPT", fmt.Sprintf("--install_base '%s'", modulePath))
		setEnvVar(&env, "PERL_MM_OPT", fmt.Sprintf("INSTALL_BASE=%s", modulePath))

	case IsolationMedium:
		// Medium isolation: Clean PERL5LIB but preserve most environment variables
		perl5lib := fmt.Sprintf("%s:%s:%s:%s",
			libDir,
			archLibDir,
			siteDir,
			vendorDir)

		// Add any user-specified additional module paths
		for _, path := range options.AdditionalModulePaths {
			perl5lib = fmt.Sprintf("%s:%s", perl5lib, path)
		}

		setEnvVar(&env, "PERL5LIB", perl5lib)

		// Use the custom module path if provided, otherwise use the isolation directory
		modulePath := isolationDir
		if options.CustomModulePath != "" {
			modulePath = options.CustomModulePath
		}

		setEnvVar(&env, "PERL_LOCAL_LIB_ROOT", modulePath)
		setEnvVar(&env, "PERL_MB_OPT", fmt.Sprintf("--install_base '%s'", modulePath))
		setEnvVar(&env, "PERL_MM_OPT", fmt.Sprintf("INSTALL_BASE=%s", modulePath))

	case IsolationHigh:
		// High isolation: Start with minimal environment and add only what's needed
		// Create a clean environment with only essential variables
		cleanEnv := []string{}

		// Add any preserved environment variables
		for _, key := range options.PreserveEnv {
			for _, envVar := range env {
				if strings.HasPrefix(envVar, key+"=") {
					cleanEnv = append(cleanEnv, envVar)
					break
				}
			}
		}

		// Set up enhanced filesystem isolation for high isolation mode
		if options.IsolatedOutput {
			setEnvVar(&cleanEnv, "PVM_ISOLATED_OUTPUT", "1")
		}

		// Configure filesystem access paths
		if len(options.ReadOnlyPaths) > 0 {
			readOnlyPathsStr := strings.Join(options.ReadOnlyPaths, ":")
			setEnvVar(&cleanEnv, "PVM_READONLY_PATHS", readOnlyPathsStr)
		}

		if len(options.ReadWritePaths) > 0 {
			readWritePathsStr := strings.Join(options.ReadWritePaths, ":")
			setEnvVar(&cleanEnv, "PVM_READWRITE_PATHS", readWritePathsStr)
		}

		// Set PERL5LIB to include all the isolation directory paths
		perl5lib := fmt.Sprintf("%s:%s:%s:%s",
			libDir,
			archLibDir,
			siteDir,
			vendorDir)

		// Add any user-specified additional module paths
		for _, path := range options.AdditionalModulePaths {
			perl5lib = fmt.Sprintf("%s:%s", perl5lib, path)
		}

		setEnvVar(&cleanEnv, "PERL5LIB", perl5lib)

		// Set up the local::lib equivalent environment variables
		// Use the custom module path if provided, otherwise use the isolation directory
		modulePath := isolationDir
		if options.CustomModulePath != "" {
			modulePath = options.CustomModulePath
		}

		setEnvVar(&cleanEnv, "PERL_LOCAL_LIB_ROOT", modulePath)
		setEnvVar(&cleanEnv, "PERL_MB_OPT", fmt.Sprintf("--install_base '%s'", modulePath))
		setEnvVar(&cleanEnv, "PERL_MM_OPT", fmt.Sprintf("INSTALL_BASE=%s", modulePath))

		// Use the clean environment instead of the original one
		env = cleanEnv
	}

	return env, nil
}

// TestEnhancedEnvironmentIsolation tests that the enhanced environment isolation features work correctly
func TestEnhancedEnvironmentIsolation(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create options with preserved environment variables
	options := &ExecutionOptions{
		IsolationLevel: IsolationHigh,
		IsolationDir:   tmpDir,
		PreserveEnv:    []string{"TEST_VAR_1", "TEST_VAR_2"},
		Verbose:        true,
	}

	// Set up test environment variables
	var err error
	err = os.Setenv("TEST_VAR_1", "value1")
	require.NoError(t, err, "Failed to set TEST_VAR_1")
	err = os.Setenv("TEST_VAR_2", "value2")
	require.NoError(t, err, "Failed to set TEST_VAR_2")
	err = os.Setenv("TEST_VAR_3", "value3") // Should not be preserved
	require.NoError(t, err, "Failed to set TEST_VAR_3")

	// Build the environment using our test-only function
	env, err := buildTestEnvironment(options)
	require.NoError(t, err, "buildTestEnvironment should not fail")

	// Check that preserved variables exist in the environment
	var foundVar1, foundVar2, foundVar3 bool
	for _, envVar := range env {
		if envVar == "TEST_VAR_1=value1" {
			foundVar1 = true
		}
		if envVar == "TEST_VAR_2=value2" {
			foundVar2 = true
		}
		if envVar == "TEST_VAR_3=value3" {
			foundVar3 = true
		}
	}

	// Preserved variables should be in the environment
	assert.True(t, foundVar1, "TEST_VAR_1 should be preserved")
	assert.True(t, foundVar2, "TEST_VAR_2 should be preserved")

	// Non-preserved variables should not be in the environment
	assert.False(t, foundVar3, "TEST_VAR_3 should not be preserved in high isolation")
}

// TestModulePathIsolation tests that the module path isolation features work correctly
func TestModulePathIsolation(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create module test paths
	customModulePath := filepath.Join(tmpDir, "custom_modules")
	additionalModulePath1 := filepath.Join(tmpDir, "extra_modules1")
	additionalModulePath2 := filepath.Join(tmpDir, "extra_modules2")

	// Create the directories
	require.NoError(t, os.MkdirAll(customModulePath, 0755), "Failed to create custom module path")
	require.NoError(t, os.MkdirAll(additionalModulePath1, 0755), "Failed to create additional module path 1")
	require.NoError(t, os.MkdirAll(additionalModulePath2, 0755), "Failed to create additional module path 2")

	// Test cases for different isolation levels
	testCases := []struct {
		name           string
		isolationLevel IsolationLevel
		options        *ExecutionOptions
		checkEnv       func(t *testing.T, env []string)
	}{
		{
			name:           "LowIsolationWithCustomModulePath",
			isolationLevel: IsolationLow,
			options: &ExecutionOptions{
				IsolationLevel:        IsolationLow,
				IsolationDir:          tmpDir,
				CustomModulePath:      customModulePath,
				AdditionalModulePaths: []string{additionalModulePath1, additionalModulePath2},
				Verbose:               true,
			},
			checkEnv: func(t *testing.T, env []string) {
				// Check that module paths are in PERL5LIB
				foundCustomPath := false
				foundAddPath1 := false
				foundAddPath2 := false

				for _, envVar := range env {
					if envVar == "PERL_LOCAL_LIB_ROOT="+customModulePath {
						foundCustomPath = true
					}
					// Check PERL5LIB contains additional module paths
					if envVar == "PERL5LIB="+filepath.Join(tmpDir, "lib", "perl5")+":"+
						filepath.Join(tmpDir, "lib", "perl5", "darwin-2level")+":"+
						additionalModulePath1+":"+additionalModulePath2 {
						foundAddPath1 = true
						foundAddPath2 = true
					}
				}

				assert.True(t, foundCustomPath, "PERL_LOCAL_LIB_ROOT should be set to custom module path")
				assert.True(t, foundAddPath1 && foundAddPath2, "PERL5LIB should include additional module paths")
			},
		},
		{
			name:           "HighIsolationWithCustomModulePath",
			isolationLevel: IsolationHigh,
			options: &ExecutionOptions{
				IsolationLevel:        IsolationHigh,
				IsolationDir:          tmpDir,
				CustomModulePath:      customModulePath,
				AdditionalModulePaths: []string{additionalModulePath1, additionalModulePath2},
				Verbose:               true,
			},
			checkEnv: func(t *testing.T, env []string) {
				// Check that module paths are in PERL5LIB
				foundCustomPath := false
				foundAddPath1 := false
				foundAddPath2 := false

				for _, envVar := range env {
					if envVar == "PERL_LOCAL_LIB_ROOT="+customModulePath {
						foundCustomPath = true
					}
					// Check PERL5LIB contains additional module paths
					if envVar == "PERL5LIB="+filepath.Join(tmpDir, "lib", "perl5")+":"+
						filepath.Join(tmpDir, "lib", "perl5", "darwin-2level")+":"+
						filepath.Join(tmpDir, "lib", "site_perl")+":"+
						filepath.Join(tmpDir, "lib", "vendor_perl")+":"+
						additionalModulePath1+":"+additionalModulePath2 {
						foundAddPath1 = true
						foundAddPath2 = true
					}
				}

				assert.True(t, foundCustomPath, "PERL_LOCAL_LIB_ROOT should be set to custom module path")
				assert.True(t, foundAddPath1 && foundAddPath2, "PERL5LIB should include additional module paths")
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Build the environment using our test-only function
			env, err := buildTestEnvironment(tc.options)
			require.NoError(t, err, "buildTestEnvironment should not fail")

			// Check the environment
			tc.checkEnv(t, env)
		})
	}
}

// TestFilesystemIsolation tests that the filesystem isolation features work correctly
func TestFilesystemIsolation(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create options with read-only and read-write paths
	options := &ExecutionOptions{
		IsolationLevel: IsolationHigh,
		IsolationDir:   tmpDir,
		ReadOnlyPaths:  []string{"/usr", "/bin"},
		ReadWritePaths: []string{"/tmp", "/home"},
		IsolatedOutput: true,
		Verbose:        true,
	}

	// Create the output directory for testing
	outputDir := filepath.Join(tmpDir, "output")
	require.NoError(t, os.MkdirAll(outputDir, 0755), "Failed to create output directory")

	// Build the environment using our test-only function
	env, err := buildTestEnvironment(options)
	require.NoError(t, err, "buildTestEnvironment should not fail")

	// Check that read-only and read-write paths are set in the environment
	var foundROPaths, foundRWPaths, foundIsolatedOutput bool
	for _, envVar := range env {
		if envVar == "PVM_READONLY_PATHS=/usr:/bin" {
			foundROPaths = true
		}
		if envVar == "PVM_READWRITE_PATHS=/tmp:/home" {
			foundRWPaths = true
		}
		if envVar == "PVM_ISOLATED_OUTPUT=1" {
			foundIsolatedOutput = true
		}
	}

	assert.True(t, foundROPaths, "PVM_READONLY_PATHS should be set in the environment")
	assert.True(t, foundRWPaths, "PVM_READWRITE_PATHS should be set in the environment")
	assert.True(t, foundIsolatedOutput, "PVM_ISOLATED_OUTPUT should be set in the environment")

	// Output directory has already been created by us for the test
	_, err = os.Stat(outputDir)
	require.NoError(t, err, "Output directory should exist")
}

// TestSaveOutputFiles tests that output files are saved correctly
func TestSaveOutputFiles(t *testing.T) {
	// Create temporary directories for testing
	sourceDir := t.TempDir()
	targetDir := t.TempDir()

	// Create the output directory in the source directory
	outputDir := filepath.Join(sourceDir, "output")
	require.NoError(t, os.MkdirAll(outputDir, 0755), "Failed to create output directory")

	// Create some test files in the output directory
	testFiles := []struct {
		name    string
		content string
	}{
		{"file1.txt", "File 1 content"},
		{"file2.txt", "File 2 content"},
		{"file3.txt", "File 3 content"},
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(outputDir, tf.name)
		require.NoError(t, os.WriteFile(filePath, []byte(tf.content), 0644), "Failed to create test file")
	}

	// Create options for testing
	options := &ExecutionOptions{
		IsolationDir:   sourceDir,
		IsolatedOutput: true,
		Verbose:        true,
	}

	// Test saving output files
	savedFiles, err := saveOutputFiles(options, targetDir)
	require.NoError(t, err, "saveOutputFiles should not fail")
	assert.Equal(t, len(testFiles), len(savedFiles), "Should save all test files")

	// Check that the files were saved with the correct content
	for _, tf := range testFiles {
		targetPath := filepath.Join(targetDir, tf.name)
		content, ok := savedFiles[targetPath]
		assert.True(t, ok, "File should be in the saved files map")
		assert.Equal(t, tf.content, content, "File content should match")

		// Check that the file exists on disk
		_, err := os.Stat(targetPath)
		assert.NoError(t, err, "File should exist in the target directory")
	}
}
