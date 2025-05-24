package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/mcp/generation"
	"tamarou.com/pvm/internal/mcp/validation"
)

// Mock implementations for testing
type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) ValidateCode(ctx context.Context, code string, projectPath string) (*validation.ValidationResult, error) {
	args := m.Called(code)
	return args.Get(0).(*validation.ValidationResult), args.Error(1)
}

type MockAutoFixer struct {
	mock.Mock
}

func (m *MockAutoFixer) AutoFix(ctx context.Context, code string, errors []validation.ValidationError, projectPath string) ([]validation.FixError, error) {
	args := m.Called(code, errors)
	return args.Get(0).([]validation.FixError), args.Error(1)
}

func TestNewCodeGenerator(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	generator := NewCodeGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	assert.NotNil(t, generator)
	assert.Equal(t, validator, generator.validator)
	assert.Equal(t, autoFixer, generator.autoFixer)
	assert.Equal(t, samplingClient, generator.samplingClient)
	assert.Equal(t, memoryManager, generator.memoryManager)
	assert.Equal(t, logger, generator.logger)
}

func TestCodeGenerator_Generate_Function(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	generator := NewCodeGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	// Mock validation response
	validationResult := &validation.ValidationResult{
		Errors:   []validation.ValidationError{},
		Warnings: []validation.ValidationWarning{},
	}
	validator.On("ValidateCode", mock.AnythingOfType("string")).Return(validationResult, nil)

	request := GenerationRequest{
		Type:          "function",
		Specification: "Create a function that adds two numbers",
		Context:       "use v5.40;",
		SessionID:     "test-session",
	}

	result, err := generator.Generate(request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Status)
	assert.NotEmpty(t, result.GeneratedCode)
	assert.Contains(t, result.GeneratedCode, "sub ")
	assert.NotZero(t, result.Iterations)
	assert.NotEmpty(t, result.MemoryContext)
	assert.NotEmpty(t, result.Decisions)

	validator.AssertExpectations(t)
}

func TestCodeGenerator_Generate_Class(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	generator := NewCodeGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	// Mock validation response
	validationResult := &validation.ValidationResult{
		Errors:   []validation.ValidationError{},
		Warnings: []validation.ValidationWarning{},
	}
	validator.On("ValidateCode", mock.AnythingOfType("string")).Return(validationResult, nil)

	request := GenerationRequest{
		Type:          "class",
		Specification: "Create a Person class with name and age fields",
		Context:       "use v5.40;",
		SessionID:     "test-session",
	}

	result, err := generator.Generate(request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Status)
	assert.NotEmpty(t, result.GeneratedCode)
	assert.True(t, strings.Contains(result.GeneratedCode, "class ") || strings.Contains(result.GeneratedCode, "package "))
	assert.NotZero(t, result.Iterations)

	validator.AssertExpectations(t)
}

func TestCodeGenerator_Generate_Test(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	generator := NewCodeGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	// Mock validation response
	validationResult := &validation.ValidationResult{
		Errors:   []validation.ValidationError{},
		Warnings: []validation.ValidationWarning{},
	}
	validator.On("ValidateCode", mock.AnythingOfType("string")).Return(validationResult, nil)

	request := GenerationRequest{
		Type:          "test",
		Specification: "Create tests for a calculator function",
		Context:       "sub add { return $_[0] + $_[1]; }",
		SessionID:     "test-session",
	}

	result, err := generator.Generate(request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Status)
	assert.NotEmpty(t, result.GeneratedCode)
	assert.True(t, strings.Contains(result.GeneratedCode, "ok(") || strings.Contains(result.GeneratedCode, "is("))
	assert.NotZero(t, result.Iterations)

	validator.AssertExpectations(t)
}

func TestCodeGenerator_Generate_InvalidType(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	generator := NewCodeGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	request := GenerationRequest{
		Type:          "invalid",
		Specification: "Something",
		SessionID:     "test-session",
	}

	result, err := generator.Generate(request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported generation type")
}

func TestCodeGenerator_ValidateAndFix_WithErrors(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	generator := &CodeGenerator{
		validator:     validator,
		autoFixer:     autoFixer,
		memoryManager: memoryManager,
		logger:        logger,
	}

	memory := memoryManager.CreateSession("test-session")
	code := "sub broken { my $x = ; }"

	// Mock validation with errors
	errors := []validation.ValidationError{
		{Message: "Syntax error", Line: 1, Column: 20},
	}
	validationResult := &validation.ValidationResult{
		Errors:   errors,
		Warnings: []validation.ValidationWarning{},
	}
	validator.On("ValidateCode", code).Return(validationResult, nil)

	// Mock auto-fix
	fixedCode := "sub broken { my $x = 0; }"
	fixedResult := &validation.ValidationResult{
		Errors:   []validation.ValidationError{},
		Warnings: []validation.ValidationWarning{},
	}
	fixes := []validation.FixError{
		{FixedCode: fixedCode, Explanation: "Fixed syntax", Confidence: 0.9},
	}
	autoFixer.On("AutoFix", code, errors).Return(fixes, nil)
	validator.On("ValidateCode", fixedCode).Return(fixedResult, nil)

	result, err := generator.validateAndFix(code, memory)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Errors)

	validator.AssertExpectations(t)
	autoFixer.AssertExpectations(t)
}

