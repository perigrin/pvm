// ABOUTME: Type checking implementation for PSC
// ABOUTME: Validates type annotations in Perl code

package parser

import (
	"fmt"
	"sort"
	"strings"

	"tamarou.com/pvm/internal/errors"
)

// The NewTypeChecker function has been moved to interface.go

// The initializeValidationPatterns method has been moved to interface.go

// The CheckAST method has been moved to interface.go
// This is the real implementation of the method

// SkipFlowChecks controls whether to skip flow-sensitive type checks
// but still perform type refinements based on control flow
var SkipFlowChecks bool

// performFlowSensitiveAnalysis performs flow-sensitive type analysis on the AST
// This function is called by CheckAST in interface.go
// nolint:unused
func (tc *TypeChecker) performFlowSensitiveAnalysis(ast *AST) []error {
	var errors []error

	// Create a snapshot of the current variable types to initialize flow analysis
	for varName, varType := range tc.VariableTypes {
		if tc.TypeState != nil {
			tc.TypeState.VariableTypes[varName] = varType
		}
	}

	// Process the AST for flow-sensitive analysis
	if ast.Root != nil {
		tc.analyzeNodeFlow(ast.Root, &errors)
	}

	return errors
}

// analyzeNodeFlow performs flow-sensitive analysis on a node and its children
func (tc *TypeChecker) analyzeNodeFlow(node Node, errors *[]error) {
	// Different node types require different analysis
	switch node.Type() {
	case "if_statement":
		tc.analyzeIfStatement(node, errors)

	case "while_statement":
		tc.analyzeLoopStatement(node, errors)

	case "for_statement":
		tc.analyzeLoopStatement(node, errors)

	case "function_call":
		tc.analyzePatternValidation(node, errors)

	default:
		// Other node types may need custom analysis
		// Check for expressions that could be validation patterns
		tc.analyzePatternValidation(node, errors)

		// Process children recursively
		for _, child := range node.Children() {
			tc.analyzeNodeFlow(child, errors)
		}
	}
}

// analyzeIfStatement performs flow-sensitive analysis on an if statement
func (tc *TypeChecker) analyzeIfStatement(node Node, errors *[]error) {
	// Extract the condition node
	var conditionNode Node
	var thenBlock Node
	var elseBlock Node

	// In a real implementation, we would properly extract these nodes
	// For now, we'll use a simplified approach
	for _, child := range node.Children() {
		if child.Type() == "condition" {
			conditionNode = child
		} else if child.Type() == "block" && thenBlock == nil {
			thenBlock = child
		} else if child.Type() == "block" {
			elseBlock = child
		}
	}

	if conditionNode == nil {
		// Process the if statement normally without flow analysis
		for _, child := range node.Children() {
			tc.analyzeNodeFlow(child, errors)
		}
		return
	}

	// Analyze the condition for type refinements
	condition, refinements := tc.analyzeCondition(conditionNode)

	// Push the current state onto the stack
	if tc.TypeState != nil {
		tc.TypeStateStack = append(tc.TypeStateStack, tc.cloneTypeState(tc.TypeState))
	}

	// Apply refinements for the then branch
	tc.applyRefinements(refinements)

	// Process the then branch with the refined types
	if thenBlock != nil {
		tc.analyzeNodeFlow(thenBlock, errors)
	}

	// If there's an else branch, create a new state with the negated condition
	if elseBlock != nil {
		// Restore the original state
		if len(tc.TypeStateStack) > 0 {
			tc.TypeState = tc.TypeStateStack[len(tc.TypeStateStack)-1]
			tc.TypeStateStack = tc.TypeStateStack[:len(tc.TypeStateStack)-1]
		}

		// Push the current state for the else branch
		if tc.TypeState != nil {
			tc.TypeStateStack = append(tc.TypeStateStack, tc.cloneTypeState(tc.TypeState))
		}

		// Apply negated refinements for the else branch
		negatedRefinements := tc.negateRefinements(condition, refinements)
		tc.applyRefinements(negatedRefinements)

		// Process the else branch
		tc.analyzeNodeFlow(elseBlock, errors)
	}

	// Restore the original state
	if len(tc.TypeStateStack) > 0 {
		tc.TypeState = tc.TypeStateStack[len(tc.TypeStateStack)-1]
		tc.TypeStateStack = tc.TypeStateStack[:len(tc.TypeStateStack)-1]
	}

	// Process any siblings after the if statement
	for _, child := range node.Children() {
		if child != conditionNode && child != thenBlock && child != elseBlock {
			tc.analyzeNodeFlow(child, errors)
		}
	}
}

// analyzeLoopStatement performs flow-sensitive analysis on a loop statement
func (tc *TypeChecker) analyzeLoopStatement(node Node, errors *[]error) {
	// Extract the condition node and body block
	var conditionNode Node
	var bodyBlock Node

	// In a real implementation, we would properly extract these nodes
	// For now, we'll use a simplified approach
	for _, child := range node.Children() {
		if child.Type() == "condition" {
			conditionNode = child
		} else if child.Type() == "block" {
			bodyBlock = child
		}
	}

	if conditionNode == nil {
		// Process the loop statement normally without flow analysis
		for _, child := range node.Children() {
			tc.analyzeNodeFlow(child, errors)
		}
		return
	}

	// Analyze the condition for type refinements
	// We're intentionally ignoring the refinements for loops as we're being conservative
	_, _ = tc.analyzeCondition(conditionNode)

	// For loops, we're more conservative with refinements
	// We push the current state but don't apply refinements automatically
	if tc.TypeState != nil {
		tc.TypeStateStack = append(tc.TypeStateStack, tc.cloneTypeState(tc.TypeState))
		// In a more sophisticated implementation, we might selectively apply refinements
	}

	// Process the loop body
	if bodyBlock != nil {
		tc.analyzeNodeFlow(bodyBlock, errors)
	}

	// Restore the original state
	if len(tc.TypeStateStack) > 0 {
		tc.TypeState = tc.TypeStateStack[len(tc.TypeStateStack)-1]
		tc.TypeStateStack = tc.TypeStateStack[:len(tc.TypeStateStack)-1]
	}

	// Process any siblings
	for _, child := range node.Children() {
		if child != conditionNode && child != bodyBlock {
			tc.analyzeNodeFlow(child, errors)
		}
	}
}

