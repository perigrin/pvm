// ABOUTME: Method signature validation and type checking for typed Perl methods
// ABOUTME: Validates parameter types, return types, and signature compatibility

package typechecker

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/errors"
)

// MethodSignature represents a parsed method signature with full type information
type MethodSignature struct {
	Name       string
	Parameters []MethodParameter
	ReturnType *ast.TypeExpression
	IsVariadic bool
	ClassName  string // Class context for the method
	SourcePos  ast.Position
}

// MethodParameter represents a typed parameter in a method signature
type MethodParameter struct {
	Name       string
	Type       *ast.TypeExpression
	IsOptional bool
	IsNamed    bool
	Position   int
}

// MethodSignatureValidator validates method signatures and parameter compatibility
type MethodSignatureValidator struct {
	typeChecker *TypeChecker
	signatures  map[string]*MethodSignature // method_name -> signature
}

// NewMethodSignatureValidator creates a new method signature validator
func NewMethodSignatureValidator(tc *TypeChecker) *MethodSignatureValidator {
	return &MethodSignatureValidator{
		typeChecker: tc,
		signatures:  make(map[string]*MethodSignature),
	}
}

// ValidateMethodSignatures processes all method annotations and validates their signatures
func (msv *MethodSignatureValidator) ValidateMethodSignatures(annotations []*ast.TypeAnnotation) []error {
	var errs []error
	methodSignatures := make(map[string]*MethodSignature)

	// First pass: collect all method signatures
	for _, annotation := range annotations {
		if annotation.Kind == ast.MethodReturnAnnotation {
			methodName := annotation.AnnotatedItem
			if sig, exists := methodSignatures[methodName]; exists {
				sig.ReturnType = annotation.TypeExpression
			} else {
				methodSignatures[methodName] = &MethodSignature{
					Name:       methodName,
					ReturnType: annotation.TypeExpression,
					Parameters: []MethodParameter{},
					SourcePos:  annotation.Pos,
				}
			}
		}
	}

	// Second pass: collect method parameters
	for _, annotation := range annotations {
		if annotation.Kind == ast.MethodParamAnnotation {
			// Find the method this parameter belongs to
			// For now, we'll associate parameters with the most recently declared method
			// A more sophisticated approach would parse the full method context
			if len(methodSignatures) > 0 {
				// Find the appropriate method signature - this is simplified
				// In practice, we'd need better context from the parser
				for _, sig := range methodSignatures {
					param := MethodParameter{
						Name:     annotation.AnnotatedItem,
						Type:     annotation.TypeExpression,
						Position: len(sig.Parameters),
					}
					sig.Parameters = append(sig.Parameters, param)
					break // For now, just add to first method
				}
			}
		}
	}

	// Third pass: validate each signature
	for _, signature := range methodSignatures {
		if validationErrs := msv.validateSignature(signature); len(validationErrs) > 0 {
			errs = append(errs, validationErrs...)
		}
		msv.signatures[signature.Name] = signature
	}

	return errs
}

// validateSignature validates a single method signature
func (msv *MethodSignatureValidator) validateSignature(sig *MethodSignature) []error {
	var errs []error

	// Validate return type if present
	if sig.ReturnType != nil {
		if err := msv.validateTypeExpression(sig.ReturnType, sig.Name+" return type"); err != nil {
			errs = append(errs, err)
		}
	}

	// Validate parameter types
	for i, param := range sig.Parameters {
		if err := msv.validateTypeExpression(param.Type, fmt.Sprintf("%s parameter %d (%s)", sig.Name, i+1, param.Name)); err != nil {
			errs = append(errs, err)
		}
	}

	// Validate parameter ordering (optional parameters must come after required ones)
	foundOptional := false
	for _, param := range sig.Parameters {
		if param.IsOptional {
			foundOptional = true
		} else if foundOptional {
			errs = append(errs, errors.NewPSC103Error(
				fmt.Sprintf("Required parameter %s cannot follow optional parameters in method %s", param.Name, sig.Name),
				fmt.Sprintf("%d:%d", sig.SourcePos.Line, sig.SourcePos.Column)))
		}
	}

	return errs
}

// validateTypeExpression validates that a type expression is well-formed and resolvable
func (msv *MethodSignatureValidator) validateTypeExpression(typeExpr *ast.TypeExpression, context string) error {
	if typeExpr == nil {
		return errors.NewPSC103Error(fmt.Sprintf("Missing type expression for %s", context), "")
	}

	// Validate base type
	if typeExpr.Name == "" {
		return errors.NewPSC103Error(fmt.Sprintf("Empty type name for %s", context), "")
	}

	// Check for known built-in types
	builtInTypes := map[string]bool{
		"Int":      true,
		"Str":      true,
		"Bool":     true,
		"Num":      true,
		"ArrayRef": true,
		"HashRef":  true,
		"CodeRef":  true,
		"Ref":      true,
		"Undef":    true,
		"Any":      true,
		"Item":     true,
		"Value":    true,
		"Defined":  true,
	}

	baseType := typeExpr.Name
	if len(typeExpr.Parameters) > 0 {
		// For parameterized types like ArrayRef[Int], extract the base type
		baseType = strings.Split(typeExpr.Name, "[")[0]
	}

	// Allow class names (anything starting with uppercase) or built-in types
	if !builtInTypes[baseType] && !isValidClassName(baseType) {
		return errors.NewPSC102Error(fmt.Sprintf("Unknown type '%s' in %s", baseType, context), "")
	}

	// Validate parameterized types
	if len(typeExpr.Parameters) > 0 {
		// Validate that parameterized types have valid base types
		parameterizableTypes := map[string]bool{
			"ArrayRef": true,
			"HashRef":  true,
			"Ref":      true,
		}

		if !parameterizableTypes[baseType] {
			return errors.NewPSC103Error(fmt.Sprintf("Type '%s' cannot be parameterized in %s", baseType, context), "")
		}

		// Recursively validate parameter types
		for i, param := range typeExpr.Parameters {
			if err := msv.validateTypeExpression(param, fmt.Sprintf("%s parameter %d", context, i+1)); err != nil {
				return err
			}
		}
	}

	// Validate union types
	if typeExpr.IsUnion && len(typeExpr.UnionTypes) > 0 {
		for i, unionType := range typeExpr.UnionTypes {
			if err := msv.validateTypeExpression(unionType, fmt.Sprintf("%s union type %d", context, i+1)); err != nil {
				return err
			}
		}
	}

	return nil
}

