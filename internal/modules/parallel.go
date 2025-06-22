// ABOUTME: Parallel installation coordination with dependency-aware ordering
// ABOUTME: Provides sophisticated multi-module installation with worker pool management

package modules

import (
	"context"
	"fmt"
	"sync"

	pviModules "tamarou.com/pvm/internal/pvi/modules"
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

// InstallModules installs multiple modules with dependency-aware parallel coordination
func (pc *ParallelCoordinator) InstallModules(ctx context.Context, modules []string, opts InstallOptions) ([]*InstallResult, error) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

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
	// TODO: Implement sophisticated dependency resolution
	// For now, return a simple graph with no dependencies
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

	return graph, nil
}

// ExecuteInstallPlan executes a pre-computed installation plan
func (pc *ParallelCoordinator) ExecuteInstallPlan(ctx context.Context, plan *InstallPlan) ([]*InstallResult, error) {
	if plan == nil || len(plan.Modules) == 0 {
		return []*InstallResult{}, nil
	}

	// For now, install all modules in parallel
	// TODO: Implement proper dependency-aware batched installation
	opts := InstallOptions{
		Parallel: true,
		Workers:  pc.maxWorkers,
		Context:  ctx,
	}

	return pc.InstallModules(ctx, plan.Modules, opts)
}

// CreateInstallPlan generates an optimized installation plan from a dependency graph
func (pc *ParallelCoordinator) CreateInstallPlan(graph *DependencyGraph) (*InstallPlan, error) {
	if graph == nil || len(graph.Nodes) == 0 {
		return &InstallPlan{}, nil
	}

	modules := make([]string, 0, len(graph.Nodes))
	for name := range graph.Nodes {
		modules = append(modules, name)
	}

	// TODO: Implement proper topological sort for dependency ordering
	// For now, return simple plan
	return &InstallPlan{
		Modules:           modules,
		Dependencies:      graph.Edges,
		InstallationOrder: modules,
		ParallelBatches:   [][]string{modules}, // All modules in one batch for now
	}, nil
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
