// ABOUTME: Tests for pooled type checker functionality and session management
// ABOUTME: Ensures efficient memory management and session tracking for type checking operations

package typechecker

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	basetesting "tamarou.com/pvm/internal/testing"
	"tamarou.com/pvm/internal/typedef"
)

func TestPooledTypeCheckerFactory_CreateTypeChecker(t *testing.T) {
	basetesting.SampleTypeCheckerTest(t)
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)

	// Create test dependencies
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	// Test basic creation
	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")
	if ptc == nil {
		t.Fatal("Expected non-nil PooledTypeChecker")
	}
	if ptc.TypeChecker == nil {
		t.Error("Expected non-nil embedded TypeChecker")
	}
	if ptc.poolManager == nil {
		t.Error("Expected non-nil poolManager when pooling is enabled")
	}
	if ptc.resolutionPool == nil {
		t.Error("Expected non-nil resolutionPool when caching is enabled")
	}
	if ptc.activeSessions == nil {
		t.Error("Expected non-nil activeSessions map")
	}

	// Check that underlying type checker is properly initialized
	if ptc.TypeChecker.CurrentModule != "TestModule" {
		t.Errorf("Expected module 'TestModule', got %s", ptc.TypeChecker.CurrentModule)
	}
	if ptc.TypeChecker.Hierarchy != hierarchy {
		t.Error("Expected hierarchy to be set correctly")
	}
	if ptc.TypeChecker.SymbolTable != symbolTable {
		t.Error("Expected symbol table to be set correctly")
	}
}

func TestPooledTypeCheckerFactory_CreateTypeCheckerWithoutPooling(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	config.EnablePooling = false
	config.EnableCaching = false

	factory := NewPooledTypeCheckerFactory(config)

	// Create test dependencies
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	// Test creation without pooling
	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")
	if ptc == nil {
		t.Fatal("Expected non-nil PooledTypeChecker")
	}
	if ptc.TypeChecker == nil {
		t.Error("Expected non-nil embedded TypeChecker")
	}
	if ptc.poolManager != nil {
		t.Error("Expected nil poolManager when pooling is disabled")
	}
	if ptc.resolutionPool != nil {
		t.Error("Expected nil resolutionPool when caching is disabled")
	}
}

func TestPooledTypeCheckerFactory_GetStatistics(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)

	// Create test dependencies
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	// Initial statistics should be zero
	initialStats := factory.GetStatistics()
	if initialStats.CheckersCreated != 0 {
		t.Errorf("Expected 0 checkers created initially, got %d", initialStats.CheckersCreated)
	}

	// Create a type checker
	_ = factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")

	// Statistics should be updated
	updatedStats := factory.GetStatistics()
	if updatedStats.CheckersCreated != 1 {
		t.Errorf("Expected 1 checker created, got %d", updatedStats.CheckersCreated)
	}
}

func TestPooledTypeChecker_NewTypeCheckSession(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")

	// Test session creation
	session := ptc.NewTypeCheckSession("session1", "test.pl")
	if session == nil {
		t.Fatal("Expected non-nil TypeCheckSession")
	}
	if session.SessionID != "session1" {
		t.Errorf("Expected session ID 'session1', got %s", session.SessionID)
	}
	if session.SourceFile != "test.pl" {
		t.Errorf("Expected source file 'test.pl', got %s", session.SourceFile)
	}
	if session.StartTime == 0 {
		t.Error("Expected non-zero start time")
	}
	if session.PooledObjects == nil {
		t.Error("Expected non-nil PooledObjects")
	}

	// Test that session is tracked
	retrievedSession := ptc.GetTypeCheckSession("session1")
	if retrievedSession != session {
		t.Error("Expected to retrieve the same session")
	}

	// Test session that doesn't exist
	nonExistentSession := ptc.GetTypeCheckSession("nonexistent")
	if nonExistentSession != nil {
		t.Error("Expected nil for non-existent session")
	}
}

func TestPooledTypeChecker_CloseTypeCheckSession(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")

	// Create and close a session
	session := ptc.NewTypeCheckSession("session1", "test.pl")
	if session == nil {
		t.Fatal("Expected non-nil session")
	}

	// Verify session exists
	if ptc.GetTypeCheckSession("session1") == nil {
		t.Error("Expected session to exist before closing")
	}

	// Close the session
	ptc.CloseTypeCheckSession("session1")

	// Verify session is removed
	if ptc.GetTypeCheckSession("session1") != nil {
		t.Error("Expected session to be removed after closing")
	}

	// Closing non-existent session should not cause issues
	ptc.CloseTypeCheckSession("nonexistent")
}

