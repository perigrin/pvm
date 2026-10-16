// ABOUTME: Invariant tests for the lowered vocabulary: a Hole stands anywhere a Node may,
// ABOUTME: and every node type is usable at its zero value, because the CST it comes from is routinely broken.

package lower_test

import (
	"reflect"
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

// Every lowered field is optional (§6.1.6 item 5): a lowering over broken
// input leaves fields unset, so each kind at its zero value must walk
// without panicking, report no children, and carry a usable Meta.
func TestZeroValueNodeIsUsable(t *testing.T) {
	for _, n := range allKinds() {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%T: walking the zero value panicked: %v", n, r)
				}
			}()
			if kids := lower.Children(n); len(kids) != 0 {
				t.Errorf("%T: zero value has %d children, want none", n, len(kids))
			}
			if m := n.Meta(); m == nil || m.ID != (lower.ID{}) || m.Paren {
				t.Errorf("%T: zero value's Meta is not the zero identity", n)
			}
		}()
	}
}

// UseStmts finds every use, no and require in source order, nested ones
// included: a require inside a sub or a BEGIN block loads its module too.
func TestUseStmtsInSourceOrder(t *testing.T) {
	strict := &lower.UseStmt{Op: "use", Module: "strict"}
	inSub := &lower.UseStmt{Op: "require", Module: "Inner"}
	inBegin := &lower.UseStmt{Op: "require", Module: "Phase"}
	byPath := &lower.UseStmt{Op: "require", Expr: &lower.StrLit{Value: "foo.pl"}}
	file := &lower.File{Stmts: []lower.Node{
		strict,
		&lower.Decl{Kind: lower.DeclSub, Init: &lower.Block{Stmts: []lower.Node{inSub}}},
		&lower.PhaserStmt{Kind: "BEGIN", Body: &lower.Block{Stmts: []lower.Node{inBegin}}},
		&lower.ExprStmt{X: &lower.Hole{}},
		byPath,
	}}
	want := []*lower.UseStmt{strict, inSub, inBegin, byPath}
	if got := lower.UseStmts(file); !reflect.DeepEqual(got, want) {
		t.Errorf("UseStmts = %v, want %v", got, want)
	}
	if got := lower.UseStmts(nil); got != nil {
		t.Errorf("UseStmts(nil) = %v, want nil", got)
	}
}

// allKinds returns every lowered kind at its zero value.
func allKinds() []lower.Node {
	return []lower.Node{
		&lower.Hole{},
		&lower.NumLit{}, &lower.StrLit{}, &lower.Interp{}, &lower.QwList{}, &lower.Bareword{},
		&lower.Var{}, &lower.Deref{}, &lower.Index{}, &lower.Key{}, &lower.Slice{},
		&lower.Unary{}, &lower.Binary{}, &lower.Repeat{}, &lower.Assign{}, &lower.Ternary{},
		&lower.CmpChain{}, &lower.Range{}, &lower.List{},
		&lower.Decl{},
		&lower.Call{}, &lower.MethodCall{},
		&lower.MapGrepSort{}, &lower.DoBlock{}, &lower.DoFile{}, &lower.EvalBlock{}, &lower.EvalStr{},
		&lower.AnonSub{}, &lower.ArrayLiteral{}, &lower.HashLiteral{}, &lower.Ref{},
		&lower.FileTest{}, &lower.Readline{}, &lower.Glob{}, &lower.LoopEx{}, &lower.Wantarray{},
		&lower.Match{},
		&lower.Block{}, &lower.ExprStmt{}, &lower.IfStmt{},
		&lower.BlockStmt{}, &lower.WhileStmt{}, &lower.CForStmt{}, &lower.ForeachStmt{},
		&lower.UseStmt{}, &lower.PhaserStmt{}, &lower.File{},
	}
}
