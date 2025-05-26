// ABOUTME: Type definitions for compiler package to avoid circular dependencies
// ABOUTME: Defines AST interface that can be implemented by any parser

package compiler

// AST represents an Abstract Syntax Tree that can be compiled
type AST interface {
	// GetPath returns the source file path
	GetPath() string

	// IsValid returns true if the AST is valid for compilation
	IsValid() bool

	// GetContent returns the original source content if available
	GetContent() (string, error)
}

// SimpleAST is a basic implementation of AST for testing
type SimpleAST struct {
	Path    string
	Content string
	Valid   bool
}

// GetPath returns the source file path
func (a *SimpleAST) GetPath() string {
	return a.Path
}

// IsValid returns true if the AST is valid
func (a *SimpleAST) IsValid() bool {
	return a.Valid
}

// GetContent returns the source content
func (a *SimpleAST) GetContent() (string, error) {
	return a.Content, nil
}
