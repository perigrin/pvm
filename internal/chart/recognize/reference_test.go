// ABOUTME: A textbook Earley recognizer, test-only, and the oracle that holds the recognizer to it.
// ABOUTME: It shares no code with the recognizer; the oracle compares acceptance and Expected under every Options combination.

package recognize

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// refItem is a rule with a dot and the set it began in.
type refItem struct{ rule, dot, origin int }

// refRun is the reference's Earley sets for the terminal-set ids ids, one per
// node of a linear lattice: sets[i] holds what is true after i tokens; and
// whether it accepts. It reads tab's rules and terminal sets only, never its
// states. Each set is closed by predicting and completing until nothing is
// added, so a nullable symbol needs no special case. Where the next token
// holds PERLY_BRACE_CLOSE, or the input ends, a ; toke.c inserts is scanned
// into a set of its own past the node's (spec §5.2); the next node scans its
// token from both, and at the end either may accept. inserted is the nodes
// where such a ; was scanned, ascending.
func refRun(tab *grammar.Tables, ids []int) (sets [][]refItem, inserted []int, accepted bool) {
	var all [][]refItem // every set, a node's or an inserted ;'s; origins index it
	open := func(seed []refItem) []refItem {
		k := len(all)
		all = append(all, nil)
		seen := map[refItem]bool{}
		put := func(it refItem) {
			if !seen[it] {
				seen[it] = true
				all[k] = append(all[k], it)
			}
		}
		for _, it := range seed {
			put(it)
		}
		for grew := true; grew; {
			grew = false
			n := len(all[k])
			for i := 0; i < len(all[k]); i++ {
				it := all[k][i]
				rule := tab.Rules[it.rule]
				if it.dot < len(rule.RHS) {
					for r, p := range tab.Rules { // predict
						if p.LHS == rule.RHS[it.dot] {
							put(refItem{r, 0, k})
						}
					}
					continue
				}
				for _, w := range all[it.origin] { // complete
					if rhs := tab.Rules[w.rule].RHS; w.dot < len(rhs) && rhs[w.dot] == rule.LHS {
						put(refItem{w.rule, w.dot + 1, w.origin})
					}
				}
			}
			grew = len(all[k]) > n
		}
		return all[k]
	}
	advance := func(from []refItem, terms []int) (out []refItem) {
		for _, it := range from {
			if rhs := tab.Rules[it.rule].RHS; it.dot < len(rhs) && slices.Contains(terms, rhs[it.dot]) {
				out = append(out, refItem{it.rule, it.dot + 1, it.origin})
			}
		}
		return out
	}
	named := func(name string) int {
		return slices.IndexFunc(tab.Symbols, func(s grammar.TableSymbol) bool { return s.Terminal && s.Name == name })
	}
	semi, brace := []int{named("PERLY_SEMICOLON")}, named("PERLY_BRACE_CLOSE")
	var seed, ins []refItem
	for r, rule := range tab.Rules {
		if rule.LHS == tab.Start {
			seed = append(seed, refItem{r, 0, 0})
		}
	}
	sets = append(sets, open(seed))
	for i := 0; ; i++ {
		ins = nil
		if i == len(ids) || slices.Contains(tab.TerminalSets[ids[i]], brace) {
			if seed := advance(sets[i], semi); len(seed) > 0 {
				ins = open(seed)
				inserted = append(inserted, i)
			}
		}
		if i == len(ids) {
			break
		}
		terms := tab.TerminalSets[ids[i]]
		sets = append(sets, open(append(advance(sets[i], terms), advance(ins, terms)...)))
	}
	done := func(it refItem) bool {
		return it.origin == 0 && it.dot == len(tab.Rules[it.rule].RHS) && tab.Rules[it.rule].LHS == tab.Start
	}
	return sets, inserted, slices.ContainsFunc(sets[len(ids)], done) || slices.ContainsFunc(ins, done)
}

// refExpected is the names of the terminals that some item of items awaits,
// sorted.
func refExpected(tab *grammar.Tables, items []refItem) []string {
	var names []string
	for _, it := range items {
		if rhs := tab.Rules[it.rule].RHS; it.dot < len(rhs) && tab.Symbols[rhs[it.dot]].Terminal {
			names = append(names, tab.Symbols[rhs[it.dot]].Name)
		}
	}
	slices.Sort(names)
	return slices.Compact(names)
}

