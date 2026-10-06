// ABOUTME: Tests for typed Perl in .pmt declaration files: typed parameters and their defaults.
// ABOUTME: Typed Perl is read only from a declaration file; ordinary source keeps perl's reading.
package parse

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/types"
)

// TestPmtTypedParameters: RFC 0001 "Typed Perl, in `.pmt` only". A .pmt
// declaration is `sub NAME :attrs (Type $v = default, ...)`, its prototype
// in `:prototype(...)` as perl requires once a signature is present. Each
// parameter carries the lattice type its name resolves to and the default
// as written; a scalar with no default is required.
func TestPmtTypedParameters(t *testing.T) {
	facts := readDeclaration([]byte("sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__);\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	if got := facts.protos["bless"]; got != "($;$)" {
		t.Errorf("prototype: got %q, want %q", got, "($;$)")
	}
	want := types.Signature{Params: []types.Param{
		{Name: "ref", Sigil: '$', Type: types.Ref, Required: true},
		{Name: "class", Sigil: '$', Type: types.Str, Default: "__PACKAGE__"},
	}}
	if got, ok := facts.signatures["bless"]; !ok || !reflect.DeepEqual(got, want) {
		t.Errorf("signature: got %+v (recorded %v), want %+v", got, ok, want)
	}
}

// TestPmtSlurpyDefaultsParse: a default on a slurpy is a typed-Perl
// extension (perl's signatures reject it, "A slurpy parameter may not have a
// default value"), used for `print`'s `= ($_)` and `sort`'s `= die`. RFC 0001
// "A required argument defaults to `die`": a `die` default runs only when
// its argument is omitted, so it makes the parameter required, and is not
// recorded as a default. A slurpy with no default accepts zero arguments, as
// perl's own does.
func TestPmtSlurpyDefaultsParse(t *testing.T) {
	facts := readDeclaration([]byte("sub f (List @l = ($_));\nsub g (List @list = die);\nsub h (List @list);\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	for name, want := range map[string]types.Param{
		"f": {Name: "l", Sigil: '@', Type: types.List, Default: "($_)"},
		"g": {Name: "list", Sigil: '@', Type: types.List, Required: true},
		"h": {Name: "list", Sigil: '@', Type: types.List},
	} {
		got, ok := facts.signatures[name]
		if !ok || !reflect.DeepEqual(got, types.Signature{Params: []types.Param{want}}) {
			t.Errorf("%s: got %+v (recorded %v), want one param %+v", name, got, ok, want)
		}
	}
}

// TestTypedSyntaxOnlyInPmt: typed Perl is read only from a declaration file.
// In ordinary source, without the signatures feature, the parens after `sub
// NAME` are a prototype whatever they hold. Measured on 5.42.0, `sub f (Ref
// $x) { 1 } print prototype(\&f)` prints `Ref $x`.
func TestTypedSyntaxOnlyInPmt(t *testing.T) {
	src := []byte("sub f (Ref $x) { 1 }\n")
	facts := readModule(Parse(src))
	if got := facts.protos["f"]; got != "(Ref $x)" {
		t.Errorf("prototype: got %q, want %q", got, "(Ref $x)")
	}
	if len(facts.signatures) > 0 || len(facts.errs) > 0 {
		t.Errorf("ordinary source recorded types: %v, errors %v", facts.signatures, facts.errs)
	}
	if typed := readDeclaration(src, nil); len(typed.signatures["f"].Params) != 1 {
		t.Errorf("the same text in a .pmt: got %+v, want one typed param", typed.signatures)
	}
}

// TestTypedSyntaxNotEnabledBySignaturesFeature: the signatures feature puts
// the lexer in the signature mode a .pmt uses, but not the parser in typed
// Perl. In a .pm under `use v5.36`, `sub f (Ref $x) { 1 }` is a signature
// perl refuses -- measured on 5.42.0, "A signature parameter must start with
// '$', '@' or '%'" -- and it records no types.
func TestTypedSyntaxNotEnabledBySignaturesFeature(t *testing.T) {
	lib := t.TempDir()
	pm := "package Pm;\nuse v5.36;\nsub f (Ref $x) { 1 }\n1;\n"
	if err := os.WriteFile(filepath.Join(lib, "Pm.pm"), []byte(pm), 0o644); err != nil {
		t.Fatal(err)
	}
	r := &resolver{load: DirLoader(lib), seen: map[string]bool{}}
	facts, ok := r.resolve("Pm")
	if !ok {
		t.Fatal("Pm.pm not reached")
	}
	if len(facts.signatures) > 0 || len(facts.errs) > 0 {
		t.Errorf("a .pm under use v5.36 recorded types: %v, errors %v", facts.signatures, facts.errs)
	}
}

// TestPmtMalformedTypedDeclarationIsError: a typed signature that cannot be
// read is an error that says what is wrong, not a panic and not a partial
// record. The tree still covers every byte.
func TestPmtMalformedTypedDeclarationIsError(t *testing.T) {
	for src, want := range map[string]string{
		"sub f (Str $x":        "sub f: signature not terminated",
		"sub f (Str);\n":       "sub f: type Str names no variable",
		"sub f (Str $x = );\n": "sub f: default for $x has no expression",
	} {
		facts := readDeclaration([]byte(src), nil)
		if len(facts.errs) != 1 || facts.errs[0].Error() != want {
			t.Errorf("%q: got errors %v, want %q", src, facts.errs, want)
		}
		if sig, ok := facts.signatures["f"]; ok {
			t.Errorf("%q: recorded %+v", src, sig)
		}
		if root, _ := parseSource([]byte(src), nil, true); root.SourceText([]byte(src)) != src {
			t.Errorf("%q: round trip lost bytes", src)
		}
	}
}

// TestPmtReturnType: RFC 0001 "Typed Perl, in `.pmt` only". After the
// signature's closing paren a declaration states its return type, where a
// definition would have its body: `sub NAME :attrs (Type $v, ...) Type;`.
func TestPmtReturnType(t *testing.T) {
	facts := readDeclaration([]byte("sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__) Object;\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	sig, ok := facts.signatures["bless"]
	if !ok || sig.Returns != types.Object {
		t.Errorf("return type: got %+v (recorded %v), want %v", sig, ok, types.Object)
	}
	if len(sig.Params) != 2 {
		t.Errorf("params: got %+v, want two", sig.Params)
	}
	if got := facts.protos["bless"]; got != "($;$)" {
		t.Errorf("prototype: got %q, want %q", got, "($;$)")
	}
}

// TestPmtUnionAndContainerTypes: RFC 0001 "Type names" spells a union `A|B`,
// spaces around `|` allowed, and "A slurpy takes no bare element type" gives a
// slurpy a container type, `List[Str] @args`. The container is the param's
// type and its element is recorded beside it.
func TestPmtUnionAndContainerTypes(t *testing.T) {
	facts := readDeclaration([]byte("sub f (Str|Undef $x, Str | Undef $y) Str | Undef;\nsub g (List[Str] @args);\nsub h (Str $x) None;\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	for name, want := range map[string]types.Signature{
		"f": {Params: []types.Param{
			{Name: "x", Sigil: '$', Type: types.Str | types.Undef, Required: true},
			{Name: "y", Sigil: '$', Type: types.Str | types.Undef, Required: true},
		}, Returns: types.Str | types.Undef},
		"g": {Params: []types.Param{{Name: "args", Sigil: '@', Type: types.List, Element: types.Str}}},
		"h": {Params: []types.Param{{Name: "x", Sigil: '$', Type: types.Str, Required: true}}, Returns: types.None},
	} {
		if got, ok := facts.signatures[name]; !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %+v (recorded %v), want %+v", name, got, ok, want)
		}
	}
}

// TestPmtUnknownTypeNameIsError: RFC 0001 "Type names", a `.pmt` names types
// from the lattice and an unknown name is an error naming it, alone or as a
// member of a union. Names are the paper's, so `Bool` is unknown too: the
// lattice's spelling is `Boolean`. Nothing is recorded for the sub.
func TestPmtUnknownTypeNameIsError(t *testing.T) {
	for src, want := range map[string]string{
		"sub f (Strng $x);\n":     `sub f: unknown type name "Strng"`,
		"sub f (Str|Strng $x);\n": `sub f: unknown type name "Strng"`,
		"sub f (Bool $x);\n":      `sub f: unknown type name "Bool"`,
	} {
		facts := readDeclaration([]byte(src), nil)
		if len(facts.errs) != 1 || facts.errs[0].Error() != want {
			t.Errorf("%q: got errors %v, want %q", src, facts.errs, want)
		}
		if sig, ok := facts.signatures["f"]; ok {
			t.Errorf("%q: recorded %+v", src, sig)
		}
	}
}

// TestPmtSignatureBeforeAttributesRefused: RFC 0001 "Declaration order:
// Perl's", name, then attributes, then signature. Measured on 5.42.0 under
// `use v5.36`, `sub f :lvalue ($x) {}` compiles and both `sub f ($x) :lvalue
// {}` and `sub f ($x) :lvalue;` die "Subroutine attributes must come before
// the signature". A `.pmt` refuses the same order with perl's words, and
// records no signature for it.
func TestPmtSignatureBeforeAttributesRefused(t *testing.T) {
	src := "sub f (Str $x) :lvalue;\n"
	facts := readDeclaration([]byte(src), nil)
	want := "sub f: subroutine attributes must come before the signature"
	if len(facts.errs) != 1 || facts.errs[0].Error() != want {
		t.Errorf("got errors %v, want %q", facts.errs, want)
	}
	if sig, ok := facts.signatures["f"]; ok {
		t.Errorf("recorded %+v", sig)
	}
	if root, _ := parseSource([]byte(src), nil, true); root.SourceText([]byte(src)) != src {
		t.Errorf("round trip lost bytes")
	}
	if ok := readDeclaration([]byte("sub f :lvalue (Str $x);\n"), nil); len(ok.errs) > 0 || len(ok.signatures["f"].Params) != 1 {
		t.Errorf("perl's order: got errors %v, signature %+v", ok.errs, ok.signatures)
	}
}

// pmtRefuses reads each source as a `.pmt` and wants exactly the one error
// given for it, no signature recorded for f, and every byte still in the tree.
func pmtRefuses(t *testing.T, cases map[string]string) {
	t.Helper()
	for src, want := range cases {
		facts := readDeclaration([]byte(src), nil)
		if len(facts.errs) != 1 || facts.errs[0].Error() != want {
			t.Errorf("%q: got errors %v, want %q", src, facts.errs, want)
		}
		if sig, ok := facts.signatures["f"]; ok {
			t.Errorf("%q: recorded %+v", src, sig)
		}
		if root, _ := parseSource([]byte(src), nil, true); root.SourceText([]byte(src)) != src {
			t.Errorf("%q: round trip lost bytes", src)
		}
	}
}

// TestPmtUnknownReturnTypeIsError: an unknown name as the return type is the
// error an unknown parameter type is (RFC 0001, "Type names").
func TestPmtUnknownReturnTypeIsError(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f (Str $x) Strng;\n":     `sub f: unknown type name "Strng"`,
		"sub f (Str $x) Str|Strng;\n": `sub f: unknown type name "Strng"`,
	})
}

// TestPmtMalformedReturnTypeIsError: a return type that cannot be read is an
// error that says what is wrong, not a panic.
func TestPmtMalformedReturnTypeIsError(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f (Str $x) Str":         "sub f: return type Str not terminated",
		"sub f (Str $x) Str|;\n":     "sub f: union Str| has no type after `|`",
		"sub f (Str $x) Str { 1 }\n": "sub f: return type Str is not followed by `;`",
	})
}

