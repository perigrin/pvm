// ABOUTME: Tests for enhanced diagnostics system with symbol-aware error reporting
// ABOUTME: Validates undefined variable detection, shadowing warnings, and usage tracking

package diagnostics

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
)

func TestEnhancedDiagnosticEngine_UndefinedVariables(t *testing.T) {
	// Create a simple symbol table with one variable
	symbolTable := binder.NewSymbolTable()
	symbolTable.EnterScope(binder.ScopeBlock, nil)

	definedSymbol := &binder.Symbol{
		Name:     "$defined_var",
		Kind:     binder.SymbolScalar,
		Position: ast.Position{Line: 1, Column: 5},
		Type:     "Int",
	}
	symbolTable.AddSymbol(definedSymbol)

	// Create diagnostic engine
	sourceContent := `my Int $defined_var = 42;
print $undefined_var;  # This should trigger undefined variable error`

	engine := NewEnhancedDiagnosticEngine(symbolTable, "test.pl", sourceContent)

	// Create AST with undefined variable reference
	undefinedVar := ast.NewVariableExpr("undefined_var", "$", ast.Position{Line: 2, Column: 7}, ast.Position{Line: 2, Column: 20})

	// Analyze symbols
	diagnostics := engine.findUndefinedVariables(undefinedVar)

	// Should find one undefined variable error
	if len(diagnostics) != 1 {
		t.Errorf("Expected 1 undefined variable diagnostic, got %d", len(diagnostics))
		return
	}

	diagnostic := diagnostics[0]
	if diagnostic.Kind != DiagnosticError {
		t.Errorf("Expected DiagnosticError, got %v", diagnostic.Kind)
	}

	if diagnostic.SymbolName != "$undefined_var" {
		t.Errorf("Expected symbol name '$undefined_var', got '%s'", diagnostic.SymbolName)
	}

	if diagnostic.Code != "PSC-E001" {
		t.Errorf("Expected error code 'PSC-E001', got '%s'", diagnostic.Code)
	}
}

func TestEnhancedDiagnosticEngine_ShadowedVariables(t *testing.T) {
	// Create symbol table with shadowed variables
	symbolTable := binder.NewSymbolTable()
	symbolTable.EnterScope(binder.ScopeBlock, nil)

	// Outer variable
	outerSymbol := &binder.Symbol{
		Name:     "$var",
		Kind:     binder.SymbolScalar,
		Position: ast.Position{Line: 1, Column: 5},
		Type:     "Int",
	}
	symbolTable.AddSymbol(outerSymbol)

	symbolTable.EnterScope(binder.ScopeBlock, nil)

	// Inner variable (shadows outer)
	innerSymbol := &binder.Symbol{
		Name:     "$var",
		Kind:     binder.SymbolScalar,
		Position: ast.Position{Line: 3, Column: 9},
		Type:     "Str",
	}
	symbolTable.AddSymbol(innerSymbol)

	sourceContent := `my Int $var = 42;
{
    my Str $var = "hello";  # This should trigger shadowing warning
}`

	engine := NewEnhancedDiagnosticEngine(symbolTable, "test.pl", sourceContent)

	// Find shadowed variables
	diagnostics := engine.findShadowedVariables()

	// Should find one shadowing warning
	if len(diagnostics) != 1 {
		t.Errorf("Expected 1 shadowing diagnostic, got %d", len(diagnostics))
		return
	}

	diagnostic := diagnostics[0]
	if diagnostic.Kind != DiagnosticWarning {
		t.Errorf("Expected DiagnosticWarning, got %v", diagnostic.Kind)
	}

	if diagnostic.Code != "PSC-W001" {
		t.Errorf("Expected error code 'PSC-W001', got '%s'", diagnostic.Code)
	}
}

