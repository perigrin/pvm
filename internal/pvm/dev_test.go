// ABOUTME: Tests for development environment command functionality
// ABOUTME: Ensures proper service coordination and command structure

package pvm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/project"
)

func TestNewDevCommand(t *testing.T) {
	cmd := newDevCommand()

	assert.Equal(t, "dev", cmd.Use)
	assert.Equal(t, "Start development environment", cmd.Short)
	assert.Contains(t, cmd.Long, "integrated development environment")
	assert.NotNil(t, cmd.RunE)

	// Check flags
	flags := cmd.Flags()
	assert.True(t, flags.HasAvailableFlags())

	// Verify required flags exist
	buildFlag := flags.Lookup("build")
	require.NotNil(t, buildFlag)
	assert.Equal(t, "true", buildFlag.DefValue)

	testFlag := flags.Lookup("test")
	require.NotNil(t, testFlag)
	assert.Equal(t, "true", testFlag.DefValue)

	typecheckFlag := flags.Lookup("typecheck")
	require.NotNil(t, typecheckFlag)
	assert.Equal(t, "true", typecheckFlag.DefValue)

	verboseFlag := flags.Lookup("verbose")
	require.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue)
}

func TestDevConfig(t *testing.T) {
	config := &DevConfig{
		EnableBuild:       true,
		EnableTest:        false,
		EnableTypecheck:   true,
		TestInterval:      10 * time.Second,
		TypecheckInterval: 5 * time.Second,
		Verbose:           true,
	}

	assert.True(t, config.EnableBuild)
	assert.False(t, config.EnableTest)
	assert.True(t, config.EnableTypecheck)
	assert.Equal(t, 10*time.Second, config.TestInterval)
	assert.Equal(t, 5*time.Second, config.TypecheckInterval)
	assert.True(t, config.Verbose)
}

func TestServiceStatus(t *testing.T) {
	status := ServiceStatus{
		Name:      "Test Service",
		Running:   true,
		LastEvent: "test event",
		Errors:    []string{"error 1", "error 2"},
	}

	assert.Equal(t, "Test Service", status.Name)
	assert.True(t, status.Running)
	assert.Equal(t, "test event", status.LastEvent)
	assert.Len(t, status.Errors, 2)
}

func TestTestResult(t *testing.T) {
	result := TestResult{
		Success:   true,
		TestCount: 10,
		Passed:    8,
		Failed:    2,
		Skipped:   0,
		Duration:  time.Second,
		Output:    "test output",
		Error:     nil,
	}

	assert.True(t, result.Success)
	assert.Equal(t, 10, result.TestCount)
	assert.Equal(t, 8, result.Passed)
	assert.Equal(t, 2, result.Failed)
	assert.Equal(t, 0, result.Skipped)
	assert.Equal(t, time.Second, result.Duration)
	assert.Equal(t, "test output", result.Output)
	assert.Nil(t, result.Error)
}

func TestTypeCheckResult(t *testing.T) {
	result := TypeCheckResult{
		Success:    false,
		ErrorCount: 3,
		Warnings:   1,
		Duration:   500 * time.Millisecond,
		Output:     "type check output",
		Error:      nil,
	}

	assert.False(t, result.Success)
	assert.Equal(t, 3, result.ErrorCount)
	assert.Equal(t, 1, result.Warnings)
	assert.Equal(t, 500*time.Millisecond, result.Duration)
	assert.Equal(t, "type check output", result.Output)
	assert.Nil(t, result.Error)
}

func TestNewTestRunnerService(t *testing.T) {
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   "/test/project",
	}

	interval := 5 * time.Second
	service := NewTestRunnerService(projectCtx, interval)

	assert.NotNil(t, service)
	assert.True(t, service.enabled)
	assert.Equal(t, interval, service.interval)
	assert.Equal(t, projectCtx, service.projectCtx)
	assert.Equal(t, "Test Runner", service.Name())

	status := service.Status()
	assert.Equal(t, "Test Runner", status.Name)
	assert.True(t, status.Running)
	assert.Equal(t, "idle", status.LastEvent)
}

func TestNewTypeCheckerService(t *testing.T) {
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   "/test/project",
	}

	interval := 2 * time.Second
	service := NewTypeCheckerService(projectCtx, interval)

	assert.NotNil(t, service)
	assert.True(t, service.enabled)
	assert.Equal(t, interval, service.interval)
	assert.Equal(t, projectCtx, service.projectCtx)
	assert.Equal(t, "Type Checker", service.Name())

	status := service.Status()
	assert.Equal(t, "Type Checker", status.Name)
	assert.True(t, status.Running)
	assert.Equal(t, "idle", status.LastEvent)
}

func TestTestRunnerServiceRunTests(t *testing.T) {
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   "/test/project",
	}

	service := NewTestRunnerService(projectCtx, time.Second)
	result := service.runTests()

	// Verify placeholder result structure
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.TestCount)
	assert.Equal(t, 0, result.Passed)
	assert.Equal(t, 0, result.Failed)
	assert.Equal(t, 0, result.Skipped)
	assert.True(t, result.Duration > 0)
	assert.Empty(t, result.Output)
	assert.Nil(t, result.Error)
}

func TestTypeCheckerServiceRunTypeCheck(t *testing.T) {
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   "/test/project",
	}

	service := NewTypeCheckerService(projectCtx, time.Second)
	result := service.runTypeCheck()

	// Verify placeholder result structure
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, 0, result.Warnings)
	assert.True(t, result.Duration > 0)
	assert.Empty(t, result.Output)
	assert.Nil(t, result.Error)
}

func TestServiceStatusWithResults(t *testing.T) {
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   "/test/project",
	}

	// Test test runner with result
	testService := NewTestRunnerService(projectCtx, time.Second)
	testService.lastResult = &TestResult{
		Success:   true,
		TestCount: 5,
		Passed:    4,
		Failed:    1,
	}

	status := testService.Status()
	assert.Contains(t, status.LastEvent, "tests passed (4/5)")

	// Test type checker with result
	typeService := NewTypeCheckerService(projectCtx, time.Second)
	typeService.lastResult = &TypeCheckResult{
		Success:    false,
		ErrorCount: 3,
	}

	status = typeService.Status()
	assert.Contains(t, status.LastEvent, "type errors found (3 errors)")
}

func TestDevEnvironmentServices(t *testing.T) {
	// Test service interfaces are properly implemented
	var _ DevService = (*BuildWatcherService)(nil)
	var _ DevService = (*TestRunnerService)(nil)
	var _ DevService = (*TypeCheckerService)(nil)

	// Verify service methods exist
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   "/test/project",
	}

	testService := NewTestRunnerService(projectCtx, time.Second)
	assert.NotNil(t, testService.Start)
	assert.NotNil(t, testService.Stop)
	assert.NotNil(t, testService.Name)
	assert.NotNil(t, testService.Status)
}
