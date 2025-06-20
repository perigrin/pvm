// ABOUTME: Tests for unified module installer
// ABOUTME: Validates module installation functionality and helper functions

package modules

import (
	"context"
	"os"
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
