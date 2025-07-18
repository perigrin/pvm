// ABOUTME: Tests for JSON marshaling functionality of AST types
// ABOUTME: Ensures JSON serialization works correctly for all AST node types

package ast

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestAST_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		ast  *AST
		want map[string]interface{}
	}{
		{
			name: "simple_ast",
			ast: &AST{
				Path:            "/test.pl",
				Source:          "my $x = 42;",
				TypeAnnotations: []*TypeAnnotation{},
				Errors:          []error{},
			},
			want: map[string]interface{}{
				"path":             "/test.pl",
				"source_length":    float64(11),
				"errors":           []interface{}{},
				"type_annotations": []interface{}{},
			},
		},
		{
			name: "ast_with_errors",
			ast: &AST{
				Path:            "/error.pl",
				Source:          "my $x = ;",
				TypeAnnotations: []*TypeAnnotation{},
				Errors:          []error{errors.New("syntax error")},
			},
			want: map[string]interface{}{
				"path":             "/error.pl",
				"source_length":    float64(9),
				"errors":           []interface{}{"syntax error"},
				"type_annotations": []interface{}{},
			},
		},
		{
			name: "ast_with_type_annotations",
			ast: &AST{
				Path:   "/typed.pl",
				Source: "my Int $x = 42;",
				TypeAnnotations: []*TypeAnnotation{
					{
						AnnotatedItem: "$x",
						TypeExpression: &TypeExpression{
							BaseNode: NewBaseNode("type_expr", Position{Line: 1, Column: 1}, Position{Line: 1, Column: 4}),
							Name:     "Int",
							Kind:     SimpleTypeKind,
						},
						Pos:  Position{Line: 1, Column: 1},
						Kind: VarAnnotation,
					},
				},
				Errors: []error{},
			},
			want: map[string]interface{}{
				"path":          "/typed.pl",
				"source_length": float64(15),
				"errors":        []interface{}{},
			},
		},
		{
			name: "empty_ast",
			ast: &AST{
				Path:            "",
				Source:          "",
				TypeAnnotations: []*TypeAnnotation{},
				Errors:          []error{},
			},
			want: map[string]interface{}{
				"path":             "",
				"source_length":    float64(0),
				"errors":           []interface{}{},
				"type_annotations": []interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.ast)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}

			var got map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &got); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			// Compare basic fields
			if got["path"] != tt.want["path"] {
				t.Errorf("path = %v, want %v", got["path"], tt.want["path"])
			}
			if got["source_length"] != tt.want["source_length"] {
				t.Errorf("source_length = %v, want %v", got["source_length"], tt.want["source_length"])
			}

			// Check that errors array exists and has correct length
			if errors, ok := got["errors"]; ok {
				if errorsArray, ok := errors.([]interface{}); ok {
					wantErrors := tt.want["errors"].([]interface{})
					if len(errorsArray) != len(wantErrors) {
						t.Errorf("errors length = %v, want %v", len(errorsArray), len(wantErrors))
					}
				}
			}

			// Check that type_annotations array exists
			if _, ok := got["type_annotations"]; !ok {
				t.Error("type_annotations field missing from JSON")
			}
		})
	}
}

func TestAST_MarshalJSON_WithRoot(t *testing.T) {
	// Create a simple AST with a root node
	root := NewBaseNode("root", Position{Line: 1, Column: 1}, Position{Line: 1, Column: 10})
	root.SetText("my $x = 42;")

	ast := &AST{
		Path:            "/test.pl",
		Root:            root,
		Source:          "my $x = 42;",
		TypeAnnotations: []*TypeAnnotation{},
		Errors:          []error{},
	}

	jsonBytes, err := json.Marshal(ast)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &got); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check that root node is present
	if root, ok := got["root"]; !ok {
		t.Error("root field missing from JSON")
	} else if rootMap, ok := root.(map[string]interface{}); !ok {
		t.Error("root field is not a map")
	} else {
		if rootMap["type"] != "root" {
			t.Errorf("root.type = %v, want 'root'", rootMap["type"])
		}
		if rootMap["text"] != "my $x = 42;" {
			t.Errorf("root.text = %v, want 'my $x = 42;'", rootMap["text"])
		}
	}
}

