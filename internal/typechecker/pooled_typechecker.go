// ABOUTME: Pooled type checker factory providing memory-efficient type checking operations
// ABOUTME: Integrates type pool manager with existing type checker for optimal performance

package typechecker

import (
	"sync"
	"time"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/typedef"
)

// PooledTypeChecker extends TypeChecker with pooling capabilities
type PooledTypeChecker struct {
	*TypeChecker

	// Pool manager for allocating type objects
	poolManager *TypePoolManager

	// Resolution pool for caching type resolution results
	resolutionPool *TypeResolutionPool

	// Active sessions for tracking type checking sessions
	activeSessions map[string]*TypeCheckSession

	// Session lock for concurrent access
	sessionMu sync.RWMutex
}

// TypeCheckSession represents an active type checking session
type TypeCheckSession struct {
	// SessionID uniquely identifies the session
	SessionID string

	// SourceFile being checked
	SourceFile string

	// StartTime when the session began
	StartTime int64

	// PooledObjects tracks objects allocated from pools for this session
	PooledObjects *SessionObjectTracker

	// Statistics for this session
	Statistics TypeCheckSessionStats
}

// SessionObjectTracker tracks pooled objects for cleanup
type SessionObjectTracker struct {
	// TypeCheckResults allocated in this session
	TypeCheckResults []*TypeCheckResult

	// TypeStates allocated in this session
	TypeStates []*TypeState

	// InferenceEngines allocated in this session
	InferenceEngines []*InferenceEngine

	// InferredTypeInfos allocated in this session
	InferredTypeInfos []*InferredTypeInfo

	// ResolutionResults allocated in this session
	ResolutionResults []*TypeResolutionResult
}

// TypeCheckSessionStats contains statistics for a type checking session
type TypeCheckSessionStats struct {
	// ObjectsAllocated is the number of objects allocated from pools
	ObjectsAllocated int64

	// CacheHits is the number of cache hits during the session
	CacheHits int64

	// CacheMisses is the number of cache misses during the session
	CacheMisses int64

	// TypeResolutions is the number of type resolutions performed
	TypeResolutions int64

	// InferenceOperations is the number of inference operations performed
	InferenceOperations int64

	// MemoryReused is the estimated amount of memory reused from pools
	MemoryReused int64
}

// PooledTypeCheckerFactory creates pooled type checkers
type PooledTypeCheckerFactory struct {
	// Global pool manager
	poolManager *TypePoolManager

	// Global resolution pool
	resolutionPool *TypeResolutionPool

	// Factory configuration
	config PooledTypeCheckerConfig

	// Statistics
	checkersCreated int64
	sessionsActive  int64

	mu sync.RWMutex
}

// PooledTypeCheckerConfig contains configuration for pooled type checkers
type PooledTypeCheckerConfig struct {
	// EnablePooling controls whether object pooling is enabled
	EnablePooling bool

	// EnableCaching controls whether type resolution caching is enabled
	EnableCaching bool

	// SessionTracking controls whether session tracking is enabled
	SessionTracking bool

	// AutoCleanup controls whether automatic cleanup is enabled
	AutoCleanup bool

	// CleanupThreshold is the memory threshold for triggering cleanup
	CleanupThreshold int64

	// MaxActiveSessions is the maximum number of active sessions
	MaxActiveSessions int

	// StatisticsCollection controls whether detailed statistics are collected
	StatisticsCollection bool
}

// DefaultPooledTypeCheckerConfig returns default configuration
func DefaultPooledTypeCheckerConfig() PooledTypeCheckerConfig {
	return PooledTypeCheckerConfig{
		EnablePooling:        true,
		EnableCaching:        true,
		SessionTracking:      true,
		AutoCleanup:          true,
		CleanupThreshold:     100 * 1024 * 1024, // 100MB
		MaxActiveSessions:    100,
		StatisticsCollection: true,
	}
}

