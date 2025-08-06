// ABOUTME: Tests for cpanfile management functionality in PVI package
// ABOUTME: Covers formatting fixes for trailing newlines and dependency insertion

package pvi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCpanfileManager_TrailingNewline(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")
	manager := NewCpanfileManager(cpanfilePath)

	// Test case 1: Create new cpanfile - should have trailing newline
	err := manager.AddDependency("Test::More", "1.0", false)
	if err != nil {
		t.Fatalf("Failed to add dependency: %v", err)
	}

	content, err := os.ReadFile(cpanfilePath)
	if err != nil {
		t.Fatalf("Failed to read cpanfile: %v", err)
	}

	if !strings.HasSuffix(string(content), "\n") {
		t.Error("New cpanfile should end with trailing newline")
	}

	// Test case 2: Update existing cpanfile - should preserve trailing newline
	err = manager.AddDependency("Moose", "2.0", false)
	if err != nil {
		t.Fatalf("Failed to add second dependency: %v", err)
	}

	content, err = os.ReadFile(cpanfilePath)
	if err != nil {
		t.Fatalf("Failed to read updated cpanfile: %v", err)
	}

	if !strings.HasSuffix(string(content), "\n") {
		t.Error("Updated cpanfile should end with trailing newline")
	}
}

func TestCpanfileManager_InsertionLocation(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")

	// Create initial cpanfile with perl version and existing dependency
	initialContent := `# cpanfile
requires 'perl', '5.038000';
requires 'Existing::Module';

on 'develop' => sub {
    requires 'Test::Harness';
};`

	err := os.WriteFile(cpanfilePath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial cpanfile: %v", err)
	}

	manager := NewCpanfileManager(cpanfilePath)

	// Add a new runtime dependency
	err = manager.AddDependency("New::Module", "1.0", false)
	if err != nil {
		t.Fatalf("Failed to add new dependency: %v", err)
	}

	content, err := os.ReadFile(cpanfilePath)
	if err != nil {
		t.Fatalf("Failed to read updated cpanfile: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find the positions of key elements
	perlIndex := -1
	newModuleIndex := -1
	existingModuleIndex := -1
	developIndex := -1

	for i, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.Contains(line, "requires 'perl'"):
			perlIndex = i
		case strings.Contains(line, "requires 'New::Module'"):
			newModuleIndex = i
		case strings.Contains(line, "requires 'Existing::Module'"):
			existingModuleIndex = i
		case strings.Contains(line, "on 'develop'"):
			developIndex = i
		}
	}

	// Verify positions
	if perlIndex == -1 {
		t.Error("Perl version requirement not found")
	}
	if newModuleIndex == -1 {
		t.Error("New module requirement not found")
	}
	if existingModuleIndex == -1 {
		t.Error("Existing module requirement not found")
	}
	if developIndex == -1 {
		t.Error("Develop block not found")
	}

	// New module should be inserted after perl version but before develop block
	if newModuleIndex <= perlIndex {
		t.Errorf("New module (line %d) should be after perl version (line %d)", newModuleIndex, perlIndex)
	}
	if developIndex != -1 && newModuleIndex >= developIndex {
		t.Errorf("New module (line %d) should be before develop block (line %d)", newModuleIndex, developIndex)
	}

	// Content should still end with newline
	if !strings.HasSuffix(string(content), "\n") {
		t.Error("Updated cpanfile should end with trailing newline")
	}
}

func TestCpanfileManager_InsertionLocation_NoPerlVersion(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")

	// Create cpanfile without perl version requirement
	initialContent := `# cpanfile
requires 'Existing::Module';`

	err := os.WriteFile(cpanfilePath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial cpanfile: %v", err)
	}

	manager := NewCpanfileManager(cpanfilePath)

	// Add a new runtime dependency
	err = manager.AddDependency("New::Module", "1.0", false)
	if err != nil {
		t.Fatalf("Failed to add new dependency: %v", err)
	}

	content, err := os.ReadFile(cpanfilePath)
	if err != nil {
		t.Fatalf("Failed to read updated cpanfile: %v", err)
	}

	// Should contain both modules
	if !strings.Contains(string(content), "requires 'Existing::Module'") {
		t.Error("Should contain existing module")
	}
	if !strings.Contains(string(content), "requires 'New::Module'") {
		t.Error("Should contain new module")
	}

	// Should end with newline
	if !strings.HasSuffix(string(content), "\n") {
		t.Error("Cpanfile should end with trailing newline")
	}
}

func TestCpanfileManager_DevelopDependencies(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")
	manager := NewCpanfileManager(cpanfilePath)

	// Add a develop dependency
	err := manager.AddDependency("Test::More", "1.0", true)
	if err != nil {
		t.Fatalf("Failed to add develop dependency: %v", err)
	}

	content, err := os.ReadFile(cpanfilePath)
	if err != nil {
		t.Fatalf("Failed to read cpanfile: %v", err)
	}

	contentStr := string(content)

	// Should contain develop block
	if !strings.Contains(contentStr, "on 'develop' => sub {") {
		t.Error("Should contain develop block")
	}
	if !strings.Contains(contentStr, "requires 'Test::More'") {
		t.Error("Should contain Test::More in develop block")
	}

	// Should end with newline
	if !strings.HasSuffix(contentStr, "\n") {
		t.Error("Cpanfile should end with trailing newline")
	}
}

func TestCpanfileManager_MixedDependencies(t *testing.T) {
	tempDir := t.TempDir()
	cpanfilePath := filepath.Join(tempDir, "cpanfile")

	// Create cpanfile with perl version
	initialContent := `requires 'perl', '5.038000';`

	err := os.WriteFile(cpanfilePath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial cpanfile: %v", err)
	}

	manager := NewCpanfileManager(cpanfilePath)

	// Add runtime dependency
	err = manager.AddDependency("Moose", "2.0", false)
	if err != nil {
		t.Fatalf("Failed to add runtime dependency: %v", err)
	}

	// Add develop dependency
	err = manager.AddDependency("Test::More", "1.0", true)
	if err != nil {
		t.Fatalf("Failed to add develop dependency: %v", err)
	}

	content, err := os.ReadFile(cpanfilePath)
	if err != nil {
		t.Fatalf("Failed to read cpanfile: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find positions
	perlIndex := -1
	mooseIndex := -1
	developIndex := -1

	for i, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.Contains(line, "requires 'perl'"):
			perlIndex = i
		case strings.Contains(line, "requires 'Moose'"):
			mooseIndex = i
		case strings.Contains(line, "on 'develop'"):
			developIndex = i
		}
	}

	// Verify order: perl -> moose -> develop block
	if perlIndex == -1 || mooseIndex == -1 || developIndex == -1 {
		t.Fatal("Could not find all required elements in cpanfile")
	}

	if mooseIndex <= perlIndex {
		t.Errorf("Moose (line %d) should be after perl (line %d)", mooseIndex, perlIndex)
	}
	if developIndex <= mooseIndex {
		t.Errorf("Develop block (line %d) should be after Moose (line %d)", developIndex, mooseIndex)
	}

	// Should end with newline
	contentStr := string(content)
	if !strings.HasSuffix(contentStr, "\n") {
		t.Error("Cpanfile should end with trailing newline")
	}
}
