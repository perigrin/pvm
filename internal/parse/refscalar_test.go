// ABOUTME: Tests that a \$ prototype slot refuses what perl refuses there: constants, sub results, aggregates, wrong arity.
// ABOUTME: Each case was measured with perl 5.42.0's `perl -c`; scalar lvalues perl accepts stay parsed.
package parse

import (
	"slices"
	"testing"
	"testing/fstest"
)

// refScalarPrelude declares what every case below calls: a plain sub f, the
// `\$`-prototyped sref, and the variables the cases pass.
const refScalarPrelude = "sub f {} sub sref (\\$); my ($x, $y, %h, @a, $r); "

// refusedAs returns the refusal codes of every Unknown in a parse of src.
func refusedAs(src string) []RefusalCode {
	return refusalsIn(Parse([]byte(src)))
}

// refusalsIn returns the refusal codes of every Unknown in a tree.
func refusalsIn(root *Node) []RefusalCode {
	var codes []RefusalCode
	var walk func(*Node)
	walk = func(n *Node) {
		if n.Kind == Unknown {
			codes = append(codes, n.Refusal)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)
	return codes
}

// wantRefScalarRefusal wants each call refused, once, with the `\$` slot's
// own code.
func wantRefScalarRefusal(t *testing.T, calls ...string) {
	t.Helper()
	for _, call := range calls {
		src := refScalarPrelude + call + ";\n"
		if got := refusedAs(src); len(got) != 1 || got[0] != "ref_scalar_slot" {
			t.Errorf("%s: refusals %v, want one ref_scalar_slot", call, got)
		}
	}
}

// wantNoRefusal wants each source parsed with no Unknown at all.
func wantNoRefusal(t *testing.T, srcs ...string) {
	t.Helper()
	for _, src := range srcs {
		if got := refusedAs(src); len(got) > 0 {
			t.Errorf("%q: refused %v; perl compiles it", src, got)
		}
	}
}

// TestRefScalarSlotRefusesConstants: RFC 0001 "Refusing what perl refuses".
// A `\$` slot takes a scalar lvalue. Measured on 5.42.0 with `perl -c`:
// `sref(1)` and `sref("s")` die "Type of arg 1 to main::sref must be scalar
// (not constant item)", and `sref(f())` "(not subroutine entry)", while
// `sref($x)`, `sref($h{k})` and `sref($x = 7)` compile.
func TestRefScalarSlotRefusesConstants(t *testing.T) {
	wantRefScalarRefusal(t, "sref(1)", `sref("s")`, "sref(f())")
	wantNoRefusal(t,
		refScalarPrelude+"sref($x);\n",
		refScalarPrelude+"sref($h{k});\n",
		refScalarPrelude+"sref($x = 7);\n",
	)
}

// TestRefScalarSlotRefusesAggregatesAndArity: measured on 5.42.0, `sref(@a)`
// dies "(not private array)", `sref(%h)` "(not private hash)", `sref(@$r)`
// "(not array dereference)" and `sref(my @q)` "(not private array)";
// `sref()` and a bare `sref` die "Not enough arguments for main::sref" and
// `sref($x, $y)` "Too many arguments", which is the arity refusal's code,
// not the slot's. `sref(my $z)` and `sref(substr($x,
// 0, 1))` compile, as do a slice, `sref(@a[0])`, a parenless `sref $x, $y`
// whose comma is the enclosing list's, an optional slot left empty,
// `sopt()` under `(;\$)`, and comp/proto.t's `sreftest my $a = 'quidgley',
// $i++` under `(\$$)`.
func TestRefScalarSlotRefusesAggregatesAndArity(t *testing.T) {
	wantRefScalarRefusal(t, "sref(@a)", "sref(%h)", "sref(@$r)", "sref(%{$r})",
		"sref(my @q)", "sref(our %q)")
	wantArityRefusal(t, refScalarPrelude+"sref();\n", refScalarPrelude+"sref;\n",
		refScalarPrelude+"sref($x, $y);\n")
	wantNoRefusal(t,
		refScalarPrelude+"sref(my $z);\n",
		refScalarPrelude+"sref(substr($x, 0, 1));\n",
		refScalarPrelude+"sref(@a[0]);\n",
		refScalarPrelude+"sref $x, $y;\n",
		"sub sopt (;\\$); my $x; sopt(); sopt($x);\n",
		"sub stwo (\\$$); my $x; stwo($x, 1);\n",
		// comp/proto.t:623: the comma after an initialiser is the call's.
		"sub stwo (\\$$); my $i; stwo my $a = 'quidgley', $i++;\n",
	)
}

// TestRefScalarSlotFromModuleSource: a `\$` slot is one whichever source
// stated it. A module's own `sub sref (\$)` and a declaration file's
// `sub sref (Scalar \$x)`, which derives `(\$)`, refuse `sref(1)` alike
// and keep `sref($x)`.
func TestRefScalarSlotFromModuleSource(t *testing.T) {
	saved := declarations
	defer func() { declarations = saved }()
	none := func(string) ([]byte, bool) { return nil, false }
	for _, tc := range []struct {
		name string
		decl fstest.MapFS
		load Loader
	}{
		{".pm", fstest.MapFS{}, func(name string) ([]byte, bool) {
			return []byte("package M;\nour @EXPORT = qw(sref);\nsub sref (\\$) { }\n1;\n"), name == "M"
		}},
		{".pmt", fstest.MapFS{"declarations/M.pmt": {Data: []byte("package M;\nour @EXPORT = qw(sref);\nsub sref (Scalar \\$x);\n")}}, none},
	} {
		declarations = tc.decl
		if got := refusalsIn(ParseWithLoader([]byte("use M; sref(1);\n"), tc.load)); len(got) != 1 || got[0] != RefScalarSlot {
			t.Errorf("%s: sref(1) refusals %v, want one %s", tc.name, got, RefScalarSlot)
		}
		if got := refusalsIn(ParseWithLoader([]byte("use M; my $x; sref($x);\n"), tc.load)); len(got) > 0 {
			t.Errorf("%s: sref($x) refused %v", tc.name, got)
		}
	}
}

// TestRefScalarSlotKeepsLvalues: every scalar lvalue measured to compile in
// a `\$` slot on 5.42.0 keeps its parse, as do the subs perl does not yet
// know to be non-lvalue: one declared after the call, a lexical sub, an
// `:lvalue` sub, and an undeclared qualified one. An `&` call ignores the
// prototype, so `&sref(1)` compiles too.
func TestRefScalarSlotKeepsLvalues(t *testing.T) {
	var srcs []string
	for _, call := range []string{
		"sref($x)", "sref($h{k})", "sref($a[0])", "sref(f()->[0])", "sref($x = 7)",
		"sref(my $z)", "sref(substr($x, 0, 1))", "sref($x,)", "sref(($x, $y))",
		"sref(g()); sub g {}", "my sub g {} sref(g())", "sub g :lvalue {} sref(g())",
		"sub g :method :lvalue {} sref(g())", "sref(Foo::g())", "&sref(1)",
	} {
		srcs = append(srcs, refScalarPrelude+call+";\n")
	}
	wantNoRefusal(t, srcs...)
}

// TestRefScalarSlotLeavesUnprototypedAlone: only a `\$` slot refuses a
// constant. `sub g {} g(1)` and `sub g ($) {} g(1)` compile on 5.42.0.
func TestRefScalarSlotLeavesUnprototypedAlone(t *testing.T) {
	wantNoRefusal(t, "sub g {} g(1);\n", "sub g ($) {} g(1);\n", "sub g {} g(@ARGV);\n")
}

// TestRefScalarSlotRefusalCodeDistinct: the `\$` refusal has a code of its
// own, listed in the inventory, so a corpus file can tell it from any other
// refusal.
func TestRefScalarSlotRefusalCodeDistinct(t *testing.T) {
	site, ok := RefusalSites[RefScalarSlot]
	if !ok {
		t.Fatalf("%s is not in RefusalSites", RefScalarSlot)
	}
	for code, other := range RefusalSites {
		if code != RefScalarSlot && other.Where == site.Where {
			t.Errorf("%s shares its site with %s", RefScalarSlot, code)
		}
	}
	if got := refusedAs(refScalarPrelude + "sref(1);\n"); len(got) != 1 || got[0] != RefScalarSlot {
		t.Errorf("sref(1): refusals %v, want one %s", got, RefScalarSlot)
	}
}

// TestPmtAliasedScalarConstantDefaultRefused: `Scalar \$x = 1` derives
// `;\$` with a default perl would refuse in that slot -- `sref(1)` is "must
// be scalar (not constant item)" -- so the declaration is in error. A
// scalar lvalue default, `Scalar \$x = $_`, is not.
func TestPmtAliasedScalarConstantDefaultRefused(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f (Scalar \\$x = 1);\n":     `sub f: aliased \$x defaults to 1, which a \$ slot refuses; it takes a scalar lvalue`,
		"sub f (Scalar \\$x = 'a');\n":   `sub f: aliased \$x defaults to 'a', which a \$ slot refuses; it takes a scalar lvalue`,
		"sub f (Scalar \\$x = @ARGV);\n": `sub f: aliased \$x defaults to @ARGV, which a \$ slot refuses; it takes a scalar lvalue`,
	})
	if facts := readDeclaration([]byte("sub f (Scalar \\$x = $_);\n"), nil); len(facts.errs) > 0 {
		t.Errorf("an lvalue default: %v", facts.errs)
	}
}