// NewPooledTypeCheckerFactory creates a new pooled type checker factory
func NewPooledTypeCheckerFactory(config PooledTypeCheckerConfig) *PooledTypeCheckerFactory {
	factory := &PooledTypeCheckerFactory{
		config: config,
	}

	if config.EnablePooling {
		factory.poolManager = GlobalTypePoolManager()
	}

	if config.EnableCaching {
		factory.resolutionPool = GlobalTypeResolutionPool()
	}

	return factory
}

// CreateTypeChecker creates a new pooled type checker
func (factory *PooledTypeCheckerFactory) CreateTypeChecker(hierarchy *typedef.TypeHierarchy, symbolTable *binder.SymbolTable, moduleName string) *PooledTypeChecker {
	factory.mu.Lock()
	defer factory.mu.Unlock()

	var tc *TypeChecker

	if factory.config.EnablePooling && factory.poolManager != nil {
		// Use pooled allocation
		tc = factory.poolManager.NewTypeChecker(hierarchy, symbolTable, moduleName)
	} else {
		// Use regular allocation
		tc = NewTypeChecker(hierarchy, symbolTable, moduleName)
	}

	pooledTC := &PooledTypeChecker{
		TypeChecker:    tc,
		poolManager:    factory.poolManager,
		resolutionPool: factory.resolutionPool,
		activeSessions: make(map[string]*TypeCheckSession),
	}

	factory.checkersCreated++

	return pooledTC
}

// GetStatistics returns factory statistics
func (factory *PooledTypeCheckerFactory) GetStatistics() PooledTypeCheckerFactoryStats {
	factory.mu.RLock()
	defer factory.mu.RUnlock()

	return PooledTypeCheckerFactoryStats{
		CheckersCreated: factory.checkersCreated,
		SessionsActive:  factory.sessionsActive,
	}
}

// PooledTypeCheckerFactoryStats contains factory statistics
type PooledTypeCheckerFactoryStats struct {
	CheckersCreated int64 // Number of type checkers created
	SessionsActive  int64 // Number of active sessions
}

// NewTypeCheckSession creates a new type checking session
func (ptc *PooledTypeChecker) NewTypeCheckSession(sessionID, sourceFile string) *TypeCheckSession {
	ptc.sessionMu.Lock()
	defer ptc.sessionMu.Unlock()

	session := &TypeCheckSession{
		SessionID:     sessionID,
		SourceFile:    sourceFile,
		StartTime:     getCurrentTimeMs(),
		PooledObjects: &SessionObjectTracker{},
		Statistics:    TypeCheckSessionStats{},
	}

	ptc.activeSessions[sessionID] = session

	return session
}

// GetTypeCheckSession retrieves an active session
func (ptc *PooledTypeChecker) GetTypeCheckSession(sessionID string) *TypeCheckSession {
	ptc.sessionMu.RLock()
	defer ptc.sessionMu.RUnlock()

	return ptc.activeSessions[sessionID]
}

// CloseTypeCheckSession closes and cleans up a type checking session
func (ptc *PooledTypeChecker) CloseTypeCheckSession(sessionID string) {
	ptc.sessionMu.Lock()
	defer ptc.sessionMu.Unlock()

	session, exists := ptc.activeSessions[sessionID]
	if !exists {
		return
	}

	// Clean up pooled objects for this session
	ptc.cleanupSessionObjects(session)

	// Remove from active sessions
	delete(ptc.activeSessions, sessionID)
}

