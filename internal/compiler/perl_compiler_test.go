// ABOUTME: Tests for the unified Perl compiler that works directly with CST
// ABOUTME: Validates CST-based compilation for both clean and typed Perl targets

package compiler

import (
	"strings"
	"testing"
)

func TestCSTBasedAST(t *testing.T) {
	content := "my Int $count = 42;"

	t.Run("NewCSTBasedAST creates valid AST", func(t *testing.T) {
		ast, err := NewCSTBasedAST("test.pl", content)
		if err != nil {
			t.Fatalf("Failed to create CST-based AST: %v", err)
		}

		if ast.GetPath() != "test.pl" {
			t.Errorf("Expected path 'test.pl', got %q", ast.GetPath())
		}

		if !ast.IsValid() {
			t.Error("AST should be valid")
		}

		gotContent, err := ast.GetContent()
		if err != nil {
			t.Fatalf("Failed to get content: %v", err)
		}

		if gotContent != content {
			t.Errorf("Expected content %q, got %q", content, gotContent)
		}

		if ast.GetCSTRoot() == nil {
			t.Error("CST root should not be nil")
		}
	})

	t.Run("Empty content parses successfully", func(t *testing.T) {
		// Empty content should still parse as valid Perl (empty program)
		ast, err := NewCSTBasedAST("test.pl", "")
		if err != nil {
			t.Fatalf("Empty content should parse successfully: %v", err)
		}

		if !ast.IsValid() {
			t.Error("Empty content should be valid")
		}
	})
}

func TestPerlCompiler_Basic(t *testing.T) {
	t.Run("NewPerlCompiler creates compiler with correct target", func(t *testing.T) {
		cleanCompiler := NewPerlCompiler(TargetCleanPerl)
		if cleanCompiler.Target() != TargetCleanPerl {
			t.Errorf("Expected target %s, got %s", TargetCleanPerl, cleanCompiler.Target())
		}

		typedCompiler := NewPerlCompiler(TargetTypedPerl)
		if typedCompiler.Target() != TargetTypedPerl {
			t.Errorf("Expected target %s, got %s", TargetTypedPerl, typedCompiler.Target())
		}
	})

	t.Run("NewCleanPerlCompilerUnified creates clean compiler", func(t *testing.T) {
		compiler := NewCleanPerlCompilerUnified()
		if compiler.Target() != TargetCleanPerl {
			t.Errorf("Expected target %s, got %s", TargetCleanPerl, compiler.Target())
		}
	})

	t.Run("NewTypedPerlCompilerUnified creates typed compiler", func(t *testing.T) {
		compiler := NewTypedPerlCompilerUnified()
		if compiler.Target() != TargetTypedPerl {
			t.Errorf("Expected target %s, got %s", TargetTypedPerl, compiler.Target())
		}
	})
}

func TestPerlCompiler_Validation(t *testing.T) {
	compiler := NewPerlCompiler(TargetCleanPerl)

	t.Run("Validate with nil AST returns error", func(t *testing.T) {
		err := compiler.Validate(nil)
		if err == nil {
			t.Error("Expected error for nil AST")
		}

		if !strings.Contains(err.Error(), "AST cannot be nil") {
			t.Errorf("Expected 'AST cannot be nil' error, got %q", err.Error())
		}
	})

	t.Run("Validate with valid AST succeeds", func(t *testing.T) {
		ast, err := NewCSTBasedAST("test.pl", "my $var = 42;")
		if err != nil {
			t.Fatalf("Failed to create AST: %v", err)
		}

		err = compiler.Validate(ast)
		if err != nil {
			t.Errorf("Validation should succeed for valid AST: %v", err)
		}
	})

	t.Run("Validate with invalid AST returns error", func(t *testing.T) {
		// Create an invalid AST
		ast := &CSTBasedAST{
			Path:    "test.pl",
			Content: "",
			Root:    nil, // Invalid - no root
		}

		err := compiler.Validate(ast)
		if err == nil {
			t.Error("Expected error for invalid AST")
		}
	})
}

