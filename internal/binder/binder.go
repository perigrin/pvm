// ABOUTME: Main binder implementation that works directly with CST
// ABOUTME: Handles Perl's lexical scoping rules and variable/subroutine declarations using tree-sitter CST

package binder

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/ast"
)

// DefaultBinder implements the Binder interface using CST-based binding
type DefaultBinder struct {
	poolManager *SymbolPoolManager
	cstBinder   *CSTBinder // Embedded CST binder for direct CST operations
}

// NewBinder creates a new default binder
func NewBinder() *DefaultBinder {
	poolManager := DefaultSymbolPoolManager()
	return &DefaultBinder{
		poolManager: poolManager,
		cstBinder:   NewCSTBinderWithPool(poolManager),
	}
}

// NewBinderWithPool creates a new binder with a specific pool manager
func NewBinderWithPool(poolManager *SymbolPoolManager) *DefaultBinder {
	return &DefaultBinder{
		poolManager: poolManager,
		cstBinder:   NewCSTBinderWithPool(poolManager),
	}
}

// BindCST performs symbol binding directly on tree-sitter CST
func (b *DefaultBinder) BindCST(root *sitter.Node, content []byte, typeAnnotations []*ast.TypeAnnotation) (*SymbolTable, error) {
	// Use the embedded CST binder for direct CST binding
	return b.cstBinder.BindCST(root, content, typeAnnotations)
}