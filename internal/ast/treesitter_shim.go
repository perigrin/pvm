// ABOUTME: Tree-sitter shim implementing ast.Node interface over tree-sitter CST
// ABOUTME: Provides direct CST access while maintaining compatibility with existing AST interfaces

package ast

import (
	"errors"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// TreeSitterNode wraps a tree-sitter node to implement the ast.Node interface
// This provides a thin shim around the CST without conversion, preserving all information
type TreeSitterNode struct {
	// node is the underlying tree-sitter node
	node *sitter.Node

	// source is the original source content for text extraction
	source []byte

	// parent is the cached parent wrapper (tree-sitter nodes are immutable)
	parent Node

	// children is the cached children wrappers
	children []Node
}

// NewTreeSitterNode creates a new tree-sitter node wrapper
func NewTreeSitterNode(node *sitter.Node, source []byte) *TreeSitterNode {
	if node == nil {
		return nil
	}

	return &TreeSitterNode{
		node:   node,
		source: source,
	}
}

// GetTreeSitterNode returns the underlying tree-sitter node
// This allows direct CST access when needed
func (ts *TreeSitterNode) GetTreeSitterNode() *sitter.Node {
	return ts.node
}

// Type returns the syntactic type of the node
func (ts *TreeSitterNode) Type() string {
	if ts.node == nil {
		return "unknown"
	}
	return ts.node.Kind()
}

// Start returns the start position of the node in source
func (ts *TreeSitterNode) Start() Position {
	if ts.node == nil {
		return Position{}
	}

	pos := ts.node.StartPosition()
	return Position{
		Line:   int(pos.Row) + 1,    // tree-sitter uses 0-based rows, we use 1-based
		Column: int(pos.Column) + 1, // tree-sitter uses 0-based columns, we use 1-based
		Offset: int(ts.node.StartByte()),
	}
}

// End returns the end position of the node in source
func (ts *TreeSitterNode) End() Position {
	if ts.node == nil {
		return Position{}
	}

	pos := ts.node.EndPosition()
	return Position{
		Line:   int(pos.Row) + 1,    // tree-sitter uses 0-based rows, we use 1-based
		Column: int(pos.Column) + 1, // tree-sitter uses 0-based columns, we use 1-based
		Offset: int(ts.node.EndByte()),
	}
}

// Children returns the child nodes
func (ts *TreeSitterNode) Children() []Node {
	if ts.node == nil {
		return nil
	}

	// Cache children to avoid repeated wrapping
	if ts.children == nil {
		childCount := ts.node.ChildCount()
		ts.children = make([]Node, 0, childCount)

		for i := uint(0); i < childCount; i++ {
			child := ts.node.Child(i)
			if child != nil {
				childWrapper := NewTreeSitterNode(child, ts.source)
				childWrapper.parent = ts
				ts.children = append(ts.children, childWrapper)
			}
		}
	}

	return ts.children
}

// Text returns the source text covered by this node
func (ts *TreeSitterNode) Text() string {
	if ts.node == nil || ts.source == nil {
		return ""
	}

	return ts.node.Utf8Text(ts.source)
}

// Parent returns the parent node (nil for root)
func (ts *TreeSitterNode) Parent() Node {
	return ts.parent
}

// SetParent sets the parent node (for tree construction)
func (ts *TreeSitterNode) SetParent(parent Node) {
	ts.parent = parent
}

// TreeSitterAST wraps a tree-sitter tree to implement the AST interface
type TreeSitterAST struct {
	// Path is the source file path
	Path string

	// Root is the root node wrapper
	Root *TreeSitterNode

	// Source contains the original source text
	Source string

	// TypeAnnotations contains extracted type annotations
	TypeAnnotations []*TypeAnnotation

	// Errors contains any parse errors encountered
	Errors []error

	// tree is the underlying tree-sitter tree (kept for lifecycle management)
	tree *sitter.Tree
}

// NewTreeSitterAST creates a new tree-sitter AST wrapper
func NewTreeSitterAST(path string, tree *sitter.Tree, source string) *TreeSitterAST {
	if tree == nil {
		return nil
	}

	sourceBytes := []byte(source)
	var root *TreeSitterNode

	if tree.RootNode() != nil {
		root = NewTreeSitterNode(tree.RootNode(), sourceBytes)
	}

	return &TreeSitterAST{
		Path:            path,
		Root:            root,
		Source:          source,
		TypeAnnotations: make([]*TypeAnnotation, 0),
		Errors:          make([]error, 0),
		tree:            tree,
	}
}

// GetPath returns the source file path (compiler.AST interface)
func (ts *TreeSitterAST) GetPath() string {
	return ts.Path
}

// IsValid returns true if the AST is valid for compilation (compiler.AST interface)
func (ts *TreeSitterAST) IsValid() bool {
	return len(ts.Errors) == 0 && ts.Root != nil
}

// GetContent returns the original source content (compiler.AST interface)
func (ts *TreeSitterAST) GetContent() (string, error) {
	return ts.Source, nil
}

// GetRootNode returns the root AST node (compiler.AST interface)
func (ts *TreeSitterAST) GetRootNode() (Node, error) {
	if ts.Root == nil {
		return nil, errors.New("no root node available")
	}
	return ts.Root, nil
}

// GetTree returns the underlying tree-sitter tree
// This allows direct CST access for compilation
func (ts *TreeSitterAST) GetTree() *sitter.Tree {
	return ts.tree
}

// GetTreeSitterRoot returns the underlying tree-sitter root node
// This is a convenience method for direct CST access
func (ts *TreeSitterAST) GetTreeSitterRoot() *sitter.Node {
	if ts.tree == nil {
		return nil
	}
	return ts.tree.RootNode()
}

// Utility functions for working with tree-sitter nodes

// ConvertTreeSitterPosition converts a tree-sitter Point to our Position
func ConvertTreeSitterPosition(point sitter.Point, offset uint) Position {
	return Position{
		Line:   int(point.Row) + 1,    // Convert from 0-based to 1-based
		Column: int(point.Column) + 1, // Convert from 0-based to 1-based
		Offset: int(offset),
	}
}

// FindNodeByType searches for the first child node of a specific type
func (ts *TreeSitterNode) FindNodeByType(nodeType string) *TreeSitterNode {
	if ts.node == nil {
		return nil
	}

	// Check direct children first
	for i := uint(0); i < ts.node.ChildCount(); i++ {
		child := ts.node.Child(i)
		if child != nil && child.Kind() == nodeType {
			return NewTreeSitterNode(child, ts.source)
		}
	}

	return nil
}

// FindAllNodesByType searches for all child nodes of a specific type
func (ts *TreeSitterNode) FindAllNodesByType(nodeType string) []*TreeSitterNode {
	if ts.node == nil {
		return nil
	}

	var results []*TreeSitterNode

	for i := uint(0); i < ts.node.ChildCount(); i++ {
		child := ts.node.Child(i)
		if child != nil && child.Kind() == nodeType {
			results = append(results, NewTreeSitterNode(child, ts.source))
		}
	}

	return results
}

// GetChildByIndex returns a specific child by index
func (ts *TreeSitterNode) GetChildByIndex(index uint) *TreeSitterNode {
	if ts.node == nil || index >= ts.node.ChildCount() {
		return nil
	}

	child := ts.node.Child(index)
	if child == nil {
		return nil
	}

	return NewTreeSitterNode(child, ts.source)
}

// IsOfType checks if this node is of the specified type
func (ts *TreeSitterNode) IsOfType(nodeType string) bool {
	return ts.Type() == nodeType
}

// HasChild checks if this node has a child of the specified type
func (ts *TreeSitterNode) HasChild(nodeType string) bool {
	return ts.FindNodeByType(nodeType) != nil
}

// GetTextContent returns the text content, trimmed of whitespace
func (ts *TreeSitterNode) GetTextContent() string {
	return strings.TrimSpace(ts.Text())
}

// Additional TreeSitterNode utilities for position conversion and node traversal

// GetChildByName returns a child node by its field name (if it's a named field)
func (ts *TreeSitterNode) GetChildByName(fieldName string) *TreeSitterNode {
	if ts.node == nil {
		return nil
	}

	child := ts.node.ChildByFieldName(fieldName)
	if child == nil {
		return nil
	}

	return NewTreeSitterNode(child, ts.source)
}

// GetNamedChildren returns only the named children (excluding anonymous tokens)
func (ts *TreeSitterNode) GetNamedChildren() []*TreeSitterNode {
	if ts.node == nil {
		return nil
	}

	var results []*TreeSitterNode

	for i := uint(0); i < ts.node.NamedChildCount(); i++ {
		child := ts.node.NamedChild(i)
		if child != nil {
			results = append(results, NewTreeSitterNode(child, ts.source))
		}
	}

	return results
}

// IsNamed returns true if this is a named node (not an anonymous token)
func (ts *TreeSitterNode) IsNamed() bool {
	if ts.node == nil {
		return false
	}
	return ts.node.IsNamed()
}

// IsError returns true if this node represents a parse error
func (ts *TreeSitterNode) IsError() bool {
	if ts.node == nil {
		return false
	}
	return ts.node.IsError()
}

// IsMissing returns true if this node is missing from the parse
func (ts *TreeSitterNode) IsMissing() bool {
	if ts.node == nil {
		return false
	}
	return ts.node.IsMissing()
}

// GetFieldName returns the field name of this node if it's a named field
//
//nolint:sloppyTypeAssert // Function needs type assertions from Node interface to TreeSitterNode
func (ts *TreeSitterNode) GetFieldName() string {
	if ts.node == nil || ts.parent == nil {
		return ""
	}

	// Get the parent's tree-sitter node
	// TODO(#366): go-critic thinks this type assertion is redundant because ts.parent is Node interface,
	// but we need to assert to *TreeSitterNode specifically to access tree-sitter fields
	parentTS, ok := ts.parent.(*TreeSitterNode) //nolint:gocritic,sloppyTypeAssert
	if !ok || parentTS.node == nil {
		return ""
	}

	// Find this node among the parent's children to get its field name
	for i := uint(0); i < parentTS.node.ChildCount(); i++ {
		child := parentTS.node.Child(i)
		if child != nil && child.Id() == ts.node.Id() {
			return parentTS.node.FieldNameForChild(uint32(i))
		}
	}

	return ""
}

// ToTreeSitterPosition converts our Position to tree-sitter Point
func ToTreeSitterPosition(pos Position) sitter.Point {
	return sitter.Point{
		Row:    uint(pos.Line - 1),   // Convert from 1-based to 0-based
		Column: uint(pos.Column - 1), // Convert from 1-based to 0-based
	}
}

// WalkNodes performs a depth-first traversal of the tree starting from this node
//
//nolint:sloppyTypeAssert // Function needs type assertions from Node interface to TreeSitterNode
func (ts *TreeSitterNode) WalkNodes(visitor func(*TreeSitterNode) bool) {
	if ts.node == nil {
		return
	}

	// Visit this node first
	if !visitor(ts) {
		return // visitor returned false, stop traversal
	}

	// Then visit children
	for _, child := range ts.Children() {
		// TODO(#366): go-critic sees Node interface but we need *TreeSitterNode for tree-sitter methods
		if tsChild, ok := child.(*TreeSitterNode); ok { //nolint:gocritic,sloppyTypeAssert
			tsChild.WalkNodes(visitor)
		}
	}
}

// FindNodeRecursive searches recursively for a node matching the predicate
//
//nolint:sloppyTypeAssert // Function needs type assertions from Node interface to TreeSitterNode
func (ts *TreeSitterNode) FindNodeRecursive(predicate func(*TreeSitterNode) bool) *TreeSitterNode {
	if ts.node == nil {
		return nil
	}

	// Check this node first
	if predicate(ts) {
		return ts
	}

	// Then search children recursively
	for _, child := range ts.Children() {
		// TODO(#366): go-critic sees Node interface but we need *TreeSitterNode for recursive calls
		if tsChild, ok := child.(*TreeSitterNode); ok { //nolint:gocritic,sloppyTypeAssert
			if result := tsChild.FindNodeRecursive(predicate); result != nil {
				return result
			}
		}
	}

	return nil
}

// GetAncestor traverses up the tree to find the first ancestor of the specified type
//
//nolint:sloppyTypeAssert // Function needs type assertions from Node interface to TreeSitterNode
func (ts *TreeSitterNode) GetAncestor(nodeType string) *TreeSitterNode {
	current := ts.parent
	for current != nil {
		// TODO(#366): go-critic sees Node interface but we need *TreeSitterNode to access .parent field
		if tsParent, ok := current.(*TreeSitterNode); ok { //nolint:gocritic,sloppyTypeAssert
			if tsParent.Type() == nodeType {
				return tsParent
			}
			current = tsParent.parent
		} else {
			break
		}
	}
	return nil
}

// GetByteRange returns the start and end byte offsets for this node
func (ts *TreeSitterNode) GetByteRange() (uint, uint) {
	if ts.node == nil {
		return 0, 0
	}
	return ts.node.StartByte(), ts.node.EndByte()
}

// GetRange returns the Position range for this node
func (ts *TreeSitterNode) GetRange() (Position, Position) {
	return ts.Start(), ts.End()
}

// Utility methods for type-specific node access

// AsVarDecl attempts to interpret this node as a variable declaration
// This is a specialized method that extracts type information from CST
func (ts *TreeSitterNode) AsVarDecl() *VarDeclNode {
	if ts.node == nil || ts.Type() != "variable_declaration" {
		return nil
	}

	return &VarDeclNode{
		TreeSitterNode: ts,
	}
}

// VarDeclNode is a specialized wrapper for variable declaration nodes
// This provides direct access to type information from the CST
// It implements the VarDecl interface while preserving CST-direct access
type VarDeclNode struct {
	*TreeSitterNode

	// Cached extracted values to avoid repeated CST traversal
	cachedVariables []*VariableExpr
	cachedTypeExpr  *TypeExpression
	cachedKeyword   string
	cachedInit      Node
}

// GetVariableName extracts the variable name from the declaration
func (vd *VarDeclNode) GetVariableName() string {
	if vd.TreeSitterNode == nil {
		return ""
	}

	// Look for the variable identifier
	if varNode := vd.GetChildByName("variable"); varNode != nil {
		return varNode.GetTextContent()
	}

	// Alternative: search for scalar variable patterns
	for _, child := range vd.GetNamedChildren() {
		if child.Type() == "scalar_variable" {
			return child.GetTextContent()
		}
	}

	return ""
}

// GetTypeExpression extracts the type annotation if present
// This preserves the original type information from the CST
func (vd *VarDeclNode) GetTypeExpression() *TypeExpression {
	if vd.TreeSitterNode == nil {
		return nil
	}

	// Look for type annotation in the CST
	if typeNode := vd.GetChildByName("type"); typeNode != nil {
		start := typeNode.Start()
		end := typeNode.End()
		return NewTypeExpression(typeNode.GetTextContent(), nil, start, end)
	}

	return nil
}

// HasTypeAnnotation returns true if this variable declaration has a type annotation
func (vd *VarDeclNode) HasTypeAnnotation() bool {
	return vd.GetTypeExpression() != nil
}

// VarDecl interface implementation for compatibility with existing code

// GetKeyword returns the declaration keyword (my, our, state, local)
func (vd *VarDeclNode) GetKeyword() string {
	if vd.cachedKeyword != "" {
		return vd.cachedKeyword
	}

	if vd.TreeSitterNode == nil {
		return ""
	}

	// Look for the keyword token in the CST
	for _, child := range vd.GetNamedChildren() {
		childType := child.Type()
		if childType == "my" || childType == "our" || childType == "state" || childType == "local" {
			vd.cachedKeyword = child.GetTextContent()
			return vd.cachedKeyword
		}
	}

	// Fallback: check the first token
	if vd.node.ChildCount() > 0 {
		firstChild := vd.node.Child(0)
		if firstChild != nil {
			keyword := firstChild.Utf8Text(vd.source)
			if keyword == "my" || keyword == "our" || keyword == "state" || keyword == "local" {
				vd.cachedKeyword = keyword
				return vd.cachedKeyword
			}
		}
	}

	return ""
}

// GetVariables returns the list of variables being declared
func (vd *VarDeclNode) GetVariables() []*VariableExpr {
	if vd.cachedVariables != nil {
		return vd.cachedVariables
	}

	if vd.TreeSitterNode == nil {
		return nil
	}

	var variables []*VariableExpr

	// Search for variable nodes in the CST
	for _, child := range vd.GetNamedChildren() {
		childType := child.Type()

		// Handle different variable types
		switch childType {
		case "scalar", "array", "hash", "scalar_variable", "array_variable", "hash_variable":
			varText := child.GetTextContent()
			if varText != "" {
				// Extract variable name and sigil
				sigil := ""
				name := varText
				if len(varText) > 0 {
					sigil = string(varText[0])
					if len(varText) > 1 {
						name = varText[1:]
					}
				}

				pos := child.Start()
				end := child.End()

				varExpr := NewVariableExpr(name, sigil, pos, end)
				variables = append(variables, varExpr)
			}

		case "variable_list":
			// Handle lists of variables like (my $a, $b, $c)
			for _, listChild := range child.GetNamedChildren() {
				if listChild.Type() == "scalar" || listChild.Type() == "array" || listChild.Type() == "hash" ||
					listChild.Type() == "scalar_variable" ||
					listChild.Type() == "array_variable" ||
					listChild.Type() == "hash_variable" {
					varText := listChild.GetTextContent()
					if varText != "" {
						sigil := ""
						name := varText
						if len(varText) > 0 {
							sigil = string(varText[0])
							if len(varText) > 1 {
								name = varText[1:]
							}
						}

						pos := listChild.Start()
						end := listChild.End()

						varExpr := NewVariableExpr(name, sigil, pos, end)
						variables = append(variables, varExpr)
					}
				}
			}
		}
	}

	vd.cachedVariables = variables
	return vd.cachedVariables
}

// GetTypeExpr returns the type expression for this declaration
func (vd *VarDeclNode) GetTypeExpr() *TypeExpression {
	if vd.cachedTypeExpr != nil {
		return vd.cachedTypeExpr
	}

	// Use the existing GetTypeExpression method
	vd.cachedTypeExpr = vd.GetTypeExpression()
	return vd.cachedTypeExpr
}

// GetInit returns the initialization expression if present
func (vd *VarDeclNode) GetInit() Node {
	if vd.cachedInit != nil {
		return vd.cachedInit
	}

	if vd.TreeSitterNode == nil {
		return nil
	}

	// The initialization value is typically in the parent assignment_expression node
	// Look up to the parent to find the assignment value
	parentNode := vd.Parent()
	if parentNode != nil {
		// Check if parent is an assignment_expression
		//nolint:gocritic,sloppyTypeAssert // TODO(#366): go-critic sees Node interface but we need *TreeSitterNode for tree-sitter methods
		if tsParent, ok := parentNode.(*TreeSitterNode); ok && tsParent.Type() == "assignment_expression" {
			// Find the value part after the "=" operator
			for i := uint(0); i < tsParent.node.ChildCount(); i++ {
				child := tsParent.node.Child(i)
				if child != nil && child.Utf8Text(tsParent.source) == "=" {
					// The next sibling should be the initialization value
					if i+1 < tsParent.node.ChildCount() {
						nextChild := tsParent.node.Child(i + 1)
						if nextChild != nil {
							initNode := NewTreeSitterNode(nextChild, tsParent.source)
							vd.cachedInit = initNode
							return vd.cachedInit
						}
					}
				}
			}
		}
	}

	// Fallback: look for assignment or initialization expressions in children
	for _, child := range vd.GetNamedChildren() {
		childType := child.Type()

		// Look for assignment operators and following expressions
		if childType == "assignment_expression" ||
			childType == "expression" ||
			childType == "literal_expression" ||
			childType == "number" ||
			childType == "string" {
			vd.cachedInit = child
			return vd.cachedInit
		}
	}

	return nil
}

// GetPosition returns the position of this declaration
func (vd *VarDeclNode) GetPosition() Position {
	return vd.Start()
}

// IsTyped returns true if this declaration has a type annotation
func (vd *VarDeclNode) IsTyped() bool {
	return vd.HasTypeAnnotation()
}

// GetTypedVariable returns the first variable with its type information
// This is a convenience method for single-variable declarations
func (vd *VarDeclNode) GetTypedVariable() (*VariableExpr, *TypeExpression) {
	variables := vd.GetVariables()
	if len(variables) == 0 {
		return nil, nil
	}

	return variables[0], vd.GetTypeExpr()
}

// GetDeclarationKind returns the kind of declaration as a string
func (vd *VarDeclNode) GetDeclarationKind() string {
	// This would need to be determined from context
	// For now, assume it's a variable declaration
	return "variable"
}

// Clone creates a copy of this VarDeclNode
func (vd *VarDeclNode) Clone() *VarDeclNode {
	if vd.TreeSitterNode == nil {
		return nil
	}

	// Create a new wrapper around the same tree-sitter node
	// Note: tree-sitter nodes are immutable, so this is safe
	return &VarDeclNode{
		TreeSitterNode: &TreeSitterNode{
			node:     vd.node,
			source:   vd.source,
			parent:   vd.parent,
			children: vd.children,
		},
		// Don't copy cached values - let them be recomputed
	}
}

// String returns a string representation of this variable declaration
func (vd *VarDeclNode) String() string {
	if vd.TreeSitterNode == nil {
		return "VarDeclNode{nil}"
	}

	keyword := vd.GetKeyword()
	variables := vd.GetVariables()
	typeExpr := vd.GetTypeExpr()

	result := "VarDeclNode{" + keyword

	if typeExpr != nil {
		result += " " + typeExpr.String()
	}

	for i, variable := range variables {
		if i > 0 {
			result += ", "
		} else {
			result += " "
		}
		result += variable.Sigil + variable.Name
	}

	if init := vd.GetInit(); init != nil {
		result += " = " + init.Text()
	}

	result += "}"
	return result
}
