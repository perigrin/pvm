// ABOUTME: Tests LibraryOperators, the export of each infix operator a file's `use` statements brought into scope.
// ABOUTME: The fixture testdata/declarations/My/Lib.tpm declares `plus :infix(ADD)`; From is the byte its `use` starts at.
package parse

import (
	"slices"
	"testing"
	"testing/fstest"
)

// libraryOpsOf parses src with the My::Lib fixture in scope.
func libraryOpsOf(t *testing.T, src string) []LibraryOperator {
	t.Helper()
	useLibraryFixture(t)
	return LibraryOperators(ParseWithLoader([]byte(src), noModules))
}

// TestLibraryOperatorsExport: each infix operator a `use` brought into scope
// is reported with its class, the level recursive descent placed it at, and
// where that `use` starts. `plus :infix(ADD)` sits at `+`'s level. `use M ()`
// brings none, and Parse, which resolves no `use`, reports none.
func TestLibraryOperatorsExport(t *testing.T) {
	derivePowers()
	plus := func(from int) LibraryOperator {
		return LibraryOperator{Name: "plus", Class: "ADD", Info: infix["+"], From: from}
	}
	for src, want := range map[string][]LibraryOperator{
		"use My::Lib; 1 plus 2;":          {plus(0)},
		"1;\nuse My::Lib;\n1 plus 2;":     {plus(3)},
		"use My::Lib;\nuse My::Lib;\n":    {plus(0), plus(13)},
		"use My::Lib ();\n1;":             nil,
		"use My::Lib qw(join_all);\n1;":   {plus(0)},
		"no My::Lib;\nrequire My::Lib;\n": nil,
	} {
		if got := libraryOpsOf(t, src); !slices.Equal(got, want) {
			t.Errorf("%q: got %+v, want %+v", src, got, want)
		}
	}
	if got := LibraryOperators(Parse([]byte("use My::Lib; 1 plus 2;"))); len(got) != 0 {
		t.Errorf("Parse: got %+v, want none", got)
	}
}

// TestLibraryOperatorsExportBringsNoneWithoutInfixUse: a prefix or postfix
// declaration, a require, and a use of a module that cannot be loaded bring
// no infix operator, and none panics.
func TestLibraryOperatorsExportBringsNoneWithoutInfixUse(t *testing.T) {
	useLibraryFixture(t)
	saved := declarations
	t.Cleanup(func() { declarations = saved })
	decl := "package Unary;\nsub ! :prefix (Scalar $x) Boolean;\nsub ++ :postfix (Scalar $x) Scalar;\n"
	declarations = layeredFS{top: fstest.MapFS{"declarations/Unary.tpm": {Data: []byte(decl)}}, base: saved}
	for _, src := range []string{
		"use Unary;\n1;", "require My::Lib;\n1;", "use No::Such::Module;\n1;", "BEGIN { require My::Lib; }\n1;",
	} {
		if got := LibraryOperators(ParseWithLoader([]byte(src), noModules)); len(got) != 0 {
			t.Errorf("%q: got %+v, want none", src, got)
		}
	}
	if got := LibraryOperators(nil); len(got) != 0 {
		t.Errorf("nil root: got %+v", got)
	}
}

// TestLibraryOperatorsExportStable: a `use` the parser reads twice while it
// backtracks (a `try` block with no `catch`) is reported once, and the slice
// returned is a copy.
func TestLibraryOperatorsExportStable(t *testing.T) {
	useLibraryFixture(t)
	root := ParseWithLoader([]byte("try { use My::Lib; };\n1 plus 2;"), noModules)
	got := LibraryOperators(root)
	if len(got) != 1 || got[0].Name != "plus" || got[0].From != 6 {
		t.Fatalf("got %+v, want plus once, from 6", got)
	}
	got[0].Name = "changed"
	if again := LibraryOperators(root); again[0].Name != "plus" {
		t.Errorf("the returned slice aliases the tree's: %+v", again)
	}
}

// TestLibraryOperatorsExportFromIsUseStart: From is where the `use` statement
// starts, not the module name or the operator, also after a comment and under
// a `#line` directive.
func TestLibraryOperatorsExportFromIsUseStart(t *testing.T) {
	for src, from := range map[string]int{
		"# a comment\nuse My::Lib;\n1 plus 2;\nplus 1;": 12,
		"#line 100\nuse My::Lib;\n1 plus 2;":            10,
		"  use   My::Lib  ;1 plus 2;":                   2,
	} {
		got := libraryOpsOf(t, src)
		if len(got) != 1 || got[0].From != from {
			t.Errorf("%q: got %+v, want From %d", src, got, from)
		}
	}
}
