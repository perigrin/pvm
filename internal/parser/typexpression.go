// ABOUTME: Type expression parsing for PSC
// ABOUTME: Handles parsing of type annotations and expressions

package parser

// This file is intentionally simplified to use the main implementation in parser.go
// In the future, we'll implement a more sophisticated parser using tree-sitter

// parseTypeExpression parses a type expression string and returns a TypeExpression
// This function is intentionally unused as it delegates to ParseTypeExpression in types.go
// nolint:unused
func parseTypeExpression(text string, pos Position) (*TypeExpression, error) {
	// Delegate to the implementation in parser.go
	return ParseTypeExpression(text, pos)
}
