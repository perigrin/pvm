// ABOUTME: End-to-end integration tests for MCP server workflows
// ABOUTME: Tests complete workflows from tool registration to execution with real components

package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tamarou.com/pvm/internal/config"
)

func TestMCPServer_FullWorkflow_Integration(t *testing.T) {
	// Create a temporary test project
	tempDir := t.TempDir()
	testProjectPath := filepath.Join(tempDir, "test_project")
	err := os.MkdirAll(testProjectPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create test project directory: %v", err)
	}

	// Create a sample Perl file
	samplePerl := `use v5.40;
use experimental 'class';

class Calculator {
    field $value = 0;

    method add(Int $n) {
        $value += $n;
        return $self;
    }

    method get_value() : Int {
        return $value;
    }
}

sub factorial(Int $n) : Int {
    return 1 if $n <= 1;
    return $n * factorial($n - 1);
}
`
	err = os.WriteFile(filepath.Join(testProjectPath, "calculator.pl"), []byte(samplePerl), 0644)
	if err != nil {
		t.Fatalf("Failed to create sample Perl file: %v", err)
	}

	// Create cpanfile to mark it as a Perl project
	cpanfile := `requires 'strict';
requires 'warnings';
`
	err = os.WriteFile(filepath.Join(testProjectPath, "cpanfile"), []byte(cpanfile), 0644)
	if err != nil {
		t.Fatalf("Failed to create cpanfile: %v", err)
	}

	// Change to test project directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(testProjectPath)
	if err != nil {
		t.Fatalf("Failed to change to test project directory: %v", err)
	}

	// Create test configuration
	cfg := &config.Config{
		MCP: &config.MCPConfig{
			Port:                      3000,
			Host:                      "localhost",
			AutoDiscoverProjects:      true,
			AutoFixErrors:             true,
			ValidationCacheSize:       "10MB",
			EmbeddingProvider:         "local", // Use local provider for testing
			EmbeddingCacheSize:        "10MB",
			EmbeddingModel:            "test-model",
			GenerationMemorySize:      10,
			EnableIterativeRefinement: true,
			MaxConcurrentRequests:     5,
			RequestTimeout:            10 * time.Second,
		},
	}

	// Create server instance
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create MCP server: %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}

	// Start the server to trigger project discovery
	ctx := context.Background()
	err = server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start MCP server: %v", err)
	}
	defer server.Stop(ctx)

	// Test 1: Verify project discovery
	projects := server.GetProjects()
	if len(projects) == 0 {
		t.Error("Expected at least one project to be discovered")
	}

	foundTestProject := false
	for path, project := range projects {
		if project.HasCpanfile {
			foundTestProject = true
			t.Logf("Discovered project at %s: type=%s, cpanfile=%v", path, project.ProjectType, project.HasCpanfile)
		}
	}

	if !foundTestProject {
		t.Error("Expected to find test project with cpanfile")
	}

	// Test 2: Verify health monitoring is working
	health := server.healthMonitor.GetHealth()
	if health.Status == "" {
		t.Error("Health status should not be empty")
	}

	if len(health.Components) == 0 {
		t.Error("Expected health components to be registered")
	}

	t.Logf("Health status: %s with %d components", health.Status, len(health.Components))

	// Test 3: Verify performance manager is initialized
	stats := server.performanceManager.GetAllStats()
	if stats == nil {
		t.Error("Performance stats should not be nil")
	}

	if _, ok := stats["request_queue"]; !ok {
		t.Error("Expected request_queue stats")
	}

	if _, ok := stats["resource_monitor"]; !ok {
		t.Error("Expected resource_monitor stats")
	}

	t.Logf("Performance stats collected successfully")

	// Test 4: Verify circuit breakers are set up
	breakerStatus := server.circuitManager.GetBreakersStatus()
	if breakerStatus == nil {
		t.Error("Circuit breaker status should not be nil")
	}

	// Test 5: Verify degradation manager is working
	degradationStats := server.degradationManager.GetStats()
	if degradationStats == nil {
		t.Error("Degradation stats should not be nil")
	}

	t.Logf("All major components initialized and operational")

	// Test 6: Verify graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Stop(ctx)
	if err != nil {
		t.Errorf("Server shutdown failed: %v", err)
	}

	t.Log("End-to-end workflow test completed successfully")
}

