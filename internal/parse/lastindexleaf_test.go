// ABOUTME: `$#{ EXPR }` stays one Term leaf spanning its braces, the tree lower and Sites read,
// ABOUTME: now that the lexer emits its `$#` and its block's tokens apart.

package parse

import "testing"

// TestLastIndexBlockIsOneLeaf: internal/lexer emits `$#{ EXPR }` as a
// DerefSigil `$#` and the block's own tokens (toke.c's DOLSHARP, then a
// block), so a quoted `}` no longer ends it. The parser joins them back
// into the leaf it built when the lexer emitted one Variable.
func TestLastIndexBlockIsOneLeaf(t *testing.T) {
	for _, c := range []struct{ src, leaf string }{
		{`my $n = $#{$r->{a}};`, `$#{$r->{a}}`},
		{`my $n = $#{ ["}"] };`, `$#{ ["}"] }`},
		{`my $n = $#{ \@a } + 1;`, `$#{ \@a }`},
	} {
		var found bool
		var walk func(*Node)
		walk = func(n *Node) {
			if n.Kind == Term && n.Text == c.leaf && n.Start == len(`my $n = `) && len(n.Children) == 0 {
				found = true
			}
			for _, k := range n.Children {
				walk(k)
			}
		}
		root := Parse([]byte(c.src))
		walk(root)
		if !found || refusedNode(root) {
			t.Errorf("%q: want one Term leaf %q and no refusal", c.src, c.leaf)
		}
	}
}

func refusedNode(n *Node) bool {
	if n.Refused() {
		return true
	}
	for _, k := range n.Children {
		if refusedNode(k) {
			return true
		}
	}
	return false
}
