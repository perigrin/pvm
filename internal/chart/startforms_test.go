// ABOUTME: Tests perly.y's seven start forms: each is a start rule, and each recognizes input behind its GRAM* selector.
// ABOUTME: The program form is graded as files are; the others only on input that lexes the same from statement start.

package chart

import (
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// recognizeSelected recognizes src through chart/lex behind the start selector,
// a token perl's parser pushes before the first one it lexes and the lexer never
// emits; "" selects the program form, which has none.
func recognizeSelected(t *testing.T, selector string, src string) bool {
	t.Helper()
	tab, err := committed()
	if err != nil || tab == nil {
		t.Fatalf("the committed grammar generates no tables: %v", err)
	}
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	if selector != "" {
		sel := slices.IndexFunc(tab.Symbols, func(y grammar.TableSymbol) bool { return y.Terminal && y.Name == selector })
		if sel < 0 {
			t.Fatalf("%s is not a terminal of the committed grammar", selector)
		}
		selected := *tab
		selected.TerminalSets = append(slices.Clone(tab.TerminalSets), []int{sel})
		s.Tokens = append([]lex.Token{{TerminalSet: len(tab.TerminalSets)}}, s.Tokens...)
		tab = &selected
	}
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), lex.Linear(s))
	if err != nil {
		t.Fatal(err)
	}
	return c.Accepted()
}

// TestStartForms holds perly.y's seven start forms (perly.y:187-268) as start
// rules. Waypoint 1 grades the program form, the one a file is recognized by.
// The other six are tested only on input that lexes the same from statement
// start as it would where perl selects them: internal/lexer exports only
// statement-start states, so input whose tokens depend on another start
// state (an expression's leading `{`, a signature's sigils) is not tested
// here. Every input accepted below is accepted by the pinned perl as a program.
func TestStartForms(t *testing.T) {
	// Marked lines are cited by a rule (TestPortLedger), so these are ported.
	if err := grammar.CheckFamilyMarked([]int{187, 198, 208, 221, 234, 247, 257}, nil); err != nil {
		t.Error(err)
	}

	for _, c := range []struct {
		selector       string
		accept, refuse []string
	}{
		{"", []string{"", "use warnings;\n$x = 1;\n", "{ $x = 1 }\n$y = 2"}, []string{"use ;", "$x = 1 $y = 2"}},
		{"GRAMEXPR", []string{"", "$x = 1", "$x, $y"}, []string{"$x = 1;", "use strict;"}},
		{"GRAMBLOCK", []string{"{ $x = 1; }", "{ use strict; }"}, []string{"", "$x = 1;"}},
		{"GRAMBARESTMT", []string{"$x = 1;", "use strict;", "{ $x = 1 }", ";"}, []string{"$x = 1; $y = 2;"}},
		{"GRAMFULLSTMT", []string{"$x = 1;", "no warnings;"}, []string{"", "$x = 1; $y = 2;"}},
		{"GRAMSTMTSEQ", []string{"", "use strict;\nno warnings;\n"}, []string{"use strict; {"}},
		{"GRAMSUBSIGNATURE", []string{""}, []string{";", "$x = 1;"}},
	} {
		for _, src := range c.accept {
			if !recognizeSelected(t, c.selector, src) {
				t.Errorf("%s: %q is refused", c.selector, src)
			}
		}
		for _, src := range c.refuse {
			if recognizeSelected(t, c.selector, src) {
				t.Errorf("%s: %q is accepted", c.selector, src)
			}
		}
	}
}
