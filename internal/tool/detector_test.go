// ABOUTME: Comprehensive tests for tool execution mode detection
// ABOUTME: Tests mode detection logic and argument parsing functionality

package tool

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewDetector(t *testing.T) {
	detector := NewDetector()

	if detector == nil {
		t.Fatal("NewDetector() returned nil")
	}

	// Test that known tools are properly initialized
	knownTools := []string{"perl", "cpanm", "prove", "perltidy", "perlcritic"}
	for _, tool := range knownTools {
		if !detector.IsKnownTool(tool) {
			t.Errorf("Expected %s to be a known tool", tool)
		}
	}
}

func TestDetectExecutionMode_KnownTools(t *testing.T) {
	detector := NewDetector()

	testCases := []struct {
		name          string
		args          []string
		expectedMode  ExecutionMode
		expectedTool  string
		minConfidence float64
	}{
		{
			name:          "perl tool",
			args:          []string{"perl", "--version"},
			expectedMode:  ModeTool,
			expectedTool:  "perl",
			minConfidence: 1.0,
		},
		{
			name:          "cpanm tool",
			args:          []string{"cpanm", "App::Ack"},
			expectedMode:  ModeTool,
			expectedTool:  "cpanm",
			minConfidence: 1.0,
		},
		{
			name:          "prove tool",
			args:          []string{"prove", "-l", "t/"},
			expectedMode:  ModeTool,
			expectedTool:  "prove",
			minConfidence: 1.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := detector.DetectExecutionMode(tc.args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.Mode != tc.expectedMode {
				t.Errorf("Expected mode %v, got %v", tc.expectedMode, result.Mode)
			}

			if result.ToolName != tc.expectedTool {
				t.Errorf("Expected tool name %s, got %s", tc.expectedTool, result.ToolName)
			}

			if result.Confidence < tc.minConfidence {
				t.Errorf("Expected confidence >= %f, got %f", tc.minConfidence, result.Confidence)
			}
		})
	}
}

func TestDetectExecutionMode_ScriptFiles(t *testing.T) {
	detector := NewDetector()

	// Create a temporary script file for testing
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "test.pl")
	err := os.WriteFile(scriptPath, []byte("#!/usr/bin/env perl\nprint \"Hello\\n\";"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	testCases := []struct {
		name           string
		args           []string
		expectedMode   ExecutionMode
		expectedScript string
		minConfidence  float64
	}{
		{
			name:           "existing script file",
			args:           []string{scriptPath, "arg1", "arg2"},
			expectedMode:   ModeScript,
			expectedScript: scriptPath,
			minConfidence:  0.9,
		},
		{
			name:           "script with .pl extension",
			args:           []string{"script.pl", "arg1"},
			expectedMode:   ModeScript,
			expectedScript: "script.pl",
			minConfidence:  0.8,
		},
		{
			name:           "script with .pm extension",
			args:           []string{"Module.pm"},
			expectedMode:   ModeScript,
			expectedScript: "Module.pm",
			minConfidence:  0.8,
		},
		{
			name:           "script with path",
			args:           []string{"./script", "arg1"},
			expectedMode:   ModeScript,
			expectedScript: "./script",
			minConfidence:  0.7,
		},
		{
			name:           "script with absolute path",
			args:           []string{"/usr/local/bin/script"},
			expectedMode:   ModeScript,
			expectedScript: "/usr/local/bin/script",
			minConfidence:  0.7,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := detector.DetectExecutionMode(tc.args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.Mode != tc.expectedMode {
				t.Errorf("Expected mode %v, got %v", tc.expectedMode, result.Mode)
			}

			if result.ScriptPath != tc.expectedScript {
				t.Errorf("Expected script path %s, got %s", tc.expectedScript, result.ScriptPath)
			}

			if result.Confidence < tc.minConfidence {
				t.Errorf("Expected confidence >= %f, got %f", tc.minConfidence, result.Confidence)
			}
		})
	}
}

func TestDetectExecutionMode_InlineCode(t *testing.T) {
	detector := NewDetector()

	testCases := []struct {
		name         string
		args         []string
		expectedMode ExecutionMode
		expectedCode string
	}{
		{
			name:         "inline code with -e",
			args:         []string{"-e", "print 'hello'"},
			expectedMode: ModeInline,
			expectedCode: "print 'hello'",
		},
		{
			name:         "inline code with --execute",
			args:         []string{"--execute", "use strict; print 'hello'"},
			expectedMode: ModeInline,
			expectedCode: "use strict; print 'hello'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := detector.DetectExecutionMode(tc.args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.Mode != tc.expectedMode {
				t.Errorf("Expected mode %v, got %v", tc.expectedMode, result.Mode)
			}

			if result.InlineCode != tc.expectedCode {
				t.Errorf("Expected inline code %s, got %s", tc.expectedCode, result.InlineCode)
			}
		})
	}
}

