// ABOUTME: Assignment checking logic for the typechecker
// ABOUTME: Validates type compatibility in variable assignments

package typechecker

import (
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

// checkNodeForAssignments checks a node and its children for assignments
func (tc *TypeChecker) checkNodeForAssignments(node ast.Node, errors *[]error) {
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
		// We don't know the type of this variable, so we can't check the assignment
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
