// ABOUTME: Advanced dependency resolver with conflict resolution and optimization
// ABOUTME: Implements sophisticated algorithms for complex dependency scenarios

package deps

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// ConflictResolutionStrategy defines how to handle version conflicts
type ConflictResolutionStrategy int

const (
	// StrategyFailFast fails immediately on any conflict
	StrategyFailFast ConflictResolutionStrategy = iota
	// StrategyLatestCompatible picks the latest version that satisfies all constraints
	StrategyLatestCompatible
	// StrategyMinimalVersion picks the minimal version that satisfies all constraints
	StrategyMinimalVersion
	// StrategyPreferExisting prefers already-resolved versions when possible
	StrategyPreferExisting
)

// OptimizationStrategy defines how to optimize dependency resolution
type OptimizationStrategy int

const (
	// OptimizeNone performs no optimization
	OptimizeNone OptimizationStrategy = iota
	// OptimizeMinimalTree minimizes the number of dependencies
	OptimizeMinimalTree
	// OptimizeSharedDependencies maximizes sharing of common dependencies
	OptimizeSharedDependencies
	// OptimizeParallel uses parallel resolution for faster processing
	OptimizeParallel
)

// AdvancedResolutionOptions extends basic options with advanced features
type AdvancedResolutionOptions struct {
	*DependencyResolutionOptions

	// ConflictStrategy determines how to handle version conflicts
	ConflictStrategy ConflictResolutionStrategy

	// OptimizationStrategy determines which optimization to apply
	OptimizationStrategy OptimizationStrategy

	// MaxConflictRetries is the maximum number of times to retry resolving conflicts
	MaxConflictRetries int

	// ParallelWorkers is the number of parallel workers for resolution
	ParallelWorkers int

	// PreferredVersions maps module names to preferred versions
	PreferredVersions map[string]string

	// LockedVersions maps module names to locked versions (cannot be changed)
	LockedVersions map[string]string

	// ExcludedVersions maps module names to sets of excluded versions
	ExcludedVersions map[string]map[string]bool
}

// ResolutionContext maintains state during resolution
type ResolutionContext struct {
	// VersionCandidates maps module names to available versions
	VersionCandidates map[string][]string

	// ConstraintGraph maps modules to their constraints
	ConstraintGraph map[string]map[string]*VersionConstraint

	// ResolvedVersions maps module names to resolved versions
	ResolvedVersions map[string]string

	// ConflictHistory tracks conflict resolution attempts
	ConflictHistory []ConflictResolutionAttempt

	// Metrics tracks performance metrics
	Metrics ResolutionMetrics
}

// ConflictResolutionAttempt records an attempt to resolve a conflict
type ConflictResolutionAttempt struct {
	Module      string
	Constraints map[string]*VersionConstraint
	Attempted   string
	Success     bool
	Reason      string
	Timestamp   time.Time
}

// ResolutionMetrics tracks performance metrics
type ResolutionMetrics struct {
	StartTime         time.Time
	EndTime           time.Time
	ModulesProcessed  int
	ConflictsResolved int
	ConflictsFailed   int
	CacheHits         int
	CacheMisses       int
	ParallelTasks     int
}

// advancedResolver implements advanced dependency resolution
type advancedResolver struct {
	baseResolver DependencyResolver
	cache        *DependencyCache
	mu           sync.Mutex
	// Store advanced options for when called through interface
	defaultAdvancedOptions *AdvancedResolutionOptions
}

// workItem represents a work item for parallel processing
type workItem struct {
	node   *DependencyNode
	parent *DependencyNode
}

// NewAdvancedResolver creates a new advanced dependency resolver
func NewAdvancedResolver(cacheDir string, cacheTTL int) (DependencyResolver, error) {
	baseResolver, err := NewDefaultResolver(cacheDir, cacheTTL)
	if err != nil {
		return nil, err
	}

	var cache *DependencyCache
	if cacheDir != "" && cacheTTL > 0 {
		cache, err = NewDependencyCache(cacheDir, cacheTTL)
		if err != nil {
			return nil, err
		}
	}

	return &advancedResolver{
		baseResolver: baseResolver,
		cache:        cache,
	}, nil
}

