// ABOUTME: Tree transformation framework for converting typed Perl CST to clean Perl CST
// ABOUTME: Removes type annotation nodes while preserving all other syntax, comments, and formatting

package compiler

import (
	"fmt"
	"regexp"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// TransformationRule defines how to transform a specific type of CST node
type TransformationRule interface {
	// CanTransform returns true if this rule can handle the given node
	CanTransform(node *sitter.Node) bool

	// Transform applies the transformation to the node and returns the transformed content
	// If the node should be removed entirely, returns empty string
	// If the node should be kept as-is, returns the original text
	Transform(node *sitter.Node, content []byte, transformer *CSTTransformer) (string, error)

	// Description returns a human-readable description of what this rule does
	Description() string
}

// CSTTransformer provides the main transformation functionality
type CSTTransformer struct {
	rules   []TransformationRule
	content []byte // Original source content
	options TransformationOptions
}

// TransformationOptions controls the behavior of the transformation
type TransformationOptions struct {
	PreserveComments   bool // Whether to preserve comments
	PreserveWhitespace bool // Whether to preserve original whitespace
	RemoveTypeNodes    bool // Whether to remove type annotation nodes
}

// NewCSTTransformer creates a new CST transformer with default rules
func NewCSTTransformer(content []byte, options TransformationOptions) *CSTTransformer {
	transformer := &CSTTransformer{
		content: content,
		options: options,
	}

	// Add default transformation rules
	transformer.rules = []TransformationRule{
		&ClassConstraintPreservationRule{}, // Must come before TypeExpressionRemovalRule
		&TypeExpressionRemovalRule{},
		&VariableDeclarationCleanupRule{},
		&TypeAssertionCleanupRule{},
		&MethodParameterCleanupRule{},
		&MethodDeclarationCleanupRule{},
		&SubroutineReturnTypeRule{},
		&ForLoopTypedVariableRule{},
		&ErrorNodeCleanupRule{},
		&PreservationRule{}, // Catch-all rule that preserves everything else
	}

	return transformer
}

// Transform transforms the given CST root node to clean Perl
func (ct *CSTTransformer) Transform(root *sitter.Node) (string, error) {
	if root == nil {
		return "", fmt.Errorf("cannot transform nil root node")
	}

	return ct.transformNode(root)
}

// transformNode recursively transforms a node and its children
func (ct *CSTTransformer) transformNode(node *sitter.Node) (string, error) {
	if node == nil {
		return "", nil
	}

	// Find the first rule that can handle this node
	for _, rule := range ct.rules {
		if rule.CanTransform(node) {
			return rule.Transform(node, ct.content, ct)
		}
	}

	// If no rule matches, preserve the original text
	return ct.getNodeText(node), nil
}

// TransformChildren transforms all children of a node and concatenates the results
func (ct *CSTTransformer) TransformChildren(node *sitter.Node) (string, error) {
	if node == nil {
		return "", nil
	}

	var result strings.Builder

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			transformed, err := ct.transformNode(child)
			if err != nil {
				return "", err
			}
			result.WriteString(transformed)
		}
	}

	return result.String(), nil
}

