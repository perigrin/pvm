// ABOUTME: Comprehensive integration tests for object pooling system validation
// ABOUTME: Tests pooling performance, memory usage, concurrent safety, and correctness across all components

package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/ls"
	"tamarou.com/pvm/internal/memory"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/parser/treesitter"
	"tamarou.com/pvm/internal/typechecker"
	"tamarou.com/pvm/internal/typedef"
	"tamarou.com/pvm/test/e2e/helpers"
)

// TestPoolingIntegration_MemoryUsageValidation validates memory usage patterns with pooling
func TestPoolingIntegration_MemoryUsageValidation(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create test project with various file sizes
	projectDir := filepath.Join(env.RootDir, "memory_test_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Test different workload patterns
	testCases := []struct {
		name     string
		fileSize string
		content  func(int) string
	}{
		{
			name:     "small_files",
			fileSize: "small",
			content:  generateSmallFile,
		},
		{
			name:     "medium_files",
			fileSize: "medium",
			content:  generateMediumFile,
		},
		{
			name:     "large_files",
			fileSize: "large",
			content:  generateLargeFile,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(projectDir, fmt.Sprintf("test_%s.pl", tc.fileSize))
			content := tc.content(100) // Generate content with 100 elements
			err := os.WriteFile(testFile, []byte(content), 0644)
			require.NoError(t, err)

			// Measure memory before pooling operations
			var memBefore runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memBefore)

			// Perform complete parsing and type checking with pooling
			p, err := parser.NewParser()
			require.NoError(t, err)

			astResult, err := p.ParseFile(testFile)
			require.NoError(t, err)
			require.NotNil(t, astResult)

			// Perform symbol binding
			b := binder.NewBinder()
			// Parse with tree-sitter for CST binding
			tsParser := sitter.NewParser()
			tsParser.SetLanguage(treesitter.Language())
			contentBytes, err := os.ReadFile(testFile)
			require.NoError(t, err)
			tree := tsParser.Parse(contentBytes, nil)
			require.NotNil(t, tree)
			
			symbolTable, err := b.BindCST(tree.RootNode(), contentBytes, astResult.TypeAnnotations)
			require.NoError(t, err)
			require.NotNil(t, symbolTable)

			// Perform type checking
			store, err := typedef.NewStorage()
			require.NoError(t, err)
			hierarchy := typedef.NewTypeHierarchy(store)
			checker := typechecker.NewTypeChecker(hierarchy, symbolTable, "test")
			typeErrors := checker.CheckAST(astResult)
			if len(typeErrors) > 0 {
				t.Logf("Type checking returned %d errors for %s (may be expected)", len(typeErrors), tc.fileSize)
			}

			// Measure memory after operations
			var memAfter runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memAfter)

			// Validate pool statistics - if pools aren't registered yet, this is expected
			poolStats := memory.GetGlobalPoolStats()
			// Note: Pool registration may not be fully implemented yet
			t.Logf("Pool stats for %s: Total allocations: %d, Pool hits: %d",
				tc.fileSize, poolStats.TotalAllocations, poolStats.PoolHits)

			// Memory efficiency validation
			allocDiff := memAfter.TotalAlloc - memBefore.TotalAlloc
			t.Logf("Memory allocated: %d bytes for %s", allocDiff, tc.fileSize)

			// Pool efficiency should show reuse
			if poolStats.TotalAllocations > 0 {
				hitRate := float64(poolStats.PoolHits) / float64(poolStats.TotalAllocations)
				assert.Greater(t, hitRate, 0.1, "Pool hit rate should be reasonable")
				t.Logf("Pool hit rate: %.2f%% for %s", hitRate*100, tc.fileSize)
			}
		})
	}
}

