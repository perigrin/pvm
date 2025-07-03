// ABOUTME: Tests for advanced symbol binding features including closures, dynamic scoping, and modules.
// ABOUTME: Covers complex Perl scoping scenarios and edge cases for production readiness.

package binder

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
)

func TestAdvancedBinder_PackageQualifiedSymbols(t *testing.T) {
	st := NewSymbolTable()

	// Create a package-qualified symbol
	symbol := st.CreatePackageQualifiedSymbol("MyPackage", "test_var", SymbolScalar, SymbolFlagPackage)

	if symbol == nil {
		t.Fatal("Package-qualified symbol should be created")
	}

	if symbol.QualifiedName != "MyPackage::test_var" {
		t.Errorf("Expected qualified name 'MyPackage::test_var', got '%s'", symbol.QualifiedName)
	}

	if symbol.Flags&SymbolFlagPackageQualified == 0 {
		t.Error("Symbol should have package qualified flag")
	}

	// Test resolution
	resolved := st.ResolvePackageSymbol("MyPackage", "test_var", SymbolScalar)
	if resolved != symbol {
		t.Error("Should resolve package-qualified symbol")
	}

	// Test with qualification
	resolvedQualified := st.ResolveWithPackageQualification("MyPackage::test_var", SymbolScalar)
	if resolvedQualified != symbol {
		t.Error("Should resolve with package qualification")
	}
}

func TestAdvancedBinder_ModuleImports(t *testing.T) {
	st := NewSymbolTable()

	// Create a module symbol table
	moduleTable := NewSymbolTable()
	moduleTable.Package = "TestModule"

	// Add a symbol to the module
	moduleSymbol := &Symbol{
		Name:     "module_function",
		Kind:     SymbolSubroutine,
		Flags:    SymbolFlagExported,
		Position: ast.Position{Line: 1, Column: 1},
	}
	moduleTable.AddSymbol(moduleSymbol)
	moduleTable.ExportSymbol(moduleSymbol)

	// Import the module
	err := st.ImportModule("TestModule", moduleTable)
	if err != nil {
		t.Fatalf("Failed to import module: %v", err)
	}

	// Test module symbol table retrieval
	retrieved := st.GetModuleSymbolTable("TestModule")
	if retrieved != moduleTable {
		t.Error("Should retrieve correct module symbol table")
	}

	// Test imported symbol resolution
	resolved := st.ResolveImportedSymbol("TestModule", "module_function", SymbolSubroutine)
	if resolved != moduleSymbol {
		t.Error("Should resolve imported symbol")
	}
}

func TestAdvancedBinder_SymbolAliases(t *testing.T) {
	st := NewSymbolTable()

	// Create original symbol
	original := &Symbol{
		Name:     "original_var",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagLexical,
		Position: ast.Position{Line: 1, Column: 1},
	}
	st.AddSymbol(original)

	// Create alias
	alias := st.CreateAlias("alias_var", original)

	if alias == nil {
		t.Fatal("Alias symbol should be created")
	}

	if alias.OriginalSymbol != original {
		t.Error("Alias should point to original symbol")
	}

	if alias.Flags&SymbolFlagAlias == 0 {
		t.Error("Symbol should have alias flag")
	}

	if alias.Kind != original.Kind {
		t.Error("Alias should have same kind as original")
	}
}

func TestAdvancedBinder_ClosureCapture(t *testing.T) {
	st := NewSymbolTable()

	// Create outer scope with a variable
	st.EnterScope(ScopeSubroutine, nil)
	outerSymbol := &Symbol{
		Name:     "outer_var",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagLexical,
		Position: ast.Position{Line: 1, Column: 1},
	}
	st.AddSymbol(outerSymbol)

	// Create inner scope (closure)
	innerScope := st.EnterScope(ScopeSubroutine, nil)

	// Capture the outer symbol
	err := st.CaptureSymbol(outerSymbol, innerScope)
	if err != nil {
		t.Fatalf("Failed to capture symbol: %v", err)
	}

	// Verify capture flags
	if outerSymbol.Flags&SymbolFlagClosure == 0 {
		t.Error("Symbol should have closure flag")
	}

	if outerSymbol.Flags&SymbolFlagUpvalue == 0 {
		t.Error("Symbol should have upvalue flag")
	}

	// Verify capture relationships
	if len(outerSymbol.CapturedBy) != 1 || outerSymbol.CapturedBy[0] != innerScope {
		t.Error("Symbol should be captured by inner scope")
	}

	if len(innerScope.CapturedSymbols) != 1 || innerScope.CapturedSymbols[0] != outerSymbol {
		t.Error("Inner scope should capture outer symbol")
	}
}

