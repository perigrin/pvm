// ABOUTME: Development environment command for integrated development workflow
// ABOUTME: Coordinates build watching, testing, and type checking services

package pvm

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/build"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/integration"
	"tamarou.com/pvm/internal/project"
	"tamarou.com/pvm/internal/pvx"
	"tamarou.com/pvm/internal/typechecker"
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
		return fmt.Errorf("not in a workspace directory - use 'pvm workspace init' to initialize a workspace")
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

// runTests executes project tests using validation workflow or fallback to prove/perl
func (t *TestRunnerService) runTests() TestResult {
	startTime := time.Now()

	// Try ValidationWorkflow first for comprehensive testing
	if result := t.tryValidationWorkflow(startTime); result.Success || result.Error == nil {
		return result
	}

	// Fallback to traditional test discovery and execution
	testFiles, err := t.discoverTestFiles()
	if err != nil {
		return TestResult{
			Success:  false,
			Duration: time.Since(startTime),
			Error:    err,
		}
	}

	if len(testFiles) == 0 {
		return TestResult{
			Success:   true,
			TestCount: 0,
			Duration:  time.Since(startTime),
			Output:    "No test files found",
		}
	}

	// Execute tests using prove or direct perl execution
	return t.executeTestFiles(testFiles, startTime)
}

// tryValidationWorkflow attempts to use ValidationWorkflow for comprehensive testing
func (t *TestRunnerService) tryValidationWorkflow(startTime time.Time) TestResult {
	// Find a test script to validate with
	testScript := t.findTestScript()

	// Use ValidationWorkflow for comprehensive validation
	workflowResult, err := integration.ValidationWorkflow(testScript)
	if err != nil {
		// Return a failed result but allow fallback
		return TestResult{
			Success:  false,
			Duration: time.Since(startTime),
			Error:    err,
			Output:   fmt.Sprintf("Validation workflow failed: %v", err),
		}
	}

	if workflowResult == nil {
		return TestResult{
			Success:  false,
			Duration: time.Since(startTime),
			Error:    fmt.Errorf("validation workflow returned nil result"),
		}
	}

	// Convert workflow result to test result
	success := workflowResult.TypeCheckPassed && workflowResult.ExecutionExitCode == 0

	return TestResult{
		Success:   success,
		TestCount: 1, // Validation workflow is one comprehensive test
		Passed: func() int {
			if success {
				return 1
			}
			return 0
		}(),
		Failed: func() int {
			if success {
				return 0
			}
			return 1
		}(),
		Duration: workflowResult.Duration,
		Output:   t.formatValidationOutput(workflowResult),
		Error:    nil,
	}
}

// findTestScript finds a suitable test script for validation workflow
func (t *TestRunnerService) findTestScript() string {
	// Let ValidationWorkflow create its own test script by passing empty string
	// This ensures consistent validation across different project types
	return ""
}

// formatValidationOutput formats validation workflow results for test display
func (t *TestRunnerService) formatValidationOutput(result *integration.WorkflowResult) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Validation using Perl %s", result.VersionUsed))

	if result.TypeCheckPassed && result.ExecutionExitCode == 0 {
		output.WriteString(" - validation passed")
	} else {
		output.WriteString(" - validation failed")
		if !result.TypeCheckPassed {
			output.WriteString(fmt.Sprintf(" (type check failed: %d errors)", len(result.TypeErrors)))
		}
		if result.ExecutionExitCode != 0 {
			output.WriteString(fmt.Sprintf(" (execution failed with exit code %d)", result.ExecutionExitCode))
		}
	}

	// Add execution output if available
	if result.ExecutionOutput != "" {
		output.WriteString("\nExecution output:\n")
		output.WriteString(result.ExecutionOutput)
	}

	// Add type definition information
	if result.TypeDefGenerated {
		output.WriteString(fmt.Sprintf("\nType definitions validated: %s", result.TypeDefPath))
	}

	return output.String()
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

