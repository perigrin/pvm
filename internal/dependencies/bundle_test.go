// ABOUTME: Unit tests for bundle manager type safety
// ABOUTME: Ensures proper interface usage and eliminates runtime type assertion risks

package dependencies

import (
	"context"
	"log"
	"testing"
	"time"

	"tamarou.com/pvm/internal/cli/progress"
	"tamarou.com/pvm/internal/modules"
)

// mockManager implements the Manager interface for testing
type mockManager struct {
	installedModules map[string]bool
	installResults   []*modules.InstallResult
	installError     error
	isInstalledError error
}

func (m *mockManager) InstallBatch(ctx context.Context, modulesList []string, opts modules.InstallOptions) ([]*modules.InstallResult, error) {
	if m.installError != nil {
		return nil, m.installError
	}

	if m.installResults != nil {
		return m.installResults, nil
	}

	// Default behavior: create successful results
	results := make([]*modules.InstallResult, len(modulesList))
	for i, module := range modulesList {
		results[i] = &modules.InstallResult{
			ModuleName: module,
			Version:    "1.0.0",
			Success:    true,
		}
	}
	return results, nil
}

func (m *mockManager) IsInstalled(ctx context.Context, module string) (bool, error) {
	if m.isInstalledError != nil {
		return false, m.isInstalledError
	}

	if m.installedModules == nil {
		return false, nil
	}

	return m.installedModules[module], nil
}

// mockProgressTracker implements the progress.Tracker interface for testing
type mockProgressTracker struct {
	started   bool
	finished  bool
	operation string
	total     int
	current   int
	message   string
	startTime time.Time
}

func (m *mockProgressTracker) Start(operation string, total int) {
	m.started = true
	m.finished = false
	m.operation = operation
	m.total = total
	m.current = 0
	m.startTime = time.Now()
}

func (m *mockProgressTracker) Update(current int, message string) {
	m.current = current
	m.message = message
}

func (m *mockProgressTracker) Finish(result *progress.Result) {
	m.finished = true
}

func (m *mockProgressTracker) SetTotal(total int) {
	m.total = total
}

func (m *mockProgressTracker) SetMessage(message string) {
	m.message = message
}

func (m *mockProgressTracker) IsRunning() bool {
	return m.started && !m.finished
}

func (m *mockProgressTracker) GetProgress() *progress.Status {
	elapsed := time.Since(m.startTime)
	percentage := 0.0
	if m.total > 0 {
		percentage = float64(m.current) / float64(m.total) * 100
	}

	return &progress.Status{
		Operation:   m.operation,
		Current:     m.current,
		Total:       m.total,
		Message:     m.message,
		Percentage:  percentage,
		StartTime:   m.startTime,
		ElapsedTime: elapsed,
	}
}

func TestNewBundleManager_TypeSafety(t *testing.T) {
	resolver := &DependencyResolver{}
	manager := &mockManager{}
	tracker := &mockProgressTracker{}
	logger := log.Default()

	// This should compile without type assertions
	bundleManager := NewBundleManager(resolver, manager, tracker, logger)

	if bundleManager == nil {
		t.Fatal("NewBundleManager returned nil")
	}

	if bundleManager.manager == nil {
		t.Fatal("BundleManager.manager is nil")
	}

	// Verify the manager field is properly typed
	if bundleManager.manager != manager {
		t.Fatal("BundleManager.manager does not match provided manager")
	}
}

func TestBundleManager_InstallBundle_TypeSafety(t *testing.T) {
	ctx := context.Background()
	resolver := &DependencyResolver{}
	manager := &mockManager{
		installResults: []*modules.InstallResult{
			{
				ModuleName: "Test::Module",
				Version:    "1.0.0",
				Success:    true,
			},
		},
	}
	tracker := &mockProgressTracker{}

	bundleManager := NewBundleManager(resolver, manager, tracker, nil)

	bundle := &Bundle{
		Name:        "test-bundle",
		Description: "Test bundle",
		Modules: []*BundleEntry{
			{
				Name:         "Test::Module",
				Phase:        "runtime",
				Relationship: "requires",
			},
		},
	}

	options := BundleImportOptions{
		SkipTests:        false,
		Force:            false,
		SkipDependencies: false,
		SkipInstalled:    false,
		Parallel:         false,
		Workers:          1,
		DryRun:           false,
	}

	// This should work without any type assertions
	results, err := bundleManager.InstallBundle(ctx, bundle, options)
	if err != nil {
		t.Fatalf("InstallBundle failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].ModuleName != "Test::Module" {
		t.Fatalf("Expected module name 'Test::Module', got '%s'", results[0].ModuleName)
	}
}

func TestBundleManager_InstallBundle_SkipInstalled(t *testing.T) {
	ctx := context.Background()
	resolver := &DependencyResolver{}
	manager := &mockManager{
		installedModules: map[string]bool{
			"Already::Installed": true,
		},
	}
	tracker := &mockProgressTracker{}

	bundleManager := NewBundleManager(resolver, manager, tracker, nil)

	bundle := &Bundle{
		Name:        "test-bundle",
		Description: "Test bundle with installed module",
		Modules: []*BundleEntry{
			{
				Name:         "Already::Installed",
				Phase:        "runtime",
				Relationship: "requires",
			},
			{
				Name:         "Not::Installed",
				Phase:        "runtime",
				Relationship: "requires",
			},
		},
	}

	options := BundleImportOptions{
		SkipInstalled: true,
	}

	results, err := bundleManager.InstallBundle(ctx, bundle, options)
	if err != nil {
		t.Fatalf("InstallBundle failed: %v", err)
	}

	// Should only install the module that's not already installed
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].ModuleName != "Not::Installed" {
		t.Fatalf("Expected module name 'Not::Installed', got '%s'", results[0].ModuleName)
	}
}

func TestBundleManager_NilManager_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	resolver := &DependencyResolver{}
	tracker := &mockProgressTracker{}

	// Test with nil manager (which should be caught by type system now)
	bundleManager := NewBundleManager(resolver, nil, tracker, nil)

	bundle := &Bundle{
		Name:        "test-bundle",
		Description: "Test bundle",
		Modules: []*BundleEntry{
			{
				Name:         "Test::Module",
				Phase:        "runtime",
				Relationship: "requires",
			},
		},
	}

	options := BundleImportOptions{}

	_, err := bundleManager.InstallBundle(ctx, bundle, options)
	if err == nil {
		t.Fatal("Expected error when manager is nil")
	}

	// Should get a system error with code 407
	errorText := err.Error()
	if !containsError(errorText, "407") {
		t.Fatalf("Expected error code 407, got: %v", err)
	}
}

// Helper function to check if error contains specific text
func containsError(errorText, expectedCode string) bool {
	return len(errorText) > 0 && len(expectedCode) > 0 &&
		(errorText[4:7] == expectedCode) // SYS-407 format
}
