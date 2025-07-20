// ABOUTME: Type injection transformer for adding type annotations to untyped Perl CST
// ABOUTME: Converts untyped variable declarations and method signatures to typed equivalents

package pipeline

import (
	"fmt"
	"regexp"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/parser/treesitter"
	"tamarou.com/pvm/internal/types"
)

// TypeInjectionTransformer adds type annotations to untyped Perl code based on inference results
type TypeInjectionTransformer struct {
	BaseTransformer
	typeInfo        map[string]*types.TypeInfo
	annotationStyle string
}

// TypeInjectionOptions controls type injection behavior
type TypeInjectionOptions struct {
	AnnotationStyle     string
	PreserveFormatting  bool
	InjectVariableTypes bool
	InjectMethodTypes   bool
	InjectReturnTypes   bool
}

// NewTypeInjectionTransformer creates a new type injection transformer
func NewTypeInjectionTransformer(typeInfo map[string]*types.TypeInfo, options TypeInjectionOptions) Transformer {
	return &TypeInjectionTransformer{
		BaseTransformer: NewBaseTransformer("type_injection", "Adds type annotations to untyped Perl code"),
		typeInfo:        typeInfo,
		annotationStyle: options.AnnotationStyle,
	}
}

// Transform adds type annotations to the CST based on inference results
func (ti *TypeInjectionTransformer) Transform(input *TransformationInput) (*TransformationOutput, error) {
	// Clone the CST and content since we'll be modifying them
	content := make([]byte, len(input.Content))
	copy(content, input.Content)

	// Transform the CST by adding type annotations
	transformed, err := ti.transformNode(input.CST, &content)
	if err != nil {
		return nil, fmt.Errorf("type injection failed: %w", err)
	}

	// Validate the generated code is syntactically correct
	if transformed > 0 {
		if err := ti.validateGeneratedCode(content); err != nil {
			return nil, fmt.Errorf("type injection produced invalid Perl syntax: %w", err)
		}
	}

	// Create metrics for the transformation
	metrics := TransformationMetrics{
		NodesProcessed:  ti.countNodes(input.CST),
		BytesProcessed:  len(input.Content),
		MemoryAllocated: int64(len(content) - len(input.Content)),
	}

	// Create new CST from transformed content if needed
	// For now, we'll work with content transformation
	return &TransformationOutput{
		CST:      input.CST, // We'd need to re-parse for true CST modification
		Content:  content,
		Modified: transformed > 0,
		Metrics:  metrics,
	}, nil
}

// transformNode recursively transforms nodes, adding type annotations where appropriate
func (ti *TypeInjectionTransformer) transformNode(node *sitter.Node, content *[]byte) (int, error) {
	if node == nil {
		return 0, nil
	}

	nodesChanged := 0

	switch node.Kind() {
	case "variable_declaration":
		changed, err := ti.transformVariableDeclaration(node, content)
		if err != nil {
			return 0, err
		}
		if changed {
			nodesChanged++
		}

	case "subroutine_declaration_statement":
		changed, err := ti.transformSubroutineDeclaration(node, content)
		if err != nil {
			return 0, err
		}
		if changed {
			nodesChanged++
		}

	case "method_declaration_statement":
		changed, err := ti.transformMethodDeclaration(node, content)
		if err != nil {
			return 0, err
		}
		if changed {
			nodesChanged++
		}
	}

	// Recursively process child nodes
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			childChanged, err := ti.transformNode(child, content)
			if err != nil {
				return nodesChanged, err
			}
			nodesChanged += childChanged
		}
	}

	return nodesChanged, nil
}

// transformVariableDeclaration adds type annotations to variable declarations
func (ti *TypeInjectionTransformer) transformVariableDeclaration(node *sitter.Node, content *[]byte) (bool, error) {
	nodeID := ti.generateNodeID(node)
	typeInfo, exists := ti.typeInfo[nodeID]

	if !exists {
		return false, nil
	}

	// Find the variable name and its position
	varName, _, _ := ti.findVariableName(node, *content)
	if varName == "" {
		return false, nil
	}

	// Generate type annotation with validation
	typeAnnotation, err := ti.formatAndValidateTypeAnnotation(typeInfo.Type)
	if err != nil {
		return false, fmt.Errorf("invalid type annotation for variable %s: %w", varName, err)
	}
	if typeAnnotation == "" {
		return false, nil
	}

	// Find the "my" keyword
	myKeyword := ti.findMyKeyword(node, *content)
	if myKeyword == nil {
		return false, nil
	}

	insertPos := myKeyword.EndByte()

	// Validate the insertion position
	if insertPos >= uint(len(*content)) {
		return false, fmt.Errorf("invalid insertion position %d for content length %d", insertPos, len(*content))
	}

	// Build the new content with type annotation
	newContent := make([]byte, 0, len(*content)+len(typeAnnotation)+1)
	newContent = append(newContent, (*content)[:insertPos]...)
	newContent = append(newContent, ' ')
	newContent = append(newContent, []byte(typeAnnotation)...)
	newContent = append(newContent, (*content)[insertPos:]...)

	// Validate the syntax of the modified declaration
	if err := ti.validateVariableDeclaration(newContent, insertPos, typeAnnotation); err != nil {
		return false, fmt.Errorf("type injection would create invalid variable declaration: %w", err)
	}

	*content = newContent
	return true, nil
}

