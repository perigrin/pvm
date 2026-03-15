// ABOUTME: The "psc lsp" command stub for future Language Server Protocol support.
// ABOUTME: Currently returns "not yet implemented" and exits cleanly.

package psc

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLSPCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "lsp",
		Short: "Start a Language Server Protocol server (not yet implemented)",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = NewLSPServer()
			return fmt.Errorf("lsp: not yet implemented")
		},
	}
}
