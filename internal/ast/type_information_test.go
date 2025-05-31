// ABOUTME: Tests for enhanced type information functionality in AST nodes
// ABOUTME: Validates type visitor patterns and serialization support

package ast

import (
	"testing"
)

// TestTypeInformationExtraction tests type information extraction from AST
func TestTypeInformationExtraction(t *testing.T) {
	// Create a sample AST with type information
	ast := &AST{
		Path:   "test.pl",
		Source: "my Int $count = 42; method foo(Str $input) -> Bool { return 1; }",
	}

	// Create some typed nodes
	start := Position{Line: 1, Column: 1, Offset: 0}
	end := Position{Line: 1, Column: 10, Offset: 10}

	// Create a typed variable declaration
	intType := NewTypeExpression("Int", nil, start, end)
	varExpr := NewVariableExpr("count", "$", start, end)
	varDecl := NewVarDecl("my", []*VariableExpr{varExpr}, intType, nil, start, end)

	// Create a typed method declaration
	strType := NewTypeExpression("Str", nil, start, end)
	boolType := NewTypeExpression("Bool", nil, start, end)
	param := &Parameter{
		Name:     "input",
		TypeExpr: strType,
		Variable: NewVariableExpr("input", "$", start, end),
		Pos:      start,
	}
	methodDecl := NewSubDecl("foo", []*Parameter{param}, boolType, nil, true, start, end)

	// Add nodes to AST (simplified for testing)
	ast.Root = varDecl
	// In a real scenario, we'd have a proper tree structure

	// Extract type information
	typeInfo := ExtractTypeInformation(ast)

	// Validate results
	if typeInfo == nil {
		t.Fatal("Expected type information, got nil")
	}

	if typeInfo.FilePath != "test.pl" {
		t.Errorf("Expected file path 'test.pl', got '%s'", typeInfo.FilePath)
	}

	// Test visitor functionality
	visitor := &testTypeVisitor{}
	walker := NewTypeWalker(visitor)

	// Create a simple AST structure for testing
	walker.walkNode(varDecl)
	walker.walkNode(methodDecl)

	if visitor.variableCount != 1 {
		t.Errorf("Expected 1 typed variable, got %d", visitor.variableCount)
	}

	if visitor.methodCount != 1 {
		t.Errorf("Expected 1 typed method, got %d", visitor.methodCount)
	}
}

// TestTypeExpressionKinds tests type expression kind assignment
func TestTypeExpressionKinds(t *testing.T) {
	start := Position{Line: 1, Column: 1, Offset: 0}
	end := Position{Line: 1, Column: 10, Offset: 10}

	// Test simple type
	simpleType := NewTypeExpression("Int", nil, start, end)
	if simpleType.Kind != SimpleTypeKind {
		t.Errorf("Expected SimpleTypeKind, got %v", simpleType.Kind)
	}

	// Test parameterized type
	innerType := NewTypeExpression("Int", nil, start, end)
	paramType := NewTypeExpression("ArrayRef", []*TypeExpression{innerType}, start, end)
	if paramType.Kind != ParameterizedTypeKind {
		t.Errorf("Expected ParameterizedTypeKind, got %v", paramType.Kind)
	}
}

// TestVariableTypeInfo tests variable type information extraction
func TestVariableTypeInfo(t *testing.T) {
	start := Position{Line: 1, Column: 1, Offset: 0}
	end := Position{Line: 1, Column: 10, Offset: 10}

	// Create typed variable declaration
	intType := NewTypeExpression("Int", nil, start, end)
	varExpr := NewVariableExpr("count", "$", start, end)
	varDecl := NewVarDecl("my", []*VariableExpr{varExpr}, intType, nil, start, end)

	// Test type info extraction
	typeInfo := varDecl.GetTypeInfo()

	if typeInfo == nil {
		t.Fatal("Expected type info, got nil")
	}

	if typeInfo.DeclType != "my" {
		t.Errorf("Expected declaration type 'my', got '%s'", typeInfo.DeclType)
	}

	if typeInfo.TypeExpr == nil {
		t.Fatal("Expected type expression, got nil")
	}

	if typeInfo.TypeExpr.Name != "Int" {
		t.Errorf("Expected type name 'Int', got '%s'", typeInfo.TypeExpr.Name)
	}

	if len(typeInfo.VariableNames) != 1 {
		t.Errorf("Expected 1 variable name, got %d", len(typeInfo.VariableNames))
	}

	if typeInfo.VariableNames[0] != "$count" {
		t.Errorf("Expected variable name '$count', got '%s'", typeInfo.VariableNames[0])
	}
}

