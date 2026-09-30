// ABOUTME: Ratchets perl.git t/ walked recursively, through the LOADER-AWARE
// ABOUTME: entry point, which is the only instrument that can see require.
package parse_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// perlGitTFiles returns the perl.git t/ corpus walked RECURSIVELY, or skips
// when the checkout is absent.
//
// This is deliberately a different population from t2Files, which globs five
// named directories and finds 56 files. Walking finds 620. Conflating the two
// has already cost this project one wrong claim -- a brief asserted the T2
// metric "should move a lot" from a change that only the loader-aware parser
// could see -- so the two populations stay in separate functions with separate
// names rather than sharing a helper with a flag.
func perlGitTFiles(t *testing.T) (string, []string) {
	t.Helper()

	root := os.Getenv("PERL5_CORPUS")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("no PERL5_CORPUS and no home directory: %v", err)
		}
		root = filepath.Join(home, "dev", "perl5")
	}
	tDir := filepath.Join(root, "t")
	if _, err := os.Stat(tDir); err != nil {
		t.Skipf("perl5 corpus not present at %s: %v\n"+
			"set PERL5_CORPUS to a checkout to run the perl.git t/ ratchet", tDir, err)
	}

	var files []string
	err := filepath.Walk(tDir, func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() || !strings.HasSuffix(p, ".t") {
			return nil
		}
		rel, relErr := filepath.Rel(tDir, p)
		if relErr != nil {
			return nil
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		t.Skipf("walking %s: %v", tDir, err)
	}
	sort.Strings(files)
	return tDir, files
}

// perlGitTRoots is where perl's suite finds what it loads.
//
// Both working directories it runs from: `t/` for `require './test.pl'`, the
// tree root for `require './t/test.pl'`. And the module directories a BUILT
// perl has in lib/: test.pl's set_up_inc('../lib') puts lib/ in @INC, and the
// build copies each dual-life module there from dist/*/lib, cpan/*/lib and
// ext/*/lib. An unbuilt checkout has them only at those source paths, so
// `use Carp` and `use File::Spec::Functions` -- dist/Carp/lib and
// dist/PathTools/lib -- are searched where they are.
//
// The build also copies a distribution's top-level module and compiles its
// XS: ext/Devel-Peek/Peek.pm and Peek.xs are Devel::Peek. dist/, cpan/ and
// ext/ themselves are roots for that, where DirLoader's source-tree lookup
// finds `Devel-Peek/Peek.pm`.
func perlGitTRoots(dir string) []string {
	root := filepath.Dir(dir)
	roots := []string{dir, root, filepath.Join(root, "lib")}
	for _, tree := range []string{"dist", "cpan", "ext"} {
		libs, _ := filepath.Glob(filepath.Join(root, tree, "*", "lib"))
		roots = append(roots, libs...)
	}
	for _, tree := range []string{"dist", "cpan", "ext"} {
		roots = append(roots, filepath.Join(root, tree))
	}
	return roots
}

// TestPerlGitTRatchet holds the per-file Unknown count over perl.git t/.
//
// It calls parse.ParseFileFrom rather than parse.Parse, and that is the whole
// reason it exists. parse.Parse takes no path and no loader, so every other
// ratchet in this package is blind to require resolution: the change that took
// this corpus from 185 clean files to 278 left all six baseline files
// BYTE-IDENTICAL, because none of them could see it. A goal measured only by
// probes its author deletes is a goal nothing can regress.
//
// The count is Unknown nodes, not a rate. A rate over 620 files hides which
// file moved, and "A COUNT IS NOT A SIZE" applies here as it does to t1: a
// refusal that opens up can raise the count while the refused BYTES fall,
// because inner statements start refusing one at a time instead of vanishing
// into one swallowed span. When a count rises, explain it before regenerating.
//
// A file perl itself rejects is not counted here; see perlRejects.
func TestPerlGitTRatchet(t *testing.T) {
	dir, files := perlGitTFiles(t)
	roots := perlGitTRoots(dir)

	now := make(map[string]int, len(files))
	for _, rel := range files {
		if _, rejected := perlRejects[rel]; rejected {
			continue
		}
		n, err := parse.ParseFileFrom(filepath.Join(dir, rel), roots...)
		if err != nil {
			continue
		}
		now[rel] = countUnknown(n)
	}

	checkRatchet(t, filepath.Join("testdata", "perlgitt.ratchet"),
		"Unknown nodes per perl.git t/ file, through parse.ParseFileFrom.", now)
}

