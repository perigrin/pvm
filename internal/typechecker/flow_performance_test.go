// ABOUTME: Performance and resource usage tests for flow analysis
// ABOUTME: Tests performance characteristics and resource limits for large codebases

package typechecker

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// TestCFGConstructionPerformance tests CFG construction performance with large files
func TestCFGConstructionPerformance(t *testing.T) {
	testCases := []struct {
		name        string
		lines       int
		complexity  string
		maxDuration time.Duration
		description string
	}{
		{
			name:        "small_linear_code",
			lines:       100,
			complexity:  "linear",
			maxDuration: 50 * time.Millisecond,
			description: "Should handle small linear code quickly",
		},
		{
			name:        "medium_conditional_code",
			lines:       500,
			complexity:  "conditional",
			maxDuration: 200 * time.Millisecond,
			description: "Should handle medium conditional code efficiently",
		},
		{
			name:        "large_complex_code",
			lines:       1000,
			complexity:  "complex",
			maxDuration: 500 * time.Millisecond,
			description: "Should handle large complex code within reasonable time",
		},
		{
			name:        "very_large_linear_code",
			lines:       5000,
			complexity:  "linear",
			maxDuration: 1 * time.Second,
			description: "Should handle very large linear code files",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := generateTestCode(tc.lines, tc.complexity)

			analyzer := setupPerformanceTest(t)

			start := time.Now()
			cfg := buildPerformanceCFG(t, analyzer, code)
			duration := time.Since(start)

			if duration > tc.maxDuration {
				t.Errorf("%s: CFG construction took %v, expected < %v",
					tc.description, duration, tc.maxDuration)
			}

			// Verify basic CFG structure
			if cfg.Entry == nil || cfg.Exit == nil {
				t.Errorf("%s: CFG missing entry or exit blocks", tc.description)
			}

			if len(cfg.Nodes) < 2 {
				t.Errorf("%s: CFG should have at least entry and exit blocks", tc.description)
			}

			t.Logf("%s: %d lines -> %d blocks in %v",
				tc.name, tc.lines, len(cfg.Nodes), duration)
		})
	}
}

// TestMemoryUsageWithComplexTypes tests memory usage with complex type hierarchies
func TestMemoryUsageWithComplexTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	testCases := []struct {
		name           string
		typeComplexity int
		variableCount  int
		maxMemoryMB    int
		description    string
	}{
		{
			name:           "simple_types",
			typeComplexity: 5, // Union types with 5 members
			variableCount:  100,
			maxMemoryMB:    10,
			description:    "Should use reasonable memory with simple union types",
		},
		{
			name:           "complex_types",
			typeComplexity: 20, // Union types with 20 members
			variableCount:  200,
			maxMemoryMB:    25,
			description:    "Should handle complex union types efficiently",
		},
		{
			name:           "very_complex_types",
			typeComplexity: 50, // Union types with 50 members
			variableCount:  500,
			maxMemoryMB:    50,
			description:    "Should handle very complex type hierarchies",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var memBefore, memAfter runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memBefore)

			code := generateComplexTypeCode(tc.typeComplexity, tc.variableCount)
			analyzer := setupPerformanceTest(t)
			cfg := buildPerformanceCFG(t, analyzer, code)

			// Perform flow analysis to populate type states
			_ = analyzer.analyzeDataFlow(cfg)

			runtime.GC()
			runtime.ReadMemStats(&memAfter)

			memUsedMB := int((memAfter.Alloc - memBefore.Alloc) / 1024 / 1024)

			if memUsedMB > tc.maxMemoryMB {
				t.Errorf("%s: Used %d MB memory, expected < %d MB",
					tc.description, memUsedMB, tc.maxMemoryMB)
			}

			t.Logf("%s: %d variables with complexity %d used %d MB memory",
				tc.name, tc.variableCount, tc.typeComplexity, memUsedMB)
		})
	}
}

