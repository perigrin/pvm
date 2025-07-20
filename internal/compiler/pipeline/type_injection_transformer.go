// ABOUTME: Type injection transformer for adding type annotations to untyped Perl CST
// ABOUTME: Converts untyped variable declarations and method signatures to typed equivalents

package pipeline

import (
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"
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

	// Generate type annotation
	typeAnnotation := ti.formatTypeAnnotation(typeInfo.Type)
	if typeAnnotation == "" {
		return false, nil
	}

	// Insert type annotation between "my" and variable name
	// Pattern: "my $var" -> "my Int $var"
	myKeyword := ti.findMyKeyword(node, *content)
	if myKeyword == nil {
		return false, nil
	}

	insertPos := myKeyword.EndByte()

	// Build the new content with type annotation
	newContent := make([]byte, 0, len(*content)+len(typeAnnotation)+1)
	newContent = append(newContent, (*content)[:insertPos]...)
	newContent = append(newContent, ' ')
	newContent = append(newContent, []byte(typeAnnotation)...)
	newContent = append(newContent, (*content)[insertPos:]...)

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

	// For now, we'll add return type annotation as comments
	// A full implementation would modify the CST structure
	typeAnnotation := ti.formatTypeAnnotation(typeInfo.Type)
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

	// Insert return type annotation
	// Pattern: "sub name {" -> "sub name -> ReturnType {"
	returnTypeStr := fmt.Sprintf(" -> %s ", typeAnnotation)

	newContent := make([]byte, 0, len(*content)+len(returnTypeStr))
	newContent = append(newContent, (*content)[:bracePos]...)
	newContent = append(newContent, []byte(returnTypeStr)...)
	newContent = append(newContent, (*content)[bracePos:]...)

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

func (ti *TypeInjectionTransformer) formatTypeAnnotation(t types.Type) string {
	// Format type for Perl syntax
	if t == nil {
		return ""
	}

	switch t := t.(type) {
	case interface{ String() string }:
		return t.String()
	default:
		return fmt.Sprintf("%T", t) // Fallback
	}
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
