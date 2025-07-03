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

// Bind performs symbol binding on an AST node
func (b *DefaultBinder) Bind(node ast.Node) (*SymbolTable, error) {
	// Create a new symbol table
	symbolTable := NewSymbolTableWithPool(b.poolManager, "main")

	// For now, this is a placeholder implementation
	// In a full implementation, this would traverse the AST and bind symbols

	return symbolTable, nil
}

// BindAST performs symbol binding on an AST node (alternative to Bind)
func (b *DefaultBinder) BindAST(node ast.Node) (*SymbolTable, error) {
	// This is the same as Bind but with a different name for compatibility
	return b.Bind(node)
}

// getVariableSymbolKind determines the symbol kind for a variable
func (b *DefaultBinder) getVariableSymbolKind(sigil string) SymbolKind {
	switch sigil {
	case "$":
		return SymbolScalar
	case "@":
		return SymbolArray
	case "%":
		return SymbolHash
	case "*":
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

// bindSubroutineDeclaration binds a subroutine declaration
func (b *DefaultBinder) bindSubroutineDeclaration(node ast.Node) error {
	// Placeholder implementation for testing
	return nil
}

// bindMethodDeclaration binds a method declaration
func (b *DefaultBinder) bindMethodDeclaration(node ast.Node) error {
	// Placeholder implementation for testing
	return nil
}
