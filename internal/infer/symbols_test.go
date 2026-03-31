// ABOUTME: Tests for the symbol table used by the PSC type inference engine.
// ABOUTME: Covers Symbol, Scope, SymbolTable, and CollectDeclarations.

package infer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/types"
)

// --- Unit tests: Symbol, Scope, SymbolTable (no parser required) ---

func TestSymbolCreation(t *testing.T) {
	sym := infer.Symbol{
		Name:      "$x",
		Type:      types.Scalar,
		Kind:      infer.SymVariable,
		StartByte: 0,
		EndByte:   5,
	}

	assert.Equal(t, "$x", sym.Name)
	assert.Equal(t, types.Scalar, sym.Type)
	assert.Equal(t, infer.SymVariable, sym.Kind)
	assert.Equal(t, uint32(0), sym.StartByte)
	assert.Equal(t, uint32(5), sym.EndByte)
}

func TestScopeCreation(t *testing.T) {
	st := infer.NewSymbolTable()
	require.NotNil(t, st)
	assert.Equal(t, "main", st.CurrentPackage())
}

func TestSymbolTableDefine(t *testing.T) {
	st := infer.NewSymbolTable()

	sym := infer.Symbol{
		Name: "$x",
		Type: types.Scalar,
		Kind: infer.SymVariable,
	}
	st.Define(sym)

	found, ok := st.Lookup("$x")
	require.True(t, ok)
	assert.Equal(t, "$x", found.Name)
	assert.Equal(t, types.Scalar, found.Type)
}

func TestSymbolTableShadowing(t *testing.T) {
	st := infer.NewSymbolTable()

	outer := infer.Symbol{
		Name: "$x",
		Type: types.Scalar,
		Kind: infer.SymVariable,
	}
	st.Define(outer)

	st.EnterScope("block")

	inner := infer.Symbol{
		Name: "$x",
		Type: types.Int,
		Kind: infer.SymVariable,
	}
	st.Define(inner)

	// Inner scope should see the inner $x (Int), not the outer (Scalar)
	found, ok := st.Lookup("$x")
	require.True(t, ok)
	assert.Equal(t, types.Int, found.Type, "inner scope should shadow outer")
}

func TestSymbolTableLookupMiss(t *testing.T) {
	st := infer.NewSymbolTable()

	_, ok := st.Lookup("$notdefined")
	assert.False(t, ok)
}

func TestSymbolTableExitScope(t *testing.T) {
	st := infer.NewSymbolTable()

	outer := infer.Symbol{
		Name: "$outer",
		Type: types.Scalar,
		Kind: infer.SymVariable,
	}
	st.Define(outer)

	st.EnterScope("block")

	inner := infer.Symbol{
		Name: "$inner",
		Type: types.Int,
		Kind: infer.SymVariable,
	}
	st.Define(inner)

	st.ExitScope()

	// After exit, outer $outer is still visible
	found, ok := st.Lookup("$outer")
	require.True(t, ok)
	assert.Equal(t, types.Scalar, found.Type)

	// After exit, inner $inner is no longer visible
	_, ok = st.Lookup("$inner")
	assert.False(t, ok)
}

// --- UpdateType tests ---

func TestSymbolTableUpdateType(t *testing.T) {
	st := infer.NewSymbolTable()

	st.Define(infer.Symbol{
		Name: "$x",
		Type: types.Scalar,
		Kind: infer.SymVariable,
	})

	// Update $x from Scalar to Int
	ok := st.UpdateType("$x", types.Int)
	require.True(t, ok, "UpdateType should return true for a defined symbol")

	found, lok := st.Lookup("$x")
	require.True(t, lok)
	assert.Equal(t, types.Int, found.Type, "$x should now have type Int after UpdateType")
}

func TestSymbolTableUpdateTypeUnknown(t *testing.T) {
	st := infer.NewSymbolTable()

	// Updating a symbol that does not exist should return false
	ok := st.UpdateType("$notdefined", types.Int)
	assert.False(t, ok, "UpdateType should return false for an undefined symbol")
}

func TestSymbolTableUpdateTypeInnerScope(t *testing.T) {
	st := infer.NewSymbolTable()

	// Define $x in the outer (root) scope
	st.Define(infer.Symbol{
		Name: "$x",
		Type: types.Scalar,
		Kind: infer.SymVariable,
	})

	// Enter an inner scope
	st.EnterScope("block")

	// Update $x from within the inner scope — should walk up and find it
	ok := st.UpdateType("$x", types.Num)
	require.True(t, ok, "UpdateType should find $x in the outer scope")

	// Verify from the inner scope
	found, lok := st.Lookup("$x")
	require.True(t, lok)
	assert.Equal(t, types.Num, found.Type, "$x should be Num after UpdateType from inner scope")

	// Exit back to outer scope and verify there too
	st.ExitScope()
	found, lok = st.Lookup("$x")
	require.True(t, lok)
	assert.Equal(t, types.Num, found.Type, "$x should still be Num in the outer scope")
}

// --- Integration tests: CollectDeclarations (uses parser) ---

