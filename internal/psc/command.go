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
		Short: "Perl Script Compiler - Static type checking for Perl",
		Long: `PSC (Perl Script Compiler) provides static type checking for Perl code.

It offers these key features:
• Type checking with detailed error reporting
• Type annotation stripping for compatibility
• Continuous monitoring with file watching
• Type definition management for modules
• Flow-sensitive type analysis and refinement

PSC helps catch type errors before runtime and supports gradual typing -
you can add type annotations incrementally to existing code.

Examples:
  psc check myfile.pl                  # Check a single file
  psc check --recursive lib/           # Check all files in a directory
  psc strip myfile.pl clean.pl         # Strip type annotations
  psc watch lib/                       # Watch directory for changes
  psc def generate MyModule --save     # Generate type definitions
  psc run myfile.pl                    # Type-check and run`,
	}

	// Add PSC-specific commands
	cmd.AddCommand(
		newCheckTypeCommand(), // Use the enhanced type checking command
		newStripCommand(),
		newRunCommand(),
		newWatchCommand(),
		newDefCommand(),
		newModuleCommand(), // Type-aware module management
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