// CheckASTWithPooling performs type checking on an AST using pooled objects
func (ptc *PooledTypeChecker) CheckASTWithPooling(ast *ast.AST, sessionID string) (*TypeCheckResult, error) {
	session := ptc.GetTypeCheckSession(sessionID)
	if session == nil {
		// Create a temporary session if none exists
		session = ptc.NewTypeCheckSession(sessionID, "")
		defer ptc.CloseTypeCheckSession(sessionID)
	}

	var result *TypeCheckResult

	if ptc.poolManager != nil {
		// Use pooled allocation
		result = ptc.poolManager.NewTypeCheckResult(session.SourceFile)
		session.PooledObjects.TypeCheckResults = append(session.PooledObjects.TypeCheckResults, result)
		session.Statistics.ObjectsAllocated++
	} else {
		// Use regular allocation
		result = &TypeCheckResult{
			Path:                 session.SourceFile,
			Errors:               []TypeCheckError{},
			TypeAnnotations:      nil,
			RefinedTypes:         make(map[string]string),
			FlowSensitiveEnabled: false,
		}
	}

	// Perform the actual type checking
	errors := ptc.TypeChecker.CheckAST(ast)

	// Convert errors to TypeCheckError format
	for _, err := range errors {
		typeErr := TypeCheckError{
			Message: err.Error(),
			Path:    session.SourceFile,
			Line:    0, // Would need to extract from error
			Column:  0, // Would need to extract from error
		}
		result.Errors = append(result.Errors, typeErr)
	}

	// Copy type annotations from AST
	result.TypeAnnotations = ast.TypeAnnotations

	// Update session statistics
	session.Statistics.TypeResolutions++

	return result, nil
}

// ResolveTypeWithCaching resolves a type using caching if available
func (ptc *PooledTypeChecker) ResolveTypeWithCaching(key string, resolver func() (*TypeResolutionResult, error), sessionID string) (*TypeResolutionResult, error) {
	session := ptc.GetTypeCheckSession(sessionID)

	if ptc.resolutionPool != nil {
		// Use cached resolution
		result, err := ptc.resolutionPool.ResolveType(key, resolver)
		if err != nil {
			return nil, err
		}

		if session != nil {
			if result.CacheInfo.CacheHit {
				session.Statistics.CacheHits++
			} else {
				session.Statistics.CacheMisses++
			}
			session.PooledObjects.ResolutionResults = append(session.PooledObjects.ResolutionResults, result)
		}

		return result, nil
	} else {
		// Use direct resolution
		return resolver()
	}
}

// CreateInferenceEngineForSession creates an inference engine for a session
func (ptc *PooledTypeChecker) CreateInferenceEngineForSession(sessionID string) *InferenceEngine {
	session := ptc.GetTypeCheckSession(sessionID)

	var engine *InferenceEngine

	if ptc.poolManager != nil {
		// Use pooled allocation
		engine = ptc.poolManager.NewInferenceEngine(ptc.TypeChecker.Hierarchy, ptc.TypeChecker.SymbolTable)
		if session != nil {
			session.PooledObjects.InferenceEngines = append(session.PooledObjects.InferenceEngines, engine)
			session.Statistics.ObjectsAllocated++
		}
	} else {
		// Use regular allocation
		engine = NewInferenceEngine(ptc.TypeChecker.Hierarchy, ptc.TypeChecker.SymbolTable)
	}

	return engine
}

// PerformInferenceWithPooling performs type inference using pooled objects
func (ptc *PooledTypeChecker) PerformInferenceWithPooling(ast *ast.AST, sessionID string) error {
	engine := ptc.CreateInferenceEngineForSession(sessionID)

	// Perform inference
	err := engine.InferTypes(ast)

	if session := ptc.GetTypeCheckSession(sessionID); session != nil {
		session.Statistics.InferenceOperations++
	}

	return err
}

// GetSessionStatistics returns statistics for a specific session
func (ptc *PooledTypeChecker) GetSessionStatistics(sessionID string) *TypeCheckSessionStats {
	ptc.sessionMu.RLock()
	defer ptc.sessionMu.RUnlock()

	session, exists := ptc.activeSessions[sessionID]
	if !exists {
		return nil
	}

	return &session.Statistics
}

// GetAllSessionStatistics returns statistics for all active sessions
func (ptc *PooledTypeChecker) GetAllSessionStatistics() map[string]TypeCheckSessionStats {
	ptc.sessionMu.RLock()
	defer ptc.sessionMu.RUnlock()

	stats := make(map[string]TypeCheckSessionStats)
	for sessionID, session := range ptc.activeSessions {
		stats[sessionID] = session.Statistics
	}

	return stats
}

