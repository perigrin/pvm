// ABOUTME: Tests for PVX commands
// ABOUTME: Verifies the command-line functionality of the Perl executor

package pvx

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/cli"
)

func TestPVXDebugOutput(t *testing.T) {
	t.Run("VerboseFlag", func(t *testing.T) {
		// Test that --verbose flag is parsed correctly
		cmd := NewCommand()

		// Set up arguments with verbose flag
		cmd.SetArgs([]string{"--verbose", "-e", "print 'test'"})

		// Parse flags
		err := cmd.ParseFlags([]string{"--verbose", "-e", "print 'test'"})
		require.NoError(t, err)

		// Check that verbose flag is set
		verbose, err := cmd.Flags().GetBool("verbose")
		require.NoError(t, err)
		assert.True(t, verbose, "Verbose flag should be true")
	})

	t.Run("DebugFlag", func(t *testing.T) {
		// Test that --debug flag is parsed correctly
		cmd := NewCommand()

		// Set up arguments with debug flag
		cmd.SetArgs([]string{"--debug", "-e", "print 'test'"})

		// Parse flags
		err := cmd.ParseFlags([]string{"--debug", "-e", "print 'test'"})
		require.NoError(t, err)

		// Check that debug flag is set
		debug, err := cmd.Flags().GetBool("debug")
		require.NoError(t, err)
		assert.True(t, debug, "Debug flag should be true")
	})

	t.Run("BothVerboseAndDebug", func(t *testing.T) {
		// Test that both flags can be used together
		cmd := NewCommand()

		// Set up arguments with both flags
		cmd.SetArgs([]string{"--verbose", "--debug", "-e", "print 'test'"})

		// Parse flags
		err := cmd.ParseFlags([]string{"--verbose", "--debug", "-e", "print 'test'"})
		require.NoError(t, err)

		// Check that both flags are set
		verbose, err := cmd.Flags().GetBool("verbose")
		require.NoError(t, err)
		assert.True(t, verbose, "Verbose flag should be true")

		debug, err := cmd.Flags().GetBool("debug")
		require.NoError(t, err)
		assert.True(t, debug, "Debug flag should be true")
	})
}

func TestPVXAutoInstallFlag(t *testing.T) {
	t.Run("LongForm", func(t *testing.T) {
		// Test that --auto-install flag is parsed correctly
		cmd := NewCommand()

		// Parse flags with long form
		err := cmd.ParseFlags([]string{"--auto-install", "-e", "print 'test'"})
		require.NoError(t, err)

		// Check that auto-install flag is set
		autoInstall, err := cmd.Flags().GetBool("auto-install")
		require.NoError(t, err)
		assert.True(t, autoInstall, "Auto-install flag should be true with long form")
	})

	t.Run("ShortForm", func(t *testing.T) {
		// Test that -a flag is parsed correctly as alias for --auto-install
		cmd := NewCommand()

		// Parse flags with short form
		err := cmd.ParseFlags([]string{"-a", "-e", "print 'test'"})
		require.NoError(t, err)

		// Check that auto-install flag is set via short alias
		autoInstall, err := cmd.Flags().GetBool("auto-install")
		require.NoError(t, err)
		assert.True(t, autoInstall, "Auto-install flag should be true with short form -a")
	})

	t.Run("DefaultValue", func(t *testing.T) {
		// Test that auto-install flag defaults to false
		cmd := NewCommand()

		// Parse flags without auto-install flag
		err := cmd.ParseFlags([]string{"-e", "print 'test'"})
		require.NoError(t, err)

		// Check that auto-install flag is false by default
		autoInstall, err := cmd.Flags().GetBool("auto-install")
		require.NoError(t, err)
		assert.False(t, autoInstall, "Auto-install flag should be false by default")
	})
}

