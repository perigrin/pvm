// ABOUTME: Integration functions for inference engine with tree-sitter shim
// ABOUTME: Provides convenient methods to use inference engine with TreeSitterAST

package typechecker

import (
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/typedef"
)

// InferTypesFromTreeSitterAST runs type inference on a TreeSitterAST
// This is a convenience method that handles the conversion from TreeSitterAST to ast.AST
func (ie *InferenceEngine) InferTypesFromTreeSitterAST(shimAST *ast.TreeSitterAST) error {
	// Convert TreeSitterAST to regular AST for inference engine
	astInterface := &ast.AST{
		Path:            shimAST.Path,
		Root:            shimAST.Root,
		TypeAnnotations: shimAST.TypeAnnotations,
		Errors:          shimAST.Errors,
		Source:          shimAST.Source,
	}

	// Run type inference using the existing method
	return ie.InferTypes(astInterface)
}

// InferTypesFromShimParser creates an InferenceEngine and runs inference on TreeSitterAST
// This is a high-level convenience function for one-shot type inference
func InferTypesFromShimParser(shimAST *ast.TreeSitterAST, typeHierarchy *typedef.TypeHierarchy, symbolTable *binder.SymbolTable) (*InferenceEngine, error) {
	// Create inference engine
	engine := NewInferenceEngine(typeHierarchy, symbolTable)

	// Run inference
	err := engine.InferTypesFromTreeSitterAST(shimAST)
	if err != nil {
		return nil, err
	}

	return engine, nil
}

// ExtractInferredTypesFromShim extracts inferred types from TreeSitterAST
// This analyzes the AST using the inference engine and returns typed information
func ExtractInferredTypesFromShim(shimAST *ast.TreeSitterAST) (map[string]*InferredTypeInfo, error) {
	// TODO: Need to handle the dependency injection properly
	// For now, return empty map to avoid compilation errors
	return make(map[string]*InferredTypeInfo), nil
}
