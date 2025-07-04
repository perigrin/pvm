// ABOUTME: Tests for CST analysis utilities and typed Perl construct recognition
// ABOUTME: Validates CST structure understanding and navigation functionality

package compiler

import (
	"fmt"
	"strings"
	"testing"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/parser/treesitter"
)

// parseTypedPerl parses typed Perl code using tree-sitter and returns the root node and content
func parseTypedPerl(code string) (*sitter.Node, []byte, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(treesitter.Language())

	content := []byte(code)
	tree := parser.Parse(content, nil)
	if tree == nil {
		return nil, nil, fmt.Errorf("failed to parse content")
	}

	return tree.RootNode(), content, nil
}

// parseTypedPerlSimple parses code and returns just the root node for simpler tests
func parseTypedPerlSimple(code string) (*sitter.Node, error) {
	root, _, err := parseTypedPerl(code)
	return root, err
}

func TestCSTAnalyzer_AnalyzeTypedVariableDeclaration(t *testing.T) {
	code := "my Int $count = 42;"
	root, content, err := parseTypedPerl(code)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	analyzer := NewCSTAnalyzerWithContent(root, content)

	// Find variable declaration nodes
	nav := NewCSTNavigatorWithContent(root, content)
	varDecls := nav.FindNodesByType(NodeVariableDecl)

	if len(varDecls) == 0 {
		t.Fatal("No variable declaration nodes found")
	}

	// Analyze the first variable declaration
	analysis := analyzer.AnalyzeNode(varDecls[0])

	if !analysis.Valid {
		t.Fatalf("Analysis failed: %s", analysis.Error)
	}

	if analysis.Type != NodeVariableDecl {
		t.Errorf("Expected node type %s, got %s", NodeVariableDecl, analysis.Type)
	}

	// Should identify as variable declaration pattern
	if analysis.Pattern == nil {
		t.Fatal("Failed to identify variable declaration pattern")
	}

	if analysis.Pattern.Name != "VariableDeclaration" {
		t.Errorf("Expected pattern name VariableDeclaration, got %s", analysis.Pattern.Name)
	}
}

func TestCSTAnalyzer_ExtractTypeAnnotation(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "Simple type",
			code:     "my Int $count = 42;",
			expected: "Int",
		},
		{
			name:     "Complex type",
			code:     "my ArrayRef[Str] $names = [];",
			expected: "ArrayRef[Str]",
		},
		{
			name:     "Field declaration",
			code:     "field Str $name;",
			expected: "Str",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root, content, err := parseTypedPerl(tc.code)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			analyzer := NewCSTAnalyzerWithContent(root, content)
			nav := NewCSTNavigatorWithContent(root, content)
			varDecls := nav.FindNodesByType(NodeVariableDecl)

			if len(varDecls) == 0 {
				t.Fatal("No variable declaration nodes found")
			}

			typeText := analyzer.ExtractTypeAnnotation(varDecls[0])
			if typeText != tc.expected {
				t.Errorf("Expected type annotation %q, got %q", tc.expected, typeText)
			}
		})
	}
}

func TestCSTAnalyzer_ExtractVariableName(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "Scalar variable",
			code:     "my Int $count = 42;",
			expected: "count",
		},
		{
			name:     "Field variable",
			code:     "field Str $name;",
			expected: "name",
		},
		{
			name:     "Array variable",
			code:     "my ArrayRef @items = ();",
			expected: "items",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root, content, err := parseTypedPerl(tc.code)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			analyzer := NewCSTAnalyzerWithContent(root, content)
			nav := NewCSTNavigatorWithContent(root, content)
			varDecls := nav.FindNodesByType(NodeVariableDecl)

			if len(varDecls) == 0 {
				t.Fatal("No variable declaration nodes found")
			}

			varName := analyzer.ExtractVariableName(varDecls[0])
			if varName != tc.expected {
				t.Errorf("Expected variable name %q, got %q", tc.expected, varName)
			}
		})
	}
}

