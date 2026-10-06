// ABOUTME: Tests that a library's .pmt is read in CORE.pmt's language, RFC 0001 "One language for every .pmt".
// ABOUTME: The fixture testdata/declarations/My/Lib.pmt is read through the resolver, as an embedded declaration is.
package parse

import (
	"os"
	"reflect"
	"slices"
	"testing"

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
