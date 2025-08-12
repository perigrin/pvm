// ABOUTME: Type error identification for tree-sitter parse failures
// ABOUTME: Creates structured error objects without formatting - presentation handled by consumers

package parser

import (
	"regexp"
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

// IdentifyMalformedTypeInAST analyzes a parsed AST to detect malformed type expressions
// This catches cases where tree-sitter partially parses but creates structures with embedded ERROR nodes
// Uses conservative precision to avoid false positives on valid Perl syntax
func (tei *TypeErrorIdentifier) IdentifyMalformedTypeInAST(node ast.Node, source string) *errors.TypeParseError {
	if node == nil {
		return nil
	}

	// Check if this specific node is in a type context and validate it
	if tei.isInTypeContext(node, source) {
		// Check if this node or its children contain ERROR nodes in type contexts
		if errorNode := tei.findTypeContextError(node, source); errorNode != nil {
			// Apply additional validation to ensure this is truly a type error
			if tei.isConfirmedTypeError(errorNode, node, source) {
				return errorNode
			}
		}
	}

	// Always recursively check children - they might be in type contexts even if parent isn't
	for _, child := range node.Children() {
		if childError := tei.IdentifyMalformedTypeInAST(child, source); childError != nil {
			return childError
		}
	}

	return nil
}

// isInTypeContext determines if a node is in a clear type annotation context
// This helps avoid false positives on regular Perl syntax
func (tei *TypeErrorIdentifier) isInTypeContext(node ast.Node, source string) bool {
	nodeType := node.Type()

	// Clear type contexts
	typeContexts := []string{
		"typed_variable_declaration",
		"type_expression",
		"parameterized_type",
		"union_type",
		"intersection_type",
		"negation_type",
		"type_assertion",
		"field_declaration", // with type annotations
		"method_signature",  // with return type annotations
	}

	for _, context := range typeContexts {
		if nodeType == context {
			return true
		}
	}

	// Check if parent nodes indicate type context
	if tei.hasTypeContextParent(node) {
		return true
	}

	// For variable declarations, only consider them type contexts if they have type annotations
	if nodeType == "variable_declaration" {
		return tei.hasTypeAnnotations(node, source)
	}

	return false
}

// isConfirmedTypeError applies additional validation to confirm this is truly a type error
// and not a false positive from valid Perl syntax
func (tei *TypeErrorIdentifier) isConfirmedTypeError(errorNode *errors.TypeParseError, contextNode ast.Node, source string) bool {
	if errorNode == nil {
		return false
	}

	// Extract the text around the error for analysis
	errorText := tei.extractNodeText(contextNode, source)

	// Check for known false positive patterns and exclude them
	if tei.isKnownValidPerlPattern(errorText) {
		return false
	}

	// Only confirm errors that match known malformed type patterns
	return tei.isKnownMalformedTypePattern(errorText)
}

// hasTypeContextParent checks if any parent node indicates a type context
func (tei *TypeErrorIdentifier) hasTypeContextParent(node ast.Node) bool {
	// In a real implementation, we'd walk up the AST tree
	// For now, we'll use heuristics based on the node's properties
	nodeType := node.Type()

	// If this node is a child of a type-related node, it's in type context
	typeParentIndicators := []string{
		"type_",     // Any node starting with "type_"
		"typed_",    // Any node starting with "typed_"
		"field_",    // Field declarations
		"method_",   // Method declarations
		"signature", // Method signatures
	}

	for _, indicator := range typeParentIndicators {
		if strings.Contains(nodeType, indicator) {
			return true
		}
	}

	return false
}

// hasTypeAnnotations checks if a variable declaration has type annotations
func (tei *TypeErrorIdentifier) hasTypeAnnotations(node ast.Node, source string) bool {
	for _, child := range node.Children() {
		childType := child.Type()
		if childType == "type_expression" ||
			childType == "parameterized_type" ||
			childType == "union_type" ||
			childType == "intersection_type" {
			return true
		}
	}

	// Also check the source text for type annotation patterns
	nodeText := tei.extractNodeText(node, source)
	return tei.containsTypeAnnotationPattern(nodeText)
}

// isKnownValidPerlPattern checks if text represents valid Perl syntax that should not be flagged
func (tei *TypeErrorIdentifier) isKnownValidPerlPattern(text string) bool {
	// Patterns that are valid Perl but might trigger false positives

	// Package-qualified variable declarations
	if strings.Contains(text, "::") &&
		(strings.Contains(text, "our ") || strings.Contains(text, "my ")) {
		return true
	}

	// Logical operators in non-type contexts
	if (strings.Contains(text, "||") || strings.Contains(text, "&&")) &&
		!tei.containsTypeAnnotationPattern(text) {
		return true
	}

	// Regular expressions with brackets
	if strings.Contains(text, "qr/") || strings.Contains(text, "s/") ||
		strings.Contains(text, "m/") || strings.Contains(text, "tr/") {
		return true
	}

	// Array/hash references in non-type contexts
	if (strings.Contains(text, "[") || strings.Contains(text, "{")) &&
		!tei.containsTypeAnnotationPattern(text) {
		return true
	}

	return false
}

// isKnownMalformedTypePattern checks if text matches patterns that are definitively malformed types
func (tei *TypeErrorIdentifier) isKnownMalformedTypePattern(text string) bool {
	// Only flag patterns that are clearly malformed type expressions
	// Be conservative - the grammar supports error recovery for many patterns

	// Type annotation contexts with clear errors
	if !tei.containsTypeAnnotationPattern(text) {
		return false // Not a type annotation, so not a type error
	}

	// Be more conservative - only flag truly egregious errors, not patterns that grammar can recover from

	// Double operators in type contexts (||, &&, !!) - these are clearly wrong
	if tei.containsTypeAnnotationPattern(text) {
		if strings.Contains(text, "||") || strings.Contains(text, "&&") || strings.Contains(text, "!!") {
			return true
		}
	}

	// Incomplete type assertions - clearly incomplete syntax
	if strings.Contains(text, " as ;") || strings.HasSuffix(strings.TrimSpace(text), " as") {
		return true
	}

	// NOTE: Removed missing bracket and spacing checks as the grammar supports error recovery for these patterns
	// The parser can successfully handle:
	// - Missing closing brackets: ArrayRef[Int $var; -> parser recovers
	// - Spacing in parameterized types: ArrayRef[ Int] -> parser accepts this

	return false
}

// containsTypeAnnotationPattern checks if text contains patterns indicating type annotations
func (tei *TypeErrorIdentifier) containsTypeAnnotationPattern(text string) bool {
	// Look for clear type annotation indicators
	typeIndicators := []string{
		" as ",      // Type assertions
		"ArrayRef[", // Parameterized types
		"HashRef[",
		"Maybe[",
		"Undef", // Type names
		"Object",
		"Serializable",
		"field ",   // Field declarations
		"returns ", // Method return types
	}

	for _, indicator := range typeIndicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}

	// Check for typed variable declarations (my Type $var)
	if matched, _ := regexp.MatchString(`\b(my|our|state|field)\s+[A-Z][A-Za-z0-9]*\s+\$`, text); matched {
		return true
	}

	// Check for union/intersection types, but only with type names
	if (strings.Contains(text, "|") || strings.Contains(text, "&")) &&
		tei.containsValidTypeName(text) {
		return true
	}

	return false
}