func TestCodeGenerator_ScoreGeneration(t *testing.T) {
	generator := &CodeGenerator{}

	tests := []struct {
		name     string
		code     string
		codeType string
		minScore float64
	}{
		{
			name:     "good_function",
			code:     "use v5.40;\nsub add {\n  my ($a, $b) = @_;\n  return $a + $b;\n}",
			codeType: "function",
			minScore: 0.7,
		},
		{
			name:     "good_class",
			code:     "use v5.40;\nclass Person {\n  field $name;\n  method name { $name }\n}",
			codeType: "class",
			minScore: 0.7,
		},
		{
			name:     "good_test",
			code:     "use Test2::V0;\nok(add(2, 3) == 5, 'addition works');\ndone_testing;",
			codeType: "test",
			minScore: 0.7,
		},
		{
			name:     "poor_code",
			code:     "x",
			codeType: "function",
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := generator.scoreGeneration(tt.code, tt.codeType)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestCodeGenerator_MemoryIntegration(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	generator := NewCodeGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	// Mock validation response
	validationResult := &validation.ValidationResult{
		Errors:   []validation.ValidationError{},
		Warnings: []validation.ValidationWarning{},
	}
	validator.On("ValidateCode", mock.AnythingOfType("string")).Return(validationResult, nil)

	sessionID := "memory-test-session"

	// First generation - should establish naming convention
	request1 := GenerationRequest{
		Type:          "function",
		Specification: "Create a function that adds two numbers",
		SessionID:     sessionID,
	}

	result1, err := generator.Generate(request1)
	assert.NoError(t, err)
	assert.NotNil(t, result1)

	// Check that memory was used
	memory := memoryManager.GetSession(sessionID)
	naming, hasNaming := memory.GetNamingPattern("function")
	assert.True(t, hasNaming)
	assert.NotEmpty(t, naming)

	// Second generation - should reuse naming convention
	request2 := GenerationRequest{
		Type:          "function",
		Specification: "Create a function that multiplies two numbers",
		SessionID:     sessionID,
	}

	result2, err := generator.Generate(request2)
	assert.NoError(t, err)
	assert.NotNil(t, result2)

	// Check that same naming convention was used
	naming2, _ := memory.GetNamingPattern("function")
	assert.Equal(t, naming, naming2)

	// Check that decisions were recorded
	decisions := memory.GetDecisions()
	assert.Greater(t, len(decisions), 2) // At least some decisions recorded

	validator.AssertExpectations(t)
}

func TestValidationHelpers(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		validFn  func(string) bool
		expected bool
	}{
		{"valid_snake_case", "snake_case", isValidNamingConvention, true},
		{"valid_camelCase", "camelCase", isValidNamingConvention, true},
		{"valid_PascalCase", "PascalCase", isValidNamingConvention, true},
		{"invalid_naming", "invalid", isValidNamingConvention, false},
		{"valid_modern_class", "modern_class", isValidClassPattern, true},
		{"valid_moose", "moose", isValidClassPattern, true},
		{"valid_classic_bless", "classic_bless", isValidClassPattern, true},
		{"invalid_class", "invalid", isValidClassPattern, false},
		{"valid_Test2", "Test2::V0", isValidTestFramework, true},
		{"valid_TestMore", "Test::More", isValidTestFramework, true},
		{"valid_TestMost", "Test::Most", isValidTestFramework, true},
		{"invalid_test", "invalid", isValidTestFramework, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.validFn(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeGenerator_BuildPrompts(t *testing.T) {
	generator := &CodeGenerator{}
	memoryManager := generation.NewMemoryManager(50)
	memory := memoryManager.CreateSession("test-session")
	defer memoryManager.Close()

	// Add some decisions to memory
	memory.SetTypeChoice("counter", "Int")
	memory.SetNamingPattern("variable", "snake_case")

	request := GenerationRequest{
		Specification: "Create a counter function",
		Context:       "use v5.40;",
	}

	tests := []struct {
		name     string
		buildFn  func() string
		contains []string
	}{
		{
			name: "function_prompt",
			buildFn: func() string {
				return generator.buildFunctionPrompt(request, memory, "snake_case")
			},
			contains: []string{"Create a counter function", "snake_case", "Type Context"},
		},
		{
			name: "class_prompt",
			buildFn: func() string {
				return generator.buildClassPrompt(request, memory, "modern_class")
			},
			contains: []string{"Create a counter function", "modern_class", "Type Context"},
		},
		{
			name: "test_prompt",
			buildFn: func() string {
				return generator.buildTestPrompt(request, memory, "Test2::V0")
			},
			contains: []string{"Create a counter function", "Test2::V0", "Type Context"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := tt.buildFn()
			assert.NotEmpty(t, prompt)
			for _, expected := range tt.contains {
				assert.Contains(t, prompt, expected)
			}
		})
	}
}

func TestCodeGenerator_IdentifyIssues(t *testing.T) {
	generator := &CodeGenerator{}

	tests := []struct {
		name     string
		code     string
		score    float64
		expected []string
	}{
		{
			name:     "missing_version",
			code:     "sub test { }",
			score:    0.5,
			expected: []string{"Missing modern Perl version declaration"},
		},
		{
			name:     "missing_comments",
			code:     "use v5.40; sub test { }",
			score:    0.5,
			expected: []string{"Missing documentation comments"},
		},
		{
			name:     "low_score",
			code:     "x",
			score:    0.2,
			expected: []string{"Code structure needs improvement", "Missing best practices implementation"},
		},
		{
			name:     "good_code",
			code:     "use v5.40; # Good function\nsub test { return 1; }",
			score:    0.8,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := generator.identifyIssues(tt.code, tt.score)
			for _, expected := range tt.expected {
				found := false
				for _, issue := range issues {
					if strings.Contains(issue, expected) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected issue '%s' not found in %v", expected, issues)
			}
		})
	}
}