func TestNodeToJSON(t *testing.T) {
	tests := []struct {
		name string
		node Node
		want map[string]interface{}
	}{
		{
			name: "base_node",
			node: NewBaseNode("test", Position{Line: 1, Column: 1}, Position{Line: 1, Column: 5}),
			want: map[string]interface{}{
				"type": "test",
				"start": map[string]interface{}{
					"Line":   float64(1),
					"Column": float64(1),
					"Offset": float64(0),
				},
				"end": map[string]interface{}{
					"Line":   float64(1),
					"Column": float64(5),
					"Offset": float64(0),
				},
			},
		},
		{
			name: "nil_node",
			node: nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nodeToJSON(tt.node)

			if tt.want == nil {
				if got != nil {
					t.Errorf("nodeToJSON() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("nodeToJSON() = nil, want non-nil")
				return
			}

			// Convert to JSON and back for comparison
			jsonBytes, err := json.Marshal(got)
			if err != nil {
				t.Fatalf("Failed to marshal result: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &result); err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}

			// Compare basic fields
			if result["type"] != tt.want["type"] {
				t.Errorf("type = %v, want %v", result["type"], tt.want["type"])
			}
		})
	}
}

func TestLiteralKindToString(t *testing.T) {
	tests := []struct {
		kind LiteralKind
		want string
	}{
		{StringLiteral, "string"},
		{NumberLiteral, "number"},
		{BooleanLiteral, "boolean"},
		{UndefLiteral, "undef"},
		{RegexLiteral, "regex"},
		{LiteralKind(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := literalKindToString(tt.kind); got != tt.want {
				t.Errorf("literalKindToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeExpressionKindToString(t *testing.T) {
	tests := []struct {
		kind TypeExpressionKind
		want string
	}{
		{SimpleTypeKind, "simple"},
		{UnionTypeKind, "union"},
		{IntersectionTypeKind, "intersection"},
		{NegationTypeKind, "negation"},
		{ParameterizedTypeKind, "parameterized"},
		{ConstrainedTypeKind, "constrained"},
		{TypeExpressionKind(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := typeExpressionKindToString(tt.kind); got != tt.want {
				t.Errorf("typeExpressionKindToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAST_MarshalJSON_ErrorHandling(t *testing.T) {
	// Test with a valid AST to ensure no errors
	ast := &AST{
		Path:            "/test.pl",
		Source:          "my $x = 42;",
		TypeAnnotations: []*TypeAnnotation{},
		Errors:          []error{},
	}

	jsonBytes, err := json.Marshal(ast)
	if err != nil {
		t.Fatalf("MarshalJSON() should not error on valid AST, got: %v", err)
	}

	// Verify that the JSON is valid
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Generated JSON should be valid, got error: %v", err)
	}

	// Check that all required fields are present
	requiredFields := []string{"path", "source_length", "errors", "type_annotations"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Required field %s missing from JSON", field)
		}
	}
}

func TestAST_MarshalJSON_CompareWithStringRepresentation(t *testing.T) {
	// Create an AST with some content
	ast := &AST{
		Path:            "/test.pl",
		Source:          "my Int $x = 42;",
		TypeAnnotations: []*TypeAnnotation{},
		Errors:          []error{},
	}

	// Test both JSON and string representations
	jsonBytes, err := json.Marshal(ast)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	stringRepr := ast.String()

	// Both should be non-empty
	if len(jsonBytes) == 0 {
		t.Error("JSON representation is empty")
	}
	if len(stringRepr) == 0 {
		t.Error("String representation is empty")
	}

	// JSON should be valid
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("JSON should be valid, got error: %v", err)
	}

	// String representation should contain expected content
	if !strings.Contains(stringRepr, "Path: /test.pl") {
		t.Error("String representation should contain path")
	}
}
