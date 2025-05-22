// ABOUTME: Type definitions and interfaces for the typechecker package
// ABOUTME: Defines core types used throughout type checking

package typechecker

import (
	"fmt"

	"tamarou.com/pvm/internal/parser"
)

// PSC Error codes for type checking
const (
	ErrTypeAnnotationMismatch = "810" // Type annotation doesn't match expected type
	ErrTypeInferenceError     = "811" // Failed to infer type
	ErrTypeValidationError    = "812" // Type validation error
	ErrTypeAssignmentError    = "813" // Error in variable assignment
	ErrTypeFunctionError      = "814" // Error in function parameter or return type
	ErrTypeDeclarationError   = "815" // Error in type declaration
	ErrTypeIncompatibleError  = "816" // Incompatible types in expression
)

// TypeCheckResult represents the result of type checking a file
type TypeCheckResult struct {
	// Path is the path to the checked file
	Path string

	// Errors is a list of type errors found during checking
	Errors []TypeCheckError

	// TypeAnnotations is a list of type annotations found in the code
	TypeAnnotations []*parser.TypeAnnotation

	// RefinedTypes maps variable names to their refined types from flow-sensitive analysis
	RefinedTypes map[string]string

	// FlowSensitiveEnabled indicates if flow-sensitive analysis was enabled for this check
	FlowSensitiveEnabled bool
}

// TypeCheckError provides detailed information about a type error
type TypeCheckError struct {
	// Message is the error message
	Message string

	// Path is the file path where the error occurred
	Path string

	// Line is the line number where the error occurred
	Line int

	// Column is the column number where the error occurred
	Column int
}

// Error implements the error interface
func (e TypeCheckError) Error() string {
	if e.Path != "" && e.Line > 0 {
		return fmt.Sprintf("%s:%d:%d: %s", e.Path, e.Line, e.Column, e.Message)
	}
	return e.Message
}

// FunctionSignature represents the type signature of a function or method
type FunctionSignature struct {
	// ParameterTypes maps parameter names to their types
	ParameterTypes map[string]string

	// ReturnType is the return type of the function
	ReturnType string

	// IsMethod indicates if this is a method
	IsMethod bool
}

// GenericFunctionSignature represents a generic function signature
type GenericFunctionSignature struct {
	// TypeParameters lists the generic type parameters
	TypeParameters []string

	// ParameterTypes maps parameter names to their types (may include type parameters)
	ParameterTypes map[string]string

	// ReturnType is the return type (may include type parameters)
	ReturnType string

	// Constraints maps type parameters to their constraint types
	Constraints map[string][]string

	// IsMethod indicates if this is a method
	IsMethod bool
}

// HigherKindedTypeDefinition represents a higher-kinded type definition
type HigherKindedTypeDefinition struct {
	// Name is the name of the higher-kinded type
	Name string

	// TypeConstructors lists the type constructor parameters
	TypeConstructors []string

	// Definition is the type definition body
	Definition string
}

// TypeState represents the types of variables at a specific point in the code
// It is used for flow-sensitive analysis to track how types change based on control flow
type TypeState struct {
	// VariableTypes maps variable names to their types in this state
	VariableTypes map[string]string

	// RefinedTypes maps variable names to their refined types based on control flow
	RefinedTypes map[string]string

	// Conditions tracks the conditions that led to this state
	Conditions []Condition

	// SkipFlowChecks controls whether to skip flow-sensitive type checks
	// but still perform type refinements based on control flow
	SkipFlowChecks bool
}

// Condition represents a condition that affects type refinement
type Condition struct {
	// Variable is the variable being checked in the condition
	Variable string

	// Operator is the comparison operator used (==, !=, >, <, etc.)
	Operator string

	// Value is the value being compared against
	Value string

	// Negated indicates if the condition is negated
	Negated bool
}

// ValidationPattern represents a recognized pattern for type validation
type ValidationPattern struct {
	// Name is the name of the pattern (e.g., "defined check", "ref check")
	Name string

	// Pattern is a simplified representation of the pattern
	Pattern string

	// RefinementFunc is the function that refines the type
	RefinementFunc func(variable string, currentType string) string

	// Checker is a function that checks if code matches this pattern
	Checker func(node parser.Node) (string, bool)
}
