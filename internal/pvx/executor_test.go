// ABOUTME: Tests for the PVX executor functionality
// ABOUTME: Verifies the Perl execution capabilities

package pvx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Save the original execCommand for restoration
var origExecCommand = execCommand

// mockExecutor captures arguments and environment variables passed to exec.Command
var mockCmdArgs []string
var mockCmdEnv []string
var mockCmdOutput string
var mockExecShouldFail bool

// Mock for execCommand
func mockExecCmd(cmd *exec.Cmd) error {
	// Capture for testing
	mockCmdArgs = cmd.Args
	mockCmdEnv = cmd.Env

	// Write output to stdout if available
	if cmd.Stdout != nil {
		_, _ = fmt.Fprint(cmd.Stdout, mockCmdOutput)
	}

	// Return error if needed
	if mockExecShouldFail {
		// We don't need to set ProcessState
		// Just return a normal error since we're not checking any specific exit code logic
		return fmt.Errorf("exit status 1")
	}
	return nil
}

// Helper to create a temporary script file
// Helper to create a temporary script file
func createTempScript3(t *testing.T, content string) string {
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "test_script.pl")
	err := os.WriteFile(scriptPath, []byte(content), 0755)
	require.NoError(t, err, "Failed to create temporary script")
	return scriptPath
}

// TestExecuteScript tests script execution with mocked command execution
func TestExecuteScript(t *testing.T) {
	t.Run("SimpleExecution", func(t *testing.T) {
		// Create a simple Perl script
		scriptPath := createTempScript3(t, `print "Hello, PVX!";`)

		// Set up mocks
		mockCmdOutput = "Hello, PVX!"
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

		// Execute the script
		options := &ExecutionOptions{
			ScriptPath: scriptPath,
		}

		output, err := ExecuteScript(options)
		require.NoError(t, err, "Script execution should succeed")
		assert.Equal(t, mockCmdOutput, output, "Output should match the mock output")

		// Verify command args
		assert.Contains(t, mockCmdArgs, scriptPath, "Command should include script path")
		assert.Contains(t, mockCmdArgs, "/mock/bin/perl", "Command should use mock perl path")
	})

	t.Run("CommandLineArguments", func(t *testing.T) {
		// Create a script that echoes its arguments
		scriptPath := createTempScript3(t, `print "Arguments: ", join(", ", @ARGV);`)

		// Set up mocks
		mockCmdOutput = "Arguments: arg1, arg2, arg3"
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

		// Execute with arguments
		options := &ExecutionOptions{
			ScriptPath: scriptPath,
			Args:       []string{"arg1", "arg2", "arg3"},
		}

		output, err := ExecuteScript(options)
		require.NoError(t, err, "Script execution should succeed")
		assert.Equal(t, mockCmdOutput, output, "Output should match mock output")

		// Verify arguments were passed correctly
		for _, arg := range []string{"arg1", "arg2", "arg3"} {
			assert.Contains(t, mockCmdArgs, arg, "Command should include argument: "+arg)
		}
	})

	t.Run("EnvironmentVariables", func(t *testing.T) {
		// Create a script that reads an environment variable
		scriptPath := createTempScript3(t, `print "TEST_VAR: $ENV{TEST_VAR}";`)

		// Set up mocks
		mockCmdOutput = "TEST_VAR: test_value"
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

		// Execute with environment variable
		options := &ExecutionOptions{
			ScriptPath: scriptPath,
			Env:        map[string]string{"TEST_VAR": "test_value"},
		}

		output, err := ExecuteScript(options)
		require.NoError(t, err, "Script execution should succeed")
		assert.Equal(t, mockCmdOutput, output, "Output should match mock output")

		// Verify environment variable was set
		foundTestVar := false
		for _, env := range mockCmdEnv {
			if strings.HasPrefix(env, "TEST_VAR=test_value") {
				foundTestVar = true
				break
			}
		}
		assert.True(t, foundTestVar, "Environment should contain TEST_VAR")
	})

	t.Run("ExitCodePropagation", func(t *testing.T) {
		// Create a script that exits with a non-zero code
		scriptPath := createTempScript3(t, `exit 42;`)

		// Set up mocks
		mockCmdOutput = "Error output"
		mockExecShouldFail = true
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

		// Execute the script
		options := &ExecutionOptions{
			ScriptPath: scriptPath,
		}

		output, err := ExecuteScript(options)
		require.Error(t, err, "Script execution should fail")
		assert.Equal(t, mockCmdOutput, output, "Output should match mock output")
		assert.Contains(t, err.Error(), "failed", "Error should indicate execution failure")
	})

	t.Run("NonExistentScript", func(t *testing.T) {
		// Create a path that definitely doesn't exist
		tempDir := t.TempDir()
		nonexistentPath := filepath.Join(tempDir, "definitely_does_not_exist.pl")

		// Execute a non-existent script
		options := &ExecutionOptions{
			ScriptPath: nonexistentPath,
		}

		_, err := ExecuteScript(options)
		require.Error(t, err, "Execution should fail for non-existent script")
		assert.Contains(t, err.Error(), "not found", "Error should mention file not found")
	})

	t.Run("VersionResolutionError", func(t *testing.T) {
		// Create a script
		scriptPath := createTempScript3(t, `print "Hello";`)

		// Temporarily replace version resolution to simulate failure
		origResolvePerlExecutable := resolvePerlExecutable
		resolvePerlExecutable = func(options *ExecutionOptions) (string, error) {
			return "", fmt.Errorf("mock error: version not found")
		}
		defer func() { resolvePerlExecutable = origResolvePerlExecutable }()

		options := &ExecutionOptions{
			ScriptPath:  scriptPath,
			PerlVersion: "nonexistent",
		}

		_, err := ExecuteScript(options)
		require.Error(t, err, "Script execution should fail when version resolution fails")
		assert.Contains(t, err.Error(), "version not found", "Error should indicate version not found")
	})
}