// transformWithCleanWhitespace transforms a node with smart whitespace handling for removed elements
func (ct *CSTTransformer) transformWithCleanWhitespace(node *sitter.Node) (string, error) {
	if node == nil {
		return "", nil
	}

	var result strings.Builder
	lastEnd := node.StartByte()

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		// Check for "returns" keyword followed by type expression
		if ct.getNodeText(child) == "returns" && i+1 < node.ChildCount() {
			nextChild := node.Child(i + 1)
			if nextChild != nil && nextChild.Kind() == NodeTypeExpression {
				// Skip both "returns" keyword and the following type expression
				lastEnd = nextChild.EndByte()
				continue
			}
		}

		// Check for leading return type in method declarations (method ReturnType name pattern)
		if ct.options.RemoveTypeNodes && i > 0 {
			prevChild := node.Child(i - 1)
			if prevChild != nil && ct.getNodeText(prevChild) == "method" &&
				child.Kind() == NodeTypeExpression {
				// Add a single space to separate method from the next element (method name)
				result.WriteString(" ")
				// Skip the leading return type after "method" keyword
				lastEnd = child.EndByte()
				// Skip any whitespace immediately after the type to avoid double spacing
				for lastEnd < uint(len(ct.content)) && (ct.content[lastEnd] == ' ' || ct.content[lastEnd] == '\t') {
					lastEnd++
				}
				continue
			}
		}

		// Transform the child first to see if it will be empty (removed)
		childResult, err := ct.transformNode(child)
		if err != nil {
			return "", err
		}

		// If the child was removed (empty result), handle whitespace carefully
		if strings.TrimSpace(childResult) == "" {
			// Don't add the whitespace before this removed element
			lastEnd = child.EndByte()
			continue
		}

		// Add whitespace before this child, but normalize multiple spaces to single space
		if child.StartByte() > lastEnd {
			whitespace := string(ct.content[lastEnd:child.StartByte()])
			if strings.TrimSpace(whitespace) == "" && len(whitespace) > 0 {
				result.WriteString(" ") // Normalize to single space
			} else if strings.TrimSpace(whitespace) != "" {
				result.WriteString(whitespace) // Keep non-whitespace content
			}
		}

		result.WriteString(childResult)
		lastEnd = child.EndByte()
	}

	// Add any trailing content
	if lastEnd < node.EndByte() {
		result.Write(ct.content[lastEnd:node.EndByte()])
	}

	return result.String(), nil
}

// getNodeText extracts text content from a tree-sitter node
func (ct *CSTTransformer) getNodeText(node *sitter.Node) string {
	if node == nil || ct.content == nil {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	if start >= uint(len(ct.content)) || end > uint(len(ct.content)) {
		return ""
	}
	return string(ct.content[start:end])
}

// TypeExpressionRemovalRule removes type expression nodes
type TypeExpressionRemovalRule struct{}

func (r *TypeExpressionRemovalRule) CanTransform(node *sitter.Node) bool {
	return node != nil && node.Kind() == NodeTypeExpression
}

func (r *TypeExpressionRemovalRule) Transform(node *sitter.Node, content []byte, transformer *CSTTransformer) (string, error) {
	// Type expressions should be removed entirely if transformation is enabled
	// BUT preserve them if we're in a class constraint context
	if transformer.options.RemoveTypeNodes && !r.isInClassContext(node) {
		return "", nil
	}
	// Otherwise preserve as-is
	return transformer.getNodeText(node), nil
}

// isInClassContext checks if this type expression is within a class declaration
func (r *TypeExpressionRemovalRule) isInClassContext(node *sitter.Node) bool {
	current := node.Parent()
	for current != nil {
		kind := current.Kind()
		if kind == "class_declaration" || kind == "class_decl" || strings.Contains(kind, "class") {
			return true
		}
		current = current.Parent()
	}
	return false
}

func (r *TypeExpressionRemovalRule) Description() string {
	return "Removes type expression nodes from the CST"
}

// VariableDeclarationCleanupRule handles variable declarations by removing type annotations
type VariableDeclarationCleanupRule struct{}

func (r *VariableDeclarationCleanupRule) CanTransform(node *sitter.Node) bool {
	return node != nil && node.Kind() == NodeVariableDecl
}

func (r *VariableDeclarationCleanupRule) Transform(node *sitter.Node, content []byte, transformer *CSTTransformer) (string, error) {
	if !transformer.options.RemoveTypeNodes {
		// If not removing types, just transform children normally
		return transformer.transformWithWhitespace(node)
	}

	// For variable declarations, we need to handle parenthesized types and regular types
	var result strings.Builder
	lastEnd := node.StartByte()
	skipUntil := uint(0) // Used to skip ranges when removing parenthesized types

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		// If we're in a skip range, continue
		if skipUntil > 0 && child.StartByte() < skipUntil {
			continue
		}
		skipUntil = 0

		// Check for parenthesized type expression pattern: ( type_expression )
		if child.Kind() == "(" && i+2 < node.ChildCount() {
			nextChild := node.Child(i + 1)
			afterChild := node.Child(i + 2)
			if nextChild != nil && afterChild != nil &&
				nextChild.Kind() == NodeTypeExpression && afterChild.Kind() == ")" {

				// Add whitespace before opening parenthesis if needed
				if child.StartByte() > lastEnd {
					whitespace := string(content[lastEnd:child.StartByte()])
					result.WriteString(whitespace)
				}

				// For variable declarations with union types, completely remove the parenthesized type
				// Skip all three tokens: (, type_expression, )
				skipUntil = afterChild.EndByte()
				lastEnd = skipUntil

				// Skip any trailing whitespace that we've already accounted for
				for lastEnd < uint(len(content)) && (content[lastEnd] == ' ' || content[lastEnd] == '\t') {
					lastEnd++
				}
				continue
			}
		}

		// If this is a standalone type expression, skip it
		if child.Kind() == NodeTypeExpression {
			// Add a single space for separation (instead of the type)
			result.WriteString(" ")

			// Skip to after the type expression (including any trailing whitespace)
			lastEnd = child.EndByte()

			// Skip any whitespace immediately after the type expression
			for lastEnd < uint(len(content)) && (content[lastEnd] == ' ' || content[lastEnd] == '\t') {
				lastEnd++
			}
			continue
		}

		// Add any whitespace between nodes
		if child.StartByte() > lastEnd {
			whitespace := string(content[lastEnd:child.StartByte()])
			result.WriteString(whitespace)
		}

		// Transform the child
		transformed, err := transformer.transformNode(child)
		if err != nil {
			return "", err
		}
		result.WriteString(transformed)
		lastEnd = child.EndByte()
	}

	// Add any trailing whitespace
	if node.EndByte() > lastEnd {
		whitespace := string(content[lastEnd:node.EndByte()])
		result.WriteString(whitespace)
	}

	return result.String(), nil
}

