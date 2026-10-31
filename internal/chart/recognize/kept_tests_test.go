// ABOUTME: Holds the package's tests to toy grammars: no _test.go file may call a Guacamole loader.
// ABOUTME: The check parses the files with go/parser; a second test feeds it synthetic sources so it can fail.

package recognize

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"testing"
)

// guacamoleCallers names each function in src that calls Guacamole, bare or
// qualified (recognize.Guacamole, lex.Guacamole).
func guacamoleCallers(file string, src any) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), file, src, 0)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ast.Inspect(fn, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name := ""
			switch fun := call.Fun.(type) {
			case *ast.Ident:
				name = fun.Name
			case *ast.SelectorExpr:
				name = fun.Sel.Name
			}
			if name == "Guacamole" {
				out = append(out, file+": "+fn.Name.Name)
			}
			return true
		})
	}
	return out, nil
}

// TestKeptTestsUseNoGuacamole: the recognizer, Leo, lattice and fuzz tests
// run on toy grammars, so none of the package's _test.go files calls a
// Guacamole loader.
func TestKeptTestsUseNoGuacamole(t *testing.T) {
	files, err := filepath.Glob("*_test.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("no test files: %v", err)
	}
	for _, file := range files {
		callers, err := guacamoleCallers(file, nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range callers {
			t.Errorf("%s calls Guacamole", c)
		}
	}
}

// TestKeptTestsUseNoGuacamoleDetectsCall: the check names a function that
// calls Guacamole, bare or qualified, and passes one that does not.
func TestKeptTestsUseNoGuacamoleDetectsCall(t *testing.T) {
	for src, want := range map[string][]string{
		"package p\nfunc direct() { Guacamole(\"x\") }":                        {"s.go: direct"},
		"package p\nfunc qualified() { _, _, _ = recognize.Guacamole(\"x\") }": {"s.go: qualified"},
		"package p\nfunc clean() { toyTables(nil) }\nvar _ = \"Guacamole()\"":  nil,
	} {
		got, err := guacamoleCallers("s.go", src)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got, want) {
			t.Errorf("%q: callers %v, want %v", src, got, want)
		}
	}
	if _, err := guacamoleCallers("s.go", "package"); err == nil {
		t.Error("an unparsable file was not an error")
	}
}