func TestAdvancedBinder_LocalVariables(t *testing.T) {
	st := NewSymbolTable()

	// Create global variable
	globalSymbol := &Symbol{
		Name:     "test_var",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagPackage,
		Position: ast.Position{Line: 1, Column: 1},
	}
	st.AddSymbol(globalSymbol)

	// Enter block scope
	blockScope := st.EnterScope(ScopeBlock, nil)

	// Create local variable
	localSymbol := &Symbol{
		Name:     "test_var",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagLexical,
		Position: ast.Position{Line: 2, Column: 1},
	}

	err := st.CreateLocalSymbol(localSymbol)
	if err != nil {
		t.Fatalf("Failed to create local symbol: %v", err)
	}

	// Verify local flag
	if localSymbol.Flags&SymbolFlagLocal == 0 {
		t.Error("Symbol should have local flag")
	}

	// Verify saved value
	if savedSymbol, exists := blockScope.SavedValues["test_var"]; !exists || savedSymbol != globalSymbol {
		t.Error("Should save original symbol value")
	}

	// Test resolution (should find local)
	resolved := st.ResolveSymbol("test_var", SymbolScalar)
	if resolved != localSymbol {
		t.Error("Should resolve to local symbol")
	}

	// Exit scope and test restoration
	st.ExitScopeAdvanced()

	// Should now resolve to global again
	resolved = st.ResolveSymbol("test_var", SymbolScalar)
	if resolved != globalSymbol {
		t.Error("Should resolve to global symbol after scope exit")
	}
}

func TestAdvancedBinder_DynamicSymbols(t *testing.T) {
	st := NewSymbolTable()

	// Create dynamic symbol
	dynamicSymbol := st.CreateDynamicSymbol("dynamic_var", SymbolScalar, SymbolFlagLexical)

	if dynamicSymbol == nil {
		t.Fatal("Dynamic symbol should be created")
	}

	if dynamicSymbol.Flags&SymbolFlagDynamic == 0 {
		t.Error("Symbol should have dynamic flag")
	}

	// Verify it's in dynamic symbols map
	if stored := st.DynamicSymbols["dynamic_var"]; stored != dynamicSymbol {
		t.Error("Symbol should be stored in dynamic symbols map")
	}
}

func TestAdvancedBinder_ClosureAnalysis(t *testing.T) {
	st := NewSymbolTable()

	// Create outer scope with variables
	outerScope := st.EnterScope(ScopeSubroutine, nil)

	var1 := &Symbol{
		Name:     "var1",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagLexical,
		Position: ast.Position{Line: 1, Column: 1},
		Scope:    outerScope,
	}
	st.AddSymbol(var1)

	var2 := &Symbol{
		Name:     "var2",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagLexical,
		Position: ast.Position{Line: 2, Column: 1},
		Scope:    outerScope,
	}
	st.AddSymbol(var2)

	// Create inner closure scope
	closureScope := st.EnterScope(ScopeSubroutine, nil)

	// Simulate references to outer variables in closure
	closureScope.Symbols["var1"] = var1
	closureScope.Symbols["var2"] = var2

	// Analyze closure capture
	capturedSymbols := st.AnalyzeClosureCapture(closureScope)

	if len(capturedSymbols) != 2 {
		t.Errorf("Expected 2 captured symbols, got %d", len(capturedSymbols))
	}

	// Verify symbols are marked as captured
	if var1.Flags&SymbolFlagClosure == 0 {
		t.Error("var1 should be marked as captured")
	}

	if var2.Flags&SymbolFlagClosure == 0 {
		t.Error("var2 should be marked as captured")
	}
}

func TestAdvancedBinder_UseStatement(t *testing.T) {
	// TODO: Convert this test to use CST-based binding approach like other tests
	t.Skip("Temporarily disabled - needs conversion to CST-based binding")

	_ = NewBinder() // Suppress unused warning

	// Create use statement
	_ = ast.NewUseStmt(
		"Test::Module",
		"1.0",
		[]string{},
		ast.Position{Line: 1, Column: 1},
		ast.Position{Line: 1, Column: 20},
	)

	st := NewSymbolTable() // Placeholder for now
	err := error(nil)
	if err != nil {
		t.Fatalf("Failed to bind use statement: %v", err)
	}

	// Verify import symbol was created
	symbol := st.ResolveSymbol("Test::Module", SymbolImport)
	if symbol == nil {
		t.Fatal("Import symbol should be created")
	}

	if symbol.Kind != SymbolImport {
		t.Errorf("Expected SymbolImport, got %v", symbol.Kind)
	}

	if symbol.Flags&SymbolFlagImported == 0 {
		t.Error("Symbol should have imported flag")
	}

	// Verify module was registered
	moduleTable := st.GetModuleSymbolTable("Test::Module")
	if moduleTable == nil {
		t.Error("Module symbol table should be registered")
	}
}

