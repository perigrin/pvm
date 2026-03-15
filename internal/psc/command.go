// ABOUTME: Root command for psc (Perl Structural Checker) CLI tool.
// ABOUTME: Registers all psc subcommands and returns the root cobra.Command.

package psc

import "github.com/spf13/cobra"

// NewCommand creates the root psc command with all subcommands registered.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "psc",
		Short: "Perl Structural Checker",
		Long:  "psc analyses and inspects Perl source code using a pure-Go tree-sitter parser.",
	}

	cmd.AddCommand(newParseCommand())
	cmd.AddCommand(newAnalyzeCommand())
	cmd.AddCommand(newLSPCommand())

	return cmd
}
