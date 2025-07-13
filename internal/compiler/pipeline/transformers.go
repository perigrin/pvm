// ABOUTME: Transformer implementations for the transformation pipeline
// ABOUTME: Contains type removal, whitespace normalization, and indentation transformers

package pipeline

import (
	"fmt"
	"regexp"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Type Removal Transformers

// TypeRemovalTransformer removes type annotations from typed Perl code
type TypeRemovalTransformer struct {
	BaseTransformer
	preserveComments bool
}

// NewTypeRemovalTransformer creates a new type removal transformer
func NewTypeRemovalTransformer() Transformer {
	return &TypeRemovalTransformer{
		BaseTransformer:  NewBaseTransformer("type_removal", "Removes type annotations from typed Perl code"),
		preserveComments: true,
	}
}

// NewTypeRemovalTransformerWithOptions creates a type removal transformer with custom options
func NewTypeRemovalTransformerWithOptions(preserveComments bool) Transformer {
	return &TypeRemovalTransformer{
		BaseTransformer:  NewBaseTransformer("type_removal", "Removes type annotations from typed Perl code"),
		preserveComments: preserveComments,
	}
}

// Transform removes type annotations using direct CST transformation
func (tr *TypeRemovalTransformer) Transform(input *TransformationInput) (*TransformationOutput, error) {
	// Transform the CST directly
	transformed, nodesProcessed, err := tr.transformNode(input.CST, input.Content)
	if err != nil {
		return nil, fmt.Errorf("type removal failed: %w", err)
	}

	// Post-process to clean up any remaining type annotations
	cleanedCode := tr.cleanRemainingTypeAnnotations(transformed)

	// Normalize whitespace after type removal
	normalizedCode := tr.normalizeWhitespaceAfterTypeRemoval(cleanedCode)

	// Check if content was modified
	modified := normalizedCode != string(input.Content)

	// Create metrics
	metrics := TransformationMetrics{
		NodesProcessed:  nodesProcessed,
		BytesProcessed:  len(input.Content),
		MemoryAllocated: 0, // We don't track memory allocation in this simple implementation
	}

	return &TransformationOutput{
		CST:      input.CST, // CST structure doesn't change, only content
		Content:  []byte(normalizedCode),
		Modified: modified,
		Metrics:  metrics,
	}, nil
}

// CanSkip returns true if the CST contains no type annotations
func (tr *TypeRemovalTransformer) CanSkip(input *TransformationInput) bool {
	if !input.Context.Options.EnableOptimizations {
		return false
	}

	// Quick check: if content doesn't contain common type annotation patterns, skip
	content := string(input.Content)

	// Look for common type annotation patterns
	typePatterns := []string{
		"Int ", "Str ", "Num ", "Bool ", // Basic types
		"ArrayRef[", "HashRef[", "CodeRef[", // Parameterized types
		"|", "&", "!", // Union, intersection, negation
		" as ", "->", // Type assertions and method syntax
		"method ", "sub ", // Method/sub declarations might have types
	}

	for _, pattern := range typePatterns {
		if strings.Contains(content, pattern) {
			return false // Found potential type annotation, don't skip
		}
	}

	return true // No type patterns found, can skip
}

// transformNode recursively transforms a CST node to remove type annotations
func (tr *TypeRemovalTransformer) transformNode(node *sitter.Node, content []byte) (string, int, error) {
	if node == nil {
		return "", 0, nil
	}

	nodesProcessed := 1

	// Handle specific node types that contain type annotations
	switch node.Kind() {
	case "type_expression":
		// Remove type expressions entirely
		return "", nodesProcessed, nil

	case "variable_declaration":
		return tr.transformVariableDeclaration(node, content)

	case "method_declaration_statement":
		return tr.transformMethodDeclaration(node, content)

	case "mandatory_parameter":
		return tr.transformMethodParameter(node, content)

	case "type_assertion":
		return tr.transformTypeAssertion(node, content)

	case "ERROR":
		// Handle ERROR nodes that might contain type syntax
		return tr.transformErrorNode(node, content)
	}

	// For other nodes, transform children and preserve structure
	if node.ChildCount() == 0 {
		// Leaf node - return original text
		return tr.getNodeText(node, content), nodesProcessed, nil
	}

	// Transform children
	var result strings.Builder
	lastEnd := node.StartByte()

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		// Add whitespace before child
		if child.StartByte() > lastEnd {
			whitespace := string(content[lastEnd:child.StartByte()])
			result.WriteString(whitespace)
		}

		// Transform child
		childResult, childNodes, err := tr.transformNode(child, content)
		if err != nil {
			return "", nodesProcessed, err
		}
		nodesProcessed += childNodes

		result.WriteString(childResult)
		lastEnd = child.EndByte()
	}

	// Add trailing content
	if node.EndByte() > lastEnd {
		trailing := string(content[lastEnd:node.EndByte()])
		result.WriteString(trailing)
	}

	return result.String(), nodesProcessed, nil
}

