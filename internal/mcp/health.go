// ABOUTME: Health check and monitoring system for MCP server
// ABOUTME: Provides health status, component checks, and operational metrics

package mcp

import (
	"context"
	"sync"
	"time"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentStatus represents the status of an individual component
type ComponentStatus struct {
	Name      string        `json:"name"`
	Status    HealthStatus  `json:"status"`
	Message   string        `json:"message,omitempty"`
	LastCheck time.Time     `json:"last_check"`
	Duration  time.Duration `json:"duration"`
}

// HealthReport represents the overall health report
type HealthReport struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentStatus `json:"components"`
	Uptime     time.Duration              `json:"uptime"`
	Version    string                     `json:"version"`
}

// HealthChecker interface for component health checks
type HealthChecker interface {
	HealthCheck(ctx context.Context) ComponentStatus
}

// HealthMonitor manages health checks for all server components
type HealthMonitor struct {
	mu            sync.RWMutex
	checkers      map[string]HealthChecker
	lastReport    HealthReport
	startTime     time.Time
	checkInterval time.Duration
	checkTimeout  time.Duration
	stopChan      chan struct{}
	stopped       bool
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor() *HealthMonitor {
	return &HealthMonitor{
		checkers:      make(map[string]HealthChecker),
		startTime:     time.Now(),
		checkInterval: 30 * time.Second, // Default check interval
		checkTimeout:  10 * time.Second, // Default check timeout
		stopChan:      make(chan struct{}),
	}
}

// RegisterChecker registers a health checker for a component
func (h *HealthMonitor) RegisterChecker(name string, checker HealthChecker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers[name] = checker
}

// Start begins the health monitoring loop
func (h *HealthMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	// Initial health check
	h.performHealthCheck(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopChan:
			return
		case <-ticker.C:
			h.performHealthCheck(ctx)
		}
	}
}

// Stop stops the health monitoring
func (h *HealthMonitor) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.stopped {
		close(h.stopChan)
		h.stopped = true
	}
}

// GetHealth returns the current health report
func (h *HealthMonitor) GetHealth() HealthReport {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a copy to avoid race conditions
	report := h.lastReport
	report.Components = make(map[string]ComponentStatus)
	for k, v := range h.lastReport.Components {
		report.Components[k] = v
	}

	return report
}

// performHealthCheck checks all registered components
func (h *HealthMonitor) performHealthCheck(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	checkCtx, cancel := context.WithTimeout(ctx, h.checkTimeout)
	defer cancel()

	components := make(map[string]ComponentStatus)
	overallStatus := StatusHealthy
	unhealthyCount := 0
	degradedCount := 0

	// Check all registered components
	for name, checker := range h.checkers {
		start := time.Now()
		status := checker.HealthCheck(checkCtx)
		status.Duration = time.Since(start)
		status.LastCheck = time.Now()

		components[name] = status

		// Update overall status based on component status
		switch status.Status {
		case StatusUnhealthy:
			unhealthyCount++
		case StatusDegraded:
			degradedCount++
		}
	}

	// Determine overall status
	if unhealthyCount > 0 {
		overallStatus = StatusUnhealthy
	} else if degradedCount > 0 {
		overallStatus = StatusDegraded
	}

	// Update last report
	h.lastReport = HealthReport{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Components: components,
		Uptime:     time.Since(h.startTime),
		Version:    "1.0.0",
	}
}

// EmbeddingHealthChecker checks embedding store health
type EmbeddingHealthChecker struct {
	store interface{} // Will be *embeddings.EmbeddingStore
}

// NewEmbeddingHealthChecker creates a new embedding health checker
func NewEmbeddingHealthChecker(store interface{}) *EmbeddingHealthChecker {
	return &EmbeddingHealthChecker{store: store}
}

// HealthCheck performs health check for embedding store
func (e *EmbeddingHealthChecker) HealthCheck(ctx context.Context) ComponentStatus {
	start := time.Now()

	// Basic availability check - in real implementation would ping store
	// For now, just check if store is not nil
	if e.store == nil {
		return ComponentStatus{
			Name:      "embedding_store",
			Status:    StatusUnhealthy,
			Message:   "Embedding store is not initialized",
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// TODO: Add actual health check logic:
	// - Test connection to chromem database
	// - Check disk space for persistent storage
	// - Verify collection accessibility

	return ComponentStatus{
		Name:      "embedding_store",
		Status:    StatusHealthy,
		Message:   "Embedding store is operational",
		LastCheck: time.Now(),
		Duration:  time.Since(start),
	}
}

// ValidatorHealthChecker checks validator component health
type ValidatorHealthChecker struct {
	validator interface{} // Will be *validation.Validator
}

// NewValidatorHealthChecker creates a new validator health checker
func NewValidatorHealthChecker(validator interface{}) *ValidatorHealthChecker {
	return &ValidatorHealthChecker{validator: validator}
}

// HealthCheck performs health check for validator
func (v *ValidatorHealthChecker) HealthCheck(ctx context.Context) ComponentStatus {
	start := time.Now()

	if v.validator == nil {
		return ComponentStatus{
			Name:      "validator",
			Status:    StatusUnhealthy,
			Message:   "Validator is not initialized",
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// TODO: Add actual health check logic:
	// - Test type checking functionality
	// - Check cache health and size
	// - Verify parser availability

	return ComponentStatus{
		Name:      "validator",
		Status:    StatusHealthy,
		Message:   "Validator is operational",
		LastCheck: time.Now(),
		Duration:  time.Since(start),
	}
}

// MemoryHealthChecker checks memory manager health
type MemoryHealthChecker struct {
	manager interface{} // Will be *generation.MemoryManager
}

// NewMemoryHealthChecker creates a new memory health checker
func NewMemoryHealthChecker(manager interface{}) *MemoryHealthChecker {
	return &MemoryHealthChecker{manager: manager}
}

// HealthCheck performs health check for memory manager
func (m *MemoryHealthChecker) HealthCheck(ctx context.Context) ComponentStatus {
	start := time.Now()

	if m.manager == nil {
		return ComponentStatus{
			Name:      "memory_manager",
			Status:    StatusUnhealthy,
			Message:   "Memory manager is not initialized",
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// TODO: Add actual health check logic:
	// - Check memory usage vs limits
	// - Verify session cleanup is working
	// - Test session creation/retrieval

	return ComponentStatus{
		Name:      "memory_manager",
		Status:    StatusHealthy,
		Message:   "Memory manager is operational",
		LastCheck: time.Now(),
		Duration:  time.Since(start),
	}
}

// SamplingHealthChecker checks sampling client health
type SamplingHealthChecker struct {
	client interface{} // Will be *generation.SamplingClient
}

// NewSamplingHealthChecker creates a new sampling health checker
func NewSamplingHealthChecker(client interface{}) *SamplingHealthChecker {
	return &SamplingHealthChecker{client: client}
}

// HealthCheck performs health check for sampling client
func (s *SamplingHealthChecker) HealthCheck(ctx context.Context) ComponentStatus {
	start := time.Now()

	if s.client == nil {
		return ComponentStatus{
			Name:      "sampling_client",
			Status:    StatusUnhealthy,
			Message:   "Sampling client is not initialized",
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// TODO: Add actual health check logic:
	// - Test sampling request capability
	// - Check connection to MCP sampling endpoint
	// - Verify response parsing

	return ComponentStatus{
		Name:      "sampling_client",
		Status:    StatusHealthy,
		Message:   "Sampling client is operational",
		LastCheck: time.Now(),
		Duration:  time.Since(start),
	}
}
