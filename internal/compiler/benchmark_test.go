// ABOUTME: Performance benchmarks for unified compiler architecture
// ABOUTME: Tests compilation speed, memory usage, and caching effectiveness

package compiler

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"
)

// BenchmarkUnifiedCompilerBasic benchmarks basic compilation performance
func BenchmarkUnifiedCompilerBasic(b *testing.B) {
	testCode := `my Int $count = 42;
print "Count: $count\n";`

	compiler := NewCleanPerlCompilerUnified()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compiler.CompileString(testCode)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

// BenchmarkUnifiedCompilerWithCaching benchmarks compilation with caching
func BenchmarkUnifiedCompilerWithCaching(b *testing.B) {
	testCode := `my Int $count = 42;
print "Count: $count\n";`

	compiler := NewCachingCleanPerlCompiler(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compiler.CompileString(testCode)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}

	// Report cache effectiveness
	stats := compiler.GetCacheStats()
	b.Logf("Cache hit ratio: %.2f%%", stats.HitRatio*100)
}

// BenchmarkCompilerComparison compares unified vs legacy compiler performance
func BenchmarkCompilerComparison(b *testing.B) {
	testCode := `my Int $count = 42;
my Str $name = "test";
print "Count: $count, Name: $name\n";`

	b.Run("Unified", func(b *testing.B) {
		compiler := NewCleanPerlCompilerUnified()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := compiler.CompileString(testCode)
			if err != nil {
				b.Fatalf("Compilation failed: %v", err)
			}
		}
	})

	b.Run("Legacy", func(b *testing.B) {
		compiler := NewCleanPerlCompilerUnified()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create temporary CST-based AST for comparison
			cstAST, err := NewCSTBasedAST("", testCode)
			if err != nil {
				b.Fatalf("Failed to create AST: %v", err)
			}
			_, err = compiler.Compile(cstAST)
			if err != nil {
				b.Fatalf("Compilation failed: %v", err)
			}
		}
	})
}

// BenchmarkLargeFileCompilation benchmarks compilation of large files
func BenchmarkLargeFileCompilation(b *testing.B) {
	// Generate a large Perl file with many typed variables
	var builder strings.Builder
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(&builder, "my Int $var%d = %d;\n", i, i)
		fmt.Fprintf(&builder, "my Str $name%d = \"name%d\";\n", i, i)
		fmt.Fprintf(&builder, "print \"Var: $var%d, Name: $name%d\\n\";\n", i, i)
	}
	largeCode := builder.String()

	b.Run("WithoutCaching", func(b *testing.B) {
		compiler := NewCleanPerlCompilerUnified()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := compiler.CompileString(largeCode)
			if err != nil {
				b.Fatalf("Compilation failed: %v", err)
			}
		}
	})

	b.Run("WithCaching", func(b *testing.B) {
		compiler := NewCachingCleanPerlCompiler(10)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := compiler.CompileString(largeCode)
			if err != nil {
				b.Fatalf("Compilation failed: %v", err)
			}
		}

		stats := compiler.GetCacheStats()
		b.Logf("Cache hit ratio: %.2f%%", stats.HitRatio*100)
	})
}

// BenchmarkMemoryUsage benchmarks memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	testCode := `my Int $count = 42;
my ArrayRef[Str] $names = ["alice", "bob"];
my HashRef[Int] $scores = {alice => 95, bob => 87};`

	compiler := NewCleanPerlCompilerUnified()

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compiler.CompileString(testCode)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&after)

	allocPerOp := (after.TotalAlloc - before.TotalAlloc) / uint64(b.N)
	b.ReportMetric(float64(allocPerOp), "bytes/op")
}

// BenchmarkCacheEffectiveness tests cache effectiveness with varying patterns
func BenchmarkCacheEffectiveness(b *testing.B) {
	testCodes := []string{
		`my Int $a = 1;`,
		`my Str $b = "test";`,
		`my ArrayRef[Int] $c = [1, 2, 3];`,
		`my Int $a = 1;`,      // Repeat for cache hit
		`my Str $b = "test";`, // Repeat for cache hit
	}

	compiler := NewCachingCleanPerlCompiler(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		code := testCodes[i%len(testCodes)]
		_, err := compiler.CompileString(code)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}

	stats := compiler.GetCacheStats()
	b.Logf("Cache hits: %d, misses: %d, ratio: %.2f%%",
		stats.HitCount, stats.MissCount, stats.HitRatio*100)
}

// BenchmarkParallelCompilation benchmarks concurrent compilation
func BenchmarkParallelCompilation(b *testing.B) {
	testCode := `my Int $count = 42;
