// ABOUTME: Tests for deriving a .pmt declaration's prototype from its types and its types from its prototype.
// ABOUTME: Expected values are RFC 0001's table under "A typed signature and a prototype say the same thing".
package parse

import (
	"fmt"
	"reflect"
	"strings"
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
// and nothing wider or narrower, but for Void: List is that union and
// Void, and the empty list in a `+` slot arrives as undef (measured on
// 5.42, `f(())` passes one undef). Measured on 5.42, `f(@a)` passes an
// ARRAY reference under `+` and the count under `$`, so a `$` parameter
// typed Any or Array|Str is a `$`, not a `+`.
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

// TestAliasedParamsDerivePrototypes: RFC 0001's table rows `\$`, `\@` and
// `\%`, read both ways. An aliased parameter derives its backslashed
// character, and a backslashed prototype character derives the aliased
// parameter: an Array or Hash passed whole, or a Scalar lvalue, each one
// required argument. Measured on 5.42.0, `sub f (\@) {} f()` is "Not
// enough arguments", and `f(@x)` passes an ARRAY reference.
func TestAliasedParamsDerivePrototypes(t *testing.T) {
	rows := map[string]types.Param{
		`\$`: {Sigil: '$', Type: types.Scalar, Alias: true, Required: true},
		`\@`: {Sigil: '@', Type: types.Array, Alias: true, Required: true},
		`\%`: {Sigil: '%', Type: types.Hash, Alias: true, Required: true},
	}
	for typed, want := range map[string]string{
		`(Scalar \$x)`: `\$`,
		`(Array \@a)`:  `\@`,
		`(Hash \%h)`:   `\%`,
	} {
		facts := readDeclaration([]byte("sub f "+typed+";\n"), nil)
		if len(facts.errs) > 0 || facts.protos["f"] != "("+want+")" {
			t.Errorf("%s: errors %v, prototype %q, want (%s)", typed, facts.errs, facts.protos["f"], want)
		}
	}
	for proto, want := range rows {
		facts := readDeclaration([]byte("sub f :prototype("+proto+");\n"), nil)
		if got := facts.signatures["f"]; len(facts.errs) > 0 || !reflect.DeepEqual(got, []types.Signature{{Params: []types.Param{want}}}) {
			t.Errorf("(%s): errors %v, got %+v, want %+v", proto, facts.errs, got, want)
		}
		sig, derived, err := typesFromPrototype(proto)
		if back := prototypeFromTypes(sig); err != nil || !derived || back != proto {
			t.Errorf("(%s): derived %v, error %v, round trip gave %q", proto, derived, err, back)
		}
	}
}

// TestPushRoundTripsRefArrayAt: RFC 0001's `sub push (Array \@a, List
// @list) Int;`. In a `.pmt` a sigil is the caller's view, so only the
// final parameter is slurpy and the array before it is one argument: the
// declaration derives perl's `\@@` (measured on 5.42.0,
// `prototype("CORE::push")`), and `:prototype(\@@)` derives the two
// parameters back.
func TestPushRoundTripsRefArrayAt(t *testing.T) {
	sig := typedSignature(t, "sub f (Array \\@a, List @list) Int;\n")
	if got := prototypeFromTypes(sig); got != `\@@` {
		t.Errorf("prototype: got %q, want \\@@", got)
	}
	if !sig.Params[0].Required || sig.Params[0].Slurpy() || !sig.Params[1].Slurpy() {
		t.Errorf("only the final parameter should be slurpy: %+v", sig.Params)
	}
	facts := readDeclaration([]byte("sub f :prototype(\\@@);\n"), nil)
	want := []types.Signature{{Params: []types.Param{
		{Sigil: '@', Type: types.Array, Alias: true, Required: true},
		protoList,
	}}}
	if got := facts.signatures["f"]; len(facts.errs) > 0 || !reflect.DeepEqual(got, want) {
		t.Errorf("(\\@@): errors %v, got %+v, want %+v", facts.errs, got, want)
	}
	if back := prototypeFromTypes(want[0]); back != `\@@` {
		t.Errorf("(\\@@): round trip gave %q", back)
	}
}

// TestCodeAliasDerivesRefAmp: RFC 0001 "The scalar container", a code
// slot aliased as perlref's `\&foo = \&bar` aliases a sub, `Code \&c`,
// derives `\&` where an unbackslashed `Code &c` derives `&`. So `sub any
// (Code \&block, List @list) Boolean;` derives perl's `\&@` (measured on
// 5.42.0, `prototype("CORE::any")` and `prototype("CORE::all")`), and
// `:prototype(\&@)` derives the two parameters back.
func TestCodeAliasDerivesRefAmp(t *testing.T) {
	code := types.Param{Sigil: '&', Type: types.Code, Alias: true, Required: true}
	sig := typedSignature(t, "sub f (Code \\&c);\n")
	if named := (types.Param{Name: "c", Sigil: '&', Type: types.Code, Alias: true, Required: true}); !reflect.DeepEqual(sig.Params, []types.Param{named}) {
		t.Errorf("Code \\&c: got %+v, want %+v", sig.Params, named)
	}
	if got := prototypeFromTypes(sig); got != `\&` {
		t.Errorf("Code \\&c: prototype %q, want \\&", got)
	}
	if got := prototypeFromTypes(typedSignature(t, "sub f (Code \\&block, List @list) Boolean;\n")); got != `\&@` {
		t.Errorf("any: prototype %q, want \\&@", got)
	}
	for proto, want := range map[string][]types.Param{
		`\&`:  {code},
		`\&@`: {code, protoList},
	} {
		facts := readDeclaration([]byte("sub f :prototype("+proto+");\n"), nil)
		if got := facts.signatures["f"]; len(facts.errs) > 0 || !reflect.DeepEqual(got, []types.Signature{{Params: want}}) {
			t.Errorf("(%s): errors %v, got %+v, want %+v", proto, facts.errs, got, want)
		}
		if back := prototypeFromTypes(types.Signature{Params: want}); back != proto {
			t.Errorf("(%s): round trip gave %q", proto, back)
		}
	}
}

// TestAliasedPrototypeDisagreementIsError: `\@` is an Array passed whole
// and `@` the rest of the call, flattened (RFC 0001, "A typed signature and
// a prototype say the same thing"), so a declaration whose prototype says
// one and whose types say the other is in error.
func TestAliasedPrototypeDisagreementIsError(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f :prototype(\\@) (List @l);\n":  `sub f: :prototype(\@) disagrees with its types, which give (@)`,
		"sub f :prototype(@) (Array \\@a);\n": `sub f: :prototype(@) disagrees with its types, which give (\@)`,
	})
}

