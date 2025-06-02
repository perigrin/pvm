// ABOUTME: Performance benchmarks and regression tests for MCP server
// ABOUTME: Establishes baseline performance expectations and catches regressions

package mcp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"tamarou.com/pvm/internal/config"
)

// BenchmarkMCPServer_HealthCheck measures health check performance
func BenchmarkMCPServer_HealthCheck(b *testing.B) {
	server := createTestServer(b)
	defer shutdownTestServer(server, b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.healthMonitor.GetHealth()
	}
}

// BenchmarkMCPServer_MetricsCollection measures metrics collection performance
func BenchmarkMCPServer_MetricsCollection(b *testing.B) {
	server := createTestServer(b)
	defer shutdownTestServer(server, b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.performanceManager.GetAllStats()
	}
}

// BenchmarkMCPServer_CircuitBreaker measures circuit breaker performance
func BenchmarkMCPServer_CircuitBreaker(b *testing.B) {
	server := createTestServer(b)
	defer shutdownTestServer(server, b)

	breaker := server.circuitManager.GetBreaker("bench_breaker", DefaultCircuitBreakerConfig())
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = breaker.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
	}
}

// BenchmarkMCPServer_DegradationCache measures degradation cache performance
func BenchmarkMCPServer_DegradationCache(b *testing.B) {
	cache := NewFallbackCache(time.Hour, 1000)

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%100)
		_, _ = cache.Get(key)
	}
}

// BenchmarkMCPServer_RequestQueue measures request queue performance
func BenchmarkMCPServer_RequestQueue(b *testing.B) {
	config := DefaultPerformanceConfig()
	config.MaxConcurrentRequests = 10
	config.QueueSize = 100

	metrics := NewPerformanceMetrics()
	queue := NewRequestQueue(config, metrics)

	ctx := context.Background()
	queue.Start(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		requestID := fmt.Sprintf("bench_request_%d", i)
		_ = queue.Submit(ctx, requestID, func(ctx context.Context) error {
			return nil
		})
	}
}

// BenchmarkMCPServer_ConcurrentRequests measures concurrent request handling
func BenchmarkMCPServer_ConcurrentRequests(b *testing.B) {
	server := createTestServer(b)
	defer shutdownTestServer(server, b)

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		requestID := 0
		for pb.Next() {
			requestID++
			id := fmt.Sprintf("concurrent_request_%d", requestID)
			_ = server.performanceManager.ExecuteWithOptimization(ctx, id, func(ctx context.Context) error {
				// Simulate minimal work
				time.Sleep(1 * time.Microsecond)
				return nil
			})
		}
	})
}

// Performance regression tests
func TestMCPServer_PerformanceRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression test in short mode")
	}

	server := createTestServer(t)
	defer shutdownTestServer(server, t)

	// Test 1: Health check should complete in under 10ms
	start := time.Now()
	_ = server.healthMonitor.GetHealth()
	duration := time.Since(start)

	if duration > 10*time.Millisecond {
		t.Errorf("Health check took too long: %v (expected < 10ms)", duration)
	}

	// Test 2: Metrics collection should complete in under 50ms
	start = time.Now()
	_ = server.performanceManager.GetAllStats()
	duration = time.Since(start)

	if duration > 50*time.Millisecond {
		t.Errorf("Metrics collection took too long: %v (expected < 50ms)", duration)
	}

	// Test 3: Circuit breaker execution should complete in under 1ms
	breaker := server.circuitManager.GetBreaker("regression_breaker", DefaultCircuitBreakerConfig())
	ctx := context.Background()

	start = time.Now()
	_ = breaker.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	duration = time.Since(start)

	if duration > 1*time.Millisecond {
		t.Errorf("Circuit breaker execution took too long: %v (expected < 1ms)", duration)
	}

	// Test 4: Degradation cache operations should complete in under 1ms
	start = time.Now()
	server.degradationManager.CacheResult("test_operation", "test_result")
	duration = time.Since(start)

	if duration > 1*time.Millisecond {
		t.Errorf("Cache write took too long: %v (expected < 1ms)", duration)
	}

	start = time.Now()
	_, _ = server.degradationManager.cache.Get("test_operation")
	duration = time.Since(start)

	if duration > 1*time.Millisecond {
		t.Errorf("Cache read took too long: %v (expected < 1ms)", duration)
	}

	t.Log("All performance regression tests passed")
}

