// ABOUTME: The "psc check" command type-checks Perl files for diagnostics.
// ABOUTME: Reports arity and type errors found by the inference engine to stderr.

package psc

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parser"
)

func newCheckCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check <file|directory>",
		Short: "Type-check Perl files and report diagnostics",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			errW := cmd.ErrOrStderr()

			info, err := os.Stat(target)
			if err != nil {
				return fmt.Errorf("stat %s: %w", target, err)
			}

			p := parser.New()
			var foundDiags bool

			if info.IsDir() {
				walkErr := filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.IsDir() {
						return nil
					}
					if isPerlFile(path) {
						had, checkErr := checkFile(errW, p, path)
						if checkErr != nil {
							return checkErr
						}
						if had {
							foundDiags = true
						}
					}
					return nil
				})
				if walkErr != nil {
					return walkErr
				}
			} else {
				had, checkErr := checkFile(errW, p, target)
				if checkErr != nil {
					return checkErr
				}
				foundDiags = had
			}

			if foundDiags {
				return fmt.Errorf("diagnostics found")
			}
			return nil
		},
	}

	return cmd
}

// checkFile parses and type-checks a single Perl file, printing any diagnostics
// to w. It returns true if any diagnostics were found.
func checkFile(w interface{ Write([]byte) (int, error) }, p *parser.Parser, path string) (bool, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", path, err)
	}

	tree, err := p.Parse(source)
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", path, err)
	}

	_, diags := infer.Analyze(tree, source)

	for _, d := range diags {
		fmt.Fprintln(w, infer.FormatDiagnostic(path, source, d))
	}

	return len(diags) > 0, nil
}
