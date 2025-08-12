// ABOUTME: Unit tests for generic type inference engine functions
// ABOUTME: Tests extraction of type parameters, constraints, and parameter types from AST

package typechecker

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/typedef"
)

// mockExpressionNode is a simple mock for testing
type mockExpressionNode struct {
	*ast.BaseNode
	text string
}

func (m *mockExpressionNode) Text() string {
	return m.text
}

func (m *mockExpressionNode) IsExpression() bool {
	return true
}

func newMockExpression(text string) *mockExpressionNode {
	return &mockExpressionNode{
		BaseNode: &ast.BaseNode{},
		text:     text,
	}
}

func TestExtractTypeParameters(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.SubDecl
		expected []string
	}{
		{
			name:     "nil node",
			node:     nil,
			expected: []string{},
		},
		{
			name: "no type parameters",
			node: &ast.SubDecl{
				Name:           "test_func",
				TypeParameters: nil,
			},
			expected: []string{},
		},
		{
			name: "empty type parameters",
			node: &ast.SubDecl{
				Name:           "test_func",
				TypeParameters: []*ast.TypeParameter{},
			},
			expected: []string{},
		},
		{
			name: "single type parameter",
			node: &ast.SubDecl{
				Name: "test_func",
				TypeParameters: []*ast.TypeParameter{
					{Name: "T"},
				},
			},
			expected: []string{"T"},
		},
		{
			name: "multiple type parameters",
			node: &ast.SubDecl{
				Name: "test_func",
				TypeParameters: []*ast.TypeParameter{
					{Name: "T"},
					{Name: "U"},
					{Name: "V"},
				},
			},
			expected: []string{"T", "U", "V"},
		},
		{
			name: "type parameters with nil entry",
			node: &ast.SubDecl{
				Name: "test_func",
				TypeParameters: []*ast.TypeParameter{
					{Name: "T"},
					nil,
					{Name: "U"},
				},
			},
			expected: []string{"T", "U"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTypeParameters(tt.node)

			if len(result) != len(tt.expected) {
				t.Errorf("extractTypeParameters() length = %d, expected %d", len(result), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("extractTypeParameters()[%d] = %s, expected %s", i, result[i], expected)
				}
			}
		})
	}
}

func TestExtractConstraints(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.SubDecl
		expected int // number of constraints expected
	}{
		{
			name:     "nil node",
			node:     nil,
			expected: 0,
		},
		{
			name: "no constraints",
			node: &ast.SubDecl{
				Name:        "test_func",
				Constraints: nil,
			},
			expected: 0,
		},
		{
			name: "empty constraints",
			node: &ast.SubDecl{
				Name:        "test_func",
				Constraints: []*ast.TypeConstraint{},
			},
			expected: 0,
		},
		{
			name: "single constraint",
			node: &ast.SubDecl{
				Name: "test_func",
				Constraints: []*ast.TypeConstraint{
					{
						Parameter:  "T",
						Kind:       ast.TypeConstraintKind,
						Expression: newMockExpression("Serializable"),
					},
				},
			},
			expected: 1,
		},
		{
			name: "multiple constraints",
			node: &ast.SubDecl{
				Name: "test_func",
				Constraints: []*ast.TypeConstraint{
					{
						Parameter:  "T",
						Kind:       ast.TypeConstraintKind,
						Expression: newMockExpression("Serializable"),
					},
					{
						Parameter:  "U",
						Kind:       ast.ProtocolConstraint,
						Expression: newMockExpression("EventHandler"),
					},
				},
			},
			expected: 2,
		},
		{
			name: "constraints with nil entry",
			node: &ast.SubDecl{
				Name: "test_func",
				Constraints: []*ast.TypeConstraint{
					{
						Parameter:  "T",
						Kind:       ast.TypeConstraintKind,
						Expression: newMockExpression("Serializable"),
					},
					nil,
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractConstraints(tt.node)

			if len(result) != tt.expected {
				t.Errorf("extractConstraints() length = %d, expected %d", len(result), tt.expected)
				return
			}

			// Verify constraint structure for non-empty results
			if tt.expected > 0 && tt.node != nil && len(tt.node.Constraints) > 0 {
				firstConstraint := tt.node.Constraints[0]
				if firstConstraint != nil {
					resultConstraint := result[0]
					if resultConstraint.Parameter != firstConstraint.Parameter {
						t.Errorf("constraint parameter = %s, expected %s", resultConstraint.Parameter, firstConstraint.Parameter)
					}
				}
			}
		})
	}
}

