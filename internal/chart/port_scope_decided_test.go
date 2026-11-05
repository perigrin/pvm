// ABOUTME: Tests the scope-decided terminals: NOAMP, the *SUB tokens and PRIVATEREF, as superpositions on a bareword.
// ABOUTME: Snippets under testdata/port/scope-decided are recognized; scope-decided-refused ones refuse.

package chart

import (
	"os"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// scopeDecidedLines are the perly.y alternatives this family ports or fits. None is left for a completing issue.
var scopeDecidedLines = []int{1046, 1330, 1646, 1678, 1680, 1691}

func TestPortScopeDecided(t *testing.T) {
	if err := grammar.CheckFamilyMarked(scopeDecidedLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "scope-decided") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

// Whether a bareword is a declared sub, and of which prototype, is scope's, so the chart accepts what any
// declaration would let perl accept and refuses what none would. `print FH x 3;`, `new Foo x 3;`,
// `foo bar x 3;` and `foo { a => 1 } x 3;` stay accepted, as the indirect objects left them: perl accepts
// each with some declaration.
func TestPortScopeDecidedRefusals(t *testing.T) {
	for _, f := range snippets(t, "scope-decided-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	for _, src := range []string{"foo(1;", "foo 1 2;", "sub foo; foo 1 2;", "a::();", "Foo:: 1;", "Foo:: $x;"} {
		if v := recognizeSource(t, []byte(src)); v.accepted {
			t.Errorf("%q is accepted; perl refuses it", src)
		}
	}
	for _, src := range []string{
		"foo;", "foo(1);", "sub foo; foo 1, 2;", "my sub bar { 1 } bar();", "&foo;", "f()->{foo};",
		"print FH x 3;", "new Foo x 3;", "foo bar x 3;", "foo { a => 1 } x 3;", // scope's, above
		"new Foo::(1);", "Foo::->new;", "foo => 1;", "x(1);",
	} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d, expecting %v", src, v.at, v.expected)
		}
	}
}

// A word before => is BAREWORD and one before ( is the zero-width & and the name, PRIVATEREF here (toke.c:8176-8211),
// whatever is declared; only a word they leave undecided carries the scope-decided readings.
func TestPortScopeDecidedSuperpositionBounded(t *testing.T) {
	scoped := []string{"NOAMP", "UNIOPSUB", "FUNC0SUB"}
	for _, tc := range []struct {
		src  string
		at   int
		want []string
	}{
		{"$y = foo => 1;", 5, []string{"BAREWORD"}},
		{"$y = foo(1);", 5, []string{"PRIVATEREF"}},
		{"$y = a::(1);", 5, []string{"BAREWORD"}},
		{"$y = Foo:: 1;", 5, []string{"BAREWORD"}},
		{"next foo;", 5, []string{"BAREWORD"}},
	} {
		if got := terminalsAt(t, tc.src, tc.at); !slices.Equal(got, tc.want) {
			t.Errorf("%q at %d admits %v, want only %v", tc.src, tc.at, got, tc.want)
		}
	}
	for _, src := range []string{"$y = foo 1;", "$y = foo;", "$y = foo + 1;"} {
		got := terminalsAt(t, src, 5)
		for _, w := range append([]string{"BAREWORD"}, scoped...) {
			if !slices.Contains(got, w) {
				t.Errorf("%q: foo admits %v, want %s among them", src, got, w)
			}
		}
		if slices.Contains(got, "LSTOPSUB") {
			t.Errorf("%q: foo admits LSTOPSUB, which only a { after it allows", src)
		}
	}
	if got := terminalsAt(t, "$y = first { 1 } 2;", 5); !slices.Contains(got, "LSTOPSUB") {
		t.Errorf("first before { admits %v, want LSTOPSUB among them", got)
	}
	if got := terminalsAt(t, "sub foo { 1 }", 4); !slices.Contains(got, "PRIVATEREF") || !slices.Contains(got, "BAREWORD") {
		t.Errorf("a sub's name admits %v, want BAREWORD and PRIVATEREF", got)
	}
}