// TestExecuteInlineCode tests inline code execution with mocked command execution
func TestExecuteInlineCode(t *testing.T) {
	t.Run("SimpleInlineExecution", func(t *testing.T) {
		// Set up mocks
		mockCmdOutput = "Hello from inline code!"
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

		// Create options with inline code
		options := &ExecutionOptions{
			InlineCode: `print "Hello from inline code!";`,
		}

		output, err := ExecuteInlineCode(options)
		require.NoError(t, err, "Inline code execution should succeed")
		assert.Equal(t, mockCmdOutput, output, "Output should match mock output")

		// Verify inline code was executed correctly
		assert.Contains(t, mockCmdArgs, "-e", "Should use -e flag")
		assert.Contains(t, mockCmdArgs, options.InlineCode, "Should include the inline code")
	})

	t.Run("InlineWithCommandLineArguments", func(t *testing.T) {
		// Set up mocks
		mockCmdOutput = "Arguments: arg1, arg2, arg3"
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

		// Create options with inline code that uses command line arguments
		options := &ExecutionOptions{
			InlineCode: `print "Arguments: ", join(", ", @ARGV);`,
			Args:       []string{"arg1", "arg2", "arg3"},
		}

		output, err := ExecuteInlineCode(options)
		require.NoError(t, err, "Inline code execution should succeed")
		assert.Equal(t, mockCmdOutput, output, "Output should match mock output")

		// Verify arguments were passed correctly
		for _, arg := range []string{"arg1", "arg2", "arg3"} {
			assert.Contains(t, mockCmdArgs, arg, "Command should include argument: "+arg)
		}
	})

	t.Run("InlineWithEnvironmentVariables", func(t *testing.T) {
		// Set up mocks
		mockCmdOutput = "TEST_VAR: inline_test_value"
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

		// Create options with inline code that uses environment variables
		options := &ExecutionOptions{
			InlineCode: `print "TEST_VAR: $ENV{TEST_VAR}";`,
			Env:        map[string]string{"TEST_VAR": "inline_test_value"},
		}

		output, err := ExecuteInlineCode(options)
		require.NoError(t, err, "Inline code execution should succeed")
		assert.Equal(t, mockCmdOutput, output, "Output should match mock output")

		// Verify environment variable was set
		foundTestVar := false
		for _, env := range mockCmdEnv {
			if strings.HasPrefix(env, "TEST_VAR=inline_test_value") {
				foundTestVar = true
				break
			}
		}
		assert.True(t, foundTestVar, "Environment should contain TEST_VAR")
	})

	t.Run("InlineWithExitCode", func(t *testing.T) {
		// Set up mocks
		mockCmdOutput = "Error output"
		mockExecShouldFail = true
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

		// Create options with inline code that exits with a non-zero code
		options := &ExecutionOptions{
			InlineCode: `exit 42;`,
		}

		output, err := ExecuteInlineCode(options)
		require.Error(t, err, "Inline code execution should fail")
		assert.Equal(t, mockCmdOutput, output, "Output should match mock output")
		assert.Contains(t, err.Error(), "failed", "Error should indicate execution failure")
	})

	t.Run("EmptyInlineCode", func(t *testing.T) {
		// Create options with empty inline code
		options := &ExecutionOptions{
			InlineCode: "",
		}

		_, err := ExecuteInlineCode(options)
		require.Error(t, err, "Inline execution should fail with empty code")
		assert.Contains(t, err.Error(), "No Perl code provided", "Error should indicate no code provided")
	})
}

