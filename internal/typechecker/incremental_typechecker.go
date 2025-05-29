// ABOUTME: Incremental type checking implementation with pool reuse for efficient updates
// ABOUTME: Provides delta-based type checking that reuses pooled objects and cached results

package typechecker

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/typedef"
)

// IncrementalTypeChecker extends PooledTypeChecker with incremental capabilities
type IncrementalTypeChecker struct {
	*PooledTypeChecker

	// Delta tracking for incremental updates
	deltaTracker *DeltaTracker

	// File state cache for change detection
	fileStateCache map[string]*FileState

	// Dependency graph for invalidation
	dependencyGraph *DependencyGraph

	// Incremental configuration
	config IncrementalConfig

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// DeltaTracker tracks changes for incremental type checking
type DeltaTracker struct {
	// ChangedFiles tracks files that have been modified
	ChangedFiles map[string]*FileChange

	// ChangedSymbols tracks symbols that have been modified
	ChangedSymbols map[string]*SymbolChange

	// ChangedTypes tracks types that have been modified
	ChangedTypes map[string]*TypeChange

	// Invalidated caches that need to be updated
	InvalidatedCaches []string

	// Incremental session ID for tracking
	SessionID string

	// Last update timestamp
	LastUpdate time.Time

	mu sync.RWMutex
}

// FileChange represents a change to a source file
type FileChange struct {
	// FilePath is the path to the changed file
	FilePath string

	// ChangeType indicates the type of change
	ChangeType FileChangeType

	// OldChecksum is the checksum before the change
	OldChecksum string

	// NewChecksum is the checksum after the change
	NewChecksum string

	// AffectedRanges lists the ranges of code that changed
	AffectedRanges []SourceRange

	// Timestamp when the change was detected
	Timestamp time.Time
}

// SourceRange represents a range in source code
type SourceRange struct {
	Start ast.Position
	End   ast.Position
}

// FileChangeType categorizes file changes
type FileChangeType int

const (
	// FileAdded indicates a new file was added
	FileAdded FileChangeType = iota
	// FileModified indicates an existing file was modified
	FileModified
	// FileDeleted indicates a file was deleted
	FileDeleted
	// FileRenamed indicates a file was renamed
	FileRenamed
)

// SymbolChange represents a change to a symbol
type SymbolChange struct {
	// SymbolName is the name of the changed symbol
	SymbolName string

	// ChangeType indicates the type of change
	ChangeType SymbolChangeType

	// OldSymbol is the symbol before the change
	OldSymbol *binder.Symbol

	// NewSymbol is the symbol after the change
	NewSymbol *binder.Symbol

	// AffectedFiles lists files that use this symbol
	AffectedFiles []string

	// Timestamp when the change was detected
	Timestamp time.Time
}

// SymbolChangeType categorizes symbol changes
type SymbolChangeType int

const (
	// SymbolAdded indicates a new symbol was added
	SymbolAdded SymbolChangeType = iota
	// SymbolModified indicates an existing symbol was modified
	SymbolModified
	// SymbolDeleted indicates a symbol was deleted
	SymbolDeleted
	// SymbolTypeChanged indicates a symbol's type was changed
	SymbolTypeChanged
)

// TypeChange represents a change to a type definition
type TypeChange struct {
	// TypeName is the name of the changed type
	TypeName string

	// ChangeType indicates the type of change
	ChangeType TypeChangeType

	// OldType is the type before the change
	OldType string

	// NewType is the type after the change
	NewType string

	// AffectedSymbols lists symbols that use this type
	AffectedSymbols []string

	// Timestamp when the change was detected
	Timestamp time.Time
}

// TypeChangeType categorizes type changes
type TypeChangeType int

const (
	// TypeAdded indicates a new type was defined
	TypeAdded TypeChangeType = iota
	// TypeModified indicates an existing type was modified
	TypeModified
	// TypeDeleted indicates a type was deleted
	TypeDeleted
	// TypeHierarchyChanged indicates the type hierarchy changed
	TypeHierarchyChanged
)

// FileState represents the state of a file for change detection
type FileState struct {
	// FilePath is the path to the file
	FilePath string

	// Checksum is the content checksum for change detection
	Checksum string

	// LastModified is the last modification time
	LastModified time.Time

	// TypeCheckResult is the cached type check result
	TypeCheckResult *TypeCheckResult

	// SymbolTable is the cached symbol table for the file
	SymbolTable *binder.SymbolTable

	// Dependencies lists files this file depends on
	Dependencies []string

	// Dependents lists files that depend on this file
	Dependents []string
}

// DependencyGraph tracks dependencies between files and symbols
type DependencyGraph struct {
	// FileDependencies maps files to their dependencies
	FileDependencies map[string][]string

	// SymbolDependencies maps symbols to files that use them
	SymbolDependencies map[string][]string

	// TypeDependencies maps types to symbols that use them
	TypeDependencies map[string][]string

	// Reverse lookups for efficient invalidation
	FileDependents   map[string][]string
	SymbolDependents map[string][]string
	TypeDependents   map[string][]string

	mu sync.RWMutex
}

// IncrementalConfig contains configuration for incremental type checking
type IncrementalConfig struct {
	// EnableDeltaTracking controls whether delta tracking is enabled
	EnableDeltaTracking bool

	// EnableCaching controls whether file state caching is enabled
	EnableCaching bool

	// EnableDependencyTracking controls whether dependency tracking is enabled
	EnableDependencyTracking bool

	// CacheInvalidationStrategy determines how caches are invalidated
	CacheInvalidationStrategy string // "conservative", "optimistic", "precise"

	// MaxCachedFiles is the maximum number of files to cache
	MaxCachedFiles int

	// DeltaRetentionTime is how long to keep delta information
	DeltaRetentionTime time.Duration

	// EnableFileWatching controls whether file system watching is enabled
	EnableFileWatching bool
}

// DefaultIncrementalConfig returns default incremental configuration
func DefaultIncrementalConfig() IncrementalConfig {
	return IncrementalConfig{
		EnableDeltaTracking:       true,
		EnableCaching:             true,
		EnableDependencyTracking:  true,
		CacheInvalidationStrategy: "conservative",
		MaxCachedFiles:            1000,
		DeltaRetentionTime:        24 * time.Hour,
		EnableFileWatching:        false,
	}
}

// NewIncrementalTypeChecker creates a new incremental type checker
func NewIncrementalTypeChecker(hierarchy *typedef.TypeHierarchy, symbolTable *binder.SymbolTable, moduleName string, config IncrementalConfig) *IncrementalTypeChecker {
	pooledTC := CreatePooledTypeChecker(hierarchy, symbolTable, moduleName)

	itc := &IncrementalTypeChecker{
		PooledTypeChecker: pooledTC,
		deltaTracker:      NewDeltaTracker(),
		fileStateCache:    make(map[string]*FileState),
		dependencyGraph:   NewDependencyGraph(),
		config:            config,
	}

	return itc
}

// NewDeltaTracker creates a new delta tracker
func NewDeltaTracker() *DeltaTracker {
	return &DeltaTracker{
		ChangedFiles:      make(map[string]*FileChange),
		ChangedSymbols:    make(map[string]*SymbolChange),
		ChangedTypes:      make(map[string]*TypeChange),
		InvalidatedCaches: []string{},
		SessionID:         generateSessionID(),
		LastUpdate:        time.Now(),
	}
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		FileDependencies:   make(map[string][]string),
		SymbolDependencies: make(map[string][]string),
		TypeDependencies:   make(map[string][]string),
		FileDependents:     make(map[string][]string),
		SymbolDependents:   make(map[string][]string),
		TypeDependents:     make(map[string][]string),
	}
}