// SetAdvancedOptions allows configuring advanced options
func (r *advancedResolver) SetAdvancedOptions(opts *AdvancedResolutionOptions) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultAdvancedOptions = opts
}

// ResolveDependenciesAdvanced performs dependency resolution with advanced options
func (r *advancedResolver) ResolveDependenciesAdvanced(ctx context.Context, moduleName string, options *AdvancedResolutionOptions) (*DependencyResolutionResult, error) {
	// Temporarily set advanced options for this call
	r.SetAdvancedOptions(options)
	return r.ResolveDependencies(ctx, moduleName, options.DependencyResolutionOptions)
}

// ResolveDependencies performs advanced dependency resolution
func (r *advancedResolver) ResolveDependencies(ctx context.Context, moduleName string, options *DependencyResolutionOptions) (*DependencyResolutionResult, error) {
	// Use advanced options if available, otherwise create defaults
	var advOptions *AdvancedResolutionOptions
	if r.defaultAdvancedOptions != nil {
		advOptions = r.defaultAdvancedOptions
	} else {
		// Create advanced options with defaults
		advOptions = &AdvancedResolutionOptions{
			DependencyResolutionOptions: options,
			ConflictStrategy:            StrategyLatestCompatible,
			OptimizationStrategy:        OptimizeSharedDependencies,
			MaxConflictRetries:          3,
			ParallelWorkers:             4,
		}
	}

	// Initialize resolution context
	resCtx := &ResolutionContext{
		VersionCandidates: make(map[string][]string),
		ConstraintGraph:   make(map[string]map[string]*VersionConstraint),
		ResolvedVersions:  make(map[string]string),
		ConflictHistory:   []ConflictResolutionAttempt{},
		Metrics: ResolutionMetrics{
			StartTime: time.Now(),
		},
	}

	// Apply optimization strategy
	var result *DependencyResolutionResult
	var err error

	switch advOptions.OptimizationStrategy {
	case OptimizeParallel:
		result, err = r.resolveParallel(ctx, moduleName, advOptions, resCtx)
	case OptimizeMinimalTree:
		result, err = r.resolveMinimal(ctx, moduleName, advOptions, resCtx)
	case OptimizeSharedDependencies:
		result, err = r.resolveShared(ctx, moduleName, advOptions, resCtx)
	default:
		// Fall back to base resolver
		result, err = r.baseResolver.ResolveDependencies(ctx, moduleName, options)
	}

	// Handle conflicts if any
	if result != nil && len(result.Conflicts) > 0 {
		result, err = r.resolveConflicts(ctx, result, advOptions, resCtx)
	}

	// Update metrics
	resCtx.Metrics.EndTime = time.Now()
	if result != nil {
		resCtx.Metrics.ModulesProcessed = len(result.Modules)
		resCtx.Metrics.ConflictsFailed = len(result.Conflicts)
	}

	// Log metrics if verbose
	if advOptions.Verbose {
		log.Infof("Resolution completed in %v: %d modules, %d conflicts resolved, %d failed",
			resCtx.Metrics.EndTime.Sub(resCtx.Metrics.StartTime),
			resCtx.Metrics.ModulesProcessed,
			resCtx.Metrics.ConflictsResolved,
			resCtx.Metrics.ConflictsFailed)
	}

	return result, err
}

