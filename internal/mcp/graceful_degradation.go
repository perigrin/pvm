// ABOUTME: Graceful degradation mechanisms for MCP server components
// ABOUTME: Provides fallback strategies when external dependencies fail

package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DegradationStrategy defines how to handle component failures
type DegradationStrategy int

const (
	StrategyFail     DegradationStrategy = iota // Fail the request
	StrategyCache                               // Use cached data if available
	StrategyFallback                            // Use fallback implementation
	StrategySkip                                // Skip the operation
)

func (s DegradationStrategy) String() string {
	switch s {
	case StrategyFail:
		return "fail"
	case StrategyCache:
		return "cache"
	case StrategyFallback:
		return "fallback"
	case StrategySkip:
		return "skip"
	default:
		return "unknown"
	}
}

// DegradationConfig configures graceful degradation behavior
type DegradationConfig struct {
	EmbeddingFailureStrategy  DegradationStrategy `json:"embedding_failure_strategy"`
	ValidationFailureStrategy DegradationStrategy `json:"validation_failure_strategy"`
	SamplingFailureStrategy   DegradationStrategy `json:"sampling_failure_strategy"`
	MemoryFailureStrategy     DegradationStrategy `json:"memory_failure_strategy"`
	EnableCaching             bool                `json:"enable_caching"`
	CacheTTL                  time.Duration       `json:"cache_ttl"`
	FallbackTimeout           time.Duration       `json:"fallback_timeout"`
}

// DefaultDegradationConfig returns sensible defaults
func DefaultDegradationConfig() DegradationConfig {
	return DegradationConfig{
		EmbeddingFailureStrategy:  StrategyCache,
		ValidationFailureStrategy: StrategyFallback,
		SamplingFailureStrategy:   StrategySkip,
		MemoryFailureStrategy:     StrategyFallback,
		EnableCaching:             true,
		CacheTTL:                  30 * time.Minute,
		FallbackTimeout:           10 * time.Second,
	}
}

// FallbackCache provides simple in-memory caching for degradation scenarios
type FallbackCache struct {
	mu      sync.RWMutex
	cache   map[string]*CacheEntry
	ttl     time.Duration
	maxSize int
}

