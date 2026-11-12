// ABOUTME: Writes perl's oracle facts over the oracle-graded corpora to <outdir>/{in-tree,t1,perlgit}.txt; run by oracle-facts.sh.
// ABOUTME: Offline only: make test never runs it. usage: oracle <perl> <perl5-tree> <outdir>, from the repository root, T1_CORPUS set.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"tamarou.com/pvm/internal/chart"
	"tamarou.com/pvm/internal/parseoracle"
)

// timeout bounds one file. The CHECK walk costs 32 s on op/taint.t idle, and
// t/re/pat_psycho.t wedges for its whole budget.
const timeout = 180 * time.Second

// shim is the temporary shim tree, removed on every exit.
var shim string

func die(format string, a ...any) {
	os.RemoveAll(shim)
	fmt.Fprintf(os.Stderr, "oracle-facts: "+format+"\n", a...)
	os.Exit(1)
}

func out(name string, args ...string) string {
	b, err := exec.Command(name, args...).Output()
	if err != nil {
		die("%s %s: %v", name, strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(b))
}

// group is one facts file: the files to ask perl about, from which directory,
// under which PERL5LIB, keyed by key(file).
type group struct {
	file, corpus, cwd, root, lib string
	dir                          string
	paths                        []string
	key                          func(string) string
}

func main() {
	if len(os.Args) != 4 {
		die("usage: oracle <perl> <perl5-tree> <outdir>")
	}
	perl, tree, outdir := os.Args[1], os.Args[2], os.Args[3]
	t1 := os.Getenv("T1_CORPUS")
	config := func(k string) string { return out(perl, "-MConfig", "-e", "print $Config{"+k+"}") }
	commit := config("git_commit_id")

	// perl.git t/ compiles only from inside a shim whose lib/ holds the perl's own library.
	var err error
	if shim, err = os.MkdirTemp("", "oracle-shim-"); err != nil {
		die("%v", err)
	}
	defer os.RemoveAll(shim)
	if err := parseoracle.BuildShim(perl, tree, shim, out("git", "-C", tree, "rev-parse", "HEAD")); err != nil {
		die("building the shim: %v", err)
	}

	acceptedLabels := func(name string) []string {
		lf, err := chart.ReadLabels(filepath.Join("internal", "chart", "testdata", "labels", name))
		if err != nil {
			die("%v", err)
		}
		var p []string
		for _, l := range lf.Labels {
			if !l.Refused() {
				p = append(p, l.Path)
			}
		}
		return p
	}
	// accepted is every key the labels accept; perl.git t/'s facts are taken over every file, refused ones too.
	accepted := map[string]bool{}
	for name, prefix := range map[string]string{"in-tree.txt": "", "t1.txt": "T1/", "perlgit.txt": "perlgit/"} {
		for _, p := range acceptedLabels(name) {
			accepted[prefix+p] = true
		}
	}
	var perlgit []string
	filepath.WalkDir(filepath.Join(tree, "t"), func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".t") {
			rel, _ := filepath.Rel(filepath.Join(tree, "t"), p)
			perlgit = append(perlgit, rel)
		}
		return nil
	})
	t1top := out("git", "-C", t1, "rev-parse", "--show-toplevel")
	t1root := strings.TrimSuffix(out("git", "-C", t1, "rev-parse", "--show-prefix"), "/")
	t1names := acceptedLabels("t1.txt")
	t1paths := make([]string, len(t1names))
	for i, n := range t1names {
		t1paths[i] = t1root + "/" + n
	}
	groups := []group{
		{file: "in-tree.txt", corpus: "pvm, paths from the repository root", cwd: ".", root: ".",
			lib: "internal/chart/testdata/guacamole/stubs", paths: acceptedLabels("in-tree.txt"),
			key: func(p string) string { return p }},
		{file: "t1.txt", corpus: "PerlOnJava " + out("git", "-C", t1, "rev-parse", "--short=8", "HEAD"), cwd: t1top, root: t1root,
			dir: t1top, paths: t1paths, key: func(p string) string { return "T1/" + strings.TrimPrefix(p, t1root+"/") }},
		{file: "perlgit.txt", corpus: "perl.git t/ at " + commit[:8] + ", every .t file, compiled from a shim t/", cwd: "shim t/ (perl.git t/ beside a lib/ of the perl's @INC)",
			root: ".", dir: filepath.Join(shim, "t"), paths: perlgit, key: func(p string) string { return "perlgit/" + p }},
	}

	for _, g := range groups {
		// A perl run inherits this environment, so it holds what label.sh's env -i does.
		path, tmp := os.Getenv("PATH"), os.Getenv("TMPDIR")
		os.Clearenv()
		os.Setenv("PATH", path)
		if tmp != "" {
			os.Setenv("TMPDIR", tmp)
		}
		os.Setenv("PERL5LIB", g.lib)
		// The subject is a no-op: Run is the runner that handles -T shebangs,
		// the working directory and the per-file timeout, and keeps perl's facts.
		report, err := parseoracle.Run(context.Background(), g.paths, parseoracle.RunOptions{
			Dir: g.dir, Timeout: timeout, Perl: perl, Subject: &parseoracle.Subject{Command: []string{"true"}}})
		if err != nil {
			die("%v", err)
		}
		var b strings.Builder
		fmt.Fprintf(&b, "# perl: %s\n# version: %s\n# git_commit_id: %s\n# ccflags: %s\n# corpus: %s\n# cwd: %s\n# root: %s\n",
			perl, config("version"), commit, config("ccflags"), g.corpus, g.cwd, g.root)
		fmt.Fprintf(&b, "# PERL5LIB: %s\n# environment: PATH, TMPDIR and PERL5LIB only\n# timeout: %.0fs per file; a timeout is recorded as none\n",
			orEmpty(g.lib), timeout.Seconds())
		b.WriteString("# fields: ok, walked, op_count, sites and prototypes, as parseoracle.Facts holds them; the op sequence is not kept\n")
		bad := 0
		for _, r := range report.Files {
			rec := chart.OracleRecord{Key: g.key(r.Path), Facts: r.Facts}
			if r.Err != "" {
				rec.Missing = r.Err
			}
			// shortcut: a label-accepted file with no ok facts is reported, not refused (T1/cpan_tooling_regressions.t and
			// perlgit/re/pat_psycho.t today, pinned by TestOracleFactsLabelAcceptedWithoutOKFacts); refuse here once each is fixed.
			if accepted[rec.Key] && !r.Facts.OK {
				bad++ // a label accepted it and perl's facts did not: reported, still recorded
			}
			line, err := chart.FormatOracleRecord(rec)
			if err != nil {
				die("%s: %v", rec.Key, err)
			}
			b.WriteString(line + "\n")
		}
		if bad > 0 {
			fmt.Fprintf(os.Stderr, "oracle-facts: %s: %d files the labels accept have no ok facts\n", g.file, bad)
		}
		dst := filepath.Join(outdir, g.file)
		if err := os.WriteFile(dst+".tmp", []byte(b.String()), 0o644); err != nil || os.Rename(dst+".tmp", dst) != nil {
			die("writing %s: %v", dst, err)
		}
		fmt.Fprintf(os.Stderr, "oracle-facts: %s: %d records (%d workers)\n", dst, len(report.Files), runtime.NumCPU())
	}
}

func orEmpty(s string) string {
	if s == "" {
		return "(empty)"
	}
	return s
}
