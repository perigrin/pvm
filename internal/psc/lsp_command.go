// ABOUTME: The "psc lsp" command starts the PSC Language Server Protocol server.
// ABOUTME: Reads JSON-RPC 2.0 messages from stdin, dispatches to handler, writes responses to stdout.

package psc

import (
	"os"

	"github.com/spf13/cobra"
)

func newLSPCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "lsp",
		Short: "Start a Language Server Protocol server",
		Long:  "Start a Language Server Protocol server on stdin/stdout. Provides diagnostics, hover, definition, and document symbols for Perl files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			server := NewLSPServer()
			tr := newTransport(os.Stdin, os.Stdout)
			h := newHandler(server, tr)
			h.exitFn = os.Exit

			for {
				msg, err := tr.readMessage()
				if err != nil {
					return err
				}
				if err := h.handle(msg); err != nil {
					return err
				}
			}
		},
	}
}