func TestPooledTypeChecker_CheckASTWithPooling(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")

	// Create a simple AST for testing
	astRoot := &ast.AST{
		Root:            &ast.BaseNode{},
		TypeAnnotations: []*ast.TypeAnnotation{},
		// SymbolTable and Imports fields not available
	}

	// Test with existing session
	_ = ptc.NewTypeCheckSession("test_session", "test.pl")
	result, err := ptc.CheckASTWithPooling(astRoot, "test_session")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil TypeCheckResult")
	}
	if result.Path != "test.pl" {
		t.Errorf("Expected path 'test.pl', got %s", result.Path)
	}

	// Check that session statistics were updated
	sessionStats := ptc.GetSessionStatistics("test_session")
	if sessionStats == nil {
		t.Fatal("Expected non-nil session statistics")
	}
	if sessionStats.TypeResolutions == 0 {
		t.Error("Expected at least one type resolution to be recorded")
	}
	if sessionStats.ObjectsAllocated == 0 {
		t.Error("Expected at least one object allocation to be recorded")
	}

	// Test with automatic session creation
	result2, err := ptc.CheckASTWithPooling(astRoot, "auto_session")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result2 == nil {
		t.Fatal("Expected non-nil TypeCheckResult")
	}

	// Auto-created session should be cleaned up automatically
	if ptc.GetTypeCheckSession("auto_session") != nil {
		t.Error("Expected auto-created session to be cleaned up")
	}
}

func TestPooledTypeChecker_ResolveTypeWithCaching(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")
	_ = ptc.NewTypeCheckSession("cache_test", "test.pl")

	resolverCalled := 0
	resolver := func() (*TypeResolutionResult, error) {
		resolverCalled++
		return ptc.resolutionPool.NewTypeResolutionResult("cache_key", "CachedType", 0.9, true), nil
	}

	// First call should miss cache
	result1, err := ptc.ResolveTypeWithCaching("cache_key", resolver, "cache_test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result1 == nil {
		t.Fatal("Expected non-nil result")
	}
	if resolverCalled != 1 {
		t.Errorf("Expected resolver to be called once, was called %d times", resolverCalled)
	}

	// Check session statistics
	sessionStats := ptc.GetSessionStatistics("cache_test")
	if sessionStats.CacheMisses == 0 {
		t.Error("Expected at least one cache miss to be recorded")
	}

	// Second call should hit cache
	result2, err := ptc.ResolveTypeWithCaching("cache_key", resolver, "cache_test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result2 == nil {
		t.Fatal("Expected non-nil result")
	}
	if resolverCalled != 1 {
		t.Errorf("Expected resolver to be called once total, was called %d times", resolverCalled)
	}

	// Check that cache hit was recorded
	sessionStatsAfter := ptc.GetSessionStatistics("cache_test")
	if sessionStatsAfter.CacheHits == 0 {
		t.Error("Expected at least one cache hit to be recorded")
	}
}

func TestPooledTypeChecker_CreateInferenceEngineForSession(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")
	_ = ptc.NewTypeCheckSession("inference_test", "test.pl")

	// Test inference engine creation
	engine := ptc.CreateInferenceEngineForSession("inference_test")
	if engine == nil {
		t.Fatal("Expected non-nil InferenceEngine")
	}
	if engine.TypeHierarchy != hierarchy {
		t.Error("Expected TypeHierarchy to be set correctly")
	}
	if engine.SymbolTable != symbolTable {
		t.Error("Expected SymbolTable to be set correctly")
	}

	// Check that object was tracked in session
	sessionStats := ptc.GetSessionStatistics("inference_test")
	if sessionStats.ObjectsAllocated == 0 {
		t.Error("Expected object allocation to be recorded")
	}

	// Test with non-existent session
	engine2 := ptc.CreateInferenceEngineForSession("nonexistent")
	if engine2 == nil {
		t.Fatal("Expected non-nil InferenceEngine even without session")
	}
}

func TestPooledTypeChecker_PerformInferenceWithPooling(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")
	_ = ptc.NewTypeCheckSession("inference_session", "test.pl")

	// Create a simple AST for testing
	astRoot := &ast.AST{
		Root:            &ast.BaseNode{},
		TypeAnnotations: []*ast.TypeAnnotation{},
		// SymbolTable and Imports fields not available
	}

	// Test inference
	err := ptc.PerformInferenceWithPooling(astRoot, "inference_session")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that inference operation was recorded
	sessionStats := ptc.GetSessionStatistics("inference_session")
	if sessionStats.InferenceOperations == 0 {
		t.Error("Expected at least one inference operation to be recorded")
	}
}

