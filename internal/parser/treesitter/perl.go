// ABOUTME: Go bindings for tree-sitter-perl
// ABOUTME: Provides Go interface to tree-sitter-perl parser

package treesitter

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	tree_sitter_typed_perl "github.com/perigrin/pvm/tree-sitter-typed-perl/bindings/go"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Language returns the tree-sitter language for Typed Perl
func Language() *sitter.Language {
	return sitter.NewLanguage(tree_sitter_typed_perl.Language())
}

// PerlParser represents a parser instance for Perl code
type PerlParser struct {
	parser       *sitter.Parser
	typeQueries  *sitter.Query
	debug        bool
	typePatterns []string
}

// NewPerlParser creates a new PerlParser instance
func NewPerlParser(debug bool) (*PerlParser, error) {
	// Create a new tree-sitter parser
	parser := sitter.NewParser()
	language := Language()

	if err := parser.SetLanguage(language); err != nil {
		return nil, fmt.Errorf("failed to set language: %v", err)
	}

	p := &PerlParser{
		parser:       parser,
		typeQueries:  nil, // Will be set up later for type annotation queries
		debug:        debug,
		typePatterns: []string{},
	}

	return p, nil
}

// ParseFile parses a Perl file using tree-sitter
func (p *PerlParser) ParseFile(path string) (*PerlTree, error) {
	// Read the file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return p.ParseBytes(content)
}

// ParseString parses a string of Perl code using tree-sitter
func (p *PerlParser) ParseString(content string) (*PerlTree, error) {
	return p.ParseBytes([]byte(content))
}

// ParseBytes parses byte content as Perl code using tree-sitter
func (p *PerlParser) ParseBytes(content []byte) (*PerlTree, error) {
	// Parse the content using tree-sitter
	tree := p.parser.Parse(content, nil)
	if tree == nil {
		return nil, fmt.Errorf("failed to parse content")
	}

	perlTree := &PerlTree{
		Tree:    tree,
		Content: content,
		parser:  p,
	}

	return perlTree, nil
}

// Close frees resources used by the parser
func (p *PerlParser) Close() {
	if p.parser != nil {
		p.parser.Close()
	}
}

// PerlError represents a parsing error
type PerlError struct {
	Message string
	Row     uint32
	Column  uint32
}

// PerlTree represents a parsed Perl syntax tree
type PerlTree struct {
	Tree    *sitter.Tree
	Content []byte
	parser  *PerlParser
	errors  []PerlError
}

// Root returns the root node of the tree
func (t *PerlTree) Root() *sitter.Node {
	if t.Tree == nil {
		return nil
	}
	return t.Tree.RootNode()
}

// Errors returns any parsing errors
func (t *PerlTree) Errors() []PerlError {
	return t.errors
}

// Close frees resources used by the tree
func (t *PerlTree) Close() {
	if t.Tree != nil {
		t.Tree.Close()
	}
}

// FindTypeAnnotations finds all type annotations in the tree
func (t *PerlTree) FindTypeAnnotations() ([]*PerlTypeAnnotation, error) {
	var annotations []*PerlTypeAnnotation

	if t.Tree == nil || t.Tree.RootNode() == nil {
		return annotations, nil
	}

	// Traverse the tree looking for type annotations
	root := t.Tree.RootNode()
	t.traverseForTypeAnnotations(root, &annotations)

	return annotations, nil
}

// traverseForTypeAnnotations recursively traverses the tree looking for type annotations
func (t *PerlTree) traverseForTypeAnnotations(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	if node == nil {
		return
	}

	// Check if this node represents a type annotation pattern
	switch node.Kind() {
	case "variable_declaration":
		t.processVariableDeclaration(node, annotations)
	case "subroutine_declaration_statement":
		t.processSubroutineDeclaration(node, annotations)
	case "method_declaration":
		t.processMethodDeclaration(node, annotations)
	}

	// Recursively process all child nodes
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child != nil {
			t.traverseForTypeAnnotations(child, annotations)
		}
	}
}

