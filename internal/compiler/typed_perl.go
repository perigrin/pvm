// ABOUTME: Typed Perl compiler that preserves type annotations in AST output
// ABOUTME: Generates Perl code with full type information for PSC consumption

package compiler

import (
	"fmt"
	"tamarou.com/pvm/internal/ast"
)

// TypedPerlCompiler compiles AST to Perl code with type annotations preserved
type TypedPerlCompiler struct {
	options *CompilerOptions
}

// NewTypedPerlCompiler creates a new typed Perl compiler
func NewTypedPerlCompiler() *TypedPerlCompiler {
	return &TypedPerlCompiler{
		options: &CompilerOptions{
			PreserveComments:   true,
			PreserveFormatting: true,
			StrictMode:         false,
			CustomPatterns:     nil,
		},
	}
}

// Target returns the compilation target
func (c *TypedPerlCompiler) Target() Target {
	return TargetTypedPerl
}

// Validate checks if the AST is suitable for typed Perl compilation
func (c *TypedPerlCompiler) Validate(ast *ast.AST) error {
	if ast == nil {
		return NewCompilerError(ErrInvalidAST, "AST cannot be nil")
	}

	// Check if we can get content either from source or file path
	_, err := ast.GetContent()
	if err != nil {
		return NewCompilerError(ErrInvalidAST, "AST must have accessible source content").WithCause(err)
	}

	return nil
}

// Compile converts an AST to typed Perl code with type annotations preserved
func (c *TypedPerlCompiler) Compile(ast *ast.AST) (string, error) {
	if err := c.Validate(ast); err != nil {
		return "", err
	}

	// Get the original content since it already contains type annotations
	content, err := ast.GetContent()
	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed,
			fmt.Sprintf("failed to get source content: %v", err)).
			WithLocation(ast.GetPath(), 0, 0).
			WithCause(err)
	}

	// For now, return the original content since it already has type annotations
	// In a more complete implementation, we would regenerate the code from the AST
	// with type annotations included
	return content, nil
}

// SetOptions updates the compiler options
func (c *TypedPerlCompiler) SetOptions(options *CompilerOptions) {
	c.options = options
}