// CheckFileIncremental performs incremental type checking on a file
func (itc *IncrementalTypeChecker) CheckFileIncremental(filePath string, content []byte, ast *ast.AST) (*TypeCheckResult, error) {
	itc.mu.Lock()
	defer itc.mu.Unlock()

	// Calculate content checksum for change detection
	checksum := calculateChecksum(content)

	// Check if file has changed
	fileState, exists := itc.fileStateCache[filePath]
	if exists && fileState.Checksum == checksum {
		// File hasn't changed, return cached result if available
		if fileState.TypeCheckResult != nil {
			return fileState.TypeCheckResult, nil
		}
	}

	// Track file change
	if itc.config.EnableDeltaTracking {
		oldChecksum := ""
		if exists {
			oldChecksum = fileState.Checksum
		}
		itc.trackFileChange(filePath, FileModified, oldChecksum, checksum)
	}

	// Invalidate dependent caches
	if itc.config.EnableDependencyTracking {
		itc.invalidateDependentCaches(filePath)
	}

	// Create incremental session
	sessionID := fmt.Sprintf("incremental_%s_%d", filePath, time.Now().UnixNano())
	_ = itc.NewTypeCheckSession(sessionID, filePath)
	defer itc.CloseTypeCheckSession(sessionID)

	// Perform type checking with pool reuse
	result, err := itc.CheckASTWithPooling(ast, sessionID)
	if err != nil {
		return nil, err
	}

	// Update file state cache
	if itc.config.EnableCaching {
		itc.updateFileStateCache(filePath, checksum, result, nil)
	}

	// Update dependency graph
	if itc.config.EnableDependencyTracking {
		itc.updateDependencies(filePath, ast)
	}

	return result, nil
}

