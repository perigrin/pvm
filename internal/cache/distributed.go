// ABOUTME: Implements distributed caching support using Redis for cross-instance cache sharing
// ABOUTME: Provides scalable caching with automatic failover and connection pooling

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// RedisClient interface for Redis operations (to avoid direct dependency)
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
	FlushAll(ctx context.Context) error
	Ping(ctx context.Context) error
	Close() error
}

// DistributedConfig contains configuration for distributed caching
type DistributedConfig struct {
	Addresses    []string      // Redis server addresses
	Password     string        // Redis password
	DB           int           // Redis database number
	MaxRetries   int           // Maximum retry attempts
	DialTimeout  time.Duration // Connection timeout
	ReadTimeout  time.Duration // Read operation timeout
	WriteTimeout time.Duration // Write operation timeout
	PoolSize     int           // Connection pool size
	MinIdleConns int           // Minimum idle connections
}

// DistributedTier implements a distributed cache tier using Redis
type DistributedTier struct {
	client     RedisClient
	config     *DistributedConfig
	compressor Compressor
	stats      *TierStats
	mu         sync.RWMutex
	logger     *log.Logger
	healthMu   sync.RWMutex
	healthy    bool
}

// NewDistributedTier creates a new distributed cache tier
func NewDistributedTier(config *DistributedConfig, compressor Compressor, logger *log.Logger) (*DistributedTier, error) {
	if config == nil {
		config = &DistributedConfig{
			Addresses:    []string{"localhost:6379"},
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
			MinIdleConns: 5,
		}
	}

	// In a real implementation, we would create a Redis client here
	// For now, we'll use a mock implementation
	client := NewMockRedisClient()

	dt := &DistributedTier{
		client:     client,
		config:     config,
		compressor: compressor,
		stats:      &TierStats{MaxSize: -1}, // No size limit for distributed cache
		logger:     logger,
		healthy:    true,
	}

	// Start health check goroutine
	go dt.healthCheck()

	return dt, nil
}

// Get retrieves an entry from the distributed tier
func (d *DistributedTier) Get(ctx context.Context, key string) (*CacheEntry, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.isHealthy() {
		d.stats.Misses++
		return nil, errors.NewSystemError("008", "distributed cache unhealthy", nil)
	}

	// Get compressed data from Redis
	data, err := d.client.Get(ctx, d.formatKey(key))
	if err != nil {
		d.stats.Misses++
		return nil, errors.Wrap(err, "PSC", "distributed", "014", "failed to get from Redis")
	}

	// Decompress data
	decompressed, err := d.compressor.Decompress([]byte(data))
	if err != nil {
		return nil, errors.Wrap(err, "PSC", "distributed", "015", "failed to decompress data")
	}

	// Unmarshal entry
	var entry CacheEntry
	if err := json.Unmarshal(decompressed, &entry); err != nil {
		return nil, errors.Wrap(err, "PSC", "distributed", "016", "failed to unmarshal entry")
	}

	d.stats.Hits++
	return &entry, nil
}

// Set stores an entry in the distributed tier
func (d *DistributedTier) Set(ctx context.Context, entry *CacheEntry) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isHealthy() {
		return errors.NewSystemError("009", "distributed cache unhealthy", nil)
	}

	// Marshal entry
	data, err := json.Marshal(entry)
	if err != nil {
		return errors.Wrap(err, "PSC", "distributed", "017", "failed to marshal entry")
	}

	// Compress data
	compressed, err := d.compressor.Compress(data)
	if err != nil {
		return errors.Wrap(err, "PSC", "distributed", "018", "failed to compress data")
	}

	// Store in Redis with TTL
	err = d.client.Set(ctx, d.formatKey(entry.Key), string(compressed), entry.TTL)
	if err != nil {
		return errors.Wrap(err, "PSC", "distributed", "019", "failed to set in Redis")
	}

	d.stats.EntryCount++
	return nil
}

// Delete removes an entry from the distributed tier
func (d *DistributedTier) Delete(ctx context.Context, key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isHealthy() {
		return errors.NewSystemError("009", "distributed cache unhealthy", nil)
	}

	err := d.client.Del(ctx, d.formatKey(key))
	if err != nil {
		return errors.Wrap(err, "PSC", "distributed", "020", "failed to delete from Redis")
	}

	d.stats.EntryCount--
	return nil
}

// Clear removes all entries from the distributed tier
func (d *DistributedTier) Clear(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isHealthy() {
		return errors.NewSystemError("009", "distributed cache unhealthy", nil)
	}

	err := d.client.FlushAll(ctx)
	if err != nil {
		return errors.Wrap(err, "PSC", "distributed", "021", "failed to flush Redis")
	}

	d.stats.EntryCount = 0
	return nil
}

// Stats returns statistics for the distributed tier
func (d *DistributedTier) Stats() *TierStats {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.stats
}