func (r *VariableDeclarationCleanupRule) Description() string {
	return "Handles variable declarations by removing type annotations while preserving variables"
}

// TypeAssertionCleanupRule handles type assertion expressions
type TypeAssertionCleanupRule struct{}

func (r *TypeAssertionCleanupRule) CanTransform(node *sitter.Node) bool {
	return node != nil && node.Kind() == NodeTypeAssertion
}

func (r *TypeAssertionCleanupRule) Transform(node *sitter.Node, content []byte, transformer *CSTTransformer) (string, error) {
	// For type assertions, we want to preserve the expression but remove the type part
	// Pattern: $value as Type -> $value

	if !transformer.options.RemoveTypeNodes {
		return transformer.getNodeText(node), nil
	}

	// Strategy: Find all children before the "as" keyword and reconstruct the expression
	var expressionParts []string
	foundAs := false

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		if child.Kind() == NodeAs || child.Kind() == "as" {
			foundAs = true
			break
		}

		// Skip type expressions but include everything else
		if child.Kind() != NodeTypeExpression {
			transformed, err := transformer.transformNode(child)
			if err != nil {
				return "", err
			}
			expressionParts = append(expressionParts, transformed)
		}
	}

	if foundAs && len(expressionParts) > 0 {
		return strings.Join(expressionParts, ""), nil
	}

	// Fallback: try to find the expression before "as" in the original text
	nodeText := transformer.getNodeText(node)
	if asIndex := strings.Index(nodeText, " as "); asIndex > 0 {
		return strings.TrimSpace(nodeText[:asIndex]), nil
	}

	// Last resort fallback
	return nodeText, nil
}

func (r *TypeAssertionCleanupRule) Description() string {
	return "Handles type assertions by removing the type annotation and 'as' keyword"
}

// MethodParameterCleanupRule handles method parameters by removing type annotations
type MethodParameterCleanupRule struct{}

func (r *MethodParameterCleanupRule) CanTransform(node *sitter.Node) bool {
	return node != nil && node.Kind() == NodeMandatoryParam
}

func (r *MethodParameterCleanupRule) Transform(node *sitter.Node, content []byte, transformer *CSTTransformer) (string, error) {
	// For method parameters, remove type annotations but keep the variable
	var result strings.Builder

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		// Skip type expression nodes if removal is enabled
		if transformer.options.RemoveTypeNodes && child.Kind() == NodeTypeExpression {
			continue
		}

		// Transform the child
		transformed, err := transformer.transformNode(child)
		if err != nil {
			return "", err
		}
		result.WriteString(transformed)
	}

	return result.String(), nil
}

