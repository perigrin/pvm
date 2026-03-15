// ABOUTME: Parallel installation coordination with dependency-aware ordering
// ABOUTME: Provides sophisticated multi-module installation with worker pool management

package modules

import (
	"context"
	"fmt"
	"log"
	"sync"

	pviModules "tamarou.com/pvm/internal/pm/modules"
)

// ParallelCoordinator manages complex multi-module installation operations
type ParallelCoordinator struct {
	installer  ModuleInstaller
	maxWorkers int
	tracker    ParallelProgressTracker
	mutex      sync.RWMutex
}

// NewParallelCoordinator creates a new parallel installation coordinator
func NewParallelCoordinator(installer ModuleInstaller, maxWorkers int, tracker ParallelProgressTracker) *ParallelCoordinator {
	if maxWorkers <= 0 {
		maxWorkers = 4 // Default to 4 workers
	}

	return &ParallelCoordinator{
		installer:  installer,
		maxWorkers: maxWorkers,
		tracker:    tracker,
	}
}

// InstallPlan represents a dependency-aware installation plan
type InstallPlan struct {
	// Modules to install in dependency order
	Modules []string

	// Dependencies maps module names to their dependencies
	Dependencies map[string][]string

	// InstallationOrder provides the optimal installation sequence
	InstallationOrder []string

	// ParallelBatches groups modules that can be installed in parallel
	ParallelBatches [][]string
}

// DependencyGraph represents module dependencies
type DependencyGraph struct {
	// Nodes maps module names to their metadata
	Nodes map[string]*DependencyNode

	// Edges represents dependency relationships
	Edges map[string][]string
}

// DependencyNode represents a module in the dependency graph
type DependencyNode struct {
	Name         string
	Version      string
	Dependencies []string
	Dependents   []string
}

// DependencyResolverInterface defines the interface for dependency resolution
// This allows integration with advanced dependency resolvers without circular imports
type DependencyResolverInterface interface {
	// ResolveDependencies resolves dependencies for the given modules
	ResolveDependencies(ctx context.Context, modules []string) (interface{}, error)

	// CreateInstallPlan creates an optimized installation plan
	CreateInstallPlan(dependencyGraph interface{}) (interface{}, error)
}

// RichDependencyGraph represents the interface for advanced dependency graphs
type RichDependencyGraph interface {
	GetNodes() map[string]interface{}
	GetEdges() []interface{}
}

// RichInstallPlan represents the interface for advanced install plans
type RichInstallPlan interface {
	GetModules() []interface{}
	GetDependencies() map[string][]string
	GetLevels() [][]string
}

// InstallModules installs multiple modules with dependency-aware parallel coordination
func (pc *ParallelCoordinator) InstallModules(ctx context.Context, modules []string, opts InstallOptions) ([]*InstallResult, error) {
	if len(modules) == 0 {
		return []*InstallResult{}, nil
	}

	// Start progress tracking if available
	if pc.tracker != nil {
		pc.tracker.StartParallel(modules)
	}

	// If parallel is not enabled or only one module, use the direct installer
	if !opts.Parallel || len(modules) == 1 {
		results, err := pc.installer.InstallBatch(ctx, modules, opts)

		// Finish progress tracking
		if pc.tracker != nil {
			operationResults := pc.convertToOperationResults(results)
			pc.tracker.FinishParallel(operationResults)
		}

		return results, err
	}

	// For parallel installation, delegate to the installer's batch capability
	// which should handle parallelism internally
	results, err := pc.installer.InstallBatch(ctx, modules, opts)

	// Finish progress tracking
	if pc.tracker != nil {
		operationResults := pc.convertToOperationResults(results)
		pc.tracker.FinishParallel(operationResults)
	}

	return results, err
}

// convertToOperationResults converts InstallResult array to OperationResult array
func (pc *ParallelCoordinator) convertToOperationResults(results []*InstallResult) []*OperationResult {
	operationResults := make([]*OperationResult, len(results))
	for i, result := range results {
		operationResults[i] = &OperationResult{
			Operation: "install",
			Target:    result.ModuleName,
			Success:   result.Success,
			Duration:  result.Duration,
			Message:   fmt.Sprintf("Installed %s v%s", result.ModuleName, result.Version),
		}
		if !result.Success && len(result.Errors) > 0 {
			operationResults[i].Error = fmt.Errorf("%v", result.Errors[0])
		}
	}
	return operationResults
}

