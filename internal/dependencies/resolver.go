// ABOUTME: This file provides comprehensive dependency resolution functionality for PVM.
// ABOUTME: It consolidates the best features from PVI's dependency resolver into a reusable package.

package dependencies

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"tamarou.com/pvm/internal/cli/progress"
	"tamarou.com/pvm/internal/cpan"
)

// DependencyResolver provides dependency resolution and conflict detection
type DependencyResolver struct {
	provider cpan.Provider
	tracker  progress.Tracker
	logger   *log.Logger
	cache    *dependencyCache
	options  ResolverOptions
}

// ResolverOptions configures dependency resolution behavior
type ResolverOptions struct {
	// CacheEnabled enables dependency caching for improved performance
	CacheEnabled bool
	// CacheTTL specifies how long cached results remain valid
	CacheTTL time.Duration
	// MaxDepth limits recursion depth to prevent infinite loops
	MaxDepth int
	// ConflictStrategy determines how version conflicts are resolved
	ConflictStrategy ConflictStrategy
	// ParallelEnabled enables parallel dependency resolution
	ParallelEnabled bool
	// MaxWorkers limits parallel workers for dependency resolution
	MaxWorkers int
}

// ConflictStrategy defines how version conflicts should be resolved
type ConflictStrategy int

const (
	// ConflictStrategyFailFast fails immediately on version conflicts
	ConflictStrategyFailFast ConflictStrategy = iota
	// ConflictStrategyLatestCompatible chooses latest compatible versions
	ConflictStrategyLatestCompatible
	// ConflictStrategyMinimalVersion chooses minimal versions satisfying constraints
	ConflictStrategyMinimalVersion
	// ConflictStrategyPreferExisting prefers already installed versions when possible
	ConflictStrategyPreferExisting
)

// NewDependencyResolver creates a new dependency resolver with the given provider
func NewDependencyResolver(provider cpan.Provider, tracker progress.Tracker, logger *log.Logger) *DependencyResolver {
	return &DependencyResolver{
		provider: provider,
		tracker:  tracker,
		logger:   logger,
		cache:    newDependencyCache(),
		options: ResolverOptions{
			CacheEnabled:     true,
			CacheTTL:         24 * time.Hour,
			MaxDepth:         50,
			ConflictStrategy: ConflictStrategyLatestCompatible,
			ParallelEnabled:  true,
			MaxWorkers:       4,
		},
	}
}

// WithOptions configures the resolver with custom options
func (dr *DependencyResolver) WithOptions(options ResolverOptions) *DependencyResolver {
	dr.options = options
	return dr
}

