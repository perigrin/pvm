// ABOUTME: Code generation tool with collaborative sampling for LLM interaction
// ABOUTME: Implements function, class, and test generation using PVM's type system

package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/mcp/generation"
	"tamarou.com/pvm/internal/mcp/validation"
)

// Validator interface for code validation
type Validator interface {
	ValidateCode(ctx context.Context, code string, projectPath string) (*validation.ValidationResult, error)
}

// AutoFixer interface for automatic error fixing
type AutoFixer interface {
	AutoFix(ctx context.Context, code string, errors []validation.ValidationError, projectPath string) ([]validation.FixError, error)
}

// CodeGenerator implements collaborative code generation using sampling
type CodeGenerator struct {
	validator      Validator
	autoFixer      AutoFixer
	samplingClient *generation.SamplingClient
	memoryManager  *generation.MemoryManager
	logger         *log.Logger
}

// NewCodeGenerator creates a new code generator instance
func NewCodeGenerator(validator Validator, autoFixer AutoFixer, samplingClient *generation.SamplingClient, memoryManager *generation.MemoryManager, logger *log.Logger) *CodeGenerator {
	return &CodeGenerator{
		validator:      validator,
		autoFixer:      autoFixer,
		samplingClient: samplingClient,
		memoryManager:  memoryManager,
		logger:         logger,
	}
}

// GenerationRequest represents a code generation request
type GenerationRequest struct {
	Type          string `json:"type"`          // function, class, test
	Specification string `json:"specification"` // description of what to generate
	Context       string `json:"context"`       // optional context code
	ProjectPath   string `json:"project_path"`  // optional project path
	SessionID     string `json:"session_id"`    // memory session ID
}

// GenerationResult represents the result of code generation
type GenerationResult struct {
	Status           string                       `json:"status"`
	GeneratedCode    string                       `json:"generated_code"`
	ValidationResult *validation.ValidationResult `json:"validation_result,omitempty"`
	MemoryContext    map[string]interface{}       `json:"memory_context"`
	Iterations       int                          `json:"iterations"`
	Decisions        []generation.Decision        `json:"decisions"`
	Message          string                       `json:"message"`
	Timestamp        string                       `json:"timestamp"`
}

