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

func TestNewAdvancedGenerator(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	advGen := NewAdvancedGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	assert.NotNil(t, advGen)
	assert.NotNil(t, advGen.CodeGenerator)
}

func TestAdvancedGenerator_GenerateTestsFromTypes(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	advGen := NewAdvancedGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	// Mock validation response
	validationResult := &validation.ValidationResult{
		Errors:   []validation.ValidationError{},
		Warnings: []validation.ValidationWarning{},
	}
	validator.On("ValidateCode", mock.AnythingOfType("string")).Return(validationResult, nil)

	request := TestGenerationRequest{
		Code: `sub add {
			my ($x, $y) = @_;
			return $x + $y;
		}`,
		TypeSigs: map[string]string{
			"add": "(Int, Int) -> Int",
		},
		Framework: "Test2::V0",
		SessionID: "test-session",
	}

	result, err := advGen.GenerateTestsFromTypes(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Status)
	assert.NotEmpty(t, result.GeneratedCode)
	assert.Contains(t, result.GeneratedCode, "Test2::V0")
	assert.Contains(t, result.Message, "Generated comprehensive tests")

	validator.AssertExpectations(t)
}

func TestAdvancedGenerator_RefactorCode(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	advGen := NewAdvancedGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	// Mock validation responses
	originalValidation := &validation.ValidationResult{
		Errors:   []validation.ValidationError{},
		Warnings: []validation.ValidationWarning{},
	}
	refactoredValidation := &validation.ValidationResult{
		Errors:   []validation.ValidationError{},
		Warnings: []validation.ValidationWarning{},
	}

	validator.On("ValidateCode", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string"), "").
		Return(originalValidation, nil).Once()
	validator.On("ValidateCode", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string"), "").
		Return(refactoredValidation, nil).Once()

	tests := []struct {
		name             string
		request          RefactoringRequest
		expectedStatus   string
		expectedContains string
	}{
		{
			name: "extract_method",
			request: RefactoringRequest{
				Code:            "sub process { my $x = 5; my $y = 10; return $x + $y; }",
				RefactoringType: "extract_method",
				Target:          "my $y = 10; return $x + $y;",
				SessionID:       "test-session",
			},
			expectedStatus:   "success",
			expectedContains: "extract_method",
		},
		{
			name: "rename",
			request: RefactoringRequest{
				Code:            "sub calculate { my $value = 42; return $value * 2; }",
				RefactoringType: "rename",
				Target:          "calculate",
				NewName:         "compute",
				SessionID:       "test-session",
			},
			expectedStatus:   "success",
			expectedContains: "rename",
		},
		{
			name: "inline",
			request: RefactoringRequest{
				Code:            "sub helper { return 42; } sub main { my $x = helper(); }",
				RefactoringType: "inline",
				Target:          "helper",
				SessionID:       "test-session",
			},
			expectedStatus:   "success",
			expectedContains: "inline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := advGen.RefactorCode(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.NotEmpty(t, result.GeneratedCode)
			assert.Contains(t, result.Message, tt.expectedContains)
		})
	}

	validator.AssertExpectations(t)
}

func TestAdvancedGenerator_RefactorCode_InvalidType(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	advGen := NewAdvancedGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	// Mock validation response
	validationResult := &validation.ValidationResult{
		Errors:   []validation.ValidationError{},
		Warnings: []validation.ValidationWarning{},
	}
	validator.On("ValidateCode", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string"), "").
		Return(validationResult, nil)

	request := RefactoringRequest{
		Code:            "sub test { }",
		RefactoringType: "invalid_type",
		Target:          "test",
		SessionID:       "test-session",
	}

	result, err := advGen.RefactorCode(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported refactoring type")
}

func TestAdvancedGenerator_GenerateDocumentation(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	advGen := NewAdvancedGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	// Mock validation response
	validationResult := &validation.ValidationResult{
		Errors:   []validation.ValidationError{},
		Warnings: []validation.ValidationWarning{},
	}
	validator.On("ValidateCode", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string"), "").
		Return(validationResult, nil)

	tests := []struct {
		name           string
		request        DocumentationRequest
		expectedPrefix string
	}{
		{
			name: "pod_style",
			request: DocumentationRequest{
				Code:      "sub calculate { my ($x, $y) = @_; return $x + $y; }",
				Style:     "pod",
				SessionID: "test-session",
			},
			expectedPrefix: "=",
		},
		{
			name: "markdown_style",
			request: DocumentationRequest{
				Code:      "sub calculate { my ($x, $y) = @_; return $x + $y; }",
				Style:     "markdown",
				SessionID: "test-session",
			},
			expectedPrefix: "#",
		},
		{
			name: "inline_style",
			request: DocumentationRequest{
				Code:      "sub calculate { my ($x, $y) = @_; return $x + $y; }",
				Style:     "inline",
				SessionID: "test-session",
			},
			expectedPrefix: "#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := advGen.GenerateDocumentation(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, "success", result.Status)
			assert.NotEmpty(t, result.GeneratedCode)
			assert.True(t, strings.HasPrefix(result.GeneratedCode, tt.expectedPrefix) ||
				strings.Contains(result.GeneratedCode, tt.expectedPrefix))
			assert.Contains(t, result.Message, tt.request.Style)
		})
	}

	validator.AssertExpectations(t)
}

func TestAdvancedGenerator_CompleteCode(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	advGen := NewAdvancedGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	// Mock validation response
	validationResult := &validation.ValidationResult{
		Errors:   []validation.ValidationError{},
		Warnings: []validation.ValidationWarning{},
	}
	validator.On("ValidateCode", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string"), "").
		Return(validationResult, nil)

	request := CompletionRequest{
		PartialCode: "sub calculate { my ($x, $y) = @_; ret",
		CursorPos:   37, // After "ret"
		Context:     "use v5.40;",
		SessionID:   "test-session",
	}

	result, err := advGen.CompleteCode(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Status)
	assert.NotEmpty(t, result.GeneratedCode)
	assert.Contains(t, result.Message, "completion generated successfully")

	validator.AssertExpectations(t)
}

func TestAdvancedGenerator_CompleteCode_ValidationFailure(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	advGen := NewAdvancedGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	// Mock validation failure
	validator.On("ValidateCode", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string"), "").
		Return(nil, assert.AnError)

	request := CompletionRequest{
		PartialCode: "sub calculate { my ($x, $y) = @_; ret",
		CursorPos:   37,
		Context:     "use v5.40;",
		SessionID:   "test-session",
	}

	result, err := advGen.CompleteCode(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success_with_warnings", result.Status)
	assert.NotEmpty(t, result.GeneratedCode)
	assert.Contains(t, result.Message, "may have validation issues")

	validator.AssertExpectations(t)
}

func TestAdvancedGenerator_BatchGenerate(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	advGen := NewAdvancedGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	// Mock validation responses
	validationResult := &validation.ValidationResult{
		Errors:   []validation.ValidationError{},
		Warnings: []validation.ValidationWarning{},
	}
	validator.On("ValidateCode", mock.AnythingOfType("string")).Return(validationResult, nil)

	request := BatchGenerationRequest{
		Requests: []GenerationRequest{
			{
				Type:          "function",
				Specification: "Create add function",
			},
			{
				Type:          "class",
				Specification: "Create Person class",
			},
			{
				Type:          "test",
				Specification: "Test add function",
			},
		},
		SessionID: "batch-session",
	}

	results, err := advGen.BatchGenerate(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 3)

	for i, result := range results {
		assert.NotNil(t, result)
		assert.Equal(t, "success", result.Status)
		assert.NotEmpty(t, result.GeneratedCode)

		// Verify correct type generation
		switch request.Requests[i].Type {
		case "function":
			assert.Contains(t, result.GeneratedCode, "sub ")
		case "class":
			assert.True(t, strings.Contains(result.GeneratedCode, "class ") ||
				strings.Contains(result.GeneratedCode, "package "))
		case "test":
			assert.True(t, strings.Contains(result.GeneratedCode, "ok(") ||
				strings.Contains(result.GeneratedCode, "is("))
		}
	}

	// Verify shared session was used
	memory := memoryManager.GetSession("batch-session")
	decisions := memory.GetDecisions()
	assert.Greater(t, len(decisions), 3) // Should have decisions from all generations

	validator.AssertExpectations(t)
}

func TestAdvancedGenerator_BatchGenerate_WithErrors(t *testing.T) {
	validator := &MockValidator{}
	autoFixer := &MockAutoFixer{}
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	logger := log.NewLogger(log.LevelInfo, nil, "test")

	advGen := NewAdvancedGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	// Mock validation responses - first succeeds, second fails
	validationResult := &validation.ValidationResult{
		Errors:   []validation.ValidationError{},
		Warnings: []validation.ValidationWarning{},
	}
	validator.On("ValidateCode", mock.AnythingOfType("string")).Return(validationResult, nil).Once()
	validator.On("ValidateCode", mock.AnythingOfType("string")).Return(nil, assert.AnError).Once()

	request := BatchGenerationRequest{
		Requests: []GenerationRequest{
			{
				Type:          "function",
				Specification: "Create add function",
			},
			{
				Type:          "invalid", // This will cause an error
				Specification: "Invalid type",
			},
		},
		SessionID: "batch-error-session",
	}

	results, err := advGen.BatchGenerate(context.Background(), request)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch generation had 1 errors")
	assert.NotNil(t, results)
	assert.Len(t, results, 2)

	// First result should be success
	assert.Equal(t, "success", results[0].Status)
	assert.NotEmpty(t, results[0].GeneratedCode)

	// Second result should be error
	assert.Equal(t, "error", results[1].Status)
	assert.NotEmpty(t, results[1].Message)
}

func TestAdvancedGenerator_ExtractTypeInfo(t *testing.T) {
	advGen := &AdvancedGenerator{}

	code := `sub add {
		my ($x, $y) = @_;
		return $x + $y;
	}

	class Calculator {
		field $value;
		method compute { }
	}`

	typeSigs := map[string]string{
		"add":     "(Int, Int) -> Int",
		"compute": "() -> Int",
	}

	typeInfo := advGen.extractTypeInfo(code, typeSigs)

	assert.NotNil(t, typeInfo)
	assert.Equal(t, typeSigs, typeInfo["signatures"])

	// Type assertions with proper handling
	functionsInterface, ok := typeInfo["functions"]
	assert.True(t, ok, "functions key should exist")
	functions, ok := functionsInterface.([]string)
	assert.True(t, ok, "functions should be []string")
	assert.Len(t, functions, 2)
	assert.Contains(t, functions[0], "sub add")
	assert.Contains(t, functions[1], "method compute")

	classesInterface, ok := typeInfo["classes"]
	assert.True(t, ok, "classes key should exist")
	classes, ok := classesInterface.([]string)
	assert.True(t, ok, "classes should be []string")
	assert.Len(t, classes, 1)
	assert.Contains(t, classes[0], "class Calculator")
}

func TestAdvancedGenerator_FormatDocumentation(t *testing.T) {
	advGen := &AdvancedGenerator{}

	tests := []struct {
		name     string
		docs     string
		style    string
		expected string
	}{
		{
			name:     "pod_formatting",
			docs:     "This is documentation",
			style:    "pod",
			expected: "=pod\n\nThis is documentation\n\n=cut",
		},
		{
			name:     "pod_already_formatted",
			docs:     "=head1 NAME\n\nModule",
			style:    "pod",
			expected: "=head1 NAME\n\nModule",
		},
		{
			name:     "markdown_formatting",
			docs:     "This is documentation",
			style:    "markdown",
			expected: "# Documentation\n\nThis is documentation",
		},
		{
			name:     "markdown_already_formatted",
			docs:     "# Title\n\nContent",
			style:    "markdown",
			expected: "# Title\n\nContent",
		},
		{
			name:     "inline_formatting",
			docs:     "Line 1\nLine 2\n\nLine 3",
			style:    "inline",
			expected: "# Line 1\n# Line 2\n\n# Line 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := advGen.formatDocumentation(tt.docs, tt.style)
			assert.Equal(t, tt.expected, formatted)
		})
	}
}

func TestAdvancedGenerator_ExtractCursorContext(t *testing.T) {
	advGen := &AdvancedGenerator{}

	tests := []struct {
		name           string
		code           string
		cursorPos      int
		expectedPrefix string
		expectedSuffix string
	}{
		{
			name:           "middle_position",
			code:           "sub test { return 42; }",
			cursorPos:      10,
			expectedPrefix: "sub test {",
			expectedSuffix: " return 42; }",
		},
		{
			name:           "start_position",
			code:           "sub test { }",
			cursorPos:      0,
			expectedPrefix: "",
			expectedSuffix: "sub test { }",
		},
		{
			name:           "end_position",
			code:           "sub test { }",
			cursorPos:      12,
			expectedPrefix: "sub test { }",
			expectedSuffix: "",
		},
		{
			name:           "beyond_end",
			code:           "sub test { }",
			cursorPos:      100,
			expectedPrefix: "sub test { }",
			expectedSuffix: "",
		},
		{
			name:           "negative_position",
			code:           "sub test { }",
			cursorPos:      -5,
			expectedPrefix: "",
			expectedSuffix: "sub test { }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, suffix := advGen.extractCursorContext(tt.code, tt.cursorPos)
			assert.Equal(t, tt.expectedPrefix, prefix)
			assert.Equal(t, tt.expectedSuffix, suffix)
		})
	}
}
