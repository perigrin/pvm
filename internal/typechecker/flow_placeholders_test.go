// ABOUTME: Tests for flow control analysis placeholder implementations
// ABOUTME: Covers variable declarations, assignments, and function calls

package typechecker

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/typedef"
)

// Helper function to create a test FlowAnalyzer with minimal setup
func createTestFlowAnalyzer(t *testing.T) *FlowAnalyzer {
	t.Helper()

	// Create a type hierarchy and symbol table
	typeStore, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create type store: %v", err)
	}

	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	symbolTable.Package = "test"

	// Create a type checker
	tc := NewTypeChecker(hierarchy, symbolTable, "test")

	// Create flow analyzer with minimal initialization to avoid hanging on builtin type registry
	return &FlowAnalyzer{
		TypeChecker:     tc,
		ProcessedBlocks: make(map[int]bool),
		MaxIterations:   100,
		BuiltinTypes:    nil, // Skip builtin registry creation to avoid hanging
		OperatorTypes:   nil, // Skip operator registry creation
	}
}

// Helper function to create a test TypeState
func createTestTypeState() *TypeState {
	return &TypeState{
		VariableTypes: make(map[string]string),
		RefinedTypes:  make(map[string]string),
		Conditions:    make([]Condition, 0),
	}
}

func TestProcessVariableDeclaration(t *testing.T) {
	tests := []struct {
		name          string
		setupVarDecl  func() *ast.VarDecl
		expectVarType string
		expectError   bool
		errorContains string
	}{
		{
			name: "simple typed declaration",
			setupVarDecl: func() *ast.VarDecl {
				// Create: my Int $var
				intType := ast.NewTypeExpression("Int", nil, ast.Position{Line: 1, Column: 4, Offset: 3}, ast.Position{Line: 1, Column: 7, Offset: 6})
				variable := ast.NewVariableExpr("$var", "$", ast.Position{Line: 1, Column: 8, Offset: 7}, ast.Position{Line: 1, Column: 12, Offset: 11})
				return ast.NewVarDecl("my", []*ast.VariableExpr{variable}, intType, nil, ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 12, Offset: 11})
			},
			expectVarType: "Int",
			expectError:   false,
		},
		{
			name: "untyped declaration",
			setupVarDecl: func() *ast.VarDecl {
				// Create: my $var
				variable := ast.NewVariableExpr("$var", "$", ast.Position{Line: 1, Column: 4, Offset: 3}, ast.Position{Line: 1, Column: 8, Offset: 7})
				return ast.NewVarDecl("my", []*ast.VariableExpr{variable}, nil, nil, ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 8, Offset: 7})
			},
			expectVarType: "Any",
			expectError:   false,
		},
		{
			name: "declaration with string initializer",
			setupVarDecl: func() *ast.VarDecl {
				// Create: my $var = "hello"
				variable := ast.NewVariableExpr("$var", "$", ast.Position{Line: 1, Column: 4, Offset: 3}, ast.Position{Line: 1, Column: 8, Offset: 7})
				initializer := ast.NewLiteralExpr("hello", ast.StringLiteral, ast.Position{Line: 1, Column: 11, Offset: 10}, ast.Position{Line: 1, Column: 18, Offset: 17})
				return ast.NewVarDecl("my", []*ast.VariableExpr{variable}, nil, initializer, ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 18, Offset: 17})
			},
			expectVarType: "Str",
			expectError:   false,
		},
		{
			name: "declaration with numeric initializer",
			setupVarDecl: func() *ast.VarDecl {
				// Create: my $var = 42
				variable := ast.NewVariableExpr("$var", "$", ast.Position{Line: 1, Column: 4, Offset: 3}, ast.Position{Line: 1, Column: 8, Offset: 7})
				initializer := ast.NewLiteralExpr("42", ast.NumberLiteral, ast.Position{Line: 1, Column: 11, Offset: 10}, ast.Position{Line: 1, Column: 13, Offset: 12})
				return ast.NewVarDecl("my", []*ast.VariableExpr{variable}, nil, initializer, ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 13, Offset: 12})
			},
			expectVarType: "Num",
			expectError:   false,
		},
		{
			name: "multiple variables in one declaration",
			setupVarDecl: func() *ast.VarDecl {
				// Create: my ($var1, $var2)
				var1 := ast.NewVariableExpr("$var1", "$", ast.Position{Line: 1, Column: 5, Offset: 4}, ast.Position{Line: 1, Column: 10, Offset: 9})
				var2 := ast.NewVariableExpr("$var2", "$", ast.Position{Line: 1, Column: 12, Offset: 11}, ast.Position{Line: 1, Column: 17, Offset: 16})
				return ast.NewVarDecl("my", []*ast.VariableExpr{var1, var2}, nil, nil, ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 18, Offset: 17})
			},
			expectVarType: "Any", // Both variables should get Any type
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := createTestFlowAnalyzer(t)
			state := createTestTypeState()
			varDecl := tt.setupVarDecl()

			// Process the variable declaration
			errors := analyzer.processVariableDeclaration(varDecl, state)

			// Check for expected errors
			if tt.expectError {
				if len(errors) == 0 {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" {
					found := false
					for _, err := range errors {
						if strings.Contains(err.Error(), tt.errorContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error containing '%s', but got: %v", tt.errorContains, errors)
					}
				}
			} else if len(errors) > 0 {
				t.Errorf("Unexpected errors: %v", errors)
			}

			// Check variable type in state
			variables := varDecl.Variables()
			if len(variables) > 0 {
				varName := variables[0].Name
				actualType, exists := state.VariableTypes[varName]
				if !exists {
					t.Errorf("Variable '%s' not found in type state", varName)
				} else if actualType != tt.expectVarType {
					t.Errorf("Expected type '%s' for variable '%s', got '%s'", tt.expectVarType, varName, actualType)
				}
			}
		})
	}
}

