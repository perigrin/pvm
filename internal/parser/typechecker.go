// ABOUTME: Type checking implementation for PSC
// ABOUTME: Validates type annotations in Perl code

package parser

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/typedef"
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

	// Debug enables debug mode
	Debug bool
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

// TypeErrorInfo provides detailed information about a type error
type TypeErrorInfo struct {
	// Message is the error message
	Message string

	// Path is the file path where the error occurred
	Path string

	// Position is the position in the file where the error occurred
	Position Position

	// SourceType is the source type that caused the error (if applicable)
	SourceType string

	// TargetType is the target type that caused the error (if applicable)
	TargetType string

	// Item is the variable, function, or other item involved in the error
	Item string

	// Kind is the kind of type error
	Kind string
}

// NewTypeChecker creates a new TypeChecker with the given type hierarchy
func NewTypeChecker(hierarchy *typedef.TypeHierarchy, moduleName string) *TypeChecker {
	return &TypeChecker{
		Hierarchy:       hierarchy,
		CurrentModule:   moduleName,
		ImportedModules: make(map[string]bool),
		TypeAnnotations: make(map[string]string),
		VariableTypes:   make(map[string]string),
		FunctionTypes:   make(map[string]*FunctionSignature),
		Debug:           false,
	}
}

// CheckAST performs type checking on an entire AST
func (tc *TypeChecker) CheckAST(ast *AST) []error {
	var checkErrors []error

	// Process imported modules
	tc.extractImports(ast)

	// First pass: collect all type annotations
	for _, annotation := range ast.TypeAnnotations {
		if err := tc.collectTypeAnnotation(annotation); err != nil {
			checkErrors = append(checkErrors, err)
		}
	}

	// Second pass: check type annotations
	for _, annotation := range ast.TypeAnnotations {
		if err := tc.checkTypeAnnotation(annotation); err != nil {
			checkErrors = append(checkErrors, err)
		}
	}

	// Third pass: check assignments and function calls
	assignmentErrors := tc.CheckASTAssignments(ast)
	checkErrors = append(checkErrors, assignmentErrors...)

	// Fourth pass: check function/method return types
	returnErrors := tc.checkASTFunctionReturns(ast)
	checkErrors = append(checkErrors, returnErrors...)

	return checkErrors
}

// collectTypeAnnotation collects a type annotation without full validation
func (tc *TypeChecker) collectTypeAnnotation(annotation *TypeAnnotation) error {
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
	case VarAnnotation:
		tc.VariableTypes[annotation.AnnotatedItem] = typeStr

	case SubParamAnnotation, MethodParamAnnotation:
		// Extract the function or method name from the parameter
		funcName := tc.extractFunctionNameFromParam(annotation.AnnotatedItem)
		if funcName == "" {
			// If we can't determine the function name, just continue
			break
		}

		// Create or update the function signature
		signature, exists := tc.FunctionTypes[funcName]
		if !exists {
			signature = &FunctionSignature{
				ParameterTypes: make(map[string]string),
				IsMethod:       annotation.Kind == MethodParamAnnotation,
			}
			tc.FunctionTypes[funcName] = signature
		}
		
		// Store the parameter type
		signature.ParameterTypes[annotation.AnnotatedItem] = typeStr

	case SubReturnAnnotation, MethodReturnAnnotation:
		// Extract the function or method name from the annotation
		funcName := tc.extractFunctionNameFromReturn(annotation.AnnotatedItem)
		if funcName == "" {
			// If we can't determine the function name, just continue
			break
		}

		// Create or update the function signature
		signature, exists := tc.FunctionTypes[funcName]
		if !exists {
			signature = &FunctionSignature{
				ParameterTypes: make(map[string]string),
				IsMethod:       annotation.Kind == MethodReturnAnnotation,
			}
			tc.FunctionTypes[funcName] = signature
		}
		
		// Store the return type
		signature.ReturnType = typeStr

	case AttrAnnotation:
		// For attribute annotations, store in both variable types and annotations
		tc.VariableTypes[annotation.AnnotatedItem] = typeStr

	case TypeDeclAnnotation:
		// Type declarations are handled separately
	}

	return nil
}

// extractFunctionNameFromParam extracts the function name from a parameter name
// This is a simplified implementation that would need to be enhanced in a real system
func (tc *TypeChecker) extractFunctionNameFromParam(paramName string) string {
	// This is a placeholder implementation
	// In a real system, this would use the AST to determine the function name
	// For now, just return an empty string as a placeholder
	return ""
}

