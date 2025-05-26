// ABOUTME: DEPRECATED - AST compilation has been moved to internal/compiler package
// ABOUTME: Use compiler.CompilerRegistry for AST-to-Perl compilation

package parser

import (
	"fmt"
)

// ASTGenerator is deprecated, use internal/compiler package instead
type ASTGenerator struct {
	// DEPRECATED: Use compiler.CompilerRegistry instead
}

// GenerateFromAST is deprecated, use compiler.CompilerRegistry.Compile instead
func GenerateFromAST(ast *AST, includeTypes bool) (string, error) {
	return "", fmt.Errorf("GenerateFromAST is deprecated, use internal/compiler package instead")
}
