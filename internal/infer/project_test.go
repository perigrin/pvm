// ABOUTME: Tests for the cross-file analysis index (ProjectIndex).
// ABOUTME: Covers ResolveModule, AnalyzeFile, LookupSymbol, and Prefetch.

package infer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/types"
)

func TestResolveModule(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "lib", "Foo"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "lib", "Foo", "Bar.pm"), []byte("package Foo::Bar;\n"), 0644))
	idx := infer.NewProjectIndex(dir)
	path, err := idx.ResolveModule("Foo::Bar")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "lib", "Foo", "Bar.pm"), path)
}

func TestResolveModuleNotFound(t *testing.T) {
	dir := t.TempDir()
	idx := infer.NewProjectIndex(dir)
	_, err := idx.ResolveModule("Nonexistent::Module")
	assert.Error(t, err)
}

func TestResolveModuleTopLevel(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "lib"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "lib", "Foo.pm"), []byte("package Foo;\n"), 0644))
	idx := infer.NewProjectIndex(dir)
	path, err := idx.ResolveModule("Foo")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "lib", "Foo.pm"), path)
}

func TestAnalyzeFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "lib", "Foo"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "lib", "Foo", "Bar.pm"),
		[]byte("package Foo::Bar;\nsub baz { return 42; }\n"),
		0644,
	))
	idx := infer.NewProjectIndex(dir)
	result, err := idx.AnalyzeFile(filepath.Join(dir, "lib", "Foo", "Bar.pm"))
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Symbols)
	sym, ok := result.Symbols.Lookup("baz")
	assert.True(t, ok, "baz should be in symbol table")
	assert.Equal(t, types.Int, sym.ReturnType, "baz returns Int")
}

func TestAnalyzeFileCached(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "lib"), 0755))
	path := filepath.Join(dir, "lib", "Simple.pm")
	require.NoError(t, os.WriteFile(path, []byte("package Simple;\n"), 0644))
	idx := infer.NewProjectIndex(dir)
	r1, err1 := idx.AnalyzeFile(path)
	require.NoError(t, err1)
	r2, err2 := idx.AnalyzeFile(path)
	require.NoError(t, err2)
	assert.True(t, r1 == r2, "second call should return cached pointer")
}

func TestLookupSymbol(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "lib", "Foo"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "lib", "Foo", "Bar.pm"),
		[]byte("package Foo::Bar;\nsub baz { return 42; }\n"),
		0644,
	))
	idx := infer.NewProjectIndex(dir)
	sym, ok := idx.LookupSymbol("Foo::Bar", "baz")
	assert.True(t, ok, "should find baz")
	assert.Equal(t, types.Int, sym.ReturnType)
}

func TestLookupSymbolMissing(t *testing.T) {
	dir := t.TempDir()
	idx := infer.NewProjectIndex(dir)
	_, ok := idx.LookupSymbol("Nonexistent", "foo")
	assert.False(t, ok)
}

func TestPrefetch(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "lib", "A"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "lib", "A", "B.pm"),
		[]byte("package A::B;\nsub x { return 1; }\n"),
		0644,
	))
	idx := infer.NewProjectIndex(dir)
	idx.Prefetch()
	sym, ok := idx.LookupSymbol("A::B", "x")
	assert.True(t, ok)
	assert.Equal(t, types.Int, sym.ReturnType)
}
