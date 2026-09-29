// ABOUTME: An apostrophe after a keyword opens a string; after any other word it
// ABOUTME: is the old package separator. `eval'f()'` evals, `foo'bar` is foo::bar.

package lexer

import "testing"

// TestKeywordThenApostrophe: toke.c's yyl_keylookup scans the first word
// without package parts and looks it up as a keyword; only a non-keyword
// reaches yyl_just_a_word, which takes `'` as a separator. Measured on
// 5.42.0:
//
//	$ perl -MO=Deparse -e "my \$r = eval'fred3(5)'; my \$s = print'x'; my \$t = foo'bar;"
//	my $r = eval 'fred3(5)';
//	my $s = print('x');
//	my $t = 'foo::bar';
//
// perl.git t/op/eval.t:315 and t/op/kvaslice.t:48.
func TestKeywordThenApostrophe(t *testing.T) {
	for _, tc := range []struct{ src, word, quote string }{
		{"my $r = eval'fred3(5)';", "eval", "'fred3(5)'"},
		{"my $s = print'x';", "print", "'x'"},
	} {
		toks := significant(Tokenize([]byte(tc.src)))
		found := false
		for i, tok := range toks {
			if tok.Kind == Word && tc.src[tok.Start:tok.End] == tc.word &&
				i+1 < len(toks) && toks[i+1].Kind == Quote &&
				tc.src[toks[i+1].Start:toks[i+1].End] == tc.quote {
				found = true
			}
		}
		if !found {
			t.Errorf("%q: want Word %q then Quote %q", tc.src, tc.word, tc.quote)
		}
	}
	src := "my $t = foo'bar;"
	toks := significant(Tokenize([]byte(src)))
	if len(toks) < 4 || src[toks[3].Start:toks[3].End] != "foo'bar" {
		t.Errorf("%q: a non-keyword keeps the apostrophe as a separator", src)
	}
}
