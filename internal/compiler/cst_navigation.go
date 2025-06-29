// ABOUTME: Navigation helpers for tree-sitter CST nodes in typed Perl
// ABOUTME: Provides utilities for traversing and querying CST structure efficiently

package compiler

import (
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// CSTNavigator provides efficient navigation utilities for tree-sitter CST
type CSTNavigator struct {
	root    *sitter.Node
	content []byte // Source content for extracting node text
}

// NewCSTNavigator creates a new CST navigator for the given root node
func NewCSTNavigator(root *sitter.Node) *CSTNavigator {
	return &CSTNavigator{root: root}
}

// NewCSTNavigatorWithContent creates a new CST navigator with source content for text extraction
func NewCSTNavigatorWithContent(root *sitter.Node, content []byte) *CSTNavigator {
	return &CSTNavigator{root: root, content: content}
}

// FindNodesByType finds all nodes of the specified type in the CST
func (nav *CSTNavigator) FindNodesByType(nodeType string) []*sitter.Node {
	var found []*sitter.Node
	nav.walkCST(nav.root, func(node *sitter.Node) {
		if node.Kind() == nodeType {
			found = append(found, node)
		}
	})
	return found
}

// FindNodesByTypes finds all nodes matching any of the specified types
func (nav *CSTNavigator) FindNodesByTypes(nodeTypes []string) []*sitter.Node {
	typeSet := make(map[string]bool)
	for _, t := range nodeTypes {
		typeSet[t] = true
	}

	var found []*sitter.Node
	nav.walkCST(nav.root, func(node *sitter.Node) {
		if typeSet[node.Kind()] {
			found = append(found, node)
		}
	})
	return found
}

// FindChildByType finds the first direct child of the specified type
func (nav *CSTNavigator) FindChildByType(parent *sitter.Node, nodeType string) *sitter.Node {
	if parent == nil {
		return nil
	}

	for i := uint(0); i < parent.ChildCount(); i++ {
		child := parent.Child(i)
		if child != nil && child.Kind() == nodeType {
			return child
		}
	}
	return nil
}

// FindChildrenByType finds all direct children of the specified type
func (nav *CSTNavigator) FindChildrenByType(parent *sitter.Node, nodeType string) []*sitter.Node {
	if parent == nil {
		return nil
	}

	var children []*sitter.Node
	for i := uint(0); i < parent.ChildCount(); i++ {
		child := parent.Child(i)
		if child != nil && child.Kind() == nodeType {
			children = append(children, child)
		}
	}
	return children
}

// FindDescendantByType finds the first descendant (at any depth) of the specified type
func (nav *CSTNavigator) FindDescendantByType(parent *sitter.Node, nodeType string) *sitter.Node {
	if parent == nil {
		return nil
	}

	// Check direct children first
	if child := nav.FindChildByType(parent, nodeType); child != nil {
		return child
	}

	// Recursively search in children
	for i := uint(0); i < parent.ChildCount(); i++ {
		child := parent.Child(i)
		if child != nil {
			if found := nav.FindDescendantByType(child, nodeType); found != nil {
				return found
			}
		}
	}
	return nil
}

// FindDescendantsByType finds all descendants (at any depth) of the specified type
func (nav *CSTNavigator) FindDescendantsByType(parent *sitter.Node, nodeType string) []*sitter.Node {
	if parent == nil {
		return nil
	}

	var descendants []*sitter.Node
	nav.walkCST(parent, func(node *sitter.Node) {
		if node.Kind() == nodeType && node != parent {
			descendants = append(descendants, node)
		}
	})
	return descendants
}

// GetSiblings returns all sibling nodes of the given node
func (nav *CSTNavigator) GetSiblings(node *sitter.Node) []*sitter.Node {
	if node == nil {
		return nil
	}

	parent := node.Parent()
	if parent == nil {
		return nil
	}

	var siblings []*sitter.Node
	for i := uint(0); i < parent.ChildCount(); i++ {
		child := parent.Child(i)
		if child != nil && child != node {
			siblings = append(siblings, child)
		}
	}
	return siblings
}

// GetNextSibling returns the next sibling node, if any
func (nav *CSTNavigator) GetNextSibling(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}

	parent := node.Parent()
	if parent == nil {
		return nil
	}

	// Find the node's index
	var nodeIndex uint = 0
	found := false
	for i := uint(0); i < parent.ChildCount(); i++ {
		if parent.Child(i) == node {
			nodeIndex = i
			found = true
			break
		}
	}

	if !found || nodeIndex+1 >= parent.ChildCount() {
		return nil
	}

	return parent.Child(nodeIndex + 1)
}

