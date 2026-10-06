// ABOUTME: Tests that a library's .pmt is read in CORE.pmt's language, RFC 0001 "One language for every .pmt".
// ABOUTME: The fixture testdata/declarations/My/Lib.pmt is read through the resolver, as an embedded declaration is.
package parse

import (
	"os"
	"reflect"
	"slices"
	"testing"
	"testing/fstest"

	"tamarou.com/pvm/internal/types"
)

// useLibraryFixture makes the declarations under testdata the ones the
// resolver reads, for the rest of the test.
func useLibraryFixture(t *testing.T) {
	t.Helper()
	saved := declarations
	declarations = os.DirFS("testdata")
	t.Cleanup(func() { declarations = saved })
}

// resolveLibrary reads module's declaration through the resolver, the path
// a `use` takes.
func resolveLibrary(t *testing.T, module string) moduleFacts {
	t.Helper()
	r := &resolver{load: func(string) ([]byte, bool) { return nil, false }, seen: map[string]bool{}}
	facts, ok := r.resolve(module)
	if !ok {
		t.Fatalf("%s not resolved", module)
	}
	return facts
}

// TestLibraryPmtSpeaksCoreLanguage: RFC 0001 "One language for every
// `.pmt`". The fixture states every construct CORE.pmt states, and reading
// it through the resolver yields each one, as reading it the way CORE.pmt
// is read does.
func TestLibraryPmtSpeaksCoreLanguage(t *testing.T) {
	useLibraryFixture(t)
	facts := resolveLibrary(t, "My::Lib")
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	for name, n := range map[string]int{
		"join_all": 1, "alias_all": 1, "need": 1, "pick": 2, "when_": 2, "emit": 1, "size": 1, "hidden": 1,
	} {
		if got := len(facts.signatures[name]); got != n {
			t.Errorf("%s: %d signatures, want %d", name, got, n)
		}
	}
	if s := facts.signatures["emit"]; len(s) == 0 || s[0].Invocant == nil {
		t.Errorf("emit: no invocant")
	}
	if s := facts.signatures["size"]; len(s) == 0 || !s[0].Unary {
		t.Errorf("size: not :unary")
	}
	if len(facts.operators) != 1 || facts.operators[0].name != "plus" {
		t.Errorf("operators: %+v", facts.operators)
	}
	if _, ok := facts.syntax["await"]; !ok {
		t.Errorf("await: no declared syntax")
	}
	if want := []string{"join_all", "alias_all", "need", "pick", "when_", "emit", "size"}; !slices.Equal(facts.defaults, want) {
		t.Errorf("@EXPORT: got %v, want %v", facts.defaults, want)
	}

	// The same file read as CORE.pmt is read says the same.
	src, err := os.ReadFile("testdata/declarations/My/Lib.pmt")
	if err != nil {
		t.Fatal(err)
	}
	protos, sigs, err := coreProtos(src)
	if err != nil {
		t.Fatalf("read as CORE.pmt: %v", err)
	}
	if !reflect.DeepEqual(sigs, facts.signatures) {
		t.Errorf("signatures differ from CORE.pmt's reading:\n got %+v\nwant %+v", facts.signatures, sigs)
	}
	for name, proto := range protos {
		if got := facts.protos[name]; got != "("+proto+")" {
			t.Errorf("%s: prototype %q, CORE.pmt's reading gives (%s)", name, got, proto)
		}
	}
}

// TestLibraryPmtOperatorsAndUnary: RFC 0001 "One language for every
// `.pmt`" and "Operator declarations": a library declares an operator in
// CORE.pmt's spelling, `:infix(CLASS)` with XS::Parse::Infix's class, and
// it is read with its fixity, class and signature; its `:unary` sub is
// read with its signature, and derives no prototype, as CORE.pmt's do.
func TestLibraryPmtOperatorsAndUnary(t *testing.T) {
	useLibraryFixture(t)
	facts := resolveLibrary(t, "My::Lib")
	num := func(n string) types.Param { return types.Param{Name: n, Sigil: '$', Type: types.Num, Required: true} }
	want := []operatorDecl{{name: "plus", fixity: "infix", class: "ADD",
		sig: types.Signature{Params: []types.Param{num("x"), num("y")}, Returns: types.Num}}}
	if !reflect.DeepEqual(facts.operators, want) {
		t.Errorf("operators:\n got %+v\nwant %+v", facts.operators, want)
	}
	if _, ok := facts.signatures["plus"]; ok {
		t.Errorf("plus recorded as a sub")
	}
	size := []types.Signature{{Params: []types.Param{{Name: "x", Sigil: '$', Type: types.Scalar, Default: "$_"}}, Returns: types.Int, Unary: true}}
	if got := facts.signatures["size"]; !reflect.DeepEqual(got, size) {
		t.Errorf("size:\n got %+v\nwant %+v", got, size)
	}
	if proto := facts.protos["size"]; proto != "" {
		t.Errorf("size: derived prototype %q", proto)
	}
}

