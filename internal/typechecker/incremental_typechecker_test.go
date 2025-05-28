// ABOUTME: Tests for incremental type checking functionality and delta tracking
// ABOUTME: Ensures efficient change detection and cache management for incremental type checking

package typechecker

import (
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/typedef"
)

func TestIncrementalTypeChecker_NewIncrementalTypeChecker(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	if itc == nil {
		t.Fatal("Expected non-nil IncrementalTypeChecker")
	}
	if itc.PooledTypeChecker == nil {
		t.Error("Expected non-nil PooledTypeChecker")
	}
	if itc.deltaTracker == nil {
		t.Error("Expected non-nil deltaTracker")
	}
	if itc.fileStateCache == nil {
		t.Error("Expected non-nil fileStateCache")
	}
	if itc.dependencyGraph == nil {
		t.Error("Expected non-nil dependencyGraph")
	}
	if itc.config != config {
		t.Error("Expected config to be set correctly")
	}
}

func TestIncrementalTypeChecker_DefaultIncrementalConfig(t *testing.T) {
	config := DefaultIncrementalConfig()

	if !config.EnableDeltaTracking {
		t.Error("Expected delta tracking to be enabled by default")
	}
	if !config.EnableCaching {
		t.Error("Expected caching to be enabled by default")
	}
	if !config.EnableDependencyTracking {
		t.Error("Expected dependency tracking to be enabled by default")
	}
	if config.CacheInvalidationStrategy != "conservative" {
		t.Errorf("Expected conservative strategy, got %s", config.CacheInvalidationStrategy)
	}
	if config.MaxCachedFiles <= 0 {
		t.Error("Expected positive max cached files")
	}
	if config.DeltaRetentionTime <= 0 {
		t.Error("Expected positive delta retention time")
	}
}

func TestIncrementalTypeChecker_CheckFileIncremental_NewFile(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	// Create test AST
	astRoot := &ast.AST{
		Root:            &ast.BaseNode{},
		TypeAnnotations: []*ast.TypeAnnotation{},
		// SymbolTable and Imports fields not available in current AST structure
	}

	content := []byte("my $var = 42;")
	filePath := "test.pl"

	// Check new file
	result, err := itc.CheckFileIncremental(filePath, content, astRoot)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil TypeCheckResult")
	}

	// Verify file state was cached
	state, exists := itc.fileStateCache[filePath]
	if !exists {
		t.Error("Expected file state to be cached")
	}
	if state.Checksum == "" {
		t.Error("Expected checksum to be set")
	}
	if state.TypeCheckResult != result {
		t.Error("Expected cached result to match returned result")
	}

	// Verify delta tracking
	deltaInfo := itc.GetDeltaInformation()
	if deltaInfo.ChangedFiles == 0 {
		t.Error("Expected at least one changed file to be tracked")
	}
}

func TestIncrementalTypeChecker_CheckFileIncremental_UnchangedFile(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	// Create test AST
	astRoot := &ast.AST{
		Root:            &ast.BaseNode{},
		TypeAnnotations: []*ast.TypeAnnotation{},
		// SymbolTable and Imports fields not available in current AST structure
	}

	content := []byte("my $var = 42;")
	filePath := "test.pl"

	// First check
	result1, err := itc.CheckFileIncremental(filePath, content, astRoot)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Second check with same content
	result2, err := itc.CheckFileIncremental(filePath, content, astRoot)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should return the same cached result
	if result1 != result2 {
		t.Error("Expected same result for unchanged file")
	}
}

func TestIncrementalTypeChecker_CheckFileIncremental_ChangedFile(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	// Create test AST
	astRoot := &ast.AST{
		Root:            &ast.BaseNode{},
		TypeAnnotations: []*ast.TypeAnnotation{},
		// SymbolTable and Imports fields not available in current AST structure
	}

	filePath := "test.pl"

	// First check
	content1 := []byte("my $var = 42;")
	result1, err := itc.CheckFileIncremental(filePath, content1, astRoot)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Second check with changed content
	content2 := []byte("my $var = 'hello';")
	result2, err := itc.CheckFileIncremental(filePath, content2, astRoot)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should be different results
	if result1 == result2 {
		t.Error("Expected different results for changed file")
	}

	// Check that change was tracked
	itc.deltaTracker.mu.RLock()
	fileChange, exists := itc.deltaTracker.ChangedFiles[filePath]
	itc.deltaTracker.mu.RUnlock()

	if !exists {
		t.Error("Expected file change to be tracked")
	}
	if fileChange.ChangeType != FileModified {
		t.Errorf("Expected FileModified, got %v", fileChange.ChangeType)
	}
	if fileChange.OldChecksum == fileChange.NewChecksum {
		t.Error("Expected different checksums for changed file")
	}
}