// processVariableDeclaration looks for type annotations in variable declarations
func (t *PerlTree) processVariableDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	// Look for both patterns:
	// 1. Original: my + scalar + : + attrlist  (my $var: Type)
	// 2. Typed Perl: my + ERROR + scalar       (my Type $var)

	var varName, typeName string
	var typeNode *sitter.Node
	hasColon := false
	hasTypeError := false

	// Walk through child nodes to find the components
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		switch child.Kind() {
		case "scalar":
			varName = t.getNodeText(child)
		case ":":
			hasColon = true
		case "attrlist":
			// Look for attribute inside attrlist
			for j := 0; j < int(child.ChildCount()); j++ {
				attr := child.Child(uint(j))
				if attr != nil && attr.Kind() == "attribute" {
					typeName = t.getNodeText(attr)
					typeNode = attr // Store the attribute node for position
					break
				}
			}
		case "ERROR":
			// This might be a type name in "my Type $var" syntax
			errorText := t.getNodeText(child)
			// Check if it looks like a type name (starts with uppercase)
			if len(errorText) > 0 && errorText[0] >= 'A' && errorText[0] <= 'Z' {
				typeName = errorText
				typeNode = child // Store the ERROR node for position
				hasTypeError = true
			}
		}
	}

	// Create annotation if we found the components and have a type node for position
	// Either colon-based (my $var: Type) OR error-based (my Type $var)
	if (hasColon || hasTypeError) && varName != "" && typeName != "" && typeNode != nil {
		annotation := &PerlTypeAnnotation{
			ItemName: varName,
			TypeName: typeName,
			Kind:     "variable",
			StartPos: int(typeNode.StartByte()), // Use the specific type token position
			EndPos:   int(typeNode.EndByte()),   // Use the specific type token position
			Content:  t.getNodeText(typeNode),   // Content of just the type token
		}
		*annotations = append(*annotations, annotation)
	}
}

// processSubroutineDeclaration looks for type annotations in subroutine declarations
func (t *PerlTree) processSubroutineDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	// Extract subroutine name
	var subName string

	// Walk through child nodes to find signature and return type
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		switch child.Kind() {
		case "bareword":
			// This should be the subroutine name
			subName = t.getNodeText(child)
		}
	}

	// Create individual annotations for parameter types and return type
	// This allows the stripper to remove them independently
	if subName != "" {
		// Create annotations for parameter types and return type
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(uint(i))
			if child == nil {
				continue
			}

			if child.Kind() == "signature" {
				// Process each ERROR node in the signature
				for j := 0; j < int(child.ChildCount()); j++ {
					sigChild := child.Child(uint(j))
					if sigChild != nil && sigChild.Kind() == "ERROR" {
						errorText := t.getNodeText(sigChild)
						if len(errorText) > 0 && errorText[0] >= 'A' && errorText[0] <= 'Z' {
							annotation := &PerlTypeAnnotation{
								ItemName: fmt.Sprintf("%s_param", subName),
								TypeName: errorText,
								Kind:     "subroutine_param",
								StartPos: int(sigChild.StartByte()),
								EndPos:   int(sigChild.EndByte()),
								Content:  errorText,
							}
							*annotations = append(*annotations, annotation)
						}
					}
				}
			} else if child.Kind() == "ERROR" && strings.HasPrefix(t.getNodeText(child), "->") {
				// Return type annotation
				errorText := t.getNodeText(child)
				annotation := &PerlTypeAnnotation{
					ItemName: fmt.Sprintf("%s_return", subName),
					TypeName: strings.TrimSpace(strings.TrimPrefix(errorText, "->")),
					Kind:     "subroutine_return",
					StartPos: int(child.StartByte()),
					EndPos:   int(child.EndByte()),
					Content:  errorText,
				}
				*annotations = append(*annotations, annotation)
			}
		}
	}
}

// extractParameterTypes extracts type information from a signature node
func (t *PerlTree) extractParameterTypes(signatureNode *sitter.Node) []string {
	var paramTypes []string

	for i := 0; i < int(signatureNode.ChildCount()); i++ {
		child := signatureNode.Child(uint(i))
		if child == nil {
			continue
		}

		if child.Kind() == "ERROR" {
			// This might be a parameter type like "Str" before the parameter name
			errorText := t.getNodeText(child)
			if len(errorText) > 0 && errorText[0] >= 'A' && errorText[0] <= 'Z' {
				paramTypes = append(paramTypes, errorText)
			}
		}
	}

	return paramTypes
}