func TestMCPServer_ConcurrentOperations(t *testing.T) {
	// Test concurrent operations to verify performance optimizations
	cfg := &config.Config{
		MCP: &config.MCPConfig{
			Port:                  3001,
			Host:                  "localhost",
			AutoDiscoverProjects:  false,
			MaxConcurrentRequests: 3,
			RequestTimeout:        5 * time.Second,
			EmbeddingProvider:     "local",
			ValidationCacheSize:   "5MB",
			GenerationMemorySize:  5,
		},
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start the server to enable performance manager
	ctx := context.Background()
	err = server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop(ctx)

	// Test concurrent request handling
	numRequests := 5
	results := make(chan error, numRequests)

	// Start multiple operations concurrently
	for i := 0; i < numRequests; i++ {
		go func(id int) {
			requestID := fmt.Sprintf("test_request_%d", id)
			err := server.performanceManager.ExecuteWithOptimization(ctx, requestID, func(ctx context.Context) error {
				// Simulate some work
				time.Sleep(100 * time.Millisecond)
				return nil
			})
			results <- err
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numRequests; i++ {
		err := <-results
		if err == nil {
			successCount++
		} else {
			t.Logf("Request failed: %v", err)
		}
	}

	// With max 3 concurrent requests, some should succeed
	if successCount == 0 {
		t.Error("Expected at least some requests to succeed")
	}

	t.Logf("Concurrent operations test: %d/%d requests succeeded", successCount, numRequests)

	// Verify queue stats
	stats := server.performanceManager.GetRequestQueue().GetStats()
	if processedRequests := stats["processed_requests"]; processedRequests != nil {
		t.Logf("Queue processed %v requests", processedRequests)
	}

	// Clean shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	server.Stop(ctx)
}

func TestMCPServer_FailureRecovery(t *testing.T) {
	// Test graceful degradation and failure recovery
	cfg := &config.Config{
		MCP: &config.MCPConfig{
			Port:                 3002,
			Host:                 "localhost",
			AutoDiscoverProjects: false,
			EmbeddingProvider:    "local",
			ValidationCacheSize:  "5MB",
			RequestTimeout:       2 * time.Second,
		},
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()

	// Test embedding failure handling
	result, err := server.degradationManager.HandleEmbeddingFailure(ctx, "test_operation", context.DeadlineExceeded)
	if err != nil {
		t.Logf("Embedding failure handled with error: %v", err)
	} else {
		t.Logf("Embedding failure handled with fallback result: %v", result)
	}

	// Test validation failure handling
	result, err = server.degradationManager.HandleValidationFailure(ctx, "test_validation", context.DeadlineExceeded)
	if err != nil {
		t.Logf("Validation failure handled with error: %v", err)
	} else {
		t.Logf("Validation failure handled with fallback result: %v", result)
	}

	// Test circuit breaker functionality
	breaker := server.circuitManager.GetBreaker("test_breaker", DefaultCircuitBreakerConfig())

	// Cause failures to open the circuit
	for i := 0; i < 6; i++ {
		err := breaker.Execute(ctx, func(ctx context.Context) error {
			return context.DeadlineExceeded
		})
		if err != nil {
			t.Logf("Circuit breaker attempt %d failed as expected: %v", i+1, err)
		}
	}

	// Verify circuit is open
	if breaker.GetState() != StateOpen {
		t.Errorf("Expected circuit breaker to be open, got %s", breaker.GetState())
	}

	// Test that subsequent requests are rejected
	err = breaker.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != ErrCircuitOpen {
		t.Errorf("Expected circuit open error, got: %v", err)
	}

	t.Log("Failure recovery test completed successfully")

	// Clean shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	server.Stop(ctx)
}

func TestMCPServer_ResourceManagement(t *testing.T) {
	// Test resource monitoring and limits
	cfg := &config.Config{
		MCP: &config.MCPConfig{
			Port:                  3003,
			Host:                  "localhost",
			AutoDiscoverProjects:  false,
			EmbeddingProvider:     "local",
			MaxConcurrentRequests: 2,
			RequestTimeout:        1 * time.Second,
		},
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test resource monitoring
	resourceStats := server.performanceManager.GetResourceMonitor().GetStats()
	if resourceStats == nil {
		t.Error("Resource stats should not be nil")
	}

	if memUsage, ok := resourceStats["memory_usage_mb"]; ok {
		t.Logf("Current memory usage: %v MB", memUsage)
	}

	if goroutineCount, ok := resourceStats["goroutine_count"]; ok {
		t.Logf("Current goroutine count: %v", goroutineCount)
	}

	// Test cache management
	cacheStats := server.degradationManager.GetStats()
	if cacheStats == nil {
		t.Error("Cache stats should not be nil")
	}

	if cache := cacheStats["cache"]; cache != nil {
		t.Logf("Degradation cache info: %v", cache)
	}

	t.Log("Resource management test completed successfully")

	// Clean shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	server.Stop(ctx)
}