// CheckProjectIncremental performs incremental type checking on an entire project
func (itc *IncrementalTypeChecker) CheckProjectIncremental(files map[string]*ast.AST) (map[string]*TypeCheckResult, error) {
	itc.mu.Lock()
	defer itc.mu.Unlock()

	results := make(map[string]*TypeCheckResult)
	changedFiles := make([]string, 0)

	// Detect changed files
	for filePath, ast := range files {
		// In a real implementation, we'd get the file content
		// For now, assume content is available in AST or elsewhere
		content := []byte(ast.Root.Text()) // Simplified
		checksum := calculateChecksum(content)

		fileState, exists := itc.fileStateCache[filePath]
		if !exists || fileState.Checksum != checksum {
			changedFiles = append(changedFiles, filePath)
		}
	}

	// If no files changed, return cached results
	if len(changedFiles) == 0 {
		for filePath := range files {
			if state, exists := itc.fileStateCache[filePath]; exists && state.TypeCheckResult != nil {
				results[filePath] = state.TypeCheckResult
			}
		}
		return results, nil
	}

	// Determine which files need re-checking based on dependencies
	filesToCheck := itc.calculateFilesToRecheck(changedFiles)

	// Create incremental session for project
	sessionID := fmt.Sprintf("project_incremental_%d", time.Now().UnixNano())
	_ = itc.NewTypeCheckSession(sessionID, "project")
	defer itc.CloseTypeCheckSession(sessionID)

	// Check files in dependency order
	checkOrder := itc.calculateCheckOrder(filesToCheck)
	for _, filePath := range checkOrder {
		ast, exists := files[filePath]
		if !exists {
			continue
		}

		result, err := itc.CheckASTWithPooling(ast, sessionID)
		if err != nil {
			return nil, fmt.Errorf("error checking file %s: %w", filePath, err)
		}

		results[filePath] = result

		// Update caches and dependencies
		content := []byte(ast.Root.Text()) // Simplified
		checksum := calculateChecksum(content)

		if itc.config.EnableCaching {
			itc.updateFileStateCache(filePath, checksum, result, nil)
		}

		if itc.config.EnableDependencyTracking {
			itc.updateDependencies(filePath, ast)
		}
	}

	return results, nil
}

// trackFileChange records a file change in the delta tracker
func (itc *IncrementalTypeChecker) trackFileChange(filePath string, changeType FileChangeType, oldChecksum, newChecksum string) {
	itc.deltaTracker.mu.Lock()
	defer itc.deltaTracker.mu.Unlock()

	change := &FileChange{
		FilePath:    filePath,
		ChangeType:  changeType,
		OldChecksum: oldChecksum,
		NewChecksum: newChecksum,
		Timestamp:   time.Now(),
	}

	itc.deltaTracker.ChangedFiles[filePath] = change
	itc.deltaTracker.LastUpdate = time.Now()
}

