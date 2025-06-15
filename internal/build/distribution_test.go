// ABOUTME: Tests for distribution build functionality
// ABOUTME: Comprehensive test coverage for CPAN distribution generation

package build

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/project"
)

func TestNewDistributionBuilder(t *testing.T) {
	// Test with nil project context
	builder, err := NewDistributionBuilder(nil)
	if err != nil {
		t.Fatalf("Failed to create distribution builder with nil context: %v", err)
	}
	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}

	// Test with project context
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   "/tmp/test-project",
	}
	builder, err = NewDistributionBuilder(projectCtx)
	if err != nil {
		t.Fatalf("Failed to create distribution builder with project context: %v", err)
	}
	if builder.projectCtx != projectCtx {
		t.Error("Project context not set correctly")
	}
}

func TestDistributionBuilder_Build(t *testing.T) {
	// Create temporary project directory
	tempDir, err := os.MkdirTemp("", "pvm-dist-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up test project structure
	projectCtx := setupTestProject(t, tempDir)

	builder, err := NewDistributionBuilder(projectCtx)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	ctx := context.Background()
	options := &BuildOptions{
		CleanFirst:     true,
		SkipTypeCheck:  true, // Skip type checking for this test
		IncludeTests:   true,
		IncludeScripts: true,
	}

	result, err := builder.Build(ctx, options)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Build was not successful. Errors: %v", result.BuildErrors)
	}

	if result.FilesProcessed == 0 {
		t.Error("No files were processed")
	}

	if !result.MetadataGenerated {
		t.Error("Metadata was not generated")
	}

	// Verify build directory exists
	buildDir := result.BuildDir
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		t.Errorf("Build directory does not exist: %s", buildDir)
	}

	// Verify required files exist
	requiredFiles := []string{
		"META.json",
		"META.yml",
		"Makefile.PL",
		"MANIFEST",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(buildDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Required file missing: %s", file)
		}
	}

	// Verify directory structure
	expectedDirs := []string{
		"lib",
		"t",
		"script",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(buildDir, dir)
		if info, err := os.Stat(dirPath); os.IsNotExist(err) || !info.IsDir() {
			t.Errorf("Expected directory missing: %s", dir)
		}
	}
}

func TestDistributionBuilder_generateMETAJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-meta-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectCtx := setupTestProject(t, tempDir)
	builder, err := NewDistributionBuilder(projectCtx)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	buildDir := filepath.Join(tempDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("Failed to create build directory: %v", err)
	}

	err = builder.generateMETAJSON(buildDir)
	if err != nil {
		t.Fatalf("Failed to generate META.json: %v", err)
	}

	// Verify META.json exists and is valid
	metaPath := filepath.Join(buildDir, "META.json")
	content, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read META.json: %v", err)
	}

	var metaData map[string]interface{}
	err = json.Unmarshal(content, &metaData)
	if err != nil {
		t.Fatalf("META.json is not valid JSON: %v", err)
	}

	// Check required fields
	requiredFields := []string{"name", "version", "abstract", "author", "license"}
	for _, field := range requiredFields {
		if _, exists := metaData[field]; !exists {
			t.Errorf("Required field missing from META.json: %s", field)
		}
	}

	// Check meta-spec
	if metaSpec, exists := metaData["meta-spec"]; exists {
		if spec, ok := metaSpec.(map[string]interface{}); ok {
			if version, exists := spec["version"]; !exists || version != "2" {
				t.Error("meta-spec version should be '2'")
			}
		} else {
			t.Error("meta-spec should be an object")
		}
	} else {
		t.Error("meta-spec field is required")
	}
}

