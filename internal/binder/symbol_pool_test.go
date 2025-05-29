// ABOUTME: Tests for symbol pooling implementation
// ABOUTME: Ensures symbol pool manager correctly manages Symbol, Scope, and SymbolTable objects

package binder

import (
	"sync"
	"testing"

	"tamarou.com/pvm/internal/ast"
)

func TestSymbolPoolManager_NewSymbol(t *testing.T) {
	manager := NewSymbolPoolManager(SymbolPoolHooks{})

	// Create symbol
	pos := ast.Position{Line: 1, Column: 1, Offset: 0}
	symbol := manager.NewSymbol("test_var", SymbolScalar, SymbolFlagLexical, nil, pos)

	if symbol == nil {
		t.Fatal("Expected non-nil symbol")
	}
	if symbol.Name != "test_var" {
		t.Errorf("Expected name 'test_var', got '%s'", symbol.Name)
	}
	if symbol.Kind != SymbolScalar {
		t.Errorf("Expected SymbolScalar kind, got %d", symbol.Kind)
	}
	if symbol.Flags != SymbolFlagLexical {
		t.Errorf("Expected SymbolFlagLexical flags, got %d", symbol.Flags)
	}
	if symbol.Position != pos {
		t.Errorf("Expected position %+v, got %+v", pos, symbol.Position)
	}

	// Check manager statistics
	if manager.SymbolCount() != 1 {
		t.Errorf("Expected symbol count 1, got %d", manager.SymbolCount())
	}
}

func TestSymbolPoolManager_NewScope(t *testing.T) {
	manager := NewSymbolPoolManager(SymbolPoolHooks{})

	// Create parent scope
	pos := ast.Position{Line: 1, Column: 1, Offset: 0}
	parentScope := manager.NewScope(ScopeGlobal, nil, nil, pos)

	if parentScope == nil {
		t.Fatal("Expected non-nil parent scope")
	}
	if parentScope.Kind != ScopeGlobal {
		t.Errorf("Expected ScopeGlobal kind, got %d", parentScope.Kind)
	}
	if parentScope.Parent != nil {
		t.Error("Expected nil parent for global scope")
	}

	// Create child scope
	childPos := ast.Position{Line: 2, Column: 1, Offset: 10}
	childScope := manager.NewScope(ScopeSubroutine, parentScope, nil, childPos)

	if childScope == nil {
		t.Fatal("Expected non-nil child scope")
	}
	if childScope.Kind != ScopeSubroutine {
		t.Errorf("Expected ScopeSubroutine kind, got %d", childScope.Kind)
	}
	if childScope.Parent != parentScope {
		t.Error("Expected parent scope to match")
	}

	// Check parent's children
	if len(parentScope.Children) != 1 {
		t.Errorf("Expected 1 child scope, got %d", len(parentScope.Children))
	}
	if parentScope.Children[0] != childScope {
		t.Error("Expected child scope to be added to parent")
	}

	// Check manager statistics
	if manager.ScopeCount() != 2 {
		t.Errorf("Expected scope count 2, got %d", manager.ScopeCount())
	}
}

func TestSymbolPoolManager_NewSymbolTable(t *testing.T) {
	manager := NewSymbolPoolManager(SymbolPoolHooks{})

	// Create symbol table
	table := manager.NewSymbolTable("Test::Package")

	if table == nil {
		t.Fatal("Expected non-nil symbol table")
	}
	if table.Package != "Test::Package" {
		t.Errorf("Expected package 'Test::Package', got '%s'", table.Package)
	}
	if table.Scopes == nil {
		t.Error("Expected initialized Scopes map")
	}
	if table.Symbols == nil {
		t.Error("Expected initialized Symbols map")
	}
	if table.ModuleSymbols == nil {
		t.Error("Expected initialized ModuleSymbols map")
	}
}