func TestPooledTypeChecker_GetSessionStatistics(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")

	// Test with non-existent session
	stats := ptc.GetSessionStatistics("nonexistent")
	if stats != nil {
		t.Error("Expected nil statistics for non-existent session")
	}

	// Create session and test statistics
	session := ptc.NewTypeCheckSession("stats_test", "test.pl")
	if session == nil {
		t.Fatal("Expected non-nil session")
	}

	stats = ptc.GetSessionStatistics("stats_test")
	if stats == nil {
		t.Fatal("Expected non-nil statistics")
	}

	// Initial statistics should be zero
	if stats.ObjectsAllocated != 0 {
		t.Errorf("Expected 0 objects allocated initially, got %d", stats.ObjectsAllocated)
	}
	if stats.CacheHits != 0 {
		t.Errorf("Expected 0 cache hits initially, got %d", stats.CacheHits)
	}
	if stats.CacheMisses != 0 {
		t.Errorf("Expected 0 cache misses initially, got %d", stats.CacheMisses)
	}
}

func TestPooledTypeChecker_GetAllSessionStatistics(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")

	// No sessions initially
	allStats := ptc.GetAllSessionStatistics()
	if len(allStats) != 0 {
		t.Errorf("Expected 0 sessions initially, got %d", len(allStats))
	}

	// Create multiple sessions
	ptc.NewTypeCheckSession("session1", "test1.pl")
	ptc.NewTypeCheckSession("session2", "test2.pl")

	allStats = ptc.GetAllSessionStatistics()
	if len(allStats) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(allStats))
	}

	// Check that both sessions are present
	if _, exists := allStats["session1"]; !exists {
		t.Error("Expected session1 to be present in all statistics")
	}
	if _, exists := allStats["session2"]; !exists {
		t.Error("Expected session2 to be present in all statistics")
	}
}

func TestPooledTypeChecker_CleanupAllSessions(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")

	// Create multiple sessions
	ptc.NewTypeCheckSession("cleanup1", "test1.pl")
	ptc.NewTypeCheckSession("cleanup2", "test2.pl")

	// Verify sessions exist
	if len(ptc.GetAllSessionStatistics()) != 2 {
		t.Error("Expected 2 sessions before cleanup")
	}

	// Cleanup all sessions
	ptc.CleanupAllSessions()

	// Verify all sessions are removed
	if len(ptc.GetAllSessionStatistics()) != 0 {
		t.Error("Expected 0 sessions after cleanup")
	}

	// Verify sessions are actually gone
	if ptc.GetTypeCheckSession("cleanup1") != nil {
		t.Error("Expected cleanup1 session to be removed")
	}
	if ptc.GetTypeCheckSession("cleanup2") != nil {
		t.Error("Expected cleanup2 session to be removed")
	}
}

func TestPooledTypeChecker_GetPoolStatistics(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")

	// Test getting pool statistics
	stats := ptc.GetPoolStatistics()
	if stats == nil {
		t.Fatal("Expected non-nil pool statistics")
	}

	// Should have some baseline values from pool warming
	if stats.TypeCreated == 0 {
		t.Error("Expected some types to be created during pool warming")
	}
}

func TestPooledTypeChecker_GetCacheStatistics(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")

	// Test getting cache statistics
	stats := ptc.GetCacheStatistics()
	if stats == nil {
		t.Fatal("Expected non-nil cache statistics")
	}

	// Should have some baseline values from cache warming
	if stats.Entries == 0 {
		t.Error("Expected some cache entries from cache warming")
	}
}

func TestPooledTypeChecker_ResetPools(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")

	// Create some objects to populate pools
	_ = ptc.NewTypeCheckSession("reset_test", "test.pl")
	_ = ptc.CreateInferenceEngineForSession("reset_test")

	// Get initial statistics
	_ = ptc.GetPoolStatistics()
	cacheStatsBefore := ptc.GetCacheStatistics()

	// Reset pools
	ptc.ResetPools()

	// Pool statistics might change after reset (pools are reset but not cleared)
	poolStatsAfter := ptc.GetPoolStatistics()
	cacheStatsAfter := ptc.GetCacheStatistics()

	// Cache should be cleared
	if cacheStatsAfter.Entries >= cacheStatsBefore.Entries {
		t.Error("Expected cache entries to be reduced after reset")
	}

	// Pools should still be functional
	newEngine := ptc.CreateInferenceEngineForSession("reset_test")
	if newEngine == nil {
		t.Error("Expected pools to be functional after reset")
	}

	// Actually use the pools to trigger statistics update
	if ptc.poolManager != nil {
		// Create a type check result to increment the counter
		result := ptc.poolManager.NewTypeCheckResult("test_path")
		if result != nil {
			// Return it to the pool
			ptc.poolManager.Reset()
		}
	}

	// Pool statistics should reflect continued operation
	poolStatsAfterUse := ptc.GetPoolStatistics()
	if poolStatsAfterUse.TypeCheckCount <= poolStatsAfter.TypeCheckCount {
		t.Error("Expected pool statistics to increase after use following reset")
	}
}

