// ABOUTME: Tests for advanced code generation features
// ABOUTME: Validates test generation, refactoring, documentation, and completion functionality

package tools

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/mcp/generation"
	"tamarou.com/pvm/internal/mcp/validation"
)

// Mock implementations for testing

type mockTypeParser struct {
	parseTypeErr   error
	extractTypeErr error
	parsedType     *SimpleType
	extractedTypes []*SimpleType
}

func (m *mockTypeParser) ParseTypeSignature(signature string) (*SimpleType, error) {
	if m.parseTypeErr != nil {
		return nil, m.parseTypeErr
	}
	if m.parsedType != nil {
		return m.parsedType, nil
	}
	// Return a simple type for testing
	return &SimpleType{Name: "Int"}, nil
}

func (m *mockTypeParser) ExtractTypeFromCode(code string) ([]*SimpleType, error) {
	if m.extractTypeErr != nil {
		return nil, m.extractTypeErr
	}
	if m.extractedTypes != nil {
		return m.extractedTypes, nil
	}
	// Return some default types
	return []*SimpleType{
		{Name: "Int"},
		{Name: "Str"},
	}, nil
}

type mockValidator struct {
	result *validation.ValidationResult
	err    error
}

func (m *mockValidator) ValidateCode(ctx context.Context, code string, projectPath string) (*validation.ValidationResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return &validation.ValidationResult{
		Valid:  true,
		Errors: []validation.ValidationError{},
	}, nil
}

func TestTestGenerator_GenerateTestsFromType(t *testing.T) {
	logger := log.NewLogger(log.LevelError, os.Stderr, "test")
	samplingClient := generation.NewSamplingClient(true)
	typeParser := &mockTypeParser{}

	testGen := NewTestGenerator(samplingClient, typeParser, logger)

	tests := []struct {
		name     string
		request  TestGenRequest
		wantErr  bool
		errMsg   string
		validate func(t *testing.T, result *TestGenerationResult)
	}{
		{
			name: "successful test generation with default framework",
			request: TestGenRequest{
				TypeSignature: "Int -> Str",
				FunctionName:  "int_to_string",
				Context:       "Converts integer to string",
			},
			wantErr: false,
			validate: func(t *testing.T, result *TestGenerationResult) {
				assert.NotEmpty(t, result.TestCode)
				assert.Contains(t, result.TestCode, "use Test2::V0")
				assert.Contains(t, result.TestCode, "int_to_string")
				assert.Len(t, result.TestCases, 5)
				assert.Contains(t, result.TestCases[0], "valid input")
				assert.Contains(t, result.TestCases[1], "invalid input")
				assert.Contains(t, result.TestCases[2], "edge cases")
				assert.Contains(t, result.TestCases[3], "type constraints")
				assert.Contains(t, result.TestCases[4], "return type correctness")
				assert.Greater(t, result.Coverage, 0.0)
				assert.LessOrEqual(t, result.Coverage, 1.0)
			},
		},
		{
			name: "test generation with custom framework",
			request: TestGenRequest{
				TypeSignature: "ArrayRef[Int] -> Int",
				FunctionName:  "sum_array",
				Context:       "Sums all integers in array",
				Framework:     "Test::More",
			},
			wantErr: false,
			validate: func(t *testing.T, result *TestGenerationResult) {
				assert.NotEmpty(t, result.TestCode)
				assert.Contains(t, result.TestCode, "use Test::More")
				assert.Contains(t, result.TestCode, "sum_array")
			},
		},
		{
			name: "type parsing error",
			request: TestGenRequest{
				TypeSignature: "Invalid::Type",
				FunctionName:  "bad_function",
			},
			wantErr: true,
			errMsg:  "failed to parse type signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr && strings.Contains(tt.errMsg, "parse type signature") {
				typeParser.parseTypeErr = assert.AnError
			} else {
				typeParser.parseTypeErr = nil
			}

			result, err := testGen.GenerateTestsFromType(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				tt.validate(t, result)
			}
		})
	}
}

