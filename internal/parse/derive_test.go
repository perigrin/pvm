// ABOUTME: Tests for deriving a .pmt declaration's prototype from its types and its types from its prototype.
// ABOUTME: Expected values are RFC 0001's table under "A typed signature and a prototype say the same thing".
package parse

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/types"
)

// Each table row's parameter, as a prototype-only declaration derives it.
// A derived parameter has no name: a prototype names none. Measured on
// 5.42.0, `f()` is "Not enough arguments" under `$`, `+`, `&`, `*` and
// `$;$`, and compiles under `_`, `@` and `%`.
var (
	protoScalar   = types.Param{Sigil: '$', Type: types.Scalar, Required: true}
	protoOptional = types.Param{Sigil: '$', Type: types.Scalar}
	protoList     = types.Param{Sigil: '@', Type: types.List}
	protoHash     = types.Param{Sigil: '%', Type: types.List}
	protoPlus     = types.Param{Sigil: '$', Type: types.Array | types.Hash | types.Scalar, Required: true}
	protoCode     = types.Param{Sigil: '&', Type: types.Code, Required: true}
	protoGlob     = types.Param{Sigil: '*', Type: types.Glob, Required: true}
	protoTopic    = types.Param{Sigil: '$', Type: types.Scalar, Default: "$_"}
)

// TestDeriveTypesFromPrototype: RFC 0001, "A typed signature and a
// prototype say the same thing". A declaration with only a prototype gets
// the coarse typed signature the table gives, one case per row that takes
// no backslash: `sub foo :prototype($$)` reads as `(Scalar $, Scalar $)`.
func TestDeriveTypesFromPrototype(t *testing.T) {
	for proto, want := range map[string][]types.Param{
		"$$":  {protoScalar, protoScalar},
		"@":   {protoList},
		"%":   {protoHash},
		"+":   {protoPlus},
		"&@":  {protoCode, protoList},
		"*":   {protoGlob},
		"_":   {protoTopic},
		"$;$": {protoScalar, protoOptional},
	} {
		facts := readDeclaration([]byte("sub f :prototype("+proto+");\n"), nil)
		if len(facts.errs) > 0 {
			t.Errorf("(%s): errors %v", proto, facts.errs)
			continue
		}
		if got := facts.signatures["f"]; !reflect.DeepEqual(got, []types.Signature{{Params: want}}) {
			t.Errorf("(%s): got %+v, want %+v", proto, got, want)
		}
	}
}

// typedSignature reads one typed declaration's signature, failing the
// test on any error.
func typedSignature(t *testing.T, src string) types.Signature {
	t.Helper()
	facts := readDeclaration([]byte(src), nil)
	if len(facts.errs) > 0 || len(facts.signatures["f"]) != 1 {
		t.Fatalf("%q: errors %v, signatures %+v", src, facts.errs, facts.signatures)
	}
	return facts.signatures["f"][0]
}

// TestDerivePrototypeFromTypes: a typed declaration gets its prototype from
// its types and needs no `:prototype(...)`, one case per row that takes no
// backslash. Types are finer than prototypes, so `Str $x` gives `$` as
// `Scalar $x` does; and prototype to types to prototype round-trips for
// every row.
func TestDerivePrototypeFromTypes(t *testing.T) {
	for typed, want := range map[string]string{
		"(Scalar $x, Str $y)":             "$$",
		"(List @l)":                       "@",
		"(List %h)":                       "%",
		"(Array|Hash|Scalar $x)":          "+",
		"(Code &block, List @l)":          "&@",
		"(Glob *fh)":                      "*",
		"(Scalar $x = $_)":                "_",
		"(Scalar $x, Scalar $y = 1)":      "$;$",
		"(Scalar $x = $_, Int $y = 0)":    "_;$",
		"(Scalar $x = 1, Scalar $y = $_)": ";$_",
	} {
		if got := prototypeFromTypes(typedSignature(t, "sub f "+typed+";\n")); got != want {
			t.Errorf("%s: got %q, want %q", typed, got, want)
		}
	}

	facts := readDeclaration([]byte("sub f (Scalar $x, Str $y);\n"), nil)
	if len(facts.errs) > 0 || facts.protos["f"] != "($$)" {
		t.Errorf("declaration: errors %v, prototype %q, want ($$)", facts.errs, facts.protos["f"])
	}

	for _, proto := range []string{"$$", "@", "%", "+", "&@", "*", "_", "$;$", "_;$", ";$_", ""} {
		sig, derived, err := typesFromPrototype(proto)
		if err != nil || !derived {
			t.Errorf("(%s): derived %v, error %v", proto, derived, err)
			continue
		}
		if back := prototypeFromTypes(sig); back != proto {
			t.Errorf("(%s): round trip gave %q", proto, back)
		}
	}
}

