// ABOUTME: A signature followed by a comment before its body parses: perl.git
// ABOUTME: t/op/signatures.t's t086 and t087 put a comment after every token.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestSignatureThenComment: measured on 5.42.0, `sub f ($a) #c)))` with its
// body on the next line deparses as `sub f ($a) { $a; }`. See
// lexer.TestBlockAfterComment for the brace.
func TestSignatureThenComment(t *testing.T) {
	src := "use feature 'signatures';\nsub f ($a) #c)))\n { $a }\n%h = ();\n"
	root := parse.Parse([]byte(src))
	if containsKind(root, parse.Unknown) {
		t.Errorf("%q: perl accepts this; got %s", src, shape(root))
	}
}