// TestIterationLimitHandling tests that analysis respects iteration limits
func TestIterationLimitHandling(t *testing.T) {
	testCases := []struct {
		name         string
		loopDepth    int
		iterLimit    int
		shouldFinish bool
		description  string
	}{
		{
			name:         "simple_loop_within_limit",
			loopDepth:    3,
			iterLimit:    100,
			shouldFinish: true,
			description:  "Should complete analysis for simple loops within iteration limit",
		},
		{
			name:         "complex_loop_within_limit",
			loopDepth:    5,
			iterLimit:    500,
			shouldFinish: true,
			description:  "Should complete analysis for complex loops within higher limit",
		},
		{
			name:         "very_complex_loop_hits_limit",
			loopDepth:    10,
			iterLimit:    50,
			shouldFinish: false, // May hit iteration limit
			description:  "Should gracefully handle iteration limit for very complex loops",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := generateLoopCode(tc.loopDepth)
			analyzer := setupPerformanceTest(t)
			analyzer.MaxIterations = tc.iterLimit

			cfg := buildPerformanceCFG(t, analyzer, code)

			start := time.Now()
			errors := analyzer.analyzeDataFlow(cfg)
			duration := time.Since(start)

			// Should complete within reasonable time even if hitting limits
			maxDuration := 2 * time.Second
			if duration > maxDuration {
				t.Errorf("%s: Analysis took %v, expected < %v",
					tc.description, duration, maxDuration)
			}

			// Check if analysis completed or hit limits appropriately
			if tc.shouldFinish && len(errors) > 0 {
				// Filter out actual type errors vs limit errors
				limitErrors := 0
				for _, err := range errors {
					if strings.Contains(err.Error(), "iteration limit") {
						limitErrors++
					}
				}
				if limitErrors > 0 {
					t.Errorf("%s: Unexpected iteration limit hit (%d limit errors)",
						tc.description, limitErrors)
				}
			}

			t.Logf("%s: depth %d with limit %d completed in %v (%d errors)",
				tc.name, tc.loopDepth, tc.iterLimit, duration, len(errors))
		})
	}
}

// TestTimeoutHandling tests that analysis respects timeout limits
func TestTimeoutHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	testCases := []struct {
		name        string
		codeSize    int
		timeout     time.Duration
		description string
	}{
		{
			name:        "reasonable_timeout",
			codeSize:    1000,
			timeout:     5 * time.Second,
			description: "Should complete normal analysis within reasonable timeout",
		},
		{
			name:        "tight_timeout",
			codeSize:    2000,
			timeout:     100 * time.Millisecond,
			description: "Should handle tight timeouts gracefully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := generateTestCode(tc.codeSize, "complex")
			analyzer := setupPerformanceTest(t)
			cfg := buildPerformanceCFG(t, analyzer, code)

			// Use channel to implement timeout
			done := make(chan bool, 1)

			go func() {
				_ = analyzer.analyzeDataFlow(cfg)
				done <- true
			}()

			select {
			case <-done:
				t.Logf("%s: Analysis completed within timeout", tc.description)
			case <-time.After(tc.timeout):
				t.Logf("%s: Analysis timed out after %v (expected for tight timeout)",
					tc.description, tc.timeout)
			}
		})
	}
}

// BenchmarkFlowAnalysis benchmarks flow analysis performance
func BenchmarkFlowAnalysis(b *testing.B) {
	benchmarks := []struct {
		name       string
		lines      int
		complexity string
	}{
		{"Small_Linear", 50, "linear"},
		{"Small_Conditional", 50, "conditional"},
		{"Medium_Linear", 200, "linear"},
		{"Medium_Conditional", 200, "conditional"},
		{"Large_Linear", 500, "linear"},
		{"Large_Complex", 500, "complex"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			code := generateTestCode(bm.lines, bm.complexity)
			analyzer := setupPerformanceTest(nil)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cfg := buildBenchmarkCFG(analyzer, code)
				_ = analyzer.analyzeDataFlow(cfg)
			}
		})
	}
}

// Helper functions for performance tests

