// ABOUTME: Tests for the module installer
// ABOUTME: Tests the module installation functionality

package modules

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/project"
)

// Test helper functions for module installer testing

// createMockProvider creates a mock CPAN provider for testing
func createMockProvider() *cpan.ProviderMock {
	return &cpan.ProviderMock{
		GetModuleInfoFunc: func(ctx context.Context, moduleName string) (*cpan.ModuleInfo, error) {
			if moduleName == "Test::Module" {
				return &cpan.ModuleInfo{
					Name:    "Test::Module",
					Version: "1.0.0",
					Author:  "TEST",
				}, nil
			}
			return nil, &cpan.ProviderError{
				Source:  "mock",
				Code:    "MODULE_NOT_FOUND",
				Message: "Module not found",
			}
		},
		SearchModulesFunc: func(ctx context.Context, query string, limit int) (*cpan.SearchResults, error) {
			return &cpan.SearchResults{
				Query:   query,
				Total:   1,
				Results: []*cpan.SearchResult{},
				Source:  "mock",
			}, nil
		},
		GetDependenciesFunc: func(ctx context.Context, moduleName string) ([]*cpan.Dependency, error) {
			return []*cpan.Dependency{}, nil
		},
		GetModuleVersionsFunc: func(ctx context.Context, moduleName string) ([]string, error) {
			return []string{"1.0.0"}, nil
		},
		GetAuthorInfoFunc: func(ctx context.Context, authorID string) (map[string]interface{}, error) {
			return map[string]interface{}{"name": "Test Author"}, nil
		},
		IsCoreModuleFunc: func(ctx context.Context, moduleName string, perlVersion string) (bool, error) {
			return false, nil
		},
		NameFunc: func() string {
			return "mock"
		},
		BaseURLFunc: func() string {
			return "https://mock.cpan.org"
		},
	}
}

// TestInstallModule_Basic tests basic functionality of the module installer
func TestInstallModule_Basic(t *testing.T) {
	// Save original function variables
	origDownloadModule := DownloadModule
	origExtractModuleArchive := ExtractModuleArchive
	origBuildAndInstallModule := BuildAndInstallModule

	// Restore original functions after test
	defer func() {
		DownloadModule = origDownloadModule
		ExtractModuleArchive = origExtractModuleArchive
		BuildAndInstallModule = origBuildAndInstallModule
	}()

	// Set up mocks
	DownloadModule = func(options *DownloadOptions) (*DownloadResult, error) {
		return &DownloadResult{
			Path:       "/tmp/test-module.tar.gz",
			ModuleName: "Test::Module",
			Version:    "1.0.0",
			Size:       12345,
		}, nil
	}

	ExtractModuleArchive = func(archivePath, targetDir string, ctx context.Context) (*ExtractionResult, error) {
		return &ExtractionResult{
			ExtractedDir: "/tmp/extracted/Test-Module-1.0.0",
			ModuleName:   "Test::Module",
			ArchivePath:  archivePath,
			Distribution: "Test-Module-1.0.0",
		}, nil
	}

	BuildAndInstallModule = func(options *ModuleBuildOptions) (*ModuleBuildResult, error) {
		return &ModuleBuildResult{
			ModuleName:   "Test::Module",
			Distribution: "Test-Module-1.0.0",
			Success:      true,
			Installed:    true,
			Warnings:     []string{},
			Errors:       []string{},
		}, nil
	}

	// Create a mock provider
	mockProvider := createMockProvider()

	// Create install options
	options := &ModuleInstallOptions{
		ModuleName:       "Test::Module",
		Provider:         mockProvider,
		SkipDependencies: true, // Skip dependencies for basic test
		Verbose:          false,
	}

	// Execute installation
	result, err := InstallModule(options)

	// Verify results
	if err != nil {
		t.Fatalf("InstallModule should not return error: %v", err)
	}

	if result == nil {
		t.Fatal("InstallModule should return a result")
	}

	if !result.Success {
		t.Errorf("Installation should be successful, errors: %v", result.Errors)
	}

	if result.ModuleName != "Test::Module" {
		t.Errorf("Expected module name 'Test::Module', got '%s'", result.ModuleName)
	}

	if result.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", result.Version)
	}

	if len(result.Errors) > 0 {
		t.Errorf("Expected no errors, got: %v", result.Errors)
	}
}

