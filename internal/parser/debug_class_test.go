// ABOUTME: Debug test to see what AST nodes are created for class declarations
// ABOUTME: Temporary file for debugging Step 20 implementation

package parser

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
)

func TestDebugClassParsing(t *testing.T) {
	input := `class User {
    field Str $name;

    method User new(Str $name) {
        return bless { name => $name }, __PACKAGE__;
    }
}`

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	astResult, err := parser.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	if astResult == nil {
		t.Fatal("Parser returned nil AST")
	}

	t.Logf("AST:\n%s", astResult.String())
	printASTNodes(t, astResult.Root, 0)
}

// printASTNodes prints AST node structure for debugging
func printASTNodes(t *testing.T, node interface{}, depth int) {
	if node == nil {
		return
	}

	// Use type assertion to get ast.Node interface
	if astNode, ok := node.(interface {
		Type() string
		Text() string
		Children() []interface{}
	}); ok {
		indent := strings.Repeat("  ", depth)
		t.Logf("%s%s: %q", indent, astNode.Type(), astNode.Text())

		if children := astNode.Children(); children != nil {
			for _, child := range children {
				printASTNodes(t, child, depth+1)
			}
		}
	}
}
func TestSimpleClassDebug(t *testing.T) {
	input := `class User {
    field Str $name;
    field Int $age;

    method User new(Str $name, Int $age) {
        return bless { name => $name, age => $age }, __PACKAGE__;
    }

    method Str get_name() {
        return $name;
    }
}`

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	astResult, err := parser.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	if astResult == nil {
		t.Fatal("Parser returned nil AST")
	}

	// Log AST structure
	t.Logf("AST Root Type: %T", astResult.Root)
	t.Logf("AST Root: %s", astResult.Root.Type())

	// Check children of root node
	children := astResult.Root.Children()
	t.Logf("Root has %d children", len(children))
	for i, child := range children {
		t.Logf("Child %d: %T - %s", i, child, child.Type())
		if classDecl, ok := child.(*ast.ClassDecl); ok {
			t.Logf("Found ClassDecl: %s with %d fields and %d methods", classDecl.Name, len(classDecl.Fields), len(classDecl.Methods))
		}
	}
}