func TestSymbolPoolManager_PoolingEfficiency(t *testing.T) {
	manager := NewSymbolPoolManager(SymbolPoolHooks{})

	// Create many symbols to test pooling
	const numSymbols = 100
	pos := ast.Position{Line: 1, Column: 1, Offset: 0}

	for i := 0; i < numSymbols; i++ {
		symbol := manager.NewSymbol("test", SymbolScalar, SymbolFlagNone, nil, pos)
		if symbol == nil {
			t.Fatalf("Failed to create symbol %d", i)
		}
	}

	if manager.SymbolCount() != numSymbols {
		t.Errorf("Expected %d symbols created, got %d", numSymbols, manager.SymbolCount())
	}

	// Test efficiency calculation
	efficiency := manager.PoolEfficiency()
	if efficiency < 0 || efficiency > 100 {
		t.Errorf("Expected efficiency between 0-100%%, got %f", efficiency)
	}
}

func TestSymbolPoolManager_Hooks(t *testing.T) {
	var createdSymbols []*Symbol
	var createdScopes []*Scope
	var resetSymbols []*Symbol
	var resetScopes []*Scope

	hooks := SymbolPoolHooks{
		OnSymbolCreate: func(symbol *Symbol) {
			createdSymbols = append(createdSymbols, symbol)
		},
		OnScopeCreate: func(scope *Scope) {
			createdScopes = append(createdScopes, scope)
		},
		OnSymbolReset: func(symbol *Symbol) {
			resetSymbols = append(resetSymbols, symbol)
		},
		OnScopeReset: func(scope *Scope) {
			resetScopes = append(resetScopes, scope)
		},
	}

	manager := NewSymbolPoolManager(hooks)

	// Create a symbol and scope
	pos := ast.Position{Line: 1, Column: 1, Offset: 0}
	symbol := manager.NewSymbol("test", SymbolScalar, SymbolFlagNone, nil, pos)
	scope := manager.NewScope(ScopeGlobal, nil, nil, pos)

	if len(createdSymbols) != 1 {
		t.Errorf("Expected 1 created symbol, got %d", len(createdSymbols))
	}
	if createdSymbols[0] != symbol {
		t.Error("Expected created symbol to match")
	}

	if len(createdScopes) != 1 {
		t.Errorf("Expected 1 created scope, got %d", len(createdScopes))
	}
	if createdScopes[0] != scope {
		t.Error("Expected created scope to match")
	}

	// Reset manager to trigger reset hooks
	manager.Reset()

	// Create another symbol and scope to trigger reuse and reset
	manager.NewSymbol("test2", SymbolScalar, SymbolFlagNone, nil, pos)
	manager.NewScope(ScopeBlock, nil, nil, pos)

	// Should have triggered reset hooks when reusing objects
	if len(resetSymbols) == 0 {
		t.Error("Expected reset hook to be called during symbol reuse")
	}
	if len(resetScopes) == 0 {
		t.Error("Expected reset hook to be called during scope reuse")
	}
}

