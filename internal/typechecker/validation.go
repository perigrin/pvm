// ABOUTME: Type validation logic for the typechecker
// ABOUTME: Validates type annotations and type compatibility

package typechecker

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
)

// validateType validates that a type string is valid
func (tc *TypeChecker) validateType(typeStr string) error {
	// Basic type validation - check if it's in our known types
	basicTypes := map[string]bool{
		"Int":      true,
		"Str":      true,
		"Num":      true,
		"Bool":     true,
		"ArrayRef": true,
		"HashRef":  true,
		"Any":      true,
		"Undef":    true,
	}

	if basicTypes[typeStr] {
		return nil
	}

	// For now, accept any type that starts with an uppercase letter
	// In a real implementation, we'd check against the type hierarchy
	if len(typeStr) > 0 && typeStr[0] >= 'A' && typeStr[0] <= 'Z' {
		return nil
	}

	return fmt.Errorf("unknown type: %s", typeStr)
}

// checkTypeAnnotation validates a single type annotation
func (tc *TypeChecker) checkTypeAnnotation(annotation *parser.TypeAnnotation) error {
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
	case parser.VarAnnotation:
		// For variable annotations, we check if there's an assignment
		// that conflicts with the annotation
		return tc.checkVariableAnnotation(annotation)
	}

	return nil
}

// checkVariableAnnotation checks a variable annotation
func (tc *TypeChecker) checkVariableAnnotation(annotation *parser.TypeAnnotation) error {
	// This would check variable-specific constraints
	// For now, it's a placeholder that always succeeds
	return nil
}

// CheckAssignment checks if an assignment is type-compatible
func (tc *TypeChecker) CheckAssignment(fromType, toType string, pos parser.Position) error {
	// If types are the same, assignment is valid
	if fromType == toType {
		return nil
	}

	// Some type conversions are allowed in Perl
	allowedConversions := map[string]map[string]bool{
		"Int": {
			"Str":  true, // Numbers can be stringified
			"Bool": true, // Numbers can be used as booleans
		},
		"Str": {
			"Bool": true, // Strings can be used as booleans
		},
		"Float": {
			"Str":  true, // Floats can be stringified
			"Bool": true, // Floats can be used as booleans
		},
	}

	if conversions, ok := allowedConversions[fromType]; ok {
		if conversions[toType] {
			return nil // Conversion is allowed
		}
	}

	// Assignment is not compatible
	return fmt.Errorf("Cannot assign %s to %s at line %d, column %d", fromType, toType, pos.Line, pos.Column)
}

// initializeValidationPatterns sets up the recognized validation patterns
func (tc *TypeChecker) initializeValidationPatterns() {
	// Add basic validation patterns for common Perl idioms

	// defined() check for Maybe types
	tc.ValidationPatterns = append(tc.ValidationPatterns, ValidationPattern{
		Name:    "defined check",
		Pattern: "defined($var)",
		RefinementFunc: func(varName, currentType string) string {
			// For Maybe[T] types, refine to T
			if strings.HasPrefix(currentType, "Maybe[") {
				baseType, params := ExtractTypeAndParams(currentType)
				if baseType == "Maybe" && len(params) > 0 {
					return params[0]
				}
			}
			return currentType
		},
		Checker: func(node parser.Node) (string, bool) {
			// Check if this is a defined() expression
			if node.Type() != "defined_expression" && !strings.Contains(node.Type(), "function_call") {
				return "", false
			}

			// Extract the variable name from defined($var)
			nodeText := getNodeText(node)
			if strings.HasPrefix(nodeText, "defined(") && strings.HasSuffix(nodeText, ")") {
				varName := strings.TrimPrefix(nodeText, "defined(")
				varName = strings.TrimSuffix(varName, ")")
				varName = strings.TrimSpace(varName)

				// Only handle variables
				if strings.HasPrefix(varName, "$") {
					return varName, true
				}
			}

			return "", false
		},
	})

	// Add ref() check for reftype refinement
	tc.ValidationPatterns = append(tc.ValidationPatterns, ValidationPattern{
		Name:    "ref check",
		Pattern: "ref($var) eq 'TYPE'",
		RefinementFunc: func(varName, currentType string) string {
			// When checking ref type, we can refine Ref to a specific reference type
			if currentType == "Ref" || currentType == "Any" {
				// The specific type would be determined based on the ref type string
				// For now, we're just demonstrating with ArrayRef
				return "ArrayRef"
			}
			return currentType
		},
		Checker: func(node parser.Node) (string, bool) {
			// This is a simplified check that would need to be enhanced in a real implementation
			nodeText := getNodeText(node)

			// Look for patterns like 'ref($var) eq "ARRAY"'
			if strings.Contains(nodeText, "ref(") &&
				(strings.Contains(nodeText, "'ARRAY'") || strings.Contains(nodeText, "\"ARRAY\"")) {
				// Extract the variable from ref($var)
				start := strings.Index(nodeText, "ref(") + 4
				end := strings.Index(nodeText[start:], ")")
				if end > 0 {
					varName := nodeText[start : start+end]
					varName = strings.TrimSpace(varName)
					if strings.HasPrefix(varName, "$") {
						return varName, true
					}
				}
			}

			return "", false
		},
	})

	// Add more patterns as needed
}