// TestPmtMalformedContainerTypeRefused: a container type that cannot be read
// is an error naming the bad part, not a panic (RFC 0001, "A slurpy takes no
// bare element type").
func TestPmtMalformedContainerTypeRefused(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f (List[Str @args);\n":    "sub f: container type List[Str is not closed by `]`",
		"sub f (List[] @args);\n":      "sub f: container type List[] has no element type",
		"sub f (List[Strng] @args);\n": `sub f: unknown type name "Strng"`,
		"sub f (Strng[Str] @args);\n":  `sub f: unknown type name "Strng"`,
	})
}

// TestContainerTypeOnlyInPmt: container types are typed Perl, read only from
// a declaration file. In ordinary source, without the signatures feature,
// the parens are a prototype whatever they hold. Measured on 5.42.0, `sub f
// (List[Str] @args) { 1 } print prototype(\&f)` prints `List[Str] @args`.
func TestContainerTypeOnlyInPmt(t *testing.T) {
	src := []byte("sub f (List[Str] @args) { 1 }\n")
	facts := readModule(Parse(src))
	if got := facts.protos["f"]; got != "(List[Str] @args)" {
		t.Errorf("prototype: got %q, want %q", got, "(List[Str] @args)")
	}
	if len(facts.signatures) > 0 || len(facts.errs) > 0 {
		t.Errorf("ordinary source recorded types: %v, errors %v", facts.signatures, facts.errs)
	}
}

