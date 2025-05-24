// ABOUTME: Integration layer for advanced generation features
// ABOUTME: Combines test generation, refactoring, documentation, completion, and batch operations

package tools

import (
	"context"
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/mcp/generation"
	"tamarou.com/pvm/internal/mcp/validation"
)

// AdvancedGenerator provides advanced code generation capabilities
type AdvancedGenerator struct {
	testGenerator *TestGenerator
	refactorer    *RefactoringEngine
	docGenerator  *DocumentationGenerator
	completer     *CompletionEngine
	codeGenerator *CodeGenerator
	typeParser    TypeParser
	logger        *log.Logger
}

// NewAdvancedGenerator creates a new advanced generator instance
func NewAdvancedGenerator(validator Validator, autoFixer AutoFixer, samplingClient *generation.SamplingClient, memoryManager *generation.MemoryManager, logger *log.Logger) *AdvancedGenerator {
	// Create a simple type parser implementation
	typeParser := &simpleTypeParser{}

	return &AdvancedGenerator{
		testGenerator: NewTestGenerator(samplingClient, typeParser, logger),
		refactorer:    NewRefactoringEngine(samplingClient, typeParser, validator, logger),
		docGenerator:  NewDocumentationGenerator(samplingClient, logger),
		completer:     NewCompletionEngine(samplingClient, typeParser, logger),
		codeGenerator: NewCodeGenerator(validator, autoFixer, samplingClient, memoryManager, logger),
		typeParser:    typeParser,
		logger:        logger,
	}
}

// TestGenerationRequest represents a request to generate tests
type TestGenerationRequest struct {
	Code        string            `json:"code"`
	TypeSigs    map[string]string `json:"type_signatures"`
	Framework   string            `json:"framework"`
	SessionID   string            `json:"session_id"`
	ProjectPath string            `json:"project_path"`
}

// GenerateTestsFromTypes generates tests from type signatures
func (ag *AdvancedGenerator) GenerateTestsFromTypes(ctx context.Context, request TestGenerationRequest) (*GenerationResult, error) {
	ag.logger.Infof("Generating tests from types for session: %s", request.SessionID)

	// If no type signatures provided, try to extract from code
	if len(request.TypeSigs) == 0 {
		ag.logger.Debugf("No type signatures provided, attempting to extract from code")
		// For now, generate tests without explicit type signatures
		// In a full implementation, we would parse the code to extract types
	}

	// Generate tests for each function with type signature
	var allTests []string
	var iterations int

	for functionName, typeSig := range request.TypeSigs {
		testReq := TestGenRequest{
			TypeSignature: typeSig,
			FunctionName:  functionName,
			Context:       request.Code,
			Framework:     request.Framework,
		}

		result, err := ag.testGenerator.GenerateTestsFromType(testReq)
		if err != nil {
			ag.logger.Warningf("Failed to generate tests for %s: %v", functionName, err)
			continue
		}

		allTests = append(allTests, result.TestCode)
		iterations++
	}

	// If no type signatures, generate general tests for the code
	if len(request.TypeSigs) == 0 {
		testReq := TestGenRequest{
			TypeSignature: "",
			FunctionName:  "main",
			Context:       request.Code,
			Framework:     request.Framework,
		}

		result, err := ag.testGenerator.GenerateTestsFromType(testReq)
		if err != nil {
			return nil, fmt.Errorf("test generation failed: %w", err)
		}

		allTests = append(allTests, result.TestCode)
		iterations = 1
	}

	// Combine all tests
	combinedTests := strings.Join(allTests, "\n\n")

	return &GenerationResult{
		Status:        "success",
		GeneratedCode: combinedTests,
		Iterations:    iterations,
		Message:       fmt.Sprintf("Generated tests for %d functions using %s", len(allTests), request.Framework),
	}, nil
}

// RefactoringRequestInternal represents a refactoring request for the MCP server
type RefactoringRequestInternal struct {
	Code            string `json:"code"`
	RefactoringType string `json:"refactoring_type"`
	Target          string `json:"target"`
	NewName         string `json:"new_name"`
	SessionID       string `json:"session_id"`
}

// RefactorCode performs code refactoring
func (ag *AdvancedGenerator) RefactorCode(ctx context.Context, request RefactoringRequestInternal) (*GenerationResult, error) {
	ag.logger.Infof("Refactoring code: type=%s, target=%s", request.RefactoringType, request.Target)

	refReq := RefactoringRequest{
		Code:            request.Code,
		RefactoringType: request.RefactoringType,
		Target:          request.Target,
		NewName:         request.NewName,
		PreserveTypes:   true, // Always preserve types
	}

	result, err := ag.refactorer.Refactor(refReq)
	if err != nil {
		return nil, fmt.Errorf("refactoring failed: %w", err)
	}

	// Create validation result if types were not preserved
	var validationResult *validation.ValidationResult
	if !result.TypesSafe {
		validationResult = &validation.ValidationResult{
			Valid: false,
			Warnings: []validation.ValidationWarning{
				{
					Message: "Type safety may have been compromised during refactoring",
					Line:    0,
					Column:  0,
					Code:    "type_safety",
				},
			},
		}
	}

	return &GenerationResult{
		Status:           "success",
		GeneratedCode:    result.RefactoredCode,
		ValidationResult: validationResult,
		Iterations:       1,
		Message:          fmt.Sprintf("Refactoring completed: %s", strings.Join(result.Changes, ", ")),
	}, nil
}

