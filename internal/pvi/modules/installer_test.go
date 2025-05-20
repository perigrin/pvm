// ABOUTME: Tests for the module installer
// ABOUTME: Tests the module installation functionality

package modules

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/pvi/deps"
)

// Mock CPAN provider for testing
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) GetModuleInfo(ctx context.Context, moduleName string) (*cpan.ModuleInfo, error) {
	args := m.Called(ctx, moduleName)
	return args.Get(0).(*cpan.ModuleInfo), args.Error(1)
}

func (m *MockProvider) SearchModules(ctx context.Context, query string, limit int) (*cpan.SearchResults, error) {
	args := m.Called(ctx, query, limit)
	return args.Get(0).(*cpan.SearchResults), args.Error(1)
}

func (m *MockProvider) GetDependencies(ctx context.Context, moduleName string) ([]*cpan.Dependency, error) {
	args := m.Called(ctx, moduleName)
	return args.Get(0).([]*cpan.Dependency), args.Error(1)
}

func (m *MockProvider) GetModuleVersions(ctx context.Context, moduleName string) ([]string, error) {
	args := m.Called(ctx, moduleName)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockProvider) GetAuthorInfo(ctx context.Context, authorID string) (map[string]interface{}, error) {
	args := m.Called(ctx, authorID)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockProvider) IsCoreModule(ctx context.Context, moduleName string, perlVersion string) (bool, error) {
	args := m.Called(ctx, moduleName, perlVersion)
	return args.Bool(0), args.Error(1)
}

func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) BaseURL() string {
	args := m.Called()
	return args.String(0)
}

// Mock resolver for testing
type MockResolver struct {
	mock.Mock
}

func (m *MockResolver) ResolveDependencies(ctx context.Context, moduleName string, options *deps.DependencyResolutionOptions) (*deps.DependencyResolutionResult, error) {
	args := m.Called(ctx, moduleName, options)
	return args.Get(0).(*deps.DependencyResolutionResult), args.Error(1)
}

func (m *MockResolver) CheckVersionConstraint(version, constraint string) (bool, error) {
	args := m.Called(version, constraint)
	return args.Bool(0), args.Error(1)
}

func (m *MockResolver) GetFlattenedDependencies(result *deps.DependencyResolutionResult) []*deps.DependencyNode {
	args := m.Called(result)
	return args.Get(0).([]*deps.DependencyNode)
}

func (m *MockResolver) PrintDependencyTree(node *deps.DependencyNode) string {
	args := m.Called(node)
	return args.String(0)
}

// TestInstallModule_Basic tests basic functionality of the module installer
func TestInstallModule_Basic(t *testing.T) {
	// Skip if testing is in short mode
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create mock provider
	mockProvider := new(MockProvider)
	mockProvider.On("GetModuleInfo", mock.Anything, "Test::Module").Return(&cpan.ModuleInfo{
		Name:             "Test::Module",
		Version:          "1.0.0",
		Author:           "TESTAUTHOR",
		Description:      "Test module for testing",
		Distribution:     "Test-Module",
		DistributionFile: "authors/id/T/TE/TESTAUTHOR/Test-Module-1.0.0.tar.gz",
	}, nil)
	mockProvider.On("Name").Return("mockprovider")

	// Create mock resolver with no dependencies
	mockResolver := new(MockResolver)
	mockResolver.On("ResolveDependencies", mock.Anything, "Test::Module", mock.Anything).Return(&deps.DependencyResolutionResult{
		Root: &deps.DependencyNode{
			Name:    "Test::Module",
			Version: "1.0.0",
			IsRoot:  true,
		},
		Modules: map[string]*deps.DependencyNode{
			"Test::Module": {
				Name:    "Test::Module",
				Version: "1.0.0",
				IsRoot:  true,
			},
		},
	}, nil)
	mockResolver.On("GetFlattenedDependencies", mock.Anything).Return([]*deps.DependencyNode{
		{
			Name:    "Test::Module",
			Version: "1.0.0",
			IsRoot:  true,
		},
	})

	// Create a mock downloader function that returns success without actually downloading
	originalDownloadModule := DownloadModule
	defer func() { DownloadModule = originalDownloadModule }()

	// Create a dummy archive path
	archivePath := filepath.Join(tempDir, "Test-Module-1.0.0.tar.gz")

	// Mock the download function
	DownloadModule = func(options *DownloadOptions) (*DownloadResult, error) {
		return &DownloadResult{
			Path:         archivePath,
			ModuleName:   options.ModuleName,
			Version:      "1.0.0",
			FromCache:    false,
			Duration:     time.Millisecond * 100,
			Distribution: "Test-Module",
			Author:       "TESTAUTHOR",
		}, nil
	}

	// Create a mock extractor function
	originalExtractModuleArchive := ExtractModuleArchive
	defer func() { ExtractModuleArchive = originalExtractModuleArchive }()

	// Create a dummy extracted directory
	extractedDir := filepath.Join(tempDir, "build", "Test-Module-1.0.0")
	if err := os.MkdirAll(extractedDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	// Mock the extraction function
	ExtractModuleArchive = func(archivePath, targetDir string, ctx context.Context) (*ExtractionResult, error) {
		return &ExtractionResult{
			ExtractedDir: extractedDir,
			ModuleName:   "Test::Module",
			ArchivePath:  archivePath,
			Distribution: "Test-Module",
			RootDir:      "Test-Module-1.0.0",
		}, nil
	}

	// Create a mock builder function
	originalBuildAndInstallModule := BuildAndInstallModule
	defer func() { BuildAndInstallModule = originalBuildAndInstallModule }()

	// Mock the build function
	BuildAndInstallModule = func(options *ModuleBuildOptions) (*ModuleBuildResult, error) {
		return &ModuleBuildResult{
			ModuleName:   options.ModuleName,
			Distribution: options.Distribution,
			Success:      true,
			Installed:    true,
			TestsPassed:  true,
			Warnings:     []string{},
			Errors:       []string{},
			Duration:     time.Millisecond * 200,
		}, nil
	}

	// Create install options
	options := &ModuleInstallOptions{
		ModuleName:         "Test::Module",
		VersionConstraint:  "",
		PerlPath:           "/usr/bin/perl", // Fake path for testing
		InstallDir:         filepath.Join(tempDir, "perl"),
		RunTests:           true,
		Force:              false,
		Cleanup:            true,
		Verbose:            true,
		SkipDependencies:   true, // Skip to simplify test
		Provider:           mockProvider,
		DependencyResolver: mockResolver,
		ProgressCallback:   nil,
		Context:            context.Background(),
	}

	// Test the installation
	result, err := InstallModule(options)

	// Check results
	assert.NoError(t, err, "Installation should succeed")
	assert.NotNil(t, result, "Result should not be nil")
	assert.True(t, result.Success, "Installation should be successful")
	assert.Equal(t, "Test::Module", result.ModuleName, "Module name should match")
	assert.Equal(t, "1.0.0", result.Version, "Version should match")
	assert.Empty(t, result.Errors, "Should have no errors")
	assert.Empty(t, result.Warnings, "Should have no warnings")
	assert.NotZero(t, result.Duration, "Duration should be recorded")

	// Verify mock expectations
	mockProvider.AssertExpectations(t)
	mockResolver.AssertExpectations(t)
}
