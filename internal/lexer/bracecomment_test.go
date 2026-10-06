// ABOUTME: A `)` whose `{` follows a comment opens a block: perl's skipspace skips
// ABOUTME: comments as well as whitespace before it looks for the brace.

package lexer

import "testing"

// TestBlockAfterComment: toke.c's yyl_rightparen skips space with skipspace,
// which skips comments too, before testing for `{`. Measured on 5.42.0, a
// file of
//
//	use feature "signatures";
//	sub f ($a) #c)))
//	 { $a }
//	if ($x) # c
//	{ print 1; }
//	%h = ();
//
// deparses the sub body and the conditional's block as blocks and `%h = ()`
// as an assignment. perl.git t/op/signatures.t:1034 and :1059.
func TestBlockAfterComment(t *testing.T) {
	src := "use feature \"signatures\";\nsub f ($a) #c)))\n { $a }\nif ($x) # c\n{ print 1; }\n%h = ();\n"
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == Operator && src[tok.Start:tok.End] == "{" && !tok.OpensBlock {
			t.Errorf("the `{` at %d opens a block", tok.Start)
		}
		if tok.Kind == Operator && src[tok.Start:tok.End] == "%" {
			t.Errorf("`%%h` after the block is a hash, not modulus")
		}
	}
}
