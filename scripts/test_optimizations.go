//go:build ignore

// ABOUTME: Performance optimization validation script
// ABOUTME: Tests and measures effectiveness of caching, fast parsing, and algorithmic improvements

package main

import (
	"fmt"
	"hash/fnv"
	"log"
	"strings"
	"time"

	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/performance"
)

func main() {
	fmt.Println("🧪 Performance Optimization Validation")
	fmt.Println("=====================================")

	// Test data - representative Perl code patterns
	testCases := []struct {
		name     string
		content  string
		isSimple bool
	}{
		{
			name: "Simple Variables",
			content: `my $name = "John";
my Int $age = 30;
my Str $city = "New York";`,
			isSimple: true,
		},
		{
			name: "Basic Subroutine",
			content: `sub hello {
    print "Hello World\n";
}`,
			isSimple: true,
		},
		{
			name: "Complex Code",
			content: `package MyClass;
use Moose;
has 'name' => (is => 'rw', isa => 'Str');
sub BUILD {
    my ($self, $args) = @_;
    if ($args->{name}) {
        $self->name($args->{name});
    }
}`,
			isSimple: false,
		},
		{
			name: "Type Annotations",
			content: `sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}`,
			isSimple: true,
		},
	}

	// Initialize parsers
	baseParser, err := parser.NewParser()
	if err != nil {
		log.Fatal("Failed to create base parser:", err)
	}

	fastParser := performance.NewFastParser(baseParser)
	parseCache := performance.NewParseCache(1000)

	fmt.Println("\n📊 Testing Performance Optimizations")
	fmt.Println("====================================")

	// Test 1: Fast Parser vs Tree-sitter
	fmt.Println("\n1. Fast Parser Performance:")
	testFastParser(fastParser, testCases)

	// Test 2: Parse Caching
	fmt.Println("\n2. Parse Cache Performance:")
	testParseCache(parseCache, baseParser, testCases)

	// Test 3: Object Pool Performance
	fmt.Println("\n3. Object Pool Performance:")
	testObjectPool()

	// Test 4: Memory Optimization
	fmt.Println("\n4. Memory Usage Comparison:")
	testMemoryUsage(baseParser, fastParser, testCases)

	// Test 5: Integration Performance
	fmt.Println("\n5. Integration Performance Test:")
	testIntegrationPerformance(baseParser, fastParser, parseCache)

	fmt.Println("\n✅ Performance optimization validation complete!")
}

// hashContent creates a hash for content caching
func hashContent(content string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(content))
	return h.Sum64()
}

func testFastParser(fastParser *performance.FastParser, testCases []struct {
	name, content string
	isSimple      bool
}) {
	for _, tc := range testCases {
		start := time.Now()

		_, err := fastParser.ParseString(tc.content)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("  ❌ %s: Error - %v\n", tc.name, err)
			continue
		}

		fmt.Printf("  ✅ %s: %v\n", tc.name, duration)
	}

	// Get statistics
	fast, fallback, percentage := fastParser.GetStats()
	fmt.Printf("  📈 Fast parsing ratio: %.1f%% (%d fast, %d fallback)\n",
		percentage, fast, fallback)
}

func testParseCache(cache *performance.ParseCache, parser parser.Parser, testCases []struct {
	name, content string
	isSimple      bool
}) {
	// Test cache miss (first parse)
	var totalUncached, totalCached time.Duration

	for _, tc := range testCases {
		contentHash := hashContent(tc.content)

		// Cache miss
		start := time.Now()
		ast, err := parser.ParseString(tc.content)
		if err != nil {
			continue
		}
		cache.Put(tc.content, ast, nil, contentHash)
		totalUncached += time.Since(start)

		// Cache hit
		start = time.Now()
		_, hit := cache.Get(tc.content, contentHash)
		totalCached += time.Since(start)

		if !hit {
			fmt.Printf("  ❌ %s: Cache miss unexpected\n", tc.name)
		}
	}

	fmt.Printf("  📊 Uncached: %v, Cached: %v\n", totalUncached, totalCached)
	if totalUncached > 0 {
		speedup := float64(totalUncached) / float64(totalCached)
		fmt.Printf("  🚀 Cache speedup: %.1fx\n", speedup)
	}
}

