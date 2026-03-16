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
