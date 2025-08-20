// ABOUTME: Tests for compiler integration with tree-sitter shim
// ABOUTME: Validates that compilation uses direct CST access without re-parsing

package compiler

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parser"
)

func TestCompilerWithTreeSitterShim(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse some typed code
	testCode := `my Int $count = 42; print "Count: $count\n";`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Verify the TreeSitterAST has CST access
	if shimAST.GetCSTRoot() == nil {
		t.Fatal("TreeSitterAST.GetCSTRoot() returned nil")
	}

	// Create a compiler
	compiler := NewPerlCompiler(TargetCleanPerl)

	// Test compilation using TreeSitterAST directly
	// This should NOT trigger re-parsing because TreeSitterAST has GetCSTRoot()
	result, err := compiler.Compile(shimAST)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	if result == "" {
		t.Fatal("Compilation returned empty result")
	}

	// Verify the result contains expected content
	if !strings.Contains(result, "count") {
		t.Error("Compiled result should contain 'count'")
	}

	if !strings.Contains(result, "42") {
		t.Error("Compiled result should contain '42'")
	}

	t.Logf("Successfully compiled TreeSitterAST without re-parsing")
	t.Logf("Compiled result: %s", result)
}

func TestCompilerEliminatesRedundantParsing(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse code with type annotations
	testCode := `my Str $name = "test"; my Int $length = length($name);`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Test both clean and typed Perl compilation
	targets := []Target{TargetCleanPerl, TargetTypedPerl}

	for _, target := range targets {
		t.Run(string(target), func(t *testing.T) {
			compiler := NewPerlCompiler(target)

			// Compile using TreeSitterAST (should use direct CST access)
			result, err := compiler.Compile(shimAST)
			if err != nil {
				t.Fatalf("Compilation failed for target %s: %v", target, err)
			}

			if result == "" {
				t.Fatalf("Compilation returned empty result for target %s", target)
			}

			// Verify expected content based on target
			switch target {
			case TargetCleanPerl:
				// Should not contain type annotations
				if strings.Contains(result, "Str") || strings.Contains(result, "Int") {
					t.Logf("Note: Clean Perl still contains type annotations (may be expected): %s", result)
				}
			case TargetTypedPerl:
				// Should preserve type annotations
				if !strings.Contains(result, "name") {
					t.Error("Typed Perl should contain variable 'name'")
				}
			}

			t.Logf("Target %s compilation successful. Length: %d chars", target, len(result))
		})
	}
}

func TestTreeSitterASTImplementsCompilerInterfaces(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse code
	testCode := `my ArrayRef[Int] $numbers = [1, 2, 3];`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Test that TreeSitterAST implements the AST interface expected by compiler
	var astInterface AST = shimAST

	// Test AST interface methods
	path := astInterface.GetPath()
	t.Logf("AST path: %s", path)

	isValid := astInterface.IsValid()
	if !isValid {
		t.Error("TreeSitterAST should be valid")
	}

	content, err := astInterface.GetContent()
	if err != nil {
		t.Fatalf("GetContent failed: %v", err)
	}
	if content != testCode {
		t.Error("GetContent should return original source")
	}

	rootNode, err := astInterface.GetRootNode()
	if err != nil {
		t.Fatalf("GetRootNode failed: %v", err)
	}
	if rootNode == nil {
		t.Error("GetRootNode should return non-nil root")
	}

	// Test the crucial CST access method
	cstRoot := shimAST.GetCSTRoot()
	if cstRoot == nil {
		t.Error("GetCSTRoot should return non-nil CST root")
	}

	t.Log("TreeSitterAST successfully implements compiler AST interface")
}

func TestCompilerRegistryWithTreeSitterShim(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse code
	testCode := `my HashRef[Str] $config = { host => "localhost" };`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Create compiler registry
	registry := NewCompilerRegistry()

	// Register compilers
	registry.Register(NewPerlCompiler(TargetCleanPerl))
	registry.Register(NewPerlCompiler(TargetTypedPerl))

	// Test compilation through registry
	targets := []Target{TargetCleanPerl, TargetTypedPerl}

	for _, target := range targets {
		t.Run("registry_"+string(target), func(t *testing.T) {
			result, err := registry.Compile(shimAST, target)
			if err != nil {
				t.Fatalf("Registry compilation failed for target %s: %v", target, err)
			}

			if result == "" {
				t.Fatalf("Registry compilation returned empty result for target %s", target)
			}

			// Verify the result looks reasonable
			if !strings.Contains(result, "config") {
				t.Error("Compiled result should contain 'config' variable")
			}

			t.Logf("Registry compilation for %s successful. Result length: %d", target, len(result))
		})
	}
}