// analyzePatternValidation checks if a node matches any validation pattern
func (tc *TypeChecker) analyzePatternValidation(node Node, errors *[]error) {
	// Check against all validation patterns
	for _, pattern := range tc.ValidationPatterns {
		if pattern.Checker != nil {
			if varName, matches := pattern.Checker(node); matches {
				// The node matches this validation pattern
				// Apply the refinement function to the variable
				if currentType, ok := tc.GetVariableType(varName); ok && pattern.RefinementFunc != nil {
					refinedType := pattern.RefinementFunc(varName, currentType)

					// Update the refined type in the current state
					if tc.TypeState != nil {
						tc.TypeState.RefinedTypes[varName] = refinedType
					}
				}
			}
		}
	}

	// Process children recursively
	for _, child := range node.Children() {
		tc.analyzeNodeFlow(child, errors)
	}
}

// analyzeCondition extracts conditions and type refinements from a condition node
func (tc *TypeChecker) analyzeCondition(node Node) (Condition, map[string]string) {
	refinements := make(map[string]string)
	var condition Condition

	// Extract conditions based on node type
	switch node.Type() {
	case "eq_expression", "ne_expression", "gt_expression", "lt_expression",
		"ge_expression", "le_expression":
		// Parse comparison expressions
		text := getNodeText(node)
		parts := strings.Split(text, getOperatorFromNodeType(node.Type()))

		if len(parts) >= 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			// Set up the condition
			condition = Condition{
				Variable: left,
				Operator: getOperatorFromNodeType(node.Type()),
				Value:    right,
				Negated:  node.Type() == "ne_expression",
			}

			// Try to refine type based on the condition
			if strings.HasPrefix(left, "$") {
				// Variable on the left
				refinements[left] = tc.refineTypeFromComparison(left, right, node.Type())
			} else if strings.HasPrefix(right, "$") {
				// Variable on the right
				refinements[right] = tc.refineTypeFromComparison(right, left, reverseComparisonType(node.Type()))
			}
		}

	case "defined_expression":
		// Handle defined($var) expressions
		text := getNodeText(node)
		if strings.HasPrefix(text, "defined(") && strings.HasSuffix(text, ")") {
			varName := strings.TrimPrefix(text, "defined(")
			varName = strings.TrimSuffix(varName, ")")
			varName = strings.TrimSpace(varName)

			condition = Condition{
				Variable: varName,
				Operator: "defined",
				Value:    "",
				Negated:  false,
			}

			// Refine the type
			if currentType, ok := tc.GetVariableType(varName); ok {
				if strings.HasPrefix(currentType, "Maybe[") {
					baseType, params := ExtractTypeAndParams(currentType)
					if baseType == "Maybe" && len(params) > 0 {
						refinements[varName] = params[0]
					}
				}
			}
		}

	case "ref_expression":
		// Handle ref($var) eq 'ARRAY' expressions
		// This would need more complex parsing in a real implementation

	case "isa_expression":
		// Handle $var->isa('Class') expressions
		// This would need more complex parsing in a real implementation
	}

	return condition, refinements
}

// refineTypeFromComparison refines a type based on a comparison expression
func (tc *TypeChecker) refineTypeFromComparison(varName, value, nodeType string) string {
	currentType, ok := tc.GetVariableType(varName)
	if !ok {
		return ""
	}

	// Specific refinements based on comparison type and value
	switch {
	case nodeType == "eq_expression" && value == "undef":
		// $var == undef -> Maybe[T] remains Maybe[T]
		return currentType

	case nodeType == "ne_expression" && value == "undef":
		// $var != undef can refine Maybe[T] to T
		if strings.HasPrefix(currentType, "Maybe[") {
			baseType, params := ExtractTypeAndParams(currentType)
			if baseType == "Maybe" && len(params) > 0 {
				return params[0]
			}
		}

	case (nodeType == "eq_expression" && value == "''" || value == "\"\"") && currentType == "Str":
		// Empty string check - no refinement needed
		return currentType

	case nodeType == "eq_expression" && isStringLiteral(value) && currentType == "Str":
		// Specific string value - we could refine to a more specific type if needed
		return currentType

	case (nodeType == "gt_expression" || nodeType == "lt_expression" ||
		nodeType == "ge_expression" || nodeType == "le_expression") && isNumericLiteral(value):
		// Numeric comparison - no refinement needed for basic types
		return currentType
	}

	// Default: no refinement
	return currentType
}

// getOperatorFromNodeType converts a node type to an operator string
func getOperatorFromNodeType(nodeType string) string {
	switch nodeType {
	case "eq_expression":
		return "=="
	case "ne_expression":
		return "!="
	case "gt_expression":
		return ">"
	case "lt_expression":
		return "<"
	case "ge_expression":
		return ">="
	case "le_expression":
		return "<="
	default:
		return ""
	}
}

// reverseComparisonType reverses a comparison type (for swapping operands)
func reverseComparisonType(nodeType string) string {
	switch nodeType {
	case "gt_expression":
		return "lt_expression"
	case "lt_expression":
		return "gt_expression"
	case "ge_expression":
		return "le_expression"
	case "le_expression":
		return "ge_expression"
	default:
		return nodeType // eq and ne are symmetric
	}
}