func TestDetectExecutionMode_AmbiguousCases(t *testing.T) {
	detector := NewDetector()
	detector.SetOptions(true, false) // Allow ambiguous, prefer tool mode

	testCases := []struct {
		name         string
		args         []string
		expectedMode ExecutionMode
		shouldError  bool
	}{
		{
			name:         "simple name - ambiguous",
			args:         []string{"myapp", "arg1"},
			expectedMode: ModeTool, // Should default to tool mode when ambiguous
			shouldError:  false,
		},
		{
			name:         "hyphenated name - ambiguous",
			args:         []string{"my-app", "arg1"},
			expectedMode: ModeTool,
			shouldError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := detector.DetectExecutionMode(tc.args)

			if tc.shouldError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tc.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result.Mode != tc.expectedMode {
				t.Errorf("Expected mode %v, got %v", tc.expectedMode, result.Mode)
			}
		})
	}
}

func TestDetectExecutionMode_AmbiguousWithStrictMode(t *testing.T) {
	detector := NewDetector()
	detector.SetOptions(false, false) // Don't allow ambiguous

	testCases := []struct {
		name        string
		args        []string
		shouldError bool
		errorType   string
	}{
		{
			name:        "ambiguous simple name",
			args:        []string{"myapp"},
			shouldError: true,
			errorType:   ErrAmbiguousMode,
		},
		{
			name:        "known tool should not error",
			args:        []string{"perl"},
			shouldError: false,
		},
		{
			name:        "clear script should not error",
			args:        []string{"script.pl"},
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := detector.DetectExecutionMode(tc.args)

			if tc.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				} else {
					var toolErr *ToolError
					if !errors.As(err, &toolErr) {
						t.Errorf("Expected ToolError, got %T", err)
					} else if toolErr.Code != tc.errorType {
						t.Errorf("Expected error code %s, got %s", tc.errorType, toolErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
				}
			}
		})
	}
}

func TestDetectExecutionMode_EmptyArgs(t *testing.T) {
	detector := NewDetector()

	result, err := detector.DetectExecutionMode([]string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Mode != ModeAmbiguous {
		t.Errorf("Expected ModeAmbiguous for empty args, got %v", result.Mode)
	}

	if result.Confidence != 0.0 {
		t.Errorf("Expected confidence 0.0 for empty args, got %f", result.Confidence)
	}
}

func TestDetectExecutionMode_ArgumentParsing(t *testing.T) {
	detector := NewDetector()

	testCases := []struct {
		name         string
		args         []string
		expectedArgs []string
	}{
		{
			name:         "tool with args",
			args:         []string{"cpanm", "App::Ack", "--verbose"},
			expectedArgs: []string{"App::Ack", "--verbose"},
		},
		{
			name:         "script with args",
			args:         []string{"script.pl", "arg1", "arg2", "--flag"},
			expectedArgs: []string{"arg1", "arg2", "--flag"},
		},
		{
			name:         "single argument",
			args:         []string{"perl"},
			expectedArgs: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := detector.DetectExecutionMode(tc.args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result.Arguments) != len(tc.expectedArgs) {
				t.Errorf("Expected %d arguments, got %d", len(tc.expectedArgs), len(result.Arguments))
			}

			for i, arg := range tc.expectedArgs {
				if i >= len(result.Arguments) || result.Arguments[i] != arg {
					t.Errorf("Expected argument %d to be %s, got %s", i, arg, result.Arguments[i])
				}
			}
		})
	}
}

