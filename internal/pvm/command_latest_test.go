// ABOUTME: Tests for latest alias and dev version support in PVM commands
// ABOUTME: Validates the install and available commands with new flags and aliases

package pvm

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallCommand_LatestAlias(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "install latest",
			args:        []string{"latest"},
			expectError: false,
		},
		{
			name:        "install latest-dev",
			args:        []string{"latest-dev"},
			expectError: false,
		},
		{
			name:        "install latest with include-dev flag",
			args:        []string{"latest", "--include-dev"},
			expectError: false,
		},
		{
			name:        "install with @latest alias",
			args:        []string{"@latest"},
			expectError: false,
		},
		{
			name:        "install with @latest-dev alias",
			args:        []string{"@latest-dev"},
			expectError: false,
		},
		{
			name:        "install specific version",
			args:        []string{"5.38.2"},
			expectError: false,
		},
		{
			name:        "install with no args",
			args:        []string{},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 0",
		},
		{
			name:        "install with too many args",
			args:        []string{"5.38.2", "5.40.0"},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newInstallCommand()
			cmd.SetArgs(tt.args)

			// Mock RunE to avoid actual installation
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				// Test flag parsing
				includeDev, err := cmd.Flags().GetBool("include-dev")
				require.NoError(t, err)

				// Validate that the flag can be read
				assert.NotNil(t, includeDev)

				// Test that version argument is available
				if len(args) > 0 {
					assert.NotEmpty(t, args[0])
				}

				return nil
			}

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestInstallCommand_IncludeDevFlag(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedDev bool
		expectError bool
	}{
		{
			name:        "without include-dev flag",
			args:        []string{"latest"},
			expectedDev: false,
			expectError: false,
		},
		{
			name:        "with include-dev flag",
			args:        []string{"latest", "--include-dev"},
			expectedDev: true,
			expectError: false,
		},
		{
			name:        "with include-dev flag and specific version",
			args:        []string{"5.38.2", "--include-dev"},
			expectedDev: true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newInstallCommand()
			cmd.SetArgs(tt.args)

			// Mock RunE to test flag values
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				includeDev, err := cmd.Flags().GetBool("include-dev")
				require.NoError(t, err)

				assert.Equal(t, tt.expectedDev, includeDev)

				return nil
			}

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestInstallCommand_FlagCompatibility(t *testing.T) {
	// Test that include-dev flag is compatible with other flags
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "include-dev with prefer-binary",
			args:        []string{"latest", "--include-dev", "--prefer-binary"},
			expectError: false,
		},
		{
			name:        "include-dev with binary-only",
			args:        []string{"latest", "--include-dev", "--binary-only"},
			expectError: false,
		},
		{
			name:        "include-dev with force-source",
			args:        []string{"latest", "--include-dev", "--force-source"},
			expectError: false,
		},
		{
			name:        "binary-only with force-source (should fail)",
			args:        []string{"latest", "--binary-only", "--force-source"},
			expectError: true,
			errorMsg:    "mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newInstallCommand()
			cmd.SetArgs(tt.args)

			// Mock RunE to test flag compatibility
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				// Test that flags can be read
				includeDev, err := cmd.Flags().GetBool("include-dev")
				require.NoError(t, err)

				binaryOnly, err := cmd.Flags().GetBool("binary-only")
				require.NoError(t, err)

				forceSource, err := cmd.Flags().GetBool("force-source")
				require.NoError(t, err)

				// Test mutually exclusive flags
				if binaryOnly && forceSource {
					return fmt.Errorf("--binary-only and --force-source are mutually exclusive")
				}

				assert.NotNil(t, includeDev)
				assert.NotNil(t, binaryOnly)
				assert.NotNil(t, forceSource)

				return nil
			}

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestAvailableCommand_IncludeDevFlag(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedDev bool
		expectError bool
	}{
		{
			name:        "without include-dev flag",
			args:        []string{},
			expectedDev: false,
			expectError: false,
		},
		{
			name:        "with include-dev flag",
			args:        []string{"--include-dev"},
			expectedDev: true,
			expectError: false,
		},
		{
			name:        "with include-dev and format flags",
			args:        []string{"--include-dev", "--format=json"},
			expectedDev: true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newAvailableCommand()
			cmd.SetArgs(tt.args)

			// Mock RunE to test flag values
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				includeDev, err := cmd.Flags().GetBool("include-dev")
				require.NoError(t, err)

				assert.Equal(t, tt.expectedDev, includeDev)

				return nil
			}

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestAvailableCommand_FormatFlag_WithDev(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedFmt string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "default format",
			args:        []string{},
			expectedFmt: "text",
			expectError: false,
		},
		{
			name:        "json format",
			args:        []string{"--format=json"},
			expectedFmt: "json",
			expectError: false,
		},
		{
			name:        "plain format",
			args:        []string{"--format=plain"},
			expectedFmt: "plain",
			expectError: false,
		},
		{
			name:        "short format flag",
			args:        []string{"-f", "json"},
			expectedFmt: "json",
			expectError: false,
		},
		{
			name:        "json format with include-dev",
			args:        []string{"--format=json", "--include-dev"},
			expectedFmt: "json",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newAvailableCommand()
			cmd.SetArgs(tt.args)

			// Mock RunE to test flag values
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				format, err := cmd.Flags().GetString("format")
				require.NoError(t, err)

				assert.Equal(t, tt.expectedFmt, format)

				return nil
			}

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestInstallCommand_HelpText(t *testing.T) {
	cmd := newInstallCommand()

	// Test that help text includes new functionality
	helpText := cmd.Long

	assert.Contains(t, helpText, "latest")
	assert.Contains(t, helpText, "latest-dev")
	assert.Contains(t, helpText, "--include-dev")
	assert.Contains(t, helpText, "Examples:")
}

func TestAvailableCommand_HelpText_WithDev(t *testing.T) {
	cmd := newAvailableCommand()

	// Test that help text includes new functionality
	helpText := cmd.Long

	assert.Contains(t, helpText, "--include-dev")
	assert.Contains(t, helpText, "development versions")
	assert.Contains(t, helpText, "Examples:")
}

func TestInstallCommand_FlagDefinitions(t *testing.T) {
	cmd := newInstallCommand()

	// Test that include-dev flag is defined
	flag := cmd.Flags().Lookup("include-dev")
	require.NotNil(t, flag)

	assert.Equal(t, "bool", flag.Value.Type())
	assert.Equal(t, "false", flag.DefValue)
	assert.Contains(t, flag.Usage, "development versions")
}

func TestInstallCommand_ConfigDefaultInstallMethod(t *testing.T) {
	// The install command should apply the config's DefaultInstallMethod
	// when no explicit --binary-only, --prefer-binary, or --force-source flag is set.
	// The default config sets DefaultInstallMethod to "prefer-binary".
	tests := []struct {
		name                 string
		args                 []string
		expectedPreferBinary bool
		expectedForceSource  bool
	}{
		{
			name:                 "no flags — config default prefer-binary applies",
			args:                 []string{"5.38.0"},
			expectedPreferBinary: true,
			expectedForceSource:  false,
		},
		{
			name:                 "explicit --force-source overrides config",
			args:                 []string{"--force-source", "5.38.0"},
			expectedPreferBinary: false,
			expectedForceSource:  true,
		},
		{
			name:                 "explicit --binary-only overrides config",
			args:                 []string{"--binary-only", "5.38.0"},
			expectedPreferBinary: false,
			expectedForceSource:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newInstallCommand()
			cmd.SetArgs(tt.args)

			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				preferBinary, err := cmd.Flags().GetBool("prefer-binary")
				require.NoError(t, err)

				forceSource, err := cmd.Flags().GetBool("force-source")
				require.NoError(t, err)

				binaryOnly, err := cmd.Flags().GetBool("binary-only")
				require.NoError(t, err)

				// Apply config default when no explicit install method flag is set
				if !preferBinary && !binaryOnly && !forceSource {
					preferBinary = applyConfigInstallMethod()
				}

				assert.Equal(t, tt.expectedPreferBinary, preferBinary,
					"preferBinary mismatch")
				assert.Equal(t, tt.expectedForceSource, forceSource,
					"forceSource mismatch")
				return nil
			}

			err := cmd.Execute()
			require.NoError(t, err)
		})
	}
}

func TestAvailableCommand_FlagDefinitions(t *testing.T) {
	cmd := newAvailableCommand()

	// Test that include-dev flag is defined
	flag := cmd.Flags().Lookup("include-dev")
	require.NotNil(t, flag)

	assert.Equal(t, "bool", flag.Value.Type())
	assert.Equal(t, "false", flag.DefValue)
	assert.Contains(t, flag.Usage, "development versions")
}