// containsValidTypeName checks if the text contains valid type names (helper method)
func (tei *TypeErrorIdentifier) containsValidTypeName(text string) bool {
	// Look for capitalized words that could be type names
	words := strings.Fields(text)
	for _, word := range words {
		// Remove punctuation for checking
		cleanWord := strings.Trim(word, "[](){}|&!;,")
		if len(cleanWord) > 0 && cleanWord[0] >= 'A' && cleanWord[0] <= 'Z' {
			// Check if it's a reasonable type name length
			if len(cleanWord) >= 2 && len(cleanWord) <= 30 {
				return true
			}
		}
	}
	return false
}

// findTypeContextError looks for ERROR nodes in type-related contexts and classifies them
func (tei *TypeErrorIdentifier) findTypeContextError(node ast.Node, source string) *errors.TypeParseError {
	nodeType := node.Type()

	// Check for ERROR nodes and content validation in typed variable declarations
	if nodeType == "typed_variable_declaration" {
		// First check for ERROR children
		for _, child := range node.Children() {
			if child.Type() == "ERROR" {
				return tei.classifyTypedVarDeclError(node, child, source)
			}
		}

		// Also check for malformed parameterized types (even if no ERROR nodes)
		if malformedErr := tei.validateTypedVarDeclaration(node, source); malformedErr != nil {
			return malformedErr
		}
	}

	// Check variable declarations that have type expressions (parsed as variable_declaration + type_expression)
	if nodeType == "variable_declaration" {
		hasTypeExpr := false
		for _, child := range node.Children() {
			if child.Type() == "type_expression" {
				hasTypeExpr = true
				// Check for ERROR nodes in the type expression
				for _, grandchild := range child.Children() {
					if grandchild.Type() == "ERROR" {
						return tei.classifyTypeExpressionError(child, grandchild, source)
					}
				}
				break
			}
		}

		// If it has type expressions, validate the content even without ERROR nodes
		if hasTypeExpr {
			if malformedErr := tei.validateTypedVarDeclaration(node, source); malformedErr != nil {
				return malformedErr
			}
		}
	}

	// Note: variable_declaration handling is done above

	// Check for malformed assignment expressions and general syntax (like incomplete type assertions)
	if nodeType == "assignment_expression" || nodeType == "expression_statement" ||
		nodeType == "var_decl" || nodeType == "expression_stmt" {
		if malformedErr := tei.validateGeneralSyntax(node, source); malformedErr != nil {
			return malformedErr
		}
	}

	// Check for ERROR nodes in type expressions
	if nodeType == "type_expression" || nodeType == "parameterized_type" ||
		nodeType == "union_type" || nodeType == "intersection_type" {
		for _, child := range node.Children() {
			if child.Type() == "ERROR" {
				return tei.classifyTypeExpressionError(node, child, source)
			}
		}
	}

	// Check if this node itself is an ERROR in a type context
	// Only classify ERROR nodes that are confirmed to be in type contexts
	if nodeType == "ERROR" {
		// Only classify if we're in a confirmed type context and match malformed patterns
		if tei.isInTypeContext(node, source) {
			nodeText := tei.extractNodeText(node, source)

			// Only flag if this matches known malformed type patterns
			if tei.isKnownMalformedTypePattern(nodeText) && !tei.isKnownValidPerlPattern(nodeText) {
				return tei.classifyGenericTypeError(node, source)
			}
		}
		return nil
	}

	// Also check children for ERROR nodes in specific contexts
	for _, child := range node.Children() {
		if child.Type() == "ERROR" {
			// Only classify as type error if we're in a clear type context with confirmed malformed patterns
			if nodeType == "type_expression" || nodeType == "parameterized_type" ||
				nodeType == "union_type" || nodeType == "intersection_type" {

				childText := tei.extractNodeText(child, source)
				// Apply precision validation
				if tei.isKnownMalformedTypePattern(childText) && !tei.isKnownValidPerlPattern(childText) {
					return tei.classifyTypeExpressionError(node, child, source)
				}
			}
			// Otherwise, don't classify as type error
		}
	}

	return nil
}

