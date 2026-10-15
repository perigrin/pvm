// ABOUTME: Asserts the tree-sitter path is gone from internal/infer: no grammar name in a case arm,
// ABOUTME: and no inference rule reading source text to learn which operator or operation it has.

package infer_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// treeSitterGrammarNames is the tree-sitter-perl vocabulary internal/infer
// once switched on: the 46 distinct string literals in the case arms of a
// switch on x.Kind(), or on a variable assigned from one, across infer.go,
// symbols.go, project.go and diagnostics.go, measured at 1cf680fa by the
// method chapter 4 §4.14 states.
var treeSitterGrammarNames = []string{
	"ambiguous_function_call_expression", "anonymous_array_expression",
	"anonymous_hash_expression", "anonymous_subroutine_expression", "array",
	"array_deref_expression", "array_element_expression", "arraylen",
	"assignment_expression", "bareword", "binary_expression", "block",
	"conditional_expression", "conditional_statement", "else", "elsif",
	"equality_expression", "expression_statement", "for_statement",
	"func0op_call_expression", "func1op_call_expression",
	"function_call_expression", "hash", "hash_deref_expression",
	"hash_element_expression", "interpolated_string_literal",
	"list_expression", "loop_statement", "lowprec_logical_expression",
	"map_grep_expression", "method_call_expression", "number",
	"package_statement", "quoted_regexp", "refgen_expression",
	"relational_expression", "return_expression", "scalar", "signature",
	"sort_expression", "string_literal", "subroutine_declaration_statement",
	"unary_expression", "undef_expression", "use_statement",
	"variable_declaration",
}

// inferSources parses the package's non-test .go files.
func inferSources(t *testing.T) map[string]*ast.File {
	t.Helper()
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	files := map[string]*ast.File{}
	for _, p := range paths {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, p, nil, 0)
		if err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		files[p] = f
	}
	if len(files) == 0 {
		t.Fatal("no non-test .go files found")
	}
	return files
}

func TestNoTreeSitterVocabularyRemains(t *testing.T) {
	forbidden := map[string]bool{}
	for _, n := range treeSitterGrammarNames {
		forbidden[n] = true
	}
	for path, f := range inferSources(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			cc, ok := n.(*ast.CaseClause)
			if !ok {
				return true
			}
			for _, e := range cc.List {
				lit, ok := e.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				if s, _ := strconv.Unquote(lit.Value); forbidden[s] {
					t.Errorf("%s: case arm names the tree-sitter grammar name %q", path, s)
				}
			}
			return true
		})
	}
}

// sourceFormatters are the functions that may take source bytes: they turn
// a diagnostic's byte span into a line and column for display, and decide
// nothing about what the program does.
var sourceFormatters = map[string]bool{
	"FormatDiagnostic":    true,
	"byteOffsetToLineCol": true,
}

func TestNoSourceTextRederivation(t *testing.T) {
	for path, f := range inferSources(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.FuncDecl:
				if n.Name.Name == "findOperatorText" {
					t.Errorf("%s: findOperatorText re-derives an operator from source text", path)
				}
				if !sourceFormatters[n.Name.Name] && takesBytes(n.Type) {
					t.Errorf("%s: %s takes source bytes; the lowered tree carries what it reads", path, n.Name.Name)
				}
			case *ast.FuncLit:
				if takesBytes(n.Type) {
					t.Errorf("%s: a func literal takes source bytes; the lowered tree carries what it reads", path)
				}
			case *ast.CallExpr:
				sel, ok := n.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "HasPrefix" {
					return true
				}
				for _, arg := range n.Args {
					if lit, ok := arg.(*ast.BasicLit); ok && lit.Value == `"grep"` {
						t.Errorf("%s: HasPrefix(..., \"grep\") re-derives an operation from source text", path)
					}
				}
			}
			return true
		})
	}
}

// takesBytes reports whether a function type has a []byte parameter.
func takesBytes(ft *ast.FuncType) bool {
	for _, p := range ft.Params.List {
		arr, ok := p.Type.(*ast.ArrayType)
		if !ok || arr.Len != nil {
			continue
		}
		if id, ok := arr.Elt.(*ast.Ident); ok && id.Name == "byte" {
			return true
		}
	}
	return false
}
