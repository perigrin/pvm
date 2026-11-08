// ABOUTME: Tests the subscripts: element and arrow subscripts, calls through them, list slices, and a lone word in a subscript's braces.
// ABOUTME: Snippets under testdata/port/subscripts are recognized through internal/lexer, chart/lex and the recognizer; subscripts-refused ones refuse.

package chart

import (
	"os"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// subscriptsLines are the perly.y alternatives this family ports or fits. None is left for a completing issue.
// 1349 (a glob's subscript) and 1614, 1624 (slices) are other families': they share the fitted `{ expr }`.
var subscriptsLines = []int{1353, 1356, 1361, 1366, 1369, 1373, 1377, 1383, 1391, 1398, 1404, 1406, 1408, 1592}

func TestPortSubscripts(t *testing.T) {
	if err := grammar.CheckFamilyMarked(subscriptsLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "subscripts") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortSubscriptsRefusals(t *testing.T) {
	for _, f := range snippets(t, "subscripts-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	// The accepted form beside each refusal; $h{1;} and +{ a => 1; } are refused because the ; before } is
	// toke.c's, never the source's, so no rule here expects one.
	// perl refuses these as a string terminator or not as a syntax error; the quote-like q takes the newline as its delimiter.
	for _, src := range []string{"$y = $h{q\n};", "$y = $h{s\n};", "$y = $h{ - q\n};"} {
		if v := recognizeSource(t, []byte(src)); v.accepted {
			t.Errorf("%q is accepted; perl refuses it", src)
		}
	}
	for _, src := range []string{"$a[0];", "$h{foo};", "$h->{foo};", "&f()->{foo};", "$x[0][1];", "$h{1};", "$y = +{ a => 1 };"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d, expecting %v", src, v.at, v.expected)
		}
	}
}

// toke.c's XOPERATOR { (yyl_leftcurly, toke.c:6656-6679) forces a word that stands alone before the closing } to
// BAREWORD before it looks the word up, a minus before it included, so the call reading is gone from a keyword there.
// A { that opens an anonymous hash is XTERM's and keeps it.
func TestPortSubscriptsHashWordIsString(t *testing.T) {
	for _, tc := range []struct {
		src, word string
		call      bool
	}{
		{"$y = $h{time};", "time", false},
		{"$y = $h{shift};", "shift", false},
		{"$y = $h->{time};", "time", false},
		{"$y = $h->{shift};", "shift", false},
		{"$y = $h{ shift };", "shift", false},
		{"$y = $h{if};", "if", false},
		{"$y = $h{x};", "x", false},
		{"$y = $h{-if};", "if", false},
		{"$y = $h{a}{time};", "time", false},
		{"$y = $h{\tif};", "if", false},
		{"$y = $h{ - if };", "if", false},
		// toke.c skips only spaces and tabs before the word, after a minus and before the }: a newline or a comment
		// there leaves the word to the grammar (perl calls time in the first, and refuses the keywords).
		{"$y = $h{time\n};", "time", true},
		{"$y = $h{\ntime};", "time", true},
		{"$y = $h{time # c\n};", "time", true},
		{"$y = $h{-time\n};", "time", true},
		{"$y = $h{if\n};", "if", true},
		{"$y = $h{ -\nif};", "if", true},
		{"$y = +{ time };", "time", true},
		{"$y = +{ shift };", "shift", true},
		{"$y = $h{time + 1};", "time", true},
		{"$y = $h{shift @a};", "shift", true},
		{"$y = $h{shift()};", "shift", true},
		{"$y = $h{time, 1};", "time", true},
	} {
		got := terminalsAt(t, tc.src, strings.Index(tc.src, tc.word))
		isCall := slices.ContainsFunc(got, func(n string) bool { return n != "BAREWORD" })
		if isCall != tc.call {
			t.Errorf("%q: %s admits %v, want a call reading: %v", tc.src, tc.word, got, tc.call)
		}
	}
}

// scan_word(..., FALSE) reads no package part, so a package-qualified word is not forced: the Blanks bit is clear
// on it and a Lone class cannot cut its readings. The set is the one the same word has in an anonymous hash.
func TestPortSubscriptsHashWordPackageIsNotForced(t *testing.T) {
	for _, w := range []string{"Foo::bar", "Foo'bar", "::foo", "foo::", "Foo::time"} {
		sub := terminalsAt(t, "$y = $h{"+w+"};", strings.Index("$y = $h{"+w+"};", w))
		hash := terminalsAt(t, "$y = +{"+w+"};", strings.Index("$y = +{"+w+"};", w))
		if !slices.Equal(sub, hash) {
			t.Errorf("%s: %v in a subscript, %v in an anonymous hash", w, sub, hash)
		}
	}
}

// A lone word in a subscript is a string at each position the two previous tokens decide, and the terminal set holds
// BAREWORD alone, so the key decides it and scope is not left to.
func TestPortSubscriptWordIsString(t *testing.T) {
	for _, src := range []string{"$h{foo};", "$h->{foo};", "f()->{foo};", "$x[0]{foo};", "$h{a}{foo};", "$h->[0]{foo};"} {
		if got := terminalsAt(t, src, strings.Index(src, "foo")); !slices.Equal(got, []string{"BAREWORD"}) {
			t.Errorf("%q: foo admits %v, want BAREWORD alone", src, got)
		}
	}
}

// toke.c leaves PL_expect at XOPERATOR after a subscript's ] and }, so x after them is repetition.
func TestPortSubscriptsXRule(t *testing.T) {
	for src, want := range map[string]bool{
		"$a[0] x 3;":              true,
		"$h{a} x 3;":              true,
		"$r->[0] x 3;":            true,
		"$r->{a} x 3;":            true,
		"$r->() x 3;":             true,
		"$y = $h{a} x 3;":         true,
		"$y = $a[0] * ref x 3;":   false,
		"$y = $h{a} * ref x 3;":   false,
		"$y = $h{a} * shift x 3;": false,
	} {
		if v := recognizeSource(t, []byte(src)); v.accepted != want {
			t.Errorf("%q: accepted %v, want %v (refused at %d, expecting %v)", src, v.accepted, want, v.at, v.expected)
		}
	}
}

// Known limits, pinned so they are noticed when they end. (`f()->{foo}`, once one, is the scope-decided family's
// testdata/port/scope-decided/call-paren.pl.)
func TestPortSubscriptsKnownLimits(t *testing.T) {
	// internal/lexer emits -e (any filetest letter) as one token, so the chart reads the filetest where toke.c
	// forces the string -e (v0.1 01a1a805-a94e-7b39-8021-c61fe44b13ec). This fails when the lexer splits it.
	if got := terminalsAt(t, "$y = $h{-e};", 8); !slices.Equal(got, []string{"UNIOP"}) {
		t.Errorf("$h{-e}: -e admits %v; the known limit is gone, so drop the README line", got)
	}
}