// validateTypedVarDeclaration checks for malformed typed variable declarations even without ERROR nodes
func (tei *TypeErrorIdentifier) validateTypedVarDeclaration(node ast.Node, source string) *errors.TypeParseError {
	position := node.Start()

	// Get the text of the node from the source using its position
	// Since tree-sitter nodes might not have text populated, extract from source
	nodeText := tei.extractNodeText(node, source)

	// Apply precision validation - only flag if this is confirmed to be a type pattern
	if !tei.containsTypeAnnotationPattern(nodeText) {
		return nil // Not a type annotation, don't validate as type error
	}

	// Additional validation to avoid false positives
	if tei.isKnownValidPerlPattern(nodeText) {
		return nil // Known valid Perl pattern, don't flag as error
	}

	// Only flag truly malformed patterns, not those that grammar can recover from

	// Check for invalid union syntax (||) - only in confirmed type context
	if strings.Contains(nodeText, "||") && tei.containsValidTypeName(nodeText) {
		return &errors.TypeParseError{
			ErrorType:  "InvalidUnionSyntaxError",
			Message:    "Invalid union type syntax - use single '|' between types",
			Position:   position,
			Suggestion: "Change '||' to '|' for union types",
			Context:    "union type expression",
			ErrorCode:  errors.InvalidUnionSyntaxError,
			Source:     source,
			SourceLine: tei.getSourceLine(source, position.Line),
		}
	}

	// Check for invalid intersection syntax (&&) - only in confirmed type context
	if strings.Contains(nodeText, "&&") && tei.containsValidTypeName(nodeText) {
		return &errors.TypeParseError{
			ErrorType:  "InvalidIntersectionSyntaxError",
			Message:    "Invalid intersection type syntax - use single '&' between types",
			Position:   position,
			Suggestion: "Change '&&' to '&' for intersection types",
			Context:    "intersection type expression",
			ErrorCode:  errors.InvalidIntersectionSyntaxError,
			Source:     source,
			SourceLine: tei.getSourceLine(source, position.Line),
		}
	}

	// Check for invalid negation syntax (!!) - only in confirmed type context
	if strings.Contains(nodeText, "!!") && tei.containsValidTypeName(nodeText) {
		return &errors.TypeParseError{
			ErrorType:  "InvalidNegationSyntaxError",
			Message:    "Invalid negation type syntax - use single '!' before type",
			Position:   position,
			Suggestion: "Change '!!' to '!' for negation types",
			Context:    "negation type expression",
			ErrorCode:  errors.InvalidNegationSyntaxError,
			Source:     source,
			SourceLine: tei.getSourceLine(source, position.Line),
		}
	}

	// NOTE: Removed bracket and spacing checks as the grammar supports error recovery
	// The parser handles these patterns gracefully:
	// - Missing brackets: grammar recovers and parses successfully
	// - Spacing variations: grammar accepts different spacing patterns

	return nil
}