// resolveParallel performs parallel dependency resolution
func (r *advancedResolver) resolveParallel(ctx context.Context, moduleName string, options *AdvancedResolutionOptions, resCtx *ResolutionContext) (*DependencyResolutionResult, error) {
	// Create channels for work distribution
	workChan := make(chan workItem, 100)
	resultChan := make(chan *DependencyNode, 100)
	errorChan := make(chan error, 1)

	// Create wait group for workers
	var wg sync.WaitGroup

	// Start workers
	numWorkers := options.ParallelWorkers
	if numWorkers <= 0 {
		numWorkers = 4
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range workChan {
				// Process dependency
				if err := r.processNodeParallel(ctx, item.node, item.parent, options, resCtx, workChan, resultChan); err != nil {
					select {
					case errorChan <- err:
					default:
					}
					return
				}
			}
		}()
	}

	// Create result structure
	result := &DependencyResolutionResult{
		Modules:            make(map[string]*DependencyNode),
		VersionConstraints: make(map[string]map[string]bool),
		Conflicts:          []*DependencyConflict{},
		Warnings:           []string{},
	}

	// Get root module info
	moduleInfo, err := options.Provider.GetModuleInfo(ctx, moduleName)
	if err != nil {
		return nil, errors.NewSystemError(
			ErrModuleNotFound,
			fmt.Sprintf("Failed to get information for module %s", moduleName),
			err)
	}

	// Create root node
	root := &DependencyNode{
		Name:    moduleInfo.Name,
		Version: moduleInfo.Version,
		IsRoot:  true,
		Path:    []string{moduleInfo.Name},
		Depth:   0,
	}

	result.Root = root
	result.Modules[root.Name] = root

	// Start processing from root
	workChan <- workItem{node: root, parent: nil}

	// Collect results
	go func() {
		wg.Wait()
		close(workChan)
		close(resultChan)
	}()

	// Process results
	for node := range resultChan {
		r.mu.Lock()
		result.Modules[node.Name] = node
		r.mu.Unlock()
	}

	// Check for errors
	select {
	case err := <-errorChan:
		return result, err
	default:
	}

	resCtx.Metrics.ParallelTasks = numWorkers
	return result, nil
}

// processNodeParallel processes a node in parallel mode
func (r *advancedResolver) processNodeParallel(ctx context.Context, node *DependencyNode, parent *DependencyNode, options *AdvancedResolutionOptions, resCtx *ResolutionContext, workChan chan<- workItem, resultChan chan<- *DependencyNode) error {
	// Get module info
	moduleInfo, err := options.Provider.GetModuleInfo(ctx, node.Name)
	if err != nil {
		return errors.NewSystemError(
			ErrModuleNotFound,
			fmt.Sprintf("Failed to get information for module %s", node.Name),
			err)
	}

	// Process dependencies
	for _, dep := range moduleInfo.Dependencies {
		if !shouldIncludeDependency(dep, options.DependencyResolutionOptions) {
			continue
		}

		// Check if already processed
		r.mu.Lock()
		_, exists := resCtx.ResolvedVersions[dep.Name]
		r.mu.Unlock()

		if exists {
			continue
		}

		// Create dependency node
		depNode := &DependencyNode{
			Name:              dep.Name,
			VersionConstraint: dep.Version,
			Phase:             dep.Phase,
			Type:              dep.Type,
			Parent:            node,
			Path:              append(append([]string{}, node.Path...), dep.Name),
			Depth:             node.Depth + 1,
		}

		// Add to work queue
		workChan <- workItem{node: depNode, parent: node}

		// Send to results
		resultChan <- depNode

		// Record as processed
		r.mu.Lock()
		resCtx.ResolvedVersions[dep.Name] = "" // Placeholder
		r.mu.Unlock()
	}

	return nil
}

// resolveMinimal performs minimal tree resolution
func (r *advancedResolver) resolveMinimal(ctx context.Context, moduleName string, options *AdvancedResolutionOptions, resCtx *ResolutionContext) (*DependencyResolutionResult, error) {
	// First get the full dependency tree
	result, err := r.baseResolver.ResolveDependencies(ctx, moduleName, options.DependencyResolutionOptions)
	if err != nil {
		return nil, err
	}

	// Build constraint graph
	r.buildConstraintGraph(result, resCtx)

	// Find minimal set of versions that satisfy all constraints
	minimalVersions := r.findMinimalVersionSet(ctx, resCtx, options)

	// Rebuild tree with minimal versions
	return r.rebuildWithVersions(ctx, result, minimalVersions, options)
}

