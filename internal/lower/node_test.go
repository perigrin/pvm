// ABOUTME: Invariant tests for the lowered vocabulary: a Hole stands anywhere a Node may,
// ABOUTME: and every node type is usable at its zero value, because the CST it comes from is routinely broken.

package lower_test

import (
	"testing"

	"tamarou.com/pvm/internal/lower"
)

// reaches reports whether a walk from root visits target.
func reaches(root, target lower.Node) bool {
	if root == target {
		return true
	}
	for _, c := range lower.Children(root) {
		if reaches(c, target) {
			return true
		}
	}
	return false
}

// The CST puts an Unknown as a binary operand, an index base or subscript,
// and under a declaration (§4.14.2, Hole row), so the lowered fields in those
// positions must take a Hole and a walk must reach it.
func TestHoleStandsAnywhereANodeMay(t *testing.T) {
	decl, base, target, l, r := &lower.Hole{}, &lower.Hole{}, &lower.Hole{}, &lower.Hole{}, &lower.Hole{}
	cases := []struct {
		name string
		node lower.Node
		hole lower.Node
	}{
		{"Decl.Init", &lower.Decl{Vars: []lower.Node{&lower.Var{Sigil: "$", Name: "x"}}, Init: decl}, decl},
		{"Index.Base", &lower.Index{Base: base, Idx: &lower.NumLit{Text: "0", IsInt: true}}, base},
		{"Match.Target", &lower.Match{Op: "m", Target: target}, target},
		{"Binary.L", &lower.Binary{Op: "+", L: l, R: &lower.NumLit{Text: "1", IsInt: true}}, l},
		{"Binary.R", &lower.Binary{Op: "+", L: &lower.NumLit{Text: "1", IsInt: true}, R: r}, r},
	}
	for _, c := range cases {
		if !reaches(c.node, c.hole) {
			t.Errorf("%s: a walk from the node does not reach its Hole", c.name)
		}
	}
}
