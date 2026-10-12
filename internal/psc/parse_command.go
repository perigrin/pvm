// ABOUTME: The "psc parse" command parses a Perl file and displays internal/parse's CST or the lowered tree.
// ABOUTME: --tree picks cst (default) or lowered; --format picks tree (default) or sexpr for either.

package psc

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

func newParseCommand() *cobra.Command {
	var format, tree string

	cmd := &cobra.Command{
		Use:   "parse <file>",
		Short: "Parse a Perl file and display its syntax tree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := args[0]

			source, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("read %s: %w", filename, err)
			}

			root := parse.Parse(source)
			sexpr := strings.ToLower(format) == "sexpr"
			w := cmd.OutOrStdout()

			switch tree {
			case "cst":
				if sexpr {
					fmt.Fprintln(w, cstSexpr(root))
				} else {
					printCST(w, root, 0)
				}
			case "lowered":
				file := lower.Lower(root)
				if sexpr {
					fmt.Fprintln(w, lower.Sexpr(file))
				} else {
					fmt.Fprint(w, lower.Tree(file))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "tree", "output format: tree or sexpr")
	cmd.Flags().StringVar(&tree, "tree", "cst", "which tree: cst (internal/parse) or lowered")

	return cmd
}

// printCST prints n and its subtree as an outline, two spaces a level.
func printCST(w io.Writer, n *parse.Node, depth int) {
	fmt.Fprintf(w, "%s%s\n", strings.Repeat("  ", depth), cstLabel(n))
	for _, c := range n.Children {
		printCST(w, c, depth+1)
	}
}

// cstSexpr renders n and its subtree on one line, as printCST does.
func cstSexpr(n *parse.Node) string {
	var b strings.Builder
	b.WriteString("(" + cstLabel(n))
	for _, c := range n.Children {
		b.WriteString(" " + cstSexpr(c))
	}
	b.WriteString(")")
	return b.String()
}

// cstLabel is a node's kind, its Text, the flags set on it, and an
// Unknown's refusal code.
func cstLabel(n *parse.Node) string {
	parts := []string{n.Kind.String()}
	if n.Text != "" {
		parts = append(parts, fmt.Sprintf("%q", n.Text))
	}
	for _, f := range []struct {
		name string
		set  bool
	}{
		{"resolved", n.Resolved}, {"arrow", n.Arrow}, {"paren", n.Paren},
		{"fat", n.Fat}, {"handle", n.Handle}, {"comparator", n.Comparator},
		{"indirect", n.Indirect}, {"modifier", n.Modifier},
		{"heredoc_body", n.HeredocBody}, {"loop_var", n.LoopVar},
		{"cstyle", n.CStyle}, {"ref_slot", n.RefSlot},
	} {
		if f.set {
			parts = append(parts, f.name)
		}
	}
	if n.Refusal != "" {
		parts = append(parts, "refusal="+string(n.Refusal))
	}
	return strings.Join(parts, " ")
}