// resolveShared performs shared dependency optimization
func (r *advancedResolver) resolveShared(ctx context.Context, moduleName string, options *AdvancedResolutionOptions, resCtx *ResolutionContext) (*DependencyResolutionResult, error) {
	// First get the full dependency tree
	result, err := r.baseResolver.ResolveDependencies(ctx, moduleName, options.DependencyResolutionOptions)
	if err != nil {
		return nil, err
	}

	// Build constraint graph from collected constraints
	r.buildConstraintGraph(result, resCtx)

	// Analyze dependency sharing opportunities
	sharingOpportunities := r.analyzeSharingOpportunities(result)

	// Optimize for maximum sharing
	optimizedVersions := r.optimizeForSharing(ctx, sharingOpportunities, resCtx, options)

	// Also check for excluded versions on non-shared dependencies
	if len(options.ExcludedVersions) > 0 {
		excludedOptimized := r.optimizeExcludedVersions(ctx, result, resCtx, options)
		// Merge the results
		for module, version := range excludedOptimized {
			if _, exists := optimizedVersions[module]; !exists {
				optimizedVersions[module] = version
			}
		}
	}

	// Rebuild tree with optimized versions
	return r.rebuildWithVersions(ctx, result, optimizedVersions, options)
}

// resolveConflicts attempts to resolve version conflicts
func (r *advancedResolver) resolveConflicts(ctx context.Context, result *DependencyResolutionResult, options *AdvancedResolutionOptions, resCtx *ResolutionContext) (*DependencyResolutionResult, error) {
	if len(result.Conflicts) == 0 {
		return result, nil
	}

	switch options.ConflictStrategy {
	case StrategyFailFast:
		// Return immediately with conflicts
		return result, errors.NewSystemError(
			ErrVersionConflict,
			fmt.Sprintf("Version conflicts detected for %d modules", len(result.Conflicts)),
			nil)

	case StrategyLatestCompatible:
		return r.resolveLatestCompatible(ctx, result, options, resCtx)

	case StrategyMinimalVersion:
		return r.resolveMinimalVersion(ctx, result, options, resCtx)

	case StrategyPreferExisting:
		return r.resolvePreferExisting(ctx, result, options, resCtx)

	default:
		return result, nil
	}
}

// resolveLatestCompatible finds the latest version that satisfies all constraints
func (r *advancedResolver) resolveLatestCompatible(ctx context.Context, result *DependencyResolutionResult, options *AdvancedResolutionOptions, resCtx *ResolutionContext) (*DependencyResolutionResult, error) {
	resolvedConflicts := 0

	for _, conflict := range result.Conflicts {
		// Get all available versions for the module
		versions, err := options.Provider.GetModuleVersions(ctx, conflict.Module)
		if err != nil {
			log.Warnf("Failed to get available versions for %s: %v", conflict.Module, err)
			continue
		}

		// Sort versions in descending order (latest first)
		sort.Sort(sort.Reverse(VersionSlice(versions)))

		// Find the latest version that satisfies all constraints
		var compatibleVersion string
		for _, version := range versions {
			// Check if version is excluded
			if excluded, ok := options.ExcludedVersions[conflict.Module]; ok && excluded[version] {
				continue
			}

			// Check if version satisfies all constraints
			allSatisfied := true
			for constraint := range conflict.Requirements {
				satisfied, err := r.CheckVersionConstraint(version, constraint)
				if err != nil || !satisfied {
					allSatisfied = false
					break
				}
			}

			if allSatisfied {
				compatibleVersion = version
				break
			}
		}

		if compatibleVersion != "" {
			// Update the resolved version
			if node, ok := result.Modules[conflict.Module]; ok {
				node.Version = compatibleVersion
			}
			resCtx.ResolvedVersions[conflict.Module] = compatibleVersion
			resolvedConflicts++

			// Record successful resolution
			resCtx.ConflictHistory = append(resCtx.ConflictHistory, ConflictResolutionAttempt{
				Module:      conflict.Module,
				Constraints: r.extractConstraints(conflict),
				Attempted:   compatibleVersion,
				Success:     true,
				Reason:      "Found latest compatible version",
				Timestamp:   time.Now(),
			})
		} else {
			// Record failed resolution
			resCtx.ConflictHistory = append(resCtx.ConflictHistory, ConflictResolutionAttempt{
				Module:      conflict.Module,
				Constraints: r.extractConstraints(conflict),
				Attempted:   "",
				Success:     false,
				Reason:      "No compatible version found",
				Timestamp:   time.Now(),
			})
		}
	}

	// Remove resolved conflicts from the result
	if resolvedConflicts > 0 {
		newConflicts := []*DependencyConflict{}
		for _, conflict := range result.Conflicts {
			if _, resolved := resCtx.ResolvedVersions[conflict.Module]; !resolved {
				newConflicts = append(newConflicts, conflict)
			}
		}
		result.Conflicts = newConflicts
		resCtx.Metrics.ConflictsResolved = resolvedConflicts
	}

	return result, nil
}

