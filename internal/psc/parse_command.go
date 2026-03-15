// ABOUTME: The "psc parse" command parses a Perl file and displays the AST.
// ABOUTME: Supports tree (default) and sexpr output formats via --format flag.

package psc

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"tamarou.com/pvm/internal/parser"
)

func newParseCommand() *cobra.Command {
	var format string

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

			p := parser.New()
			tree, err := p.Parse(source)
			if err != nil {
				return fmt.Errorf("parse %s: %w", filename, err)
			}

			root := tree.RootNode()
			if root == nil {
				return fmt.Errorf("parse %s: empty tree", filename)
			}

			w := cmd.OutOrStdout()

			switch strings.ToLower(format) {
			case "sexpr":
				fmt.Fprintln(w, root.SExpr())
			default:
				printTree(w, root, source, "", true)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "tree", "output format: tree or sexpr")

	return cmd
}

// printTree prints an indented tree representation of the syntax tree node.
func printTree(w io.Writer, n *parser.Node, source []byte, indent string, isLast bool) {
	if n == nil {
		return
	}

	branch := "├── "
	childIndent := indent + "│   "
	if isLast {
		branch = "└── "
		childIndent = indent + "    "
	}

	kind := n.Kind()
	named := ""
	if !n.IsNamed() {
		named = " (anonymous)"
	}

	nodeText := ""
	if n.ChildCount() == 0 {
		text := n.Text(source)
		// Truncate long leaf text for display
		if len(text) > 40 {
			text = text[:37] + "..."
		}
		nodeText = " " + fmt.Sprintf("%q", text)
	}

	errorMark := ""
	if n.IsError() {
		errorMark = " [ERROR]"
	} else if n.HasError() {
		errorMark = " [has error]"
	}

	fmt.Fprintf(w, "%s%s%s%s%s%s\n", indent, branch, kind, named, nodeText, errorMark)

	count := n.ChildCount()
	for i := 0; i < count; i++ {
		child := n.Child(i)
		printTree(w, child, source, childIndent, i == count-1)
	}
}