// runTypeCheck executes type checking using integration workflows
func (tc *TypeCheckerService) runTypeCheck() TypeCheckResult {
	startTime := time.Now()

	// Find the main script or library entry point for type checking
	scriptPath, err := tc.findMainScriptPath()
	if err != nil {
		return TypeCheckResult{
			Success:  false,
			Duration: time.Since(startTime),
			Error:    fmt.Errorf("failed to find script for type checking: %w", err),
		}
	}

	if scriptPath == "" {
		return TypeCheckResult{
			Success:  true,
			Duration: time.Since(startTime),
			Output:   "No suitable files found for type checking",
		}
	}

	// Use TypeCheckWorkflow for comprehensive type checking
	perlVersion := tc.resolvePerlVersion()
	workflowResult, err := integration.TypeCheckWorkflow(scriptPath, perlVersion, false)

	// Convert WorkflowResult to TypeCheckResult
	return tc.convertWorkflowResult(workflowResult, err, time.Since(startTime))
}

// discoverTestFiles finds all test files in the project
func (t *TestRunnerService) discoverTestFiles() ([]string, error) {
	var testDirs []string

	if t.projectCtx.IsProject {
		// Standard Perl test directories
		testDirs = []string{
			filepath.Join(t.projectCtx.RootDir, "t"),
			filepath.Join(t.projectCtx.RootDir, "tests"),
		}
	} else {
		// Look in current directory
		testDirs = []string{"t", "tests", "."}
	}

	var testFiles []string

	for _, dir := range testDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			// Look for .t files (standard Perl test files)
			if strings.HasSuffix(path, ".t") {
				testFiles = append(testFiles, path)
			}

			return nil
		})

		if err != nil {
			return []string{}, fmt.Errorf("failed to walk test directory %s: %w", dir, err)
		}
	}

	// Always return a valid slice, even if empty
	if testFiles == nil {
		testFiles = []string{}
	}

	return testFiles, nil
}

// executeTestFiles runs test files using the most appropriate method
func (t *TestRunnerService) executeTestFiles(testFiles []string, startTime time.Time) TestResult {
	// Try prove first (standard Perl test runner)
	if result := t.tryExecuteWithProve(testFiles, startTime); result.Success || result.Error == nil {
		return result
	}

	// Fallback to direct perl execution
	return t.executeWithDirectPerl(testFiles, startTime)
}

// tryExecuteWithProve attempts to use prove for test execution
func (t *TestRunnerService) tryExecuteWithProve(testFiles []string, startTime time.Time) TestResult {
	// Create a prove command that will run all test files
	proveCode := fmt.Sprintf("exec { 'prove' } 'prove', '-v', %s",
		strings.Join(func() []string {
			var quoted []string
			for _, file := range testFiles {
				quoted = append(quoted, fmt.Sprintf("'%s'", file))
			}
			return quoted
		}(), ", "))

	// Set up execution options for prove
	options := &pvx.ExecutionOptions{
		InlineCode:     proveCode,
		PerlVersion:    t.resolvePerlVersion(),
		IsolationLevel: pvx.IsolationLocal, // Use local isolation for test environment
		Env: map[string]string{
			"PERL_TEST_HARNESS_DUMP_TAP": "1", // Enable TAP dumping
		},
	}

	// Set up project-aware environment
	if t.projectCtx.IsProject {
		if options.Env == nil {
			options.Env = make(map[string]string)
		}

		// Add project lib paths to PERL5LIB
		projectLibPaths := []string{
			filepath.Join(t.projectCtx.RootDir, "lib"),
			filepath.Join(t.projectCtx.RootDir, "local", "lib", "perl5"),
		}

		for _, path := range projectLibPaths {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				if existingPerl5Lib := options.Env["PERL5LIB"]; existingPerl5Lib != "" {
					options.Env["PERL5LIB"] = path + string(os.PathListSeparator) + existingPerl5Lib
				} else {
					options.Env["PERL5LIB"] = path
				}
			}
		}
	}

	output, err := pvx.ExecuteInlineCode(options)

	result := TestResult{
		Duration: time.Since(startTime),
		Output:   output,
		Error:    err,
	}

	if err != nil {
		result.Success = false
		result.Failed = len(testFiles) // Assume all failed if prove fails
		return result
	}

	// Parse TAP output from prove
	t.parseTAPOutput(output, &result)
	return result
}

