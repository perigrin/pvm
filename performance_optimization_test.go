// ABOUTME: Performance optimization tests for PVI command refactoring validation
// ABOUTME: Measures memory usage and CPU performance of extracted packages
package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"testing"
	"time"

	"tamarou.com/pvm/internal/cli/progress"
	"tamarou.com/pvm/internal/dependencies"
	"tamarou.com/pvm/internal/modules"
)

// BenchmarkMemoryUsage_ModuleOperations measures memory allocations
func BenchmarkMemoryUsage_ModuleOperations(b *testing.B) {
	// Force garbage collection before benchmark
	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate common module operations
		module := &modules.Module{
			Name:        fmt.Sprintf("Test::Module%d", i),
			Version:     "1.00",
			Description: "Test module for memory profiling",
			Author:      "TEST",
		}

		// JSON operations
		data, _ := json.Marshal(module)
		var unmarshaled modules.Module
		json.Unmarshal(data, &unmarshaled)

		// Progress tracking
		tracker := progress.NewTracker()
		tracker.Start("test-operation", 100)
		tracker.Update(50, "Halfway done")
		tracker.Finish(&progress.Result{
			Success: true,
			Message: "Completed",
		})
	}

	runtime.ReadMemStats(&m2)
	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
}

// BenchmarkCPUUsage_DependencyProcessing measures CPU performance
func BenchmarkCPUUsage_DependencyProcessing(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create dependency data
		cpanfile := &dependencies.CPANFile{
			Requirements: make([]dependencies.Requirement, 0, 100),
		}

		// Add many dependencies
		for j := 0; j < 100; j++ {
			req := dependencies.Requirement{
				Module:       fmt.Sprintf("Dep::Module%d", j),
				Version:      "1.00",
				Phase:        "runtime",
				Relationship: "requires",
			}
			cpanfile.Requirements = append(cpanfile.Requirements, req)
		}

		// Process dependencies
		for _, req := range cpanfile.Requirements {
			_ = fmt.Sprintf("%s-%s", req.Module, req.Version)
		}
	}
}

// BenchmarkScalability_LargeOperations measures performance with large datasets
func BenchmarkScalability_LargeOperations(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Create large module list
				moduleList := make([]modules.Module, size)
				for j := 0; j < size; j++ {
					moduleList[j] = modules.Module{
						Name:        fmt.Sprintf("Module%04d", j),
						Version:     "1.00",
						Description: "Scalability test module",
					}
				}

				// Process modules
				for _, mod := range moduleList {
					_ = fmt.Sprintf("%s-%s", mod.Name, mod.Version)
				}
			}
		})
	}
}

// BenchmarkParallelism_ConcurrentOperations measures concurrent performance
func BenchmarkParallelism_ConcurrentOperations(b *testing.B) {
	b.SetParallelism(10) // Set parallelism level
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate concurrent module operations
			tracker := progress.NewTracker()
			tracker.Start("parallel-op", 10)

			for i := 0; i < 10; i++ {
				tracker.Update(i, fmt.Sprintf("Step %d", i))
				time.Sleep(time.Microsecond) // Simulate work
			}

			tracker.Finish(&progress.Result{
				Success: true,
				Message: "Parallel operation completed",
			})
		}
	})
}

// BenchmarkGarbageCollection_Impact measures GC impact
func BenchmarkGarbageCollection_Impact(b *testing.B) {
	// Track GC stats
	var gcBefore, gcAfter runtime.MemStats
	runtime.ReadMemStats(&gcBefore)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create many temporary objects
		data := make([]byte, 1024)
		for j := 0; j < len(data); j++ {
			data[j] = byte(j % 256)
		}

		// Force some allocations
		temp := make(map[string]interface{})
		for j := 0; j < 10; j++ {
			temp[fmt.Sprintf("key%d", j)] = fmt.Sprintf("value%d", j)
		}
	}

	runtime.ReadMemStats(&gcAfter)
	b.ReportMetric(float64(gcAfter.NumGC-gcBefore.NumGC), "gc_cycles")
}
