package tools

import (
	"context"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/mcp/generation"
	"tamarou.com/pvm/internal/mcp/validation"
)

// Mock implementations for testing (avoid conflicts with existing mocks)

type testValidator struct{}

func (m *testValidator) ValidateCode(ctx context.Context, code string, projectPath string) (*validation.ValidationResult, error) {
	return &validation.ValidationResult{
		Valid:     true,
		Errors:    []validation.ValidationError{},
		Warnings:  []validation.ValidationWarning{},
		TypeInfo:  make(map[string]validation.TypeInfo),
		Timestamp: time.Now(),
	}, nil
}

type testAutoFixer struct{}

func (m *testAutoFixer) AutoFix(ctx context.Context, code string, errors []validation.ValidationError, projectPath string) ([]validation.FixError, error) {
	return []validation.FixError{}, nil
}

func createTestCodeGenerator() *CodeGenerator {
	logger := log.NewLogger(log.LevelError, nil, "test")
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(100)
	validator := &testValidator{}
	autoFixer := &testAutoFixer{}

	return NewCodeGenerator(validator, autoFixer, samplingClient, memoryManager, logger)
}

func TestNewCodeGenerator(t *testing.T) {
	cg := createTestCodeGenerator()
	if cg == nil {
		t.Fatal("NewCodeGenerator returned nil")
	}
	if cg.validator == nil {
		t.Error("CodeGenerator validator is nil")
	}
	if cg.autoFixer == nil {
		t.Error("CodeGenerator autoFixer is nil")
	}
	if cg.samplingClient == nil {
		t.Error("CodeGenerator samplingClient is nil")
	}
	if cg.memoryManager == nil {
		t.Error("CodeGenerator memoryManager is nil")
	}
}

func TestCodeGenerator_Generate_Function(t *testing.T) {
	cg := createTestCodeGenerator()

	request := GenerationRequest{
		Type:          "function",
		Specification: "Create a function to add two numbers",
		Context:       "",
		ProjectPath:   "/test/project",
		SessionID:     "test-session-1",
	}

	result, err := cg.Generate(request)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result == nil {
		t.Fatal("Generate returned nil result")
	}

	if result.Status != "success" && result.Status != "success_with_warnings" {
		t.Errorf("Expected success status, got %s", result.Status)
	}

	if result.GeneratedCode == "" {
		t.Error("Generated code is empty")
	}

	// Check that generated code contains function-like structure
	code := result.GeneratedCode
	if !strings.Contains(code, "sub ") && !strings.Contains(code, "add") {
		t.Error("Generated code doesn't appear to contain a function")
	}

	if result.Iterations <= 0 {
		t.Error("Expected at least 1 iteration")
	}
}

func TestCodeGenerator_Generate_Class(t *testing.T) {
	cg := createTestCodeGenerator()

	request := GenerationRequest{
		Type:          "class",
		Specification: "Create a Person class with name and age",
		Context:       "",
		ProjectPath:   "/test/project",
		SessionID:     "test-session-2",
	}

	result, err := cg.Generate(request)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result == nil {
		t.Fatal("Generate returned nil result")
	}

	if result.Status != "success" && result.Status != "success_with_warnings" {
		t.Errorf("Expected success status, got %s", result.Status)
	}

	if result.GeneratedCode == "" {
		t.Error("Generated code is empty")
	}

	// Check that generated code contains class-like structure
	code := result.GeneratedCode
	if !strings.Contains(code, "package ") && !strings.Contains(code, "class ") {
		t.Error("Generated code doesn't appear to contain a class")
	}
}

func TestCodeGenerator_Generate_Test(t *testing.T) {
	cg := createTestCodeGenerator()

	request := GenerationRequest{
		Type:          "test",
		Specification: "Create tests for add_numbers function",
		Context:       "sub add_numbers { my ($a, $b) = @_; return $a + $b; }",
		ProjectPath:   "/test/project",
		SessionID:     "test-session-3",
	}

	result, err := cg.Generate(request)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result == nil {
		t.Fatal("Generate returned nil result")
	}

	if result.Status != "success" && result.Status != "success_with_warnings" {
		t.Errorf("Expected success status, got %s", result.Status)
	}

	if result.GeneratedCode == "" {
		t.Error("Generated code is empty")
	}

	// Check that generated code contains test-like structure
	code := result.GeneratedCode
	if !strings.Contains(code, "ok(") && !strings.Contains(code, "is(") {
		t.Error("Generated code doesn't appear to contain test assertions")
	}

	if !strings.Contains(code, "done_testing") {
		t.Error("Generated code doesn't contain done_testing()")
	}
}

func TestCodeGenerator_Generate_InvalidType(t *testing.T) {
	cg := createTestCodeGenerator()

	request := GenerationRequest{
		Type:          "invalid_type",
		Specification: "Create something invalid",
		Context:       "",
		ProjectPath:   "/test/project",
		SessionID:     "test-session-4",
	}

	_, err := cg.Generate(request)
	if err == nil {
		t.Error("Expected error for invalid generation type")
	}

	if err != nil && !strings.Contains(err.Error(), "unsupported generation type") {
		t.Errorf("Expected unsupported generation type error, got: %v", err)
	}
}

func TestCodeGenerator_ValidateAndFix_WithErrors(t *testing.T) {
	cg := createTestCodeGenerator()
	memory := cg.memoryManager.CreateSession("test-validation")

	// Test with code that should pass validation (using mock validator)
	code := "my $test = 42;"
	result, err := cg.validateAndFix(code, memory)

	if err != nil {
		t.Fatalf("validateAndFix failed: %v", err)
	}

	if result == nil {
		t.Fatal("validateAndFix returned nil result")
	}

	if !result.Valid {
		t.Error("Expected valid result from mock validator")
	}
}

