// ABOUTME: Comprehensive tests for cpanfile management functionality
// ABOUTME: Tests cpanfile parsing, writing, snapshot generation, and validation

package dependencies

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mockLogger is a simple logger for testing
type mockLogger struct {
	logs []string
}

func (l *mockLogger) Printf(format string, args ...interface{}) {
	// Store log message for testing
	l.logs = append(l.logs, fmt.Sprintf(format, args...))
}

func TestNewCpanfileManager(t *testing.T) {
	tempDir := t.TempDir()
	logger := &mockLogger{}

	manager := NewCpanfileManager(tempDir, logger)

	expectedPath := filepath.Join(tempDir, "cpanfile")
	if manager.Path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, manager.Path)
	}

	if manager.ProjectDir != tempDir {
		t.Errorf("Expected project dir %s, got %s", tempDir, manager.ProjectDir)
	}
}

func TestLoadCpanfile_NonExistent(t *testing.T) {
	tempDir := t.TempDir()
	logger := &mockLogger{}
	manager := NewCpanfileManager(tempDir, logger)

	cpanfile, err := manager.LoadCpanfile()
	if err != nil {
		t.Fatalf("Expected no error for non-existent cpanfile, got %v", err)
	}

	if cpanfile == nil {
		t.Fatal("Expected cpanfile to be non-nil")
	}

	if len(cpanfile.Requirements) != 0 {
		t.Errorf("Expected empty requirements, got %d", len(cpanfile.Requirements))
	}
}

func TestLoadCpanfile_Existing(t *testing.T) {
	tempDir := t.TempDir()
	logger := &mockLogger{}
	manager := NewCpanfileManager(tempDir, logger)

	// Create a test cpanfile
	content := `requires 'Moose', '2.0';
requires 'DBI';

on 'develop' => sub {
    requires 'Perl::Critic';
};`

	err := os.WriteFile(manager.Path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test cpanfile: %v", err)
	}

	cpanfile, err := manager.LoadCpanfile()
	if err != nil {
		t.Fatalf("Failed to load cpanfile: %v", err)
	}

	if len(cpanfile.Requirements) == 0 {
		t.Error("Expected requirements to be loaded")
	}

	// Check that we have the expected modules
	foundMoose := false
	foundDBI := false
	foundCritic := false

	for _, req := range cpanfile.Requirements {
		switch req.Module {
		case "Moose":
			foundMoose = true
			if req.Version != "2.0" {
				t.Errorf("Expected Moose version 2.0, got %s", req.Version)
			}
		case "DBI":
			foundDBI = true
		case "Perl::Critic":
			foundCritic = true
			if req.Phase != "develop" {
				t.Errorf("Expected Perl::Critic to be in develop phase, got %s", req.Phase)
			}
		}
	}

	if !foundMoose {
		t.Error("Expected to find Moose requirement")
	}
	if !foundDBI {
		t.Error("Expected to find DBI requirement")
	}
	if !foundCritic {
		t.Error("Expected to find Perl::Critic requirement")
	}
}

