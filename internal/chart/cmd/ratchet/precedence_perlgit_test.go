//go:build precedencereference

// ABOUTME: The perl.git precedence differential (decision §3.4): production Constrain against ConstrainReference on every recognized t/ file.
// ABOUTME: make chart-precedence-reference runs it, out of go test ./... because CI's other jobs have no perl.git.

package main

import (
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/chart/forest"
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
	prec, err := forest.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	perl, err := labelVerdicts(filepath.Join(testdata, "labels"))
	if err != nil {
		t.Fatal(err)
	}
	scope := chartInputs(tab, srcs, nil, perl, realDirs).Scope
	r := recognize.New(tab, recognize.Options{})
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
