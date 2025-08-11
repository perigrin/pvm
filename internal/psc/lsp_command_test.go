package psc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/config"
)

func TestLSPCommand_DefaultConfiguration(t *testing.T) {
	// Test that LSP command uses default configuration when no config file exists
	cmd := &cobra.Command{}
	cmd.Flags().Bool("stdio", false, "")
	cmd.Flags().Bool("tcp", false, "")
	cmd.Flags().Int("port", 0, "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().String("log-file", "", "")

	opts := &lspOptions{}

	// Create a default config with LSP settings
	defaultConfig := config.NewDefaultConfig()

	// Apply configuration
	applyLSPConfig(opts, defaultConfig.PSC.LSP, cmd)

	// Verify default values are applied
	assert.Equal(t, 9999, opts.port, "Default TCP port should be 9999")
	assert.Equal(t, false, opts.verbose, "Default verbose should be false")
	assert.Equal(t, "", opts.logFile, "Default log file should be empty")
}

func TestLSPCommand_ConfigurationFileOverride(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "psc_lsp_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test configuration file
	configPath := filepath.Join(tempDir, "pvm.toml")
	configContent := `
[psc.lsp]
log_file = "/tmp/test-lsp.log"
log_level = "debug"
verbose = true
default_mode = "tcp"
tcp_port = 8080
enable_hover = true
enable_completion = true
max_cache_size = 2000
exclude_patterns = ["**/test/**", "**/temp/**"]
include_directories = ["lib", "script", "bin"]
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Change to the temp directory so config can be found
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	os.Chdir(tempDir)

	// Parse the config file
	cfg, err := config.ParseFile(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg.PSC)
	require.NotNil(t, cfg.PSC.LSP)

	// Set up command with flags not changed
	cmd := &cobra.Command{}
	cmd.Flags().Bool("stdio", false, "")
	cmd.Flags().Bool("tcp", false, "")
	cmd.Flags().Int("port", 9999, "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().String("log-file", "", "")

	opts := &lspOptions{
		port: 9999, // Default value
	}

	// Apply configuration
	applyLSPConfig(opts, cfg.PSC.LSP, cmd)

	// Verify configuration values are applied
	assert.Equal(t, 8080, opts.port, "Config TCP port should override default")
	assert.Equal(t, true, opts.verbose, "Config verbose should override default")
	assert.Equal(t, "/tmp/test-lsp.log", opts.logFile, "Config log file should be set")
	assert.Equal(t, false, opts.stdio, "stdio should be false when default_mode is tcp")
	assert.Equal(t, true, opts.tcp, "tcp should be true when default_mode is tcp")
}

func TestLSPCommand_CommandLineFlagsPrecedence(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "psc_lsp_test_precedence_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test configuration file with specific values
	configPath := filepath.Join(tempDir, "pvm.toml")
	configContent := `
[psc.lsp]
verbose = true
tcp_port = 8080
log_file = "/config/log.txt"
default_mode = "tcp"
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Parse the config
	cfg, err := config.ParseFile(configPath)
	require.NoError(t, err)

	// Set up command with flags explicitly changed (simulating CLI usage)
	cmd := &cobra.Command{}
	cmd.Flags().Bool("stdio", false, "")
	cmd.Flags().Bool("tcp", false, "")
	cmd.Flags().Int("port", 9999, "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().String("log-file", "", "")

	// Mark flags as changed to simulate command-line usage
	cmd.Flags().Set("port", "7777")
	cmd.Flags().Set("verbose", "false")
	cmd.Flags().Set("log-file", "/cli/log.txt")
	cmd.Flags().Set("stdio", "true")

	opts := &lspOptions{
		port:    7777,
		verbose: false,
		logFile: "/cli/log.txt",
		stdio:   true,
		tcp:     false,
	}

	// Apply configuration
	applyLSPConfig(opts, cfg.PSC.LSP, cmd)

	// Verify command-line flags take precedence over config
	assert.Equal(t, 7777, opts.port, "CLI port should override config")
	assert.Equal(t, false, opts.verbose, "CLI verbose should override config")
	assert.Equal(t, "/cli/log.txt", opts.logFile, "CLI log file should override config")
	assert.Equal(t, true, opts.stdio, "CLI stdio should override config default_mode")
	assert.Equal(t, false, opts.tcp, "CLI stdio should override config default_mode")
}

func TestLSPCommand_ExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		setup    func()
		expected string
	}{
		{
			name:  "simple relative path",
			input: "logs/psc.log",
			setup: func() {},
			// expected will be checked as absolute path
		},
		{
			name:  "environment variable expansion",
			input: "$HOME/logs/psc.log",
			setup: func() {
				os.Setenv("HOME", "/test/home")
			},
			expected: "/test/home/logs/psc.log",
		},
		{
			name:  "XDG_CACHE_HOME expansion",
			input: "$XDG_CACHE_HOME/pvm/lsp.log",
			setup: func() {
				os.Setenv("XDG_CACHE_HOME", "/test/cache")
			},
			expected: "/test/cache/pvm/lsp.log",
		},
		{
			name:  "XDG_CACHE_HOME fallback to HOME/.cache",
			input: "$XDG_CACHE_HOME/pvm/lsp.log",
			setup: func() {
				os.Unsetenv("XDG_CACHE_HOME")
				os.Setenv("HOME", "/test/home")
			},
			expected: "/test/home/.cache/pvm/lsp.log",
		},
		{
			name:     "absolute path unchanged",
			input:    "/absolute/path/to/log.txt",
			setup:    func() {},
			expected: "/absolute/path/to/log.txt",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Clean up environment
			defer func() {
				os.Unsetenv("HOME")
				os.Unsetenv("XDG_CACHE_HOME")
			}()

			// Set up test environment
			test.setup()

			result := expandPath(test.input)

			if test.expected != "" {
				assert.Equal(t, test.expected, result)
			} else {
				// For relative paths, just check it becomes absolute
				assert.True(t, filepath.IsAbs(result), "Path should be converted to absolute")
			}
		})
	}
}

