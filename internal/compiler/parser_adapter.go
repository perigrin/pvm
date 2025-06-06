// ABOUTME: Adapter to connect parser AST to compiler AST interface
// ABOUTME: Allows using parser.AST with the compiler package

package compiler

import (
	"os"

	"tamarou.com/pvm/internal/ast"
)

// ParserASTAdapter adapts ast.AST to implement the compiler AST interface
type ParserASTAdapter struct {
	ast *ast.AST
}

// NewParserASTAdapter creates a new adapter for ast.AST
func NewParserASTAdapter(ast *ast.AST) *ParserASTAdapter {
	return &ParserASTAdapter{ast: ast}
}

// GetPath returns the source file path
func (a *ParserASTAdapter) GetPath() string {
	if a.ast == nil {
		return ""
	}
	return a.ast.Path
}

// IsValid returns true if the AST is valid for compilation
func (a *ParserASTAdapter) IsValid() bool {
	if a.ast == nil {
		return false
	}

	// Check if there are any parse errors
	return len(a.ast.Errors) == 0
}

// GetContent returns the original source content
func (a *ParserASTAdapter) GetContent() (string, error) {
	if a.ast == nil {
		return "", NewCompilerError(ErrInvalidAST, "AST is nil")
	}

	// First try to get content from the AST's Source field (for string parsing)
	if a.ast.Source != "" {
		return a.ast.Source, nil
	}

	// Fall back to reading from file if no source content is stored
	if a.ast.Path == "" {
		return "", NewCompilerError(ErrInvalidAST, "AST has no source content or file path")
	}

	content, err := os.ReadFile(a.ast.Path)
	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed, "failed to read source file").WithCause(err)
	}

	return string(content), nil
}

// GetRootNode returns the root AST node
func (a *ParserASTAdapter) GetRootNode() (ast.Node, error) {
	if a.ast == nil {
		return nil, NewCompilerError(ErrInvalidAST, "adapter contains nil AST")
	}
	if a.ast.Root == nil {
		return nil, NewCompilerError(ErrInvalidAST, "AST has no root node")
	}
	return a.ast.Root, nil
}