func TestEnhancedDiagnosticEngine_UnusedVariables(t *testing.T) {
	// Create symbol table with unused variable
	symbolTable := binder.NewSymbolTable()
	symbolTable.EnterScope(binder.ScopeBlock, nil)

	unusedSymbol := &binder.Symbol{
		Name:     "$unused_var",
		Kind:     binder.SymbolScalar,
		Position: ast.Position{Line: 1, Column: 5},
		Type:     "Int",
	}
	symbolTable.AddSymbol(unusedSymbol)

	sourceContent := `my Int $unused_var = 42;  # This variable is never used`

	engine := NewEnhancedDiagnosticEngine(symbolTable, "test.pl", sourceContent)

	// Create empty AST (no usage of the variable)
	emptyRoot := ast.NewBaseNode("root", ast.Position{}, ast.Position{})

	// Track usage (should find no usage)
	engine.usageTracker.TrackUsage(emptyRoot, symbolTable)

	// Find unused variables
	diagnostics := engine.findUnusedVariables()

	// Should find one unused variable warning
	if len(diagnostics) != 1 {
		t.Errorf("Expected 1 unused variable diagnostic, got %d", len(diagnostics))
		return
	}

	diagnostic := diagnostics[0]
	if diagnostic.Kind != DiagnosticWarning {
		t.Errorf("Expected DiagnosticWarning, got %v", diagnostic.Kind)
	}

	if diagnostic.Code != "PSC-W002" {
		t.Errorf("Expected error code 'PSC-W002', got '%s'", diagnostic.Code)
	}
}

func TestDiagnostic_FormatDiagnostic(t *testing.T) {
	diagnostic := Diagnostic{
		Kind:        DiagnosticError,
		Message:     "Undefined variable '$typo'",
		Pos:         ast.Position{Line: 5, Column: 10},
		SymbolName:  "$typo",
		FilePath:    "test.pl",
		LineText:    "print $typo;",
		Suggestion:  "Did you mean '$type'?",
		DidYouMean:  []string{"$type", "$temp"},
		Code:        "PSC-E001",
		HelpMessage: "Variables must be declared before use",
	}

	// Test without color
	formatted := diagnostic.FormatDiagnostic(false)

	expectedSubstrings := []string{
		"test.pl:5:10:",
		"error:",
		"Undefined variable '$typo'",
		"[PSC-E001]",
		"5 | print $typo;",
		"^",
		"help: Did you mean '$type'?",
		"note: Variables must be declared before use",
		"note: Did you mean: $type, $temp",
	}

	for _, substr := range expectedSubstrings {
		if !containsString(formatted, substr) {
			t.Errorf("Expected formatted diagnostic to contain '%s', got:\n%s", substr, formatted)
		}
	}
}

func TestSymbolUsageTracker_TrackUsage(t *testing.T) {
	tracker := NewSymbolUsageTracker()
	symbolTable := binder.NewSymbolTable()
	symbolTable.EnterScope(binder.ScopeBlock, nil)

	// Add a symbol
	symbol := &binder.Symbol{
		Name:     "$var",
		Kind:     binder.SymbolScalar,
		Position: ast.Position{Line: 1, Column: 5},
		Type:     "Int",
	}
	symbolTable.AddSymbol(symbol)

	// Create AST with variable reference
	varRef := ast.NewVariableExpr("var", "$", ast.Position{Line: 2, Column: 7}, ast.Position{Line: 2, Column: 11})

	// Track usage
	tracker.TrackUsage(varRef, symbolTable)

	// Check if symbol is used
	if !tracker.IsSymbolUsed("$var") {
		t.Error("Expected symbol '$var' to be marked as used")
	}

	// Check reference positions
	refs := tracker.GetSymbolReferences("$var")
	if len(refs) != 1 {
		t.Errorf("Expected 1 reference, got %d", len(refs))
	}

	if refs[0].Line != 2 || refs[0].Column != 7 {
		t.Errorf("Expected reference at line 2, column 7, got line %d, column %d", refs[0].Line, refs[0].Column)
	}
}

func TestEditDistance(t *testing.T) {
	engine := &EnhancedDiagnosticEngine{}

	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"hello", "hello", 0},
		{"hello", "helo", 1},
		{"hello", "help", 2},
		{"hello", "world", 4},
		{"$var", "$var1", 1},
		{"$typo", "$type", 1},
	}

	for _, test := range tests {
		result := engine.editDistance(test.s1, test.s2)
		if result != test.expected {
			t.Errorf("editDistance('%s', '%s') = %d, expected %d", test.s1, test.s2, result, test.expected)
		}
	}
}

// Helper function to check if a string contains a substring
func containsString(text, substr string) bool {
	return len(text) >= len(substr) && findSubstring(text, substr) != -1
}

// Simple substring search
func findSubstring(text, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(text) {
		return -1
	}

	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