// Memory usage regression test
func TestMCPServer_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	server := createTestServer(t)
	defer shutdownTestServer(server, t)

	// Get initial memory stats
	initialStats := server.performanceManager.GetResourceMonitor().GetStats()
	initialMemory := uint64(0)
	if memUsage := initialStats["memory_usage_mb"]; memUsage != nil {
		// Type switch to handle the interface{} value
		switch v := memUsage.(type) {
		case uint64:
			initialMemory = v
		case int64:
			initialMemory = uint64(v)
		case int:
			initialMemory = uint64(v)
		default:
			t.Fatal("Could not get initial memory usage")
		}
	}

	// Perform operations that might allocate memory
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		requestID := fmt.Sprintf("memory_test_%d", i)
		_ = server.performanceManager.ExecuteWithOptimization(ctx, requestID, func(ctx context.Context) error {
			// Simulate some memory allocation
			data := make([]byte, 1024)
			_ = data
			return nil
		})

		// Add items to degradation cache
		server.degradationManager.CacheResult(fmt.Sprintf("operation_%d", i), fmt.Sprintf("result_%d", i))

		// Get health status
		_ = server.healthMonitor.GetHealth()
	}

	// Get final memory stats
	finalStats := server.performanceManager.GetResourceMonitor().GetStats()
	finalMemory := uint64(0)
	if memUsage := finalStats["memory_usage_mb"]; memUsage != nil {
		// Type switch to handle the interface{} value
		switch v := memUsage.(type) {
		case uint64:
			finalMemory = v
		case int64:
			finalMemory = uint64(v)
		case int:
			finalMemory = uint64(v)
		default:
			t.Fatal("Could not get final memory usage")
		}
	}

	memoryIncrease := finalMemory - initialMemory
	t.Logf("Memory usage increased by %d MB during test", memoryIncrease)

	// Check for reasonable memory usage (should not increase by more than 50MB for this test)
	if memoryIncrease > 50 {
		t.Errorf("Memory usage increased too much: %d MB (expected < 50MB)", memoryIncrease)
	}
}

// Goroutine leak detection test
func TestMCPServer_GoroutineLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping goroutine leak test in short mode")
	}

	server := createTestServer(t)

	// Get initial goroutine count
	initialStats := server.performanceManager.GetResourceMonitor().GetStats()
	initialGoroutines := 0
	if gorCount := initialStats["goroutine_count"]; gorCount != nil {
		// Type switch to handle the interface{} value
		switch v := gorCount.(type) {
		case int:
			initialGoroutines = v
		case int64:
			initialGoroutines = int(v)
		default:
			t.Fatal("Could not get initial goroutine count")
		}
	}

	// Perform operations that might create goroutines
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		requestID := fmt.Sprintf("goroutine_test_%d", i)
		_ = server.performanceManager.ExecuteWithOptimization(ctx, requestID, func(ctx context.Context) error {
			time.Sleep(1 * time.Millisecond)
			return nil
		})
	}

	// Wait for operations to complete
	time.Sleep(100 * time.Millisecond)

	// Shutdown server
	shutdownTestServer(server, t)

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Create a new server to get clean goroutine count
	cleanServer := createTestServer(t)
	cleanStats := cleanServer.performanceManager.GetResourceMonitor().GetStats()
	cleanGoroutines := 0
	if gorCount := cleanStats["goroutine_count"]; gorCount != nil {
		// Type switch to handle the interface{} value
		switch v := gorCount.(type) {
		case int:
			cleanGoroutines = v
		case int64:
			cleanGoroutines = int(v)
		default:
			t.Fatal("Could not get clean goroutine count")
		}
	}
	shutdownTestServer(cleanServer, t)

	goroutineIncrease := cleanGoroutines - initialGoroutines
	t.Logf("Goroutine count difference: %d", goroutineIncrease)

	// Allow for some variance in goroutine count (up to 5 extra goroutines)
	if goroutineIncrease > 5 {
		t.Errorf("Potential goroutine leak detected: %d extra goroutines", goroutineIncrease)
	}
}

// Helper functions
func createTestServer(tb testing.TB) *Server {
	cfg := &config.Config{
		MCP: &config.MCPConfig{
			Port:                  3100,
			Host:                  "localhost",
			AutoDiscoverProjects:  false,
			EmbeddingProvider:     "local",
			ValidationCacheSize:   "5MB",
			EmbeddingCacheSize:    "5MB",
			GenerationMemorySize:  10,
			MaxConcurrentRequests: 5,
			RequestTimeout:        5 * time.Second,
		},
	}

	server, err := NewServer(cfg)
	if err != nil {
		tb.Fatalf("Failed to create test server: %v", err)
	}

	// Start the server to properly initialize all components
	ctx := context.Background()
	err = server.Start(ctx)
	if err != nil {
		tb.Fatalf("Failed to start test server: %v", err)
	}

	return server
}

func shutdownTestServer(server *Server, tb testing.TB) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := server.Stop(ctx)
	if err != nil {
		tb.Errorf("Failed to shutdown test server: %v", err)
	}
}