func TestDistributionBuilder_generateMETAYML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-meta-yml-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectCtx := setupTestProject(t, tempDir)
	builder, err := NewDistributionBuilder(projectCtx)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	buildDir := filepath.Join(tempDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("Failed to create build directory: %v", err)
	}

	err = builder.generateMETAYML(buildDir)
	if err != nil {
		t.Fatalf("Failed to generate META.yml: %v", err)
	}

	// Verify META.yml exists
	metaPath := filepath.Join(buildDir, "META.yml")
	content, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read META.yml: %v", err)
	}

	contentStr := string(content)

	// Check for required fields
	requiredFields := []string{"abstract:", "author:", "name:", "version:", "license:"}
	for _, field := range requiredFields {
		if !strings.Contains(contentStr, field) {
			t.Errorf("Required field missing from META.yml: %s", field)
		}
	}

	// Should start with YAML document separator
	if !strings.HasPrefix(contentStr, "---") {
		t.Error("META.yml should start with '---'")
	}
}

func TestDistributionBuilder_generateMakefilePL(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-makefile-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectCtx := setupTestProject(t, tempDir)
	builder, err := NewDistributionBuilder(projectCtx)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	buildDir := filepath.Join(tempDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("Failed to create build directory: %v", err)
	}

	err = builder.generateMakefilePL(buildDir)
	if err != nil {
		t.Fatalf("Failed to generate Makefile.PL: %v", err)
	}

	// Verify Makefile.PL exists
	makefilePath := filepath.Join(buildDir, "Makefile.PL")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("Failed to read Makefile.PL: %v", err)
	}

	contentStr := string(content)

	// Check for required components
	required := []string{
		"use ExtUtils::MakeMaker",
		"WriteMakefile(",
		"NAME",
		"VERSION",
		"ABSTRACT",
		"AUTHOR",
		"LICENSE",
	}

	for _, req := range required {
		if !strings.Contains(contentStr, req) {
			t.Errorf("Required component missing from Makefile.PL: %s", req)
		}
	}
}

func TestDistributionBuilder_generateMANIFEST(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-manifest-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	buildDir := filepath.Join(tempDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("Failed to create build directory: %v", err)
	}

	// Create some test files
	testFiles := []string{
		"lib/Test/Module.pm",
		"t/basic.t",
		"script/myscript.pl",
		"META.json",
		"Makefile.PL",
	}

	for _, file := range testFiles {
		filePath := filepath.Join(buildDir, file)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", file, err)
		}
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	projectCtx := &project.ProjectContext{IsProject: false}
	builder, err := NewDistributionBuilder(projectCtx)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	err = builder.generateMANIFEST(buildDir)
	if err != nil {
		t.Fatalf("Failed to generate MANIFEST: %v", err)
	}

	// Verify MANIFEST exists
	manifestPath := filepath.Join(buildDir, "MANIFEST")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read MANIFEST: %v", err)
	}

	contentStr := string(content)
	lines := strings.Split(strings.TrimSpace(contentStr), "\n")

	// Should contain all test files plus MANIFEST itself
	expectedCount := len(testFiles) + 1
	if len(lines) != expectedCount {
		t.Errorf("Expected %d files in MANIFEST, got %d", expectedCount, len(lines))
	}

	// Check that all test files are listed
	for _, file := range testFiles {
		if !strings.Contains(contentStr, file) {
			t.Errorf("File missing from MANIFEST: %s", file)
		}
	}

	// MANIFEST should list itself
	if !strings.Contains(contentStr, "MANIFEST") {
		t.Error("MANIFEST should list itself")
	}
}

func TestDistributionBuilder_extractModuleName(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-module-name-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create lib directory with a module
	libDir := filepath.Join(tempDir, "lib")
	moduleDir := filepath.Join(libDir, "My", "Test")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatalf("Failed to create module directory: %v", err)
	}

	modulePath := filepath.Join(moduleDir, "Module.pm")
	moduleContent := `package My::Test::Module;
use v5.36;

our $VERSION = '1.23';

1;
`
	if err := os.WriteFile(modulePath, []byte(moduleContent), 0644); err != nil {
		t.Fatalf("Failed to create module file: %v", err)
	}

	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   tempDir,
	}

	builder, err := NewDistributionBuilder(projectCtx)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	name, err := builder.extractModuleName()
	if err != nil {
		t.Fatalf("Failed to extract module name: %v", err)
	}

	expected := "My-Test-Module"
	if name != expected {
		t.Errorf("Expected module name %s, got %s", expected, name)
	}
}

