// ABOUTME: Tests for build configuration validation and merging
// ABOUTME: Validates the new build system configuration functionality

package config

import (
	"testing"
)

func TestBuildConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *BuildConfig
		wantError bool
	}{
		{
			name: "valid_configuration",
			config: &BuildConfig{
				Mode:             "distribution",
				OutputDir:        "build",
				CleanBeforeBuild: true,
				TypeCheck: &BuildTypeCheckConfig{
					Strict:       false,
					Experimental: false,
					TargetPerl:   "5.36",
				},
				Files: &BuildFilesConfig{
					Include:   []string{"lib/**/*.pm"},
					Exclude:   []string{"local/**"},
					WatchDirs: []string{"lib", "t"},
				},
				Distribution: &BuildDistributionConfig{
					IncludeTests:   true,
					IncludeScripts: true,
					Installer:      "ExtUtils::MakeMaker",
				},
			},
			wantError: false,
		},
		{
			name: "invalid_build_mode",
			config: &BuildConfig{
				Mode:      "invalid",
				OutputDir: "build",
			},
			wantError: true,
		},
		{
			name: "empty_output_dir",
			config: &BuildConfig{
				Mode:      "distribution",
				OutputDir: "",
			},
			wantError: true,
		},
		{
			name: "invalid_installer",
			config: &BuildConfig{
				Mode:      "distribution",
				OutputDir: "build",
				Distribution: &BuildDistributionConfig{
					Installer: "InvalidInstaller",
				},
			},
			wantError: true,
		},
		{
			name: "invalid_target_perl",
			config: &BuildConfig{
				Mode:      "distribution",
				OutputDir: "build",
				TypeCheck: &BuildTypeCheckConfig{
					TargetPerl: "invalid",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Validate()
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("BuildConfig.Validate() error = %v, wantError %v", errors, tt.wantError)
			}
		})
	}
}

func TestProjectConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *ProjectConfig
		wantError bool
	}{
		{
			name: "valid_configuration",
			config: &ProjectConfig{
				Name:        "TestProject",
				Version:     "1.0.0",
				PerlVersion: "5.36",
				License:     "perl_5",
			},
			wantError: false,
		},
		{
			name: "invalid_perl_version",
			config: &ProjectConfig{
				PerlVersion: "invalid",
			},
			wantError: true,
		},
		{
			name: "invalid_license",
			config: &ProjectConfig{
				License: "invalid_license",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Validate()
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("ProjectConfig.Validate() error = %v, wantError %v", errors, tt.wantError)
			}
		})
	}
}

func TestBuildConfigMerging(t *testing.T) {
	target := &BuildConfig{
		Mode:             "inline",
		OutputDir:        "build",
		CleanBeforeBuild: false,
		TypeCheck: &BuildTypeCheckConfig{
			Strict:     false,
			TargetPerl: "5.36",
		},
	}

	source := &BuildConfig{
		Mode:             "distribution",
		CleanBeforeBuild: true,
		TypeCheck: &BuildTypeCheckConfig{
			Strict: true,
		},
		Files: &BuildFilesConfig{
			Include: []string{"lib/**/*.pm", "script/**/*.pl"},
		},
	}

	mergeBuildConfig(target, source)

	// Check that source values override target values
	if target.Mode != "distribution" {
		t.Errorf("Expected Mode to be 'distribution', got %s", target.Mode)
	}
	if !target.CleanBeforeBuild {
		t.Errorf("Expected CleanBeforeBuild to be true")
	}
	if !target.TypeCheck.Strict {
		t.Errorf("Expected TypeCheck.Strict to be true")
	}
	if target.TypeCheck.TargetPerl != "5.36" {
		t.Errorf("Expected TypeCheck.TargetPerl to remain '5.36', got %s", target.TypeCheck.TargetPerl)
	}
	if target.Files == nil || len(target.Files.Include) != 2 {
		t.Errorf("Expected Files.Include to have 2 items")
	}
}

func TestBuildConfigFromTOML(t *testing.T) {
	tomlContent := `
[project]
name = "TestProject"
version = "1.0.0"
perl_version = "5.36"
license = "perl_5"

[build]
mode = "distribution"
output_dir = "dist"
clean_before_build = true

[build.typecheck]
strict = true
experimental = false
target_perl = "5.36"

[build.files]
include = ["lib/**/*.pm", "script/**/*.pl"]
exclude = ["local/**", "build/**"]
watch_dirs = ["lib", "script", "t"]

[build.distribution]
include_tests = true
include_scripts = true
installer = "ExtUtils::MakeMaker"
`

	config, err := ParseString(tomlContent)
	if err != nil {
		t.Fatalf("Failed to parse TOML: %v", err)
	}

	// Check Project configuration
	if config.Project == nil {
		t.Fatal("Project configuration is nil")
	}
	if config.Project.Name != "TestProject" {
		t.Errorf("Expected project name 'TestProject', got %s", config.Project.Name)
	}
	if config.Project.PerlVersion != "5.36" {
		t.Errorf("Expected perl version '5.36', got %s", config.Project.PerlVersion)
	}

	// Check Build configuration
	if config.Build == nil {
		t.Fatal("Build configuration is nil")
	}
	if config.Build.Mode != "distribution" {
		t.Errorf("Expected build mode 'distribution', got %s", config.Build.Mode)
	}
	if config.Build.OutputDir != "dist" {
		t.Errorf("Expected output dir 'dist', got %s", config.Build.OutputDir)
	}
	if !config.Build.CleanBeforeBuild {
		t.Errorf("Expected CleanBeforeBuild to be true")
	}

	// Check TypeCheck configuration
	if config.Build.TypeCheck == nil {
		t.Fatal("TypeCheck configuration is nil")
	}
	if !config.Build.TypeCheck.Strict {
		t.Errorf("Expected TypeCheck.Strict to be true")
	}

	// Check Files configuration
	if config.Build.Files == nil {
		t.Fatal("Files configuration is nil")
	}
	if len(config.Build.Files.Include) != 2 {
		t.Errorf("Expected 2 include patterns, got %d", len(config.Build.Files.Include))
	}

	// Check Distribution configuration
	if config.Build.Distribution == nil {
		t.Fatal("Distribution configuration is nil")
	}
	if config.Build.Distribution.Installer != "ExtUtils::MakeMaker" {
		t.Errorf("Expected installer 'ExtUtils::MakeMaker', got %s", config.Build.Distribution.Installer)
	}
}
