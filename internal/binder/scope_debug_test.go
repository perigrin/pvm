package binder

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/parser"
)

func TestScopeDebugBug(t *testing.T) {
	// Enable debug mode
	DebugScoping = true
	defer func() { DebugScoping = false }()

	// Test the exact problematic code from the prompt plan
	// This SHOULD work (methods with same local variable names should be isolated)
	// but currently fails due to scope sharing bug
	code := `method process0() {
    my $result = 1;
    return $result;
}

method process1() {
    my $result = 2;
    return $result;
}`

	t.Logf("=== DEBUGGING METHOD SCOPE SHARING ===")
	t.Logf("Testing code:\n%s", code)

	// Parse the code
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	astTree, err := p.ParseString(code)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Log the AST structure to understand what we're working with
	t.Logf("AST Root: %T", astTree.Root)
	if astTree.Root != nil {
		logASTStructure(t, astTree.Root, 0)
	}

	t.Logf("=== STARTING SYMBOL BINDING ===")

	// Create binder and bind
	binder := NewBinder()
	symbolTable, err := binder.BindAST(astTree)

	if err != nil {
		t.Logf("=== BINDING ERROR OCCURRED ===")
		t.Logf("Error: %v", err)
		t.Errorf("Binding should not fail for isolated method scopes, but got: %v", err)
	} else {
		t.Logf("=== BINDING SUCCEEDED ===")
		t.Logf("Binding succeeded - methods have properly isolated scopes")
	}

	// Log the scope structure if we got a symbol table
	if symbolTable != nil {
		t.Logf("=== SCOPE HIERARCHY ANALYSIS ===")
		logScopeStructure(t, symbolTable.GlobalScope, 0)

		// Analyze the symbol table structure
		t.Logf("=== SYMBOL TABLE ANALYSIS ===")
		t.Logf("Total symbols: %d", len(symbolTable.Symbols))

		for name, symbols := range symbolTable.Symbols {
			t.Logf("Symbol '%s': %d instances", name, len(symbols))
			for i, symbol := range symbols {
				t.Logf("  [%d] %s %s in %s scope ID=%d",
					i, symbol.Kind.String(), symbol.Name,
					symbol.Scope.Kind.String(), symbol.Scope.ScopeID)
			}
		}
	}
}

func logASTStructure(t *testing.T, node interface{}, depth int) {
	if node == nil {
		return
	}

	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	t.Logf("%sNode: %T", indent, node)

	// Log method-specific information
	switch n := node.(type) {
	case *ast.MethodDecl:
		t.Logf("%s  Method Name: %s", indent, n.Name)
		t.Logf("%s  Has Body: %v", indent, n.Body != nil)
		if n.Body != nil {
			t.Logf("%s  Body Type: %T", indent, n.Body)
		}
	}

	// Try to get children if this is a container node
	switch n := node.(type) {
	case interface{ Children() []interface{} }:
		children := n.Children()
		for _, child := range children {
			logASTStructure(t, child, depth+1)
		}
	}
}

func logScopeStructure(t *testing.T, scope *Scope, depth int) {
	if scope == nil {
		return
	}

	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	t.Logf("%sScope ID=%d Kind=%s Symbols=%d",
		indent, scope.ScopeID, scope.Kind.String(), len(scope.Symbols))

	for name, symbol := range scope.Symbols {
		t.Logf("%s  Symbol: %s %s", indent, symbol.Kind.String(), name)
	}

	for _, child := range scope.Children {
		logScopeStructure(t, child, depth+1)
	}
}