// TestPmtUnionWithNone: None is the bottom type, so a union with it is the
// other member -- its sentinel bit must not leak into the recorded type, as
// types.Join holds for inferred types. Likewise through FromName.
func TestPmtUnionWithNone(t *testing.T) {
	facts := readDeclaration([]byte("sub f (Str|None $x) None|Str;\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	want := types.Signature{Params: []types.Param{{Name: "x", Sigil: '$', Type: types.Str, Required: true}}, Returns: types.Str}
	if got := facts.signatures["f"]; !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if got, err := types.FromName("None|Str"); err != nil || got != types.Str {
		t.Errorf("FromName(None|Str) = %v, %v; want Str", got, err)
	}
}

// TestPmtAttributesBeforeNameRefused: RFC 0001 "Declaration order", the name
// comes first. Measured on 5.42, `sub :lvalue f;` is no declaration at all
// -- perl reads the label `sub:` and the indirect call `'f'->lvalue` -- and
// under `use v5.36`, the bundle a .pmt is read with, it is a syntax error.
// Any statement a .pmt cannot read is reported, never silently skipped;
// here, as in perl, `sub:` reads as a label and the rest is unread.
func TestPmtAttributesBeforeNameRefused(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub :lvalue f;\n": `not a declaration: "lvalue f;"`,
	})
}