// TestDerivePrototypeFromMultiCandidates: RFC 0001's `\[$@%]` row, "any one
// of those containers". A multi still has one prototype, so `each`'s two
// candidates, which differ only in the container their one aliased
// parameter takes, derive `\[%@]`, the containers in declaration order.
// Measured on 5.42, `prototype("CORE::each")` is `\[%@]`.
func TestDerivePrototypeFromMultiCandidates(t *testing.T) {
	src := "multi sub f (Hash \\%h) List;\nmulti sub f (Array \\@a) List;\n"
	facts := readDeclaration([]byte(src), nil)
	if len(facts.errs) > 0 || facts.protos["f"] != `(\[%@])` {
		t.Errorf("each: errors %v, prototype %q, want (\\[%%@])", facts.errs, facts.protos["f"])
	}
}

// TestDeriveMultiCandidatesGroupByParameters: candidates group by their
// parameters only, so `:context` and the return type never split a group.
// `keys`' four candidates are two groups, Hash and Array, and derive
// `\[%@]`; `localtime`'s two share `(Int $time = time)` and derive `;$`.
// Measured on 5.42, `prototype("CORE::keys")` is `\[%@]` and
// `prototype("CORE::localtime")` is `;$`.
func TestDeriveMultiCandidatesGroupByParameters(t *testing.T) {
	for src, want := range map[string]string{
		"multi sub f :context($) (Hash \\%h) Int;\n" +
			"multi sub f :context($) (Array \\@a) Int;\n" +
			"multi sub f :context(@) (Hash \\%h) List;\n" +
			"multi sub f :context(@) (Array \\@a) List;\n": `(\[%@])`,
		"multi sub f :context($) (Int $time = time) Str;\n" +
			"multi sub f :context(@) (Int $time = time) List[Int];\n": "(;$)",
	} {
		facts := readDeclaration([]byte(src), nil)
		if len(facts.errs) > 0 || facts.protos["f"] != want {
			t.Errorf("%q: errors %v, prototype %q, want %s", src, facts.errs, facts.protos["f"], want)
		}
	}
}