// GetPreviousSibling returns the previous sibling node, if any
func (nav *CSTNavigator) GetPreviousSibling(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}

	parent := node.Parent()
	if parent == nil {
		return nil
	}

	// Find the node's index
	var nodeIndex uint = 0
	found := false
	for i := uint(0); i < parent.ChildCount(); i++ {
		if parent.Child(i) == node {
			nodeIndex = i
			found = true
			break
		}
	}

	if !found || nodeIndex == 0 {
		return nil
	}

	return parent.Child(nodeIndex - 1)
}

// GetNodePath returns the path from root to the specified node
func (nav *CSTNavigator) GetNodePath(target *sitter.Node) []NodePathElement {
	if target == nil {
		return nil
	}

	var path []NodePathElement
	current := target

	for current != nil {
		element := NodePathElement{
			Node:  current,
			Type:  current.Kind(),
			Index: nav.getChildIndex(current),
		}
		path = append([]NodePathElement{element}, path...) // Prepend
		current = current.Parent()
	}

	return path
}

// NodePathElement represents one element in a path from root to a node
type NodePathElement struct {
	Node  *sitter.Node // The node at this path element
	Type  string       // Node type
	Index int          // Index among siblings (-1 if root)
}

// getChildIndex finds the index of a node among its siblings
func (nav *CSTNavigator) getChildIndex(node *sitter.Node) int {
	if node == nil {
		return -1
	}

	parent := node.Parent()
	if parent == nil {
		return -1 // Root node
	}

	for i := uint(0); i < parent.ChildCount(); i++ {
		if parent.Child(i) == node {
			return int(i)
		}
	}

	return -1 // Not found (shouldn't happen)
}