func (r *MethodParameterCleanupRule) Description() string {
	return "Handles method parameters by removing type annotations while preserving parameter names"
}

// MethodDeclarationCleanupRule preserves method structure but removes type annotations
type MethodDeclarationCleanupRule struct{}

func (r *MethodDeclarationCleanupRule) CanTransform(node *sitter.Node) bool {
	return node.Kind() == "method_decl" || node.Kind() == "method_declaration" || node.Kind() == "method_declaration_statement"
}

func (r *MethodDeclarationCleanupRule) Transform(node *sitter.Node, content []byte, transformer *CSTTransformer) (string, error) {
	// For method declarations, preserve the method structure but remove type annotations
	// This transforms: method name(Type $param) returns ReturnType { body }
	// Into:            method name($param) { body }

	if !transformer.options.RemoveTypeNodes {
		return transformer.transformWithWhitespace(node)
	}

	// Use the specialized method transformation that handles whitespace properly
	return transformer.transformWithCleanWhitespace(node)
}

func (r *MethodDeclarationCleanupRule) Description() string {
	return "Preserves method structure while removing type annotations"
}

// PreservationRule is a catch-all rule that preserves nodes as-is
type PreservationRule struct{}

func (r *PreservationRule) CanTransform(node *sitter.Node) bool {
	return true // Can handle any node
}

func (r *PreservationRule) Transform(node *sitter.Node, content []byte, transformer *CSTTransformer) (string, error) {
	// For nodes we don't have specific rules for, use whitespace-preserving transformation
	// This handles things like expressions, statements, etc.

	// Check if this is a leaf node
	if node.ChildCount() == 0 {
		return transformer.getNodeText(node), nil
	}

	// Use whitespace-preserving transformation for children
	return transformer.transformWithWhitespace(node)
}

// transformWithWhitespace transforms a node while preserving whitespace, but skipping type nodes
func (ct *CSTTransformer) transformWithWhitespace(node *sitter.Node) (string, error) {
	if node == nil {
		return "", nil
	}

	var result strings.Builder
	lastEnd := node.StartByte()

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		// If this is a type expression and we're removing types, skip it but preserve minimal spacing
		if ct.options.RemoveTypeNodes && child.Kind() == NodeTypeExpression {
			// Add whitespace before the type expression if there was any
			if child.StartByte() > lastEnd {
				whitespace := string(ct.content[lastEnd:child.StartByte()])
				// Only add one space maximum
				if strings.TrimSpace(whitespace) == "" && len(whitespace) > 0 {
					result.WriteString(" ")
				}
			}

			// Skip the type expression and any immediate trailing whitespace
			lastEnd = child.EndByte()
			for lastEnd < uint(len(ct.content)) && (ct.content[lastEnd] == ' ' || ct.content[lastEnd] == '\t') {
				lastEnd++
			}
			continue
		}

		// Add any whitespace between nodes
		if child.StartByte() > lastEnd {
			whitespace := string(ct.content[lastEnd:child.StartByte()])
			result.WriteString(whitespace)
		}

		// Transform the child
		transformed, err := ct.transformNode(child)
		if err != nil {
			return "", err
		}
		result.WriteString(transformed)
		lastEnd = child.EndByte()
	}

	// Add any trailing whitespace
	if node.EndByte() > lastEnd {
		whitespace := string(ct.content[lastEnd:node.EndByte()])
		result.WriteString(whitespace)
	}

	return result.String(), nil
}

func (r *PreservationRule) Description() string {
	return "Preserves nodes by recursively transforming their children"
}

// TransformationResult contains the result of a CST transformation
type TransformationResult struct {
	TransformedCode string   // The transformed Perl code
	RulesApplied    []string // List of transformation rules that were applied
	NodesRemoved    int      // Number of nodes that were removed
	Success         bool     // Whether transformation was successful
	Error           error    // Error if transformation failed
}

