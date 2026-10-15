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

// A file with nothing to report gets an empty diagnostics slice, not nil: a
// caller that serialises it writes [], not null.
func TestLoweredDiagnosticsAreNeverNil(t *testing.T) {
	src := []byte("42;")
	root := parse.Parse(src)
	f := lower.Lower(root)
	_, diags, _ := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil, infer.Options{})
	if diags == nil {
		t.Error("diagnostics are nil, want an empty slice")
	}
}

// A sub that returns a list leaves its last element in scalar context: the
// comma operator. A sub's return type is its scalar-context value, so a
// returned List gives its last element's type, whichever of the three forms
// returns it. In list context the call is the whole list, so an array it
// fills does not record the last element's type as its element type.
//
// perl 5.42: with `sub r { return ("a", 2) }`, `my $v = r()` is 2 and
// `my @l = r()` is ("a", 2); `sub s2 { ("x", 2.5) }` gives 2.5 and
// `sub t { return 1, "b" }` gives "b".
func TestLoweredListReturnIsItsLastElement(t *testing.T) {
	src := []byte(`sub r { return ("a", 2) }
sub s2 { ("x", 2.5) }
sub t { return 1, "b" }
my $v = r();
my $w = s2();
my $u = t();
my @l = r();
`)
	root := parse.Parse(src)
	f := lower.Lower(root)
	_, _, st := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil, infer.Options{})

	for name, want := range map[string]types.Type{
		"r":  types.Int,
		"s2": types.Num,
		"t":  types.Str,
		"$v": types.Int,
		"$w": types.Num,
		"$u": types.Str,
	} {
		sym, ok := st.LookupAll(name)
		if !ok {
			t.Errorf("%s: no symbol", name)
			continue
		}
		got := sym.Type
		if sym.Kind == infer.SymSubroutine {
			got = sym.ReturnType
		}
		if got != want {
			t.Errorf("%s: %v, want %v", name, got, want)
		}
	}

	sym, ok := st.Lookup("@l")
	if !ok {
		t.Fatal("@l: no symbol")
	}
	if sym.ElemType == types.Int {
		t.Errorf("@l element type %v: the last element's type, but @l holds (\"a\", 2)", sym.ElemType)
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
// Strict: a command's output is not typed, so
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

// perl truncates a repeat count toward zero, so a finite Num count is no
// mismatch: perl 5.42 -w is silent on `"ab" x 2.5` and `"ab" x 2.9`, each
// "abab", and on `(1,2) x 1.5`, (1, 2). A count that is not a number at
// all is still one: a reference repeats by its address, a lossy coercion
// reported under WarnLossyCoercion.
func TestRepeatCountAcceptsNum(t *testing.T) {
	for _, src := range []string{
		"my $s = \"ab\" x 2.5;\n",
		"my $n = 2.9;\nmy $s = \"ab\" x $n;\n",
		"my @l = (1, 2) x 1.5;\n",
	} {
		root := parse.Parse([]byte(src))
		f := lower.Lower(root)
		_, diags, _ := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil, infer.Options{})
		for _, d := range diags {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}
	root := parse.Parse([]byte("my $s = \"ab\" x {};\n"))
	f := lower.Lower(root)
	_, diags, _ := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil, infer.Options{WarnLossyCoercion: true})
	if len(diags) != 1 || diags[0].Code != infer.CodeLossyCoercion {
		t.Errorf(`"ab" x {}: diagnostics %v, want one lossy coercion`, diags)
	}
}

// A call with an argument perl rejects is analysed without a panic, and
// draws no arity or argument diagnostic, by default or under Strict: the
// Hole that stands for the argument counts as one, and is typed Any.
func TestLoweredCallsOverBrokenInput(t *testing.T) {
	for _, src := range []string{
		`length(;`,
		`Foo->new(;`,
		`bless(;`,
	} {
		root := parse.Parse([]byte(src))
		f := lower.Lower(root)
		if len(holes(f)) == 0 {
			t.Fatalf("%q lowered with no Hole:\n%s", src, lower.Tree(f))
		}
		for _, opts := range []infer.Options{{}, {Strict: true}} {
			_, diags, _ := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil, opts)
			for _, d := range diags {
				if d.Code == infer.CodeArityMismatch || d.Code == infer.CodeTypeMismatch {
					t.Errorf("%q (Strict %v): diagnostic %q, want none", src, opts.Strict, d.Message)
				}
			}
		}
	}
}

