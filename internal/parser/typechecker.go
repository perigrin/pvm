// ABOUTME: Type checking implementation for PSC
// ABOUTME: Validates type annotations in Perl code

package parser

import (
	"fmt"
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