func TestPerlCompiler_Compilation(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		cleanExpected string
		typedExpected string
	}{
		{
			name:          "Simple typed variable",
			input:         "my Int $count = 42;",
			cleanExpected: "use v5.38.2;\nmy $count = 42;",
			typedExpected: "my Int $count = 42;",
		},
		{
			name:          "Field declaration",
			input:         "field Str $name;",
			cleanExpected: "use v5.38.2;\nfield $name;",
			typedExpected: "field Str $name;",
		},
		{
			name:          "Union type",
			input:         "my (Int|Str) $value;",
			cleanExpected: "use v5.38.2;\nmy $value;",
			typedExpected: "my (Int|Str) $value;",
		},
		{
			name:          "Untyped variable",
			input:         "my $plain = 123;",
			cleanExpected: "use v5.38.2;\nmy $plain = 123;",
			typedExpected: "my $plain = 123;",
		},
		{
			name:          "Type assertion",
			input:         "my $typed = $value as Int;",
			cleanExpected: "use v5.38.2;\nmy $typed = $value;",
			typedExpected: "my $typed = $value as Int;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test clean compilation
			t.Run("Clean Perl", func(t *testing.T) {
				cleanCompiler := NewPerlCompiler(TargetCleanPerl)
				ast, err := NewCSTBasedAST("test.pl", tc.input)
				if err != nil {
					t.Fatalf("Failed to create AST: %v", err)
				}

				result, err := cleanCompiler.Compile(ast)
				if err != nil {
					t.Fatalf("Compilation failed: %v", err)
				}

				result = strings.TrimSpace(result)
				expected := strings.TrimSpace(tc.cleanExpected)

				if result != expected {
					t.Errorf("Expected %q, got %q", expected, result)
				}
			})

			// Test typed compilation
			t.Run("Typed Perl", func(t *testing.T) {
				typedCompiler := NewPerlCompiler(TargetTypedPerl)
				ast, err := NewCSTBasedAST("test.pl", tc.input)
				if err != nil {
					t.Fatalf("Failed to create AST: %v", err)
				}

				result, err := typedCompiler.Compile(ast)
				if err != nil {
					t.Fatalf("Compilation failed: %v", err)
				}

				result = strings.TrimSpace(result)
				expected := strings.TrimSpace(tc.typedExpected)

				if result != expected {
					t.Errorf("Expected %q, got %q", expected, result)
				}
			})
		})
	}
}

func TestPerlCompiler_CompileString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		target   Target
	}{
		{
			name:     "Clean Perl from string",
			input:    "my Int $count = 42;",
			expected: "use v5.38.2;\nmy $count = 42;",
			target:   TargetCleanPerl,
		},
		{
			name:     "Typed Perl from string",
			input:    "my Int $count = 42;",
			expected: "my Int $count = 42;",
			target:   TargetTypedPerl,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			compiler := NewPerlCompiler(tc.target)

			result, err := compiler.CompileString(tc.input)
			if err != nil {
				t.Fatalf("CompileString failed: %v", err)
			}

			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tc.expected)

			if result != expected {
				t.Errorf("Expected %q, got %q", expected, result)
			}
		})
	}
}

func TestPerlCompiler_BackwardCompatibility(t *testing.T) {
	// Test that the unified compiler works with existing AST types (like SimpleAST)
	t.Run("Works with SimpleAST", func(t *testing.T) {
		compiler := NewPerlCompiler(TargetCleanPerl)

		// Create a SimpleAST (for backward compatibility testing)
		simpleAST := &SimpleAST{
			Path:    "test.pl",
			Content: "my Int $count = 42;",
			Valid:   true,
		}

		result, err := compiler.Compile(simpleAST)
		if err != nil {
			t.Fatalf("Compilation with SimpleAST failed: %v", err)
		}

		result = strings.TrimSpace(result)
		expected := "use v5.38.2;\nmy $count = 42;"

		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})
}

