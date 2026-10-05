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
