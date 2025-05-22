// ABOUTME: Type inference logic for the typechecker
// ABOUTME: Infers types from expressions and literals

package typechecker

import (
	"strings"
)

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
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	// Handle negative numbers
	if strings.HasPrefix(s, "-") {
		s = s[1:]
	}

	if s == "" {
		return false
	}

	// Check if all characters are digits or decimal point
	hasDecimal := false
	for _, c := range s {
		if c == '.' {
			if hasDecimal {
				return false // Multiple decimal points
			}
			hasDecimal = true
		} else if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// isStringLiteral checks if a string is a string literal
func isStringLiteral(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'"))
}
