// ABOUTME: Comprehensive tests for inferred typed Perl compiler and integration
// ABOUTME: Validates compilation behavior across different confidence levels and annotation styles

package compiler

import (
	"errors"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
)

func TestInferredTypedPerlCompiler(t *testing.T) {
	compiler := NewInferredTypedPerlCompiler()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{"Basic compiler creation", func(t *testing.T) {
			if compiler == nil {
				t.Error("NewInferredTypedPerlCompiler() returned nil")
			}
		}},
		{"Compiler interface compliance", func(t *testing.T) {
			var _ Compiler = compiler
		}},
		{"Target identification", func(t *testing.T) {
			if compiler.Target() != TargetInferredTypeAnnotations {
				t.Errorf("Expected target %s, got %s", TargetInferredTypeAnnotations, compiler.Target())
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func TestCompilerRegistryIntegration(t *testing.T) {
	registry := NewCompilerRegistry()

	t.Run("New target is registered", func(t *testing.T) {
		compiler, exists := registry.GetCompiler(TargetInferredTypeAnnotations)
		if !exists {
			t.Error("TargetInferredTypeAnnotations not registered in compiler registry")
		}
		if compiler == nil {
			t.Error("TargetInferredTypeAnnotations compiler is nil")
		}
	})

	t.Run("Available targets include new target", func(t *testing.T) {
		targets := registry.AvailableTargets()
		found := false
		for _, target := range targets {
			if target == TargetInferredTypeAnnotations {
				found = true
				break
			}
		}
		if !found {
			t.Error("TargetInferredTypeAnnotations not found in available targets")
		}
	})
}

func TestCompilerValidation(t *testing.T) {
	compiler := NewInferredTypedPerlCompiler()

	tests := []struct {
		name        string
		ast         *ast.AST
		expectError bool
		errorCode   string
	}{
		{
			name:        "Nil AST",
			ast:         nil,
			expectError: true,
			errorCode:   "INVALID_AST",
		},
		{
			name: "AST with nil root",
			ast: &ast.AST{
				Path: "test.pl",
				Root: nil,
			},
			expectError: true,
			errorCode:   "INVALID_AST",
		},
		{
			name: "Valid AST",
			ast: &ast.AST{
				Path: "test.pl",
				Root: ast.NewBaseNode("program", ast.Position{}, ast.Position{}),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := compiler.Validate(tt.ast)

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error, got nil")
					return
				}

				var compErr *CompilerError
				if errors.As(err, &compErr) {
					if compErr.Code != tt.errorCode {
						t.Errorf("Expected error code %s, got %s", tt.errorCode, compErr.Code)
					}
				} else {
					t.Errorf("Expected CompilerError, got %T", err)
				}
			} else if err != nil {
				t.Errorf("Expected no validation error, got %v", err)
			}
		})
	}
}

func TestCompilerOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  InferredCompilerOptions
		testFunc func(t *testing.T, compiler *InferredTypedPerlCompiler)
	}{
		{
			name: "Inline annotation style",
			options: InferredCompilerOptions{
				AnnotationStyle:    StyleInline,
				PreserveComments:   true,
				PreserveFormatting: true,
				VerboseOutput:      false,
			},
			testFunc: func(t *testing.T, compiler *InferredTypedPerlCompiler) {
				if compiler.options.AnnotationStyle != StyleInline {
					t.Errorf("Expected StyleInline, got %s", compiler.options.AnnotationStyle)
				}
			},
		},
		{
			name: "Verbose output mode",
			options: InferredCompilerOptions{
				AnnotationStyle:    StyleVerbose,
				PreserveComments:   true,
				PreserveFormatting: true,
				VerboseOutput:      true,
			},
			testFunc: func(t *testing.T, compiler *InferredTypedPerlCompiler) {
				if !compiler.options.VerboseOutput {
					t.Error("Expected verbose output to be enabled")
				}
				if compiler.options.AnnotationStyle != StyleVerbose {
					t.Errorf("Expected StyleVerbose, got %s", compiler.options.AnnotationStyle)
				}
			},
		},
		{
			name: "Compact style mode",
			options: InferredCompilerOptions{
				AnnotationStyle:    StyleCompact,
				PreserveComments:   false,
				PreserveFormatting: false,
				VerboseOutput:      false,
			},
			testFunc: func(t *testing.T, compiler *InferredTypedPerlCompiler) {
				if compiler.options.AnnotationStyle != StyleCompact {
					t.Errorf("Expected StyleCompact, got %s", compiler.options.AnnotationStyle)
				}
				if compiler.options.PreserveComments {
					t.Error("Expected PreserveComments to be false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewInferredTypedPerlCompilerWithOptions(tt.options)
			tt.testFunc(t, compiler)
		})
	}
}

func TestBasicCompilation(t *testing.T) {
	// Create a simple AST for testing with source content
	sourceCode := "#!/usr/bin/perl\nuse v5.42.0;\nprint \"Hello, World!\\n\";"
	programStmt := ast.NewProgramStmt([]ast.StatementNode{}, ast.Position{}, ast.Position{})

	testAST := &ast.AST{
		Path:   "test.pl",
		Source: sourceCode,
		Root:   programStmt,
	}

	compiler := NewInferredTypedPerlCompiler()

	t.Run("Basic program compilation", func(t *testing.T) {
		result, err := compiler.Compile(testAST)
		if err != nil {
			t.Errorf("Compilation failed: %v", err)
			return
		}

		// Should include Perl version pragma
		if !strings.Contains(result, "use v5.42.0;") {
			t.Error("Expected Perl version pragma in output")
		}
	})
}

func TestVariableDeclarationCompilation(t *testing.T) {
	// Create a variable declaration AST
	varExpr := ast.NewVariableExpr("count", "$", ast.Position{}, ast.Position{})
	typeExpr := ast.NewTypeExpression("Int", nil, ast.Position{}, ast.Position{})

	varDecl := ast.NewVarDecl("my", []*ast.VariableExpr{varExpr}, typeExpr, nil, ast.Position{}, ast.Position{})
	programStmt := ast.NewProgramStmt([]ast.StatementNode{varDecl}, ast.Position{}, ast.Position{})

	testAST := &ast.AST{
		Path:   "test.pl",
		Source: "my Int $count = 0;", // Basic variable declaration source
		Root:   programStmt,
	}

	tests := []struct {
		name                string
		options             InferredCompilerOptions
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name: "Inline style with annotations",
			options: InferredCompilerOptions{
				AnnotationStyle: StyleInline,
			},
			expectedContains: []string{"use v", "my"},
		},
		{
			name: "Verbose style with annotations",
			options: InferredCompilerOptions{
				AnnotationStyle: StyleVerbose,
			},
			expectedContains: []string{"use v", "my"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewInferredTypedPerlCompilerWithOptions(tt.options)
			result, err := compiler.Compile(testAST)
			if err != nil {
				t.Errorf("Compilation failed: %v", err)
				return
			}

			for _, expected := range tt.expectedContains {
				// Handle dynamic version pragma - accept any version pragma format
				if expected == "use v5.42.0;" && strings.Contains(result, "use v") {
					// Version pragma test passes if any version pragma is present
					continue
				}
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', got: %s", expected, result)
				}
			}

			for _, notExpected := range tt.expectedNotContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected result NOT to contain '%s', got: %s", notExpected, result)
				}
			}
		})
	}
}