// TestLibraryPmtCoinedClassRefused: RFC 0001 "Operator declarations".
// BITAND, BITOR, SHIFT and RANGE are coined for CORE.pmt: XS::Parse::Infix
// classes no operator at those levels, so it cannot register one there,
// and a library .pmt naming one is in error, naming its module. CORE.pmt
// names them.
func TestLibraryPmtCoinedClassRefused(t *testing.T) {
	saved := declarations
	t.Cleanup(func() { declarations = saved })
	for _, class := range []string{"BITAND", "BITOR", "SHIFT", "RANGE"} {
		src := "package Coined;\nsub op :infix(" + class + ") (Int $x, Int $y) Int;\n"
		declarations = fstest.MapFS{"declarations/Coined.pmt": {Data: []byte(src)}}
		root := ParseWithLoader([]byte("use Coined;\n"), func(string) ([]byte, bool) { return nil, false })
		want := "Coined: sub op: operator class " + class + " is CORE.pmt's; XS::Parse::Infix registers no operator at its level"
		if errs := DeclarationErrors(root); len(errs) != 1 || errs[0].Error() != want {
			t.Errorf("%s: got %v, want %q", class, errs, want)
		}
		if facts := resolveLibrary(t, "Coined"); len(facts.operators) > 0 {
			t.Errorf("%s: recorded %+v", class, facts.operators)
		}
	}
	if _, _, err := coreProtos([]byte("sub & :infix(BITAND) (Int $x, Int $y) Int;\n")); err != nil {
		t.Errorf("CORE.pmt's reading refused BITAND: %v", err)
	}
	// Reached by a `use`, CORE.pmt is still the interpreter's file.
	declarations = saved
	if errs := DeclarationErrors(ParseWithLoader([]byte("use CORE;\n"), func(string) ([]byte, bool) { return nil, false })); len(errs) > 0 {
		t.Errorf("use CORE: %v", errs)
	}
}

// TestPmtUnaryTakesOneOperand: `:unary` marks a named unary (RFC 0001,
// "Builtins with no prototype"), which takes at most one operand, so a
// signature stating two is an error and records nothing; an invocant is
// an operand too. One operand, or a list form as chomp's, is a unary's.
func TestPmtUnaryTakesOneOperand(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f :unary (Scalar $x, Scalar $y) Int;\n":   "sub f: :unary takes at most one operand, the signature has 2",
		"sub f :unary (FileHandle $fh: Str $x) Int;\n": "sub f: :unary takes at most one operand, the signature has 2",
	})
	for _, src := range []string{
		"sub f :unary (Scalar $x = $_) Int;\n",
		"sub f :unary (List[Str] \\(@args = ($_))) Int;\n",
	} {
		typedSignature(t, src)
	}
}

// TestPmtAmbiguousOperatorCandidatesIsError: RFC 0001 "Multi
// declarations" holds for an operator's candidates as for a sub's (RFC
// 0001, "Operators that fork" makes them multis): candidates with no
// single most specific one are ambiguous, the declaration is an error
// naming the operator, and none of its candidates is recorded.
func TestPmtAmbiguousOperatorCandidatesIsError(t *testing.T) {
	src := "multi sub plus :infix(ADD) (Int $a, Num $b) Num;\nmulti sub plus :infix(ADD) (Num $a, Int $b) Num;\n"
	facts := readDeclaration([]byte(src), nil)
	want := "sub plus: candidates (Int $a, Num $b) and (Num $a, Int $b) are ambiguous for (Int, Int)"
	if len(facts.errs) != 1 || facts.errs[0].Error() != want {
		t.Errorf("got errors %v, want %q", facts.errs, want)
	}
	if len(facts.operators) > 0 {
		t.Errorf("recorded %+v", facts.operators)
	}
	// A prefix and an infix operator of one symbol are two operators.
	both := readDeclaration([]byte("sub - :prefix (Num $x) Num;\nsub - :infix(ADD) (Num $x, Num $y) Num;\n"), nil)
	if len(both.errs) > 0 || len(both.operators) != 2 {
		t.Errorf("prefix and infix -: errors %v, operators %+v", both.errs, both.operators)
	}
}

// TestLibraryPmtAliasedAndContainerParams: RFC 0001 "The scalar
// container" and "A typed signature and a prototype say the same thing",
// in a library's .pmt. Its aliased parameters derive `\$`, `\@` and `\&`
// and its `\(@args)` list form is an aliased list; its `= die` parameter
// is required and derives no `;` before it; and its `List[Str] @parts`
// keeps its container type.
func TestLibraryPmtAliasedAndContainerParams(t *testing.T) {
	useLibraryFixture(t)
	facts := resolveLibrary(t, "My::Lib")
	for name, want := range map[string]string{
		"alias_all": `(\$\@\&@)`,
		"need":      "($)",
		"join_all":  "($@)",
	} {
		if got := facts.protos[name]; got != want {
			t.Errorf("%s: prototype %q, want %q", name, got, want)
		}
	}
	alias := []types.Signature{{Params: []types.Param{
		{Name: "x", Sigil: '$', Type: types.Scalar, Alias: true, Required: true},
		{Name: "a", Sigil: '@', Type: types.Array, Alias: true, Required: true},
		{Name: "c", Sigil: '&', Type: types.Code, Alias: true, Required: true},
		{Name: "args", Sigil: '@', Type: types.List, Element: types.Str, AliasEach: true},
	}, Returns: types.Int}}
	if got := facts.signatures["alias_all"]; !reflect.DeepEqual(got, alias) {
		t.Errorf("alias_all:\n got %+v\nwant %+v", got, alias)
	}
	need := []types.Signature{{Params: []types.Param{{Name: "name", Sigil: '$', Type: types.Str, Required: true}}, Returns: types.Str}}
	if got := facts.signatures["need"]; !reflect.DeepEqual(got, need) {
		t.Errorf("need:\n got %+v\nwant %+v", got, need)
	}
	join := []types.Signature{{Params: []types.Param{
		{Name: "sep", Sigil: '$', Type: types.Str, Required: true},
		{Name: "parts", Sigil: '@', Type: types.List, Element: types.Str},
	}, Returns: types.Str}}
	if got := facts.signatures["join_all"]; !reflect.DeepEqual(got, join) {
		t.Errorf("join_all:\n got %+v\nwant %+v", got, join)
	}
}