func TestSymbolPoolManager_ConcurrentAccess(t *testing.T) {
	manager := NewSymbolPoolManager(SymbolPoolHooks{})
	const numGoroutines = 10
	const symbolsPerGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	pos := ast.Position{Line: 1, Column: 1, Offset: 0}

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < symbolsPerGoroutine; j++ {
				symbol := manager.NewSymbol("concurrent", SymbolScalar, SymbolFlagNone, nil, pos)
				if symbol == nil {
					t.Errorf("Goroutine %d failed to create symbol %d", id, j)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	expectedSymbols := int64(numGoroutines * symbolsPerGoroutine)
	if manager.SymbolCount() != expectedSymbols {
		t.Errorf("Expected %d symbols from concurrent creation, got %d", expectedSymbols, manager.SymbolCount())
	}
}

func TestSymbolPoolManager_MapAndSliceAllocation(t *testing.T) {
	manager := NewSymbolPoolManager(SymbolPoolHooks{})

	// Test symbol map allocation
	symbolMap := manager.NewSymbolMap(10)
	if len(symbolMap) != 0 {
		t.Errorf("Expected empty symbol map, got length %d", len(symbolMap))
	}

	// Test symbol slice allocation
	symbolSlice := manager.NewSymbolSlice(10)
	if cap(symbolSlice) < 10 {
		t.Errorf("Expected symbol slice capacity >= 10, got %d", cap(symbolSlice))
	}

	// Test scope slice allocation
	scopeSlice := manager.NewScopeSlice(5)
	if cap(scopeSlice) < 5 {
		t.Errorf("Expected scope slice capacity >= 5, got %d", cap(scopeSlice))
	}

	// Test module map allocation
	moduleMap := manager.NewModuleMap(8)
	if len(moduleMap) != 0 {
		t.Errorf("Expected empty module map, got length %d", len(moduleMap))
	}

	// Test string map allocation
	stringMap := manager.NewStringMap(4)
	if len(stringMap) != 0 {
		t.Errorf("Expected empty string map, got length %d", len(stringMap))
	}

	// Test node map allocation
	nodeMap := manager.NewNodeMap(16)
	if len(nodeMap) != 0 {
		t.Errorf("Expected empty node map, got length %d", len(nodeMap))
	}
}

func TestSymbolPoolManager_WarmPools(t *testing.T) {
	var warmedPools []string

	hooks := SymbolPoolHooks{
		OnPoolWarming: func(poolType string) {
			warmedPools = append(warmedPools, poolType)
		},
	}

	manager := NewSymbolPoolManager(hooks)

	// Warm pools
	manager.WarmPools()

	// Check that warming hooks were called
	if len(warmedPools) == 0 {
		t.Error("Expected pool warming hooks to be called")
	}

	// Should have warmed both symbols and scopes
	foundSymbols := false
	foundScopes := false
	for _, poolType := range warmedPools {
		if poolType == "symbols" {
			foundSymbols = true
		}
		if poolType == "scopes" {
			foundScopes = true
		}
	}

	if !foundSymbols {
		t.Error("Expected symbols pool to be warmed")
	}
	if !foundScopes {
		t.Error("Expected scopes pool to be warmed")
	}
}

func TestSymbolPoolManager_ResetAndClear(t *testing.T) {
	manager := NewSymbolPoolManager(SymbolPoolHooks{})

	// Create some symbols and scopes
	pos := ast.Position{Line: 1, Column: 1, Offset: 0}
	manager.NewSymbol("test1", SymbolScalar, SymbolFlagNone, nil, pos)
	manager.NewSymbol("test2", SymbolArray, SymbolFlagNone, nil, pos)
	manager.NewScope(ScopeGlobal, nil, nil, pos)

	initialSymbolCount := manager.SymbolCount()
	initialScopeCount := manager.ScopeCount()

	if initialSymbolCount != 2 {
		t.Errorf("Expected 2 symbols initially, got %d", initialSymbolCount)
	}
	if initialScopeCount != 1 {
		t.Errorf("Expected 1 scope initially, got %d", initialScopeCount)
	}

	// Reset should keep statistics but clear pools for reuse
	manager.Reset()
	if manager.SymbolCount() != initialSymbolCount {
		t.Error("Reset should not change symbol count")
	}
	if manager.ScopeCount() != initialScopeCount {
		t.Error("Reset should not change scope count")
	}

	// Clear should reset everything
	manager.Clear()
	if manager.SymbolCount() != 0 {
		t.Errorf("Expected 0 symbols after clear, got %d", manager.SymbolCount())
	}
	if manager.ScopeCount() != 0 {
		t.Errorf("Expected 0 scopes after clear, got %d", manager.ScopeCount())
	}
}

func TestDefaultSymbolPoolManager(t *testing.T) {
	manager := DefaultSymbolPoolManager()
	if manager == nil {
		t.Fatal("Expected non-nil default symbol pool manager")
	}

	// Test that we can create symbols with default manager
	pos := ast.Position{Line: 1, Column: 1, Offset: 0}
	symbol := manager.NewSymbol("default", SymbolScalar, SymbolFlagNone, nil, pos)
	if symbol == nil {
		t.Fatal("Expected non-nil symbol from default manager")
	}

	// Test setting custom default manager
	customManager := NewSymbolPoolManager(SymbolPoolHooks{})
	SetDefaultSymbolPoolManager(customManager)

	if DefaultSymbolPoolManager() != customManager {
		t.Error("Expected default manager to be updated")
	}
}

func TestSymbolPoolManager_SymbolFieldsReset(t *testing.T) {
	manager := NewSymbolPoolManager(SymbolPoolHooks{})

	// Create a symbol with various fields set
	pos := ast.Position{Line: 1, Column: 1, Offset: 0}
	symbol := manager.NewSymbol("test", SymbolScalar, SymbolFlagLexical, nil, pos)

	// Set some fields that should be reset
	symbol.Type = "Int"
	symbol.Package = "Test::Package"
	symbol.QualifiedName = "Test::Package::test"
	symbol.CapturedBy = []*Scope{{}}
	symbol.Upvalues = []*Symbol{{}}

	// Reset the symbol
	manager.resetSymbol(symbol)

	// Check that fields were properly reset
	if symbol.Type != "Int" { // Type should be preserved
		t.Error("Type should be preserved during reset")
	}
	if len(symbol.CapturedBy) != 0 {
		t.Error("CapturedBy slice should be empty after reset")
	}
	if len(symbol.Upvalues) != 0 {
		t.Error("Upvalues slice should be empty after reset")
	}
	if symbol.OriginalSymbol != nil {
		t.Error("OriginalSymbol should be nil after reset")
	}
}

func TestSymbolPoolManager_ScopeFieldsReset(t *testing.T) {
	manager := NewSymbolPoolManager(SymbolPoolHooks{})

	// Create a scope with various fields set
	pos := ast.Position{Line: 1, Column: 1, Offset: 0}
	scope := manager.NewScope(ScopeSubroutine, nil, nil, pos)

	// Add some symbols and children
	scope.Symbols["test"] = &Symbol{Name: "test"}
	scope.LocalSymbols["local"] = &Symbol{Name: "local"}
	scope.ImportedModules["Module"] = "path"
	childScope := manager.NewScope(ScopeBlock, scope, nil, pos)
	scope.CapturedSymbols = append(scope.CapturedSymbols, &Symbol{Name: "captured"})

	// Reset the scope
	manager.resetScope(scope)

	// Check that maps and slices were properly reset
	if len(scope.Symbols) != 0 {
		t.Error("Symbols map should be empty after reset")
	}
	if len(scope.LocalSymbols) != 0 {
		t.Error("LocalSymbols map should be empty after reset")
	}
	if len(scope.ImportedModules) != 0 {
		t.Error("ImportedModules map should be empty after reset")
	}
	if len(scope.Children) != 0 {
		t.Error("Children slice should be empty after reset")
	}
	if len(scope.CapturedSymbols) != 0 {
		t.Error("CapturedSymbols slice should be empty after reset")
	}
	if scope.Parent != nil {
		t.Error("Parent should be nil after reset")
	}

	// But the child scope should still exist (it's managed independently)
	if childScope == nil {
		t.Error("Child scope should still exist")
	}
}

func BenchmarkSymbolPoolManager_NewSymbol(b *testing.B) {
	manager := NewSymbolPoolManager(SymbolPoolHooks{})
	pos := ast.Position{Line: 1, Column: 1, Offset: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		symbol := manager.NewSymbol("benchmark", SymbolScalar, SymbolFlagNone, nil, pos)
		if symbol == nil {
			b.Fatal("Failed to create symbol")
		}
	}
}

func BenchmarkSymbolPoolManager_NewScope(b *testing.B) {
	manager := NewSymbolPoolManager(SymbolPoolHooks{})
	pos := ast.Position{Line: 1, Column: 1, Offset: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scope := manager.NewScope(ScopeSubroutine, nil, nil, pos)
		if scope == nil {
			b.Fatal("Failed to create scope")
		}
	}
}

func BenchmarkSymbolPoolManager_Concurrent(b *testing.B) {
	manager := NewSymbolPoolManager(SymbolPoolHooks{})
	pos := ast.Position{Line: 1, Column: 1, Offset: 0}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			symbol := manager.NewSymbol("concurrent", SymbolScalar, SymbolFlagNone, nil, pos)
			if symbol == nil {
				b.Fatal("Failed to create symbol concurrently")
			}
		}
	})
}