func TestAdvancedBinder_ComplexScoping(t *testing.T) {
	st := NewSymbolTable()

	// Create complex nested structure with multiple scoping scenarios

	// Package level
	st.EnterScope(ScopePackage, nil)
	packageVar := &Symbol{
		Name:     "package_var",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagPackage,
		Position: ast.Position{Line: 1, Column: 1},
	}
	st.AddSymbol(packageVar)

	// Subroutine level
	st.EnterScope(ScopeSubroutine, nil)
	subVar := &Symbol{
		Name:     "sub_var",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagLexical,
		Position: ast.Position{Line: 2, Column: 1},
	}
	st.AddSymbol(subVar)

	// Block level
	st.EnterScope(ScopeBlock, nil)
	blockVar := &Symbol{
		Name:     "block_var",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagLexical,
		Position: ast.Position{Line: 3, Column: 1},
	}
	st.AddSymbol(blockVar)

	// Test resolution from innermost scope
	if resolved := st.ResolveSymbol("block_var", SymbolScalar); resolved != blockVar {
		t.Error("Should resolve block variable from block scope")
	}

	if resolved := st.ResolveSymbol("sub_var", SymbolScalar); resolved != subVar {
		t.Error("Should resolve subroutine variable from block scope")
	}

	if resolved := st.ResolveSymbol("package_var", SymbolScalar); resolved != packageVar {
		t.Error("Should resolve package variable from block scope")
	}

	// Exit block scope
	st.ExitScopeAdvanced()

	// Should no longer resolve block variable
	if resolved := st.ResolveSymbol("block_var", SymbolScalar); resolved != nil {
		t.Error("Should not resolve block variable after exiting block scope")
	}

	// But should still resolve others
	if resolved := st.ResolveSymbol("sub_var", SymbolScalar); resolved != subVar {
		t.Error("Should still resolve subroutine variable")
	}
}

func TestAdvancedBinder_FlagCombinations(t *testing.T) {
	tests := []struct {
		name          string
		flags         SymbolFlags
		expectedParts []string
	}{
		{
			name:          "lexical with type annotation",
			flags:         SymbolFlagLexical | SymbolFlagTypeAnnotated,
			expectedParts: []string{"lexical", "typed"},
		},
		{
			name:          "package qualified closure",
			flags:         SymbolFlagPackageQualified | SymbolFlagClosure,
			expectedParts: []string{"qualified", "closure"},
		},
		{
			name:          "exported method alias",
			flags:         SymbolFlagExported | SymbolFlagMethod | SymbolFlagAlias,
			expectedParts: []string{"exported", "method", "alias"},
		},
		{
			name:          "local dynamic upvalue",
			flags:         SymbolFlagLocal | SymbolFlagDynamic | SymbolFlagUpvalue,
			expectedParts: []string{"local", "dynamic", "upvalue"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flagString := test.flags.String()
			for _, expectedPart := range test.expectedParts {
				if !strings.Contains(flagString, expectedPart) {
					t.Errorf("Flag string '%s' should contain '%s'", flagString, expectedPart)
				}
			}
		})
	}
}

func TestAdvancedBinder_ErrorCases(t *testing.T) {
	st := NewSymbolTable()

	// Test creating local symbol without current scope
	st.CurrentScope = nil
	symbol := &Symbol{
		Name:     "test_var",
		Kind:     SymbolScalar,
		Flags:    SymbolFlagLexical,
		Position: ast.Position{Line: 1, Column: 1},
	}

	err := st.CreateLocalSymbol(symbol)
	if err == nil {
		t.Error("Should fail to create local symbol without current scope")
	}

	// Restore current scope
	st.CurrentScope = st.GlobalScope

	// Test resolving non-existent package symbol
	resolved := st.ResolvePackageSymbol("NonExistent", "var", SymbolScalar)
	if resolved != nil {
		t.Error("Should not resolve non-existent package symbol")
	}

	// Test resolving non-existent imported symbol
	resolved = st.ResolveImportedSymbol("NonExistent", "var", SymbolScalar)
	if resolved != nil {
		t.Error("Should not resolve non-existent imported symbol")
	}
}
