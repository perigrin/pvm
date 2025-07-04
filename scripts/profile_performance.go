// ABOUTME: Performance profiling tool for identifying bottlenecks across PVM components
// ABOUTME: Profiles CPU, memory, and goroutine usage for parser, binder, typechecker, and LSP operations

//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/parser/treesitter"
	"tamarou.com/pvm/internal/typechecker"
	"tamarou.com/pvm/internal/typedef"
)

// ProfileConfig represents profiling configuration
type ProfileConfig struct {
	OutputDir    string
	SampleSize   int
	TestFiles    []string
	ProfileTypes []string
}

// ProfileResult represents profiling results
type ProfileResult struct {
	Component      string        `json:"component"`
	Duration       time.Duration `json:"duration"`
	MemoryBefore   uint64        `json:"memory_before"`
	MemoryAfter    uint64        `json:"memory_after"`
	MemoryAlloced  uint64        `json:"memory_alloced"`
	GoroutineCount int           `json:"goroutine_count"`
	CPUProfile     string        `json:"cpu_profile,omitempty"`
	MemProfile     string        `json:"mem_profile,omitempty"`
}

func main() {
	config := ProfileConfig{
		OutputDir:    "performance-profiles",
		SampleSize:   10, // Reduced for faster profiling
		TestFiles:    generateTestFiles(),
		ProfileTypes: []string{"cpu", "mem", "goroutine", "block"},
	}

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	fmt.Println("🔍 Starting comprehensive performance profiling...")
	fmt.Printf("Output directory: %s\n", config.OutputDir)
	fmt.Printf("Sample size: %d iterations\n", config.SampleSize)
	fmt.Printf("Test files: %d files\n", len(config.TestFiles))

	// Profile each major component
	results := []ProfileResult{
		profileParser(config),
		profileBinder(config),
		profileTypeChecker(config),
		profileIntegratedPipeline(config),
	}

	// Generate summary report
	generateSummaryReport(results, config.OutputDir)

	fmt.Println("✅ Performance profiling complete!")
	fmt.Printf("📊 Results saved to %s/\n", config.OutputDir)
	fmt.Println("📈 Run 'go tool pprof' on .prof files for detailed analysis")
}

func generateTestFiles() []string {
	// Generate various Perl code samples for testing
	samples := []struct {
		name    string
		content string
	}{
		{
			"simple_script.pl",
			`my $count = 42;
my $name = "test";
print "$name: $count\n";`,
		},
		{
			"typed_script.pl",
			`my Int $count = 42;
my Str $name = "test";
my ArrayRef[Int] $numbers = [1, 2, 3, 4, 5];

sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

my $result = add($count, 10);`,
		},
		{
			"complex_script.pl",
			strings.Repeat(`my Int $var_%d = %d;
my Str $msg_%d = "message_%d";

sub process_%d(Int $input) -> Int {
    return $input * 2 + 1;
}

`, 10), // Reduced repetitions for faster profiling
		},
		{
			"union_types.pl",
			`type NumberOrString = Int|Str;
type OptionalData = HashRef|Undef;
type ComplexUnion = Int|Str|ArrayRef[Int]|HashRef[Str];

my NumberOrString $value = 42;
$value = "hello";

my OptionalData $data = { key => "value" };
$data = undef;`,
		},
	}

	var files []string
	for i, sample := range samples {
		// Expand template variables for complex script
		content := sample.content
		if strings.Contains(content, "%d") {
			expanded := ""
			for j := 0; j < 10; j++ { // Reduced for faster profiling
				expanded += fmt.Sprintf(content, j, j, j, j, j)
			}
			content = expanded
		}

		filename := fmt.Sprintf("test_%d_%s", i, sample.name)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			log.Printf("Warning: failed to create test file %s: %v", filename, err)
			continue
		}
		files = append(files, filename)
	}

	return files
}