// CleanupAllSessions closes and cleans up all active sessions
func (ptc *PooledTypeChecker) CleanupAllSessions() {
	ptc.sessionMu.Lock()
	defer ptc.sessionMu.Unlock()

	for sessionID, session := range ptc.activeSessions {
		ptc.cleanupSessionObjects(session)
		delete(ptc.activeSessions, sessionID)
	}
}

// cleanupSessionObjects cleans up pooled objects for a session
func (ptc *PooledTypeChecker) cleanupSessionObjects(session *TypeCheckSession) {
	if ptc.poolManager == nil {
		return
	}

	// Reset and return objects to pools
	for _, result := range session.PooledObjects.TypeCheckResults {
		ptc.poolManager.resetTypeCheckResult(result)
	}

	for _, state := range session.PooledObjects.TypeStates {
		ptc.poolManager.resetTypeState(state)
	}

	for _, engine := range session.PooledObjects.InferenceEngines {
		ptc.poolManager.resetInferenceEngine(engine)
	}

	for _, info := range session.PooledObjects.InferredTypeInfos {
		ptc.poolManager.resetInferredTypeInfo(info)
	}

	// Update memory reused statistics
	session.Statistics.MemoryReused += int64(len(session.PooledObjects.TypeCheckResults) * 1024) // Estimated
}

// GetPoolStatistics returns pool manager statistics
func (ptc *PooledTypeChecker) GetPoolStatistics() *TypePoolStats {
	if ptc.poolManager == nil {
		return nil
	}

	stats := ptc.poolManager.GetDetailedStats()
	return &stats
}

// GetCacheStatistics returns resolution pool cache statistics
func (ptc *PooledTypeChecker) GetCacheStatistics() *TypeResolutionCacheStats {
	if ptc.resolutionPool == nil {
		return nil
	}

	stats := ptc.resolutionPool.GetCacheStats()
	return &stats
}

// ResetPools resets all pools for incremental type checking
func (ptc *PooledTypeChecker) ResetPools() {
	if ptc.poolManager != nil {
		ptc.poolManager.Reset()
	}

	if ptc.resolutionPool != nil {
		ptc.resolutionPool.Reset()
	}
}

// SetPoolHooks sets hooks for pool events
func (ptc *PooledTypeChecker) SetPoolHooks(hooks TypePoolHooks) {
	if ptc.poolManager != nil {
		ptc.poolManager.hooks = hooks
	}
}

// SetResolutionPoolHooks sets hooks for resolution pool events
func (ptc *PooledTypeChecker) SetResolutionPoolHooks(hooks TypeResolutionPoolHooks) {
	if ptc.resolutionPool != nil {
		ptc.resolutionPool.hooks = hooks
	}
}

// Helper function to get current time in milliseconds
func getCurrentTimeMs() int64 {
	return getCurrentTime().UnixNano() / 1000000
}

// Helper function to get current time (mockable for testing)
var getCurrentTime = func() time.Time {
	return time.Now()
}

// Global pooled type checker factory instance
var globalPooledTypeCheckerFactory *PooledTypeCheckerFactory
var pooledFactoryOnce sync.Once

// GlobalPooledTypeCheckerFactory returns the global pooled type checker factory
func GlobalPooledTypeCheckerFactory() *PooledTypeCheckerFactory {
	pooledFactoryOnce.Do(func() {
		globalPooledTypeCheckerFactory = NewPooledTypeCheckerFactory(DefaultPooledTypeCheckerConfig())
	})
	return globalPooledTypeCheckerFactory
}

// CreatePooledTypeChecker creates a pooled type checker using the global factory
func CreatePooledTypeChecker(hierarchy *typedef.TypeHierarchy, symbolTable *binder.SymbolTable, moduleName string) *PooledTypeChecker {
	factory := GlobalPooledTypeCheckerFactory()
	return factory.CreateTypeChecker(hierarchy, symbolTable, moduleName)
}
