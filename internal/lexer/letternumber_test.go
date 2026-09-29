// ABOUTME: Under `use utf8` a letter number (Unicode Nl) is an identifier character:
// ABOUTME: `package bugⅲⅱⅴⅵⅱ` and `my $ⅷ` name as perl does.

package lexer

import "testing"

// TestLetterNumberIdentifier: perl's identifier start is \p{XID_Start} and
// \p{Word}, and \p{Word} takes Nl through Alphabetic. identStart and
// identContinue asked unicode.IsLetter, which is L alone. Measured on 5.42.0:
//
//	$ perl -Mutf8 -e 'package bugⅲⅱⅴⅵⅱ { sub f { __PACKAGE__ } }
//	      binmode STDOUT, ":utf8"; print bugⅲⅱⅴⅵⅱ::f(), "\n";
//	      my $ⅷ = 8; print "$ⅷ\n";'
//	bugⅲⅱⅴⅵⅱ
//	8
//
// perl.git t/uni/package.t:94.
func TestLetterNumberIdentifier(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"use utf8; package bugⅲⅱⅴⅵⅱ { 1 }", "bugⅲⅱⅴⅵⅱ"},
		{"use utf8; my $ⅷ = 8;", "$ⅷ"},
	} {
		found := false
		for _, tok := range Tokenize([]byte(tc.src)) {
			if tc.src[tok.Start:tok.End] == tc.want {
				found = true
			}
		}
		if !found {
			t.Errorf("%q: want one token %q", tc.src, tc.want)
		}
	}
}
