// ABOUTME: The "psc analyze" command walks Perl files and extracts dependencies.
// ABOUTME: Lists the modules each .pl/.pm/.t file loads, read from its lowered use, no and require statements.

package psc

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
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

			// Each file is parsed as psc check parses it.
			sess := parse.NewSession()
			roots := incRoots(target)

			if info.IsDir() {
				return filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.IsDir() {
						return nil
					}
					if isPerlFile(path) {
						return analyzeFile(w, sess, roots, path)
					}
					return nil
				})
			}

			return analyzeFile(w, sess, roots, target)
		},
	}

	return cmd
}

// isPerlFile reports whether the path looks like a Perl source file.
func isPerlFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".pl" || ext == ".pm" || ext == ".t"
}

// analyzeFile parses path against its own directory and roots, lowers it,
// and prints the modules its use, no and require statements name. sess
// shares what each module says about itself across the run's files.
func analyzeFile(w interface{ Write([]byte) (int, error) }, sess *parse.Session, roots []string, path string) error {
	root, err := sess.ParseFileFrom(path, roots...)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	deps := collectDependencies(lower.Lower(root))

	if len(deps) > 0 {
		fmt.Fprintf(w, "%s:\n", path)
		for _, dep := range deps {
			fmt.Fprintf(w, "  %s\n", dep)
		}
	}

	return nil
}

// collectDependencies lists the module each use, no and require names, in
// source order. `use VERSION`, `require VERSION` and `require EXPR` name no
// module and list nothing.
func collectDependencies(file *lower.File) []string {
	var deps []string
	for _, u := range lower.UseStmts(file) {
		if u.Module != "" {
			deps = append(deps, u.Module)
		}
	}
	return deps
}
