// ABOUTME: Advanced code generation features including test generation and refactoring
// ABOUTME: Provides sophisticated generation capabilities using type information

package tools

import (
	"context"
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/mcp/generation"
	"tamarou.com/pvm/internal/mcp/validation"
)

// AdvancedGenerator extends CodeGenerator with sophisticated capabilities
type AdvancedGenerator struct {
	*CodeGenerator
}

// NewAdvancedGenerator creates a new advanced generator instance
func NewAdvancedGenerator(validator Validator, autoFixer AutoFixer, samplingClient *generation.SamplingClient, memoryManager *generation.MemoryManager, logger *log.Logger) *AdvancedGenerator {
	return &AdvancedGenerator{
		CodeGenerator: NewCodeGenerator(validator, autoFixer, samplingClient, memoryManager, logger),
	}
}

// TestGenerationRequest represents a request to generate tests from code
type TestGenerationRequest struct {
	Code        string            `json:"code"`         // code to generate tests for
	TypeSigs    map[string]string `json:"type_sigs"`    // type signatures
	Framework   string            `json:"framework"`    // test framework to use
	SessionID   string            `json:"session_id"`   // memory session ID
	ProjectPath string            `json:"project_path"` // optional project path
}

// RefactoringRequest represents a code refactoring request
type RefactoringRequest struct {
	Code            string `json:"code"`             // code to refactor
	RefactoringType string `json:"refactoring_type"` // extract_method, rename, inline, etc.
	Target          string `json:"target"`           // target element to refactor
	NewName         string `json:"new_name"`         // new name (for rename operations)
	SessionID       string `json:"session_id"`       // memory session ID
}

// DocumentationRequest represents a documentation generation request
type DocumentationRequest struct {
	Code      string `json:"code"`       // code to document
	Style     string `json:"style"`      // pod, markdown, inline
	SessionID string `json:"session_id"` // memory session ID
}

// CompletionRequest represents a code completion request
type CompletionRequest struct {
	PartialCode string `json:"partial_code"` // incomplete code
	CursorPos   int    `json:"cursor_pos"`   // cursor position
	Context     string `json:"context"`      // surrounding code context
	SessionID   string `json:"session_id"`   // memory session ID
}

// BatchGenerationRequest represents multiple generation requests
type BatchGenerationRequest struct {
	Requests  []GenerationRequest `json:"requests"`   // multiple requests
	SessionID string              `json:"session_id"` // shared session ID
}

// GenerateTestsFromTypes generates comprehensive test suites from type signatures
func (ag *AdvancedGenerator) GenerateTestsFromTypes(ctx context.Context, request TestGenerationRequest) (*GenerationResult, error) {
	ag.logger.Infof("Generating tests from types for session: %s", request.SessionID)

	// Get or create memory session
	memory := ag.memoryManager.GetSession(request.SessionID)
	memory.AddDecision("test_generation", "from_types", request.Framework,
		"Generating tests from type signatures")

	// Parse type signatures and code structure
	typeInfo := ag.extractTypeInfo(request.Code, request.TypeSigs)

	// Build test generation prompt
	prompt := ag.buildTestFromTypesPrompt(request, typeInfo, memory)

	// Generate tests using collaborative sampling
	tests, iterations, err := ag.generateWithRefinement(prompt, memory, "test")
	if err != nil {
		return nil, fmt.Errorf("test generation failed: %w", err)
	}

	// Validate generated tests
	validationResult, err := ag.validateAndFix(tests, memory)
	if err != nil {
		return nil, fmt.Errorf("test validation failed: %w", err)
	}

	return &GenerationResult{
		Status:           "success",
		GeneratedCode:    tests,
		ValidationResult: validationResult,
		Iterations:       iterations,
		Message:          fmt.Sprintf("Generated comprehensive tests using %s", request.Framework),
	}, nil
}