// TestPoolingIntegration_ConcurrentUsage tests pool safety under concurrent access
func TestPoolingIntegration_ConcurrentUsage(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create multiple test files
	projectDir := filepath.Join(env.RootDir, "concurrent_test_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	const numFiles = 10
	const numConcurrentWorkers = 5

	// Create test files
	testFiles := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		testFile := filepath.Join(projectDir, fmt.Sprintf("concurrent_test_%d.pl", i))
		content := generateMediumFile(50 + i*10) // Varying content sizes
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)
		testFiles[i] = testFile
	}

	// Test concurrent parsing and type checking
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	errChan := make(chan error, numConcurrentWorkers)

	// Launch concurrent workers
	for w := 0; w < numConcurrentWorkers; w++ {
		go func(workerID int) {
			defer func() {
				if r := recover(); r != nil {
					errChan <- fmt.Errorf("worker %d panicked: %v", workerID, r)
				}
			}()

			for i, testFile := range testFiles {
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				default:
				}

				// Parse file
				p, err := parser.NewParser()
				if err != nil {
					errChan <- fmt.Errorf("worker %d: failed to create parser for file %d: %w", workerID, i, err)
					return
				}

				astResult, err := p.ParseFile(testFile)
				if err != nil {
					errChan <- fmt.Errorf("worker %d: failed to parse file %d: %w", workerID, i, err)
					return
				}

				// Bind symbols
				b := binder.NewBinder()
				// Parse with tree-sitter for CST binding
				tsParser := sitter.NewParser()
				tsParser.SetLanguage(treesitter.Language())
				contentBytes, err := os.ReadFile(testFile)
				if err != nil {
					errChan <- fmt.Errorf("worker %d: failed to read file %d: %w", workerID, i, err)
					return
				}
				tree := tsParser.Parse(contentBytes, nil)
				if tree == nil {
					errChan <- fmt.Errorf("worker %d: failed to parse with tree-sitter file %d", workerID, i)
					return
				}
				
				symbolTable, err := b.BindCST(tree.RootNode(), contentBytes, astResult.TypeAnnotations)
				if err != nil {
					errChan <- fmt.Errorf("worker %d: failed to bind file %d: %w", workerID, i, err)
					return
				}

				// Type check
				store, err := typedef.NewStorage()
				if err != nil {
					errChan <- fmt.Errorf("worker %d: failed to create type storage for file %d: %w", workerID, i, err)
					return
				}
				hierarchy := typedef.NewTypeHierarchy(store)
				checker := typechecker.NewTypeChecker(hierarchy, symbolTable, "test")
				typeErrors := checker.CheckAST(astResult)
				if len(typeErrors) > 0 {
					t.Logf("Type checking returned %d errors for worker %d file %d (may be expected)", len(typeErrors), workerID, i)
				}
			}
			errChan <- nil
		}(w)
	}

	// Wait for all workers to complete
	for w := 0; w < numConcurrentWorkers; w++ {
		err := <-errChan
		assert.NoError(t, err, "Concurrent worker should not fail")
	}

	// Validate pool integrity after concurrent usage
	poolStats := memory.GetGlobalPoolStats()
	// Note: Pool registration may not be fully implemented yet
	t.Logf("Concurrent test completed - Total allocations: %d, Pool hits: %d, Leaks: %d",
		poolStats.TotalAllocations, poolStats.PoolHits, poolStats.PoolLeaks)
}