func TestProcessAssignment(t *testing.T) {
	tests := []struct {
		name          string
		setupAssign   func() *ast.AssignmentExpr
		initialState  map[string]string
		expectType    string
		expectError   bool
		errorContains string
	}{
		{
			name: "simple assignment to existing variable",
			setupAssign: func() *ast.AssignmentExpr {
				// Create: $var = 42 (assigning to Num variable to avoid Int/Num compatibility issues)
				left := ast.NewVariableExpr("$var", "$", ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 5, Offset: 4})
				right := ast.NewLiteralExpr("42", ast.NumberLiteral, ast.Position{Line: 1, Column: 8, Offset: 7}, ast.Position{Line: 1, Column: 10, Offset: 9})
				return ast.NewAssignmentExpr(left, right, "=", ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 10, Offset: 9})
			},
			initialState: map[string]string{"$var": "Num"},
			expectType:   "Num", // Should remain Num
			expectError:  false,
		},
		{
			name: "assignment to undeclared variable",
			setupAssign: func() *ast.AssignmentExpr {
				// Create: $new_var = "hello"
				left := ast.NewVariableExpr("$new_var", "$", ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 9, Offset: 8})
				right := ast.NewLiteralExpr("hello", ast.StringLiteral, ast.Position{Line: 1, Column: 12, Offset: 11}, ast.Position{Line: 1, Column: 19, Offset: 18})
				return ast.NewAssignmentExpr(left, right, "=", ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 19, Offset: 18})
			},
			initialState: map[string]string{},
			expectType:   "Str", // Should infer Str from literal
			expectError:  false,
		},
		{
			name: "compound assignment",
			setupAssign: func() *ast.AssignmentExpr {
				// Create: $var += 10
				left := ast.NewVariableExpr("$var", "$", ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 5, Offset: 4})
				right := ast.NewLiteralExpr("10", ast.NumberLiteral, ast.Position{Line: 1, Column: 9, Offset: 8}, ast.Position{Line: 1, Column: 11, Offset: 10})
				return ast.NewAssignmentExpr(left, right, "+=", ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 11, Offset: 10})
			},
			initialState: map[string]string{"$var": "Num"},
			expectType:   "Num", // Should remain Num
			expectError:  false,
		},
		{
			name: "string concatenation assignment",
			setupAssign: func() *ast.AssignmentExpr {
				// Create: $str .= " world"
				left := ast.NewVariableExpr("$str", "$", ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 5, Offset: 4})
				right := ast.NewLiteralExpr(" world", ast.StringLiteral, ast.Position{Line: 1, Column: 9, Offset: 8}, ast.Position{Line: 1, Column: 17, Offset: 16})
				return ast.NewAssignmentExpr(left, right, ".=", ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 17, Offset: 16})
			},
			initialState: map[string]string{"$str": "Str"},
			expectType:   "Str", // Should remain Str
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := createTestFlowAnalyzer(t)
			state := createTestTypeState()

			// Set up initial state
			for varName, varType := range tt.initialState {
				state.VariableTypes[varName] = varType
			}

			assign := tt.setupAssign()

			// Process the assignment
			errors := analyzer.processAssignment(assign, state)

			// Check for expected errors
			if tt.expectError {
				if len(errors) == 0 {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" {
					found := false
					for _, err := range errors {
						if strings.Contains(err.Error(), tt.errorContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error containing '%s', but got: %v", tt.errorContains, errors)
					}
				}
			} else if len(errors) > 0 {
				t.Errorf("Unexpected errors: %v", errors)
			}

			// Check variable type in state
			if varExpr, ok := assign.Left.(*ast.VariableExpr); ok {
				varName := varExpr.Name
				actualType, exists := state.VariableTypes[varName]
				if !exists {
					t.Errorf("Variable '%s' not found in type state after assignment", varName)
				} else if actualType != tt.expectType {
					t.Errorf("Expected type '%s' for variable '%s', got '%s'", tt.expectType, varName, actualType)
				}
			}
		})
	}
}