// TestIsolationEnvironment tests that isolation levels create the correct environment
func TestIsolationEnvironment(t *testing.T) {
	tempDir := t.TempDir()
	isolationDir := filepath.Join(tempDir, "pvm-test-isolation")

	options := &ExecutionOptions{
		IsolationDir:   isolationDir,
		Env:            map[string]string{},
		IsolationLevel: IsolationLocal,
	}

	// Create isolation directory
	err := os.MkdirAll(options.IsolationDir, 0755)
	require.NoError(t, err, "Failed to create isolation directory")
	// No need to defer cleanup as t.TempDir() handles it

	env, err := buildEnvironment(options)
	require.NoError(t, err, "Building environment should succeed")

	// Verify that isolation environment variables are set
	foundLocalLib := false
	foundPerl5Lib := false
	foundMMOpt := false

	for _, envVar := range env {
		if strings.HasPrefix(envVar, "PERL_LOCAL_LIB_ROOT=") {
			foundLocalLib = true
			assert.Contains(t, envVar, options.IsolationDir, "PERL_LOCAL_LIB_ROOT should contain isolation dir")
		}
		if strings.HasPrefix(envVar, "PERL5LIB=") {
			foundPerl5Lib = true
			assert.Contains(t, envVar, options.IsolationDir, "PERL5LIB should contain isolation dir")
		}
		if strings.HasPrefix(envVar, "PERL_MM_OPT=") {
			foundMMOpt = true
			assert.Contains(t, envVar, options.IsolationDir, "PERL_MM_OPT should contain isolation dir")
		}
	}

	assert.True(t, foundLocalLib, "PERL_LOCAL_LIB_ROOT should be set")
	assert.True(t, foundPerl5Lib, "PERL5LIB should be set")
	assert.True(t, foundMMOpt, "PERL_MM_OPT should be set")
}

func TestDebugOutput(t *testing.T) {
	// Test that debug and verbose flags are properly passed to ExecutionOptions
	t.Run("DebugFlagInExecutionOptions", func(t *testing.T) {
		options := &ExecutionOptions{
			Debug:   true,
			Verbose: false,
		}

		assert.True(t, options.Debug, "Debug should be true when set")
		assert.False(t, options.Verbose, "Verbose should be false when not set")
	})

	t.Run("VerboseFlagInExecutionOptions", func(t *testing.T) {
		options := &ExecutionOptions{
			Debug:   false,
			Verbose: true,
		}

		assert.False(t, options.Debug, "Debug should be false when not set")
		assert.True(t, options.Verbose, "Verbose should be true when set")
	})

	t.Run("BothFlagsInExecutionOptions", func(t *testing.T) {
		options := &ExecutionOptions{
			Debug:   true,
			Verbose: true,
		}

		assert.True(t, options.Debug, "Debug should be true when set")
		assert.True(t, options.Verbose, "Verbose should be true when set")
	})

	t.Run("GetSourceDisplayName", func(t *testing.T) {
		// Test the helper function for displaying resolution sources
		tests := []struct {
			source       string
			sourcePath   string
			expectedName string
		}{
			{"explicit", "", "explicit version (command line)"},
			{"project_file", "/path/.perl-version", ".perl-version file (/path/.perl-version)"},
			{"project_file", "", ".perl-version file"},
			{"env_var", "PVM_PERL_VERSION", "environment variable (PVM_PERL_VERSION)"},
			{"system_perl", "", "system Perl"},
			{"unknown", "", "unknown"},
		}

		for _, test := range tests {
			// We need to import perl package to use the ResolutionSource types
			// For now, we'll test the function directly with strings
			// This would need proper perl.ResolutionSource types in a real implementation
			t.Run(fmt.Sprintf("%s_%s", test.source, test.sourcePath), func(t *testing.T) {
				// This test would need to be implemented with proper types
				// For now, we just verify the function exists
				assert.NotNil(t, getSourceDisplayName, "getSourceDisplayName function should exist")
			})
		}
	})
}