func TestPerlCompiler_Options(t *testing.T) {
	compiler := NewPerlCompiler(TargetCleanPerl)

	t.Run("Default options are set", func(t *testing.T) {
		options := compiler.GetOptions()

		if !options.PreserveComments {
			t.Error("PreserveComments should be true by default")
		}

		if !options.PreserveFormatting {
			t.Error("PreserveFormatting should be true by default")
		}

		if options.StrictMode {
			t.Error("StrictMode should be false by default")
		}
	})

	t.Run("Options can be updated", func(t *testing.T) {
		newOptions := CompilerOptions{
			PreserveComments:   false,
			PreserveFormatting: false,
			StrictMode:         true,
			CustomPatterns:     map[string]string{"test": "value"},
		}

		compiler.SetOptions(newOptions)
		options := compiler.GetOptions()

		if options.PreserveComments {
			t.Error("PreserveComments should be false after update")
		}

		if options.PreserveFormatting {
			t.Error("PreserveFormatting should be false after update")
		}

		if !options.StrictMode {
			t.Error("StrictMode should be true after update")
		}

		if options.CustomPatterns["test"] != "value" {
			t.Error("CustomPatterns should be updated")
		}
	})
}

func TestPerlCompiler_TargetSupport(t *testing.T) {
	compiler := NewPerlCompiler(TargetCleanPerl)

	t.Run("SupportsTarget works correctly", func(t *testing.T) {
		if !compiler.SupportsTarget(TargetCleanPerl) {
			t.Error("Should support TargetCleanPerl")
		}

		if !compiler.SupportsTarget(TargetTypedPerl) {
			t.Error("Should support TargetTypedPerl")
		}

		if compiler.SupportsTarget(TargetInferredTypeAnnotations) {
			t.Error("Should not support TargetInferredTypeAnnotations")
		}
	})

	t.Run("GetSupportedTargets returns correct targets", func(t *testing.T) {
		targets := compiler.GetSupportedTargets()

		expectedTargets := []Target{TargetCleanPerl, TargetTypedPerl}
		if len(targets) != len(expectedTargets) {
			t.Errorf("Expected %d targets, got %d", len(expectedTargets), len(targets))
		}

		targetMap := make(map[Target]bool)
		for _, target := range targets {
			targetMap[target] = true
		}

		for _, expected := range expectedTargets {
			if !targetMap[expected] {
				t.Errorf("Missing expected target: %s", expected)
			}
		}
	})
}

func TestCreateUnifiedCompilerForTarget(t *testing.T) {
	t.Run("Creates compiler for supported targets", func(t *testing.T) {
		cleanCompiler, err := CreateUnifiedCompilerForTarget(TargetCleanPerl)
		if err != nil {
			t.Fatalf("Failed to create clean compiler: %v", err)
		}

		if cleanCompiler.Target() != TargetCleanPerl {
			t.Errorf("Expected target %s, got %s", TargetCleanPerl, cleanCompiler.Target())
		}

		typedCompiler, err := CreateUnifiedCompilerForTarget(TargetTypedPerl)
		if err != nil {
			t.Fatalf("Failed to create typed compiler: %v", err)
		}

		if typedCompiler.Target() != TargetTypedPerl {
			t.Errorf("Expected target %s, got %s", TargetTypedPerl, typedCompiler.Target())
		}
	})

	t.Run("Returns error for unsupported targets", func(t *testing.T) {
		_, err := CreateUnifiedCompilerForTarget(TargetInferredTypeAnnotations)
		if err == nil {
			t.Error("Expected error for unsupported target")
		}

		if !strings.Contains(err.Error(), "unsupported target") {
			t.Errorf("Expected 'unsupported target' error, got %q", err.Error())
		}
	})
}

func TestPerlCompiler_ErrorHandling(t *testing.T) {
	t.Run("Compilation error for invalid target", func(t *testing.T) {
		// Create a compiler with an invalid target
		invalidCompiler := &PerlCompiler{
			target: Target("invalid_target"),
		}

		ast, err := NewCSTBasedAST("test.pl", "my $var = 42;")
		if err != nil {
			t.Fatalf("Failed to create AST: %v", err)
		}

		_, err = invalidCompiler.Compile(ast)
		if err == nil {
			t.Error("Expected error for invalid target")
		}

		if !strings.Contains(err.Error(), "unsupported compilation target") {
			t.Errorf("Expected 'unsupported compilation target' error, got %q", err.Error())
		}
	})
}
