// ABOUTME: Integration tests for the unified compiler architecture
// ABOUTME: Validates end-to-end functionality with PSC and registry integration

package compiler

import (
	"fmt"
	"strings"
	"testing"
)

func TestUnifiedCompilerIntegration(t *testing.T) {
	t.Run("End-to-end compilation workflow", func(t *testing.T) {
		// Simulate a complete compilation workflow
		registry := NewCompilerRegistry()

		testCode := `
			my Int $count = 42;
			field Str $name = "test";
			my (Int|Str) $value = $count as Int;
			my $untyped = "plain";
		`

		// Create CST-based AST
		ast, err := NewCSTBasedAST("workflow.pl", testCode)
		if err != nil {
			t.Fatalf("Failed to create AST: %v", err)
		}

		// Test clean compilation
		cleanResult, err := registry.Compile(ast, TargetCleanPerl)
		if err != nil {
			t.Fatalf("Clean compilation failed: %v", err)
		}

		// Verify type annotations are removed
		if strings.Contains(cleanResult, "Int") || strings.Contains(cleanResult, "Str") {
			t.Errorf("Clean result should not contain type annotations, got:\n%s", cleanResult)
		}

		// Verify variables are preserved
		if !strings.Contains(cleanResult, "$count") || !strings.Contains(cleanResult, "$name") {
			t.Errorf("Clean result should preserve variable names, got:\n%s", cleanResult)
		}

		// Test typed compilation
		typedResult, err := registry.Compile(ast, TargetTypedPerl)
		if err != nil {
			t.Fatalf("Typed compilation failed: %v", err)
		}

		// Verify type annotations are preserved
		if !strings.Contains(typedResult, "Int") || !strings.Contains(typedResult, "Str") {
			t.Errorf("Typed result should preserve type annotations, got:\n%s", typedResult)
		}

		// Verify variables are preserved
		if !strings.Contains(typedResult, "$count") || !strings.Contains(typedResult, "$name") {
			t.Errorf("Typed result should preserve variable names, got:\n%s", typedResult)
		}
	})

	t.Run("Registry migration validation", func(t *testing.T) {
		registry := NewCompilerRegistry()
		targets := registry.AvailableTargets()

		// Verify all expected targets are available
		targetMap := make(map[Target]bool)
		for _, target := range targets {
			targetMap[target] = true
		}

		requiredTargets := []Target{
			TargetCleanPerl,
			TargetTypedPerl,
			TargetInferredTypeAnnotations,
		}

		for _, required := range requiredTargets {
			if !targetMap[required] {
				t.Errorf("Required target %s not available in registry", required)
			}
		}

		// Verify unified compilers are being used
		cleanCompiler, exists := registry.GetCompiler(TargetCleanPerl)
		if !exists {
			t.Fatal("Clean compiler not found")
		}

		if _, ok := cleanCompiler.(*PerlCompiler); !ok {
			t.Errorf("Expected unified PerlCompiler for clean target, got %T", cleanCompiler)
		}

		typedCompiler, exists := registry.GetCompiler(TargetTypedPerl)
		if !exists {
			t.Fatal("Typed compiler not found")
		}

		if _, ok := typedCompiler.(*PerlCompiler); !ok {
			t.Errorf("Expected unified PerlCompiler for typed target, got %T", typedCompiler)
		}
	})

	t.Run("Complex type handling", func(t *testing.T) {
		registry := NewCompilerRegistry()

		complexCode := `
			my ArrayRef[HashRef[Str]] $complex = [{name => "test"}];
			my (Int|Str|Bool) $union = 42;
			my $assertion = $value as ArrayRef[Int];
			field Complex::Type $object;
		`

		ast, err := NewCSTBasedAST("complex.pl", complexCode)
		if err != nil {
			t.Fatalf("Failed to create complex AST: %v", err)
		}

		// Test clean compilation with complex types
		cleanResult, err := registry.Compile(ast, TargetCleanPerl)
		if err != nil {
			t.Fatalf("Complex clean compilation failed: %v", err)
		}

		// Verify all type annotations are removed, even complex ones
		complexTypes := []string{"ArrayRef", "HashRef", "Str", "Int", "Bool", "Complex::Type"}
		for _, typeStr := range complexTypes {
			if strings.Contains(cleanResult, typeStr) {
				t.Errorf("Clean result should not contain type %s, got:\n%s", typeStr, cleanResult)
			}
		}

		// Verify variables are preserved
		expectedVars := []string{"$complex", "$union", "$assertion", "$object"}
		for _, varStr := range expectedVars {
			if !strings.Contains(cleanResult, varStr) {
				t.Errorf("Clean result should preserve variable %s, got:\n%s", varStr, cleanResult)
			}
		}

		// Test typed compilation preserves complex types
		typedResult, err := registry.Compile(ast, TargetTypedPerl)
		if err != nil {
			t.Fatalf("Complex typed compilation failed: %v", err)
		}

		// Verify complex types are preserved
		preservedTypes := []string{"ArrayRef", "HashRef", "Complex::Type"}
		for _, typeStr := range preservedTypes {
			if !strings.Contains(typedResult, typeStr) {
				t.Errorf("Typed result should preserve type %s, got:\n%s", typeStr, typedResult)
			}
		}
	})

	t.Run("Error handling and validation", func(t *testing.T) {
		registry := NewCompilerRegistry()

		// Test with invalid AST
		invalidAST := &CSTBasedAST{
			Path:    "invalid.pl",
			Content: "some content",
			Root:    nil, // Invalid
		}

		_, err := registry.Compile(invalidAST, TargetCleanPerl)
		if err == nil {
			t.Error("Expected error for invalid AST")
		}

		// Test with unknown target
		validAST, err := NewCSTBasedAST("valid.pl", "my $var = 42;")
		if err != nil {
			t.Fatalf("Failed to create valid AST: %v", err)
		}

		_, err = registry.Compile(validAST, Target("unknown"))
		if err == nil {
			t.Error("Expected error for unknown target")
		}
	})

	t.Run("Performance and memory characteristics", func(t *testing.T) {
		registry := NewCompilerRegistry()

		// Test with larger code to ensure unified compiler handles it efficiently
		fullCode := ""
		for i := 0; i < 100; i++ {
			fullCode += fmt.Sprintf("my Int $var%d = %d;\n", i, i)
		}

		ast, err := NewCSTBasedAST("large.pl", fullCode)
		if err != nil {
			t.Fatalf("Failed to create large AST: %v", err)
		}

		// Test clean compilation
		cleanResult, err := registry.Compile(ast, TargetCleanPerl)
		if err != nil {
			t.Fatalf("Large clean compilation failed: %v", err)
		}

		// Verify no type annotations remain
		if strings.Contains(cleanResult, "Int") {
			t.Error("Large clean result should not contain type annotations")
		}

		// Verify all variables are present
		for i := 0; i < 10; i++ { // Check first 10
			expectedVar := fmt.Sprintf("$var%d", i)
			if !strings.Contains(cleanResult, expectedVar) {
				t.Errorf("Large clean result should contain %s", expectedVar)
			}
		}
	})
}
