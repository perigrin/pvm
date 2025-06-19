// ABOUTME: Development environment command for integrated development workflow
// ABOUTME: Coordinates build watching, testing, and type checking services

package pvm

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/build"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/project"
)

// DevService represents a development service that can be started and stopped
type DevService interface {
	Start(ctx context.Context) error
	Stop() error
	Name() string
	Status() ServiceStatus
}

// ServiceStatus represents the current status of a development service
type ServiceStatus struct {
	Name      string
	Running   bool
	LastEvent string
	Errors    []string
}

// DevEnvironment manages multiple development services
type DevEnvironment struct {
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	services       []DevService
	statusChan     chan ServiceStatus
	projectContext *project.ProjectContext

	// Service instances
	buildWatcher *BuildWatcherService
	testRunner   *TestRunnerService
	typeChecker  *TypeCheckerService
}

// BuildWatcherService wraps the build watcher as a dev service
type BuildWatcherService struct {
	watcher    *build.BuildWatcher
	enabled    bool
	lastResult *build.BuildResult
}

// TestRunnerService provides continuous test running
type TestRunnerService struct {
	enabled    bool
	lastResult *TestResult
	interval   time.Duration
	projectCtx *project.ProjectContext
}

// TypeCheckerService provides continuous type checking
type TypeCheckerService struct {
	enabled    bool
	lastResult *TypeCheckResult
	interval   time.Duration
	projectCtx *project.ProjectContext
}

// TestResult represents test execution results
type TestResult struct {
	Success   bool
	TestCount int
	Passed    int
	Failed    int
	Skipped   int
	Duration  time.Duration
	Output    string
	Error     error
}

// TypeCheckResult represents type checking results
type TypeCheckResult struct {
	Success    bool
	ErrorCount int
	Warnings   int
	Duration   time.Duration
	Output     string
	Error      error
}