// extractFunctionNameFromReturn extracts the function name from a return annotation
// This is a simplified implementation that would need to be enhanced in a real system
func (tc *TypeChecker) extractFunctionNameFromReturn(returnName string) string {
	// This is a placeholder implementation
	// In a real system, this would use the AST to determine the function name
	// For now, just return an empty string as a placeholder
	return ""
}

// extractImports identifies imported modules in the AST
func (tc *TypeChecker) extractImports(ast *AST) {
	// Scan the AST for use statements
	// This is a simplified implementation that would need to be expanded
	// for a real parser that properly captures use statements
	
	// Start by analyzing nodes in the tree to find use statements
	if ast.Root != nil {
		tc.processNodeForImports(ast.Root)
	}
}

// processNodeForImports processes a node and its children to find imports
func (tc *TypeChecker) processNodeForImports(node Node) {
	// Look for use statements
	if node.Type() == "use_statement" {
		// Extract module name from the use statement
		nodeText := getNodeText(node)
		parts := strings.Fields(nodeText)
		if len(parts) >= 2 && parts[0] == "use" {
			// Basic handling of "use Module" statements
			moduleName := parts[1]
			// Remove trailing semicolon if present
			moduleName = strings.TrimSuffix(moduleName, ";")
			// Record the imported module
			tc.ImportedModules[moduleName] = true
		}
	}
	
	// Process children recursively
	for _, child := range node.Children() {
		tc.processNodeForImports(child)
	}
}

// getNodeText extracts the text from a node
func getNodeText(node Node) string {
	if textNode, ok := node.(*TreeSitterNode); ok {
		return textNode.Text
	}
	return ""
}

// checkTypeAnnotation validates a single type annotation
func (tc *TypeChecker) checkTypeAnnotation(annotation *TypeAnnotation) error {
	// First, validate that the type exists and is valid
	typeExpr := annotation.TypeExpression
	if typeExpr == nil {
		return errors.NewTypeError(
			ErrTypeValidationError,
			fmt.Sprintf("Missing type expression for %s", annotation.AnnotatedItem),
			nil,
		)
	}
	
	// Convert type expression to string form for validation
	typeStr := typeExpr.String()
	
	// Validate the type
	err := tc.validateType(typeStr)
	if err != nil {
		posInfo := fmt.Sprintf(" at line %d, column %d", 
			annotation.Pos.Line, annotation.Pos.Column)
		return errors.NewTypeError(
			ErrTypeValidationError,
			fmt.Sprintf("Invalid type %s for %s%s", 
				typeStr, annotation.AnnotatedItem, posInfo),
			err,
		)
	}
	
	// Additional checks based on annotation kind
	switch annotation.Kind {
	case VarAnnotation:
		// For variable annotations, we check if there's an assignment
		// that conflicts with the annotation
		return tc.checkVariableAnnotation(annotation)
		
	case SubParamAnnotation, MethodParamAnnotation:
		// For parameter annotations, check that the parameter type is valid
		return tc.checkParameterAnnotation(annotation)
		
	case SubReturnAnnotation, MethodReturnAnnotation:
		// For return annotations, check that the return type is valid
		return tc.checkReturnAnnotation(annotation)
		
	case AttrAnnotation:
		// For attribute annotations, check that the attribute type is valid
		return tc.checkAttributeAnnotation(annotation)
		
	case TypeDeclAnnotation:
		// For type declarations, check that the type definition is valid
		return tc.checkTypeDeclarationAnnotation(annotation)
	}
	
	return nil
}

// checkVariableAnnotation checks a variable type annotation
func (tc *TypeChecker) checkVariableAnnotation(annotation *TypeAnnotation) error {
	// In a full implementation, we would check for any conflicting assignments
	// or uses of the variable that don't match the declared type.
	// For now, we'll just accept the annotation as valid since we've already
	// validated the type itself.
	
	// Record the type for this variable
	tc.VariableTypes[annotation.AnnotatedItem] = annotation.TypeExpression.String()
	
	return nil
}

// checkParameterAnnotation checks a function parameter type annotation
func (tc *TypeChecker) checkParameterAnnotation(annotation *TypeAnnotation) error {
	// In a full implementation, we would check that the parameter type
	// is compatible with how the function is called.
	// For now, we'll just accept the annotation as valid.
	
	return nil
}

// checkReturnAnnotation checks a function return type annotation
func (tc *TypeChecker) checkReturnAnnotation(annotation *TypeAnnotation) error {
	// In a full implementation, we would check that all return statements
	// in the function return compatible types.
	// For now, we'll just accept the annotation as valid.
	
	return nil
}