my Str $name = "concurrent";
print "Count: $count, Name: $name\n";`

	compiler := NewCachingCleanPerlCompiler(100)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := compiler.CompileString(testCode)
			if err != nil {
				b.Fatalf("Compilation failed: %v", err)
			}
		}
	})

	stats := compiler.GetCacheStats()
	b.Logf("Parallel cache hit ratio: %.2f%%", stats.HitRatio*100)
}

// BenchmarkRegistryPerformance benchmarks the optimized compiler registry
func BenchmarkRegistryPerformance(b *testing.B) {
	registry := NewOptimizedCompilerRegistry(100)
	testCode := `my Int $value = 123;`

	b.Run("CleanPerl", func(b *testing.B) {
		cstAST, err := NewCSTBasedAST("", testCode)
		if err != nil {
			b.Fatalf("Failed to create AST: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := registry.CompileOptimized(cstAST, TargetCleanPerl)
			if err != nil {
				b.Fatalf("Compilation failed: %v", err)
			}
		}
	})

	b.Run("TypedPerl", func(b *testing.B) {
		cstAST, err := NewCSTBasedAST("", testCode)
		if err != nil {
			b.Fatalf("Failed to create AST: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := registry.CompileOptimized(cstAST, TargetTypedPerl)
			if err != nil {
				b.Fatalf("Compilation failed: %v", err)
			}
		}
	})

	// Report aggregated statistics
	stats := registry.GetAggregatedStats()
	b.Logf("Aggregated cache hit ratio: %.2f%%", stats.CacheHitRatio*100)
	b.Logf("Average compilation time: %v", stats.AverageTime)
}

// BenchmarkTypeTransformation benchmarks specific type transformation patterns
func BenchmarkTypeTransformation(b *testing.B) {
	testCases := map[string]string{
		"SimpleTypes": `my Int $a = 1;
my Str $b = "test";
my Bool $c = 1;`,

		"ComplexTypes": `my ArrayRef[HashRef[Str]] $complex = [{name => "test"}];
my Union[Int, Str, Bool] $union = 42;
my Maybe[Object] $optional = undef;`,

		"TypeAssertions": `my $value = get_value();
my $typed = $value as Int;
my $converted = convert($data) as ArrayRef[Str];`,

		"MethodSignatures": `method Int calculate(Int $a, Int $b) {
    return $a + $b;
}
method HashRef[Int] process(ArrayRef[Str] $items) {
    return {};
}`,
	}

	for name, code := range testCases {
		b.Run(name, func(b *testing.B) {
			compiler := NewCleanPerlCompilerUnified()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := compiler.CompileString(code)
				if err != nil {
					b.Fatalf("Compilation failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkCacheEviction benchmarks cache eviction performance
func BenchmarkCacheEviction(b *testing.B) {
	// Use a small cache to force evictions
	compiler := NewCachingCleanPerlCompiler(5)

	// Generate many different code samples to exceed cache size
	codes := make([]string, 20)
	for i := 0; i < 20; i++ {
		codes[i] = fmt.Sprintf("my Int $var%d = %d;", i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		code := codes[i%len(codes)]
		_, err := compiler.CompileString(code)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}

	stats := compiler.GetCacheStats()
	b.Logf("Evictions: %d, final cache size: %d",
		stats.EvictCount, stats.CleanSize+stats.TypedSize)
}

// TestPerformanceRegression tests for performance regressions
func TestPerformanceRegression(t *testing.T) {
	testCode := `my Int $count = 42;
my Str $name = "test";
print "Count: $count, Name: $name\n";`

	// Baseline measurement
	compiler := NewCleanPerlCompilerUnified()

	start := time.Now()
	iterations := 1000
	for i := 0; i < iterations; i++ {
		_, err := compiler.CompileString(testCode)
		if err != nil {
			t.Fatalf("Compilation failed: %v", err)
		}
	}
	duration := time.Since(start)

	avgTime := duration / time.Duration(iterations)

	// Performance expectations (adjust based on your requirements)
	maxAvgTime := 10 * time.Millisecond // Maximum acceptable average compilation time

	if avgTime > maxAvgTime {
		t.Errorf("Performance regression detected: average compilation time %v exceeds maximum %v",
			avgTime, maxAvgTime)
	}

	t.Logf("Performance test passed: average compilation time %v", avgTime)
}

// TestMemoryLeaks tests for memory leaks in compilation
func TestMemoryLeaks(t *testing.T) {
	testCode := `my Int $count = 42;`

	compiler := NewCachingCleanPerlCompiler(100)

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	// Perform many compilations
	for i := 0; i < 10000; i++ {
		_, err := compiler.CompileString(testCode)
		if err != nil {
			t.Fatalf("Compilation failed: %v", err)
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&after)

	// Check memory growth (handle potential underflow if GC reduced memory)
	var memGrowth uint64
	if after.Alloc >= before.Alloc {
		memGrowth = after.Alloc - before.Alloc
	} else {
		// Memory actually decreased due to GC - this is good, not a leak
		t.Logf("Memory test passed: memory decreased by %d bytes (GC occurred)", before.Alloc-after.Alloc)
		return
	}

	maxGrowth := uint64(10 * 1024 * 1024) // 10MB maximum growth

	if memGrowth > maxGrowth {
		t.Errorf("Possible memory leak detected: memory grew by %d bytes (max allowed: %d)",
			memGrowth, maxGrowth)
	}

	t.Logf("Memory test passed: growth %d bytes", memGrowth)
}
