// ABOUTME: Performance optimization integration for PVM components
// ABOUTME: Orchestrates caching, pooling, and optimization strategies across the entire pipeline

package performance

import (
	"context"
	"sync"
	"time"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typechecker"
	"tamarou.com/pvm/internal/typedef"
)

// OptimizedPipeline integrates all performance optimizations into a single interface
type OptimizedPipeline struct {
	// Core components
	fastParser      *FastParser
	optimizedParser *OptimizedParser
	memBinder       *MemoryOptimizedBinder

	// Caching and optimization
	parseCache   *ParseCache
	typeCache    *TypeCache
	objectPool   *ObjectPool
	lazyResolver *LazyTypeResolver

	// Monitoring
	perfMonitor *PerformanceMonitor

	// Configuration
	config *OptimizationConfig

	// Statistics
	mu                  sync.RWMutex
	totalOperations     int64
	optimizedOperations int64
	cacheHits           int64

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// OptimizationConfig controls optimization behavior
type OptimizationConfig struct {
	EnableFastParser      bool
	EnableParseCache      bool
	EnableTypeCache       bool
	EnableObjectPooling   bool
	EnableLazyEvaluation  bool
	CacheSize             int
	ObjectPoolSize        int
	MaxParallelOperations int
	OptimizationLevel     OptimizationLevel
}

// OptimizationLevel controls how aggressively to optimize
type OptimizationLevel int

const (
	OptimizationConservative OptimizationLevel = iota
	OptimizationBalanced
	OptimizationAggressive
)

// NewOptimizedPipeline creates a fully optimized parsing and analysis pipeline
func NewOptimizedPipeline(baseParser parser.Parser, config *OptimizationConfig) *OptimizedPipeline {
	if config == nil {
		config = DefaultOptimizationConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create optimized components
	fastParser := NewFastParser(baseParser)
	optimizedParser := NewOptimizedParser(baseParser)
	memBinder := NewMemoryOptimizedBinder()

	// Create caches and pools
	parseCache := NewParseCache(config.CacheSize)
	typeCache := NewTypeCache(config.CacheSize)
	objectPool := NewObjectPool()
	lazyResolver := NewLazyTypeResolver()

	// Create monitoring
	perfMonitor := NewPerformanceMonitor()

	return &OptimizedPipeline{
		fastParser:      fastParser,
		optimizedParser: optimizedParser,
		memBinder:       memBinder,
		parseCache:      parseCache,
		typeCache:       typeCache,
		objectPool:      objectPool,
		lazyResolver:    lazyResolver,
		perfMonitor:     perfMonitor,
		config:          config,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// DefaultOptimizationConfig returns a balanced optimization configuration
func DefaultOptimizationConfig() *OptimizationConfig {
	return &OptimizationConfig{
		EnableFastParser:      true,
		EnableParseCache:      true,
		EnableTypeCache:       true,
		EnableObjectPooling:   true,
		EnableLazyEvaluation:  true,
		CacheSize:             1000,
		ObjectPoolSize:        500,
		MaxParallelOperations: 4,
		OptimizationLevel:     OptimizationBalanced,
	}
}

// ProcessFile processes a single file through the optimized pipeline
func (op *OptimizedPipeline) ProcessFile(filename, content string) (*ProcessingResult, error) {
	start := time.Now()

	defer func() {
		op.mu.Lock()
		op.totalOperations++
		op.mu.Unlock()

		duration := time.Since(start)
		op.perfMonitor.RecordParse(duration, false) // Will be updated with cache info
	}()

	// Step 1: Optimized parsing
	astResult, parseFromCache, err := op.optimizedParse(content)
	if err != nil {
		return nil, err
	}

	// Step 2: Optimized symbol binding
	symbolTable, bindFromCache, err := op.optimizedBind(astResult)
	if err != nil {
		return nil, err
	}

	// Step 3: Optimized type checking (if enabled)
	var typeErrors []string
	var typeFromCache bool
	if op.config.OptimizationLevel >= OptimizationBalanced {
		typeErrors, typeFromCache, err = op.optimizedTypeCheck(astResult, symbolTable)
		if err != nil {
			return nil, err
		}
	}

	// Update statistics
	if parseFromCache || bindFromCache || typeFromCache {
		op.mu.Lock()
		op.cacheHits++
		op.optimizedOperations++
		op.mu.Unlock()
	}

	return &ProcessingResult{
		Filename:    filename,
		AST:         astResult,
		SymbolTable: symbolTable,
		TypeErrors:  typeErrors,
		ParseTime:   time.Since(start),
		CacheHits:   parseFromCache || bindFromCache || typeFromCache,
	}, nil
}

// ProcessingResult contains the results of processing a file
type ProcessingResult struct {
	Filename    string
	AST         *ast.AST
	SymbolTable *binder.SymbolTable
	TypeErrors  []string
	ParseTime   time.Duration
	CacheHits   bool
}

// optimizedParse performs parsing with all optimizations enabled
func (op *OptimizedPipeline) optimizedParse(content string) (*ast.AST, bool, error) {
	// Try cache first if enabled
	if op.config.EnableParseCache {
		contentHash := fastHash([]byte(content))
		if entry, found := op.parseCache.Get(content, contentHash); found {
			return entry.AST, true, nil
		}
	}

	// Try fast parser for simple content
	if op.config.EnableFastParser {
		if result, err := op.fastParser.ParseString(content); err == nil {
			// Cache the result if caching is enabled
			if op.config.EnableParseCache {
				contentHash := fastHash([]byte(content))
				op.parseCache.Put(content, result, nil, contentHash)
			}
			return result, false, nil
		}
	}

	// Fall back to optimized parser
	result, err := op.optimizedParser.ParseString(content)
	if err != nil {
		return nil, false, err
	}

	// Cache the result if caching is enabled
	if op.config.EnableParseCache {
		contentHash := fastHash([]byte(content))
		op.parseCache.Put(content, result, nil, contentHash)
	}

	return result, false, nil
}

// optimizedBind performs symbol binding with optimizations
func (op *OptimizedPipeline) optimizedBind(astResult *ast.AST) (*binder.SymbolTable, bool, error) {
	// Create binder with optimizations
	b := binder.NewBinder()

	// Use object pool for temporary allocations if enabled
	if op.config.EnableObjectPooling {
		// Get reusable objects from pool
		symbolMap := op.objectPool.GetSymbolMap()
		defer op.objectPool.PutSymbolMap(symbolMap)
	}

	// Perform binding
	symbolTable, err := b.BindAST(astResult)
	if err != nil {
		return nil, false, err
	}

	return symbolTable, false, nil
}

// optimizedTypeCheck performs type checking with optimizations
func (op *OptimizedPipeline) optimizedTypeCheck(astResult *ast.AST, symbolTable *binder.SymbolTable) ([]string, bool, error) {
	// Create type checker with optimized storage
	store, err := typedef.NewStorage()
	if err != nil {
		return nil, false, err
	}

	hierarchy := typedef.NewTypeHierarchy(store)
	tc := typechecker.NewTypeChecker(hierarchy, symbolTable, "optimized_module")

	// Perform type checking
	errors := tc.CheckAST(astResult)

	// Convert errors to strings
	var errorStrings []string
	for _, err := range errors {
		errorStrings = append(errorStrings, err.Error())
	}

	return errorStrings, false, nil
}

// ProcessBatch processes multiple files in parallel with optimization
func (op *OptimizedPipeline) ProcessBatch(files map[string]string) ([]*ProcessingResult, error) {
	results := make([]*ProcessingResult, 0, len(files))
	resultsCh := make(chan *ProcessingResult, len(files))
	errorsCh := make(chan error, len(files))

	// Create worker pool
	maxWorkers := op.config.MaxParallelOperations
	if maxWorkers <= 0 {
		maxWorkers = 1
	}

	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	// Process files in parallel
	for filename, content := range files {
		wg.Add(1)
		go func(fname, fcontent string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Process file
			result, err := op.ProcessFile(fname, fcontent)
			if err != nil {
				errorsCh <- err
				return
			}

			resultsCh <- result
		}(filename, content)
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultsCh)
		close(errorsCh)
	}()

	// Collect results
	for result := range resultsCh {
		results = append(results, result)
	}

	// Check for errors
	select {
	case err := <-errorsCh:
		return results, err
	default:
		return results, nil
	}
}

// GetOptimizationStats returns comprehensive optimization statistics
func (op *OptimizedPipeline) GetOptimizationStats() *OptimizationStats {
	op.mu.RLock()
	defer op.mu.RUnlock()

	// Get cache statistics
	parseHits, parseMisses, parseHitRate := op.parseCache.GetStats()
	typeHits, typeMisses, typeHitRate, typeCacheSize := op.typeCache.GetCacheStats()

	// Get parser statistics
	fastCount, fallbackCount, fastPercentage := op.fastParser.GetStats()

	// Get performance statistics
	perfOps, perfCacheRate, avgTime := op.perfMonitor.GetStats()

	return &OptimizationStats{
		TotalOperations:     op.totalOperations,
		OptimizedOperations: op.optimizedOperations,
		OptimizationRate:    float64(op.optimizedOperations) / float64(op.totalOperations),

		ParseCacheHits:    parseHits,
		ParseCacheMisses:  parseMisses,
		ParseCacheHitRate: parseHitRate,

		TypeCacheHits:    typeHits,
		TypeCacheMisses:  typeMisses,
		TypeCacheHitRate: typeHitRate,
		TypeCacheSize:    typeCacheSize,

		FastParseCount:      fastCount,
		FallbackParseCount:  fallbackCount,
		FastParsePercentage: fastPercentage,

		PerformanceOperations: perfOps,
		PerformanceCacheRate:  perfCacheRate,
		AverageParseTime:      avgTime,
	}
}

// OptimizationStats contains comprehensive optimization statistics
type OptimizationStats struct {
	TotalOperations     int64
	OptimizedOperations int64
	OptimizationRate    float64

	ParseCacheHits    int64
	ParseCacheMisses  int64
	ParseCacheHitRate float64

	TypeCacheHits    int64
	TypeCacheMisses  int64
	TypeCacheHitRate float64
	TypeCacheSize    int

	FastParseCount      int64
	FallbackParseCount  int64
	FastParsePercentage float64

	PerformanceOperations int64
	PerformanceCacheRate  float64
	AverageParseTime      time.Duration
}

// Shutdown gracefully shuts down the optimized pipeline
func (op *OptimizedPipeline) Shutdown() {
	op.cancel()
}

// WarmupCache pre-warms caches with common operations
func (op *OptimizedPipeline) WarmupCache(commonFiles []string) error {
	for _, content := range commonFiles {
		_, err := op.ProcessFile("warmup", content)
		if err != nil {
			return err
		}
	}
	return nil
}

// TunePerformance automatically adjusts optimization parameters based on usage patterns
func (op *OptimizedPipeline) TunePerformance() {
	stats := op.GetOptimizationStats()

	// Adjust cache sizes based on hit rates
	if stats.ParseCacheHitRate < 0.5 && op.config.CacheSize < 5000 {
		op.config.CacheSize *= 2
		op.parseCache = NewParseCache(op.config.CacheSize)
	}

	if stats.TypeCacheHitRate < 0.5 && op.config.CacheSize < 5000 {
		op.typeCache = NewTypeCache(op.config.CacheSize)
	}

	// Adjust optimization level based on fast parse success rate
	if stats.FastParsePercentage > 80 {
		op.config.OptimizationLevel = OptimizationAggressive
	} else if stats.FastParsePercentage < 20 {
		op.config.OptimizationLevel = OptimizationConservative
	}
}
