// ABOUTME: Tests that AnalyzeLowered's annotations are keyed by node identity: no two nodes share a key, and an edit above a sub keeps the keys inside it.
// ABOUTME: These are what byte-offset keys could not promise: a parent and child starting at one byte, and offsets shifting after an edit.

package infer_test

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// firstOfKind returns the first node under n, in preorder, of the same kind
// as like.
func firstOfKind(n lower.Node, like lower.Node) lower.Node {
	if reflect.TypeOf(n) == reflect.TypeOf(like) {
		return n
	}
	for _, c := range lower.Children(n) {
		if f := firstOfKind(c, like); f != nil {
			return f
		}
	}
	return nil
}

// `$y[0]` and its base `$y` start at the same byte. Keyed by offset, the
// element's annotation overwrote the base's; keyed by id, each keeps its
// own: the element is an Int, and the base is the array @y.
func TestAnnotationsKeyedByIdentity(t *testing.T) {
	src := []byte(`my $x = 5; my @y = ($x); my $z = $y[0];`)
	root := parse.Parse(src)
	f := lower.Lower(root)
	pos := lower.NewPositions(root, f)
	ann, _, st := infer.AnalyzeLowered(f, pos, nil, infer.Options{})

	idx, ok := firstOfKind(f, &lower.Index{}).(*lower.Index)
	if !ok {
		t.Fatalf("no Index in:\n%s", lower.Tree(f))
	}
	is, _, _ := pos.Span(idx.ID)
	bs, _, _ := pos.Span(idx.Base.Meta().ID)
	if is != bs {
		t.Fatalf("$y[0] starts at %d and its base at %d; the test needs them to share a byte", is, bs)
	}
	if idx.ID == idx.Base.Meta().ID {
		t.Fatalf("$y[0] and its base share the id %v", idx.ID)
	}
	if got := ann[idx.ID]; got != types.Int {
		t.Errorf("$y[0] annotated %v, want Int", got)
	}
	y, _ := st.Lookup("@y")
	if got, ok := ann[idx.Base.Meta().ID]; !ok || got != y.Type {
		t.Errorf("the base of $y[0] annotated %v (present %v), want @y's type %v", got, ok, y.Type)
	}
}

// The annotations inside a named sub are keyed within that sub, so a
// statement inserted above both subs changes none of their keys or types.
func TestAnnotationsSurviveEditAbove(t *testing.T) {
	inSubs := func(src string) map[lower.ID]types.Type {
		root := parse.Parse([]byte(src))
		f := lower.Lower(root)
		ann, _, _ := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil, infer.Options{})
		out := map[lower.ID]types.Type{}
		for id, typ := range ann {
			if id.Anchor != "" {
				out[id] = typ
			}
		}
		return out
	}
	const subs = "sub a { my $x = 1; } sub b { my $y = 2; }\n"
	before := inSubs(subs)
	after := inSubs("my $z = 0;\n" + subs)

	anchors := map[string]bool{}
	for id := range before {
		anchors[id.Anchor] = true
	}
	if !anchors["sub a#0"] || !anchors["sub b#0"] {
		t.Fatalf("no annotations keyed inside both subs: %v", before)
	}
	if !reflect.DeepEqual(before, after) {
		t.Errorf("annotations inside the subs changed after an edit above them:\nbefore %v\nafter  %v", before, after)
	}
}