// getNodeText extracts text content from a tree-sitter node
func (tr *TypeRemovalTransformer) getNodeText(node *sitter.Node, content []byte) string {
	if node == nil || content == nil {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	if start >= uint(len(content)) || end > uint(len(content)) {
		return ""
	}
	return string(content[start:end])
}

// transformVariableDeclaration handles variable declarations by removing type annotations
func (tr *TypeRemovalTransformer) transformVariableDeclaration(node *sitter.Node, content []byte) (string, int, error) {
	var result strings.Builder
	lastEnd := node.StartByte()
	nodesProcessed := 1
	skipUntil := uint(0)

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		// Skip ranges when removing parenthesized types
		if skipUntil > 0 && child.StartByte() < skipUntil {
			continue
		}
		skipUntil = 0

		// Check for parenthesized type expression pattern: ( type_expression )
		if child.Kind() == "(" && i+2 < node.ChildCount() {
			nextChild := node.Child(i + 1)
			afterChild := node.Child(i + 2)
			if nextChild != nil && afterChild != nil &&
				nextChild.Kind() == "type_expression" && afterChild.Kind() == ")" {
				// Add whitespace before opening parenthesis if needed
				if child.StartByte() > lastEnd {
					whitespace := string(content[lastEnd:child.StartByte()])
					result.WriteString(whitespace)
				}
				// Skip all three tokens: (, type_expression, )
				skipUntil = afterChild.EndByte()
				lastEnd = skipUntil
				continue
			}
		}

		// Skip standalone type expressions
		if child.Kind() == "type_expression" {
			result.WriteString(" ") // Add single space for separation
			lastEnd = child.EndByte()
			// Skip any whitespace immediately after the type expression
			for lastEnd < uint(len(content)) && (content[lastEnd] == ' ' || content[lastEnd] == '\t') {
				lastEnd++
			}
			continue
		}

		// Add whitespace between nodes
		if child.StartByte() > lastEnd {
			whitespace := string(content[lastEnd:child.StartByte()])
			result.WriteString(whitespace)
		}

		// Transform the child
		transformed, childNodes, err := tr.transformNode(child, content)
		if err != nil {
			return "", nodesProcessed, err
		}
		nodesProcessed += childNodes
		result.WriteString(transformed)
		lastEnd = child.EndByte()
	}

	// Add any trailing whitespace
	if node.EndByte() > lastEnd {
		whitespace := string(content[lastEnd:node.EndByte()])
		result.WriteString(whitespace)
	}

	return result.String(), nodesProcessed, nil
}

