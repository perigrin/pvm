// ABOUTME: Tests for AST navigation utilities
// ABOUTME: Ensures navigation, search, and traversal functions work correctly

package astnav

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
)

func createTestAST() *ast.AST {
	// Create a simple test AST: my Int $x = 42;
	start := ast.Position{Line: 1, Column: 1, Offset: 0}
	end := ast.Position{Line: 1, Column: 15, Offset: 14}

	root := ast.NewBaseNode("root", start, end)

	// Variable declaration: my Int $x = 42;
	varStart := ast.Position{Line: 1, Column: 1, Offset: 0}
	varEnd := ast.Position{Line: 1, Column: 15, Offset: 14}

	variable := ast.NewVariableExpr("x", "$", ast.Position{Line: 1, Column: 8, Offset: 7}, ast.Position{Line: 1, Column: 10, Offset: 9})
	literal := ast.NewLiteralExpr("42", ast.NumberLiteral, ast.Position{Line: 1, Column: 13, Offset: 12}, ast.Position{Line: 1, Column: 15, Offset: 14})
	literal.SetText("42")

	typeExpr := &ast.TypeExpression{BaseType: "Int"}
	varDecl := ast.NewVarDecl("my", []*ast.VariableExpr{variable}, typeExpr, literal, varStart, varEnd)

	root.AddChild(varDecl)

	return &ast.AST{
		Path:   "/test/file.pl",
		Root:   root,
		Source: "my Int $x = 42;",
	}
}

func TestNewNavigator(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	if nav.root != testAST.Root {
		t.Error("Navigator root not properly set")
	}
}

func TestNavigator_FindNodeAt(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	// Find node at position where variable should be
	node := nav.FindNodeAt(1, 9) // Should find variable $x

	if node == nil {
		t.Fatal("Expected to find a node at position 1:9")
	}

	// Should find the variable expression
	if varExpr, ok := node.(*ast.VariableExpr); ok { //nolint:go-critic
		if varExpr.Name != "x" {
			t.Errorf("Expected variable 'x', got %q", varExpr.Name)
		}
	} else {
		t.Errorf("Expected VariableExpr, got %T", node)
	}
}

func TestNavigator_FindNodeAt_OutOfRange(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	// Find node at position outside the AST
	node := nav.FindNodeAt(10, 10)

	if node != nil {
		t.Error("Expected nil for position outside AST range")
	}
}

func TestNavigator_FindNodesByType(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	// Find all variable nodes
	variables := nav.FindNodesByType("variable")

	if len(variables) != 1 {
		t.Errorf("Expected 1 variable node, got %d", len(variables))
	}

	// Find all literal nodes
	literals := nav.FindNodesByType("literal")

	if len(literals) != 1 {
		t.Errorf("Expected 1 literal node, got %d", len(literals))
	}
}

func TestNavigator_FindVariables(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	variables := nav.FindVariables()

	if len(variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(variables))
	}

	if variables[0].Name != "x" {
		t.Errorf("Expected variable 'x', got %q", variables[0].Name)
	}
}

func TestNavigator_FindVariableByName(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	// Find by name
	variables := nav.FindVariableByName("x")

	if len(variables) != 1 {
		t.Errorf("Expected 1 variable named 'x', got %d", len(variables))
	}

	// Find by full name
	variables = nav.FindVariableByName("$x")

	if len(variables) != 1 {
		t.Errorf("Expected 1 variable named '$x', got %d", len(variables))
	}

	// Find non-existent variable
	variables = nav.FindVariableByName("y")

	if len(variables) != 0 {
		t.Errorf("Expected 0 variables named 'y', got %d", len(variables))
	}
}

func TestNavigator_FindTypeAnnotations(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	annotations := nav.FindTypeAnnotations()

	if len(annotations) != 1 {
		t.Errorf("Expected 1 type annotation, got %d", len(annotations))
	}

	if annotations[0].AnnotatedItem != "$x" {
		t.Errorf("Expected annotation for '$x', got %q", annotations[0].AnnotatedItem)
	}

	if annotations[0].Kind != ast.VarAnnotation {
		t.Errorf("Expected VarAnnotation, got %v", annotations[0].Kind)
	}
}

func TestNavigator_GetParent(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	// Find a child node
	variables := nav.FindVariables()
	if len(variables) == 0 {
		t.Fatal("Need at least one variable for this test")
	}

	variable := variables[0]
	parent := nav.GetParent(variable)

	if parent == nil {
		t.Error("Expected variable to have a parent")
	}

	// Root should have no parent
	rootParent := nav.GetParent(testAST.Root)
	if rootParent != nil {
		t.Error("Expected root to have no parent")
	}
}

