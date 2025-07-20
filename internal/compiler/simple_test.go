// ABOUTME: Simple tests for compiler without parser dependencies
// ABOUTME: Tests core compiler logic and registry functionality

package compiler

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
)

func TestCompilerError(t *testing.T) {
	err := NewCompilerError("TEST_CODE", "test message")

	if err.Code != "TEST_CODE" {
		t.Errorf("Expected code TEST_CODE, got %s", err.Code)
	}

	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got %s", err.Message)
	}

	// Test error string
	errStr := err.Error()
	expected := "[TEST_CODE] test message"
	if errStr != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, errStr)
	}

	// Test with location
	err.WithLocation("/test/file.pl", 10, 5)
	errStr = err.Error()
	expected = "/test/file.pl:10:5: [TEST_CODE] test message"
	if errStr != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, errStr)
	}
}

func TestCompilerRegistry_Basic(t *testing.T) {
	registry := NewCompilerRegistry()

	// Test that default compilers are registered
	targets := registry.AvailableTargets()
	if len(targets) != 3 {
		t.Errorf("Expected 3 targets, got %d", len(targets))
	}

	// Test getting clean Perl compiler
	cleanCompiler, exists := registry.GetCompiler(TargetCleanPerl)
	if !exists {
		t.Error("Clean Perl compiler should be available")
	}
	if cleanCompiler.Target() != TargetCleanPerl {
		t.Errorf("Expected target %s, got %s", TargetCleanPerl, cleanCompiler.Target())
	}

	// Test getting typed Perl compiler
	typedCompiler, exists := registry.GetCompiler(TargetTypedPerl)
	if !exists {
		t.Error("Typed Perl compiler should be available")
	}
	if typedCompiler.Target() != TargetTypedPerl {
		t.Errorf("Expected target %s, got %s", TargetTypedPerl, typedCompiler.Target())
	}

	// Test unknown target
	_, exists = registry.GetCompiler("unknown")
	if exists {
		t.Error("Unknown target should not be available")
	}
}

func TestCleanPerlCompiler_Target(t *testing.T) {
	compiler := NewCleanPerlCompilerUnified()
	if compiler.Target() != TargetCleanPerl {
		t.Errorf("Expected target %s, got %s", TargetCleanPerl, compiler.Target())
	}
}

func TestTypedPerlCompiler_Target(t *testing.T) {
	compiler := NewTypedPerlCompilerUnified()
	if compiler.Target() != TargetTypedPerl {
		t.Errorf("Expected target %s, got %s", TargetTypedPerl, compiler.Target())
	}
}

func TestCleanPerlCompiler_ASTBased(t *testing.T) {
	compiler := NewCleanPerlCompilerUnified()

	t.Run("Variable declaration with type annotation", func(t *testing.T) {
		// Create a simple AST with a typed variable declaration
		// my Int $count = 42;

		// Create variable expression
		countVar := ast.NewVariableExpr("count", "$", ast.Position{Line: 1, Column: 8}, ast.Position{Line: 1, Column: 14})

		// Create literal expression for initializer
		literal := ast.NewLiteralExpr("42", ast.NumberLiteral, ast.Position{Line: 1, Column: 17}, ast.Position{Line: 1, Column: 19})

		// Create type expression (this will be stripped in clean output)
		typeExpr := ast.NewTypeExpression("Int", nil, ast.Position{Line: 1, Column: 4}, ast.Position{Line: 1, Column: 7})

		// Create variable declaration
		varDecl := ast.NewVarDecl("my", []*ast.VariableExpr{countVar}, typeExpr, literal, ast.Position{Line: 1, Column: 1}, ast.Position{Line: 1, Column: 19})

		// Create program with the declaration
		program := ast.NewProgramStmt([]ast.StatementNode{varDecl}, ast.Position{Line: 1, Column: 1}, ast.Position{Line: 1, Column: 19})

		// Create AST
		testAST := &ast.AST{
			Path:   "test.pl",
			Root:   program,
			Source: "my Int $count = 42;",
		}

		// Compile using AST-based approach
		result, err := compiler.Compile(testAST)
		if err != nil {
			t.Fatalf("Compilation failed: %v", err)
		}

		// Should generate clean Perl without type annotation
		// Note: version will be PVM-managed, not hardcoded
		if !strings.Contains(result, "my $count = 42;") {
			t.Errorf("Expected result to contain 'my $count = 42;', got '%s'", result)
		}

		if !strings.HasPrefix(result, "use v") {
			t.Errorf("Expected result to start with version pragma, got '%s'", result)
		}
	})

	t.Run("Handles SimpleAST by re-parsing", func(t *testing.T) {
		// Test that SimpleAST without root node gets handled by re-parsing content
		simpleAST := &SimpleAST{
			Path:    "test.pl",
			Content: "my Int $name = \"test\";",
			Valid:   true,
		}

		result, err := compiler.Compile(simpleAST)
		if err != nil {
			t.Fatalf("Unexpected compilation failure: %v", err)
		}

		// Should successfully compile by re-parsing the content
		if !strings.Contains(result, "my $name = \"test\";") {
			t.Errorf("Expected result to contain cleaned variable declaration, got: %v", result)
		}
	})
}