func TestRefactoringEngine_Refactor(t *testing.T) {
	logger := log.NewLogger(log.LevelError, os.Stderr, "test")
	samplingClient := generation.NewSamplingClient(true)
	typeParser := &mockTypeParser{}
	validator := &mockValidator{}

	refactorer := NewRefactoringEngine(samplingClient, typeParser, validator, logger)

	tests := []struct {
		name     string
		request  RefactoringRequest
		wantErr  bool
		validate func(t *testing.T, result *RefactoringResult)
	}{
		{
			name: "extract method refactoring",
			request: RefactoringRequest{
				Code: `sub process {
    my $data = shift;
    # validation logic
    die "Invalid data" unless $data;
    die "Data too large" if length($data) > 100;
    # processing logic
    return uc($data);
}`,
				RefactoringType: "extract_method",
				Target:          "# validation logic",
				PreserveTypes:   true,
			},
			wantErr: false,
			validate: func(t *testing.T, result *RefactoringResult) {
				assert.NotEmpty(t, result.RefactoredCode)
				assert.Contains(t, result.RefactoredCode, "sub ")
				assert.True(t, result.TypesSafe)
				assert.Contains(t, result.Changes, "Extracted code into new method")
			},
		},
		{
			name: "rename refactoring",
			request: RefactoringRequest{
				Code: `my $foo = 42;
print $foo;`,
				RefactoringType: "rename",
				Target:          "foo",
				NewName:         "meaningful_name",
				PreserveTypes:   true,
			},
			wantErr: false,
			validate: func(t *testing.T, result *RefactoringResult) {
				assert.NotEmpty(t, result.RefactoredCode)
				assert.Contains(t, result.RefactoredCode, "meaningful_name")
				assert.Contains(t, result.Changes, "Renamed identifiers throughout code")
			},
		},
		{
			name: "inline refactoring",
			request: RefactoringRequest{
				Code: `sub get_pi { return 3.14159; }
my $circle_area = get_pi() * $r * $r;`,
				RefactoringType: "inline",
				Target:          "get_pi",
				PreserveTypes:   false,
			},
			wantErr: false,
			validate: func(t *testing.T, result *RefactoringResult) {
				assert.NotEmpty(t, result.RefactoredCode)
				assert.Contains(t, result.RefactoredCode, "3.14159")
				assert.Contains(t, result.Changes, "Inlined method/variable at call sites")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := refactorer.Refactor(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				tt.validate(t, result)
			}
		})
	}
}

func TestDocumentationGenerator_GenerateDocumentation(t *testing.T) {
	logger := log.NewLogger(log.LevelError, os.Stderr, "test")
	samplingClient := generation.NewSamplingClient(true)

	docGen := NewDocumentationGenerator(samplingClient, logger)

	tests := []struct {
		name     string
		request  DocumentationRequest
		wantErr  bool
		validate func(t *testing.T, result *DocumentationResult)
	}{
		{
			name: "generate POD documentation",
			request: DocumentationRequest{
				Code: `sub calculate_discount {
    my ($price, $percentage) = @_;
    return $price * (1 - $percentage / 100);
}`,
				DocType:      "pod",
				IncludeTypes: true,
				Verbose:      false,
			},
			wantErr: false,
			validate: func(t *testing.T, result *DocumentationResult) {
				assert.NotEmpty(t, result.Documentation)
				assert.Contains(t, result.Documentation, "=head")
				assert.NotEmpty(t, result.Sections)
			},
		},
		{
			name: "generate inline documentation",
			request: DocumentationRequest{
				Code: `my Int $count = 0;
$count++;`,
				DocType:      "inline",
				IncludeTypes: true,
				Verbose:      true,
			},
			wantErr: false,
			validate: func(t *testing.T, result *DocumentationResult) {
				assert.NotEmpty(t, result.Documentation)
				assert.NotEmpty(t, result.TypeInfo)
				assert.Contains(t, result.TypeInfo[0], "$count: Int")
			},
		},
		{
			name: "generate both POD and inline",
			request: DocumentationRequest{
				Code: `package MyModule;
use v5.40;`,
				DocType:      "both",
				IncludeTypes: false,
				Verbose:      false,
			},
			wantErr: false,
			validate: func(t *testing.T, result *DocumentationResult) {
				assert.NotEmpty(t, result.Documentation)
				assert.NotEmpty(t, result.Sections)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := docGen.GenerateDocumentation(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				tt.validate(t, result)
			}
		})
	}
}

