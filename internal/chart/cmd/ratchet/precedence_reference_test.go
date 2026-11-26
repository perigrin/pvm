// ABOUTME: The in-tree corpus differential (decision §3.4): production Constrain against ConstrainReference on every recognized file.
// ABOUTME: Each file's forest after LibraryOperators under the pipeline's own scope, verdict and every node's kept alternatives.

package main

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/prectab"
	"tamarou.com/pvm/internal/chart/recognize"
	"tamarou.com/pvm/internal/parse"
)

// precedenceOutcome is one file's comparison: whether the chart recognized it, whether LibraryOperators refused it,
// the reference's verdict, whether the reference dropped an alternative, and where production disagreed ("" for nowhere).
type precedenceOutcome struct {
	recognized, prelibRefused, accepted, pruned bool
	diff                                        string
}

// comparePrecedence judges src's forest after LibraryOperators under scope twice, with production
// Constrain on a copy and with ConstrainReference on the original, and compares their verdicts and every node's kept alternatives. Only this
// file's forests are held, and they are dropped on return.
func comparePrecedence(tab *grammar.Tables, r *recognize.Recognizer, prec *prectab.Precedence, src []byte, scope forest.Scope) (precedenceOutcome, error) {
	var o precedenceOutcome
	s, _ := lex.Lex(src, tab.Tokens)
	l := lex.Linear(s)
	c, err := recognize.Recognize(r, l)
	if err != nil {
		return o, err
	}
	f := forest.Build(tab, c, l)
	if o.recognized = f != nil; !o.recognized {
		return o, nil
	}
	if o.prelibRefused = forest.LibraryOperators(tab, f, s, src, scope) >= 0; o.prelibRefused {
		return o, nil
	}
	before, live := keptAlts(f), reachable(f)
	prod, ref := f.Copy(), f
	gotV, wantV := productionConstrain(prod, prec), ref.ConstrainReference(prec)
	o.accepted = wantV
	if gotV != wantV {
		o.diff = fmt.Sprintf("Constrain says %t, ConstrainReference %t", gotV, wantV)
		return o, nil
	}
	got, want := keptAlts(prod), keptAlts(ref)
	for n := range want {
		if !slices.Equal(got[n], want[n]) {
			o.diff = fmt.Sprintf("node %d %+v: Constrain keeps alternatives %v, ConstrainReference %v", n, ref.Nodes[n], got[n], want[n])
			return o, nil
		}
		o.pruned = o.pruned || live[n] && len(want[n]) < len(before[n])
	}
	return o, nil
}

// productionConstrain is the judgement comparePrecedence compares with the reference;
// TestPrecedenceHarnessesDoNotAlias swaps it for one that over-prunes.
var productionConstrain = (*forest.Forest).Constrain

// reachable is which of f's nodes its root reaches: both judgements drop every alternative of an unreachable node,
// so pruning counts only reachable ones.
func reachable(f *forest.Forest) []bool {
	seen := make([]bool, len(f.Nodes))
	work := []int32{f.Root}
	for len(work) > 0 {
		n := work[len(work)-1]
		work = work[:len(work)-1]
		if seen[n] {
			continue
		}
		seen[n] = true
		for a := f.Nodes[n].Alt; a != 0; a = f.Alts[a].Next {
			for _, c := range []int32{f.Alts[a].Left, f.Alts[a].Right} {
				if c != 0 {
					work = append(work, c)
				}
			}
		}
	}
	return seen
}

// keptAlts is every node's alternatives, in order.
func keptAlts(f *forest.Forest) [][]int32 {
	out := make([][]int32, len(f.Nodes))
	for n, node := range f.Nodes {
		for a := node.Alt; a != 0; a = f.Alts[a].Next {
			out[n] = append(out[n], a)
		}
	}
	return out
}

