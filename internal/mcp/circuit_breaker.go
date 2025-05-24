// ABOUTME: Circuit breaker implementation for external dependencies
// ABOUTME: Provides resilience patterns for API calls and external services

package mcp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig configures a circuit breaker
type CircuitBreakerConfig struct {
	MaxFailures      int           `json:"max_failures"`      // Number of failures before opening
	ResetTimeout     time.Duration `json:"reset_timeout"`     // Time to wait before half-open
	SuccessThreshold int           `json:"success_threshold"` // Successes needed in half-open to close
	Timeout          time.Duration `json:"timeout"`           // Individual request timeout
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:      5,
		ResetTimeout:     60 * time.Second,
		SuccessThreshold: 3,
		Timeout:          30 * time.Second,
	}
}

// CircuitBreakerMetrics tracks circuit breaker metrics
type CircuitBreakerMetrics struct {
	TotalRequests   int64     `json:"total_requests"`
	SuccessCount    int64     `json:"success_count"`
	FailureCount    int64     `json:"failure_count"`
	CircuitOpens    int64     `json:"circuit_opens"`
	CircuitCloses   int64     `json:"circuit_closes"`
	LastFailureTime time.Time `json:"last_failure_time"`
	LastSuccessTime time.Time `json:"last_success_time"`
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	name            string
	config          CircuitBreakerConfig
	state           CircuitBreakerState
	failures        int
	successes       int
	lastFailureTime time.Time
	lastSuccessTime time.Time
	metrics         CircuitBreakerMetrics
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		name:   name,
		config: config,
		state:  StateClosed,
	}
}

// ErrCircuitOpen is returned when the circuit breaker is open
var ErrCircuitOpen = errors.New("circuit breaker is open")

// Execute runs a function through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	// Check if we can execute
	if err := cb.allowRequest(); err != nil {
		return err
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, cb.config.Timeout)
	defer cancel()

	// Execute the function
	cb.mu.Lock()
	cb.metrics.TotalRequests++
	cb.mu.Unlock()

	err := fn(timeoutCtx)

	// Record the result
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// allowRequest checks if a request should be allowed
func (cb *CircuitBreaker) allowRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailureTime) > cb.config.ResetTimeout {
			cb.state = StateHalfOpen
			cb.successes = 0
			return nil
		}
		return ErrCircuitOpen
	case StateHalfOpen:
		return nil
	default:
		return ErrCircuitOpen
	}
}

// recordSuccess records a successful execution
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.metrics.SuccessCount++
	cb.lastSuccessTime = time.Now()
	cb.metrics.LastSuccessTime = cb.lastSuccessTime

	switch cb.state {
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
			cb.metrics.CircuitCloses++
		}
	case StateClosed:
		// Reset failure count on success
		cb.failures = 0
	}
}

// recordFailure records a failed execution
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.metrics.FailureCount++
	cb.failures++
	cb.lastFailureTime = time.Now()
	cb.metrics.LastFailureTime = cb.lastFailureTime

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.state = StateOpen
			cb.metrics.CircuitOpens++
		}
	case StateHalfOpen:
		cb.state = StateOpen
		cb.successes = 0
		cb.metrics.CircuitOpens++
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns current metrics
func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.metrics
}

// GetName returns the circuit breaker name
func (cb *CircuitBreaker) GetName() string {
	return cb.name
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetBreaker gets or creates a circuit breaker
func (m *CircuitBreakerManager) GetBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	m.mu.Lock()
	defer m.mu.Unlock()

	if breaker, exists := m.breakers[name]; exists {
		return breaker
	}

	breaker := NewCircuitBreaker(name, config)
	m.breakers[name] = breaker
	return breaker
}

// GetAllBreakers returns all circuit breakers
func (m *CircuitBreakerManager) GetAllBreakers() map[string]*CircuitBreaker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*CircuitBreaker)
	for name, breaker := range m.breakers {
		result[name] = breaker
	}
	return result
}

// GetBreakersStatus returns status of all circuit breakers
func (m *CircuitBreakerManager) GetBreakersStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]interface{})
	for name, breaker := range m.breakers {
		status[name] = map[string]interface{}{
			"state":   breaker.GetState().String(),
			"metrics": breaker.GetMetrics(),
		}
	}
	return status
}

// BreakersHealthChecker implements health checking for circuit breakers
type BreakersHealthChecker struct {
	manager *CircuitBreakerManager
}

// NewBreakersHealthChecker creates a health checker for circuit breakers
func NewBreakersHealthChecker(manager *CircuitBreakerManager) *BreakersHealthChecker {
	return &BreakersHealthChecker{manager: manager}
}

// HealthCheck checks the health of all circuit breakers
func (b *BreakersHealthChecker) HealthCheck(ctx context.Context) ComponentStatus {
	start := time.Now()

	breakers := b.manager.GetAllBreakers()
	openCount := 0
	totalCount := len(breakers)

	var openBreakers []string
	for name, breaker := range breakers {
		if breaker.GetState() == StateOpen {
			openCount++
			openBreakers = append(openBreakers, name)
		}
	}

	var status HealthStatus
	var message string

	switch {
	case openCount == 0:
		status = StatusHealthy
		message = fmt.Sprintf("All %d circuit breakers are operational", totalCount)
	case openCount == totalCount && totalCount > 0:
		status = StatusUnhealthy
		message = fmt.Sprintf("All %d circuit breakers are open: %v", totalCount, openBreakers)
	default:
		status = StatusDegraded
		message = fmt.Sprintf("%d of %d circuit breakers are open: %v", openCount, totalCount, openBreakers)
	}

	return ComponentStatus{
		Name:      "circuit_breakers",
		Status:    status,
		Message:   message,
		LastCheck: time.Now(),
		Duration:  time.Since(start),
	}
}
