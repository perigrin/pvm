// ABOUTME: Tests the per-state expected-terminal sets, Leo eligibility, the lexer probes and the conflicting pairs.
// ABOUTME: Small grammars pin each definition; the cut Guacamole grammar's pairs are held to the committed list.

package grammar

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// stateNames is each state's expected terminals, by name, in id order.
func stateNames(tab *Tables) []string {
	var out []string
	for _, s := range tab.States {
		var names []string
		for _, id := range s.Expected {
			names = append(names, tab.Symbols[id].Name)
		}
		out = append(out, strings.Join(names, " "))
	}
	return out
}

// TestExpectedTerminals: a state expects exactly the terminals after a dot
// in its items. In S ::= A S | B, state 0 predicts both alternatives, the
// kernel after A predicts S again through state 0, and complete states
// expect nothing.
func TestExpectedTerminals(t *testing.T) {
	tab, err := BuildTables(cut(t, "S ::= A S | B\nA ~ 'a'\nB ~ 'b'\n", CutSpec{}))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"A B", "", "", ""}
	if got := stateNames(tab); !slices.Equal(got, want) {
		t.Errorf("expected terminals per state = %q, want %q\n%s", got, want, tab.Dump())
	}
}

// leoStates is the item text of each Leo-eligible state's item.
func leoStates(tab *Tables) []string {
	var out []string
	for _, s := range tab.States {
		if s.Leo {
			out = append(out, tab.itemText(s.Items[0]))
		}
	}
	return out
}

// TestLeoEligibility: in S ::= A S | B the kernel state after A holds the one
// item S ::= A . S, a right-recursive completion candidate.
func TestLeoEligibility(t *testing.T) {
	tab, err := BuildTables(cut(t, "S ::= A S | B\nA ~ 'a'\nB ~ 'b'\n", CutSpec{}))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := leoStates(tab), []string{"S ::= A . S"}; !slices.Equal(got, want) {
		t.Errorf("Leo-eligible items = %q, want %q", got, want)
	}
}

// TestLeoEligibilityRefusals: a left-recursive rule's single item L ::= L . X
// has a nonterminal last, but X never derives L at its right edge; and in
// S ::= A S | A the kernel after A holds two items, S ::= A . S and S ::= A .
func TestLeoEligibilityRefusals(t *testing.T) {
	for _, src := range []string{
		"L ::= L X | Y\nX ::= C\nC ~ 'c'\nY ~ 'y'\n",
		"S ::= A S | A\nA ~ 'a'\n",
	} {
		tab, err := BuildTables(cut(t, src, CutSpec{}))
		if err != nil {
			t.Fatal(err)
		}
		if got := leoStates(tab); len(got) > 0 {
			t.Errorf("%q: Leo-eligible items = %q, want none\n%s", src, got, tab.Dump())
		}
	}
}

// toyDivRe is a toy grammar where the text / maps to two predicates: Div,
// an Operator, and Re, a Quote. The lexer reads /x/ as Div then more in
// XOperator and as one Re token in XTerm. The literal / alone is an
// unterminated regex in XTerm, which maps to no terminal and makes no pair.
const toyDivRe = "S ::= T | S Div T\nT ::= Var | Re\nDiv ~ '/'\nRe ~ '/' Body '/'\nBody ~ [a-z]+\nVar ~ 'v'\n"

var toyDivReSpec = CutSpec{Tokens: []TokenDef{{Name: "Re", Kind: "Quote"}}}

func conflictLines(cs []Conflict) []string {
	var out []string
	for _, c := range cs {
		out = append(out, c.String())
	}
	return out
}

// TestConflictsOneTextTwoPredicates: / is Div after a value and Re where a
// term starts, found by the sample /x/.
func TestConflictsOneTextTwoPredicates(t *testing.T) {
	cs := FindConflicts(cut(t, toyDivRe, toyDivReSpec))
	want := []string{`/ Div Re "/x/"`}
	if got := conflictLines(cs); !slices.Equal(got, want) {
		t.Errorf("conflicts = %q, want %q", got, want)
	}
}

// TestConflictsNoneWithoutSharedText: + and - lex the same in both states,
// so there is no conflicting pair.
func TestConflictsNoneWithoutSharedText(t *testing.T) {
	src := "S ::= T | S Plus T\nT ::= Var | Neg T\nPlus ~ '+'\nNeg ~ '-'\nVar ~ 'v'\n"
	if cs := FindConflicts(cut(t, src, CutSpec{})); len(cs) > 0 {
		t.Errorf("conflicts = %q, want none", conflictLines(cs))
	}
}

// TestProbesPerStartCharacter: a pattern terminal is probed once for each
// character its texts can start with, a literal terminal by its text.
func TestProbesPerStartCharacter(t *testing.T) {
	src := "S ::= W | L\nW ~ [ab] Tail 'c'\nTail ~ [a-z]*\nL ~ 'l'\n"
	c := cut(t, src, CutSpec{Tokens: []TokenDef{{Name: "W", Kind: "Word"}}})
	var got []string
	for _, p := range Probes(c) {
		got = append(got, fmt.Sprintf("%s %q %v", p.Terminal, p.Text, p.Sample))
	}
	want := []string{`L "l" false`, `W "axc" true`, `W "bxc" true`}
	if !slices.Equal(got, want) {
		t.Errorf("probes = %q, want %q", got, want)
	}
}

func guacamoleConflicts(t *testing.T) []Conflict {
	t.Helper()
	return FindConflicts(guacamoleCut(t))
}

func conflictTexts(cs []Conflict) []string {
	var out []string
	for _, c := range cs {
		out = append(out, c.Text)
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// TestGuacamoleConflictsExcludeLateJoiners: { (block or hash) and ? (the
// ternary or ?PATTERN?) conflict only once the ratchet adds what makes them.
func TestGuacamoleConflictsExcludeLateJoiners(t *testing.T) {
	cs := guacamoleConflicts(t)
	for _, c := range cs {
		if c.Text == "{" || c.Text == "?" {
			t.Errorf("conflict %s, want none on { or ?", c)
		}
	}
}

// TestGuacamoleConflictsIncludePerlOpTerm: the cut Guacamole grammar's pairs
// equal the committed list and include perl's operator/term texts. -X is not
// among them: perl lexes a filetest by its text, not its position.
func TestGuacamoleConflictsIncludePerlOpTerm(t *testing.T) {
	cs := guacamoleConflicts(t)
	texts := conflictTexts(cs)
	for _, want := range []string{"/", "<", "%", "&", "*", "x"} {
		if !slices.Contains(texts, want) {
			t.Errorf("no conflict named %q; the pairs found:\n%s", want, strings.Join(conflictLines(cs), "\n"))
		}
	}
	committed, err := os.ReadFile(filepath.Join(tablesDir, "conflicts.txt"))
	if err != nil {
		t.Fatalf("%v (regenerate with make chart-tables)", err)
	}
	got := strings.Join(conflictLines(cs), "\n") + "\n"
	if got != string(committed) {
		t.Errorf("%s/conflicts.txt differs from the pairs found:\n%s", tablesDir, firstDifference(tablesDir+"/conflicts.txt", "the generator", lines([]byte(got)), lines(committed)))
	}
}
