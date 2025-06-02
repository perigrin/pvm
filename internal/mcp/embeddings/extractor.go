// ABOUTME: Code block extraction for embedding generation
// ABOUTME: Extracts meaningful code blocks from Perl files for semantic search

package embeddings

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	chromem "github.com/philippgille/chromem-go"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
)

// Error codes for extraction operations
const (
	ErrExtractionFailed  = "MCP-8101" // Failed to extract code blocks
	ErrParsingFailed     = "MCP-8102" // Failed to parse file
	ErrInvalidCodeBlock  = "MCP-8103" // Invalid code block structure
	ErrBatchCreationFail = "MCP-8104" // Failed to create document batch
)

// CodeBlock represents an extracted code block with metadata
type CodeBlock struct {
	// ID is a unique identifier for the block (project/file/block)
	ID string

	// Content is the actual code text
	Content string

	// Type is the kind of code block (function, method, class, etc.)
	Type string

	// Name is the name of the code element (function name, class name, etc.)
	Name string

	// File is the source file path
	File string

	// StartLine is the starting line number
	StartLine int

	// EndLine is the ending line number
	EndLine int

	// TypeInfo contains type annotations if present
	TypeInfo map[string]string

	// Imports are the modules imported by this file
	Imports []string

	// Context provides surrounding context (e.g., parent class/package)
	Context string
}

// Extractor extracts code blocks from Perl files
type Extractor struct {
	// Note: Using parser pool instead of shared instance for thread safety
}

// NewExtractor creates a new code block extractor
func NewExtractor() (*Extractor, error) {
	return &Extractor{}, nil
}

// ExtractFromFile extracts code blocks from a single Perl file
func (e *Extractor) ExtractFromFile(projectID, filePath string) ([]*CodeBlock, error) {
	// Parse the file using parser pool for thread safety
	return parser.PooledParserFunc(func(p parser.Parser) ([]*CodeBlock, error) {
		ast, err := p.ParseFile(filePath)
		if err != nil {
			return nil, errors.NewSystemError(
				ErrParsingFailed,
				fmt.Sprintf("Failed to parse file: %s", filePath),
				err,
			).WithLocation(filePath)
		}

		// Check for parse errors
		if len(ast.Errors) > 0 {
			// We can still extract from files with errors, just log them
			// In a real system, you might want to handle this differently
		}

		// Extract code blocks from the AST
		blocks := e.extractBlocksFromAST(ast, projectID, filePath)

		return blocks, nil
	})
}

// extractBlocksFromAST extracts code blocks from a parsed AST
func (e *Extractor) extractBlocksFromAST(ast *parser.AST, projectID, filePath string) []*CodeBlock {
	var blocks []*CodeBlock

	// Get base filename without path for IDs
	baseFile := filepath.Base(filePath)

	// Extract package/module context
	packageName := e.extractPackageName(ast.Root)
	imports := e.extractImports(ast.Root)

	// Walk the AST to find code blocks
	e.walkNode(ast.Root, func(node parser.Node) bool {
		switch node.Type() {
		case "sub_decl": // Tree-sitter parser produces sub_decl nodes
			if block := e.extractSubroutineWithSource(node, projectID, filePath, baseFile, packageName, imports, ast.Source); block != nil {
				blocks = append(blocks, block)
			}
		case "subroutine_declaration_statement", "subroutine_declaration":
			if block := e.extractSubroutine(node, projectID, filePath, baseFile, packageName, imports); block != nil {
				blocks = append(blocks, block)
			}
		case "method_declaration_statement", "method_declaration":
			if block := e.extractMethod(node, projectID, filePath, baseFile, packageName, imports); block != nil {
				blocks = append(blocks, block)
			}
		case "class_declaration":
			if block := e.extractClass(node, projectID, filePath, baseFile, packageName, imports); block != nil {
				blocks = append(blocks, block)
			}
		case "package_declaration":
			// Package declarations are captured for context, not as separate blocks
		case "expression_stmt", "literal":
			// Workaround: Check if this expression statement contains a typed subroutine
			// that wasn't parsed correctly by tree-sitter
			if block := e.extractTypedSubroutineFromExpressionWorkaround(node, projectID, filePath, baseFile, packageName, imports, ast.Source); block != nil {
				blocks = append(blocks, block)
			}
		}
		return true // Continue walking
	})

	// If no specific blocks found, create a file-level block
	if len(blocks) == 0 && ast.Root != nil {
		// Read file content directly since ast.Root.Text() may be empty with pooled parsers
		content, err := os.ReadFile(filePath)
		contentStr := ""
		if err == nil {
			contentStr = string(content)
		}

		blocks = append(blocks, &CodeBlock{
			ID:        fmt.Sprintf("%s/%s/file", projectID, baseFile),
			Content:   contentStr,
			Type:      "file",
			Name:      baseFile,
			File:      filePath,
			StartLine: 1,
			EndLine:   len(strings.Split(contentStr, "\n")),
			TypeInfo:  e.extractTypeInfo(ast),
			Imports:   imports,
			Context:   packageName,
		})
	}

	return blocks
}

