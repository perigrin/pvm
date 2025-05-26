package modules

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleInstallTask(t *testing.T) {
	t.Run("TaskInterfaceImplementation", func(t *testing.T) {
		options := &ModuleInstallOptions{
			ModuleName: "Test::Module",
		}

		task := NewModuleInstallTask(options, 5)

		assert.Equal(t, "install-Test::Module", task.ID())
		assert.Equal(t, 5, task.Priority())
		assert.Nil(t, task.GetResult()) // No result before execution
	})
}

func TestInstallModulesParallel(t *testing.T) {
	t.Run("EmptyModuleList", func(t *testing.T) {
		options := &ParallelInstallOptions{
			Modules: []*ModuleInstallOptions{},
		}

		result, err := InstallModulesParallel(options)
		require.NoError(t, err)
		assert.Equal(t, 0, len(result.Results))
		assert.Equal(t, 0, len(result.Failures))
		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 0, result.FailureCount)
	})

	t.Run("NilOptions", func(t *testing.T) {
		result, err := InstallModulesParallel(nil)
		require.NoError(t, err)
		assert.Equal(t, 0, len(result.Results))
		assert.Equal(t, 0, len(result.Failures))
	})

	t.Run("WorkerCountConfiguration", func(t *testing.T) {
		modules := []*ModuleInstallOptions{
			{ModuleName: "Module1"},
			{ModuleName: "Module2"},
		}

		options := &ParallelInstallOptions{
			Modules:     modules,
			Workers:     2,
			StopOnError: false,
			Timeout:     5 * time.Second,
		}

		// This test mainly verifies the configuration is accepted
		// Actual installation would require mocking the InstallModule function
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		options.Context = ctx

		// Note: This may fail due to actual module installation attempts
		// In a real test environment, we'd mock InstallModule
		_, err := InstallModulesParallel(options)
		// We expect this to fail since we don't have real modules to install
		// but it should not fail due to configuration issues
		assert.Error(t, err) // Expected since we can't actually install modules in test
	})
}

func TestInstallModulesBatch(t *testing.T) {
	t.Run("EmptyModuleNames", func(t *testing.T) {
		result, err := InstallModulesBatch([]string{}, nil)
		require.NoError(t, err)
		assert.Equal(t, 0, len(result.Results))
	})

	t.Run("ModuleNameConversion", func(t *testing.T) {
		moduleNames := []string{"Module1", "Module2", "Module3"}

		options := &ModuleInstallOptions{
			PerlPath: "/usr/bin/perl",
			RunTests: false,
			Force:    true,
		}

		// This will fail due to actual installation attempts, but we can verify
		// the setup doesn't crash
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		options.Context = ctx

		_, err := InstallModulesBatch(moduleNames, options)
		// Expected to fail since we can't install real modules, but should not panic
		assert.Error(t, err)
	})
}

func TestCountCompletedTasks(t *testing.T) {
	t.Run("NoCompletedTasks", func(t *testing.T) {
		tasks := []*ModuleInstallTask{
			NewModuleInstallTask(&ModuleInstallOptions{ModuleName: "Module1"}, 1),
			NewModuleInstallTask(&ModuleInstallOptions{ModuleName: "Module2"}, 2),
		}

		completed := countCompletedTasks(tasks)
		assert.Equal(t, 0, completed)
	})

	t.Run("SomeCompletedTasks", func(t *testing.T) {
		tasks := []*ModuleInstallTask{
			NewModuleInstallTask(&ModuleInstallOptions{ModuleName: "Module1"}, 1),
			NewModuleInstallTask(&ModuleInstallOptions{ModuleName: "Module2"}, 2),
		}

		// Simulate one task completion by setting a result
		tasks[0].result = &ModuleInstallResult{
			ModuleName: "Module1",
			Success:    true,
		}

		completed := countCompletedTasks(tasks)
		assert.Equal(t, 1, completed)
	})
}

func TestModuleInstallFailure(t *testing.T) {
	t.Run("FailureStructure", func(t *testing.T) {
		failure := ModuleInstallFailure{
			ModuleName: "Failed::Module",
			Error:      assert.AnError,
			Duration:   5 * time.Second,
		}

		assert.Equal(t, "Failed::Module", failure.ModuleName)
		assert.Equal(t, assert.AnError, failure.Error)
		assert.Equal(t, 5*time.Second, failure.Duration)
	})
}

func TestParallelInstallResult(t *testing.T) {
	t.Run("ResultCalculations", func(t *testing.T) {
		result := &ParallelInstallResult{
			Results: []*ModuleInstallResult{
				{ModuleName: "Module1", Success: true},
				{ModuleName: "Module2", Success: false},
				{ModuleName: "Module3", Success: true},
			},
			Failures: []ModuleInstallFailure{
				{ModuleName: "Module2"},
			},
			Duration:          10 * time.Second,
			SuccessCount:      2,
			FailureCount:      1,
			InstallationOrder: []string{"Module1", "Module3", "Module2"},
		}

		assert.Equal(t, 3, len(result.Results))
		assert.Equal(t, 1, len(result.Failures))
		assert.Equal(t, 2, result.SuccessCount)
		assert.Equal(t, 1, result.FailureCount)
		assert.Equal(t, 10*time.Second, result.Duration)
		assert.Equal(t, []string{"Module1", "Module3", "Module2"}, result.InstallationOrder)
	})
}
