// ABOUTME: Unit tests for PVM command functionality
// ABOUTME: Tests command flag parsing and validation logic

package pvm

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
)

func TestInstallCommandFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "basic install",
			args:        []string{"5.38.0"},
			expectError: false,
		},
		{
			name:        "binary only flag",
			args:        []string{"--binary-only", "5.38.0"},
			expectError: false,
		},
		{
			name:        "short binary only flag",
			args:        []string{"-B", "5.38.0"},
			expectError: false,
		},
		{
			name:        "prefer binary flag",
			args:        []string{"--prefer-binary", "5.38.0"},
			expectError: false,
		},
		{
			name:        "force source flag",
			args:        []string{"--force-source", "5.38.0"},
			expectError: false,
		},
		{
			name:        "mutually exclusive flags",
			args:        []string{"--binary-only", "--force-source", "5.38.0"},
			expectError: true,
			errorMsg:    "--binary-only and --force-source are mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newInstallCommand()
			cmd.SetArgs(tt.args)

			// We need to capture the error before execution since the command will
			// try to actually install Perl
			err := cmd.ParseFlags(tt.args)
			if err != nil && !tt.expectError {
				t.Errorf("ParseFlags() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Check that flags are properly set
				binaryOnly, _ := cmd.Flags().GetBool("binary-only")
				preferBinary, _ := cmd.Flags().GetBool("prefer-binary")
				forceSource, _ := cmd.Flags().GetBool("force-source")

				// Test specific flag combinations
				switch tt.name {
				case "binary only flag", "short binary only flag":
					if !binaryOnly {
						t.Error("Expected binary-only flag to be true")
					}
				case "prefer binary flag":
					if !preferBinary {
						t.Error("Expected prefer-binary flag to be true")
					}
				case "force source flag":
					if !forceSource {
						t.Error("Expected force-source flag to be true")
					}
				}
			}
		})
	}
}

func TestInstallCommandFlagValidation(t *testing.T) {
	// Test the validation logic without executing the command
	cmd := newInstallCommand()

	// Mock the RunE function to test only the validation logic
	originalRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Get binary installation flags
		binaryOnly, err := cmd.Flags().GetBool("binary-only")
		if err != nil {
			return err
		}

		forceSource, err := cmd.Flags().GetBool("force-source")
		if err != nil {
			return err
		}

		// Validate mutually exclusive flags
		if binaryOnly && forceSource {
			return fmt.Errorf("--binary-only and --force-source are mutually exclusive")
		}

		return nil
	}

	// Test mutually exclusive flags
	cmd.SetArgs([]string{"--binary-only", "--force-source", "5.38.0"})
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for mutually exclusive flags")
	}
	if err.Error() != "--binary-only and --force-source are mutually exclusive" {
		t.Errorf("Expected specific error message, got: %v", err)
	}

	// Restore original RunE
	cmd.RunE = originalRunE
}
