// ABOUTME: Tests for enhanced help system with context-aware suggestions
// ABOUTME: Validates help content generation and command suggestions

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/project"
)

func TestHelpManager_GetContextualHelp(t *testing.T) {
	// Test help in non-project directory
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	os.Chdir(tempDir)
	project.ClearDetectionCache()

	helpManager := NewHelpManager()
	categories := helpManager.GetContextualHelp()

	// Should have basic categories for non-project
	if len(categories) == 0 {
		t.Error("Expected help categories, got none")
	}

	// Should contain getting started
	hasGettingStarted := false
	hasProjectSetup := false
	for _, cat := range categories {
		if cat.Name == "Getting Started" {
			hasGettingStarted = true
		}
		if cat.Name == "Project Setup" {
			hasProjectSetup = true
		}
	}

	if !hasGettingStarted {
		t.Error("Expected Getting Started category")
	}
	if !hasProjectSetup {
		t.Error("Expected Project Setup category for non-project directory")
	}
}

func TestHelpManager_GetContextualHelp_InProject(t *testing.T) {
	// Create a project directory
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create project markers
	os.WriteFile(filepath.Join(tempDir, ".perl-version"), []byte("5.38.0"), 0644)
	os.WriteFile(filepath.Join(tempDir, "cpanfile"), []byte("requires 'DBI';"), 0644)

	os.Chdir(tempDir)
	project.ClearDetectionCache()

	helpManager := NewHelpManager()
	categories := helpManager.GetContextualHelp()

	// Should have project-specific categories
	hasProjectWorkflow := false
	hasBuildAndTest := false
	hasModuleManagement := false

	for _, cat := range categories {
		switch cat.Name {
		case "Project Workflow":
			hasProjectWorkflow = true
		case "Build & Test":
			hasBuildAndTest = true
		case "Module Management":
			hasModuleManagement = true
		}
	}

	if !hasProjectWorkflow {
		t.Error("Expected Project Workflow category in project directory")
	}
	if !hasBuildAndTest {
		t.Error("Expected Build & Test category in project directory")
	}
	if !hasModuleManagement {
		t.Error("Expected Module Management category in project directory")
	}
}

func TestHelpManager_SuggestNextSteps(t *testing.T) {
	// Test in non-project directory
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	os.Chdir(tempDir)
	project.ClearDetectionCache()

	helpManager := NewHelpManager()
	suggestions := helpManager.SuggestNextSteps()

	if len(suggestions) == 0 {
		t.Error("Expected next step suggestions")
	}

	// Should suggest project initialization
	hasInitSuggestion := false
	for _, suggestion := range suggestions {
		if strings.Contains(suggestion, "pvm workspace init") {
			hasInitSuggestion = true
			break
		}
	}

	if !hasInitSuggestion {
		t.Error("Expected workspace initialization suggestion for non-workspace directory")
	}
}

func TestHelpManager_SuggestNextSteps_InProject(t *testing.T) {
	// Create a project directory without cpanfile
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	os.WriteFile(filepath.Join(tempDir, ".perl-version"), []byte("5.38.0"), 0644)

	os.Chdir(tempDir)
	project.ClearDetectionCache()

	helpManager := NewHelpManager()
	suggestions := helpManager.SuggestNextSteps()

	// Should suggest creating cpanfile
	hasCpanfileSuggestion := false
	for _, suggestion := range suggestions {
		if strings.Contains(suggestion, "cpanfile") {
			hasCpanfileSuggestion = true
			break
		}
	}

	if !hasCpanfileSuggestion {
		t.Error("Expected cpanfile creation suggestion for project without cpanfile")
	}
}

func TestHelpManager_GetWorkflowHelp(t *testing.T) {
	helpManager := NewHelpManager()
	workflows := helpManager.GetWorkflowHelp()

	expectedWorkflows := []string{
		"new-project",
		"existing-project",
		"module-development",
		"testing",
		"building",
	}

	for _, expected := range expectedWorkflows {
		if _, exists := workflows[expected]; !exists {
			t.Errorf("Expected workflow '%s' not found", expected)
		}
	}

	// Check content quality
	if content, exists := workflows["new-project"]; exists {
		if !strings.Contains(content, "pvm workspace init") {
			t.Error("new-project workflow should mention pvm workspace init")
		}
	}
}