func TestCodeGenerator_ScoreGeneration(t *testing.T) {
	cg := createTestCodeGenerator()

	tests := []struct {
		name     string
		code     string
		codeType string
		minScore float64
	}{
		{
			name:     "function with good practices",
			code:     "use v5.40;\nsub test { my $x = 1; return $x; } # good function",
			codeType: "function",
			minScore: 0.7,
		},
		{
			name:     "test with assertions",
			code:     "use Test2::V0;\nok(1, 'test'); done_testing();",
			codeType: "test",
			minScore: 0.7,
		},
		{
			name:     "simple class",
			code:     "package MyClass; sub new { bless {}, shift; }",
			codeType: "class",
			minScore: 0.4,
		},
		{
			name:     "minimal code",
			code:     "1;",
			codeType: "function",
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := cg.scoreGeneration(tt.code, tt.codeType)
			if score < tt.minScore {
				t.Errorf("%s: expected score >= %.2f, got %.2f", tt.name, tt.minScore, score)
			}
			if score > 1.0 {
				t.Errorf("%s: score should not exceed 1.0, got %.2f", tt.name, score)
			}
		})
	}
}

func TestCodeGenerator_MemoryIntegration(t *testing.T) {
	cg := createTestCodeGenerator()

	// First generation to establish memory
	request1 := GenerationRequest{
		Type:          "function",
		Specification: "Create add function",
		Context:       "",
		ProjectPath:   "/test/project",
		SessionID:     "memory-test",
	}

	result1, err := cg.Generate(request1)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	// Check that memory context is populated
	if len(result1.MemoryContext) == 0 {
		t.Error("Expected memory context to be populated")
	}

	if len(result1.Decisions) == 0 {
		t.Error("Expected decisions to be recorded")
	}

	// Second generation using same session should have memory
	request2 := GenerationRequest{
		Type:          "function",
		Specification: "Create subtract function",
		Context:       "",
		ProjectPath:   "/test/project",
		SessionID:     "memory-test", // Same session
	}

	result2, err := cg.Generate(request2)
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	// Should have more decisions due to accumulated memory
	if len(result2.Decisions) <= len(result1.Decisions) {
		t.Error("Expected more decisions in second generation due to memory")
	}
}

func TestValidationHelpers(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		function func(string) bool
		expected bool
	}{
		{"valid snake_case", "snake_case", isValidNamingConvention, true},
		{"valid camelCase", "camelCase", isValidNamingConvention, true},
		{"valid PascalCase", "PascalCase", isValidNamingConvention, true},
		{"invalid convention", "invalid-convention", isValidNamingConvention, false},
		{"valid modern_class", "modern_class", isValidClassPattern, true},
		{"valid moose", "moose", isValidClassPattern, true},
		{"valid classic_bless", "classic_bless", isValidClassPattern, true},
		{"invalid pattern", "invalid_pattern", isValidClassPattern, false},
		{"valid Test2::V0", "Test2::V0", isValidTestFramework, true},
		{"valid Test::More", "Test::More", isValidTestFramework, true},
		{"valid Test::Most", "Test::Most", isValidTestFramework, true},
		{"invalid framework", "Invalid::Framework", isValidTestFramework, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.value)
			if result != tt.expected {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, result)
			}
		})
	}
}

func TestCodeGenerator_BuildPrompts(t *testing.T) {
	cg := createTestCodeGenerator()
	memory := cg.memoryManager.CreateSession("prompt-test")

	request := GenerationRequest{
		Type:          "function",
		Specification: "Create a test function",
		Context:       "some context",
		ProjectPath:   "/test",
		SessionID:     "prompt-test",
	}

	// Test function prompt building
	functionPrompt := cg.buildFunctionPrompt(request, memory, "snake_case")
	if !strings.Contains(functionPrompt, "Create a test function") {
		t.Error("Function prompt should contain specification")
	}
	if !strings.Contains(functionPrompt, "snake_case") {
		t.Error("Function prompt should contain naming convention")
	}
	if !strings.Contains(functionPrompt, "some context") {
		t.Error("Function prompt should contain context")
	}

	// Test class prompt building
	classPrompt := cg.buildClassPrompt(request, memory, "modern_class")
	if !strings.Contains(classPrompt, "Create a test function") {
		t.Error("Class prompt should contain specification")
	}
	if !strings.Contains(classPrompt, "modern_class") {
		t.Error("Class prompt should contain class pattern")
	}

	// Test test prompt building
	testPrompt := cg.buildTestPrompt(request, memory, "Test2::V0")
	if !strings.Contains(testPrompt, "Create a test function") {
		t.Error("Test prompt should contain specification")
	}
	if !strings.Contains(testPrompt, "Test2::V0") {
		t.Error("Test prompt should contain test framework")
	}
}

func TestCodeGenerator_IdentifyIssues(t *testing.T) {
	cg := createTestCodeGenerator()

	tests := []struct {
		name      string
		code      string
		score     float64
		maxIssues int
	}{
		{
			name:      "good code",
			code:      "use v5.40; # good comment\nsub test { return 1; }",
			score:     0.9,
			maxIssues: 0,
		},
		{
			name:      "missing version",
			code:      "sub test { return 1; }",
			score:     0.6,
			maxIssues: 2,
		},
		{
			name:      "very poor score",
			code:      "1",
			score:     0.2,
			maxIssues: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := cg.identifyIssues(tt.code, tt.score)
			if len(issues) > tt.maxIssues {
				t.Errorf("%s: expected <= %d issues, got %d: %v", tt.name, tt.maxIssues, len(issues), issues)
			}
		})
	}
}