// transformMethodDeclaration handles method declarations by removing type annotations
func (tr *TypeRemovalTransformer) transformMethodDeclaration(node *sitter.Node, content []byte) (string, int, error) {
	var result strings.Builder
	lastEnd := node.StartByte()
	nodesProcessed := 1

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		// Check for leading return type in method declarations
		if i > 0 {
			prevChild := node.Child(i - 1)
			if prevChild != nil && tr.getNodeText(prevChild, content) == "method" &&
				child.Kind() == "type_expression" {
				// Add a single space to separate method from the next element
				result.WriteString(" ")
				// Skip the leading return type after "method" keyword
				lastEnd = child.EndByte()
				// Skip any whitespace immediately after the type
				for lastEnd < uint(len(content)) && (content[lastEnd] == ' ' || content[lastEnd] == '\t') {
					lastEnd++
				}
				continue
			}
		}

		// Add whitespace before child
		if child.StartByte() > lastEnd {
			whitespace := string(content[lastEnd:child.StartByte()])
			result.WriteString(whitespace)
		}

		// Transform child
		transformed, childNodes, err := tr.transformNode(child, content)
		if err != nil {
			return "", nodesProcessed, err
		}
		nodesProcessed += childNodes
		result.WriteString(transformed)
		lastEnd = child.EndByte()
	}

	// Add trailing content
	if node.EndByte() > lastEnd {
		trailing := string(content[lastEnd:node.EndByte()])
		result.WriteString(trailing)
	}

	return result.String(), nodesProcessed, nil
}

// transformMethodParameter handles method parameters by removing type annotations
func (tr *TypeRemovalTransformer) transformMethodParameter(node *sitter.Node, content []byte) (string, int, error) {
	var result strings.Builder
	nodesProcessed := 1

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		// Skip type expression nodes
		if child.Kind() == "type_expression" {
			continue
		}

		// Transform the child
		transformed, childNodes, err := tr.transformNode(child, content)
		if err != nil {
			return "", nodesProcessed, err
		}
		nodesProcessed += childNodes
		result.WriteString(transformed)
	}

	return result.String(), nodesProcessed, nil
}

// transformTypeAssertion handles type assertion expressions
func (tr *TypeRemovalTransformer) transformTypeAssertion(node *sitter.Node, content []byte) (string, int, error) {
	// For type assertions, preserve the expression but remove the type part
	// Pattern: $value as Type -> $value
	nodesProcessed := 1
	var expressionParts []string
	foundAs := false

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		if child.Kind() == "as" || tr.getNodeText(child, content) == "as" {
			foundAs = true
			break
		}

		// Skip type expressions but include everything else
		if child.Kind() != "type_expression" {
			transformed, childNodes, err := tr.transformNode(child, content)
			if err != nil {
				return "", nodesProcessed, err
			}
			nodesProcessed += childNodes
			expressionParts = append(expressionParts, transformed)
		}
	}

	if foundAs && len(expressionParts) > 0 {
		return strings.Join(expressionParts, ""), nodesProcessed, nil
	}

	// Fallback: try to find the expression before "as" in the original text
	nodeText := tr.getNodeText(node, content)
	if asIndex := strings.Index(nodeText, " as "); asIndex > 0 {
		return strings.TrimSpace(nodeText[:asIndex]), nodesProcessed, nil
	}

	// Last resort fallback
	return nodeText, nodesProcessed, nil
}

// transformErrorNode handles ERROR nodes that might contain type annotations
func (tr *TypeRemovalTransformer) transformErrorNode(node *sitter.Node, content []byte) (string, int, error) {
	nodeText := tr.getNodeText(node, content)
	nodesProcessed := 1

	// Check if this ERROR node contains type-like content
	if tr.looksLikeTypeAnnotation(strings.TrimSpace(nodeText)) {
		return "", nodesProcessed, nil // Remove type-like ERROR nodes
	}

	// Otherwise preserve the ERROR node (might be legitimate syntax error)
	return nodeText, nodesProcessed, nil
}

