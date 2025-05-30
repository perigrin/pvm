// ABOUTME: Advanced code generation features including test generation, refactoring, and documentation
// ABOUTME: Extends basic generation with type-aware transformations and batch operations

package tools

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/mcp/generation"
)

// SimpleType represents a simple type for internal use
type SimpleType struct {
	Name string
}

// TypeParser interface for parsing type signatures
type TypeParser interface {
	ParseTypeSignature(signature string) (*SimpleType, error)
	ExtractTypeFromCode(code string) ([]*SimpleType, error)
}

// DocumentationGenerator generates documentation from typed code
type DocumentationGenerator struct {
	samplingClient *generation.SamplingClient
	logger         *log.Logger
}

// NewDocumentationGenerator creates a new documentation generator
func NewDocumentationGenerator(samplingClient *generation.SamplingClient, logger *log.Logger) *DocumentationGenerator {
	return &DocumentationGenerator{
		samplingClient: samplingClient,
		logger:         logger,
	}
}

// TestGenerator generates test suites from type signatures
type TestGenerator struct {
	samplingClient *generation.SamplingClient
	typeParser     TypeParser
	logger         *log.Logger
}

// NewTestGenerator creates a new test generator
func NewTestGenerator(samplingClient *generation.SamplingClient, typeParser TypeParser, logger *log.Logger) *TestGenerator {
	return &TestGenerator{
		samplingClient: samplingClient,
		typeParser:     typeParser,
		logger:         logger,
	}
}

// RefactoringEngine performs type-preserving code transformations
type RefactoringEngine struct {
	samplingClient *generation.SamplingClient
	typeParser     TypeParser
	validator      Validator
	logger         *log.Logger
}

// NewRefactoringEngine creates a new refactoring engine
func NewRefactoringEngine(samplingClient *generation.SamplingClient, typeParser TypeParser, validator Validator, logger *log.Logger) *RefactoringEngine {
	return &RefactoringEngine{
		samplingClient: samplingClient,
		typeParser:     typeParser,
		validator:      validator,
		logger:         logger,
	}
}

// CompletionEngine provides code completion suggestions
type CompletionEngine struct {
	samplingClient *generation.SamplingClient
	typeParser     TypeParser
	logger         *log.Logger
}

// NewCompletionEngine creates a new completion engine
func NewCompletionEngine(samplingClient *generation.SamplingClient, typeParser TypeParser, logger *log.Logger) *CompletionEngine {
	return &CompletionEngine{
		samplingClient: samplingClient,
		typeParser:     typeParser,
		logger:         logger,
	}
}

// TestGenRequest represents a request to generate tests from types
type TestGenRequest struct {
	TypeSignature string `json:"type_signature"`
	FunctionName  string `json:"function_name"`
	Context       string `json:"context"`
	Framework     string `json:"framework"`
}

// TestGenerationResult represents generated test code
type TestGenerationResult struct {
	TestCode  string   `json:"test_code"`
	TestCases []string `json:"test_cases"`
	Coverage  float64  `json:"coverage"`
	Warnings  []string `json:"warnings"`
}

// GenerateTestsFromType generates comprehensive test suites from type signatures
func (tg *TestGenerator) GenerateTestsFromType(request TestGenRequest) (*TestGenerationResult, error) {
	tg.logger.Infof("Generating tests for function: %s with type: %s", request.FunctionName, request.TypeSignature)

	// Parse the type signature
	typeInfo, err := tg.typeParser.ParseTypeSignature(request.TypeSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to parse type signature: %w", err)
	}

	// Determine test framework
	framework := request.Framework
	if framework == "" {
		framework = "Test2::V0" // default
	}

	// Generate test cases based on type constraints
	testCases := tg.generateTestCasesFromType(typeInfo, request.FunctionName)

	// Build comprehensive test prompt
	prompt := tg.buildTestGenerationPrompt(request.FunctionName, typeInfo, testCases, framework, request.Context)

	// Sample for test implementation
	response, err := tg.samplingClient.Sample(context.Background(), prompt, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate tests: %w", err)
	}

	// Calculate coverage estimate
	coverage := tg.estimateTestCoverage(testCases, typeInfo)

	// Identify any warnings
	warnings := tg.identifyTestWarnings(typeInfo, testCases)

	return &TestGenerationResult{
		TestCode:  response.Content,
		TestCases: testCases,
		Coverage:  coverage,
		Warnings:  warnings,
	}, nil
}