func TestIncrementalTypeChecker_CheckProjectIncremental(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	// Create test ASTs
	files := map[string]*ast.AST{
		"file1.pl": {
			Root:            &ast.BaseNode{},
			TypeAnnotations: []*ast.TypeAnnotation{},
			// SymbolTable and Imports fields not available
		},
		"file2.pl": {
			Root:            &ast.BaseNode{},
			TypeAnnotations: []*ast.TypeAnnotation{},
			// SymbolTable and Imports fields not available
		},
	}

	// First project check
	results1, err := itc.CheckProjectIncremental(files)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(results1) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results1))
	}

	// Second project check (should use cache)
	results2, err := itc.CheckProjectIncremental(files)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(results2) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results2))
	}

	// Results should be the same (from cache)
	for filePath := range files {
		if results1[filePath] != results2[filePath] {
			t.Errorf("Expected same result for file %s", filePath)
		}
	}
}

func TestDeltaTracker_NewDeltaTracker(t *testing.T) {
	tracker := NewDeltaTracker()

	if tracker == nil {
		t.Fatal("Expected non-nil DeltaTracker")
	}
	if tracker.ChangedFiles == nil {
		t.Error("Expected non-nil ChangedFiles map")
	}
	if tracker.ChangedSymbols == nil {
		t.Error("Expected non-nil ChangedSymbols map")
	}
	if tracker.ChangedTypes == nil {
		t.Error("Expected non-nil ChangedTypes map")
	}
	if tracker.InvalidatedCaches == nil {
		t.Error("Expected non-nil InvalidatedCaches slice")
	}
	if tracker.SessionID == "" {
		t.Error("Expected non-empty SessionID")
	}
	if tracker.LastUpdate.IsZero() {
		t.Error("Expected non-zero LastUpdate time")
	}
}

func TestDependencyGraph_NewDependencyGraph(t *testing.T) {
	graph := NewDependencyGraph()

	if graph == nil {
		t.Fatal("Expected non-nil DependencyGraph")
	}
	if graph.FileDependencies == nil {
		t.Error("Expected non-nil FileDependencies map")
	}
	if graph.SymbolDependencies == nil {
		t.Error("Expected non-nil SymbolDependencies map")
	}
	if graph.TypeDependencies == nil {
		t.Error("Expected non-nil TypeDependencies map")
	}
	if graph.FileDependents == nil {
		t.Error("Expected non-nil FileDependents map")
	}
	if graph.SymbolDependents == nil {
		t.Error("Expected non-nil SymbolDependents map")
	}
	if graph.TypeDependents == nil {
		t.Error("Expected non-nil TypeDependents map")
	}
}