// Generate performs collaborative code generation
func (cg *CodeGenerator) Generate(request GenerationRequest) (*GenerationResult, error) {
	cg.logger.Infof("Starting code generation: type=%s, session=%s", request.Type, request.SessionID)

	// Get or create memory session
	memory := cg.memoryManager.GetSession(request.SessionID)

	// Record generation start
	memory.AddDecision("generation_start", request.Type, request.Specification,
		fmt.Sprintf("Starting %s generation", request.Type))

	var result *GenerationResult
	var err error

	switch request.Type {
	case "function":
		result, err = cg.generateFunction(request, memory)
	case "class":
		result, err = cg.generateClass(request, memory)
	case "test":
		result, err = cg.generateTest(request, memory)
	default:
		return nil, fmt.Errorf("unsupported generation type: %s", request.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	// Add final context and decisions
	result.MemoryContext = memory.GetContext()
	result.Decisions = memory.GetDecisions()
	result.Timestamp = time.Now().UTC().Format(time.RFC3339)

	cg.logger.Infof("Code generation completed: status=%s, iterations=%d", result.Status, result.Iterations)
	return result, nil
}

// generateFunction generates a Perl function using collaborative sampling
func (cg *CodeGenerator) generateFunction(request GenerationRequest, memory *generation.GenerationMemory) (*GenerationResult, error) {
	cg.logger.Debugf("Generating function with spec: %s", request.Specification)

	// Check memory for previous naming conventions
	namingConvention, hasNaming := memory.GetNamingPattern("function")
	if !hasNaming {
		// Sample for naming convention preference
		namePrompt := fmt.Sprintf(`What naming convention should I use for Perl functions?
Specification: %s
Context: %s
Options: snake_case, camelCase, PascalCase
Please respond with just the convention name.`, request.Specification, request.Context)

		nameResponse, err := cg.samplingClient.Sample(context.Background(), namePrompt, "")
		if err != nil {
			cg.logger.Warningf("Failed to sample naming convention: %v", err)
			namingConvention = "snake_case" // default
		} else {
			namingConvention = strings.TrimSpace(nameResponse.Content)
			if !isValidNamingConvention(namingConvention) {
				namingConvention = "snake_case" // fallback
			}
		}
		memory.SetNamingPattern("function", namingConvention)
	}

	// Sample for function signature and implementation
	functionPrompt := cg.buildFunctionPrompt(request, memory, namingConvention)
	code, iterations, err := cg.generateWithRefinement(functionPrompt, memory, "function")
	if err != nil {
		return nil, fmt.Errorf("function generation failed: %w", err)
	}

	// Validate the generated code
	validationResult, err := cg.validateAndFix(code, memory)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	status := "success"
	if validationResult != nil && len(validationResult.Errors) > 0 {
		status = "success_with_warnings"
	}

	return &GenerationResult{
		Status:           status,
		GeneratedCode:    code,
		ValidationResult: validationResult,
		Iterations:       iterations,
		Message:          fmt.Sprintf("Function generated successfully using %s naming", namingConvention),
	}, nil
}

// generateClass generates a Perl class using collaborative sampling
func (cg *CodeGenerator) generateClass(request GenerationRequest, memory *generation.GenerationMemory) (*GenerationResult, error) {
	cg.logger.Debugf("Generating class with spec: %s", request.Specification)

	// Check memory for class patterns
	classPattern, hasPattern := memory.GetNamingPattern("class")
	if !hasPattern {
		// Sample for class structure preference
		patternPrompt := fmt.Sprintf(`What Perl class pattern should I use?
Specification: %s
Context: %s
Options: modern_class (use v5.40 class), moose, classic_bless
Please respond with just the pattern name.`, request.Specification, request.Context)

		patternResponse, err := cg.samplingClient.Sample(context.Background(), patternPrompt, "")
		if err != nil {
			cg.logger.Warningf("Failed to sample class pattern: %v", err)
			classPattern = "modern_class" // default
		} else {
			classPattern = strings.TrimSpace(patternResponse.Content)
			if !isValidClassPattern(classPattern) {
				classPattern = "modern_class" // fallback
			}
		}
		memory.SetNamingPattern("class", classPattern)
	}

	// Sample for class implementation
	classPrompt := cg.buildClassPrompt(request, memory, classPattern)
	code, iterations, err := cg.generateWithRefinement(classPrompt, memory, "class")
	if err != nil {
		return nil, fmt.Errorf("class generation failed: %w", err)
	}

	// Validate the generated code
	validationResult, err := cg.validateAndFix(code, memory)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	status := "success"
	if validationResult != nil && len(validationResult.Errors) > 0 {
		status = "success_with_warnings"
	}

	return &GenerationResult{
		Status:           status,
		GeneratedCode:    code,
		ValidationResult: validationResult,
		Iterations:       iterations,
		Message:          fmt.Sprintf("Class generated successfully using %s pattern", classPattern),
	}, nil
}

// generateTest generates Perl test code using collaborative sampling
func (cg *CodeGenerator) generateTest(request GenerationRequest, memory *generation.GenerationMemory) (*GenerationResult, error) {
	cg.logger.Debugf("Generating test with spec: %s", request.Specification)

	// Check memory for test framework preference
	testFramework, hasFramework := memory.GetNamingPattern("test_framework")
	if !hasFramework {
		// Sample for test framework preference
		frameworkPrompt := fmt.Sprintf(`What Perl test framework should I use?
Specification: %s
Context: %s
Options: Test2::V0, Test::More, Test::Most
Please respond with just the framework name.`, request.Specification, request.Context)

		frameworkResponse, err := cg.samplingClient.Sample(context.Background(), frameworkPrompt, "")
		if err != nil {
			cg.logger.Warningf("Failed to sample test framework: %v", err)
			testFramework = "Test2::V0" // default
		} else {
			testFramework = strings.TrimSpace(frameworkResponse.Content)
			if !isValidTestFramework(testFramework) {
				testFramework = "Test2::V0" // fallback
			}
		}
		memory.SetNamingPattern("test_framework", testFramework)
	}

	// Sample for test implementation
	testPrompt := cg.buildTestPrompt(request, memory, testFramework)
	code, iterations, err := cg.generateWithRefinement(testPrompt, memory, "test")
	if err != nil {
		return nil, fmt.Errorf("test generation failed: %w", err)
	}

	// Validate the generated code
	validationResult, err := cg.validateAndFix(code, memory)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	status := "success"
	if validationResult != nil && len(validationResult.Errors) > 0 {
		status = "success_with_warnings"
	}

	return &GenerationResult{
		Status:           status,
		GeneratedCode:    code,
		ValidationResult: validationResult,
		Iterations:       iterations,
		Message:          fmt.Sprintf("Test generated successfully using %s", testFramework),
	}, nil
}

// generateWithRefinement performs iterative generation with LLM feedback
func (cg *CodeGenerator) generateWithRefinement(prompt string, memory *generation.GenerationMemory, codeType string) (string, int, error) {
	maxIterations := 3
	var bestCode string
	var bestScore float64 = -1

	for iteration := 1; iteration <= maxIterations; iteration++ {
		cg.logger.Debugf("Generation iteration %d/%d for %s", iteration, maxIterations, codeType)

		// Generate code
		codeResponse, err := cg.samplingClient.Sample(context.Background(), prompt, "")
		if err != nil {
			return "", iteration, fmt.Errorf("sampling failed at iteration %d: %w", iteration, err)
		}
		code := codeResponse.Content

		// Basic validation
		score := cg.scoreGeneration(code, codeType)
		memory.AddDecision("generation_iteration", fmt.Sprintf("iteration_%d", iteration), code,
			fmt.Sprintf("Generated code with score %.2f", score))

		if score > bestScore {
			bestCode = code
			bestScore = score
		}

		// If score is good enough, stop early
		if score >= 0.8 {
			cg.logger.Debugf("Early termination at iteration %d with score %.2f", iteration, score)
			return bestCode, iteration, nil
		}

		// Prepare refinement prompt for next iteration
		if iteration < maxIterations {
			prompt = cg.buildRefinementPrompt(prompt, code, score, memory)
		}
	}

	cg.logger.Debugf("Generation completed after %d iterations with best score %.2f", maxIterations, bestScore)
	return bestCode, maxIterations, nil
}

// validateAndFix validates generated code and applies auto-fixes if needed
func (cg *CodeGenerator) validateAndFix(code string, memory *generation.GenerationMemory) (*validation.ValidationResult, error) {
	// Validate the code
	result, err := cg.validator.ValidateCode(context.Background(), code, "")
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// If there are errors and auto-fix is enabled, try to fix them
	if len(result.Errors) > 0 && cg.autoFixer != nil {
		cg.logger.Debugf("Attempting auto-fix for %d errors", len(result.Errors))

		fixes, fixErr := cg.autoFixer.AutoFix(context.Background(), code, result.Errors, "")
		if fixErr == nil && len(fixes) > 0 {
			// Use the first successful fix
			fixedCode := fixes[0].FixedCode
			if fixedCode != code {
				// Re-validate the fixed code
				fixedResult, fixValidationErr := cg.validator.ValidateCode(context.Background(), fixedCode, "")
				if fixValidationErr == nil && len(fixedResult.Errors) < len(result.Errors) {
					memory.AddDecision("auto_fix", "validation_errors", "applied_fixes",
						fmt.Sprintf("Reduced errors from %d to %d", len(result.Errors), len(fixedResult.Errors)))
					return fixedResult, nil
				}
			}
		}
	}

	return result, nil
}

// Helper functions for prompt building and validation

func (cg *CodeGenerator) buildFunctionPrompt(request GenerationRequest, memory *generation.GenerationMemory, namingConvention string) string {
	// Get recent type decisions for context
	recentDecisions := memory.GetRecentDecisions(10)
	typeContext := cg.extractTypeContext(recentDecisions)

	prompt := fmt.Sprintf(`Generate a Perl function following these requirements:

Specification: %s
Naming Convention: %s
Context Code: %s
Type Context: %s

Requirements:
- Use modern Perl with 'use v5.40;'
- Include type annotations where beneficial
- Follow %s naming convention
- Include proper error handling
- Add brief documentation comments
- Ensure code is production-ready

Generate only the function code, no explanations:`,
		request.Specification, namingConvention, request.Context, typeContext, namingConvention)

	return prompt
}

func (cg *CodeGenerator) buildClassPrompt(request GenerationRequest, memory *generation.GenerationMemory, classPattern string) string {
	recentDecisions := memory.GetRecentDecisions(10)
	typeContext := cg.extractTypeContext(recentDecisions)

	prompt := fmt.Sprintf(`Generate a Perl class following these requirements:

Specification: %s
Class Pattern: %s
Context Code: %s
Type Context: %s

Requirements:
- Use %s pattern for class structure
- Use modern Perl with 'use v5.40;' if using modern_class
- Include type annotations for fields and methods
- Implement proper constructor and methods
- Add documentation for public interface
- Follow Perl best practices

Generate only the class code, no explanations:`,
		request.Specification, classPattern, request.Context, typeContext, classPattern)

	return prompt
}

func (cg *CodeGenerator) buildTestPrompt(request GenerationRequest, memory *generation.GenerationMemory, testFramework string) string {
	recentDecisions := memory.GetRecentDecisions(10)
	typeContext := cg.extractTypeContext(recentDecisions)

	prompt := fmt.Sprintf(`Generate Perl test code following these requirements:

Specification: %s
Test Framework: %s
Context Code: %s
Type Context: %s

Requirements:
- Use %s test framework
- Include comprehensive test cases
- Test both success and failure scenarios
- Use descriptive test names
- Include setup and teardown if needed
- Follow testing best practices

Generate only the test code, no explanations:`,
		request.Specification, testFramework, request.Context, typeContext, testFramework)

	return prompt
}

func (cg *CodeGenerator) buildRefinementPrompt(originalPrompt, previousCode string, score float64, memory *generation.GenerationMemory) string {
	issues := cg.identifyIssues(previousCode, score)

	prompt := fmt.Sprintf(`%s

Previous attempt (score: %.2f):
%s

Issues to address:
%s

Please improve the code addressing these issues:`,
		originalPrompt, score, previousCode, strings.Join(issues, "\n"))

	return prompt
}

func (cg *CodeGenerator) extractTypeContext(decisions []generation.Decision) string {
	var typeChoices []string
	for _, decision := range decisions {
		if decision.Type == "type_choice" {
			typeChoices = append(typeChoices, fmt.Sprintf("%s: %s", decision.Context, decision.Choice))
		}
	}

	if len(typeChoices) == 0 {
		return "No previous type decisions"
	}

	return "Previous type decisions: " + strings.Join(typeChoices, ", ")
}

func (cg *CodeGenerator) scoreGeneration(code, codeType string) float64 {
	score := 0.0
	maxScore := 5.0

	// Basic syntax checks
	if strings.Contains(code, "use v5.") {
		score += 1.0
	}
	if strings.Contains(code, "sub ") || strings.Contains(code, "method ") {
		score += 1.0
	}
	if strings.Contains(code, "#") { // has comments
		score += 0.5
	}

	// Type-specific checks
	switch codeType {
	case "function":
		if strings.Contains(code, "my ") && strings.Contains(code, "return") {
			score += 1.0
		}
	case "class":
		if strings.Contains(code, "class ") || strings.Contains(code, "package ") {
			score += 1.0
		}
	case "test":
		if strings.Contains(code, "ok(") || strings.Contains(code, "is(") {
			score += 2.0 // Give tests higher scores for basic assertions
		}
		if strings.Contains(code, "use Test") || strings.Contains(code, "use Test2") {
			score += 1.0 // Extra points for using test framework
		}
		if strings.Contains(code, "done_testing") {
			score += 0.5 // Extra points for proper test completion
		}
	}

	// Length check (not too short, not too long)
	lines := strings.Split(code, "\n")
	if len(lines) >= 5 && len(lines) <= 100 {
		score += 0.5
	}

	return score / maxScore
}

func (cg *CodeGenerator) identifyIssues(code string, score float64) []string {
	var issues []string

	if !strings.Contains(code, "use v5.") {
		issues = append(issues, "- Missing modern Perl version declaration")
	}
	if !strings.Contains(code, "#") {
		issues = append(issues, "- Missing documentation comments")
	}
	if score < 0.3 {
		issues = append(issues, "- Code structure needs improvement")
	}
	if score < 0.5 {
		issues = append(issues, "- Missing best practices implementation")
	}

	return issues
}

// Validation helper functions

func isValidNamingConvention(convention string) bool {
	valid := map[string]bool{
		"snake_case": true,
		"camelCase":  true,
		"PascalCase": true,
	}
	return valid[convention]
}

func isValidClassPattern(pattern string) bool {
	valid := map[string]bool{
		"modern_class":  true,
		"moose":         true,
		"classic_bless": true,
	}
	return valid[pattern]
}

func isValidTestFramework(framework string) bool {
	valid := map[string]bool{
		"Test2::V0":  true,
		"Test::More": true,
		"Test::Most": true,
	}
	return valid[framework]
}