func TestAddDependency(t *testing.T) {
	tempDir := t.TempDir()
	logger := &mockLogger{}
	manager := NewCpanfileManager(tempDir, logger)

	// Add a dependency
	err := manager.AddDependency("Test::More", "1.0", "runtime")
	if err != nil {
		t.Fatalf("Failed to add dependency: %v", err)
	}

	// Verify it was added
	cpanfile, err := manager.LoadCpanfile()
	if err != nil {
		t.Fatalf("Failed to load cpanfile after adding dependency: %v", err)
	}

	found := false
	for _, req := range cpanfile.Requirements {
		if req.Module == "Test::More" {
			found = true
			if req.Version != "1.0" {
				t.Errorf("Expected version 1.0, got %s", req.Version)
			}
			if req.Phase != "runtime" {
				t.Errorf("Expected runtime phase, got %s", req.Phase)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find Test::More dependency")
	}
}

func TestAddDependency_UpdateExisting(t *testing.T) {
	tempDir := t.TempDir()
	logger := &mockLogger{}
	manager := NewCpanfileManager(tempDir, logger)

	// Add initial dependency
	err := manager.AddDependency("Test::More", "1.0", "runtime")
	if err != nil {
		t.Fatalf("Failed to add initial dependency: %v", err)
	}

	// Update the dependency
	err = manager.AddDependency("Test::More", "2.0", "runtime")
	if err != nil {
		t.Fatalf("Failed to update dependency: %v", err)
	}

	// Verify it was updated
	cpanfile, err := manager.LoadCpanfile()
	if err != nil {
		t.Fatalf("Failed to load cpanfile after updating dependency: %v", err)
	}

	count := 0
	for _, req := range cpanfile.Requirements {
		if req.Module == "Test::More" {
			count++
			if req.Version != "2.0" {
				t.Errorf("Expected version 2.0, got %s", req.Version)
			}
		}
	}

	if count == 0 {
		t.Error("Expected to find Test::More dependency")
	} else if count > 1 {
		t.Errorf("Expected only one Test::More dependency, found %d", count)
	}
}

func TestRemoveDependency(t *testing.T) {
	tempDir := t.TempDir()
	logger := &mockLogger{}
	manager := NewCpanfileManager(tempDir, logger)

	// Add a dependency
	err := manager.AddDependency("Test::More", "1.0", "runtime")
	if err != nil {
		t.Fatalf("Failed to add dependency: %v", err)
	}

	// Remove the dependency
	err = manager.RemoveDependency("Test::More", "runtime")
	if err != nil {
		t.Fatalf("Failed to remove dependency: %v", err)
	}

	// Verify it was removed
	cpanfile, err := manager.LoadCpanfile()
	if err != nil {
		t.Fatalf("Failed to load cpanfile after removing dependency: %v", err)
	}

	for _, req := range cpanfile.Requirements {
		if req.Module == "Test::More" {
			t.Error("Expected Test::More to be removed")
		}
	}
}

func TestRemoveDependency_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	logger := &mockLogger{}
	manager := NewCpanfileManager(tempDir, logger)

	// Try to remove a non-existent dependency
	err := manager.RemoveDependency("NonExistent::Module", "runtime")
	if err == nil {
		t.Error("Expected error when removing non-existent dependency")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error to contain 'not found', got: %v", err)
	}
}

func TestGenerateSnapshot(t *testing.T) {
	tempDir := t.TempDir()
	logger := &mockLogger{}
	manager := NewCpanfileManager(tempDir, logger)

	// Add some dependencies
	err := manager.AddDependency("Test::More", "1.0", "runtime")
	if err != nil {
		t.Fatalf("Failed to add dependency: %v", err)
	}

	err = manager.AddDependency("Moose", "2.0", "runtime")
	if err != nil {
		t.Fatalf("Failed to add dependency: %v", err)
	}

	// Generate snapshot
	snapshot, err := manager.GenerateSnapshot()
	if err != nil {
		t.Fatalf("Failed to generate snapshot: %v", err)
	}

	if snapshot == nil {
		t.Fatal("Expected snapshot to be non-nil")
	}

	if snapshot.GeneratedBy != "PVM" {
		t.Errorf("Expected GeneratedBy to be 'PVM', got %s", snapshot.GeneratedBy)
	}

	if len(snapshot.Modules) == 0 {
		t.Error("Expected snapshot to contain modules")
	}

	// Check that modules are included
	foundTestMore := false
	foundMoose := false
	for _, module := range snapshot.Modules {
		if module.Name == "Test::More" {
			foundTestMore = true
		}
		if module.Name == "Moose" {
			foundMoose = true
		}
	}

	if !foundTestMore {
		t.Error("Expected to find Test::More in snapshot")
	}
	if !foundMoose {
		t.Error("Expected to find Moose in snapshot")
	}
}

func TestWriteAndReadSnapshot(t *testing.T) {
	tempDir := t.TempDir()
	logger := &mockLogger{}
	manager := NewCpanfileManager(tempDir, logger)

	// Create a test snapshot
	snapshot := &Snapshot{
		GeneratedAt: time.Now(),
		GeneratedBy: "PVM",
		PerlVersion: "5.38.0",
		Modules: []*SnapshotModule{
			{
				Name:         "Test::More",
				Version:      "1.302195",
				Distribution: "Test-Simple-1.302195",
				Source:       "E/EX/EXODIST/Test-Simple-1.302195.tar.gz",
				Dependencies: []string{"Scalar::Util"},
			},
			{
				Name:         "Moose",
				Version:      "2.2206",
				Distribution: "Moose-2.2206",
				Source:       "E/ET/ETHER/Moose-2.2206.tar.gz",
				Dependencies: []string{"Class::MOP", "Data::OptList"},
			},
		},
	}

	// Write snapshot
	err := manager.WriteSnapshot(snapshot)
	if err != nil {
		t.Fatalf("Failed to write snapshot: %v", err)
	}

	// Read snapshot back
	readSnapshot, err := manager.ReadSnapshot()
	if err != nil {
		t.Fatalf("Failed to read snapshot: %v", err)
	}

	if readSnapshot == nil {
		t.Fatal("Expected read snapshot to be non-nil")
	}

	if len(readSnapshot.Modules) != len(snapshot.Modules) {
		t.Errorf("Expected %d modules, got %d", len(snapshot.Modules), len(readSnapshot.Modules))
	}

	// Check that modules were preserved
	for _, originalModule := range snapshot.Modules {
		found := false
		for _, readModule := range readSnapshot.Modules {
			if readModule.Name == originalModule.Name {
				found = true
				if readModule.Version != originalModule.Version {
					t.Errorf("Expected module %s version %s, got %s",
						originalModule.Name, originalModule.Version, readModule.Version)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected to find module %s in read snapshot", originalModule.Name)
		}
	}
}

func TestValidateSnapshot(t *testing.T) {
	tempDir := t.TempDir()
	logger := &mockLogger{}
	manager := NewCpanfileManager(tempDir, logger)

	// Add dependencies to cpanfile
	err := manager.AddDependency("Test::More", "1.0", "runtime")
	if err != nil {
		t.Fatalf("Failed to add dependency: %v", err)
	}

	// Create matching snapshot
	snapshot := &Snapshot{
		GeneratedAt: time.Now(),
		GeneratedBy: "PVM",
		Modules: []*SnapshotModule{
			{
				Name:    "Test::More",
				Version: "1.302195",
			},
		},
	}

	// Validate snapshot - should pass
	err = manager.ValidateSnapshot(snapshot)
	if err != nil {
		t.Errorf("Expected snapshot validation to pass, got error: %v", err)
	}

	// Create snapshot missing required module
	incompleteSnapshot := &Snapshot{
		GeneratedAt: time.Now(),
		GeneratedBy: "PVM",
		Modules:     []*SnapshotModule{},
	}

	// Validate incomplete snapshot - should fail
	err = manager.ValidateSnapshot(incompleteSnapshot)
	if err == nil {
		t.Error("Expected snapshot validation to fail for incomplete snapshot")
	}

	if !strings.Contains(err.Error(), "missing from snapshot") {
		t.Errorf("Expected error to mention missing module, got: %v", err)
	}
}

func TestCreateNewCpanfile(t *testing.T) {
	tempDir := t.TempDir()
	logger := &mockLogger{}
	manager := NewCpanfileManager(tempDir, logger)

	cpanfile := &CPANFile{
		Requirements: []Requirement{
			{
				Module:       "Test::More",
				Version:      "1.0",
				Phase:        "runtime",
				Relationship: "requires",
			},
			{
				Module:       "Perl::Critic",
				Version:      "",
				Phase:        "develop",
				Relationship: "requires",
			},
		},
		Features:  make(map[string][]Requirement),
		Platforms: make(map[string][]Requirement),
	}

	err := manager.SaveCpanfile(cpanfile)
	if err != nil {
		t.Fatalf("Failed to save cpanfile: %v", err)
	}

	// Verify file was created
	if !fileExists(manager.Path) {
		t.Error("Expected cpanfile to be created")
	}

	// Read content and verify structure
	content, err := os.ReadFile(manager.Path)
	if err != nil {
		t.Fatalf("Failed to read created cpanfile: %v", err)
	}

	contentStr := string(content)

	// Check that runtime requirements are present
	if !strings.Contains(contentStr, "requires 'Test::More', '1.0';") {
		t.Error("Expected to find Test::More requirement in cpanfile")
	}

	// Check that develop block is present
	if !strings.Contains(contentStr, "on 'develop' => sub {") {
		t.Error("Expected to find develop block in cpanfile")
	}

	if !strings.Contains(contentStr, "requires 'Perl::Critic';") {
		t.Error("Expected to find Perl::Critic requirement in develop block")
	}
}

func TestNoBackupCreation(t *testing.T) {
	tempDir := t.TempDir()
	logger := &mockLogger{}
	manager := NewCpanfileManager(tempDir, logger)

	// Create initial cpanfile
	initialContent := "requires 'Initial::Module';"
	err := os.WriteFile(manager.Path, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial cpanfile: %v", err)
	}

	// Save updated cpanfile
	cpanfile := &CPANFile{
		Requirements: []Requirement{
			{
				Module:       "Updated::Module",
				Phase:        "runtime",
				Relationship: "requires",
			},
		},
		Features:  make(map[string][]Requirement),
		Platforms: make(map[string][]Requirement),
	}

	err = manager.SaveCpanfile(cpanfile)
	if err != nil {
		t.Fatalf("Failed to save updated cpanfile: %v", err)
	}

	// Check that NO backup was created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "cpanfile.backup.") {
			t.Errorf("Backup file should not be created, but found: %s", file.Name())
		}
	}
}

func TestLogging(t *testing.T) {
	tempDir := t.TempDir()
	logger := &mockLogger{}
	manager := NewCpanfileManager(tempDir, logger)

	// Add a dependency first so there's something to check
	err := manager.AddDependency("NonExistent::Module", "1.0", "runtime")
	if err != nil {
		t.Fatalf("Failed to add dependency: %v", err)
	}

	// Generate snapshot (which should log warnings for missing modules)
	_, err = manager.GenerateSnapshot()
	if err != nil {
		t.Fatalf("Failed to generate snapshot: %v", err)
	}

	// Check that logger was used (should have warnings about missing modules)
	if len(logger.logs) == 0 {
		t.Error("Expected logger to be used")
	}

	// Check that the log contains the expected warning
	found := false
	for _, log := range logger.logs {
		if strings.Contains(log, "Warning: module NonExistent::Module not installed") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find warning about missing module in logs: %v", logger.logs)
	}
}
