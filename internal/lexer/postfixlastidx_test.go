// ABOUTME: `->$#*` is the postfix last-index dereference: one token, as `->@*` is,
// ABOUTME: not `$#` followed by a multiplication.

package lexer

import "testing"

// TestPostfixLastIndex: after `->` the star closes the dereference.
// Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my $r = [1,2]; my $n = $r->$#*; (\my @a)->$#*++;'
//	my $n = $r->$#*;
//	++(\my @a)->$#*;
//
// perl.git t/op/array.t:613 and 628. Outside a postfix dereference `$#` then
// `*` is still the variable named `#` times something.
func TestPostfixLastIndex(t *testing.T) {
	src := "my $n = $r->$#*;"
	toks := significant(Tokenize([]byte(src)))
	found := false
	for _, tok := range toks {
		if src[tok.Start:tok.End] == "$#*" && tok.Kind == Variable {
			found = true
		}
	}
	if !found {
		t.Errorf("%q: want one Variable `$#*` after the arrow", src)
	}
	src = "my $n = $# * 2;"
	for _, tok := range significant(Tokenize([]byte(src))) {
		if src[tok.Start:tok.End] == "$#*" {
			t.Errorf("%q: `$#` outside a postfix dereference must not take the star", src)
		}
	}
}