// executeWithDirectPerl runs tests directly with perl (fallback method)
func (t *TestRunnerService) executeWithDirectPerl(testFiles []string, startTime time.Time) TestResult {
	totalTests := 0
	passedFiles := 0
	failedFiles := 0
	var outputs []string

	for _, testFile := range testFiles {
		options := &pvx.ExecutionOptions{
			ScriptPath:     testFile,
			PerlVersion:    t.resolvePerlVersion(),
			IsolationLevel: pvx.IsolationLocal,
		}

		// Set up project-aware environment
		if t.projectCtx.IsProject {
			options.Env = make(map[string]string)

			// Add project lib paths to PERL5LIB
			projectLibPaths := []string{
				filepath.Join(t.projectCtx.RootDir, "lib"),
				filepath.Join(t.projectCtx.RootDir, "local", "lib", "perl5"),
			}

			for _, path := range projectLibPaths {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					if existingPerl5Lib := options.Env["PERL5LIB"]; existingPerl5Lib != "" {
						options.Env["PERL5LIB"] = path + string(os.PathListSeparator) + existingPerl5Lib
					} else {
						options.Env["PERL5LIB"] = path
					}
				}
			}
		}

		output, err := pvx.ExecuteScript(options)
		outputs = append(outputs, fmt.Sprintf("=== %s ===\n%s", testFile, output))

		if err != nil {
			failedFiles++
		} else {
			passedFiles++
			// Try to extract test count from TAP output
			totalTests += t.parseTestCount(output)
		}
	}

	return TestResult{
		Success:   failedFiles == 0,
		TestCount: totalTests,
		Passed:    passedFiles,
		Failed:    failedFiles,
		Duration:  time.Since(startTime),
		Output:    strings.Join(outputs, "\n"),
	}
}

// resolvePerlVersion determines which Perl version to use for testing
func (t *TestRunnerService) resolvePerlVersion() string {
	// Use project-specific version if available
	if t.projectCtx.IsProject && t.projectCtx.PerlVersion != "" {
		return t.projectCtx.PerlVersion
	}

	// Look for .perl-version file
	if version := t.readVersionFile(".perl-version"); version != "" {
		return version
	}

	// Fall back to system Perl (empty string lets PVX handle resolution)
	return ""
}

// readVersionFile reads version from a file
func (t *TestRunnerService) readVersionFile(filename string) string {
	var filePath string
	if t.projectCtx.IsProject {
		filePath = filepath.Join(t.projectCtx.RootDir, filename)
	} else {
		filePath = filename
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

// parseTAPOutput parses TAP output from prove and updates TestResult
func (t *TestRunnerService) parseTAPOutput(output string, result *TestResult) {
	lines := strings.Split(output, "\n")

	var totalTests, passedTests, failedTests int

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for test plan (1..N)
		if matches := regexp.MustCompile(`^1\.\.(\d+)`).FindStringSubmatch(line); matches != nil {
			if count, err := strconv.Atoi(matches[1]); err == nil {
				totalTests += count
			}
		}

		// Count passed tests
		if strings.HasPrefix(line, "ok ") {
			passedTests++
		}

		// Count failed tests
		if strings.HasPrefix(line, "not ok ") {
			failedTests++
		}

		// Look for prove summary
		if strings.Contains(line, "All tests successful") {
			result.Success = true
		}

		// Look for failure summary
		if matches := regexp.MustCompile(`(\d+)/(\d+) subtests failed`).FindStringSubmatch(line); matches != nil {
			result.Success = false
			if failed, err := strconv.Atoi(matches[1]); err == nil {
				failedTests = failed
			}
			if total, err := strconv.Atoi(matches[2]); err == nil {
				totalTests = total
			}
		}
	}

	result.TestCount = totalTests
	result.Passed = passedTests
	result.Failed = failedTests

	if result.Success && failedTests > 0 {
		result.Success = false
	}

	// If we don't have explicit success indication and no failures, assume success
	if !result.Success && failedTests == 0 && totalTests > 0 {
		result.Success = true
	}
}

// parseTestCount extracts test count from TAP output (for direct perl execution)
func (t *TestRunnerService) parseTestCount(output string) int {
	// Look for "1..N" pattern in TAP output
	re := regexp.MustCompile(`^1\.\.(\d+)`)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if matches := re.FindStringSubmatch(strings.TrimSpace(line)); matches != nil {
			if len(matches) > 1 {
				if count, err := strconv.Atoi(matches[1]); err == nil {
					return count
				}
			}
		}
	}

	// Fallback: count "ok" and "not ok" lines
	count := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "ok ") || strings.HasPrefix(trimmed, "not ok ") {
			count++
		}
	}
	return count
}