func TestApplyLSPConfig_DefaultModeSelection(t *testing.T) {
	tests := []struct {
		name          string
		defaultMode   string
		expectedStdio bool
		expectedTCP   bool
	}{
		{
			name:          "stdio mode",
			defaultMode:   "stdio",
			expectedStdio: true,
			expectedTCP:   false,
		},
		{
			name:          "tcp mode",
			defaultMode:   "tcp",
			expectedStdio: false,
			expectedTCP:   true,
		},
		{
			name:          "empty mode (no change)",
			defaultMode:   "",
			expectedStdio: false,
			expectedTCP:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lspConfig := &config.PSCLSPConfig{
				DefaultMode: test.defaultMode,
			}

			cmd := &cobra.Command{}
			cmd.Flags().Bool("stdio", false, "")
			cmd.Flags().Bool("tcp", false, "")
			cmd.Flags().Int("port", 9999, "")
			cmd.Flags().Bool("verbose", false, "")
			cmd.Flags().String("log-file", "", "")

			opts := &lspOptions{}

			applyLSPConfig(opts, lspConfig, cmd)

			assert.Equal(t, test.expectedStdio, opts.stdio, "stdio flag should match expected")
			assert.Equal(t, test.expectedTCP, opts.tcp, "tcp flag should match expected")
		})
	}
}

func TestLSPCommand_ConfigValidationIntegration(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "psc_lsp_validation_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a configuration file with validation errors
	configPath := filepath.Join(tempDir, "pvm.toml")
	configContent := `
[psc.lsp]
log_level = "invalid_level"
tcp_port = 70000
max_cache_size = -1
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Parse config without automatic validation
	cfg, err := config.ParseFileWithoutValidation(configPath)
	require.NoError(t, err, "Config parsing should succeed even with invalid values")

	// Manually validate the config
	errors := cfg.Validate()
	require.NotEmpty(t, errors, "Config should have validation errors")

	// Check specific validation errors
	errorMessages := make([]string, len(errors))
	for i, err := range errors {
		errorMessages[i] = err.Error()
	}

	assert.Contains(t, errorMessages, "LSP log_level must be one of: debug, info, warn, error")
	assert.Contains(t, errorMessages, "LSP tcp_port must be between 1 and 65535")
	assert.Contains(t, errorMessages, "LSP max_cache_size cannot be negative")
}