func TestPVXCommand(t *testing.T) {
	// We need to import the setupTestMocks function from executor_test.go
	// For now, we'll skip the integration test since it's more complex to mock
	t.Run("ShowsHelpWhenNoArgs", func(t *testing.T) {
		// Create a new PVX command
		cmd := NewCommand()

		// Capture output
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetErr(output)

		// Execute with no arguments
		cmd.SetArgs([]string{})
		err := cmd.Execute()

		// Should succeed and output should contain help text
		require.NoError(t, err)
		assert.Contains(t, output.String(), "pvx [options] script.pl [args...]")
	})

	t.Run("IntegrationWithExecutor", func(t *testing.T) {
		// Reset global UI state first
		cli.ResetGlobalState()

		// Create a temporary script
		tempDir := t.TempDir()
		scriptPath := filepath.Join(tempDir, "test.pl")
		err := os.WriteFile(scriptPath, []byte(`print "Hello from PVX CLI";`), 0755)
		require.NoError(t, err)

		// Set up mocks for the executor
		origExecCommand := execCommand
		origResolvePerlExecutable := resolvePerlExecutable

		defer func() {
			execCommand = origExecCommand
			resolvePerlExecutable = origResolvePerlExecutable
			// Reset global UI state
			cli.ResetGlobalState()
		}()

		// Mock command execution
		execCommand = func(cmd *exec.Cmd) error {
			// Write mock output
			if cmd.Stdout != nil {
				cmd.Stdout.Write([]byte("Hello from PVX CLI"))
			}
			return nil
		}

		// Mock perl executable resolution
		resolvePerlExecutable = func(options *ExecutionOptions) (string, error) {
			return "/usr/bin/perl", nil
		}

		// Create a new PVX command
		cmd := NewCommand()

		// Capture output
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetErr(output)

		// Override os.Exit to prevent test termination
		origOsExit := osExit
		defer func() { osExit = origOsExit }()

		exitCode := 0
		osExit = func(code int) {
			exitCode = code
		}

		// Execute with script path
		cmd.SetArgs([]string{scriptPath})
		err = cmd.Execute()

		// Should succeed
		require.NoError(t, err)
		assert.Equal(t, 0, exitCode, "Exit code should be 0")

		// Verify the mock output was captured
		assert.Contains(t, output.String(), "Hello from PVX CLI")
	})

	t.Run("InlineCodeExecution", func(t *testing.T) {
		// Reset global UI state first
		cli.ResetGlobalState()

		// Set up mocks for the executor
		origExecCommand := execCommand
		origResolvePerlExecutable := resolvePerlExecutable

		defer func() {
			execCommand = origExecCommand
			resolvePerlExecutable = origResolvePerlExecutable
			// Reset global UI state
			cli.ResetGlobalState()
		}()

		// Mock command execution
		execCommand = func(cmd *exec.Cmd) error {
			// Write mock output
			if cmd.Stdout != nil {
				cmd.Stdout.Write([]byte("Hello from inline Perl!"))
			}
			return nil
		}

		// Mock perl executable resolution
		resolvePerlExecutable = func(options *ExecutionOptions) (string, error) {
			return "/usr/bin/perl", nil
		}

		// Create a new PVX command
		cmd := NewCommand()

		// Capture output
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetErr(output)

		// Override os.Exit to prevent test termination
		origOsExit := osExit
		defer func() { osExit = origOsExit }()

		exitCode := 0
		osExit = func(code int) {
			exitCode = code
		}

		// Set the -e flag with inline Perl code
		cmd.SetArgs([]string{"-e", `print "Hello from inline Perl!";`})
		err := cmd.Execute()

		// Should succeed
		require.NoError(t, err)
		assert.Equal(t, 0, exitCode, "Exit code should be 0")

		// Verify the mock output was captured
		assert.Contains(t, output.String(), "Hello from inline Perl!")
	})
}

// Test-specific init function to prevent os.Exit from terminating tests
func init() {
	// Use a no-op function for os.Exit to prevent test termination
	// The original osExit is already defined in command.go
	origExit := osExit
	osExit = func(code int) {
		// We need to keep the original function reference but not do anything
		// in global init to prevent termination
		if origExit == nil {
			// Should never happen, but avoid nil panic
			return
		}
	}
}