// RefactoringRequest represents a code refactoring request
type RefactoringRequest struct {
	Code            string `json:"code"`
	RefactoringType string `json:"refactoring_type"` // extract_method, rename, inline, etc.
	Target          string `json:"target"`           // what to refactor
	NewName         string `json:"new_name"`         // for rename operations
	PreserveTypes   bool   `json:"preserve_types"`   // ensure type safety
}

// RefactoringResult represents the result of a refactoring operation
type RefactoringResult struct {
	RefactoredCode string   `json:"refactored_code"`
	Changes        []string `json:"changes"`
	TypesSafe      bool     `json:"types_safe"`
	Warnings       []string `json:"warnings"`
}

// Refactor performs type-preserving code transformations
func (re *RefactoringEngine) Refactor(request RefactoringRequest) (*RefactoringResult, error) {
	re.logger.Infof("Performing %s refactoring on target: %s", request.RefactoringType, request.Target)

	// Extract current types from code
	currentTypes, err := re.typeParser.ExtractTypeFromCode(request.Code)
	if err != nil {
		re.logger.Warningf("Failed to extract types: %v", err)
	}

	// Build refactoring prompt
	prompt := re.buildRefactoringPrompt(request, currentTypes)

	// Sample for refactored code
	response, err := re.samplingClient.Sample(context.Background(), prompt, "")
	if err != nil {
		return nil, fmt.Errorf("failed to refactor code: %w", err)
	}

	refactoredCode := response.Content

	// Validate type preservation if requested
	typesSafe := true
	if request.PreserveTypes && re.validator != nil {
		originalResult, _ := re.validator.ValidateCode(context.Background(), request.Code, "")
		refactoredResult, _ := re.validator.ValidateCode(context.Background(), refactoredCode, "")

		if refactoredResult != nil && originalResult != nil {
			typesSafe = len(refactoredResult.Errors) <= len(originalResult.Errors)
		}
	}

	// Identify changes
	changes := re.identifyChanges(request.Code, refactoredCode, request.RefactoringType)

	// Generate warnings
	warnings := re.generateRefactoringWarnings(request, typesSafe)

	return &RefactoringResult{
		RefactoredCode: refactoredCode,
		Changes:        changes,
		TypesSafe:      typesSafe,
		Warnings:       warnings,
	}, nil
}

// DocumentationRequest represents a documentation generation request
type DocumentationRequest struct {
	Code         string `json:"code"`
	DocType      string `json:"doc_type"`      // pod, inline, both
	IncludeTypes bool   `json:"include_types"` // include type information
	Verbose      bool   `json:"verbose"`       // detailed documentation
}

// DocumentationResult represents generated documentation
type DocumentationResult struct {
	Documentation string            `json:"documentation"`
	Sections      map[string]string `json:"sections"`
	TypeInfo      []string          `json:"type_info"`
}

// GenerateDocumentation creates documentation from typed code
func (dg *DocumentationGenerator) GenerateDocumentation(request DocumentationRequest) (*DocumentationResult, error) {
	dg.logger.Infof("Generating %s documentation", request.DocType)

	// Build documentation prompt
	prompt := dg.buildDocumentationPrompt(request)

	// Sample for documentation
	response, err := dg.samplingClient.Sample(context.Background(), prompt, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate documentation: %w", err)
	}

	// Parse sections from generated documentation
	sections := dg.parseDocumentationSections(response.Content)

	// Extract type information if requested
	var typeInfo []string
	if request.IncludeTypes {
		typeInfo = dg.extractTypeDocumentation(request.Code)
	}

	return &DocumentationResult{
		Documentation: response.Content,
		Sections:      sections,
		TypeInfo:      typeInfo,
	}, nil
}

// CompletionRequest represents a code completion request
type CompletionRequest struct {
	PartialCode    string `json:"partial_code"`
	CursorPosition int    `json:"cursor_position"`
	Context        string `json:"context"`
	MaxSuggestions int    `json:"max_suggestions"`
}

// CompletionResult represents code completion suggestions
type CompletionResult struct {
	Suggestions []CompletionSuggestion `json:"suggestions"`
	TypeHints   []string               `json:"type_hints"`
}

// CompletionSuggestion represents a single completion suggestion
type CompletionSuggestion struct {
	Text        string  `json:"text"`
	Description string  `json:"description"`
	TypeInfo    string  `json:"type_info"`
	Score       float64 `json:"score"`
}