// TestPrecedenceCorpusMatchesReference: on every file of the ratchet's in-tree corpora the chart recognizes, perl-refused
// ones included, production Constrain and ConstrainReference reach the same verdict and keep the same alternatives on
// every node of the forest after LibraryOperators under the pipeline's own scope (chartInputs's, C39). A file
// LibraryOperators refuses has no constraint input and is counted, not compared. Every corpus must contribute a compared
// file, so a missing or empty one fails rather than passing on nothing.
func TestPrecedenceCorpusMatchesReference(t *testing.T) {
	t.Chdir("../../../..")
	files, srcs, err := corpus()
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
	r := recognize.New(tab, recognize.Options{})
	type counts struct{ compared, prelib, precedence, pruned, perlRefused int }
	stats := map[string]*counts{}
	for _, file := range files {
		st := stats[file.Corpus]
		if st == nil {
			st = &counts{}
			stats[file.Corpus] = st
		}
		o, err := comparePrecedence(tab, r, prec, srcs[file.Path], scope(file.Path))
		switch {
		case err != nil:
			t.Fatalf("%s: %v", file.Path, err)
		case !o.recognized:
			continue
		case o.prelibRefused:
			st.prelib++
			continue
		case o.diff != "":
			t.Errorf("%s: %s", file.Path, o.diff)
		}
		st.compared++
		if !o.accepted {
			st.precedence++
		}
		if o.pruned {
			st.pruned++
		}
		if v := perl(file.Path); v != nil && v.Refused {
			st.perlRefused++
		}
	}
	for _, name := range slices.Sorted(maps.Keys(stats)) {
		st := stats[name]
		t.Logf("corpus=%s compared=%d prelib-refused=%d precedence-refused=%d pruned=%d perl-refused-compared=%d",
			name, st.compared, st.prelib, st.precedence, st.pruned, st.perlRefused)
		if st.compared == 0 {
			t.Errorf("corpus %s contributed no compared file", name)
		}
	}
}

// TestComparePrecedenceUsesScope: comparePrecedence judges the forest LibraryOperators leaves under the scope it is
// given, so a file whose plugin operator only its own scope declares is compared under that scope and refused before
// the constraint under an empty one.
func TestComparePrecedenceUsesScope(t *testing.T) {
	tab := committedTables(t)
	prec, err := prectab.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(parse.OverlayDeclarations(os.DirFS(filepath.Join("..", "..", "testdata", "libops"))))
	src := []byte("use My::Ops; 1 plus 2 + 3;\n")
	root := parse.ParseWithLoader(src, func(string) ([]byte, bool) { return nil, false })
	r := recognize.New(tab, recognize.Options{})
	for _, c := range []struct {
		scope  forest.Scope
		prelib bool
	}{{forest.Scope{Imports: parse.Imports(root), LibraryOperators: parse.LibraryOperators(root)}, false}, {forest.Scope{}, true}} {
		o, err := comparePrecedence(tab, r, prec, src, c.scope)
		if err != nil || !o.recognized || o.prelibRefused != c.prelib || o.diff != "" {
			t.Errorf("scope %+v: %+v, %v; want recognized, refused before the constraint %t, no diff", c.scope, o, err, c.prelib)
		}
	}
}

// TestPrecedenceHarnessesDoNotAlias: comparePrecedence reports a production judgement that drops one alternative
// more than Constrain, which it could not if production judged the forest the reference judges.
func TestPrecedenceHarnessesDoNotAlias(t *testing.T) {
	tab := committedTables(t)
	prec, err := prectab.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	old := productionConstrain
	t.Cleanup(func() { productionConstrain = old })
	productionConstrain = func(f *forest.Forest, p *prectab.Precedence) bool {
		ok := f.Constrain(p)
		f.Nodes[f.Root].Alt = f.Alts[f.Nodes[f.Root].Alt].Next
		return ok
	}
	o, err := comparePrecedence(tab, recognize.New(tab, recognize.Options{}), prec, []byte("my $x = 1 + 2 * 3;\n"), forest.Scope{})
	if err != nil || !o.recognized || o.prelibRefused {
		t.Fatalf("%+v, %v; want a compared file", o, err)
	}
	if o.diff == "" {
		t.Error("comparePrecedence reports no difference for a judgement that over-prunes the root")
	}
}
