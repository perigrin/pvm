// ABOUTME: Visitor pattern implementation for AST traversal
// ABOUTME: Provides type-safe visitor patterns for different AST node types

package astnav

import (
	"tamarou.com/pvm/internal/ast"
)

// Visitor provides a type-safe visitor pattern for AST nodes
type Visitor interface {
	// Expression visitors
	VisitLiteralExpr(expr *ast.LiteralExpr) interface{}
	VisitVariableExpr(expr *ast.VariableExpr) interface{}
	VisitBinaryExpr(expr *ast.BinaryExpr) interface{}
	VisitUnaryExpr(expr *ast.UnaryExpr) interface{}
	VisitCallExpr(expr *ast.CallExpr) interface{}
	VisitArrayRefExpr(expr *ast.ArrayRefExpr) interface{}
	VisitHashRefExpr(expr *ast.HashRefExpr) interface{}
	VisitTypeAssertionExpr(expr *ast.TypeAssertionExpr) interface{}
	VisitConditionalExpr(expr *ast.ConditionalExpr) interface{}
	VisitListExpr(expr *ast.ListExpr) interface{}

	// Statement visitors
	VisitExpressionStmt(stmt *ast.ExpressionStmt) interface{}
	VisitVarDecl(stmt *ast.VarDecl) interface{}
	VisitSubDecl(stmt *ast.SubDecl) interface{}
	VisitMethodDecl(stmt *ast.MethodDecl) interface{}
	VisitFieldDecl(stmt *ast.FieldDecl) interface{}
	VisitTypeDecl(stmt *ast.TypeDecl) interface{}
	VisitBlockStmt(stmt *ast.BlockStmt) interface{}
	VisitIfStmt(stmt *ast.IfStmt) interface{}
	VisitWhileStmt(stmt *ast.WhileStmt) interface{}
	VisitForStmt(stmt *ast.ForStmt) interface{}
	VisitReturnStmt(stmt *ast.ReturnStmt) interface{}
	VisitUseStmt(stmt *ast.UseStmt) interface{}
	VisitPackageStmt(stmt *ast.PackageStmt) interface{}

	// Generic node visitor (fallback)
	VisitNode(node ast.Node) interface{}
}

// BaseVisitor provides a default implementation of the Visitor interface
// Users can embed this and override only the methods they need
type BaseVisitor struct{}

// Expression visitor implementations (default: return nil)
func (bv *BaseVisitor) VisitLiteralExpr(expr *ast.LiteralExpr) interface{}             { return nil }
func (bv *BaseVisitor) VisitVariableExpr(expr *ast.VariableExpr) interface{}           { return nil }
func (bv *BaseVisitor) VisitBinaryExpr(expr *ast.BinaryExpr) interface{}               { return nil }
func (bv *BaseVisitor) VisitUnaryExpr(expr *ast.UnaryExpr) interface{}                 { return nil }
func (bv *BaseVisitor) VisitCallExpr(expr *ast.CallExpr) interface{}                   { return nil }
func (bv *BaseVisitor) VisitArrayRefExpr(expr *ast.ArrayRefExpr) interface{}           { return nil }
func (bv *BaseVisitor) VisitHashRefExpr(expr *ast.HashRefExpr) interface{}             { return nil }
func (bv *BaseVisitor) VisitTypeAssertionExpr(expr *ast.TypeAssertionExpr) interface{} { return nil }
func (bv *BaseVisitor) VisitConditionalExpr(expr *ast.ConditionalExpr) interface{}     { return nil }
func (bv *BaseVisitor) VisitListExpr(expr *ast.ListExpr) interface{}                   { return nil }