// FilterNodes filters nodes based on a predicate function
func (nav *CSTNavigator) FilterNodes(nodes []*sitter.Node, predicate func(*sitter.Node) bool) []*sitter.Node {
	var filtered []*sitter.Node
	for _, node := range nodes {
		if predicate(node) {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

// HasChildOfType checks if a node has a direct child of the specified type
func (nav *CSTNavigator) HasChildOfType(parent *sitter.Node, nodeType string) bool {
	return nav.FindChildByType(parent, nodeType) != nil
}

// HasDescendantOfType checks if a node has a descendant of the specified type
func (nav *CSTNavigator) HasDescendantOfType(parent *sitter.Node, nodeType string) bool {
	return nav.FindDescendantByType(parent, nodeType) != nil
}

// GetTextContent extracts clean text content from a node, handling whitespace
func (nav *CSTNavigator) GetTextContent(node *sitter.Node) string {
	if node == nil {
		return ""
	}
	return strings.TrimSpace(nav.getNodeText(node))
}

// GetSourceRange returns the source range (start, end) for a node
func (nav *CSTNavigator) GetSourceRange(node *sitter.Node) (uint, uint) {
	if node == nil {
		return 0, 0
	}
	return node.StartByte(), node.EndByte()
}

// IsLeafNode returns true if the node has no children
func (nav *CSTNavigator) IsLeafNode(node *sitter.Node) bool {
	return node != nil && node.ChildCount() == 0
}

// GetLeafNodes returns all leaf nodes under the given node
func (nav *CSTNavigator) GetLeafNodes(root *sitter.Node) []*sitter.Node {
	var leaves []*sitter.Node
	nav.walkCST(root, func(node *sitter.Node) {
		if nav.IsLeafNode(node) {
			leaves = append(leaves, node)
		}
	})
	return leaves
}

// walkCST recursively walks the CST and calls the visitor function for each node
func (nav *CSTNavigator) walkCST(node *sitter.Node, visitor func(*sitter.Node)) {
	if node == nil {
		return
	}

	visitor(node)

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			nav.walkCST(child, visitor)
		}
	}
}

// TypeAnnotationQuery provides specialized queries for type annotation nodes
type TypeAnnotationQuery struct {
	nav *CSTNavigator
}

// NewTypeAnnotationQuery creates a new type annotation query helper
func NewTypeAnnotationQuery(root *sitter.Node) *TypeAnnotationQuery {
	return &TypeAnnotationQuery{
		nav: NewCSTNavigator(root),
	}
}

// NewTypeAnnotationQueryWithContent creates a new type annotation query helper with content
func NewTypeAnnotationQueryWithContent(root *sitter.Node, content []byte) *TypeAnnotationQuery {
	return &TypeAnnotationQuery{
		nav: NewCSTNavigatorWithContent(root, content),
	}
}

// FindAllTypeExpressions finds all type expression nodes in the CST
func (taq *TypeAnnotationQuery) FindAllTypeExpressions() []*sitter.Node {
	typeNodes := []string{
		NodeTypeExpression,
		NodeSimpleType,
	}
	return taq.nav.FindNodesByTypes(typeNodes)
}

// FindAllTypeAssertions finds all type assertion expressions
func (taq *TypeAnnotationQuery) FindAllTypeAssertions() []*sitter.Node {
	return taq.nav.FindNodesByType(NodeTypeAssertion)
}

// FindAllVariableDeclarations finds all variable declaration nodes
func (taq *TypeAnnotationQuery) FindAllVariableDeclarations() []*sitter.Node {
	return taq.nav.FindNodesByType(NodeVariableDecl)
}

// FindTypedVariableDeclarations finds variable declarations with type annotations
func (taq *TypeAnnotationQuery) FindTypedVariableDeclarations() []*sitter.Node {
	allVarDecls := taq.FindAllVariableDeclarations()
	return taq.nav.FilterNodes(allVarDecls, func(node *sitter.Node) bool {
		return taq.nav.HasDescendantOfType(node, NodeTypeExpression)
	})
}

// FindUntypedVariableDeclarations finds variable declarations without type annotations
func (taq *TypeAnnotationQuery) FindUntypedVariableDeclarations() []*sitter.Node {
	allVarDecls := taq.FindAllVariableDeclarations()
	return taq.nav.FilterNodes(allVarDecls, func(node *sitter.Node) bool {
		return !taq.nav.HasDescendantOfType(node, NodeTypeExpression)
	})
}

// FindMethodDeclarations finds all method declaration nodes
func (taq *TypeAnnotationQuery) FindMethodDeclarations() []*sitter.Node {
	return taq.nav.FindNodesByType(NodeMethodDecl)
}

// FindErrorNodes finds all ERROR nodes which may indicate parsing issues
func (taq *TypeAnnotationQuery) FindErrorNodes() []*sitter.Node {
	return taq.nav.FindNodesByType(NodeError)
}

// GetVariableDeclarationType extracts the type annotation from a variable declaration
func (taq *TypeAnnotationQuery) GetVariableDeclarationType(varDecl *sitter.Node) string {
	typeExpr := taq.nav.FindDescendantByType(varDecl, NodeTypeExpression)
	if typeExpr == nil {
		return ""
	}

	simpleType := taq.nav.FindDescendantByType(typeExpr, NodeSimpleType)
	if simpleType == nil {
		return taq.nav.GetTextContent(typeExpr)
	}

	return taq.nav.GetTextContent(simpleType)
}

// GetVariableDeclarationName extracts the variable name from a variable declaration
func (taq *TypeAnnotationQuery) GetVariableDeclarationName(varDecl *sitter.Node) string {
	scalar := taq.nav.FindDescendantByType(varDecl, NodeScalar)
	if scalar == nil {
		return ""
	}

	varname := taq.nav.FindDescendantByType(scalar, NodeVarname)
	if varname == nil {
		return ""
	}

	return taq.nav.GetTextContent(varname)
}

// getNodeText extracts text content from a tree-sitter node
func (nav *CSTNavigator) getNodeText(node *sitter.Node) string {
	if node == nil || nav.content == nil {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	if start >= uint(len(nav.content)) || end > uint(len(nav.content)) {
		return ""
	}
	return string(nav.content[start:end])
}
