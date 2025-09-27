// ABOUTME: Core type checking logic and AST analysis
// ABOUTME: Implements the TypeChecker that analyzes ASTs for type errors

package typechecker

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/typedef"
)

// TypeChecker performs type checking on parsed Perl code
type TypeChecker struct {
	// Hierarchy is the type hierarchy used for checking
	Hierarchy *typedef.TypeHierarchy

	// SymbolTable contains symbol information from the binder phase
	SymbolTable *binder.SymbolTable

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

	// InferenceEngine for advanced type inference
	InferenceEngine *InferenceEngine

	// MaybeHandler provides comprehensive Maybe type operations
	MaybeHandler *MaybeTypeHandler

	// Debug enables debug mode
	Debug bool

	// SafetyAnalysisEnabled controls whether safety analysis is performed
	SafetyAnalysisEnabled bool
}

// NewTypeChecker creates a new TypeChecker with the given type hierarchy and symbol table
func NewTypeChecker(hierarchy *typedef.TypeHierarchy, symbolTable *binder.SymbolTable, moduleName string) *TypeChecker {
	// Create initial type state
	initialState := &TypeState{
		VariableTypes: make(map[string]string),
		RefinedTypes:  make(map[string]string),
		Conditions:    []Condition{},
	}

	tc := &TypeChecker{
		Hierarchy:                 hierarchy,
		SymbolTable:               symbolTable,
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

	// Initialize Maybe type handler
	tc.MaybeHandler = NewMaybeTypeHandler(tc)

	// Initialize validation patterns
	tc.initializeValidationPatterns()

	return tc
}

// populateTypesFromSymbolTable extracts type information from the symbol table
func (tc *TypeChecker) populateTypesFromSymbolTable() {
	if tc.SymbolTable == nil {
		return
	}

	// Extract variable types from symbols
	for name, symbols := range tc.SymbolTable.Symbols {
		for _, symbol := range symbols {
			if symbol.Type != "" {
				tc.VariableTypes[name] = symbol.Type
				tc.TypeAnnotations[name] = symbol.Type
			}
		}
	}

	// Extract function types from subroutine and method symbols
	for _, symbols := range tc.SymbolTable.Symbols {
		for _, symbol := range symbols {
			if symbol.Kind == binder.SymbolSubroutine || symbol.Kind == binder.SymbolMethod {
				if symbol.Type != "" {
					// For functions, the type is the return type
					signature := &FunctionSignature{
						ParameterTypes: make(map[string]string),
						ReturnType:     symbol.Type,
						IsMethod:       symbol.Kind == binder.SymbolMethod,
					}
					tc.FunctionTypes[symbol.Name] = signature
				}
			}
		}
	}
}

// checkSymbolUsage checks for unused variables and shadowing warnings
func (tc *TypeChecker) checkSymbolUsage() []error {
	var warnings []error
	if tc.SymbolTable == nil {
		return warnings
	}

	// Check for unused variables
	for _, symbols := range tc.SymbolTable.Symbols {
		for _, symbol := range symbols {
			if symbol.Flags&binder.SymbolFlagLexical != 0 {
				// For now, just collect symbols - in a real implementation,
				// we'd track usage and report unused ones
				_ = symbol
			}
		}
	}

	// Check for shadowing
	tc.checkShadowing(&warnings)

	return warnings
}

// checkShadowing checks for variable shadowing and adds warnings
func (tc *TypeChecker) checkShadowing(warnings *[]error) {
	// Walk through scopes and check for shadowed variables
	tc.walkScopes(tc.SymbolTable.GlobalScope, warnings)
}

// walkScopes recursively walks scopes to find shadowing
func (tc *TypeChecker) walkScopes(scope *binder.Scope, warnings *[]error) {
	if scope == nil {
		return
	}

	// Check each symbol in this scope
	for name, symbol := range scope.Symbols {
		// Look for same-named symbols in parent scopes
		parent := scope.Parent
		for parent != nil {
			if parentSymbol, exists := parent.Symbols[name]; exists {
				// Found shadowing
				warning := fmt.Errorf("variable '%s' at line %d shadows outer scope variable at line %d",
					name, symbol.Position.Line, parentSymbol.Position.Line)
				*warnings = append(*warnings, warning)
				break
			}
			parent = parent.Parent
		}
	}

	// Recursively check child scopes
	for _, child := range scope.Children {
		tc.walkScopes(child, warnings)
	}
}

// NewTypeCheckerLegacy creates a new TypeChecker without symbol table for backward compatibility
func NewTypeCheckerLegacy(hierarchy *typedef.TypeHierarchy, moduleName string) *TypeChecker {
	// Create empty symbol table for backward compatibility
	symbolTable := binder.NewSymbolTable()
	symbolTable.Package = moduleName
	return NewTypeChecker(hierarchy, symbolTable, moduleName)
}

// CheckAST performs type checking on an entire AST using symbol table information
func (tc *TypeChecker) CheckAST(ast *ast.AST) []error {
	var typeErrors []error

	// Extract information about imported modules
	tc.extractImports(ast)

	// Use symbol table to populate type information
	tc.populateTypesFromSymbolTable()

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

	// Check for symbol usage issues (unused variables, shadowing)
	usageWarnings := tc.checkSymbolUsage()
	if len(usageWarnings) > 0 {
		typeErrors = append(typeErrors, usageWarnings...)
	}

	// Enhanced Maybe type safety analysis
	if tc.MaybeHandler != nil {
		maybeErrors := tc.checkMaybeTypeSafety(ast)
		if len(maybeErrors) > 0 {
			typeErrors = append(typeErrors, maybeErrors...)
		}
	}

	return typeErrors
}

// extractImports extracts import information from the AST (placeholder)
func (tc *TypeChecker) extractImports(ast *ast.AST) {
	// This would be implemented to extract module imports
	// For now, it's a placeholder
}

// collectTypeAnnotation collects a type annotation without full validation
func (tc *TypeChecker) collectTypeAnnotation(annotation *ast.TypeAnnotation) error {
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
	case ast.VarAnnotation:
		tc.VariableTypes[annotation.AnnotatedItem] = typeStr
	}

	return nil
}