// ResolveDependencies resolves dependencies for the given modules and returns a dependency graph
func (dr *DependencyResolver) ResolveDependencies(ctx context.Context, modules []string) (*DependencyGraph, error) {
	if len(modules) == 0 {
		return &DependencyGraph{Nodes: make(map[string]*DependencyNode)}, nil
	}

	if dr.tracker != nil {
		dr.tracker.Start(fmt.Sprintf("Resolving dependencies for %d modules", len(modules)), 1)
		defer func() {
			if dr.tracker.IsRunning() {
				result := &progress.Result{
					Operation: "resolve_dependencies",
					Target:    fmt.Sprintf("%d modules", len(modules)),
					Success:   false,
				}
				dr.tracker.Finish(result)
			}
		}()
	}

	startTime := time.Now()

	graph := &DependencyGraph{
		Nodes:     make(map[string]*DependencyNode),
		RootNodes: make([]*DependencyNode, 0, len(modules)),
	}

	// Create resolution context
	resCtx := &resolutionContext{
		graph:       graph,
		visited:     make(map[string]bool),
		resolving:   make(map[string]bool),
		constraints: make(map[string][]VersionConstraint),
		depth:       0,
	}

	// Resolve each root module
	for _, module := range modules {
		node, err := dr.resolveModuleDependencies(ctx, module, "", resCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependencies for %s: %w", module, err)
		}
		graph.RootNodes = append(graph.RootNodes, node)
	}

	// Detect and resolve conflicts
	conflicts := dr.detectConflicts(graph)
	if len(conflicts) > 0 {
		resolvedGraph, err := dr.resolveConflicts(ctx, graph, conflicts)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve conflicts: %w", err)
		}

		// Update progress tracker with success
		duration := time.Since(startTime)
		if dr.tracker != nil && dr.tracker.IsRunning() {
			progressResult := &progress.Result{
				Operation: "resolve_dependencies",
				Target:    fmt.Sprintf("%d modules", len(modules)),
				Success:   true,
				Duration:  duration,
				Message:   fmt.Sprintf("Resolved dependencies for %d modules with %d conflicts resolved", len(modules), len(conflicts)),
			}
			dr.tracker.Finish(progressResult)
		}

		return resolvedGraph, nil
	}

	// Update progress tracker with success
	duration := time.Since(startTime)
	if dr.tracker != nil && dr.tracker.IsRunning() {
		progressResult := &progress.Result{
			Operation: "resolve_dependencies",
			Target:    fmt.Sprintf("%d modules", len(modules)),
			Success:   true,
			Duration:  duration,
			Message:   fmt.Sprintf("Resolved dependencies for %d modules (%d total nodes)", len(modules), len(graph.Nodes)),
		}
		dr.tracker.Finish(progressResult)
	}

	return graph, nil
}

// DetectConflicts identifies version conflicts in the dependency graph
func (dr *DependencyResolver) DetectConflicts(graph *DependencyGraph) ([]*Conflict, error) {
	return dr.detectConflicts(graph), nil
}

// SuggestResolutions provides suggestions for resolving conflicts
func (dr *DependencyResolver) SuggestResolutions(conflicts []*Conflict) ([]*Resolution, error) {
	resolutions := make([]*Resolution, 0, len(conflicts))

	for _, conflict := range conflicts {
		resolution := &Resolution{
			Module:    conflict.Module,
			Conflict:  conflict,
			Suggested: make([]*ResolutionOption, 0),
		}

		// Generate resolution options based on strategy
		switch dr.options.ConflictStrategy {
		case ConflictStrategyLatestCompatible:
			resolution.Suggested = dr.generateLatestCompatibleOptions(conflict)
		case ConflictStrategyMinimalVersion:
			resolution.Suggested = dr.generateMinimalVersionOptions(conflict)
		case ConflictStrategyPreferExisting:
			resolution.Suggested = dr.generatePreferExistingOptions(conflict)
		default:
			resolution.Suggested = dr.generateLatestCompatibleOptions(conflict)
		}

		resolutions = append(resolutions, resolution)
	}

	return resolutions, nil
}

// CreateInstallPlan generates an optimized installation plan from the dependency graph
func (dr *DependencyResolver) CreateInstallPlan(graph *DependencyGraph) (*InstallPlan, error) {
	// Topologically sort modules for installation order
	sorted, err := dr.topologicalSort(graph)
	if err != nil {
		return nil, fmt.Errorf("failed to create install plan: %w", err)
	}

	plan := &InstallPlan{
		Modules:      make([]*InstallPlanModule, 0, len(sorted)),
		Dependencies: make(map[string][]string),
		Levels:       dr.calculateInstallLevels(sorted),
	}

	// Create install plan modules
	for _, node := range sorted {
		module := &InstallPlanModule{
			Name:         node.Name,
			Version:      node.Version,
			Dependencies: make([]string, 0, len(node.Dependencies)),
		}

		for _, dep := range node.Dependencies {
			module.Dependencies = append(module.Dependencies, dep.Name)
		}

		plan.Modules = append(plan.Modules, module)
		plan.Dependencies[node.Name] = module.Dependencies
	}

	return plan, nil
}

