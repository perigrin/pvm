// ABOUTME: Comprehensive tests for automatic type annotation generation functionality
// ABOUTME: Tests core functions like extractSubroutineInfo, parameter parsing, and type lookup

package compiler

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/types"
)

// Test extractSubroutineInfo function
func TestExtractSubroutineInfo(t *testing.T) {
	tests := []struct {
		name            string
		nodeText        string
		isMethod        bool
		expectedName    string
		expectedLexical bool
		expectedParams  int
		expectError     bool
	}{
		{
			name:            "simple_subroutine",
			nodeText:        "sub add_numbers {\n    my ($a, $b) = @_;\n    return $a + $b;\n}",
			isMethod:        false,
			expectedName:    "add_numbers",
			expectedLexical: false,
			expectedParams:  0, // No signature in this style
			expectError:     false,
		},
		{
			name:            "lexical_subroutine",
			nodeText:        "my sub calculate($x, $y) {\n    return $x * $y;\n}",
			isMethod:        false,
			expectedName:    "calculate",
			expectedLexical: true,
			expectedParams:  2,
			expectError:     false,
		},
		{
			name:            "method_with_signature",
			nodeText:        "method process($self, $data) {\n    return $self->transform($data);\n}",
			isMethod:        true,
			expectedName:    "process",
			expectedLexical: false,
			expectedParams:  2,
			expectError:     false,
		},
		{
			name:            "lexical_method",
			nodeText:        "my method validate($self, $input, $options) {\n    return 1;\n}",
			isMethod:        true,
			expectedName:    "validate",
			expectedLexical: true,
			expectedParams:  3,
			expectError:     false,
		},
		{
			name:            "empty_text",
			nodeText:        "",
			isMethod:        false,
			expectedName:    "",
			expectedLexical: false,
			expectedParams:  0,
			expectError:     false, // Empty text doesn't error, just produces empty name
		},
		{
			name:            "complex_signature",
			nodeText:        "sub complex_func($param1, $param2 = 'default', @rest) {\n    # body\n}",
			isMethod:        false,
			expectedName:    "complex_func",
			expectedLexical: false,
			expectedParams:  3,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock node compiler
			nc := &nodeCompiler{
				options: InferredCompilerOptions{
					MinimumConfidence: 0.8,
				},
			}

			// Create a mock AST node
			mockNode := &mockASTNode{
				text: tt.nodeText,
				pos:  ast.Position{Line: 1, Column: 1, Offset: 0},
			}

			info, err := nc.extractSubroutineInfo(mockNode, tt.isMethod)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.name)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.name, err)
				return
			}

			if info.Name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, info.Name)
			}

			if info.IsMethod != tt.isMethod {
				t.Errorf("Expected IsMethod %v, got %v", tt.isMethod, info.IsMethod)
			}

			if info.IsLexical != tt.expectedLexical {
				t.Errorf("Expected IsLexical %v, got %v", tt.expectedLexical, info.IsLexical)
			}

			if len(info.Parameters) != tt.expectedParams {
				t.Errorf("Expected %d parameters, got %d", tt.expectedParams, len(info.Parameters))
			}
		})
	}
}

// Test parseParametersFromSignature function
func TestParseParametersFromSignature(t *testing.T) {
	tests := []struct {
		name           string
		signature      string
		expectedParams []string
		expectedVars   []string
	}{
		{
			name:           "simple_parameters",
			signature:      "$a, $b, $c",
			expectedParams: []string{"a", "b", "c"},
			expectedVars:   []string{"$a", "$b", "$c"},
		},
		{
			name:           "typed_parameters",
			signature:      "Int $count, Str $message, Bool $flag",
			expectedParams: []string{"count", "message", "flag"},
			expectedVars:   []string{"$count", "$message", "$flag"},
		},
		{
			name:           "mixed_parameters",
			signature:      "$self, ArrayRef[Int] $numbers, HashRef $options",
			expectedParams: []string{"self", "numbers", "options"},
			expectedVars:   []string{"$self", "$numbers", "$options"},
		},
		{
			name:           "array_and_hash_params",
			signature:      "@items, %config",
			expectedParams: []string{"items", "config"},
			expectedVars:   []string{"@items", "%config"},
		},
		{
			name:           "empty_signature",
			signature:      "",
			expectedParams: []string{},
			expectedVars:   []string{},
		},
		{
			name:           "complex_nested_types",
			signature:      "ArrayRef[HashRef[Int]] $data, Optional[CodeRef] $processor",
			expectedParams: []string{"data", "processor"},
			expectedVars:   []string{"$data", "$processor"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nc := &nodeCompiler{}
			basePos := ast.Position{Line: 1, Column: 1, Offset: 0}

			params := nc.parseParametersFromSignature(tt.signature, basePos)

			if len(params) != len(tt.expectedParams) {
				t.Errorf("Expected %d parameters, got %d", len(tt.expectedParams), len(params))
				return
			}

			for i, param := range params {
				if i >= len(tt.expectedParams) {
					break
				}

				if param.Name != tt.expectedParams[i] {
					t.Errorf("Parameter %d: expected name %s, got %s", i, tt.expectedParams[i], param.Name)
				}

				if param.Variable != tt.expectedVars[i] {
					t.Errorf("Parameter %d: expected variable %s, got %s", i, tt.expectedVars[i], param.Variable)
				}
			}
		})
	}
}