// invalidateDependentCaches invalidates caches for files that depend on the changed file
func (itc *IncrementalTypeChecker) invalidateDependentCaches(filePath string) {
	itc.dependencyGraph.mu.RLock()
	dependents := itc.dependencyGraph.FileDependents[filePath]
	itc.dependencyGraph.mu.RUnlock()

	for _, dependent := range dependents {
		// Invalidate cached result for dependent file
		if state, exists := itc.fileStateCache[dependent]; exists {
			state.TypeCheckResult = nil
		}

		// Add to invalidation list
		itc.deltaTracker.InvalidatedCaches = append(itc.deltaTracker.InvalidatedCaches, dependent)
	}
}

// updateFileStateCache updates the file state cache with new information
func (itc *IncrementalTypeChecker) updateFileStateCache(filePath, checksum string, result *TypeCheckResult, symbolTable *binder.SymbolTable) {
	// Ensure cache doesn't exceed maximum size
	if len(itc.fileStateCache) >= itc.config.MaxCachedFiles {
		itc.evictOldestCacheEntry()
	}

	state := &FileState{
		FilePath:        filePath,
		Checksum:        checksum,
		LastModified:    time.Now(),
		TypeCheckResult: result,
		SymbolTable:     symbolTable,
		Dependencies:    []string{},
		Dependents:      []string{},
	}

	itc.fileStateCache[filePath] = state
}

// updateDependencies updates the dependency graph with information from the AST
func (itc *IncrementalTypeChecker) updateDependencies(filePath string, ast *ast.AST) {
	itc.dependencyGraph.mu.Lock()
	defer itc.dependencyGraph.mu.Unlock()

	// Extract dependencies from imports (placeholder - AST doesn't have Imports field yet)
	dependencies := []string{}

	// Update file dependencies
	itc.dependencyGraph.FileDependencies[filePath] = dependencies

	// Update reverse dependencies
	for _, dep := range dependencies {
		if itc.dependencyGraph.FileDependents[dep] == nil {
			itc.dependencyGraph.FileDependents[dep] = []string{}
		}
		itc.dependencyGraph.FileDependents[dep] = append(itc.dependencyGraph.FileDependents[dep], filePath)
	}

	// Update file state cache with dependency information
	if state, exists := itc.fileStateCache[filePath]; exists {
		state.Dependencies = dependencies
	}
}

// calculateFilesToRecheck determines which files need to be re-checked based on changes
func (itc *IncrementalTypeChecker) calculateFilesToRecheck(changedFiles []string) []string {
	filesToCheck := make(map[string]bool)

	// Always check changed files
	for _, file := range changedFiles {
		filesToCheck[file] = true
	}

	// Add dependent files based on strategy
	switch itc.config.CacheInvalidationStrategy {
	case "conservative":
		// Re-check all files that transitively depend on changed files
		itc.addTransitiveDependents(changedFiles, filesToCheck)

	case "optimistic":
		// Only re-check direct dependents
		itc.addDirectDependents(changedFiles, filesToCheck)

	case "precise":
		// Analyze specific changes to determine minimal set
		itc.addPreciseDependents(changedFiles, filesToCheck)
	}

	// Convert to slice
	result := make([]string, 0, len(filesToCheck))
	for file := range filesToCheck {
		result = append(result, file)
	}

	return result
}

// addTransitiveDependents adds all transitive dependents to the files to check
func (itc *IncrementalTypeChecker) addTransitiveDependents(changedFiles []string, filesToCheck map[string]bool) {
	itc.dependencyGraph.mu.RLock()
	defer itc.dependencyGraph.mu.RUnlock()

	visited := make(map[string]bool)
	var addDependents func(string)

	addDependents = func(file string) {
		if visited[file] {
			return
		}
		visited[file] = true

		dependents := itc.dependencyGraph.FileDependents[file]
		for _, dependent := range dependents {
			filesToCheck[dependent] = true
			addDependents(dependent)
		}
	}

	for _, file := range changedFiles {
		addDependents(file)
	}
}

