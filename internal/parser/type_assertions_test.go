// ABOUTME: Tests for type assertion and constraint parsing
// ABOUTME: Validates parser correctly handles 'as' keyword and 'where' clauses

package parser

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
)

func TestBasicTypeAssertions(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	testCases := []struct {
		name     string
		input    string
		expected string // Expected type assertion structure
	}{
		{
			name:     "simple_int_assertion",
			input:    "my $number = $input as Int;",
			expected: "type_assertion(variable($input) -> Int)",
		},
		{
			name:     "string_assertion",
			input:    "my $text = $data as Str;",
			expected: "type_assertion(variable($data) -> Str)",
		},
		{
			name:     "custom_type_assertion",
			input:    "my $obj = $value as MyClass;",
			expected: "type_assertion(variable($value) -> MyClass)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseString(tc.input)
			if err != nil {
				t.Errorf("Parse error for %s: %v", tc.name, err)
				return
			}

			if result == nil {
				t.Errorf("Expected AST result for %s", tc.name)
				return
			}

			// For now, just verify we get a result without errors
			// TODO: Add more specific AST structure validation
			t.Logf("Parsed %s successfully: %s", tc.name, result.String())
		})
	}
}

func TestExpressionContextAssertions(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "parenthesized_expression",
			input: "my $result = ($a + $b) as Num;",
		},
		{
			name:  "array_access",
			input: "my $item = $array->[$index] as ItemType;",
		},
		{
			name:  "chained_with_default",
			input: "my $value = $maybe as Int // 0;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseString(tc.input)
			if err != nil {
				t.Errorf("Parse error for %s: %v", tc.name, err)
				return
			}

			if result == nil {
				t.Errorf("Expected AST result for %s", tc.name)
				return
			}

			t.Logf("Parsed %s successfully", tc.name)
		})
	}
}

func TestComplexTypeAssertions(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "parameterized_type",
			input: "my $array = $input as ArrayRef[Str];",
		},
		{
			name:  "union_type",
			input: "my $union = $data as Int|Str;",
		},
		{
			name:  "nested_parameterized",
			input: "my $complex = $data as ArrayRef[HashRef[Int]];",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseString(tc.input)
			if err != nil {
				t.Errorf("Parse error for %s: %v", tc.name, err)
				return
			}

			if result == nil {
				t.Errorf("Expected AST result for %s", tc.name)
				return
			}

			t.Logf("Parsed %s successfully", tc.name)
		})
	}
}

func TestTypeConstraints(t *testing.T) {
	t.Skip("Type constraint syntax (where clauses) not yet implemented - skipping until grammar supports constraint parsing")

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "simple_constraint",
			input: "my $positive = $num as (Int where { $_ > 0 });",
		},
		{
			name:  "range_constraint",
			input: "my $range = $val as (Num where { $_ >= 0 && $_ <= 100 });",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseString(tc.input)
			if err != nil {
				t.Errorf("Parse error for %s: %v", tc.name, err)
				return
			}

			if result == nil {
				t.Errorf("Expected AST result for %s", tc.name)
				return
			}

			t.Logf("Parsed %s successfully", tc.name)
		})
	}
}

func TestTypeAssertionAST(t *testing.T) {
	// Test AST node creation directly
	start := ast.Position{Line: 1, Column: 1, Offset: 0}
	end := ast.Position{Line: 1, Column: 20, Offset: 20}

	// Create a simple variable expression
	varExpr := ast.NewVariableExpr("input", "$", start, end)

	// Create a simple type expression
	typeExpr := ast.NewTypeExpression("Int", nil, start, end)

	// Create a type assertion
	assertion := ast.NewTypeAssertionExpr(varExpr, typeExpr, start, end)

	if assertion == nil {
		t.Fatal("Failed to create type assertion")
	}

	if !assertion.IsExpression() {
		t.Error("Type assertion should be an expression")
	}

	if assertion.Expression != varExpr {
		t.Error("Type assertion expression not set correctly")
	}

	if assertion.TargetType != typeExpr {
		t.Error("Type assertion target type not set correctly")
	}

	t.Logf("Type assertion AST created successfully: %s", assertion.Type())
}

func TestWhereClauseAST(t *testing.T) {
	// Test where clause AST creation
	start := ast.Position{Line: 1, Column: 1, Offset: 0}
	end := ast.Position{Line: 1, Column: 10, Offset: 10}

	// Create a simple comparison expression for the constraint
	varExpr := ast.NewVariableExpr("_", "$", start, end)
	numExpr := ast.NewLiteralExpr("0", ast.NumberLiteral, start, end)
	comparison := ast.NewBinaryExpr(varExpr, numExpr, ">", start, end)

	// Create where clause
	whereClause := ast.NewWhereClause(comparison, start, end)

	if whereClause == nil {
		t.Fatal("Failed to create where clause")
	}

	if whereClause.Expression != comparison {
		t.Error("Where clause expression not set correctly")
	}

	t.Logf("Where clause AST created successfully: %s", whereClause.Type())
}

func TestConstrainedTypeAST(t *testing.T) {
	// Test constrained type AST creation
	start := ast.Position{Line: 1, Column: 1, Offset: 0}
	end := ast.Position{Line: 1, Column: 20, Offset: 20}

	// Create base type
	baseType := ast.NewTypeExpression("Int", nil, start, end)

	// Create constraint expression
	varExpr := ast.NewVariableExpr("_", "$", start, end)
	numExpr := ast.NewLiteralExpr("0", ast.NumberLiteral, start, end)
	comparison := ast.NewBinaryExpr(varExpr, numExpr, ">", start, end)
	whereClause := ast.NewWhereClause(comparison, start, end)

	// Create constrained type
	constrainedType := ast.NewConstrainedType(baseType, whereClause, start, end)

	if constrainedType == nil {
		t.Fatal("Failed to create constrained type")
	}

	if constrainedType.BaseType != baseType {
		t.Error("Constrained type base type not set correctly")
	}

	if constrainedType.Constraint != whereClause {
		t.Error("Constrained type constraint not set correctly")
	}

	expectedStr := "Int where { ... }"
	if constrainedType.String() != expectedStr {
		t.Errorf("Expected constrained type string '%s', got '%s'", expectedStr, constrainedType.String())
	}

	t.Logf("Constrained type AST created successfully: %s", constrainedType.String())
}