// transformSubroutineDeclaration adds type annotations to subroutine signatures
func (ti *TypeInjectionTransformer) transformSubroutineDeclaration(node *sitter.Node, content *[]byte) (bool, error) {
	nodeID := ti.generateNodeID(node)
	typeInfo, exists := ti.typeInfo[nodeID]

	if !exists {
		return false, nil
	}

	// Find subroutine signature components
	subName := ti.findSubroutineName(node, *content)
	if subName == "" {
		return false, nil
	}

	// Generate validated type annotation
	typeAnnotation, err := ti.formatAndValidateTypeAnnotation(typeInfo.Type)
	if err != nil {
		return false, fmt.Errorf("invalid type annotation for subroutine %s: %w", subName, err)
	}
	if typeAnnotation == "" {
		return false, nil
	}

	// Find insertion point (after sub name)
	subKeywordNode := ti.findSubKeyword(node, *content)
	if subKeywordNode == nil {
		return false, nil
	}

	// Find the opening brace position
	bracePos := ti.findOpeningBrace(node, *content)
	if bracePos == 0 {
		return false, nil
	}

	// Validate the insertion position
	if bracePos >= uint(len(*content)) {
		return false, fmt.Errorf("invalid brace position %d for content length %d", bracePos, len(*content))
	}

	// Insert return type annotation
	// Pattern: "sub name {" -> "sub name -> ReturnType {"
	returnTypeStr := fmt.Sprintf(" -> %s ", typeAnnotation)

	newContent := make([]byte, 0, len(*content)+len(returnTypeStr))
	newContent = append(newContent, (*content)[:bracePos]...)
	newContent = append(newContent, []byte(returnTypeStr)...)
	newContent = append(newContent, (*content)[bracePos:]...)

	// Validate the syntax of the modified subroutine declaration
	if err := ti.validateSubroutineDeclaration(newContent, bracePos, returnTypeStr); err != nil {
		return false, fmt.Errorf("type injection would create invalid subroutine declaration: %w", err)
	}

	*content = newContent
	return true, nil
}

// transformMethodDeclaration adds type annotations to method signatures
func (ti *TypeInjectionTransformer) transformMethodDeclaration(node *sitter.Node, content *[]byte) (bool, error) {
	// Similar to subroutine declaration but for methods
	return ti.transformSubroutineDeclaration(node, content)
}

// Helper methods for finding CST components

func (ti *TypeInjectionTransformer) generateNodeID(node *sitter.Node) string {
	// Generate a stable node ID based on position and type
	return fmt.Sprintf("%s_%d_%d", node.Kind(), node.StartByte(), node.EndByte())
}

func (ti *TypeInjectionTransformer) findVariableName(node *sitter.Node, content []byte) (string, uint, uint) {
	// Find the variable identifier in a variable declaration
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Kind() == "identifier" {
			start := child.StartByte()
			end := child.EndByte()
			name := string(content[start:end])
			return name, start, end
		}
	}
	return "", 0, 0
}

func (ti *TypeInjectionTransformer) findMyKeyword(node *sitter.Node, content []byte) *sitter.Node {
	// Find the "my" keyword in a variable declaration
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Kind() == "my" {
			return child
		}
	}
	return nil
}

func (ti *TypeInjectionTransformer) findSubroutineName(node *sitter.Node, content []byte) string {
	// Find the subroutine name
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Kind() == "identifier" {
			start := child.StartByte()
			end := child.EndByte()
			return string(content[start:end])
		}
	}
	return ""
}

func (ti *TypeInjectionTransformer) findSubKeyword(node *sitter.Node, content []byte) *sitter.Node {
	// Find the "sub" keyword
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Kind() == "sub" {
			return child
		}
	}
	return nil
}

func (ti *TypeInjectionTransformer) findOpeningBrace(node *sitter.Node, content []byte) uint {
	// Find the opening brace position
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Kind() == "{" {
			return child.StartByte()
		}
	}
	return 0
}

// Deprecated: Use formatAndValidateTypeAnnotation instead
func (ti *TypeInjectionTransformer) formatTypeAnnotation(t types.Type) string {
	result, err := ti.formatAndValidateTypeAnnotation(t)
	if err != nil {
		return "" // Return empty string on validation errors for backward compatibility
	}
	return result
}

func (ti *TypeInjectionTransformer) countNodes(node *sitter.Node) int {
	if node == nil {
		return 0
	}

	count := 1
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			count += ti.countNodes(child)
		}
	}
	return count
}

