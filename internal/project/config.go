// ABOUTME: Integration between project detection and configuration system
// ABOUTME: Provides project-aware configuration loading and path resolution

package project

import (
	"path/filepath"
)

// GetProjectAwareConfigPath returns the appropriate config file path based on project context
func GetProjectAwareConfigPath(workingDir string) (string, error) {
	ctx, err := DetectProject(workingDir)
	if err != nil {
		return "", err
	}

	if ctx.IsProject && ctx.ConfigFile != "" {
		return ctx.ConfigFile, nil
	}

	// If no project config found, check for .pvm/pvm.toml in project root
	if ctx.IsProject {
		pvmConfigPath := filepath.Join(ctx.RootDir, ".pvm", "pvm.toml")
		return pvmConfigPath, nil
	}

	// Not in a project, no config path
	return "", nil
}

// GetProjectAwareLibPath returns the appropriate local lib path based on project context
func GetProjectAwareLibPath(workingDir string) (string, error) {
	ctx, err := DetectProject(workingDir)
	if err != nil {
		return "", err
	}

	if ctx.IsProject {
		return ctx.LocalLibDir, nil
	}

	// Not in a project, return empty
	return "", nil
}

// IsInProject checks if the given directory is within a project
func IsInProject(workingDir string) (bool, error) {
	ctx, err := DetectProject(workingDir)
	if err != nil {
		return false, err
	}
	return ctx.IsProject, nil
}

// GetProjectRoot returns the root directory of the project containing the given directory
func GetProjectRoot(workingDir string) (string, error) {
	ctx, err := DetectProject(workingDir)
	if err != nil {
		return "", err
	}

	if ctx.IsProject {
		return ctx.RootDir, nil
	}

	return "", nil
}

// GetProjectPerlVersion returns the Perl version specified in the project
func GetProjectPerlVersion(workingDir string) (string, error) {
	ctx, err := DetectProject(workingDir)
	if err != nil {
		return "", err
	}

	return ctx.PerlVersion, nil
}