// discoverPerlFiles finds all Perl files in the project for type checking
func (tc *TypeCheckerService) discoverPerlFiles() ([]string, error) {
	var searchDirs []string
	var perlFiles []string

	if tc.projectCtx.IsProject {
		// Standard Perl project directories
		searchDirs = []string{
			filepath.Join(tc.projectCtx.RootDir, "lib"),
			filepath.Join(tc.projectCtx.RootDir, "bin"),
			filepath.Join(tc.projectCtx.RootDir, "scripts"),
			tc.projectCtx.RootDir, // Check root for loose .pl files
		}
	} else {
		// Look in current directory
		searchDirs = []string{"."}
	}

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				// Skip hidden directories and common non-source directories
				baseName := filepath.Base(path)
				if strings.HasPrefix(baseName, ".") ||
					baseName == "t" || baseName == "tests" || baseName == "local" {
					return filepath.SkipDir
				}
				return nil
			}

			// Look for .pl and .pm files (standard Perl source files)
			if strings.HasSuffix(path, ".pl") || strings.HasSuffix(path, ".pm") {
				perlFiles = append(perlFiles, path)
			}

			return nil
		})

		if err != nil {
			return []string{}, fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}
	}

	// Always return a valid slice, even if empty
	if perlFiles == nil {
		perlFiles = []string{}
	}

	return perlFiles, nil
}

// findMainScriptPath finds the main script or entry point for type checking
func (tc *TypeCheckerService) findMainScriptPath() (string, error) {
	// Priority order for finding type check entry points:
	// 1. Main script in bin/ directory
	// 2. .pl files in root directory
	// 3. Main module in lib/ directory
	// 4. First .pl or .pm file found

	var candidates []string

	if tc.projectCtx.IsProject {
		rootDir := tc.projectCtx.RootDir

		// Check bin/ directory for main scripts
		binDir := filepath.Join(rootDir, "bin")
		if _, err := os.Stat(binDir); !os.IsNotExist(err) {
			if binFiles, err := tc.findPerlFilesInDir(binDir); err == nil {
				candidates = append(candidates, binFiles...)
			}
		}

		// Check root directory for .pl files
		if rootFiles, err := tc.findPerlFilesInDir(rootDir); err == nil {
			for _, file := range rootFiles {
				if strings.HasSuffix(file, ".pl") {
					candidates = append(candidates, file)
				}
			}
		}

		// Check lib/ directory for main modules
		libDir := filepath.Join(rootDir, "lib")
		if _, err := os.Stat(libDir); !os.IsNotExist(err) {
			if libFiles, err := tc.findPerlFilesInDir(libDir); err == nil {
				candidates = append(candidates, libFiles...)
			}
		}
	} else {
		// For non-project directories, look in current directory
		if files, err := tc.findPerlFilesInDir("."); err == nil {
			candidates = files
		}
	}

	if len(candidates) == 0 {
		return "", nil
	}

	// Return the first candidate (prioritized by order above)
	return candidates[0], nil
}

// findPerlFilesInDir finds Perl files in a specific directory (non-recursive)
func (tc *TypeCheckerService) findPerlFilesInDir(dir string) ([]string, error) {
	var perlFiles []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".pl") || strings.HasSuffix(name, ".pm") {
			fullPath := filepath.Join(dir, name)
			perlFiles = append(perlFiles, fullPath)
		}
	}

	return perlFiles, nil
}

