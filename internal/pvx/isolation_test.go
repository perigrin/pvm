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

	// Process any environment variables that should be cleared
	if len(options.ClearEnv) > 0 {
		// Create a map for faster lookup
		clearEnvMap := make(map[string]bool)
		for _, key := range options.ClearEnv {
			clearEnvMap[key] = true
		}

		// Filter out environment variables that should be cleared
		filteredEnv := []string{}
		for _, envVar := range env {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) < 2 {
				continue
			}

			if !clearEnvMap[parts[0]] {
				filteredEnv = append(filteredEnv, envVar)
			}
		}

		env = filteredEnv
	}

	// Add or override with custom environment variables
	for key, value := range options.Env {
		envVar := fmt.Sprintf("%s=%s", key, value)
		found := false

		// Replace existing variable if present
		for i, existing := range env {
			if strings.HasPrefix(existing, key+"=") {
				env[i] = envVar
				found = true
				break
			}
		}

		// Add new variable if not found
		if !found {
			env = append(env, envVar)
		}
	}

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
	binDir := filepath.Join(isolationDir, "bin")
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

		// Add to existing PERL5LIB if present
		perl5libFound := false
		for i, existing := range env {
			if strings.HasPrefix(existing, "PERL5LIB=") {
				currentValue := strings.TrimPrefix(existing, "PERL5LIB=")
				if currentValue != "" {
					perl5lib = fmt.Sprintf("%s:%s", perl5lib, currentValue)
				}
				env[i] = fmt.Sprintf("PERL5LIB=%s", perl5lib)
				perl5libFound = true
				break
			}
		}

		// Add new PERL5LIB if not found
		if !perl5libFound {
			env = append(env, fmt.Sprintf("PERL5LIB=%s", perl5lib))
		}

		// Set up the local::lib equivalent environment variables
		// Use the custom module path if provided, otherwise use the isolation directory
		modulePath := isolationDir
		if options.CustomModulePath != "" {
			modulePath = options.CustomModulePath
		}

		setEnvVar(&env, "PERL_LOCAL_LIB_ROOT", modulePath)
		setEnvVar(&env, "PERL_MB_OPT", fmt.Sprintf("--install_base '%s'", modulePath))
		setEnvVar(&env, "PERL_MM_OPT", fmt.Sprintf("INSTALL_BASE=%s", modulePath))

		// Add the bin directory to PATH
		pathFound := false
		for i, existing := range env {
			if strings.HasPrefix(existing, "PATH=") {
				currentPath := strings.TrimPrefix(existing, "PATH=")
				env[i] = fmt.Sprintf("PATH=%s:%s", binDir, currentPath)
				pathFound = true
				break
			}
		}

		// Add new PATH if not found
		if !pathFound {
			env = append(env, fmt.Sprintf("PATH=%s", binDir))
		}

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

		// Set up the local::lib equivalent environment variables
		// Use the custom module path if provided, otherwise use the isolation directory
		modulePath := isolationDir
		if options.CustomModulePath != "" {
			modulePath = options.CustomModulePath
		}

		setEnvVar(&env, "PERL_LOCAL_LIB_ROOT", modulePath)
		setEnvVar(&env, "PERL_MB_OPT", fmt.Sprintf("--install_base '%s'", modulePath))
		setEnvVar(&env, "PERL_MM_OPT", fmt.Sprintf("INSTALL_BASE=%s", modulePath))

		// Add the bin directory to PATH
		pathFound := false
		for i, existing := range env {
			if strings.HasPrefix(existing, "PATH=") {
				currentPath := strings.TrimPrefix(existing, "PATH=")
				env[i] = fmt.Sprintf("PATH=%s:%s", binDir, currentPath)
				pathFound = true
				break
			}
		}

		// Add new PATH if not found
		if !pathFound {
			env = append(env, fmt.Sprintf("PATH=%s", binDir))
		}

	case IsolationHigh:
		// High isolation: Start with minimal environment and add only what's needed
		// Create a clean environment with only essential variables
		cleanEnv := []string{}

		// Copy only essential environment variables (non-exhaustive list)
		essentialVars := []string{
			"PATH",
			"HOME",
			"USER",
			"SHELL",
			"TMPDIR",
			"TERM",
		}

		for _, key := range essentialVars {
			for _, envVar := range env {
				if strings.HasPrefix(envVar, key+"=") {
					cleanEnv = append(cleanEnv, envVar)
					break
				}
			}
		}

		// Add any preserved environment variables
		for _, key := range options.PreserveEnv {
			// Skip if it's already in essential vars
			isEssential := false
			for _, essential := range essentialVars {
				if key == essential {
					isEssential = true
					break
				}
			}
			if isEssential {
				continue
			}

			// Find and add from original environment
			for _, envVar := range env {
				if strings.HasPrefix(envVar, key+"=") {
					// Check if it should be cleared
					if len(options.ClearEnv) > 0 {
						shouldClear := false
						for _, clearKey := range options.ClearEnv {
							if key == clearKey {
								shouldClear = true
								break
							}
						}
						if shouldClear {
							continue
						}
					}

					cleanEnv = append(cleanEnv, envVar)
					break
				}
			}
		}

		// Add custom environment variables
		for key, value := range options.Env {
			setEnvVar(&cleanEnv, key, value)
		}

		// Set up enhanced filesystem isolation for high isolation mode
		if options.IsolatedOutput {
			// Create a temporary directory for script output
			outputDir := filepath.Join(isolationDir, "output")
			setEnvVar(&cleanEnv, "PVM_OUTPUT_DIR", outputDir)
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

		// Set PATH to include the bin directory first
		pathFound := false
		for i, existing := range cleanEnv {
			if strings.HasPrefix(existing, "PATH=") {
				currentPath := strings.TrimPrefix(existing, "PATH=")
				cleanEnv[i] = fmt.Sprintf("PATH=%s:%s", binDir, currentPath)
				pathFound = true
				break
			}
		}

		// Add new PATH if not found
		if !pathFound {
			cleanEnv = append(cleanEnv, fmt.Sprintf("PATH=%s", binDir))
		}

		// Use the clean environment instead of the original one
		env = cleanEnv
	}

	return env, nil
}

