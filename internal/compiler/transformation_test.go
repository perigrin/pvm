// ABOUTME: Tests for the tree transformation framework
// ABOUTME: Validates that type annotation removal works correctly while preserving other syntax

package compiler

import (
	"strings"
	"testing"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/parser/treesitter"
)

// parseAndTransform parses typed Perl code and transforms it
func parseAndTransform(code string, removeTypes bool) (string, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(treesitter.Language())

	content := []byte(code)
	tree := parser.Parse(content, nil)
	if tree == nil {
		return "", nil
	}

	root := tree.RootNode()

	var result *TransformationResult
	var err error

	if removeTypes {
		result, err = CreateCleanPerl(root, content)
	} else {
		result, err = CreateTypedPerl(root, content)
	}

	if err != nil {
		return "", err
	}

	return result.TransformedCode, nil
}

func TestTransformation_BasicVariableDeclaration(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple typed variable",
			input:    "my Int $count = 42;",
			expected: "my $count = 42;",
		},
		{
			name:     "String variable",
			input:    "my Str $name = \"hello\";",
			expected: "my $name = \"hello\";",
		},
		{
			name:     "Array variable",
			input:    "my ArrayRef @items = ();",
			expected: "my @items = ();",
		},
		{
			name:     "Field declaration",
			input:    "field Int $count;",
			expected: "field $count;",
		},
		{
			name:     "Untyped variable (should remain unchanged)",
			input:    "my $value = 123;",
			expected: "my $value = 123;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseAndTransform(tc.input, true)
			if err != nil {
				t.Fatalf("Transformation failed: %v", err)
			}

			// Normalize whitespace for comparison
			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tc.expected)

			if result != expected {
				t.Errorf("Expected %q, got %q", expected, result)
			}
		})
	}
}

func TestTransformation_ComplexTypes(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Parameterized type",
			input:    "my ArrayRef[Str] $names = [];",
			expected: "my $names = [];",
		},
		{
			name:     "Union type",
			input:    "my (Int|Str) $value;",
			expected: "my $value;",
		},
		{
			name:     "Multiple variables",
			input:    "my Int $a = 1; my Str $b = 'test';",
			expected: "my $a = 1; my $b = 'test';",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseAndTransform(tc.input, true)
			if err != nil {
				t.Fatalf("Transformation failed: %v", err)
			}

			// Normalize whitespace for comparison
			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tc.expected)

			if result != expected {
				t.Errorf("Expected %q, got %q", expected, result)
			}
		})
	}
}

func TestTransformation_TypeAssertions(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple type assertion",
			input:    "my $typed = $value as Int;",
			expected: "my $typed = $value;",
		},
		{
			name:     "Complex expression with type assertion",
			input:    "return $result as Bool;",
			expected: "return $result;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseAndTransform(tc.input, true)
			if err != nil {
				t.Fatalf("Transformation failed: %v", err)
			}

			// Normalize whitespace for comparison
			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tc.expected)

			if result != expected {
				t.Errorf("Expected %q, got %q", expected, result)
			}
		})
	}
}

func TestTransformation_PreservesTypedVersion(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Typed variable should be preserved",
			input: "my Int $count = 42;",
		},
		{
			name:  "Field with type should be preserved",
			input: "field Str $name;",
		},
		{
			name:  "Type assertion should be preserved",
			input: "my $typed = $value as Int;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseAndTransform(tc.input, false) // Don't remove types
			if err != nil {
				t.Fatalf("Transformation failed: %v", err)
			}

			// Result should be identical to input (possibly with normalized whitespace)
			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tc.input)

			if result != expected {
				t.Errorf("Expected preserved input %q, got %q", expected, result)
			}
		})
	}
}

func TestTransformation_PreservesOtherSyntax(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Expressions without types",
			input:    "print \"Hello World\";",
			expected: "print \"Hello World\";",
		},
		{
			name:     "Function calls",
			input:    "my $result = calculate(1, 2, 3);",
			expected: "my $result = calculate(1, 2, 3);",
		},
		{
			name:     "Control structures",
			input:    "if ($condition) { print \"true\"; }",
			expected: "if ($condition) { print \"true\"; }",
		},
		{
			name:     "Mixed typed and untyped",
			input:    "my Int $count = 0; print $count;",
			expected: "my $count = 0; print $count;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseAndTransform(tc.input, true)
			if err != nil {
				t.Fatalf("Transformation failed: %v", err)
			}

			// Normalize whitespace for comparison
			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tc.expected)

			if result != expected {
				t.Errorf("Expected %q, got %q", expected, result)
			}
		})
	}
}

