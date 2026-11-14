// ABOUTME: Tests F2's forest memory: first blocks sized from the file, grammar-only data built once per tables value,
// ABOUTME: and one forest whichever of the slipper's closures (kept from recognition, or recomputed) a node reads.

package forest

import (
	"os"
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// TestForestBlocksSizedForTinySources: zero- and one-token sources size
// their first blocks from the file without panic, and give the same forest
// as blocks sized for no tokens at all, which grow from the placeholder.
func TestForestBlocksSizedForTinySources(t *testing.T) {
	tab := committedTables(t)
	for _, src := range []string{"", "\n", "# a comment\n", "1", "$x"} {
		s, _ := lex.Lex([]byte(src), tab.Tokens)
		l := lex.Linear(s)
		c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
		if err != nil {
			t.Fatal(err)
		}
		if len(l.Edges) > 1 || !c.Accepted() {
			t.Fatalf("%q: %d tokens, accepted %v; want at most one, accepted", src, len(l.Edges), c.Accepted())
		}
		sized, unsized := build(tab, c, l, len(l.Edges)), build(tab, c, l, 0)
		if err := wellFormed(sized); err != nil {
			t.Fatalf("%q: %v", src, err)
		}
		if got, want := dump(sized), dump(unsized); got != want {
			t.Errorf("%q: the forest from sized blocks\n%s\ndiffers from unsized blocks'\n%s", src, got, want)
		}
	}
}

// TestForestGrammarDataOncePerTables: however many forests are built over
// one tables value, its grammar-only forest data is built once; a copy of
// the tables builds its own; and the recognizer's grammar-only tables, kept
// beside the same value, are the same value before and after.
func TestForestGrammarDataOncePerTables(t *testing.T) {
	tab := new(grammar.Tables)
	*tab = *committedTables(t) // a value of its own, so nothing built beside the shared one counts
	src, err := os.ReadFile("../testdata/std5k.pl")
	if err != nil {
		t.Fatal(err)
	}
	s, _ := lex.Lex(src, tab.Tokens)
	l := lex.Linear(s)
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
	if err != nil {
		t.Fatal(err)
	}
	recognizer := derivedValues(tab)
	before := forestGrammarBuilds.Load()
	for range 3 {
		if Build(tab, c, l) == nil {
			t.Fatal("std5k.pl has no forest")
		}
	}
	if n := forestGrammarBuilds.Load() - before; n != 1 {
		t.Errorf("three forests over one tables value built its grammar data %d times, want 1", n)
	}
	after := derivedValues(tab)
	forestType := reflect.ValueOf(reflect.TypeFor[*forestGrammar]()).Pointer()
	if len(recognizer) != 1 || len(after) != 2 || after[forestType] == 0 {
		t.Fatalf("derived values before the forests %v, after %v; want the recognizer's, then it and the forest's", recognizer, after)
	}
	for typ, v := range recognizer {
		if after[typ] != v {
			t.Errorf("the recognizer's derived value moved from %#x to %#x", v, after[typ])
		}
	}
	cp := *tab
	if Build(&cp, c, l) == nil {
		t.Fatal("std5k.pl has no forest over the copy")
	}
	if n := forestGrammarBuilds.Load() - before; n != 2 {
		t.Errorf("a copy of the tables built %d grammar data in all, want its own: 2", n)
	}
}

// derivedValues is what grammar.Derived keeps beside tab, as each value's
// address by its type's: read by reflection, since the slots are grammar's.
func derivedValues(tab *grammar.Tables) map[uintptr]uintptr {
	out := map[uintptr]uintptr{}
	m := reflect.ValueOf(tab).Elem().FieldByName("derived")
	for _, k := range m.MapKeys() {
		out[k.Elem().Pointer()] = m.MapIndex(k).Elem().FieldByName("value").Elem().Pointer()
	}
	return out
}

// TestForestSameAcrossCacheAndLeo: built and cache-served nodes, Leo on and
// off, give one forest. With the cache on, a node built during recognition
// keeps its slipper closure (Leo on) and a cache-served one recomputes it
// (Leo off); with it off every node keeps one. Over the corpus sample each
// file's dump is the same under all four.
func TestForestSameAcrossCacheAndLeo(t *testing.T) {
	tab := committedTables(t)
	opts := []recognize.Options{{}, {NoCache: true}, {NoLeo: true}, {NoCache: true, NoLeo: true}}
	rs := make([]*recognize.Recognizer, len(opts))
	for i, o := range opts {
		rs[i] = recognize.New(tab, o) // long-lived, so the cache serves nodes as it does in the ratchet
	}
	accepted := 0
	for _, p := range samplePaths(t) {
		src, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		s, _ := lex.Lex(src, tab.Tokens)
		l := lex.Linear(s)
		var want string
		for i, r := range rs {
			c, err := recognize.Recognize(r, l)
			if err != nil {
				t.Fatal(err)
			}
			var f *Forest
			if c.Accepted() {
				f = Build(tab, c, l)
			}
			got := dump(f)
			if i == 0 {
				want = got
				if f != nil {
					accepted++
				}
			} else if got != want {
				t.Errorf("%s: the forest under %+v differs from the default's", p, opts[i])
			}
		}
	}
	if accepted == 0 {
		t.Fatal("no file of the corpus sample has a forest")
	}
}
