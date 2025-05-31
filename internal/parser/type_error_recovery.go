// ABOUTME: Enhanced error recovery and position tracking for type expressions
// ABOUTME: Provides graceful error handling for malformed type annotations

package parser

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
)

// TypeError represents a type-specific parsing error with enhanced information
type TypeError struct {
	// Message is the primary error message
	Message string

	// Position is the exact location where the error occurred
	Position ast.Position

	// Suggestion provides a helpful suggestion for fixing the error
	Suggestion string

	// Context describes the type expression context where the error occurred
	Context string

	// ErrorCode is a specific error code for programmatic handling
	ErrorCode TypeErrorCode

	// Source is the original source text that caused the error
	Source string
}

// TypeErrorCode represents specific type error categories
type TypeErrorCode int

const (
	// UnknownTypeError is for unrecognized errors
	UnknownTypeError TypeErrorCode = iota

	// MissingClosingBracketError for parameterized types
	MissingClosingBracketError

	// InvalidUnionSyntaxError for malformed union types
	InvalidUnionSyntaxError

	// IncompleteTypeAssertionError for malformed 'as' expressions
	IncompleteTypeAssertionError

	// InvalidParameterizedTypeError for malformed parameterized types
	InvalidParameterizedTypeError

	// MissingTypeNameError for empty type annotations
	MissingTypeNameError

	// InvalidWhereClauseError for malformed where constraints
	InvalidWhereClauseError

	// InvalidIntersectionSyntaxError for malformed intersection types
	InvalidIntersectionSyntaxError

	// InvalidNegationSyntaxError for malformed negation types
	InvalidNegationSyntaxError

	// DeepNestingError for excessively nested types
	DeepNestingError
)

// Error implements the error interface
func (te *TypeError) Error() string {
	result := fmt.Sprintf("%d:%d: %s", te.Position.Line, te.Position.Column, te.Message)
	if te.Context != "" {
		result += fmt.Sprintf(" (in %s)", te.Context)
	}
	if te.Suggestion != "" {
		result += fmt.Sprintf(" - %s", te.Suggestion)
	}
	return result
}

// TypeErrorRecovery provides error recovery strategies for type parsing
type TypeErrorRecovery struct {
	// maxNestingDepth prevents stack overflow from deeply nested types
	maxNestingDepth int

	// syncTokens are tokens where parsing can be safely resumed
	syncTokens map[string]bool

	// currentDepth tracks current nesting depth
	currentDepth int
}

// NewTypeErrorRecovery creates a new type error recovery handler
func NewTypeErrorRecovery() *TypeErrorRecovery {
	return &TypeErrorRecovery{
		maxNestingDepth: 20, // Reasonable limit for type nesting
		syncTokens: map[string]bool{
			";":      true, // End of statement
			"{":      true, // Start of block
			"}":      true, // End of block
			"my":     true, // Variable declaration
			"our":    true, // Variable declaration
			"sub":    true, // Subroutine
			"method": true, // Method
			"class":  true, // Class
			"role":   true, // Role
		},
		currentDepth: 0,
	}
}

// RecoverFromTypeError attempts to recover from a type parsing error
func (ter *TypeErrorRecovery) RecoverFromTypeError(source string, position ast.Position, context string) *TypeError {
	// Analyze the source to determine the type of error
	errorCode, message, suggestion := ter.analyzeTypeError(source, position)

	return &TypeError{
		Message:    message,
		Position:   position,
		Suggestion: suggestion,
		Context:    context,
		ErrorCode:  errorCode,
		Source:     source,
	}
}

