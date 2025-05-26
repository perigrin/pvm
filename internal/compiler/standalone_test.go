// ABOUTME: Standalone tests for compiler package without external dependencies
// ABOUTME: Tests core functionality that doesn't require tree-sitter

package compiler

import (
	"strings"
	"testing"
)

func TestCompilerTypes(t *testing.T) {
	// Test target constants
	if TargetCleanPerl != "clean_perl" {
		t.Errorf("Expected TargetCleanPerl to be 'clean_perl', got '%s'", TargetCleanPerl)
	}

	if TargetTypedPerl != "typed_perl" {
		t.Errorf("Expected TargetTypedPerl to be 'typed_perl', got '%s'", TargetTypedPerl)
	}
}

func TestCompilerError_Standalone(t *testing.T) {
	// Test basic error creation
	err := NewCompilerError("TEST_CODE", "test message")

	if err.Code != "TEST_CODE" {
		t.Errorf("Expected code 'TEST_CODE', got '%s'", err.Code)
	}

	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", err.Message)
	}

	// Test error string without location
	errStr := err.Error()
	expected := "[TEST_CODE] test message"
	if errStr != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, errStr)
	}

	// Test with location
	err = err.WithLocation("/test/file.pl", 10, 5)
	errStr = err.Error()
	expected = "/test/file.pl:10:5: [TEST_CODE] test message"
	if errStr != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, errStr)
	}
}

func TestCleanPerlCompiler_Basic(t *testing.T) {
	compiler := NewCleanPerlCompiler()

	// Test target
	if compiler.Target() != TargetCleanPerl {
		t.Errorf("Expected target %s, got %s", TargetCleanPerl, compiler.Target())
	}

	// Test validation with nil AST
	err := compiler.Validate(nil)
	if err == nil {
		t.Error("Should fail validation with nil AST")
	}

	// Check error code through string matching to avoid type assertion
	if !strings.Contains(err.Error(), ErrInvalidAST) {
		t.Errorf("Expected error to contain %s, got %s", ErrInvalidAST, err.Error())
	}
}

func TestTypedPerlCompiler_Basic(t *testing.T) {
	compiler := NewTypedPerlCompiler()

	// Test target
	if compiler.Target() != TargetTypedPerl {
		t.Errorf("Expected target %s, got %s", TargetTypedPerl, compiler.Target())
	}

	// Test validation with nil AST
	err := compiler.Validate(nil)
	if err == nil {
		t.Error("Should fail validation with nil AST")
	}

	// Check error code through string matching to avoid type assertion
	if !strings.Contains(err.Error(), ErrInvalidAST) {
		t.Errorf("Expected error to contain %s, got %s", ErrInvalidAST, err.Error())
	}
}

func TestCompilerRegistry_Standalone(t *testing.T) {
	registry := NewCompilerRegistry()

	// Test that default compilers are registered
	targets := registry.AvailableTargets()
	if len(targets) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(targets))
	}

	// Verify both expected targets are present
	hasCleanPerl := false
	hasTypedPerl := false

	for _, target := range targets {
		if target == TargetCleanPerl {
			hasCleanPerl = true
		}
		if target == TargetTypedPerl {
			hasTypedPerl = true
		}
	}

	if !hasCleanPerl {
		t.Error("TargetCleanPerl should be available")
	}

	if !hasTypedPerl {
		t.Error("TargetTypedPerl should be available")
	}

	// Test getting compilers
	cleanCompiler, exists := registry.GetCompiler(TargetCleanPerl)
	if !exists {
		t.Error("Clean Perl compiler should be available")
	}
	if cleanCompiler.Target() != TargetCleanPerl {
		t.Errorf("Expected target %s, got %s", TargetCleanPerl, cleanCompiler.Target())
	}

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