func TestProcessFunctionCall(t *testing.T) {
	tests := []struct {
		name           string
		setupCall      func() *ast.CallExpr
		setupFunctions map[string]*FunctionSignature
		expectError    bool
		errorContains  string
	}{
		{
			name: "valid function call with correct parameters",
			setupCall: func() *ast.CallExpr {
				// Create: print("hello")
				funcName := ast.NewVariableExpr("print", "", ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 6, Offset: 5})
				arg := ast.NewLiteralExpr("hello", ast.StringLiteral, ast.Position{Line: 1, Column: 7, Offset: 6}, ast.Position{Line: 1, Column: 14, Offset: 13})
				return ast.NewCallExpr(funcName, []ast.ExpressionNode{arg}, false, ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 15, Offset: 14})
			},
			setupFunctions: map[string]*FunctionSignature{
				"print": {
					ParameterTypes: map[string]string{"msg": "Str"},
					ReturnType:     "Bool",
					IsMethod:       false,
				},
			},
			expectError: false,
		},
		{
			name: "function call with wrong parameter count",
			setupCall: func() *ast.CallExpr {
				// Create: add(1) - missing second parameter
				funcName := ast.NewVariableExpr("add", "", ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 4, Offset: 3})
				arg := ast.NewLiteralExpr("1", ast.NumberLiteral, ast.Position{Line: 1, Column: 5, Offset: 4}, ast.Position{Line: 1, Column: 6, Offset: 5})
				return ast.NewCallExpr(funcName, []ast.ExpressionNode{arg}, false, ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 7, Offset: 6})
			},
			setupFunctions: map[string]*FunctionSignature{
				"add": {
					ParameterTypes: map[string]string{"a": "Num", "b": "Num"},
					ReturnType:     "Num",
					IsMethod:       false,
				},
			},
			expectError:   true,
			errorContains: "expects 2 parameters, got 1",
		},
		{
			name: "undefined function call",
			setupCall: func() *ast.CallExpr {
				// Create: unknown_func()
				funcName := ast.NewVariableExpr("unknown_func", "", ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 13, Offset: 12})
				return ast.NewCallExpr(funcName, []ast.ExpressionNode{}, false, ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 15, Offset: 14})
			},
			setupFunctions: map[string]*FunctionSignature{},
			expectError:    false, // Should not error for undefined functions
		},
		{
			name: "function call with no parameters",
			setupCall: func() *ast.CallExpr {
				// Create: get_time()
				funcName := ast.NewVariableExpr("get_time", "", ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 9, Offset: 8})
				return ast.NewCallExpr(funcName, []ast.ExpressionNode{}, false, ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 11, Offset: 10})
			},
			setupFunctions: map[string]*FunctionSignature{
				"get_time": {
					ParameterTypes: map[string]string{},
					ReturnType:     "Int",
					IsMethod:       false,
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := createTestFlowAnalyzer(t)
			state := createTestTypeState()

			// Set up function signatures
			analyzer.TypeChecker.FunctionTypes = tt.setupFunctions

			call := tt.setupCall()

			// Process the function call
			errors := analyzer.processFunctionCall(call, state)

			// Check for expected errors
			if tt.expectError {
				if len(errors) == 0 {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" {
					found := false
					for _, err := range errors {
						if strings.Contains(err.Error(), tt.errorContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error containing '%s', but got: %v", tt.errorContains, errors)
					}
				}
			} else if len(errors) > 0 {
				t.Errorf("Unexpected errors: %v", errors)
			}
		})
	}
}