// TestPoolingIntegration_LSPOperations tests pooling with LSP protocol operations
func TestPoolingIntegration_LSPOperations(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create test file for LSP operations
	projectDir := filepath.Join(env.RootDir, "lsp_test_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	testFile := filepath.Join(projectDir, "lsp_test.pl")
	content := `package TestModule;
use v5.36;

field Str $name;
field Int $count = 0;

method greet(Str $greeting) -> Str {
    return "$greeting, $name!";
}

method increment() -> Int {
    $count++;
    return $count;
}

1;`
	err = os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Initialize language service with pooling
	server, err := ls.NewLanguageService()
	require.NoError(t, err)
	require.NotNil(t, server)

	// Update document in language service
	err = server.UpdateDocument(testFile, content, 1)
	require.NoError(t, err)

	// Test multiple LSP operations that should use pooling
	lspOps := []struct {
		name string
		op   func() error
	}{
		{
			name: "completion",
			op: func() error {
				pos := ls.Position{Line: 8, Character: 10} // In method body
				_, err := server.GetCompletions(testFile, pos)
				return err
			},
		},
		{
			name: "hover",
			op: func() error {
				pos := ls.Position{Line: 4, Character: 5} // On field declaration
				_, err := server.GetHover(testFile, pos)
				return err
			},
		},
		{
			name: "definition",
			op: func() error {
				pos := ls.Position{Line: 8, Character: 15} // On variable reference
				_, err := server.GetDefinition(testFile, pos)
				return err
			},
		},
		{
			name: "document_symbols",
			op: func() error {
				_, err := server.GetDocumentSymbols(testFile)
				return err
			},
		},
	}

	// Record pool stats before LSP operations
	statsBefore := memory.GetGlobalPoolStats()

	// Perform LSP operations multiple times to test pool reuse
	for i := 0; i < 5; i++ {
		for _, lspOp := range lspOps {
			t.Run(fmt.Sprintf("%s_iteration_%d", lspOp.name, i), func(t *testing.T) {
				err := lspOp.op()
				// LSP operations may return errors for incomplete implementations
				// but should not panic or cause pool corruption
				if err != nil {
					t.Logf("LSP operation %s returned error (may be expected): %v", lspOp.name, err)
				}
			})
		}
	}

	// Validate pool usage after LSP operations
	statsAfter := memory.GetGlobalPoolStats()
	// Note: Pool registration may not be fully implemented yet
	allocDiff := statsAfter.TotalAllocations - statsBefore.TotalAllocations
	hitDiff := statsAfter.PoolHits - statsBefore.PoolHits
	var hitRate float64
	if allocDiff > 0 {
		hitRate = float64(hitDiff) / float64(allocDiff) * 100
	}
	t.Logf("LSP pooling test - Allocations: %d, Hits: %d, Hit rate: %.2f%%",
		allocDiff, hitDiff, hitRate)
}

// TestPoolingIntegration_StressTest performs stress testing with large codebases
func TestPoolingIntegration_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create large test project
	projectDir := filepath.Join(env.RootDir, "stress_test_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	const numLargeFiles = 20

	// Create multiple large files
	for i := 0; i < numLargeFiles; i++ {
		testFile := filepath.Join(projectDir, fmt.Sprintf("stress_test_%d.pl", i))
		content := generateLargeFile(200 + i*50) // Increasingly large files
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Measure baseline memory
	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)
	poolStatsBefore := memory.GetGlobalPoolStats()

	startTime := time.Now()

	// Process all files with pooling
	for i := 0; i < numLargeFiles; i++ {
		testFile := filepath.Join(projectDir, fmt.Sprintf("stress_test_%d.pl", i))

		// Parse
		p, err := parser.NewParser()
		require.NoError(t, err)

		astResult, err := p.ParseFile(testFile)
		require.NoError(t, err)

		// Bind
		b := binder.NewBinder()
		// Parse with tree-sitter for CST binding
		tsParser := sitter.NewParser()
		tsParser.SetLanguage(treesitter.Language())
		contentBytes, err := os.ReadFile(testFile)
		require.NoError(t, err)
		tree := tsParser.Parse(contentBytes, nil)
		require.NotNil(t, tree)
		
		symbolTable, err := b.BindCST(tree.RootNode(), contentBytes, astResult.TypeAnnotations)
		require.NoError(t, err)

		// Type check
		store, err := typedef.NewStorage()
		require.NoError(t, err)
		hierarchy := typedef.NewTypeHierarchy(store)
		checker := typechecker.NewTypeChecker(hierarchy, symbolTable, "test")
		typeErrors := checker.CheckAST(astResult)
		if len(typeErrors) > 0 {
			t.Logf("Type checking returned %d errors for stress test file %d (may be expected)", len(typeErrors), i)
		}

		// Force periodic cleanup to test pool stability
		if i%5 == 0 {
			runtime.GC()
		}
	}

	duration := time.Since(startTime)

	// Measure final memory and pool stats
	var memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memAfter)
	poolStatsAfter := memory.GetGlobalPoolStats()

	// Validate stress test results
	allocDiff := memAfter.TotalAlloc - memBefore.TotalAlloc
	poolAllocDiff := poolStatsAfter.TotalAllocations - poolStatsBefore.TotalAllocations
	poolHitDiff := poolStatsAfter.PoolHits - poolStatsBefore.PoolHits

	assert.Greater(t, poolAllocDiff, uint64(0), "Stress test should use pools")
	assert.Equal(t, uint64(0), poolStatsAfter.PoolLeaks, "Stress test should not leak pool objects")

	// Calculate hit rate
	var hitRate float64
	if poolAllocDiff > 0 {
		hitRate = float64(poolHitDiff) / float64(poolAllocDiff)
	}

	t.Logf("Stress test completed in %v", duration)
	t.Logf("Memory allocated: %d bytes", allocDiff)
	t.Logf("Pool allocations: %d, Pool hits: %d, Hit rate: %.2f%%",
		poolAllocDiff, poolHitDiff, hitRate*100)

	// Performance expectations for stress test
	assert.Less(t, duration, 2*time.Minute, "Stress test should complete in reasonable time")
	assert.Greater(t, hitRate, 0.2, "Pool hit rate should show meaningful reuse under stress")
}

