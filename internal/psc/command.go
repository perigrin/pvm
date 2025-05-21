// ABOUTME: PSC-specific commands and functionality
// ABOUTME: Implements commands for Perl type checking

package psc

import (
	"github.com/spf13/cobra"
)

// NewCommand creates a new PSC command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "psc",
		Short: "Perl Script Compiler",
		Long:  "Provides static type checking for Perl code",
	}

	// Add PSC-specific commands
	cmd.AddCommand(
		newCheckTypeCommand(), // Use the enhanced type checking command
		newStripCommand(),
		newRunCommand(),
		newWatchCommand(),
		newDefCommand(),
		// Add new type definition commands for PSC-PVI integration
		newGenerateTypeCommand(),
		newImportTypeCommand(),
		newListTypesCommand(),
		// Add LSP command
		lspCmd,
	)

	return cmd
}

// Legacy command - kept for backwards compatibility but delegates to the new implementation
// The newCheckTypeCommand implementation is in check_command.go
// The newWatchCommand implementation is in watch_command.go
// The newDefCommand implementation is in def_command.go