func TestHelperMethods(t *testing.T) {
	analyzer := createTestFlowAnalyzer(t)

	t.Run("inferTypeFromLiteral", func(t *testing.T) {
		tests := []struct {
			name     string
			literal  *ast.LiteralExpr
			expected string
		}{
			{
				name:     "string literal",
				literal:  ast.NewLiteralExpr("hello", ast.StringLiteral, ast.Position{}, ast.Position{}),
				expected: "Str",
			},
			{
				name:     "number literal",
				literal:  ast.NewLiteralExpr("42", ast.NumberLiteral, ast.Position{}, ast.Position{}),
				expected: "Num",
			},
			{
				name:     "boolean literal",
				literal:  ast.NewLiteralExpr("true", ast.BooleanLiteral, ast.Position{}, ast.Position{}),
				expected: "Bool",
			},
			{
				name:     "undef literal",
				literal:  ast.NewLiteralExpr("undef", ast.UndefLiteral, ast.Position{}, ast.Position{}),
				expected: "Undef",
			},
			{
				name:     "regex literal",
				literal:  ast.NewLiteralExpr("/pattern/", ast.RegexLiteral, ast.Position{}, ast.Position{}),
				expected: "Regex",
			},
			{
				name:     "nil literal",
				literal:  nil,
				expected: "Any",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := analyzer.inferTypeFromLiteral(tt.literal)
				if result != tt.expected {
					t.Errorf("Expected '%s', got '%s'", tt.expected, result)
				}
			})
		}
	})

	t.Run("extractVariableFromExpression", func(t *testing.T) {
		tests := []struct {
			name     string
			expr     ast.ExpressionNode
			expected string
		}{
			{
				name:     "simple variable",
				expr:     ast.NewVariableExpr("$var", "$", ast.Position{}, ast.Position{}),
				expected: "$var",
			},
			{
				name:     "nil expression",
				expr:     nil,
				expected: "",
			},
			{
				name:     "non-variable expression",
				expr:     ast.NewLiteralExpr("42", ast.NumberLiteral, ast.Position{}, ast.Position{}),
				expected: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := analyzer.extractVariableFromExpression(tt.expr)
				if result != tt.expected {
					t.Errorf("Expected '%s', got '%s'", tt.expected, result)
				}
			})
		}
	})

	t.Run("isNumericType", func(t *testing.T) {
		tests := []struct {
			typeStr  string
			expected bool
		}{
			{"Int", true},
			{"Num", true},
			{"Any", true},
			{"Str", false},
			{"Bool", false},
			{"", false},
		}

		for _, tt := range tests {
			t.Run(tt.typeStr, func(t *testing.T) {
				result := analyzer.isNumericType(tt.typeStr)
				if result != tt.expected {
					t.Errorf("Expected %v for '%s', got %v", tt.expected, tt.typeStr, result)
				}
			})
		}
	})

	t.Run("isStringCompatibleType", func(t *testing.T) {
		tests := []struct {
			typeStr  string
			expected bool
		}{
			{"Str", true},
			{"Int", true},
			{"Num", true},
			{"Bool", true},
			{"Any", true},
			{"Undef", false},
			{"", false},
		}

		for _, tt := range tests {
			t.Run(tt.typeStr, func(t *testing.T) {
				result := analyzer.isStringCompatibleType(tt.typeStr)
				if result != tt.expected {
					t.Errorf("Expected %v for '%s', got %v", tt.expected, tt.typeStr, result)
				}
			})
		}
	})
}