// Complete provides code completion suggestions
func (ce *CompletionEngine) Complete(request CompletionRequest) (*CompletionResult, error) {
	ce.logger.Debugf("Generating completions at position %d", request.CursorPosition)

	// Extract context around cursor
	contextInfo := ce.extractCompletionContext(request.PartialCode, request.CursorPosition)

	// Build completion prompt
	prompt := ce.buildCompletionPrompt(contextInfo, request)

	// Sample for completions
	response, err := ce.samplingClient.Sample(context.Background(), prompt, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate completions: %w", err)
	}

	// Parse suggestions from response
	suggestions := ce.parseSuggestions(response.Content, request.MaxSuggestions)

	// Extract type hints
	typeHints := ce.extractTypeHints(contextInfo, suggestions)

	return &CompletionResult{
		Suggestions: suggestions,
		TypeHints:   typeHints,
	}, nil
}

// BatchGenerationRequest represents a batch generation request
type BatchGenerationRequest struct {
	Requests  []GenerationRequest `json:"requests"`
	Parallel  bool                `json:"parallel"`
	SessionID string              `json:"session_id"`
}

// BatchGenerationResult represents results from batch generation
type BatchGenerationResult struct {
	Results   []*GenerationResult `json:"results"`
	Succeeded int                 `json:"succeeded"`
	Failed    int                 `json:"failed"`
	Errors    []string            `json:"errors"`
}

// GenerateBatch performs batch code generation operations
func (cg *CodeGenerator) GenerateBatch(request BatchGenerationRequest) (*BatchGenerationResult, error) {
	cg.logger.Infof("Processing batch generation with %d requests", len(request.Requests))

	result := &BatchGenerationResult{
		Results: make([]*GenerationResult, len(request.Requests)),
		Errors:  []string{},
	}

	// Process each request
	for i, req := range request.Requests {
		// Use the batch session ID if individual request doesn't have one
		if req.SessionID == "" {
			req.SessionID = request.SessionID
		}

		genResult, err := cg.Generate(req)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("Request %d failed: %v", i, err))
			result.Results[i] = &GenerationResult{
				Status:  "error",
				Message: err.Error(),
			}
		} else {
			result.Succeeded++
			result.Results[i] = genResult
		}
	}

	cg.logger.Infof("Batch generation completed: %d succeeded, %d failed", result.Succeeded, result.Failed)
	return result, nil
}

// Helper methods for TestGenerator

func (tg *TestGenerator) generateTestCasesFromType(typeInfo *SimpleType, functionName string) []string {
	var testCases []string

	// Generate boundary test cases
	testCases = append(testCases, fmt.Sprintf("Test %s with valid input", functionName))
	testCases = append(testCases, fmt.Sprintf("Test %s with invalid input", functionName))
	testCases = append(testCases, fmt.Sprintf("Test %s with edge cases", functionName))

	// Type-specific test cases
	if typeInfo != nil {
		testCases = append(testCases, fmt.Sprintf("Test %s type constraints", functionName))
		testCases = append(testCases, fmt.Sprintf("Test %s return type correctness", functionName))
	}

	return testCases
}

func (tg *TestGenerator) buildTestGenerationPrompt(functionName string, typeInfo *SimpleType, testCases []string, framework, context string) string {
	prompt := fmt.Sprintf(`Generate comprehensive Perl tests for function '%s' with the following requirements:

Function Type Signature: %v
Test Framework: %s
Context: %s

Required Test Cases:
%s

Requirements:
- Test all type constraints and boundaries
- Include both positive and negative test cases
- Use %s test framework idioms
- Add descriptive test names
- Include setup/teardown if needed
- Ensure high code coverage

Generate only the test code, no explanations:`,
		functionName, typeInfo, framework, context,
		strings.Join(testCases, "\n- "),
		framework)

	return prompt
}

func (tg *TestGenerator) estimateTestCoverage(testCases []string, typeInfo *SimpleType) float64 {
	// Simple heuristic: more test cases = better coverage
	baseCoverage := float64(len(testCases)) * 0.15
	if baseCoverage > 0.95 {
		baseCoverage = 0.95
	}
	return baseCoverage
}

func (tg *TestGenerator) identifyTestWarnings(typeInfo *SimpleType, testCases []string) []string {
	var warnings []string

	if len(testCases) < 3 {
		warnings = append(warnings, "Limited test coverage - consider adding more test cases")
	}

	if typeInfo == nil {
		warnings = append(warnings, "No type information available - tests may not cover all constraints")
	}

	return warnings
}

