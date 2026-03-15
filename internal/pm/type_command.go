// ABOUTME: Type command stub for PM
// ABOUTME: Returns "not yet available" since type commands require the typedef system

package pm

import (
	"fmt"

	"github.com/spf13/cobra"
)

// createTypeCommands creates stub type subcommands that report they are not yet available
func createTypeCommands(cmd *cobra.Command) {
	cmd.AddCommand(newTypeListCommand())
}

func newTypeListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List type definitions (not yet available)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("type commands are not yet available in this build")
		},
	}
}