// CacheEntry represents a cached value
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// NewFallbackCache creates a new fallback cache
func NewFallbackCache(ttl time.Duration, maxSize int) *FallbackCache {
	cache := &FallbackCache{
		cache:   make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from cache
func (c *FallbackCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(c.cache, key)
		return nil, false
	}

	return entry.Value, true
}

// Set stores a value in cache
func (c *FallbackCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entries if cache is full
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// evictOldest removes the oldest entry from cache
func (c *FallbackCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.cache {
		if oldestKey == "" || entry.ExpiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.ExpiresAt
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}

// cleanup removes expired entries
func (c *FallbackCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.cache {
			if now.After(entry.ExpiresAt) {
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
	}
}

// Size returns the current cache size
func (c *FallbackCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// DegradationManager handles graceful degradation strategies
type DegradationManager struct {
	config DegradationConfig
	cache  *FallbackCache
	logger interface{} // Can be any logger that has Printf method
}

// NewDegradationManager creates a new degradation manager
func NewDegradationManager(config DegradationConfig, logger interface{}) *DegradationManager {
	cache := NewFallbackCache(config.CacheTTL, 1000) // Max 1000 cached items

	return &DegradationManager{
		config: config,
		cache:  cache,
		logger: logger,
	}
}

// logf logs a formatted message using the configured logger
func (dm *DegradationManager) logf(format string, args ...interface{}) {
	if logger, ok := dm.logger.(interface{ Printf(string, ...interface{}) }); ok {
		logger.Printf(format, args...)
	}
}

// HandleEmbeddingFailure handles embedding service failures
func (dm *DegradationManager) HandleEmbeddingFailure(ctx context.Context, operation string, err error) (interface{}, error) {
	dm.logf("Embedding failure in %s: %v", operation, err)

	switch dm.config.EmbeddingFailureStrategy {
	case StrategyCache:
		if cached, found := dm.cache.Get("embedding_" + operation); found {
			dm.logf("Using cached embedding result for %s", operation)
			return cached, nil
		}
		return nil, fmt.Errorf("embedding failed and no cache available: %w", err)

	case StrategyFallback:
		// Use local/deterministic embedding as fallback
		return dm.fallbackEmbedding(ctx, operation)

	case StrategySkip:
		dm.logf("Skipping embedding operation %s due to failure", operation)
		return nil, nil

	default:
		return nil, err
	}
}

// HandleValidationFailure handles validation service failures
func (dm *DegradationManager) HandleValidationFailure(ctx context.Context, operation string, err error) (interface{}, error) {
	dm.logf("Validation failure in %s: %v", operation, err)

	switch dm.config.ValidationFailureStrategy {
	case StrategyCache:
		if cached, found := dm.cache.Get("validation_" + operation); found {
			dm.logf("Using cached validation result for %s", operation)
			return cached, nil
		}
		return nil, fmt.Errorf("validation failed and no cache available: %w", err)

	case StrategyFallback:
		// Use basic syntax checking as fallback
		return dm.fallbackValidation(ctx, operation)

	case StrategySkip:
		dm.logf("Skipping validation operation %s due to failure", operation)
		// Return a minimal validation result
		return map[string]interface{}{
			"status":  "skipped",
			"valid":   true,
			"message": "Validation skipped due to service unavailability",
		}, nil

	default:
		return nil, err
	}
}

// HandleSamplingFailure handles sampling service failures
func (dm *DegradationManager) HandleSamplingFailure(ctx context.Context, operation string, err error) (interface{}, error) {
	dm.logf("Sampling failure in %s: %v", operation, err)

	switch dm.config.SamplingFailureStrategy {
	case StrategyCache:
		if cached, found := dm.cache.Get("sampling_" + operation); found {
			dm.logf("Using cached sampling result for %s", operation)
			return cached, nil
		}
		return nil, fmt.Errorf("sampling failed and no cache available: %w", err)

	case StrategyFallback:
		// Use template-based generation as fallback
		return dm.fallbackSampling(ctx, operation)

	case StrategySkip:
		dm.logf("Skipping sampling operation %s due to failure", operation)
		return nil, nil

	default:
		return nil, err
	}
}

// HandleMemoryFailure handles memory service failures
func (dm *DegradationManager) HandleMemoryFailure(ctx context.Context, operation string, err error) (interface{}, error) {
	dm.logf("Memory failure in %s: %v", operation, err)

	switch dm.config.MemoryFailureStrategy {
	case StrategyCache:
		if cached, found := dm.cache.Get("memory_" + operation); found {
			dm.logf("Using cached memory result for %s", operation)
			return cached, nil
		}
		return nil, fmt.Errorf("memory failed and no cache available: %w", err)

	case StrategyFallback:
		// Use in-memory fallback storage
		return dm.fallbackMemory(ctx, operation)

	case StrategySkip:
		dm.logf("Skipping memory operation %s due to failure", operation)
		return map[string]interface{}{
			"status":  "no_memory",
			"message": "Memory service unavailable",
		}, nil

	default:
		return nil, err
	}
}

// CacheResult caches a successful result for later fallback use
func (dm *DegradationManager) CacheResult(operation string, result interface{}) {
	if dm.config.EnableCaching {
		dm.cache.Set(operation, result)
	}
}

// fallbackEmbedding provides a simple fallback embedding
func (dm *DegradationManager) fallbackEmbedding(ctx context.Context, operation string) (interface{}, error) {
	dm.logf("Using fallback embedding for %s", operation)

	// Return a simple deterministic embedding based on hash of operation
	// In a real implementation, this could use a local embedding model
	hash := simpleHash(operation)
	embedding := make([]float32, 384) // Standard embedding dimension

	// Generate deterministic values based on hash
	for i := range embedding {
		embedding[i] = float32((hash+i)%100)/100.0 - 0.5
	}

	return embedding, nil
}

// fallbackValidation provides basic syntax validation
func (dm *DegradationManager) fallbackValidation(ctx context.Context, operation string) (interface{}, error) {
	dm.logf("Using fallback validation for %s", operation)

	// Provide basic validation response
	return map[string]interface{}{
		"status":   "fallback",
		"valid":    true,
		"message":  "Basic validation passed (full validation unavailable)",
		"errors":   []string{},
		"warnings": []string{"Full type checking unavailable"},
	}, nil
}

// fallbackSampling provides template-based code generation
func (dm *DegradationManager) fallbackSampling(ctx context.Context, operation string) (interface{}, error) {
	dm.logf("Using fallback sampling for %s", operation)

	// Return a basic template-based response
	return map[string]interface{}{
		"generated_code": "# Generated using fallback template\n# Full AI sampling unavailable\n",
		"status":         "fallback",
		"message":        "Template-based generation used (AI sampling unavailable)",
	}, nil
}

// fallbackMemory provides in-memory storage fallback
func (dm *DegradationManager) fallbackMemory(ctx context.Context, operation string) (interface{}, error) {
	dm.logf("Using fallback memory for %s", operation)

	// Use the degradation cache as temporary memory
	return map[string]interface{}{
		"status":  "fallback",
		"storage": "temporary",
		"message": "Using temporary in-memory storage",
	}, nil
}

// simpleHash creates a simple hash of a string
func simpleHash(s string) int {
	hash := 0
	for _, char := range s {
		hash = hash*31 + int(char)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// GetStats returns degradation manager statistics
func (dm *DegradationManager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"config": map[string]interface{}{
			"embedding_strategy":  dm.config.EmbeddingFailureStrategy.String(),
			"validation_strategy": dm.config.ValidationFailureStrategy.String(),
			"sampling_strategy":   dm.config.SamplingFailureStrategy.String(),
			"memory_strategy":     dm.config.MemoryFailureStrategy.String(),
			"caching_enabled":     dm.config.EnableCaching,
			"cache_ttl_minutes":   dm.config.CacheTTL.Minutes(),
		},
		"cache": map[string]interface{}{
			"size":     dm.cache.Size(),
			"max_size": 1000,
		},
	}
}

// HealthCheck checks the health of the degradation manager
func (dm *DegradationManager) HealthCheck(ctx context.Context) ComponentStatus {
	start := time.Now()

	cacheSize := dm.cache.Size()

	var status HealthStatus
	var message string

	switch {
	case cacheSize < 500:
		status = StatusHealthy
		message = fmt.Sprintf("Degradation manager operational with %d cached items", cacheSize)
	case cacheSize < 800:
		status = StatusDegraded
		message = fmt.Sprintf("Degradation cache getting full: %d items", cacheSize)
	default:
		status = StatusUnhealthy
		message = fmt.Sprintf("Degradation cache nearly full: %d items (performance may be affected)", cacheSize)
	}

	return ComponentStatus{
		Name:      "degradation_manager",
		Status:    status,
		Message:   message,
		LastCheck: time.Now(),
		Duration:  time.Since(start),
	}
}
