// ABOUTME: AST navigation utilities following TypeScript-Go astnav patterns
// ABOUTME: Provides efficient traversal, search, and manipulation of AST trees

package astnav

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
)

// Navigator provides utilities for navigating and manipulating AST trees
type Navigator struct {
	root ast.Node
}

// NewNavigator creates a new AST navigator for the given root node
func NewNavigator(root ast.Node) *Navigator {
	return &Navigator{root: root}
}

// FindNodeAt finds the deepest node that contains the given position
func (n *Navigator) FindNodeAt(line, column int) ast.Node {
	pos := ast.Position{Line: line, Column: column}
	return n.findNodeAtPosition(n.root, pos)
}

// findNodeAtPosition recursively searches for the node containing the position
func (n *Navigator) findNodeAtPosition(node ast.Node, pos ast.Position) ast.Node {
	if node == nil {
		return nil
	}

	// Check if position is within this node
	start := node.Start()
	end := node.End()

	if !n.positionInRange(pos, start, end) {
		return nil
	}

	// Check children first (deepest match wins)
	for _, child := range node.Children() {
		if found := n.findNodeAtPosition(child, pos); found != nil {
			return found
		}
	}

	// If no child contains the position, this node is the deepest match
	return node
}

// positionInRange checks if a position falls within a range
func (n *Navigator) positionInRange(pos, start, end ast.Position) bool {
	if pos.Line < start.Line || pos.Line > end.Line {
		return false
	}

	if pos.Line == start.Line && pos.Column < start.Column {
		return false
	}

	if pos.Line == end.Line && pos.Column > end.Column {
		return false
	}

	return true
}

// FindNodesByType finds all nodes of the specified type
func (n *Navigator) FindNodesByType(nodeType string) []ast.Node {
	var nodes []ast.Node
	n.Walk(n.root, func(node ast.Node) bool {
		if node.Type() == nodeType {
			nodes = append(nodes, node)
		}
		return true // Continue traversal
	})
	return nodes
}

// FindVariables finds all variable references
func (n *Navigator) FindVariables() []*ast.VariableExpr {
	var variables []*ast.VariableExpr
	n.Walk(n.root, func(node ast.Node) bool {
		if varExpr, ok := node.(*ast.VariableExpr); ok {
			variables = append(variables, varExpr)
		}
		return true
	})
	return variables
}

// FindVariableByName finds all references to a specific variable
func (n *Navigator) FindVariableByName(name string) []*ast.VariableExpr {
	var variables []*ast.VariableExpr
	n.Walk(n.root, func(node ast.Node) bool {
		if varExpr, ok := node.(*ast.VariableExpr); ok {
			if varExpr.Name == name || varExpr.FullName() == name {
				variables = append(variables, varExpr)
			}
		}
		return true
	})
	return variables
}

// FindTypeAnnotations finds all type annotations in the tree
func (n *Navigator) FindTypeAnnotations() []*ast.TypeAnnotation {
	var annotations []*ast.TypeAnnotation

	// Check if root is an AST with TypeAnnotations
	if rootAST, ok := n.root.(*ast.AST); ok {
		return rootAST.TypeAnnotations
	}

	// Otherwise, search for typed declarations
	n.Walk(n.root, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.VarDecl:
			if typed.IsTyped() {
				// Create type annotation from declaration
				for _, variable := range typed.LogicalVariables() {
					annotation := &ast.TypeAnnotation{
						AnnotatedItem:  variable.FullName(),
						TypeExpression: typed.TypeExpr,
						Pos:            typed.Start(),
						Kind:           ast.VarAnnotation,
					}
					annotations = append(annotations, annotation)
				}
			}
		case *ast.SubDecl:
			if typed.IsTyped() {
				// Add parameter annotations
				for _, param := range typed.LogicalParameters() {
					if param.TypeExpr != nil {
						annotation := &ast.TypeAnnotation{
							AnnotatedItem:  param.Name,
							TypeExpression: param.TypeExpr,
							Pos:            param.Pos,
							Kind:           ast.SubParamAnnotation,
						}
						annotations = append(annotations, annotation)
					}
				}
				// Add return type annotation
				if typed.ReturnType != nil {
					annotation := &ast.TypeAnnotation{
						AnnotatedItem:  typed.Name,
						TypeExpression: typed.ReturnType,
						Pos:            typed.Start(),
						Kind:           ast.SubReturnAnnotation,
					}
					annotations = append(annotations, annotation)
				}
			}
		case *ast.FieldDecl:
			if typed.TypeExpr != nil {
				annotation := &ast.TypeAnnotation{
					AnnotatedItem:  typed.Variable.FullName(),
					TypeExpression: typed.TypeExpr,
					Pos:            typed.Start(),
					Kind:           ast.FieldAnnotation,
				}
				annotations = append(annotations, annotation)
			}
		}
		return true
	})

	return annotations
}

