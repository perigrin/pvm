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
	cstBinder   *CSTBinder   // Embedded CST binder for direct CST operations
	symbolTable *SymbolTable // Symbol table for testing access
}

// NewBinder creates a new default binder
func NewBinder() *DefaultBinder {
	poolManager := DefaultSymbolPoolManager()
	symbolTable := NewSymbolTableWithPool(poolManager, "main")
	return &DefaultBinder{
		poolManager: poolManager,
		cstBinder:   NewCSTBinderWithPool(poolManager),
		symbolTable: symbolTable,
	}
}

// NewBinderWithPool creates a new binder with a specific pool manager
func NewBinderWithPool(poolManager *SymbolPoolManager) *DefaultBinder {
	symbolTable := NewSymbolTableWithPool(poolManager, "main")
	return &DefaultBinder{
		poolManager: poolManager,
		cstBinder:   NewCSTBinderWithPool(poolManager),
		symbolTable: symbolTable,
	}
}

// BindCST performs symbol binding directly on tree-sitter CST
func (b *DefaultBinder) BindCST(root *sitter.Node, content []byte, typeAnnotations []*ast.TypeAnnotation) (*SymbolTable, error) {
	// Use the embedded CST binder for direct CST binding
	return b.cstBinder.BindCST(root, content, typeAnnotations)
}

// Note: AST-based binding methods removed in favor of CST-based binding only.
// All binding now goes through BindCST() for consistency and better integration
// with tree-sitter parsing infrastructure.

// getVariableSymbolKind determines the symbol kind for a variable
func (b *DefaultBinder) getVariableSymbolKind(name string) SymbolKind {
	if len(name) == 0 {
		return SymbolScalar
	}

	switch name[0] {
	case '$':
		return SymbolScalar
	case '@':
		return SymbolArray
	case '%':
		return SymbolHash
	case '*':
		return SymbolGlob
	default:
		return SymbolScalar
	}
}

// stripSigil removes the sigil from a variable name
func (b *DefaultBinder) stripSigil(name string) string {
	if len(name) > 1 && (name[0] == '$' || name[0] == '@' || name[0] == '%' || name[0] == '*') {
		return name[1:]
	}
	return name
}