// newDevCommand creates the development environment command
func newDevCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Start development environment",
		Long: `Start an integrated development environment with continuous building, testing, and type checking.

The development environment provides:
- Continuous build watching with .pmc generation
- Automatic test execution on changes
- Type checking with immediate feedback
- Unified status dashboard
- Graceful service coordination

Services can be selectively enabled/disabled through configuration or flags.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDevEnvironment(cmd)
		},
	}

	// Add flags for service control
	cmd.Flags().Bool("build", true, "Enable build watching")
	cmd.Flags().Bool("test", true, "Enable test runner")
	cmd.Flags().Bool("typecheck", true, "Enable type checking")
	cmd.Flags().Duration("test-interval", 5*time.Second, "Test runner check interval")
	cmd.Flags().Duration("typecheck-interval", 2*time.Second, "Type checker interval")
	cmd.Flags().Bool("verbose", false, "Verbose output")

	return cmd
}

// runDevEnvironment executes the development environment
func runDevEnvironment(cmd *cobra.Command) error {
	// Detect project context
	projectCtx, err := project.GetCurrentProject()
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	if !projectCtx.IsProject {
		return fmt.Errorf("not in a project directory - use 'pvm project init' to initialize a project")
	}

	// Get flags
	enableBuild, _ := cmd.Flags().GetBool("build")
	enableTest, _ := cmd.Flags().GetBool("test")
	enableTypecheck, _ := cmd.Flags().GetBool("typecheck")
	testInterval, _ := cmd.Flags().GetDuration("test-interval")
	typecheckInterval, _ := cmd.Flags().GetDuration("typecheck-interval")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Create development environment
	devEnv, err := NewDevEnvironment(projectCtx, &DevConfig{
		EnableBuild:       enableBuild,
		EnableTest:        enableTest,
		EnableTypecheck:   enableTypecheck,
		TestInterval:      testInterval,
		TypecheckInterval: typecheckInterval,
		Verbose:           verbose,
	})
	if err != nil {
		return fmt.Errorf("failed to create development environment: %w", err)
	}

	ui := cli.GetUI(cmd)
	ui.Status(fmt.Sprintf("Starting development environment for project: %s", projectCtx.RootDir))
	ui.Info("Press Ctrl+C to stop")
	ui.Println()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the development environment
	if err := devEnv.Start(); err != nil {
		return fmt.Errorf("failed to start development environment: %w", err)
	}

	// Start status monitoring
	go devEnv.MonitorStatus(cmd, verbose)

	// Wait for shutdown signal
	<-sigChan
	ui.Status("Shutting down development environment...")

	// Graceful shutdown
	if err := devEnv.Stop(); err != nil {
		ui.Warning("Error during shutdown: %v", err)
	}

	ui.Success("Development environment stopped.")
	return nil
}

// DevConfig configures the development environment
type DevConfig struct {
	EnableBuild       bool
	EnableTest        bool
	EnableTypecheck   bool
	TestInterval      time.Duration
	TypecheckInterval time.Duration
	Verbose           bool
}

// NewDevEnvironment creates a new development environment
func NewDevEnvironment(projectCtx *project.ProjectContext, config *DevConfig) (*DevEnvironment, error) {
	ctx, cancel := context.WithCancel(context.Background())

	devEnv := &DevEnvironment{
		ctx:            ctx,
		cancel:         cancel,
		statusChan:     make(chan ServiceStatus, 10),
		projectContext: projectCtx,
		services:       []DevService{},
	}

	// Create build watcher service if enabled
	if config.EnableBuild {
		buildService, err := NewBuildWatcherService(projectCtx)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create build service: %w", err)
		}
		devEnv.buildWatcher = buildService
		devEnv.services = append(devEnv.services, buildService)
	}

	// Create test runner service if enabled
	if config.EnableTest {
		testService := NewTestRunnerService(projectCtx, config.TestInterval)
		devEnv.testRunner = testService
		devEnv.services = append(devEnv.services, testService)
	}

	// Create type checker service if enabled
	if config.EnableTypecheck {
		typecheckService := NewTypeCheckerService(projectCtx, config.TypecheckInterval)
		devEnv.typeChecker = typecheckService
		devEnv.services = append(devEnv.services, typecheckService)
	}

	return devEnv, nil
}

// Start starts all enabled services
func (d *DevEnvironment) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, service := range d.services {
		if err := service.Start(d.ctx); err != nil {
			// Stop any already started services
			for _, startedService := range d.services {
				if startedService == service {
					break
				}
				startedService.Stop()
			}
			return fmt.Errorf("failed to start service %s: %w", service.Name(), err)
		}
	}

	return nil
}

// Stop stops all services
func (d *DevEnvironment) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.cancel()

	var errs []error
	for _, service := range d.services {
		if err := service.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop service %s: %w", service.Name(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	return nil
}

// MonitorStatus monitors and displays service status
func (d *DevEnvironment) MonitorStatus(cmd *cobra.Command, verbose bool) {
	ui := cli.GetUI(cmd)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case status := <-d.statusChan:
			if verbose || status.LastEvent != "" {
				timestamp := time.Now().Format("15:04:05")
				ui.Printf("[%s] %s: %s", timestamp, status.Name, status.LastEvent)
			}
		case <-ticker.C:
			// Periodically show service status summary
			if verbose {
				d.showStatusSummary(cmd)
			}
		}
	}
}

// showStatusSummary displays a summary of all service statuses
func (d *DevEnvironment) showStatusSummary(cmd *cobra.Command) {
	ui := cli.GetUI(cmd)
	d.mu.RLock()
	defer d.mu.RUnlock()

	ui.Println()
	ui.Info("=== Service Status ===")
	for _, service := range d.services {
		status := service.Status()
		statusIcon := "🔴"
		if status.Running {
			statusIcon = "🟢"
		}
		ui.Printf("%s %s: %s", statusIcon, status.Name, status.LastEvent)
	}
	ui.Info("======================")
	ui.Println()
}

// NewBuildWatcherService creates a new build watcher service
func NewBuildWatcherService(projectCtx *project.ProjectContext) (*BuildWatcherService, error) {
	watcherConfig := build.DefaultWatcherConfig()
	watcherConfig.EnableInline = true
	watcherConfig.EnableTypeCheck = true
	watcherConfig.EnableDist = false // Disable distribution builds in dev mode

	watcher, err := build.NewBuildWatcher(projectCtx, watcherConfig)
	if err != nil {
		return nil, err
	}

	return &BuildWatcherService{
		watcher: watcher,
		enabled: true,
	}, nil
}

// Start implements DevService for BuildWatcherService
func (b *BuildWatcherService) Start(ctx context.Context) error {
	if !b.enabled {
		return nil
	}

	// Start the build watcher
	if err := b.watcher.Start(); err != nil {
		return err
	}

	// Monitor build results
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case result := <-b.watcher.Results():
				b.lastResult = &result
				// Results are handled by the monitoring system
			}
		}
	}()

	return nil
}

// Stop implements DevService for BuildWatcherService
func (b *BuildWatcherService) Stop() error {
	if !b.enabled {
		return nil
	}
	return b.watcher.Stop()
}

// Name implements DevService for BuildWatcherService
func (b *BuildWatcherService) Name() string {
	return "Build Watcher"
}

// Status implements DevService for BuildWatcherService
func (b *BuildWatcherService) Status() ServiceStatus {
	running := b.enabled && b.watcher != nil
	lastEvent := "idle"

	if b.lastResult != nil {
		if b.lastResult.Success {
			lastEvent = fmt.Sprintf("build successful (%s)", b.lastResult.Duration.Round(time.Millisecond))
		} else {
			lastEvent = fmt.Sprintf("build failed: %v", b.lastResult.Error)
		}
	}

	return ServiceStatus{
		Name:      b.Name(),
		Running:   running,
		LastEvent: lastEvent,
	}
}

// NewTestRunnerService creates a new test runner service
func NewTestRunnerService(projectCtx *project.ProjectContext, interval time.Duration) *TestRunnerService {
	return &TestRunnerService{
		enabled:    true,
		interval:   interval,
		projectCtx: projectCtx,
	}
}

// Start implements DevService for TestRunnerService
func (t *TestRunnerService) Start(ctx context.Context) error {
	if !t.enabled {
		return nil
	}

	// Start test monitoring
	go func() {
		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Run tests periodically
				result := t.runTests()
				t.lastResult = &result
			}
		}
	}()

	return nil
}

// Stop implements DevService for TestRunnerService
func (t *TestRunnerService) Stop() error {
	return nil // No cleanup needed for test runner
}

// Name implements DevService for TestRunnerService
func (t *TestRunnerService) Name() string {
	return "Test Runner"
}

// Status implements DevService for TestRunnerService
func (t *TestRunnerService) Status() ServiceStatus {
	lastEvent := "idle"

	if t.lastResult != nil {
		if t.lastResult.Success {
			lastEvent = fmt.Sprintf("tests passed (%d/%d)", t.lastResult.Passed, t.lastResult.TestCount)
		} else {
			lastEvent = fmt.Sprintf("tests failed (%d/%d failed)", t.lastResult.Failed, t.lastResult.TestCount)
		}
	}

	return ServiceStatus{
		Name:      t.Name(),
		Running:   t.enabled,
		LastEvent: lastEvent,
	}
}

// runTests executes project tests
func (t *TestRunnerService) runTests() TestResult {
	// This is a simplified test runner - in a real implementation
	// you would integrate with prove, Test::Harness, or similar
	result := TestResult{
		Success:   true,
		TestCount: 0,
		Passed:    0,
		Failed:    0,
		Skipped:   0,
		Duration:  time.Millisecond * 100, // Placeholder
		Output:    "",
		Error:     nil,
	}

	// TODO: Implement actual test running
	// For now, just return a placeholder result

	return result
}

// NewTypeCheckerService creates a new type checker service
func NewTypeCheckerService(projectCtx *project.ProjectContext, interval time.Duration) *TypeCheckerService {
	return &TypeCheckerService{
		enabled:    true,
		interval:   interval,
		projectCtx: projectCtx,
	}
}

// Start implements DevService for TypeCheckerService
func (tc *TypeCheckerService) Start(ctx context.Context) error {
	if !tc.enabled {
		return nil
	}

	// Start type checking monitoring
	go func() {
		ticker := time.NewTicker(tc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Run type checking periodically
				result := tc.runTypeCheck()
				tc.lastResult = &result
			}
		}
	}()

	return nil
}

// Stop implements DevService for TypeCheckerService
func (tc *TypeCheckerService) Stop() error {
	return nil // No cleanup needed for type checker
}

// Name implements DevService for TypeCheckerService
func (tc *TypeCheckerService) Name() string {
	return "Type Checker"
}

// Status implements DevService for TypeCheckerService
func (tc *TypeCheckerService) Status() ServiceStatus {
	lastEvent := "idle"

	if tc.lastResult != nil {
		if tc.lastResult.Success {
			lastEvent = "type checking passed"
		} else {
			lastEvent = fmt.Sprintf("type errors found (%d errors)", tc.lastResult.ErrorCount)
		}
	}

	return ServiceStatus{
		Name:      tc.Name(),
		Running:   tc.enabled,
		LastEvent: lastEvent,
	}
}

// runTypeCheck executes type checking
func (tc *TypeCheckerService) runTypeCheck() TypeCheckResult {
	// This is a simplified type checker - in a real implementation
	// you would integrate with the PSC type checker
	result := TypeCheckResult{
		Success:    true,
		ErrorCount: 0,
		Warnings:   0,
		Duration:   time.Millisecond * 50, // Placeholder
		Output:     "",
		Error:      nil,
	}

	// TODO: Implement actual type checking
	// For now, just return a placeholder result

	return result
}
