// ABOUTME: Tests for PVX isolation levels
// ABOUTME: Verifies that different isolation levels are applied correctly

package pvx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsolationLevelsCommand tests the environment creation for different isolation levels
func TestIsolationLevelsCommand(t *testing.T) {
	// Set up a temporary directory for the tests
	tmpDir := t.TempDir()

	// Create a test script
	scriptPath := filepath.Join(tmpDir, "test_script.pl")
	err := os.WriteFile(scriptPath, []byte("print 'Hello world';\n"), 0755)
	require.NoError(t, err, "Failed to create test script")

	// Define test cases for different isolation levels
	testCases := []struct {
		name           string
		isolationLevel IsolationLevel
		checkEnv       func(t *testing.T, env []string)
	}{
		{
			name:           "IsolationGlobal",
			isolationLevel: IsolationGlobal,
			checkEnv: func(t *testing.T, env []string) {
				// Should not contain PERL_LOCAL_LIB_ROOT
				for _, envVar := range env {
					if strings.HasPrefix(envVar, "PERL_LOCAL_LIB_ROOT=") {
						t.Errorf("IsolationGlobal should not set PERL_LOCAL_LIB_ROOT, but found: %s", envVar)
					}
				}
			},
		},
		{
			name:           "IsolationLocal",
			isolationLevel: IsolationLocal,
			checkEnv: func(t *testing.T, env []string) {
				// Should contain PERL_LOCAL_LIB_ROOT and preserve original PERL5LIB
				foundLocalLib := false
				for _, envVar := range env {
					if strings.HasPrefix(envVar, "PERL_LOCAL_LIB_ROOT=") {
						foundLocalLib = true
						break
					}
				}
				if !foundLocalLib {
					t.Errorf("IsolationLocal should set PERL_LOCAL_LIB_ROOT, but not found in env")
				}
			},
		},
		{
			name:           "IsolationClean",
			isolationLevel: IsolationClean,
			checkEnv: func(t *testing.T, env []string) {
				// Should only have PERL5LIB pointing to isolated directories
				foundLocalLib := false
				cleanPerl5Lib := false

				for _, envVar := range env {
					if strings.HasPrefix(envVar, "PERL_LOCAL_LIB_ROOT=") {
						foundLocalLib = true
					}
					if strings.HasPrefix(envVar, "PERL5LIB=") {
						// Should not contain system paths
						perl5lib := strings.TrimPrefix(envVar, "PERL5LIB=")
						for _, path := range strings.Split(perl5lib, ":") {
							if strings.Contains(path, "isolated") || strings.Contains(path, "test-isolation") {
								cleanPerl5Lib = true
								break
							}
						}
					}
				}

				if !foundLocalLib {
					t.Errorf("IsolationClean should set PERL_LOCAL_LIB_ROOT, but not found in env")
				}
				if !cleanPerl5Lib {
					t.Errorf("IsolationClean should have a clean PERL5LIB with isolation dirs")
				}
			},
		},
		// Note: High isolation level was eliminated
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mocks
			mockCmdOutput = "Hello world"
			mockExecShouldFail = false
			mockCmdArgs = nil
			mockCmdEnv = nil
			execCommand = mockExecCmd
			defer func() { execCommand = origExecCommand }()

			// Temporarily replace version resolution to avoid system call
			origResolvePerlExecutable := resolvePerlExecutable
			resolvePerlExecutable = func(options *ExecutionOptions) (string, error) {
				return "/mock/bin/perl", nil
			}
			defer func() { resolvePerlExecutable = origResolvePerlExecutable }()

			// Create execution options with the test isolation level
			options := &ExecutionOptions{
				ScriptPath:     scriptPath,
				IsolationLevel: tc.isolationLevel,
				Env:            map[string]string{},
				Verbose:        true,
				PerlVersion:    "/mock/bin/perl", // We're using a mock resolver but this helps create the proper execution path
			}

			// Execute the script
			_, err := ExecuteScript(options)
			require.NoError(t, err, "ExecuteScript failed")

			// Check the environment
			tc.checkEnv(t, mockCmdEnv)
		})
	}
}