// FindFunctions finds all function/subroutine declarations
func (n *Navigator) FindFunctions() []*ast.SubDecl {
	var functions []*ast.SubDecl
	n.Walk(n.root, func(node ast.Node) bool {
		if subDecl, ok := node.(*ast.SubDecl); ok {
			functions = append(functions, subDecl)
		}
		if methodDecl, ok := node.(*ast.MethodDecl); ok {
			functions = append(functions, methodDecl.SubDecl)
		}
		return true
	})
	return functions
}

// GetParent returns the parent of the given node
func (n *Navigator) GetParent(node ast.Node) ast.Node {
	if node == nil {
		return nil
	}
	return node.Parent()
}

// GetAncestors returns all ancestors of the given node (from parent to root)
func (n *Navigator) GetAncestors(node ast.Node) []ast.Node {
	var ancestors []ast.Node
	current := node.Parent()

	for current != nil {
		ancestors = append(ancestors, current)
		current = current.Parent()
	}

	return ancestors
}

// GetSiblings returns all sibling nodes of the given node
func (n *Navigator) GetSiblings(node ast.Node) []ast.Node {
	parent := node.Parent()
	if parent == nil {
		return []ast.Node{}
	}

	var siblings []ast.Node
	for _, child := range parent.Children() {
		if child != node {
			siblings = append(siblings, child)
		}
	}

	return siblings
}

// Walk performs a depth-first traversal of the AST
func (n *Navigator) Walk(node ast.Node, visitor ast.NodeVisitor) {
	if node == nil || visitor == nil {
		return
	}

	// Visit current node
	if !visitor(node) {
		return // Visitor returned false, stop traversal
	}

	// Visit children
	for _, child := range node.Children() {
		n.Walk(child, visitor)
	}
}

// WalkWithContext performs a depth-first traversal with enter/exit hooks
func (n *Navigator) WalkWithContext(node ast.Node, walkFunc ast.WalkFunc) {
	if node == nil {
		return
	}

	// Enter the node
	shouldVisitChildren := true
	if walkFunc.Enter != nil {
		shouldVisitChildren = walkFunc.Enter(node)
	}

	// Visit children if allowed
	if shouldVisitChildren {
		for _, child := range node.Children() {
			n.WalkWithContext(child, walkFunc)
		}
	}

	// Exit the node
	if walkFunc.Exit != nil {
		walkFunc.Exit(node)
	}
}

// GetNodePath returns a path description from root to the given node
func (n *Navigator) GetNodePath(node ast.Node) string {
	if node == nil {
		return ""
	}

	var path []string
	current := node

	// Build path from node to root
	for current != nil {
		nodeDesc := current.Type()

		// Add position for disambiguation
		start := current.Start()
		nodeDesc += fmt.Sprintf("@%d:%d", start.Line, start.Column)

		path = append([]string{nodeDesc}, path...) // Prepend to reverse order
		current = current.Parent()
	}

	return strings.Join(path, " > ")
}

// CountNodes returns the total number of nodes in the tree
func (n *Navigator) CountNodes() int {
	count := 0
	n.Walk(n.root, func(node ast.Node) bool {
		count++
		return true
	})
	return count
}

// GetDepth returns the maximum depth of the tree
func (n *Navigator) GetDepth() int {
	maxDepth := 0

	var calculateDepth func(ast.Node, int)
	calculateDepth = func(node ast.Node, currentDepth int) {
		if currentDepth > maxDepth {
			maxDepth = currentDepth
		}

		for _, child := range node.Children() {
			calculateDepth(child, currentDepth+1)
		}
	}

	calculateDepth(n.root, 0)
	return maxDepth
}

// FindNodeByText finds nodes containing specific text
func (n *Navigator) FindNodeByText(text string) []ast.Node {
	var nodes []ast.Node
	n.Walk(n.root, func(node ast.Node) bool {
		if strings.Contains(node.Text(), text) {
			nodes = append(nodes, node)
		}
		return true
	})
	return nodes
}

// IsDescendantOf checks if child is a descendant of ancestor
func (n *Navigator) IsDescendantOf(child, ancestor ast.Node) bool {
	current := child.Parent()

	for current != nil {
		if current == ancestor {
			return true
		}
		current = current.Parent()
	}

	return false
}

// FindCommonAncestor finds the lowest common ancestor of two nodes
func (n *Navigator) FindCommonAncestor(node1, node2 ast.Node) ast.Node {
	if node1 == nil || node2 == nil {
		return nil
	}

	// Get all ancestors of node1
	ancestors1 := make(map[ast.Node]bool)
	current := node1
	for current != nil {
		ancestors1[current] = true
		current = current.Parent()
	}

	// Find first common ancestor in node2's ancestry
	current = node2
	for current != nil {
		if ancestors1[current] {
			return current
		}
		current = current.Parent()
	}

	return nil
}