// looksLikeTypeAnnotation checks if text looks like a type annotation
func (tr *TypeRemovalTransformer) looksLikeTypeAnnotation(text string) bool {
	if text == "" {
		return false
	}

	// Common type patterns
	typePatterns := []string{
		"Int", "Str", "Bool", "Num", "Any", "Undef",
		"ArrayRef", "HashRef", "CodeRef", "Maybe",
		"->", "returns",
	}

	for _, pattern := range typePatterns {
		if text == pattern || strings.HasPrefix(text, pattern+"[") || strings.Contains(text, "->") || strings.Contains(text, "returns ") {
			return true
		}
	}

	// Check if it starts with capital letter (likely a type)
	if len(text) > 0 && text[0] >= 'A' && text[0] <= 'Z' {
		return true
	}

	return false
}

// cleanRemainingTypeAnnotations post-processes code to remove any type annotations
// that the CST transformation missed due to grammar limitations
func (tr *TypeRemovalTransformer) cleanRemainingTypeAnnotations(code string) string {
	// Pattern to match type annotations in for loops: "for my Type $var"
	forLoopTypePattern := regexp.MustCompile(`\bfor\s+my\s+\w+(?:\[[^\]]*\])*\s+(\$\w+)`)
	code = forLoopTypePattern.ReplaceAllString(code, "for my $1")

	// Pattern to match any remaining standalone type expressions with brackets
	complexTypePattern := regexp.MustCompile(`\b(?:ArrayRef|HashRef|CodeRef|Maybe|Optional|Union|Intersection)\[[^\]]*\]`)
	code = complexTypePattern.ReplaceAllString(code, "")

	// Pattern to match parenthesized union/intersection types
	parenthesizedUnionPattern := regexp.MustCompile(`\((?:[A-Z]\w*(?:\[[^\]]*\])?)\|(?:[A-Z]\w*(?:\[[^\]]*\])?(?:\|[A-Z]\w*(?:\[[^\]]*\])?)*)\)`)
	code = parenthesizedUnionPattern.ReplaceAllString(code, "(|)")

	return code
}

// normalizeWhitespaceAfterTypeRemoval fixes malformed whitespace left behind after type removal
func (tr *TypeRemovalTransformer) normalizeWhitespaceAfterTypeRemoval(code string) string {
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

			// Skip empty lines between variable declarations
			if tr.isVariableDeclarationContext(prevNonEmpty, nextNonEmpty) {
				continue
			}
		}

		// Handle lines that are just whitespace with a variable (e.g., "  $var;")
		if strings.HasPrefix(trimmed, "$") && i > 0 {
			prevLine := strings.TrimSpace(lines[i-1])
			if prevLine == "my" || prevLine == "our" || prevLine == "local" {
				// Combine with previous line
				if len(normalizedLines) > 0 {
					normalizedLines[len(normalizedLines)-1] = prevLine + " " + trimmed
					continue
				}
			}
		}

		// Keep the line
		normalizedLines = append(normalizedLines, line)
	}

	return strings.Join(normalizedLines, "\n")
}

// isVariableDeclarationContext determines if empty lines are between variable declarations
func (tr *TypeRemovalTransformer) isVariableDeclarationContext(prev, next string) bool {
	// Check if previous line looks like start of variable declaration
	prevIsVarStart := strings.HasPrefix(prev, "my ") || strings.HasPrefix(prev, "our ") ||
		strings.HasPrefix(prev, "local ") || prev == "my" || prev == "our" || prev == "local"

	// Check if next line looks like variable name
	nextIsVarName := strings.HasPrefix(next, "$") || strings.HasPrefix(next, "@") || strings.HasPrefix(next, "%")

	return prevIsVarStart && nextIsVarName
}

// TypePreservationTransformer preserves type annotations (identity transformation for typed targets)
type TypePreservationTransformer struct {
	BaseTransformer
}

// NewTypePreservationTransformer creates a transformer that preserves type annotations
func NewTypePreservationTransformer() Transformer {
	return &TypePreservationTransformer{
		BaseTransformer: NewBaseTransformer("type_preservation", "Preserves type annotations in typed Perl code"),
	}
}