func TestDistributionBuilder_extractModuleVersion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-version-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "simple version",
			content:  "our $VERSION = '1.23';\n",
			expected: "1.23",
		},
		{
			name:     "version with quotes",
			content:  `our $VERSION = "0.01";`,
			expected: "0.01",
		},
		{
			name:     "version declaration",
			content:  "use version; our $VERSION = version->declare('v1.2.3');\n",
			expected: "v1.2.3",
		},
		{
			name:     "no version",
			content:  "package My::Module;\n1;\n",
			expected: "undef",
		},
	}

	builder, err := NewDistributionBuilder(nil)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tempDir, "test.pm")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			version, err := builder.extractModuleVersion(testFile)
			if err != nil && tt.expected != "undef" {
				t.Fatalf("Failed to extract version: %v", err)
			}

			if version != tt.expected {
				t.Errorf("Expected version %s, got %s", tt.expected, version)
			}
		})
	}
}

func TestDistributionBuilder_Validate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-validate-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	builder, err := NewDistributionBuilder(nil)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	buildDir := filepath.Join(tempDir, "build")

	// Test validation of non-existent directory
	err = builder.Validate(buildDir)
	if err == nil {
		t.Error("Expected error for non-existent build directory")
	}

	// Create build directory
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("Failed to create build directory: %v", err)
	}

	// Test validation without required files
	err = builder.Validate(buildDir)
	if err == nil {
		t.Error("Expected error for missing required files")
	}

	// Create required files
	requiredFiles := []string{
		"META.json",
		"META.yml",
		"Makefile.PL",
		"MANIFEST",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(buildDir, file)
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create required file %s: %v", file, err)
		}
	}

	// Test validation without lib directory
	err = builder.Validate(buildDir)
	if err == nil {
		t.Error("Expected error for missing lib directory")
	}

	// Create lib directory but no modules
	libDir := filepath.Join(buildDir, "lib")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	err = builder.Validate(buildDir)
	if err == nil {
		t.Error("Expected error for lib directory with no modules")
	}

	// Create a module
	moduleFile := filepath.Join(libDir, "Test.pm")
	if err := os.WriteFile(moduleFile, []byte("package Test; 1;"), 0644); err != nil {
		t.Fatalf("Failed to create module file: %v", err)
	}

	// Now validation should pass
	err = builder.Validate(buildDir)
	if err != nil {
		t.Errorf("Validation failed unexpectedly: %v", err)
	}
}

