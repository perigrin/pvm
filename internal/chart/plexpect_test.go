// ABOUTME: Tests the PL_expect map (testdata/pl-expect-map.txt) and the comparison of expectation classes with it.
// ABOUTME: The class at a position is the union over its Earley set, shown on a toy grammar where states disagree.

package chart

import (
	"os"
	"slices"
	"strings"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
	"tamarou.com/pvm/internal/lexer"
)

var (
	guacOnce sync.Once
	guacTab  *grammar.Tables
	guacErr  error
)

// guacamoleConflicts is the pairs the generator emits for the cut Guacamole
// grammar, the grammar in force.
func guacamoleConflicts(t *testing.T) []grammar.Conflict {
	t.Helper()
	guacOnce.Do(func() {
		guacTab, _, guacErr = recognize.Guacamole("grammar/testdata/guacamole/guacamole.slif")
	})
	if guacErr != nil {
		t.Fatal(guacErr)
	}
	return guacTab.Conflicts
}

// TestPLExpectMapCommitted: the committed map names each of the eleven
// states exactly once, and only conflicts the generator can emit.
func TestPLExpectMapCommitted(t *testing.T) {
	text, err := os.ReadFile(PLExpectMapFile)
	if err != nil {
		t.Fatal(err)
	}
	m, err := ReadPLExpectMap(PLExpectMapFile, text, guacamoleConflicts(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != len(PLExpectStates) {
		t.Errorf("%s holds %d states, want %d", PLExpectMapFile, len(m), len(PLExpectStates))
	}
}

var toyConflicts = []grammar.Conflict{
	{Text: "/", Operator: "Div", Term: "Re"},
	{Text: "<", Operator: "Lt", Term: "Diamond"},
	{Text: "<", Operator: "Lt", Term: "Heredoc"},
}

func TestReadPLExpectMapFailures(t *testing.T) {
	full := strings.Join(PLExpectStates, " /\n") + " /\n"
	for name, tc := range map[string]struct{ text, want string }{
		"missing":   {strings.Replace(full, "XREF /\n", "", 1), "XREF"},
		"duplicate": {full + "# again\nXBLOCK <\n", "duplicate"},
		"unknown":   {full + "XNOTASTATE /\n", "XNOTASTATE"},
		"bare":      {strings.Replace(full, "XREF /\n", "XREF\n", 1), "XREF"},
	} {
		_, err := ReadPLExpectMap("pl-expect-map.txt", []byte(tc.text), toyConflicts)
		if err == nil || !strings.Contains(err.Error(), tc.want) || !strings.Contains(err.Error(), "pl-expect-map.txt") {
			t.Errorf("%s: err = %v, want one naming pl-expect-map.txt and %q", name, err, tc.want)
		}
	}
	if _, err := ReadPLExpectMap("pl-expect-map.txt", []byte("# a comment\n\n"+full), toyConflicts); err != nil {
		t.Errorf("a full map: %v", err)
	}
}

// TestPLExpectUnknownConflict: a map may name a conflict the generator
// emits, or a late joiner ({ and ?), and nothing else.
func TestPLExpectUnknownConflict(t *testing.T) {
	full := strings.Join(PLExpectStates, " / { ?\n") + " / { ?\n"
	if _, err := ReadPLExpectMap("pl-expect-map.txt", []byte(full), toyConflicts); err != nil {
		t.Fatalf("emitted and late-joining conflicts: %v", err)
	}
	for _, bad := range []string{"-X", "~~", "%", ":"} {
		text := strings.Replace(full, "XSTATE / { ?", "XSTATE / "+bad, 1)
		_, err := ReadPLExpectMap("pl-expect-map.txt", []byte(text), toyConflicts)
		if err == nil || !strings.Contains(err.Error(), "XSTATE") || !strings.Contains(err.Error(), bad) {
			t.Errorf("XSTATE naming %q: err = %v, want one naming XSTATE and %q", bad, err, bad)
		}
	}
}

func TestSettled(t *testing.T) {
	for _, tc := range []struct {
		expected, want []string
	}{
		{[]string{"Div", "Lt"}, []string{"/", "<"}},
		{[]string{"Re", "Diamond", "Heredoc"}, []string{"/", "<"}},
		{[]string{"Div", "Re"}, nil},                     // both readings: open
		{[]string{"Semicolon"}, nil},                     // neither reading: not settled
		{[]string{"Div", "Diamond"}, []string{"/"}},      // < is open on its Heredoc pair
		{[]string{"Re", "Lt", "Heredoc"}, []string{"/"}}, // Lt and Heredoc both held
	} {
		if got := Settled(tc.expected, toyConflicts); !slices.Equal(got, tc.want) {
			t.Errorf("Settled(%v) = %v, want %v", tc.expected, got, tc.want)
		}
	}
}

// TestPLExpectReproduced: a state is reproduced when some position's class
// settles exactly its in-force conflicts.
func TestPLExpectReproduced(t *testing.T) {
	m := map[string][]string{"XOPERATOR": {"/", "<", "{"}, "XTERMORDORDOR": {"/"}, "XREF": {"<"}}
	classes := map[string]bool{ClassKey([]string{"/", "<"}): true, ClassKey(nil): true}
	got, counted := Unreproduced(m, toyConflicts, classes)
	if counted != 3 || !slices.Equal(got, []string{"XREF", "XTERMORDORDOR"}) {
		t.Errorf("Unreproduced = %v of %d, want [XREF XTERMORDORDOR] of 3", got, counted)
	}
	classes[ClassKey([]string{"/"})] = true
	if got, _ := Unreproduced(m, toyConflicts, classes); !slices.Equal(got, []string{"XREF"}) {
		t.Errorf("with a class settling only /: %v, want [XREF]", got)
	}
}

// TestPLExpectOutOfForceNotCounted: a state whose conflicts are all out of
// force, such as one settling only {, is neither counted nor unreproduced.
func TestPLExpectOutOfForceNotCounted(t *testing.T) {
	m := map[string][]string{"XBLOCK": {"{"}, "XSTATE": {"{", "?"}, "XOPERATOR": {"/"}}
	got, counted := Unreproduced(m, toyConflicts, map[string]bool{ClassKey([]string{"/"}): true})
	if counted != 1 || len(got) != 0 {
		t.Errorf("Unreproduced = %v of %d, want none of 1", got, counted)
	}
}

// TestClassAtPositionIsUnionOfSet: after A, one Earley set holds a state
// expecting Div (S ::= P . Div) and another expecting Re (S ::= Q . Re).
// Each alone would settle /; the class at the position leaves it open.
func TestClassAtPositionIsUnionOfSet(t *testing.T) {
	g, err := grammar.ReadSLIF("toy.slif", []byte("S ::= P Div | Q Re\nP ::= A\nQ ::= A\nA ~ 'a'\nDiv ~ '/'\nRe ~ 'r'\n"))
	if err != nil {
		t.Fatal(err)
	}
	cut, err := grammar.ApplyCut(g, grammar.CutSpec{})
	if err != nil {
		t.Fatal(err)
	}
	tab, err := grammar.BuildTables(cut)
	if err != nil {
		t.Fatal(err)
	}
	l := &lex.Lattice[lex.None]{
		Nodes: []lex.Node[lex.None]{{Offset: 0}, {Offset: 1}},
		Edges: []lex.Edge{{From: 0, To: 1, Token: lex.Token{Token: lexer.Token{Start: 0, End: 1}, Terminals: []string{"A"}}}},
	}
	// Without lookahead the stored set holds every state; Expected is the same
	// either way, and the class is read from both.
	c := recognize.Recognize(recognize.New(tab, recognize.Options{NoLookahead: true}), l)
	pair := []grammar.Conflict{{Text: "/", Operator: "Div", Term: "Re"}}

	var alone [][]string
	for _, it := range c.Sets[1].Items() {
		var exp []string
		for _, id := range tab.States[it.State].Expected {
			exp = append(exp, tab.Symbols[id].Name)
		}
		if len(exp) > 0 {
			alone = append(alone, exp)
			if got := Settled(exp, pair); !slices.Equal(got, []string{"/"}) {
				t.Errorf("state %d alone expects %v and settles %v, want [/]", it.State, exp, got)
			}
		}
	}
	if len(alone) != 2 {
		t.Fatalf("node 1's set has %d states expecting a terminal (%v), want 2", len(alone), alone)
	}
	for _, opts := range []recognize.Options{{}, {NoLookahead: true}} {
		got, err := Class(recognize.Recognize(recognize.New(tab, opts), l), 1, pair)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Errorf("%+v: the class at node 1 settles %v, want / left open", opts, got)
		}
	}
}