// A constructor's result records its class on the variable that takes it,
// declared or assigned. The golden format shows no class, so this test
// holds it.
func TestLoweredConstructorRecordsClass(t *testing.T) {
	src := []byte("my $obj = Foo->new();\nmy $o;\n$o = Bar->new;\nmy $n = Baz->make;\n")
	root := parse.Parse(src)
	f := lower.Lower(root)
	_, _, st := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil, infer.Options{})
	for name, want := range map[string]string{"$obj": "Foo", "$o": "Bar", "$n": ""} {
		sym, ok := st.Lookup(name)
		if !ok {
			t.Fatalf("no %s in the symbol table", name)
		}
		if sym.ClassType != want {
			t.Errorf("%s: class %q, want %q", name, sym.ClassType, want)
		}
	}
}

// A call the project index resolves takes the return type of the sub it
// names: a fully-qualified sub, and a method on an object whose class a
// constructor recorded.
func TestLoweredCallsResolveThroughTheIndex(t *testing.T) {
	dir := t.TempDir()
	for path, body := range map[string]string{
		"lib/Foo/Bar.pm": "package Foo::Bar;\nsub baz { return 42; }\n",
		"lib/Counter.pm": "package Counter;\nsub count { return 4.5; }\n",
	} {
		full := filepath.Join(dir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	idx := infer.NewProjectIndex(dir)

	src := []byte("use Foo::Bar;\nuse Counter;\nmy $x = Foo::Bar::baz();\nmy $c = Counter->new();\nmy $n = $c->count();\n")
	root := parse.Parse(src)
	f := lower.Lower(root)
	_, _, st := infer.AnalyzeLowered(f, lower.NewPositions(root, f), idx, infer.Options{})

	for name, want := range map[string]types.Type{"$x": types.Int, "$n": types.Num} {
		sym, ok := st.Lookup(name)
		if !ok {
			t.Fatalf("no %s in the symbol table", name)
		}
		if sym.Type != want {
			t.Errorf("%s: %v, want %v", name, sym.Type, want)
		}
	}
}

// A named sub's parameters are typed by what the operators and builtins
// its body passes them to expect, whether a signature, `my (...) = @_` or
// a run of `my $v = shift` names them, and whether a builtin's arguments
// stand in parentheses or not: perl compiles `length($x)` and `length $x`
// alike. A sub none of whose parameters is constrained records none. The golden format shows no parameter types,
// so this test holds them.
func TestLoweredParamTypes(t *testing.T) {
	src := []byte(`use v5.36;
sub calc { my $x = shift; my $result = $x + 1; }
sub passthrough { my $x = shift; return $x; }
sub two { my $self = shift; my $s = shift; return $s . "!"; }
sub unpack { my ($p, $q) = @_; return $p * $q; }
sub sig ($a, $b) { return lc $a; }
sub paren { my $x = shift; return length($x); }
`)
	root := parse.Parse(src)
	f := lower.Lower(root)
	_, _, st := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil, infer.Options{})

	for name, want := range map[string][]types.Type{
		"calc":        {types.Num},
		"passthrough": nil,
		"two":         {types.Unknown, types.Str},
		"unpack":      {types.Num, types.Num},
		"sig":         {types.Str, types.Unknown},
		"paren":       {types.Str},
	} {
		sym, ok := st.LookupAll(name)
		if !ok {
			t.Fatalf("%s: no symbol", name)
		}
		if !reflect.DeepEqual(sym.ParamTypes, want) {
			t.Errorf("%s: parameter types %v, want %v", name, sym.ParamTypes, want)
		}
	}
}

// A map or eval block perl rejects lowers to a Hole, which is typed Any and
// never inspected: the analysis does not panic, the Hole is annotated Any,
// and the array map fills records Any as its element type rather than a
// type guessed from the broken body.
func TestLoweredSpecialFormsOverBrokenInput(t *testing.T) {
	for _, tc := range []struct {
		src, name string
		elem      types.Type
	}{
		{`my @b = map {;`, "@b", types.Any},
		{`my $r = eval {;`, "$r", types.Unknown},
	} {
		root := parse.Parse([]byte(tc.src))
		f := lower.Lower(root)
		hs := holes(f)
		if len(hs) == 0 {
			t.Fatalf("%q lowered with no Hole:\n%s", tc.src, lower.Tree(f))
		}
		ann, diags, st := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil, infer.Options{})
		for _, h := range hs {
			if got := ann[h.Meta().ID]; got != types.Any {
				t.Errorf("%q: Hole annotated %v, want Any", tc.src, got)
			}
		}
		for _, d := range diags {
			t.Errorf("%q: diagnostic %q, want none", tc.src, d.Message)
		}
		sym, ok := st.Lookup(tc.name)
		if !ok {
			t.Fatalf("%q: %s is not in the symbol table", tc.src, tc.name)
		}
		if sym.ElemType != tc.elem {
			t.Errorf("%q: %s element type %v, want %v", tc.src, tc.name, sym.ElemType, tc.elem)
		}
	}
}

