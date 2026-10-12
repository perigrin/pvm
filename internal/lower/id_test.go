// ABOUTME: Invariant tests for node identity: ids survive an edit above them and never collide.
// ABOUTME: Byte offsets and whole-file preorder indexes both fail the first test, which is why it exists.

package lower_test

import (
	"testing"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// namedSubs returns the named sub declarations of a tree, by name.
func namedSubs(n *parse.Node) map[string]*parse.Node {
	subs := map[string]*parse.Node{}
	var walk func(*parse.Node)
	walk = func(n *parse.Node) {
		if n.Kind == parse.Declaration && n.Text == "sub" &&
			len(n.Children) > 0 && n.Children[0].Kind == parse.Term {
			subs[n.Children[0].Text] = n
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(n)
	return subs
}

// sameIDs walks two subtrees of the same shape in step and fails on the
// first node whose id differs.
func sameIDs(t *testing.T, what string, a *parse.Node, aIDs map[*parse.Node]lower.ID, b *parse.Node, bIDs map[*parse.Node]lower.ID) {
	t.Helper()
	if a.Kind != b.Kind || len(a.Children) != len(b.Children) {
		t.Fatalf("%s: shapes differ: %s/%d vs %s/%d", what, a.Kind, len(a.Children), b.Kind, len(b.Children))
	}
	if aIDs[a] != bIDs[b] {
		t.Errorf("%s: %s %q changed id: %v -> %v", what, a.Kind, a.Text, aIDs[a], bIDs[b])
	}
	for i := range a.Children {
		sameIDs(t, what, a.Children[i], aIDs, b.Children[i], bIDs)
	}
}

// everyNode returns every node of a tree in preorder.
func everyNode(n *parse.Node) []*parse.Node {
	out := []*parse.Node{n}
	for _, c := range n.Children {
		out = append(out, everyNode(c)...)
	}
	return out
}

// distinct fails unless every node of the tree has an id and no two share one.
func distinct(t *testing.T, root *parse.Node, ids map[*parse.Node]lower.ID) {
	t.Helper()
	nodes := everyNode(root)
	byID := map[lower.ID]*parse.Node{}
	for _, n := range nodes {
		id, ok := ids[n]
		if !ok {
			t.Errorf("%s %q at %d has no id", n.Kind, n.Text, n.Start)
			continue
		}
		if prev, dup := byID[id]; dup {
			t.Errorf("id %v shared by %s %q and %s %q", id, prev.Kind, prev.Text, n.Kind, n.Text)
		}
		byID[id] = n
	}
	if len(byID) != len(nodes) {
		t.Errorf("a map keyed by id holds %d entries for %d nodes", len(byID), len(nodes))
	}
}

func TestIDStableUnderEditAbove(t *testing.T) {
	before := parse.Parse([]byte("sub a { my $x = 1; } sub b { my $y = 2; }"))
	after := parse.Parse([]byte("my $z = 0;\nsub a { my $x = 1; } sub b { my $y = 2; }"))
	beforeIDs, afterIDs := lower.AssignIDs(before), lower.AssignIDs(after)
	distinct(t, before, beforeIDs)
	distinct(t, after, afterIDs)

	beforeSubs, afterSubs := namedSubs(before), namedSubs(after)
	for _, name := range []string{"a", "b"} {
		if beforeSubs[name] == nil || afterSubs[name] == nil {
			t.Fatalf("sub %s not found in both trees", name)
		}
		sameIDs(t, "sub "+name, beforeSubs[name], beforeIDs, afterSubs[name], afterIDs)
	}
}

func TestIDDistinguishesSameSpanNodes(t *testing.T) {
	root := parse.Parse([]byte("my $x = 5; my @y = ($x); my $z = $y[0];"))
	ids := lower.AssignIDs(root)

	atZero := 0
	for _, n := range everyNode(root) {
		if n.Start == 0 {
			atZero++
		}
	}
	if atZero < 2 {
		t.Fatalf("want several nodes starting at byte 0, got %d", atZero)
	}
	distinct(t, root, ids)
}

func TestIDDistinguishesSiblingsWithSameShape(t *testing.T) {
	root := parse.Parse([]byte("sub f { my $a = 1; my $a = 1; }"))
	ids := lower.AssignIDs(root)

	var decls []*parse.Node
	for _, n := range everyNode(root) {
		if n.Kind == parse.Declaration && n.Text == "my" {
			decls = append(decls, n)
		}
	}
	if len(decls) != 2 {
		t.Fatalf("want two my declarations, got %d", len(decls))
	}
	if ids[decls[0]] == ids[decls[1]] {
		t.Errorf("identical siblings share id %v", ids[decls[0]])
	}
	distinct(t, root, ids)
}

func TestIDOverBrokenTree(t *testing.T) {
	unclosed := parse.Parse([]byte("if ($x) {"))
	var withChildren *parse.Node
	for _, n := range everyNode(unclosed) {
		if n.Kind == parse.Unknown && len(n.Children) > 0 {
			withChildren = n
		}
	}
	if withChildren == nil {
		t.Fatal("want an Unknown with children in `if ($x) {`")
	}
	distinct(t, unclosed, lower.AssignIDs(unclosed))

	broken := parse.Parse([]byte("sub a { my $x = ; } sub b { 1 }"))
	clean := parse.Parse([]byte("sub b { 1 }"))
	brokenIDs, cleanIDs := lower.AssignIDs(broken), lower.AssignIDs(clean)
	distinct(t, broken, brokenIDs)

	sawUnknown := false
	for _, n := range everyNode(namedSubs(broken)["a"]) {
		sawUnknown = sawUnknown || n.Kind == parse.Unknown
	}
	if !sawUnknown {
		t.Fatal("want an Unknown inside the broken sub a")
	}
	sameIDs(t, "sub b", namedSubs(broken)["b"], brokenIDs, namedSubs(clean)["b"], cleanIDs)
}

// TestAnonSubWithSignatureOpensNoAnchor: an anonymous sub opens no anchor,
// signature or not. `sub ($x) { 1 }` parses with its one parameter as a
// parenthesised term in the slot a named sub's name takes, and that term is
// not a name: were it read as one, two anonymous subs with the same
// one-parameter signature would anchor as "sub $x#0" and "sub $x#1", and
// inserting one above the other would renumber it.
func TestAnonSubWithSignatureOpensNoAnchor(t *testing.T) {
	src := "use v5.36; my $f = sub ($x) { 1 }; my $g = sub ($x) { 2 };"
	root := parse.Parse([]byte(src))
	ids := lower.AssignIDs(root)
	for _, n := range everyNode(root) {
		if id := ids[n]; id.Anchor != "" {
			t.Errorf("%s %q at %d is anchored %q; an anonymous sub opens no anchor", n.Kind, n.Text, n.Start, id.Anchor)
		}
	}

	named := parse.Parse([]byte("use v5.36; sub f ($x) { 1 }"))
	namedIDs := lower.AssignIDs(named)
	f, ok := namedSubs(named)["f"]
	if !ok {
		t.Fatal("no sub f in the parse")
	}
	if got := namedIDs[f.Children[1]].Anchor; got != "sub f#0" {
		t.Errorf("sub f's signature is anchored %q, want %q", got, "sub f#0")
	}
}

// TestDerivedIDs: a lowered node with no parse node of its own takes its
// parent's Anchor and Local under a nonzero Role. A named sub's Decl and
// its AnonSub come from one parse node and differ only in Role; the Hole a
// childless deref holds as its Expr is not the root's zero id.
func TestDerivedIDs(t *testing.T) {
	decl := lowerLast(t, "sub f { 1 }")
	kids := lower.Children(decl) // the Var &f, then the AnonSub
	sub := kids[len(kids)-1]
	d, s := decl.Meta().ID, sub.Meta().ID
	if d == s || d.Role != 0 || s != (lower.ID{Anchor: d.Anchor, Local: d.Local, Role: lower.RoleSubValue}) {
		t.Errorf("sub f: Decl id %v, AnonSub id %v; want the same node under RoleSubValue", d, s)
	}

	for _, src := range []string{"$#{$x};", "$#$x;", "&$x;"} {
		root := parse.Parse([]byte(src))
		deref := lowerLast(t, src)
		hole := lower.Children(deref)[0]
		want := deref.Meta().ID
		want.Role = lower.RoleDerefOperand
		if got := hole.Meta().ID; got != want || got == lower.AssignIDs(root)[root] {
			t.Errorf("%q: the deref's Hole has id %v, want %v", src, got, want)
		}
	}
}