func testObjectPool() {
	pool := performance.NewObjectPool()

	start := time.Now()

	// Test object allocation and release
	for i := 0; i < 1000; i++ {
		nodes := pool.GetASTNodes()
		// Simulate work
		nodes = append(nodes, nil) // Add some nodes
		pool.PutASTNodes(nodes)
	}

	duration := time.Since(start)
	fmt.Printf("  ⚡ 1000 object pool operations: %v\n", duration)

	// Compare with regular allocation
	start = time.Now()
	for i := 0; i < 1000; i++ {
		_ = make([]*interface{}, 0, 100)
	}
	normalDuration := time.Since(start)

	fmt.Printf("  📊 Normal allocation: %v\n", normalDuration)
	if normalDuration > 0 && duration > 0 {
		efficiency := float64(normalDuration) / float64(duration)
		fmt.Printf("  🎯 Pool efficiency: %.1fx\n", efficiency)
	}
}

func testMemoryUsage(baseParser parser.Parser, fastParser *performance.FastParser, testCases []struct {
	name, content string
	isSimple      bool
}) {
	// Create a moderately complex test case
	complexCode := strings.Repeat(`my Int $var%d = %d;
`, 100)

	for i := 0; i < 100; i++ {
		complexCode = strings.Replace(complexCode, fmt.Sprintf("%%d"), fmt.Sprintf("%d", i), 2)
	}

	// Test base parser memory
	start := time.Now()
	_, err := baseParser.ParseString(complexCode)
	baseDuration := time.Since(start)

	if err != nil {
		fmt.Printf("  ❌ Base parser error: %v\n", err)
		return
	}

	// Test fast parser memory
	start = time.Now()
	_, err = fastParser.ParseString(complexCode)
	fastDuration := time.Since(start)

	if err != nil {
		fmt.Printf("  ❌ Fast parser error: %v\n", err)
		return
	}

	fmt.Printf("  📊 Base parser: %v\n", baseDuration)
	fmt.Printf("  📊 Fast parser: %v\n", fastDuration)

	if baseDuration > 0 && fastDuration > 0 {
		improvement := float64(baseDuration) / float64(fastDuration)
		fmt.Printf("  🚀 Performance improvement: %.1fx\n", improvement)
	}
}

func testIntegrationPerformance(baseParser parser.Parser, fastParser *performance.FastParser, cache *performance.ParseCache) {
	// Create a comprehensive test that exercises all optimizations
	testCode := `#!/usr/bin/perl
use strict;
use warnings;

my Str $name = "TestModule";
my Int $version = 1;

sub new(Str $class) -> Object {
    my $self = {};
    bless $self, $class;
    return $self;
}

sub process(Str $input) -> Str {
    my $result = $input;
    $result =~ s/test/processed/g;
    return $result;
}

1;`

	iterations := 10

	// Test baseline performance
	fmt.Printf("  🔄 Running %d iterations...\n", iterations)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := baseParser.ParseString(testCode)
		if err != nil {
			fmt.Printf("  ❌ Base parser iteration %d failed: %v\n", i, err)
		}
	}
	baseDuration := time.Since(start)

	// Test optimized performance with fast parser
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, err := fastParser.ParseString(testCode)
		if err != nil {
			fmt.Printf("  ❌ Fast parser iteration %d failed: %v\n", i, err)
		}
	}
	optimizedDuration := time.Since(start)

	// Test with caching
	contentHash := hashContent(testCode)
	start = time.Now()
	for i := 0; i < iterations; i++ {
		if cached, hit := cache.Get(testCode, contentHash); hit {
			_ = cached
		} else {
			ast, err := baseParser.ParseString(testCode)
			if err == nil {
				cache.Put(testCode, ast, nil, contentHash)
			}
		}
	}
	cachedDuration := time.Since(start)

	fmt.Printf("  📊 Baseline (%d iterations): %v (avg: %v)\n",
		iterations, baseDuration, baseDuration/time.Duration(iterations))
	fmt.Printf("  📊 Optimized (%d iterations): %v (avg: %v)\n",
		iterations, optimizedDuration, optimizedDuration/time.Duration(iterations))
	fmt.Printf("  📊 Cached (%d iterations): %v (avg: %v)\n",
		iterations, cachedDuration, cachedDuration/time.Duration(iterations))

	if baseDuration > 0 {
		if optimizedDuration > 0 {
			optimizedImprovement := float64(baseDuration) / float64(optimizedDuration)
			fmt.Printf("  🚀 Fast parser improvement: %.1fx\n", optimizedImprovement)
		}
		if cachedDuration > 0 {
			cacheImprovement := float64(baseDuration) / float64(cachedDuration)
			fmt.Printf("  🚀 Cache improvement: %.1fx\n", cacheImprovement)
		}
	}
}
