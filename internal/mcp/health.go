// ABOUTME: Health check and monitoring system for MCP server
// ABOUTME: Provides health status, component checks, and operational metrics

package mcp

import (
	"context"
	"fmt"
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

	// TODO: See issue #353 - Implement comprehensive MCP component health monitoring

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

	// TODO: See issue #353 - Implement comprehensive MCP component health monitoring

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

	// TODO: See issue #353 - Implement comprehensive MCP component health monitoring

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

// SamplingClientHealthInterface defines the health check interface for sampling clients
type SamplingClientHealthInterface interface {
	HealthCheck(ctx context.Context) error
	IsEnabled() bool
	GetMode() interface{} // Returns SamplingMode, but using interface{} to avoid import cycles
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

	// Try to cast to the health check interface
	healthClient, ok := s.client.(SamplingClientHealthInterface)
	if !ok {
		return ComponentStatus{
			Name:      "sampling_client",
			Status:    StatusDegraded,
			Message:   "Sampling client does not support health checks",
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// Check if sampling is enabled
	if !healthClient.IsEnabled() {
		return ComponentStatus{
			Name:      "sampling_client",
			Status:    StatusDegraded,
			Message:   "Sampling is disabled",
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// Perform actual health check
	healthCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := healthClient.HealthCheck(healthCtx)
	if err != nil {
		status := StatusUnhealthy
		message := fmt.Sprintf("Sampling health check failed: %v", err)

		// Determine if this is a degraded or unhealthy state
		errorStr := err.Error()
		if contains(errorStr, "circuit breaker") || contains(errorStr, "timeout") {
			status = StatusDegraded
			message = fmt.Sprintf("Sampling degraded: %v", err)
		}

		return ComponentStatus{
			Name:      "sampling_client",
			Status:    status,
			Message:   message,
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// Determine message based on mode
	mode := healthClient.GetMode()
	message := "Sampling client is operational"
	if mode != nil {
		if modeStr, ok := mode.(string); ok && modeStr == "mock" {
			message = "Sampling client is operational (mock mode)"
		} else {
			message = "Sampling client is operational (real MCP mode)"
		}
	}

	return ComponentStatus{
		Name:      "sampling_client",
		Status:    StatusHealthy,
		Message:   message,
		LastCheck: time.Now(),
		Duration:  time.Since(start),
	}
}

// MCPClientHealthChecker checks real MCP client health
type MCPClientHealthChecker struct {
	client interface{} // Will be *client.MCPClient
}

// MCPClientHealthInterface defines the health check interface for MCP clients
type MCPClientHealthInterface interface {
	HealthCheck(ctx context.Context) error
	IsConnected() bool
	GetCapabilities() interface{} // Returns *MCPCapabilities, but using interface{} to avoid import cycles
}

// NewMCPClientHealthChecker creates a new MCP client health checker
func NewMCPClientHealthChecker(client interface{}) *MCPClientHealthChecker {
	return &MCPClientHealthChecker{client: client}
}

// HealthCheck performs health check for MCP client
func (m *MCPClientHealthChecker) HealthCheck(ctx context.Context) ComponentStatus {
	start := time.Now()

	if m.client == nil {
		return ComponentStatus{
			Name:      "mcp_client",
			Status:    StatusUnhealthy,
			Message:   "MCP client is not initialized",
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// Try to cast to the health check interface
	healthClient, ok := m.client.(MCPClientHealthInterface)
	if !ok {
		return ComponentStatus{
			Name:      "mcp_client",
			Status:    StatusDegraded,
			Message:   "MCP client does not support health checks",
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// Check connection status
	if !healthClient.IsConnected() {
		return ComponentStatus{
			Name:      "mcp_client",
			Status:    StatusUnhealthy,
			Message:   "MCP client is not connected",
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// Perform actual health check
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := healthClient.HealthCheck(healthCtx)
	if err != nil {
		status := StatusUnhealthy
		message := fmt.Sprintf("MCP client health check failed: %v", err)

		// Determine if this is a degraded or unhealthy state
		errorStr := err.Error()
		if contains(errorStr, "circuit breaker") || contains(errorStr, "timeout") {
			status = StatusDegraded
			message = fmt.Sprintf("MCP client degraded: %v", err)
		}

		return ComponentStatus{
			Name:      "mcp_client",
			Status:    status,
			Message:   message,
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// Get capabilities information
	capabilities := healthClient.GetCapabilities()
	message := "MCP client is connected and operational"
	if capabilities != nil {
		message = "MCP client is connected with full capabilities"
	}

	return ComponentStatus{
		Name:      "mcp_client",
		Status:    StatusHealthy,
		Message:   message,
		LastCheck: time.Now(),
		Duration:  time.Since(start),
	}
}

// ResourceManagerHealthChecker checks resource manager health
type ResourceManagerHealthChecker struct {
	manager interface{} // Will be *ResourceManager
}

// ResourceManagerHealthInterface defines the health check interface for resource managers
type ResourceManagerHealthInterface interface {
	IsEnabled() bool
	RefreshResources(ctx context.Context) error
}

// NewResourceManagerHealthChecker creates a new resource manager health checker
func NewResourceManagerHealthChecker(manager interface{}) *ResourceManagerHealthChecker {
	return &ResourceManagerHealthChecker{manager: manager}
}

// HealthCheck performs health check for resource manager
func (r *ResourceManagerHealthChecker) HealthCheck(ctx context.Context) ComponentStatus {
	start := time.Now()

	if r.manager == nil {
		return ComponentStatus{
			Name:      "resource_manager",
			Status:    StatusUnhealthy,
			Message:   "Resource manager is not initialized",
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// Try to cast to the health check interface
	healthManager, ok := r.manager.(ResourceManagerHealthInterface)
	if !ok {
		return ComponentStatus{
			Name:      "resource_manager",
			Status:    StatusDegraded,
			Message:   "Resource manager does not support health checks",
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// Check if resource management is enabled
	if !healthManager.IsEnabled() {
		return ComponentStatus{
			Name:      "resource_manager",
			Status:    StatusDegraded,
			Message:   "Resource management is disabled",
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// Perform actual health check by trying to refresh resources
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := healthManager.RefreshResources(healthCtx)
	if err != nil {
		status := StatusUnhealthy
		message := fmt.Sprintf("Resource manager health check failed: %v", err)

		// Determine if this is a degraded or unhealthy state
		errorStr := err.Error()
		if contains(errorStr, "permission denied") || contains(errorStr, "timeout") {
			status = StatusDegraded
			message = fmt.Sprintf("Resource manager degraded: %v", err)
		}

		return ComponentStatus{
			Name:      "resource_manager",
			Status:    status,
			Message:   message,
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}
	}

	return ComponentStatus{
		Name:      "resource_manager",
		Status:    StatusHealthy,
		Message:   "Resource manager is operational",
		LastCheck: time.Now(),
		Duration:  time.Since(start),
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