// TestLegacyIsolationFlag tests that the deprecated Isolated flag still works
func TestLegacyIsolationFlag(t *testing.T) {
	// Set up a temporary directory for the tests
	tmpDir := t.TempDir()

	// Create a test script
	scriptPath := filepath.Join(tmpDir, "test_script.pl")
	err := os.WriteFile(scriptPath, []byte("print 'Hello world';\n"), 0755)
	require.NoError(t, err, "Failed to create test script")

	// Set up mocks
	mockCmdOutput = "Hello world"
	mockExecShouldFail = false
	mockCmdArgs = nil
	mockCmdEnv = nil
	execCommand = mockExecCmd
	defer func() { execCommand = origExecCommand }()

	// Temporarily replace version resolution to avoid system call
	origResolvePerlExecutable := resolvePerlExecutable
	resolvePerlExecutable = func(options *ExecutionOptions) (string, error) {
		return "/mock/bin/perl", nil
	}
	defer func() { resolvePerlExecutable = origResolvePerlExecutable }()

	// Test with legacy isolation flag
	options := &ExecutionOptions{
		ScriptPath:  scriptPath,
		Isolated:    true,
		Env:         map[string]string{},
		Verbose:     true,
		PerlVersion: "/mock/bin/perl",
	}

	// Execute the script
	_, err = ExecuteScript(options)
	require.NoError(t, err, "ExecuteScript failed")

	// Should contain PERL_LOCAL_LIB_ROOT (indicating it was converted to IsolationLocal)
	foundLocalLib := false
	for _, envVar := range mockCmdEnv {
		if strings.HasPrefix(envVar, "PERL_LOCAL_LIB_ROOT=") {
			foundLocalLib = true
			break
		}
	}
	if !foundLocalLib {
		t.Errorf("Legacy isolation flag should set PERL_LOCAL_LIB_ROOT, but not found in env")
	}
}

// TestEnvironmentBuilding tests the buildEnvironment function directly
func TestEnvironmentBuilding(t *testing.T) {
	// Test buildEnvironment with different isolation levels
	testCases := []struct {
		name           string
		isolationLevel IsolationLevel
		checkEnv       func(t *testing.T, env []string)
	}{
		{
			name:           "BuildEnvironmentNone",
			isolationLevel: IsolationGlobal,
			checkEnv: func(t *testing.T, env []string) {
				originalEnvCount := len(os.Environ())
				if len(env) < originalEnvCount {
					t.Errorf("Expected at least %d environment variables, got %d",
						originalEnvCount, len(env))
				}
			},
		},
		{
			name:           "BuildEnvironmentLow",
			isolationLevel: IsolationLocal,
			checkEnv: func(t *testing.T, env []string) {
				foundLocalLib := false
				perl5libSet := false

				for _, envVar := range env {
					if strings.HasPrefix(envVar, "PERL_LOCAL_LIB_ROOT=") {
						foundLocalLib = true
					}
					if strings.HasPrefix(envVar, "PERL5LIB=") {
						perl5libSet = true
					}
				}

				if !foundLocalLib {
					t.Errorf("Expected PERL_LOCAL_LIB_ROOT to be set")
				}
				if !perl5libSet {
					t.Errorf("Expected PERL5LIB to be set")
				}
			},
		},
		{
			name:           "BuildEnvironmentMedium",
			isolationLevel: IsolationClean,
			checkEnv: func(t *testing.T, env []string) {
				foundLocalLib := false
				perl5libSet := false

				for _, envVar := range env {
					if strings.HasPrefix(envVar, "PERL_LOCAL_LIB_ROOT=") {
						foundLocalLib = true
					}
					if strings.HasPrefix(envVar, "PERL5LIB=") {
						perl5libSet = true
					}
				}

				if !foundLocalLib {
					t.Errorf("Expected PERL_LOCAL_LIB_ROOT to be set")
				}
				if !perl5libSet {
					t.Errorf("Expected PERL5LIB to be set")
				}
			},
		},
		// Note: High isolation level test was eliminated
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			isolationDir := filepath.Join(tempDir, "pvm-test-env-building")

			options := &ExecutionOptions{
				IsolationLevel: tc.isolationLevel,
				IsolationDir:   isolationDir,
				Env:            map[string]string{},
			}

			env, err := buildEnvironment(options)
			require.NoError(t, err, "buildEnvironment failed")

			tc.checkEnv(t, env)
		})
	}
}

// TestInvalidIsolationLevel tests that an invalid isolation level is rejected
func TestInvalidIsolationLevel(t *testing.T) {
	options := &ExecutionOptions{
		IsolationLevel: "invalid_level",
		Env:            map[string]string{},
	}

	_, err := buildEnvironment(options)
	require.Error(t, err, "Invalid isolation level should be rejected")
	assert.Contains(t, err.Error(), "Invalid isolation level", "Error should indicate invalid level")
}