// TestDeclaredSubIsPackageScoped: a `sub f` declares f in the package it is
// read in, and an unqualified call names a sub in the CURRENT package.
// Measured on 5.42.0 with `perl -c`: after `package Foo; sub f {}`, main's
// `f()` is an undeclared sub, so `sref(f())` compiles.
func TestDeclaredSubIsPackageScoped(t *testing.T) {
	wantNoRefusal(t,
		"package Foo; sub f {} package main; sub sref (\\$) {} sref(f());\n",
		"package Foo { sub f {} } sub sref (\\$) {} sref(f());\n",
	)
	if got := refusedAs("package Foo; sub f {} sub sref (\\$) {} sref(f());\n"); len(got) != 1 || got[0] != RefScalarSlot {
		t.Errorf("Foo's own f(): refusals %v, want one %s", got, RefScalarSlot)
	}
}

// TestRefScalarSlotIsPackageScoped: a `\$` prototype applies where perl
// resolves the call to the sub that carries it. Measured on 5.42.0 with
// `perl -c`, each `ok` compiles and each `refused` dies "Type of arg 1 to
// A::s2 must be scalar (not constant item)".
func TestRefScalarSlotIsPackageScoped(t *testing.T) {
	wantNoRefusal(t,
		"package A; sub s2 (\\$) {} package B; s2(1);\n",
		"sub sref (\\$) {} package Foo; sref(1);\n",
		"package A; sub s2 (\\$) {} package B; ::s2(1);\n",
		"package A { sub s2 (\\$) {} } s2(1);\n",
		"package A; sub s2 (\\$) {} package main; s2(1);\n",
		"package Foo; sub sref (\\$) {} package main; my $x; Foo::sref($x);\n",
	)
	for _, src := range []string{
		"package A; sub s2 (\\$) {} package B; A::s2(1);\n",
		"package A { sub s2 (\\$) {} s2(1); }\n",
		"package A; sub s2 (\\$) {} package B; package A; s2(1);\n",
		"sub sref (\\$) {} package Foo; main::sref(1);\n",
		"sub sref (\\$) {} package Foo; ::sref(1);\n",
		"sub Foo::s2 (\\$) {} package Foo; s2(1);\n",
		"package A; sub s2 (\\$) {} { package B; } s2(1);\n",
	} {
		if got := refusedAs(src); len(got) != 1 || got[0] != RefScalarSlot {
			t.Errorf("%q: refusals %v, want one %s", src, got, RefScalarSlot)
		}
	}
}