// Helper methods

// buildConstraintGraph builds a graph of version constraints
func (r *advancedResolver) buildConstraintGraph(result *DependencyResolutionResult, resCtx *ResolutionContext) {
	for moduleName, constraints := range result.VersionConstraints {
		if _, ok := resCtx.ConstraintGraph[moduleName]; !ok {
			resCtx.ConstraintGraph[moduleName] = make(map[string]*VersionConstraint)
		}

		for constraint := range constraints {
			parsed, err := ParseVersionConstraint(constraint)
			if err == nil {
				resCtx.ConstraintGraph[moduleName][constraint] = parsed
			}
		}
	}
}

// findMinimalVersionSet finds the minimal set of versions that satisfy all constraints
func (r *advancedResolver) findMinimalVersionSet(ctx context.Context, resCtx *ResolutionContext, options *AdvancedResolutionOptions) map[string]string {
	minimalVersions := make(map[string]string)

	for moduleName, constraints := range resCtx.ConstraintGraph {
		// Get available versions
		versions, err := options.Provider.GetModuleVersions(ctx, moduleName)
		if err != nil {
			continue
		}

		// Sort versions in ascending order (minimal first)
		sort.Sort(VersionSlice(versions))

		// Find minimal version that satisfies all constraints
		for _, version := range versions {
			allSatisfied := true
			for _, constraint := range constraints {
				satisfied := r.checkConstraintsSatisfied(version, constraint)
				if !satisfied {
					allSatisfied = false
					break
				}
			}

			if allSatisfied {
				minimalVersions[moduleName] = version
				break
			}
		}
	}

	return minimalVersions
}

// checkConstraintsSatisfied checks if a version satisfies a parsed constraint
func (r *advancedResolver) checkConstraintsSatisfied(version string, constraint *VersionConstraint) bool {
	for _, c := range constraint.Constraints {
		satisfied, err := checkSingleConstraint(version, c)
		if err != nil || !satisfied {
			return false
		}
	}
	return true
}

// analyzeSharingOpportunities analyzes where dependencies can be shared
func (r *advancedResolver) analyzeSharingOpportunities(result *DependencyResolutionResult) map[string][]string {
	opportunities := make(map[string][]string)

	// Find modules that appear multiple times in the tree
	for _, node := range result.Modules {
		if len(node.Children) > 0 {
			for _, child := range node.Children {
				opportunities[child.Name] = append(opportunities[child.Name], node.Name)
			}
		}
	}

	// Filter to only those with multiple parents
	filtered := make(map[string][]string)
	for module, parents := range opportunities {
		if len(parents) > 1 {
			filtered[module] = parents
		}
	}

	return filtered
}

