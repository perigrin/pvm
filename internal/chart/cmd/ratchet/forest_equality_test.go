// ABOUTME: The forest-equality differential over the ratchet's in-tree corpora: the candidate's forest against V1's, file by file.
// ABOUTME: Each file is compared under its own scope; the perl.git differential behind the precedencereference tag shares compareForests.

package main

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/parse"
)

// minCorpusFiles is the floor under the in-tree corpus's size, so a truncated T1 cannot pass as the whole.
const minCorpusFiles = 1800

// forestDiffer is where the candidate's forest of src under scope differs from V1's; "" for nowhere. Both sides
// judge the same scope, which is the file's own.
func forestDiffer(tab *grammar.Tables, src []byte, scope forest.Scope) string {
	cand, _ := forest.CandidateForest(tab, src, scope)
	v1, _ := forest.V1Forest(tab, src, scope)
	return forest.ForestDiff(cand, v1)
}

// compareForests compares each file's candidate forest with V1's under scope(file), counting a file only once it
// has been compared, and returns that count with each difference named by file.
func compareForests(tab *grammar.Tables, files []File, srcs map[string][]byte, scope func(string) forest.Scope) (int, []string) {
	var diffs []string
	compared := 0
	for _, f := range files {
		d := forestDiffer(tab, srcs[f.Path], scope(f.Path))
		compared++
		if d != "" {
			diffs = append(diffs, f.Path+": "+d)
		}
	}
	return compared, diffs
}

// reportForests logs the comparison's size and fails on every difference, and on any file left uncompared.
func reportForests(t *testing.T, label string, total, compared int, diffs []string) {
	t.Helper()
	t.Logf("%s: compared %d of %d files", label, compared, total)
	if compared == 0 || compared != total {
		t.Errorf("%s: compared %d of %d files; want every one, and some", label, compared, total)
	}
	for _, d := range diffs {
		t.Errorf("%s: %s", label, d)
	}
}

// fileScope is the scope each file of srcs is judged under in the pipeline (chartInputs's, C39).
func fileScope(t *testing.T, tab *grammar.Tables, srcs map[string][]byte) func(string) forest.Scope {
	t.Helper()
	perl, err := labelVerdicts(filepath.Join(testdata, "labels"))
	if err != nil {
		t.Fatal(err)
	}
	return chartInputs(tab, srcs, nil, perl, realDirs).Scope
}

// TestForestEqualityCorpus: on every file of the ratchet's in-tree corpora, perl-refused and unrecognized ones
// included, the candidate's forest equals V1's under the file's own scope. A missing or empty corpus fails in
// corpus(), and one that compares fewer than minCorpusFiles files fails here.
func TestForestEqualityCorpus(t *testing.T) {
	t.Chdir("../../../..")
	files, srcs, err := corpus()
	if err != nil {
		t.Fatal(err)
	}
	tab := committedTables(t)
	compared, diffs := compareForests(tab, files, srcs, fileScope(t, tab, srcs))
	reportForests(t, "in-tree corpus", len(files), compared, diffs)
	if compared < minCorpusFiles {
		t.Errorf("compared %d files, want more than %d", compared, minCorpusFiles)
	}
}

// TestForestEqualityCorpusUsesFileScope: a file whose plugin operator only its own scope declares has a V1 forest
// under that scope and a refusal under an empty one, and the differential compares it under its own: with the
// candidate swapped for one that judges an empty scope, it reports the file, so the scope is passed to both sides.
func TestForestEqualityCorpusUsesFileScope(t *testing.T) {
	tab := committedTables(t)
	t.Cleanup(parse.OverlayDeclarations(os.DirFS(filepath.Join("..", "..", "testdata", "libops"))))
	src := []byte("use My::Ops; 1 plus 2 + 3;\n")
	root := parse.ParseWithLoader(src, func(string) ([]byte, bool) { return nil, false })
	own := forest.Scope{Imports: parse.Imports(root), LibraryOperators: parse.LibraryOperators(root)}
	if v1, _ := forest.V1Forest(tab, src, own); v1 == nil {
		t.Fatal("V1 refuses the file under its own scope")
	}
	if v1, _ := forest.V1Forest(tab, src, forest.Scope{}); v1 != nil {
		t.Fatal("V1 accepts the file under an empty scope, so the scope proves nothing")
	}
	if d := forestDiffer(tab, src, own); d != "" {
		t.Errorf("the candidate differs from V1 under the file's own scope: %s", d)
	}
	cand0 := forest.CandidateForest
	t.Cleanup(func() { forest.CandidateForest = cand0 })
	forest.CandidateForest = func(tab *grammar.Tables, src []byte, _ forest.Scope) (*forest.Forest, forest.Refusal) {
		return cand0(tab, src, forest.Scope{})
	}
	if forestDiffer(tab, src, own) == "" {
		t.Error("a candidate judging an empty scope is not reported, so the file's scope is not what is compared")
	}
}

// TestForestEqualityCorpusReportsDifference: compareForests reports each file on which the candidate differs, and
// counts every file it compared, so a candidate that refuses what V1 accepts fails the corpus differential.
func TestForestEqualityCorpusReportsDifference(t *testing.T) {
	tab := committedTables(t)
	srcs := map[string][]byte{"a.pl": []byte("1;\n"), "b.pl": []byte("$x = 1 + 2 * 3;\n")}
	files := []File{{"a.pl", "x"}, {"b.pl", "x"}}
	none := func(string) forest.Scope { return forest.Scope{} }
	if n, diffs := compareForests(tab, files, srcs, none); n != 2 || len(diffs) != 0 {
		t.Fatalf("equal candidate: compared %d, diffs %v; want 2, none", n, diffs)
	}
	cand0 := forest.CandidateForest
	t.Cleanup(func() { forest.CandidateForest = cand0 })
	forest.CandidateForest = func(*grammar.Tables, []byte, forest.Scope) (*forest.Forest, forest.Refusal) {
		return nil, forest.Refusal{Reason: "refuses everything"}
	}
	if n, diffs := compareForests(tab, files, srcs, none); n != 2 || len(diffs) != 2 {
		t.Errorf("refusing candidate: compared %d, diffs %v; want 2 compared and 2 reported", n, diffs)
	}
}