func TestExtractParameterTypes(t *testing.T) {
	tests := []struct {
		name     string
		sig      *ast.MethodSignature
		expected int // number of parameter types expected
	}{
		{
			name:     "nil signature",
			sig:      nil,
			expected: 0,
		},
		{
			name: "no parameters",
			sig: &ast.MethodSignature{
				Name:       "test_method",
				Parameters: nil,
			},
			expected: 0,
		},
		{
			name: "empty parameters",
			sig: &ast.MethodSignature{
				Name:       "test_method",
				Parameters: []*ast.ParameterInfo{},
			},
			expected: 0,
		},
		{
			name: "single typed parameter",
			sig: &ast.MethodSignature{
				Name: "test_method",
				Parameters: []*ast.ParameterInfo{
					{
						Name: "param1",
						Type: &ast.TypeExpression{
							Kind: ast.SimpleTypeKind,
							Name: "Int",
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "multiple typed parameters",
			sig: &ast.MethodSignature{
				Name: "test_method",
				Parameters: []*ast.ParameterInfo{
					{
						Name: "param1",
						Type: &ast.TypeExpression{
							Kind: ast.SimpleTypeKind,
							Name: "Int",
						},
					},
					{
						Name: "param2",
						Type: &ast.TypeExpression{
							Kind: ast.SimpleTypeKind,
							Name: "Str",
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "parameter without type",
			sig: &ast.MethodSignature{
				Name: "test_method",
				Parameters: []*ast.ParameterInfo{
					{
						Name: "param1",
						Type: nil,
					},
				},
			},
			expected: 1, // Should default to Any
		},
		{
			name: "parameterized type parameter",
			sig: &ast.MethodSignature{
				Name: "test_method",
				Parameters: []*ast.ParameterInfo{
					{
						Name: "param1",
						Type: &ast.TypeExpression{
							Kind: ast.ParameterizedTypeKind,
							Name: "ArrayRef",
							Parameters: []*ast.TypeExpression{
								{
									Kind: ast.SimpleTypeKind,
									Name: "Int",
								},
							},
						},
					},
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractParameterTypes(tt.sig)

			if len(result) != tt.expected {
				t.Errorf("extractParameterTypes() length = %d, expected %d", len(result), tt.expected)
				return
			}

			// Verify type structure for specific cases
			if tt.expected > 0 && tt.sig != nil && len(tt.sig.Parameters) > 0 {
				firstParam := tt.sig.Parameters[0]
				resultType := result[0]

				if firstParam.Type == nil {
					// Should default to Any
					if simpleType, ok := resultType.(*typedef.SimpleType); ok {
						if simpleType.Name != "Any" {
							t.Errorf("parameter type = %s, expected Any for nil type", simpleType.Name)
						}
					}
				} else {
					// Verify type conversion worked
					if resultType == nil {
						t.Errorf("parameter type conversion failed, got nil")
					}
				}
			}
		})
	}
}

func TestConvertASTConstraintKind(t *testing.T) {
	tests := []struct {
		name     string
		astKind  ast.ConstraintKind
		expected typedef.ConstraintKind
	}{
		{
			name:     "TypeConstraintKind to TraitConstraint",
			astKind:  ast.TypeConstraintKind,
			expected: typedef.TraitConstraint,
		},
		{
			name:     "ProtocolConstraint to ProtocolConstraint",
			astKind:  ast.ProtocolConstraint,
			expected: typedef.ProtocolConstraint,
		},
		{
			name:     "CapabilityConstraint to CapabilityConstraint",
			astKind:  ast.CapabilityConstraint,
			expected: typedef.CapabilityConstraint,
		},
		{
			name:     "ValueConstraint to ValueConstraint",
			astKind:  ast.ValueConstraint,
			expected: typedef.ValueConstraint,
		},
		{
			name:     "VersionConstraint to TraitConstraint (fallback)",
			astKind:  ast.VersionConstraint,
			expected: typedef.TraitConstraint,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertASTConstraintKind(tt.astKind)
			if result != tt.expected {
				t.Errorf("convertASTConstraintKind() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestConvertExpressionToType(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.ExpressionNode
		expected string // expected type name
	}{
		{
			name:     "nil expression",
			expr:     nil,
			expected: "Any",
		},
		{
			name:     "identifier expression",
			expr:     newMockExpression("Serializable"),
			expected: "Serializable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertExpressionToType(tt.expr)

			if simpleType, ok := result.(*typedef.SimpleType); ok {
				if simpleType.Name != tt.expected {
					t.Errorf("convertExpressionToType() = %s, expected %s", simpleType.Name, tt.expected)
				}
			} else {
				t.Errorf("convertExpressionToType() returned non-SimpleType: %T", result)
			}
		})
	}
}