// refAccepted is whether the reference accepts the terminal names words.
func refAccepted(tab *grammar.Tables, words string) bool {
	var ids []int
	for _, w := range strings.Fields(words) {
		ids = append(ids, termSet(tab, w))
	}
	_, _, ok := refRun(tab, ids)
	return ok
}

// TestReferenceRecognizerEdgeCases: the edges a textbook recognizer gets
// wrong. The empty input is accepted by a nullable start and refused by a
// non-nullable one, a unit cycle terminates, and left and right recursion
// accept their language and refuse a token past it.
func TestReferenceRecognizerEdgeCases(t *testing.T) {
	for _, tc := range []struct {
		name    string
		rules   []string
		accepts []string
		refuses []string
	}{
		{"nullable start", []string{"S ::= A S |"}, []string{"", "A", "A A"}, []string{"B"}},
		{"non-nullable start", []string{"S ::= A S | A"}, []string{"A", "A A"}, []string{""}},
		{"nullable through a chain", []string{"S ::= X Y", "X ::= Z", "Z ::=", "Y ::= A |"}, []string{"", "A"}, []string{"A A"}},
		{"nullable fixed point", []string{"S ::= X Y", "X ::=", "Y ::= X B"}, []string{"B"}, []string{"", "B B"}},
		{"start on a right-hand side", []string{"S ::= B S C | A"}, []string{"A", "B A C"}, []string{"B A", "B B A C"}},
		{"unit cycle", []string{"S ::= S | A"}, []string{"A"}, []string{"", "A A"}},
		{"left recursion", []string{"S ::= S A | A", "Z ::= B"}, []string{"A", "A A A"}, []string{"", "A A A B"}},
		{"right recursion", []string{"S ::= A S | A", "Z ::= B"}, []string{"A", "A A A"}, []string{"", "A A A B"}},
	} {
		tab := toyTables(t, tc.rules...)
		for _, in := range tc.accepts {
			if !refAccepted(tab, in) {
				t.Errorf("%s: the reference refused %q", tc.name, in)
			}
		}
		for _, in := range tc.refuses {
			if refAccepted(tab, in) {
				t.Errorf("%s: the reference accepted %q", tc.name, in)
			}
		}
	}
}

// declared is the names of the functions, methods and types file declares at
// package scope.
func declared(file string) (map[string]bool, error) {
	f, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			out[d.Name.Name] = true
		case *ast.GenDecl:
			for _, s := range d.Specs {
				if ts, ok := s.(*ast.TypeSpec); ok {
					out[ts.Name.Name] = true
				}
			}
		}
	}
	return out, nil
}

// usedFrom is the names in banned that src, a file called file, mentions,
// sorted.
func usedFrom(file string, src any, banned map[string]bool) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), file, src, 0)
	if err != nil {
		return nil, err
	}
	var out []string
	ast.Inspect(f, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && banned[id.Name] {
			out = append(out, id.Name)
		}
		return true
	})
	slices.Sort(out)
	return slices.Compact(out), nil
}

// TestReferenceSharesNoCode: reference_test.go mentions no function, method
// or type the package's non-test files declare, so the reference cannot
// agree with the recognizer by borrowing from it. Fed a synthetic reference
// that calls one, the check names it.
func TestReferenceSharesNoCode(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	banned := map[string]bool{}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		names, err := declared(file)
		if err != nil {
			t.Fatal(err)
		}
		for n := range names {
			banned[n] = true
		}
	}
	if !banned["Recognize"] || !banned["Expected"] {
		t.Fatalf("the check found %d declarations and not Recognize and Expected", len(banned))
	}
	got, err := usedFrom("reference_test.go", nil, banned)
	if err != nil || len(got) > 0 {
		t.Errorf("reference_test.go uses the recognizer's %v (err %v)", got, err)
	}
	got, err = usedFrom("synthetic.go", "package p\nfunc ref() { c.Accepted(); New(nil, Options{}); mine() }", banned)
	if err != nil || !slices.Equal(got, []string{"Accepted", "New", "Options"}) {
		t.Errorf("the synthetic reference's borrowings were %v (err %v), want Accepted, New and Options", got, err)
	}
}
