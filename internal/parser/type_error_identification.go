// ABOUTME: Type error identification for tree-sitter parse failures
// ABOUTME: Creates structured error objects without formatting - presentation handled by consumers

package parser

import (
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/errors"
)

// TypeErrorIdentifier analyzes tree-sitter parse failures to classify type errors
type TypeErrorIdentifier struct {
	// maxNestingDepth prevents identification of unreasonably nested types
	maxNestingDepth int
}

// NewTypeErrorIdentifier creates a new type error identifier
func NewTypeErrorIdentifier() *TypeErrorIdentifier {
	return &TypeErrorIdentifier{
		maxNestingDepth: 20, // Reasonable limit for type nesting
	}
}

// IdentifyTypeError analyzes failed parse input to determine the specific type error
func (tei *TypeErrorIdentifier) IdentifyTypeError(source string, position ast.Position, context string) *errors.TypeParseError {
	// Extract the relevant portion of source around the error
	lines := strings.Split(source, "\n")
	if position.Line <= 0 || position.Line > len(lines) {
		return &errors.TypeParseError{
			ErrorType:  "UnknownTypeError",
			Message:    "Unknown type parsing error",
			Position:   position,
			Suggestion: "Check syntax",
			Context:    context,
			ErrorCode:  errors.UnknownTypeError,
			Source:     source,
			SourceLine: "",
		}
	}

	line := lines[position.Line-1]
	errorCode, message, suggestion, errorType := tei.analyzeTypeError(line, position)

	return &errors.TypeParseError{
		ErrorType:  errorType,
		Message:    message,
		Position:   position,
		Suggestion: suggestion,
		Context:    context,
		ErrorCode:  errorCode,
		Source:     source,
		SourceLine: line,
	}
}

// analyzeTypeError analyzes the source line to determine the specific type error
func (tei *TypeErrorIdentifier) analyzeTypeError(line string, position ast.Position) (errors.TypeErrorCode, string, string, string) {
	// Check for common error patterns in order of specificity

	// Deep nesting check
	if strings.Count(line, "[") > 20 || strings.Count(line, "ArrayRef") > 10 {
		return errors.DeepNestingError,
			"Type nesting too deep",
			"Reduce type nesting to less than 20 levels",
			"DeepNestingError"
	}

	// Type assertion errors (highest priority for 'as' keyword)
	if strings.Contains(line, " as ") && (strings.HasSuffix(strings.TrimSpace(line), "as") ||
		strings.HasSuffix(strings.TrimSpace(line), "as ;")) {
		return errors.IncompleteTypeAssertionError,
			"Incomplete type assertion - missing target type",
			"Add the target type after 'as' keyword",
			"IncompleteTypeAssertionError"
	}

	// Negation syntax errors (before general pattern checks)
	if strings.Contains(line, "!!") && containsTypeLikePattern(line) {
		return errors.InvalidNegationSyntaxError,
			"Invalid negation type syntax - use single '!' before type",
			"Change '!!' to '!' for negation types",
			"errors.InvalidNegationSyntaxError"
	}

	// Missing closing bracket
	if strings.Contains(line, "ArrayRef[") && !strings.Contains(line, "]") {
		return errors.MissingClosingBracketError,
			"Missing closing bracket in parameterized type",
			"Add closing ']' to complete the parameterized type",
			"errors.MissingClosingBracketError"
	}

	// Missing closing bracket for HashRef
	if strings.Contains(line, "HashRef[") && !strings.Contains(line, "]") {
		return errors.MissingClosingBracketError,
			"Missing closing bracket in parameterized type",
			"Add closing ']' to complete the parameterized type",
			"errors.MissingClosingBracketError"
	}

	// Union syntax errors
	if strings.Contains(line, "||") && containsTypeLikePattern(line) {
		return errors.InvalidUnionSyntaxError,
			"Invalid union type syntax - use single '|' between types",
			"Change '||' to '|' for union types",
			"errors.InvalidUnionSyntaxError"
	}

	// Parameterized type errors
	if strings.Contains(line, "ArrayRef[") && strings.Contains(line, "ArrayRef[ ") {
		return errors.InvalidParameterizedTypeError,
			"Invalid parameterized type - unexpected space after '['",
			"Remove space after '[' in parameterized type",
			"errors.InvalidParameterizedTypeError"
	}

	// Where clause errors
	if strings.Contains(line, "where") && strings.Contains(line, "where $") {
		return errors.InvalidWhereClauseError,
			"Invalid where clause syntax",
			"Use 'where Type: Constraint' syntax for type constraints",
			"errors.InvalidWhereClauseError"
	}

	// Intersection syntax errors
	if strings.Contains(line, "&&") && containsTypeLikePattern(line) {
		return errors.InvalidIntersectionSyntaxError,
			"Invalid intersection type syntax - use single '&' between types",
			"Change '&&' to '&' for intersection types",
			"errors.InvalidIntersectionSyntaxError"
	}

	// Missing type name (lowest priority - most general)
	if strings.Contains(line, "my ") && strings.Contains(line, "$") &&
		!containsValidTypeName(line) {
		return errors.MissingTypeNameError,
			"Missing or invalid type name in variable declaration",
			"Add a valid type name before the variable",
			"errors.MissingTypeNameError"
	}

	return errors.UnknownTypeError, "Syntax error in type expression", "Check type syntax", "UnknownTypeError"
}

// containsValidTypeName checks if the line contains a valid type name pattern
func containsValidTypeName(line string) bool {
	// Simple heuristic: look for capitalized words that could be type names
	words := strings.Fields(line)
	for _, word := range words {
		if len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z' {
			// Could be a type name like Int, Str, ArrayRef, etc.
			return true
		}
	}
	return false
}

// containsTypeLikePattern checks if the line contains patterns that look like types
func containsTypeLikePattern(line string) bool {
	return containsValidTypeName(line) ||
		strings.Contains(line, "ArrayRef") ||
		strings.Contains(line, "HashRef") ||
		strings.Contains(line, "CodeRef") ||
		strings.Contains(line, "Undef") ||
		strings.Contains(line, "Object") ||
		strings.Contains(line, "Serializable")
}

// IdentifyTreeSitterError analyzes tree-sitter parse failures and converts them to specific type errors
func (tei *TypeErrorIdentifier) IdentifyTreeSitterError(source string, treeSitterError error, context string) error {
	// Extract position information from tree-sitter error if possible
	position := ast.Position{Line: 1, Column: 1} // Default position

	// Try to extract position from error message
	if treeSitterError != nil {
		errorStr := treeSitterError.Error()
		// Look for position patterns in tree-sitter errors
		if strings.Contains(errorStr, "parse error") {
			// For now, use simple heuristics to find likely error positions
			// This could be enhanced to parse tree-sitter error details more precisely
			lines := strings.Split(source, "\n")
			for i, line := range lines {
				// Look for lines with type-like syntax that might have errors
				if containsTypeLikePattern(line) && (strings.Contains(line, "[") || strings.Contains(line, "|") || strings.Contains(line, "&")) {
					position = ast.Position{Line: i + 1, Column: 1}
					break
				}
			}
		}
	}

	// Identify the specific type error
	typeError := tei.IdentifyTypeError(source, position, context)
	return typeError
}
