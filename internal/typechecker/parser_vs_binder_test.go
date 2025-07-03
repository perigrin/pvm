package typechecker

import (
	"fmt"
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/parser"
)

func TestParserVsBinderMismatch(t *testing.T) {
	code := `my Int $count = 42;
my Str $name = "Alice";`

	// Parse
	p, err := parser.NewParser()
	if err != nil {
		t.Fatal(err)
	}

	astTree, err := p.ParseString(code)
	if err != nil {
		t.Fatal(err)
	}

	// Check what parser found
	fmt.Printf("=== PARSER TYPE ANNOTATIONS ===\n")
	fmt.Printf("AST has %d type annotations:\n", len(astTree.TypeAnnotations))
	for i, annotation := range astTree.TypeAnnotations {
		fmt.Printf("  [%d] %s\n", i, annotation.String())
	}

	// Check what VarDecl.LogicalVariables() returns
	fmt.Printf("\n=== VARDECL LOGICAL VARIABLES ===\n")
	findVarDeclsLogicalVars(astTree.Root)
}

func findVarDeclsLogicalVars(node interface{}) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.VarDecl:
		fmt.Printf("VarDecl found:\n")
		vars := n.LogicalVariables()
		fmt.Printf("  LogicalVariables count: %d\n", len(vars))
		for i, v := range vars {
			if v != nil {
				fmt.Printf("    [%d] Name: '%s'\n", i, v.Name)
			} else {
				fmt.Printf("    [%d] nil\n", i)
			}
		}
	}

	// Recurse into children
	switch n := node.(type) {
	case interface{ Children() []interface{} }:
		children := n.Children()
		for _, child := range children {
			findVarDeclsLogicalVars(child)
		}
	}
}
