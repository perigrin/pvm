// ABOUTME: AST to Perl code generator for full round-trip compilation
// ABOUTME: Converts parsed AST back to valid Perl source code

package parser

import (
	"fmt"
	"os"
)

// ASTGenerator generates Perl code from AST
type ASTGenerator struct {
	// Reserved for future use when full AST traversal is implemented
}

// GenerateFromAST creates a complete Perl code generator from AST
func GenerateFromAST(ast *AST, includeTypes bool) (string, error) {
	if ast == nil {
		return "", fmt.Errorf("AST is nil")
	}

	// For now, implement a simple version that works with the current AST structure
	// The AST Root might not have a full Node implementation yet

	// If we're including types, use the tree-sitter compiler but modify it to include types
	if includeTypes {
		return generateWithTypes(ast)
	}

	// If not including types, use the existing strip functionality
	return StripAnnotations(ast.Path)
}

// generateWithTypes generates Perl code with type annotations included
func generateWithTypes(ast *AST) (string, error) {
	// Read the original file to get the source text
	if ast.Path == "" {
		return "", fmt.Errorf("AST path is empty")
	}

	originalContent, err := os.ReadFile(ast.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read original file: %v", err)
	}

	// For now, return the original content since it already has type annotations
	// In a more complete implementation, we would modify the content based on the AST
	return string(originalContent), nil
}