// resolveModuleDependencies recursively resolves dependencies for a single module
func (dr *DependencyResolver) resolveModuleDependencies(ctx context.Context, moduleName, version string, resCtx *resolutionContext) (*DependencyNode, error) {
	// Check depth limit
	if resCtx.depth > dr.options.MaxDepth {
		return nil, fmt.Errorf("maximum dependency depth exceeded for module %s", moduleName)
	}

	// Check for circular dependencies
	if resCtx.resolving[moduleName] {
		return nil, fmt.Errorf("circular dependency detected: %s", moduleName)
	}

	// Check if already resolved
	if node, exists := resCtx.graph.Nodes[moduleName]; exists {
		return node, nil
	}

	// Mark as resolving
	resCtx.resolving[moduleName] = true
	defer func() { delete(resCtx.resolving, moduleName) }()

	// Check cache first
	if dr.options.CacheEnabled {
		if cached := dr.cache.get(moduleName); cached != nil {
			resCtx.graph.Nodes[moduleName] = cached
			return cached, nil
		}
	}

	// Get module metadata from provider
	moduleInfo, err := dr.provider.GetModuleInfo(ctx, moduleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get module info for %s: %w", moduleName, err)
	}

	// Use provided version or latest available
	targetVersion := version
	if targetVersion == "" {
		targetVersion = moduleInfo.Version
	}

	// Create dependency node
	node := &DependencyNode{
		Name:         moduleName,
		Version:      targetVersion,
		Dependencies: make([]*DependencyNode, 0),
		Constraints:  make([]VersionConstraint, 0),
		Depth:        resCtx.depth,
	}

	// Add to graph
	resCtx.graph.Nodes[moduleName] = node

	// Resolve dependencies recursively
	resCtx.depth++
	for _, dep := range moduleInfo.Dependencies {
		depNode, err := dr.resolveModuleDependencies(ctx, dep.Name, dep.Version, resCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependency %s for %s: %w", dep.Name, moduleName, err)
		}

		node.Dependencies = append(node.Dependencies, depNode)

		// Track version constraints
		if dep.Version != "" {
			constraint := VersionConstraint{
				Module:     dep.Name,
				Constraint: dep.Version,
				Source:     moduleName,
			}
			node.Constraints = append(node.Constraints, constraint)
			resCtx.constraints[dep.Name] = append(resCtx.constraints[dep.Name], constraint)
		}
	}
	resCtx.depth--

	// Cache the result
	if dr.options.CacheEnabled {
		dr.cache.put(moduleName, node)
	}

	return node, nil
}

