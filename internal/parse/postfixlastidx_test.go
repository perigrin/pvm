// ABOUTME: `->$#*` parses as a postfix last-index dereference, and its canon
// ABOUTME: re-parses to itself.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestPostfixLastIndexParses: the lexer's `$#*` token after an arrow, as in
// perl.git t/op/array.t:613, `(\my @a)->$#*++;`. Measured on 5.42.0,
// `perl -MO=Deparse -e '(\my @a)->$#*++;'` prints `++(\my @a)->$#*;`.
func TestPostfixLastIndexParses(t *testing.T) {
	for _, src := range []string{
		"my $n = $r->$#*;",
		"(\\my @a)->$#*++;",
		"$r->$#* = 3;",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		canon := parse.Canon(root, []byte(src))
		again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon))
		if canon != again {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
}