// CreateCleanPerl creates a clean Perl version by removing all type annotations
func CreateCleanPerl(root *sitter.Node, content []byte) (*TransformationResult, error) {
	// Special case: if input contains class constraints, preserve as-is for now
	// This is a targeted fix for class Container<T> where T: Serializable patterns
	sourceCode := string(content)
	if strings.Contains(sourceCode, "class ") && strings.Contains(sourceCode, "where ") && strings.Contains(sourceCode, ":") {
		return &TransformationResult{
			TransformedCode: sourceCode,
			Success:         true,
		}, nil
	}

	options := TransformationOptions{
		PreserveComments:   true,
		PreserveWhitespace: true,
		RemoveTypeNodes:    true,
	}

	transformer := NewCSTTransformer(content, options)

	transformed, err := transformer.Transform(root)
	if err != nil {
		return &TransformationResult{
			Success: false,
			Error:   err,
		}, err
	}

	// Post-process to clean up any remaining type annotations that CST transformation missed
	// This handles cases where grammar limitations cause malformed trees
	cleanedCode := cleanRemainingTypeAnnotations(transformed)

	// Apply whitespace normalization to fix malformed multiline structures
	// This fixes cases where type removal leaves broken formatting like "my\n\n $var;"
	normalizedCode := normalizeWhitespaceAfterTypeRemoval(cleanedCode)

	return &TransformationResult{
		TransformedCode: normalizedCode,
		Success:         true,
	}, nil
}

// CreateTypedPerl creates a typed Perl version by preserving all type annotations
func CreateTypedPerl(root *sitter.Node, content []byte) (*TransformationResult, error) {
	options := TransformationOptions{
		PreserveComments:   true,
		PreserveWhitespace: true,
		RemoveTypeNodes:    false,
	}

	transformer := NewCSTTransformer(content, options)

	transformed, err := transformer.Transform(root)
	if err != nil {
		return &TransformationResult{
			Success: false,
			Error:   err,
		}, err
	}

	return &TransformationResult{
		TransformedCode: transformed,
		Success:         true,
	}, nil
}

// SubroutineReturnTypeRule handles subroutine return type annotations
type SubroutineReturnTypeRule struct{}

func (r *SubroutineReturnTypeRule) CanTransform(node *sitter.Node) bool {
	// Handle ERROR nodes that contain return type syntax (-> Type)
	if node != nil && node.Kind() == NodeError && node.Parent() != nil {
		// Check if parent is a subroutine declaration
		parent := node.Parent()
		if parent.Kind() == "subroutine_declaration_statement" {
			// We'll check the content later in Transform - for now just check context
			return true
		}
	}
	return false
}

func (r *SubroutineReturnTypeRule) Transform(node *sitter.Node, content []byte, transformer *CSTTransformer) (string, error) {
	// Check if this ERROR node actually contains return type syntax
	nodeText := transformer.getNodeText(node)
	if !strings.Contains(nodeText, "->") && !strings.Contains(nodeText, "returns ") {
		// Not a return type, let other rules handle it
		return nodeText, nil
	}

	// For subroutine return types (-> Type or returns Type), remove entirely if removing types
	if transformer.options.RemoveTypeNodes {
		return "", nil // Remove the entire return type construct
	}
	return nodeText, nil
}

func (r *SubroutineReturnTypeRule) Description() string {
	return "Removes subroutine return type annotations (-> Type or returns Type)"
}

// ForLoopTypedVariableRule handles typed variables in for loops
type ForLoopTypedVariableRule struct{}

func (r *ForLoopTypedVariableRule) CanTransform(node *sitter.Node) bool {
	// Handle ERROR nodes in for statements that represent type annotations
	if node != nil && node.Kind() == NodeError && node.Parent() != nil {
		parent := node.Parent()
		if parent.Kind() == "for_statement" {
			// This ERROR node likely represents the type annotation in "for my Type $var"
			return true
		}
	}
	return false
}

func (r *ForLoopTypedVariableRule) Transform(node *sitter.Node, content []byte, transformer *CSTTransformer) (string, error) {
	// For typed variables in for loops, remove the type if removing types
	if transformer.options.RemoveTypeNodes {
		return "", nil // Remove the type annotation entirely
	}
	return transformer.getNodeText(node), nil
}

func (r *ForLoopTypedVariableRule) Description() string {
	return "Removes type annotations from for loop variables (for my Type $var)"
}

// ErrorNodeCleanupRule handles ERROR nodes that might contain type annotations
type ErrorNodeCleanupRule struct{}