// TestSetupIsolationEnvironment_ProjectContext tests project-aware installation directory setup
func TestSetupIsolationEnvironment_ProjectContext(t *testing.T) {
	tests := []struct {
		name           string
		projectContext *project.ProjectContext
		forceGlobal    bool
		expectedInPath string
		expectedType   string
	}{
		{
			name:           "No project context",
			projectContext: nil,
			forceGlobal:    false,
			expectedInPath: "perl/modules", // Should use XDG data directory
			expectedType:   "global",
		},
		{
			name: "Project context with local installation",
			projectContext: &project.ProjectContext{
				IsProject:   true,
				RootDir:     "/tmp/test-project",
				LocalLibDir: "/tmp/test-project/local", // Changed from "lib" to "local"
			},
			forceGlobal:    false,
			expectedInPath: "/tmp/test-project/local", // Should use project local
			expectedType:   "project",
		},
		{
			name: "Project context with forced global installation",
			projectContext: &project.ProjectContext{
				IsProject:   true,
				RootDir:     "/tmp/test-project",
				LocalLibDir: "/tmp/test-project/local", // Set LocalLibDir field
			},
			forceGlobal:    true,
			expectedInPath: "perl/modules", // Should use global despite project context
			expectedType:   "global",
		},
		{
			name: "Non-project context (IsProject=false)",
			projectContext: &project.ProjectContext{
				IsProject: false,
				RootDir:   "",
			},
			forceGlobal:    false,
			expectedInPath: "perl/modules", // Should use global
			expectedType:   "global",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for testing
			tempDir := t.TempDir()

			// Create test options
			options := &ModuleInstallOptions{
				ModuleName:     "Test::Module",
				ProjectContext: tt.projectContext,
				ForceGlobal:    tt.forceGlobal,
			}

			// Update project context to use temp directory if needed
			if options.ProjectContext != nil && tt.expectedType == "project" {
				options.ProjectContext.RootDir = tempDir
				options.ProjectContext.LocalLibDir = filepath.Join(tempDir, "lib")
			}

			// Test setup isolation environment
			installDir, envVars, err := setupIsolationEnvironment(options)
			if err != nil {
				t.Fatalf("setupIsolationEnvironment failed: %v", err)
			}

			// Check install directory
			if tt.expectedType == "project" {
				expectedDir := filepath.Join(tempDir, "lib")
				if installDir != expectedDir {
					t.Errorf("Expected install dir %s, got %s", expectedDir, installDir)
				}
			} else {
				if !filepath.IsAbs(installDir) {
					t.Errorf("Expected absolute path for install dir, got %s", installDir)
				}
				if !strings.HasSuffix(installDir, tt.expectedInPath) {
					t.Errorf("Expected install dir to contain %s, got %s", tt.expectedInPath, installDir)
				}
			}

			// Check environment variables
			requiredEnvVars := []string{
				"PERL_LOCAL_LIB_ROOT",
				"PERL_MB_OPT",
				"PERL_MM_OPT",
				"PERL5LIB",
				"PATH",
			}

			for _, envVar := range requiredEnvVars {
				if _, exists := envVars[envVar]; !exists {
					t.Errorf("Expected environment variable %s to be set", envVar)
				}
			}

			// Check PERL5LIB contains the lib directory
			perl5lib := envVars["PERL5LIB"]
			expectedPerlLib := filepath.Join(installDir, "lib", "perl5")
			if perl5lib == "" || !containsPath(perl5lib, expectedPerlLib) {
				t.Errorf("Expected PERL5LIB to contain %s, got %s", expectedPerlLib, perl5lib)
			}

			// For project context, check if project lib is in PERL5LIB
			if tt.projectContext != nil && tt.projectContext.IsProject && !tt.forceGlobal {
				projectLib := filepath.Join(tempDir, "lib")
				if !containsPath(perl5lib, projectLib) {
					t.Errorf("Expected PERL5LIB to contain project lib %s, got %s", projectLib, perl5lib)
				}
			}
		})
	}
}

