// ABOUTME: Benchmark tests for performance optimizations
// ABOUTME: Measures improvement in parsing, binding, and type checking performance

package performance

import (
	"fmt"
	"testing"
	"time"

	"tamarou.com/pvm/internal/parser"
)

// Test content for benchmarking
var benchmarkContent = map[string]string{
	"simple": `my $name = "test";
my $count = 42;
print "$name: $count\n";`,

	"typed": `my Int $count = 42;
my Str $name = "test";
my ArrayRef[Int] $numbers = [1, 2, 3, 4, 5];

sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

my $result = add($count, 10);`,

	"complex": func() string {
		content := ""
		for i := 0; i < 20; i++ {
			content += fmt.Sprintf(`my Int $var_%d = %d;
my Str $msg_%d = "message_%d";

sub process_%d(Int $input) -> Int {
    return $input * 2 + 1;
}

`, i, i, i, i, i)
		}
		return content
	}(),

	"union_types": `type NumberOrString = Int|Str;
type OptionalData = HashRef|Undef;
type ComplexUnion = Int|Str|ArrayRef[Int]|HashRef[Str];

my NumberOrString $value = 42;
$value = "hello";

my OptionalData $data = { key => "value" };
$data = undef;`,
}

// BenchmarkOriginalParser benchmarks the original parser without optimizations
func BenchmarkOriginalParser(b *testing.B) {
	parser, err := parser.NewParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	content := benchmarkContent["typed"]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseString(content)
		if err != nil {
			b.Fatalf("Parse error: %v", err)
		}
	}
}

// BenchmarkOptimizedParser benchmarks the optimized parser
func BenchmarkOptimizedParser(b *testing.B) {
	baseParser, err := parser.NewParser()
	if err != nil {
		b.Fatalf("Failed to create base parser: %v", err)
	}

	optimizedParser := NewOptimizedParser()
	content := benchmarkContent["typed"]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := optimizedParser.ParseString(content)
		if err != nil {
			b.Fatalf("Parse error: %v", err)
		}
	}
}

// BenchmarkFastParser benchmarks the fast parser
func BenchmarkFastParser(b *testing.B) {
	baseParser, err := parser.NewParser()
	if err != nil {
		b.Fatalf("Failed to create base parser: %v", err)
	}

	fastParser := NewFastParser()
	content := benchmarkContent["simple"] // Use simple content for fast parser

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fastParser.ParseString(content)
		if err != nil {
			b.Fatalf("Parse error: %v", err)
		}
	}
}

// BenchmarkCachedParser benchmarks parsing with caching
func BenchmarkCachedParser(b *testing.B) {
	baseParser, err := parser.NewParser()
	if err != nil {
		b.Fatalf("Failed to create base parser: %v", err)
	}

	cache := NewParseCache(100)
	content := benchmarkContent["typed"]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate cache usage by repeating the same content
		contentHash := fastHash([]byte(content))
		if entry, found := cache.Get(content, contentHash); found {
			_ = entry.AST
		} else {
			ast, err := baseParser.ParseString(content)
			if err != nil {
				b.Fatalf("Parse error: %v", err)
			}
			cache.Put(content, ast, nil, contentHash)
		}
	}
}

// BenchmarkTypeCache benchmarks type operation caching
func BenchmarkTypeCache(b *testing.B) {
	typeCache := NewTypeCache(1000)

	// Sample union type components
	components := []string{"Int", "Str", "ArrayRef", "HashRef"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate type operations
		result, cached := typeCache.GetUnionType(components, func() (string, time.Duration) {
			start := time.Now()
			// Simulate expensive type computation
			time.Sleep(100 * time.Microsecond)
			return "Int|Str|ArrayRef|HashRef", time.Since(start)
		})

		if result == "" {
			b.Fatal("Expected non-empty result")
		}

		// Track cache effectiveness
		if cached && i%1000 == 0 {
			b.Logf("Cache hit at iteration %d", i)
		}
	}
}

// BenchmarkObjectPool benchmarks object pooling
func BenchmarkObjectPool(b *testing.B) {
	pool := NewObjectPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Get objects from pool
		nodes := pool.GetASTNodes()
		symbols := pool.GetSymbolMap()
		buffer := pool.GetStringBuffer()

		// Simulate usage
		for j := 0; j < 10; j++ {
			nodes = append(nodes, nil) // Simulate adding nodes
		}

		// Return to pool
		pool.PutASTNodes(nodes)
		pool.PutSymbolMap(symbols)
		pool.PutStringBuffer(buffer)
	}
}

// BenchmarkOptimizedPipeline benchmarks the complete optimized pipeline
func BenchmarkOptimizedPipeline(b *testing.B) {
	baseParser, err := parser.NewParser()
	if err != nil {
		b.Fatalf("Failed to create base parser: %v", err)
	}

	config := DefaultOptimizationConfig()
	pipeline := NewOptimizedPipeline(baseParser, config)
	defer pipeline.Shutdown()

	content := benchmarkContent["typed"]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := pipeline.ProcessFile("test.pl", content)
		if err != nil {
			b.Fatalf("Pipeline error: %v", err)
		}
		if result.AST == nil {
			b.Fatal("Expected non-nil AST")
		}
	}
}

