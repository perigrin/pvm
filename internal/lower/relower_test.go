// ABOUTME: Tests Relower: re-lowering an edited file's damaged top-level subs equals lowering the edited file afresh.
// ABOUTME: Ids and shape are compared node by node; undamaged subs are shared with the old tree, which stays unchanged.

package lower_test

import (
	"fmt"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// shape is n's s-expression with every node's id beside it, so two trees
// equal under it have the same kinds, fields, children and ids.
func shape(n lower.Node) string {
	var b strings.Builder
	var walk func(n lower.Node)
	walk = func(n lower.Node) {
		fmt.Fprintf(&b, "(%v %s", n.Meta().ID, lower.Sexpr(n))
		for _, c := range lower.Children(n) {
			walk(c)
		}
		b.WriteString(")")
	}
	walk(n)
	return b.String()
}

// declOf is the lowered Decl whose id is anchor's own, or nil.
func declOf(n lower.Node, anchor string) lower.Node {
	if n.Meta().ID == (lower.ID{Anchor: anchor}) {
		return n
	}
	for _, c := range lower.Children(n) {
		if d := declOf(c, anchor); d != nil {
			return d
		}
	}
	return nil
}

func TestRelowerEqualsFreshLowering(t *testing.T) {
	src := `package P;
my $top = 1;
sub a { my $x = 1; return $x }
{ sub b { sub inner { 2 } return inner() } }
package Q { sub a { return 'q' } }
my $f = sub { sub c { 3 } };
sub d;
1;
`
	cases := []struct {
		name, old, new string
		anchors        []string
	}{
		{"first sub", "my $x = 1", "my $x = 12", []string{"sub a#0"}},
		{"sub in a bare block", "{ 2 }", "{ 2 + 2 }", []string{"sub b#0"}},
		{"same-named sub in a package block", "'q'", "'qq'", []string{"sub a#1"}},
		{"sub in an anonymous sub", "{ 3 }", "{ 3; 4 }", []string{"sub c#0"}},
		{"two subs", "{ 3 }", "{ 4 }", []string{"sub a#0", "sub c#0"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			edited := strings.Replace(src, c.old, c.new, 1)
			if c.name == "two subs" {
				edited = strings.Replace(edited, "my $x = 1", "my $x = 2", 1)
			}
			oldRoot, newRoot := parse.Parse([]byte(src)), parse.Parse([]byte(edited))
			old := lower.Lower(oldRoot)
			before := shape(old)
			got, pos, ok := lower.Relower(old, lower.NewPositions(oldRoot, old), newRoot, c.anchors)
			if !ok {
				t.Fatal("Relower reported false")
			}
			fresh := lower.Lower(newRoot)
			if want := shape(fresh); shape(got) != want {
				t.Errorf("Relower = %s\nwant      %s", shape(got), want)
			}
			// Its position map answers as a fresh one of the edited tree.
			freshPos := lower.NewPositions(newRoot, fresh)
			var spans func(n lower.Node)
			spans = func(n lower.Node) {
				gs, ge, gok := pos.Span(n.Meta().ID)
				ws, we, wok := freshPos.Span(n.Meta().ID)
				if gs != ws || ge != we || gok != wok {
					t.Errorf("Span(%v) = %d, %d, %v; want %d, %d, %v", n.Meta().ID, gs, ge, gok, ws, we, wok)
				}
				for _, k := range lower.Children(n) {
					spans(k)
				}
			}
			spans(fresh)
			for off := range len(edited) + 1 {
				if g, w := pos.At(off), freshPos.At(off); (g == nil) != (w == nil) || g != nil && g.Meta().ID != w.Meta().ID {
					t.Errorf("At(%d) = %v, want %v", off, g, w)
				}
			}
			if shape(old) != before {
				t.Error("Relower changed the old tree")
			}
			// Every other top-level sub is the old tree's own node.
			for _, a := range []string{"sub a#0", "sub b#0", "sub a#1", "sub c#0", "sub d#0"} {
				damaged := false
				for _, d := range c.anchors {
					damaged = damaged || d == a
				}
				if shared := declOf(got, a) == declOf(old, a); shared == damaged {
					t.Errorf("%s: shared with the old tree = %v, damaged = %v", a, shared, damaged)
				}
			}
		})
	}
}

func TestRelowerUnknownAnchor(t *testing.T) {
	root := parse.Parse([]byte("sub a { 1 }\n"))
	f := lower.Lower(root)
	pos := lower.NewPositions(root, f)
	for _, anchors := range [][]string{{"sub z#0"}, {"sub a#1"}, {"sub a#0", "sub z#0"}} {
		if got, p, ok := lower.Relower(f, pos, root, anchors); ok || got != nil || p != nil {
			t.Errorf("Relower(%v) = %v, %v, %v; want nil, nil, false", anchors, got, p, ok)
		}
	}
	if got, _, ok := lower.Relower(f, pos, root, nil); !ok || got != f {
		t.Errorf("Relower(nil) = %v, %v; want f itself, true", got, ok)
	}
}
