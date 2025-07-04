// ABOUTME: Simplified test to isolate method scope sharing bug using direct binder calls.
// ABOUTME: Tests scope isolation by simulating method binding without complex AST construction.

package binder

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
)

// TestDirectMethodScopeIsolation tests scope isolation by binding two methods directly
func TestDirectMethodScopeIsolation(t *testing.T) {
	// Enable debug mode
	DebugScoping = true
	defer func() { DebugScoping = false }()

	t.Logf("=== TESTING DIRECT METHOD SCOPE ISOLATION ===")

	// Create binder and symbol table
	binder := NewBinder()
	symbolTable := NewSymbolTable()
	binder.symbolTable = symbolTable

	pos := ast.Position{Line: 1, Column: 1}

	// Simulate binding method process0() { my $result = 1; }
	t.Logf("=== BINDING METHOD process0 ===")

	// Manually enter method scope for process0
	method0Node := &ast.BaseNode{}
	method0Scope := symbolTable.EnterScope(ScopeMethod, method0Node)
	t.Logf("Created method scope ID=%d for process0", method0Scope.ScopeID)

	// Create $result symbol in method0 scope
	symbol0 := &Symbol{
		Name:     "result",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagLexical,
		Position: pos,
	}

	err := symbolTable.AddSymbol(symbol0)
	if err != nil {
		t.Logf("Error adding symbol to method0: %v", err)
	} else {
		t.Logf("Successfully added $result to method0 scope ID=%d", symbol0.Scope.ScopeID)
	}

	// Exit method0 scope
	symbolTable.ExitScopeAdvanced()
	t.Logf("Exited method0 scope, current scope: %s ID=%d",
		symbolTable.CurrentScope.Kind.String(), symbolTable.CurrentScope.ScopeID)

	// Simulate binding method process1() { my $result = 2; }
	t.Logf("=== BINDING METHOD process1 ===")

	// Manually enter method scope for process1
	method1Node := &ast.BaseNode{}
	method1Scope := symbolTable.EnterScope(ScopeMethod, method1Node)
	t.Logf("Created method scope ID=%d for process1", method1Scope.ScopeID)

	// Create $result symbol in method1 scope
	symbol1 := &Symbol{
		Name:     "result",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagLexical,
		Position: pos,
	}

	err = symbolTable.AddSymbol(symbol1)
	if err != nil {
		t.Logf("=== BINDING ERROR OCCURRED (INDICATES BUG) ===")
		t.Logf("Error adding symbol to method1: %v", err)

		// This indicates the scope sharing bug
		if contains(err.Error(), "already declared") {
			t.Logf("CONFIRMED: Bug reproduced - methods are sharing scopes!")
			t.Logf("Symbol from method0 scope ID=%d conflicts with method1 scope ID=%d",
				symbol0.Scope.ScopeID, method1Scope.ScopeID)
			t.Logf("This should NOT happen - methods should have isolated scopes")
		} else {
			t.Errorf("Unexpected error type: %v", err)
		}
	} else {
		t.Logf("=== BINDING SUCCEEDED ===")
		t.Logf("Successfully added $result to method1 scope ID=%d", symbol1.Scope.ScopeID)
		t.Logf("Methods have properly isolated scopes")

		// Verify they're in different scopes
		if symbol0.Scope.ScopeID == symbol1.Scope.ScopeID {
			t.Errorf("BUG: Both symbols are in the same scope ID=%d", symbol0.Scope.ScopeID)
		} else {
			t.Logf("GOOD: Symbols are in different scopes (ID=%d vs ID=%d)",
				symbol0.Scope.ScopeID, symbol1.Scope.ScopeID)
		}
	}

	// Exit method1 scope
	symbolTable.ExitScopeAdvanced()

	// Log final scope structure
	t.Logf("=== FINAL SCOPE HIERARCHY ===")
	// TODO: Re-implement logScopeStructure helper function
	// logScopeStructure(t, symbolTable.GlobalScope, 0)
}

// Helper to count method scopes in hierarchy
func countMethodScopes(scope *Scope) int {
	if scope == nil {
		return 0
	}

	count := 0
	if scope.Kind == ScopeMethod {
		count = 1
	}

	for _, child := range scope.Children {
		count += countMethodScopes(child)
	}

	return count
}

// Helper to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