// validateGeneralSyntax checks for malformed syntax patterns like incomplete type assertions
func (tei *TypeErrorIdentifier) validateGeneralSyntax(node ast.Node, source string) *errors.TypeParseError {
	position := node.Start()

	// Get the text of the node from the source using its position
	nodeText := tei.extractNodeText(node, source)

	// Apply precision validation - avoid false positives
	if tei.isKnownValidPerlPattern(nodeText) {
		return nil // Known valid Perl pattern, don't flag as error
	}

	// Check for incomplete type assertion: "as ;" pattern or just "as " at end
	// This is a clear type assertion context, so safe to flag
	if strings.Contains(nodeText, " as ;") || strings.Contains(nodeText, "as ;") {
		return &errors.TypeParseError{
			ErrorType:  "IncompleteTypeAssertionError",
			Message:    "Incomplete type assertion - missing target type",
			Position:   position,
			Suggestion: "Add the target type after 'as' keyword",
			Context:    "type assertion",
			ErrorCode:  errors.IncompleteTypeAssertionError,
			Source:     source,
			SourceLine: tei.getSourceLine(source, position.Line),
		}
	}

	// Check for incomplete type assertion: "as " at end of node (fragment pattern)
	// Only flag if this appears to be in a type context
	if (strings.HasSuffix(strings.TrimSpace(nodeText), "as") || strings.HasSuffix(strings.TrimSpace(nodeText), "as ")) &&
		tei.containsTypeAnnotationPattern(nodeText) {
		return &errors.TypeParseError{
			ErrorType:  "IncompleteTypeAssertionError",
			Message:    "Incomplete type assertion - missing target type",
			Position:   position,
			Suggestion: "Add the target type after 'as' keyword",
			Context:    "type assertion",
			ErrorCode:  errors.IncompleteTypeAssertionError,
			Source:     source,
			SourceLine: tei.getSourceLine(source, position.Line),
		}
	}

	// Check for arrow syntax patterns (method name() -> Type or sub name() -> Type)
	if tei.containsArrowSyntax(nodeText) {
		return &errors.TypeParseError{
			ErrorType:  "ArrowSyntaxError",
			Message:    "Arrow syntax is not supported",
			Position:   position,
			Suggestion: tei.suggestCorrectSyntax(nodeText),
			Context:    "method or subroutine declaration",
			ErrorCode:  errors.ArrowSyntaxError,
			Source:     source,
			SourceLine: tei.getSourceLine(source, position.Line),
		}
	}

	return nil
}

