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

		ann, diags, _ := infer.AnalyzeLowered(f, pos, nil, infer.Options{})

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
	_, _, st := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil, infer.Options{})

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
	infer.AnalyzeLowered(f, lower.NewPositions(root, f), idx, infer.Options{})

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

		ann, diags, _ := infer.AnalyzeLowered(f, pos, nil, infer.Options{})

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

// A subscript perl rejects is analysed without a panic, and its Hole draws
// no index diagnostic, while the variable it subscripts is still typed: the
// base of `$a[` is @a, and the base of `$h{` is %h.
func TestLoweredSubscriptOverBrokenInput(t *testing.T) {
	for _, c := range []struct {
		src, agg string
		like     lower.Node
	}{
		{`my @a = (1); my $e = $a[;`, "@a", &lower.Index{}},
		{`my %h; my $v = $h{;`, "%h", &lower.Key{}},
	} {
		root := parse.Parse([]byte(c.src))
		f := lower.Lower(root)
		pos := lower.NewPositions(root, f)
		if len(holes(f)) == 0 {
			t.Fatalf("%q lowered with no Hole:\n%s", c.src, lower.Tree(f))
		}

		ann, diags, st := infer.AnalyzeLowered(f, pos, nil, infer.Options{})

		for _, d := range diags {
			t.Errorf("%q: diagnostic %q, want none", c.src, d.Message)
		}
		sub := firstOfKind(f, c.like)
		if sub == nil {
			t.Fatalf("%q: no %T in:\n%s", c.src, c.like, lower.Tree(f))
		}
		base := lower.Children(sub)[0]
		sym, _ := st.Lookup(c.agg)
		if got, ok := ann[base.Meta().ID]; !ok || got != sym.Type {
			t.Errorf("%q: the subscript's base annotated %v (present %v), want %s's type %v", c.src, got, ok, c.agg, sym.Type)
		}
	}
}

// An interpolated string is a Str, whatever it interpolates.
func TestLoweredInterpIsStr(t *testing.T) {
	s := &lower.Interp{Parts: []lower.Node{&lower.StrLit{Value: "n="}, &lower.Var{Sigil: "$", Name: "n"}}}
	s.ID = lower.ID{Local: 1}
	s.Parts[0].Meta().ID = lower.ID{Local: 2}
	s.Parts[1].Meta().ID = lower.ID{Local: 3}
	f := &lower.File{Stmts: []lower.Node{s}}
	ann, _, _ := infer.AnalyzeLowered(f, nil, nil, infer.Options{})
	if got := ann[s.ID]; got != types.Str {
		t.Errorf("Interp annotated %v, want Str", got)
	}
}

// An operand of no known type is passed over by default and reported under
// Strict, as on the legacy path: a command's output is not typed, so
// adding 1 to it is unchecked until Strict asks for a narrowing.
func TestLoweredStrictSeesUnknownOperand(t *testing.T) {
	src := []byte("my $n = `true` + 1;\n")
	root := parse.Parse(src)
	f := lower.Lower(root)
	pos := lower.NewPositions(root, f)

	_, permissive, _ := infer.AnalyzeLowered(f, pos, nil, infer.Options{})
	_, strict, _ := infer.AnalyzeLowered(f, pos, nil, infer.Options{Strict: true})

	if len(permissive) != 0 {
		t.Errorf("the default reported %v, want nothing", permissive)
	}
	if len(strict) != 1 || strict[0].Code != infer.CodeCoercionMismatch {
		t.Errorf("Strict reported %v, want one coercion mismatch", strict)
	}
}

// An operator with an operand perl rejects is analysed without a panic, and
// nothing is said about the Hole that stands for the operand, by default or
// under Strict, while the operator is still typed from what is intact.
func TestLoweredOperatorsOverBrokenInput(t *testing.T) {
	for _, c := range []struct {
		src  string
		like lower.Node
		want types.Type
	}{
		{`my $x = 1 +;`, &lower.Binary{}, types.Num | types.Inf},
		{`my $s = "a" x;`, &lower.Repeat{}, types.Str},
	} {
		root := parse.Parse([]byte(c.src))
		f := lower.Lower(root)
		pos := lower.NewPositions(root, f)
		if len(holes(f)) == 0 {
			t.Fatalf("%q lowered with no Hole:\n%s", c.src, lower.Tree(f))
		}
		op := firstOfKind(f, c.like)
		if op == nil {
			t.Fatalf("%q: no %T in:\n%s", c.src, c.like, lower.Tree(f))
		}

		for _, opts := range []infer.Options{{}, {Strict: true}} {
			ann, diags, _ := infer.AnalyzeLowered(f, pos, nil, opts)
			for _, d := range diags {
				t.Errorf("%q (Strict %v): diagnostic %q, want none", c.src, opts.Strict, d.Message)
			}
			if got := ann[op.Meta().ID]; got != c.want {
				t.Errorf("%q (Strict %v): %T annotated %v, want %v", c.src, opts.Strict, c.like, got, c.want)
			}
		}
	}
}

// loweredSymbolType analyses src on the lowered path and returns name's
// type and the diagnostics.
func loweredSymbolType(t *testing.T, src, name string) (types.Type, []infer.Diagnostic) {
	t.Helper()
	root := parse.Parse([]byte(src))
	f := lower.Lower(root)
	_, diags, st := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil, infer.Options{})
	sym, ok := st.Lookup(name)
	if !ok {
		t.Fatalf("%q: no %s in the symbol table", src, name)
	}
	return sym.Type, diags
}

// An expression in scalar context evaluates each arm of a ternary, and the
// left operand of a logical operator, in that context, and an array there is
// its count. perl 5.42, -w silent: with @a = (7,8,9), `(1 ? @a : 42) + 1`
// is 4, `my $n = 1 ? @a : 42` is 3, a nested arm is counted too, and
// `(@a || 0) + 1` is 4. In list context the array stays a list:
// `my @t = 1 ? @a : 42` is (7, 8, 9).
func TestTernaryArmTakesNumericContext(t *testing.T) {
	for _, c := range []struct {
		src, name string
		want      types.Type
	}{
		{"my @a = (7, 8, 9);\nmy $x = (1 ? @a : 42) + 1;\n", "$x", types.Num | types.Inf},
		{"my @a = (7, 8, 9);\nmy $n = 1 ? @a : 42;\n", "$n", types.Int},
		{"my @a = (7, 8, 9);\nmy $m = (1 ? (0 ? 1 : @a) : 42) + 1;\n", "$m", types.Num | types.Inf},
		{"my @a = (7, 8, 9);\nmy $o = (@a || 0) + 1;\n", "$o", types.Num | types.Inf},
		{"my @a = (7, 8, 9);\nmy @t = 1 ? @a : 42;\n", "@t", types.Join(types.Array, types.Int)},
	} {
		got, diags := loweredSymbolType(t, c.src, c.name)
		for _, d := range diags {
			t.Errorf("%q: diagnostic %q, want none", c.src, d.Message)
		}
		if got != c.want {
			t.Errorf("%q: %s is %v, want %v", c.src, c.name, got, c.want)
		}
	}
}
