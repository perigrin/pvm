// ABOUTME: Parallel module installer implementing cpm-style batch installation
// ABOUTME: Uses worker pool for concurrent module installation with dependency ordering

package modules

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/parallel"
)

// ModuleInstallTask implements the parallel.Task interface for module installation
type ModuleInstallTask struct {
	options  *ModuleInstallOptions
	result   *ModuleInstallResult
	priority int
	id       string
	mutex    sync.Mutex
}

// NewModuleInstallTask creates a new module installation task
func NewModuleInstallTask(options *ModuleInstallOptions, priority int) *ModuleInstallTask {
	return &ModuleInstallTask{
		options:  options,
		priority: priority,
		id:       fmt.Sprintf("install-%s", options.ModuleName),
	}
}

// Execute implements the Task interface
func (t *ModuleInstallTask) Execute(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Create a copy of options with the provided context
	opts := *t.options
	opts.Context = ctx

	result, err := InstallModule(&opts)
	t.result = result
	return err
}

// Priority implements the Task interface
func (t *ModuleInstallTask) Priority() int {
	return t.priority
}

// ID implements the Task interface
func (t *ModuleInstallTask) ID() string {
	return t.id
}

// GetResult returns the installation result (safe to call after Execute)
func (t *ModuleInstallTask) GetResult() *ModuleInstallResult {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.result
}

// ParallelInstallOptions contains options for parallel module installation
type ParallelInstallOptions struct {
	// Modules to install with their individual options
	Modules []*ModuleInstallOptions

	// Number of parallel workers (0 = auto-detect)
	Workers int

	// Whether to stop on first error
	StopOnError bool

	// Maximum time to wait for all installations
	Timeout time.Duration

	// Progress callback for overall progress
	ProgressCallback func(completed, total int, currentModule string, stage InstallProgressStage)

	// Context for cancellation
	Context context.Context
}

// ParallelInstallResult contains results from parallel installation
type ParallelInstallResult struct {
	// Individual results for each module
	Results []*ModuleInstallResult

	// Modules that failed to install
	Failures []ModuleInstallFailure

	// Total time taken
	Duration time.Duration

	// Number of successful installations
	SuccessCount int

	// Number of failed installations
	FailureCount int

	// Installation order (modules installed in parallel may complete out of order)
	InstallationOrder []string
}

// ModuleInstallFailure represents a failed module installation
type ModuleInstallFailure struct {
	ModuleName string
	Error      error
	Duration   time.Duration
}

// InstallModulesParallel installs multiple modules in parallel using the worker pool
func InstallModulesParallel(options *ParallelInstallOptions) (*ParallelInstallResult, error) {
	startTime := time.Now()

	if options == nil || len(options.Modules) == 0 {
		return &ParallelInstallResult{
			Results:  []*ModuleInstallResult{},
			Failures: []ModuleInstallFailure{},
			Duration: time.Since(startTime),
		}, nil
	}

	// Determine number of workers
	workers := options.Workers
	if workers <= 0 {
		workers = min(len(options.Modules), 4) // Default to 4 workers or number of modules, whichever is smaller
	}

	// Create worker pool configuration
	config := parallel.DefaultPoolConfig()
	config.NumWorkers = workers
	config.QueueSize = len(options.Modules) * 2 // Buffer for retries

	// Create worker pool with a logger
	logger := log.NewLogger(log.LevelInfo, os.Stderr, "pvi-parallel")
	pool := parallel.NewWorkerPool(config, logger)

	// Create context with timeout if specified
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	// Start the pool
	err := pool.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start worker pool: %w", err)
	}
	defer pool.Stop()

	// Create tasks and track them
	tasks := make([]*ModuleInstallTask, len(options.Modules))

	for i, moduleOpts := range options.Modules {
		// Create progress callback wrapper
		if options.ProgressCallback != nil {
			originalCallback := moduleOpts.ProgressCallback
			moduleOpts.ProgressCallback = func(stage InstallProgressStage, moduleName string, details string, progress float64) {
				// Call original callback if it exists
				if originalCallback != nil {
					originalCallback(stage, moduleName, details, progress)
				}
				// Call overall progress callback
				completed := countCompletedTasks(tasks)
				options.ProgressCallback(completed, len(tasks), moduleName, stage)
			}
		}

		// Create task with priority based on dependency order (lower index = higher priority)
		task := NewModuleInstallTask(moduleOpts, len(options.Modules)-i)
		tasks[i] = task

		// Submit task to pool
		err = pool.Submit(task)
		if err != nil {
			return nil, fmt.Errorf("failed to submit task for module %s: %w", moduleOpts.ModuleName, err)
		}
	}

	// Wait for completion and collect results
	results := make([]*ModuleInstallResult, 0, len(tasks))
	failures := make([]ModuleInstallFailure, 0)
	installationOrder := make([]string, 0, len(tasks))

	// Wait for all tasks to complete
	pool.WaitForCompletion()

	// Collect results
	for _, task := range tasks {
		result := task.GetResult()
		if result != nil {
			results = append(results, result)
			installationOrder = append(installationOrder, result.ModuleName)

			if !result.Success {
				failures = append(failures, ModuleInstallFailure{
					ModuleName: result.ModuleName,
					Error:      fmt.Errorf("installation failed"),
					Duration:   result.Duration,
				})
			}
		} else {
			// Task didn't complete or failed before producing a result
			failures = append(failures, ModuleInstallFailure{
				ModuleName: task.options.ModuleName,
				Error:      fmt.Errorf("task failed to complete"),
				Duration:   0,
			})
		}
	}

	return &ParallelInstallResult{
		Results:           results,
		Failures:          failures,
		Duration:          time.Since(startTime),
		SuccessCount:      len(results) - len(failures),
		FailureCount:      len(failures),
		InstallationOrder: installationOrder,
	}, nil
}

// countCompletedTasks counts how many tasks have completed
func countCompletedTasks(tasks []*ModuleInstallTask) int {
	completed := 0
	for _, task := range tasks {
		if task.GetResult() != nil {
			completed++
		}
	}
	return completed
}

// InstallModulesBatch is a convenience function for batch installation with sensible defaults
func InstallModulesBatch(moduleNames []string, options *ModuleInstallOptions) (*ParallelInstallResult, error) {
	if len(moduleNames) == 0 {
		return &ParallelInstallResult{}, nil
	}

	// Create module options for each module
	modules := make([]*ModuleInstallOptions, len(moduleNames))
	for i, name := range moduleNames {
		opts := &ModuleInstallOptions{
			ModuleName: name,
		}

		// Copy common options if provided
		if options != nil {
			opts.VersionConstraint = options.VersionConstraint
			opts.PerlPath = options.PerlPath
			opts.InstallDir = options.InstallDir
			opts.Force = options.Force
			opts.RunTests = options.RunTests
			opts.SkipDependencies = options.SkipDependencies
			opts.Cleanup = options.Cleanup
			opts.Verbose = options.Verbose
			opts.BuildArgs = options.BuildArgs
			opts.Provider = options.Provider
			opts.DependencyResolver = options.DependencyResolver
			opts.ProgressCallback = options.ProgressCallback
			opts.Context = options.Context
		}

		modules[i] = opts
	}

	parallelOpts := &ParallelInstallOptions{
		Modules:     modules,
		Workers:     0, // Auto-detect
		StopOnError: false,
		Context:     options.Context,
	}

	return InstallModulesParallel(parallelOpts)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