// TestDeriveDieDefaultIsRequired: RFC 0001 "A required argument defaults
// to `die`". A `die` default runs only when its argument is omitted, so
// the parameter is required and derivation puts no `;` before it; only a
// default other than `die` makes a parameter optional.
func TestDeriveDieDefaultIsRequired(t *testing.T) {
	for typed, want := range map[string]string{
		"(Scalar $x = die)":            "$",
		"(Scalar $x = 1)":              ";$",
		"(Scalar $x, Scalar $y = die)": "$$",
	} {
		if got := prototypeFromTypes(typedSignature(t, "sub f "+typed+";\n")); got != want {
			t.Errorf("%s: got %q, want %q", typed, got, want)
		}
	}
}

// TestDerivedAtMeansNoPrototype: a typed declaration whose types derive
// `@` has no prototype, recorded as perl records a sub with none -- the
// empty prototype string. Measured on 5.42.0, `prototype(\&f)` is undef
// for `sub f;`, and a sub with none parses as one with `@` (RFC 0001,
// "Builtins that keep their own parse").
func TestDerivedAtMeansNoPrototype(t *testing.T) {
	facts := readDeclaration([]byte("sub f (List @args);\n"), nil)
	if proto, ok := facts.protos["f"]; len(facts.errs) > 0 || !ok || proto != "" {
		t.Errorf("errors %v, prototype %q (declared %v), want none", facts.errs, proto, ok)
	}
}

// TestDeriveEmptyPrototypeIsNotAbsent: as with perl's `prototype`, absent
// and empty differ. `sub f :prototype();` takes no arguments and derives
// no parameters; `sub g;` has no prototype and derives `List @`, the rest
// of the call. A declaration whose typed signature is in error derives
// nothing: it has a signature, just not one that can be read.
func TestDeriveEmptyPrototypeIsNotAbsent(t *testing.T) {
	facts := readDeclaration([]byte("sub f :prototype();\nsub g;\nsub h (Str);\n"), nil)
	if f := facts.signatures["f"]; len(f) != 1 || len(f[0].Params) != 0 {
		t.Errorf("(): got %+v, want one signature with no parameters", f)
	}
	if g := facts.signatures["g"]; !reflect.DeepEqual(g, []types.Signature{{Params: []types.Param{protoList}}}) {
		t.Errorf("none: got %+v, want (List @)", g)
	}
	if h, ok := facts.signatures["h"]; ok {
		t.Errorf("in error: got %+v, want no signature", h)
	}
}

// TestPrototypeTypeDisagreementIsError: a declaration with both a
// prototype and types must have them agree, or the declaration file is
// in error. `(Scalar $x)` gives `$`, which is not `$$`; `(Str $x)` gives
// `$`, which is `:prototype($)`, since types are finer than prototypes.
func TestPrototypeTypeDisagreementIsError(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f :prototype($$) (Scalar $x);\n": "sub f: :prototype($$) disagrees with its types, which give ($)",
	})
	facts := readDeclaration([]byte("sub f :prototype($) (Str $x);\n"), nil)
	if len(facts.errs) > 0 || len(facts.signatures["f"]) != 1 || facts.protos["f"] != "($)" {
		t.Errorf("agreeing: errors %v, signatures %+v, prototype %q", facts.errs, facts.signatures["f"], facts.protos["f"])
	}
}

// TestPrototypeTypeDisagreementKinds: a disagreement is caught whatever it
// is about, not only the count of parameters: optionality, a `;` before a
// parameter with no default, and the slot, a Code `&` where the types
// state a Scalar `$`.
func TestPrototypeTypeDisagreementKinds(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f :prototype($;$) (Scalar $x, Scalar $y);\n": "sub f: :prototype($;$) disagrees with its types, which give ($$)",
		"sub f :prototype(&@) (Scalar $f, List @l);\n":    "sub f: :prototype(&@) disagrees with its types, which give ($@)",
	})
}