func TestTransformationRules(t *testing.T) {
	t.Run("TypeExpressionRemovalRule", func(t *testing.T) {
		rule := &TypeExpressionRemovalRule{}

		// Create a mock type expression node
		parser := sitter.NewParser()
		parser.SetLanguage(treesitter.Language())

		content := []byte("my Int $var;")
		tree := parser.Parse(content, nil)
		root := tree.RootNode()

		// Find a type expression node
		navigator := NewCSTNavigator(root)
		typeExprs := navigator.FindNodesByType(NodeTypeExpression)

		if len(typeExprs) == 0 {
			t.Skip("No type expression found for testing")
		}

		typeExpr := typeExprs[0]

		if !rule.CanTransform(typeExpr) {
			t.Error("Rule should be able to transform type expression nodes")
		}

		// Test with removal enabled
		transformer := NewCSTTransformer(content, TransformationOptions{RemoveTypeNodes: true})
		result, err := rule.Transform(typeExpr, content, transformer)
		if err != nil {
			t.Fatalf("Transform failed: %v", err)
		}

		if result != "" {
			t.Errorf("Expected empty result when removing type nodes, got %q", result)
		}

		// Test with removal disabled
		transformer = NewCSTTransformer(content, TransformationOptions{RemoveTypeNodes: false})
		result, err = rule.Transform(typeExpr, content, transformer)
		if err != nil {
			t.Fatalf("Transform failed: %v", err)
		}

		if result == "" {
			t.Error("Expected non-empty result when preserving type nodes")
		}
	})

	t.Run("VariableDeclarationCleanupRule", func(t *testing.T) {
		rule := &VariableDeclarationCleanupRule{}

		parser := sitter.NewParser()
		parser.SetLanguage(treesitter.Language())

		content := []byte("my Int $var = 42;")
		tree := parser.Parse(content, nil)
		root := tree.RootNode()

		// Find variable declaration node
		navigator := NewCSTNavigator(root)
		varDecls := navigator.FindNodesByType(NodeVariableDecl)

		if len(varDecls) == 0 {
			t.Skip("No variable declaration found for testing")
		}

		varDecl := varDecls[0]

		if !rule.CanTransform(varDecl) {
			t.Error("Rule should be able to transform variable declaration nodes")
		}

		// Test transformation
		transformer := NewCSTTransformer(content, TransformationOptions{RemoveTypeNodes: true})
		result, err := rule.Transform(varDecl, content, transformer)
		if err != nil {
			t.Fatalf("Transform failed: %v", err)
		}

		// Result should not contain "Int" but should contain "$var"
		if strings.Contains(result, "Int") {
			t.Errorf("Result should not contain type annotation 'Int', got %q", result)
		}

		if !strings.Contains(result, "$var") {
			t.Errorf("Result should contain variable '$var', got %q", result)
		}
	})
}

func TestTransformationOptions(t *testing.T) {
	input := "my Int $count = 42;"

	t.Run("RemoveTypeNodes enabled", func(t *testing.T) {
		result, err := parseAndTransform(input, true)
		if err != nil {
			t.Fatalf("Transformation failed: %v", err)
		}

		if strings.Contains(result, "Int") {
			t.Errorf("Result should not contain 'Int' when removing types, got %q", result)
		}
	})

	t.Run("RemoveTypeNodes disabled", func(t *testing.T) {
		result, err := parseAndTransform(input, false)
		if err != nil {
			t.Fatalf("Transformation failed: %v", err)
		}

		if !strings.Contains(result, "Int") {
			t.Errorf("Result should contain 'Int' when preserving types, got %q", result)
		}
	})
}

func TestTransformationError_Handling(t *testing.T) {
	transformer := NewCSTTransformer([]byte("test"), TransformationOptions{})

	// Test with nil node
	result, err := transformer.Transform(nil)
	if err == nil {
		t.Error("Expected error when transforming nil node")
	}

	if result != "" {
		t.Errorf("Expected empty result for nil node, got %q", result)
	}
}

func TestCreateCleanPerl(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(treesitter.Language())

	content := []byte("my Int $count = 42;")
	tree := parser.Parse(content, nil)
	root := tree.RootNode()

	result, err := CreateCleanPerl(root, content)
	if err != nil {
		t.Fatalf("CreateCleanPerl failed: %v", err)
	}

	if !result.Success {
		t.Error("CreateCleanPerl should succeed")
	}

	if strings.Contains(result.TransformedCode, "Int") {
		t.Errorf("Clean Perl should not contain type annotations, got %q", result.TransformedCode)
	}

	if !strings.Contains(result.TransformedCode, "$count") {
		t.Errorf("Clean Perl should contain variable name, got %q", result.TransformedCode)
	}
}

func TestCreateTypedPerl(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(treesitter.Language())

	content := []byte("my Int $count = 42;")
	tree := parser.Parse(content, nil)
	root := tree.RootNode()

	result, err := CreateTypedPerl(root, content)
	if err != nil {
		t.Fatalf("CreateTypedPerl failed: %v", err)
	}

	if !result.Success {
		t.Error("CreateTypedPerl should succeed")
	}

	if !strings.Contains(result.TransformedCode, "Int") {
		t.Errorf("Typed Perl should contain type annotations, got %q", result.TransformedCode)
	}
}
