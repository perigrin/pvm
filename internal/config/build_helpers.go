// ABOUTME: Helper functions for build configuration access
// ABOUTME: Provides convenient access to build settings with project-aware defaults

package config

import (
	"os"
	"path/filepath"

	"tamarou.com/pvm/internal/project"
)

// GetEffectiveBuildConfig returns build configuration with project context awareness
func GetEffectiveBuildConfig() (*BuildConfig, error) {
	config, err := LoadEffectiveConfig()
	if err != nil {
		return nil, err
	}

	if config.Build == nil {
		return NewDefaultConfig().Build, nil
	}

	return config.Build, nil
}

// GetProjectBuildConfig returns build configuration specific to a project directory
func GetProjectBuildConfig(workingDir string) (*BuildConfig, error) {
	// Check if we're in a project
	ctx, err := project.DetectProject(workingDir)
	if err != nil {
		return nil, err
	}

	if !ctx.IsProject {
		// Not in a project, return defaults
		return NewDefaultConfig().Build, nil
	}

	// Load project-specific configuration directly
	var projectConfig *Config
	if ctx.ConfigFile != "" {
		// Load from the detected config file
		projectConfig, err = ParseFile(ctx.ConfigFile)
		if err != nil {
			return nil, err
		}
	} else {
		// Check for .pvm/pvm.toml in project root
		pvmConfigPath := filepath.Join(ctx.RootDir, ".pvm", "pvm.toml")
		if _, err := os.Stat(pvmConfigPath); err == nil {
			projectConfig, err = ParseFile(pvmConfigPath)
			if err != nil {
				return nil, err
			}
		}
	}

	// Start with defaults and merge project config
	result := NewDefaultConfig()
	if projectConfig != nil && projectConfig.Build != nil {
		mergeBuildConfig(result.Build, projectConfig.Build)
	}

	// Ensure output directory is relative to project root
	if !filepath.IsAbs(result.Build.OutputDir) {
		result.Build.OutputDir = filepath.Join(ctx.RootDir, result.Build.OutputDir)
	}

	return result.Build, nil
}

// GetBuildOutputDir returns the effective build output directory for a project
func GetBuildOutputDir(workingDir string) (string, error) {
	buildConfig, err := GetProjectBuildConfig(workingDir)
	if err != nil {
		return "", err
	}

	return buildConfig.OutputDir, nil
}

// GetBuildMode returns the effective build mode for a project
func GetBuildMode(workingDir string) (string, error) {
	buildConfig, err := GetProjectBuildConfig(workingDir)
	if err != nil {
		return "", err
	}

	return buildConfig.Mode, nil
}

// GetBuildIncludePatterns returns the file patterns to include in builds
func GetBuildIncludePatterns(workingDir string) ([]string, error) {
	buildConfig, err := GetProjectBuildConfig(workingDir)
	if err != nil {
		return nil, err
	}

	if buildConfig.Files == nil || len(buildConfig.Files.Include) == 0 {
		return []string{"lib/**/*.pm"}, nil
	}

	return buildConfig.Files.Include, nil
}

// GetBuildExcludePatterns returns the file patterns to exclude from builds
func GetBuildExcludePatterns(workingDir string) ([]string, error) {
	buildConfig, err := GetProjectBuildConfig(workingDir)
	if err != nil {
		return nil, err
	}

	if buildConfig.Files == nil || len(buildConfig.Files.Exclude) == 0 {
		return []string{"local/**", "build/**", "**/.git/**"}, nil
	}

	return buildConfig.Files.Exclude, nil
}

// GetBuildWatchDirs returns the directories to watch in watch mode
func GetBuildWatchDirs(workingDir string) ([]string, error) {
	buildConfig, err := GetProjectBuildConfig(workingDir)
	if err != nil {
		return nil, err
	}

	if buildConfig.Files == nil || len(buildConfig.Files.WatchDirs) == 0 {
		return []string{"lib", "script", "t"}, nil
	}

	return buildConfig.Files.WatchDirs, nil
}

// GetTypeCheckConfig returns the type checking configuration
func GetTypeCheckConfig(workingDir string) (*BuildTypeCheckConfig, error) {
	buildConfig, err := GetProjectBuildConfig(workingDir)
	if err != nil {
		return nil, err
	}

	if buildConfig.TypeCheck == nil {
		return &BuildTypeCheckConfig{
			Strict:       false,
			Experimental: false,
			TargetPerl:   "5.36",
		}, nil
	}

	return buildConfig.TypeCheck, nil
}

// GetDistributionConfig returns the distribution build configuration
func GetDistributionConfig(workingDir string) (*BuildDistributionConfig, error) {
	buildConfig, err := GetProjectBuildConfig(workingDir)
	if err != nil {
		return nil, err
	}

	if buildConfig.Distribution == nil {
		return &BuildDistributionConfig{
			IncludeTests:   true,
			IncludeScripts: true,
			Installer:      "ExtUtils::MakeMaker",
		}, nil
	}

	return buildConfig.Distribution, nil
}

// ShouldCleanBeforeBuild returns whether to clean output before building
func ShouldCleanBeforeBuild(workingDir string) (bool, error) {
	buildConfig, err := GetProjectBuildConfig(workingDir)
	if err != nil {
		return false, err
	}

	return buildConfig.CleanBeforeBuild, nil
}

