// ABOUTME: Adapter to connect parser AST to compiler AST interface
// ABOUTME: Allows using parser.AST with the compiler package

package compiler

import (
	"os"

	"tamarou.com/pvm/internal/parser"
)

// ParserASTAdapter adapts parser.AST to implement the compiler AST interface
type ParserASTAdapter struct {
	ast *parser.AST
}

// NewParserASTAdapter creates a new adapter for parser.AST
func NewParserASTAdapter(ast *parser.AST) *ParserASTAdapter {
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

	if a.ast.Path == "" {
		return "", NewCompilerError(ErrInvalidAST, "AST has no file path")
	}

	content, err := os.ReadFile(a.ast.Path)
	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed, "failed to read source file").WithCause(err)
	}

	return string(content), nil
}