// Helper methods for RefactoringEngine

func (re *RefactoringEngine) buildRefactoringPrompt(request RefactoringRequest, currentTypes []*SimpleType) string {
	typeContext := ""
	if len(currentTypes) > 0 {
		typeContext = fmt.Sprintf("\nCurrent Types: %v", currentTypes)
	}

	prompt := fmt.Sprintf(`Perform %s refactoring on the following Perl code:

Code:
%s

Target: %s
%s

Requirements:
- Preserve all type annotations and constraints
- Maintain the same behavior
- Follow Perl best practices
- Keep the code readable and maintainable
- %s

Generate only the refactored code, no explanations:`,
		request.RefactoringType, request.Code, request.Target, typeContext,
		re.getRefactoringSpecificRequirements(request))

	return prompt
}

func (re *RefactoringEngine) getRefactoringSpecificRequirements(request RefactoringRequest) string {
	switch request.RefactoringType {
	case "extract_method":
		return "Extract the target code into a well-named method with appropriate parameters"
	case "rename":
		return fmt.Sprintf("Rename all occurrences of '%s' to '%s'", request.Target, request.NewName)
	case "inline":
		return "Inline the target method/variable at all call sites"
	case "simplify":
		return "Simplify the code while maintaining functionality"
	default:
		return "Apply the requested refactoring"
	}
}

func (re *RefactoringEngine) identifyChanges(original, refactored, refactoringType string) []string {
	var changes []string

	// Simple line-based diff identification
	originalLines := strings.Split(original, "\n")
	refactoredLines := strings.Split(refactored, "\n")

	if len(originalLines) != len(refactoredLines) {
		changes = append(changes, fmt.Sprintf("Line count changed from %d to %d", len(originalLines), len(refactoredLines)))
	}

	switch refactoringType {
	case "extract_method":
		changes = append(changes, "Extracted code into new method")
	case "rename":
		changes = append(changes, "Renamed identifiers throughout code")
	case "inline":
		changes = append(changes, "Inlined method/variable at call sites")
	}

	return changes
}

func (re *RefactoringEngine) generateRefactoringWarnings(request RefactoringRequest, typesSafe bool) []string {
	var warnings []string

	if !typesSafe && request.PreserveTypes {
		warnings = append(warnings, "Type safety may have been compromised - please verify")
	}

	if request.RefactoringType == "extract_method" && len(request.Target) > 100 {
		warnings = append(warnings, "Large extraction - consider breaking into smaller methods")
	}

	return warnings
}

// Helper methods for DocumentationGenerator

func (dg *DocumentationGenerator) buildDocumentationPrompt(request DocumentationRequest) string {
	docTypeInstructions := ""
	switch request.DocType {
	case "pod":
		docTypeInstructions = "Generate POD (Plain Old Documentation) format documentation"
	case "inline":
		docTypeInstructions = "Generate inline comments throughout the code"
	case "both":
		docTypeInstructions = "Generate both POD documentation and inline comments"
	}

	verbosity := "concise"
	if request.Verbose {
		verbosity = "detailed"
	}

	prompt := fmt.Sprintf(`Generate %s Perl documentation for the following code:

Code:
%s

Requirements:
- %s
- Include parameter descriptions and return values
- Document any type constraints if present
- Add usage examples where appropriate
- Make documentation %s
- Follow Perl documentation best practices

Generate the documentation:`,
		verbosity, request.Code, docTypeInstructions, verbosity)

	return prompt
}

func (dg *DocumentationGenerator) parseDocumentationSections(documentation string) map[string]string {
	sections := make(map[string]string)

	// Parse POD sections - handle nested headers correctly
	lines := strings.Split(documentation, "\n")
	var currentSection string
	var content []string

	for _, line := range lines {
		if strings.HasPrefix(line, "=head1 ") {
			// Save previous section if exists
			if currentSection != "" {
				sections[currentSection] = strings.TrimSpace(strings.Join(content, "\n"))
			}
			// Start new section
			currentSection = strings.TrimSpace(strings.TrimPrefix(line, "=head1"))
			content = []string{}
		} else if currentSection != "" {
			content = append(content, line)
		}
	}

	// Save the last section
	if currentSection != "" {
		sections[currentSection] = strings.TrimSpace(strings.Join(content, "\n"))
	}

	// If no POD sections found, treat as single section
	if len(sections) == 0 {
		sections["main"] = documentation
	}

	return sections
}