func TestIncrementalTypeChecker_trackFileChange(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	filePath := "test.pl"
	oldChecksum := "abc123"
	newChecksum := "def456"

	// Track file change
	itc.trackFileChange(filePath, FileModified, oldChecksum, newChecksum)

	// Verify change was tracked
	itc.deltaTracker.mu.RLock()
	change, exists := itc.deltaTracker.ChangedFiles[filePath]
	itc.deltaTracker.mu.RUnlock()

	if !exists {
		t.Fatal("Expected file change to be tracked")
	}
	if change.FilePath != filePath {
		t.Errorf("Expected file path %s, got %s", filePath, change.FilePath)
	}
	if change.ChangeType != FileModified {
		t.Errorf("Expected FileModified, got %v", change.ChangeType)
	}
	if change.OldChecksum != oldChecksum {
		t.Errorf("Expected old checksum %s, got %s", oldChecksum, change.OldChecksum)
	}
	if change.NewChecksum != newChecksum {
		t.Errorf("Expected new checksum %s, got %s", newChecksum, change.NewChecksum)
	}
	if change.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestIncrementalTypeChecker_updateDependencies(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	filePath := "test.pl"
	_ = []string{"dep1.pl", "dep2.pl"} // dependencies not used in current AST structure

	// Create test AST with imports
	astRoot := &ast.AST{
		Root:            &ast.BaseNode{},
		TypeAnnotations: []*ast.TypeAnnotation{},
		// SymbolTable and Imports fields not available
	}

	// Update dependencies
	itc.updateDependencies(filePath, astRoot)

	// Verify dependencies were recorded
	itc.dependencyGraph.mu.RLock()
	recordedDeps := itc.dependencyGraph.FileDependencies[filePath]
	dep1Dependents := itc.dependencyGraph.FileDependents["dep1.pl"]
	dep2Dependents := itc.dependencyGraph.FileDependents["dep2.pl"]
	itc.dependencyGraph.mu.RUnlock()

	if len(recordedDeps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(recordedDeps))
	}
	if recordedDeps[0] != "dep1.pl" || recordedDeps[1] != "dep2.pl" {
		t.Errorf("Expected deps [dep1.pl, dep2.pl], got %v", recordedDeps)
	}

	// Check reverse dependencies
	if len(dep1Dependents) == 0 || dep1Dependents[len(dep1Dependents)-1] != filePath {
		t.Errorf("Expected %s to be dependent of dep1.pl", filePath)
	}
	if len(dep2Dependents) == 0 || dep2Dependents[len(dep2Dependents)-1] != filePath {
		t.Errorf("Expected %s to be dependent of dep2.pl", filePath)
	}
}

func TestIncrementalTypeChecker_invalidateDependentCaches(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	// Set up dependencies
	changedFile := "base.pl"
	dependentFile := "dependent.pl"

	itc.dependencyGraph.FileDependents[changedFile] = []string{dependentFile}

	// Create cached result for dependent file
	itc.fileStateCache[dependentFile] = &FileState{
		FilePath: dependentFile,
		TypeCheckResult: &TypeCheckResult{
			Path: dependentFile,
		},
	}

	// Invalidate caches
	itc.invalidateDependentCaches(changedFile)

	// Verify dependent cache was invalidated
	state := itc.fileStateCache[dependentFile]
	if state.TypeCheckResult != nil {
		t.Error("Expected dependent cache to be invalidated")
	}

	// Verify invalidation was tracked
	if len(itc.deltaTracker.InvalidatedCaches) == 0 {
		t.Error("Expected invalidated cache to be tracked")
	}
	if itc.deltaTracker.InvalidatedCaches[0] != dependentFile {
		t.Errorf("Expected %s to be in invalidated caches", dependentFile)
	}
}

func TestIncrementalTypeChecker_calculateFilesToRecheck(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()
	config.CacheInvalidationStrategy = "conservative"

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	// Set up dependency chain: base.pl -> mid.pl -> top.pl
	itc.dependencyGraph.FileDependents["base.pl"] = []string{"mid.pl"}
	itc.dependencyGraph.FileDependents["mid.pl"] = []string{"top.pl"}

	changedFiles := []string{"base.pl"}
	filesToCheck := itc.calculateFilesToRecheck(changedFiles)

	// Should include all files in the dependency chain
	expectedFiles := map[string]bool{
		"base.pl": true,
		"mid.pl":  true,
		"top.pl":  true,
	}

	if len(filesToCheck) != len(expectedFiles) {
		t.Errorf("Expected %d files to check, got %d", len(expectedFiles), len(filesToCheck))
	}

	for _, file := range filesToCheck {
		if !expectedFiles[file] {
			t.Errorf("Unexpected file %s in files to check", file)
		}
	}
}

func TestIncrementalTypeChecker_calculateCheckOrder(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	// Set up dependencies: dep.pl <- main.pl
	itc.dependencyGraph.FileDependencies["main.pl"] = []string{"dep.pl"}

	files := []string{"main.pl", "dep.pl"}
	checkOrder := itc.calculateCheckOrder(files)

	// dep.pl should come before main.pl
	if len(checkOrder) != 2 {
		t.Errorf("Expected 2 files in check order, got %d", len(checkOrder))
	}

	depIndex := -1
	mainIndex := -1
	for i, file := range checkOrder {
		if file == "dep.pl" {
			depIndex = i
		} else if file == "main.pl" {
			mainIndex = i
		}
	}

	if depIndex == -1 || mainIndex == -1 {
		t.Error("Expected both files to be in check order")
	}
	if depIndex > mainIndex {
		t.Error("Expected dep.pl to come before main.pl in check order")
	}
}

