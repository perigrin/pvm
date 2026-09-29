// ABOUTME: A statement follows a format body: the `{` after the closing `.`
// ABOUTME: opens a block, labelled or bare, not an anonymous hash.

package lexer

import "testing"

// TestStatementAfterFormat: toke.c brackets a format like a block -- the `=`
// goes through yyl_leftcurly, which saves XSTATE on the bracket stack
// (toke.c:6689), and the `.` through yyl_rightcurly, which pops it
// (toke.c:6862). Measured on 5.42.0, a file of `format X =`, `T`, `.` and
// `SKIP: { print 1; $x = 1; }` deparses the block with both statements in it.
//
// perl.git t/op/write.t:1683.
func TestStatementAfterFormat(t *testing.T) {
	for _, src := range []string{
		"format X =\nT\n.\nSKIP: { f(1); $x = 1; }\n",
		"format X =\nT\n.\n{ f(1); $x = 1; }\n",
	} {
		found := false
		for _, tok := range Tokenize([]byte(src)) {
			if src[tok.Start:tok.End] == "{" {
				found = true
				if !tok.OpensBlock {
					t.Errorf("%q: the `{` after the format opens a block", src)
				}
			}
		}
		if !found {
			t.Errorf("%q: no `{` token", src)
		}
	}
}