// isValidClassName checks if a string is a valid class name (starts with uppercase)
func isValidClassName(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

// GetMethodSignature returns the signature for a given method name
func (msv *MethodSignatureValidator) GetMethodSignature(methodName string) (*MethodSignature, bool) {
	sig, exists := msv.signatures[methodName]
	return sig, exists
}

// ValidateMethodCall validates a method call against its signature
func (msv *MethodSignatureValidator) ValidateMethodCall(methodName string, argTypes []*ast.TypeExpression, pos ast.Position) error {
	sig, exists := msv.signatures[methodName]
	if !exists {
		return errors.NewPSC102Error(fmt.Sprintf("Unknown method '%s'", methodName), fmt.Sprintf("%d:%d", pos.Line, pos.Column))
	}

	// Check argument count
	requiredParams := 0
	for _, param := range sig.Parameters {
		if !param.IsOptional {
			requiredParams++
		}
	}

	if len(argTypes) < requiredParams {
		return errors.NewPSC103Error(
			fmt.Sprintf("Method '%s' requires %d arguments, got %d", methodName, requiredParams, len(argTypes)),
			fmt.Sprintf("%d:%d", pos.Line, pos.Column))
	}

	if !sig.IsVariadic && len(argTypes) > len(sig.Parameters) {
		return errors.NewPSC103Error(
			fmt.Sprintf("Method '%s' accepts at most %d arguments, got %d", methodName, len(sig.Parameters), len(argTypes)),
			fmt.Sprintf("%d:%d", pos.Line, pos.Column))
	}

	// Validate argument types
	for i, argType := range argTypes {
		if i >= len(sig.Parameters) {
			break // Variadic arguments handled elsewhere
		}

		param := sig.Parameters[i]
		if !msv.isTypeCompatible(argType, param.Type) {
			return errors.NewPSC101Error(
				fmt.Sprintf("Method '%s' parameter %d (%s) expects type %s, got %s", methodName, i+1, param.Name, param.Type.String(), argType.String()),
				fmt.Sprintf("%d:%d", pos.Line, pos.Column))
		}
	}

	return nil
}

// isTypeCompatible checks if two types are compatible for assignment
func (msv *MethodSignatureValidator) isTypeCompatible(provided, expected *ast.TypeExpression) bool {
	if provided == nil || expected == nil {
		return false
	}

	// Exact match
	if provided.String() == expected.String() {
		return true
	}

	// Any type accepts everything
	if expected.Name == "Any" || expected.Name == "Item" {
		return true
	}

	// Undef is compatible with anything that can be undefined
	if provided.Name == "Undef" {
		return true
	}

	// More sophisticated type compatibility would go here
	// For now, keep it simple
	return false
}

// ValidateGenericMethodSignature validates method signatures with generic type parameters
func (msv *MethodSignatureValidator) ValidateGenericMethodSignature(sig *MethodSignature, typeParams []string) error {
	// Check that all generic type parameters are used
	usedParams := make(map[string]bool)

	// Check return type for generic parameters
	if sig.ReturnType != nil {
		msv.findGenericTypesInExpression(sig.ReturnType, typeParams, usedParams)
	}

	// Check parameter types for generic parameters
	for _, param := range sig.Parameters {
		msv.findGenericTypesInExpression(param.Type, typeParams, usedParams)
	}

	// Warn about unused generic parameters
	for _, typeParam := range typeParams {
		if !usedParams[typeParam] {
			// This would ideally be a warning, but for now we'll allow it
		}
	}

	return nil
}

// findGenericTypesInExpression recursively finds generic type parameters in a type expression
func (msv *MethodSignatureValidator) findGenericTypesInExpression(expr *ast.TypeExpression, typeParams []string, used map[string]bool) {
	if expr == nil {
		return
	}

	// Check if this type name is a generic parameter
	for _, param := range typeParams {
		if expr.Name == param {
			used[param] = true
		}
	}

	// Recursively check parameters
	for _, paramExpr := range expr.Parameters {
		msv.findGenericTypesInExpression(paramExpr, typeParams, used)
	}

	// Recursively check union types
	for _, unionExpr := range expr.UnionTypes {
		msv.findGenericTypesInExpression(unionExpr, typeParams, used)
	}
}
