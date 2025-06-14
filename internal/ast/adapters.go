// ABOUTME: Factory functions for creating concrete AST nodes  
// ABOUTME: Provides utilities for constructing typed AST structures

package ast

// CreateConcreteAST creates concrete AST nodes from generic interfaces
// This is used when we need to construct proper typed AST nodes from parsed content
func CreateConcreteAST(path, source string, root Node, annotations []*TypeAnnotation, errors []error) *AST {
	return &AST{
		Path:            path,
		Root:            root,
		TypeAnnotations: annotations,
		Errors:          errors,
		Source:          source,
	}
}
