// ABOUTME: Core AST type definitions consolidating scattered node types
// ABOUTME: Provides unified interfaces following TypeScript-Go astnav patterns

package ast

import (
	"strings"
)

// Position represents a position in source code
// Consolidated from multiple package definitions
type Position struct {
	Line   int
	Column int
	Offset int
}

// Node represents a node in the Abstract Syntax Tree
// Unified interface from parser and treesitter packages
type Node interface {
	// Type returns the syntactic type of the node
	Type() string

	// Start returns the start position of the node in source
	Start() Position

	// End returns the end position of the node in source
	End() Position

	// Children returns the child nodes
	Children() []Node

	// Text returns the source text covered by this node
	Text() string

	// Parent returns the parent node (nil for root)
	Parent() Node

	// SetParent sets the parent node (for tree construction)
	SetParent(parent Node)
}

// AST represents the Abstract Syntax Tree of a parsed file
// Consolidated from parser and treesitter AST types
type AST struct {
	// Path is the source file path
	Path string

	// Root is the root node of the syntax tree
	Root Node

	// TypeAnnotations contains all type annotations found
	TypeAnnotations []*TypeAnnotation

	// Errors contains any parse errors encountered
	Errors []error

	// Source contains the original source text
	Source string
}

// AST implements Node interface to allow navigation
func (a *AST) Type() string {
	return "ast"
}

func (a *AST) Start() Position {
	if a.Root != nil {
		return a.Root.Start()
	}
	return Position{Line: 1, Column: 1, Offset: 0}
}

func (a *AST) End() Position {
	if a.Root != nil {
		return a.Root.End()
	}
	return Position{Line: 1, Column: 1, Offset: 0}
}

func (a *AST) Children() []Node {
	if a.Root != nil {
		return []Node{a.Root}
	}
	return []Node{}
}

func (a *AST) Text() string {
	return a.Source
}

func (a *AST) Parent() Node {
	return nil // AST is the root, has no parent
}

func (a *AST) SetParent(parent Node) {
	// AST is the root, cannot set parent
}

// TypeAnnotation represents a type annotation in the code
// Consolidated from parser and treesitter TypeAnnotation types
type TypeAnnotation struct {
	// AnnotatedItem is the name of the annotated item (variable, function, etc.)
	AnnotatedItem string

	// TypeExpression is the type expression itself
	TypeExpression *TypeExpression

	// Position is the location of the annotation in source
	Pos Position

	// Kind indicates what kind of item is being annotated
	Kind AnnotationKind
}

// AnnotationKind represents the kind of type annotation
type AnnotationKind int

const (
	// VarAnnotation is a variable type annotation (my Type $var)
	VarAnnotation AnnotationKind = iota

	// SubParamAnnotation is a subroutine parameter annotation
	SubParamAnnotation

	// SubReturnAnnotation is a subroutine return type annotation
	SubReturnAnnotation

	// MethodParamAnnotation is a method parameter annotation
	MethodParamAnnotation

	// MethodReturnAnnotation is a method return type annotation
	MethodReturnAnnotation

	// FieldAnnotation is a field/attribute annotation
	FieldAnnotation

	// TypeDeclAnnotation is a type declaration
	TypeDeclAnnotation
)

// String returns a string representation of the annotation kind
func (ak AnnotationKind) String() string {
	switch ak {
	case VarAnnotation:
		return "variable"
	case SubParamAnnotation:
		return "subroutine_parameter"
	case SubReturnAnnotation:
		return "subroutine_return"
	case MethodParamAnnotation:
		return "method_parameter"
	case MethodReturnAnnotation:
		return "method_return"
	case FieldAnnotation:
		return "field"
	case TypeDeclAnnotation:
		return "type_declaration"
	default:
		return "unknown"
	}
}

// TypeExpression represents a type expression
// Consolidated and enhanced from parser and treesitter implementations
type TypeExpression struct {
	// BaseType is the primary type name
	BaseType string

	// Parameters are type parameters for parameterized types (e.g., ArrayRef[Int])
	Parameters []*TypeExpression

	// IsUnion indicates this is a union type (Type1|Type2)
	IsUnion bool

	// IsIntersection indicates this is an intersection type (Type1&Type2)
	IsIntersection bool

	// IsNegation indicates this is a negation type (!Type)
	IsNegation bool

	// UnionTypes contains the types in a union (for multi-way unions)
	UnionTypes []*TypeExpression

	// IntersectionTypes contains the types in an intersection
	IntersectionTypes []*TypeExpression

	// NegatedType is the type being negated (for negation types)
	NegatedType *TypeExpression

	// OriginalString preserves the original source text
	OriginalString string

	// Pos is the position in source where this type appears
	Pos Position
}

