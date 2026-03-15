// ABOUTME: MCP server command stub for PVM
// ABOUTME: Returns "not yet available" since MCP requires type-system components not present in this build

package pvm

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newMCPCommand creates a stub mcp subcommand that reports it is unavailable
func newMCPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start the Model Context Protocol server (not yet available)",
		Long:  `The MCP server requires type-system components that are not available in this build.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("mcp server is not yet available in this build")
		},
	}

	return cmd
}
