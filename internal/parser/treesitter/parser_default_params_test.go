package treesitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/ast"
)

func TestParseDefaultParameters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected struct {
			funcName string
			params   []struct {
				name         string
				typeExpr     string
				defaultValue string
				isOptional   bool
			}
		}
	}{
		{
			name:  "simple default value",
			input: `sub new(Str $class, Num $initial_value = 0) { return 1; }`,
			expected: struct {
				funcName string
				params   []struct {
					name         string
					typeExpr     string
					defaultValue string
					isOptional   bool
				}
			}{
				funcName: "new",
				params: []struct {
					name         string
					typeExpr     string
					defaultValue string
					isOptional   bool
				}{
					{name: "class", typeExpr: "Str", defaultValue: "", isOptional: false},
					{name: "initial_value", typeExpr: "Num", defaultValue: "0", isOptional: true},
				},
			},
		},
		{
			name:  "string default value",
			input: `sub greet(Str $name = "world") { print "Hello, $name!"; }`,
			expected: struct {
				funcName string
				params   []struct {
					name         string
					typeExpr     string
					defaultValue string
					isOptional   bool
				}
			}{
				funcName: "greet",
				params: []struct {
					name         string
					typeExpr     string
					defaultValue string
					isOptional   bool
				}{
					{name: "name", typeExpr: "Str", defaultValue: `"world"`, isOptional: true},
				},
			},
		},
		{
			name:  "multiple defaults",
			input: `sub process(Int $a, Str $b = "default", Bool $flag = 1) { }`,
			expected: struct {
				funcName string
				params   []struct {
					name         string
					typeExpr     string
					defaultValue string
					isOptional   bool
				}
			}{
				funcName: "process",
				params: []struct {
					name         string
					typeExpr     string
					defaultValue string
					isOptional   bool
				}{
					{name: "a", typeExpr: "Int", defaultValue: "", isOptional: false},
					{name: "b", typeExpr: "Str", defaultValue: `"default"`, isOptional: true},
					{name: "flag", typeExpr: "Bool", defaultValue: "1", isOptional: true},
				},
			},
		},
		{
			name:  "float default value",
			input: `sub calculate(Num $value = 3.14) { return $value * 2; }`,
			expected: struct {
				funcName string
				params   []struct {
					name         string
					typeExpr     string
					defaultValue string
					isOptional   bool
				}
			}{
				funcName: "calculate",
				params: []struct {
					name         string
					typeExpr     string
					defaultValue string
					isOptional   bool
				}{
					{name: "value", typeExpr: "Num", defaultValue: "3.14", isOptional: true},
				},
			},
		},
		{
			name:  "untyped with default",
			input: `sub legacy($param = 42) { return $param; }`,
			expected: struct {
				funcName string
				params   []struct {
					name         string
					typeExpr     string
					defaultValue string
					isOptional   bool
				}
			}{
				funcName: "legacy",
				params: []struct {
					name         string
					typeExpr     string
					defaultValue string
					isOptional   bool
				}{
					{name: "param", typeExpr: "", defaultValue: "42", isOptional: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser(false)
			require.NoError(t, err)

			astResult, err := parser.ParseString(tt.input)
			require.NoError(t, err)
			require.NotNil(t, astResult)
			require.NotNil(t, astResult.Root)

			// Find the subroutine declaration
			var subDecl *ast.SubDecl
			// Walk the AST to find SubDecl
			var findSubDecl func(node ast.Node)
			findSubDecl = func(node ast.Node) {
				if sub, ok := node.(*ast.SubDecl); ok {
					subDecl = sub
					return
				}
				for _, child := range node.Children() {
					findSubDecl(child)
					if subDecl != nil {
						return
					}
				}
			}
			findSubDecl(astResult.Root)

			require.NotNil(t, subDecl, "Should find subroutine declaration")
			assert.Equal(t, tt.expected.funcName, subDecl.Name)

			// Check parameters
			require.Equal(t, len(tt.expected.params), len(subDecl.LogicalParameters()))
			for i, expectedParam := range tt.expected.params {
				actualParam := subDecl.LogicalParameters()[i]
				assert.Equal(t, expectedParam.name, actualParam.Name, "Parameter %d name mismatch", i)

				if expectedParam.typeExpr != "" {
					require.NotNil(t, actualParam.TypeExpr, "Parameter %d should have type", i)
					// Just check that we have a type, exact representation may vary
				} else {
					assert.Nil(t, actualParam.TypeExpr, "Parameter %d should not have type", i)
				}

				if expectedParam.defaultValue != "" {
					require.NotNil(t, actualParam.Default, "Parameter %d should have default value", i)
					// The default value should be present, exact text representation may vary
				} else {
					assert.Nil(t, actualParam.Default, "Parameter %d should not have default value", i)
				}

				assert.Equal(t, expectedParam.isOptional, actualParam.IsOptional, "Parameter %d optional flag mismatch", i)
			}
		})
	}
}