// RefactorCode performs type-preserving code refactoring
func (ag *AdvancedGenerator) RefactorCode(ctx context.Context, request RefactoringRequest) (*GenerationResult, error) {
	ag.logger.Infof("Refactoring code: type=%s, session=%s", request.RefactoringType, request.SessionID)

	// Get or create memory session
	memory := ag.memoryManager.GetSession(request.SessionID)
	memory.AddDecision("refactoring", request.RefactoringType, request.Target,
		fmt.Sprintf("Refactoring %s", request.Target))

	// Validate original code to get type information
	originalValidation, err := ag.validator.ValidateCode(ctx, request.Code, "")
	if err != nil {
		return nil, fmt.Errorf("failed to validate original code: %w", err)
	}

	// Perform refactoring based on type
	var refactoredCode string
	var iterations int

	switch request.RefactoringType {
	case "extract_method":
		refactoredCode, iterations, err = ag.extractMethod(request, memory, originalValidation)
	case "rename":
		refactoredCode, iterations, err = ag.renameElement(request, memory, originalValidation)
	case "inline":
		refactoredCode, iterations, err = ag.inlineCode(request, memory, originalValidation)
	default:
		return nil, fmt.Errorf("unsupported refactoring type: %s", request.RefactoringType)
	}

	if err != nil {
		return nil, fmt.Errorf("refactoring failed: %w", err)
	}

	// Validate refactored code
	refactoredValidation, err := ag.validator.ValidateCode(ctx, refactoredCode, "")
	if err != nil {
		return nil, fmt.Errorf("failed to validate refactored code: %w", err)
	}

	// Ensure types are preserved
	if !ag.typesPreserved(originalValidation, refactoredValidation) {
		return nil, fmt.Errorf("refactoring broke type preservation")
	}

	return &GenerationResult{
		Status:           "success",
		GeneratedCode:    refactoredCode,
		ValidationResult: refactoredValidation,
		Iterations:       iterations,
		Message:          fmt.Sprintf("Successfully refactored code with %s", request.RefactoringType),
	}, nil
}

// GenerateDocumentation generates documentation from typed code
func (ag *AdvancedGenerator) GenerateDocumentation(ctx context.Context, request DocumentationRequest) (*GenerationResult, error) {
	ag.logger.Infof("Generating documentation: style=%s, session=%s", request.Style, request.SessionID)

	// Get or create memory session
	memory := ag.memoryManager.GetSession(request.SessionID)
	memory.AddDecision("documentation", request.Style, "generate",
		fmt.Sprintf("Generating %s documentation", request.Style))

	// Validate code to extract type information
	validationResult, err := ag.validator.ValidateCode(ctx, request.Code, "")
	if err != nil {
		return nil, fmt.Errorf("failed to validate code: %w", err)
	}

	// Build documentation prompt with type information
	prompt := ag.buildDocumentationPrompt(request, validationResult, memory)

	// Generate documentation
	docs, iterations, err := ag.generateWithRefinement(prompt, memory, "documentation")
	if err != nil {
		return nil, fmt.Errorf("documentation generation failed: %w", err)
	}

	// Format documentation based on style
	formattedDocs := ag.formatDocumentation(docs, request.Style)

	return &GenerationResult{
		Status:        "success",
		GeneratedCode: formattedDocs,
		Iterations:    iterations,
		Message:       fmt.Sprintf("Generated %s documentation", request.Style),
	}, nil
}

// CompleteCode provides intelligent code completion suggestions
func (ag *AdvancedGenerator) CompleteCode(ctx context.Context, request CompletionRequest) (*GenerationResult, error) {
	ag.logger.Infof("Completing code at position %d, session=%s", request.CursorPos, request.SessionID)

	// Get or create memory session
	memory := ag.memoryManager.GetSession(request.SessionID)

	// Extract context around cursor
	prefix, suffix := ag.extractCursorContext(request.PartialCode, request.CursorPos)

	// Build completion prompt
	prompt := ag.buildCompletionPrompt(prefix, suffix, request.Context, memory)

	// Generate completion
	completionResponse, err := ag.samplingClient.Sample(ctx, prompt, "")
	if err != nil {
		return nil, fmt.Errorf("completion generation failed: %w", err)
	}

	completion := strings.TrimSpace(completionResponse.Content)

	// Build completed code
	completedCode := prefix + completion + suffix

	// Validate completed code
	validationResult, err := ag.validator.ValidateCode(ctx, completedCode, "")
	if err != nil {
		// If validation fails, return completion anyway with warning
		return &GenerationResult{
			Status:        "success_with_warnings",
			GeneratedCode: completion,
			Iterations:    1,
			Message:       "Completion generated but may have validation issues",
		}, nil
	}

	return &GenerationResult{
		Status:           "success",
		GeneratedCode:    completion,
		ValidationResult: validationResult,
		Iterations:       1,
		Message:          "Code completion generated successfully",
	}, nil
}

