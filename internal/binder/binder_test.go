// ABOUTME: Comprehensive tests for symbol binding functionality in Perl code analysis.
// ABOUTME: Tests symbol resolution, scope management, and Perl scoping semantics following TypeScript-Go patterns.

package binder

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
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

	// Create test variable declaration
	variable := &ast.VariableExpr{
		BaseNode: ast.NewBaseNode("variable", ast.Position{Line: 1, Column: 5}, ast.Position{Line: 1, Column: 14}),
		Name:     "$test_var",
		Sigil:    "$",
	}

	typeExpr := &ast.TypeExpression{
		BaseType: "Int",
		Pos:      ast.Position{Line: 1, Column: 1},
	}

	varDecl := ast.NewVarDecl(
		"my",
		[]*ast.VariableExpr{variable},
		typeExpr,
		nil,
		ast.Position{Line: 1, Column: 1},
		ast.Position{Line: 1, Column: 20},
	)

	st, err := binder.Bind(varDecl)
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

	// Create test subroutine declaration
	param := &ast.Parameter{
		Name:     "$param1",
		TypeExpr: &ast.TypeExpression{BaseType: "Int"},
		Pos:      ast.Position{Line: 1, Column: 10},
	}

	returnType := &ast.TypeExpression{
		BaseType: "Str",
	}

	body := ast.NewBlockStmt(
		[]ast.StatementNode{},
		ast.Position{Line: 1, Column: 20},
		ast.Position{Line: 1, Column: 22},
	)

	subDecl := ast.NewSubDecl(
		"test_sub",
		[]*ast.Parameter{param},
		returnType,
		body,
		false,
		ast.Position{Line: 1, Column: 1},
		ast.Position{Line: 1, Column: 25},
	)

	st, err := binder.Bind(subDecl)
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

	if symbol.Type != "Str" {
		t.Errorf("Expected return type 'Str', got '%s'", symbol.Type)
	}

	// Verify scope was created for subroutine
	subScope := st.FindScope(subDecl)
	if subScope == nil {
		t.Fatal("Subroutine scope should be created")
	}

	if subScope.Kind != ScopeSubroutine {
		t.Errorf("Expected ScopeSubroutine, got %v", subScope.Kind)
	}

	// Verify parameter symbol was created in subroutine scope
	paramSymbol := st.ResolveInScope(subScope, "param1", SymbolScalar)
	if paramSymbol == nil {
		t.Fatal("Parameter symbol should be created")
	}

	if paramSymbol.Flags&SymbolFlagLexical == 0 {
		t.Error("Parameter should have lexical flag")
	}
}

func TestBinder_PackageDeclarations(t *testing.T) {
	binder := NewBinder()

	// Create test package declaration
	pkgStmt := ast.NewPackageStmt(
		"Test::Package",
		"",
		ast.Position{Line: 1, Column: 1},
		ast.Position{Line: 1, Column: 20},
	)

	st, err := binder.Bind(pkgStmt)
	if err != nil {
		t.Fatalf("Failed to bind package declaration: %v", err)
	}

	// Verify package was set
	if st.Package != "Test::Package" {
		t.Errorf("Expected package 'Test::Package', got '%s'", st.Package)
	}

	// Verify package symbol was created
	symbol := st.ResolveSymbol("Test::Package", SymbolPackage)
	if symbol == nil {
		t.Fatal("Package symbol should be created")
	}

	if symbol.Kind != SymbolPackage {
		t.Errorf("Expected SymbolPackage, got %v", symbol.Kind)
	}
}

func TestBinder_ScopeNesting(t *testing.T) {
	binder := NewBinder()

	// Create variable declaration for block
	variable := &ast.VariableExpr{
		BaseNode: ast.NewBaseNode("variable", ast.Position{Line: 2, Column: 5}, ast.Position{Line: 2, Column: 15}),
		Name:     "$outer_var",
		Sigil:    "$",
	}

	varDecl := ast.NewVarDecl(
		"my",
		[]*ast.VariableExpr{variable},
		nil,
		nil,
		ast.Position{Line: 2, Column: 1},
		ast.Position{Line: 2, Column: 20},
	)

	// Create nested block structure
	blockStmt := ast.NewBlockStmt(
		[]ast.StatementNode{varDecl},
		ast.Position{Line: 1, Column: 1},
		ast.Position{Line: 3, Column: 1},
	)

	st, err := binder.Bind(blockStmt)
	if err != nil {
		t.Fatalf("Failed to bind block statement: %v", err)
	}

	// Verify block scope was created
	blockScope := st.FindScope(blockStmt)
	if blockScope == nil {
		t.Fatal("Block scope should be created")
	}

	if blockScope.Kind != ScopeBlock {
		t.Errorf("Expected ScopeBlock, got %v", blockScope.Kind)
	}

	if blockScope.Parent != st.GlobalScope {
		t.Error("Block scope parent should be global scope")
	}

	// Verify variable is in block scope
	symbol := st.ResolveInScope(blockScope, "outer_var", SymbolScalar)
	if symbol == nil {
		t.Fatal("Variable should be in block scope")
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