// optimizeForSharing finds versions that maximize dependency sharing
func (r *advancedResolver) optimizeForSharing(ctx context.Context, opportunities map[string][]string, resCtx *ResolutionContext, options *AdvancedResolutionOptions) map[string]string {
	optimized := make(map[string]string)

	// For each shared dependency, find a version that works for all parents
	for module, parents := range opportunities {
		// Collect all constraints from parents
		allConstraints := make([]*VersionConstraint, 0)
		// Note: In a real implementation, we would check constraints specific to each parent
		// For now, we just collect all constraints for the module
		_ = parents // Mark as intentionally unused
		if constraints, ok := resCtx.ConstraintGraph[module]; ok {
			for _, constraint := range constraints {
				allConstraints = append(allConstraints, constraint)
			}
		}

		// Find a version that satisfies all constraints
		versions, err := options.Provider.GetModuleVersions(ctx, module)
		if err != nil {
			continue
		}

		// Sort versions based on conflict strategy
		switch options.ConflictStrategy {
		case StrategyMinimalVersion:
			sort.Sort(VersionSlice(versions)) // Ascending order (minimal first)
		default:
			sort.Sort(sort.Reverse(VersionSlice(versions))) // Descending order (latest first)
		}

		for _, version := range versions {
			// Check if version is excluded
			if excluded, ok := options.ExcludedVersions[module]; ok && excluded[version] {
				continue
			}

			allSatisfied := true
			for _, constraint := range allConstraints {
				if !r.checkConstraintsSatisfied(version, constraint) {
					allSatisfied = false
					break
				}
			}

			if allSatisfied {
				optimized[module] = version
				break
			}
		}
	}

	return optimized
}

// optimizeExcludedVersions finds non-excluded versions for all modules with excluded versions
func (r *advancedResolver) optimizeExcludedVersions(ctx context.Context, result *DependencyResolutionResult, resCtx *ResolutionContext, options *AdvancedResolutionOptions) map[string]string {
	optimized := make(map[string]string)

	// Check each module that has excluded versions
	for moduleName, excludedVersions := range options.ExcludedVersions {
		if _, exists := result.Modules[moduleName]; !exists {
			continue // Module not in dependency tree
		}

		// Get available versions
		versions, err := options.Provider.GetModuleVersions(ctx, moduleName)
		if err != nil {
			continue
		}

		// Get constraints for this module
		var allConstraints []*VersionConstraint
		if constraints, ok := resCtx.ConstraintGraph[moduleName]; ok {
			for _, constraint := range constraints {
				allConstraints = append(allConstraints, constraint)
			}
		}

		// Sort based on conflict strategy (same logic as optimizeForSharing)
		switch options.ConflictStrategy {
		case StrategyMinimalVersion:
			sort.Sort(VersionSlice(versions)) // Ascending order (minimal first)
		default:
			sort.Sort(sort.Reverse(VersionSlice(versions))) // Descending order (latest first)
		}

		// Find best non-excluded version
		for _, version := range versions {
			// Skip excluded versions
			if excludedVersions[version] {
				continue
			}

			// Check if version satisfies all constraints
			allSatisfied := true
			for _, constraint := range allConstraints {
				if !r.checkConstraintsSatisfied(version, constraint) {
					allSatisfied = false
					break
				}
			}

			if allSatisfied {
				optimized[moduleName] = version
				break
			}
		}
	}

	return optimized
}

// rebuildWithVersions rebuilds the dependency tree with specific versions
func (r *advancedResolver) rebuildWithVersions(ctx context.Context, original *DependencyResolutionResult, versions map[string]string, options *AdvancedResolutionOptions) (*DependencyResolutionResult, error) {
	// Update versions in the original result
	for moduleName, version := range versions {
		if node, ok := original.Modules[moduleName]; ok {
			node.Version = version
		}
	}

	// Clear conflicts that have been resolved
	newConflicts := []*DependencyConflict{}
	for _, conflict := range original.Conflicts {
		if _, resolved := versions[conflict.Module]; !resolved {
			newConflicts = append(newConflicts, conflict)
		}
	}
	original.Conflicts = newConflicts

	return original, nil
}

// extractConstraints extracts constraints from a conflict
func (r *advancedResolver) extractConstraints(conflict *DependencyConflict) map[string]*VersionConstraint {
	constraints := make(map[string]*VersionConstraint)

	for constraint := range conflict.Requirements {
		parsed, err := ParseVersionConstraint(constraint)
		if err == nil {
			constraints[constraint] = parsed
		}
	}

	return constraints
}