// TestDeriveRejectsUnknownPrototypeCharacter: a prototype character
// outside the table is an error in the declaration file, not a guessed
// signature. perl 5.42.0 only warns, "Illegal character in prototype for
// main::f : Q", and for `\x` the same.
func TestDeriveRejectsUnknownPrototypeCharacter(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f :prototype(Q);\n":    `sub f: prototype (Q) has "Q", which is not a prototype character`,
		"sub f :prototype($\\x);\n": `sub f: prototype ($\x) has "\\x", which is not a prototype character`,
	})
}

// TestDeriveMandatoryAfterOptionalIsError: a required parameter after an
// optional one has no prototype, since everything after a `;` is
// optional, and is an error as perl 5.42.0 makes it one: `sub f ($x = 1,
// $y) {}` dies "Mandatory parameter follows optional parameter".
func TestDeriveMandatoryAfterOptionalIsError(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f (Str $x = 1, Str $y);\n":      "sub f: mandatory parameter $y follows optional parameter $x",
		"sub f (Scalar $x = $_, Code &c);\n": "sub f: mandatory parameter &c follows optional parameter $x",
	})
}

// TestDerivePrototypeAfterAtIsError: a prototype with anything after `@`
// or `%` derives no signature: the `@` takes every argument, so what
// follows can never be filled (perl 5.42.0 warns "Prototype after '@'").
// It is the final-List error, pointing to `\@` for a single array.
func TestDerivePrototypeAfterAtIsError(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f :prototype(@$);\n":   `sub f: prototype (@$): List parameter @ is not last; a single array followed by more parameters is Array \@, as in (\@$)`,
		"sub f :prototype($%;$);\n": `sub f: prototype ($%;$): List parameter % is not last; a single hash followed by more parameters is Hash \%, as in ($\%;$)`,
	})
}

// TestDerivePlusOnlyFromList: `+` is the table's `Array|Hash|Scalar $x`
// and nothing wider or narrower. Measured on 5.42, `f(@a)` passes an ARRAY
// reference under `+` and the count under `$`, so a `$` parameter typed Any
// or Array|Str is a `$`, not a `+`.
func TestDerivePlusOnlyFromList(t *testing.T) {
	for typed, want := range map[string]string{
		"(Any $x)":       "$",
		"(Array|Str $x)": "$",
		"(Scalar $x)":    "$",
		"(List $x)":      "+",
	} {
		if got := prototypeFromTypes(typedSignature(t, "sub f "+typed+";\n")); got != want {
			t.Errorf("%s: got %q, want %q", typed, got, want)
		}
	}
}

// TestDeclaredTrailingSemicolonIsKept: types cannot say a trailing `;`.
// Measured on 5.42, `f 1, 2` under `($;)` is "Too many arguments" -- a list
// operator -- while under `($)` it reads `f(1), 2`; and `g 1` under `(;)` is
// "Too many arguments" where `()` is a syntax error. So `$;` and `;` do not
// round-trip through types, a declared `:prototype($;)` agrees with
// `(Scalar $x)`, and the declared prototype is what the declaration keeps.
func TestDeclaredTrailingSemicolonIsKept(t *testing.T) {
	for src, want := range map[string]string{
		"sub f :prototype($;) (Scalar $x);\n": "($;)",
		"sub g :prototype(;) ();\n":           "(;)",
	} {
		facts := readDeclaration([]byte(src), nil)
		name := src[4:5]
		if len(facts.errs) > 0 || facts.protos[name] != want {
			t.Errorf("%q: errors %v, prototype %q, want %s", src, facts.errs, facts.protos[name], want)
		}
	}
	for proto, back := range map[string]string{"$;": "$", ";": ""} {
		sig, _, _ := typesFromPrototype(proto)
		if got := prototypeFromTypes(sig); got != back {
			t.Errorf("(%s) through types gave %q, want %q: a trailing ; is not a type", proto, got, back)
		}
	}
}
