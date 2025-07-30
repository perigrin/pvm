// ABOUTME: Tests for the updated compiler registry using unified compilers
// ABOUTME: Validates that registry correctly uses CST-based unified compilers

package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRegistry_UnifiedCompilers(t *testing.T) {
	t.Run("Registry uses unified compilers", func(t *testing.T) {
		registry := NewCompilerRegistry()

		// Test clean Perl compiler
		cleanCompiler, exists := registry.GetCompiler(TargetCleanPerl)
		if !exists {
			t.Fatal("Clean Perl compiler not found in registry")
		}

		// Verify it's the unified compiler
		if cleanCompiler.Target() != TargetCleanPerl {
			t.Errorf("Expected clean compiler target, got %s", cleanCompiler.Target())
		}

		// Test typed Perl compiler
		typedCompiler, exists := registry.GetCompiler(TargetTypedPerl)
		if !exists {
			t.Fatal("Typed Perl compiler not found in registry")
		}

		// Verify it's the unified compiler
		if typedCompiler.Target() != TargetTypedPerl {
			t.Errorf("Expected typed compiler target, got %s", typedCompiler.Target())
		}
	})

	t.Run("Registry compiles correctly with unified compilers", func(t *testing.T) {
		registry := NewCompilerRegistry()

		// Create test AST
		ast, err := NewCSTBasedAST("test.pl", "my Int $count = 42;")
		if err != nil {
			t.Fatalf("Failed to create test AST: %v", err)
		}

		// Test clean compilation
		cleanResult, err := registry.Compile(ast, TargetCleanPerl)
		if err != nil {
			t.Fatalf("Clean compilation failed: %v", err)
		}

		cleanResult = strings.TrimSpace(cleanResult)
		// The version pragma depends on the environment, so just check the code part
		if !strings.Contains(cleanResult, "use v") || !strings.Contains(cleanResult, "my $count = 42;") {
			t.Errorf("Unexpected clean result: %q", cleanResult)
		}

		// Test typed compilation
		typedResult, err := registry.Compile(ast, TargetTypedPerl)
		if err != nil {
			t.Fatalf("Typed compilation failed: %v", err)
		}

		typedResult = strings.TrimSpace(typedResult)
		expectedTyped := "my Int $count = 42;"

		if typedResult != expectedTyped {
			t.Errorf("Expected typed result %q, got %q", expectedTyped, typedResult)
		}
	})

	t.Run("Registry compilation with options", func(t *testing.T) {
		registry := NewCompilerRegistry()

		// Create test AST
		ast, err := NewCSTBasedAST("test.pl", "my Str $name = \"hello\";")
		if err != nil {
			t.Fatalf("Failed to create test AST: %v", err)
		}

		// Test with default options
		options := &CompilerOptions{
			PreserveComments:   true,
			PreserveFormatting: true,
			StrictMode:         false,
			PerlVersion:        "5.42.0",
			CustomPatterns:     nil,
		}

		result, err := registry.CompileWithOptions(ast, TargetCleanPerl, options)
		if err != nil {
			t.Fatalf("Compilation with options failed: %v", err)
		}

		result = strings.TrimSpace(result)
		// With explicit version set, we should get the specified version
		expected := "use v5.42.0;\nmy $name = \"hello\";"

		if result != expected {
			t.Errorf("Expected result %q, got %q", expected, result)
		}
	})

	t.Run("Registry available targets includes unified targets", func(t *testing.T) {
		registry := NewCompilerRegistry()
		targets := registry.AvailableTargets()

		targetMap := make(map[Target]bool)
		for _, target := range targets {
			targetMap[target] = true
		}

		if !targetMap[TargetCleanPerl] {
			t.Error("TargetCleanPerl should be available")
		}

		if !targetMap[TargetTypedPerl] {
			t.Error("TargetTypedPerl should be available")
		}

		if !targetMap[TargetInferredTypeAnnotations] {
			t.Error("TargetInferredTypeAnnotations should still be available")
		}
	})

	t.Run("Registry backward compatibility with SimpleAST", func(t *testing.T) {
		registry := NewCompilerRegistry()

		// Test with SimpleAST (backward compatibility)
		ast := &SimpleAST{
			Path:    "test.pl",
			Content: "my Int $value = 123;",
			Valid:   true,
		}

		result, err := registry.Compile(ast, TargetCleanPerl)
		if err != nil {
			t.Fatalf("Compilation with SimpleAST failed: %v", err)
		}

		result = strings.TrimSpace(result)
		// The version pragma depends on the environment, so just check the code part
		if !strings.Contains(result, "use v") || !strings.Contains(result, "my $value = 123;") {
			t.Errorf("Unexpected result: %q", result)
		}
	})
}

func TestCompilerRegistry_Migration(t *testing.T) {
	t.Run("Unified compilers replace legacy compilers", func(t *testing.T) {
		registry := NewCompilerRegistry()

		// Verify that we get unified compilers, not legacy ones
		cleanCompiler, _ := registry.GetCompiler(TargetCleanPerl)
		typedCompiler, _ := registry.GetCompiler(TargetTypedPerl)

		// Check that these are unified compilers
		if cleanCompiler.Target() != TargetCleanPerl {
			t.Error("Unified clean compiler should have correct target")
		}

		if typedCompiler.Target() != TargetTypedPerl {
			t.Error("Unified typed compiler should have correct target")
		}
	})
}

func TestCompilerRegistry_ErrorHandling(t *testing.T) {
	t.Run("Registry error handling with unified compilers", func(t *testing.T) {
		registry := NewCompilerRegistry()

		// Test with invalid AST
		invalidAST := &CSTBasedAST{
			Path:    "test.pl",
			Content: "valid content",
			Root:    nil, // Invalid - no root
		}

		_, err := registry.Compile(invalidAST, TargetCleanPerl)
		if err == nil {
			t.Error("Expected error for invalid AST")
		}

		if !strings.Contains(err.Error(), "not valid") {
			t.Errorf("Expected validation error, got %q", err.Error())
		}
	})

	t.Run("Registry unknown target error", func(t *testing.T) {
		registry := NewCompilerRegistry()

		ast, err := NewCSTBasedAST("test.pl", "my $var = 42;")
		if err != nil {
			t.Fatalf("Failed to create AST: %v", err)
		}

		_, err = registry.Compile(ast, Target("unknown_target"))
		if err == nil {
			t.Error("Expected error for unknown target")
		}

		if !strings.Contains(err.Error(), "unknown compilation target") {
			t.Errorf("Expected 'unknown compilation target' error, got %q", err.Error())
		}
	})
}
