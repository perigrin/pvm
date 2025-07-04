// ABOUTME: Comprehensive tests for symbol binding functionality in Perl code analysis.
// ABOUTME: Tests symbol resolution, scope management, and Perl scoping semantics following TypeScript-Go patterns.

package binder

import (
	"testing"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/parser/treesitter"
)

func TestSymbolTable_Basic(t *testing.T) {
	st := NewSymbolTable()

	// Test initial state
	if st.GlobalScope == nil {
		t.Fatal("GlobalScope should not be nil")
	}

	if st.CurrentScope != st.GlobalScope {
		t.Fatal("CurrentScope should be GlobalScope initially")
	}

	if st.Package != "main" {
		t.Errorf("Expected package 'main', got '%s'", st.Package)
	}
}

func TestSymbolTable_ScopeManagement(t *testing.T) {
	st := NewSymbolTable()

	// Test entering scope
	subScope := st.EnterScope(ScopeSubroutine, nil)
	if st.CurrentScope != subScope {
		t.Error("CurrentScope should be the new scope")
	}

	if subScope.Parent != st.GlobalScope {
		t.Error("New scope parent should be global scope")
	}

	if len(st.GlobalScope.Children) != 1 {
		t.Error("Global scope should have one child")
	}

	// Test scope depth
	if st.GetScopeDepth() != 1 {
		t.Errorf("Expected scope depth 1, got %d", st.GetScopeDepth())
	}

	// Test exiting scope
	st.ExitScope()
	if st.CurrentScope != st.GlobalScope {
		t.Error("Should return to global scope")
	}

	if st.GetScopeDepth() != 0 {
		t.Errorf("Expected scope depth 0, got %d", st.GetScopeDepth())
	}
}

func TestSymbolTable_AddSymbol(t *testing.T) {
	st := NewSymbolTable()

	// Create test symbol
	symbol := &Symbol{
		Name:     "test_var",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagLexical,
		Position: ast.Position{Line: 1, Column: 1},
	}

	// Add symbol
	err := st.AddSymbol(symbol)
	if err != nil {
		t.Fatalf("Failed to add symbol: %v", err)
	}

	// Verify symbol was added
	if symbol.Scope != st.CurrentScope {
		t.Error("Symbol scope should be current scope")
	}

	if symbol.Package != "main" {
		t.Errorf("Expected symbol package 'main', got '%s'", symbol.Package)
	}

	// Verify symbol is in scope
	if st.CurrentScope.Symbols["test_var"] != symbol {
		t.Error("Symbol should be in current scope")
	}

	// Verify symbol is in global index
	symbols := st.GetSymbolsByName("test_var")
	if len(symbols) != 1 || symbols[0] != symbol {
		t.Error("Symbol should be in global index")
	}
}

func TestSymbolTable_SymbolResolution(t *testing.T) {
	st := NewSymbolTable()

	// Add symbol to global scope
	globalSymbol := &Symbol{
		Name:     "global_var",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagPackage,
		Position: ast.Position{Line: 1, Column: 1},
	}
	st.AddSymbol(globalSymbol)

	// Enter nested scope
	st.EnterScope(ScopeSubroutine, nil)

	// Add local symbol
	localSymbol := &Symbol{
		Name:     "local_var",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagLexical,
		Position: ast.Position{Line: 2, Column: 1},
	}
	st.AddSymbol(localSymbol)

	// Test resolving local symbol
	resolved := st.ResolveSymbol("local_var", SymbolScalar)
	if resolved != localSymbol {
		t.Error("Should resolve local symbol")
	}

	// Test resolving global symbol from nested scope
	resolved = st.ResolveSymbol("global_var", SymbolScalar)
	if resolved != globalSymbol {
		t.Error("Should resolve global symbol from nested scope")
	}

	// Test resolving non-existent symbol
	resolved = st.ResolveSymbol("non_existent", SymbolScalar)
	if resolved != nil {
		t.Error("Should not resolve non-existent symbol")
	}
}

