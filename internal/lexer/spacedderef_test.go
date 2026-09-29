// ABOUTME: A deref sigil may be separated from the sigil it applies to by spaces:
// ABOUTME: `$ $name` is `$$name`, and `$$ ;` is still the process id.

package lexer

import "testing"

// TestSpacedDeref: scan_ident skips space after a sigil before deciding what
// follows, for a second sigil as for a brace or a name. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my $x = $ $name1; my @y = @ $r; my $z = $$ ;'
//	my $x = $$name1;
//	my(@y) = @$r;
//	my $z = $$;
//
// perl.git t/op/postfixderef.t:224, `is ($ $name1, undef, ...)`.
func TestSpacedDeref(t *testing.T) {
	for _, src := range []string{"my $x = $ $name1;", "my @y = @ $r;"} {
		toks := significant(Tokenize([]byte(src)))
		if len(toks) < 5 || toks[3].Kind != DerefSigil || toks[4].Kind != Variable {
			t.Errorf("%q: want a DerefSigil then the variable it applies to", src)
		}
	}
	src := "my $z = $$ ;"
	toks := significant(Tokenize([]byte(src)))
	if len(toks) < 4 || src[toks[3].Start:toks[3].End] != "$$" {
		t.Errorf("%q: `$$` with nothing named after it is the process id", src)
	}
}