func profileParser(config ProfileConfig) ProfileResult {
	fmt.Println("🔧 Profiling parser performance...")

	// Start CPU profiling
	cpuFile := filepath.Join(config.OutputDir, "parser_cpu.prof")
	cpuProfile, err := os.Create(cpuFile)
	if err != nil {
		log.Printf("Warning: failed to create CPU profile: %v", err)
	} else {
		defer cpuProfile.Close()
		if err := pprof.StartCPUProfile(cpuProfile); err != nil {
			log.Printf("Warning: failed to start CPU profiling: %v", err)
		} else {
			defer pprof.StopCPUProfile()
		}
	}

	// Get initial memory stats
	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	start := time.Now()
	goroutinesBefore := runtime.NumGoroutine()

	// Run parser performance test
	p, err := parser.NewParser()
	if err != nil {
		log.Printf("Error: Failed to create parser: %v", err)
		return ProfileResult{}
	}

	for i := 0; i < config.SampleSize; i++ {
		for _, testFile := range config.TestFiles {
			content, err := os.ReadFile(testFile)
			if err != nil {
				continue
			}

			_, err = p.ParseString(string(content))
			if err != nil {
				log.Printf("Warning: parse error in %s: %v", testFile, err)
			}
		}
	}

	duration := time.Since(start)

	// Get final memory stats
	var memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	// Create memory profile
	memFile := filepath.Join(config.OutputDir, "parser_mem.prof")
	memProfile, err := os.Create(memFile)
	if err != nil {
		log.Printf("Warning: failed to create memory profile: %v", err)
	} else {
		defer memProfile.Close()
		if err := pprof.WriteHeapProfile(memProfile); err != nil {
			log.Printf("Warning: failed to write memory profile: %v", err)
		}
	}

	result := ProfileResult{
		Component:      "Parser",
		Duration:       duration,
		MemoryBefore:   memBefore.Alloc,
		MemoryAfter:    memAfter.Alloc,
		MemoryAlloced:  memAfter.TotalAlloc - memBefore.TotalAlloc,
		GoroutineCount: runtime.NumGoroutine() - goroutinesBefore,
		CPUProfile:     cpuFile,
		MemProfile:     memFile,
	}

	fmt.Printf("  Duration: %v\n", result.Duration)
	fmt.Printf("  Memory allocated: %d bytes\n", result.MemoryAlloced)
	fmt.Printf("  Goroutines created: %d\n", result.GoroutineCount)

	return result
}

func profileBinder(config ProfileConfig) ProfileResult {
	fmt.Println("🔗 Profiling binder performance...")

	cpuFile := filepath.Join(config.OutputDir, "binder_cpu.prof")
	cpuProfile, err := os.Create(cpuFile)
	if err != nil {
		log.Printf("Warning: failed to create CPU profile: %v", err)
	} else {
		defer cpuProfile.Close()
		if err := pprof.StartCPUProfile(cpuProfile); err != nil {
			log.Printf("Warning: failed to start CPU profiling: %v", err)
		} else {
			defer pprof.StopCPUProfile()
		}
	}

	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	start := time.Now()
	goroutinesBefore := runtime.NumGoroutine()

	// Run binder performance test
	p, err := parser.NewParser()
	if err != nil {
		log.Printf("Error: Failed to create parser: %v", err)
		return ProfileResult{}
	}

	b := binder.NewBinder()

	for i := 0; i < config.SampleSize; i++ {
		for _, testFile := range config.TestFiles {
			content, err := os.ReadFile(testFile)
			if err != nil {
				continue
			}

			ast, err := p.ParseString(string(content))
			if err != nil {
				continue
			}

			// Parse with tree-sitter for CST binding
			tsParser := sitter.NewParser()
			tsParser.SetLanguage(treesitter.Language())
			contentBytes := []byte(string(content))
			tree := tsParser.Parse(contentBytes, nil)
			if tree == nil {
				log.Printf("Warning: tree-sitter parse failed in %s", testFile)
				continue
			}

			_, err = b.BindCST(tree.RootNode(), contentBytes, ast.TypeAnnotations)
			if err != nil {
				log.Printf("Warning: bind error in %s: %v", testFile, err)
			}
		}
	}

	duration := time.Since(start)

	var memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	memFile := filepath.Join(config.OutputDir, "binder_mem.prof")
	memProfile, err := os.Create(memFile)
	if err != nil {
		log.Printf("Warning: failed to create memory profile: %v", err)
	} else {
		defer memProfile.Close()
		if err := pprof.WriteHeapProfile(memProfile); err != nil {
			log.Printf("Warning: failed to write memory profile: %v", err)
		}
	}

	result := ProfileResult{
		Component:      "Binder",
		Duration:       duration,
		MemoryBefore:   memBefore.Alloc,
		MemoryAfter:    memAfter.Alloc,
		MemoryAlloced:  memAfter.TotalAlloc - memBefore.TotalAlloc,
		GoroutineCount: runtime.NumGoroutine() - goroutinesBefore,
		CPUProfile:     cpuFile,
		MemProfile:     memFile,
	}

	fmt.Printf("  Duration: %v\n", result.Duration)
	fmt.Printf("  Memory allocated: %d bytes\n", result.MemoryAlloced)
	fmt.Printf("  Goroutines created: %d\n", result.GoroutineCount)

	return result
}

