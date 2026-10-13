// ABOUTME: Tests for AnalyzeLowered that a golden cannot express: what it does with the Holes of input perl rejects.
// ABOUTME: Holes are found by walking the lowered tree; kinds are compared with reflect, not type assertions.

package infer_test

import (
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

		ann, diags, _ := infer.AnalyzeLowered(f, pos)

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
	_, _, st := infer.AnalyzeLowered(f, lower.NewPositions(root, f))

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