// walkNode recursively walks the AST nodes
func (e *Extractor) walkNode(node parser.Node, visitor func(parser.Node) bool) {
	if node == nil {
		return
	}

	// Visit this node
	if !visitor(node) {
		return // Stop if visitor returns false
	}

	// Visit children
	for _, child := range node.Children() {
		e.walkNode(child, visitor)
	}
}

// extractPackageName extracts the package name from the AST
func (e *Extractor) extractPackageName(root parser.Node) string {
	var packageName string

	e.walkNode(root, func(node parser.Node) bool {
		if node.Type() == "package_declaration" {
			// Extract package name from the declaration
			text := node.Text()
			if parts := strings.Fields(text); len(parts) >= 2 {
				packageName = strings.TrimSuffix(parts[1], ";")
			}
			return false // Stop after finding first package
		}
		return true
	})

	if packageName == "" {
		packageName = "main"
	}

	return packageName
}

// extractImports extracts import statements from the AST
func (e *Extractor) extractImports(root parser.Node) []string {
	var imports []string
	seen := make(map[string]bool)

	e.walkNode(root, func(node parser.Node) bool {
		if node.Type() == "use_statement" {
			// Extract module name from use statement
			text := node.Text()
			if parts := strings.Fields(text); len(parts) >= 2 {
				module := strings.TrimSuffix(parts[1], ";")
				// Remove version or import list
				if idx := strings.IndexAny(module, " ("); idx > 0 {
					module = module[:idx]
				}
				if !seen[module] {
					imports = append(imports, module)
					seen[module] = true
				}
			}
		}
		return true
	})

	return imports
}

// extractSubroutine extracts a subroutine block
func (e *Extractor) extractSubroutine(node parser.Node, projectID, filePath, baseFile, packageName string, imports []string) *CodeBlock {
	// Extract subroutine name
	name := e.extractSubroutineName(node)
	if name == "" {
		return nil
	}

	// Extract type information
	typeInfo := e.extractSubroutineTypes(node)

	return &CodeBlock{
		ID:        fmt.Sprintf("%s/%s/sub/%s", projectID, baseFile, name),
		Content:   node.Text(),
		Type:      "function",
		Name:      name,
		File:      filePath,
		StartLine: node.Start().Line,
		EndLine:   node.End().Line,
		TypeInfo:  typeInfo,
		Imports:   imports,
		Context:   packageName,
	}
}

// extractSubroutineWithSource extracts a subroutine block with source access for text extraction
func (e *Extractor) extractSubroutineWithSource(node parser.Node, projectID, filePath, baseFile, packageName string, imports []string, source string) *CodeBlock {
	// Extract subroutine name using source
	name := e.extractSubroutineNameWithSource(node, source)
	if name == "" {
		return nil
	}

	// Extract type information
	typeInfo := e.extractSubroutineTypes(node)

	// Extract content using source positions
	content := e.extractNodeTextFromSource(node, source)

	return &CodeBlock{
		ID:        fmt.Sprintf("%s/%s/sub/%s", projectID, baseFile, name),
		Content:   content,
		Type:      "function",
		Name:      name,
		File:      filePath,
		StartLine: node.Start().Line,
		EndLine:   node.End().Line,
		TypeInfo:  typeInfo,
		Imports:   imports,
		Context:   packageName,
	}
}

// extractMethod extracts a method block
func (e *Extractor) extractMethod(node parser.Node, projectID, filePath, baseFile, packageName string, imports []string) *CodeBlock {
	// Extract method name
	name := e.extractMethodName(node)
	if name == "" {
		return nil
	}

	// Extract type information
	typeInfo := e.extractMethodTypes(node)

	return &CodeBlock{
		ID:        fmt.Sprintf("%s/%s/method/%s", projectID, baseFile, name),
		Content:   node.Text(),
		Type:      "method",
		Name:      name,
		File:      filePath,
		StartLine: node.Start().Line,
		EndLine:   node.End().Line,
		TypeInfo:  typeInfo,
		Imports:   imports,
		Context:   packageName,
	}
}