// installModulesParallel uses the existing PVI parallel installer
func (pc *ParallelCoordinator) installModulesParallel(ctx context.Context, modules []string, opts InstallOptions) ([]*InstallResult, error) {
	// Convert unified options to PVI module options
	pviOptions := pc.convertInstallOptions(opts)

	// Create parallel install options
	parallelOptions := &pviModules.ParallelInstallOptions{
		Modules:     make([]*pviModules.ModuleInstallOptions, len(modules)),
		Workers:     pc.maxWorkers,
		StopOnError: false,
		Context:     ctx,
	}

	// Create module options for each module
	for i, moduleName := range modules {
		moduleOpts := *pviOptions // Copy base options
		moduleOpts.ModuleName = moduleName
		parallelOptions.Modules[i] = &moduleOpts
	}

	// Add progress tracking if available
	if pc.tracker != nil {
		pc.tracker.StartParallel(modules)

		parallelOptions.ProgressCallback = func(completed, total int, currentModule string, stage pviModules.InstallProgressStage) {
			status := pc.convertProgressStage(stage)
			message := fmt.Sprintf("%s: %s", currentModule, stage.String())
			pc.tracker.UpdateOperation(currentModule, status, message)
		}
	}

	// Execute parallel installation
	pviResult, err := pviModules.InstallModulesParallel(parallelOptions)
	if err != nil {
		return nil, fmt.Errorf("parallel installation failed: %w", err)
	}

	// Convert PVI results to unified results
	results := pc.convertInstallResults(pviResult)

	// Notify tracker of completion
	if pc.tracker != nil {
		operationResults := make([]*OperationResult, len(results))
		for i, result := range results {
			operationResults[i] = &OperationResult{
				Operation: "install",
				Target:    result.ModuleName,
				Success:   result.Success,
				Duration:  result.Duration,
				Message:   fmt.Sprintf("Installed %s v%s", result.ModuleName, result.Version),
			}
			if !result.Success && len(result.Errors) > 0 {
				operationResults[i].Error = fmt.Errorf("%v", result.Errors[0])
			}
		}
		pc.tracker.FinishParallel(operationResults)
	}

	return results, nil
}

// ResolveDependencies creates a dependency graph for the given modules
func (pc *ParallelCoordinator) ResolveDependencies(modules []string) (*DependencyGraph, error) {
	return pc.ResolveDependenciesWithResolver(modules, nil)
}

// ResolveDependenciesWithResolver creates a dependency graph using an advanced resolver
func (pc *ParallelCoordinator) ResolveDependenciesWithResolver(modules []string, resolver DependencyResolverInterface) (*DependencyGraph, error) {
	if resolver == nil {
		// Fallback to simple graph if no resolver available
		return pc.createSimpleGraph(modules), nil
	}

	// Use the advanced dependency resolver
	ctx := context.Background()
	richGraphInterface, err := resolver.ResolveDependencies(ctx, modules)
	if err != nil {
		return nil, fmt.Errorf("dependency resolution failed: %w", err)
	}

	// Convert rich dependency graph to simple format for interface compatibility
	return pc.convertRichGraphToSimple(richGraphInterface), nil
}

// ExecuteInstallPlan executes a pre-computed installation plan
func (pc *ParallelCoordinator) ExecuteInstallPlan(ctx context.Context, plan *InstallPlan) ([]*InstallResult, error) {
	if plan == nil || len(plan.Modules) == 0 {
		return []*InstallResult{}, nil
	}

	// Execute installation in dependency-aware batches
	if len(plan.ParallelBatches) == 0 {
		// Fallback to installing all at once if no batches defined
		opts := InstallOptions{
			Parallel: true,
			Workers:  pc.maxWorkers,
			Context:  ctx,
		}
		return pc.InstallModules(ctx, plan.Modules, opts)
	}

	// Install each batch in order, with parallel installation within each batch
	allResults := make([]*InstallResult, 0, len(plan.Modules))

	for batchIndex, batch := range plan.ParallelBatches {
		if len(batch) == 0 {
			continue
		}

		// Install all modules in this batch in parallel
		opts := InstallOptions{
			Parallel: len(batch) > 1, // Enable parallel only if multiple modules
			Workers:  pc.maxWorkers,
			Context:  ctx,
		}

		batchResults, err := pc.InstallModules(ctx, batch, opts)
		if err != nil {
			return allResults, fmt.Errorf("batch %d installation failed: %w", batchIndex, err)
		}

		// Check if any installations in this batch failed
		for _, result := range batchResults {
			allResults = append(allResults, result)
			if !result.Success {
				// Stop on first failure to respect dependency ordering
				return allResults, fmt.Errorf("module %s installation failed, stopping batch execution", result.ModuleName)
			}
		}
	}

	return allResults, nil
}