func TestSuggestCommand(t *testing.T) {
	availableCommands := []string{
		"build", "test", "install", "version", "help", "status", "init",
	}

	tests := []struct {
		input       string
		shouldMatch string // Command that should be in suggestions (if any)
		expectEmpty bool   // Whether to expect no suggestions
	}{
		{
			input:       "biuld",
			shouldMatch: "build",
			expectEmpty: false,
		},
		{
			input:       "tset",
			shouldMatch: "test",
			expectEmpty: false,
		},
		{
			input:       "versio",
			shouldMatch: "version",
			expectEmpty: false,
		},
		{
			input:       "xyz", // No good matches
			shouldMatch: "",
			expectEmpty: true,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			suggestions := SuggestCommand(test.input, availableCommands)

			if test.expectEmpty {
				if len(suggestions) > 0 {
					t.Errorf("Expected no suggestions for '%s', got %v", test.input, suggestions)
				}
				return
			}

			if test.shouldMatch == "" {
				return // No specific expectation
			}

			// Check if expected command is in suggestions
			found := false
			for _, suggestion := range suggestions {
				if suggestion == test.shouldMatch {
					found = true
					break
				}
			}

			if !found && len(suggestions) == 0 {
				t.Errorf("Expected suggestions including '%s' for '%s', got none", test.shouldMatch, test.input)
			} else if !found {
				t.Errorf("Expected suggestions including '%s' for '%s', got %v", test.shouldMatch, test.input, suggestions)
			}
		})
	}
}

func TestCreateHelpCommand(t *testing.T) {
	cmd := CreateHelpCommand()

	if cmd.Use != "guide [topic]" {
		t.Errorf("Expected Use 'guide [topic]', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("Expected non-empty Long description")
	}

	if cmd.RunE == nil {
		t.Error("Expected RunE function to be set")
	}
}

func TestCalculateSimilarity(t *testing.T) {
	tests := []struct {
		s1          string
		s2          string
		expectRange [2]float64 // [min, max] range for expected result
	}{
		{"build", "build", [2]float64{1.0, 1.0}},     // Exact match
		{"build", "biuld", [2]float64{0.6, 0.9}},     // Most chars match, some positional
		{"test", "tset", [2]float64{0.5, 0.8}},       // All chars match, some positional
		{"abc", "xyz", [2]float64{0.0, 0.1}},         // No characters match
		{"", "", [2]float64{0.0, 0.0}},               // Both empty
		{"abc", "", [2]float64{0.0, 0.0}},            // One empty
		{"install", "instal", [2]float64{0.7, 0.95}}, // Close match
	}

	for _, test := range tests {
		t.Run(test.s1+"_"+test.s2, func(t *testing.T) {
			result := calculateSimilarity(test.s1, test.s2)
			if result < test.expectRange[0] || result > test.expectRange[1] {
				t.Errorf("calculateSimilarity('%s', '%s') = %f, expected range [%f, %f]",
					test.s1, test.s2, result, test.expectRange[0], test.expectRange[1])
			}
		})
	}
}

func TestHasTestDirectory(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	os.Chdir(tempDir)

	// Should return false when no test directory exists
	if hasTestDirectory() {
		t.Error("Expected false when no test directory exists")
	}

	// Create t/ directory
	os.Mkdir("t", 0755)
	if !hasTestDirectory() {
		t.Error("Expected true when t/ directory exists")
	}

	// Remove t/ and create test/ directory
	os.RemoveAll("t")
	os.Mkdir("test", 0755)
	if !hasTestDirectory() {
		t.Error("Expected true when test/ directory exists")
	}
}

func TestShowTypesHelp(t *testing.T) {
	// Create a test command and help manager
	cmd := &cobra.Command{
		Use: "test",
	}

	// Create a UI context that captures output
	var output strings.Builder
	ctx := &ui.UIContext{
		Writer:      &output,
		ErrorWriter: &output,
		ColorMode:   ui.ColorNever, // Disable colors for testing
		Quiet:       false,
		Verbose:     false,
		Interactive: false,
		RawMarkdown: true,
	}

	// Temporarily override global UI
	oldUI := globalUI
	defer func() { globalUI = oldUI }()
	globalUI = ui.NewOutput(ctx)

	helpManager := NewHelpManager()

	// Call ShowTypesHelp
	err := ShowTypesHelp(cmd, helpManager)
	if err != nil {
		t.Fatalf("ShowTypesHelp returned unexpected error: %v", err)
	}

	// Verify output contains expected sections
	outputStr := output.String()

	expectedSections := []string{
		"PVM Type System Reference",
		"Basic Type Syntax",
		"Parameterized Types",
		"Union & Intersection Types",
		"Types::Standard Migration",
		"Bool Type Validation",
		"Type Hierarchy Overview",
		"Common Type Patterns",
		"Getting More Help",
	}

	for _, section := range expectedSections {
		if !strings.Contains(outputStr, section) {
			t.Errorf("Expected output to contain section '%s'", section)
		}
	}

	// Check for specific content examples
	expectedContent := []string{
		"my Int $count = 42;",
		"my ArrayRef[Int] @numbers;",
		"my Map[Str, Int] %mapping;",
		"my Int|Str $flexible;",
		"Bool type has strict validation rules:",
		"Any",
		"├── Defined",
		"Maybe[Str] $name",
		"pvm help workflows",
		"pvm dev",
	}

	for _, content := range expectedContent {
		if !strings.Contains(outputStr, content) {
			t.Errorf("Expected output to contain content '%s'", content)
		}
	}

	// Verify no error message in output
	if strings.Contains(outputStr, "Error") || strings.Contains(outputStr, "error") {
		t.Error("Output should not contain error messages")
	}
}
