// ABOUTME: Tree transformation framework for converting typed Perl CST to clean Perl CST
// ABOUTME: Removes type annotation nodes while preserving all other syntax, comments, and formatting

package compiler

import (
	"fmt"
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
		&TypeExpressionRemovalRule{},
		&VariableDeclarationCleanupRule{},
		&TypeAssertionCleanupRule{},
		&MethodParameterCleanupRule{},
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
	if transformer.options.RemoveTypeNodes {
		return "", nil
	}
	// Otherwise preserve as-is
	return transformer.getNodeText(node), nil
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

				// Add a single space instead of the entire parenthesized type
				result.WriteString(" ")

				// Skip all three tokens: (, type_expression, )
				skipUntil = afterChild.EndByte()
				lastEnd = skipUntil

				// Skip any trailing whitespace
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

	// Find the expression part (before "as")
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Kind() != NodeAs && child.Kind() != NodeTypeExpression {
			// This should be the expression we want to keep
			return transformer.transformNode(child)
		}
	}

	// Fallback to original text if we can't parse it
	return transformer.getNodeText(node), nil
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

	return &TransformationResult{
		TransformedCode: transformed,
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
	if !strings.Contains(nodeText, "->") {
		// Not a return type, let other rules handle it
		return nodeText, nil
	}

	// For subroutine return types (-> Type), remove entirely if removing types
	if transformer.options.RemoveTypeNodes {
		return "", nil // Remove the entire -> Type construct
	}
	return nodeText, nil
}

func (r *SubroutineReturnTypeRule) Description() string {
	return "Removes subroutine return type annotations (-> Type)"
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

		// Check for common type patterns
		if r.looksLikeTypeAnnotation(trimmed) {
			return "", nil // Remove type-like ERROR nodes
		}
	}

	// Otherwise preserve the ERROR node (might be legitimate syntax error)
	return nodeText, nil
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