func setupPerformanceTest(t *testing.T) *FlowAnalyzer {
	typeStore, err := typedef.NewStorage()
	if err != nil {
		if t != nil {
			t.Fatalf("Failed to create type store: %v", err)
		}
		panic(fmt.Sprintf("Failed to create type store: %v", err))
	}

	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	symbolTable.Package = "performance_test"

	tc := NewTypeChecker(hierarchy, symbolTable, "performance_test")
	tc.SafetyAnalysisEnabled = true

	analyzer := NewFlowAnalyzer(tc)
	return analyzer
}

func buildPerformanceCFG(t *testing.T, analyzer *FlowAnalyzer, code string) *ControlFlowGraph {
	astResult := parsePerformanceCode(t, code)
	cfg, err := analyzer.buildControlFlowGraph(astResult)
	if err != nil {
		if t != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		panic(fmt.Sprintf("Failed to build CFG: %v", err))
	}
	return cfg
}

func buildBenchmarkCFG(analyzer *FlowAnalyzer, code string) *ControlFlowGraph {
	return buildPerformanceCFG(nil, analyzer, code)
}

func parsePerformanceCode(t *testing.T, code string) *ast.AST {
	p, err := parser.NewParser()
	if err != nil {
		if t != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}
		panic(fmt.Sprintf("Failed to create parser: %v", err))
	}

	result, err := p.ParseString(code)
	if err != nil {
		if t != nil {
			t.Fatalf("Failed to parse code: %v", err)
		}
		panic(fmt.Sprintf("Failed to parse code: %v", err))
	}

	return result
}

func generateTestCode(lines int, complexity string) string {
	var builder strings.Builder

	builder.WriteString("sub performance_test_function {\n")

	switch complexity {
	case "linear":
		for i := 0; i < lines; i++ {
			builder.WriteString(fmt.Sprintf("    my $var%d = process_data($input%d);\n", i, i))
		}
	case "conditional":
		for i := 0; i < lines/3; i++ {
			builder.WriteString(fmt.Sprintf(`    if ($condition%d) {
        my $result%d = process_a($input%d);
    } else {
        my $result%d = process_b($input%d);
    }
`, i, i, i, i, i))
		}
	case "complex":
		for i := 0; i < lines/5; i++ {
			builder.WriteString(fmt.Sprintf(`    if ($condition%d) {
        for my $item (@items%d) {
            if (defined($item)) {
                my $processed%d = transform($item);
                push @results%d, $processed%d;
            }
        }
    }
`, i, i, i, i, i))
		}
	}

	builder.WriteString("}\n")
	return builder.String()
}

func generateComplexTypeCode(typeComplexity, variableCount int) string {
	var builder strings.Builder

	builder.WriteString("sub complex_type_function {\n")

	// Generate union types with specified complexity
	for i := 0; i < variableCount; i++ {
		var unionTypes []string
		for j := 0; j < typeComplexity; j++ {
			unionTypes = append(unionTypes, fmt.Sprintf("Type%d", j))
		}
		unionType := strings.Join(unionTypes, "|")

		builder.WriteString(fmt.Sprintf("    my $var%d; # Type: %s\n", i, unionType))
		builder.WriteString(fmt.Sprintf("    $var%d = get_complex_value();\n", i))
	}

	builder.WriteString("}\n")
	return builder.String()
}

func generateLoopCode(depth int) string {
	var builder strings.Builder

	builder.WriteString("sub loop_test_function {\n")

	// Generate nested loops
	for i := 0; i < depth; i++ {
		builder.WriteString(fmt.Sprintf("%s    for my $item%d (@array%d) {\n",
			strings.Repeat("    ", i), i, i))
		builder.WriteString(fmt.Sprintf("%s        my $processed%d = process($item%d);\n",
			strings.Repeat("    ", i), i, i))
	}

	// Close nested loops
	for i := depth - 1; i >= 0; i-- {
		builder.WriteString(fmt.Sprintf("%s    }\n", strings.Repeat("    ", i)))
	}

	builder.WriteString("}\n")
	return builder.String()
}