// checkASTFunctionReturns checks function return types
func (tc *TypeChecker) checkASTFunctionReturns(ast *ast.AST) []error {
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

// CheckAssignment checks if a source type can be assigned to a target type
func (tc *TypeChecker) CheckAssignment(sourceType, targetType string, pos ast.Position) error {
	return tc.Hierarchy.CheckTypeCompatibility(sourceType, targetType)
}

// refineTypeAfterCondition refines a variable's type based on a conditional check
func (tc *TypeChecker) refineTypeAfterCondition(variable, condition string, positive bool) {
	currentType, exists := tc.VariableTypes[variable]
	if !exists {
		return
	}

	var refinedType string

	switch condition {
	case "defined":
		if positive {
			// After defined($var), exclude Undef from the type
			refinedType = tc.excludeTypeFromUnion(currentType, "Undef")
		} else {
			// After !defined($var), the variable must be Undef
			if tc.Hierarchy.IsUnionType(currentType) {
				unionType := tc.Hierarchy.ParseUnionType(currentType)
				if unionType != nil && unionType.ContainsMember("Undef") {
					refinedType = "Undef"
				}
			} else if currentType == "Maybe["+tc.extractMaybeParameter(currentType)+"]" {
				refinedType = "Undef"
			}
		}
	case "ref":
		// ref() checks are handled differently - type refinement is done in flow analyzer
		// since it depends on the specific ref type being checked
		return
	}

	if refinedType != "" {
		tc.TypeState.RefinedTypes[variable] = refinedType
	}
}

// excludeTypeFromUnion removes a specific type from a union type
func (tc *TypeChecker) excludeTypeFromUnion(unionType, excludeType string) string {
	// Handle Maybe[T] special case
	if strings.HasPrefix(unionType, "Maybe[") {
		param := tc.extractMaybeParameter(unionType)
		if param != "" {
			return param // Maybe[T] without Undef becomes T
		}
	}

	// Handle regular union types
	if tc.Hierarchy.IsUnionType(unionType) {
		unionInstance := tc.Hierarchy.ParseUnionType(unionType)
		if unionInstance != nil {
			members := unionInstance.GetMembers()
			var newMembers []string
			for _, member := range members {
				if member != excludeType {
					newMembers = append(newMembers, member)
				}
			}

			if len(newMembers) == 1 {
				return newMembers[0] // Single type remaining
			} else if len(newMembers) > 1 {
				return strings.Join(newMembers, "|") // Remaining union
			}
		}
	}

	return unionType // No change
}

// extractMaybeParameter extracts the parameter from Maybe[T]
func (tc *TypeChecker) extractMaybeParameter(maybeType string) string {
	if strings.HasPrefix(maybeType, "Maybe[") && strings.HasSuffix(maybeType, "]") {
		return maybeType[6 : len(maybeType)-1]
	}
	return ""
}

// inferExpressionType infers the type of a literal expression
func (tc *TypeChecker) inferExpressionType(expression string) string {
	expression = strings.TrimSpace(expression)

	// Integer literals
	if strings.Trim(expression, "0123456789") == "" && expression != "" {
		return "Int"
	}

	// Float literals
	if strings.Contains(expression, ".") {
		if strings.Trim(expression, "0123456789.") == "" {
			return "Float"
		}
	}

	// String literals
	if (strings.HasPrefix(expression, "\"") && strings.HasSuffix(expression, "\"")) ||
		(strings.HasPrefix(expression, "'") && strings.HasSuffix(expression, "'")) {
		return "Str"
	}

	// Special values
	switch expression {
	case "undef":
		return "Undef"
	case "[]":
		return "ArrayRef"
	case "{}":
		return "HashRef"
	}

	// Default to unknown
	return "Unknown"
}

// initializeValidationPatterns initializes the validation patterns for flow-sensitive analysis
func (tc *TypeChecker) initializeValidationPatterns() {
	// This would initialize flow-sensitive validation patterns
	// For now, it's a placeholder
}

// validateType validates that a type string is valid
func (tc *TypeChecker) validateType(typeStr string) error {
	return tc.Hierarchy.ValidateType(typeStr)
}

// checkTypeAnnotation validates a single type annotation
func (tc *TypeChecker) checkTypeAnnotation(annotation *ast.TypeAnnotation) error {
	if annotation == nil || annotation.TypeExpression == nil {
		return errors.NewTypeError(
			ErrTypeValidationError,
			"Nil type annotation",
			nil,
		)
	}

	typeStr := annotation.TypeExpression.String()
	return tc.validateType(typeStr)
}

// checkMaybeTypeSafety performs comprehensive Maybe type safety analysis
func (tc *TypeChecker) checkMaybeTypeSafety(ast *ast.AST) []error {
	var maybeErrors []error

	if tc.MaybeHandler == nil {
		return maybeErrors
	}

	// Analyze all variable usages for Maybe type safety
	// Note: In real implementation, this would walk the AST to find actual variable usage positions
	for varName, varType := range tc.VariableTypes {
		if tc.MaybeHandler.IsMaybeType(varType) {
			// This would be implemented to track actual usage positions
			// For now, skip the position-dependent checks
			_ = varName
		}
	}

	// Analyze Perl idiom integration patterns
	// This would typically walk the AST to find expressions with //, ||=, etc.
	// For now, we'll simulate some common patterns
	tc.analyzePerlIdioms(&maybeErrors)

	return maybeErrors
}

// analyzePerlIdioms analyzes Perl idiom usage with Maybe types
func (tc *TypeChecker) analyzePerlIdioms(errors *[]error) {
	// In a real implementation, this would walk the AST looking for:
	// 1. defined-or operator (//) usage
	// 2. logical-or assignment (||=) usage
	// 3. defined() checks
	// 4. exists() checks
	//
	// For now, we'll check if we have good practices vs unsafe accesses
	if tc.MaybeHandler != nil {
		report := tc.MaybeHandler.GenerateSafetyReport()

		// If safety score is low, add a general recommendation
		if report.SafetyScore < 70.0 {
			generalError := fmt.Errorf("WARNING_MAYBE_TYPE_SAFETY: Maybe type safety score: %.1f%%. Consider using // operator and defined() checks", report.SafetyScore)
			*errors = append(*errors, generalError)
		}
	}
}