func TestValidateToolName(t *testing.T) {
	detector := NewDetector()

	testCases := []struct {
		name        string
		toolName    string
		shouldError bool
		errorCode   string
	}{
		{
			name:        "valid simple name",
			toolName:    "cpanm",
			shouldError: false,
		},
		{
			name:        "valid name with hyphen",
			toolName:    "my-tool",
			shouldError: false,
		},
		{
			name:        "valid name with underscore",
			toolName:    "my_tool",
			shouldError: false,
		},
		{
			name:        "valid name with numbers",
			toolName:    "tool123",
			shouldError: false,
		},
		{
			name:        "empty name",
			toolName:    "",
			shouldError: true,
			errorCode:   ErrInvalidToolName,
		},
		{
			name:        "name starting with number",
			toolName:    "123tool",
			shouldError: true,
			errorCode:   ErrInvalidToolName,
		},
		{
			name:        "name with invalid characters",
			toolName:    "tool@name",
			shouldError: true,
			errorCode:   ErrInvalidToolName,
		},
		{
			name:        "reserved name",
			toolName:    "help",
			shouldError: true,
			errorCode:   ErrInvalidToolName,
		},
		{
			name:        "too long name",
			toolName:    "a" + strings.Repeat("b", 64),
			shouldError: true,
			errorCode:   ErrInvalidToolName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := detector.ValidateToolName(tc.toolName)

			if tc.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				} else {
					var toolErr *ToolError
					if !errors.As(err, &toolErr) {
						t.Errorf("Expected ToolError, got %T", err)
					} else if toolErr.Code != tc.errorCode {
						t.Errorf("Expected error code %s, got %s", tc.errorCode, toolErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAddKnownTool(t *testing.T) {
	detector := NewDetector()

	customTool := "mytool"
	if detector.IsKnownTool(customTool) {
		t.Errorf("Tool %s should not be known initially", customTool)
	}

	detector.AddKnownTool(customTool)

	if !detector.IsKnownTool(customTool) {
		t.Errorf("Tool %s should be known after adding", customTool)
	}

	// Test that it's now detected as a tool
	result, err := detector.DetectExecutionMode([]string{customTool, "arg1"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Mode != ModeTool {
		t.Errorf("Expected ModeTool for added tool, got %v", result.Mode)
	}

	if result.ToolName != customTool {
		t.Errorf("Expected tool name %s, got %s", customTool, result.ToolName)
	}
}

func TestGetKnownTools(t *testing.T) {
	detector := NewDetector()

	tools := detector.GetKnownTools()
	if len(tools) == 0 {
		t.Error("Expected some known tools, got none")
	}

	// Check that common tools are included
	expectedTools := []string{"perl", "cpanm", "prove"}
	for _, expected := range expectedTools {
		found := false
		for _, tool := range tools {
			if tool == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tool %s to be in known tools list", expected)
		}
	}
}

func TestSetOptions(t *testing.T) {
	detector := NewDetector()

	// Test default behavior (no ambiguous allowed)
	detector.SetOptions(false, false)
	result, err := detector.DetectExecutionMode([]string{"ambiguous"})
	if err == nil {
		t.Error("Expected error for ambiguous case with strict mode")
	}

	// Test allowing ambiguous with tool preference
	detector.SetOptions(true, false)
	result, err = detector.DetectExecutionMode([]string{"ambiguous"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Mode != ModeTool {
		t.Errorf("Expected ModeTool with tool preference, got %v", result.Mode)
	}

	// Test allowing ambiguous with script preference
	detector.SetOptions(true, true)
	result, err = detector.DetectExecutionMode([]string{"ambiguous"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Mode != ModeScript {
		t.Errorf("Expected ModeScript with script preference, got %v", result.Mode)
	}
}
