// ABOUTME: Unit tests for shim manager functionality
// ABOUTME: Tests shim creation, updating, removal, and conflict detection

package shim

import (
	"os"
	"runtime"
	"testing"
	"time"

	"tamarou.com/pvm/internal/tool"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager.shimDir == "" {
		t.Error("Shim directory should not be empty")
	}

	if manager.platform != runtime.GOOS {
		t.Errorf("Expected platform %s, got %s", runtime.GOOS, manager.platform)
	}

	// Check that shim directory exists
	if _, err := os.Stat(manager.shimDir); os.IsNotExist(err) {
		t.Errorf("Shim directory should exist: %s", manager.shimDir)
	}
}

func TestCreateShim(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create test tool info
	toolInfo := &tool.ToolInfo{
		Name:        "test-tool",
		Module:      "Test::Tool",
		Version:     "1.0.0",
		InstallDate: time.Now(),
	}

	// Test creating shim
	err = manager.CreateShim("test-tool", toolInfo)
	if err != nil {
		t.Fatalf("Failed to create shim: %v", err)
	}

	// Check that shim exists
	if !manager.ShimExists("test-tool") {
		t.Error("Shim should exist after creation")
	}

	// Clean up
	defer manager.RemoveShim("test-tool")

	// Verify shim content
	shimPath := manager.getShimPath("test-tool")
	content, err := os.ReadFile(shimPath)
	if err != nil {
		t.Fatalf("Failed to read shim file: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "PVX Global Tool Shim") {
		t.Error("Shim should contain PVX header")
	}

	if !contains(contentStr, "test-tool") {
		t.Error("Shim should contain tool name")
	}

	if !contains(contentStr, "Test::Tool") {
		t.Error("Shim should contain module name")
	}
}

func TestCreateShimErrors(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test empty tool name
	err = manager.CreateShim("", &tool.ToolInfo{})
	if err == nil {
		t.Error("Should error on empty tool name")
	}

	// Test nil tool info
	err = manager.CreateShim("test", nil)
	if err == nil {
		t.Error("Should error on nil tool info")
	}
}

func TestUpdateShim(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create initial shim
	toolInfo := &tool.ToolInfo{
		Name:        "test-tool",
		Module:      "Test::Tool",
		Version:     "1.0.0",
		InstallDate: time.Now(),
	}

	err = manager.CreateShim("test-tool", toolInfo)
	if err != nil {
		t.Fatalf("Failed to create initial shim: %v", err)
	}
	defer manager.RemoveShim("test-tool")

	// Update shim with new version
	updatedInfo := &tool.ToolInfo{
		Name:        "test-tool",
		Module:      "Test::Tool",
		Version:     "2.0.0",
		InstallDate: time.Now(),
	}

	err = manager.UpdateShim("test-tool", updatedInfo)
	if err != nil {
		t.Fatalf("Failed to update shim: %v", err)
	}

	// Verify updated content
	shimPath := manager.getShimPath("test-tool")
	content, err := os.ReadFile(shimPath)
	if err != nil {
		t.Fatalf("Failed to read updated shim: %v", err)
	}

	if !contains(string(content), "2.0.0") {
		t.Error("Updated shim should contain new version")
	}
}

func TestRemoveShim(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create shim
	toolInfo := &tool.ToolInfo{
		Name:        "test-tool",
		Module:      "Test::Tool",
		Version:     "1.0.0",
		InstallDate: time.Now(),
	}

	err = manager.CreateShim("test-tool", toolInfo)
	if err != nil {
		t.Fatalf("Failed to create shim: %v", err)
	}

	// Verify it exists
	if !manager.ShimExists("test-tool") {
		t.Error("Shim should exist before removal")
	}

	// Remove shim
	err = manager.RemoveShim("test-tool")
	if err != nil {
		t.Fatalf("Failed to remove shim: %v", err)
	}

	// Verify it's gone
	if manager.ShimExists("test-tool") {
		t.Error("Shim should not exist after removal")
	}

	// Test removing non-existent shim (should not error)
	err = manager.RemoveShim("non-existent")
	if err != nil {
		t.Errorf("Removing non-existent shim should not error: %v", err)
	}
}

func TestListShims(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create multiple shims
	tools := []string{"tool1", "tool2", "tool3"}
	toolInfo := &tool.ToolInfo{
		Name:        "test-tool",
		Module:      "Test::Tool",
		Version:     "1.0.0",
		InstallDate: time.Now(),
	}

	for _, toolName := range tools {
		err = manager.CreateShim(toolName, toolInfo)
		if err != nil {
			t.Fatalf("Failed to create shim %s: %v", toolName, err)
		}
		defer manager.RemoveShim(toolName)
	}

	// List shims
	shims, err := manager.ListShims()
	if err != nil {
		t.Fatalf("Failed to list shims: %v", err)
	}

	// Verify all tools are listed
	for _, toolName := range tools {
		found := false
		for _, shim := range shims {
			if shim == toolName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Tool %s not found in shim list", toolName)
		}
	}
}

func TestGetShimPath(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	shimPath := manager.getShimPath("test-tool")

	// Check path contains tool name
	if !contains(shimPath, "test-tool") {
		t.Error("Shim path should contain tool name")
	}

	// Check platform-specific extension
	if runtime.GOOS == "windows" {
		if !contains(shimPath, ".bat") {
			t.Error("Windows shim path should have .bat extension")
		}
	} else {
		if contains(shimPath, ".bat") {
			t.Error("Unix shim path should not have .bat extension")
		}
	}
}

func TestShimDirectory(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	shimDir := manager.GetShimDirectory()
	if shimDir == "" {
		t.Error("Shim directory should not be empty")
	}

	// Check directory exists
	if _, err := os.Stat(shimDir); os.IsNotExist(err) {
		t.Errorf("Shim directory should exist: %s", shimDir)
	}

	// Check it's actually a directory
	if info, err := os.Stat(shimDir); err == nil && !info.IsDir() {
		t.Error("Shim path should be a directory")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