// processMethodDeclaration looks for type annotations in method declarations
func (t *PerlTree) processMethodDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
	content := t.getNodeText(node)

	if len(content) > 0 && (containsMethodTypePattern(content)) {
		annotation := &PerlTypeAnnotation{
			ItemName: extractMethodName(content),
			TypeName: extractMethodTypes(content),
			Kind:     "method",
			StartPos: int(node.StartByte()),
			EndPos:   int(node.EndByte()),
			Content:  content,
		}
		*annotations = append(*annotations, annotation)
	}
}

// PerlTypeAnnotation represents a type annotation found in Perl code
type PerlTypeAnnotation struct {
	ItemName string
	TypeName string
	Kind     string // variable, subroutine, method, etc.
	StartPos int    // byte position
	EndPos   int    // byte position
	Content  string // original source text
	Children []*PerlTypeAnnotation
}

// String returns a string representation of the annotation
func (a *PerlTypeAnnotation) String() string {
	return fmt.Sprintf("%s: %s", a.ItemName, a.TypeName)
}

// These extraction functions are placeholders

// GetNodeText extracts the text content of a tree-sitter node
func (t *PerlTree) GetNodeText(node *sitter.Node) string {
	if node == nil {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	if start >= uint(len(t.Content)) || end > uint(len(t.Content)) {
		return ""
	}
	return string(t.Content[start:end])
}

// PerlPosition represents a position in the source code
type PerlPosition struct {
	Row    uint32
	Column uint32
}

// GetPosition returns the position of a node
func (t *PerlTree) GetPosition(node interface{}) PerlPosition {
	return PerlPosition{}
}

// Helper functions for extracting type information from text

// getNodeText is an alias for GetNodeText to match the usage in the methods above
func (t *PerlTree) getNodeText(node *sitter.Node) string {
	return t.GetNodeText(node)
}

// containsTypePattern checks if content contains variable type annotations
func containsTypePattern(content string) bool {
	// Look for patterns like: my $var: Type  or  my Type $var
	typePattern := `:\s*[A-Z][a-zA-Z0-9_]*|my\s+[A-Z][a-zA-Z0-9_]*\s+\$`
	matched, _ := regexp.MatchString(typePattern, content)
	return matched
}

// containsSubTypePattern checks if content contains subroutine type annotations
func containsSubTypePattern(content string) bool {
	// Look for patterns like: sub name(Type $param): ReturnType
	typePattern := `sub\s+\w+\s*\([^)]*:[^)]*\)|sub\s+\w+.*->`
	matched, _ := regexp.MatchString(typePattern, content)
	return matched
}

// containsMethodTypePattern checks if content contains method type annotations
func containsMethodTypePattern(content string) bool {
	// Look for patterns like: method name(Type $param): ReturnType
	typePattern := `method\s+\w+\s*\([^)]*:[^)]*\)|method\s+\w+.*->`
	matched, _ := regexp.MatchString(typePattern, content)
	return matched
}

// extractVariableName extracts variable name from declaration
func extractVariableName(content string) string {
	// Look for $variableName pattern
	varPattern := `\$([a-zA-Z_][a-zA-Z0-9_]*)`
	re := regexp.MustCompile(varPattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return "$" + matches[1]
	}
	return ""
}

// extractTypeName extracts type name from variable declaration
func extractTypeName(content string) string {
	// Look for : Type pattern or my Type pattern
	colonTypePattern := `:\s*([A-Z][a-zA-Z0-9_]*)`
	myTypePattern := `my\s+([A-Z][a-zA-Z0-9_]*)\s+\$`

	re := regexp.MustCompile(colonTypePattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}

	re = regexp.MustCompile(myTypePattern)
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// extractSubroutineName extracts subroutine name from declaration
func extractSubroutineName(content string) string {
	subPattern := `sub\s+([a-zA-Z_][a-zA-Z0-9_]*)`
	re := regexp.MustCompile(subPattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractSubroutineTypes extracts type information from subroutine
func extractSubroutineTypes(content string) string {
	// This is a simplified extraction - full implementation would parse parameters and return types
	return "subroutine_types"
}

// extractMethodName extracts method name from declaration
func extractMethodName(content string) string {
	methodPattern := `method\s+([a-zA-Z_][a-zA-Z0-9_]*)`
	re := regexp.MustCompile(methodPattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractMethodTypes extracts type information from method
func extractMethodTypes(content string) string {
	// This is a simplified extraction - full implementation would parse parameters and return types
	return "method_types"
}