func TestDistributionBuilder_Clean(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-clean-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	buildDir := filepath.Join(tempDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("Failed to create build directory: %v", err)
	}

	// Create some files in build directory
	testFile := filepath.Join(buildDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	builder, err := NewDistributionBuilder(nil)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	// Clean the build directory
	err = builder.Clean(buildDir)
	if err != nil {
		t.Fatalf("Failed to clean build directory: %v", err)
	}

	// Verify build directory is gone
	if _, err := os.Stat(buildDir); !os.IsNotExist(err) {
		t.Error("Build directory should have been removed")
	}
}

// Helper function to set up a test project
func setupTestProject(t *testing.T, tempDir string) *project.ProjectContext {
	// Create project structure
	dirs := []string{
		"lib/My/Test",
		"t",
		"script",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test module
	moduleContent := `package My::Test::Module;
use v5.36;

our $VERSION = '1.23';

sub new {
    my $class = shift;
    return bless {}, $class;
}

1;
`
	modulePath := filepath.Join(tempDir, "lib", "My", "Test", "Module.pm")
	if err := os.WriteFile(modulePath, []byte(moduleContent), 0644); err != nil {
		t.Fatalf("Failed to create module file: %v", err)
	}

	// Create test file
	testContent := `#!/usr/bin/env perl
use v5.36;
use Test::More tests => 1;

BEGIN { use_ok('My::Test::Module') };
`
	testPath := filepath.Join(tempDir, "t", "basic.t")
	if err := os.WriteFile(testPath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create script file
	scriptContent := `#!/usr/bin/env perl
use v5.36;
use My::Test::Module;

my $obj = My::Test::Module->new();
print "Hello World\n";
`
	scriptPath := filepath.Join(tempDir, "script", "myscript.pl")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to create script file: %v", err)
	}

	// Create cpanfile
	cpanfileContent := `requires 'Test::More', '0';

on 'develop' => sub {
    requires 'Test::Pod', '1.22';
};
`
	cpanfilePath := filepath.Join(tempDir, "cpanfile")
	if err := os.WriteFile(cpanfilePath, []byte(cpanfileContent), 0644); err != nil {
		t.Fatalf("Failed to create cpanfile: %v", err)
	}

	return &project.ProjectContext{
		IsProject:   true,
		RootDir:     tempDir,
		ConfigFile:  filepath.Join(tempDir, "pvm.toml"),
		LocalLibDir: filepath.Join(tempDir, "lib"),
	}
}

func TestDistributionBuildOptions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-options-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectCtx := setupTestProject(t, tempDir)
	builder, err := NewDistributionBuilder(projectCtx)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	ctx := context.Background()

	// Test with custom output directory
	customBuildDir := filepath.Join(tempDir, "custom-build")
	options := &BuildOptions{
		OutputDir:     customBuildDir,
		CleanFirst:    true,
		SkipTypeCheck: true,
	}

	result, err := builder.Build(ctx, options)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if result.BuildDir != customBuildDir {
		t.Errorf("Expected build dir %s, got %s", customBuildDir, result.BuildDir)
	}

	// Verify custom build directory was created
	if _, err := os.Stat(customBuildDir); os.IsNotExist(err) {
		t.Error("Custom build directory was not created")
	}
}

func TestDistributionBuilder_BuildWithErrors(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-error-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create minimal project without proper structure
	projectCtx := &project.ProjectContext{
		IsProject: true,
		RootDir:   tempDir,
	}

	builder, err := NewDistributionBuilder(projectCtx)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	ctx := context.Background()
	options := &BuildOptions{
		SkipTypeCheck: true,
	}

	result, err := builder.Build(ctx, options)
	// Build should succeed even with minimal structure
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Should still generate metadata even without modules
	if !result.MetadataGenerated {
		t.Error("Metadata should be generated even for empty project")
	}
}

func TestDistributionBuilder_ContextCancellation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-cancel-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectCtx := setupTestProject(t, tempDir)
	builder, err := NewDistributionBuilder(projectCtx)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	options := &BuildOptions{
		SkipTypeCheck: true,
	}

	_, err = builder.Build(ctx, options)
	if err == nil {
		t.Error("Expected error due to cancelled context")
	}

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

func TestDistributionBuilder_BuildDuration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-duration-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectCtx := setupTestProject(t, tempDir)
	builder, err := NewDistributionBuilder(projectCtx)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	ctx := context.Background()
	options := &BuildOptions{
		SkipTypeCheck: true,
	}

	start := time.Now()
	result, err := builder.Build(ctx, options)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	elapsed := time.Since(start)

	// Duration should be reasonable and match actual elapsed time
	if result.Duration <= 0 {
		t.Error("Build duration should be positive")
	}

	if result.Duration > elapsed+time.Millisecond*100 {
		t.Errorf("Reported duration %v seems too long compared to actual %v", result.Duration, elapsed)
	}
}
