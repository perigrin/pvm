// ABOUTME: The "psc check" command type-checks Perl files for diagnostics.
// ABOUTME: Reports arity and type errors found by the inference engine to stderr.

package psc

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/perl"
)

func newCheckCommand() *cobra.Command {
	var strict, warnLossy bool

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

			opts := infer.Options{Strict: strict, WarnLossyCoercion: warnLossy}
			sess := parse.NewSession()
			roots, why := incRoots(target)
			if why != "" {
				fmt.Fprintf(errW, "psc: note: %s; @INC is PERL5LIB alone\n", printable(why))
			}
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
						had, checkErr := checkFile(errW, sess, roots, path, opts)
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
				had, checkErr := checkFile(errW, sess, roots, target, opts)
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

	cmd.Flags().BoolVar(&strict, "strict", false,
		"report values whose type could not be inferred instead of accepting them")
	cmd.Flags().BoolVar(&warnLossy, "warn-lossy-coercion", false,
		"report coercions perl makes silently that lose the value, such as a reference printed as a string")

	return cmd
}

// incRoots is the @INC psc check parses each file against, so its use
// statements resolve as perl compiles them: PERL5LIB, then the library
// directories of the perl target's project names (perl.StaticPerl). It is
// built once per run, and never by running perl or asking the network:
// psc check reads untrusted code, running perl runs whatever PERL5OPT or
// sitecustomize.pl says, and a version file must not make the checker
// phone home. With no resolved perl it is PERL5LIB alone, and with no
// PERL5LIB either it is empty, and each file is parsed against its own
// directory alone. why says why a perl the project names was not
// resolved, and is "" when one was or the project names none.
func incRoots(target string) (roots []string, why string) {
	projectDir := target
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		projectDir = filepath.Dir(target)
	}
	perlPath, why := perl.StaticPerl(projectDir)
	return perl.INC(perlPath, os.Getenv("PERL5LIB")), why
}

// checkFile parses path against its own directory and roots, lowers and
// type-checks it, printing any diagnostics to w. sess shares what each
// module says about itself across the run's files. It returns true if any
// diagnostics were found.
func checkFile(w interface{ Write([]byte) (int, error) }, sess *parse.Session, roots []string, path string, opts infer.Options) (bool, error) {
	source, err := parse.ReadSource(path)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", path, err)
	}

	// Parsed from the bytes read above, so each diagnostic's line and
	// column are found in the text its span was taken from.
	root := sess.ParseSourceFrom(path, source, roots...)

	diags := analyze(root, opts).Diagnostics

	for _, d := range diags {
		fmt.Fprintln(w, infer.FormatDiagnostic(printable(path), source, d))
	}

	return len(diags) > 0, nil
}

// printable is s with each control character, and each byte that is not
// UTF-8, written as a \xNN escape: a file name in a cloned repository is
// untrusted, and an escape sequence in one must not reach the terminal.
func printable(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			fmt.Fprintf(&b, "\\x%02x", s[i])
		case unicode.IsControl(r):
			fmt.Fprintf(&b, "\\x%02x", r)
		default:
			b.WriteString(s[i : i+size])
		}
		i += size
	}
	return b.String()
}

// analyze lowers and type-checks root, and reports every refusal in it
// beside what inference reports, in source order. psc check and the LSP
// both analyse through it, so they report the same.
func analyze(root *parse.Node, opts infer.Options) *infer.FileAnalysis {
	fa := infer.AnalyzeParsed(root, nil, opts)
	fa.Diagnostics = append(refusals(root), fa.Diagnostics...)
	slices.SortStableFunc(fa.Diagnostics, func(a, b infer.Diagnostic) int {
		return int(a.StartByte) - int(b.StartByte)
	})
	return fa
}

// refusals reports each Unknown in n's tree: source the parser declines
// because perl refuses to compile it, as its refusal code over its span. An
// Unknown inside another is part of the outer refusal and is not reported
// again. Inference types the Hole an Unknown lowers to as Any and says
// nothing about it, so this is the only report it gets.
func refusals(n *parse.Node) []infer.Diagnostic {
	if n.Kind == parse.Unknown {
		return []infer.Diagnostic{{
			StartByte: uint32(n.Start),
			EndByte:   uint32(n.End),
			Severity:  infer.Error,
			Code:      string(n.Refusal),
			Message:   "perl refuses this: " + parse.RefusalSites[n.Refusal].What,
		}}
	}
	var out []infer.Diagnostic
	for _, c := range n.Children {
		out = append(out, refusals(c)...)
	}
	return out
}
