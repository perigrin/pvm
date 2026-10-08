// ABOUTME: Tests for typed Perl in .pmt declaration files: typed parameters and their defaults.
// ABOUTME: Typed Perl is read only from a declaration file; ordinary source keeps perl's reading.
package parse

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

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
	if got, ok := facts.signatures["bless"]; !ok || !reflect.DeepEqual(got, []types.Signature{want}) {
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
		if !ok || !reflect.DeepEqual(got, []types.Signature{{Params: []types.Param{want}}}) {
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
	if typed := readDeclaration(src, nil); len(typed.signatures["f"]) != 1 || len(typed.signatures["f"][0].Params) != 1 {
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
	sigs, ok := facts.signatures["bless"]
	if !ok || len(sigs) != 1 {
		t.Fatalf("got %+v, want one signature", sigs)
	}
	sig := sigs[0]
	if sig.Returns != types.Object {
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
		if got, ok := facts.signatures[name]; !ok || !reflect.DeepEqual(got, []types.Signature{want}) {
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
	if ok := readDeclaration([]byte("sub f :lvalue (Str $x);\n"), nil); len(ok.errs) > 0 || len(ok.signatures["f"]) != 1 || len(ok.signatures["f"][0].Params) != 1 {
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
	if got := facts.signatures["f"]; !reflect.DeepEqual(got, []types.Signature{want}) {
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
	facts := readDeclaration([]byte("sub f (@args);\nsub g (List[Str] @args);\nsub h (List[Str] %h);\nsub i (Array|Hash|Scalar @args);\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	for name, want := range map[string]types.Param{
		"f": {Name: "args", Sigil: '@', Type: types.Unknown},
		"g": {Name: "args", Sigil: '@', Type: types.List, Element: types.Str},
		"h": {Name: "h", Sigil: '%', Type: types.List, Element: types.Str},
		"i": {Name: "args", Sigil: '@', Type: types.Array | types.Hash | types.Scalar},
	} {
		if got, ok := facts.signatures[name]; !ok || !reflect.DeepEqual(got, []types.Signature{{Params: []types.Param{want}}}) {
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
		"sub f (Array @a);\n":     `sub f: container type Array with flattening sigil @a; a parameter that takes the caller's container is Array \@a`,
		"sub f (Hash %h);\n":      `sub f: container type Hash with flattening sigil %h; a parameter that takes the caller's container is Hash \%h`,
		"sub f (Hash[Str] %h);\n": `sub f: container type Hash[Str] with flattening sigil %h; a parameter that takes the caller's container is Hash[Str] \%h`,
	}
	pmtRefuses(t, cases)
	for src := range cases {
		if proto := readDeclaration([]byte(src), nil).protos["f"]; proto != "" {
			t.Errorf("%q: derived prototype %q", src, proto)
		}
	}
}

// TestPmtNonFinalListParam: RFC 0001 "A typed signature and a prototype say
// the same thing", a List parameter must be last. Measured on 5.42.0, the
// signature `(@a, $x)` dies "Slurpy parameter not last", `(@a, @b)` dies
// "Multiple slurpy parameters not allowed", and the prototype `(@$)` warns
// "Prototype after '@'": the `@` takes every argument. The prototype `\@$`
// passes the array whole and fills its `$`, so the error points to
// `Array \@a`. An untyped slurpy is a List parameter too.
func TestPmtNonFinalListParam(t *testing.T) {
	want := `sub f: List parameter @a is not last; a single array followed by more parameters is Array \@a`
	pmtRefuses(t, map[string]string{
		"sub f (List @a, Str $x);\n":  want,
		"sub f (List @a, List @b);\n": want,
		"sub f (@a, $x);\n":           want,
	})
}

// TestPmtModuloAfterIndexInDefault: a `%` after a container type's `]` is
// a hash sigil, `List[Str] %h`, but a `%` after an element's `]` in a
// default is still modulo.
func TestPmtModuloAfterIndexInDefault(t *testing.T) {
	facts := readDeclaration([]byte("sub f (Int $x = $a[0] % 2);\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	want := []types.Signature{{Params: []types.Param{{Name: "x", Sigil: '$', Type: types.Int, Default: "$a[0] % 2"}}}}
	if got := facts.signatures["f"]; !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// TestPmtDefaultBeforeAnotherParameter: a default ends at the comma, so a
// parameter can follow it. Measured on 5.42.0, `sub f ($x = 1, $y = $_)`
// gives "1 u" for `f()` with `$_ = "u"`.
func TestPmtDefaultBeforeAnotherParameter(t *testing.T) {
	facts := readDeclaration([]byte("sub f (Scalar $x = 1, Scalar $y = $_);\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	want := []types.Signature{{Params: []types.Param{
		{Name: "x", Sigil: '$', Type: types.Scalar, Default: "1"},
		{Name: "y", Sigil: '$', Type: types.Scalar, Default: "$_"},
	}}}
	if got := facts.signatures["f"]; !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// TestPmtCodeAndGlobParameters: RFC 0001's table gives a leading `&` as
// `Code &` and `*` as `Glob *`, the slots a prototype names by those
// characters, as `multi sub grep (Code &block, List @list) List;` writes
// one. Each is required, as a `$` is; neither flattens, so a parameter may
// follow it.
func TestPmtCodeAndGlobParameters(t *testing.T) {
	src := "sub f (Code &block, Glob *fh, List @list);\n"
	facts := readDeclaration([]byte(src), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	want := []types.Signature{{Params: []types.Param{
		{Name: "block", Sigil: '&', Type: types.Code, Required: true},
		{Name: "fh", Sigil: '*', Type: types.Glob, Required: true},
		{Name: "list", Sigil: '@', Type: types.List},
	}}}
	if got := facts.signatures["f"]; !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if root, _ := parseSource([]byte(src), nil, true); root.SourceText([]byte(src)) != src {
		t.Errorf("round trip lost bytes")
	}
}

// TestPmtMultiSubKeepsEveryCandidate: RFC 0001 "Multi declarations". A
// `multi sub` gives one of several signatures for its name, and every one is
// kept, in the order declared.
func TestPmtMultiSubKeepsEveryCandidate(t *testing.T) {
	facts := readDeclaration([]byte("multi sub select (FileHandle $fh) Str;\nmulti sub select ($r, $w, $e, Num $timeout) Int;\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	want := []types.Signature{
		{Params: []types.Param{{Name: "fh", Sigil: '$', Type: types.FileHandle, Required: true}}, Returns: types.Str},
		{Params: []types.Param{
			{Name: "r", Sigil: '$', Required: true},
			{Name: "w", Sigil: '$', Required: true},
			{Name: "e", Sigil: '$', Required: true},
			{Name: "timeout", Sigil: '$', Type: types.Num, Required: true},
		}, Returns: types.Int},
	}
	if got := facts.signatures["select"]; !reflect.DeepEqual(got, want) {
		t.Errorf("candidates: got %+v, want %+v", got, want)
	}
}

// TestPmtPlainAndMultiSubMixIsAMultiWithAWarning: RFC 0001 "Multi
// declarations". A name declared both as `sub` and as `multi sub` is a
// multi: every candidate is kept, in the order declared, whichever comes
// first, and the file carries one warning naming the sub and no error. Two
// plain `sub f` lines with no `multi sub f` keep only the last, silently.
// Once mixed, the multi rules apply: candidates whose return types answer
// different contexts fork on context.
func TestPmtPlainAndMultiSubMixIsAMultiWithAWarning(t *testing.T) {
	intSig := types.Signature{Params: []types.Param{{Name: "x", Sigil: '$', Type: types.Int, Required: true}}, Returns: types.Str}
	strSig := types.Signature{Params: []types.Param{{Name: "x", Sigil: '$', Type: types.Str, Required: true}}, Returns: types.Int}
	numSig := types.Signature{Params: []types.Param{{Name: "x", Sigil: '$', Type: types.Num, Required: true}}, Returns: types.Str}
	warning := "sub f: declared both as sub and as multi sub; read as a multi"
	for _, tc := range []struct {
		src  string
		want []types.Signature
	}{
		{"sub f (Int $x) Str;\nmulti sub f (Str $x) Int;\n", []types.Signature{intSig, strSig}},
		{"multi sub f (Int $x) Str;\nsub f (Str $x) Int;\n", []types.Signature{intSig, strSig}},
		{"sub f (Int $x) Str;\nsub f (Num $x) Str;\nmulti sub f (Str $x) Int;\n", []types.Signature{intSig, numSig, strSig}},
	} {
		facts := readDeclaration([]byte(tc.src), nil)
		if len(facts.errs) > 0 {
			t.Errorf("%q: errors %v", tc.src, facts.errs)
		}
		if len(facts.warns) != 1 || facts.warns[0].Error() != warning {
			t.Errorf("%q: got warnings %v, want %q", tc.src, facts.warns, warning)
		}
		if got := facts.signatures["f"]; !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%q: candidates got %+v, want %+v", tc.src, got, tc.want)
		}
	}

	plain := readDeclaration([]byte("sub f (Int $x) Str;\nsub f (Str $x) Int;\n"), nil)
	if len(plain.errs) > 0 || len(plain.warns) > 0 {
		t.Errorf("two plain subs: errors %v, warnings %v", plain.errs, plain.warns)
	}
	if got := plain.signatures["f"]; !reflect.DeepEqual(got, []types.Signature{strSig}) {
		t.Errorf("two plain subs: got %+v, want only the last", got)
	}

	ctx := readDeclaration([]byte("sub g () Str;\nmulti sub g () List;\n"), nil)
	if len(ctx.errs) > 0 || len(ctx.warns) != 1 {
		t.Errorf("context mix: errors %v, warnings %v", ctx.errs, ctx.warns)
	}
	if got := ctx.signatures["g"]; len(got) != 2 || got[0].Returns != types.Str || got[1].Returns != types.List {
		t.Errorf("context mix: got %+v, want a scalar and a list candidate", got)
	}
}

// TestPmtMultiCandidateRefusedAlone: a `multi sub` candidate refused for its
// declaration order (RFC 0001, "Declaration order: Perl's") is not recorded,
// and the candidates read before it are kept.
func TestPmtMultiCandidateRefusedAlone(t *testing.T) {
	facts := readDeclaration([]byte("multi sub f (Int $x) Str;\nmulti sub f (Str $x) :lvalue;\n"), nil)
	want := "sub f: subroutine attributes must come before the signature"
	if len(facts.errs) != 1 || facts.errs[0].Error() != want {
		t.Errorf("got errors %v, want %q", facts.errs, want)
	}
	if got := facts.signatures["f"]; len(got) != 1 || got[0].Params[0].Type != types.Int {
		t.Errorf("candidates: got %+v, want only the Int one", got)
	}
}

// TestMultiOnlyInPmt: `multi` declares only in a declaration file. In
// ordinary source it is the name it has always been: `sub multi { 1 }
// multi(1);` declares and calls a sub named multi, as perl 5.42.0 runs it,
// and records no candidates.
func TestMultiOnlyInPmt(t *testing.T) {
	src := []byte("sub multi { 1 } multi(1);\n")
	root := Parse(src)
	if got, want := Canon(root, src), "sub multi {1;} multi(1);"; got != want {
		t.Errorf("canon: got %q, want %q", got, want)
	}
	facts := readModule(root)
	if _, ok := facts.protos["multi"]; !ok {
		t.Errorf("sub multi not declared: %v", facts.protos)
	}
	if len(facts.signatures) > 0 || len(facts.errs) > 0 {
		t.Errorf("ordinary source recorded candidates: %v, errors %v", facts.signatures, facts.errs)
	}

	// Nor is `multi sub` a declaration there: perl 5.42.0 refuses `multi
	// sub f { 1 }`, "syntax error at -e line 1, near "multi sub f "", and
	// the parse does not accept it.
	bad := Parse([]byte("multi sub f { 1 }\n"))
	if len(bad.Children) == 0 || bad.Children[0].Kind != Unknown {
		t.Errorf("multi sub in ordinary source: got %+v, want it refused as Unknown", bad.Children)
	}
}

// TestMultiAmbiguousCandidatesIsError: RFC 0001 "Multi declarations".
// Candidates with no single most specific one for some argument types are
// ambiguous, and the declaration is an error naming the sub: two Ints fit
// both `(Int $a, Num $b)` and `(Num $a, Int $b)`, and neither is more
// specific. A third candidate for exactly two Ints, more specific than
// both, decides the call and makes the declaration valid.
func TestMultiAmbiguousCandidatesIsError(t *testing.T) {
	pair := "multi sub f (Int $a, Num $b) Str;\nmulti sub f (Num $a, Int $b) Str;\n"
	pmtRefuses(t, map[string]string{
		pair: "sub f: candidates (Int $a, Num $b) and (Num $a, Int $b) are ambiguous for (Int, Int)",
	})
	facts := readDeclaration([]byte(pair+"multi sub f (Int $a, Int $b) Str;\n"), nil)
	if len(facts.errs) > 0 || len(facts.signatures["f"]) != 3 {
		t.Errorf("with (Int $a, Int $b): got errors %v, candidates %+v", facts.errs, facts.signatures["f"])
	}
}

// TestPmtDeclarationErrorsSurface: a declaration file in error is reported,
// not dropped. `sub f (Str);` names a type and no variable, so the file is
// in error whether CORE.pmt is read for the builtin table or a module's
// declaration is read in place of its source; and every shipped declaration
// reads clean, with no error and no warning.
func TestPmtDeclarationErrorsSurface(t *testing.T) {
	bad := []byte("sub f (Str);\n")
	if _, _, err := coreProtos(bad); err == nil || err.Error() != "sub f: type Str names no variable" {
		t.Errorf("core table: got %v, want the declaration's error", err)
	}

	saved := declarations
	declarations = layeredFS{top: fstest.MapFS{"declarations/Bad.pmt": {Data: bad}}, base: saved}
	defer func() { declarations = saved }()
	// Used directly, and from a required helper.
	helper := func(name string) ([]byte, bool) { return []byte("use Bad;\n"), name == "./h.pl" }
	for _, src := range []string{"use Bad;\n", "require './h.pl';\n"} {
		root := ParseWithLoader([]byte(src), helper)
		if errs := DeclarationErrors(root); len(errs) != 1 || errs[0].Error() != "Bad: sub f: type Str names no variable" {
			t.Errorf("%q: got %v, want the declaration's error", src, errs)
		}
	}
	declarations = saved

	for _, e := range shippedDeclarationErrors(t, declarations) {
		t.Error(e)
	}
}

// TestDeclarationWarningsReachTheRoot: a declaration file's warnings travel
// beside its errors to the parse root, each naming its module, whether the
// module is used directly or from a required helper. A warning is not an
// error: the file mixing `sub f` and `multi sub f` reports none.
func TestDeclarationWarningsReachTheRoot(t *testing.T) {
	mix := []byte("package Mix;\nsub f (Int $x) Str;\nmulti sub f (Str $x) Int;\n")
	saved := declarations
	declarations = layeredFS{top: fstest.MapFS{"declarations/Mix.pmt": {Data: mix}}, base: saved}
	defer func() { declarations = saved }()
	helper := func(name string) ([]byte, bool) { return []byte("use Mix;\n"), name == "./h.pl" }
	want := "Mix: sub f: declared both as sub and as multi sub; read as a multi"
	for _, src := range []string{"use Mix;\n", "require './h.pl';\n"} {
		root := ParseWithLoader([]byte(src), helper)
		if warns := DeclarationWarnings(root); len(warns) != 1 || warns[0].Error() != want {
			t.Errorf("%q: got warnings %v, want %q", src, warns, want)
		}
		if errs := DeclarationErrors(root); len(errs) > 0 {
			t.Errorf("%q: got errors %v, want none", src, errs)
		}
	}
}

// shippedDeclarationErrors reads every declaration file in fsys as the
// resolver would -- CORE.pmt in CORE's language, every other file as a
// library's -- and returns each error and warning with its file's name.
func shippedDeclarationErrors(t *testing.T, fsys fs.FS) []string {
	t.Helper()
	var out []string
	err := fs.WalkDir(fsys, "declarations", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		src, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}
		read := readLibraryDeclaration
		if name == "declarations/CORE.pmt" {
			read = readDeclaration
		}
		facts := read(src, nil)
		for _, e := range facts.errs {
			out = append(out, fmt.Sprintf("%s: %v", name, e))
		}
		for _, w := range facts.warns {
			out = append(out, fmt.Sprintf("%s: warning: %v", name, w))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// TestShippedLibraryPmtRefusesUnclassedLevel: RFC 0001 "One language for
// every .pmt" -- a library may not join a level XS::Parse::Infix classes no
// operator at, as CORE.pmt's own operators do, so the check over the
// shipped files reads each library file as a library.
func TestShippedLibraryPmtRefusesUnclassedLevel(t *testing.T) {
	line := "sub zz :infix :equiv(&) (Int $x, Int $y) Int;\n"
	fsys := fstest.MapFS{
		"declarations/CORE.pmt": {Data: []byte(line)},
		"declarations/Foo.pmt":  {Data: []byte(line)},
	}
	errs := shippedDeclarationErrors(t, fsys)
	if len(errs) != 1 || !strings.HasPrefix(errs[0], "declarations/Foo.pmt: ") {
		t.Errorf("got %q; want one error, for the library file", errs)
	}
}

// TestPmtAliasedParamsParse: RFC 0001 "The scalar container". A parameter
// that aliases the caller's container is backslashed, as perlref's
// refaliasing writes it: `Scalar \$x`, `Array \@a`, `Hash \%h`, and a
// container type before one, `Array[Str] \@a`. Each takes exactly one
// argument, so each is required. `\(@args)` is perlref's list form, a
// reference to each element: `List[Str] \(@args = ($_))` aliases every
// argument and, like any List parameter, takes the rest of the call.
func TestPmtAliasedParamsParse(t *testing.T) {
	for src, want := range map[string]types.Param{
		"sub f (Scalar \\$x);\n":                {Name: "x", Sigil: '$', Type: types.Scalar, Alias: true, Required: true},
		"sub f (Array \\@a);\n":                 {Name: "a", Sigil: '@', Type: types.Array, Alias: true, Required: true},
		"sub f (Hash \\%h);\n":                  {Name: "h", Sigil: '%', Type: types.Hash, Alias: true, Required: true},
		"sub f (Array[Str] \\@a);\n":            {Name: "a", Sigil: '@', Type: types.Array, Element: types.Str, Alias: true, Required: true},
		"sub f (Hash[Str] \\%h);\n":             {Name: "h", Sigil: '%', Type: types.Hash, Element: types.Str, Alias: true, Required: true},
		"sub f (List[Str] \\(@args = ($_)));\n": {Name: "args", Sigil: '@', Type: types.List, Element: types.Str, AliasEach: true, Default: "($_)"},
	} {
		facts := readDeclaration([]byte(src), nil)
		if len(facts.errs) > 0 {
			t.Errorf("%q: errors %v", src, facts.errs)
			continue
		}
		if got := facts.signatures["f"]; !reflect.DeepEqual(got, []types.Signature{{Params: []types.Param{want}}}) {
			t.Errorf("%q: got %+v, want one param %+v", src, got, want)
		}
		if root, _ := parseSource([]byte(src), nil, true); root.SourceText([]byte(src)) != src {
			t.Errorf("%q: round trip lost bytes", src)
		}
	}
}

// TestPmtMalformedAliasedParamRefused: a backslash RFC 0001 "The scalar
// container" does not define is an error, not a guess. An aliased array
// or hash is the caller's container, so its type is Array or Hash; the
// list form must be closed and hold an array; and a glob slot, `\*`, has
// no spelling (open question 11): perl cannot alias a glob, measured on
// 5.42.0 `\*G = \*STDOUT` is "Can't modify reference to ref-to-glob cast".
func TestPmtMalformedAliasedParamRefused(t *testing.T) {
	cases := map[string]string{
		"sub f (Str \\@a);\n":                    `sub f: aliased \@a has type Str; the caller's array is Array \@a`,
		"sub f (Array \\%h);\n":                  `sub f: aliased \%h has type Array; the caller's hash is Hash \%h`,
		"sub f (List[Str] \\(@args, Str $x));\n": "sub f: list form \\(@args is not closed by `)`",
		"sub f (List[Str] \\(Str $x));\n":        `sub f: list form \( holds "Str"; it holds an array, as in \(@args)`,
		"sub f (Glob \\*g);\n":                   `sub f: aliased glob \*g has no spelling (RFC 0001, open question 11)`,
	}
	pmtRefuses(t, cases)
	for src := range cases {
		if proto := readDeclaration([]byte(src), nil).protos["f"]; proto != "" {
			t.Errorf("%q: derived prototype %q", src, proto)
		}
	}
}

// TestPmtSlotTypeMismatchRefused: a parameter's type is what its slot
// holds, by RFC 0001's table ("A typed signature and a prototype say the
// same thing"): a `&` or `\&` slot holds Code, a `\$` slot a Scalar type,
// and a `$` slot one scalar, so a type with no scalar member (Array, Hash,
// Code) is no `$`. A `*` slot holds a bareword handle or any scalar:
// measured on 5.42.0, `sub star (*)` receives `star(*STDOUT)` and
// `star($fh)` as GLOB references, `star(STDOUT)` and `star("f")` as
// strings, `star(@a)` as the count and `star(sub {1})` as a CODE
// reference, so a glob or any scalar type fits it and an aggregate or
// Code does not. A mismatch derives nothing.
func TestPmtSlotTypeMismatchRefused(t *testing.T) {
	cases := map[string]string{
		"sub f (Str \\&c);\n":                   `sub f: \&c has type Str; a \& slot holds Code`,
		"sub f (Str &c);\n":                     `sub f: &c has type Str; a & slot holds Code`,
		"sub f (Array *g);\n":                   `sub f: *g has type Array; a * slot holds a glob or a scalar`,
		"sub f (Code *g);\n":                    `sub f: *g has type Code; a * slot holds a glob or a scalar`,
		"sub f (Array \\$x);\n":                 `sub f: \$x has type Array; a \$ slot holds a Scalar type`,
		"sub f (Code \\$x);\n":                  `sub f: \$x has type Code; a \$ slot holds a Scalar type`,
		"sub f :prototype(\\$) (Array \\$x);\n": `sub f: \$x has type Array; a \$ slot holds a Scalar type`,
		"sub f (Array $x);\n":                   `sub f: $x has type Array; a $ slot holds a scalar`,
		"sub f (Array|Hash $x);\n":              `sub f: $x has type Array|Hash; a $ slot holds a scalar`,
	}
	pmtRefuses(t, cases)
	// Each type the table gives its slot fits it, as do the shapes
	// CORE.pmt states.
	for _, src := range []string{
		"sub f (Code &c);\n", "sub f (Code \\&c);\n", "sub f (Glob *g);\n",
		"sub f (FileHandle *g);\n", "sub f (GlobRef *g);\n", "sub f (Str *g);\n",
		"sub f (Str|FileHandle *g);\n", "sub f (Int \\$x);\n",
		"sub f (Scalar \\$x);\n", "sub f (FileHandle $fh);\n", "sub f (Any $x);\n",
		"sub f (Array|Hash|Scalar $x);\n", "sub f (Str|FileHandle $x);\n", "sub f (Ref $x);\n",
	} {
		typedSignature(t, src)
	}
}

// TestPmtMessagesPrintAlias: an error names an aliased parameter as the
// declaration writes it, with its backslash, so a suggestion never
// spells `Array @a`, which is not valid (RFC 0001, "The scalar
// container"), and `Str \(@a)` is told `List[Str] \(@a)`, keeping the
// alias.
func TestPmtMessagesPrintAlias(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f (Str \\(@a));\n":                    `sub f: slurpy \(@a) has a bare element type Str; write a container type, List[Str] \(@a)`,
		"sub f (Array \\(@a));\n":                  `sub f: container type Array with flattening sigil \(@a); a parameter that takes the caller's container is Array \@a`,
		"sub f (List[Str] \\(@a), Str $x);\n":      `sub f: List parameter \(@a) is not last; a single array followed by more parameters is Array \@a`,
		"sub f (Scalar \\$x = $y, Scalar \\$z);\n": `sub f: mandatory parameter \$z follows optional parameter \$x`,
		"sub f (Scalar \\$x Str $y);\n":            "sub f: parameter \\$x is not followed by `,` or `)`",
		"sub f (List[Str] \\(@a): Str $x);\n":      `sub f: invocant \(@a) is a List; an invocant slot holds one item`,
		"sub f (Scalar $s, Array \\@a: Str $x);\n": `sub f: invocant colon after \@a; only the first parameter is an invocant`,
		"sub f (Scalar \\$x =);\n":                 `sub f: default for \$x has no expression`,
	})
}

// TestParamAttributeRefused: the backslash replaces an earlier `:lvalue`
// parameter attribute (RFC 0001, "The scalar container"), and no
// parameter takes an attribute: perl 5.42.0 refuses `sub f ($x :lvalue)`,
// "Illegal operator following parameter in a subroutine signature". The
// declaration is in error and derives no `\$`.
func TestParamAttributeRefused(t *testing.T) {
	cases := map[string]string{
		"sub f (Scalar $x :lvalue);\n": "sub f: parameter $x is not followed by `,` or `)`",
		"sub f (Scalar $x :bogus);\n":  "sub f: parameter $x is not followed by `,` or `)`",
	}
	pmtRefuses(t, cases)
	for src := range cases {
		if proto := readDeclaration([]byte(src), nil).protos["f"]; proto != "" {
			t.Errorf("%q: derived prototype %q", src, proto)
		}
	}
}

// TestAliasedParamOnlyInPmt: the backslashed parameter is typed Perl, read
// only from a declaration file. In ordinary source without signatures the
// parens after `sub NAME` are a prototype whatever they hold: measured on
// 5.42.0, `sub f (\$x) { 1 } print prototype(\&f)` prints `\$x`.
func TestAliasedParamOnlyInPmt(t *testing.T) {
	for src, want := range map[string]string{
		"sub f (\\$x) { 1 }\n":        `(\$x)`,
		"sub f (Scalar \\$x) { 1 }\n": `(Scalar \$x)`,
	} {
		facts := readModule(Parse([]byte(src)))
		if got := facts.protos["f"]; got != want {
			t.Errorf("%q: prototype %q, want %q", src, got, want)
		}
		if len(facts.signatures) > 0 || len(facts.errs) > 0 {
			t.Errorf("%q: ordinary source recorded types: %v, errors %v", src, facts.signatures, facts.errs)
		}
	}
}

// TestPmtUnaryWithoutSignatureIsError: `:unary` says how a no-prototype
// builtin's one operand parses, so it needs the signature that states the
// operand. Without one the attribute would be dropped and the line read as
// a list operator, the parse `:unary` exists to deny.
func TestPmtUnaryWithoutSignatureIsError(t *testing.T) {
	facts := readDeclaration([]byte("sub foo :unary;\n"), nil)
	want := "sub foo: :unary needs a typed signature stating its operand"
	if len(facts.errs) != 1 || facts.errs[0].Error() != want {
		t.Errorf("got errors %v, want %q", facts.errs, want)
	}
}

// TestPmtListopWithoutSignatureIsError: `:listop` says a builtin with no
// prototype has typed positional parameters (RFC 0001, "Builtins with no
// prototype"), so a line stating no signature has nothing for it to say:
// it is an error, as `:unary` without one is.
func TestPmtListopWithoutSignatureIsError(t *testing.T) {
	facts := readDeclaration([]byte("sub foo :listop;\n"), nil)
	want := "sub foo: :listop needs a typed signature stating its parameters"
	if len(facts.errs) != 1 || facts.errs[0].Error() != want {
		t.Errorf("got errors %v, want %q", facts.errs, want)
	}
}

// TestPmtListopAndUnaryIsError: `:unary` and `:listop` are two parses, a
// named unary and a list operator, and a builtin parses as one. A line
// stating both is an error and records nothing.
func TestPmtListopAndUnaryIsError(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f :unary :listop (Scalar $x) Int;\n": "sub f: :unary and :listop are two parses; a builtin parses as one",
	})
}

// TestPmtUnaryAndListopAcrossCandidatesRefused: RFC 0001 "Builtins with no
// prototype" -- `:listop` beside `:unary` is two parses for one builtin,
// an error whether one line states both or a multi's candidates split
// them, in either order; otherwise the shape would follow line order.
func TestPmtUnaryAndListopAcrossCandidatesRefused(t *testing.T) {
	unary := "multi sub f :unary (Scalar $x) Int;\n"
	listop := "multi sub f :listop (Scalar $x) List;\n"
	want := "sub f: :unary and :listop are two parses; a builtin parses as one"
	for _, src := range []string{unary + listop, listop + unary} {
		facts := readDeclaration([]byte(src), nil)
		if len(facts.errs) != 1 || facts.errs[0].Error() != want {
			t.Errorf("%q: got errors %v, want %q", src, facts.errs, want)
		}
		if sig, ok := facts.signatures["f"]; ok {
			t.Errorf("%q: recorded %+v", src, sig)
		}
	}
}

// wantReturns checks that a call to cands with args in context ctx selects
// a candidate returning want.
func wantReturns(t *testing.T, what string, cands []types.Signature, args []types.Type, ctx types.Context, want types.Type) {
	t.Helper()
	if sel := types.Select(cands, args, ctx); sel.Outcome != types.Selected || sel.Returns != want {
		t.Errorf("%s in %v: selects %+v; want %v", what, ctx, sel, want)
	}
}

// TestContextSelectsByReturnType: RFC 0001 "Context selects by return
// type". A call site's context is a demand on its result, so of candidates
// with the same parameters, list context selects the one whose return type
// is a list and scalar context the one whose return type is a scalar.
// reverse is the pure case: measured on 5.42, `my $s = reverse "ab", "cd"`
// is "dcba" and `my @r = reverse "ab", "cd"` is ("cd", "ab"). CORE.pmt's
// split, each, keys and localtime fork the same way.
func TestContextSelectsByReturnType(t *testing.T) {
	facts := readDeclaration([]byte("multi sub rev (List @list) Str;\nmulti sub rev (List @list) List;\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("rev: errors %v", facts.errs)
	}
	strs := []types.Type{types.Str, types.Str}
	wantReturns(t, "rev", facts.signatures["rev"], strs, types.ScalarCtx, types.Str)
	wantReturns(t, "rev", facts.signatures["rev"], strs, types.ListCtx, types.List)

	core := coreSignatures()
	for _, c := range []struct {
		name         string
		args         []types.Type
		scalar, list types.Type
	}{
		{"reverse", strs, types.Str, types.List},
		{"split", strs, types.Int, types.List},
		{"each", []types.Type{types.Hash}, types.Str | types.Undef, types.List},
		{"each", []types.Type{types.Array}, types.Int | types.Undef, types.List},
		{"keys", []types.Type{types.Hash}, types.Int, types.List},
		{"keys", []types.Type{types.Array}, types.Int, types.List},
		{"localtime", []types.Type{types.Int}, types.Str | types.Undef, types.List},
	} {
		wantReturns(t, c.name, core[c.name], c.args, types.ScalarCtx, c.scalar)
		wantReturns(t, c.name, core[c.name], c.args, types.ListCtx, c.list)
	}
}

// TestVoidContextSelectsVoidReturn: a candidate returning Void, the empty
// list, is the one void context selects, over candidates returning a
// scalar or a list. Void context is a form of scalar context (perlglossary,
// "void context"), so with no Void candidate it selects the scalar one:
// localtime in void context is its Str|Undef candidate's.
func TestVoidContextSelectsVoidReturn(t *testing.T) {
	src := "multi sub f (Int $x) Void;\nmulti sub f (Int $x) Str;\nmulti sub f (Int $x) List;\n"
	facts := readDeclaration([]byte(src), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors %v", facts.errs)
	}
	ints := []types.Type{types.Int}
	wantReturns(t, "f", facts.signatures["f"], ints, types.VoidCtx, types.Void)
	wantReturns(t, "f", facts.signatures["f"], ints, types.ScalarCtx, types.Str)
	wantReturns(t, "f", facts.signatures["f"], ints, types.ListCtx, types.List)
	wantReturns(t, "localtime", coreSignatures()["localtime"], ints, types.VoidCtx, types.Str|types.Undef)
}

// TestReturnTypeContextAmbiguity: RFC 0001 "Multi declarations". Two
// candidates with the same parameters whose return types answer the same
// context share every call, so the declaration is ambiguous: two lists,
// two scalars or two Voids. A list beside a scalar is a context fork.
func TestReturnTypeContextAmbiguity(t *testing.T) {
	ambiguous := "sub f: candidates (Int $x) and (Int $x) are ambiguous for (Int)"
	pmtRefuses(t, map[string]string{
		"multi sub f (Int $x) List;\nmulti sub f (Int $x) List[Str];\n": ambiguous,
		"multi sub f (Int $x) Str;\nmulti sub f (Int $x) Int;\n":        ambiguous,
		"multi sub f (Int $x) Void;\nmulti sub f (Int $x) Void;\n":      ambiguous,
	})
	fork := readDeclaration([]byte("multi sub f (Int $x) Str;\nmulti sub f (Int $x) List[Str];\n"), nil)
	if len(fork.errs) > 0 || len(fork.signatures["f"]) != 2 {
		t.Errorf("scalar and list: errors %v, candidates %+v", fork.errs, fork.signatures["f"])
	}
}

// TestContextAttributeRefused: a candidate's return type says which
// context selects it, so a `.pmt` line stating `:context(...)` is a
// declaration error naming the attribute.
func TestContextAttributeRefused(t *testing.T) {
	refused := "sub f: :context is no attribute; a candidate's return type says which context selects it"
	pmtRefuses(t, map[string]string{
		"sub f :context($) (Str $x) Str;\n":        refused,
		"multi sub f :context(@) (Str $x) List;\n": refused,
		"sub f :context() (Str $x) Void;\n":        refused,
		"sub f :context(%) (Str $x) Str;\n":        refused,
		"sub f :context($) ;\n":                    refused,
		"sub f :context($ (Str $x) Str;\n":         refused,
	})
}