// TestPmtBareSlurpyTypeRefused: RFC 0001 "A slurpy takes no bare element
// type". `Str @args` is ambiguous twice over, in its container and in what
// `Str` applies to, so it is refused. A slurpy is untyped or carries an
// explicit container type.
func TestPmtBareSlurpyTypeRefused(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f (Str @args);\n": "sub f: slurpy @args has a bare element type Str; write a container type, List[Str] @args",
		"sub f (Int %h);\n":    "sub f: slurpy %h has a bare element type Int; write a container type, List[Int] %h",
	})
	facts := readDeclaration([]byte("sub f (@args);\nsub g (List[Str] @args);\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	for name, want := range map[string]types.Param{
		"f": {Name: "args", Sigil: '@', Type: types.Unknown},
		"g": {Name: "args", Sigil: '@', Type: types.List, Element: types.Str},
	} {
		if got, ok := facts.signatures[name]; !ok || !reflect.DeepEqual(got, types.Signature{Params: []types.Param{want}}) {
			t.Errorf("%s: got %+v (recorded %v), want one param %+v", name, got, ok, want)
		}
	}
}

// TestPmtBareUnionSlurpyRefused: a union is still a bare element type, so
// `Str|Undef @args` is refused as `Str @args` is (RFC 0001, "A slurpy takes
// no bare element type").
func TestPmtBareUnionSlurpyRefused(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f (Str|Undef @args);\n": "sub f: slurpy @args has a bare element type Str|Undef; write a container type, List[Str|Undef] @args",
	})
}

// TestPmtContainerWithFlatteningSigilRefused: RFC 0001 "The scalar
// container", `Array @a` (a container type with a flattening sigil) is not
// valid. A parameter that takes the caller's container is backslashed,
// `Array \@a`, and the error names that spelling. `Array @a` is not quietly
// read as the aliased array: it records no parameter and derives no
// prototype.
func TestPmtContainerWithFlatteningSigilRefused(t *testing.T) {
	cases := map[string]string{
		"sub f (Array @a);\n": `sub f: container type Array with flattening sigil @a; a parameter that takes the caller's container is Array \@a`,
		"sub f (Hash %h);\n":  `sub f: container type Hash with flattening sigil %h; a parameter that takes the caller's container is Hash \%h`,
	}
	pmtRefuses(t, cases)
	for src := range cases {
		if proto := readDeclaration([]byte(src), nil).protos["f"]; proto != "" {
			t.Errorf("%q: derived prototype %q", src, proto)
		}
	}
}
