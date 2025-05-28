// ABOUTME: Tests for core AST types and interfaces
// ABOUTME: Ensures AST node creation and manipulation works correctly

package ast

import (
	"testing"
)

func TestPosition(t *testing.T) {
	pos := Position{Line: 5, Column: 10, Offset: 42}

	if pos.Line != 5 {
		t.Errorf("Expected line 5, got %d", pos.Line)
	}
	if pos.Column != 10 {
		t.Errorf("Expected column 10, got %d", pos.Column)
	}
	if pos.Offset != 42 {
		t.Errorf("Expected offset 42, got %d", pos.Offset)
	}
}

func TestBaseNode(t *testing.T) {
	start := Position{Line: 1, Column: 1, Offset: 0}
	end := Position{Line: 1, Column: 5, Offset: 4}

	node := NewBaseNode("test_node", start, end)

	if node.Type() != "test_node" {
		t.Errorf("Expected type 'test_node', got %q", node.Type())
	}

	if node.Start() != start {
		t.Errorf("Expected start %+v, got %+v", start, node.Start())
	}

	if node.End() != end {
		t.Errorf("Expected end %+v, got %+v", end, node.End())
	}

	if len(node.Children()) != 0 {
		t.Errorf("Expected no children, got %d", len(node.Children()))
	}

	if node.Parent() != nil {
		t.Error("Expected no parent initially")
	}
}

func TestBaseNode_ParentChild(t *testing.T) {
	start := Position{Line: 1, Column: 1, Offset: 0}
	end := Position{Line: 1, Column: 5, Offset: 4}

	parent := NewBaseNode("parent", start, end)
	child := NewBaseNode("child", start, end)

	parent.AddChild(child)

	if len(parent.Children()) != 1 {
		t.Errorf("Expected 1 child, got %d", len(parent.Children()))
	}

	if parent.Children()[0] != child {
		t.Error("Child not properly added to parent")
	}

	if child.Parent() != parent {
		t.Error("Parent not properly set on child")
	}
}

func TestBaseNode_Text(t *testing.T) {
	start := Position{Line: 1, Column: 1, Offset: 0}
	end := Position{Line: 1, Column: 5, Offset: 4}

	node := NewBaseNode("test", start, end)

	if node.Text() != "" {
		t.Errorf("Expected empty text initially, got %q", node.Text())
	}

	node.SetText("test content")

	if node.Text() != "test content" {
		t.Errorf("Expected 'test content', got %q", node.Text())
	}
}

func TestTypeExpression_Simple(t *testing.T) {
	expr := &TypeExpression{
		Name: "Int",
	}

	if !expr.IsSimple() {
		t.Error("Simple type should report as simple")
	}

	if expr.String() != "Int" {
		t.Errorf("Expected 'Int', got %q", expr.String())
	}

	types := expr.GetAllTypes()
	if len(types) != 1 || types[0] != "Int" {
		t.Errorf("Expected ['Int'], got %v", types)
	}
}

func TestTypeExpression_Parameterized(t *testing.T) {
	param := &TypeExpression{Name: "Int"}
	expr := &TypeExpression{
		Name:       "ArrayRef",
		Parameters: []*TypeExpression{param},
	}

	if expr.IsSimple() {
		t.Error("Parameterized type should not report as simple")
	}

	expected := "ArrayRef[Int]"
	if expr.String() != expected {
		t.Errorf("Expected %q, got %q", expected, expr.String())
	}
}

func TestTypeExpression_Union(t *testing.T) {
	type1 := &TypeExpression{Name: "Int"}
	type2 := &TypeExpression{Name: "Str"}

	expr := &TypeExpression{
		IsUnion:    true,
		UnionTypes: []*TypeExpression{type1, type2},
	}

	if expr.IsSimple() {
		t.Error("Union type should not report as simple")
	}

	expected := "Int|Str"
	if expr.String() != expected {
		t.Errorf("Expected %q, got %q", expected, expr.String())
	}

	types := expr.GetAllTypes()
	if len(types) != 2 || types[0] != "Int" || types[1] != "Str" {
		t.Errorf("Expected ['Int', 'Str'], got %v", types)
	}
}