func (dg *DocumentationGenerator) extractTypeDocumentation(code string) []string {
	var typeInfo []string

	// Extract type annotations
	typeRegex := regexp.MustCompile(`my\s+(\w+(?:\[[\w\[\]]+\])?)\s+\$(\w+)`)
	matches := typeRegex.FindAllStringSubmatch(code, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			typeInfo = append(typeInfo, fmt.Sprintf("$%s: %s", match[2], match[1]))
		}
	}

	return typeInfo
}

// Helper methods for CompletionEngine

func (ce *CompletionEngine) extractCompletionContext(code string, position int) map[string]string {
	context := make(map[string]string)

	// Extract line context
	lines := strings.Split(code, "\n")
	currentLine := 0
	currentPos := 0

	for i, line := range lines {
		if currentPos+len(line)+1 > position {
			currentLine = i
			break
		}
		currentPos += len(line) + 1
	}

	if currentLine < len(lines) {
		context["current_line"] = lines[currentLine]
		context["line_number"] = fmt.Sprintf("%d", currentLine+1)
	}

	// Extract preceding token - handle partial words
	if position > 0 && position <= len(code) {
		// Find the word at the cursor position
		start := position
		end := position

		// Move start backward to find word boundary
		for start > 0 && (isAlphaNumeric(code[start-1]) || code[start-1] == '_' || code[start-1] == '$' || code[start-1] == '@' || code[start-1] == '%') {
			start--
		}

		// Move end forward to find word boundary
		for end < len(code) && (isAlphaNumeric(code[end]) || code[end] == '_') {
			end++
		}

		if start < end {
			context["preceding_token"] = code[start:end]
		} else {
			// Fallback to previous complete word
			beforeCursor := code[:position]
			tokens := strings.Fields(beforeCursor)
			if len(tokens) > 0 {
				context["preceding_token"] = tokens[len(tokens)-1]
			}
		}
	}

	return context
}

// isAlphaNumeric checks if character is alphanumeric
func isAlphaNumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func (ce *CompletionEngine) buildCompletionPrompt(contextInfo map[string]string, request CompletionRequest) string {
	prompt := fmt.Sprintf(`Provide code completion suggestions for the following Perl code:

Partial Code:
%s

Cursor Position: %d
Current Line: %s
Preceding Token: %s

Additional Context:
%s

Requirements:
- Suggest the most likely completions
- Include method names, variable names, or keywords as appropriate
- Consider Perl syntax and idioms
- Provide up to %d suggestions
- Include brief descriptions for each suggestion

Format each suggestion as:
SUGGESTION: <text>
DESCRIPTION: <brief description>
TYPE: <type info if available>

Generate suggestions:`,
		request.PartialCode, request.CursorPosition,
		contextInfo["current_line"], contextInfo["preceding_token"],
		request.Context, request.MaxSuggestions)

	return prompt
}

func (ce *CompletionEngine) parseSuggestions(response string, maxSuggestions int) []CompletionSuggestion {
	var suggestions []CompletionSuggestion

	// Parse structured suggestions from response
	suggestionRegex := regexp.MustCompile(`SUGGESTION:\s*(.+?)\nDESCRIPTION:\s*(.+?)(?:\nTYPE:\s*(.+?))?(?:\n|$)`)
	matches := suggestionRegex.FindAllStringSubmatch(response, -1)

	for i, match := range matches {
		if i >= maxSuggestions {
			break
		}

		suggestion := CompletionSuggestion{
			Text:        strings.TrimSpace(match[1]),
			Description: strings.TrimSpace(match[2]),
			Score:       1.0 - (float64(i) * 0.1), // Simple scoring based on order
		}

		if len(match) > 3 && match[3] != "" {
			suggestion.TypeInfo = strings.TrimSpace(match[3])
		}

		suggestions = append(suggestions, suggestion)
	}

	return suggestions
}

func (ce *CompletionEngine) extractTypeHints(contextInfo map[string]string, suggestions []CompletionSuggestion) []string {
	var hints []string
	seen := make(map[string]bool)

	for _, suggestion := range suggestions {
		if suggestion.TypeInfo != "" && !seen[suggestion.TypeInfo] {
			hints = append(hints, suggestion.TypeInfo)
			seen[suggestion.TypeInfo] = true
		}
	}

	return hints
}