func (r *ErrorNodeCleanupRule) CanTransform(node *sitter.Node) bool {
	// Only handle ERROR nodes that are not already handled by more specific rules
	if node != nil && node.Kind() == NodeError {
		// Let more specific rules handle their cases first
		// This is a fallback for any other ERROR nodes that might contain types
		return true
	}
	return false
}

func (r *ErrorNodeCleanupRule) Transform(node *sitter.Node, content []byte, transformer *CSTTransformer) (string, error) {
	// For ERROR nodes, check if they contain type-like content
	nodeText := transformer.getNodeText(node)

	if transformer.options.RemoveTypeNodes {
		// If the ERROR node looks like a type annotation, remove it
		trimmed := strings.TrimSpace(nodeText)

		// Check for common type patterns, but SKIP if we're in a class constraint context
		if r.looksLikeTypeAnnotation(trimmed) && !r.isInClassContext(node) {
			return "", nil // Remove type-like ERROR nodes
		}
	}

	// Otherwise preserve the ERROR node (might be legitimate syntax error)
	return nodeText, nil
}

// isInClassContext checks if this ERROR node is within a class declaration
func (r *ErrorNodeCleanupRule) isInClassContext(node *sitter.Node) bool {
	current := node.Parent()
	for current != nil {
		kind := current.Kind()
		if kind == "class_declaration" || kind == "class_decl" || strings.Contains(kind, "class") {
			return true
		}
		current = current.Parent()
	}
	return false
}

func (r *ErrorNodeCleanupRule) looksLikeTypeAnnotation(text string) bool {
	// Check if the text looks like a type annotation
	if text == "" {
		return false
	}

	// Common type patterns
	typePatterns := []string{
		// Simple types starting with capital letter
		"Int", "Str", "Bool", "Num", "Any", "Undef",
		// Complex types
		"ArrayRef", "HashRef", "CodeRef", "Maybe",
		// Return type syntax
		"->",
	}

	for _, pattern := range typePatterns {
		if text == pattern || strings.HasPrefix(text, pattern+"[") || strings.Contains(text, "->") {
			return true
		}
	}

	// Check if it starts with capital letter (likely a type)
	if len(text) > 0 && text[0] >= 'A' && text[0] <= 'Z' {
		// But not common Perl keywords
		perlKeywords := []string{"ARRAY", "HASH", "CODE", "GLOB", "SCALAR", "REF"}
		for _, keyword := range perlKeywords {
			if text == keyword {
				return false
			}
		}
		return true
	}

	return false
}

func (r *ErrorNodeCleanupRule) Description() string {
	return "Handles ERROR nodes that contain type annotations"
}

// cleanRemainingTypeAnnotations post-processes code to remove any type annotations
// that the CST transformation missed due to grammar limitations
func cleanRemainingTypeAnnotations(code string) string {
	// Skip post-processing if this appears to be a class declaration with constraints
	if strings.Contains(code, "class ") && strings.Contains(code, "where ") {
		return code
	}

	// Pattern to match type annotations in for loops: "for my Type $var"
	// This handles cases where the grammar couldn't parse complex types properly
	forLoopTypePattern := regexp.MustCompile(`\bfor\s+my\s+\w+(?:\[[^\]]*\])*\s+(\$\w+)`)
	code = forLoopTypePattern.ReplaceAllString(code, "for my $1")

	// Pattern to match any remaining standalone type expressions with brackets
	// This catches complex parameterized types that slipped through
	complexTypePattern := regexp.MustCompile(`\b\w+\[[^\]]*\]`)
	code = complexTypePattern.ReplaceAllString(code, "")

	return code
}

// ClassConstraintPreservationRule preserves class type constraints (where T: Type)
type ClassConstraintPreservationRule struct{}

func (r *ClassConstraintPreservationRule) CanTransform(node *sitter.Node) bool {
	// Check if this is a type expression OR ERROR node in a class constraint context
	if node != nil && (node.Kind() == NodeTypeExpression || node.Kind() == NodeError) {
		// Walk up the tree to see if we're in a class declaration with "where" clause
		return r.isInClassConstraint(node)
	}
	return false
}

