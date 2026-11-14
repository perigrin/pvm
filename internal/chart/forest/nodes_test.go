// ABOUTME: Tests that the forest's nodes and alternatives hold no pointers, so the collector traces nothing in them.
// ABOUTME: Also that one forest's indexes never reach into the next, and that refused and empty sources are well formed.

package forest

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// TestForestNodesPointerFree: Node and Alt hold no pointer, slice, map,
// string, interface, func or channel, at any depth, and the Forest holds
// them in slices of exactly those types. Every field is an int32, the width
// the forest's overflow shortcut names.
func TestForestNodesPointerFree(t *testing.T) {
	var traced func(reflect.Type) bool
	traced = func(ty reflect.Type) bool {
		switch ty.Kind() {
		case reflect.Pointer, reflect.UnsafePointer, reflect.Slice, reflect.Map, reflect.String,
			reflect.Interface, reflect.Func, reflect.Chan:
			return true
		case reflect.Array:
			return traced(ty.Elem())
		case reflect.Struct:
			for i := range ty.NumField() {
				if traced(ty.Field(i).Type) {
					return true
				}
			}
		}
		return false
	}
	held := map[reflect.Type]bool{}
	ft := reflect.TypeFor[Forest]()
	for i := range ft.NumField() {
		if f := ft.Field(i); f.Type.Kind() == reflect.Slice {
			held[f.Type.Elem()] = true
		}
	}
	for _, ty := range []reflect.Type{reflect.TypeFor[Node](), reflect.TypeFor[Alt]()} {
		if traced(ty) {
			t.Errorf("%v holds a field the collector traces", ty)
		}
		if !held[ty] {
			t.Errorf("Forest holds no slice of %v", ty)
		}
		for i := range ty.NumField() { // pins the shortcut: indexes wrap past 2^31-1
			if f := ty.Field(i); f.Type.Kind() != reflect.Int32 {
				t.Errorf("%v.%s is %v, not int32: widen the overflow note in forest.go and the README with it", ty, f.Name, f.Type)
			}
		}
	}
}

// wellFormed reports how f's indexes leave its slices, or nil: the root and
// every Alt, Left, Right and Next index a node or alternative, 0 being none,
// and the placeholders at 0 empty.
func wellFormed(f *Forest) error {
	nodes, alts := int32(len(f.Nodes)), int32(len(f.Alts))
	if nodes == 0 || alts == 0 || f.Nodes[0] != (Node{}) || f.Alts[0] != (Alt{}) {
		return fmt.Errorf("the placeholders are missing or set: %d nodes, %d alts", nodes, alts)
	}
	if f.Root <= 0 || f.Root >= nodes {
		return fmt.Errorf("root %d outside 1..%d", f.Root, nodes-1)
	}
	for i, n := range f.Nodes {
		if n.Alt < 0 || n.Alt >= alts {
			return fmt.Errorf("node %d's Alt %d outside 0..%d", i, n.Alt, alts-1)
		}
	}
	for i, a := range f.Alts {
		for _, x := range []int32{a.Left, a.Right} {
			if x < 0 || x >= nodes {
				return fmt.Errorf("alt %d's child %d outside 0..%d", i, x, nodes-1)
			}
		}
		if a.Next < 0 || a.Next >= alts {
			return fmt.Errorf("alt %d's Next %d outside 0..%d", i, a.Next, alts-1)
		}
	}
	return nil
}

// TestForestReparseIndependent: indexes from one forest never resolve into
// the next. Parsing a large file and then a small one, and the reverse,
// gives each the dump a fresh parse gives it, leaves the first forest's
// dump as it was, and keeps every index of each inside its own slices.
func TestForestReparseIndependent(t *testing.T) {
	tab := committedTables(t)
	big, err := os.ReadFile(filepath.Join("..", "testdata", "std5k.pl"))
	if err != nil {
		t.Fatal(err)
	}
	small := []byte("if ($x) { print 1 } else { my @a = (1, 2); }\n")
	for _, pair := range [][2][]byte{{big, small}, {small, big}} {
		fresh := dump(parsed(t, tab, pair[1]))
		first := parsed(t, tab, pair[0])
		before := dump(first)
		second := parsed(t, tab, pair[1])
		if got := dump(second); got != fresh {
			t.Errorf("a file parsed after another of %d bytes differs from its fresh parse", len(pair[0]))
		}
		if dump(first) != before {
			t.Errorf("parsing a second file changed the first's forest")
		}
		for _, f := range []*Forest{first, second} {
			if err := wellFormed(f); err != nil {
				t.Error(err)
			}
		}
	}
}

// parsed is src's forest, failing t where the chart refuses it.
func parsed(t *testing.T, tab *grammar.Tables, src []byte) *Forest {
	t.Helper()
	f, ref := Parse(tab, src)
	if f == nil {
		t.Fatalf("%.40q: refused %+v", src, ref)
	}
	return f
}

// TestForestRefusedAndEmptySources: a refused source is a refusal with no
// forest and no panic; an empty source the chart accepts gives a well-formed
// forest whose every node is zero-width at 0, with one derivation of the
// start symbol.
func TestForestRefusedAndEmptySources(t *testing.T) {
	tab := committedTables(t)
	for _, src := range []string{"my $x = ;\n", "1 +", "}", "print <<EOT;\nno terminator\n"} {
		if f, ref := Parse(tab, []byte(src)); f != nil || ref == (Refusal{}) {
			t.Errorf("%q: forest %v, refusal %+v; want none and a refusal", src, f != nil, ref)
		}
	}
	for _, src := range []string{"", "\n", "# a comment\n"} {
		f, ref := Parse(tab, []byte(src))
		if f == nil {
			t.Fatalf("%q: refused %+v", src, ref)
		}
		if err := wellFormed(f); err != nil {
			t.Fatalf("%q: %v", src, err)
		}
		if r := f.Nodes[f.Root]; int(r.Sym) != tab.Start || r.From != 0 || r.To != 0 {
			t.Errorf("%q: root %+v, want the start symbol over 0-0", src, r)
		}
		for _, i := range reachable(f) {
			if n := f.Nodes[i]; n.From != 0 || n.To != 0 {
				t.Errorf("%q: node %+v is not zero-width at 0", src, n)
			}
		}
		if trees, cut := f.Derivations(f.Root, 2); len(trees) != 1 || cut {
			t.Errorf("%q: %d derivations, cut %v; want exactly one", src, len(trees), cut)
		}
	}
}
