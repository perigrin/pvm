// ABOUTME: `q(...)` and `qq(...)` are literal names wherever a literal is read, so
// ABOUTME: `require q(./test.pl)` resolves its helper the way `require './test.pl'` does.
package parse

import "testing"

// TestLiteralNameQuoteForms holds literalNameList's reading of perl's generic
// quote operators. `q(./test.pl)` is the string "./test.pl" with any of perl's
// delimiters, and 61 perl.git t/ files spell their require that way.
//
// A `qq` whose body would interpolate is NOT a literal: it names a file only
// at runtime, which is the rule resolveRequiredFile's `interpolates` check
// already applies to `"..."`.
func TestLiteralNameQuoteForms(t *testing.T) {
	cases := []struct {
		text string
		want string
		ok   bool
	}{
		{"q(./test.pl)", "./test.pl", true},
		{"q{./test.pl}", "./test.pl", true},
		{"q[./test.pl]", "./test.pl", true},
		{"q<./test.pl>", "./test.pl", true},
		{"q!./test.pl!", "./test.pl", true},
		{"q (./test.pl)", "./test.pl", true},
		{"qq(./test.pl)", "./test.pl", true},
		{"qq(./$name.pl)", "", false},
		{"qq{@dirs}", "", false},
	}
	for _, c := range cases {
		got, ok := literalNameList(&Node{Kind: Term, Text: c.text})
		if ok != c.ok {
			t.Errorf("literalNameList(%q) ok = %v, want %v", c.text, ok, c.ok)
			continue
		}
		if ok && (len(got) != 1 || got[0] != c.want) {
			t.Errorf("literalNameList(%q) = %v, want [%q]", c.text, got, c.want)
		}
	}
}