// cloneTypeState creates a deep copy of a type state
func (tc *TypeChecker) cloneTypeState(state *TypeState) *TypeState {
	if state == nil {
		return nil
	}

	clone := &TypeState{
		VariableTypes: make(map[string]string),
		RefinedTypes:  make(map[string]string),
		Conditions:    make([]Condition, len(state.Conditions)),
	}

	// Copy variable types
	for k, v := range state.VariableTypes {
		clone.VariableTypes[k] = v
	}

	// Copy refined types
	for k, v := range state.RefinedTypes {
		clone.RefinedTypes[k] = v
	}

	// Copy conditions
	copy(clone.Conditions, state.Conditions)

	return clone
}

// applyRefinements applies a set of type refinements to the current state
func (tc *TypeChecker) applyRefinements(refinements map[string]string) {
	if tc.TypeState == nil || len(refinements) == 0 {
		return
	}

	// Apply each refinement
	for varName, refinedType := range refinements {
		if refinedType != "" {
			tc.TypeState.RefinedTypes[varName] = refinedType
		}
	}
}

// negateRefinements creates negated refinements for the else branch
func (tc *TypeChecker) negateRefinements(condition Condition, refinements map[string]string) map[string]string {
	negatedRefinements := make(map[string]string)

	// For simple defined checks, negation is straightforward
	if condition.Operator == "defined" {
		// For defined($var), the negation means the variable is definitely undef
		// We keep Maybe[T] as is, since it already indicates the variable could be undef
		return negatedRefinements
	}

	// For equality checks with undef, negation can refine Maybe[T] to T
	if condition.Operator == "==" && condition.Value == "undef" {
		currentType, ok := tc.GetVariableType(condition.Variable)
		if ok && strings.HasPrefix(currentType, "Maybe[") {
			baseType, params := ExtractTypeAndParams(currentType)
			if baseType == "Maybe" && len(params) > 0 {
				negatedRefinements[condition.Variable] = params[0]
			}
		}
	}

	// For more complex conditions, we would need more sophisticated analysis
	// This is simplified for now

	return negatedRefinements
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
// This function is referenced by CheckAST but not used directly in the current implementation
// nolint:unused
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
// This function is called by extractImports but not used directly in the current implementation
// nolint:unused
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

// The getNodeText function has been moved to interface.go

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

// The ExtractTypeAndParams function is defined in interface.go

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

// CheckStrictParameterTypeCompatibility provides stricter type checking for function parameters
// This allows contravariant parameter typing - parameters can accept supertypes but not arbitrary subtypes
func (tc *TypeChecker) CheckStrictParameterTypeCompatibility(argType, paramType string) error {
	// Exact type match is always allowed
	if argType == paramType {
		return nil
	}

	// For function parameters, we want strict typing with controlled coercion
	// Rather than allowing contravariance, we'll explicitly allow safe conversions below

	// In practice, we'll be more permissive and allow some safe coercions
	// Allow Int -> Num (numeric widening is safe)
	if paramType == "Num" && argType == "Int" {
		return nil
	}

	// Allow Num -> Scalar, Int -> Scalar, Str -> Scalar (all safe)
	if paramType == "Scalar" && (argType == "Int" || argType == "Num" || argType == "Str") {
		return nil
	}

	// Allow anything -> Any
	if paramType == "Any" {
		return nil
	}

	// Special handling for Maybe types
	if strings.HasPrefix(paramType, "Maybe[") {
		_, params := ExtractTypeAndParams(paramType)
		if len(params) > 0 {
			// Maybe[T] parameter can accept T or Maybe[T]
			if argType == params[0] {
				return nil
			}
			return tc.CheckStrictParameterTypeCompatibility(argType, params[0])
		}
	}

	// For everything else, types must be compatible through the standard hierarchy
	// but we reject conversions that would lose information or require implicit conversion
	
	// Reject numeric -> string conversions (would be coercion)
	if argType == "Int" && (paramType == "Str") {
		return fmt.Errorf("cannot pass %s to parameter expecting %s (implicit conversion not allowed)", argType, paramType)
	}
	if argType == "Num" && (paramType == "Str") {
		return fmt.Errorf("cannot pass %s to parameter expecting %s (implicit conversion not allowed)", argType, paramType)
	}

	// Reject string -> numeric conversions (would be parsing/coercion)
	if argType == "Str" && (paramType == "Int" || paramType == "Num") {
		return fmt.Errorf("cannot pass %s to parameter expecting %s (implicit conversion not allowed)", argType, paramType)
	}

	// Reject other narrowing conversions that could lose data
	if argType == "Scalar" && (paramType == "Int" || paramType == "Num" || paramType == "Str") {
		return fmt.Errorf("cannot pass %s to parameter expecting %s (narrowing conversion not allowed)", argType, paramType)
	}

	// Use standard type compatibility for remaining cases (like inheritance hierarchies)
	return tc.CheckTypeCompatibility(argType, paramType)
}

// CheckStrictReturnTypeCompatibility provides stricter type checking for function return types
func (tc *TypeChecker) CheckStrictReturnTypeCompatibility(returnType, expectedReturnType string) error {
	// Exact type match is always allowed
	if returnType == expectedReturnType {
		return nil
	}

	// For return types, we allow covariant subtyping BUT only for safe, non-conversion subtypes
	// We'll be explicit about what's allowed rather than using the full hierarchy

	// Allow safe numeric widening: Int -> Num, Int -> Scalar, Num -> Scalar
	if expectedReturnType == "Num" && returnType == "Int" {
		return nil
	}
	if expectedReturnType == "Scalar" && (returnType == "Int" || returnType == "Num" || returnType == "Str") {
		return nil
	}
	if expectedReturnType == "Any" {
		return nil // anything can be returned as Any
	}

	// Special handling for Maybe types
	if strings.HasPrefix(expectedReturnType, "Maybe[") {
		_, params := ExtractTypeAndParams(expectedReturnType)
		if len(params) > 0 {
			// Can return T for Maybe[T]
			if returnType == params[0] {
				return nil
			}
			return tc.CheckStrictReturnTypeCompatibility(returnType, params[0])
		}
	}

	// Reject conversions that require implicit conversion or could lose data
	// These are the opposite rules from parameter checking

	// Reject numeric -> string conversions for return types too (unless overridden by hierarchy)
	if returnType == "Int" && expectedReturnType == "Str" {
		return fmt.Errorf("cannot return %s as %s (implicit conversion not allowed)", returnType, expectedReturnType)
	}
	if returnType == "Num" && expectedReturnType == "Str" {
		return fmt.Errorf("cannot return %s as %s (implicit conversion not allowed)", returnType, expectedReturnType)
	}

	// Reject string -> numeric conversions 
	if returnType == "Str" && (expectedReturnType == "Int" || expectedReturnType == "Num") {
		return fmt.Errorf("cannot return %s as %s (implicit conversion not allowed)", returnType, expectedReturnType)
	}

	// For other mismatches, return a generic incompatibility error
	return fmt.Errorf("return type %s is not compatible with expected %s", returnType, expectedReturnType)
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
	// First check for refined types in the current state
	if tc.TypeState != nil {
		if refinedType, ok := tc.TypeState.RefinedTypes[variable]; ok {
			return refinedType, true
		}
	}

	// Fall back to global variable types
	typ, ok := tc.VariableTypes[variable]
	return typ, ok
}

// GetFunctionSignature retrieves the signature for a function
func (tc *TypeChecker) GetFunctionSignature(funcName string) (*FunctionSignature, bool) {
	sig, ok := tc.FunctionTypes[funcName]
	return sig, ok
}

// CheckFunctionCall validates a function call against its signature (Phase 5)
func (tc *TypeChecker) CheckFunctionCall(funcName string, argTypes []string) error {
	// Get the function signature
	signature, ok := tc.GetFunctionSignature(funcName)
	if !ok {
		return fmt.Errorf("unknown function: %s", funcName)
	}

	// Check parameter count
	expectedCount := len(signature.ParameterTypes)
	actualCount := len(argTypes)
	if actualCount != expectedCount {
		return fmt.Errorf("function %s expects %d parameters, got %d", funcName, expectedCount, actualCount)
	}

	// Validate parameter types
	return tc.ValidateParameterTypes(funcName, signature, argTypes)
}

// ValidateParameterTypes checks that argument types are compatible with parameter types (Phase 5)
func (tc *TypeChecker) ValidateParameterTypes(funcName string, signature *FunctionSignature, argTypes []string) error {
	// For this simplified implementation, we'll assume the parameter types map
	// is ordered consistently. In a real implementation, we'd use AST parameter order.
	
	// Check parameter count
	if len(argTypes) != len(signature.ParameterTypes) {
		return fmt.Errorf("parameter count mismatch for function %s: expected %d, got %d", 
			funcName, len(signature.ParameterTypes), len(argTypes))
	}

	// Convert parameter types map to ordered slice for comparison
	// For methods, we need to ensure $self comes first, then others alphabetically
	var paramNames []string
	var hasSelfs bool
	
	// Check if this is a method (has $self parameter)
	for paramName := range signature.ParameterTypes {
		if paramName == "$self" {
			hasSelfs = true
			break
		}
	}
	
	if hasSelfs {
		// For methods, put $self first, then sort the rest
		paramNames = append(paramNames, "$self")
		for paramName := range signature.ParameterTypes {
			if paramName != "$self" {
				paramNames = append(paramNames, paramName)
			}
		}
		// Sort the non-self parameters
		if len(paramNames) > 1 {
			sort.Strings(paramNames[1:])
		}
	} else {
		// For functions, just sort all parameters alphabetically
		for paramName := range signature.ParameterTypes {
			paramNames = append(paramNames, paramName)
		}
		sort.Strings(paramNames)
	}

	// Check each argument type against expected parameter type
	for i, argType := range argTypes {
		if i >= len(paramNames) {
			return fmt.Errorf("too many arguments for function %s", funcName)
		}

		paramName := paramNames[i]
		expectedType := signature.ParameterTypes[paramName]
		
		// Check strict type compatibility for function parameters
		err := tc.CheckStrictParameterTypeCompatibility(argType, expectedType)
		if err != nil {
			return fmt.Errorf("parameter %d (%s) of function %s: incompatible type %s (expected %s): %w", 
				i+1, paramName, funcName, argType, expectedType, err)
		}
	}

	return nil
}

// CheckReturnType validates that a return expression type matches the function's return type (Phase 5)
func (tc *TypeChecker) CheckReturnType(funcName string, returnType string) error {
	// Get the function signature
	signature, ok := tc.GetFunctionSignature(funcName)
	if !ok {
		return fmt.Errorf("unknown function: %s", funcName)
	}

	// Check strict return type compatibility
	err := tc.CheckStrictReturnTypeCompatibility(returnType, signature.ReturnType)
	if err != nil {
		return fmt.Errorf("function %s return type: incompatible type %s (expected %s): %w", 
			funcName, returnType, signature.ReturnType, err)
	}

	return nil
}

// CheckSubroutineCall validates a subroutine call (Phase 5)
func (tc *TypeChecker) CheckSubroutineCall(subName string, argTypes []string) error {
	// Get the function signature
	signature, ok := tc.GetFunctionSignature(subName)
	if !ok {
		return fmt.Errorf("unknown subroutine: %s", subName)
	}

	// Check that it's actually a subroutine, not a method
	if signature.IsMethod {
		return fmt.Errorf("%s is not a subroutine (it's a method)", subName)
	}

	// Validate the call using standard function call validation
	return tc.CheckFunctionCall(subName, argTypes)
}

// CheckMethodCall validates a method call (Phase 5)
func (tc *TypeChecker) CheckMethodCall(className string, methodName string, argTypes []string) error {
	// Construct the full method name (Class::method convention)
	fullMethodName := fmt.Sprintf("%s::%s", className, methodName)
	
	// Get the method signature
	signature, ok := tc.GetFunctionSignature(fullMethodName)
	if !ok {
		// Also try just the method name without class prefix
		signature, ok = tc.GetFunctionSignature(methodName)
		if !ok {
			return fmt.Errorf("unknown method: %s::%s", className, methodName)
		}
	}

	// Check that it's actually a method, not a subroutine
	if !signature.IsMethod {
		return fmt.Errorf("%s is not a method (it's a subroutine)", methodName)
	}

	// Check parameter count (including self parameter for methods)
	expectedCount := len(signature.ParameterTypes)
	actualCount := len(argTypes)
	if actualCount != expectedCount {
		return fmt.Errorf("method %s::%s expects %d parameters, got %d", className, methodName, expectedCount, actualCount)
	}

	// Validate parameter types
	return tc.ValidateParameterTypes(fullMethodName, signature, argTypes)
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

	// Check if it's a decimal number or scientific notation
	hasDecimal := false
	hasDigit := false
	hasExponent := false

	for i := startIndex; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			hasDigit = true
		} else if c == '.' && !hasDecimal && !hasExponent {
			hasDecimal = true
		} else if (c == 'e' || c == 'E') && !hasExponent && hasDigit {
			hasExponent = true
			// Check if next character is a sign or digit
			if i+1 < len(s) {
				next := s[i+1]
				if next == '+' || next == '-' {
					i++ // Skip the sign
				}
			}
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

// inferLiteralType infers the most specific type for a literal value
// This implements Phase 1: Pure Type Inference for literals
func (tc *TypeChecker) inferLiteralType(literal string) string {
	// Handle empty input
	if literal == "" {
		return ""
	}

	// Handle undef literal
	if literal == "undef" {
		return "Undef"
	}

	// Handle boolean literals
	if literal == "True" || literal == "False" {
		return "Bool"
	}

	// Handle string literals - check this before numeric to avoid false positives
	if isStringLiteral(literal) {
		return "Str"
	}

	// Handle numeric literals
	if isNumericLiteral(literal) {
		// Check if it's a float (contains decimal point or scientific notation)
		if strings.Contains(literal, ".") || strings.Contains(strings.ToLower(literal), "e") {
			return "Float"
		}
		// Otherwise it's an integer
		return "Int"
	}

	// If we can't identify the literal type, return empty string
	return ""
}

// InferBinaryOperatorType infers the result type of a binary operator expression
// This implements Phase 4: Operator Type Checking
func (tc *TypeChecker) InferBinaryOperatorType(leftType, operator, rightType string) (string, error) {
	// First, check that both operand types are valid
	if err := tc.validateType(leftType); err != nil {
		return "", fmt.Errorf("invalid left operand type '%s': %w", leftType, err)
	}
	if err := tc.validateType(rightType); err != nil {
		return "", fmt.Errorf("invalid right operand type '%s': %w", rightType, err)
	}

	switch operator {
	// Numeric arithmetic operators: +, -, *, /, %, **
	case "+", "-", "*", "/", "%", "**":
		return tc.inferArithmeticOperatorType(leftType, operator, rightType)

	// Numeric comparison operators: ==, !=, <, >, <=, >=, <=>
	case "==", "!=", "<", ">", "<=", ">=", "<=>":
		return tc.inferNumericComparisonType(leftType, operator, rightType)

	// String operators: . (concatenation)
	case ".":
		return tc.inferStringConcatenationType(leftType, operator, rightType)

	// String comparison operators: eq, ne, lt, gt, le, ge, cmp
	case "eq", "ne", "lt", "gt", "le", "ge", "cmp":
		return tc.inferStringComparisonType(leftType, operator, rightType)

	// Logical operators: &&, ||, //
	case "&&", "||", "//":
		return tc.inferLogicalOperatorType(leftType, operator, rightType)

	// Pattern matching operators: =~, !~
	case "=~", "!~":
		return tc.inferPatternMatchType(leftType, operator, rightType)

	// Bitwise operators: &, |, ^, <<, >>
	case "&", "|", "^", "<<", ">>":
		return tc.inferBitwiseOperatorType(leftType, operator, rightType)

	default:
		return "", fmt.Errorf("unknown binary operator: %s", operator)
	}
}

// InferUnaryOperatorType infers the result type of a unary operator expression
func (tc *TypeChecker) InferUnaryOperatorType(operator, operandType string) (string, error) {
	// First, check that the operand type is valid
	if err := tc.validateType(operandType); err != nil {
		return "", fmt.Errorf("invalid operand type '%s': %w", operandType, err)
	}

	switch operator {
	// Logical NOT
	case "!":
		// ! can be applied to any type (everything has boolean context in Perl)
		return "Bool", nil

	// Numeric negation
	case "-":
		return tc.inferNumericNegationType(operandType)

	// Numeric positive (unary +)
	case "+":
		return tc.inferNumericPositiveType(operandType)

	// Bitwise NOT
	case "~":
		return tc.inferBitwiseNotType(operandType)

	// Reference operator
	case "\\":
		return tc.inferReferenceType(operandType)

	// Dereference operators (handled separately as they're more complex)
	case "*", "@", "%", "&":
		return tc.inferDereferenceType(operator, operandType)

	default:
		return "", fmt.Errorf("unknown unary operator: %s", operator)
	}
}

// Helper methods for specific operator categories

func (tc *TypeChecker) inferArithmeticOperatorType(leftType, operator, rightType string) (string, error) {
	// Both operands must be numeric-compatible
	if !tc.isNumericCompatible(leftType) {
		return "", fmt.Errorf("left operand of '%s' must be numeric, got '%s'", operator, leftType)
	}
	if !tc.isNumericCompatible(rightType) {
		return "", fmt.Errorf("right operand of '%s' must be numeric, got '%s'", operator, rightType)
	}

	// Arithmetic operations generally return Num
	// Special case: if both operands are Int and operator preserves integer property
	if leftType == "Int" && rightType == "Int" && (operator == "+" || operator == "-" || operator == "*" || operator == "%") {
		return "Num", nil // In Perl, even integer arithmetic can produce non-integers
	}

	return "Num", nil
}

func (tc *TypeChecker) inferNumericComparisonType(leftType, operator, rightType string) (string, error) {
	// Both operands must be numeric-compatible
	if !tc.isNumericCompatible(leftType) {
		return "", fmt.Errorf("left operand of '%s' must be numeric, got '%s'", operator, leftType)
	}
	if !tc.isNumericCompatible(rightType) {
		return "", fmt.Errorf("right operand of '%s' must be numeric, got '%s'", operator, rightType)
	}

	// All numeric comparisons return Bool, except <=> which returns Int
	if operator == "<=>" {
		return "Int", nil // Returns -1, 0, or 1
	}
	return "Bool", nil
}

func (tc *TypeChecker) inferStringConcatenationType(leftType, operator, rightType string) (string, error) {
	// For typed Perl, we require explicit string types or types that are already strings
	// This prevents implicit numeric-to-string coercion
	if !tc.isDirectlyStringCompatible(leftType) {
		return "", fmt.Errorf("left operand of '.' must be string-compatible, got '%s'", leftType)
	}
	if !tc.isDirectlyStringCompatible(rightType) {
		return "", fmt.Errorf("right operand of '.' must be string-compatible, got '%s'", rightType)
	}

	return "Str", nil
}

func (tc *TypeChecker) inferStringComparisonType(leftType, operator, rightType string) (string, error) {
	// Both operands must be string-compatible
	if !tc.isStringifiable(leftType) {
		return "", fmt.Errorf("left operand of '%s' must be stringifiable, got '%s'", operator, leftType)
	}
	if !tc.isStringifiable(rightType) {
		return "", fmt.Errorf("right operand of '%s' must be stringifiable, got '%s'", operator, rightType)
	}

	// All string comparisons return Bool, except cmp which returns Int
	if operator == "cmp" {
		return "Int", nil // Returns -1, 0, or 1
	}
	return "Bool", nil
}

func (tc *TypeChecker) inferLogicalOperatorType(leftType, operator, rightType string) (string, error) {
	// Logical operators can work with any types (boolean context)
	// && and || can return either operand type depending on which is returned
	// For simplicity in typed Perl, we'll say they return Bool
	// // (defined-or) is a bit different but we'll treat it similarly
	return "Bool", nil
}

func (tc *TypeChecker) inferPatternMatchType(leftType, operator, rightType string) (string, error) {
	// Left side should be stringifiable, right side should be a regexp pattern
	if !tc.isStringifiable(leftType) {
		return "", fmt.Errorf("left operand of '%s' must be stringifiable, got '%s'", operator, leftType)
	}
	// For now, we'll accept any type for the right side (could be Regexp, Str, etc.)
	return "Bool", nil
}

func (tc *TypeChecker) inferBitwiseOperatorType(leftType, operator, rightType string) (string, error) {
	// Bitwise operations require integer-compatible types
	if !tc.isIntegerCompatible(leftType) {
		return "", fmt.Errorf("left operand of '%s' must be integer-compatible, got '%s'", operator, leftType)
	}
	if !tc.isIntegerCompatible(rightType) {
		return "", fmt.Errorf("right operand of '%s' must be integer-compatible, got '%s'", operator, rightType)
	}

	return "Int", nil
}

func (tc *TypeChecker) inferNumericNegationType(operandType string) (string, error) {
	if !tc.isNumericCompatible(operandType) {
		return "", fmt.Errorf("unary '-' requires numeric operand, got '%s'", operandType)
	}

	// Preserve the specific numeric type
	if operandType == "Int" {
		return "Int", nil
	}
	if operandType == "Float" {
		return "Float", nil
	}
	return "Num", nil
}

func (tc *TypeChecker) inferNumericPositiveType(operandType string) (string, error) {
	if !tc.isNumericCompatible(operandType) {
		return "", fmt.Errorf("unary '+' requires numeric operand, got '%s'", operandType)
	}

	// Preserve the specific numeric type
	if operandType == "Int" {
		return "Int", nil
	}
	if operandType == "Float" {
		return "Float", nil
	}
	return "Num", nil
}

func (tc *TypeChecker) inferBitwiseNotType(operandType string) (string, error) {
	if !tc.isIntegerCompatible(operandType) {
		return "", fmt.Errorf("bitwise '~' requires integer-compatible operand, got '%s'", operandType)
	}
	return "Int", nil
}

func (tc *TypeChecker) inferReferenceType(operandType string) (string, error) {
	// Taking a reference to a type creates a reference type
	switch operandType {
	case "Scalar", "Str", "Num", "Int", "Float", "Bool":
		return "ScalarRef", nil
	case "Array":
		return "ArrayRef", nil
	case "Hash":
		return "HashRef", nil
	case "Code":
		return "CodeRef", nil
	case "Glob":
		return "GlobRef", nil
	default:
		// For other types, create a generic Ref
		return "Ref", nil
	}
}

func (tc *TypeChecker) inferDereferenceType(operator, operandType string) (string, error) {
	// This is complex - dereferencing depends on the reference type
	// For now, simplified implementation
	if !strings.HasSuffix(operandType, "Ref") && operandType != "Ref" {
		return "", fmt.Errorf("cannot dereference non-reference type '%s'", operandType)
	}

	switch operator {
	case "*": // Scalar dereference
		if operandType == "ScalarRef" || operandType == "Ref" {
			return "Scalar", nil
		}
		return "", fmt.Errorf("cannot scalar dereference '%s'", operandType)
	case "@": // Array dereference
		if operandType == "ArrayRef" || operandType == "Ref" {
			return "Array", nil
		}
		return "", fmt.Errorf("cannot array dereference '%s'", operandType)
	case "%": // Hash dereference
		if operandType == "HashRef" || operandType == "Ref" {
			return "Hash", nil
		}
		return "", fmt.Errorf("cannot hash dereference '%s'", operandType)
	case "&": // Code dereference
		if operandType == "CodeRef" || operandType == "Ref" {
			return "Code", nil
		}
		return "", fmt.Errorf("cannot code dereference '%s'", operandType)
	default:
		return "", fmt.Errorf("unknown dereference operator: %s", operator)
	}
}

// Helper methods for type compatibility checking

func (tc *TypeChecker) isNumericCompatible(typeName string) bool {
	// Check if type is numeric or can be treated as numeric
	return tc.Hierarchy.IsSubtypeOf(typeName, "Num") || typeName == "Num"
}

func (tc *TypeChecker) isStringifiable(typeName string) bool {
	// In our type system, anything that's a subtype of Str is stringifiable
	// This includes Int, Num, etc. due to our hierarchy
	return tc.Hierarchy.IsSubtypeOf(typeName, "Str") || typeName == "Str"
}

func (tc *TypeChecker) isDirectlyStringCompatible(typeName string) bool {
	// For strict string operations, only allow types that are directly string-related
	// This excludes numeric types that would require implicit conversion
	switch typeName {
	case "Str", "Scalar":
		return true
	default:
		// Check for parameterized string types like Maybe[Str]
		if strings.HasPrefix(typeName, "Maybe[") {
			baseType, params := ExtractTypeAndParams(typeName)
			if baseType == "Maybe" && len(params) > 0 {
				return tc.isDirectlyStringCompatible(params[0])
			}
		}
		return false
	}
}

func (tc *TypeChecker) isIntegerCompatible(typeName string) bool {
	// Only Int and its subtypes are considered integer-compatible for bitwise ops
	return tc.Hierarchy.IsSubtypeOf(typeName, "Int") || typeName == "Int"
}

// Container Type Checking Methods (Phase 6)

// CheckArrayElementAccess validates array element access type compatibility
func (tc *TypeChecker) CheckArrayElementAccess(arrayType string, index interface{}, expectedElementType string) error {
	// Extract element type from array type
	elementType, err := tc.InferArrayElementType(arrayType)
	if err != nil {
		return fmt.Errorf("invalid array type '%s': %w", arrayType, err)
	}

	// Check if the expected element type is compatible with the actual element type
	err = tc.CheckStrictParameterTypeCompatibility(expectedElementType, elementType)
	if err != nil {
		return fmt.Errorf("array element type mismatch: expected %s, but array contains %s", elementType, expectedElementType)
	}

	return nil
}

// CheckArrayElementAssignment validates array element assignment
func (tc *TypeChecker) CheckArrayElementAssignment(arrayType string, index interface{}, valueType string) error {
	// Extract element type from array type
	elementType, err := tc.InferArrayElementType(arrayType)
	if err != nil {
		return fmt.Errorf("invalid array type '%s': %w", arrayType, err)
	}

	// Check if the value type can be assigned to the element type
	err = tc.CheckStrictParameterTypeCompatibility(valueType, elementType)
	if err != nil {
		return fmt.Errorf("cannot assign %s to array element of type %s: %w", valueType, elementType, err)
	}

	return nil
}

// CheckHashElementAccess validates hash element access type compatibility
func (tc *TypeChecker) CheckHashElementAccess(hashType string, key interface{}, expectedValueType string) error {
	// Extract key and value types from hash type
	keyType, valueType, err := tc.InferHashTypes(hashType)
	if err != nil {
		return fmt.Errorf("invalid hash type '%s': %w", hashType, err)
	}

	// Validate the key type (if provided as specific type)
	if keyStr, ok := key.(string); ok {
		err = tc.CheckStrictParameterTypeCompatibility("Str", keyType)
		if err != nil {
			return fmt.Errorf("hash key type mismatch: expected %s, got string '%s'", keyType, keyStr)
		}
	}

	// Check if the expected value type is compatible with the actual value type
	err = tc.CheckStrictParameterTypeCompatibility(expectedValueType, valueType)
	if err != nil {
		return fmt.Errorf("hash element type mismatch: expected %s, but hash contains %s", valueType, expectedValueType)
	}

	return nil
}

// CheckHashKeyAccess validates hash key type compatibility
func (tc *TypeChecker) CheckHashKeyAccess(hashType string, key interface{}) error {
	// Extract key type from hash type
	keyType, _, err := tc.InferHashTypes(hashType)
	if err != nil {
		return fmt.Errorf("invalid hash type '%s': %w", hashType, err)
	}

	// Determine the type of the provided key
	var providedKeyType string
	switch key.(type) {
	case string:
		providedKeyType = "Str"
	case int, int32, int64:
		providedKeyType = "Int"
	case float32, float64:
		providedKeyType = "Num"
	default:
		providedKeyType = fmt.Sprintf("%T", key)
	}

	// Check if the provided key type is compatible with the expected key type
	err = tc.CheckStrictParameterTypeCompatibility(providedKeyType, keyType)
	if err != nil {
		return fmt.Errorf("hash key type mismatch: expected %s, got %s", keyType, providedKeyType)
	}

	return nil
}

// CheckContainerTypeCompatibility checks covariant subtyping for container types
func (tc *TypeChecker) CheckContainerTypeCompatibility(sourceType, targetType string) error {
	// Parse source container type
	sourceBase, sourceParams, err := tc.ExtractContainerInfo(sourceType)
	if err != nil {
		return fmt.Errorf("invalid source container type '%s': %w", sourceType, err)
	}

	// Parse target container type
	targetBase, targetParams, err := tc.ExtractContainerInfo(targetType)
	if err != nil {
		return fmt.Errorf("invalid target container type '%s': %w", targetType, err)
	}

	// Base types must match
	if sourceBase != targetBase {
		return fmt.Errorf("container base types do not match: %s vs %s", sourceBase, targetBase)
	}

	// Check parameter compatibility based on container type
	switch sourceBase {
	case "ArrayRef":
		// ArrayRef[T1] ⊆ ArrayRef[T2] if T1 ⊆ T2 (covariance)
		if len(sourceParams) != 1 || len(targetParams) != 1 {
			return fmt.Errorf("ArrayRef requires exactly 1 parameter")
		}
		return tc.CheckContainerElementCompatibility(sourceParams[0], targetParams[0])

	case "HashRef":
		// Handle both HashRef[V] and HashRef[K,V] forms
		sourceKey, sourceValue := tc.normalizeHashParams(sourceParams)
		targetKey, targetValue := tc.normalizeHashParams(targetParams)

		// Key types must match exactly (no covariance)
		if sourceKey != targetKey {
			return fmt.Errorf("hash key types must match exactly: %s vs %s", sourceKey, targetKey)
		}

		// Value types can be covariant: HashRef[K,V1] ⊆ HashRef[K,V2] if V1 ⊆ V2
		return tc.CheckContainerElementCompatibility(sourceValue, targetValue)

	case "Maybe", "Optional":
		// Maybe[T1] ⊆ Maybe[T2] if T1 ⊆ T2 (covariance)
		if len(sourceParams) != 1 || len(targetParams) != 1 {
			return fmt.Errorf("%s requires exactly 1 parameter", sourceBase)
		}
		return tc.CheckContainerElementCompatibility(sourceParams[0], targetParams[0])

	default:
		return fmt.Errorf("unsupported container type: %s", sourceBase)
	}
}

// CheckContainerElementCompatibility checks if one element type is compatible with another (with covariance)
func (tc *TypeChecker) CheckContainerElementCompatibility(sourceElementType, targetElementType string) error {
	// Exact match is always allowed
	if sourceElementType == targetElementType {
		return nil
	}

	// Check if both are container types (for nested containers)
	if tc.isContainerType(sourceElementType) && tc.isContainerType(targetElementType) {
		return tc.CheckContainerTypeCompatibility(sourceElementType, targetElementType)
	}

	// For non-container element types, use subtype relationships
	// This allows ArrayRef[Int] ⊆ ArrayRef[Num] (since Int ⊆ Num)
	if tc.Hierarchy.IsSubtypeOf(sourceElementType, targetElementType) {
		return nil
	}

	return fmt.Errorf("element type %s is not compatible with %s", sourceElementType, targetElementType)
}

// normalizeHashParams normalizes hash parameters to (key, value) pair
func (tc *TypeChecker) normalizeHashParams(params []string) (string, string) {
	switch len(params) {
	case 1:
		// HashRef[V] defaults to HashRef[Str,V]
		return "Str", params[0]
	case 2:
		// HashRef[K,V]
		return params[0], params[1]
	default:
		// Invalid, but we'll return defaults to avoid panics
		return "Str", "Any"
	}
}

// isContainerType checks if a type is a container type
func (tc *TypeChecker) isContainerType(typeName string) bool {
	return strings.HasPrefix(typeName, "ArrayRef[") ||
		strings.HasPrefix(typeName, "HashRef[") ||
		strings.HasPrefix(typeName, "Maybe[") ||
		strings.HasPrefix(typeName, "Optional[")
}

// Array operation methods

// CheckArrayPushOperation validates array push/append operations
func (tc *TypeChecker) CheckArrayPushOperation(arrayType string, valueType string) error {
	// Extract element type from array type
	elementType, err := tc.InferArrayElementType(arrayType)
	if err != nil {
		return fmt.Errorf("invalid array type '%s': %w", arrayType, err)
	}

	// Check if the value type can be pushed to the array
	err = tc.CheckStrictParameterTypeCompatibility(valueType, elementType)
	if err != nil {
		return fmt.Errorf("cannot push %s to array of %s: %w", valueType, elementType, err)
	}

	return nil
}

// InferArrayElementType extracts the element type from an array type
func (tc *TypeChecker) InferArrayElementType(arrayType string) (string, error) {
	baseType, params, err := tc.ExtractContainerInfo(arrayType)
	if err != nil {
		return "", err
	}

	if baseType != "ArrayRef" {
		return "", fmt.Errorf("type '%s' is not an array type", arrayType)
	}

	if len(params) != 1 {
		return "", fmt.Errorf("ArrayRef requires exactly 1 parameter, got %d", len(params))
	}

	return params[0], nil
}

// Hash operation methods

// InferHashTypes extracts key and value types from a hash type
func (tc *TypeChecker) InferHashTypes(hashType string) (string, string, error) {
	baseType, params, err := tc.ExtractContainerInfo(hashType)
	if err != nil {
		return "", "", err
	}

	if baseType != "HashRef" {
		return "", "", fmt.Errorf("type '%s' is not a hash type", hashType)
	}

	keyType, valueType := tc.normalizeHashParams(params)
	return keyType, valueType, nil
}

// ExtractContainerInfo extracts base type and parameters from a container type
func (tc *TypeChecker) ExtractContainerInfo(containerType string) (string, []string, error) {
	// Check if it's actually a container type
	if !strings.Contains(containerType, "[") {
		return "", nil, fmt.Errorf("type '%s' is not a container type", containerType)
	}

	// Use the existing ExtractTypeAndParams function
	baseType, params := ExtractTypeAndParams(containerType)
	if baseType == "" {
		return "", nil, fmt.Errorf("failed to parse container type '%s'", containerType)
	}

	return baseType, params, nil
}

// checkASTFunctionReturns checks return types in an AST
// This function is referenced by CheckAST but commented out in the current implementation
// nolint:unused
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
