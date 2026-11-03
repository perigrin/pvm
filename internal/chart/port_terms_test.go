// ABOUTME: Tests the remaining term rules: anonymous arrays and hashes, do, the ternary, refgen, local, parentheses, qw lists and &-calls.
// ABOUTME: Snippets under testdata/port/terms are recognized through internal/lexer, chart/lex and the recognizer; the -refused ones refuse.

package chart

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// termsLines are the perly.y alternatives this family ports or fits; 1536 is fitted without the ;
// before the } that toke.c inserts. None is left for the completing issue.
var termsLines = []int{1534, 1536, 1558, 1560, 1566, 1567, 1568, 1570, 1574, 1576, 1578, 1580, 1638, 1641}

func TestPortTerms(t *testing.T) {
	if err := grammar.CheckFamilyMarked(termsLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "terms") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
	// The bare-term statements, which have no refused prefix for the truncation test to cut at.
	for _, src := range []string{"[1, 2];", "+{ foo => 1 };", "do { 1 };", "$c ? 1 : 2;", "\\$x;", "local $x;", "(1);", "();", "qw(a b);", "&foo();"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d (expected %v)", src, v.at, v.expected)
		}
	}
}

func TestPortTermsRefusals(t *testing.T) {
	rows := map[string]struct {
		at        string // the refused byte is the first occurrence of this text; "" is end of file
		neighbour string // an accepted form beside it
	}{
		"unclosed-array.pl":         {";", "[1, 2];"},
		"ternary-no-colon.pl":       {";", "$c ? 1 : 2;"},
		"ternary-no-middle.pl":      {":", "$c ? 1 : 2;"},
		"refgen-no-operand.pl":      {";", "\\$x;"},
		"hash-written-semicolon.pl": {";", "+{ foo => 1 };"},
		"array-two-terms.pl":        {"2", "$y = [1, 2];"},
		"do-no-operand.pl":          {";", "$y = do { 1 };"},
		"local-no-operand.pl":       {";", "$y = local $x;"},
		"unclosed-paren.pl":         {";", "$y = (1, 2);"},
		"unclosed-call.pl":          {";", "$y = &foo(1);"},
		"qwlist-two-terms.pl":       {"qw(b)", "$y = qw(a);"},
	}
	for name, c := range rows {
		src, err := os.ReadFile(filepath.Join("testdata", "port", "terms-refused", name))
		if err != nil {
			t.Fatal(err)
		}
		want := len(src)
		if c.at != "" {
			want = strings.Index(string(src), c.at)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", name)
		} else if v.at != want {
			t.Errorf("%s is refused at byte %d, want %d", name, v.at, want)
		}
		if v := recognizeSource(t, []byte(c.neighbour)); !v.accepted {
			t.Errorf("%s's accepted neighbour %q is refused at %d", name, c.neighbour, v.at)
		}
	}
	if len(snippets(t, "terms-refused")) != len(rows) {
		t.Error("a refused snippet has no row here")
	}
}

// terminalsOf is the names of the terminals admitting the token of src that starts with text.
func terminalsOf(t *testing.T, src, text string) []string {
	t.Helper()
	tab, err := committed()
	if err != nil || tab == nil {
		t.Fatalf("the committed grammar generates no tables: %v", err)
	}
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	for _, tok := range s.Tokens {
		if raw := s.Raw[tok.Index]; strings.HasPrefix(src[raw.Start:], text) && raw.End-raw.Start == len(text) {
			var names []string
			for _, id := range tab.TerminalSets[tok.TerminalSet] {
				names = append(names, tab.Symbols[id].Name)
			}
			return names
		}
	}
	t.Fatalf("no token %q in %q", text, src)
	return nil
}

// A word inside +{ ... } keeps the call reading its terminals give it, and the brace itself is the
// anonymous hash's, which a subscript's is not; so the two braces key the word after them differently.
func TestPortTermsHashWordIsCall(t *testing.T) {
	for _, src := range []string{"$y = +{ time };", "$y = +{ time => 1 };"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d (expected %v)", src, v.at, v.expected)
		}
	}
	if got := terminalsOf(t, "$y = +{ time };", "time"); !slices.Contains(got, "FUNC0") {
		t.Errorf("time inside +{ } admits %v, want the FUNC0 call", got)
	}
	if got := terminalsOf(t, "$y = +{ time };", "{"); !slices.Equal(got, []string{"HASHBRACK"}) {
		t.Errorf("the { of +{ } admits %v, want HASHBRACK", got)
	}
	for _, src := range []string{"$h{time};", "$h->{time};", "$h[0]{time};", "$h{a}{time};", "f(){time};"} {
		if got := terminalsOf(t, src, "{"); slices.Contains(got, "HASHBRACK") {
			t.Errorf("the { of %q admits %v, a subscript's brace is not an anonymous hash's", src, got)
		}
	}
}
