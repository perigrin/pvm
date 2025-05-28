// ABOUTME: Simplified performance benchmarking system for core object pools
// ABOUTME: Provides basic benchmark comparison between pooled and direct allocation

package monitoring

import (
	"context"
	"runtime"
	"sync"
	"time"

	"tamarou.com/pvm/internal/core"
)

// SimpleBenchmarkSuite manages basic performance benchmarks for core pools
type SimpleBenchmarkSuite struct {
	mu      sync.RWMutex
	results []BenchmarkResult
}

// NewSimpleBenchmarkSuite creates a new simplified benchmark suite
func NewSimpleBenchmarkSuite() *SimpleBenchmarkSuite {
	return &SimpleBenchmarkSuite{
		results: make([]BenchmarkResult, 0),
	}
}

// RunCorePoolBenchmarks runs benchmarks for core pool functionality
func (bs *SimpleBenchmarkSuite) RunCorePoolBenchmarks(ctx context.Context) []BenchmarkResult {
	var results []BenchmarkResult

	configs := []BenchmarkConfig{
		{
			Operations:        1000,
			Concurrency:       1,
			ObjectSize:        64,
			AllocationPattern: "sequential",
			TestDuration:      time.Second * 5,
		},
		{
			Operations:        5000,
			Concurrency:       4,
			ObjectSize:        128,
			AllocationPattern: "concurrent",
			TestDuration:      time.Second * 10,
		},
		{
			Operations:        10000,
			Concurrency:       8,
			ObjectSize:        32,
			AllocationPattern: "high_concurrency",
			TestDuration:      time.Second * 15,
		},
	}

	for _, config := range configs {
		// Test pooled allocation
		pooledResult := bs.benchmarkCorePooled(config)
		pooledResult.Name = "CorePool_Pooled_" + config.AllocationPattern

		// Test direct allocation
		directResult := bs.benchmarkCoreDirect(config)
		directResult.Name = "CorePool_Direct_" + config.AllocationPattern

		// Create comparison result
		comparison := BenchmarkResult{
			Name:              "CorePool_Comparison_" + config.AllocationPattern,
			Timestamp:         time.Now(),
			PooledAllocations: pooledResult.PooledAllocations,
			DirectAllocations: directResult.DirectAllocations,
			TestConfiguration: config,
		}

		// Calculate improvement
		if directResult.DirectAllocations.TotalTime > 0 {
			timeSaving := directResult.DirectAllocations.TotalTime - pooledResult.PooledAllocations.TotalTime
			comparison.ImprovementPercent = float64(timeSaving) / float64(directResult.DirectAllocations.TotalTime) * 100
		}

		// Calculate memory savings
		memorySaving := directResult.DirectAllocations.MemoryAllocated - pooledResult.PooledAllocations.MemoryAllocated
		comparison.MemorySavings = memorySaving

		results = append(results, comparison)
	}

	bs.mu.Lock()
	bs.results = append(bs.results, results...)
	bs.mu.Unlock()

	return results
}