// extractClass extracts a class block
func (e *Extractor) extractClass(node parser.Node, projectID, filePath, baseFile, packageName string, imports []string) *CodeBlock {
	// Extract class name
	name := e.extractClassName(node)
	if name == "" {
		return nil
	}

	return &CodeBlock{
		ID:        fmt.Sprintf("%s/%s/class/%s", projectID, baseFile, name),
		Content:   node.Text(),
		Type:      "class",
		Name:      name,
		File:      filePath,
		StartLine: node.Start().Line,
		EndLine:   node.End().Line,
		TypeInfo:  make(map[string]string), // TODO: Extract field types
		Imports:   imports,
		Context:   packageName,
	}
}

// Helper methods to extract names and types

func (e *Extractor) extractSubroutineName(node parser.Node) string {
	// Look for the subroutine name in children
	for _, child := range node.Children() {
		if child.Type() == "name" || child.Type() == "identifier" {
			return child.Text()
		}
		// Sometimes the name is nested deeper
		if child.Type() == "subroutine_signature" {
			for _, grandchild := range child.Children() {
				if grandchild.Type() == "name" || grandchild.Type() == "identifier" {
					return grandchild.Text()
				}
			}
		}
	}

	// For sub_decl nodes, extract name from source text using position
	text := node.Text()
	if text == "" {
		// If Text() is empty (common with pooled parsers), try to extract from position
		text = e.extractTextFromPosition(node)
	}

	// Extract name from subroutine declaration text
	if strings.Contains(text, "sub ") {
		// Find "sub " and extract the next word
		subIndex := strings.Index(text, "sub ")
		if subIndex >= 0 {
			remaining := text[subIndex+4:] // Skip "sub "
			parts := strings.Fields(remaining)
			if len(parts) >= 1 {
				name := parts[0]
				// Remove signature or body
				if idx := strings.IndexAny(name, "({"); idx > 0 {
					name = name[:idx]
				}
				return name
			}
		}
	}

	return ""
}

// extractTextFromPosition extracts text from a node using its source position
// This is needed when Text() returns empty (common with pooled parsers)
func (e *Extractor) extractTextFromPosition(node parser.Node) string {
	// This is a simplified implementation - in a real system you'd need access to the source
	// For now, we'll rely on the existing text extraction that should work
	return node.Text()
}

// extractSubroutineNameWithSource extracts subroutine name using source text
func (e *Extractor) extractSubroutineNameWithSource(node parser.Node, source string) string {
	// Try the regular method first
	if name := e.extractSubroutineName(node); name != "" {
		return name
	}

	// If that fails, extract from source using position
	text := e.extractNodeTextFromSource(node, source)

	// Handle the case where text might not start with "sub " due to position offset
	// Look for "ub " (missing 's') or "sub " anywhere in the text
	patterns := []string{"sub ", "ub "}
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			// Find pattern and extract the next word
			patternIndex := strings.Index(text, pattern)
			if patternIndex >= 0 {
				remaining := text[patternIndex+len(pattern):] // Skip pattern
				parts := strings.Fields(remaining)
				if len(parts) >= 1 {
					name := parts[0]
					// Remove signature or body
					if idx := strings.IndexAny(name, "({"); idx > 0 {
						name = name[:idx]
					}
					return name
				}
			}
		}
	}

	return ""
}