// Transform performs identity transformation (no changes)
func (tp *TypePreservationTransformer) Transform(input *TransformationInput) (*TransformationOutput, error) {
	metrics := TransformationMetrics{
		NodesProcessed:  1, // Just count the root
		BytesProcessed:  len(input.Content),
		MemoryAllocated: 0,
	}

	return &TransformationOutput{
		CST:      input.CST,
		Content:  input.Content,
		Modified: false,
		Metrics:  metrics,
	}, nil
}

// CanSkip always returns true when optimizations are enabled since this is identity transformation
func (tp *TypePreservationTransformer) CanSkip(input *TransformationInput) bool {
	return input.Context.Options.EnableOptimizations
}

// Whitespace Transformers

// WhitespaceNormalizerTransformer normalizes whitespace, especially after type removal
type WhitespaceNormalizerTransformer struct {
	BaseTransformer
	preserveTypes bool
}

// NewWhitespaceNormalizerTransformer creates a new whitespace normalizer
func NewWhitespaceNormalizerTransformer() Transformer {
	return &WhitespaceNormalizerTransformer{
		BaseTransformer: NewBaseTransformer("whitespace_normalizer", "Normalizes whitespace and fixes formatting issues"),
		preserveTypes:   false,
	}
}

// NewWhitespaceNormalizerWithOptions creates a whitespace normalizer with custom options
func NewWhitespaceNormalizerWithOptions(preserveTypes bool) Transformer {
	return &WhitespaceNormalizerTransformer{
		BaseTransformer: NewBaseTransformer("whitespace_normalizer", "Normalizes whitespace and fixes formatting issues"),
		preserveTypes:   preserveTypes,
	}
}

// Transform normalizes whitespace in the content
func (wn *WhitespaceNormalizerTransformer) Transform(input *TransformationInput) (*TransformationOutput, error) {
	originalContent := string(input.Content)
	normalizedContent := wn.normalizeWhitespace(originalContent)

	// Check if content was modified
	modified := normalizedContent != originalContent

	metrics := TransformationMetrics{
		NodesProcessed:  1, // Process the entire content as one unit
		BytesProcessed:  len(input.Content),
		MemoryAllocated: int64(len(normalizedContent) - len(originalContent)),
	}

	return &TransformationOutput{
		CST:      input.CST,
		Content:  []byte(normalizedContent),
		Modified: modified,
		Metrics:  metrics,
	}, nil
}

// CanSkip returns true if the content appears to already be well-formatted
func (wn *WhitespaceNormalizerTransformer) CanSkip(input *TransformationInput) bool {
	if !input.Context.Options.EnableOptimizations {
		return false
	}

	content := string(input.Content)

	// Quick heuristics to detect if normalization is needed
	problemPatterns := []string{
		"\nmy\n  $",    // Split variable declarations
		"\nour\n  $",   // Split our declarations
		"\nlocal\n  $", // Split local declarations
		"\n\n\nmy ",    // Excessive empty lines between declarations
		"\n\n\n$",      // Malformed spacing around variables
	}

	for _, pattern := range problemPatterns {
		if strings.Contains(content, pattern) {
			return false // Found formatting issues, don't skip
		}
	}

	return true // No obvious formatting issues found
}

// normalizeWhitespace fixes malformed whitespace left behind after type removal
// This is extracted from the existing normalizeWhitespaceAfterTypeRemoval function
func (wn *WhitespaceNormalizerTransformer) normalizeWhitespace(code string) string {
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
			if wn.isVariableDeclarationContext(prevNonEmpty, nextNonEmpty) {
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

		// Additional normalization: handle excessive blank lines
		if trimmed == "" && len(normalizedLines) > 0 {
			// Check if we already have an empty line
			lastLine := ""
			if len(normalizedLines) > 0 {
				lastLine = strings.TrimSpace(normalizedLines[len(normalizedLines)-1])
			}

			// Don't add multiple consecutive empty lines
			if lastLine == "" {
				continue
			}
		}

		// Keep the line (might be empty for intentional spacing)
		normalizedLines = append(normalizedLines, line)
	}

	result := strings.Join(normalizedLines, "\n")

	// Additional cleanup: normalize multiple consecutive newlines
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}

	return result
}

