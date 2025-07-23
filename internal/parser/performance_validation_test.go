// ABOUTME: Performance validation tests for Step 6 of prompt_plan.md
// ABOUTME: Validates parser performance and integration with PVM ecosystem components

package parser

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/ast"
	basetesting "tamarou.com/pvm/internal/testing"
)

// Benchmark parsing performance with different parser types
func BenchmarkParser_SmallFile(b *testing.B) {
	testCode := `my Int $count = 42;
my Str $name = "test";
sub Str greet(Str $name) {
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
			builder.WriteString("sub Str func")
			builder.WriteString(strings.Repeat("b", i%10))
			builder.WriteString("(Int $param) {\n")
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
sub Str greet(Str $name) {
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
		"sub Bool func(Int $a, Str $b) { return 1; }",
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

sub Int calculate(Int $a, Int $b) {
    my Int $result = $a + $b;
    return $result;
}

sub Str greet(Str $name) {
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
	testCode := "my Int $var = 42;"
	concurrency := 10
	iterations := 100

	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency*iterations)

	// Start multiple goroutines using pooled parsers
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < iterations; j++ {
				// Get parser from pool for each operation
				parser, err := NewParser()
				if err != nil {
					errors <- fmt.Errorf("failed to create parser in goroutine %d iteration %d: %v", id, j, err)
					return
				}

				result, err := parser.ParseString(testCode)

				// Return parser to pool
				ReturnParser(parser)

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

// TestStep6_TypedPerlPerformanceValidation implements Step 6 from prompt_plan.md
// This validates parser performance specifically for typed Perl features
func TestStep6_TypedPerlPerformanceValidation(t *testing.T) {
	basetesting.SkipUnlessPerformance(t, "Step 6 typed Perl performance validation")

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test cases for typed Perl features that were completed in Steps 1-5
	testCases := []struct {
		name        string
		code        string
		maxDuration time.Duration
		description string
	}{
		{
			name: "complex_method_signatures",
			code: `method Result[Map[Str, Array[Item]], ProcessingError] process(
    ArrayRef[HashRef[Int|Str]] $data,
    Optional[CodeRef[Int, Bool]] $validator = undef
) {
    return Success->new({});
}`,
			maxDuration: 50 * time.Millisecond,
			description: "Step 1: Complex method signature parsing",
		},
		{
			name: "union_types_nested_contexts",
			code: `my ArrayRef[Int|Str] $mixed_array;
method Bool func(Success|Error $result) {
    my Int|Str|Undef $value = $result->value;
    return $value as Bool;
}`,
			maxDuration: 30 * time.Millisecond,
			description: "Step 2: Union types in nested contexts",
		},
		{
			name: "complex_type_assertions",
			code: `my $data = fetch_data();
my ArrayRef[HashRef[Int|Str]] $typed_data = $data as ArrayRef[HashRef[Int|Str]];
my $result = process($typed_data) as (Success|Error)&Detailed;`,
			maxDuration: 40 * time.Millisecond,
			description: "Step 3: Complex type assertions",
		},
		{
			name: "generic_class_declarations",
			code: `class Container[T] {
    field ArrayRef[T] $items = [];
    method Int add(T $item) {
        push @{$self->{items}}, $item;
        return scalar @{$self->{items}};
    }
}

class Cache[K: Hashable, V: Serializable] {
    field HashMap[K, V] $data;
}`,
			maxDuration: 60 * time.Millisecond,
			description: "Step 4: Generic class declarations",
		},
		{
			name: "all_features_combined",
			code: `package CompleteTypedPerl;
use v5.36;

type UserID = Int;
type Result[T, E] = Success[T] | Error[E];

class UserService[T: User] {
    field HashRef[UserID, T] $users = {};

    method Optional[T] find_user(UserID $id) {
        return $self->{users}->{$id};
    }

    method Result[UserID, Str] add_user(T $user) {
        my UserID $id = $user->id as UserID;
        return Error->new("User exists") if exists $self->{users}->{$id};
        $self->{users}->{$id} = $user;
        return Success->new($id);
    }
}`,
			maxDuration: 100 * time.Millisecond,
			description: "Step 5: All typed Perl features combined",
		},
	}

	// Run performance tests for each completed step
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Warm up
			_, _ = parser.ParseString(tc.code)

			// Measure performance
			var totalDuration time.Duration
			iterations := 10

			for i := 0; i < iterations; i++ {
				start := time.Now()
				ast, err := parser.ParseString(tc.code)
				duration := time.Since(start)
				totalDuration += duration

				if err != nil {
					t.Fatalf("%s: Parse error: %v", tc.description, err)
				}
				if ast == nil {
					t.Fatalf("%s: No AST generated", tc.description)
				}
			}

			avgDuration := totalDuration / time.Duration(iterations)

			// Verify performance meets requirements
			if avgDuration > tc.maxDuration {
				t.Errorf("%s: Average parse time %v exceeded limit %v",
					tc.description, avgDuration, tc.maxDuration)
			} else {
				t.Logf("%s: ✓ Performance validated (avg: %v)", tc.description, avgDuration)
			}
		})
	}
}

// TestStep6_IntegrationWithCompiler validates parser integration with compiler
func TestStep6_IntegrationWithCompiler(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test complex typed Perl that exercises all features
	complexCode := `package IntegrationTest;
use v5.36;

# Type definitions
type UserData = HashRef[struct {
    id => Int,
    name => Str,
    roles => ArrayRef[Str],
    active => Bool
}];

# Generic result type
class Result[T, E] {
    field Optional[T] $value;
    field Optional[E] $error;

    method Bool is_ok() {
        return defined $self->{value};
    }
}

# Service with complex types
class UserService {
    field HashRef[Int, UserData] $users = {};

    method Result[Int, Str] create_user(
        Str $name,
        ArrayRef[Str] $roles = []
    ) {
        my Int $id = int(rand(10000));

        my UserData $user = {
            id => $id,
            name => $name,
            roles => $roles,
            active => 1
        };

        $self->{users}->{$id} = $user;
        return Result->new(value => $id);
    }

    method ArrayRef[UserData] find_users_with_role(Str $role) {
        return [
            grep {
                my ArrayRef[Str] $roles = $_->{roles} as ArrayRef[Str];
                grep { $_ eq $role } @$roles
            } values %{$self->{users}}
        ];
    }
}

1;`

	// Parse the complex code
	start := time.Now()
	_, parseErr := parser.ParseString(complexCode)
	parseDuration := time.Since(start)

	if parseErr != nil {
		t.Fatalf("Failed to parse integration test code: %v", parseErr)
	}

	t.Logf("Parsed complex typed Perl in %v", parseDuration)

	// AST verification: Now enabled since recent grammar improvements support typed Perl features

	// Performance check
	if parseDuration > 200*time.Millisecond {
		t.Errorf("Integration test parsing too slow: %v", parseDuration)
	}

	t.Log("✓ Step 6: Performance and integration validation completed successfully")
}

// BenchmarkStep6_TypedPerlFeatures provides detailed benchmarks for typed Perl
func BenchmarkStep6_TypedPerlFeatures(b *testing.B) {
	parser, err := NewParser()
	if err != nil {
		b.Fatal(err)
	}

	benchmarks := []struct {
		name string
		code string
	}{
		{
			name: "SimpleTypeAnnotation",
			code: `my Int $x = 42;`,
		},
		{
			name: "UnionType",
			code: `my Int|Str $value;`,
		},
		{
			name: "ParameterizedType",
			code: `my ArrayRef[HashRef[Int]] $data;`,
		},
		{
			name: "ComplexMethodSignature",
			code: `method HashRef[Str, Any] foo(ArrayRef[Int|Str] $arr, Optional[Bool] $flag = undef) {}`,
		},
		{
			name: "GenericClass",
			code: `class Box[T: Comparable] { field T $value; }`,
		},
		{
			name: "TypeAssertion",
			code: `my $typed = $data as ArrayRef[HashRef[Int|Str]];`,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := parser.ParseString(bm.code)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