func TestIncrementalTypeChecker_GetDeltaInformation(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	// Add some changes
	itc.trackFileChange("test1.pl", FileModified, "old", "new")
	itc.trackFileChange("test2.pl", FileAdded, "", "new")

	deltaInfo := itc.GetDeltaInformation()

	if deltaInfo == nil {
		t.Fatal("Expected non-nil DeltaInformation")
	}
	if deltaInfo.SessionID == "" {
		t.Error("Expected non-empty SessionID")
	}
	if deltaInfo.ChangedFiles != 2 {
		t.Errorf("Expected 2 changed files, got %d", deltaInfo.ChangedFiles)
	}
	if deltaInfo.LastUpdate.IsZero() {
		t.Error("Expected non-zero LastUpdate time")
	}
}

func TestIncrementalTypeChecker_GetCacheStatistics(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()
	config.MaxCachedFiles = 10

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	// Add some cached files
	itc.fileStateCache["file1.pl"] = &FileState{
		FilePath:        "file1.pl",
		TypeCheckResult: &TypeCheckResult{},
	}
	itc.fileStateCache["file2.pl"] = &FileState{
		FilePath:        "file2.pl",
		TypeCheckResult: nil, // Cache miss
	}

	stats := itc.GetCacheStatistics()

	if stats == nil {
		t.Fatal("Expected non-nil IncrementalCacheStats")
	}
	if stats.CachedFiles != 2 {
		t.Errorf("Expected 2 cached files, got %d", stats.CachedFiles)
	}
	if stats.MaxFiles != 10 {
		t.Errorf("Expected max files 10, got %d", stats.MaxFiles)
	}
	if stats.CacheHits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", stats.CacheHits)
	}
	if stats.CacheMisses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.CacheMisses)
	}
	if stats.HitRate != 50.0 {
		t.Errorf("Expected hit rate 50%%, got %f", stats.HitRate)
	}
	if stats.Utilization != 20.0 {
		t.Errorf("Expected utilization 20%%, got %f", stats.Utilization)
	}
}

func TestIncrementalTypeChecker_ClearCache(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	// Add cached files
	itc.fileStateCache["file1.pl"] = &FileState{}
	itc.fileStateCache["file2.pl"] = &FileState{}

	if len(itc.fileStateCache) != 2 {
		t.Error("Expected 2 cached files before clear")
	}

	// Clear cache
	itc.ClearCache()

	if len(itc.fileStateCache) != 0 {
		t.Error("Expected 0 cached files after clear")
	}
}

func TestIncrementalTypeChecker_ClearDeltas(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	// Add some deltas
	itc.trackFileChange("test.pl", FileModified, "old", "new")
	itc.deltaTracker.InvalidatedCaches = append(itc.deltaTracker.InvalidatedCaches, "cache1")

	if len(itc.deltaTracker.ChangedFiles) == 0 {
		t.Error("Expected changed files before clear")
	}
	if len(itc.deltaTracker.InvalidatedCaches) == 0 {
		t.Error("Expected invalidated caches before clear")
	}

	// Clear deltas
	itc.ClearDeltas()

	if len(itc.deltaTracker.ChangedFiles) != 0 {
		t.Error("Expected 0 changed files after clear")
	}
	if len(itc.deltaTracker.InvalidatedCaches) != 0 {
		t.Error("Expected 0 invalidated caches after clear")
	}
}

func TestIncrementalTypeChecker_ResetIncremental(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	// Add some state
	itc.fileStateCache["file1.pl"] = &FileState{}
	itc.trackFileChange("test.pl", FileModified, "old", "new")
	itc.dependencyGraph.FileDependencies["test.pl"] = []string{"dep.pl"}

	// Reset
	itc.ResetIncremental()

	// Verify all state is cleared
	if len(itc.fileStateCache) != 0 {
		t.Error("Expected cache to be cleared")
	}
	if len(itc.deltaTracker.ChangedFiles) != 0 {
		t.Error("Expected deltas to be cleared")
	}
	if len(itc.dependencyGraph.FileDependencies) != 0 {
		t.Error("Expected dependency graph to be cleared")
	}
}