// Statement visitor implementations (default: return nil)
func (bv *BaseVisitor) VisitExpressionStmt(stmt *ast.ExpressionStmt) interface{} { return nil }
func (bv *BaseVisitor) VisitVarDecl(stmt *ast.VarDecl) interface{}               { return nil }
func (bv *BaseVisitor) VisitSubDecl(stmt *ast.SubDecl) interface{}               { return nil }
func (bv *BaseVisitor) VisitMethodDecl(stmt *ast.MethodDecl) interface{}         { return nil }
func (bv *BaseVisitor) VisitFieldDecl(stmt *ast.FieldDecl) interface{}           { return nil }
func (bv *BaseVisitor) VisitTypeDecl(stmt *ast.TypeDecl) interface{}             { return nil }
func (bv *BaseVisitor) VisitBlockStmt(stmt *ast.BlockStmt) interface{}           { return nil }
func (bv *BaseVisitor) VisitIfStmt(stmt *ast.IfStmt) interface{}                 { return nil }
func (bv *BaseVisitor) VisitWhileStmt(stmt *ast.WhileStmt) interface{}           { return nil }
func (bv *BaseVisitor) VisitForStmt(stmt *ast.ForStmt) interface{}               { return nil }
func (bv *BaseVisitor) VisitReturnStmt(stmt *ast.ReturnStmt) interface{}         { return nil }
func (bv *BaseVisitor) VisitUseStmt(stmt *ast.UseStmt) interface{}               { return nil }
func (bv *BaseVisitor) VisitPackageStmt(stmt *ast.PackageStmt) interface{}       { return nil }

// Generic node visitor (default: return nil)
func (bv *BaseVisitor) VisitNode(node ast.Node) interface{} { return nil }

// Visit dispatches to the appropriate visitor method based on node type
func Visit(visitor Visitor, node ast.Node) interface{} {
	if node == nil {
		return nil
	}

	// Type switch to dispatch to appropriate visitor method
	switch n := node.(type) {
	// Expression types
	case *ast.LiteralExpr:
		return visitor.VisitLiteralExpr(n)
	case *ast.VariableExpr:
		return visitor.VisitVariableExpr(n)
	case *ast.BinaryExpr:
		return visitor.VisitBinaryExpr(n)
	case *ast.UnaryExpr:
		return visitor.VisitUnaryExpr(n)
	case *ast.CallExpr:
		return visitor.VisitCallExpr(n)
	case *ast.ArrayRefExpr:
		return visitor.VisitArrayRefExpr(n)
	case *ast.HashRefExpr:
		return visitor.VisitHashRefExpr(n)
	case *ast.TypeAssertionExpr:
		return visitor.VisitTypeAssertionExpr(n)
	case *ast.ConditionalExpr:
		return visitor.VisitConditionalExpr(n)
	case *ast.ListExpr:
		return visitor.VisitListExpr(n)

	// Statement types
	case *ast.ExpressionStmt:
		return visitor.VisitExpressionStmt(n)
	case *ast.VarDecl:
		return visitor.VisitVarDecl(n)
	case *ast.SubDecl:
		return visitor.VisitSubDecl(n)
	case *ast.MethodDecl:
		return visitor.VisitMethodDecl(n)
	case *ast.FieldDecl:
		return visitor.VisitFieldDecl(n)
	case *ast.TypeDecl:
		return visitor.VisitTypeDecl(n)
	case *ast.BlockStmt:
		return visitor.VisitBlockStmt(n)
	case *ast.IfStmt:
		return visitor.VisitIfStmt(n)
	case *ast.WhileStmt:
		return visitor.VisitWhileStmt(n)
	case *ast.ForStmt:
		return visitor.VisitForStmt(n)
	case *ast.ReturnStmt:
		return visitor.VisitReturnStmt(n)
	case *ast.UseStmt:
		return visitor.VisitUseStmt(n)
	case *ast.PackageStmt:
		return visitor.VisitPackageStmt(n)

	// Fallback to generic node visitor
	default:
		return visitor.VisitNode(node)
	}
}

// WalkVisitor combines visitor pattern with tree walking
func WalkVisitor(visitor Visitor, node ast.Node) {
	if node == nil {
		return
	}

	// Visit current node
	Visit(visitor, node)

	// Recursively visit children
	for _, child := range node.Children() {
		WalkVisitor(visitor, child)
	}
}

// CollectVisitor is a visitor that collects nodes of specific types
type CollectVisitor struct {
	*BaseVisitor
	Variables    []*ast.VariableExpr
	Functions    []*ast.SubDecl
	Types        []*ast.TypeDecl
	Declarations []ast.StatementNode
}

