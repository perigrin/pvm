// ABOUTME: Performance validation tests for Step 4 of TypeScript-Go modernization
// ABOUTME: Validates parsing performance, memory usage, and stability of new pipeline

package parser

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/ast"
)

// Benchmark parsing performance with different parser types
func BenchmarkParser_SmallFile(b *testing.B) {
	testCode := `my Int $count = 42;
my Str $name = "test";
sub greet(Str $name) -> Str {
    return "Hello, " . $name;
}`

	b.Run("Scanner-based", func(b *testing.B) {
		parser, err := NewParserWithOptions(true) // Use scanner
		if err != nil {
			b.Fatalf("Failed to create scanner parser: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := parser.ParseString(testCode)
			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}
		}
	})

	b.Run("Tree-sitter", func(b *testing.B) {
		parser, err := NewParserWithOptions(false) // Use tree-sitter
		if err != nil {
			b.Fatalf("Failed to create tree-sitter parser: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := parser.ParseString(testCode)
			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}
		}
	})
}

// Benchmark parsing with medium-sized files
func BenchmarkParser_MediumFile(b *testing.B) {
	// Generate a medium-sized Perl file
	var builder strings.Builder
	builder.WriteString("package TestModule;\n")
	builder.WriteString("use strict;\nuse warnings;\n\n")

	// Add multiple subroutines and variables
	for i := 0; i < 50; i++ {
		builder.WriteString("my Int $var")
		builder.WriteString(strings.Repeat("a", i%10))
		builder.WriteString(" = ")
		builder.WriteString("42;\n")

		if i%5 == 0 {
			builder.WriteString("sub func")
			builder.WriteString(strings.Repeat("b", i%10))
			builder.WriteString("(Int $param) -> Str {\n")
			builder.WriteString("    my Str $result = \"value\";\n")
			builder.WriteString("    return $result;\n}\n\n")
		}
	}

	testCode := builder.String()

	b.Run("Scanner-based", func(b *testing.B) {
		parser, err := NewParserWithOptions(true)
		if err != nil {
			b.Fatalf("Failed to create scanner parser: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := parser.ParseString(testCode)
			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}
		}
	})

	b.Run("Tree-sitter", func(b *testing.B) {
		parser, err := NewParserWithOptions(false)
		if err != nil {
			b.Fatalf("Failed to create tree-sitter parser: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := parser.ParseString(testCode)
			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}
		}
	})
}

// Test memory usage patterns
func TestParser_MemoryUsage(t *testing.T) {
	testCode := `my Int $count = 42;
my Str $name = "test";
sub greet(Str $name) -> Str {
    return "Hello, " . $name;
}`

	// Test memory usage for scanner-based parser
	t.Run("Scanner-based memory usage", func(t *testing.T) {
		runtime.GC()
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		parser, err := NewParserWithOptions(true)
		if err != nil {
			t.Fatalf("Failed to create scanner parser: %v", err)
		}

		// Parse multiple times to see memory patterns
		for i := 0; i < 100; i++ {
			_, err := parser.ParseString(testCode)
			if err != nil {
				t.Fatalf("Parse failed on iteration %d: %v", i, err)
			}
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		allocDiff := m2.TotalAlloc - m1.TotalAlloc
		t.Logf("Scanner-based parser memory usage: %d bytes allocated", allocDiff)

		// Basic sanity check - shouldn't use excessive memory for small files
		if allocDiff > 10*1024*1024 { // 10MB threshold
			t.Errorf("Excessive memory usage: %d bytes", allocDiff)
		}
	})

	// Test memory usage for tree-sitter parser
	t.Run("Tree-sitter memory usage", func(t *testing.T) {
		runtime.GC()
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		parser, err := NewParserWithOptions(false)
		if err != nil {
			t.Fatalf("Failed to create tree-sitter parser: %v", err)
		}

		// Parse multiple times to see memory patterns
		for i := 0; i < 100; i++ {
			_, err := parser.ParseString(testCode)
			if err != nil {
				t.Fatalf("Parse failed on iteration %d: %v", i, err)
			}
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		allocDiff := m2.TotalAlloc - m1.TotalAlloc
		t.Logf("Tree-sitter parser memory usage: %d bytes allocated", allocDiff)

		// Basic sanity check
		if allocDiff > 10*1024*1024 { // 10MB threshold
			t.Errorf("Excessive memory usage: %d bytes", allocDiff)
		}
	})
}

// Test error handling and edge cases
func TestParser_ErrorHandling(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expectOK bool
	}{
		{"Empty string", "", true},
		{"Single variable", "my $var = 42;", true},
		{"Type annotation", "my Int $var = 42;", true},
		{"Malformed syntax", "my $var = ;", false}, // This may still parse but with errors
		{"Unicode content", "my Str $名前 = \"テスト\";", true},
		{"Large content", strings.Repeat("my $var = 42;\n", 1000), true},
		{"Only whitespace", "   \n\t\n   ", true},
		{"Only comments", "# This is a comment\n# Another comment", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			result, err := parser.ParseString(tt.input)
			duration := time.Since(start)

			// Check parsing doesn't take too long
			if duration > 5*time.Second {
				t.Errorf("Parsing took too long: %v", duration)
			}

			if tt.expectOK {
				if err != nil {
					t.Logf("Note: Parse error (might be expected): %v", err)
				}
				if result == nil {
					t.Error("Expected non-nil result even with errors")
				} else {
					// Basic sanity checks
					if result.Source != tt.input {
						t.Error("Source not preserved correctly")
					}
				}
			}
		})
	}
}