// extractNodeText extracts the text content of a node from the source using its position
func (tei *TypeErrorIdentifier) extractNodeText(node ast.Node, source string) string {
	if node == nil {
		return ""
	}

	start := node.Start()
	end := node.End()

	// Convert source to lines for easier extraction
	lines := strings.Split(source, "\n")

	// Handle single-line nodes
	if start.Line == end.Line {
		if start.Line-1 < len(lines) {
			line := lines[start.Line-1]
			if start.Column-1 < len(line) && end.Column <= len(line) {
				return line[start.Column-1 : end.Column]
			}
		}
		return ""
	}

	// Handle multi-line nodes
	var result strings.Builder
	for lineNum := start.Line; lineNum <= end.Line; lineNum++ {
		if lineNum-1 >= len(lines) {
			break
		}
		line := lines[lineNum-1]

		if lineNum == start.Line {
			// First line - start from start.Column
			if start.Column-1 < len(line) {
				result.WriteString(line[start.Column-1:])
			}
		} else if lineNum == end.Line {
			// Last line - end at end.Column
			if end.Column <= len(line) {
				result.WriteString(line[:end.Column])
			}
		} else {
			// Middle lines - take the whole line
			result.WriteString(line)
		}

		// Add newline except for the last line
		if lineNum < end.Line {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// isKnownMalformedPattern checks if the text matches clearly malformed patterns we should catch
func (tei *TypeErrorIdentifier) isKnownMalformedPattern(nodeText, source string) bool {
	// Only catch patterns that are definitely malformed, not complex valid syntax

	// Missing closing brackets (clear error)
	if (strings.Contains(nodeText, "ArrayRef[") || strings.Contains(nodeText, "HashRef[")) &&
		!strings.Contains(nodeText, "]") && strings.Contains(nodeText, ";") {
		return true
	}

	// Double operators (clear errors)
	if strings.Contains(nodeText, "||") || strings.Contains(nodeText, "&&") || strings.Contains(nodeText, "!!") {
		return true
	}

	// Incomplete type assertions (clear error)
	if strings.Contains(nodeText, " as ;") || strings.HasSuffix(strings.TrimSpace(nodeText), " as") {
		return true
	}

	// Invalid spacing in parameterized types (clear error)
	if strings.Contains(nodeText, "[ ") && strings.Contains(nodeText, "]") {
		return true
	}

	// Don't flag other patterns as malformed - they may be valid complex syntax
	// that the grammar doesn't support yet
	return false
}

// classifyTypedVarDeclError classifies errors within typed variable declarations
func (tei *TypeErrorIdentifier) classifyTypedVarDeclError(parent, errorNode ast.Node, source string) *errors.TypeParseError {
	parentText := parent.Text()
	position := errorNode.Start()

	// Missing closing bracket in parameterized types
	if strings.Contains(parentText, "[") && !strings.Contains(parentText, "]") {
		return &errors.TypeParseError{
			ErrorType:  "MissingClosingBracketError",
			Message:    "Missing closing bracket in parameterized type",
			Position:   position,
			Suggestion: "Add closing ']' to complete the parameterized type",
			Context:    "parameterized type in variable declaration",
			ErrorCode:  errors.MissingClosingBracketError,
			Source:     source,
			SourceLine: tei.getSourceLine(source, position.Line),
		}
	}

	// Invalid spacing in parameterized types
	if strings.Contains(parentText, "[ ") {
		return &errors.TypeParseError{
			ErrorType:  "InvalidParameterizedTypeError",
			Message:    "Invalid spacing in parameterized type",
			Position:   position,
			Suggestion: "Remove space after '[' in parameterized type",
			Context:    "parameterized type in variable declaration",
			ErrorCode:  errors.InvalidParameterizedTypeError,
			Source:     source,
			SourceLine: tei.getSourceLine(source, position.Line),
		}
	}

	return tei.classifyGenericTypeError(errorNode, source)
}

// classifyTypeExpressionError classifies errors within type expressions
func (tei *TypeErrorIdentifier) classifyTypeExpressionError(parent, errorNode ast.Node, source string) *errors.TypeParseError {
	parentText := parent.Text()
	parentType := parent.Type()
	position := errorNode.Start()

	// Union type with invalid syntax
	if parentType == "union_type" && strings.Contains(parentText, "||") {
		return &errors.TypeParseError{
			ErrorType:  "InvalidUnionSyntaxError",
			Message:    "Invalid union type syntax - use single '|' between types",
			Position:   position,
			Suggestion: "Change '||' to '|' for union types",
			Context:    "union type expression",
			ErrorCode:  errors.InvalidUnionSyntaxError,
			Source:     source,
			SourceLine: tei.getSourceLine(source, position.Line),
		}
	}

	// Intersection type with invalid syntax
	if parentType == "intersection_type" && strings.Contains(parentText, "&&") {
		return &errors.TypeParseError{
			ErrorType:  "InvalidIntersectionSyntaxError",
			Message:    "Invalid intersection type syntax - use single '&' between types",
			Position:   position,
			Suggestion: "Change '&&' to '&' for intersection types",
			Context:    "intersection type expression",
			ErrorCode:  errors.InvalidIntersectionSyntaxError,
			Source:     source,
			SourceLine: tei.getSourceLine(source, position.Line),
		}
	}

	return tei.classifyGenericTypeError(errorNode, source)
}

// classifyGenericTypeError provides a fallback classification for unspecified type errors
func (tei *TypeErrorIdentifier) classifyGenericTypeError(errorNode ast.Node, source string) *errors.TypeParseError {
	position := errorNode.Start()

	return &errors.TypeParseError{
		ErrorType:  "UnknownTypeError",
		Message:    "Syntax error in type expression",
		Position:   position,
		Suggestion: "Check type syntax for malformed expressions",
		Context:    "type expression",
		ErrorCode:  errors.UnknownTypeError,
		Source:     source,
		SourceLine: tei.getSourceLine(source, position.Line),
	}
}

// getSourceLine extracts a specific line from source code
func (tei *TypeErrorIdentifier) getSourceLine(source string, lineNum int) string {
	lines := strings.Split(source, "\n")
	if lineNum <= 0 || lineNum > len(lines) {
		return ""
	}
	return lines[lineNum-1]
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

	// Arrow syntax error check (high priority)
	if tei.containsArrowSyntax(line) {
		return errors.ArrowSyntaxError,
			"Arrow syntax is not supported",
			tei.suggestCorrectSyntax(line),
			"ArrowSyntaxError"
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
	// Enhanced heuristic: look for capitalized words that could be type names
	// Also handle complex type expressions with parentheses, brackets, etc.

	// First try the simple approach - split by fields and look for capitalized words
	words := strings.Fields(line)
	for _, word := range words {
		if len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z' {
			// Could be a type name like Int, Str, ArrayRef, etc.
			return true
		}
	}

	// For complex expressions, scan for capitalized sequences within the line
	// This handles cases like "(ArrayRef[Int]|HashRef[Str])" where field splitting doesn't work
	for i := 0; i < len(line); i++ {
		if line[i] >= 'A' && line[i] <= 'Z' {
			// Found start of potential type name, check if it's a complete word
			j := i + 1
			for j < len(line) && ((line[j] >= 'A' && line[j] <= 'Z') ||
				(line[j] >= 'a' && line[j] <= 'z') ||
				(line[j] >= '0' && line[j] <= '9') || line[j] == '_') {
				j++
			}
			// If we found a sequence of 2+ characters starting with capital, it's likely a type
			if j-i >= 2 {
				return true
			}
			i = j - 1 // Continue scanning from where we left off
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

// containsArrowSyntax checks if the text contains arrow syntax patterns
func (tei *TypeErrorIdentifier) containsArrowSyntax(text string) bool {
	// Look for patterns like "method name() -> Type" or "sub name() -> Type"
	if strings.Contains(text, " -> ") {
		// Check if it's in a method or sub context
		return strings.Contains(text, "method ") || strings.Contains(text, "sub ")
	}
	return false
}

// suggestCorrectSyntax suggests the correct syntax based on arrow syntax pattern
func (tei *TypeErrorIdentifier) suggestCorrectSyntax(text string) string {
	if strings.Contains(text, "method ") {
		if strings.Contains(text, "()") {
			return "Use 'method Type name()' instead of 'method name() -> Type'"
		}
		return "Use 'method Type name(ParamType $param)' instead of 'method name(ParamType $param) -> Type'"
	}
	if strings.Contains(text, "sub ") {
		return "Use 'sub name()' or 'method Type name()' instead of 'sub name() -> Type'"
	}
	return "Use 'method Type name()' instead of 'method name() -> Type'"
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