func TestNavigator_GetAncestors(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	// Find a deeply nested node
	variables := nav.FindVariables()
	if len(variables) == 0 {
		t.Fatal("Need at least one variable for this test")
	}

	variable := variables[0]
	ancestors := nav.GetAncestors(variable)

	// Should have at least the variable declaration and root as ancestors
	if len(ancestors) < 2 {
		t.Errorf("Expected at least 2 ancestors, got %d", len(ancestors))
	}

	// Last ancestor should be root
	if ancestors[len(ancestors)-1] != testAST.Root {
		t.Error("Last ancestor should be root")
	}
}

func TestNavigator_Walk(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	visitedNodes := 0
	nav.Walk(testAST.Root, func(node ast.Node) bool {
		visitedNodes++
		return true // Continue traversal
	})

	if visitedNodes == 0 {
		t.Error("Expected to visit at least one node")
	}

	// Test early termination
	visitedWithStop := 0
	nav.Walk(testAST.Root, func(node ast.Node) bool {
		visitedWithStop++
		return false // Stop traversal
	})

	if visitedWithStop != 1 {
		t.Errorf("Expected to visit exactly 1 node with early termination, got %d", visitedWithStop)
	}
}

func TestNavigator_WalkWithContext(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	enteredNodes := 0
	exitedNodes := 0

	walkFunc := ast.WalkFunc{
		Enter: func(node ast.Node) bool {
			enteredNodes++
			return true // Continue to children
		},
		Exit: func(node ast.Node) {
			exitedNodes++
		},
	}

	nav.WalkWithContext(testAST.Root, walkFunc)

	if enteredNodes == 0 {
		t.Error("Expected to enter at least one node")
	}

	if exitedNodes != enteredNodes {
		t.Errorf("Expected equal enter (%d) and exit (%d) counts", enteredNodes, exitedNodes)
	}
}

func TestNavigator_GetNodePath(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	// Get path for root
	rootPath := nav.GetNodePath(testAST.Root)
	if rootPath == "" {
		t.Error("Expected non-empty path for root")
	}

	// Get path for a variable
	variables := nav.FindVariables()
	if len(variables) > 0 {
		varPath := nav.GetNodePath(variables[0])
		if varPath == "" {
			t.Error("Expected non-empty path for variable")
		}

		// Variable path should be longer than root path
		if len(varPath) <= len(rootPath) {
			t.Error("Expected variable path to be longer than root path")
		}
	}
}

func TestNavigator_CountNodes(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	count := nav.CountNodes()

	if count == 0 {
		t.Error("Expected to count at least one node")
	}

	// Should count all nodes in the tree
	expectedMinimum := 4 // root, var_decl, variable, literal
	if count < expectedMinimum {
		t.Errorf("Expected at least %d nodes, got %d", expectedMinimum, count)
	}
}

func TestNavigator_GetDepth(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	depth := nav.GetDepth()

	if depth < 2 {
		t.Errorf("Expected depth of at least 2, got %d", depth)
	}
}

func TestNavigator_FindNodeByText(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	// Find nodes containing "42"
	nodes := nav.FindNodeByText("42")

	if len(nodes) == 0 {
		t.Error("Expected to find at least one node containing '42'")
	}
}

func TestNavigator_IsDescendantOf(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	variables := nav.FindVariables()
	if len(variables) == 0 {
		t.Fatal("Need at least one variable for this test")
	}

	variable := variables[0]

	// Variable should be descendant of root
	if !nav.IsDescendantOf(variable, testAST.Root) {
		t.Error("Expected variable to be descendant of root")
	}

	// Root should not be descendant of variable
	if nav.IsDescendantOf(testAST.Root, variable) {
		t.Error("Expected root to not be descendant of variable")
	}

	// Node should not be descendant of itself
	if nav.IsDescendantOf(variable, variable) {
		t.Error("Expected node to not be descendant of itself")
	}
}

func TestNavigator_FindCommonAncestor(t *testing.T) {
	testAST := createTestAST()
	nav := NewNavigator(testAST.Root)

	variables := nav.FindVariables()
	literals := nav.FindNodesByType("literal")

	if len(variables) == 0 || len(literals) == 0 {
		t.Fatal("Need at least one variable and one literal for this test")
	}

	variable := variables[0]
	literal := literals[0]

	ancestor := nav.FindCommonAncestor(variable, literal)

	if ancestor == nil {
		t.Error("Expected to find common ancestor")
	}

	// Both nodes should be descendants of the common ancestor
	if !nav.IsDescendantOf(variable, ancestor) {
		t.Error("Expected variable to be descendant of common ancestor")
	}

	if !nav.IsDescendantOf(literal, ancestor) {
		t.Error("Expected literal to be descendant of common ancestor")
	}
}