// BatchGenerate handles multiple generation requests efficiently
func (ag *AdvancedGenerator) BatchGenerate(ctx context.Context, request BatchGenerationRequest) ([]*GenerationResult, error) {
	ag.logger.Infof("Batch generating %d requests, session=%s", len(request.Requests), request.SessionID)

	// Get or create memory session (shared across batch)
	memory := ag.memoryManager.GetSession(request.SessionID)
	memory.AddDecision("batch_generation", "start", fmt.Sprintf("%d requests", len(request.Requests)),
		"Starting batch generation")

	results := make([]*GenerationResult, len(request.Requests))
	var errors []error

	// Process requests sequentially to maintain memory context
	// In a production system, you might parallelize with proper memory synchronization
	for i, req := range request.Requests {
		// Override session ID to use shared session
		req.SessionID = request.SessionID

		result, err := ag.Generate(req)
		if err != nil {
			errors = append(errors, fmt.Errorf("request %d failed: %w", i, err))
			// Create error result
			results[i] = &GenerationResult{
				Status:  "error",
				Message: err.Error(),
			}
		} else {
			results[i] = result
		}
	}

	// Record batch completion
	memory.AddDecision("batch_generation", "complete",
		fmt.Sprintf("%d succeeded, %d failed", len(request.Requests)-len(errors), len(errors)),
		"Batch generation completed")

	if len(errors) > 0 {
		return results, fmt.Errorf("batch generation had %d errors", len(errors))
	}

	return results, nil
}

// Helper methods for advanced generation

func (ag *AdvancedGenerator) extractTypeInfo(code string, typeSigs map[string]string) map[string]interface{} {
	typeInfo := make(map[string]interface{})
	typeInfo["signatures"] = typeSigs

	// Extract function signatures from code
	functions := ag.extractFunctions(code)
	typeInfo["functions"] = functions

	// Extract class information
	classes := ag.extractClasses(code)
	typeInfo["classes"] = classes

	return typeInfo
}

func (ag *AdvancedGenerator) buildTestFromTypesPrompt(request TestGenerationRequest, typeInfo map[string]interface{}, memory *generation.GenerationMemory) string {
	// Get test framework from memory or request
	framework := request.Framework
	if framework == "" {
		if saved, exists := memory.GetNamingPattern("test_framework"); exists {
			framework = saved
		} else {
			framework = "Test2::V0"
		}
	}

	prompt := fmt.Sprintf(`Generate comprehensive tests for the following Perl code:

Code:
%s

Type Signatures:
%v

Requirements:
- Use %s test framework
- Test all type constraints and edge cases
- Include positive and negative test cases
- Test type coercion and validation
- Add descriptive test names
- Include setup/teardown if needed
- Ensure 100%% code coverage

Generate only the test code:`,
		request.Code, typeInfo["signatures"], framework)

	return prompt
}

func (ag *AdvancedGenerator) extractMethod(request RefactoringRequest, memory *generation.GenerationMemory, validation *validation.ValidationResult) (string, int, error) {
	prompt := fmt.Sprintf(`Extract the following code into a method:

Original Code:
%s

Target to Extract:
%s

Requirements:
- Preserve all type information
- Create appropriate method signature
- Handle parameters and return values correctly
- Update calling code to use new method
- Maintain code functionality

Refactored code:`,
		request.Code, request.Target)

	return ag.generateWithRefinement(prompt, memory, "refactoring")
}