// String returns a string representation of the type expression
func (te *TypeExpression) String() string {
	if te == nil {
		return ""
	}

	// Use original string if available
	if te.OriginalString != "" {
		return te.OriginalString
	}

	// Handle negation
	if te.IsNegation && te.NegatedType != nil {
		return "!" + te.NegatedType.String()
	}

	// Handle union types
	if te.IsUnion && len(te.UnionTypes) > 0 {
		var parts []string
		for _, unionType := range te.UnionTypes {
			parts = append(parts, unionType.String())
		}
		return strings.Join(parts, "|")
	}

	// Handle intersection types
	if te.IsIntersection && len(te.IntersectionTypes) > 0 {
		var parts []string
		for _, intType := range te.IntersectionTypes {
			parts = append(parts, intType.String())
		}
		return strings.Join(parts, "&")
	}

	// Handle parameterized types
	if len(te.Parameters) > 0 {
		var params []string
		for _, param := range te.Parameters {
			params = append(params, param.String())
		}
		return te.BaseType + "[" + strings.Join(params, ", ") + "]"
	}

	// Simple type
	return te.BaseType
}

// IsSimple returns true if this is a simple, non-compound type
func (te *TypeExpression) IsSimple() bool {
	return !te.IsUnion && !te.IsIntersection && !te.IsNegation && len(te.Parameters) == 0
}

// GetAllTypes returns all types mentioned in this expression (for unions/intersections)
func (te *TypeExpression) GetAllTypes() []string {
	var types []string

	switch {
	case te.IsUnion && len(te.UnionTypes) > 0:
		for _, unionType := range te.UnionTypes {
			types = append(types, unionType.GetAllTypes()...)
		}
	case te.IsIntersection && len(te.IntersectionTypes) > 0:
		for _, intType := range te.IntersectionTypes {
			types = append(types, intType.GetAllTypes()...)
		}
	case te.IsNegation && te.NegatedType != nil:
		types = append(types, te.NegatedType.GetAllTypes()...)
	default:
		types = append(types, te.BaseType)
	}

	return types
}

// BaseNode provides a common implementation for AST nodes
// This reduces boilerplate for concrete node implementations
type BaseNode struct {
	nodeType string
	start    Position
	end      Position
	children []Node
	parent   Node
	text     string
}

// NewBaseNode creates a new base node with the given type
func NewBaseNode(nodeType string, start, end Position) *BaseNode {
	return &BaseNode{
		nodeType: nodeType,
		start:    start,
		end:      end,
		children: make([]Node, 0),
	}
}

// Type implements Node interface
func (bn *BaseNode) Type() string {
	return bn.nodeType
}

// Start implements Node interface
func (bn *BaseNode) Start() Position {
	return bn.start
}

// End implements Node interface
func (bn *BaseNode) End() Position {
	return bn.end
}

// Children implements Node interface
func (bn *BaseNode) Children() []Node {
	return bn.children
}

// Text implements Node interface
func (bn *BaseNode) Text() string {
	return bn.text
}

// Parent implements Node interface
func (bn *BaseNode) Parent() Node {
	return bn.parent
}

// SetParent implements Node interface
func (bn *BaseNode) SetParent(parent Node) {
	bn.parent = parent
}

// AddChild adds a child node and sets its parent
func (bn *BaseNode) AddChild(child Node) {
	if child != nil {
		child.SetParent(bn)
		bn.children = append(bn.children, child)
	}
}

// SetText sets the text content of this node
func (bn *BaseNode) SetText(text string) {
	bn.text = text
}

// NodeVisitor defines a function that visits AST nodes
// Used for AST traversal operations
type NodeVisitor func(node Node) bool

// WalkFunc defines a function for AST walking with enter/exit hooks
type WalkFunc struct {
	Enter func(node Node) bool // Return false to skip children
	Exit  func(node Node)      // Called when leaving node
}
