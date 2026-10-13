// ABOUTME: Tests for AnalyzeLowered that a golden cannot express: what it does with the Holes of input perl rejects.
// ABOUTME: Holes are found by walking the lowered tree; kinds are compared with reflect, not type assertions.

package infer_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// holes returns every Hole under n, n included.
func holes(n lower.Node) []lower.Node {
	var out []lower.Node
	if reflect.TypeOf(n) == reflect.TypeOf(&lower.Hole{}) {
		out = append(out, n)
	}
	for _, c := range lower.Children(n) {
		out = append(out, holes(c)...)
	}
	return out
}

// A Hole is typed Any and never inspected (§4.14.2, Hole row): it is
// annotated Any, no diagnostic points inside it, and analysing it does not
// panic. Each input is one perl rejects.
func TestHoleIsTypedAnyAndNeverInspected(t *testing.T) {
	for _, src := range []string{
		`my $x = (;`,
		`"abc`,
	} {
		root := parse.Parse([]byte(src))
		f := lower.Lower(root)
		pos := lower.NewPositions(root, f)
		hs := holes(f)
		if len(hs) == 0 {
			t.Fatalf("%q lowered with no Hole:\n%s", src, lower.Tree(f))
		}

		ann, diags, _ := infer.AnalyzeLowered(f, pos, nil)

		for _, h := range hs {
			id := h.Meta().ID
			if got, ok := ann[id]; !ok || got != types.Any {
				t.Errorf("%q: Hole %v annotated %v (present %v), want Any", src, id, got, ok)
			}
			start, end, _ := pos.Span(id)
			for _, d := range diags {
				if int(d.StartByte) >= start && int(d.StartByte) < end {
					t.Errorf("%q: diagnostic %q points inside the Hole at [%d,%d)", src, d.Message, start, end)
				}
			}
		}
	}
}

// A named sub's return type is the join of its explicit returns, a bare
// one giving Undef, and the value its last statement leaves, through both
// bodies of a closing conditional. The golden format shows no return
// types, so this test holds them.
func TestLoweredSubReturnTypes(t *testing.T) {
	src := []byte(`sub one { return 1; }
sub str { "s" }
sub bare { return; }
sub both { return 1 if 0; "s" }
sub cond { if (1) { 1 } else { "s" } }
sub none { }
`)
	root := parse.Parse(src)
	f := lower.Lower(root)
	_, _, st := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil)

	for name, want := range map[string]types.Type{
		"one":  types.Int,
		"str":  types.Str,
		"bare": types.Undef,
		"both": types.Int | types.Str,
		"cond": types.Int | types.Str,
		"none": types.Unknown,
	} {
		sym, ok := st.LookupAll(name)
		if !ok {
			t.Errorf("%s: no symbol", name)
			continue
		}
		if sym.ReturnType != want {
			t.Errorf("%s: return type %v, want %v", name, sym.ReturnType, want)
		}
	}
}

// A use of a module the project index resolves analyses that module while
// the using file is analysed, so its symbols are there for a cross-file
// lookup. The module's file is rewritten afterwards without its sub: a
// lookup that finds the sub found what the use cached, not what a later
// read would.
func TestLoweredUseAnalysesTheModule(t *testing.T) {
	dir := t.TempDir()
	pm := filepath.Join(dir, "lib", "Foo.pm")
	if err := os.MkdirAll(filepath.Dir(pm), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pm, []byte("package Foo;\nsub bar { return 42; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	idx := infer.NewProjectIndex(dir)

	src := []byte("use Foo;\n")
	root := parse.Parse(src)
	f := lower.Lower(root)
	infer.AnalyzeLowered(f, lower.NewPositions(root, f), idx)

	if err := os.WriteFile(pm, []byte("package Foo;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := idx.LookupSymbol("Foo", "bar"); !ok {
		t.Error("use Foo did not analyse lib/Foo.pm")
	}
}

// declOf returns the declaration under n that declares the variable
// sigil+name, or nil.
func declOf(n lower.Node, sigil, name string) *lower.Decl {
	if d, ok := n.(*lower.Decl); ok {
		for _, v := range d.Vars {
			if v, ok := v.(*lower.Var); ok && v.Sigil == sigil && v.Name == name {
				return d
			}
		}
	}
	for _, c := range lower.Children(n) {
		if d := declOf(c, sigil, name); d != nil {
			return d
		}
	}
	return nil
}

// A conditional perl rejects is analysed without a panic, and nothing is
// said about its Holes, while what is intact is still typed: the block of
// `if (;)` declares $y, and its initialiser makes it an Int.
func TestLoweredFlowOverBrokenInput(t *testing.T) {
	for _, src := range []string{
		`if (;) { my $y = 1; }`,
		`my $x; if ($x) {`,
	} {
		root := parse.Parse([]byte(src))
		f := lower.Lower(root)
		pos := lower.NewPositions(root, f)
		hs := holes(f)
		if len(hs) == 0 {
			t.Fatalf("%q lowered with no Hole:\n%s", src, lower.Tree(f))
		}

		ann, diags, _ := infer.AnalyzeLowered(f, pos, nil)

		for _, h := range hs {
			start, end, _ := pos.Span(h.Meta().ID)
			for _, d := range diags {
				if int(d.StartByte) >= start && int(d.StartByte) < end {
					t.Errorf("%q: diagnostic %q points inside the Hole at [%d,%d)", src, d.Message, start, end)
				}
			}
		}
		if d := declOf(f, "$", "y"); d != nil {
			if got := ann[d.ID]; got != types.Int {
				t.Errorf("%q: my $y = 1 annotated %v, want Int", src, got)
			}
		}
	}
}
