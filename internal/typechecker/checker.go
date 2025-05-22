// ABOUTME: Core type checking logic and AST analysis
// ABOUTME: Implements the TypeChecker that analyzes ASTs for type errors

package typechecker

import (
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// TypeChecker performs type checking on parsed Perl code
type TypeChecker struct {
	// Hierarchy is the type hierarchy used for checking
	Hierarchy *typedef.TypeHierarchy

	// CurrentModule is the current module being checked
	CurrentModule string

	// ImportedModules tracks imported modules
	ImportedModules map[string]bool

	// TypeAnnotations tracks annotated types
	TypeAnnotations map[string]string

	// VariableTypes maps variable names to their types (from annotations or inference)
	VariableTypes map[string]string

	// FunctionTypes maps function names to their signature information
	FunctionTypes map[string]*FunctionSignature

	// TypeState tracks the current type state for flow-sensitive analysis
	TypeState *TypeState

	// TypeStateStack holds type states for different code paths
	TypeStateStack []*TypeState

	// ValidationPatterns holds recognized validation patterns
	ValidationPatterns []ValidationPattern

	// ContextSensitiveFunctions maps function names to their context-dependent return types
	ContextSensitiveFunctions map[string]map[string]string

	// TypeAliases maps alias names to their target types
	TypeAliases map[string]string

	// GenericFunctions maps function names to their generic signature information
	GenericFunctions map[string]*GenericFunctionSignature

	// ModuleTypes maps module names to their exported types
	ModuleTypes map[string]map[string]string

	// HigherKindedTypes maps type names to their higher-kinded definitions
	HigherKindedTypes map[string]*HigherKindedTypeDefinition

	// Debug enables debug mode
	Debug bool
}

// NewTypeChecker creates a new TypeChecker with the given type hierarchy
func NewTypeChecker(hierarchy *typedef.TypeHierarchy, moduleName string) *TypeChecker {
	// Create initial type state
	initialState := &TypeState{
		VariableTypes: make(map[string]string),
		RefinedTypes:  make(map[string]string),
		Conditions:    []Condition{},
	}

	tc := &TypeChecker{
		Hierarchy:                 hierarchy,
		CurrentModule:             moduleName,
		ImportedModules:           make(map[string]bool),
		TypeAnnotations:           make(map[string]string),
		VariableTypes:             make(map[string]string),
		FunctionTypes:             make(map[string]*FunctionSignature),
		TypeState:                 initialState,
		TypeStateStack:            []*TypeState{},
		ValidationPatterns:        []ValidationPattern{},
		ContextSensitiveFunctions: make(map[string]map[string]string),
		TypeAliases:               make(map[string]string),
		GenericFunctions:          make(map[string]*GenericFunctionSignature),
		ModuleTypes:               make(map[string]map[string]string),
		HigherKindedTypes:         make(map[string]*HigherKindedTypeDefinition),
		Debug:                     false,
	}

	// Initialize validation patterns
	tc.initializeValidationPatterns()

	return tc
}

// CheckAST performs type checking on an entire AST
func (tc *TypeChecker) CheckAST(ast *parser.AST) []error {
	var typeErrors []error

	// Extract information about imported modules
	tc.extractImports(ast)

	// First pass: collect all type annotations without validating them yet
	for _, annotation := range ast.TypeAnnotations {
		if err := tc.collectTypeAnnotation(annotation); err != nil {
			typeErrors = append(typeErrors, err)
		}
	}

	// Second pass: validate all type annotations
	for _, annotation := range ast.TypeAnnotations {
		if err := tc.checkTypeAnnotation(annotation); err != nil {
			typeErrors = append(typeErrors, err)
		}
	}

	// Third pass: validate usage of types in code
	assignmentErrors := tc.CheckASTAssignments(ast)
	if len(assignmentErrors) > 0 {
		typeErrors = append(typeErrors, assignmentErrors...)
	}

	// Fourth pass: validate function return types
	returnErrors := tc.checkASTFunctionReturns(ast)
	if len(returnErrors) > 0 {
		typeErrors = append(typeErrors, returnErrors...)
	}

	// Finally, perform flow-sensitive type analysis if enabled
	if tc.TypeState != nil {
		flowErrors := tc.performFlowSensitiveAnalysis(ast)
		if len(flowErrors) > 0 {
			typeErrors = append(typeErrors, flowErrors...)
		}
	}

	return typeErrors
}

// extractImports extracts import information from the AST (placeholder)
func (tc *TypeChecker) extractImports(ast *parser.AST) {
	// This would be implemented to extract module imports
	// For now, it's a placeholder
}

// collectTypeAnnotation collects a type annotation without full validation
func (tc *TypeChecker) collectTypeAnnotation(annotation *parser.TypeAnnotation) error {
	if annotation == nil || annotation.TypeExpression == nil {
		return errors.NewTypeError(
			ErrTypeValidationError,
			"Nil type annotation",
			nil,
		)
	}

	typeStr := annotation.TypeExpression.String()

	// Store the type annotation for lookup during validation
	tc.TypeAnnotations[annotation.AnnotatedItem] = typeStr

	// Record additional type information based on the kind of annotation
	switch annotation.Kind {
	case parser.VarAnnotation:
		tc.VariableTypes[annotation.AnnotatedItem] = typeStr
	}

	return nil
}

// checkASTFunctionReturns checks function return types
func (tc *TypeChecker) checkASTFunctionReturns(ast *parser.AST) []error {
	// This would check function return type compatibility
	// For now, it's a placeholder that returns no errors
	return []error{}
}

// GetVariableType returns the type of a variable if known
func (tc *TypeChecker) GetVariableType(varName string) (string, bool) {
	// First check if we have an explicit type annotation
	if varType, ok := tc.VariableTypes[varName]; ok {
		return varType, true
	}

	// Then check type annotations
	if varType, ok := tc.TypeAnnotations[varName]; ok {
		return varType, true
	}

	return "", false
}