// resolveMinimalVersion finds the minimal version that satisfies all constraints
func (r *advancedResolver) resolveMinimalVersion(ctx context.Context, result *DependencyResolutionResult, options *AdvancedResolutionOptions, resCtx *ResolutionContext) (*DependencyResolutionResult, error) {
	// Similar to resolveLatestCompatible but sorts versions in ascending order
	resolvedConflicts := 0

	for _, conflict := range result.Conflicts {
		versions, err := options.Provider.GetModuleVersions(ctx, conflict.Module)
		if err != nil {
			continue
		}

		// Sort versions in ascending order (minimal first)
		sort.Sort(VersionSlice(versions))

		// Find the minimal version that satisfies all constraints
		var compatibleVersion string
		for _, version := range versions {
			allSatisfied := true
			for constraint := range conflict.Requirements {
				satisfied, err := r.CheckVersionConstraint(version, constraint)
				if err != nil || !satisfied {
					allSatisfied = false
					break
				}
			}

			if allSatisfied {
				compatibleVersion = version
				break
			}
		}

		if compatibleVersion != "" {
			if node, ok := result.Modules[conflict.Module]; ok {
				node.Version = compatibleVersion
			}
			resCtx.ResolvedVersions[conflict.Module] = compatibleVersion
			resolvedConflicts++
		}
	}

	// Update conflicts
	if resolvedConflicts > 0 {
		newConflicts := []*DependencyConflict{}
		for _, conflict := range result.Conflicts {
			if _, resolved := resCtx.ResolvedVersions[conflict.Module]; !resolved {
				newConflicts = append(newConflicts, conflict)
			}
		}
		result.Conflicts = newConflicts
		resCtx.Metrics.ConflictsResolved = resolvedConflicts
	}

	return result, nil
}

// resolvePreferExisting prefers already-resolved versions when possible
func (r *advancedResolver) resolvePreferExisting(ctx context.Context, result *DependencyResolutionResult, options *AdvancedResolutionOptions, resCtx *ResolutionContext) (*DependencyResolutionResult, error) {
	// Check if preferred or locked versions can resolve conflicts
	resolvedConflicts := 0

	for _, conflict := range result.Conflicts {
		var candidateVersion string

		// First check locked versions
		if locked, ok := options.LockedVersions[conflict.Module]; ok {
			candidateVersion = locked
		} else if preferred, ok := options.PreferredVersions[conflict.Module]; ok {
			candidateVersion = preferred
		} else if existing, ok := resCtx.ResolvedVersions[conflict.Module]; ok && existing != "" {
			candidateVersion = existing
		}

		if candidateVersion != "" {
			// Check if candidate satisfies all constraints
			allSatisfied := true
			for constraint := range conflict.Requirements {
				satisfied, err := r.CheckVersionConstraint(candidateVersion, constraint)
				if err != nil || !satisfied {
					allSatisfied = false
					break
				}
			}

			if allSatisfied {
				if node, ok := result.Modules[conflict.Module]; ok {
					node.Version = candidateVersion
				}
				resCtx.ResolvedVersions[conflict.Module] = candidateVersion
				resolvedConflicts++
			}
		}
	}

	// If some conflicts remain, fall back to latest compatible
	if resolvedConflicts < len(result.Conflicts) {
		return r.resolveLatestCompatible(ctx, result, options, resCtx)
	}

	return result, nil
}

// Implement remaining interface methods

func (r *advancedResolver) CheckVersionConstraint(version, constraint string) (bool, error) {
	return r.baseResolver.CheckVersionConstraint(version, constraint)
}

func (r *advancedResolver) GetFlattenedDependencies(result *DependencyResolutionResult) []*DependencyNode {
	return r.baseResolver.GetFlattenedDependencies(result)
}

func (r *advancedResolver) PrintDependencyTree(node *DependencyNode) string {
	return r.baseResolver.PrintDependencyTree(node)
}

// VersionSlice implements sort.Interface for version strings
type VersionSlice []string

func (v VersionSlice) Len() int      { return len(v) }
func (v VersionSlice) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v VersionSlice) Less(i, j int) bool {
	return compareVersions(v[i], v[j]) < 0
}
