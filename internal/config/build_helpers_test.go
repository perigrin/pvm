// ABOUTME: Tests for build configuration helper functions
// ABOUTME: Validates project-aware build configuration access

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetProjectBuildConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm_build_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Non-project directory (should return defaults)
	buildConfig, err := GetProjectBuildConfig(tempDir)
	if err != nil {
		t.Fatalf("GetProjectBuildConfig failed: %v", err)
	}

	if buildConfig.Mode != "distribution" {
		t.Errorf("Expected default mode 'distribution', got %s", buildConfig.Mode)
	}

	// Test 2: Project directory with custom config
	projectDir := filepath.Join(tempDir, "project")
	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create .perl-version to make it a project
	perlVersionPath := filepath.Join(projectDir, ".perl-version")
	err = os.WriteFile(perlVersionPath, []byte("5.36.0\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create .perl-version: %v", err)
	}

	// Create project config with custom build settings
	configDir := filepath.Join(projectDir, ".pvm")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configContent := `
[build]
mode = "inline"
output_dir = "custom_build"
clean_before_build = false

[build.typecheck]
strict = true
target_perl = "5.38"
`
	configPath := filepath.Join(configDir, "pvm.toml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test with project directory
	buildConfig, err = GetProjectBuildConfig(projectDir)
	if err != nil {
		t.Fatalf("GetProjectBuildConfig failed: %v", err)
	}

	// Check that custom settings are applied
	if buildConfig.Mode != "inline" {
		t.Errorf("Expected mode 'inline', got %s", buildConfig.Mode)
	}

	// Output directory should be made absolute relative to project root
	expectedOutputDir := filepath.Join(projectDir, "custom_build")
	if buildConfig.OutputDir != expectedOutputDir {
		t.Errorf("Expected output dir %s, got %s", expectedOutputDir, buildConfig.OutputDir)
	}

	if buildConfig.CleanBeforeBuild {
		t.Errorf("Expected CleanBeforeBuild to be false")
	}

	if buildConfig.TypeCheck == nil {
		t.Fatal("TypeCheck config is nil")
	}

	if !buildConfig.TypeCheck.Strict {
		t.Errorf("Expected TypeCheck.Strict to be true")
	}

	if buildConfig.TypeCheck.TargetPerl != "5.38" {
		t.Errorf("Expected TargetPerl '5.38', got %s", buildConfig.TypeCheck.TargetPerl)
	}
}

func TestBuildConfigHelpers(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm_build_helpers_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test helper functions with non-project directory
	mode, err := GetBuildMode(tempDir)
	if err != nil {
		t.Fatalf("GetBuildMode failed: %v", err)
	}
	if mode != "distribution" {
		t.Errorf("Expected default mode 'distribution', got %s", mode)
	}

	outputDir, err := GetBuildOutputDir(tempDir)
	if err != nil {
		t.Fatalf("GetBuildOutputDir failed: %v", err)
	}
	if outputDir != "build" {
		t.Errorf("Expected default output dir 'build', got %s", outputDir)
	}

	includePatterns, err := GetBuildIncludePatterns(tempDir)
	if err != nil {
		t.Fatalf("GetBuildIncludePatterns failed: %v", err)
	}
	if len(includePatterns) != 1 || includePatterns[0] != "lib/**/*.pm" {
		t.Errorf("Expected default include patterns, got %v", includePatterns)
	}

	excludePatterns, err := GetBuildExcludePatterns(tempDir)
	if err != nil {
		t.Fatalf("GetBuildExcludePatterns failed: %v", err)
	}
	expectedExcludes := []string{"local/**", "build/**", "**/.git/**"}
	if len(excludePatterns) != len(expectedExcludes) {
		t.Errorf("Expected %d exclude patterns, got %d", len(expectedExcludes), len(excludePatterns))
	}

	watchDirs, err := GetBuildWatchDirs(tempDir)
	if err != nil {
		t.Fatalf("GetBuildWatchDirs failed: %v", err)
	}
	expectedWatchDirs := []string{"lib", "script", "t"}
	if len(watchDirs) != len(expectedWatchDirs) {
		t.Errorf("Expected %d watch dirs, got %d", len(expectedWatchDirs), len(watchDirs))
	}

	typeCheckConfig, err := GetTypeCheckConfig(tempDir)
	if err != nil {
		t.Fatalf("GetTypeCheckConfig failed: %v", err)
	}
	if typeCheckConfig.Strict {
		t.Errorf("Expected default strict to be false")
	}
	if typeCheckConfig.TargetPerl != "5.36" {
		t.Errorf("Expected default target perl '5.36', got %s", typeCheckConfig.TargetPerl)
	}

	distConfig, err := GetDistributionConfig(tempDir)
	if err != nil {
		t.Fatalf("GetDistributionConfig failed: %v", err)
	}
	if !distConfig.IncludeTests {
		t.Errorf("Expected default IncludeTests to be true")
	}
	if distConfig.Installer != "ExtUtils::MakeMaker" {
		t.Errorf("Expected default installer 'ExtUtils::MakeMaker', got %s", distConfig.Installer)
	}

	shouldClean, err := ShouldCleanBeforeBuild(tempDir)
	if err != nil {
		t.Fatalf("ShouldCleanBeforeBuild failed: %v", err)
	}
	if !shouldClean {
		t.Errorf("Expected default CleanBeforeBuild to be true")
	}
}
