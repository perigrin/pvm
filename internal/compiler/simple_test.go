// ABOUTME: Simple tests for compiler without parser dependencies
// ABOUTME: Tests core compiler logic and registry functionality

package compiler

import (
	"testing"
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
	compiler := NewCleanPerlCompiler()
	if compiler.Target() != TargetCleanPerl {
		t.Errorf("Expected target %s, got %s", TargetCleanPerl, compiler.Target())
	}
}

func TestTypedPerlCompiler_Target(t *testing.T) {
	compiler := NewTypedPerlCompiler()
	if compiler.Target() != TargetTypedPerl {
		t.Errorf("Expected target %s, got %s", TargetTypedPerl, compiler.Target())
	}
}

func TestCleanPerlCompiler_StripLine(t *testing.T) {
	compiler := NewCleanPerlCompiler()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "my Int $x = 42;",
			expected: "my $x = 42;",
		},
		{
			input:    "my Str $name = 'test';",
			expected: "my $name = 'test';",
		},
		{
			input:    "sub greet(Str $name) -> Str {",
			expected: "sub greet($name) {",
		},
		{
			input:    "field HashRef[Int] $data;",
			expected: "field $data;",
		},
		{
			input:    "for my Int $i (1..10) {",
			expected: "for my $i (1..10) {",
		},
		{
			input:    "type MyType = Int|Str;",
			expected: "",
		},
		{
			input:    "my $regular = 123;",
			expected: "my $regular = 123;",
		},
	}

	for _, test := range tests {
		result := compiler.cleanLine(test.input)
		if result != test.expected {
			t.Errorf("For input '%s', expected '%s', got '%s'", test.input, test.expected, result)
		}
	}
}