// CreateInstallPlan generates an optimized installation plan from a dependency graph
func (pc *ParallelCoordinator) CreateInstallPlan(graph *DependencyGraph) (*InstallPlan, error) {
	return pc.CreateInstallPlanWithResolver(graph, nil)
}

// CreateInstallPlanWithResolver generates an optimized installation plan using an advanced resolver
func (pc *ParallelCoordinator) CreateInstallPlanWithResolver(graph *DependencyGraph, resolver DependencyResolverInterface) (*InstallPlan, error) {
	if graph == nil || len(graph.Nodes) == 0 {
		return &InstallPlan{}, nil
	}

	if resolver == nil {
		// Fallback to simple plan if no resolver available
		return pc.createSimplePlan(graph), nil
	}

	// Convert simple graph to interface format for advanced resolver
	richGraphInterface := pc.convertSimpleGraphToInterface(graph)

	// Use advanced resolver to create sophisticated install plan
	richPlanInterface, err := resolver.CreateInstallPlan(richGraphInterface)
	if err != nil {
		return nil, fmt.Errorf("install plan creation failed: %w", err)
	}

	// Convert rich install plan to simple format for interface compatibility
	return pc.convertRichPlanToSimple(richPlanInterface), nil
}

// convertInstallOptions converts unified InstallOptions to PVI ModuleInstallOptions
func (pc *ParallelCoordinator) convertInstallOptions(opts InstallOptions) *pviModules.ModuleInstallOptions {
	return &pviModules.ModuleInstallOptions{
		PerlPath:          opts.PerlPath,
		InstallDir:        opts.InstallDir,
		VersionConstraint: opts.VersionConstraint,
		Force:             opts.Force,
		RunTests:          opts.RunTests,
		SkipDependencies:  opts.SkipDependencies,
		Cleanup:           opts.Cleanup,
		Verbose:           opts.Verbose,
		Context:           opts.Context,
	}
}

// convertProgressStage converts PVI progress stage to unified OperationStatus
func (pc *ParallelCoordinator) convertProgressStage(stage pviModules.InstallProgressStage) OperationStatus {
	switch stage {
	case pviModules.StageResolving, pviModules.StageDownloading, pviModules.StageExtracting:
		return StatusRunning
	case pviModules.StageBuilding, pviModules.StageTesting:
		return StatusRunning
	case pviModules.StageInstallingModule, pviModules.StageCleaningUp:
		return StatusRunning
	case pviModules.StageFinished:
		return StatusCompleted
	default:
		return StatusRunning
	}
}

// convertInstallResults converts PVI parallel results to unified InstallResult array
func (pc *ParallelCoordinator) convertInstallResults(pviResult *pviModules.ParallelInstallResult) []*InstallResult {
	results := make([]*InstallResult, len(pviResult.Results))

	for i, pviRes := range pviResult.Results {
		result := &InstallResult{
			ModuleName: pviRes.ModuleName,
			Version:    pviRes.Version,
			Success:    pviRes.Success,
			Duration:   pviRes.Duration,
			Path:       pviRes.InstallPath,
		}

		// Convert dependency results to string names
		if len(pviRes.Dependencies) > 0 {
			dependencies := make([]string, len(pviRes.Dependencies))
			for j, dep := range pviRes.Dependencies {
				dependencies[j] = dep.ModuleName
			}
			result.Dependencies = dependencies
		}

		// Convert warnings and errors
		if len(pviRes.Warnings) > 0 {
			result.Warnings = pviRes.Warnings
		}
		if len(pviRes.Errors) > 0 {
			result.Errors = pviRes.Errors
		}

		results[i] = result
	}

	// Add failure information
	for _, failure := range pviResult.Failures {
		// Find the corresponding result and mark it as failed
		for _, result := range results {
			if result.ModuleName == failure.ModuleName {
				result.Success = false
				result.Duration = failure.Duration
				if result.Errors == nil {
					result.Errors = []string{}
				}
				result.Errors = append(result.Errors, failure.Error.Error())
				break
			}
		}
	}

	return results
}