func TestPooledTypeChecker_SetPoolHooks(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")

	hookCalled := false
	hooks := TypePoolHooks{
		OnTypeCheckCreate: func(result *TypeCheckResult) {
			hookCalled = true
		},
	}

	// Set hooks
	ptc.SetPoolHooks(hooks)

	// Create an object that should trigger the hook
	if ptc.poolManager != nil {
		_ = ptc.poolManager.NewTypeCheckResult("hook_test.pl")
	}

	// Verify hook was called
	if !hookCalled {
		t.Error("Expected pool hook to be called")
	}
}

func TestPooledTypeChecker_SetResolutionPoolHooks(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")

	hookCalled := false
	hooks := TypeResolutionPoolHooks{
		OnResolution: func(result *TypeResolutionResult) {
			hookCalled = true
		},
	}

	// Set hooks
	ptc.SetResolutionPoolHooks(hooks)

	// Create an object that should trigger the hook
	if ptc.resolutionPool != nil {
		_ = ptc.resolutionPool.NewTypeResolutionResult("hook_test", "HookType", 0.9, true)
	}

	// Verify hook was called
	if !hookCalled {
		t.Error("Expected resolution pool hook to be called")
	}
}

func TestPooledTypeChecker_DefaultPooledTypeCheckerConfig(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()

	// Test default values
	if !config.EnablePooling {
		t.Error("Expected pooling to be enabled by default")
	}
	if !config.EnableCaching {
		t.Error("Expected caching to be enabled by default")
	}
	if !config.SessionTracking {
		t.Error("Expected session tracking to be enabled by default")
	}
	if !config.AutoCleanup {
		t.Error("Expected auto cleanup to be enabled by default")
	}
	if config.CleanupThreshold <= 0 {
		t.Error("Expected positive cleanup threshold")
	}
	if config.MaxActiveSessions <= 0 {
		t.Error("Expected positive max active sessions")
	}
	if !config.StatisticsCollection {
		t.Error("Expected statistics collection to be enabled by default")
	}
}

func TestPooledTypeChecker_GlobalPooledTypeCheckerFactory(t *testing.T) {
	// Test that global instance is created correctly
	global1 := GlobalPooledTypeCheckerFactory()
	if global1 == nil {
		t.Fatal("Expected non-nil global pooled type checker factory")
	}

	// Test that subsequent calls return the same instance
	global2 := GlobalPooledTypeCheckerFactory()
	if global1 != global2 {
		t.Error("Expected same instance from multiple calls to GlobalPooledTypeCheckerFactory")
	}
}

func TestPooledTypeChecker_CreatePooledTypeChecker(t *testing.T) {
	// Test convenience function
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := CreatePooledTypeChecker(hierarchy, symbolTable, "TestModule")
	if ptc == nil {
		t.Fatal("Expected non-nil PooledTypeChecker")
	}
	if ptc.TypeChecker == nil {
		t.Error("Expected non-nil embedded TypeChecker")
	}
	if ptc.TypeChecker.CurrentModule != "TestModule" {
		t.Errorf("Expected module 'TestModule', got %s", ptc.TypeChecker.CurrentModule)
	}
}

func TestPooledTypeChecker_SessionObjectTracker(t *testing.T) {
	config := DefaultPooledTypeCheckerConfig()
	factory := NewPooledTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	ptc := factory.CreateTypeChecker(hierarchy, symbolTable, "TestModule")
	session := ptc.NewTypeCheckSession("tracker_test", "test.pl")

	// Create various pooled objects
	result := ptc.poolManager.NewTypeCheckResult("test.pl")
	state := ptc.poolManager.NewTypeState()
	_ = ptc.CreateInferenceEngineForSession("tracker_test")

	// Manually add to session tracker (in real usage this happens automatically)
	session.PooledObjects.TypeCheckResults = append(session.PooledObjects.TypeCheckResults, result)
	session.PooledObjects.TypeStates = append(session.PooledObjects.TypeStates, state)
	// engine should already be tracked from CreateInferenceEngineForSession

	// Verify objects are tracked
	if len(session.PooledObjects.TypeCheckResults) != 1 {
		t.Errorf("Expected 1 tracked TypeCheckResult, got %d", len(session.PooledObjects.TypeCheckResults))
	}
	if len(session.PooledObjects.TypeStates) != 1 {
		t.Errorf("Expected 1 tracked TypeState, got %d", len(session.PooledObjects.TypeStates))
	}
	if len(session.PooledObjects.InferenceEngines) != 1 {
		t.Errorf("Expected 1 tracked InferenceEngine, got %d", len(session.PooledObjects.InferenceEngines))
	}

	// Close session and verify cleanup
	ptc.CloseTypeCheckSession("tracker_test")

	// Session should be removed
	if ptc.GetTypeCheckSession("tracker_test") != nil {
		t.Error("Expected session to be removed after closing")
	}
}