// DocumentationRequestInternal represents a documentation generation request for the MCP server
type DocumentationRequestInternal struct {
	Code      string `json:"code"`
	Style     string `json:"style"`
	SessionID string `json:"session_id"`
}

// GenerateDocumentation generates documentation
func (ag *AdvancedGenerator) GenerateDocumentation(ctx context.Context, request DocumentationRequestInternal) (*GenerationResult, error) {
	ag.logger.Infof("Generating documentation: style=%s", request.Style)

	// Map style to doc type
	docType := request.Style
	if request.Style == "markdown" {
		docType = "inline" // Convert markdown to inline for now
	}

	docReq := DocumentationRequest{
		Code:         request.Code,
		DocType:      docType,
		IncludeTypes: true,
		Verbose:      false,
	}

	result, err := ag.docGenerator.GenerateDocumentation(docReq)
	if err != nil {
		return nil, fmt.Errorf("documentation generation failed: %w", err)
	}

	return &GenerationResult{
		Status:        "success",
		GeneratedCode: result.Documentation,
		Iterations:    1,
		Message:       fmt.Sprintf("Generated %s documentation with %d sections", request.Style, len(result.Sections)),
	}, nil
}

// CompletionRequestInternal represents a code completion request for the MCP server
type CompletionRequestInternal struct {
	PartialCode string `json:"partial_code"`
	CursorPos   int    `json:"cursor_position"`
	Context     string `json:"context"`
	SessionID   string `json:"session_id"`
}

// CompleteCode provides code completion
func (ag *AdvancedGenerator) CompleteCode(ctx context.Context, request CompletionRequestInternal) (*GenerationResult, error) {
	ag.logger.Infof("Completing code at position: %d", request.CursorPos)

	compReq := CompletionRequest{
		PartialCode:    request.PartialCode,
		CursorPosition: request.CursorPos,
		Context:        request.Context,
		MaxSuggestions: 5,
	}

	result, err := ag.completer.Complete(compReq)
	if err != nil {
		return nil, fmt.Errorf("code completion failed: %w", err)
	}

	// Format completions as code
	var completions []string
	for _, suggestion := range result.Suggestions {
		completions = append(completions, fmt.Sprintf("# %s: %s\n%s",
			suggestion.Description,
			suggestion.TypeInfo,
			suggestion.Text))
	}

	return &GenerationResult{
		Status:        "success",
		GeneratedCode: strings.Join(completions, "\n\n"),
		Iterations:    1,
		Message:       fmt.Sprintf("Generated %d completion suggestions", len(result.Suggestions)),
	}, nil
}

// BatchGenerationRequestInternal represents a batch generation request for the MCP server
type BatchGenerationRequestInternal struct {
	Requests  []GenerationRequest `json:"requests"`
	SessionID string              `json:"session_id"`
}

// BatchGenerate performs batch code generation
func (ag *AdvancedGenerator) BatchGenerate(ctx context.Context, request BatchGenerationRequestInternal) ([]*GenerationResult, error) {
	ag.logger.Infof("Processing batch generation with %d requests", len(request.Requests))

	batchReq := BatchGenerationRequest{
		Requests:  request.Requests,
		Parallel:  false, // Sequential for now
		SessionID: request.SessionID,
	}

	result, err := ag.codeGenerator.GenerateBatch(batchReq)
	if err != nil {
		// Note: err here indicates partial failure
		ag.logger.Warningf("Batch generation had errors: %v", err)
	}

	// Extract error information
	var batchErr error
	if result.Failed > 0 {
		batchErr = fmt.Errorf("batch generation: %d succeeded, %d failed. Errors: %s",
			result.Succeeded, result.Failed, strings.Join(result.Errors, "; "))
	}

	return result.Results, batchErr
}

// simpleTypeParser provides basic type parsing functionality
type simpleTypeParser struct{}

func (p *simpleTypeParser) ParseTypeSignature(signature string) (*SimpleType, error) {
	// Simple parsing - just extract the base type name
	parts := strings.Split(signature, "->")
	if len(parts) > 0 {
		typeName := strings.TrimSpace(parts[0])
		// Remove any brackets or parameters
		if idx := strings.Index(typeName, "["); idx > 0 {
			typeName = typeName[:idx]
		}
		return &SimpleType{Name: typeName}, nil
	}
	return &SimpleType{Name: "Any"}, nil
}

func (p *simpleTypeParser) ExtractTypeFromCode(code string) ([]*SimpleType, error) {
	// Simple extraction - look for type annotations
	var types []*SimpleType

	// Look for "my Type $var" patterns
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "my ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				// Check if second part looks like a type (starts with uppercase)
				if len(parts[1]) > 0 && strings.ToUpper(parts[1][:1]) == parts[1][:1] {
					types = append(types, &SimpleType{Name: parts[1]})
				}
			}
		}
	}

	return types, nil
}
