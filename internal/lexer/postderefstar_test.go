// ABOUTME: `$r->&*` and `$r->**` are whole postfix dereferences, one token each, as
// ABOUTME: `$r->@*` is: not the & sigil and a star, or the exponent operator.

package lexer

import "testing"

// TestPostDerefCodeAndGlob: measured on 5.42.0 with -MO=Deparse,
//
//	$x = $r->&*;   $x = &$r;
//	$g = $r->**;   $g = *$r;
//	ok(!($sub_ref->&* ne "Hello, World!"), "x");
//	               ok(!(&$sub_ref ne 'Hello, World!'), 'x');
//
// PerlOnJava unit/postderef.t:31.
func TestPostDerefCodeAndGlob(t *testing.T) {
	for src, want := range map[string]string{
		"$x = $r->&*;":                "&*",
		"$g = $r->**;":                "**",
		`ok(!($s->&* ne "Hi"), "x");`: "&*",
	} {
		found := false
		for _, tok := range Tokenize([]byte(src)) {
			if text := src[tok.Start:tok.End]; tok.Kind == Variable && text == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%q: no Variable %q", src, want)
		}
	}
}