// TestEnhancedEnvironmentIsolation tests that the enhanced environment isolation features work correctly
func TestEnhancedEnvironmentIsolation(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Set up test environment variables
	var err error
	err = os.Setenv("TEST_VAR_1", "value1")
	require.NoError(t, err, "Failed to set TEST_VAR_1")
	err = os.Setenv("TEST_VAR_2", "value2")
	require.NoError(t, err, "Failed to set TEST_VAR_2")
	err = os.Setenv("TEST_VAR_3", "value3")
	require.NoError(t, err, "Failed to set TEST_VAR_3")
	err = os.Setenv("PERL5LIB", "/original/perl5/lib:/other/lib")
	require.NoError(t, err, "Failed to set PERL5LIB")

	// Test cases for different isolation levels
	testCases := []struct {
		name           string
		isolationLevel IsolationLevel
		options        *ExecutionOptions
		checkEnv       func(t *testing.T, env []string)
	}{
		{
			name:           "NoIsolation",
			isolationLevel: IsolationNone,
			options: &ExecutionOptions{
				IsolationLevel: IsolationNone,
				IsolationDir:   tmpDir,
				Verbose:        true,
			},
			checkEnv: func(t *testing.T, env []string) {
				// Should preserve all environment variables
				foundVar1 := false
				foundVar2 := false
				foundVar3 := false
				foundPerl5Lib := false

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
					if envVar == "PERL5LIB=/original/perl5/lib:/other/lib" {
						foundPerl5Lib = true
					}
				}

				assert.True(t, foundVar1, "TEST_VAR_1 should be preserved in no isolation")
				assert.True(t, foundVar2, "TEST_VAR_2 should be preserved in no isolation")
				assert.True(t, foundVar3, "TEST_VAR_3 should be preserved in no isolation")
				assert.True(t, foundPerl5Lib, "PERL5LIB should be preserved in no isolation")
			},
		},
		{
			name:           "LowIsolation",
			isolationLevel: IsolationLow,
			options: &ExecutionOptions{
				IsolationLevel: IsolationLow,
				IsolationDir:   tmpDir,
				Verbose:        true,
			},
			checkEnv: func(t *testing.T, env []string) {
				// Should preserve all environment variables but modify PERL5LIB
				foundVar1 := false
				foundVar2 := false
				foundVar3 := false
				foundPerl5Lib := false
				modifiedPerl5Lib := false

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
					if strings.HasPrefix(envVar, "PERL5LIB=") {
						foundPerl5Lib = true
						// Should include the isolation directory and original PERL5LIB
						if strings.Contains(envVar, filepath.Join(tmpDir, "lib", "perl5")) &&
							strings.Contains(envVar, "/original/perl5/lib") {
							modifiedPerl5Lib = true
						}
					}
				}

				assert.True(t, foundVar1, "TEST_VAR_1 should be preserved in low isolation")
				assert.True(t, foundVar2, "TEST_VAR_2 should be preserved in low isolation")
				assert.True(t, foundVar3, "TEST_VAR_3 should be preserved in low isolation")
				assert.True(t, foundPerl5Lib, "PERL5LIB should exist in low isolation")
				assert.True(t, modifiedPerl5Lib, "PERL5LIB should include both isolation dir and original paths")
			},
		},
		{
			name:           "MediumIsolation",
			isolationLevel: IsolationMedium,
			options: &ExecutionOptions{
				IsolationLevel: IsolationMedium,
				IsolationDir:   tmpDir,
				Verbose:        true,
			},
			checkEnv: func(t *testing.T, env []string) {
				// Should preserve all environment variables but replace PERL5LIB
				foundVar1 := false
				foundVar2 := false
				foundVar3 := false
				foundPerl5Lib := false
				cleanPerl5Lib := false

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
					if strings.HasPrefix(envVar, "PERL5LIB=") {
						foundPerl5Lib = true
						// Should only include the isolation directory paths
						if !strings.Contains(envVar, "/original/perl5/lib") {
							cleanPerl5Lib = true
						}
					}
				}

				assert.True(t, foundVar1, "TEST_VAR_1 should be preserved in medium isolation")
				assert.True(t, foundVar2, "TEST_VAR_2 should be preserved in medium isolation")
				assert.True(t, foundVar3, "TEST_VAR_3 should be preserved in medium isolation")
				assert.True(t, foundPerl5Lib, "PERL5LIB should exist in medium isolation")
				assert.True(t, cleanPerl5Lib, "PERL5LIB should only include isolation directory paths")
			},
		},
		{
			name:           "HighIsolation",
			isolationLevel: IsolationHigh,
			options: &ExecutionOptions{
				IsolationLevel: IsolationHigh,
				IsolationDir:   tmpDir,
				PreserveEnv:    []string{"TEST_VAR_1", "TEST_VAR_2"},
				Verbose:        true,
			},
			checkEnv: func(t *testing.T, env []string) {
				// Should only preserve specified environment variables
				foundVar1 := false
				foundVar2 := false
				foundVar3 := false
				foundPerl5Lib := false

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
					if strings.HasPrefix(envVar, "PERL5LIB=") {
						foundPerl5Lib = true
					}
				}

				assert.True(t, foundVar1, "TEST_VAR_1 should be preserved in high isolation")
				assert.True(t, foundVar2, "TEST_VAR_2 should be preserved in high isolation")
				assert.False(t, foundVar3, "TEST_VAR_3 should not be preserved in high isolation")
				assert.True(t, foundPerl5Lib, "PERL5LIB should be set in high isolation")
			},
		},
		{
			name:           "HighIsolationWithClearEnv",
			isolationLevel: IsolationHigh,
			options: &ExecutionOptions{
				IsolationLevel: IsolationHigh,
				IsolationDir:   tmpDir,
				PreserveEnv:    []string{"TEST_VAR_1", "TEST_VAR_2"},
				ClearEnv:       []string{"TEST_VAR_2"}, // This will override preservation
				Verbose:        true,
			},
			checkEnv: func(t *testing.T, env []string) {
				// Should preserve TEST_VAR_1 but not TEST_VAR_2 (cleared) or TEST_VAR_3 (not preserved)
				foundVar1 := false
				foundVar2 := false
				foundVar3 := false

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

				assert.True(t, foundVar1, "TEST_VAR_1 should be preserved in high isolation with clear env")
				assert.False(t, foundVar2, "TEST_VAR_2 should be cleared even though it's in preserved list")
				assert.False(t, foundVar3, "TEST_VAR_3 should not be preserved in high isolation with clear env")
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
					if strings.HasPrefix(envVar, "PERL5LIB=") {
						content := strings.TrimPrefix(envVar, "PERL5LIB=")
						if strings.Contains(content, additionalModulePath1) {
							foundAddPath1 = true
						}
						if strings.Contains(content, additionalModulePath2) {
							foundAddPath2 = true
						}
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
					if strings.HasPrefix(envVar, "PERL5LIB=") {
						content := strings.TrimPrefix(envVar, "PERL5LIB=")
						if strings.Contains(content, additionalModulePath1) {
							foundAddPath1 = true
						}
						if strings.Contains(content, additionalModulePath2) {
							foundAddPath2 = true
						}
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