// resolvePerlVersion determines which Perl version to use for type checking
func (tc *TypeCheckerService) resolvePerlVersion() string {
	// Use project-specific version if available
	if tc.projectCtx.IsProject && tc.projectCtx.PerlVersion != "" {
		return tc.projectCtx.PerlVersion
	}

	// Look for .perl-version file
	if version := tc.readVersionFile(".perl-version"); version != "" {
		return version
	}

	// Fall back to empty string (let integration workflow handle system Perl)
	return ""
}

// readVersionFile reads version from a file
func (tc *TypeCheckerService) readVersionFile(filename string) string {
	var filePath string
	if tc.projectCtx.IsProject {
		filePath = filepath.Join(tc.projectCtx.RootDir, filename)
	} else {
		filePath = filename
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

// convertWorkflowResult converts integration.WorkflowResult to TypeCheckResult
func (tc *TypeCheckerService) convertWorkflowResult(workflowResult *integration.WorkflowResult, workflowErr error, duration time.Duration) TypeCheckResult {
	if workflowErr != nil {
		return TypeCheckResult{
			Success:    false,
			ErrorCount: 1,
			Duration:   duration,
			Error:      workflowErr,
			Output:     fmt.Sprintf("Workflow failed: %v", workflowErr),
		}
	}

	if workflowResult == nil {
		return TypeCheckResult{
			Success:  false,
			Duration: duration,
			Error:    fmt.Errorf("workflow returned nil result"),
		}
	}

	// Extract type checking information from workflow result
	typeCheckPassed := workflowResult.TypeCheckPassed
	errorCount := len(workflowResult.TypeErrors)

	// Estimate warnings from type errors (using existing heuristic)
	warningCount := 0
	for _, typeErr := range workflowResult.TypeErrors {
		if strings.Contains(strings.ToLower(typeErr.Message), "warning") ||
			strings.Contains(strings.ToLower(typeErr.Message), "unused") ||
			strings.Contains(strings.ToLower(typeErr.Message), "deprecated") {
			warningCount++
		}
	}

	// Format output with rich information from workflow
	output := tc.formatWorkflowOutput(workflowResult, errorCount, warningCount)

	return TypeCheckResult{
		Success:    typeCheckPassed && errorCount == 0,
		ErrorCount: errorCount,
		Warnings:   warningCount,
		Duration:   workflowResult.Duration,
		Output:     output,
		Error:      nil,
	}
}

// formatWorkflowOutput formats workflow results for display
func (tc *TypeCheckerService) formatWorkflowOutput(result *integration.WorkflowResult, errorCount, warningCount int) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Type checked using Perl %s", result.VersionUsed))

	if errorCount == 0 && warningCount == 0 {
		output.WriteString(" - no issues found")
	} else {
		if errorCount > 0 {
			output.WriteString(fmt.Sprintf(" - %d error(s)", errorCount))
		}
		if warningCount > 0 {
			if errorCount > 0 {
				output.WriteString(",")
			}
			output.WriteString(fmt.Sprintf(" %d warning(s)", warningCount))
		}
	}

	// Add type definition information
	if result.TypeDefGenerated {
		output.WriteString(fmt.Sprintf("\nType definitions generated: %s", result.TypeDefPath))
	}

	// Add error details (limited for dev environment)
	if len(result.TypeErrors) > 0 {
		output.WriteString("\n\nType checking issues:")
		maxErrors := 5 // Fewer errors for dev display
		for i, typeErr := range result.TypeErrors {
			if i >= maxErrors {
				output.WriteString(fmt.Sprintf("\n... and %d more issues (use 'psc check' for complete output)", len(result.TypeErrors)-maxErrors))
				break
			}
			output.WriteString(fmt.Sprintf("\n  %s:%d:%d: %s",
				typeErr.Path, typeErr.Line, typeErr.Column, typeErr.Message))
		}
	}

	// Add module information if relevant
	if len(result.ModulesInstalled) > 0 {
		output.WriteString(fmt.Sprintf("\nModules installed: %d", len(result.ModulesInstalled)))
	}

	return output.String()
}

// runFileLevelTypeCheck performs incremental type checking on individual files
func (tc *TypeCheckerService) runFileLevelTypeCheck(perlFiles []string, startTime time.Time) TypeCheckResult {
	// Create a new type checker instance
	typeChecker, err := typechecker.NewTypeCheck()
	if err != nil {
		return TypeCheckResult{
			Success:  false,
			Duration: time.Since(startTime),
			Error:    fmt.Errorf("failed to create type checker: %w", err),
		}
	}

	var allErrors []string
	var totalErrorCount int
	var totalWarningCount int
	checkedFiles := 0

	// Check each file individually (incremental approach)
	for _, file := range perlFiles {
		// Only check recently modified files in development environment
		if tc.shouldSkipFile(file) {
			continue
		}

		result, err := typeChecker.CheckFile(file)
		if err != nil {
			// Don't fail entire check for individual file errors in dev environment
			allErrors = append(allErrors, fmt.Sprintf("%s: %v", file, err))
			continue
		}

		checkedFiles++

		if len(result.Errors) > 0 {
			totalErrorCount += len(result.Errors)

			// Format errors for display
			for _, typeErr := range result.Errors {
				errorMsg := fmt.Sprintf("%s:%d:%d: %s",
					file,
					typeErr.Line,
					typeErr.Column,
					typeErr.Message)
				allErrors = append(allErrors, errorMsg)
			}
		}

		// Count warnings if available
		// Note: The current typechecker doesn't expose warnings directly,
		// but we can estimate based on less severe errors
		totalWarningCount += tc.estimateWarnings(result)
	}

	success := totalErrorCount == 0
	duration := time.Since(startTime)

	return TypeCheckResult{
		Success:    success,
		ErrorCount: totalErrorCount,
		Warnings:   totalWarningCount,
		Duration:   duration,
		Output:     tc.formatTypeCheckOutput(checkedFiles, totalErrorCount, totalWarningCount, allErrors),
		Error:      nil,
	}
}

// shouldSkipFile determines if a file should be skipped in incremental checking
func (tc *TypeCheckerService) shouldSkipFile(filePath string) bool {
	// In development environment, we might want to skip files that haven't
	// been modified recently to improve performance. For now, check all files
	// but this could be enhanced with file modification time tracking.

	// Skip if file doesn't exist or can't be read
	if _, err := os.Stat(filePath); err != nil {
		return true
	}

	return false
}

// estimateWarnings estimates warning count from type check results
func (tc *TypeCheckerService) estimateWarnings(result *typechecker.TypeCheckResult) int {
	// This is a heuristic since the typechecker doesn't currently expose warnings separately
	// We could enhance this by examining error severity or types
	warningCount := 0

	for _, err := range result.Errors {
		// Consider certain error patterns as warnings rather than errors
		if strings.Contains(strings.ToLower(err.Message), "warning") ||
			strings.Contains(strings.ToLower(err.Message), "unused") ||
			strings.Contains(strings.ToLower(err.Message), "deprecated") {
			warningCount++
		}
	}

	return warningCount
}

// formatTypeCheckOutput formats the type checking results for display
func (tc *TypeCheckerService) formatTypeCheckOutput(checkedFiles, errorCount, warningCount int, errors []string) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Type checked %d file(s)", checkedFiles))

	if errorCount == 0 && warningCount == 0 {
		output.WriteString(" - no issues found")
	} else {
		if errorCount > 0 {
			output.WriteString(fmt.Sprintf(" - %d error(s)", errorCount))
		}
		if warningCount > 0 {
			if errorCount > 0 {
				output.WriteString(",")
			}
			output.WriteString(fmt.Sprintf(" %d warning(s)", warningCount))
		}
	}

	if len(errors) > 0 {
		output.WriteString("\n\nIssues found:")
		// Limit output in development environment to avoid overwhelming the console
		maxErrors := 10
		for i, err := range errors {
			if i >= maxErrors {
				output.WriteString(fmt.Sprintf("\n... and %d more issues (use 'psc check' for complete output)", len(errors)-maxErrors))
				break
			}
			output.WriteString("\n  ")
			output.WriteString(err)
		}
	}

	return output.String()
}