func TestTypeExpression_Intersection(t *testing.T) {
	type1 := &TypeExpression{Name: "Object"}
	type2 := &TypeExpression{Name: "Serializable"}

	expr := &TypeExpression{
		IsIntersection:    true,
		IntersectionTypes: []*TypeExpression{type1, type2},
	}

	if expr.IsSimple() {
		t.Error("Intersection type should not report as simple")
	}

	expected := "Object&Serializable"
	if expr.String() != expected {
		t.Errorf("Expected %q, got %q", expected, expr.String())
	}
}

func TestTypeExpression_Negation(t *testing.T) {
	innerType := &TypeExpression{Name: "Undef"}
	expr := &TypeExpression{
		IsNegation:  true,
		NegatedType: innerType,
	}

	if expr.IsSimple() {
		t.Error("Negation type should not report as simple")
	}

	expected := "!Undef"
	if expr.String() != expected {
		t.Errorf("Expected %q, got %q", expected, expr.String())
	}
}

func TestTypeExpression_OriginalString(t *testing.T) {
	expr := &TypeExpression{
		Name:           "Int",
		OriginalString: "Integer",
	}

	// Original string should take precedence
	if expr.String() != "Integer" {
		t.Errorf("Expected 'Integer', got %q", expr.String())
	}
}

func TestAnnotationKind_String(t *testing.T) {
	tests := []struct {
		kind     AnnotationKind
		expected string
	}{
		{VarAnnotation, "VarAnnotation"},
		{SubParamAnnotation, "SubParamAnnotation"},
		{SubReturnAnnotation, "SubReturnAnnotation"},
		{MethodParamAnnotation, "MethodParamAnnotation"},
		{MethodReturnAnnotation, "MethodReturnAnnotation"},
		{FieldAnnotation, "FieldAnnotation"},
		{TypeDeclAnnotation, "TypeDeclAnnotation"},
		{AnnotationKind(999), "AnnotationKind(999)"},
	}

	for _, test := range tests {
		result := test.kind.String()
		if result != test.expected {
			t.Errorf("AnnotationKind %d: expected %q, got %q",
				int(test.kind), test.expected, result)
		}
	}
}

func TestTypeAnnotation(t *testing.T) {
	typeExpr := &TypeExpression{Name: "Int"}
	pos := Position{Line: 5, Column: 10, Offset: 42}

	annotation := &TypeAnnotation{
		AnnotatedItem:  "$count",
		TypeExpression: typeExpr,
		Pos:            pos,
		Kind:           VarAnnotation,
	}

	if annotation.AnnotatedItem != "$count" {
		t.Errorf("Expected '$count', got %q", annotation.AnnotatedItem)
	}

	if annotation.TypeExpression != typeExpr {
		t.Error("Type expression not properly set")
	}

	if annotation.Pos != pos {
		t.Errorf("Expected position %+v, got %+v", pos, annotation.Pos)
	}

	if annotation.Kind != VarAnnotation {
		t.Errorf("Expected VarAnnotation, got %v", annotation.Kind)
	}
}

func TestAST(t *testing.T) {
	start := Position{Line: 1, Column: 1, Offset: 0}
	end := Position{Line: 10, Column: 1, Offset: 100}
	root := NewBaseNode("root", start, end)

	typeExpr := &TypeExpression{Name: "Int"}
	annotation := &TypeAnnotation{
		AnnotatedItem:  "$x",
		TypeExpression: typeExpr,
		Pos:            Position{Line: 2, Column: 5, Offset: 10},
		Kind:           VarAnnotation,
	}

	ast := &AST{
		Path:            "/test/file.pl",
		Root:            root,
		TypeAnnotations: []*TypeAnnotation{annotation},
		Errors:          []error{},
		Source:          "my Int $x = 42;",
	}

	if ast.Path != "/test/file.pl" {
		t.Errorf("Expected '/test/file.pl', got %q", ast.Path)
	}

	if ast.Root != root {
		t.Error("Root not properly set")
	}

	if len(ast.TypeAnnotations) != 1 {
		t.Errorf("Expected 1 type annotation, got %d", len(ast.TypeAnnotations))
	}

	if ast.TypeAnnotations[0] != annotation {
		t.Error("Type annotation not properly set")
	}

	if len(ast.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(ast.Errors))
	}

	if ast.Source != "my Int $x = 42;" {
		t.Errorf("Expected 'my Int $x = 42;', got %q", ast.Source)
	}
}
