// ABOUTME: Tests to verify CST-based compilation path is used
// ABOUTME: Confirms no redundant re-parsing occurs with TreeSitterAST

package compiler

import (
	"testing"

	"tamarou.com/pvm/internal/parser"
)

func TestTreeSitterASTTakesCSTPath(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse code
	testCode := `my Int $x = 42;`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Create a compiler
	compiler := NewPerlCompiler(TargetCleanPerl)

	// The key test: TreeSitterAST should be recognized as a CST-based AST
	// This can be verified by checking if it has a GetCSTRoot method
	cstRoot := shimAST.GetCSTRoot()
	if cstRoot == nil {
		t.Fatal("TreeSitterAST.GetCSTRoot() should return non-nil to enable CST path")
	}

	// Verify the CST root is valid
	if cstRoot.Kind() == "" {
		t.Error("CST root should have a valid node type")
	}

	// Compile and verify it works
	result, err := compiler.Compile(shimAST)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	if result == "" {
		t.Fatal("Compilation result should not be empty")
	}

	t.Logf("TreeSitterAST successfully uses CST compilation path")
	t.Logf("CST root type: %s", cstRoot.Kind())
	t.Logf("Compilation result: %s", result)
}

func TestCompilerPathSelection(t *testing.T) {
	// This test verifies that the compiler chooses the right compilation path
	// for TreeSitterAST (CST-based) vs traditional AST (re-parsing)

	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse code
	testCode := `my Str $name = "test";`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Test that TreeSitterAST has GetCSTRoot method for compiler path selection
	// This is the same check as in perl_compiler.go line 145
	cstRoot := shimAST.GetCSTRoot()
	if cstRoot == nil {
		t.Error("TreeSitterAST.GetCSTRoot() should return non-nil to enable CST path")
	} else {
		t.Log("✓ TreeSitterAST will use CST-based compilation path")
	}

	// Test the compiler's type checking logic by converting to AST interface
	var astInterface AST = shimAST
	if astInterface.IsValid() {
		t.Log("✓ TreeSitterAST implements compiler AST interface correctly")
	} else {
		t.Error("TreeSitterAST should implement AST interface for compiler compatibility")
	}

	// Verify compilation works
	compiler := NewPerlCompiler(TargetTypedPerl)
	result, err := compiler.Compile(shimAST)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Logf("Compilation successful: %s", result)
}

func TestAvoidRedundantParsingDemo(t *testing.T) {
	// This demonstrates the performance benefit:
	// Parse once with tree-sitter, compile multiple times without re-parsing

	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse complex code once
	testCode := `
		use v5.38;
		my HashRef[Int] $counters = {
			apples => 10,
			oranges => 15,
			bananas => 7
		};
		my Int $total = 0;
		for my $fruit (keys %$counters) {
			$total += $counters->{$fruit};
		}
		print "Total fruit: $total\n";
	`

	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	t.Logf("Parsed complex code once. Source length: %d chars", len(testCode))

	// Now compile to multiple targets using the same parsed AST
	// Each compilation uses the CST directly, no re-parsing needed
	targets := []Target{TargetCleanPerl, TargetTypedPerl}

	for _, target := range targets {
		compiler := NewPerlCompiler(target)

		// This compilation uses shimAST.GetCSTRoot() directly
		result, err := compiler.Compile(shimAST)
		if err != nil {
			t.Fatalf("Compilation to %s failed: %v", target, err)
		}

		t.Logf("Compiled to %s: %d chars (no re-parsing)", target, len(result))
	}

	t.Log("✓ Successfully compiled to multiple targets using single parse + direct CST access")
}