// TestImportIsPackageScoped: `use M` imports into the package it is read
// in. Measured on 5.42.0 with an Exporter module M exporting `sub sref
// (\$)`: `package B; use M; sref(1)` and `B::sref(1)` from package C die
// "must be scalar (not constant item)", while main's `sref(1)` after `package
// B; use M; package main;` and B's after `use M; package B;` compile.
func TestImportIsPackageScoped(t *testing.T) {
	load := func(name string) ([]byte, bool) {
		return []byte("package M;\nour @EXPORT = qw(sref);\nsub sref (\\$) { }\n1;\n"), name == "M"
	}
	for src, refused := range map[string]bool{
		"package B; use M; sref(1);\n":               true,
		"package B; use M; package C; B::sref(1);\n": true,
		"package B; use M; package main; sref(1);\n": false,
		"use M; package B; sref(1);\n":               false,
	} {
		var want []RefusalCode
		if refused {
			want = []RefusalCode{RefScalarSlot}
		}
		if got := refusalsIn(ParseWithLoader([]byte(src), load)); !slices.Equal(got, want) {
			t.Errorf("%q: refusals %v, want %v", src, got, want)
		}
	}
}

// TestLexicalSubCrossesPackages: a lexical sub, and a `use builtin` import,
// is in scope whatever package the parse moves to. Measured on 5.42.0 with
// `perl -c`: after `my sub s2 (\$) {}`, `our sub` or `state sub`, `package
// B; s2(1)` dies "must be scalar (not constant item)"; after `use builtin
// "refaddr"`, `package P; my $a = refaddr $x, 1` compiles, refaddr a named
// unary (porting/podcheck.t:715).
func TestLexicalSubCrossesPackages(t *testing.T) {
	for _, decl := range []string{"my", "our", "state"} {
		src := "use feature 'state'; " + decl + " sub s2 (\\$) {} package B; s2(1);\n"
		if got := refusedAs(src); !slices.Equal(got, []RefusalCode{RefScalarSlot}) {
			t.Errorf("%q: refusals %v, want one %s", src, got, RefScalarSlot)
		}
	}
	wantNoRefusal(t, "use builtin 'refaddr'; package P; my $x; my $a = refaddr $x, 1;\n")
}

