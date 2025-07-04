// ABOUTME: CST analysis utilities for understanding tree-sitter typed Perl structure
// ABOUTME: Documents node types, provides navigation helpers, and maps type annotation patterns

package compiler

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// CST Node Types for Typed Perl Constructs
const (
	// Core Perl nodes
	NodeSourceFile = "source_file"
	NodeExpression = "expression_statement"
	NodeAssignment = "assignment_expression"
	NodeBlock      = "block"
	NodeStatement  = "statement"

	// Variable declaration nodes
	NodeVariableDecl = "variable_declaration"
	NodeScalar       = "scalar"
	NodeArray        = "array"
	NodeVarname      = "varname"
	NodeMy           = "my"
	NodeField        = "field"

	// Type annotation nodes
	NodeTypeExpression = "type_expression"
	NodeSimpleType     = "simple_type"
	NodeTypeAssertion  = "type_assertion_expression"
	NodeAs             = "as"

	// Method declaration nodes
	NodeMethodDecl     = "method_declaration_statement"
	NodeMethod         = "method"
	NodeSignature      = "signature"
	NodeMandatoryParam = "mandatory_parameter"

	// Error and ambiguous nodes
	NodeError         = "ERROR"
	NodeAmbiguousCall = "ambiguous_function_call_expression"

	// Common syntax nodes
	NodeBareword   = "bareword"
	NodeNumber     = "number"
	NodeString     = "string"
	NodeBinaryExpr = "binary_expression"
)

// ConstructPattern represents a pattern for identifying Perl constructs in CST
type ConstructPattern struct {
	Name        string // Human readable name
	NodeType    string // Primary CST node type
	Description string // Description of what this construct represents
}

// GetConstructPatterns returns all known patterns for Perl constructs
func GetConstructPatterns() []ConstructPattern {
	return []ConstructPattern{
		{
			Name:        "VariableDeclaration",
			NodeType:    NodeVariableDecl,
			Description: "Variable declaration: my [Type] $var = value; or field [Type] $name;",
		},
		{
			Name:        "TypeAssertion",
			NodeType:    NodeTypeAssertion,
			Description: "Type assertion: $value as Type",
		},
		{
			Name:        "MethodParameter",
			NodeType:    NodeMandatoryParam,
			Description: "Method parameter: [Type] $param",
		},
	}
}

// CSTAnalyzer provides utilities for analyzing tree-sitter CST structure
type CSTAnalyzer struct {
	root    *sitter.Node
	content []byte // Source content for extracting node text
}

// NewCSTAnalyzer creates a new CST analyzer for the given root node
func NewCSTAnalyzer(root *sitter.Node) *CSTAnalyzer {
	return &CSTAnalyzer{root: root}
}

// NewCSTAnalyzerWithContent creates a new CST analyzer with source content for text extraction
func NewCSTAnalyzerWithContent(root *sitter.Node, content []byte) *CSTAnalyzer {
	return &CSTAnalyzer{root: root, content: content}
}

// AnalyzeNode provides detailed information about a CST node
func (ca *CSTAnalyzer) AnalyzeNode(node *sitter.Node) NodeAnalysis {
	if node == nil {
		return NodeAnalysis{Valid: false, Error: "nil node"}
	}

	analysis := NodeAnalysis{
		Valid:      true,
		Type:       node.Kind(),
		StartByte:  node.StartByte(),
		EndByte:    node.EndByte(),
		ChildCount: int(node.ChildCount()),
		Children:   make([]ChildInfo, 0, node.ChildCount()),
	}

	// Analyze children
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			analysis.Children = append(analysis.Children, ChildInfo{
				Index: int(i),
				Type:  child.Kind(),
				Text:  ca.getNodeText(child),
			})
		}
	}

	// Check if this matches any construct patterns
	analysis.Pattern = ca.IdentifyPattern(node)

	return analysis
}

// NodeAnalysis contains detailed information about a CST node
type NodeAnalysis struct {
	Valid      bool              // Whether the node is valid
	Type       string            // Node type
	StartByte  uint              // Start position in source
	EndByte    uint              // End position in source
	ChildCount int               // Number of child nodes
	Children   []ChildInfo       // Information about child nodes
	Pattern    *ConstructPattern // Matched construct pattern, if any
	Error      string            // Error message if not valid
}

// ChildInfo contains information about a child node
type ChildInfo struct {
	Index int    // Index of child
	Type  string // Node type
	Text  string // Node text content
}

// IdentifyPattern attempts to identify which construct pattern this node matches
func (ca *CSTAnalyzer) IdentifyPattern(node *sitter.Node) *ConstructPattern {
	if node == nil {
		return nil
	}

	patterns := GetConstructPatterns()
	for _, pattern := range patterns {
		if node.Kind() == pattern.NodeType {
			return &pattern
		}
	}

	return nil
}

// IsFieldDeclaration checks if a variable declaration is a field declaration
func (ca *CSTAnalyzer) IsFieldDeclaration(node *sitter.Node) bool {
	if node == nil {
		return false
	}

	// Look for "field" keyword in the variable declaration
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Kind() == NodeField {
			return true
		}
	}

	return false
}

// ExtractTypeAnnotation extracts type annotation text from a construct (returns empty string if no type)
func (ca *CSTAnalyzer) ExtractTypeAnnotation(node *sitter.Node) string {
	if node == nil {
		return ""
	}

	// Find the type expression node (works for any construct)
	typeExpr := ca.findTypeExpression(node)
	if typeExpr == nil {
		return ""
	}

	return strings.TrimSpace(ca.getNodeText(typeExpr))
}