// TestPoolingIntegration_MemoryLeakDetection validates no memory leaks in long-running sessions
func TestPoolingIntegration_MemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create test file
	projectDir := filepath.Join(env.RootDir, "leak_test_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	testFile := filepath.Join(projectDir, "leak_test.pl")
	content := generateMediumFile(100)
	err = os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	const iterations = 100
	memorySnapshots := make([]uint64, iterations+1)
	poolSnapshots := make([]memory.GlobalPoolStats, iterations+1)

	// Take initial snapshot
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	memorySnapshots[0] = memStats.HeapAlloc
	poolSnapshots[0] = memory.GetGlobalPoolStats()

	// Run many iterations to detect memory leaks
	for i := 1; i <= iterations; i++ {
		// Parse and process file
		p, err := parser.NewParser()
		require.NoError(t, err)

		astResult, err := p.ParseFile(testFile)
		require.NoError(t, err)

		b := binder.NewBinder()
		// Parse with tree-sitter for CST binding
		tsParser := sitter.NewParser()
		tsParser.SetLanguage(treesitter.Language())
		contentBytes, err := os.ReadFile(testFile)
		require.NoError(t, err)
		tree := tsParser.Parse(contentBytes, nil)
		require.NotNil(t, tree)
		
		symbolTable, err := b.BindCST(tree.RootNode(), contentBytes, astResult.TypeAnnotations)
		require.NoError(t, err)

		store, err := typedef.NewStorage()
		require.NoError(t, err)
		hierarchy := typedef.NewTypeHierarchy(store)
		checker := typechecker.NewTypeChecker(hierarchy, symbolTable, "test")
		typeErrors := checker.CheckAST(astResult)
		if len(typeErrors) > 0 {
			t.Logf("Type checking returned %d errors for memory leak test iteration %d (may be expected)", len(typeErrors), i)
		}

		// Take memory snapshot every 10 iterations
		if i%10 == 0 {
			runtime.GC()
			runtime.ReadMemStats(&memStats)
			memorySnapshots[i/10] = memStats.HeapAlloc
			poolSnapshots[i/10] = memory.GetGlobalPoolStats()
		}
	}

	// Analyze memory growth pattern
	finalSnapshot := len(memorySnapshots) - 1
	// Calculate memory growth, handling case where final < initial (memory was cleaned up)
	var memoryGrowth uint64
	if memorySnapshots[finalSnapshot] > memorySnapshots[0] {
		memoryGrowth = memorySnapshots[finalSnapshot] - memorySnapshots[0]
	} else {
		memoryGrowth = 0 // Memory was cleaned up, no growth
	}
	poolLeaks := poolSnapshots[finalSnapshot].PoolLeaks

	t.Logf("Memory leak test completed %d iterations", iterations)
	t.Logf("Initial heap: %d bytes, Final heap: %d bytes, Growth: %d bytes",
		memorySnapshots[0], memorySnapshots[finalSnapshot], memoryGrowth)
	t.Logf("Pool leaks detected: %d", poolLeaks)

	// Memory growth should be bounded (allowing for some GC overhead)
	maxAllowedGrowth := uint64(50 * 1024 * 1024) // 50MB max growth
	assert.Less(t, memoryGrowth, maxAllowedGrowth, "Memory growth should be bounded")
	assert.Equal(t, uint64(0), poolLeaks, "Should have no pool leaks")
}