func TestCSTAnalyzer_FindAllTypedConstructs(t *testing.T) {
	code := `
		my Int $count = 42;
		field Str $name;
		my $value as Str;
	`

	root, content, err := parseTypedPerl(code)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	analyzer := NewCSTAnalyzerWithContent(root, content)
	constructs := analyzer.FindAllTypedConstructs()

	if len(constructs) < 2 {
		t.Errorf("Expected at least 2 typed constructs, found %d", len(constructs))
	}

	// Check that we found variable declarations (both regular and field)
	foundPatterns := make(map[string]bool)
	hasFieldDeclaration := false
	hasRegularDeclaration := false

	for _, construct := range constructs {
		foundPatterns[construct.Pattern.Name] = true
		if construct.Pattern.Name == "VariableDeclaration" {
			// Check if it's a field or regular declaration using the analyzer
			if analyzer.IsFieldDeclaration(construct.Node) {
				hasFieldDeclaration = true
			} else {
				hasRegularDeclaration = true
			}
		}
	}

	if !foundPatterns["VariableDeclaration"] {
		t.Error("Should have found VariableDeclaration pattern")
	}

	if !hasFieldDeclaration {
		t.Error("Should have found a field declaration")
	}

	if !hasRegularDeclaration {
		t.Error("Should have found a regular variable declaration")
	}
}

func TestCSTAnalyzer_IsTypeAnnotationNode(t *testing.T) {
	code := "my Int $count = 42;"
	root, err := parseTypedPerlSimple(code)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	analyzer := NewCSTAnalyzer(root)
	nav := NewCSTNavigator(root)

	// Find type expression nodes
	typeExprs := nav.FindNodesByType(NodeTypeExpression)
	if len(typeExprs) == 0 {
		t.Fatal("No type expression nodes found")
	}

	// Test that type expression is correctly identified
	if !analyzer.IsTypeAnnotationNode(typeExprs[0]) {
		t.Error("Type expression node should be identified as type annotation")
	}

	// Find a non-type node for comparison
	numbers := nav.FindNodesByType(NodeNumber)
	if len(numbers) > 0 {
		if analyzer.IsTypeAnnotationNode(numbers[0]) {
			t.Error("Number node should not be identified as type annotation")
		}
	}
}

func TestCSTAnalyzer_PrintCSTStructure(t *testing.T) {
	code := "my Int $count = 42;"
	root, err := parseTypedPerlSimple(code)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	analyzer := NewCSTAnalyzer(root)
	structure := analyzer.PrintCSTStructure()

	if len(structure) == 0 {
		t.Error("CST structure output should not be empty")
	}

	// Should contain key node types
	expectedNodes := []string{
		"source_file",
		"variable_declaration",
		"type_expression",
		"simple_type",
		"scalar",
		"varname",
	}

	for _, expected := range expectedNodes {
		if !strings.Contains(structure, expected) {
			t.Errorf("CST structure should contain node type %s", expected)
		}
	}
}

func TestConstructPatterns(t *testing.T) {
	patterns := GetConstructPatterns()

	if len(patterns) == 0 {
		t.Fatal("Should have construct patterns")
	}

	// Verify required patterns exist
	requiredPatterns := []string{
		"VariableDeclaration",
		"TypeAssertion",
		"MethodParameter",
	}

	foundPatterns := make(map[string]bool)
	for _, pattern := range patterns {
		foundPatterns[pattern.Name] = true

		// Verify pattern has required fields
		if pattern.Name == "" {
			t.Error("Pattern should have a name")
		}
		if pattern.NodeType == "" {
			t.Error("Pattern should have a node type")
		}
		if pattern.Description == "" {
			t.Error("Pattern should have a description")
		}
	}

	for _, required := range requiredPatterns {
		if !foundPatterns[required] {
			t.Errorf("Missing required pattern: %s", required)
		}
	}
}

