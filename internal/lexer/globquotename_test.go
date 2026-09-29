// ABOUTME: After a glob's `*` a quote-operator word is the glob's name: `*y`
// ABOUTME: is the glob y, not a transliteration; `$b*s/a/b/` still multiplies.

package lexer

import "testing"

// TestGlobQuoteOpName: measured on 5.42.0,
//
//	$ perl -MO=Deparse -e '*x = *y; *s = *tr; my $r = \*q; $a = $b*s/a/b/;'
//	*x = *y;
//	*s = *tr;
//	my $r = \*q;
//	$a = $b * s/a/b/;
//
// perl.git t/op/gv.t:1132.
func TestGlobQuoteOpName(t *testing.T) {
	for _, name := range []string{"m", "q", "qq", "qr", "qw", "qx", "s", "tr", "y"} {
		src := "*x = *" + name + "; $n = 1;"
		for _, tok := range Tokenize([]byte(src)) {
			if tok.Kind == Quote || tok.Kind == UnknownRest || tok.Kind == Error {
				t.Errorf("%q: `%s` after * is a glob name, got %v %q",
					src, name, tok.Kind, src[tok.Start:tok.End])
			}
		}
	}
	// In operator position the star multiplies and the word is a quote.
	src := "$a = $b*s/a/b/;"
	if _, ok := firstOfKind(src, Quote); !ok {
		t.Errorf("%q: `s/a/b/` after a multiplying * is a substitution", src)
	}
}