// perlRejects are the perl.git t/ files perl cannot compile, by design, with
// the line perl reports the error on. For these a refusal is the right
// answer and accepting the file would be the bug, so they leave the Unknown
// count and are held to that instead, by TestPerlGitTRejects.
//
// comp/final_line_num.t tests perl's own error reporting: it ends in `print
// 1+`, and its BEGIN __DIE__ handler prints "ok 1" only when the syntax error
// is reported on the line it recorded, 13. Without the handler, 5.42.0 dies
// "syntax error at ... at EOF" with status 255. A `perl -c` sweep of every
// file finds no other that 5.42.0 rejects outside blead-only syntax (the
// named parameters of op/signatures.t, the refaliased multi-variable foreach
// of op/for-many.t), which blead accepts.
var perlRejects = map[string]int{
	"comp/final_line_num.t": 13,
}

// TestPerlGitTRejects: each file perl rejects is refused, and first on the
// line perl reports. Accepting it, or refusing it anywhere else, fails.
func TestPerlGitTRejects(t *testing.T) {
	dir, _ := perlGitTFiles(t)
	for rel, line := range perlRejects {
		path := filepath.Join(dir, rel)
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		n, err := parse.ParseFileFrom(path, perlGitTRoots(dir)...)
		if err != nil {
			t.Fatal(err)
		}
		if got := firstUnknownLine(n, src); got != line {
			t.Errorf("%s: first refusal on line %d, perl rejects it on line %d (0 = accepted)", rel, got, line)
		}
	}
}

// firstUnknownLine is the line the first Unknown starts on, or 0.
func firstUnknownLine(n *parse.Node, src []byte) int {
	if n.Kind == parse.Unknown {
		return strings.Count(string(src[:n.Start]), "\n") + 1
	}
	for _, c := range n.Children {
		if line := firstUnknownLine(c, src); line != 0 {
			return line
		}
	}
	return 0
}

// TestPerlGitTRatchetHeaderAgreesWithItsBody reads the header back.
//
// parseRatchetFile skips every line starting with "#", so the summary line
// renderRatchet writes -- "N files, M clean (P%), T total" -- is derived on
// WRITE and never checked on READ. Nothing else in this package reads it, so
// left alone it could say 620 files and 999 clean and every ratchet test would
// still pass. That is issue 01a0ddfb, and the conformance corpus had the same
// hole until today: its header was proved unchecked by editing it to
// "999 cases, 1 clean, 998 refusing" and watching TestCorpusRatchet pass.
//
// This asserts the header against the body it sits on, so a hand-edited or
// stale summary fails rather than misinforming whoever reads the file.
func TestPerlGitTRatchetHeaderAgreesWithItsBody(t *testing.T) {
	path := filepath.Join("testdata", "perlgitt.ratchet")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("no baseline at %s yet: %v", path, err)
	}

	var files, clean, total int
	var header string
	for _, line := range strings.Split(string(data), "\n") {
		switch {
		case strings.HasPrefix(line, "# ") && strings.Contains(line, " files, "):
			header = strings.TrimPrefix(line, "# ")
		case line == "" || strings.HasPrefix(line, "#"):
		default:
			count, _, ok := strings.Cut(line, " ")
			if !ok {
				continue
			}
			n, convErr := strconv.Atoi(count)
			if convErr != nil {
				continue
			}
			files++
			total += n
			if n == 0 {
				clean++
			}
		}
	}
	if header == "" {
		t.Fatalf("%s has no summary line; renderRatchet writes one and it must be readable", path)
	}

	want := fmt.Sprintf("%d files, %d clean (%.1f%%), %d total.",
		files, clean, 100*float64(clean)/float64(files), total)
	if header != want {
		t.Errorf("%s header disagrees with its body:\n  header: %s\n  body:   %s",
			path, header, want)
	}
}

// TestPerlGitTIsTheWalkedCorpus pins the denominator, so an empty or moved
// checkout fails loudly instead of passing on nothing.
//
// TestT2CoreParses pins its 56 the same way, and that pin is the only reason a
// corpus root holding empty directories fails rather than reporting "0 of 0
// files clean (NaN%)". The question that separates a real guard from a vacuous
// one is whether anyone has made it fail on purpose; this one was checked
// against a root with t/ present and no .t files in it.
//
// It is a FLOOR rather than an exact match. perl.git is upstream and gains test
// files; an exact pin would fail on the next `git pull` and teach whoever hits
// it to edit the number without thinking. A floor fails only on LOSS.
func TestPerlGitTIsTheWalkedCorpus(t *testing.T) {
	_, files := perlGitTFiles(t)

	const floor = 620
	if len(files) < floor {
		t.Errorf("perl.git t/ walked recursively is %d files, want at least %d;"+
			" a shrinking corpus means a moved or partial checkout, not progress",
			len(files), floor)
	}
}
