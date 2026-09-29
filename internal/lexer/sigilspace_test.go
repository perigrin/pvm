// ABOUTME: Whitespace and comments may stand between a sigil and its name, as
// ABOUTME: toke.c's scan_ident skips them: `my $ bits` declares $bits.

package lexer

import "testing"

// TestSpaceAfterSigil: scan_ident skips space after the sigil,
//
//	if (isSPACE(*s) || !*s)
//	    s = skipspace(s);
//
// and skipspace skips comments too. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my @ a = (1); my % h = (); my $ # c
//	    b = 2; print $ b, @ a;'
//	my(@a) = 1;
//	my(%h) = ();
//	my $b = 2;
//	print $b, @a;
//
// perl.git t/op/caller.t:279 has `my $ bits`, and op/signatures.t's t085 to
// t087 spell signature elements across lines and comments. The token keeps
// the bytes between, so the source round-trips.
func TestSpaceAfterSigil(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"my $ bits = 1;", "$ bits"},
		{"my @ a = (1);", "@ a"},
		{"my % h = ();", "% h"},
		{"my $ # c\n b = 2;", "$ # c\n b"},
	} {
		toks := significant(Tokenize([]byte(tc.src)))
		if len(toks) < 2 || toks[1].Kind != Variable || tc.src[toks[1].Start:toks[1].End] != tc.want {
			t.Errorf("%q: want one Variable %q", tc.src, tc.want)
		}
	}
	// No name after the space: the sigil does not reach across to an
	// operator. `$a % $b` is still a modulus.
	src := "my $n = $a % $b;"
	toks := significant(Tokenize([]byte(src)))
	found := false
	for _, tok := range toks {
		if src[tok.Start:tok.End] == "%" && tok.Kind == Operator {
			found = true
		}
	}
	if !found {
		t.Errorf("%q: `%%` after a term is the modulus operator", src)
	}
}
