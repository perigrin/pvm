// ABOUTME: Compilation result caching for improved performance with repeated operations
// ABOUTME: Provides intelligent caching with memory management and performance monitoring

package compiler

import (
	"runtime"
	"sync"
	"time"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// CachingPerlCompiler wraps the unified compiler with intelligent caching
type CachingPerlCompiler struct {
	*PerlCompiler
	cache           *CompilationCache
	memoryPool      *MemoryPool
	perfMonitor     *PerformanceMonitor
	enableProfiling bool
}

// NewCachingPerlCompiler creates a new caching compiler with the specified target
func NewCachingPerlCompiler(target Target, cacheSize int) *CachingPerlCompiler {
	return &CachingPerlCompiler{
		PerlCompiler:    NewPerlCompiler(target),
		cache:           NewCompilationCache(cacheSize),
		memoryPool:      NewMemoryPool(),
		perfMonitor:     &PerformanceMonitor{},
		enableProfiling: true,
	}
}

// NewCachingCleanPerlCompiler creates a caching clean Perl compiler
func NewCachingCleanPerlCompiler(cacheSize int) *CachingPerlCompiler {
	return NewCachingPerlCompiler(TargetCleanPerl, cacheSize)
}

// NewCachingTypedPerlCompiler creates a caching typed Perl compiler
func NewCachingTypedPerlCompiler(cacheSize int) *CachingPerlCompiler {
	return NewCachingPerlCompiler(TargetTypedPerl, cacheSize)
}

// Compile compiles AST with caching support
func (c *CachingPerlCompiler) Compile(ast AST) (string, error) {
	startTime := time.Now()
	var memBefore runtime.MemStats

	if c.enableProfiling {
		runtime.ReadMemStats(&memBefore)
	}

	// Get source content for cache key
	content, err := ast.GetContent()
	if err != nil {
		return "", NewCompilerError(ErrInvalidAST, "failed to get AST content").WithCause(err)
	}

	contentBytes := []byte(content)

	// Check cache first
	if cached, hit := c.cache.Get(contentBytes, c.target); hit {
		if c.enableProfiling {
			duration := time.Since(startTime)
			c.perfMonitor.RecordCompilation(duration, true, 0)
		}
		return cached, nil
	}

	// Cache miss - compile normally
	result, err := c.PerlCompiler.Compile(ast)
	if err != nil {
		return "", err
	}

	// Store in cache for future use
	c.cache.Put(contentBytes, c.target, result)

	if c.enableProfiling {
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)
		duration := time.Since(startTime)
		memoryUsed := int64(memAfter.Alloc - memBefore.Alloc)
		c.perfMonitor.RecordCompilation(duration, false, memoryUsed)
	}

	return result, nil
}

// CompileString compiles source code directly with caching
func (c *CachingPerlCompiler) CompileString(content string) (string, error) {
	startTime := time.Now()
	var memBefore runtime.MemStats

	if c.enableProfiling {
		runtime.ReadMemStats(&memBefore)
	}

	contentBytes := []byte(content)

	// Check cache first
	if cached, hit := c.cache.Get(contentBytes, c.target); hit {
		if c.enableProfiling {
			duration := time.Since(startTime)
			c.perfMonitor.RecordCompilation(duration, true, 0)
		}
		return cached, nil
	}

	// Cache miss - compile normally
	result, err := c.PerlCompiler.CompileString(content)
	if err != nil {
		return "", err
	}

	// Store in cache for future use
	c.cache.Put(contentBytes, c.target, result)

	if c.enableProfiling {
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)
		duration := time.Since(startTime)
		memoryUsed := int64(memAfter.Alloc - memBefore.Alloc)
		c.perfMonitor.RecordCompilation(duration, false, memoryUsed)
	}

	return result, nil
}

// compileFromCSTOptimized performs optimized CST compilation with caching
func (c *CachingPerlCompiler) compileFromCSTOptimized(root *sitter.Node, content []byte) (string, error) {
	// Check cache first
	if cached, hit := c.cache.Get(content, c.target); hit {
		return cached, nil
	}

	// Create optimized transformer
	options := TransformationOptions{
		PreserveComments:   c.options.PreserveComments,
		PreserveWhitespace: c.options.PreserveFormatting,
		RemoveTypeNodes:    c.target == TargetCleanPerl,
	}

	transformer := NewOptimizedCSTTransformer(content, options)
	defer transformer.ClearCache() // Clean up after transformation

	var result *TransformationResult
	var err error

	switch c.target {
	case TargetCleanPerl:
		transformedCode, transformErr := transformer.TransformOptimized(root)
		if transformErr != nil {
			err = transformErr
		} else {
			result = &TransformationResult{
				TransformedCode: transformedCode,
				Success:         true,
			}
		}
	case TargetTypedPerl:
		transformedCode, transformErr := transformer.TransformOptimized(root)
		if transformErr != nil {
			err = transformErr
		} else {
			result = &TransformationResult{
				TransformedCode: transformedCode,
				Success:         true,
			}
		}
	default:
		return "", NewCompilerError(ErrUnsupportedTarget, "unsupported compilation target: "+string(c.target))
	}

	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed, "compilation failed").WithCause(err)
	}

	if !result.Success {
		errorMsg := "transformation failed"
		if result.Error != nil {
			errorMsg = result.Error.Error()
		}
		return "", NewCompilerError(ErrCompilationFailed, errorMsg)
	}

	// Add version pragma for clean Perl if needed
	finalCode := result.TransformedCode
	if c.target == TargetCleanPerl {
		var addPragmaErr error
		finalCode, addPragmaErr = c.addVersionPragma(result.TransformedCode)
		if addPragmaErr != nil {
			return "", NewCompilerError(ErrCompilationFailed, "failed to add version pragma").WithCause(addPragmaErr)
		}
	}

	// Cache the result
	c.cache.Put(content, c.target, finalCode)

	return finalCode, nil
}

