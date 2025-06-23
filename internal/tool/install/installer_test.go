// ABOUTME: Tests for core installation logic
// ABOUTME: Tests for global tool installation using PVM local-lib system

package install

import (
	"context"
	"os"
	"testing"
	"time"

	"tamarou.com/pvm/internal/tool"
)

func TestNewToolInstaller(t *testing.T) {
	installer, err := NewToolInstaller()
	if err != nil {
		t.Fatalf("Failed to create tool installer: %v", err)
	}

	if installer == nil {
		t.Fatal("Tool installer is nil")
	}

	if installer.storage == nil {
		t.Fatal("Tool storage is nil")
	}

	if installer.toolMapping == nil {
		t.Fatal("Tool mapping is nil")
	}
}

func TestValidateInstallOptions(t *testing.T) {
	installer, err := NewToolInstaller()
	if err != nil {
		t.Fatalf("Failed to create tool installer: %v", err)
	}

	// Test nil options
	err = installer.validateInstallOptions(nil)
	if err == nil {
		t.Error("Should fail with nil options")
	}

	// Test empty tool name
	options := &InstallOptions{
		ToolName: "",
	}
	err = installer.validateInstallOptions(options)
	if err == nil {
		t.Error("Should fail with empty tool name")
	}

	// Test invalid tool name
	options = &InstallOptions{
		ToolName: "invalid-tool-name-with-@-symbols",
	}
	err = installer.validateInstallOptions(options)
	if err == nil {
		t.Error("Should fail with invalid tool name")
	}

	// Test valid options
	options = &InstallOptions{
		ToolName: "valid-tool-name",
	}
	err = installer.validateInstallOptions(options)
	if err != nil {
		t.Errorf("Should pass with valid options: %v", err)
	}
}

func TestIsToolInstalled(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-tool-installer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage with temp directory
	storage := &ToolStorage{baseDir: tempDir}
	installer := &ToolInstaller{
		storage:     storage,
		toolMapping: tool.NewToolMapping(),
	}

	// Tool should not be installed initially
	if installer.IsToolInstalled("test-tool") {
		t.Error("Tool should not be installed initially")
	}

	// Create metadata for tool
	metadata := &ToolMetadata{
		ToolName:    "test-tool",
		ModuleName:  "Test::Tool",
		Version:     "1.0.0",
		InstallDate: time.Now(),
		Status:      "installed",
	}

	err = storage.SaveMetadata(metadata)
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Tool should be installed now
	if !installer.IsToolInstalled("test-tool") {
		t.Error("Tool should be installed after saving metadata")
	}
}

func TestGetInstalledVersion(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-tool-installer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage with temp directory
	storage := &ToolStorage{baseDir: tempDir}
	installer := &ToolInstaller{
		storage:     storage,
		toolMapping: tool.NewToolMapping(),
	}

	// Should fail for non-existent tool
	_, err = installer.GetInstalledVersion("nonexistent-tool")
	if err == nil {
		t.Error("Should fail for non-existent tool")
	}

	// Create metadata for tool
	expectedVersion := "2.1.0"
	metadata := &ToolMetadata{
		ToolName:    "versioned-tool",
		ModuleName:  "Versioned::Tool",
		Version:     expectedVersion,
		InstallDate: time.Now(),
		Status:      "installed",
	}

	err = storage.SaveMetadata(metadata)
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Should return correct version
	version, err := installer.GetInstalledVersion("versioned-tool")
	if err != nil {
		t.Fatalf("Failed to get installed version: %v", err)
	}

	if version != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, version)
	}
}

func TestInstallToolValidation(t *testing.T) {
	installer, err := NewToolInstaller()
	if err != nil {
		t.Fatalf("Failed to create tool installer: %v", err)
	}

	// Test installation with invalid options
	result, err := installer.InstallTool(nil)
	if err == nil {
		t.Error("Should fail with nil options")
	}
	if result != nil {
		t.Error("Result should be nil on validation failure")
	}

	// Test installation with empty tool name
	options := &InstallOptions{
		ToolName: "",
	}
	result, err = installer.InstallTool(options)
	if err == nil {
		t.Error("Should fail with empty tool name")
	}
	if result != nil {
		t.Error("Result should be nil on validation failure")
	}
}