// isVariableDeclarationContext determines if empty lines are between variable declarations
func (wn *WhitespaceNormalizerTransformer) isVariableDeclarationContext(prev, next string) bool {
	// Check if previous line looks like start of variable declaration
	prevIsVarStart := strings.HasPrefix(prev, "my ") || strings.HasPrefix(prev, "our ") ||
		strings.HasPrefix(prev, "local ") || prev == "my" || prev == "our" || prev == "local"

	// Check if next line looks like variable name
	nextIsVarName := strings.HasPrefix(next, "$") || strings.HasPrefix(next, "@") || strings.HasPrefix(next, "%")

	return prevIsVarStart && nextIsVarName
}

// Indentation Transformers

// IndentationNormalizerTransformer normalizes indentation to a consistent style
type IndentationNormalizerTransformer struct {
	BaseTransformer
	indentSize int
	useTabs    bool
}

// NewIndentationNormalizerTransformer creates a new indentation normalizer
func NewIndentationNormalizerTransformer(indentSize int, useTabs bool) Transformer {
	description := "Normalizes indentation"
	if useTabs {
		description += " using tabs"
	} else {
		description += " using spaces"
	}

	return &IndentationNormalizerTransformer{
		BaseTransformer: NewBaseTransformer("indentation_normalizer", description),
		indentSize:      indentSize,
		useTabs:         useTabs,
	}
}

// Transform normalizes indentation in the content
func (in *IndentationNormalizerTransformer) Transform(input *TransformationInput) (*TransformationOutput, error) {
	originalContent := string(input.Content)
	normalizedContent := in.normalizeIndentation(originalContent)

	// Check if content was modified
	modified := normalizedContent != originalContent

	metrics := TransformationMetrics{
		NodesProcessed:  1,
		BytesProcessed:  len(input.Content),
		MemoryAllocated: int64(len(normalizedContent) - len(originalContent)),
	}

	return &TransformationOutput{
		CST:      input.CST,
		Content:  []byte(normalizedContent),
		Modified: modified,
		Metrics:  metrics,
	}, nil
}

// CanSkip returns true if indentation appears consistent
func (in *IndentationNormalizerTransformer) CanSkip(input *TransformationInput) bool {
	if !input.Context.Options.EnableOptimizations {
		return false
	}

	// Simple heuristic: if content doesn't have mixed tabs/spaces, might be OK
	content := string(input.Content)
	hasTabs := strings.Contains(content, "\t")
	hasSpaceIndent := strings.Contains(content, "\n    ") || strings.Contains(content, "\n  ")

	// If it has both tabs and space indentation, need normalization
	return !(hasTabs && hasSpaceIndent)
}

// normalizeIndentation converts indentation to consistent style
func (in *IndentationNormalizerTransformer) normalizeIndentation(code string) string {
	lines := strings.Split(code, "\n")
	var normalizedLines []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			// Keep empty lines as-is
			normalizedLines = append(normalizedLines, line)
			continue
		}

		// Count leading whitespace
		leadingSpaces := 0
		leadingTabs := 0

		for i, char := range line {
			switch char {
			case ' ':
				leadingSpaces++
			case '\t':
				leadingTabs++
			default:
				// Found non-whitespace character
				content := line[i:]

				// Calculate normalized indentation level
				var indentLevel int
				if in.useTabs {
					indentLevel = leadingTabs + (leadingSpaces / in.indentSize)
				} else {
					indentLevel = (leadingSpaces / in.indentSize) + leadingTabs
				}

				// Build normalized line
				var normalizedIndent string
				if in.useTabs {
					normalizedIndent = strings.Repeat("\t", indentLevel)
				} else {
					normalizedIndent = strings.Repeat(strings.Repeat(" ", in.indentSize), indentLevel)
				}

				normalizedLines = append(normalizedLines, normalizedIndent+content)
				break
			}
		}
	}

	return strings.Join(normalizedLines, "\n")
}