// $1 reads its group from the quote token of the match before it, whatever
// its delimiters: perl 5.42 -w is silent on `$1 + 1` after each of these,
// all digit groups. A group of letters is a Str.
func TestLoweredCaptureReadsEveryQuoteForm(t *testing.T) {
	for src, want := range map[string]types.Type{
		`my $s = "a1"; $s =~ m{(\d+)}; my $c = $1;`:             types.Int,
		`my $s = "a1"; $s =~ s/(\d+)/x/; my $c = $1;`:           types.Int,
		`my $s = "a/1"; $s =~ m/a\/(\d)/; my $c = $1;`:          types.Int,
		`my $s = "a1"; $s =~ m{(?:[a-z]{1})(\d+)}; my $c = $1;`: types.Int,
		`my $r = qr/(\d+)/; my $c = $1;`:                        types.Int,
		`my $s = "ab"; $s =~ m{([a-z]+)}; my $c = $1;`:          types.Str,
	} {
		root := parse.Parse([]byte(src))
		f := lower.Lower(root)
		_, _, st := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil, infer.Options{})
		if sym, ok := st.Lookup("$c"); !ok || sym.Type != want {
			t.Errorf("%q: $c is %v (found %v), want %v", src, sym.Type, ok, want)
		}
	}
}

// loweredDiags analyses src on the lowered path and returns its diagnostics.
func loweredDiags(src string) []infer.Diagnostic {
	root := parse.Parse([]byte(src))
	f := lower.Lower(root)
	_, diags, _ := infer.AnalyzeLowered(f, lower.NewPositions(root, f), nil, infer.Options{})
	return diags
}

