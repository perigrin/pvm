// ABOUTME: The "psc analyze" command walks Perl files and extracts dependencies.
// ABOUTME: Lists use/require statements found in each .pl/.pm/.t file.

package psc

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tamarou.com/pvm/internal/parser"
)

func newAnalyzeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze <file|directory>",
		Short: "Analyze Perl files and list their dependencies",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			w := cmd.OutOrStdout()

			info, err := os.Stat(target)
			if err != nil {
				return fmt.Errorf("stat %s: %w", target, err)
			}

			p := parser.New()

			if info.IsDir() {
				return filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.IsDir() {
						return nil
					}
					if isPerlFile(path) {
						return analyzeFile(w, p, path)
					}
					return nil
				})
			}

			return analyzeFile(w, p, target)
		},
	}

	return cmd
}

// isPerlFile reports whether the path looks like a Perl source file.
func isPerlFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".pl" || ext == ".pm" || ext == ".t"
}

// analyzeFile parses a single Perl file and prints its use/require dependencies.
func analyzeFile(w interface{ Write([]byte) (int, error) }, p *parser.Parser, path string) error {
	source, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	tree, err := p.Parse(source)
	if err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	root := tree.RootNode()
	if root == nil {
		return nil
	}

	deps := collectDependencies(root, source)

	if len(deps) > 0 {
		fmt.Fprintf(w, "%s:\n", path)
		for _, dep := range deps {
			fmt.Fprintf(w, "  %s\n", dep)
		}
	}

	return nil
}

// collectDependencies walks the AST and collects use/require dependency names.
func collectDependencies(n *parser.Node, source []byte) []string {
	if n == nil {
		return nil
	}

	var deps []string

	kind := n.Kind()
	switch kind {
	case "use_statement", "require_expression":
		// Try to extract the module name from child nodes
		for i := 0; i < n.ChildCount(); i++ {
			child := n.Child(i)
			if child == nil {
				continue
			}
			childKind := child.Kind()
			if childKind == "package" || childKind == "bareword" {
				text := strings.TrimSpace(child.Text(source))
				if text != "" && text != "use" && text != "require" {
					deps = append(deps, text)
				}
			}
		}
	}

	// Recurse into children
	for i := 0; i < n.ChildCount(); i++ {
		child := n.Child(i)
		childDeps := collectDependencies(child, source)
		deps = append(deps, childDeps...)
	}

	return deps
}
