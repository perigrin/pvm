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
				fmt.Fprintf(errW, "psc: note: %s; @INC is PERL5LIB alone\n", why)
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
	source, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", path, err)
	}

	root, err := sess.ParseFileFrom(path, roots...)
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", path, err)
	}

	diags := infer.AnalyzeParsed(root, nil, opts).Diagnostics

	for _, d := range diags {
		fmt.Fprintln(w, infer.FormatDiagnostic(path, source, d))
	}

	return len(diags) > 0, nil
}