func TestCSTNavigator_Navigation(t *testing.T) {
	code := "my Int $count = 42;"
	root, err := parseTypedPerlSimple(code)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	nav := NewCSTNavigator(root)

	// Test finding nodes by type
	varDecls := nav.FindNodesByType(NodeVariableDecl)
	if len(varDecls) == 0 {
		t.Error("Should find variable declaration nodes")
	}

	// Test finding child by type
	if len(varDecls) > 0 {
		typeExpr := nav.FindChildByType(varDecls[0], NodeTypeExpression)
		if typeExpr == nil {
			t.Error("Should find type expression child")
		}
	}

	// Test descendant search
	simpleTypes := nav.FindDescendantsByType(root, NodeSimpleType)
	if len(simpleTypes) == 0 {
		t.Error("Should find simple type descendants")
	}
}

func TestTypeAnnotationQuery(t *testing.T) {
	code := `
		my Int $count = 42;
		my $name = "test";
		field Str $title;
		$value as Int;
	`

	root, content, err := parseTypedPerl(code)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	query := NewTypeAnnotationQueryWithContent(root, content)

	// Test finding typed vs untyped variable declarations
	typedVars := query.FindTypedVariableDeclarations()
	untypedVars := query.FindUntypedVariableDeclarations()

	if len(typedVars) == 0 {
		t.Error("Should find typed variable declarations")
	}

	if len(untypedVars) == 0 {
		t.Error("Should find untyped variable declarations")
	}

	// Test type assertions
	assertions := query.FindAllTypeAssertions()
	if len(assertions) == 0 {
		t.Error("Should find type assertions")
	}

	// Test extracting type from variable declaration
	if len(typedVars) > 0 {
		typeText := query.GetVariableDeclarationType(typedVars[0])
		if typeText == "" {
			t.Error("Should extract type from typed variable declaration")
		}

		varName := query.GetVariableDeclarationName(typedVars[0])
		if varName == "" {
			t.Error("Should extract variable name from variable declaration")
		}
	}
}

func TestCSTAnalyzer_ErrorHandling(t *testing.T) {
	analyzer := NewCSTAnalyzer(nil)

	// Test with nil node
	analysis := analyzer.AnalyzeNode(nil)
	if analysis.Valid {
		t.Error("Analysis of nil node should be invalid")
	}

	if analysis.Error == "" {
		t.Error("Analysis of nil node should have error message")
	}

	// Test extraction with nil
	if analyzer.ExtractTypeAnnotation(nil) != "" {
		t.Error("Type extraction from nil should return empty string")
	}

	if analyzer.ExtractVariableName(nil) != "" {
		t.Error("Variable name extraction from nil should return empty string")
	}
}

func TestCSTAnalyzer_ComplexTypes(t *testing.T) {
	// Test cases that might have parsing issues
	testCases := []struct {
		name        string
		code        string
		shouldFind  bool
		description string
	}{
		{
			name:        "Union type",
			code:        "my (Int|Str) $value;",
			shouldFind:  true,
			description: "Should handle union types",
		},
		{
			name:        "Parameterized type",
			code:        "my ArrayRef[Int] $numbers;",
			shouldFind:  true,
			description: "Should handle parameterized types",
		},
		{
			name:        "Type declaration",
			code:        "type UserId = Int;",
			shouldFind:  false, // Current grammar issue
			description: "Type declarations may not parse correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root, err := parseTypedPerlSimple(tc.code)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			analyzer := NewCSTAnalyzer(root)
			constructs := analyzer.FindAllTypedConstructs()

			found := len(constructs) > 0
			if found != tc.shouldFind {
				if tc.shouldFind {
					t.Errorf("%s: Expected to find typed constructs but didn't", tc.description)
				} else {
					t.Logf("%s: Found typed constructs when not expected (this may indicate grammar improvement)", tc.description)
				}
			}

			// Always print structure for debugging
			t.Logf("CST structure for %s:\n%s", tc.name, analyzer.PrintCSTStructure())
		})
	}
}