// TestDeriveMultiCandidatesRefusesInexpressible: candidates that differ in
// two aliased positions, or in arity, or in a position that is not
// aliased, derive no single prototype; nor does a candidate with an
// invocant colon. Where the declaration requires a prototype by stating
// one, that is an error naming the sub, never a guess. A stated `\[...]`
// must also list its containers in the candidates' declaration order:
// perl keeps a prototype as written (measured on 5.42, `sub f (\[@%])`
// reports `\[@%]` and `sub g (\[%@])` reports `\[%@]`). Without a stated
// prototype the candidates have none, as perl reports none for `grep`
// and `select`.
func TestDeriveMultiCandidatesRefusesInexpressible(t *testing.T) {
	twoPositions := "multi sub f :prototype(%s) (Hash \\%%h, Hash \\%%g) List;\nmulti sub f :prototype(%s) (Array \\@a, Array \\@b) List;\n"
	arity := "multi sub f :prototype(%s) (Hash \\%%h) List;\nmulti sub f :prototype(%s) (Array \\@a, Scalar $x) List;\n"
	unaliased := "multi sub f :prototype(%s) (Code &block, List @list) List;\nmulti sub f :prototype(%s) (Scalar $expr, List @list) List;\n"
	invocant := "multi sub f :prototype(%s) :context($) (FileHandle $fh: List @l) Int;\nmulti sub f :prototype(%s) :context(@) (FileHandle $fh: List @l) List;\n"
	declared := func(src, proto string) string { return fmt.Sprintf(src, proto, proto) }
	pmtRefuses(t, map[string]string{
		declared(twoPositions, `\[%@]\[%@]`): `sub f: :prototype(\[%@]\[%@]) disagrees with its types: its candidates give (\%\%), (\@\@), which no one prototype states`,
		declared(arity, `\[%@]`):             `sub f: :prototype(\[%@]) disagrees with its types: its candidates give (\%), (\@$), which no one prototype states`,
		declared(unaliased, `&@`):            `sub f: :prototype(&@) disagrees with its types: its candidates give (&@), ($@), which no one prototype states`,
		declared(invocant, `@`):              `sub f: :prototype(@) disagrees with its types: an invocant colon derives no prototype`,
		"multi sub f :prototype(\\[@%]) (Hash \\%h) List;\nmulti sub f :prototype(\\[@%]) (Array \\@a) List;\n": `sub f: :prototype(\[@%]) disagrees with its types, which give (\[%@])`,
	})
	for _, src := range []string{twoPositions, arity, unaliased, invocant} {
		src = strings.ReplaceAll(fmt.Sprintf(src, "", ""), " :prototype()", "")
		facts := readDeclaration([]byte(src), nil)
		if proto, ok := facts.protos["f"]; len(facts.errs) > 0 || proto != "" || len(facts.signatures["f"]) != 2 {
			t.Errorf("%q: errors %v, prototype %q (recorded %v), signatures %+v; want two candidates and no prototype", src, facts.errs, proto, ok, facts.signatures["f"])
		}
	}
}

// TestDeriveMultiCandidatesIgnoreReturnType: candidates with the same
// parameters and different return types, and no `:context`, are one
// group: they derive one prototype, and do not count as a position
// differing. The pair is read before the ambiguity check, which refuses
// such a declaration for selection, not for its prototype.
func TestDeriveMultiCandidatesIgnoreReturnType(t *testing.T) {
	for src, want := range map[string]string{
		"multi sub f (Hash \\%h) Int;\nmulti sub f (Hash \\%h) List;\n":                                 `\%`,
		"multi sub f (Hash \\%h) Int;\nmulti sub f (Hash \\%h) List;\nmulti sub f (Array \\@a) List;\n": `\[%@]`,
	} {
		_, p := parseSource([]byte(src), nil, true)
		if len(p.typedErrs) > 0 {
			t.Fatalf("%q: errors %v", src, p.typedErrs)
		}
		if got, err := prototypeFromCandidates(p.signatures["f"]); err != nil || got != want {
			t.Errorf("%q: got %q, error %v, want %s", src, got, err, want)
		}
	}
}

// TestDeriveSingleMultiCandidateIsPlain: a multi with one candidate derives
// what the plain derivation gives it, `\%` and not `\[%]`.
func TestDeriveSingleMultiCandidateIsPlain(t *testing.T) {
	facts := readDeclaration([]byte("multi sub f (Hash \\%h) List;\n"), nil)
	if len(facts.errs) > 0 || facts.protos["f"] != `(\%)` {
		t.Errorf("errors %v, prototype %q, want (\\%%)", facts.errs, facts.protos["f"])
	}
	_, p := parseSource([]byte("multi sub f (Hash \\%h) List;\n"), nil, true)
	if got, err := prototypeFromCandidates(p.signatures["f"]); err != nil || got != `\%` {
		t.Errorf("one candidate: got %q, error %v, want \\%%", got, err)
	}
}

// TestStatedUnionPrototypeAgrees: a stated `\[...]` is compared with the
// prototype the candidates give, for one candidate and for several, so
// one they do not give is an error, never accepted as written. A trailing
// `;` is not a type, so `:prototype(\[%@];)` agrees with each's
// candidates as `:prototype($;)` agrees with `(Scalar $x)`.
func TestStatedUnionPrototypeAgrees(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"multi sub f :prototype(\\[@$]) :context($) (Hash \\%h) Int;\nmulti sub f :prototype(\\[@$]) :context(@) (Hash \\%h) List;\n": `sub f: :prototype(\[@$]) disagrees with its types, which give (\%)`,
		"multi sub f :prototype(\\[@$]) (Hash \\%h) List;\n":                                                                          `sub f: :prototype(\[@$]) disagrees with its types, which give (\%)`,
		"sub f :prototype(\\[@$]) (Hash \\%h) List;\n":                                                                                `sub f: :prototype(\[@$]) disagrees with its types, which give (\%)`,
	})
	facts := readDeclaration([]byte("multi sub f :prototype(\\[%@];) (Hash \\%h) List;\nmulti sub f :prototype(\\[%@];) (Array \\@a) List;\n"), nil)
	if len(facts.errs) > 0 || facts.protos["f"] != `(\[%@];)` {
		t.Errorf(`\[%%@];: errors %v, prototype %q, want (\[%%@];)`, facts.errs, facts.protos["f"])
	}
}