// addDirectDependents adds only direct dependents to the files to check
func (itc *IncrementalTypeChecker) addDirectDependents(changedFiles []string, filesToCheck map[string]bool) {
	itc.dependencyGraph.mu.RLock()
	defer itc.dependencyGraph.mu.RUnlock()

	for _, file := range changedFiles {
		dependents := itc.dependencyGraph.FileDependents[file]
		for _, dependent := range dependents {
			filesToCheck[dependent] = true
		}
	}
}

// addPreciseDependents adds dependents based on precise change analysis
func (itc *IncrementalTypeChecker) addPreciseDependents(changedFiles []string, filesToCheck map[string]bool) {
	// For now, use conservative strategy
	// In a full implementation, this would analyze the specific changes
	// and only invalidate what's actually affected
	itc.addTransitiveDependents(changedFiles, filesToCheck)
}

// calculateCheckOrder determines the order to check files based on dependencies
func (itc *IncrementalTypeChecker) calculateCheckOrder(files []string) []string {
	itc.dependencyGraph.mu.RLock()
	defer itc.dependencyGraph.mu.RUnlock()

	// Simple topological sort
	visited := make(map[string]bool)
	result := make([]string, 0, len(files))

	var visit func(string)
	visit = func(file string) {
		if visited[file] {
			return
		}
		visited[file] = true

		// Visit dependencies first
		dependencies := itc.dependencyGraph.FileDependencies[file]
		for _, dep := range dependencies {
			// Only visit if it's in our files to check
			for _, f := range files {
				if f == dep {
					visit(dep)
					break
				}
			}
		}

		result = append(result, file)
	}

	for _, file := range files {
		visit(file)
	}

	return result
}

// evictOldestCacheEntry removes the oldest entry from the file state cache
func (itc *IncrementalTypeChecker) evictOldestCacheEntry() {
	var oldestFile string
	var oldestTime time.Time = time.Now()

	for filePath, state := range itc.fileStateCache {
		if state.LastModified.Before(oldestTime) {
			oldestTime = state.LastModified
			oldestFile = filePath
		}
	}

	if oldestFile != "" {
		delete(itc.fileStateCache, oldestFile)
	}
}

// GetDeltaInformation returns current delta tracking information
func (itc *IncrementalTypeChecker) GetDeltaInformation() *DeltaInformation {
	itc.deltaTracker.mu.RLock()
	defer itc.deltaTracker.mu.RUnlock()

	return &DeltaInformation{
		SessionID:         itc.deltaTracker.SessionID,
		LastUpdate:        itc.deltaTracker.LastUpdate,
		ChangedFiles:      len(itc.deltaTracker.ChangedFiles),
		ChangedSymbols:    len(itc.deltaTracker.ChangedSymbols),
		ChangedTypes:      len(itc.deltaTracker.ChangedTypes),
		InvalidatedCaches: len(itc.deltaTracker.InvalidatedCaches),
	}
}

// DeltaInformation contains summary information about deltas
type DeltaInformation struct {
	SessionID         string    // Incremental session ID
	LastUpdate        time.Time // Last update timestamp
	ChangedFiles      int       // Number of changed files
	ChangedSymbols    int       // Number of changed symbols
	ChangedTypes      int       // Number of changed types
	InvalidatedCaches int       // Number of invalidated caches
}