// analyzeTypeError analyzes the source to determine the specific type error
func (ter *TypeErrorRecovery) analyzeTypeError(source string, position ast.Position) (TypeErrorCode, string, string) {
	// Extract the relevant portion of source around the error
	lines := strings.Split(source, "\n")
	if position.Line <= 0 || position.Line > len(lines) {
		return UnknownTypeError, "Unknown type parsing error", "Check syntax"
	}

	line := lines[position.Line-1]

	// Check for common error patterns in order of specificity

	// Deep nesting check
	if strings.Count(line, "[") > 20 || strings.Count(line, "ArrayRef") > 10 {
		return DeepNestingError,
			"Type nesting too deep",
			"Reduce type nesting to less than 20 levels"
	}

	// Type assertion errors (highest priority for 'as' keyword)
	if strings.Contains(line, " as ") && (strings.HasSuffix(strings.TrimSpace(line), "as") ||
		strings.HasSuffix(strings.TrimSpace(line), "as ;")) {
		return IncompleteTypeAssertionError,
			"Incomplete type assertion - missing target type",
			"Add the target type after 'as' keyword"
	}

	// Negation syntax errors (before general pattern checks)
	if strings.Contains(line, "!!") && containsTypeLikePattern(line) {
		return InvalidNegationSyntaxError,
			"Invalid negation type syntax - use single '!' before type",
			"Change '!!' to '!' for negation types"
	}

	// Missing closing bracket
	if strings.Contains(line, "ArrayRef[") && !strings.Contains(line, "]") {
		return MissingClosingBracketError,
			"Missing closing bracket in parameterized type",
			"Add closing ']' to complete the parameterized type"
	}

	// Union syntax errors
	if strings.Contains(line, "||") && containsTypeLikePattern(line) {
		return InvalidUnionSyntaxError,
			"Invalid union type syntax - use single '|' between types",
			"Change '||' to '|' for union types"
	}

	// Parameterized type errors
	if strings.Contains(line, "ArrayRef[") && strings.Contains(line, "ArrayRef[ ") {
		return InvalidParameterizedTypeError,
			"Invalid parameterized type - unexpected space after '['",
			"Remove space after '[' in parameterized type"
	}

	// Where clause errors
	if strings.Contains(line, "where") && strings.Contains(line, "where $") {
		return InvalidWhereClauseError,
			"Invalid where clause syntax",
			"Use 'where Type: Constraint' syntax for type constraints"
	}

	// Intersection syntax errors
	if strings.Contains(line, "&&") && containsTypeLikePattern(line) {
		return InvalidIntersectionSyntaxError,
			"Invalid intersection type syntax - use single '&' between types",
			"Change '&&' to '&' for intersection types"
	}

	// Missing type name (lowest priority - most general)
	if strings.Contains(line, "my ") && strings.Contains(line, "$") &&
		!containsValidTypeName(line) {
		return MissingTypeNameError,
			"Missing or invalid type name in variable declaration",
			"Add a valid type name before the variable"
	}

	return UnknownTypeError, "Syntax error in type expression", "Check type syntax"
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

// CheckNestingDepth validates that type nesting doesn't exceed limits
func (ter *TypeErrorRecovery) CheckNestingDepth(depth int, position ast.Position) *TypeError {
	if depth > ter.maxNestingDepth {
		return &TypeError{
			Message:    fmt.Sprintf("Type nesting too deep (%d levels)", depth),
			Position:   position,
			Suggestion: fmt.Sprintf("Reduce type nesting to less than %d levels", ter.maxNestingDepth),
			Context:    "type expression",
			ErrorCode:  DeepNestingError,
		}
	}
	return nil
}

// FindSynchronizationPoint finds the next safe point to resume parsing
func (ter *TypeErrorRecovery) FindSynchronizationPoint(source string, position ast.Position) ast.Position {
	lines := strings.Split(source, "\n")
	if position.Line <= 0 || position.Line > len(lines) {
		return position
	}

	// Start from current position and look for sync tokens
	for lineNum := position.Line; lineNum <= len(lines); lineNum++ {
		if lineNum > len(lines) {
			break
		}

		line := lines[lineNum-1]
		startCol := 0
		if lineNum == position.Line {
			// On the same line, start searching after current position
			startCol = position.Column - 1 // Convert to 0-based
			if startCol < 0 {
				startCol = 0
			}
		}

		// Find the earliest sync token
		bestCol := -1
		bestToken := ""
		for token := range ter.syncTokens {
			// Find token starting from the appropriate position
			col := strings.Index(line[startCol:], token)
			if col != -1 && (bestCol == -1 || col < bestCol) {
				bestCol = col
				bestToken = token
			}
		}

		if bestCol != -1 {
			// Found a synchronization point
			actualCol := startCol + bestCol
			return ast.Position{
				Line:   lineNum,
				Column: actualCol + len(bestToken) + 1, // +1 for 1-based column numbering
				Offset: position.Offset + len(line[:actualCol]) + len(bestToken),
			}
		}
	}

	// If no sync point found, go to end of current line
	if position.Line <= len(lines) {
		line := lines[position.Line-1]
		return ast.Position{
			Line:   position.Line,
			Column: len(line) + 1, // +1 for 1-based column numbering
			Offset: position.Offset + len(line) - position.Column,
		}
	}

	return position
}

// ValidateTypeExpression performs comprehensive validation of a type expression
func (ter *TypeErrorRecovery) ValidateTypeExpression(expr *ast.TypeExpression, source string) []*TypeError {
	var errors []*TypeError

	if expr == nil {
		return errors
	}

	// Check for deep nesting in parameterized types
	depth := ter.calculateTypeDepth(expr)
	// Only check depth if expr has valid position information
	pos := expr.Start()
	if pos.Line > 0 && pos.Column >= 0 {
		if depthError := ter.CheckNestingDepth(depth, pos); depthError != nil {
			errors = append(errors, depthError)
		}
	}

	// Validate union types
	if expr.IsUnion && len(expr.UnionTypes) < 2 {
		errors = append(errors, &TypeError{
			Message:    "Union type must have at least 2 alternatives",
			Position:   expr.Start(),
			Suggestion: "Add more types to the union or remove '|' syntax",
			Context:    "union type",
			ErrorCode:  InvalidUnionSyntaxError,
			Source:     source,
		})
	}

	// Validate intersection types
	if expr.IsIntersection && len(expr.IntersectionTypes) < 2 {
		errors = append(errors, &TypeError{
			Message:    "Intersection type must have at least 2 components",
			Position:   expr.Start(),
			Suggestion: "Add more types to the intersection or remove '&' syntax",
			Context:    "intersection type",
			ErrorCode:  InvalidIntersectionSyntaxError,
			Source:     source,
		})
	}

	// Validate parameterized types
	if len(expr.Parameters) > 0 && expr.Name == "" {
		errors = append(errors, &TypeError{
			Message:    "Parameterized type must have a base type name",
			Position:   expr.Start(),
			Suggestion: "Add the base type name before the brackets",
			Context:    "parameterized type",
			ErrorCode:  InvalidParameterizedTypeError,
			Source:     source,
		})
	}

	return errors
}

// calculateTypeDepth calculates the maximum nesting depth of a type expression
func (ter *TypeErrorRecovery) calculateTypeDepth(expr *ast.TypeExpression) int {
	if expr == nil {
		return 0
	}

	maxDepth := 1

	// Check parameters
	for _, param := range expr.Parameters {
		paramDepth := ter.calculateTypeDepth(param)
		if paramDepth+1 > maxDepth {
			maxDepth = paramDepth + 1
		}
	}

	// Check union types
	for _, unionType := range expr.UnionTypes {
		unionDepth := ter.calculateTypeDepth(unionType)
		if unionDepth > maxDepth {
			maxDepth = unionDepth
		}
	}

	// Check intersection types
	for _, intersectionType := range expr.IntersectionTypes {
		intersectionDepth := ter.calculateTypeDepth(intersectionType)
		if intersectionDepth > maxDepth {
			maxDepth = intersectionDepth
		}
	}

	return maxDepth
}

// CreateErrorWithSuggestion creates a type error with helpful suggestions
func CreateErrorWithSuggestion(message, suggestion, context string, position ast.Position, code TypeErrorCode) *TypeError {
	return &TypeError{
		Message:    message,
		Position:   position,
		Suggestion: suggestion,
		Context:    context,
		ErrorCode:  code,
	}
}
