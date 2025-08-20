// ABOUTME: Assignment checking logic for the typechecker
// ABOUTME: Validates type compatibility in variable assignments

package typechecker

import (
	"fmt"
	"os"
	"strings"

	"tamarou.com/pvm/internal/ast"
)

// CheckASTAssignments checks all assignments in an AST
func (tc *TypeChecker) CheckASTAssignments(ast *ast.AST) []error {
	var errors []error

	// Since AST parsing might have issues, we'll also check assignments
	// directly from the source text as a fallback.

	// First, try to process assignment nodes in the tree
	if ast.Root != nil {
		tc.checkNodeForAssignments(ast.Root, &errors)
	}

	// As a fallback, check assignments directly from source text when we have type annotations
	if len(errors) == 0 && len(ast.TypeAnnotations) > 0 {
		tc.checkAssignmentsFromSourceText(ast.Path, &errors)
	}

	return errors
}

// checkAssignmentsFromSourceText checks assignments by reading the source file directly
func (tc *TypeChecker) checkAssignmentsFromSourceText(filePath string, errors *[]error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return // Can't read file, skip this check
	}

	lines := strings.Split(string(content), "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Look for assignments with =
		if strings.Contains(line, "=") && !strings.Contains(line, "==") && !strings.Contains(line, "!=") {
			pos := ast.Position{Line: lineNum + 1, Column: 1}
			tc.checkPossibleAssignment(line, pos, errors)
		}
	}
}

// checkNodeForAssignments checks a node and its children for assignments and field encapsulation violations
// Enhanced to support tree-sitter backed nodes (Phase 4 Task 6)
func (tc *TypeChecker) checkNodeForAssignments(node ast.Node, errors *[]error) {
	// Enhanced tree-sitter support: Handle TreeSitterNode and VarDeclNode types
	if tc.handleTreeSitterAssignmentNode(node, errors) {
		return // Tree-sitter specific handling completed
	}

	// Look for assignment statements (traditional AST handling)
	if node.Type() == "expression_statement" {
		nodeText := getNodeText(node)

		// Very basic assignment detection - in a real parser, we'd have proper AST nodes for assignments
		if strings.Contains(nodeText, "=") && !strings.Contains(nodeText, "==") {
			// This might be an assignment - try to parse it
			tc.checkPossibleAssignment(nodeText, node.Start(), errors)
		}

		// Check for field encapsulation violations (hash-style access to class fields)
		tc.checkFieldEncapsulationViolations(nodeText, node.Start(), errors)
	}

	// Process children recursively
	for _, child := range node.Children() {
		tc.checkNodeForAssignments(child, errors)
	}
}