func TestIncrementalTypeChecker_calculateChecksum(t *testing.T) {
	content1 := []byte("my $var = 42;")
	content2 := []byte("my $var = 43;")
	content3 := []byte("my $var = 42;") // Same as content1

	checksum1 := calculateChecksum(content1)
	checksum2 := calculateChecksum(content2)
	checksum3 := calculateChecksum(content3)

	if checksum1 == "" {
		t.Error("Expected non-empty checksum")
	}
	if checksum1 == checksum2 {
		t.Error("Expected different checksums for different content")
	}
	if checksum1 != checksum3 {
		t.Error("Expected same checksums for same content")
	}
}

func TestIncrementalTypeChecker_generateSessionID(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()

	if id1 == "" {
		t.Error("Expected non-empty session ID")
	}
	if id2 == "" {
		t.Error("Expected non-empty session ID")
	}
	if id1 == id2 {
		t.Error("Expected different session IDs")
	}
	if !strings.HasPrefix(id1, "session_") {
		t.Error("Expected session ID to start with 'session_'")
	}
}

func TestIncrementalTypeCheckerFactory_CreateIncrementalTypeChecker(t *testing.T) {
	config := DefaultIncrementalConfig()
	factory := NewIncrementalTypeCheckerFactory(config)
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	itc := factory.CreateIncrementalTypeChecker(hierarchy, symbolTable, "TestModule")

	if itc == nil {
		t.Fatal("Expected non-nil IncrementalTypeChecker")
	}
	if itc.config != config {
		t.Error("Expected config to be set correctly")
	}
	if itc.TypeChecker.CurrentModule != "TestModule" {
		t.Errorf("Expected module 'TestModule', got %s", itc.TypeChecker.CurrentModule)
	}
}

func TestIncrementalTypeChecker_GlobalIncrementalTypeCheckerFactory(t *testing.T) {
	// Test that global instance is created correctly
	global1 := GlobalIncrementalTypeCheckerFactory()
	if global1 == nil {
		t.Fatal("Expected non-nil global incremental type checker factory")
	}

	// Test that subsequent calls return the same instance
	global2 := GlobalIncrementalTypeCheckerFactory()
	if global1 != global2 {
		t.Error("Expected same instance from multiple calls to GlobalIncrementalTypeCheckerFactory")
	}
}

func TestIncrementalTypeChecker_CreateIncrementalTypeChecker(t *testing.T) {
	// Test convenience function
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	itc := CreateIncrementalTypeChecker(hierarchy, symbolTable, "TestModule")
	if itc == nil {
		t.Fatal("Expected non-nil IncrementalTypeChecker")
	}
	if itc.TypeChecker.CurrentModule != "TestModule" {
		t.Errorf("Expected module 'TestModule', got %s", itc.TypeChecker.CurrentModule)
	}
}

func TestIncrementalTypeChecker_CacheEviction(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()
	config := DefaultIncrementalConfig()
	config.MaxCachedFiles = 2 // Very small cache for testing

	itc := NewIncrementalTypeChecker(hierarchy, symbolTable, "TestModule", config)

	// Add files to fill cache beyond capacity
	oldTime := time.Now().Add(-1 * time.Hour)
	itc.fileStateCache["old.pl"] = &FileState{
		FilePath:     "old.pl",
		LastModified: oldTime,
	}
	itc.fileStateCache["newer.pl"] = &FileState{
		FilePath:     "newer.pl",
		LastModified: time.Now(),
	}

	// This should trigger eviction of the oldest entry
	itc.updateFileStateCache("newest.pl", "checksum", &TypeCheckResult{}, symbolTable)

	// The old file should be evicted
	if _, exists := itc.fileStateCache["old.pl"]; exists {
		t.Error("Expected oldest file to be evicted from cache")
	}
	if _, exists := itc.fileStateCache["newer.pl"]; !exists {
		t.Error("Expected newer file to remain in cache")
	}
	if _, exists := itc.fileStateCache["newest.pl"]; !exists {
		t.Error("Expected newest file to be added to cache")
	}
}