func profileTypeChecker(config ProfileConfig) ProfileResult {
	fmt.Println("🎯 Profiling type checker performance...")

	cpuFile := filepath.Join(config.OutputDir, "typechecker_cpu.prof")
	cpuProfile, err := os.Create(cpuFile)
	if err != nil {
		log.Printf("Warning: failed to create CPU profile: %v", err)
	} else {
		defer cpuProfile.Close()
		if err := pprof.StartCPUProfile(cpuProfile); err != nil {
			log.Printf("Warning: failed to start CPU profiling: %v", err)
		} else {
			defer pprof.StopCPUProfile()
		}
	}

	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	start := time.Now()
	goroutinesBefore := runtime.NumGoroutine()

	// Run type checker performance test
	p, err := parser.NewParser()
	if err != nil {
		log.Printf("Error: Failed to create parser: %v", err)
		return ProfileResult{}
	}

	b := binder.NewBinder()

	for i := 0; i < config.SampleSize; i++ {
		for _, testFile := range config.TestFiles {
			content, err := os.ReadFile(testFile)
			if err != nil {
				continue
			}

			ast, err := p.ParseString(string(content))
			if err != nil {
				continue
			}

			// Parse with tree-sitter for CST binding
			tsParser := sitter.NewParser()
			tsParser.SetLanguage(treesitter.Language())
			contentBytes := []byte(string(content))
			tree := tsParser.Parse(contentBytes, nil)
			if tree == nil {
				log.Printf("Warning: tree-sitter parse failed in %s", testFile)
				continue
			}

			symbolTable, err := b.BindCST(tree.RootNode(), contentBytes, ast.TypeAnnotations)
			if err != nil {
				continue
			}

			store, _ := typedef.NewStorage()
			hierarchy := typedef.NewTypeHierarchy(store)
			tc := typechecker.NewTypeChecker(hierarchy, symbolTable, "test_module")
			errors := tc.CheckAST(ast)
			if len(errors) > 0 {
				log.Printf("Warning: %d type check errors in %s", len(errors), testFile)
			}
		}
	}

	duration := time.Since(start)

	var memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	memFile := filepath.Join(config.OutputDir, "typechecker_mem.prof")
	memProfile, err := os.Create(memFile)
	if err != nil {
		log.Printf("Warning: failed to create memory profile: %v", err)
	} else {
		defer memProfile.Close()
		if err := pprof.WriteHeapProfile(memProfile); err != nil {
			log.Printf("Warning: failed to write memory profile: %v", err)
		}
	}

	result := ProfileResult{
		Component:      "TypeChecker",
		Duration:       duration,
		MemoryBefore:   memBefore.Alloc,
		MemoryAfter:    memAfter.Alloc,
		MemoryAlloced:  memAfter.TotalAlloc - memBefore.TotalAlloc,
		GoroutineCount: runtime.NumGoroutine() - goroutinesBefore,
		CPUProfile:     cpuFile,
		MemProfile:     memFile,
	}

	fmt.Printf("  Duration: %v\n", result.Duration)
	fmt.Printf("  Memory allocated: %d bytes\n", result.MemoryAlloced)
	fmt.Printf("  Goroutines created: %d\n", result.GoroutineCount)

	return result
}