func TestInstallToolExistingTool(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-tool-installer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage with temp directory
	storage := &ToolStorage{baseDir: tempDir}
	installer := &ToolInstaller{
		storage:     storage,
		toolMapping: tool.NewToolMapping(),
	}

	// Create existing tool
	metadata := &ToolMetadata{
		ToolName:    "existing-tool",
		ModuleName:  "Existing::Tool",
		Version:     "1.0.0",
		InstallDate: time.Now(),
		Status:      "installed",
	}

	err = storage.SaveMetadata(metadata)
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Try to install existing tool without force
	options := &InstallOptions{
		ToolName: "existing-tool",
		Context:  context.Background(),
	}

	result, err := installer.InstallTool(options)
	if err == nil {
		t.Error("Should fail when tool already exists without force flag")
	}
	if result == nil {
		t.Error("Result should not be nil")
	}
	if result.Success {
		t.Error("Result should not indicate success")
	}
}

func TestInstallToolToolMapping(t *testing.T) {
	installer, err := NewToolInstaller()
	if err != nil {
		t.Fatalf("Failed to create tool installer: %v", err)
	}

	// Test tool that should resolve through built-in mapping
	options := &InstallOptions{
		ToolName: "ack", // Should resolve to App::Ack
		Context:  context.Background(),
	}

	// We can't test the full installation without CPAN provider,
	// but we can test that the validation passes and tool resolution works
	err = installer.validateInstallOptions(options)
	if err != nil {
		t.Errorf("Validation should pass for known tool: %v", err)
	}

	// Test explicit module name
	options = &InstallOptions{
		ToolName:   "custom-tool",
		ModuleName: "Custom::Tool::Module",
		Context:    context.Background(),
	}

	err = installer.validateInstallOptions(options)
	if err != nil {
		t.Errorf("Validation should pass with explicit module name: %v", err)
	}
}

func TestInstallOptionsDefaults(t *testing.T) {
	installer, err := NewToolInstaller()
	if err != nil {
		t.Fatalf("Failed to create tool installer: %v", err)
	}

	// Test that default context is set
	options := &InstallOptions{
		ToolName: "test-tool",
		// Context is nil, should be set to Background()
	}

	// Mock the storage to avoid actual installation
	tempDir, err := os.MkdirTemp("", "pvm-tool-installer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	installer.storage = &ToolStorage{baseDir: tempDir}

	// Try to install - it will fail due to missing CPAN provider,
	// but it should set the default context first
	result, err := installer.InstallTool(options)
	if err == nil {
		t.Error("Should fail due to missing CPAN provider")
	}

	// The context should have been set even though installation failed
	if options.Context == nil {
		t.Error("Context should have been set to default")
	}

	if result == nil {
		t.Error("Result should not be nil")
	}
}

func TestInstallProgressCallback(t *testing.T) {
	installer, err := NewToolInstaller()
	if err != nil {
		t.Fatalf("Failed to create tool installer: %v", err)
	}

	var progressCalls []string
	progressCallback := func(stage string, progress float64, message string) {
		progressCalls = append(progressCalls, stage)
	}

	options := &InstallOptions{
		ToolName:         "progress-test-tool",
		Context:          context.Background(),
		ProgressCallback: progressCallback,
	}

	// Mock the storage
	tempDir, err := os.MkdirTemp("", "pvm-tool-installer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	installer.storage = &ToolStorage{baseDir: tempDir}

	// Try to install - it will fail, but should call progress callback
	_, err = installer.InstallTool(options)
	if err == nil {
		t.Error("Should fail due to missing CPAN provider")
	}

	// Should have at least called progress for resolve stage
	if len(progressCalls) == 0 {
		t.Error("Progress callback should have been called")
	}

	// First call should be resolve stage
	if progressCalls[0] != "resolve" {
		t.Errorf("First progress call should be 'resolve', got '%s'", progressCalls[0])
	}
}

func TestGetToolStorage(t *testing.T) {
	installer, err := NewToolInstaller()
	if err != nil {
		t.Fatalf("Failed to create tool installer: %v", err)
	}

	storage := installer.GetToolStorage()
	if storage == nil {
		t.Error("Storage should not be nil")
	}

	// Should be the same storage instance
	if storage != installer.storage {
		t.Error("Returned storage should be the same instance")
	}
}

func TestSetProviders(t *testing.T) {
	installer, err := NewToolInstaller()
	if err != nil {
		t.Fatalf("Failed to create tool installer: %v", err)
	}

	// These are just testing that the setters don't panic
	// since we don't have actual implementations to test with
	installer.SetCPANProvider(nil)
	installer.SetDependencyResolver(nil)

	// Should not panic and should maintain installer state
	if installer.storage == nil {
		t.Error("Storage should still be available after setting providers")
	}
}