// TestGlobAliasCarriesPrototype: `*bar::ff = *ff` in a BEGIN makes bar::ff
// main::ff, prototype and all -- measured on 5.42.0, `package bar; ff(1)`
// then dies "Type of arg 1 to main::ff must be scalar". op/lexsub.t and
// op/lvref.t alias test.pl's `is` into another package this way.
func TestGlobAliasCarriesPrototype(t *testing.T) {
	for _, src := range []string{
		"sub ff (\\$) {} BEGIN { *bar::ff = *ff } package bar; ff(1);\n",
		"sub ff (\\$) {} package bar { BEGIN { *ff = *main::ff } ff(1); }\n",
	} {
		if got := refusedAs(src); !slices.Equal(got, []RefusalCode{RefScalarSlot}) {
			t.Errorf("%q: refusals %v, want one %s", src, got, RefScalarSlot)
		}
	}
}

// TestIsaEndsAnArgumentList: under the isa feature, `isa` is infix and takes
// no argument's place. Measured on 5.42.0 with -MO=Deparse,-p, `not(undef
// isa "BaseClass")` is `(!((undef) isa 'BaseClass'))` (op/isa.t), and
// `shift isa "X"` is `(shift(@ARGV) isa 'X')`.
func TestIsaEndsAnArgumentList(t *testing.T) {
	wantNoRefusal(t,
		"use feature 'isa'; my $r = not(undef isa \"BaseClass\");\n",
		"use feature 'isa'; package C { sub isa {} } my $r = not(undef isa \"BaseClass\");\n",
	)
}

// TestPrintWordBeforePackageIsAMethod: `print WORD PACKAGE ...` with WORD
// no sub of the current package is an indirect method call, not a
// filehandle. Measured on 5.42.0 with -MO=Deparse: `package Bar; sub x {}
// package main; print FOO Bar "x";` is `print 'Bar'->FOO('x');`. PerlOnJava
// unit/print_indirect_method_postfix.t declares its `create` inside the
// class's own package block.
func TestPrintWordBeforePackageIsAMethod(t *testing.T) {
	wantNoRefusal(t,
		"package Bar; sub x {} package main; print FOO Bar \"x\";\n",
		"{ package Bar; sub create {} } print create Bar sub { 1 };\n",
	)
}