// A filehandle slot takes a Glob, a GlobRef, an IO handle object, or a
// Scalar that may hold one. perl 5.42 -w is silent on each: a lexical
// handle from open, and an IO::File, in seek, read, stat, -s and close; a
// GlobRef; and stat and -e given a filename.
func TestLexicalFilehandleSatisfiesHandleSlot(t *testing.T) {
	for _, src := range []string{
		"open(my $fh, '<', 'f.txt') or die;\nseek($fh, 0, 0);\nread($fh, my $b, 3);\nmy @s = stat($fh);\nmy $n = -s $fh;\nclose($fh);\n",
		"use IO::File;\nmy $fh = IO::File->new('f.txt', '<');\nseek($fh, 0, 0);\nread($fh, my $b, 3);\nmy @s = stat($fh);\nclose($fh);\n",
		"my $g = \\*STDOUT;\nmy $t = tell($g);\n",
		"my $f = 'f.txt';\nmy @s = stat($f);\nmy @t = stat('f.txt');\nmy $e = -e $f;\n",
	} {
		for _, d := range loweredDiags(src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}
}

// A value that cannot be a handle is still reported in a filehandle slot:
// perl 5.42 dies "Not a GLOB reference" on close({}) and seek([], 0, 0),
// and -w warns "on unopened filehandle" for close(42) and seek("FH", 0, 0).
// A HashRef is not a filename either.
func TestNonHandleIsReportedInHandleSlot(t *testing.T) {
	for _, src := range []string{
		"close({});\n",
		"seek([], 0, 0);\n",
		"close(42);\n",
		"seek(\"FH\", 0, 0);\n",
		"my @s = stat({});\n",
	} {
		if diags := loweredDiags(src); len(diags) != 1 || diags[0].Code != infer.CodeTypeMismatch {
			t.Errorf("%q: diagnostics %v, want one type mismatch", src, diags)
		}
	}
}

// A Boolean in a numeric slot numifies to 1 or 0, and perl 5.42 -w says
// nothing: `my $s = "-" x ($n > 2); my $q = ($n > 2) / 1` prints `[-] 1`,
// and B/Deparse.pm writes `'()' x (7 < $cx)` and `(@kids != 1)/2`. So a
// Boolean operand of an arithmetic operator, a repeat count, or an Int or
// Num argument of a builtin is no mismatch. A reference there still is,
// under WarnLossyCoercion.
func TestBooleanInNumericSlotIsSilent(t *testing.T) {
	for _, src := range []string{
		"my $n = 3;\nmy $s = \"-\" x ($n > 2);\n",
		"my $n = 3;\nmy $q = ($n > 2) / 1;\n",
		"my $cx = 9;\nmy $p = '()' x (7 < $cx);\n",
		"my @kids = (1);\nmy $h = (@kids != 1)/2;\n",
		"my $f = 7 < 3;\nmy $y = $f + 1;\n",
		"my $f = 7 < 3;\nmy $c = chr($f);\n",
		"my $v = abs(1 < 2);\n",
	} {
		_, diags := analyzeSource(t, []byte(src))
		for _, d := range diags {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	// A broad type that may hold a number still satisfies the slot as it
	// is: neither a Scalar nor a Hole's Any is reported there.
	for _, src := range []string{
		"my $m;\nmy $v = abs($m);\n",
		"my $m;\nmy $c = chr($m);\n",
		"abs(;",
	} {
		_, diags := analyzeSource(t, []byte(src))
		for _, d := range diags {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	for _, src := range []string{
		"my $r = {};\nmy $y = $r + 1;\n",
		"my $r = {};\nmy $c = chr($r);\n",
	} {
		if _, diags, _ := analyzeSourceWithOptions(t, []byte(src), infer.Options{WarnLossyCoercion: true}); len(diags) == 0 {
			t.Errorf("%q: no diagnostic, want a reference in a numeric slot reported", src)
		}
	}
}

// Arithmetic is typed Num|Inf because it can overflow, and `**` Num|NaN|Inf,
// but a value that only MIGHT be Inf or NaN is checked as finite: perl 5.42
// -w is silent on `my $q = 20 / 8; $q + 1` (3.5), on `"\t" x (20 / 8)` (two
// tabs) and on `2 ** 0.5 + 1` (2.414...). perigrin ruled 2026-10-14 that only
// a value KNOWN to be Inf or NaN is reported.
func TestMaybeInfIsCheckedAsFinite(t *testing.T) {
	for _, src := range []string{
		"my $lvl = 20;\nmy $q = $lvl / 8;\nmy $r = $q + 1;\n",
		"my $lvl = 20;\nmy $t = \"\\t\" x ($lvl / 8);\n",
		"my $x = 2;\nmy $y = $x ** 0.5 + 1;\n",
		"my $lvl = 20;\nmy $v = abs($lvl / 8);\n",
	} {
		_, diags := analyzeSource(t, []byte(src))
		for _, d := range diags {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	for _, c := range []struct{ actual, want types.Type }{
		{types.Num | types.Inf, types.Num},
		{types.Int | types.Num | types.Inf, types.Int | types.Num},
		{types.Num | types.NaN | types.Inf, types.Num},
		{types.Int | types.NaN, types.Int},
	} {
		if got := infer.Finite(c.actual); got != c.want {
			t.Errorf("Finite(%v) = %v, want %v", c.actual, got, c.want)
		}
	}

	// No expression is typed exactly Inf or NaN yet; a value known to be
	// one is still no Num, as perl warns "Non-finite repeat count does
	// nothing" on `"ab" x 9**9**9`.
	for _, known := range []types.Type{types.Inf, types.NaN, types.Inf | types.NaN} {
		got := infer.Finite(known)
		if got != known {
			t.Errorf("Finite(%v) = %v, want it unchanged", known, got)
		}
		if types.TypeSatisfies(got, types.Num) {
			t.Errorf("Finite(%v) = %v satisfies Num, want it reported", known, got)
		}
	}
}