func (ag *AdvancedGenerator) renameElement(request RefactoringRequest, memory *generation.GenerationMemory, validation *validation.ValidationResult) (string, int, error) {
	prompt := fmt.Sprintf(`Rename element in the following code:

Original Code:
%s

Element to Rename:
%s

New Name:
%s

Requirements:
- Update all references to the element
- Preserve type information
- Maintain code functionality
- Update documentation if present

Refactored code:`,
		request.Code, request.Target, request.NewName)

	return ag.generateWithRefinement(prompt, memory, "refactoring")
}

func (ag *AdvancedGenerator) inlineCode(request RefactoringRequest, memory *generation.GenerationMemory, validation *validation.ValidationResult) (string, int, error) {
	prompt := fmt.Sprintf(`Inline the following element:

Original Code:
%s

Element to Inline:
%s

Requirements:
- Replace all calls with inlined code
- Preserve type information
- Maintain code functionality
- Remove the original definition

Refactored code:`,
		request.Code, request.Target)

	return ag.generateWithRefinement(prompt, memory, "refactoring")
}

func (ag *AdvancedGenerator) typesPreserved(original, refactored *validation.ValidationResult) bool {
	// Compare type information between original and refactored code
	// This is a simplified check - in production, you'd do deeper analysis
	return len(refactored.Errors) <= len(original.Errors)
}

func (ag *AdvancedGenerator) buildDocumentationPrompt(request DocumentationRequest, validation *validation.ValidationResult, memory *generation.GenerationMemory) string {
	prompt := fmt.Sprintf(`Generate %s documentation for the following Perl code:

Code:
%s

Requirements:
- Document all functions, methods, and classes
- Include parameter types and return types
- Add usage examples where appropriate
- Follow %s documentation standards
- Be concise but comprehensive

Generate only the documentation:`,
		request.Style, request.Code, request.Style)

	return prompt
}

func (ag *AdvancedGenerator) formatDocumentation(docs, style string) string {
	// Format documentation based on style
	switch style {
	case "pod":
		// Ensure POD formatting
		if !strings.HasPrefix(docs, "=") {
			docs = "=pod\n\n" + docs + "\n\n=cut"
		}
	case "markdown":
		// Ensure markdown formatting
		if !strings.HasPrefix(docs, "#") {
			docs = "# Documentation\n\n" + docs
		}
	case "inline":
		// Format as inline comments
		lines := strings.Split(docs, "\n")
		for i, line := range lines {
			if line != "" {
				lines[i] = "# " + line
			}
		}
		docs = strings.Join(lines, "\n")
	}

	return docs
}

func (ag *AdvancedGenerator) extractCursorContext(code string, cursorPos int) (string, string) {
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(code) {
		cursorPos = len(code)
	}

	prefix := code[:cursorPos]
	suffix := code[cursorPos:]

	return prefix, suffix
}

func (ag *AdvancedGenerator) buildCompletionPrompt(prefix, suffix, context string, memory *generation.GenerationMemory) string {
	// Get recent type choices for context
	recentDecisions := memory.GetRecentDecisions(5)
	typeContext := ag.extractTypeContext(recentDecisions)

	prompt := fmt.Sprintf(`Complete the Perl code at the cursor position:

Code before cursor:
%s

Code after cursor:
%s

Context:
%s

Type Context:
%s

Provide only the completion text that should be inserted at cursor position:`,
		prefix, suffix, context, typeContext)

	return prompt
}

func (ag *AdvancedGenerator) extractFunctions(code string) []string {
	var functions []string
	lines := strings.Split(code, "\n")

	for _, line := range lines {
		if strings.Contains(line, "sub ") || strings.Contains(line, "method ") {
			functions = append(functions, strings.TrimSpace(line))
		}
	}

	return functions
}

func (ag *AdvancedGenerator) extractClasses(code string) []string {
	var classes []string
	lines := strings.Split(code, "\n")

	for _, line := range lines {
		if strings.Contains(line, "class ") || strings.Contains(line, "package ") {
			classes = append(classes, strings.TrimSpace(line))
		}
	}

	return classes
}