// HasTypeAnnotation checks if a construct has a type annotation
func (ca *CSTAnalyzer) HasTypeAnnotation(node *sitter.Node) bool {
	return ca.findTypeExpression(node) != nil
}

// findTypeExpression finds the first type expression in a node
func (ca *CSTAnalyzer) findTypeExpression(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}

	var typeExpr *sitter.Node
	ca.walkCST(node, func(n *sitter.Node) {
		if n.Kind() == NodeTypeExpression && typeExpr == nil {
			typeExpr = n
		}
	})

	return typeExpr
}

// ExtractVariableName extracts variable name from a construct (returns empty string if no variable)
func (ca *CSTAnalyzer) ExtractVariableName(node *sitter.Node) string {
	if node == nil {
		return ""
	}

	// Find variable name - works for any construct that has variables
	varName := ca.findVariableName(node)
	if varName == nil {
		return ""
	}

	return strings.TrimSpace(ca.getNodeText(varName))
}

// findVariableName finds the variable name node, handling both scalar and array variables
func (ca *CSTAnalyzer) findVariableName(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}

	// Look for varname nodes in both scalar and array contexts
	var varnameNode *sitter.Node
	ca.walkCST(node, func(n *sitter.Node) {
		if n.Kind() == NodeVarname && varnameNode == nil {
			// Check if this varname is in a scalar or array context
			parent := n.Parent()
			if parent != nil && (parent.Kind() == NodeScalar || parent.Kind() == NodeArray) {
				varnameNode = n
			}
		}
	})

	return varnameNode
}

// IsTypeAnnotationNode returns true if the node represents a type annotation
func (ca *CSTAnalyzer) IsTypeAnnotationNode(node *sitter.Node) bool {
	if node == nil {
		return false
	}

	typeNodes := []string{
		NodeTypeExpression,
		NodeSimpleType,
		NodeTypeAssertion,
		"parameterized_type",
		"union_type",
		"type_parameter_list",
		"intersection_type",
		"negation_type",
	}

	for _, nodeType := range typeNodes {
		if node.Kind() == nodeType {
			return true
		}
	}

	return false
}

// FindAllConstructs finds all recognized constructs in the CST
func (ca *CSTAnalyzer) FindAllConstructs() []Construct {
	var constructs []Construct
	ca.walkCST(ca.root, func(node *sitter.Node) {
		if pattern := ca.IdentifyPattern(node); pattern != nil {
			construct := Construct{
				Node:         node,
				Pattern:      *pattern,
				TypeText:     ca.ExtractTypeAnnotation(node),
				VariableName: ca.ExtractVariableName(node),
				HasType:      ca.HasTypeAnnotation(node),
				StartByte:    node.StartByte(),
				EndByte:      node.EndByte(),
			}
			constructs = append(constructs, construct)
		}
	})
	return constructs
}

// FindAllTypedConstructs finds all constructs that have type annotations
func (ca *CSTAnalyzer) FindAllTypedConstructs() []Construct {
	allConstructs := ca.FindAllConstructs()
	var typedConstructs []Construct
	for _, construct := range allConstructs {
		if construct.HasType {
			typedConstructs = append(typedConstructs, construct)
		}
	}
	return typedConstructs
}

// Construct represents a construct found in the CST
type Construct struct {
	Node         *sitter.Node     // The CST node
	Pattern      ConstructPattern // The pattern it matches
	TypeText     string           // Extracted type annotation (empty if no type)
	VariableName string           // Extracted variable name (empty if no variable)
	HasType      bool             // Whether this construct has a type annotation
	StartByte    uint             // Start position
	EndByte      uint             // End position
}

// walkCST recursively walks the CST and calls the visitor function for each node
func (ca *CSTAnalyzer) walkCST(node *sitter.Node, visitor func(*sitter.Node)) {
	if node == nil {
		return
	}

	visitor(node)

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			ca.walkCST(child, visitor)
		}
	}
}

// PrintCSTStructure prints a hierarchical view of the CST structure
func (ca *CSTAnalyzer) PrintCSTStructure() string {
	var builder strings.Builder
	ca.printNode(ca.root, "", true, &builder)
	return builder.String()
}

// printNode recursively prints a node and its children with indentation
func (ca *CSTAnalyzer) printNode(node *sitter.Node, prefix string, isLast bool, builder *strings.Builder) {
	if node == nil {
		return
	}

	// Print current node
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	content := strings.ReplaceAll(ca.getNodeText(node), "\n", "\\n")
	if len(content) > 50 {
		content = content[:47] + "..."
	}

	builder.WriteString(fmt.Sprintf("%s%s%s [%d:%d] \"%s\"\n",
		prefix, connector, node.Kind(), node.StartByte(), node.EndByte(), content))

	// Print children
	childCount := node.ChildCount()
	for i := uint(0); i < childCount; i++ {
		child := node.Child(i)
		if child != nil {
			isChildLast := i == childCount-1
			childPrefix := prefix
			if isLast {
				childPrefix += "    "
			} else {
				childPrefix += "│   "
			}
			ca.printNode(child, childPrefix, isChildLast, builder)
		}
	}
}

// getNodeText extracts text content from a tree-sitter node
func (ca *CSTAnalyzer) getNodeText(node *sitter.Node) string {
	if node == nil || ca.content == nil {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	if start >= uint(len(ca.content)) || end > uint(len(ca.content)) {
		return ""
	}
	return string(ca.content[start:end])
}