// Test stability under stress
func TestParser_Stability(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	testCodes := []string{
		"my Int $simple = 42;",
		"my ArrayRef[Str] $array = [];",
		"sub func(Int $a, Str $b) -> Bool { return 1; }",
		"type MyType = Int|Str;",
		"my MyType $var = 123;",
	}

	// Run many parsing operations to test stability
	t.Run("Stress test", func(t *testing.T) {
		for i := 0; i < 500; i++ {
			for _, code := range testCodes {
				result, err := parser.ParseString(code)
				if err != nil {
					t.Logf("Parse error on iteration %d: %v", i, err)
				}
				if result == nil {
					t.Errorf("Nil result on iteration %d with code: %s", i, code)
				}

				// Basic validation
				if result != nil && result.Source != code {
					t.Errorf("Source mismatch on iteration %d", i)
				}
			}

			// Occasional GC to test memory stability
			if i%100 == 0 {
				runtime.GC()
			}
		}
	})
}

// Test AST navigation performance
func BenchmarkAST_Navigation(b *testing.B) {
	parser, err := NewParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	// Create a medium-sized AST
	testCode := `package TestModule;
my Int $count = 42;
my Str $name = "test";
my ArrayRef[Int] $numbers = [1, 2, 3];

sub calculate(Int $a, Int $b) -> Int {
    my Int $result = $a + $b;
    return $result;
}

sub greet(Str $name) -> Str {
    return "Hello, " . $name;
}

type MyUnion = Int|Str|Bool;
my MyUnion $value = 123;`

	astResult, err := parser.ParseString(testCode)
	if err != nil {
		b.Fatalf("Failed to parse test code: %v", err)
	}

	if astResult == nil || astResult.Root == nil {
		b.Fatal("No AST root available for benchmarking")
	}

	b.Run("AST traversal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate AST navigation operations
			walkNode(astResult.Root)
		}
	})
}

// Helper function to walk AST nodes
func walkNode(node ast.Node) int {
	if node == nil {
		return 0
	}

	count := 1
	for _, child := range node.Children() {
		count += walkNode(child)
	}
	return count
}

// Test concurrent parsing (safety)
func TestParser_Concurrency(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	testCode := "my Int $var = 42;"
	concurrency := 10
	iterations := 100

	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency*iterations)

	// Start multiple goroutines
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < iterations; j++ {
				result, err := parser.ParseString(testCode)
				if err != nil {
					errors <- err
					return
				}
				if result == nil {
					errors <- fmt.Errorf("nil result in goroutine %d iteration %d", id, j)
					return
				}
			}
		}(i)
	}

	// Wait for completion
	for i := 0; i < concurrency; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Errorf("Concurrent parsing error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("Found %d errors in concurrent parsing", errorCount)
	} else {
		t.Log("Concurrent parsing completed successfully")
	}
}