// checkAttributeAnnotation checks an attribute type annotation
func (tc *TypeChecker) checkAttributeAnnotation(annotation *TypeAnnotation) error {
	// In a full implementation, we would check that any assignments to
	// the attribute are compatible with the declared type.
	// For now, we'll just accept the annotation as valid.
	
	// Record the type for this attribute
	tc.VariableTypes[annotation.AnnotatedItem] = annotation.TypeExpression.String()
	
	return nil
}

// checkTypeDeclarationAnnotation checks a type declaration annotation
func (tc *TypeChecker) checkTypeDeclarationAnnotation(annotation *TypeAnnotation) error {
	// In a full implementation, we would check that the type declaration
	// is valid and doesn't conflict with existing types.
	// For now, we'll just accept the annotation as valid.
	
	return nil
}

// validateType checks if a type is valid
func (tc *TypeChecker) validateType(typeName string) error {
	// Handle union types
	if strings.Contains(typeName, "|") {
		parts := strings.Split(typeName, "|")
		for _, part := range parts {
			if err := tc.validateType(strings.TrimSpace(part)); err != nil {
				return err
			}
		}
		return nil
	}
	
	// Handle intersection types
	if strings.Contains(typeName, "&") {
		parts := strings.Split(typeName, "&")
		for _, part := range parts {
			if err := tc.validateType(strings.TrimSpace(part)); err != nil {
				return err
			}
		}
		return nil
	}
	
	// Handle negation types
	if strings.HasPrefix(typeName, "!") {
		return tc.validateType(typeName[1:])
	}
	
	// Use the type hierarchy to validate the type
	return tc.Hierarchy.ValidateType(typeName)
}

// CheckTypeCompatibility checks if two types are compatible
func (tc *TypeChecker) CheckTypeCompatibility(sourceType, targetType string) error {
	return tc.Hierarchy.CheckTypeCompatibility(sourceType, targetType)
}

// CheckAssignment checks if a value of sourceType can be assigned to a variable of targetType
func (tc *TypeChecker) CheckAssignment(sourceType, targetType string, pos Position) error {
	err := tc.CheckTypeCompatibility(sourceType, targetType)
	if err != nil {
		posInfo := fmt.Sprintf(" at line %d, column %d", pos.Line, pos.Column)
		return errors.NewTypeError(
			ErrTypeAnnotationMismatch,
			fmt.Sprintf("Cannot assign %s to %s%s", sourceType, targetType, posInfo),
			err,
		)
	}
	return nil
}

// GetAnnotatedType retrieves the type for an annotated item
func (tc *TypeChecker) GetAnnotatedType(item string) (string, bool) {
	typ, ok := tc.TypeAnnotations[item]
	return typ, ok
}

// GetVariableType retrieves the type for a variable
func (tc *TypeChecker) GetVariableType(variable string) (string, bool) {
	typ, ok := tc.VariableTypes[variable]
	return typ, ok
}

// GetFunctionSignature retrieves the signature for a function
func (tc *TypeChecker) GetFunctionSignature(funcName string) (*FunctionSignature, bool) {
	sig, ok := tc.FunctionTypes[funcName]
	return sig, ok
}

// CheckASTAssignments checks all assignments in an AST
// Note: This would require AST nodes for assignments, which our simplified
// parser doesn't yet create. This is a placeholder for future implementation.
func (tc *TypeChecker) CheckASTAssignments(ast *AST) []error {
	var errors []error
	
	// This would require traversing the AST looking for assignment nodes,
	// extracting the left-hand side and right-hand side, determining their types,
	// and checking compatibility.
	
	// For each assignment node in the AST:
	// 1. Determine the type of the left-hand side (variable)
	// 2. Determine the type of the right-hand side (expression)
	// 3. Check if the right-hand side type is compatible with the left-hand side type
	// 4. If not, add an error
	
	// Process assignment nodes in the tree
	if ast.Root != nil {
		tc.checkNodeForAssignments(ast.Root, &errors)
	}
	
	return errors
}

// checkNodeForAssignments checks a node and its children for assignments
func (tc *TypeChecker) checkNodeForAssignments(node Node, errors *[]error) {
	// Look for assignment statements
	if node.Type() == "expression_statement" {
		nodeText := getNodeText(node)
		
		// Very basic assignment detection - in a real parser, we'd have proper AST nodes for assignments
		if strings.Contains(nodeText, "=") && !strings.Contains(nodeText, "==") {
			// This might be an assignment - try to parse it
			tc.checkPossibleAssignment(nodeText, node.Start(), errors)
		}
	}
	
	// Process children recursively
	for _, child := range node.Children() {
		tc.checkNodeForAssignments(child, errors)
	}
}

