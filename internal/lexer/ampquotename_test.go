// ABOUTME: After the `&` sigil a quote-operator word is a sub name: `&m('x')` calls
// ABOUTME: the sub m, it does not open a match.

package lexer

import "testing"

// TestAmpQuoteOpName: measured on 5.42.0,
//
//	$ perl -MO=Deparse -e 'sub m {1} sub s {2} sub y {3}
//	      my @r = (&m("a"), &s("b"), &y("c"), &qw("d"));'
//	my(@r) = (&m('a'), &s('b'), &y('c'), &qw('d'));
//
// perl.git t/comp/opsubs.t:122-179, one call for each quote operator.
func TestAmpQuoteOpName(t *testing.T) {
	for _, name := range []string{"m", "q", "qq", "qr", "qw", "qx", "s", "tr", "y"} {
		src := "is( &" + name + "('amper'), 'x' );"
		for _, tok := range Tokenize([]byte(src)) {
			if tok.Kind == Quote && src[tok.Start] != '\'' {
				t.Errorf("%q: `%s` after & is a name, got quote %q", src, name, src[tok.Start:tok.End])
			}
			if tok.Kind == UnknownRest {
				t.Errorf("%q: `%s` after & opened an unterminated quote", src, name)
			}
		}
	}
}