func profileIntegratedPipeline(config ProfileConfig) ProfileResult {
	fmt.Println("🚀 Profiling integrated pipeline performance...")

	cpuFile := filepath.Join(config.OutputDir, "pipeline_cpu.prof")
	cpuProfile, err := os.Create(cpuFile)
	if err != nil {
		log.Printf("Warning: failed to create CPU profile: %v", err)
	} else {
		defer cpuProfile.Close()
		if err := pprof.StartCPUProfile(cpuProfile); err != nil {
			log.Printf("Warning: failed to start CPU profiling: %v", err)
		} else {
			defer pprof.StopCPUProfile()
		}
	}

	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	start := time.Now()
	goroutinesBefore := runtime.NumGoroutine()

	// Run full pipeline performance test
	for i := 0; i < config.SampleSize; i++ {
		for _, testFile := range config.TestFiles {
			content, err := os.ReadFile(testFile)
			if err != nil {
				continue
			}

			// Full pipeline: Parse → Bind → Check
			p, err := parser.NewParser()
			if err != nil {
				continue
			}

			ast, err := p.ParseString(string(content))
			if err != nil {
				continue
			}

			b := binder.NewBinder()
			// Parse with tree-sitter for CST binding
			tsParser := sitter.NewParser()
			tsParser.SetLanguage(treesitter.Language())
			contentBytes := []byte(string(content))
			tree := tsParser.Parse(contentBytes, nil)
			if tree == nil {
				log.Printf("Warning: tree-sitter parse failed in %s", testFile)
				continue
			}

			symbolTable, err := b.BindCST(tree.RootNode(), contentBytes, ast.TypeAnnotations)
			if err != nil {
				continue
			}

			store, _ := typedef.NewStorage()
			hierarchy := typedef.NewTypeHierarchy(store)
			tc := typechecker.NewTypeChecker(hierarchy, symbolTable, "test_module")
			_ = tc.CheckAST(ast)
			if err != nil {
				log.Printf("Warning: pipeline error in %s: %v", testFile, err)
			}
		}
	}

	duration := time.Since(start)

	var memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	memFile := filepath.Join(config.OutputDir, "pipeline_mem.prof")
	memProfile, err := os.Create(memFile)
	if err != nil {
		log.Printf("Warning: failed to create memory profile: %v", err)
	} else {
		defer memProfile.Close()
		if err := pprof.WriteHeapProfile(memProfile); err != nil {
			log.Printf("Warning: failed to write memory profile: %v", err)
		}
	}

	result := ProfileResult{
		Component:      "IntegratedPipeline",
		Duration:       duration,
		MemoryBefore:   memBefore.Alloc,
		MemoryAfter:    memAfter.Alloc,
		MemoryAlloced:  memAfter.TotalAlloc - memBefore.TotalAlloc,
		GoroutineCount: runtime.NumGoroutine() - goroutinesBefore,
		CPUProfile:     cpuFile,
		MemProfile:     memFile,
	}

	fmt.Printf("  Duration: %v\n", result.Duration)
	fmt.Printf("  Memory allocated: %d bytes\n", result.MemoryAlloced)
	fmt.Printf("  Goroutines created: %d\n", result.GoroutineCount)

	return result
}

func generateSummaryReport(results []ProfileResult, outputDir string) {
	reportFile := filepath.Join(outputDir, "performance_summary.md")
	report := `# Performance Profiling Summary

Generated at: ` + time.Now().Format("2006-01-02 15:04:05") + `

## Component Performance

| Component | Duration | Memory Allocated | Goroutines |
|-----------|----------|------------------|------------|
`

	for _, result := range results {
		report += fmt.Sprintf("| %s | %v | %d bytes | %d |\n",
			result.Component,
			result.Duration,
			result.MemoryAlloced,
			result.GoroutineCount)
	}

	report += `
## Analysis Commands

To analyze the profiles in detail, use these commands:

### CPU Profiles
`
	for _, result := range results {
		if result.CPUProfile != "" {
			report += fmt.Sprintf("```bash\ngo tool pprof %s\n```\n", result.CPUProfile)
		}
	}

	report += `
### Memory Profiles
`
	for _, result := range results {
		if result.MemProfile != "" {
			report += fmt.Sprintf("```bash\ngo tool pprof %s\n```\n", result.MemProfile)
		}
	}

	report += `
## Optimization Recommendations

Based on the profiling results:

1. **Highest Duration Component**: Focus optimization efforts here
2. **Memory Usage**: Consider object pooling for high-allocation components
3. **Goroutine Count**: Monitor for potential goroutine leaks

Run this profiler regularly to track optimization progress.
`

	if err := os.WriteFile(reportFile, []byte(report), 0644); err != nil {
		log.Printf("Warning: failed to write summary report: %v", err)
	}
}