func TestCompletionEngine_Complete(t *testing.T) {
	logger := log.NewLogger(log.LevelError, os.Stderr, "test")
	samplingClient := generation.NewSamplingClient(true)
	typeParser := &mockTypeParser{}

	completer := NewCompletionEngine(samplingClient, typeParser, logger)

	tests := []struct {
		name     string
		request  CompletionRequest
		wantErr  bool
		validate func(t *testing.T, result *CompletionResult)
	}{
		{
			name: "method completion",
			request: CompletionRequest{
				PartialCode: `my $str = "hello";
$str->`,
				CursorPosition: 24,
				Context:        "String object methods",
				MaxSuggestions: 5,
			},
			wantErr: false,
			validate: func(t *testing.T, result *CompletionResult) {
				assert.NotEmpty(t, result.Suggestions)
				assert.LessOrEqual(t, len(result.Suggestions), 5)
				if len(result.Suggestions) > 0 {
					assert.NotEmpty(t, result.Suggestions[0].Text)
					assert.NotEmpty(t, result.Suggestions[0].Description)
					assert.Greater(t, result.Suggestions[0].Score, 0.0)
				}
			},
		},
		{
			name: "variable completion",
			request: CompletionRequest{
				PartialCode: `my $count = 0;
my $c`,
				CursorPosition: 20,
				Context:        "Variable names",
				MaxSuggestions: 3,
			},
			wantErr: false,
			validate: func(t *testing.T, result *CompletionResult) {
				assert.NotEmpty(t, result.Suggestions)
				assert.LessOrEqual(t, len(result.Suggestions), 3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := completer.Complete(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				tt.validate(t, result)
			}
		})
	}
}

func TestCodeGenerator_GenerateBatch(t *testing.T) {
	logger := log.NewLogger(log.LevelError, os.Stderr, "test")
	samplingClient := generation.NewSamplingClient(true)
	memoryManager := generation.NewMemoryManager(50)
	validator := &mockValidator{}
	autoFixer := &mockAutoFixer{}

	codeGen := NewCodeGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	tests := []struct {
		name     string
		request  BatchGenerationRequest
		wantErr  bool
		validate func(t *testing.T, result *BatchGenerationResult)
	}{
		{
			name: "successful batch generation",
			request: BatchGenerationRequest{
				Requests: []GenerationRequest{
					{
						Type:          "function",
						Specification: "Add two numbers",
						Context:       "",
						SessionID:     "batch-1",
					},
					{
						Type:          "class",
						Specification: "User class with name and email",
						Context:       "",
						SessionID:     "batch-1",
					},
					{
						Type:          "test",
						Specification: "Test the add function",
						Context:       "sub add { my ($a, $b) = @_; return $a + $b; }",
						SessionID:     "batch-1",
					},
				},
				Parallel:  false,
				SessionID: "batch-1",
			},
			wantErr: false,
			validate: func(t *testing.T, result *BatchGenerationResult) {
				assert.Equal(t, 3, len(result.Results))
				assert.Equal(t, 3, result.Succeeded)
				assert.Equal(t, 0, result.Failed)
				assert.Empty(t, result.Errors)

				// Check individual results
				for _, res := range result.Results {
					assert.NotNil(t, res)
					assert.Equal(t, "success", res.Status)
					assert.NotEmpty(t, res.GeneratedCode)
					assert.NotEmpty(t, res.Message)
					assert.Greater(t, res.Iterations, 0)
				}
			},
		},
		{
			name: "batch with invalid request",
			request: BatchGenerationRequest{
				Requests: []GenerationRequest{
					{
						Type:          "function",
						Specification: "Valid function",
						SessionID:     "batch-2",
					},
					{
						Type:          "invalid_type",
						Specification: "This should fail",
						SessionID:     "batch-2",
					},
				},
				SessionID: "batch-2",
			},
			wantErr: false, // Batch itself doesn't error, individual requests do
			validate: func(t *testing.T, result *BatchGenerationResult) {
				assert.Equal(t, 2, len(result.Results))
				assert.Equal(t, 1, result.Succeeded)
				assert.Equal(t, 1, result.Failed)
				assert.Len(t, result.Errors, 1)
				assert.Contains(t, result.Errors[0], "unsupported generation type")

				// First should succeed
				assert.Equal(t, "success", result.Results[0].Status)
				// Second should fail
				assert.Equal(t, "error", result.Results[1].Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := codeGen.GenerateBatch(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				tt.validate(t, result)
			}
		})
	}
}

// Test helper functions

func TestGenerateTestCasesFromType(t *testing.T) {
	logger := log.NewLogger(log.LevelError, os.Stderr, "test")
	samplingClient := generation.NewSamplingClient(true)
	typeParser := &mockTypeParser{}

	testGen := NewTestGenerator(samplingClient, typeParser, logger)

	tests := []struct {
		name         string
		typeInfo     *SimpleType
		functionName string
		wantCases    int
		checkContent func(t *testing.T, cases []string)
	}{
		{
			name:         "with type info",
			typeInfo:     &SimpleType{Name: "Int"},
			functionName: "calculate",
			wantCases:    5,
			checkContent: func(t *testing.T, cases []string) {
				assert.Contains(t, cases[3], "type constraints")
				assert.Contains(t, cases[4], "return type correctness")
			},
		},
		{
			name:         "without type info",
			typeInfo:     nil,
			functionName: "process",
			wantCases:    3,
			checkContent: func(t *testing.T, cases []string) {
				assert.Contains(t, cases[0], "valid input")
				assert.Contains(t, cases[1], "invalid input")
				assert.Contains(t, cases[2], "edge cases")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cases := testGen.generateTestCasesFromType(tt.typeInfo, tt.functionName)
			assert.Len(t, cases, tt.wantCases)
			tt.checkContent(t, cases)
		})
	}
}

func TestParseDocumentationSections(t *testing.T) {
	logger := log.NewLogger(log.LevelError, os.Stderr, "test")
	samplingClient := generation.NewSamplingClient(true)
	docGen := NewDocumentationGenerator(samplingClient, logger)

	tests := []struct {
		name     string
		doc      string
		expected map[string]string
	}{
		{
			name: "POD sections",
			doc: `=head1 NAME

MyModule - A test module

=head1 DESCRIPTION

This module does things.

=head1 METHODS

=head2 new

Constructor`,
			expected: map[string]string{
				"NAME":        "MyModule - A test module",
				"DESCRIPTION": "This module does things.",
				"METHODS":     "=head2 new\n\nConstructor",
			},
		},
		{
			name: "no POD sections",
			doc:  "Just plain text documentation",
			expected: map[string]string{
				"main": "Just plain text documentation",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := docGen.parseDocumentationSections(tt.doc)
			assert.Equal(t, len(tt.expected), len(sections))
			for key, expectedValue := range tt.expected {
				assert.Contains(t, sections, key)
				assert.Equal(t, expectedValue, sections[key])
			}
		})
	}
}

func TestExtractCompletionContext(t *testing.T) {
	logger := log.NewLogger(log.LevelError, os.Stderr, "test")
	samplingClient := generation.NewSamplingClient(true)
	typeParser := &mockTypeParser{}
	completer := NewCompletionEngine(samplingClient, typeParser, logger)

	tests := []struct {
		name          string
		code          string
		position      int
		expectedLine  string
		expectedToken string
	}{
		{
			name: "middle of line",
			code: `my $foo = 42;
$foo->`,
			position:      20,
			expectedLine:  "$foo->",
			expectedToken: "$foo->",
		},
		{
			name: "start of line",
			code: `use v5.40;
my`,
			position:      11,
			expectedLine:  "my",
			expectedToken: "my",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := completer.extractCompletionContext(tt.code, tt.position)
			assert.Equal(t, tt.expectedLine, context["current_line"])
			if tt.expectedToken != "" {
				assert.Equal(t, tt.expectedToken, context["preceding_token"])
			}
		})
	}
}

// Mock auto-fixer for testing
type mockAutoFixer struct{}

func (m *mockAutoFixer) AutoFix(ctx context.Context, code string, errors []validation.ValidationError, projectPath string) ([]validation.FixError, error) {
	return []validation.FixError{}, nil
}