// GetCacheStats returns cache performance statistics
func (c *CachingPerlCompiler) GetCacheStats() CacheStats {
	return c.cache.GetStats()
}

// GetPerformanceStats returns compilation performance statistics
func (c *CachingPerlCompiler) GetPerformanceStats() PerformanceStats {
	return c.perfMonitor.GetStats()
}

// ClearCache clears the compilation cache
func (c *CachingPerlCompiler) ClearCache() {
	c.cache = NewCompilationCache(c.cache.maxSize)
}

// SetProfilingEnabled enables or disables performance profiling
func (c *CachingPerlCompiler) SetProfilingEnabled(enabled bool) {
	c.enableProfiling = enabled
}

// OptimizedCompilerRegistry provides a compiler registry with performance optimizations
type OptimizedCompilerRegistry struct {
	*CompilerRegistry
	cacheSize        int
	cachingCompilers map[Target]*CachingPerlCompiler
	mutex            sync.RWMutex
}

// NewOptimizedCompilerRegistry creates a new optimized compiler registry
func NewOptimizedCompilerRegistry(cacheSize int) *OptimizedCompilerRegistry {
	registry := &OptimizedCompilerRegistry{
		CompilerRegistry: NewCompilerRegistry(),
		cacheSize:        cacheSize,
		cachingCompilers: make(map[Target]*CachingPerlCompiler),
	}

	// Register optimized compilers
	registry.RegisterOptimized(NewCachingCleanPerlCompiler(cacheSize))
	registry.RegisterOptimized(NewCachingTypedPerlCompiler(cacheSize))

	return registry
}

// RegisterOptimized adds an optimized compiler to the registry
func (r *OptimizedCompilerRegistry) RegisterOptimized(compiler *CachingPerlCompiler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.cachingCompilers[compiler.Target()] = compiler
	r.CompilerRegistry.Register(compiler)
}

// GetOptimizedCompiler returns the optimized compiler for the specified target
func (r *OptimizedCompilerRegistry) GetOptimizedCompiler(target Target) (*CachingPerlCompiler, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	compiler, exists := r.cachingCompilers[target]
	return compiler, exists
}

// CompileOptimized compiles an AST using optimized compilation with caching
func (r *OptimizedCompilerRegistry) CompileOptimized(ast AST, target Target) (string, error) {
	if compiler, exists := r.GetOptimizedCompiler(target); exists {
		return compiler.Compile(ast)
	}

	// Fallback to regular compilation
	return r.CompilerRegistry.Compile(ast, target)
}

// GetAggregatedStats returns aggregated performance statistics across all compilers
func (r *OptimizedCompilerRegistry) GetAggregatedStats() AggregatedStats {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var totalCompilations int64
	var totalTime time.Duration
	var totalCacheHits int64
	var totalCacheMisses int64
	var totalMemory int64

	for _, compiler := range r.cachingCompilers {
		perfStats := compiler.GetPerformanceStats()
		cacheStats := compiler.GetCacheStats()

		totalCompilations += perfStats.TotalCompilations
		totalTime += perfStats.TotalTime
		totalCacheHits += cacheStats.HitCount
		totalCacheMisses += cacheStats.MissCount
		totalMemory += perfStats.MemoryUsage
	}

	avgTime := time.Duration(0)
	if totalCompilations > 0 {
		avgTime = totalTime / time.Duration(totalCompilations)
	}

	cacheHitRatio := float64(0)
	if totalCacheHits+totalCacheMisses > 0 {
		cacheHitRatio = float64(totalCacheHits) / float64(totalCacheHits+totalCacheMisses)
	}

	return AggregatedStats{
		TotalCompilations: totalCompilations,
		AverageTime:       avgTime,
		TotalTime:         totalTime,
		CacheHitRatio:     cacheHitRatio,
		MemoryUsage:       totalMemory,
		CompilerCount:     len(r.cachingCompilers),
	}
}

// AggregatedStats provides aggregate performance statistics
type AggregatedStats struct {
	TotalCompilations int64
	AverageTime       time.Duration
	TotalTime         time.Duration
	CacheHitRatio     float64
	MemoryUsage       int64
	CompilerCount     int
}

// ClearAllCaches clears caches for all compilers
func (r *OptimizedCompilerRegistry) ClearAllCaches() {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, compiler := range r.cachingCompilers {
		compiler.ClearCache()
	}
}

// SetProfilingForAll enables or disables profiling for all compilers
func (r *OptimizedCompilerRegistry) SetProfilingForAll(enabled bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, compiler := range r.cachingCompilers {
		compiler.SetProfilingEnabled(enabled)
	}
}