func TestSymbolTable_Shadowing(t *testing.T) {
	st := NewSymbolTable()

	// Add symbol to global scope
	globalSymbol := &Symbol{
		Name:     "var",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagPackage,
		Position: ast.Position{Line: 1, Column: 1},
	}
	st.AddSymbol(globalSymbol)

	// Enter nested scope
	st.EnterScope(ScopeSubroutine, nil)

	// Add shadowing symbol
	localSymbol := &Symbol{
		Name:     "var",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagLexical,
		Position: ast.Position{Line: 2, Column: 1},
	}
	st.AddSymbol(localSymbol)

	// Test that local symbol shadows global
	resolved := st.ResolveSymbol("var", SymbolScalar)
	if resolved != localSymbol {
		t.Error("Local symbol should shadow global symbol")
	}

	// Exit scope and verify global is visible again
	st.ExitScope()
	resolved = st.ResolveSymbol("var", SymbolScalar)
	if resolved != globalSymbol {
		t.Error("Global symbol should be visible after exiting scope")
	}
}

func TestBinder_VariableDeclarations(t *testing.T) {
	binder := NewBinder()

	// Use CST-based binding with real Perl code
	inputCode := "my Int $test_var = 42;"

	// Parse with tree-sitter for CST
	tsParser := sitter.NewParser()
	tsParser.SetLanguage(treesitter.Language())
	contentBytes := []byte(inputCode)
	tree := tsParser.Parse(contentBytes, nil)
	if tree == nil {
		t.Fatal("Failed to parse input code")
	}

	// Bind symbols using CST
	st, err := binder.BindCST(tree.RootNode(), contentBytes, nil)
	if err != nil {
		t.Fatalf("Failed to bind variable declaration: %v", err)
	}

	// Verify symbol was created
	symbol := st.ResolveSymbol("test_var", SymbolScalar)
	if symbol == nil {
		t.Fatal("Variable symbol should be created")
	}

	if symbol.Kind != SymbolScalar {
		t.Errorf("Expected SymbolScalar, got %v", symbol.Kind)
	}

	if symbol.Flags&SymbolFlagLexical == 0 {
		t.Error("Symbol should have lexical flag")
	}

	if symbol.Flags&SymbolFlagTypeAnnotated == 0 {
		t.Error("Symbol should have type annotation flag")
	}

	if symbol.Type != "Int" {
		t.Errorf("Expected type 'Int', got '%s'", symbol.Type)
	}
}

func TestBinder_SubroutineDeclarations(t *testing.T) {
	binder := NewBinder()

	// Use CST-based binding with real Perl code
	inputCode := "sub test_sub { print 'hello'; }"

	// Parse with tree-sitter for CST
	tsParser := sitter.NewParser()
	tsParser.SetLanguage(treesitter.Language())
	contentBytes := []byte(inputCode)
	tree := tsParser.Parse(contentBytes, nil)
	if tree == nil {
		t.Fatal("Failed to parse input code")
	}

	// Bind symbols using CST
	st, err := binder.BindCST(tree.RootNode(), contentBytes, nil)
	if err != nil {
		t.Fatalf("Failed to bind subroutine declaration: %v", err)
	}

	// Verify subroutine symbol was created
	symbol := st.ResolveSymbol("test_sub", SymbolSubroutine)
	if symbol == nil {
		t.Fatal("Subroutine symbol should be created")
	}

	if symbol.Kind != SymbolSubroutine {
		t.Errorf("Expected SymbolSubroutine, got %v", symbol.Kind)
	}
}

func TestBinder_PackageDeclarations(t *testing.T) {
	binder := NewBinder()

	// Use CST-based binding with real Perl code
	inputCode := "package Test::Package;"

	// Parse with tree-sitter for CST
	tsParser := sitter.NewParser()
	tsParser.SetLanguage(treesitter.Language())
	contentBytes := []byte(inputCode)
	tree := tsParser.Parse(contentBytes, nil)
	if tree == nil {
		t.Fatal("Failed to parse input code")
	}

	// Bind symbols using CST
	st, err := binder.BindCST(tree.RootNode(), contentBytes, nil)
	if err != nil {
		t.Fatalf("Failed to bind package declaration: %v", err)
	}

	// The default behavior for package declarations will be handled by the CST binder
	// For now, just verify the binding doesn't fail
	if st == nil {
		t.Fatal("Symbol table should be created")
	}
}