// TestMethodSignature tests method signature generation
func TestMethodSignature(t *testing.T) {
	start := Position{Line: 1, Column: 1, Offset: 0}
	end := Position{Line: 1, Column: 10, Offset: 10}

	// Create typed method declaration
	strType := NewTypeExpression("Str", nil, start, end)
	boolType := NewTypeExpression("Bool", nil, start, end)
	param := &Parameter{
		Name:     "input",
		TypeExpr: strType,
		Variable: NewVariableExpr("input", "$", start, end),
		Pos:      start,
	}
	methodDecl := NewSubDecl("process", []*Parameter{param}, boolType, nil, true, start, end)

	// Test method signature generation
	signature := methodDecl.GetMethodSignature()

	if signature == nil {
		t.Fatal("Expected method signature, got nil")
	}

	if signature.Name != "process" {
		t.Errorf("Expected method name 'process', got '%s'", signature.Name)
	}

	if signature.ReturnType == nil {
		t.Fatal("Expected return type, got nil")
	}

	if signature.ReturnType.Name != "Bool" {
		t.Errorf("Expected return type 'Bool', got '%s'", signature.ReturnType.Name)
	}

	if len(signature.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(signature.Parameters))
	}

	if signature.Parameters[0].Name != "input" {
		t.Errorf("Expected parameter name 'input', got '%s'", signature.Parameters[0].Name)
	}
}

// TestTypeAssertionInfo tests type assertion information
func TestTypeAssertionInfo(t *testing.T) {
	start := Position{Line: 1, Column: 1, Offset: 0}
	end := Position{Line: 1, Column: 10, Offset: 10}

	// Create type assertion expression
	varExpr := NewVariableExpr("value", "$", start, end)
	intType := NewTypeExpression("Int", nil, start, end)
	assertion := NewTypeAssertionExpr(varExpr, intType, start, end)

	// Test type assertion info
	info := assertion.GetTypeAssertionInfo()

	if info == nil {
		t.Fatal("Expected type assertion info, got nil")
	}

	if info.Expression == nil {
		t.Fatal("Expected expression, got nil")
	}

	if info.TargetType == nil {
		t.Fatal("Expected target type, got nil")
	}

	if info.TargetType.Name != "Int" {
		t.Errorf("Expected target type 'Int', got '%s'", info.TargetType.Name)
	}
}

// TestClassDeclaration tests class declaration functionality
func TestClassDeclaration(t *testing.T) {
	start := Position{Line: 1, Column: 1, Offset: 0}
	end := Position{Line: 1, Column: 10, Offset: 10}

	// Create class declaration
	classDecl := NewClassDecl("MyClass", start, end)

	// Add field
	strType := NewTypeExpression("Str", nil, start, end)
	nameVar := NewVariableExpr("name", "$", start, end)
	field := NewFieldDecl("name", strType, nameVar, nil, start, end)
	classDecl.AddField(field)

	// Add method
	boolType := NewTypeExpression("Bool", nil, start, end)
	method := NewMethodDecl("isValid", nil, boolType, nil, start, end)
	classDecl.AddMethod(method)

	// Test class type info
	info := classDecl.GetClassTypeInfo()

	if info == nil {
		t.Fatal("Expected class type info, got nil")
	}

	if info.Name != "MyClass" {
		t.Errorf("Expected class name 'MyClass', got '%s'", info.Name)
	}

	if info.FieldCount != 1 {
		t.Errorf("Expected 1 field, got %d", info.FieldCount)
	}

	if info.MethodCount != 1 {
		t.Errorf("Expected 1 method, got %d", info.MethodCount)
	}

	if !info.IsExported {
		t.Error("Expected class to be exported (starts with uppercase)")
	}
}

// TestFieldTypeInfo tests field type information
func TestFieldTypeInfo(t *testing.T) {
	start := Position{Line: 1, Column: 1, Offset: 0}
	end := Position{Line: 1, Column: 10, Offset: 10}

	// Create field declaration
	strType := NewTypeExpression("Str", nil, start, end)
	nameVar := NewVariableExpr("name", "$", start, end)
	field := NewFieldDecl("name", strType, nameVar, nil, start, end)

	// Test field type info
	info := field.GetFieldTypeInfo()

	if info == nil {
		t.Fatal("Expected field type info, got nil")
	}

	if info.Name != "name" {
		t.Errorf("Expected field name 'name', got '%s'", info.Name)
	}

	if info.TypeExpr == nil {
		t.Fatal("Expected type expression, got nil")
	}

	if info.TypeExpr.Name != "Str" {
		t.Errorf("Expected type 'Str', got '%s'", info.TypeExpr.Name)
	}

	if info.AccessLevel != "public" {
		t.Errorf("Expected access level 'public', got '%s'", info.AccessLevel)
	}
}

// testTypeVisitor implements TypeVisitor for testing
type testTypeVisitor struct {
	variableCount  int
	methodCount    int
	fieldCount     int
	assertionCount int
	typeExprCount  int
	typeDeclCount  int
}

func (tv *testTypeVisitor) VisitTypeExpression(node *TypeExpression) error {
	tv.typeExprCount++
	return nil
}

func (tv *testTypeVisitor) VisitTypedVariable(node *VarDecl) error {
	tv.variableCount++
	return nil
}

func (tv *testTypeVisitor) VisitTypedMethod(node *SubDecl) error {
	tv.methodCount++
	return nil
}

func (tv *testTypeVisitor) VisitTypeAssertion(node *TypeAssertionExpr) error {
	tv.assertionCount++
	return nil
}

func (tv *testTypeVisitor) VisitFieldDeclaration(node *FieldDecl) error {
	tv.fieldCount++
	return nil
}

func (tv *testTypeVisitor) VisitTypeDeclaration(node *TypeDecl) error {
	tv.typeDeclCount++
	return nil
}