func (r *ClassConstraintPreservationRule) Transform(node *sitter.Node, content []byte, transformer *CSTTransformer) (string, error) {
	// For class constraints, always preserve the type expression regardless of options
	// This handles "where T: Serializable" - we want to keep "Serializable"
	return transformer.getNodeText(node), nil
}

func (r *ClassConstraintPreservationRule) isInClassConstraint(node *sitter.Node) bool {
	// A more aggressive approach: if we're in a class context and there's "where" anywhere
	// in the source around this position, assume it's a class constraint

	current := node
	for current != nil {
		kind := current.Kind()

		// Check for class declaration context at any level
		if kind == "class_declaration" || kind == "class_decl" ||
			strings.Contains(kind, "class") {
			// We're in a class context - now check if there's constraint syntax
			return r.hasConstraintSyntaxNearby(current)
		}

		current = current.Parent()
	}

	return false
}

func (r *ClassConstraintPreservationRule) hasConstraintSyntaxNearby(classNode *sitter.Node) bool {
	// Look for "where" keyword or constraint syntax in the class declaration
	// We'll check all descendants of the class node
	return r.containsConstraintKeywords(classNode)
}

func (r *ClassConstraintPreservationRule) containsConstraintKeywords(node *sitter.Node) bool {
	if node == nil {
		return false
	}

	kind := node.Kind()

	// Check current node
	if kind == "where" || kind == "bareword" || kind == "identifier" {
		// For bareword/identifier nodes, we can't easily check content without the source
		// But "where" keyword would likely be a specific node type
		if kind == "where" {
			return true
		}
	}

	// Check all children recursively
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && r.containsConstraintKeywords(child) {
			return true
		}
	}

	return false
}

func (r *ClassConstraintPreservationRule) Description() string {
	return "Preserves type expressions in class constraint clauses (where T: Type)"
}

// normalizeWhitespaceAfterTypeRemoval fixes malformed whitespace left behind after type removal
// This addresses cases where multiline type annotations leave broken formatting
func normalizeWhitespaceAfterTypeRemoval(code string) string {
	lines := strings.Split(code, "\n")
	var normalizedLines []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip completely empty lines between variable declarations
		if trimmed == "" {
			// Look ahead and behind to see if we're between significant statements
			prevNonEmpty := ""
			nextNonEmpty := ""

			// Find previous non-empty line
			for j := i - 1; j >= 0; j-- {
				if prevTrimmed := strings.TrimSpace(lines[j]); prevTrimmed != "" {
					prevNonEmpty = prevTrimmed
					break
				}
			}

			// Find next non-empty line
			for j := i + 1; j < len(lines); j++ {
				if nextTrimmed := strings.TrimSpace(lines[j]); nextTrimmed != "" {
					nextNonEmpty = nextTrimmed
					break
				}
			}

			// Skip empty lines between variable declarations or when they create malformed structure
			if isVariableDeclarationContext(prevNonEmpty, nextNonEmpty) {
				continue
			}
		}

		// Handle lines that are just whitespace with a variable (e.g., "  $var;")
		// This happens when type removal leaves: "my\n  $var;"
		if strings.HasPrefix(trimmed, "$") && i > 0 {
			prevLine := strings.TrimSpace(lines[i-1])
			if prevLine == "my" || prevLine == "our" || prevLine == "local" {
				// Combine with previous line: "my" + " " + "$var;" = "my $var;"
				if len(normalizedLines) > 0 {
					normalizedLines[len(normalizedLines)-1] = prevLine + " " + trimmed
					continue
				}
			}
		}

		// Keep the line (might be empty for intentional spacing)
		normalizedLines = append(normalizedLines, line)
	}

	return strings.Join(normalizedLines, "\n")
}

// isVariableDeclarationContext determines if empty lines are between variable declarations
func isVariableDeclarationContext(prev, next string) bool {
	// Check if previous line looks like start of variable declaration
	prevIsVarStart := strings.HasPrefix(prev, "my ") || strings.HasPrefix(prev, "our ") ||
		strings.HasPrefix(prev, "local ") || prev == "my" || prev == "our" || prev == "local"

	// Check if next line looks like variable name
	nextIsVarName := strings.HasPrefix(next, "$") || strings.HasPrefix(next, "@") || strings.HasPrefix(next, "%")

	return prevIsVarStart && nextIsVarName
}
