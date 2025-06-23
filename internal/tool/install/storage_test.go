// ABOUTME: Tests for tool storage and isolation management
// ABOUTME: Comprehensive tests for global tool storage functionality

package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewToolStorage(t *testing.T) {
	// Test successful creation
	storage, err := NewToolStorage()
	if err != nil {
		t.Fatalf("Failed to create tool storage: %v", err)
	}

	if storage == nil {
		t.Fatal("Tool storage is nil")
	}

	if storage.baseDir == "" {
		t.Fatal("Base directory is empty")
	}
}

func TestCreateToolDirectory(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-tool-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := &ToolStorage{baseDir: tempDir}

	// Test creating tool directory
	toolName := "test-tool"
	err = storage.CreateToolDirectory(toolName)
	if err != nil {
		t.Fatalf("Failed to create tool directory: %v", err)
	}

	// Verify directory structure
	toolPath := storage.GetToolPath(toolName)
	expectedDirs := []string{
		toolPath,
		filepath.Join(toolPath, "lib", "perl5"),
		filepath.Join(toolPath, "bin"),
		filepath.Join(toolPath, "man"),
		filepath.Join(toolPath, "share"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory does not exist: %s", dir)
		}
	}
}

func TestSaveAndLoadMetadata(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-tool-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := &ToolStorage{baseDir: tempDir}

	// Create test metadata
	metadata := &ToolMetadata{
		ToolName:     "test-tool",
		ModuleName:   "Test::Tool",
		Version:      "1.0.0",
		InstallDate:  time.Now(),
		InstallPath:  "/test/path",
		LocalLibPath: "/test/path/lib/perl5",
		BinPath:      "/test/path/bin",
		Dependencies: []string{"Module::A", "Module::B"},
		PerlVersion:  "5.36.0",
		BuildArgs:    []string{"--verbose"},
		Status:       "installed",
		LastVerified: time.Now(),
		CustomData:   map[string]string{"key": "value"},
	}

	// Test saving metadata
	err = storage.SaveMetadata(metadata)
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Test loading metadata
	loadedMetadata, err := storage.LoadMetadata("test-tool")
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	// Verify metadata fields
	if loadedMetadata.ToolName != metadata.ToolName {
		t.Errorf("Expected tool name %s, got %s", metadata.ToolName, loadedMetadata.ToolName)
	}

	if loadedMetadata.ModuleName != metadata.ModuleName {
		t.Errorf("Expected module name %s, got %s", metadata.ModuleName, loadedMetadata.ModuleName)
	}

	if loadedMetadata.Version != metadata.Version {
		t.Errorf("Expected version %s, got %s", metadata.Version, loadedMetadata.Version)
	}

	if len(loadedMetadata.Dependencies) != len(metadata.Dependencies) {
		t.Errorf("Expected %d dependencies, got %d", len(metadata.Dependencies), len(loadedMetadata.Dependencies))
	}

	if loadedMetadata.CustomData["key"] != "value" {
		t.Errorf("Expected custom data value 'value', got '%s'", loadedMetadata.CustomData["key"])
	}
}

func TestToolExists(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-tool-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := &ToolStorage{baseDir: tempDir}

	// Tool should not exist initially
	if storage.ToolExists("nonexistent-tool") {
		t.Error("Tool should not exist")
	}

	// Create metadata for a tool
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

	// Tool should exist now
	if !storage.ToolExists("existing-tool") {
		t.Error("Tool should exist")
	}
}