// benchmarkCorePooled benchmarks core pool allocation
func (bs *SimpleBenchmarkSuite) benchmarkCorePooled(config BenchmarkConfig) BenchmarkResult {
	var pool core.Pool[TestObject]

	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	start := time.Now()

	// Run allocation benchmark
	var wg sync.WaitGroup
	allocationsPerWorker := config.Operations / config.Concurrency

	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < allocationsPerWorker; j++ {
				// Allocate from pool
				obj := pool.New()
				obj.ID = int64(workerID*1000 + j)
				obj.Data = "test_data"
				obj.Value = float64(j)

				// Simulate usage
				_ = obj.ID
				_ = obj.Data
				_ = obj.Value
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	return BenchmarkResult{
		PooledAllocations: BenchmarkMetrics{
			TotalTime:        duration,
			OperationsPerSec: float64(config.Operations) / duration.Seconds(),
			AllocationsCount: int64(config.Operations),
			MemoryAllocated:  int64(memAfter.TotalAlloc - memBefore.TotalAlloc),
			GCPauses:         int(memAfter.NumGC - memBefore.NumGC),
			CPUUsage:         calculateCPUUsage(start, duration),
		},
		TestConfiguration: config,
		Timestamp:         time.Now(),
	}
}

// benchmarkCoreDirect benchmarks direct allocation
func (bs *SimpleBenchmarkSuite) benchmarkCoreDirect(config BenchmarkConfig) BenchmarkResult {
	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	start := time.Now()

	// Run allocation benchmark
	var wg sync.WaitGroup
	allocationsPerWorker := config.Operations / config.Concurrency

	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < allocationsPerWorker; j++ {
				// Allocate directly
				obj := &TestObject{
					ID:    int64(workerID*1000 + j),
					Data:  "test_data",
					Value: float64(j),
				}

				// Simulate usage
				_ = obj.ID
				_ = obj.Data
				_ = obj.Value
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	return BenchmarkResult{
		DirectAllocations: BenchmarkMetrics{
			TotalTime:        duration,
			OperationsPerSec: float64(config.Operations) / duration.Seconds(),
			AllocationsCount: int64(config.Operations),
			MemoryAllocated:  int64(memAfter.TotalAlloc - memBefore.TotalAlloc),
			GCPauses:         int(memAfter.NumGC - memBefore.NumGC),
			CPUUsage:         calculateCPUUsage(start, duration),
		},
		TestConfiguration: config,
		Timestamp:         time.Now(),
	}
}

// TestObject represents a test object for benchmarking
type TestObject struct {
	ID    int64
	Data  string
	Value float64
}

// RunMemoryEfficiencyBenchmarks runs benchmarks focused on memory efficiency
func (bs *SimpleBenchmarkSuite) RunMemoryEfficiencyBenchmarks(ctx context.Context) []BenchmarkResult {
	var results []BenchmarkResult

	// Different object sizes to test memory efficiency
	sizes := []struct {
		name string
		size int
	}{
		{"small", 64},
		{"medium", 256},
		{"large", 1024},
	}

	for _, size := range sizes {
		config := BenchmarkConfig{
			Operations:        2000,
			Concurrency:       4,
			ObjectSize:        size.size,
			AllocationPattern: "memory_efficiency",
			TestDuration:      time.Second * 10,
		}

		// Run benchmark for this size
		pooledResult := bs.benchmarkMemoryPooled(config)
		pooledResult.Name = "Memory_Pooled_" + size.name

		directResult := bs.benchmarkMemoryDirect(config)
		directResult.Name = "Memory_Direct_" + size.name

		// Create comparison
		comparison := BenchmarkResult{
			Name:              "Memory_Comparison_" + size.name,
			Timestamp:         time.Now(),
			PooledAllocations: pooledResult.PooledAllocations,
			DirectAllocations: directResult.DirectAllocations,
			TestConfiguration: config,
		}

		// Calculate improvements
		if directResult.DirectAllocations.TotalTime > 0 {
			timeSaving := directResult.DirectAllocations.TotalTime - pooledResult.PooledAllocations.TotalTime
			comparison.ImprovementPercent = float64(timeSaving) / float64(directResult.DirectAllocations.TotalTime) * 100
		}

		memorySaving := directResult.DirectAllocations.MemoryAllocated - pooledResult.PooledAllocations.MemoryAllocated
		comparison.MemorySavings = memorySaving

		results = append(results, comparison)
	}

	return results
}

// benchmarkMemoryPooled benchmarks memory allocation with pools
func (bs *SimpleBenchmarkSuite) benchmarkMemoryPooled(config BenchmarkConfig) BenchmarkResult {
	var pool core.Pool[byte]

	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	start := time.Now()

	// Run allocation benchmark
	var wg sync.WaitGroup
	allocationsPerWorker := config.Operations / config.Concurrency

	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < allocationsPerWorker; j++ {
				// Allocate bytes from pool
				for k := 0; k < config.ObjectSize; k++ {
					bytePtr := pool.New()
					*bytePtr = byte(k % 256)

					// Simulate usage
					_ = *bytePtr
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	return BenchmarkResult{
		PooledAllocations: BenchmarkMetrics{
			TotalTime:        duration,
			OperationsPerSec: float64(config.Operations) / duration.Seconds(),
			AllocationsCount: int64(config.Operations * config.ObjectSize),
			MemoryAllocated:  int64(memAfter.TotalAlloc - memBefore.TotalAlloc),
			GCPauses:         int(memAfter.NumGC - memBefore.NumGC),
			CPUUsage:         calculateCPUUsage(start, duration),
		},
		TestConfiguration: config,
		Timestamp:         time.Now(),
	}
}

// benchmarkMemoryDirect benchmarks direct memory allocation
func (bs *SimpleBenchmarkSuite) benchmarkMemoryDirect(config BenchmarkConfig) BenchmarkResult {
	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	start := time.Now()

	// Run allocation benchmark
	var wg sync.WaitGroup
	allocationsPerWorker := config.Operations / config.Concurrency

	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < allocationsPerWorker; j++ {
				// Allocate bytes directly
				for k := 0; k < config.ObjectSize; k++ {
					byteVal := byte(k % 256)

					// Simulate usage
					_ = byteVal
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	return BenchmarkResult{
		DirectAllocations: BenchmarkMetrics{
			TotalTime:        duration,
			OperationsPerSec: float64(config.Operations) / duration.Seconds(),
			AllocationsCount: int64(config.Operations * config.ObjectSize),
			MemoryAllocated:  int64(memAfter.TotalAlloc - memBefore.TotalAlloc),
			GCPauses:         int(memAfter.NumGC - memBefore.NumGC),
			CPUUsage:         calculateCPUUsage(start, duration),
		},
		TestConfiguration: config,
		Timestamp:         time.Now(),
	}
}

// GetAllResults returns all benchmark results
func (bs *SimpleBenchmarkSuite) GetAllResults() []BenchmarkResult {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	results := make([]BenchmarkResult, len(bs.results))
	copy(results, bs.results)
	return results
}

// ClearResults clears all benchmark results
func (bs *SimpleBenchmarkSuite) ClearResults() {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.results = bs.results[:0]
}

// CalculateCPUUsage calculates CPU usage during the benchmark
func calculateCPUUsage(start time.Time, duration time.Duration) float64 {
	// Simplified CPU usage calculation
	// In a real implementation, this would use more sophisticated CPU monitoring
	return float64(duration.Nanoseconds()) / float64(time.Since(start).Nanoseconds()) * 100
}