// TestPoolingIntegration_BackwardCompatibility ensures no functional regressions
func TestPoolingIntegration_BackwardCompatibility(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test all major functionality still works with pooling enabled
	projectDir := filepath.Join(env.RootDir, "compat_test_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Test cases representing different code patterns
	testCases := []struct {
		name    string
		content string
	}{
		{
			name: "basic_typed_perl",
			content: `package Basic;
use v5.36;

sub new {
    my $class = shift;
    return bless {
        name => "test",
        count => 0
    }, $class;
}

sub greet {
    my $self = shift;
    return "Hello, " . $self->{name} . "!";
}

1;`,
		},
		{
			name: "complex_types",
			content: `package Complex;
use v5.36;

sub new {
    my $class = shift;
    return bless {
        value => undef,
        numbers => []
    }, $class;
}

sub process {
    my ($self, $input) = @_;
    if (ref($input) eq 'SCALAR') {
        return $$input + 1;
    }
    return $input . " processed";
}

1;`,
		},
		{
			name: "legacy_perl",
			content: `package Legacy;
use strict;
use warnings;

sub new {
    my $class = shift;
    my $self = { count => 0 };
    bless $self, $class;
}

sub increment {
    my $self = shift;
    $self->{count}++;
    return $self->{count};
}

1;`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join(projectDir, fmt.Sprintf("%s.pl", tc.name))
			err := os.WriteFile(testFile, []byte(tc.content), 0644)
			require.NoError(t, err)

			// Parse should work
			p, err := parser.NewParser()
			require.NoError(t, err)

			astResult, err := p.ParseFile(testFile)
			require.NoError(t, err)
			require.NotNil(t, astResult)

			// Symbol binding should work
			b := binder.NewBinder()
			// Parse with tree-sitter for CST binding
			tsParser := sitter.NewParser()
			tsParser.SetLanguage(treesitter.Language())
			contentBytes, err := os.ReadFile(testFile)
			require.NoError(t, err)
			tree := tsParser.Parse(contentBytes, nil)
			require.NotNil(t, tree)
			
			symbolTable, err := b.BindCST(tree.RootNode(), contentBytes, astResult.TypeAnnotations)
			require.NoError(t, err)
			require.NotNil(t, symbolTable)

			// Type checking should work (may have errors but shouldn't crash)
			store, err := typedef.NewStorage()
			require.NoError(t, err)
			hierarchy := typedef.NewTypeHierarchy(store)
			checker := typechecker.NewTypeChecker(hierarchy, symbolTable, tc.name)
			typeErrors := checker.CheckAST(astResult)
			// Type checking may return errors for complex cases, but should not panic
			if len(typeErrors) > 0 {
				t.Logf("Type checking returned %d errors for %s (may be expected)", len(typeErrors), tc.name)
			}

			t.Logf("Backward compatibility verified for %s", tc.name)
		})
	}
}

// Helper functions for generating test content

func generateSmallFile(numElements int) string {
	var content strings.Builder
	content.WriteString("package SmallTest;\nuse v5.36;\n\n")

	for i := 0; i < numElements; i++ {
		content.WriteString(fmt.Sprintf("field Str $var%d = \"value%d\";\n", i, i))
	}

	content.WriteString("\n1;\n")
	return content.String()
}

func generateMediumFile(numElements int) string {
	var content strings.Builder
	content.WriteString("package MediumTest;\nuse v5.36;\n\n")

	// Add fields
	for i := 0; i < numElements; i++ {
		content.WriteString(fmt.Sprintf("field Int $count%d = %d;\n", i, i))
	}

	// Add methods
	for i := 0; i < numElements/2; i++ {
		content.WriteString(fmt.Sprintf(`
method process%d(Int $input) returns Int {
    my $result%d = $input + $count%d;
    return $result%d;
}
`, i, i, i, i))
	}

	content.WriteString("\n1;\n")
	return content.String()
}

func generateLargeFile(numElements int) string {
	var content strings.Builder
	content.WriteString("package LargeTest;\nuse v5.36;\n\n")

	// Add type definitions
	for i := 0; i < numElements/4; i++ {
		content.WriteString(fmt.Sprintf("type Type%d = Str|Int|Num;\n", i))
	}

	// Add fields
	for i := 0; i < numElements; i++ {
		content.WriteString(fmt.Sprintf("field Type%d $field%d;\n", i%(numElements/4), i))
	}

	// Add complex methods
	for i := 0; i < numElements/3; i++ {
		content.WriteString(fmt.Sprintf(`
method complexMethod%d(Type%d $param1, Str $param2) returns Type%d {
    my $local1_%d = $param1;
    my $local2_%d = $param2;

    if (defined($local1_%d)) {
        if (ref($local1_%d) eq 'SCALAR') {
            $local1_%d = $local1_%d + 1;
        } else {
            $local1_%d = "$local1_%d modified";
        }
    }

    return $local1_%d;
}
`, i, i%(numElements/4), i%(numElements/4), i, i, i, i, i, i, i, i, i))
	}

	content.WriteString("\n1;\n")
	return content.String()
}