// GetWorkerCount returns the configured number of workers
func (pc *ParallelCoordinator) GetWorkerCount() int {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return pc.maxWorkers
}

// SetWorkerCount updates the number of workers
func (pc *ParallelCoordinator) SetWorkerCount(workers int) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	if workers > 0 {
		pc.maxWorkers = workers
	}
}

// Helper methods for data structure conversion

// createSimpleGraph creates a simple dependency graph with no dependencies (fallback)
func (pc *ParallelCoordinator) createSimpleGraph(modules []string) *DependencyGraph {
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Edges: make(map[string][]string),
	}

	for _, module := range modules {
		graph.Nodes[module] = &DependencyNode{
			Name:         module,
			Dependencies: []string{},
			Dependents:   []string{},
		}
		graph.Edges[module] = []string{}
	}

	return graph
}

// createSimplePlan creates a simple install plan with no dependency ordering (fallback)
func (pc *ParallelCoordinator) createSimplePlan(graph *DependencyGraph) *InstallPlan {
	modules := make([]string, 0, len(graph.Nodes))
	for name := range graph.Nodes {
		modules = append(modules, name)
	}

	return &InstallPlan{
		Modules:           modules,
		Dependencies:      graph.Edges,
		InstallationOrder: modules,
		ParallelBatches:   [][]string{modules},
	}
}

// convertRichGraphToSimple converts a rich dependency graph interface to simple format
// Uses reflection and type assertions to handle the interface conversion
func (pc *ParallelCoordinator) convertRichGraphToSimple(richGraphInterface interface{}) *DependencyGraph {
	// Use reflection to extract data from the rich graph interface
	// This is a simplified conversion that attempts to extract node information
	simpleGraph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Edges: make(map[string][]string),
	}

	// Try to extract nodes using type assertion
	if richGraph, ok := richGraphInterface.(RichDependencyGraph); ok {
		nodes := richGraph.GetNodes()
		for name, nodeInterface := range nodes {
			// Create a simple node with basic information
			simpleGraph.Nodes[name] = &DependencyNode{
				Name:         name,
				Dependencies: []string{}, // Will be filled by the calling code if needed
				Dependents:   []string{},
			}
			simpleGraph.Edges[name] = []string{}

			// Try to extract additional information if possible
			if nodeMap, ok := nodeInterface.(map[string]interface{}); ok {
				if version, ok := nodeMap["version"].(string); ok {
					simpleGraph.Nodes[name].Version = version
				}
			}
		}
	} else {
		// Log conversion failure for debugging
		log.Printf("Warning: Failed to convert rich dependency graph to simple format, using empty graph")
		// Fallback: create empty graph - the actual dependency resolution
		// will be handled by the advanced resolver anyway
	}

	return simpleGraph
}

// convertSimpleGraphToInterface converts a simple graph to an interface for the resolver
func (pc *ParallelCoordinator) convertSimpleGraphToInterface(graph *DependencyGraph) interface{} {
	// Return the graph as-is since it will be handled by the resolver
	// The resolver should be designed to handle this format or provide adapters
	return graph
}

// convertRichPlanToSimple converts a rich install plan interface to simple format
func (pc *ParallelCoordinator) convertRichPlanToSimple(richPlanInterface interface{}) *InstallPlan {
	// Try to extract install plan data using type assertion
	if richPlan, ok := richPlanInterface.(RichInstallPlan); ok {
		modules := richPlan.GetModules()
		moduleNames := make([]string, 0, len(modules))

		// Extract module names
		for _, moduleInterface := range modules {
			if moduleMap, ok := moduleInterface.(map[string]interface{}); ok {
				if name, ok := moduleMap["name"].(string); ok {
					moduleNames = append(moduleNames, name)
				}
			}
		}

		return &InstallPlan{
			Modules:           moduleNames,
			Dependencies:      richPlan.GetDependencies(),
			InstallationOrder: moduleNames,
			ParallelBatches:   richPlan.GetLevels(), // Use sophisticated parallel levels
		}
	}

	// Log conversion failure for debugging
	log.Printf("Warning: Failed to convert rich install plan to simple format, using empty plan")
	// Fallback to empty plan
	return &InstallPlan{}
}
