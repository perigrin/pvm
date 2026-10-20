// ABOUTME: Scans the perl-held test sources for a t.Skip gated on runtime.GOOS, the shape that
// ABOUTME: drops a whole test (and every name it would exercise) off a platform instead of just the names that differ.

package perltest

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// perlHeldTestSources are the files 01a15b21 narrowed from whole-test
// platform skips to per-OS exemption maps: every perl-held test that
// iterates several module names lives in one of these two.
var perlHeldTestSources = []string{
	filepath.Join("..", "infer", "returntypes_test.go"),
	filepath.Join("..", "parse", "heldtoperl_test.go"),
}

// TestNoWholeTestPlatformSkips: no t.Skip or t.Skipf in either file sits
// under an if or switch whose condition reads runtime.GOOS -- the shape
// that skips an entire test, and every name it would have checked, on
// one platform rather than exempting just the names that differ there.
func TestNoWholeTestPlatformSkips(t *testing.T) {
	for _, path := range perlHeldTestSources {
		for _, loc := range goosGatedSkips(t, path) {
			t.Errorf("%s: %s", path, loc)
		}
	}
}

// goosGatedSkips parses path and reports, one line each, every t.Skip or
// t.Skipf call textually inside an if block whose condition mentions
// runtime.GOOS, or a switch case list that does the same.
func goosGatedSkips(t *testing.T, path string) []string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	require.NoError(t, err)

	var found []string
	reportSkipsIn := func(n ast.Node) {
		ast.Inspect(n, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if ok && (sel.Sel.Name == "Skip" || sel.Sel.Name == "Skipf") {
				found = append(found, fmt.Sprintf("line %d: %s gated by a runtime.GOOS check", fset.Position(call.Pos()).Line, sel.Sel.Name))
			}
			return true
		})
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.IfStmt:
			if mentionsRuntimeGOOS(s.Cond) {
				reportSkipsIn(s.Body)
			}
		case *ast.SwitchStmt:
			if s.Tag != nil && mentionsRuntimeGOOS(s.Tag) {
				reportSkipsIn(s.Body)
			}
			for _, stmt := range s.Body.List {
				cc, ok := stmt.(*ast.CaseClause)
				if !ok {
					continue
				}
				gated := false
				for _, e := range cc.List {
					if mentionsRuntimeGOOS(e) {
						gated = true
					}
				}
				if gated {
					for _, body := range cc.Body {
						reportSkipsIn(body)
					}
				}
			}
		}
		return true
	})
	return found
}

// mentionsRuntimeGOOS reports whether expr reads runtime.GOOS anywhere
// in it.
func mentionsRuntimeGOOS(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if ok {
			if id, ok := sel.X.(*ast.Ident); ok && id.Name == "runtime" && sel.Sel.Name == "GOOS" {
				found = true
			}
		}
		return true
	})
	return found
}
