// ABOUTME: PSC-specific commands and functionality
// ABOUTME: Implements commands for Perl type checking

package psc

import (
	"fmt"

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
		newParseCommand(),     // Parse files with various output formats
		newCompileCommand(),   // Compile between different Perl variants
		newInferCommand(),     // Type inference and annotation generation
		newStripCommand(),
		newFormatCommand(), // Format code using transformation pipelines
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
		// Add MCP command (alias to pvm tool mcp)
		newPSCMCPCommand(),
	)

	return cmd
}

// newPSCMCPCommand creates an MCP command that delegates to pvm tool mcp
func newPSCMCPCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start the Model Context Protocol server",
		Long: `Start the MCP server that provides LLMs with:
- Perl code analysis using PVM's type system
- Semantic code search via embeddings
- Intelligent code generation with collaborative sampling
- Rich context awareness and project-scoped operations

This is an alias for 'pvm tool mcp' - you can use either command.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// This is a simple redirect - in a real implementation,
			// you might want to call the actual MCP server code directly
			return fmt.Errorf("Please use 'pvm tool mcp' to start the MCP server")
		},
	}
}

// Legacy command - kept for backwards compatibility but delegates to the new implementation
// The newCheckTypeCommand implementation is in check_command.go
// The newWatchCommand implementation is in watch_command.go
// The newDefCommand implementation is in def_command.go