// GetCacheStatistics returns cache statistics for the incremental checker
func (itc *IncrementalTypeChecker) GetCacheStatistics() *IncrementalCacheStats {
	itc.mu.RLock()
	defer itc.mu.RUnlock()

	cacheHits := 0
	cacheMisses := 0

	for _, state := range itc.fileStateCache {
		if state.TypeCheckResult != nil {
			cacheHits++
		} else {
			cacheMisses++
		}
	}

	totalRequests := cacheHits + cacheMisses
	hitRate := float64(0)
	if totalRequests > 0 {
		hitRate = float64(cacheHits) / float64(totalRequests) * 100
	}

	return &IncrementalCacheStats{
		CachedFiles: len(itc.fileStateCache),
		MaxFiles:    itc.config.MaxCachedFiles,
		CacheHits:   cacheHits,
		CacheMisses: cacheMisses,
		HitRate:     hitRate,
		Utilization: float64(len(itc.fileStateCache)) / float64(itc.config.MaxCachedFiles) * 100,
	}
}

// IncrementalCacheStats contains cache statistics for incremental checking
type IncrementalCacheStats struct {
	CachedFiles int     // Number of files in cache
	MaxFiles    int     // Maximum files that can be cached
	CacheHits   int     // Number of cache hits
	CacheMisses int     // Number of cache misses
	HitRate     float64 // Cache hit rate as percentage
	Utilization float64 // Cache utilization as percentage
}

// ClearCache clears all cached file states
func (itc *IncrementalTypeChecker) ClearCache() {
	itc.mu.Lock()
	defer itc.mu.Unlock()

	itc.fileStateCache = make(map[string]*FileState)
}

// ClearDeltas clears all delta tracking information
func (itc *IncrementalTypeChecker) ClearDeltas() {
	itc.deltaTracker.mu.Lock()
	defer itc.deltaTracker.mu.Unlock()

	itc.deltaTracker.ChangedFiles = make(map[string]*FileChange)
	itc.deltaTracker.ChangedSymbols = make(map[string]*SymbolChange)
	itc.deltaTracker.ChangedTypes = make(map[string]*TypeChange)
	itc.deltaTracker.InvalidatedCaches = []string{}
	itc.deltaTracker.LastUpdate = time.Now()
}

// ResetIncremental resets all incremental state while preserving pools
func (itc *IncrementalTypeChecker) ResetIncremental() {
	itc.ClearCache()
	itc.ClearDeltas()
	itc.dependencyGraph = NewDependencyGraph()
	itc.ResetPools()
}

// Helper functions

// calculateChecksum calculates a SHA256 checksum for content
func calculateChecksum(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// Global incremental type checker factory
var globalIncrementalFactory *IncrementalTypeCheckerFactory
var incrementalFactoryOnce sync.Once

// IncrementalTypeCheckerFactory creates incremental type checkers
type IncrementalTypeCheckerFactory struct {
	config IncrementalConfig
}

// NewIncrementalTypeCheckerFactory creates a new incremental type checker factory
func NewIncrementalTypeCheckerFactory(config IncrementalConfig) *IncrementalTypeCheckerFactory {
	return &IncrementalTypeCheckerFactory{
		config: config,
	}
}

// CreateIncrementalTypeChecker creates a new incremental type checker
func (factory *IncrementalTypeCheckerFactory) CreateIncrementalTypeChecker(hierarchy *typedef.TypeHierarchy, symbolTable *binder.SymbolTable, moduleName string) *IncrementalTypeChecker {
	return NewIncrementalTypeChecker(hierarchy, symbolTable, moduleName, factory.config)
}

// GlobalIncrementalTypeCheckerFactory returns the global incremental type checker factory
func GlobalIncrementalTypeCheckerFactory() *IncrementalTypeCheckerFactory {
	incrementalFactoryOnce.Do(func() {
		globalIncrementalFactory = NewIncrementalTypeCheckerFactory(DefaultIncrementalConfig())
	})
	return globalIncrementalFactory
}

// CreateIncrementalTypeChecker creates an incremental type checker using the global factory
func CreateIncrementalTypeChecker(hierarchy *typedef.TypeHierarchy, symbolTable *binder.SymbolTable, moduleName string) *IncrementalTypeChecker {
	factory := GlobalIncrementalTypeCheckerFactory()
	return factory.CreateIncrementalTypeChecker(hierarchy, symbolTable, moduleName)
}