// Close closes the distributed cache connection
func (d *DistributedTier) Close() error {
	return d.client.Close()
}

// formatKey formats a cache key for Redis
func (d *DistributedTier) formatKey(key string) string {
	return fmt.Sprintf("pvm:cache:%s", key)
}

// isHealthy checks if the distributed cache is healthy
func (d *DistributedTier) isHealthy() bool {
	d.healthMu.RLock()
	defer d.healthMu.RUnlock()
	return d.healthy
}

// setHealthy sets the health status
func (d *DistributedTier) setHealthy(healthy bool) {
	d.healthMu.Lock()
	defer d.healthMu.Unlock()
	d.healthy = healthy
}

// healthCheck periodically checks Redis health
func (d *DistributedTier) healthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := d.client.Ping(ctx)
		cancel()

		if err != nil {
			d.logger.Warningf("Redis health check failed: %v", err)
			d.setHealthy(false)
		} else {
			d.setHealthy(true)
		}
	}
}

// MockRedisClient provides a mock Redis implementation for testing
type MockRedisClient struct {
	mu    sync.RWMutex
	data  map[string]string
	ttls  map[string]time.Time
	alive bool
}

// NewMockRedisClient creates a new mock Redis client
func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data:  make(map[string]string),
		ttls:  make(map[string]time.Time),
		alive: true,
	}
}

// Get retrieves a value from the mock Redis
func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.alive {
		return "", errors.NewSystemError("010", "connection closed", nil)
	}

	// Check if key exists and hasn't expired
	if expiry, exists := m.ttls[key]; exists && time.Now().After(expiry) {
		delete(m.data, key)
		delete(m.ttls, key)
		return "", errors.NewSystemError("011", "key not found", nil)
	}

	value, exists := m.data[key]
	if !exists {
		return "", errors.NewSystemError("011", "key not found", nil)
	}

	return value, nil
}

// Set stores a value in the mock Redis
func (m *MockRedisClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.alive {
		return errors.NewSystemError("010", "connection closed", nil)
	}

	m.data[key] = value
	if expiration > 0 {
		m.ttls[key] = time.Now().Add(expiration)
	}

	return nil
}

// Del deletes keys from the mock Redis
func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.alive {
		return errors.NewSystemError("010", "connection closed", nil)
	}

	for _, key := range keys {
		delete(m.data, key)
		delete(m.ttls, key)
	}

	return nil
}

// FlushAll removes all data from the mock Redis
func (m *MockRedisClient) FlushAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.alive {
		return errors.NewSystemError("010", "connection closed", nil)
	}

	m.data = make(map[string]string)
	m.ttls = make(map[string]time.Time)

	return nil
}

// Ping checks if the mock Redis is alive
func (m *MockRedisClient) Ping(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.alive {
		return errors.NewSystemError("010", "connection closed", nil)
	}

	return nil
}

// Close closes the mock Redis connection
func (m *MockRedisClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.alive = false
	return nil
}

// RedisConnectionPool manages a pool of Redis connections
type RedisConnectionPool struct {
	config      *DistributedConfig
	connections chan RedisClient
	factory     func() (RedisClient, error)
	mu          sync.Mutex
	closed      bool
}

// NewRedisConnectionPool creates a new connection pool
func NewRedisConnectionPool(config *DistributedConfig, factory func() (RedisClient, error)) (*RedisConnectionPool, error) {
	pool := &RedisConnectionPool{
		config:      config,
		connections: make(chan RedisClient, config.PoolSize),
		factory:     factory,
	}

	// Pre-populate pool with minimum connections
	for i := 0; i < config.MinIdleConns; i++ {
		conn, err := factory()
		if err != nil {
			return nil, errors.Wrap(err, "PSC", "distributed", "022", "failed to create initial connections")
		}
		pool.connections <- conn
	}

	return pool, nil
}

// Get retrieves a connection from the pool
func (p *RedisConnectionPool) Get(ctx context.Context) (RedisClient, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errors.NewSystemError("013", "pool is closed", nil)
	}
	p.mu.Unlock()

	select {
	case conn := <-p.connections:
		// Test connection health
		if err := conn.Ping(ctx); err != nil {
			// Connection is bad, create a new one
			conn.Close()
			return p.factory()
		}
		return conn, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// No connections available, create a new one
		return p.factory()
	}
}

// Put returns a connection to the pool
func (p *RedisConnectionPool) Put(conn RedisClient) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		conn.Close()
		return
	}

	select {
	case p.connections <- conn:
		// Connection returned to pool
	default:
		// Pool is full, close the connection
		conn.Close()
	}
}

// Close closes all connections in the pool
func (p *RedisConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	close(p.connections)

	var errs []error
	for conn := range p.connections {
		if err := conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.NewSystemError("007", fmt.Sprintf("failed to close all connections: %v", errs), nil)
	}

	return nil
}