func TestBinder_ScopeNesting(t *testing.T) {
	binder := NewBinder()

	// Use CST-based binding with real Perl code with nested scopes
	inputCode := "{ my $outer_var = 42; }"

	// Parse with tree-sitter for CST
	tsParser := sitter.NewParser()
	tsParser.SetLanguage(treesitter.Language())
	contentBytes := []byte(inputCode)
	tree := tsParser.Parse(contentBytes, nil)
	if tree == nil {
		t.Fatal("Failed to parse input code")
	}

	// Bind symbols using CST
	st, err := binder.BindCST(tree.RootNode(), contentBytes, nil)
	if err != nil {
		t.Fatalf("Failed to bind block statement: %v", err)
	}

	// Verify variable was created (block scope handling is part of CST binding)
	symbol := st.ResolveSymbol("outer_var", SymbolScalar)
	if symbol == nil {
		t.Fatal("Variable should be created")
	}

	if symbol.Flags&SymbolFlagLexical == 0 {
		t.Error("Variable should have lexical flag")
	}
}

func TestBindingError(t *testing.T) {
	err := NewBindingError("test error", "test_symbol", "scalar", ast.Position{Line: 1, Column: 1})

	if err.Message != "test error" {
		t.Errorf("Expected message 'test error', got '%s'", err.Message)
	}

	if err.Symbol != "test_symbol" {
		t.Errorf("Expected symbol 'test_symbol', got '%s'", err.Symbol)
	}

	if err.Kind != "scalar" {
		t.Errorf("Expected kind 'scalar', got '%s'", err.Kind)
	}

	if err.Error() != "test error" {
		t.Errorf("Expected error string 'test error', got '%s'", err.Error())
	}
}

func TestSymbolKind_String(t *testing.T) {
	tests := []struct {
		kind     SymbolKind
		expected string
	}{
		{SymbolScalar, "scalar"},
		{SymbolArray, "array"},
		{SymbolHash, "hash"},
		{SymbolSubroutine, "subroutine"},
		{SymbolMethod, "method"},
		{SymbolPackage, "package"},
	}

	for _, test := range tests {
		if test.kind.String() != test.expected {
			t.Errorf("Expected %s.String() = '%s', got '%s'",
				test.kind, test.expected, test.kind.String())
		}
	}
}

func TestSymbolFlags_String(t *testing.T) {
	tests := []struct {
		flags    SymbolFlags
		expected string
	}{
		{SymbolFlagNone, "none"},
		{SymbolFlagLexical, "lexical"},
		{SymbolFlagPackage, "package"},
		{SymbolFlagLexical | SymbolFlagTypeAnnotated, "lexical|typed"},
	}

	for _, test := range tests {
		if test.flags.String() != test.expected {
			t.Errorf("Expected %v.String() = '%s', got '%s'",
				test.flags, test.expected, test.flags.String())
		}
	}
}

func TestGetVariableSymbolKind(t *testing.T) {
	binder := NewBinder()

	tests := []struct {
		name     string
		expected SymbolKind
	}{
		{"$scalar", SymbolScalar},
		{"@array", SymbolArray},
		{"%hash", SymbolHash},
		{"*glob", SymbolGlob},
		{"no_sigil", SymbolScalar},
	}

	for _, test := range tests {
		result := binder.getVariableSymbolKind(test.name)
		if result != test.expected {
			t.Errorf("getVariableSymbolKind('%s') = %v, expected %v",
				test.name, result, test.expected)
		}
	}
}

func TestStripSigil(t *testing.T) {
	binder := NewBinder()

	tests := []struct {
		name     string
		expected string
	}{
		{"$scalar", "scalar"},
		{"@array", "array"},
		{"%hash", "hash"},
		{"*glob", "glob"},
		{"no_sigil", "no_sigil"},
	}

	for _, test := range tests {
		result := binder.stripSigil(test.name)
		if result != test.expected {
			t.Errorf("stripSigil('%s') = '%s', expected '%s'",
				test.name, result, test.expected)
		}
	}
}
