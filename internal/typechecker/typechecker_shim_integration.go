// ABOUTME: Integration functions for typechecker with tree-sitter shim
// ABOUTME: Provides convenient methods to use TypeChecker with TreeSitterAST

package typechecker

import (
	"fmt"
	"os"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
)

// CheckTreeSitterAST performs type checking on a TreeSitterAST
// This is a convenience method that handles the conversion from TreeSitterAST to ast.AST
func (tc *TypeCheck) CheckTreeSitterAST(shimAST *ast.TreeSitterAST, moduleName string) (*TypeCheckResult, error) {
	// Convert TreeSitterAST to regular AST for typechecker
	astInterface := &ast.AST{
		Path:            shimAST.Path,
		Root:            shimAST.Root,
		TypeAnnotations: shimAST.TypeAnnotations,
		Errors:          shimAST.Errors,
		Source:          shimAST.Source,
	}

	// Check for parser errors first
	if len(astInterface.Errors) > 0 {
		result := &TypeCheckResult{
			Path:                 shimAST.Path,
			Errors:               []TypeCheckError{},
			TypeAnnotations:      shimAST.TypeAnnotations,
			RefinedTypes:         make(map[string]string),
			FlowSensitiveEnabled: tc.EnableFlowSensitiveAnalysis,
		}

		// Convert parser errors to type check errors
		for _, parseErr := range astInterface.Errors {
			typErr := TypeCheckError{
				Message: parseErr.Error(),
				Line:    0,
				Column:  0,
				Path:    shimAST.Path,
			}
			result.Errors = append(result.Errors, typErr)
		}

		return result, nil
	}

	// Use CST-based binding for symbol resolution
	symbolTable, err := tc.bindWithTreeSitterCST(shimAST)
	if err != nil {
		return nil, fmt.Errorf("CST symbol binding failed: %w", err)
	}

	// Create a type checker with symbol table
	checker := NewTypeChecker(tc.TypeHierarchy, symbolTable, moduleName)

	// Update inference engine with symbol table
	if tc.EnableTypeInference {
		checker.InferenceEngine = NewInferenceEngine(tc.TypeHierarchy, symbolTable)
	}

	// Configure flow-sensitive analysis options
	if tc.EnableFlowSensitiveAnalysis {
		if tc.SkipFlowChecks {
			checker.TypeState.SkipFlowChecks = true
		}

		if len(tc.FlowPatterns) > 0 {
			checker.AddFlowPatterns(tc.FlowPatterns)
		}
	}

	// Perform type inference if enabled
	if tc.EnableTypeInference && tc.InferenceEngine != nil {
		if err := tc.InferenceEngine.InferTypesFromTreeSitterAST(shimAST); err != nil {
			// Log inference error but don't fail type checking
			fmt.Fprintf(os.Stderr, "Type inference warning: %v\n", err)
		}

		// Apply inferred types to the checker
		inferredTypes := tc.InferenceEngine.GetAllInferredTypes()
		for varName, inferredType := range inferredTypes {
			// Only apply inferred type if variable doesn't already have an explicit type
			if _, hasType := checker.VariableTypes[varName]; !hasType {
				checker.VariableTypes[varName] = inferredType
			}
		}
	}

	// Check the AST for type errors
	typeErrors := checker.CheckAST(astInterface)

	// Build result
	result := &TypeCheckResult{
		Path:                 shimAST.Path,
		TypeAnnotations:      shimAST.TypeAnnotations,
		RefinedTypes:         make(map[string]string), // TODO: Extract refined types from checker
		FlowSensitiveEnabled: tc.EnableFlowSensitiveAnalysis,
	}

	// Convert type errors to TypeCheckError format
	for _, err := range typeErrors {
		typErr := TypeCheckError{
			Message: err.Error(),
			Line:    0, // TODO: Extract line/column from error if available
			Column:  0,
			Path:    shimAST.Path,
		}
		result.Errors = append(result.Errors, typErr)
	}

	// Cache the result if caching is enabled
	if tc.fileCache != nil && shimAST.Path != "" {
		tc.fileCache[shimAST.Path] = result
		tc.manageCacheSize()
	}

	return result, nil
}

// bindWithTreeSitterCST performs symbol binding using TreeSitterAST's CST directly
// This leverages the tree-sitter tree for more efficient binding
func (tc *TypeCheck) bindWithTreeSitterCST(shimAST *ast.TreeSitterAST) (*binder.SymbolTable, error) {
	// Get the underlying tree-sitter tree from the shim
	tree := shimAST.GetTree()
	if tree == nil {
		return nil, fmt.Errorf("no tree-sitter tree available in shimAST")
	}

	// Get the root node
	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, fmt.Errorf("no root node available in tree-sitter tree")
	}

	// Use CST-based binding with the tree-sitter tree
	sourceBytes := []byte(shimAST.Source)
	symbolTable, err := tc.Binder.BindCST(rootNode, sourceBytes, shimAST.TypeAnnotations)
	if err != nil {
		return nil, fmt.Errorf("failed to bind symbols from tree-sitter CST: %w", err)
	}

	return symbolTable, nil
}

// CheckShimParserFile is a high-level convenience function for type checking files using tree-sitter shim
// This parses the file with the shim parser and performs type checking in one step
func CheckShimParserFile(path string) (*TypeCheckResult, error) {
	// This would require a shim parser instance - leaving as a placeholder for now
	// In practice, you'd create a shim parser, parse the file, and call CheckTreeSitterAST
	return nil, fmt.Errorf("CheckShimParserFile not yet implemented - use CheckTreeSitterAST with parsed shimAST")
}