// Integration test that combines all three methods
func TestIntegratedFlowAnalysis(t *testing.T) {
	analyzer := createTestFlowAnalyzer(t)
	state := createTestTypeState()

	// Set up a function signature
	analyzer.TypeChecker.FunctionTypes["process"] = &FunctionSignature{
		ParameterTypes: map[string]string{"data": "Str"},
		ReturnType:     "Bool",
		IsMethod:       false,
	}

	// Step 1: Variable declaration - my Str $input
	strType := ast.NewTypeExpression("Str", nil, ast.Position{Line: 1, Column: 4, Offset: 3}, ast.Position{Line: 1, Column: 7, Offset: 6})
	variable := ast.NewVariableExpr("$input", "$", ast.Position{Line: 1, Column: 8, Offset: 7}, ast.Position{Line: 1, Column: 14, Offset: 13})
	varDecl := ast.NewVarDecl("my", []*ast.VariableExpr{variable}, strType, nil, ast.Position{Line: 1, Column: 1, Offset: 0}, ast.Position{Line: 1, Column: 14, Offset: 13})

	errors := analyzer.processVariableDeclaration(varDecl, state)
	if len(errors) > 0 {
		t.Fatalf("Variable declaration failed: %v", errors)
	}

	// Check that variable is in state with correct type
	if varType, exists := state.VariableTypes["$input"]; !exists {
		t.Fatal("Variable $input not found in state")
	} else if varType != "Str" {
		t.Fatalf("Expected type 'Str' for $input, got '%s'", varType)
	}

	// Step 2: Assignment - $input = "hello"
	left := ast.NewVariableExpr("$input", "$", ast.Position{Line: 2, Column: 1, Offset: 0}, ast.Position{Line: 2, Column: 7, Offset: 6})
	right := ast.NewLiteralExpr("hello", ast.StringLiteral, ast.Position{Line: 2, Column: 10, Offset: 9}, ast.Position{Line: 2, Column: 17, Offset: 16})
	assign := ast.NewAssignmentExpr(left, right, "=", ast.Position{Line: 2, Column: 1, Offset: 0}, ast.Position{Line: 2, Column: 17, Offset: 16})

	errors = analyzer.processAssignment(assign, state)
	if len(errors) > 0 {
		t.Fatalf("Assignment failed: %v", errors)
	}

	// Step 3: Function call - process($input)
	funcName := ast.NewVariableExpr("process", "", ast.Position{Line: 3, Column: 1, Offset: 0}, ast.Position{Line: 3, Column: 8, Offset: 7})
	arg := ast.NewVariableExpr("$input", "$", ast.Position{Line: 3, Column: 9, Offset: 8}, ast.Position{Line: 3, Column: 15, Offset: 14})
	call := ast.NewCallExpr(funcName, []ast.ExpressionNode{arg}, false, ast.Position{Line: 3, Column: 1, Offset: 0}, ast.Position{Line: 3, Column: 16, Offset: 15})

	errors = analyzer.processFunctionCall(call, state)
	if len(errors) > 0 {
		t.Fatalf("Function call failed: %v", errors)
	}

	t.Log("Integrated flow analysis completed successfully")
}
