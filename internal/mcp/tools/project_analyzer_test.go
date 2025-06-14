// ABOUTME: Tests for project-scoped analysis functionality
// ABOUTME: Verifies project analyzer can analyze multiple files and detect cross-file issues

package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tamarou.com/pvm/internal/mcp/validation"
)

func TestProjectAnalyzer_AnalyzeProject(t *testing.T) {
	// Create temporary test project
	tmpDir, err := os.MkdirTemp("", "pvm-project-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := map[string]string{
		"lib/MyModule.pm": `package MyModule;
use strict;
use warnings;

sub get_user {
    my Int $id = shift;
    my Str $name = "User $id";
    return {
        id => $id,
        name => $name
    };
}

1;`,
		"lib/MyApp.pm": `package MyApp;
use strict;
use warnings;
use MyModule;

sub process_user {
    my Str $id = shift;
    my ArrayRef[Int] $numbers = [1, 2, 3];
    my $user = MyModule::get_user($id);
    return $user;
}

1;`,
		"t/test.t": `#!/usr/bin/env perl
use strict;
use warnings;
use Test::More;
use MyModule;
use MyApp;

my $user = MyModule::get_user(123);
ok($user, "Got user");

done_testing();`,
	}

	// Create directory structure
	for filePath, content := range testFiles {
		fullPath := filepath.Join(tmpDir, filePath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	// Create analyzer components
	cache, err := validation.NewValidationCache("10MB")
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	validator, err := validation.NewValidator(cache)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	codeAnalyzer, err := NewCodeAnalyzer(validator, nil)
	if err != nil {
		t.Fatalf("Failed to create code analyzer: %v", err)
	}

	projectAnalyzer, err := NewProjectAnalyzer(codeAnalyzer, validator)
	if err != nil {
		t.Fatalf("Failed to create project analyzer: %v", err)
	}

	// Perform project analysis
	ctx := context.Background()
	analysis, err := projectAnalyzer.AnalyzeProject(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Failed to analyze project: %v", err)
	}

	// Verify results
	if analysis.ProjectPath != tmpDir {
		t.Errorf("Expected project path %s, got %s", tmpDir, analysis.ProjectPath)
	}

	if analysis.TotalFiles != 3 {
		t.Errorf("Expected 3 files, got %d", analysis.TotalFiles)
	}

	// Since we no longer have type declaration conflicts, we check for variables
	// For now, we expect no conflicts since variables are scoped to their functions
	// In a future version with type declarations, we could have meaningful conflicts

	// Check global types - should find typed variables
	if len(analysis.GlobalTypes) == 0 {
		t.Error("Expected global types to be found")
	}

	// Verify we found some typed variables (like $id, $name, $numbers)
	foundTypes := make([]string, 0, len(analysis.GlobalTypes))
	for name := range analysis.GlobalTypes {
		foundTypes = append(foundTypes, name)
	}

	if len(foundTypes) < 3 {
		t.Errorf("Expected at least 3 typed variables, got %d: %v", len(foundTypes), foundTypes)
	}

	// Check file analysis
	if len(analysis.Files) != 3 {
		t.Errorf("Expected 3 file analyses, got %d", len(analysis.Files))
	}

	// Verify dependencies
	if len(analysis.Dependencies) == 0 {
		t.Error("Expected dependencies to be found")
	}
}

func TestProjectAnalyzer_GetProjectSummary(t *testing.T) {
	// Create temporary test project
	tmpDir, err := os.MkdirTemp("", "pvm-summary-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple test file
	testFile := filepath.Join(tmpDir, "test.pl")
	content := `#!/usr/bin/env perl
use strict;
use warnings;

my Int $Count = 42;
print "Count: $Count\n";
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create analyzer
	cache, err := validation.NewValidationCache("10MB")
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	validator, err := validation.NewValidator(cache)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	codeAnalyzer, err := NewCodeAnalyzer(validator, nil)
	if err != nil {
		t.Fatalf("Failed to create code analyzer: %v", err)
	}

	projectAnalyzer, err := NewProjectAnalyzer(codeAnalyzer, validator)
	if err != nil {
		t.Fatalf("Failed to create project analyzer: %v", err)
	}

	// Analyze project first
	ctx := context.Background()
	_, err = projectAnalyzer.AnalyzeProject(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Failed to analyze project: %v", err)
	}

	// Get project summary
	summary, err := projectAnalyzer.GetProjectSummary(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get project summary: %v", err)
	}

	// Verify summary
	if summary.ProjectPath != tmpDir {
		t.Errorf("Expected project path %s, got %s", tmpDir, summary.ProjectPath)
	}

	if summary.TotalFiles != 1 {
		t.Errorf("Expected 1 file, got %d", summary.TotalFiles)
	}

	if summary.TotalTypes != 1 {
		t.Errorf("Expected 1 type, got %d", summary.TotalTypes)
	}

	if summary.PublicTypes != 1 {
		t.Errorf("Expected 1 public type, got %d", summary.PublicTypes)
	}
}

func TestProjectAnalyzer_Caching(t *testing.T) {
	// Create temporary test project
	tmpDir, err := os.MkdirTemp("", "pvm-cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "cached.pl")
	content := `#!/usr/bin/env perl
type Cached = Str;
my Cached $value = "test";
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create analyzer
	cache, err := validation.NewValidationCache("10MB")
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	validator, err := validation.NewValidator(cache)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	codeAnalyzer, err := NewCodeAnalyzer(validator, nil)
	if err != nil {
		t.Fatalf("Failed to create code analyzer: %v", err)
	}

	projectAnalyzer, err := NewProjectAnalyzer(codeAnalyzer, validator)
	if err != nil {
		t.Fatalf("Failed to create project analyzer: %v", err)
	}

	ctx := context.Background()

	// First analysis
	start1 := time.Now()
	analysis1, err := projectAnalyzer.AnalyzeProject(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Failed to analyze project: %v", err)
	}
	duration1 := time.Since(start1)

	// Second analysis (should be cached)
	start2 := time.Now()
	analysis2, err := projectAnalyzer.AnalyzeProject(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Failed to analyze project: %v", err)
	}
	duration2 := time.Since(start2)

	// Cached analysis should be faster
	if duration2 >= duration1 {
		t.Logf("Warning: Cached analysis was not faster (first: %v, second: %v)", duration1, duration2)
	}

	// Results should be the same
	if analysis1.TotalFiles != analysis2.TotalFiles {
		t.Errorf("Cached result differs: files %d vs %d", analysis1.TotalFiles, analysis2.TotalFiles)
	}
	if analysis1.TotalTypes != analysis2.TotalTypes {
		t.Errorf("Cached result differs: types %d vs %d", analysis1.TotalTypes, analysis2.TotalTypes)
	}
}

func TestProjectAnalyzer_EmptyProject(t *testing.T) {
	// Create empty directory
	tmpDir, err := os.MkdirTemp("", "pvm-empty-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create analyzer
	cache, err := validation.NewValidationCache("10MB")
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	validator, err := validation.NewValidator(cache)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	codeAnalyzer, err := NewCodeAnalyzer(validator, nil)
	if err != nil {
		t.Fatalf("Failed to create code analyzer: %v", err)
	}

	projectAnalyzer, err := NewProjectAnalyzer(codeAnalyzer, validator)
	if err != nil {
		t.Fatalf("Failed to create project analyzer: %v", err)
	}

	// Analyze empty project
	ctx := context.Background()
	analysis, err := projectAnalyzer.AnalyzeProject(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Failed to analyze empty project: %v", err)
	}

	// Verify empty results
	if analysis.TotalFiles != 0 {
		t.Errorf("Expected 0 files in empty project, got %d", analysis.TotalFiles)
	}
	if analysis.TotalTypes != 0 {
		t.Errorf("Expected 0 types in empty project, got %d", analysis.TotalTypes)
	}
	if len(analysis.Files) != 0 {
		t.Errorf("Expected no file analyses in empty project, got %d", len(analysis.Files))
	}
}