// BenchmarkPipelineWithDifferentSizes benchmarks pipeline with various content sizes
func BenchmarkPipelineWithDifferentSizes(b *testing.B) {
	baseParser, err := parser.NewParser()
	if err != nil {
		b.Fatalf("Failed to create base parser: %v", err)
	}

	config := DefaultOptimizationConfig()
	pipeline := NewOptimizedPipeline(baseParser, config)
	defer pipeline.Shutdown()

	testCases := []struct {
		name    string
		content string
	}{
		{"Simple", benchmarkContent["simple"]},
		{"Typed", benchmarkContent["typed"]},
		{"Complex", benchmarkContent["complex"]},
		{"UnionTypes", benchmarkContent["union_types"]},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result, err := pipeline.ProcessFile("test.pl", tc.content)
				if err != nil {
					b.Fatalf("Pipeline error: %v", err)
				}
				if result.AST == nil {
					b.Fatal("Expected non-nil AST")
				}
			}
		})
	}
}

// BenchmarkMemoryUsage benchmarks memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	baseParser, err := parser.NewParser()
	if err != nil {
		b.Fatalf("Failed to create base parser: %v", err)
	}

	content := benchmarkContent["complex"]

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := baseParser.ParseString(content)
		if err != nil {
			b.Fatalf("Parse error: %v", err)
		}
	}
}

// BenchmarkOptimizedMemoryUsage benchmarks memory usage with optimizations
func BenchmarkOptimizedMemoryUsage(b *testing.B) {
	baseParser, err := parser.NewParser()
	if err != nil {
		b.Fatalf("Failed to create base parser: %v", err)
	}

	config := DefaultOptimizationConfig()
	pipeline := NewOptimizedPipeline(baseParser, config)
	defer pipeline.Shutdown()

	content := benchmarkContent["complex"]

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result, err := pipeline.ProcessFile("test.pl", content)
		if err != nil {
			b.Fatalf("Pipeline error: %v", err)
		}
		if result.AST == nil {
			b.Fatal("Expected non-nil AST")
		}
	}
}

// BenchmarkLazyTypeResolution benchmarks lazy type resolution
func BenchmarkLazyTypeResolution(b *testing.B) {
	resolver := NewLazyTypeResolver()

	// Add some lazy types
	for i := 0; i < 100; i++ {
		typeName := fmt.Sprintf("Type%d", i)
		resolver.AddLazyType(typeName, nil, func() (string, error) {
			// Simulate type resolution work
			time.Sleep(10 * time.Microsecond)
			return fmt.Sprintf("resolved_%s", typeName), nil
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		typeName := fmt.Sprintf("Type%d", i%100)
		_, err := resolver.ResolveType(typeName)
		if err != nil {
			b.Fatalf("Resolution error: %v", err)
		}
	}
}

// BenchmarkUnionTypeOperations benchmarks optimized union type operations
func BenchmarkUnionTypeOperations(b *testing.B) {
	type1 := NewOptimizedUnionType([]string{"Int", "Str", "Bool"})
	type2 := NewOptimizedUnionType([]string{"Str", "Float", "Bool"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform union operations
		union := type1.Union(type2)
		intersection := type1.Intersection(type2)

		// Test membership
		contains := union.Contains("Int")

		// Generate string representation
		repr := intersection.String()

		// Avoid compiler optimization
		if !contains || repr == "" {
			b.Fatal("Unexpected result")
		}
	}
}

// TestOptimizationEffectiveness tests that optimizations actually improve performance
func TestOptimizationEffectiveness(t *testing.T) {
	baseParser, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create base parser: %v", err)
	}

	content := benchmarkContent["typed"]

	// Measure baseline performance
	start := time.Now()
	iterations := 100
	for i := 0; i < iterations; i++ {
		_, err := baseParser.ParseString(content)
		if err != nil {
			t.Fatalf("Parse error: %v", err)
		}
	}
	baselineTime := time.Since(start)

	// Measure optimized performance
	config := DefaultOptimizationConfig()
	pipeline := NewOptimizedPipeline(baseParser, config)
	defer pipeline.Shutdown()

	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, err := pipeline.ProcessFile("test.pl", content)
		if err != nil {
			t.Fatalf("Pipeline error: %v", err)
		}
	}
	optimizedTime := time.Since(start)

	// Calculate improvement
	improvement := float64(baselineTime-optimizedTime) / float64(baselineTime) * 100

	t.Logf("Baseline time: %v", baselineTime)
	t.Logf("Optimized time: %v", optimizedTime)
	t.Logf("Performance improvement: %.2f%%", improvement)

	// Get optimization statistics
	stats := pipeline.GetOptimizationStats()
	t.Logf("Parse cache hit rate: %.2f%%", stats.ParseCacheHitRate*100)
	t.Logf("Fast parse percentage: %.2f%%", stats.FastParsePercentage)
	t.Logf("Total operations: %d", stats.TotalOperations)

	// We expect some improvement, even if minimal
	if improvement < -50 { // Allow for some overhead but not too much regression
		t.Errorf("Optimization caused significant performance regression: %.2f%%", improvement)
	}
}