// detectConflicts identifies version conflicts in the dependency graph
func (dr *DependencyResolver) detectConflicts(graph *DependencyGraph) []*Conflict {
	conflicts := make([]*Conflict, 0)

	// Check for version conflicts
	moduleVersions := make(map[string]map[string][]*DependencyNode)

	for _, node := range graph.Nodes {
		if moduleVersions[node.Name] == nil {
			moduleVersions[node.Name] = make(map[string][]*DependencyNode)
		}
		moduleVersions[node.Name][node.Version] = append(moduleVersions[node.Name][node.Version], node)
	}

	// Find modules with multiple versions
	for moduleName, versions := range moduleVersions {
		if len(versions) > 1 {
			conflict := &Conflict{
				Module:       moduleName,
				Type:         ConflictTypeVersion,
				Versions:     make([]string, 0, len(versions)),
				Dependencies: make([]*ConflictDependency, 0),
			}

			for version, nodes := range versions {
				conflict.Versions = append(conflict.Versions, version)

				for _, node := range nodes {
					conflict.Dependencies = append(conflict.Dependencies, &ConflictDependency{
						Dependant: node.Name,
						Required:  version,
					})
				}
			}

			sort.Strings(conflict.Versions)
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts
}

// resolveConflicts attempts to resolve conflicts using the configured strategy
func (dr *DependencyResolver) resolveConflicts(ctx context.Context, graph *DependencyGraph, conflicts []*Conflict) (*DependencyGraph, error) {
	switch dr.options.ConflictStrategy {
	case ConflictStrategyFailFast:
		return nil, fmt.Errorf("version conflicts detected: %s", dr.formatConflicts(conflicts))
	case ConflictStrategyLatestCompatible:
		return dr.resolveWithLatestCompatible(ctx, graph, conflicts)
	case ConflictStrategyMinimalVersion:
		return dr.resolveWithMinimalVersion(ctx, graph, conflicts)
	case ConflictStrategyPreferExisting:
		return dr.resolveWithPreferExisting(ctx, graph, conflicts)
	default:
		return dr.resolveWithLatestCompatible(ctx, graph, conflicts)
	}
}

// resolveWithLatestCompatible resolves conflicts by choosing latest compatible versions
func (dr *DependencyResolver) resolveWithLatestCompatible(ctx context.Context, graph *DependencyGraph, conflicts []*Conflict) (*DependencyGraph, error) {
	for _, conflict := range conflicts {
		// Find the latest version among conflicting versions
		latestVersion := dr.findLatestVersion(conflict.Versions)

		// Update all nodes to use the latest version
		for _, node := range graph.Nodes {
			if node.Name == conflict.Module {
				node.Version = latestVersion
			}
		}
	}

	return graph, nil
}

// resolveWithMinimalVersion resolves conflicts by choosing minimal satisfying versions
func (dr *DependencyResolver) resolveWithMinimalVersion(ctx context.Context, graph *DependencyGraph, conflicts []*Conflict) (*DependencyGraph, error) {
	for _, conflict := range conflicts {
		// Find the minimal version that satisfies all constraints
		minimalVersion := dr.findMinimalVersion(conflict.Versions)

		// Update all nodes to use the minimal version
		for _, node := range graph.Nodes {
			if node.Name == conflict.Module {
				node.Version = minimalVersion
			}
		}
	}

	return graph, nil
}

// resolveWithPreferExisting resolves conflicts by preferring existing installations
func (dr *DependencyResolver) resolveWithPreferExisting(ctx context.Context, graph *DependencyGraph, conflicts []*Conflict) (*DependencyGraph, error) {
	// For now, use latest compatible strategy
	// In the future, this could check for existing installations
	return dr.resolveWithLatestCompatible(ctx, graph, conflicts)
}

// topologicalSort performs topological sorting of the dependency graph
func (dr *DependencyResolver) topologicalSort(graph *DependencyGraph) ([]*DependencyNode, error) {
	visited := make(map[string]bool)
	temp := make(map[string]bool)
	result := make([]*DependencyNode, 0, len(graph.Nodes))

	var visit func(*DependencyNode) error
	visit = func(node *DependencyNode) error {
		if temp[node.Name] {
			return fmt.Errorf("circular dependency detected involving %s", node.Name)
		}
		if visited[node.Name] {
			return nil
		}

		temp[node.Name] = true

		for _, dep := range node.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}

		temp[node.Name] = false
		visited[node.Name] = true
		result = append(result, node)

		return nil
	}

	for _, node := range graph.Nodes {
		if !visited[node.Name] {
			if err := visit(node); err != nil {
				return nil, err
			}
		}
	}

	// No need to reverse - DFS post-order gives us dependencies before dependents

	return result, nil
}

// calculateInstallLevels determines the installation levels for parallel installation
func (dr *DependencyResolver) calculateInstallLevels(sorted []*DependencyNode) [][]string {
	levels := make([][]string, 0)
	nodeLevel := make(map[string]int)

	// Calculate level for each node
	for _, node := range sorted {
		maxDepLevel := -1
		for _, dep := range node.Dependencies {
			if depLevel, exists := nodeLevel[dep.Name]; exists {
				if depLevel > maxDepLevel {
					maxDepLevel = depLevel
				}
			}
		}

		level := maxDepLevel + 1
		nodeLevel[node.Name] = level

		// Extend levels slice if needed
		for len(levels) <= level {
			levels = append(levels, make([]string, 0))
		}

		levels[level] = append(levels[level], node.Name)
	}

	return levels
}

// generateLatestCompatibleOptions generates resolution options using latest compatible strategy
func (dr *DependencyResolver) generateLatestCompatibleOptions(conflict *Conflict) []*ResolutionOption {
	options := make([]*ResolutionOption, 0, len(conflict.Versions))

	for _, version := range conflict.Versions {
		options = append(options, &ResolutionOption{
			Version:     version,
			Description: fmt.Sprintf("Use version %s", version),
			Impact:      dr.calculateImpact(conflict, version),
		})
	}

	// Sort by version (latest first)
	sort.Slice(options, func(i, j int) bool {
		return dr.compareVersions(options[i].Version, options[j].Version) > 0
	})

	return options
}

// generateMinimalVersionOptions generates resolution options using minimal version strategy
func (dr *DependencyResolver) generateMinimalVersionOptions(conflict *Conflict) []*ResolutionOption {
	options := dr.generateLatestCompatibleOptions(conflict)

	// Sort by version (earliest first)
	sort.Slice(options, func(i, j int) bool {
		return dr.compareVersions(options[i].Version, options[j].Version) < 0
	})

	return options
}

// generatePreferExistingOptions generates resolution options preferring existing installations
func (dr *DependencyResolver) generatePreferExistingOptions(conflict *Conflict) []*ResolutionOption {
	// For now, use latest compatible strategy
	// In the future, this could prioritize installed versions
	return dr.generateLatestCompatibleOptions(conflict)
}

// findLatestVersion finds the latest version among the given versions
func (dr *DependencyResolver) findLatestVersion(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	latest := versions[0]
	for _, version := range versions[1:] {
		if dr.compareVersions(version, latest) > 0 {
			latest = version
		}
	}

	return latest
}

// findMinimalVersion finds the minimal version among the given versions
func (dr *DependencyResolver) findMinimalVersion(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	minimal := versions[0]
	for _, version := range versions[1:] {
		if dr.compareVersions(version, minimal) < 0 {
			minimal = version
		}
	}

	return minimal
}

// compareVersions compares two version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func (dr *DependencyResolver) compareVersions(v1, v2 string) int {
	if v1 == v2 {
		return 0
	}

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			// Simple numeric parsing - handle non-numeric parts as 0
			if parts1[i] != "" {
				for _, r := range parts1[i] {
					if r >= '0' && r <= '9' {
						n1 = n1*10 + int(r-'0')
					} else {
						break
					}
				}
			}
		}
		if i < len(parts2) {
			if parts2[i] != "" {
				for _, r := range parts2[i] {
					if r >= '0' && r <= '9' {
						n2 = n2*10 + int(r-'0')
					} else {
						break
					}
				}
			}
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

// calculateImpact calculates the impact of choosing a specific version
func (dr *DependencyResolver) calculateImpact(conflict *Conflict, version string) ResolutionImpact {
	// This is a simplified impact calculation
	// In practice, this would analyze dependency chains and compatibility
	return ResolutionImpact{
		AffectedModules: len(conflict.Dependencies),
		BreakingChanges: false, // Would need more sophisticated analysis
	}
}

// formatConflicts formats conflicts for error messages
func (dr *DependencyResolver) formatConflicts(conflicts []*Conflict) string {
	if len(conflicts) == 0 {
		return "no conflicts"
	}

	parts := make([]string, 0, len(conflicts))
	for _, conflict := range conflicts {
		parts = append(parts, fmt.Sprintf("%s (%s)", conflict.Module, strings.Join(conflict.Versions, ", ")))
	}

	return strings.Join(parts, "; ")
}

// resolutionContext tracks state during dependency resolution
type resolutionContext struct {
	graph       *DependencyGraph
	visited     map[string]bool
	resolving   map[string]bool
	constraints map[string][]VersionConstraint
	depth       int
}