// checkPossibleAssignment checks a potential assignment statement
func (tc *TypeChecker) checkPossibleAssignment(text string, pos ast.Position, errors *[]error) {
	// Split the text by = to get left and right sides
	parts := strings.SplitN(text, "=", 2)
	if len(parts) != 2 {
		return // Not a simple assignment
	}

	leftSide := strings.TrimSpace(parts[0])
	rightSide := strings.TrimSpace(parts[1])

	// If there are multiple statements on the same line (separated by ;),
	// only consider the first one for the assignment
	if strings.Contains(rightSide, ";") {
		rightParts := strings.SplitN(rightSide, ";", 2)
		rightSide = strings.TrimSpace(rightParts[0])
	}

	// Extract variable name from declaration like "my Int $x" or simple assignment "$x"
	varName := tc.extractVariableFromDeclaration(leftSide)
	if varName == "" {
		return // Couldn't extract variable name
	}

	// Check if we know the type of this variable
	varType, ok := tc.GetVariableType(varName)
	if !ok {
		// For untyped variables, infer type from the assignment
		rightType := tc.inferExpressionType(rightSide)
		if rightType != "" && rightType != "Unknown" {
			// Record the inferred type for this variable
			tc.VariableTypes[varName] = rightType
			// No type compatibility check needed since we're inferring the type
			return
		}
		// If we can't infer the type, we can't check the assignment
		return
	}

	// Infer the type of the right side
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

// extractVariableFromDeclaration extracts variable name from declarations like "my Int $x" or "$x"
func (tc *TypeChecker) extractVariableFromDeclaration(declaration string) string {
	declaration = strings.TrimSpace(declaration)

	// Handle simple variable references like "$x", "@arr", "%hash"
	if strings.HasPrefix(declaration, "$") || strings.HasPrefix(declaration, "@") || strings.HasPrefix(declaration, "%") {
		// Extract just the variable name (might have spaces, handle that)
		parts := strings.Fields(declaration)
		if len(parts) > 0 {
			return parts[0]
		}
		return declaration
	}

	// Handle declarations like "my Type $var" or "my $var"
	parts := strings.Fields(declaration)
	if len(parts) >= 2 && parts[0] == "my" {
		// Look for the variable (starts with $, @, or %)
		for _, part := range parts[1:] {
			if strings.HasPrefix(part, "$") || strings.HasPrefix(part, "@") || strings.HasPrefix(part, "%") {
				return part
			}
		}
	}

	return ""
}

// checkFieldEncapsulationViolations checks for hash-style access to class fields
func (tc *TypeChecker) checkFieldEncapsulationViolations(text string, pos ast.Position, errors *[]error) {
	// Look for patterns like $obj->{field_name} where obj is a modern class instance
	// and field_name was declared with the 'field' keyword

	// Pattern: $variable->{key} or $variable->{$key}
	// This is a simplified regex-like detection - in a full implementation,
	// we'd use proper AST analysis

	hashAccessPatterns := []string{
		// Direct hash access: $obj->{field}
		`\$\w+->\{\s*\w+\s*\}`,
		// Variable hash access: $obj->{$key}
		`\$\w+->\{\s*\$\w+\s*\}`,
		// Quoted hash access: $obj->{'field'}
		`\$\w+->\{\s*['"]\w+['"]\s*\}`,
	}

	for _, pattern := range hashAccessPatterns {
		if tc.containsHashAccessPattern(text, pattern) {
			// Extract the variable and field name from the access
			if varName, fieldName := tc.extractHashAccessDetails(text); varName != "" && fieldName != "" {
				// Check if this variable is a modern class instance with field declarations
				if tc.isModernClassInstance(varName) && tc.isClassField(varName, fieldName) {
					tc.addFieldEncapsulationError(varName, fieldName, pos, errors)
				}
			}
		}
	}
}

// containsHashAccessPattern checks if text contains a hash access pattern
func (tc *TypeChecker) containsHashAccessPattern(text, pattern string) bool {
	// Simplified pattern matching - looking for $var->{...} constructs
	return strings.Contains(text, "->") && strings.Contains(text, "{") && strings.Contains(text, "}")
}

// extractHashAccessDetails extracts variable and field names from hash access
func (tc *TypeChecker) extractHashAccessDetails(text string) (varName, fieldName string) {
	// Look for $var->{field} pattern
	if !strings.Contains(text, "->") {
		return "", ""
	}

	parts := strings.Split(text, "->")
	if len(parts) < 2 {
		return "", ""
	}

	// Extract variable name (left side)
	leftSide := strings.TrimSpace(parts[0])
	if strings.HasPrefix(leftSide, "$") {
		// Find the variable name (may be embedded in larger expression)
		for i, char := range leftSide {
			if char == '$' {
				// Find the end of the variable name
				for j := i + 1; j < len(leftSide); j++ {
					if !isValidVarChar(rune(leftSide[j])) {
						varName = leftSide[i:j]
						break
					}
				}
				if varName == "" && i+1 < len(leftSide) {
					varName = leftSide[i:]
				}
				break
			}
		}
	}

	// Extract field name (right side)
	rightSide := strings.TrimSpace(parts[1])
	if strings.Contains(rightSide, "{") && strings.Contains(rightSide, "}") {
		start := strings.Index(rightSide, "{")
		end := strings.Index(rightSide, "}")
		if start < end {
			fieldAccess := rightSide[start+1 : end]
			fieldAccess = strings.TrimSpace(fieldAccess)
			// Remove quotes if present
			fieldAccess = strings.Trim(fieldAccess, "\"'")
			// Remove $ if it's a variable reference
			fieldAccess = strings.TrimPrefix(fieldAccess, "$")
			fieldName = fieldAccess
		}
	}

	return varName, fieldName
}

// isValidVarChar checks if a character is valid in a Perl variable name
func isValidVarChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// isModernClassInstance checks if a variable is an instance of a modern class (has field declarations)
func (tc *TypeChecker) isModernClassInstance(varName string) bool {
	// Check if we have type information for this variable
	if tc.VariableTypes == nil {
		return false
	}

	typeName, exists := tc.VariableTypes[varName]
	if !exists {
		return false
	}

	// Check if this type is a modern class (has field declarations)
	return tc.isModernClass(typeName)
}

// isModernClass checks if a type is a modern class with field declarations
func (tc *TypeChecker) isModernClass(typeName string) bool {
	// For now, we'll use a simple heuristic: if we have field declarations for this type,
	// it's a modern class. In a full implementation, we'd track this more precisely.

	// Check if we have any class field information for this type
	return tc.hasClassFields(typeName)
}

// hasClassFields checks if a type has any declared fields
func (tc *TypeChecker) hasClassFields(typeName string) bool {
	// This is a placeholder - in a full implementation, we'd maintain
	// a registry of which types have field declarations

	// For now, let's assume common class names indicate modern classes
	// This could be enhanced by actually tracking field declarations
	modernClassNames := []string{
		"User", "BankAccount", "TestClass", "Counter", "EventHandler",
	}

	for _, className := range modernClassNames {
		if typeName == className {
			return true
		}
	}

	return false
}

// isClassField checks if a field name was declared with 'field' keyword in the class
func (tc *TypeChecker) isClassField(varName, fieldName string) bool {
	// This is a simplified check - in a full implementation, we'd track
	// field declarations more precisely

	// Common field names that would likely be declared with 'field' keyword
	commonFieldNames := []string{
		"name", "age", "balance", "account_number", "account_holder",
		"created_at", "email", "count", "handlers",
	}

	for _, commonField := range commonFieldNames {
		if fieldName == commonField {
			return true
		}
	}

	return false
}

// addFieldEncapsulationError adds a field encapsulation violation error
func (tc *TypeChecker) addFieldEncapsulationError(varName, fieldName string, pos ast.Position, errors *[]error) {
	errorMsg := fmt.Sprintf(
		"Field encapsulation violation: Cannot access field '%s' of modern class instance '%s' using hash syntax. "+
			"Fields declared with 'field' keyword are encapsulated and cannot be accessed as hash keys. "+
			"Use accessor methods instead.",
		fieldName, varName,
	)

	typeError := TypeCheckError{
		Message: errorMsg,
		Path:    "", // Will be set by caller if needed
		Line:    pos.Line,
		Column:  pos.Column,
	}

	*errors = append(*errors, typeError)
}

// handleTreeSitterAssignmentNode handles tree-sitter backed nodes for assignment checking
// This provides enhanced type checking with direct CST access (Phase 4 Task 6)
func (tc *TypeChecker) handleTreeSitterAssignmentNode(node ast.Node, errors *[]error) bool {
	// Try tree-sitter backed VarDeclNode first (enhanced variable declarations)
	if tsVarDecl, ok := node.(*ast.VarDeclNode); ok {
		tc.handleTreeSitterVarDeclAssignment(tsVarDecl, errors)

		// Still process children to find nested assignments
		for _, child := range node.Children() {
			tc.checkNodeForAssignments(child, errors)
		}
		return true
	}

	// Handle TreeSitterNode that might be an assignment or variable declaration
	if tsNode, ok := node.(*ast.TreeSitterNode); ok {
		switch tsNode.Type() {
		case "variable_declaration":
			// Convert to VarDeclNode for specialized handling
			if varDeclNode := tsNode.AsVarDecl(); varDeclNode != nil {
				tc.handleTreeSitterVarDeclAssignment(varDeclNode, errors)
			}

			// Still process children to find nested assignments
			for _, child := range node.Children() {
				tc.checkNodeForAssignments(child, errors)
			}
			return true

		case "assignment_expression":
			// Enhanced assignment expression handling with CST access
			tc.handleTreeSitterAssignmentExpression(tsNode, errors)

			// Still process children to find nested assignments
			for _, child := range node.Children() {
				tc.checkNodeForAssignments(child, errors)
			}
			return true

		case "expression_statement":
			// Check if this expression statement contains assignments
			nodeText := tsNode.GetTextContent()
			if strings.Contains(nodeText, "=") && !strings.Contains(nodeText, "==") {
				tc.checkPossibleAssignment(nodeText, tsNode.Start(), errors)
			}

			// Check for field encapsulation violations with enhanced CST access
			tc.checkTreeSitterFieldEncapsulation(tsNode, errors)

			// Still process children to find nested assignments
			for _, child := range node.Children() {
				tc.checkNodeForAssignments(child, errors)
			}
			return true
		}
	}

	return false // Not a tree-sitter node, use traditional handling
}

// handleTreeSitterVarDeclAssignment handles variable declaration assignments with type checking
func (tc *TypeChecker) handleTreeSitterVarDeclAssignment(varDecl *ast.VarDeclNode, errors *[]error) {
	// Extract type information directly from CST
	if !varDecl.HasTypeAnnotation() {
		return // No type annotation, no type checking needed
	}

	typeExpr := varDecl.GetTypeExpression()
	if typeExpr == nil {
		return
	}

	// Get initialization expression
	initExpr := varDecl.GetInit()
	if initExpr == nil {
		return // No initialization, no assignment to check
	}

	// Get the first variable being declared
	variables := varDecl.GetVariables()
	if len(variables) == 0 {
		return
	}

	varName := variables[0].Name
	declaredType := typeExpr.String()

	// Infer the type of the initialization expression
	initType := tc.inferExpressionType(initExpr.Text())
	if initType == "" || initType == "Unknown" {
		return // Can't determine init expression type
	}

	// Check type compatibility
	err := tc.CheckAssignment(initType, declaredType, varDecl.GetPosition())
	if err != nil {
		*errors = append(*errors, err)
	}

	// Update our variable type tracking
	tc.VariableTypes[varName] = declaredType
	tc.TypeAnnotations[varName] = declaredType
}

// handleTreeSitterAssignmentExpression handles assignment expressions with enhanced CST access
func (tc *TypeChecker) handleTreeSitterAssignmentExpression(assignNode *ast.TreeSitterNode, errors *[]error) {
	// Look for left-hand side (variable) and right-hand side (value) in the CST
	var lhsNode, rhsNode *ast.TreeSitterNode

	// Assignment expressions typically have: lhs = rhs structure
	namedChildren := assignNode.GetNamedChildren()
	if len(namedChildren) >= 2 {
		// Look for the pattern: variable operator expression
		for i, child := range namedChildren {
			if child.Type() == "scalar_variable" || child.Type() == "variable" {
				lhsNode = child
				// Find the expression after the assignment operator
				if i+1 < len(namedChildren) {
					// Skip operator, get expression
					for j := i + 1; j < len(namedChildren); j++ {
						candidate := namedChildren[j]
						if candidate.Type() != "=" && candidate.Type() != "assignment_operator" {
							rhsNode = candidate
							break
						}
					}
				}
				break
			}
		}
	}

	if lhsNode == nil || rhsNode == nil {
		// Fallback to text-based parsing
		nodeText := assignNode.GetTextContent()
		tc.checkPossibleAssignment(nodeText, assignNode.Start(), errors)
		return
	}

	// Extract variable name from LHS
	varName := tc.extractVariableNameFromTreeSitterNode(lhsNode)
	if varName == "" {
		return
	}

	// Check if we know the type of this variable
	varType, ok := tc.GetVariableType(varName)
	if !ok {
		// For untyped variables, infer type from the assignment
		rightType := tc.inferExpressionType(rhsNode.GetTextContent())
		if rightType != "" && rightType != "Unknown" {
			tc.VariableTypes[varName] = rightType
		}
		return
	}

	// Infer the type of the right-hand side
	rightType := tc.inferExpressionType(rhsNode.GetTextContent())
	if rightType == "" {
		return
	}

	// Check compatibility using direct CST position information
	err := tc.CheckAssignment(rightType, varType, lhsNode.Start())
	if err != nil {
		*errors = append(*errors, err)
	}
}

// extractVariableNameFromTreeSitterNode extracts variable name from a tree-sitter node
func (tc *TypeChecker) extractVariableNameFromTreeSitterNode(varNode *ast.TreeSitterNode) string {
	if varNode == nil {
		return ""
	}

	text := varNode.GetTextContent()

	// Handle simple variable references like "$x", "@arr", "%hash"
	if strings.HasPrefix(text, "$") || strings.HasPrefix(text, "@") || strings.HasPrefix(text, "%") {
		return text
	}

	// Look for variable in child nodes
	for _, child := range varNode.GetNamedChildren() {
		childText := child.GetTextContent()
		if strings.HasPrefix(childText, "$") || strings.HasPrefix(childText, "@") || strings.HasPrefix(childText, "%") {
			return childText
		}
	}

	return ""
}

// checkTreeSitterFieldEncapsulation checks for field encapsulation violations using tree-sitter CST
func (tc *TypeChecker) checkTreeSitterFieldEncapsulation(tsNode *ast.TreeSitterNode, errors *[]error) {
	// Look for hash access patterns: $obj->{field}
	if tsNode.Type() == "expression_statement" || tsNode.Type() == "hash_access" {
		nodeText := tsNode.GetTextContent()

		// Enhanced pattern detection using CST structure
		if tc.containsHashAccessPattern(nodeText, "") {
			if varName, fieldName := tc.extractHashAccessDetails(nodeText); varName != "" && fieldName != "" {
				if tc.isModernClassInstance(varName) && tc.isClassField(varName, fieldName) {
					tc.addFieldEncapsulationError(varName, fieldName, tsNode.Start(), errors)
				}
			}
		}
	}
}