// TestSetupIsolationEnvironment_PERL5LIB_Preservation tests that existing PERL5LIB is preserved
func TestSetupIsolationEnvironment_PERL5LIB_Preservation(t *testing.T) {
	// Set up existing PERL5LIB
	originalPERL5LIB := os.Getenv("PERL5LIB")
	testPERL5LIB := "/existing/perl/lib:/another/lib"
	os.Setenv("PERL5LIB", testPERL5LIB)
	defer func() {
		if originalPERL5LIB != "" {
			os.Setenv("PERL5LIB", originalPERL5LIB)
		} else {
			os.Unsetenv("PERL5LIB")
		}
	}()

	options := &ModuleInstallOptions{
		ModuleName: "Test::Module",
	}

	_, envVars, err := setupIsolationEnvironment(options)
	if err != nil {
		t.Fatalf("setupIsolationEnvironment failed: %v", err)
	}

	perl5lib := envVars["PERL5LIB"]
	if !containsPath(perl5lib, "/existing/perl/lib") {
		t.Errorf("Expected PERL5LIB to preserve existing path /existing/perl/lib, got %s", perl5lib)
	}
	if !containsPath(perl5lib, "/another/lib") {
		t.Errorf("Expected PERL5LIB to preserve existing path /another/lib, got %s", perl5lib)
	}
}

// TestInstallModule_ProjectContextDetection tests automatic project context detection
func TestInstallModule_ProjectContextDetection(t *testing.T) {
	// Create a temporary project directory
	tempDir := t.TempDir()
	originalWD, _ := os.Getwd()
	defer os.Chdir(originalWD)

	// Create project marker file
	perlVersionFile := filepath.Join(tempDir, ".perl-version")
	err := os.WriteFile(perlVersionFile, []byte("5.38.0\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create .perl-version file: %v", err)
	}

	// Change to project directory
	os.Chdir(tempDir)

	// Clear project detection cache
	project.ClearDetectionCache()

	// Test that project context is auto-detected
	// We can't easily test the full installation without mocking,
	// but we can test the project detection logic
	projectCtx, err := project.GetCurrentProject()
	if err != nil {
		t.Fatalf("Failed to detect project: %v", err)
	}

	if !projectCtx.IsProject {
		t.Error("Expected project to be detected")
	}

	// Resolve both paths to handle macOS symlinks (/var -> /private/var)
	expectedRoot, err := filepath.EvalSymlinks(tempDir)
	if err != nil {
		expectedRoot = tempDir // fallback if symlink resolution fails
	}
	actualRoot, err := filepath.EvalSymlinks(projectCtx.RootDir)
	if err != nil {
		actualRoot = projectCtx.RootDir // fallback if symlink resolution fails
	}

	if actualRoot != expectedRoot {
		t.Errorf("Expected project root %s, got %s", expectedRoot, actualRoot)
	}

	if projectCtx.PerlVersion != "5.38.0" {
		t.Errorf("Expected Perl version 5.38.0, got %s", projectCtx.PerlVersion)
	}
}

// Helper function to check if a colon-separated path list contains a specific path
func containsPath(pathList, targetPath string) bool {
	if pathList == "" {
		return false
	}

	paths := filepath.SplitList(pathList)
	for _, path := range paths {
		if path == targetPath {
			return true
		}
	}
	return false
}