// NewCollectVisitor creates a new collector visitor
func NewCollectVisitor() *CollectVisitor {
	return &CollectVisitor{
		BaseVisitor:  &BaseVisitor{},
		Variables:    make([]*ast.VariableExpr, 0),
		Functions:    make([]*ast.SubDecl, 0),
		Types:        make([]*ast.TypeDecl, 0),
		Declarations: make([]ast.StatementNode, 0),
	}
}

// VisitVariableExpr collects variable expressions
func (cv *CollectVisitor) VisitVariableExpr(expr *ast.VariableExpr) interface{} {
	cv.Variables = append(cv.Variables, expr)
	return nil
}

// VisitSubDecl collects function declarations
func (cv *CollectVisitor) VisitSubDecl(stmt *ast.SubDecl) interface{} {
	cv.Functions = append(cv.Functions, stmt)
	cv.Declarations = append(cv.Declarations, stmt)
	return nil
}

// VisitMethodDecl collects method declarations
func (cv *CollectVisitor) VisitMethodDecl(stmt *ast.MethodDecl) interface{} {
	cv.Functions = append(cv.Functions, stmt.SubDecl)
	cv.Declarations = append(cv.Declarations, stmt)
	return nil
}

// VisitTypeDecl collects type declarations
func (cv *CollectVisitor) VisitTypeDecl(stmt *ast.TypeDecl) interface{} {
	cv.Types = append(cv.Types, stmt)
	cv.Declarations = append(cv.Declarations, stmt)
	return nil
}

// VisitVarDecl collects variable declarations
func (cv *CollectVisitor) VisitVarDecl(stmt *ast.VarDecl) interface{} {
	cv.Declarations = append(cv.Declarations, stmt)
	return nil
}

// VisitFieldDecl collects field declarations
func (cv *CollectVisitor) VisitFieldDecl(stmt *ast.FieldDecl) interface{} {
	cv.Declarations = append(cv.Declarations, stmt)
	return nil
}

// TransformVisitor is a visitor that can transform the AST
type TransformVisitor struct {
	*BaseVisitor
	transformations map[ast.Node]ast.Node
}

// NewTransformVisitor creates a new transform visitor
func NewTransformVisitor() *TransformVisitor {
	return &TransformVisitor{
		BaseVisitor:     &BaseVisitor{},
		transformations: make(map[ast.Node]ast.Node),
	}
}

// ReplaceNode records a node replacement
func (tv *TransformVisitor) ReplaceNode(old, new ast.Node) {
	tv.transformations[old] = new
}

// GetTransformations returns all recorded transformations
func (tv *TransformVisitor) GetTransformations() map[ast.Node]ast.Node {
	return tv.transformations
}

// ApplyTransformations applies all recorded transformations to the tree
// This would typically be implemented with a separate tree reconstruction pass
func (tv *TransformVisitor) ApplyTransformations(root ast.Node) ast.Node {
	// This is a placeholder - actual implementation would require
	// a full tree reconstruction algorithm
	return root
}

// PrintVisitor is a visitor that prints the AST structure
type PrintVisitor struct {
	*BaseVisitor
	indent int
	output []string
}

// NewPrintVisitor creates a new print visitor
func NewPrintVisitor() *PrintVisitor {
	return &PrintVisitor{
		BaseVisitor: &BaseVisitor{},
		indent:      0,
		output:      make([]string, 0),
	}
}

// VisitNode prints generic node information
func (pv *PrintVisitor) VisitNode(node ast.Node) interface{} {
	indentStr := ""
	for i := 0; i < pv.indent; i++ {
		indentStr += "  "
	}

	line := indentStr + node.Type()
	if text := node.Text(); text != "" && len(text) < 50 {
		line += ": " + text
	}

	pv.output = append(pv.output, line)
	return nil
}

// GetOutput returns the collected output lines
func (pv *PrintVisitor) GetOutput() []string {
	return pv.output
}

// WalkPrint performs a walk with printing, handling indentation
func WalkPrint(visitor *PrintVisitor, node ast.Node) {
	if node == nil {
		return
	}

	// Visit current node
	Visit(visitor, node)

	// Increase indent for children
	visitor.indent++

	// Recursively visit children
	for _, child := range node.Children() {
		WalkPrint(visitor, child)
	}

	// Decrease indent when done with children
	visitor.indent--
}