func TestCollectVariableDeclarations(t *testing.T) {
	p := parser.New()
	src := []byte("my $x = 1; my @arr; my %hash;")
	tree, err := p.Parse(src)
	require.NoError(t, err)

	st := infer.CollectDeclarations(tree, src)
	require.NotNil(t, st)

	scalar, ok := st.Lookup("$x")
	require.True(t, ok, "expected $x to be defined")
	assert.Equal(t, types.Scalar, scalar.Type)
	assert.Equal(t, infer.SymVariable, scalar.Kind)

	arr, ok := st.Lookup("@arr")
	require.True(t, ok, "expected @arr to be defined")
	assert.Equal(t, types.Array, arr.Type)
	assert.Equal(t, infer.SymVariable, arr.Kind)

	hash, ok := st.Lookup("%hash")
	require.True(t, ok, "expected %hash to be defined")
	assert.Equal(t, types.Hash, hash.Type)
	assert.Equal(t, infer.SymVariable, hash.Kind)
}

func TestCollectSubroutineDeclarations(t *testing.T) {
	p := parser.New()
	src := []byte("sub greet { my $x = 1; return $x; }")
	tree, err := p.Parse(src)
	require.NoError(t, err)

	st := infer.CollectDeclarations(tree, src)
	require.NotNil(t, st)

	sub, ok := st.Lookup("greet")
	require.True(t, ok, "expected sub greet to be defined")
	assert.Equal(t, infer.SymSubroutine, sub.Kind)
	assert.Equal(t, types.Code, sub.Type)
}

func TestCollectBlockScoping(t *testing.T) {
	p := parser.New()
	// $outer is in the top-level scope; $inner is only in the if-block scope.
	src := []byte("my $outer = 1; if ($outer) { my $inner = 2; }")
	tree, err := p.Parse(src)
	require.NoError(t, err)

	st := infer.CollectDeclarations(tree, src)
	require.NotNil(t, st)

	// $outer should be visible at the top-level scope after collection
	outer, ok := st.Lookup("$outer")
	require.True(t, ok, "expected $outer to be visible")
	assert.Equal(t, types.Scalar, outer.Type)

	// $inner should NOT be visible at top-level; it was scoped to the if-block
	_, ok = st.Lookup("$inner")
	assert.False(t, ok, "$inner should not be visible in the top-level scope after collection")
}

func TestSymbolReturnType(t *testing.T) {
	st := infer.NewSymbolTable()
	st.Define(infer.Symbol{
		Name:       "foo",
		Type:       types.Code,
		Kind:       infer.SymSubroutine,
		ReturnType: types.Int,
	})
	sym, ok := st.Lookup("foo")
	require.True(t, ok)
	assert.Equal(t, types.Int, sym.ReturnType, "ReturnType should be preserved")
}

func TestUpdateReturnType(t *testing.T) {
	st := infer.NewSymbolTable()
	st.Define(infer.Symbol{
		Name: "bar",
		Type: types.Code,
		Kind: infer.SymSubroutine,
	})
	ok := st.UpdateReturnType("bar", types.Num)
	assert.True(t, ok, "should find and update bar")
	sym, found := st.Lookup("bar")
	require.True(t, found)
	assert.Equal(t, types.Num, sym.ReturnType)

	ok = st.UpdateReturnType("nonexistent", types.Int)
	assert.False(t, ok, "nonexistent should return false")
}

func TestLookupAll(t *testing.T) {
	st := infer.NewSymbolTable()

	// Define one symbol in the root scope.
	st.Define(infer.Symbol{
		Name: "$root",
		Type: types.Scalar,
		Kind: infer.SymVariable,
	})

	// Enter a child scope and define a symbol there.
	st.EnterScope("block")
	st.Define(infer.Symbol{
		Name: "$child",
		Type: types.Int,
		Kind: infer.SymVariable,
	})
	st.ExitScope()

	// After ExitScope, current is back at root. Lookup cannot see $child.
	_, ok := st.Lookup("$child")
	assert.False(t, ok, "Lookup should not find $child after exiting its scope")

	// LookupAll must find $child by searching all scopes.
	found, ok := st.LookupAll("$child")
	require.True(t, ok, "LookupAll should find $child in a child scope")
	assert.Equal(t, types.Int, found.Type)

	// LookupAll must also find symbols in the root scope.
	foundRoot, ok := st.LookupAll("$root")
	require.True(t, ok, "LookupAll should find $root in the root scope")
	assert.Equal(t, types.Scalar, foundRoot.Type)

	// LookupAll must return false for names that do not exist anywhere.
	_, ok = st.LookupAll("$nowhere")
	assert.False(t, ok, "LookupAll should return false for an undefined symbol")
}

func TestAllSymbols(t *testing.T) {
	st := infer.NewSymbolTable()

	// Define one symbol in the root scope
	st.Define(infer.Symbol{
		Name: "$root",
		Type: types.Scalar,
		Kind: infer.SymVariable,
	})

	// Enter a child scope and define two more symbols
	st.EnterScope("block")
	st.Define(infer.Symbol{
		Name: "$child1",
		Type: types.Int,
		Kind: infer.SymVariable,
	})
	st.Define(infer.Symbol{
		Name: "$child2",
		Type: types.Num,
		Kind: infer.SymVariable,
	})
	st.ExitScope()

	all := st.AllSymbols()
	assert.Len(t, all, 3, "AllSymbols should return all symbols across all scopes")

	names := make(map[string]bool, len(all))
	for _, sym := range all {
		names[sym.Name] = true
	}
	assert.True(t, names["$root"], "AllSymbols should include $root from the root scope")
	assert.True(t, names["$child1"], "AllSymbols should include $child1 from the child scope")
	assert.True(t, names["$child2"], "AllSymbols should include $child2 from the child scope")
}