func TestListTools(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-tool-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := &ToolStorage{baseDir: tempDir}

	// Should return empty list initially
	tools, err := storage.ListTools()
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	if len(tools) != 0 {
		t.Errorf("Expected 0 tools, got %d", len(tools))
	}

	// Create some tools
	toolNames := []string{"tool1", "tool2", "tool3"}
	for i, toolName := range toolNames {
		metadata := &ToolMetadata{
			ToolName:    toolName,
			ModuleName:  "Test::Tool" + string(rune('1'+i)),
			Version:     "1.0.0",
			InstallDate: time.Now(),
			Status:      "installed",
		}

		err = storage.SaveMetadata(metadata)
		if err != nil {
			t.Fatalf("Failed to save metadata for %s: %v", toolName, err)
		}
	}

	// List tools again
	tools, err = storage.ListTools()
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	if len(tools) != len(toolNames) {
		t.Errorf("Expected %d tools, got %d", len(toolNames), len(tools))
	}

	// Verify tool names
	foundTools := make(map[string]bool)
	for _, tool := range tools {
		foundTools[tool.ToolName] = true
	}

	for _, toolName := range toolNames {
		if !foundTools[toolName] {
			t.Errorf("Tool %s not found in list", toolName)
		}
	}
}

func TestRemoveTool(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-tool-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := &ToolStorage{baseDir: tempDir}

	// Create a tool
	metadata := &ToolMetadata{
		ToolName:    "removable-tool",
		ModuleName:  "Removable::Tool",
		Version:     "1.0.0",
		InstallDate: time.Now(),
		Status:      "installed",
	}

	err = storage.SaveMetadata(metadata)
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Verify tool exists
	if !storage.ToolExists("removable-tool") {
		t.Fatal("Tool should exist before removal")
	}

	// Remove the tool
	err = storage.RemoveTool("removable-tool")
	if err != nil {
		t.Fatalf("Failed to remove tool: %v", err)
	}

	// Verify tool no longer exists
	if storage.ToolExists("removable-tool") {
		t.Error("Tool should not exist after removal")
	}

	// Verify directory is removed
	toolPath := storage.GetToolPath("removable-tool")
	if _, err := os.Stat(toolPath); !os.IsNotExist(err) {
		t.Error("Tool directory should be removed")
	}
}

func TestGetPaths(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-tool-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := &ToolStorage{baseDir: tempDir}

	toolName := "test-tool"

	// Test path getters
	toolPath := storage.GetToolPath(toolName)
	expectedToolPath := filepath.Join(tempDir, toolName)
	if toolPath != expectedToolPath {
		t.Errorf("Expected tool path %s, got %s", expectedToolPath, toolPath)
	}

	metadataPath := storage.GetMetadataPath(toolName)
	expectedMetadataPath := filepath.Join(expectedToolPath, MetadataFileName)
	if metadataPath != expectedMetadataPath {
		t.Errorf("Expected metadata path %s, got %s", expectedMetadataPath, metadataPath)
	}

	localLibPath := storage.GetToolLocalLibPath(toolName)
	expectedLocalLibPath := filepath.Join(expectedToolPath, "lib", "perl5")
	if localLibPath != expectedLocalLibPath {
		t.Errorf("Expected local lib path %s, got %s", expectedLocalLibPath, localLibPath)
	}

	binPath := storage.GetToolBinPath(toolName)
	expectedBinPath := filepath.Join(expectedToolPath, "bin")
	if binPath != expectedBinPath {
		t.Errorf("Expected bin path %s, got %s", expectedBinPath, binPath)
	}
}

