// ABOUTME: A declaration's compound initialiser operator, `my $x -= 1`, is kept on the tree and in canon.
// ABOUTME: Dropping it made `my $x -= 1` the same tree as `my $x = 1`, which perl evaluates differently.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestDeclCompoundOperatorKept: measured on 5.42.0 with -MO=Deparse,-p,
// `my $x -= 1` is `(my $x -= 1)` and leaves $x at -1, where `my $x = 1`
// leaves 1. The parser read any assignment operator after the variable and
// recorded only the initialiser, so the two had one tree and canon wrote
// `-=` back as `=`. `my $x = 1` is the control: AssignOp stays empty.
func TestDeclCompoundOperatorKept(t *testing.T) {
	for src, op := range map[string]string{
		"my $x -= 1;":        "-=",
		"my $x ||= 1;":       "||=",
		"our $x .= 'a';":     ".=",
		"local $h{a} //= 1;": "//=",
		"state $n += 2;":     "+=",
		"my $x = 1;":         "",
	} {
		root := parse.Parse([]byte(src))
		if u := firstUnknown(root); u != nil {
			t.Errorf("%q: refused as %s", src, u.Refusal)
			continue
		}
		d := root.Children[0]
		if d.Kind == parse.Statement {
			d = d.Children[0]
		}
		if d.Kind != parse.Declaration || d.AssignOp != op {
			t.Errorf("%q: got %v AssignOp %q, want a Declaration with %q", src, d.Kind, d.AssignOp, op)
		}
		canon := parse.Canon(root, []byte(src))
		want := " = "
		if op != "" {
			want = " " + op + " "
		}
		if !strings.Contains(canon, want) {
			t.Errorf("%q: canon %q lost the operator %q", src, canon, want)
		}
		if ok, why := canonMatches([]byte(src)); !ok {
			t.Errorf("%q: canon is not a fixpoint: %s", src, why)
		}
	}
}
