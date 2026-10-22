// ABOUTME: The "psc lsp" command starts the PSC Language Server Protocol server.
// ABOUTME: Reads JSON-RPC 2.0 messages from stdin, dispatches to handler, writes responses to stdout.

package psc

import (
	"os"

	"github.com/spf13/cobra"

	"tamarou.com/pvm/internal/infer"
)

func newLSPCommand() *cobra.Command {
	var strict, warnLossy, warnPossibleUndef, warnPossibleMismatch bool

	cmd := &cobra.Command{
		Use:   "lsp",
		Short: "Start a Language Server Protocol server",
		Long:  "Start a Language Server Protocol server on stdin/stdout. Provides diagnostics, hover, definition, and document symbols for Perl files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			server := NewLSPServer()
			server.Options = infer.Options{Strict: strict, WarnLossyCoercion: warnLossy, WarnPossibleUndef: warnPossibleUndef, WarnPossibleMismatch: warnPossibleMismatch}
			tr := newTransport(cmd.InOrStdin(), cmd.OutOrStdout())
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

	// The same switches as psc check's, so the editor reports what check
	// reports under them.
	cmd.Flags().BoolVar(&strict, "strict", false,
		"report values whose type could not be inferred instead of accepting them")
	cmd.Flags().BoolVar(&warnLossy, "warn-lossy-coercion", false,
		"report coercions perl makes silently that lose the value, such as a reference printed as a string")
	cmd.Flags().BoolVar(&warnPossibleUndef, "warn-possible-undef", false,
		"report an undef from a declared call's return that the call's own arguments do not prove, left silent by default")
	cmd.Flags().BoolVar(&warnPossibleMismatch, "warn-possible-mismatch", false,
		"report an argument whose union type has a member that fits the parameter beside one that does not, left silent by default")

	return cmd
}