// checkPossibleAssignment checks a potential assignment statement
func (tc *TypeChecker) checkPossibleAssignment(text string, pos Position, errors *[]error) {
	// Split the text by = to get left and right sides
	parts := strings.SplitN(text, "=", 2)
	if len(parts) != 2 {
		return // Not a simple assignment
	}
	
	leftSide := strings.TrimSpace(parts[0])
	rightSide := strings.TrimSpace(parts[1])
	
	// Check if the left side is a variable we know the type of
	varType, ok := tc.GetVariableType(leftSide)
	if !ok {
		// We don't know the type of this variable, so we can't check the assignment
		return
	}
	
	// In a real implementation, we would analyze the right side expression to determine its type
	// For now, we'll use a simplified approach that only handles direct assignments of literals or variables
	
	// Infer the type of the right side (very simplified)
	rightType := tc.inferExpressionType(rightSide)
	if rightType == "" {
		// Couldn't determine the type
		return
	}
	
	// Check compatibility
	err := tc.CheckAssignment(rightType, varType, pos)
	if err != nil {
		*errors = append(*errors, err)
	}
}

// inferExpressionType infers the type of an expression (simplified implementation)
func (tc *TypeChecker) inferExpressionType(expr string) string {
	// Remove any trailing semicolon
	expr = strings.TrimSuffix(expr, ";")
	
	// First, check if this is a variable we know the type of
	if strings.HasPrefix(expr, "$") || strings.HasPrefix(expr, "@") || strings.HasPrefix(expr, "%") {
		if varType, ok := tc.GetVariableType(expr); ok {
			return varType
		}
	}
	
	// Check for literals
	if isNumericLiteral(expr) {
		// Is it an integer or float?
		if strings.Contains(expr, ".") {
			return "Float"
		}
		return "Int"
	}
	
	if isStringLiteral(expr) {
		return "Str"
	}
	
	if expr == "undef" {
		return "Undef"
	}
	
	if expr == "1" || expr == "0" || expr == "True" || expr == "False" {
		return "Bool"
	}
	
	// Check for array references
	if strings.HasPrefix(expr, "[") && strings.HasSuffix(expr, "]") {
		return "ArrayRef"
	}
	
	// Check for hash references
	if strings.HasPrefix(expr, "{") && strings.HasSuffix(expr, "}") {
		return "HashRef"
	}
	
	// For more complex expressions, we'd need to do proper expression parsing
	// For now, return an empty string to indicate we couldn't determine the type
	return ""
}

// isNumericLiteral checks if a string is a numeric literal
func isNumericLiteral(s string) bool {
	// Check if the string is a numeric literal
	// This is a simplified check - in a real implementation, we'd use a proper parser
	if len(s) == 0 {
		return false
	}
	
	// Allow leading minus sign for negative numbers
	startIndex := 0
	if s[0] == '-' {
		startIndex = 1
		if len(s) == 1 {
			return false // Just a minus sign
		}
	}
	
	// Check if it's a decimal number
	hasDecimal := false
	hasDigit := false
	
	for i := startIndex; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			hasDigit = true
		} else if s[i] == '.' && !hasDecimal {
			hasDecimal = true
		} else {
			return false // Not a numeric character
		}
	}
	
	return hasDigit
}

// isStringLiteral checks if a string is a string literal
func isStringLiteral(s string) bool {
	// Check if the string is a string literal
	// This is a simplified check - in a real implementation, we'd use a proper parser
	return (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) ||
		(strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "`") && strings.HasSuffix(s, "`"))
}

// checkASTFunctionReturns checks return types in an AST
func (tc *TypeChecker) checkASTFunctionReturns(ast *AST) []error {
	var errors []error
	
	// This would require traversing the AST looking for return statements,
	// determining the type of the returned expression, and checking compatibility
	// with the function's declared return type.
	
	// For each function in the AST:
	// 1. Find the function signature (if any)
	// 2. Find all return statements in the function
	// 3. For each return statement, check if the return type is compatible with the function signature
	// 4. If not, add an error
	
	// This is a simplified implementation that doesn't actually check anything yet
	// In a real system, we'd need to properly parse the AST
	
	return errors
}

// FormatTypeError formats a type error message with location information
func FormatTypeError(err error, pos Position, path string) string {
	return fmt.Sprintf("%s:%d:%d: %s", path, pos.Line, pos.Column, err.Error())
}