// extractTypedSubroutineFromExpressionWorkaround attempts to extract typed subroutines
// from expression statements when tree-sitter fails to parse them correctly
func (e *Extractor) extractTypedSubroutineFromExpressionWorkaround(node parser.Node, projectID, filePath, baseFile, packageName string, imports []string, source string) *CodeBlock {
	// Extract text content from the node
	text := e.extractNodeTextFromSource(node, source)
	if text == "" {
		text = node.Text()
	}

	// Check if this looks like a typed subroutine declaration
	// Pattern: sub name(Type $param, ...) -> ReturnType {
	if !strings.Contains(text, "sub ") {
		return nil
	}

	// Use regex to match typed subroutine patterns
	// This is a heuristic approach for the cases where tree-sitter fails
	typedSubPattern := `sub\s+(\w+)\s*(?:\([^)]*\))?\s*(?:->\s*\w+(?:\[.*?\])?(?:\|\w+(?:\[.*?\])?)*\s*)?\s*\{`
	re, err := regexp.Compile(typedSubPattern)
	if err != nil {
		return nil
	}

	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return nil
	}

	subName := matches[1]
	if subName == "" {
		return nil
	}

	// Extract the full subroutine content by finding the matching closing brace
	subIndex := strings.Index(text, "sub "+subName)
	if subIndex == -1 {
		return nil
	}

	// Find the opening brace
	openBrace := strings.Index(text[subIndex:], "{")
	if openBrace == -1 {
		return nil
	}
	openBrace += subIndex

	// Find the matching closing brace
	braceCount := 1
	pos := openBrace + 1
	for pos < len(text) && braceCount > 0 {
		switch text[pos] {
		case '{':
			braceCount++
		case '}':
			braceCount--
		}
		pos++
	}

	if braceCount != 0 {
		// Couldn't find matching brace, take the whole text
		pos = len(text)
	}

	content := text[subIndex:pos]

	// Extract type information from the signature
	typeInfo := make(map[string]string)
	e.extractTypedParametersFromText(content, typeInfo)

	return &CodeBlock{
		ID:        fmt.Sprintf("%s/%s/sub/%s", projectID, baseFile, subName),
		Content:   content,
		Type:      "function",
		Name:      subName,
		File:      filePath,
		StartLine: node.Start().Line,
		EndLine:   node.End().Line,
		TypeInfo:  typeInfo,
		Imports:   imports,
		Context:   packageName,
	}
}

// extractTypedParametersFromText extracts type information from typed subroutine text
func (e *Extractor) extractTypedParametersFromText(text string, typeInfo map[string]string) {
	// Extract parameters from the signature part only (before the opening brace)
	openBrace := strings.Index(text, "{")
	if openBrace == -1 {
		return // No function body found
	}

	signature := text[:openBrace]

	// Pattern to match typed parameters: Type $var (within parentheses)
	paramPattern := `(\w+(?:\[.*?\])?(?:\|\w+(?:\[.*?\])?)*)\s+\$(\w+)`
	re, err := regexp.Compile(paramPattern)
	if err != nil {
		return
	}

	matches := re.FindAllStringSubmatch(signature, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			paramType := match[1]
			paramName := "$" + match[2]
			typeInfo[paramName] = paramType
		}
	}

	// Extract return type: -> ReturnType
	returnPattern := `->\s*(\w+(?:\[.*?\])?(?:\|\w+(?:\[.*?\])?)*)`
	returnRe, err := regexp.Compile(returnPattern)
	if err != nil {
		return
	}

	returnMatch := returnRe.FindStringSubmatch(text)
	if len(returnMatch) >= 2 {
		typeInfo["return"] = returnMatch[1]
	}
}

// extractNodeTextFromSource extracts text for a node using source and positions
func (e *Extractor) extractNodeTextFromSource(node parser.Node, source string) string {
	if source == "" {
		return node.Text()
	}

	lines := strings.Split(source, "\n")
	start := node.Start()
	end := node.End()

	if start.Line <= 0 || start.Line > len(lines) {
		return node.Text()
	}

	if start.Line == end.Line {
		// Single line
		line := lines[start.Line-1] // Lines are 1-indexed
		// Columns appear to be 1-indexed, so adjust by subtracting 1
		startCol := start.Column - 1
		endCol := end.Column - 1
		if startCol >= 0 && startCol < len(line) && endCol >= 0 && endCol <= len(line) {
			return line[startCol:endCol]
		}
	} else {
		// Multi-line
		var parts []string
		for i := start.Line - 1; i < end.Line && i < len(lines); i++ {
			line := lines[i]
			switch {
			case i == start.Line-1:
				// First line - start from column (adjust for 1-indexed)
				startCol := start.Column - 1
				if startCol >= 0 && startCol < len(line) {
					parts = append(parts, line[startCol:])
				}
			case i == end.Line-1:
				// Last line - end at column (adjust for 1-indexed)
				endCol := end.Column - 1
				if endCol >= 0 && endCol <= len(line) {
					parts = append(parts, line[:endCol])
				}
			default:
				// Middle lines - full line
				parts = append(parts, line)
			}
		}
		return strings.Join(parts, "\n")
	}

	return node.Text()
}