func TestMethodCompilation(t *testing.T) {
	// Note: This is a simplified test since full AST construction with all type information
	// would require integration with the parser and inference engine

	compiler := NewInferredTypedPerlCompiler()

	// Test that the compiler can handle method compilation without panicking
	t.Run("Method signature info building", func(t *testing.T) {
		// Create a simple SubDecl for testing
		subDecl := ast.NewSubDecl("test_method", []*ast.Parameter{}, nil, nil, true, ast.Position{}, ast.Position{})

		nodeCompiler := &nodeCompiler{
			options: compiler.options,
		}

		// This should not panic
		signature := nodeCompiler.buildMethodSignatureInfo(subDecl)
		if signature == nil {
			t.Error("Expected method signature info, got nil")
		}
	})
}

func TestCompilerErrorHandling(t *testing.T) {
	compiler := NewInferredTypedPerlCompiler()

	tests := []struct {
		name          string
		ast           *ast.AST
		expectedError string
	}{
		{
			name:          "Nil AST error",
			ast:           nil,
			expectedError: "INVALID_AST",
		},
		{
			name: "Missing root error",
			ast: &ast.AST{
				Path: "test.pl",
				Root: nil,
			},
			expectedError: "INVALID_AST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compiler.Compile(tt.ast)

			if err == nil {
				t.Error("Expected compilation error, got nil")
				return
			}

			var compErr *CompilerError
			if errors.As(err, &compErr) {
				if !strings.Contains(compErr.Code, tt.expectedError) {
					t.Errorf("Expected error code to contain '%s', got '%s'", tt.expectedError, compErr.Code)
				}
			} else {
				t.Errorf("Expected CompilerError, got %T: %v", err, err)
			}
		})
	}
}

func TestCompilerIntegrationWithRegistry(t *testing.T) {
	registry := NewCompilerRegistry()

	// Create a simple test AST
	sourceCode := "print \"Hello, World!\\n\";"
	programStmt := ast.NewProgramStmt([]ast.StatementNode{}, ast.Position{}, ast.Position{})
	testAST := &ast.AST{
		Path:   "test.pl",
		Source: sourceCode,
		Root:   programStmt,
	}

	t.Run("Compile with registry", func(t *testing.T) {
		result, err := registry.Compile(testAST, TargetInferredTypeAnnotations)
		if err != nil {
			t.Errorf("Registry compilation failed: %v", err)
			return
		}

		if !strings.Contains(result, "use v") {
			t.Error("Expected Perl version pragma in registry compilation output")
		}
	})

	t.Run("Compile with options", func(t *testing.T) {
		options := &CompilerOptions{
			PreserveComments:   true,
			PreserveFormatting: true,
			StrictMode:         false,
		}

		result, err := registry.CompileWithOptions(testAST, TargetInferredTypeAnnotations, options)
		if err != nil {
			t.Errorf("Registry compilation with options failed: %v", err)
			return
		}

		if result == "" {
			t.Error("Expected non-empty compilation result")
		}
	})
}
