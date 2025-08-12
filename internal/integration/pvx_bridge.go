// ABOUTME: Bridge layer between PVX options and workflow orchestration
// ABOUTME: Converts PVX ExecutionOptions to WorkflowOptions for seamless integration

package integration

import (
	"tamarou.com/pvm/internal/pvx"
)

// PVXToWorkflowOptions converts PVX ExecutionOptions to WorkflowOptions
// This enables PVX commands to use the workflow orchestration system
func PVXToWorkflowOptions(pvxOptions *pvx.ExecutionOptions) *WorkflowOptions {
	workflowOptions := &WorkflowOptions{
		ScriptPath:      pvxOptions.ScriptPath,
		PerlVersion:     pvxOptions.PerlVersion,
		Verbose:         pvxOptions.Verbose,
		SkipTypeCheck:   !pvxOptions.TypeCheck,
		SkipExecution:   false, // PVX always executes
		IsolationLevel:  pvxOptions.IsolationLevel,
		
		// PVX-specific options
		EnvironmentName: pvxOptions.EnvName,
		ModulePaths:     pvxOptions.AdditionalModulePaths,
		ExecuteCode:     pvxOptions.InlineCode,
		AutoDetectDeps:  pvxOptions.AutoDetectDependencies,
		ForceVersion:    pvxOptions.ForceVersion,
		NoInstall:       !pvxOptions.AutoInstallModules,
		ReadOnlyPaths:   pvxOptions.ReadOnlyPaths,
		ReadWritePaths:  pvxOptions.ReadWritePaths,
		IsolatedOutput:  pvxOptions.IsolatedOutput,
		SaveOutputDir:   pvxOptions.SaveOutputDir,
		NoCleanup:       pvxOptions.NoCleanup,
		AdditionalArgs:  pvxOptions.Args,
	}
	
	// Handle environment variables
	if len(pvxOptions.PreserveEnv) > 0 {
		workflowOptions.PreserveEnv = pvxOptions.PreserveEnv
	}
	if len(pvxOptions.ClearEnv) > 0 {
		workflowOptions.ClearEnv = pvxOptions.ClearEnv
	}
	
	// Handle required modules
	if len(pvxOptions.RequiredModules) > 0 {
		workflowOptions.RequiredModules = pvxOptions.RequiredModules
	}
	
	// Set features flag if enabled
	if pvxOptions.EnableFeatures {
		workflowOptions.ExecuteFeatures = pvxOptions.InlineCode
		workflowOptions.ExecuteCode = "" // Clear regular code when features are enabled
	}
	
	return workflowOptions
}

// WorkflowToPVXResult converts WorkflowResult back to PVX-compatible format  
// This allows PVX commands to receive workflow results in expected format
func WorkflowToPVXResult(workflowResult *WorkflowResult) *pvx.ExecuteResult {
	return &pvx.ExecuteResult{
		Output:   workflowResult.ExecutionOutput,
		ExitCode: workflowResult.ExecutionExitCode,
	}
}

// CreateWorkflowFromPVXFlags creates WorkflowOptions from individual PVX command flags
// This is useful for direct integration with cobra command flags
func CreateWorkflowFromPVXFlags(
	scriptPath, perlVersion, envName, executeCode, executeFeaturesCode string,
	typeCheck, verbose, forceVersion, autoInstall, autoDetectDeps, noCleanup bool,
	isolationLevel pvx.IsolationLevel,
	modulePaths, preserveEnv, clearEnv, readOnlyPaths, readWritePaths []string,
	args []string,
) *WorkflowOptions {
	return &WorkflowOptions{
		ScriptPath:       scriptPath,
		PerlVersion:      perlVersion,
		Verbose:          verbose,
		SkipTypeCheck:    !typeCheck,
		SkipExecution:    false,
		IsolationLevel:   isolationLevel,
		EnvironmentName:  envName,
		PreserveEnv:      preserveEnv,
		ClearEnv:         clearEnv,
		ModulePaths:      modulePaths,
		ExecuteCode:      executeCode,
		ExecuteFeatures:  executeFeaturesCode,
		AutoDetectDeps:   autoDetectDeps,
		ForceVersion:     forceVersion,
		NoInstall:        !autoInstall,
		ReadOnlyPaths:    readOnlyPaths,
		ReadWritePaths:   readWritePaths,
		NoCleanup:        noCleanup,
		AdditionalArgs:   args,
	}
}