//go:build precedencereference

// ABOUTME: The perl.git precedence differential (decision §3.4) and set-1626 observer over every t/ file the ratchet loads.
// ABOUTME: make chart-precedence-reference runs it, out of go test ./... because CI's other jobs have no perl.git.

package main

import (
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/chart/prectab"
	"tamarou.com/pvm/internal/chart/recognize"
)

// TestPrecedencePerlGitMatchesReference: every perl.git t/ file the chart recognizes, loaded as the ratchet loads it
// (perlGit refuses a tree that is missing, unclean or not at perlGitPin), is compared as comparePrecedence compares the
// in-tree corpora. A missing perl.git fails, and so does one that contributes no compared file.
func TestPrecedencePerlGitMatchesReference(t *testing.T) {
	t.Chdir("../../../..")
	srcs := map[string][]byte{}
	files, err := perlGit(perl5Root(), srcs)
	if err != nil {
		t.Fatal(err)
	}
	tab := committedTables(t)
	prec, err := prectab.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	perl, err := labelVerdicts(filepath.Join(testdata, "labels"))
	if err != nil {
		t.Fatal(err)
	}
	scope := chartInputs(tab, srcs, nil, perl, realDirs).Scope
	r := recognize.NewNoPrecedence(tab, recognize.Options{})
	var compared, prelib, precedence, pruned int
	for _, file := range files {
		o, err := comparePrecedence(tab, r, prec, srcs[file.Path], scope(file.Path))
		switch {
		case err != nil:
			t.Fatalf("%s: %v", file.Path, err)
		case !o.recognized:
			continue
		case o.prelibRefused:
			prelib++
			continue
		case o.diff != "":
			t.Errorf("%s: %s", file.Path, o.diff)
		}
		compared++
		if !o.accepted {
			precedence++
		}
		if o.pruned {
			pruned++
		}
	}
	t.Logf("perlgit files=%d compared=%d prelib-refused=%d precedence-refused=%d pruned=%d", len(files), compared, prelib, precedence, pruned)
	if compared == 0 {
		t.Error("perl.git contributed no compared file")
	}
}

// TestSet1626NeverReachedPerlGit: TestSet1626NeverReachedCorpus's observation over every perl.git t/ file, loaded as
// the ratchet loads it. A missing perl.git fails, and so does one that observes no completion of the rule.
func TestSet1626NeverReachedPerlGit(t *testing.T) {
	t.Chdir("../../../..")
	srcs := map[string][]byte{}
	files, err := perlGit(perl5Root(), srcs)
	if err != nil {
		t.Fatal(err)
	}
	o := newPairObserver(t, committedTables(t), "indirob ::= scalar", "HASHBRACK", "PERLY_BRACE_OPEN")
	for _, f := range files {
		if _, err := o.observe(srcs[f.Path]); err != nil {
			t.Fatalf("%s: %v", f.Path, err)
		}
	}
	t.Logf("perlgit files: %d tokens carrying set 1626: %d indirob ::= scalar completions: %d reaching set 1626: %d",
		o.inputs, o.carrying, o.completions, o.reaching)
	if o.inputs == 0 || o.completions == 0 {
		t.Errorf("observed nothing: %d files, %d completions", o.inputs, o.completions)
	}
	if o.reaching != 0 {
		t.Errorf("set 1626 reached %d times after indirob ::= scalar: a design §7 stop", o.reaching)
	}
}

// TestForestEqualityPerlGit: TestForestEqualityCorpus's differential over every perl.git t/ file, loaded as the ratchet
// loads it. A missing perl.git fails, and so does one that compares no file.
func TestForestEqualityPerlGit(t *testing.T) {
	t.Chdir("../../../..")
	srcs := map[string][]byte{}
	files, err := perlGit(perl5Root(), srcs)
	if err != nil {
		t.Fatal(err)
	}
	tab := committedTables(t)
	compared, diffs := compareForests(tab, files, srcs, fileScope(t, tab, srcs))
	reportForests(t, "perl.git", len(files), compared, diffs)
}