// CanSkip returns true if no type information is available
func (ti *TypeInjectionTransformer) CanSkip(input *TransformationInput) bool {
	return len(ti.typeInfo) == 0
}

// validateGeneratedCode validates that the generated Perl code is syntactically correct
func (ti *TypeInjectionTransformer) validateGeneratedCode(content []byte) error {
	// Parse the generated code with tree-sitter to check for syntax errors
	parser := sitter.NewParser()
	parser.SetLanguage(treesitter.Language())

	tree := parser.Parse(content, nil)
	if tree == nil {
		return fmt.Errorf("failed to parse generated code")
	}

	root := tree.RootNode()
	if root == nil {
		return fmt.Errorf("generated code has no root node")
	}

	// Check for error nodes in the parse tree
	if err := ti.checkForErrorNodes(root, content); err != nil {
		return fmt.Errorf("syntax errors in generated code: %w", err)
	}

	return nil
}

// checkForErrorNodes recursively checks for ERROR nodes in the CST
func (ti *TypeInjectionTransformer) checkForErrorNodes(node *sitter.Node, content []byte) error {
	if node == nil {
		return nil
	}

	if node.Kind() == "ERROR" {
		start := node.StartByte()
		end := node.EndByte()
		if end > uint(len(content)) {
			end = uint(len(content))
		}
		errorText := string(content[start:end])
		return fmt.Errorf("parse error at position %d-%d: %q", start, end, errorText)
	}

	// Recursively check child nodes
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			if err := ti.checkForErrorNodes(child, content); err != nil {
				return err
			}
		}
	}

	return nil
}

// formatAndValidateTypeAnnotation formats and validates a type annotation
func (ti *TypeInjectionTransformer) formatAndValidateTypeAnnotation(t types.Type) (string, error) {
	if t == nil {
		return "", fmt.Errorf("type cannot be nil")
	}

	var typeStr string
	switch t := t.(type) {
	case interface{ String() string }:
		typeStr = t.String()
	default:
		typeStr = fmt.Sprintf("%T", t) // Fallback
	}

	// Validate the type annotation format
	if err := ti.validateTypeAnnotationFormat(typeStr); err != nil {
		return "", err
	}

	return typeStr, nil
}

// validateTypeAnnotationFormat checks if a type annotation is valid for Perl syntax
func (ti *TypeInjectionTransformer) validateTypeAnnotationFormat(typeAnnotation string) error {
	if typeAnnotation == "" {
		return fmt.Errorf("type annotation cannot be empty")
	}

	// Check for valid type identifier format (allows alphanumeric, underscore, and some special chars)
	validTypePattern := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\[[A-Za-z0-9_|&!]+\])?(\|[A-Za-z_][A-Za-z0-9_]*(\[[A-Za-z0-9_|&!]+\])?)*$`)
	if !validTypePattern.MatchString(typeAnnotation) {
		return fmt.Errorf("invalid type annotation format: %q", typeAnnotation)
	}

	// Check for potentially problematic characters
	if strings.ContainsAny(typeAnnotation, "\"'`$@%#;\\") {
		return fmt.Errorf("type annotation contains invalid characters: %q", typeAnnotation)
	}

	return nil
}

// validateVariableDeclaration validates a modified variable declaration
func (ti *TypeInjectionTransformer) validateVariableDeclaration(content []byte, insertPos uint, typeAnnotation string) error {
	// Extract a reasonable context around the insertion point for validation
	start := uint(0)
	if insertPos > 50 {
		start = insertPos - 50
	}

	end := insertPos + uint(len(typeAnnotation)) + 100
	if end > uint(len(content)) {
		end = uint(len(content))
	}

	context := content[start:end]

	// Basic pattern matching for "my Type $var" structure
	myTypeVarPattern := regexp.MustCompile(`my\s+[A-Za-z_][A-Za-z0-9_]*(\[[^\]]+\])?\s+\$[A-Za-z_][A-Za-z0-9_]*`)
	if !myTypeVarPattern.Match(context) {
		return fmt.Errorf("modified variable declaration does not match expected pattern")
	}

	return nil
}

// validateSubroutineDeclaration validates a modified subroutine declaration
func (ti *TypeInjectionTransformer) validateSubroutineDeclaration(content []byte, insertPos uint, returnTypeStr string) error {
	// Extract a reasonable context around the insertion point for validation
	start := uint(0)
	if insertPos > 100 {
		start = insertPos - 100
	}

	end := insertPos + uint(len(returnTypeStr)) + 100
	if end > uint(len(content)) {
		end = uint(len(content))
	}

	context := content[start:end]

	// Basic pattern matching for "sub name -> ReturnType {" structure
	subReturnTypePattern := regexp.MustCompile(`sub\s+[A-Za-z_][A-Za-z0-9_]*\s*->\s*[A-Za-z_][A-Za-z0-9_]*(\[[^\]]+\])?\s*\{`)
	if !subReturnTypePattern.Match(context) {
		return fmt.Errorf("modified subroutine declaration does not match expected pattern")
	}

	return nil
}