func TestCleanupOrphanedTools(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-tool-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := &ToolStorage{baseDir: tempDir}

	// Create a valid tool
	validTool := &ToolMetadata{
		ToolName:    "valid-tool",
		ModuleName:  "Valid::Tool",
		Version:     "1.0.0",
		InstallDate: time.Now(),
		Status:      "installed",
	}

	err = storage.SaveMetadata(validTool)
	if err != nil {
		t.Fatalf("Failed to save valid tool metadata: %v", err)
	}

	// Create an orphaned tool directory (without metadata)
	orphanedPath := filepath.Join(tempDir, "orphaned-tool")
	err = os.MkdirAll(orphanedPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create orphaned directory: %v", err)
	}

	// Create some dummy files in orphaned directory
	dummyFile := filepath.Join(orphanedPath, "dummy.txt")
	err = os.WriteFile(dummyFile, []byte("dummy content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	// Verify orphaned directory exists
	if _, err := os.Stat(orphanedPath); os.IsNotExist(err) {
		t.Fatal("Orphaned directory should exist before cleanup")
	}

	// Run cleanup
	err = storage.CleanupOrphanedTools()
	if err != nil {
		t.Fatalf("Failed to cleanup orphaned tools: %v", err)
	}

	// Verify orphaned directory is removed
	if _, err := os.Stat(orphanedPath); !os.IsNotExist(err) {
		t.Error("Orphaned directory should be removed after cleanup")
	}

	// Verify valid tool still exists
	if !storage.ToolExists("valid-tool") {
		t.Error("Valid tool should still exist after cleanup")
	}
}

func TestValidateToolInstallation(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-tool-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := &ToolStorage{baseDir: tempDir}

	// Create tool directories
	toolPath := storage.GetToolPath("test-tool")
	localLibPath := storage.GetToolLocalLibPath("test-tool")
	binPath := storage.GetToolBinPath("test-tool")

	err = os.MkdirAll(localLibPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create local lib path: %v", err)
	}

	err = os.MkdirAll(binPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create bin path: %v", err)
	}

	// Create metadata with valid paths
	metadata := &ToolMetadata{
		ToolName:     "test-tool",
		ModuleName:   "Test::Tool",
		Version:      "1.0.0",
		InstallDate:  time.Now(),
		InstallPath:  toolPath,
		LocalLibPath: localLibPath,
		BinPath:      binPath,
		Status:       "installed",
	}

	err = storage.SaveMetadata(metadata)
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Validation should pass
	err = storage.ValidateToolInstallation("test-tool")
	if err != nil {
		t.Errorf("Validation should pass: %v", err)
	}

	// Remove bin path and test validation failure
	err = os.RemoveAll(binPath)
	if err != nil {
		t.Fatalf("Failed to remove bin path: %v", err)
	}

	err = storage.ValidateToolInstallation("test-tool")
	if err == nil {
		t.Error("Validation should fail when bin path is missing")
	}
}

func TestMetadataJSONFormat(t *testing.T) {
	// Test that metadata can be properly serialized and deserialized
	metadata := &ToolMetadata{
		ToolName:     "json-test-tool",
		ModuleName:   "JSON::Test::Tool",
		Version:      "2.1.0",
		InstallDate:  time.Now().Truncate(time.Second), // Truncate for JSON precision
		InstallPath:  "/test/path",
		LocalLibPath: "/test/path/lib/perl5",
		BinPath:      "/test/path/bin",
		Dependencies: []string{"Module::A", "Module::B", "Module::C"},
		PerlVersion:  "5.36.0",
		BuildArgs:    []string{"--verbose", "--test"},
		Status:       "installed",
		LastVerified: time.Now().Truncate(time.Second),
		CustomData:   map[string]string{"env": "test", "platform": "linux"},
	}

	// Serialize to JSON
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	// Deserialize from JSON
	var deserializedMetadata ToolMetadata
	err = json.Unmarshal(jsonData, &deserializedMetadata)
	if err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	// Compare all fields
	if deserializedMetadata.ToolName != metadata.ToolName {
		t.Errorf("Tool name mismatch: expected %s, got %s", metadata.ToolName, deserializedMetadata.ToolName)
	}

	if deserializedMetadata.ModuleName != metadata.ModuleName {
		t.Errorf("Module name mismatch: expected %s, got %s", metadata.ModuleName, deserializedMetadata.ModuleName)
	}

	if len(deserializedMetadata.Dependencies) != len(metadata.Dependencies) {
		t.Errorf("Dependencies length mismatch: expected %d, got %d",
			len(metadata.Dependencies), len(deserializedMetadata.Dependencies))
	}

	if deserializedMetadata.CustomData["env"] != metadata.CustomData["env"] {
		t.Errorf("Custom data mismatch: expected %s, got %s",
			metadata.CustomData["env"], deserializedMetadata.CustomData["env"])
	}
}
