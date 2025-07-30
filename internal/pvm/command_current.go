// ABOUTME: This file contains the implementation of the `pvm current` command for showing active Perl versions.
// ABOUTME: It handles various output formats including bare mode for shell scripting integration.

package pvm

import (
	"fmt"

	"github.com/spf13/cobra"

	"tamarou.com/pvm/internal/current"
)

// newCurrentCommand creates a command to show the currently active Perl version
func newCurrentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show currently active Perl version",
		Long: `Show the currently active Perl version with source attribution.

This command shows which Perl version is currently being used and where
that setting comes from (e.g., .perl-version file, environment variable,
user configuration, etc.).

The version resolution follows this precedence order:
1. Explicitly specified version
2. Environment variables (PVM_PERL_VERSION, PLENV_VERSION, PERLBREW_PERL)
3. Project-local .perl-version file
4. Project-local .pvm/pvm.toml
5. User-level configuration
6. System Perl

Examples:
  pvm current              # Show current version with source
  pvm current --bare       # Show only version (for scripting)
  pvm current --detailed   # Show comprehensive information
  pvm current --json       # Output in JSON format`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get flags
			bare, err := cmd.Flags().GetBool("bare")
			if err != nil {
				return err
			}

			detailed, err := cmd.Flags().GetBool("detailed")
			if err != nil {
				return err
			}

			jsonOutput, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			showPath, err := cmd.Flags().GetBool("path")
			if err != nil {
				return err
			}

			validate, err := cmd.Flags().GetBool("validate")
			if err != nil {
				return err
			}

			// Determine display options based on flags
			var options *current.DisplayOptions
			switch {
			case bare:
				options = current.BareDisplayOptions()
			case detailed:
				options = current.DetailedDisplayOptions()
			case jsonOutput:
				options = current.DefaultDisplayOptions()
				options.Format = current.FormatJSON
			default:
				options = current.DefaultDisplayOptions()
			}

			// Apply additional flag overrides
			if showPath {
				options.ShowPath = true
			}
			if validate {
				options.Validate = true
				options.ShowComparison = true
			}

			// Get current version information
			info, err := current.GetCurrentVersion()
			if err != nil {
				return fmt.Errorf("failed to get current version: %w", err)
			}

			// Format and display the output
			output, err := current.FormatCurrentVersion(info, options)
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			fmt.Print(output)

			// Add newline for non-bare output
			if !bare {
				fmt.Println()
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().Bool("bare", false, "Show only the version string (for scripting)")
	cmd.Flags().Bool("detailed", false, "Show comprehensive version information")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("path", false, "Include file paths in output")
	cmd.Flags().Bool("validate", false, "Validate current version and show warnings")

	return cmd
}