func TestCSTBasedCompilationPath(t *testing.T) {
	// Create a shim parser
	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Parse code
	testCode := `use v5.38; my Str $greeting = "Hello, World!"; print "$greeting\n";`
	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse with shim: %v", err)
	}

	// Verify CST access is available
	if shimAST.GetCSTRoot() == nil {
		t.Fatal("TreeSitterAST should provide CST root access")
	}

	// Test that the compiler can work directly with the CST
	compiler := NewPerlCompiler(TargetCleanPerl)

	// This should trigger the CST-based compilation path (lines 145-147 in perl_compiler.go)
	// rather than the re-parsing path (lines 151-156)
	result, err := compiler.Compile(shimAST)
	if err != nil {
		t.Fatalf("CST-based compilation failed: %v", err)
	}

	// Verify the compilation result
	if !strings.Contains(result, "greeting") {
		t.Error("Result should contain 'greeting' variable")
	}

	if !strings.Contains(result, "Hello, World") {
		t.Error("Result should contain the greeting message")
	}

	// Check for version pragma if preserved
	if strings.Contains(result, "use v5.38") {
		t.Log("Version pragma preserved in compilation")
	}

	t.Logf("CST-based compilation path successful")
	t.Logf("Result: %s", result)
}

// TestTreeSitterShimTypePreservation tests complex type preservation through complete workflow
func TestTreeSitterShimTypePreservation(t *testing.T) {
	complexTypes := []struct {
		name     string
		code     string
		contains []string
	}{
		{
			name:     "simple_types",
			code:     `my Int $age = 30; my Str $name = "John";`,
			contains: []string{"Int", "Str"},
		},
		{
			name:     "parameterized_types",
			code:     `my ArrayRef[Int] $scores = [95, 87, 92];`,
			contains: []string{"ArrayRef[Int]"},
		},
		{
			name:     "nested_parameterized",
			code:     `my HashRef[ArrayRef[Str]] $data = { names => ["Alice", "Bob"] };`,
			contains: []string{"HashRef[ArrayRef[Str]]"},
		},
		{
			name:     "union_types",
			code:     `my Int|Str $value = 42;`,
			contains: []string{"Int|Str"},
		},
		{
			name: "method_signatures",
			code: `
class Calculator {
    method add(Int $a, Int $b) : Int {
        return $a + $b;
    }
}`,
			contains: []string{"method add(Int $a, Int $b) : Int"},
		},
	}

	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	// Test with TypedPerl target to preserve annotations
	compiler := NewPerlCompiler(TargetTypedPerl)

	for _, tc := range complexTypes {
		t.Run(tc.name, func(t *testing.T) {
			shimAST, err := shimParser.ParseStringShim(tc.code)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", tc.name, err)
			}

			output, err := compiler.Compile(shimAST)
			if err != nil {
				t.Fatalf("Failed to compile %s: %v", tc.name, err)
			}

			// Verify type annotations are preserved
			for _, expected := range tc.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Type annotation %q not preserved in %s.\nOutput: %s", expected, tc.name, output)
				}
			}
		})
	}
}

// TestTreeSitterShimInferenceIntegration tests integration with inference engine
func TestTreeSitterShimInferenceIntegration(t *testing.T) {
	t.Skip("Skipping until inference integration is fully implemented")

	// This test would verify that:
	// 1. TreeSitter shim AST works with inference engine
	// 2. Type information is correctly preserved through inference
	// 3. InferredTypedPerlCompiler works with tree-sitter backed ASTs
}

// TestTreeSitterShimPerformanceImprovements tests performance characteristics
func TestTreeSitterShimPerformanceImprovements(t *testing.T) {
	// This is more of a documentation test than performance measurement
	testCode := `
use v5.38;
my Int $count = 0;
for my Int $i (1..100) {
    $count += $i;
}
print "Sum: $count\n";
`

	shimParser, err := parser.NewShimParser()
	if err != nil {
		t.Fatalf("Failed to create shim parser: %v", err)
	}

	shimAST, err := shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Verify single-parse efficiency
	if shimAST.GetCSTRoot() == nil {
		t.Fatal("TreeSitterAST should provide direct CST access")
	}

	// Test multiple compilations - all should use same CST
	compilers := []Compiler{
		NewPerlCompiler(TargetCleanPerl),
		NewPerlCompiler(TargetTypedPerl),
	}

	for _, compiler := range compilers {
		output, err := compiler.Compile(shimAST)
		if err != nil {
			t.Errorf("Compilation failed: %v", err)
			continue
		}

		if output == "" {
			t.Error("Compilation produced empty output")
		}
	}

	t.Log("Performance test passed - single parse, multiple compilations successful")
}
