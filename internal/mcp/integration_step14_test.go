// ABOUTME: Integration tests for Step 14 - Performance optimization and health monitoring
// ABOUTME: Tests circuit breakers, health checks, graceful degradation, and performance features

package mcp

import (
	"context"
	"testing"
	"time"
)

func TestStep14_HealthMonitor(t *testing.T) {
	monitor := NewHealthMonitor()

	// Test basic health monitor functionality
	if monitor == nil {
		t.Fatal("Expected health monitor to be created")
	}

	// Register a test checker
	testChecker := &TestHealthChecker{healthy: true}
	monitor.RegisterChecker("test_component", testChecker)

	// Check that health report works
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Perform one health check
	monitor.performHealthCheck(ctx)

	report := monitor.GetHealth()
	if report.Status != StatusHealthy {
		t.Errorf("Expected healthy status, got %s", report.Status)
	}

	if len(report.Components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(report.Components))
	}

	if report.Components["test_component"].Status != StatusHealthy {
		t.Errorf("Expected test component to be healthy")
	}
}

func TestStep14_CircuitBreaker(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.MaxFailures = 2
	config.ResetTimeout = 100 * time.Millisecond

	breaker := NewCircuitBreaker("test_breaker", config)

	// Test circuit breaker starts closed
	if breaker.GetState() != StateClosed {
		t.Errorf("Expected circuit breaker to start closed, got %s", breaker.GetState())
	}

	// Test successful execution
	ctx := context.Background()
	err := breaker.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected successful execution, got error: %v", err)
	}

	// Test circuit opens after failures
	for i := 0; i < 3; i++ {
		err = breaker.Execute(ctx, func(ctx context.Context) error {
			return context.DeadlineExceeded
		})
	}

	if breaker.GetState() != StateOpen {
		t.Errorf("Expected circuit breaker to be open after failures, got %s", breaker.GetState())
	}

	// Test circuit rejects requests when open
	err = breaker.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != ErrCircuitOpen {
		t.Errorf("Expected circuit open error, got: %v", err)
	}
}

func TestStep14_PerformanceManager(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.MaxConcurrentRequests = 2

	manager := NewPerformanceManager(config)

	if manager == nil {
		t.Fatal("Expected performance manager to be created")
	}

	// Test that all components are initialized
	if manager.GetRequestQueue() == nil {
		t.Error("Expected request queue to be initialized")
	}

	if manager.GetConnectionPool() == nil {
		t.Error("Expected connection pool to be initialized")
	}

	if manager.GetResourceMonitor() == nil {
		t.Error("Expected resource monitor to be initialized")
	}

	if manager.GetCircuitManager() == nil {
		t.Error("Expected circuit manager to be initialized")
	}

	// Test metrics collection
	stats := manager.GetAllStats()
	if stats == nil {
		t.Error("Expected stats to be available")
	}

	// Check that stats contain expected sections
	if _, ok := stats["request_queue"]; !ok {
		t.Error("Expected request_queue stats")
	}

	if _, ok := stats["connection_pool"]; !ok {
		t.Error("Expected connection_pool stats")
	}

	if _, ok := stats["resource_monitor"]; !ok {
		t.Error("Expected resource_monitor stats")
	}
}

func TestStep14_DegradationManager(t *testing.T) {
	config := DefaultDegradationConfig()
	logger := &TestLogger{}

	manager := NewDegradationManager(config, logger)

	if manager == nil {
		t.Fatal("Expected degradation manager to be created")
	}

	ctx := context.Background()

	// Test embedding failure handling with cache strategy
	manager.config.EmbeddingFailureStrategy = StrategyCache
	result, err := manager.HandleEmbeddingFailure(ctx, "test_operation", context.DeadlineExceeded)

	// Should fail since no cache is available
	if err == nil {
		t.Error("Expected error when no cache is available")
	}

	// Test fallback strategy
	manager.config.EmbeddingFailureStrategy = StrategyFallback
	result, err = manager.HandleEmbeddingFailure(ctx, "test_operation", context.DeadlineExceeded)

	if err != nil {
		t.Errorf("Expected fallback to succeed, got error: %v", err)
	}

	if result == nil {
		t.Error("Expected fallback result")
	}

	// Test skip strategy
	manager.config.EmbeddingFailureStrategy = StrategySkip
	result, err = manager.HandleEmbeddingFailure(ctx, "test_operation", context.DeadlineExceeded)

	if err != nil {
		t.Errorf("Expected skip to succeed, got error: %v", err)
	}

	if result != nil {
		t.Error("Expected skip to return nil result")
	}
}

func TestStep14_FallbackCache(t *testing.T) {
	cache := NewFallbackCache(100*time.Millisecond, 10)

	// Test basic set/get
	cache.Set("test_key", "test_value")

	value, found := cache.Get("test_key")
	if !found {
		t.Error("Expected to find cached value")
	}

	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %v", value)
	}

	// Test expiration
	time.Sleep(150 * time.Millisecond)

	_, found = cache.Get("test_key")
	if found {
		t.Error("Expected cached value to be expired")
	}

	// Test size limit
	for i := 0; i < 15; i++ {
		cache.Set(string(rune('a'+i)), i)
	}

	if cache.Size() > 10 {
		t.Errorf("Expected cache size to be limited to 10, got %d", cache.Size())
	}
}

// TestHealthChecker is a test implementation of HealthChecker
type TestHealthChecker struct {
	healthy bool
}

func (t *TestHealthChecker) HealthCheck(ctx context.Context) ComponentStatus {
	status := StatusHealthy
	message := "Test component is healthy"

	if !t.healthy {
		status = StatusUnhealthy
		message = "Test component is unhealthy"
	}

	return ComponentStatus{
		Name:      "test_component",
		Status:    status,
		Message:   message,
		LastCheck: time.Now(),
		Duration:  1 * time.Millisecond,
	}
}

// TestLogger is a test implementation of a logger
type TestLogger struct {
	logs []string
}

func (t *TestLogger) Printf(format string, args ...interface{}) {
	// Store log messages for testing
	t.logs = append(t.logs, format)
}