// Test splitParameters function
func TestSplitParameters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple_split",
			input:    "$a, $b, $c",
			expected: []string{"$a", " $b", " $c"},
		},
		{
			name:     "nested_parentheses",
			input:    "ArrayRef[Int] $data, CodeRef($param1, $param2) $func",
			expected: []string{"ArrayRef[Int] $data", " CodeRef($param1, $param2) $func"},
		},
		{
			name:     "complex_nesting",
			input:    "HashRef[ArrayRef[Int]] $nested, ($x, $y)",
			expected: []string{"HashRef[ArrayRef[Int]] $nested", " ($x, $y)"},
		},
		{
			name:     "empty_input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "no_commas",
			input:    "$single_param",
			expected: []string{"$single_param"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nc := &nodeCompiler{}
			result := nc.splitParameters(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d parts, got %d", len(tt.expected), len(result))
				return
			}

			for i, part := range result {
				if part != tt.expected[i] {
					t.Errorf("Part %d: expected %q, got %q", i, tt.expected[i], part)
				}
			}
		})
	}
}

// Test hasTypeInfoForSubroutine function
func TestHasTypeInfoForSubroutine(t *testing.T) {
	tests := []struct {
		name          string
		info          *SubroutineInfo
		typeInfoMap   map[string]*types.TypeInfo
		minConfidence float64
		expected      bool
	}{
		{
			name: "has_parameter_type_info",
			info: &SubroutineInfo{
				Name:     "test_func",
				IsMethod: false,
				Parameters: []ParsedParameterInfo{
					{
						Name:     "count",
						Variable: "$count",
						StartPos: ast.Position{Column: 10, Line: 1},
						EndPos:   ast.Position{Column: 20, Line: 1},
					},
				},
				StartPos: ast.Position{Column: 1, Line: 1},
				EndPos:   ast.Position{Column: 30, Line: 1},
			},
			typeInfoMap: map[string]*types.TypeInfo{
				"parameter_10_20": {
					Type:       types.NewIntType(),
					Confidence: 0.9,
				},
			},
			minConfidence: 0.8,
			expected:      true,
		},
		{
			name: "has_return_type_info",
			info: &SubroutineInfo{
				Name:       "test_func",
				IsMethod:   false,
				Parameters: []ParsedParameterInfo{},
				StartPos:   ast.Position{Column: 1, Line: 1},
				EndPos:     ast.Position{Column: 30, Line: 1},
			},
			typeInfoMap: map[string]*types.TypeInfo{
				"subroutine_return_1_30": {
					Type:       types.NewStrType(),
					Confidence: 0.85,
				},
			},
			minConfidence: 0.8,
			expected:      true,
		},
		{
			name: "low_confidence_type_info",
			info: &SubroutineInfo{
				Name:     "test_func",
				IsMethod: false,
				Parameters: []ParsedParameterInfo{
					{
						Name:     "data",
						Variable: "$data",
						StartPos: ast.Position{Column: 10, Line: 1},
						EndPos:   ast.Position{Column: 20, Line: 1},
					},
				},
				StartPos: ast.Position{Column: 1, Line: 1},
				EndPos:   ast.Position{Column: 30, Line: 1},
			},
			typeInfoMap: map[string]*types.TypeInfo{
				"parameter_10_20": {
					Type:       types.NewIntType(),
					Confidence: 0.5, // Below threshold
				},
			},
			minConfidence: 0.8,
			expected:      false,
		},
		{
			name: "no_type_info",
			info: &SubroutineInfo{
				Name:       "test_func",
				IsMethod:   false,
				Parameters: []ParsedParameterInfo{},
				StartPos:   ast.Position{Column: 1, Line: 1},
				EndPos:     ast.Position{Column: 30, Line: 1},
			},
			typeInfoMap:   map[string]*types.TypeInfo{},
			minConfidence: 0.8,
			expected:      false,
		},
		{
			name: "method_return_type_info",
			info: &SubroutineInfo{
				Name:       "test_method",
				IsMethod:   true,
				Parameters: []ParsedParameterInfo{},
				StartPos:   ast.Position{Column: 1, Line: 1},
				EndPos:     ast.Position{Column: 30, Line: 1},
			},
			typeInfoMap: map[string]*types.TypeInfo{
				"method_return_1_30": {
					Type:       types.NewBoolType(),
					Confidence: 0.95,
				},
			},
			minConfidence: 0.8,
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nc := &nodeCompiler{
				typeInfoMap: tt.typeInfoMap,
				options: InferredCompilerOptions{
					MinimumConfidence: tt.minConfidence,
				},
			}

			result := nc.hasTypeInfoForSubroutine(tt.info)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test generateTypedParameterList function
func TestGenerateTypedParameterList(t *testing.T) {
	tests := []struct {
		name        string
		info        *SubroutineInfo
		typeInfoMap map[string]*types.TypeInfo
		expected    string
	}{
		{
			name: "typed_parameters",
			info: &SubroutineInfo{
				Name:     "test_func",
				IsMethod: false,
				Parameters: []ParsedParameterInfo{
					{
						Name:     "count",
						Variable: "$count",
						StartPos: ast.Position{Column: 10, Line: 1},
						EndPos:   ast.Position{Column: 20, Line: 1},
					},
					{
						Name:     "message",
						Variable: "$message",
						StartPos: ast.Position{Column: 30, Line: 1},
						EndPos:   ast.Position{Column: 40, Line: 1},
					},
				},
			},
			typeInfoMap: map[string]*types.TypeInfo{
				"parameter_10_20": {
					Type:       types.NewIntType(),
					Confidence: 0.9,
				},
				"parameter_30_40": {
					Type:       types.NewStrType(),
					Confidence: 0.85,
				},
			},
			expected: "Int $count, Str $message",
		},
		{
			name: "method_with_self",
			info: &SubroutineInfo{
				Name:     "test_method",
				IsMethod: true,
				Parameters: []ParsedParameterInfo{
					{
						Name:     "self",
						Variable: "$self",
						StartPos: ast.Position{Column: 10, Line: 1},
						EndPos:   ast.Position{Column: 20, Line: 1},
					},
					{
						Name:     "data",
						Variable: "$data",
						StartPos: ast.Position{Column: 30, Line: 1},
						EndPos:   ast.Position{Column: 40, Line: 1},
					},
				},
			},
			typeInfoMap: map[string]*types.TypeInfo{
				"parameter_30_40": {
					Type:       types.NewArrayRefType(types.NewIntType()),
					Confidence: 0.9,
				},
			},
			expected: "Object $self, ArrayRef[Int] $data",
		},
		{
			name: "no_type_info",
			info: &SubroutineInfo{
				Name:     "test_func",
				IsMethod: false,
				Parameters: []ParsedParameterInfo{
					{
						Name:     "param",
						Variable: "$param",
						StartPos: ast.Position{Column: 10, Line: 1},
						EndPos:   ast.Position{Column: 20, Line: 1},
					},
				},
			},
			typeInfoMap: map[string]*types.TypeInfo{},
			expected:    "$param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nc := &nodeCompiler{
				typeInfoMap: tt.typeInfoMap,
				options: InferredCompilerOptions{
					MinimumConfidence: 0.8,
				},
			}

			result := nc.generateTypedParameterList(tt.info)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Mock AST node for testing
type mockASTNode struct {
	text     string
	pos      ast.Position
	children []ast.Node
	parent   ast.Node
}

func (m *mockASTNode) Type() string        { return "mock_node" }
func (m *mockASTNode) Text() string        { return m.text }
func (m *mockASTNode) Start() ast.Position { return m.pos }
func (m *mockASTNode) End() ast.Position {
	return ast.Position{Line: m.pos.Line, Column: m.pos.Column + len(m.text), Offset: m.pos.Offset + len(m.text)}
}
func (m *mockASTNode) Children() []ast.Node         { return m.children }
func (m *mockASTNode) Parent() ast.Node             { return m.parent }
func (m *mockASTNode) SetParent(parent ast.Node)    { m.parent = parent }
func (m *mockASTNode) String() string               { return m.text }
func (m *mockASTNode) ChildCount() int              { return len(m.children) }
func (m *mockASTNode) Child(index int) ast.Node     { return m.children[index] }
func (m *mockASTNode) FieldNameForChild(int) string { return "" }
func (m *mockASTNode) HasError() bool               { return false }