func (e *Extractor) extractMethodName(node parser.Node) string {
	// Similar to subroutine but for methods
	for _, child := range node.Children() {
		if child.Type() == "name" || child.Type() == "identifier" {
			return child.Text()
		}
	}

	// Fallback: try to extract from text
	text := node.Text()
	if strings.HasPrefix(text, "method ") {
		parts := strings.Fields(text)
		if len(parts) >= 2 {
			name := parts[1]
			if idx := strings.IndexAny(name, "({"); idx > 0 {
				name = name[:idx]
			}
			return name
		}
	}

	return ""
}

func (e *Extractor) extractClassName(node parser.Node) string {
	// Look for class name in children
	for _, child := range node.Children() {
		if child.Type() == "name" || child.Type() == "identifier" || child.Type() == "class_name" {
			return child.Text()
		}
	}

	// Fallback: try to extract from text
	text := node.Text()
	if strings.HasPrefix(text, "class ") {
		parts := strings.Fields(text)
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	return ""
}

func (e *Extractor) extractSubroutineTypes(node parser.Node) map[string]string {
	typeInfo := make(map[string]string)

	// Look for typed parameters and return type
	for _, child := range node.Children() {
		if child.Type() == "subroutine_signature" || child.Type() == "signature" {
			// Extract parameter types
			e.extractParameterTypes(child, typeInfo)
		}
		if child.Type() == "return_type" {
			typeInfo["return"] = child.Text()
		}
	}

	return typeInfo
}

func (e *Extractor) extractMethodTypes(node parser.Node) map[string]string {
	// Methods have similar type extraction to subroutines
	return e.extractSubroutineTypes(node)
}

func (e *Extractor) extractParameterTypes(node parser.Node, typeInfo map[string]string) {
	// Walk the signature node to find typed parameters
	e.walkNode(node, func(child parser.Node) bool {
		if child.Type() == "typed_parameter" || child.Type() == "parameter" {
			// Extract parameter name and type
			paramName := ""
			paramType := ""

			for _, grandchild := range child.Children() {
				switch grandchild.Type() {
				case "variable", "scalar_variable":
					paramName = grandchild.Text()
				case "type", "type_expression":
					paramType = grandchild.Text()
				}
			}

			if paramName != "" && paramType != "" {
				typeInfo[paramName] = paramType
			}
		}
		return true
	})
}

func (e *Extractor) extractTypeInfo(ast *parser.AST) map[string]string {
	typeInfo := make(map[string]string)

	// Extract type information from type annotations
	for _, annotation := range ast.TypeAnnotations {
		if annotation.TypeExpression != nil {
			typeInfo[annotation.AnnotatedItem] = annotation.TypeExpression.String()
		}
	}

	return typeInfo
}

// ConvertToDocuments converts code blocks to chromem documents
func ConvertToDocuments(blocks []*CodeBlock) []chromem.Document {
	docs := make([]chromem.Document, len(blocks))

	for i, block := range blocks {
		// Create metadata map
		metadata := map[string]any{
			"type":       block.Type,
			"name":       block.Name,
			"file":       block.File,
			"start_line": block.StartLine,
			"end_line":   block.EndLine,
			"context":    block.Context,
		}

		// Add imports if present
		if len(block.Imports) > 0 {
			metadata["imports"] = strings.Join(block.Imports, ",")
		}

		// Add type information if present
		for k, v := range block.TypeInfo {
			metadata["type_"+k] = v
		}

		docs[i] = chromem.Document{
			ID:       block.ID,
			Content:  block.Content,
			Metadata: toString(metadata),
		}
	}

	return docs
}

// BatchExtractAndConvert extracts code blocks from multiple files and converts to documents
func BatchExtractAndConvert(extractor *Extractor, projectID string, filePaths []string) ([]chromem.Document, error) {
	var allBlocks []*CodeBlock

	for _, filePath := range filePaths {
		blocks, err := extractor.ExtractFromFile(projectID, filePath)
		if err != nil {
			// Log error but continue with other files
			// In production, you might want to collect these errors
			continue
		}
		allBlocks = append(allBlocks, blocks...)
	}

	if len(allBlocks) == 0 {
		return nil, errors.NewConfigError(
			ErrBatchCreationFail,
			"No code blocks extracted from any files",
			nil,
		)
	}

	return ConvertToDocuments(allBlocks), nil
}

// toString converts map[string]any to map[string]string for chromem metadata
func toString(m map[string]any) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		switch val := v.(type) {
		case string:
			result[k] = val
		case int:
			result[k] = fmt.Sprintf("%d", val)
		case []string:
			result[k] = strings.Join(val, ",")
		default:
			result[k] = fmt.Sprintf("%v", val)
		}
	}
	return result
}
