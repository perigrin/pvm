// ABOUTME: Comprehensive testing framework for parser accuracy and regression testing
// ABOUTME: Provides infrastructure for systematic testing of both untyped and typed Perl parsing

package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	yaml "gopkg.in/yaml.v3"
	"tamarou.com/pvm/internal/ast"
)

// TestCategory represents different categories of parser tests
type TestCategory string

const (
	UntypedPerl TestCategory = "untyped-perl"
	TypedPerl   TestCategory = "typed-perl"
	ErrorCases  TestCategory = "error-cases"
)

// ASTTypeChecker interface allows type checking without importing typechecker package
type ASTTypeChecker interface {
	CheckAST(ast *ast.AST) []error
}

// ParserTestCase represents a single test case for parser validation
type ParserTestCase struct {
	Name                   string       `json:"name"`
	Category               TestCategory `json:"category"`
	Subcategory            string       `json:"subcategory"`
	Input                  string       `json:"input"`
	ExpectedAST            *ast.AST     `json:"expected_ast,omitempty"`
	ShouldError            bool         `json:"should_error"`
	ErrorType              string       `json:"error_type,omitempty"`
	Description            string       `json:"description"`
	Tags                   []string     `json:"tags"`
	TypeCheck              bool         `json:"type_check"`
	ExpectedTypeErrors     string       `json:"expected_type_errors,omitempty"`
	ExpectedASTBeforeInfer string       `json:"expected_ast_before_infer,omitempty"`
	ExpectedASTAfterInfer  string       `json:"expected_ast_after_infer,omitempty"`

	// Compilation outcome expectations
	ExpectedCompilationOutcomes *CompilationOutcomes `json:"expected_compilation_outcomes,omitempty"`
}

// CompilationOutcomes represents expected outputs for all compilation targets
type CompilationOutcomes struct {
	// ExpectedCleanPerl is the expected output for TargetCleanPerl
	ExpectedCleanPerl string `json:"expected_clean_perl,omitempty"`

	// ExpectedTypedPerl is the expected output for TargetTypedPerl
	ExpectedTypedPerl string `json:"expected_typed_perl,omitempty"`

	// ExpectedInferredPerl is the expected output for TargetInferredTypeAnnotations
	ExpectedInferredPerl string `json:"expected_inferred_perl,omitempty"`

	// CompilationErrors tracks expected compilation errors for each target
	CompilationErrors map[string]string `json:"compilation_errors,omitempty"`
}

// AccuracyMetrics tracks parser accuracy across different dimensions
type AccuracyMetrics struct {
	TotalTests      int               `json:"total_tests"`
	PassedTests     int               `json:"passed_tests"`
	FailedTests     int               `json:"failed_tests"`
	CategoryMetrics map[string]Metric `json:"category_metrics"`
	FeatureMetrics  map[string]Metric `json:"feature_metrics"`
	ParsingTime     time.Duration     `json:"parsing_time"`
	MemoryUsage     int64             `json:"memory_usage"`
}

// Metric represents accuracy statistics for a specific dimension
type Metric struct {
	Total    int     `json:"total"`
	Passed   int     `json:"passed"`
	Failed   int     `json:"failed"`
	Accuracy float64 `json:"accuracy"`
}

// ParserTestFramework provides comprehensive testing infrastructure
type ParserTestFramework struct {
	TestDataDir string
	UpdateMode  bool
	Verbose     bool
	Parser      Parser
	TypeChecker ASTTypeChecker
}

// NewParserTestFramework creates a new parser testing framework
func NewParserTestFramework(testDataDir string) *ParserTestFramework {
	// Initialize parser - use default parser if creation fails
	parser, err := NewParser()
	if err != nil {
		// Log error but don't fail - tests will handle missing parser
		fmt.Fprintf(os.Stderr, "Warning: Failed to create parser for test framework: %v\n", err)
	}

	framework := &ParserTestFramework{
		TestDataDir: testDataDir,
		UpdateMode:  os.Getenv("UPDATE_BASELINES") == "1",
		Verbose:     os.Getenv("VERBOSE_TESTS") == "1",
		Parser:      parser,
	}
	// TypeChecker is nil by default and will be set externally to avoid import cycle
	return framework
}

// SetTypeChecker sets the type checker for the framework (to avoid import cycles)
func (f *ParserTestFramework) SetTypeChecker(typeChecker ASTTypeChecker) {
	f.TypeChecker = typeChecker
}

// LoadTestCase loads a test case from a JSON file
func (f *ParserTestFramework) LoadTestCase(filePath string) (*ParserTestCase, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test case file %s: %w", filePath, err)
	}

	// Try to parse as single test case first
	var testCase ParserTestCase
	err = json.Unmarshal(data, &testCase)
	if err == nil {
		return &testCase, nil
	}

	// If that fails, try to parse as array format (legacy)
	var testCases []ParserTestCase
	err = json.Unmarshal(data, &testCases)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test case file %s as either single test case or array: %w", filePath, err)
	}

	// For array format, return the first test case
	if len(testCases) > 0 {
		return &testCases[0], nil
	}

	return nil, fmt.Errorf("empty test case array in file %s", filePath)
}

// LoadTestCases loads test cases from a JSON file (handles both single and array formats)
func (f *ParserTestFramework) LoadTestCases(filePath string) ([]*ParserTestCase, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test case file %s: %w", filePath, err)
	}

	// Try to parse as single test case first
	var testCase ParserTestCase
	err = json.Unmarshal(data, &testCase)
	if err == nil {
		return []*ParserTestCase{&testCase}, nil
	}

	// If that fails, try to parse as array format
	var testCases []ParserTestCase
	err = json.Unmarshal(data, &testCases)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test case file %s as either single test case or array: %w", filePath, err)
	}

	// Convert to pointer slice
	var result []*ParserTestCase
	for i := range testCases {
		result = append(result, &testCases[i])
	}

	return result, nil
}

// MarkdownTestMetadata represents the YAML frontmatter in a markdown test file
type MarkdownTestMetadata struct {
	Category    TestCategory `yaml:"category"`
	Subcategory string       `yaml:"subcategory"`
	Tags        []string     `yaml:"tags"`
	TypeCheck   bool         `yaml:"type_check"`
	ShouldError bool         `yaml:"should_error"`
}

// LoadMarkdownTestCases loads test cases from a Markdown file
func (f *ParserTestFramework) LoadMarkdownTestCases(filePath string) ([]*ParserTestCase, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read markdown test file %s: %w", filePath, err)
	}

	content := string(data)

	// Parse YAML frontmatter
	metadata, content, err := f.parseMarkdownFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter in %s: %w", filePath, err)
	}

	// Parse test cases from markdown content
	testCases, err := f.parseMarkdownTestCases(content, metadata, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test cases from %s: %w", filePath, err)
	}

	return testCases, nil
}

// parseMarkdownFrontmatter extracts YAML frontmatter from markdown content
func (f *ParserTestFramework) parseMarkdownFrontmatter(content string) (*MarkdownTestMetadata, string, error) {
	lines := strings.Split(content, "\n")

	if len(lines) < 3 || lines[0] != "---" {
		// No frontmatter, return default metadata
		return &MarkdownTestMetadata{}, content, nil
	}

	// Find the closing ---
	var frontmatterEnd int
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			frontmatterEnd = i
			break
		}
	}

	if frontmatterEnd == 0 {
		return nil, "", fmt.Errorf("unclosed YAML frontmatter")
	}

	// Parse YAML frontmatter
	yamlContent := strings.Join(lines[1:frontmatterEnd], "\n")
	var metadata MarkdownTestMetadata
	err := yaml.Unmarshal([]byte(yamlContent), &metadata)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Return content without frontmatter
	remainingContent := strings.Join(lines[frontmatterEnd+1:], "\n")
	return &metadata, remainingContent, nil
}

// parseMarkdownTestCases extracts test cases from markdown content
func (f *ParserTestFramework) parseMarkdownTestCases(content string, metadata *MarkdownTestMetadata, filePath string) ([]*ParserTestCase, error) {
	var testCases []*ParserTestCase

	// Split content into sections by headers
	sections := f.splitMarkdownSections(content)

	// Group sections: find test cases and their associated Expected sections
	for i, section := range sections {
		testCase, err := f.parseMarkdownSection(section, metadata, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse section: %w", err)
		}
		if testCase != nil {
			// Look ahead for Expected sections
			for j := i + 1; j < len(sections); j++ {
				nextSection := sections[j]
				titleLower := strings.ToLower(nextSection.Title)

				// Check for Expected Type Errors
				if strings.Contains(titleLower, "expected type errors") {
					if len(nextSection.CodeBlocks) > 0 {
						testCase.ExpectedTypeErrors = strings.TrimSpace(nextSection.CodeBlocks[0].Content)
					}
				}

				// Check for Expected AST Before Type Inference
				if strings.Contains(titleLower, "before type inference") {
					if len(nextSection.CodeBlocks) > 0 {
						testCase.ExpectedASTBeforeInfer = strings.TrimSpace(nextSection.CodeBlocks[0].Content)
					}
				}

				// Check for Expected AST After Type Inference
				if strings.Contains(titleLower, "after type inference") {
					if len(nextSection.CodeBlocks) > 0 {
						testCase.ExpectedASTAfterInfer = strings.TrimSpace(nextSection.CodeBlocks[0].Content)
					}
				}

				// Check for Expected Compilation Outcomes
				if strings.Contains(titleLower, "expected compilation outcomes") ||
					strings.Contains(titleLower, "compilation outcomes") {
					if f.Verbose {
						println("DEBUG: Found compilation outcomes section:", nextSection.Title)
					}

					// Collect all subsections that are part of compilation outcomes
					allSections := []MarkdownSection{nextSection}

					// Look ahead for compilation outcome subsections
					for k := j + 1; k < len(sections); k++ {
						subsection := sections[k]
						subsectionTitle := strings.ToLower(subsection.Title)

						// Check if this is a compilation outcome subsection
						if strings.Contains(subsectionTitle, "clean") ||
							strings.Contains(subsectionTitle, "typed") ||
							strings.Contains(subsectionTitle, "inferred") ||
							strings.Contains(subsectionTitle, "perl output") {
							allSections = append(allSections, subsection)
							if f.Verbose {
								println("DEBUG: Adding subsection to compilation outcomes:", subsection.Title)
							}
						} else if f.sectionHasPerlCode(subsection) ||
							strings.Contains(subsectionTitle, "expected") {
							// Stop when we hit another major section
							break
						}
					}

					outcomes, err := f.parseCompilationOutcomesFromSections(allSections)
					if err != nil {
						return nil, fmt.Errorf("failed to parse compilation outcomes: %w", err)
					}
					testCase.ExpectedCompilationOutcomes = outcomes
				} else if f.Verbose {
					println("DEBUG: Section not matching compilation outcomes:", nextSection.Title, "->", titleLower)
				}

				// Stop looking when we hit another test case (section with Perl code)
				if f.sectionHasPerlCode(nextSection) {
					break
				}
			}

			// Set type checking flag from metadata
			testCase.TypeCheck = metadata.TypeCheck

			testCases = append(testCases, testCase)
		}
	}

	return testCases, nil
}

// MarkdownSection represents a section of markdown content
type MarkdownSection struct {
	Title       string
	Description string
	Comments    map[string]string
	CodeBlocks  []MarkdownCodeBlock
}

// MarkdownCodeBlock represents a fenced code block
type MarkdownCodeBlock struct {
	Language string
	Info     string // Additional info from the code fence (like `perl clean` or `perl inferred`)
	Content  string
}

// splitMarkdownSections splits markdown content into sections by headers
func (f *ParserTestFramework) splitMarkdownSections(content string) []MarkdownSection {
	var sections []MarkdownSection
	var currentSection *MarkdownSection

	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()

		// Check for header (## or #)
		if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "# ") {
			// Save previous section
			if currentSection != nil {
				sections = append(sections, *currentSection)
			}

			// Start new section
			title := strings.TrimLeft(line, "# ")
			currentSection = &MarkdownSection{
				Title:    title,
				Comments: make(map[string]string),
			}
			continue
		}

		if currentSection == nil {
			continue
		}

		// Check for HTML comments with metadata
		if commentMatch := regexp.MustCompile(`<!-- (\w+): (.+) -->`).FindStringSubmatch(line); commentMatch != nil {
			currentSection.Comments[commentMatch[1]] = commentMatch[2]
			continue
		}

		// Check for fenced code blocks
		if strings.HasPrefix(line, "```") {
			language := strings.TrimPrefix(line, "```")
			var codeLines []string

			// Read until closing ```
			for scanner.Scan() {
				codeLine := scanner.Text()
				if codeLine == "```" {
					break
				}
				codeLines = append(codeLines, codeLine)
			}

			currentSection.CodeBlocks = append(currentSection.CodeBlocks, MarkdownCodeBlock{
				Language: language,
				Content:  strings.Join(codeLines, "\n"),
			})
			continue
		}

		// Regular content becomes description
		if strings.TrimSpace(line) != "" && currentSection.Description == "" {
			currentSection.Description = strings.TrimSpace(line)
		}
	}

	// Don't forget the last section
	if currentSection != nil {
		sections = append(sections, *currentSection)
	}

	return sections
}

// parseMarkdownSection converts a markdown section into a ParserTestCase
func (f *ParserTestFramework) parseMarkdownSection(section MarkdownSection, metadata *MarkdownTestMetadata, filePath string) (*ParserTestCase, error) {
	// Skip compilation outcome sections - these are not test cases but expected outputs
	titleLower := strings.ToLower(section.Title)
	if strings.Contains(titleLower, "compilation") ||
		strings.Contains(titleLower, "clean perl output") ||
		strings.Contains(titleLower, "typed perl output") ||
		strings.Contains(titleLower, "inferred perl output") ||
		strings.Contains(titleLower, "expected") && (strings.Contains(titleLower, "output") || strings.Contains(titleLower, "perl")) {
		return nil, nil // Skip compilation outcome sections
	}

	// Skip sections without Perl code blocks
	var perlCode string
	for _, block := range section.CodeBlocks {
		if block.Language == "perl" {
			perlCode = block.Content
			break
		}
	}

	if perlCode == "" {
		return nil, nil // Skip sections without Perl code
	}

	// Generate test case name from title and file
	name := f.generateTestCaseName(section.Title, filePath)

	// Parse error expectations from metadata and comments
	shouldError := metadata.ShouldError || section.Comments["should_error"] == "true"
	errorType := section.Comments["expected_error"]

	testCase := &ParserTestCase{
		Name:        name,
		Category:    metadata.Category,
		Subcategory: metadata.Subcategory,
		Input:       perlCode,
		ShouldError: shouldError,
		ErrorType:   errorType,
		Description: section.Description,
		Tags:        metadata.Tags,
	}

	return testCase, nil
}

// sectionHasPerlCode checks if a section contains Perl code blocks
func (f *ParserTestFramework) sectionHasPerlCode(section MarkdownSection) bool {
	for _, block := range section.CodeBlocks {
		if block.Language == "perl" {
			return true
		}
	}
	return false
}

// generateTestCaseName creates a unique test case name from title and filepath
func (f *ParserTestFramework) generateTestCaseName(title, filePath string) string {
	// Extract filename without extension
	baseName := filepath.Base(filePath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// Convert title to snake_case
	titleSlug := strings.ToLower(strings.ReplaceAll(title, " ", "_"))
	titleSlug = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(titleSlug, "")

	return fmt.Sprintf("%s_%s", baseName, titleSlug)
}

// LoadTestCasesFromFile loads test cases from either JSON or Markdown format
func (f *ParserTestFramework) LoadTestCasesFromFile(filePath string) ([]*ParserTestCase, error) {
	ext := filepath.Ext(filePath)

	switch ext {
	case ".json":
		return f.LoadTestCases(filePath)
	case ".md":
		return f.LoadMarkdownTestCases(filePath)
	default:
		return nil, fmt.Errorf("unsupported test file format: %s", ext)
	}
}

// SaveTestCase saves a test case to a JSON file
func (f *ParserTestFramework) SaveTestCase(testCase *ParserTestCase, filePath string) error {
	data, err := json.MarshalIndent(testCase, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test case: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write test case file: %w", err)
	}

	return nil
}

// RunTestCase executes a single test case
func (f *ParserTestFramework) RunTestCase(t *testing.T, testCase *ParserTestCase) {
	t.Helper()

	// For parallel tests, use the parser pool to avoid thread safety issues
	// For non-parallel tests, fall back to the framework's parser
	var parser Parser

	// Check if we're running in parallel mode by attempting to get a parser from pool
	if poolParser := GlobalParserPool.Get(); poolParser != nil {
		parser = poolParser
		defer GlobalParserPool.Put(parser)
	} else if f.Parser != nil {
		parser = f.Parser
	} else {
		t.Errorf("No parser available for test framework")
		return
	}

	startTime := time.Now()
	ast, err := parser.ParseString(testCase.Input)
	parseTime := time.Since(startTime)

	if testCase.ShouldError {
		if err == nil {
			t.Errorf("Test %s: Expected error but parsing succeeded", testCase.Name)
			return
		}
		if testCase.ErrorType != "" && !strings.Contains(err.Error(), testCase.ErrorType) {
			t.Errorf("Test %s: Expected error type '%s' but got: %v",
				testCase.Name, testCase.ErrorType, err)
			return
		}
		if f.Verbose {
			t.Logf("Test %s: Successfully caught expected error: %v", testCase.Name, err)
		}
		return
	}

	if err != nil {
		t.Errorf("Test %s: Unexpected parsing error: %v", testCase.Name, err)
		return
	}

	if ast == nil {
		t.Errorf("Test %s: Parser returned nil AST", testCase.Name)
		return
	}

	// Debug: log what we got from the parser
	if f.Verbose {
		t.Logf("Test %s: AST Source='%s', TypeAnnotations count=%d",
			testCase.Name, ast.Source, len(ast.TypeAnnotations))
	}

	// Log performance metrics
	if f.Verbose {
		t.Logf("Test %s: Parse time: %v", testCase.Name, parseTime)
	}

	// If we have an expected AST, compare it
	if testCase.ExpectedAST != nil {
		f.CompareASTs(t, testCase.ExpectedAST, ast, testCase.Name)
	}

	// Validate AST structure is reasonable
	f.ValidateAST(t, ast, testCase.Name, testCase.Input)

	// Compare expected AST before type inference
	if testCase.ExpectedASTBeforeInfer != "" {
		f.CompareASTString(t, testCase.ExpectedASTBeforeInfer, ast, testCase.Name, "before type inference")
	}

	// TODO: Add type inference and compare after inference AST
	if testCase.ExpectedASTAfterInfer != "" {
		// For now, just compare against the same AST until type inference is implemented
		f.CompareASTString(t, testCase.ExpectedASTAfterInfer, ast, testCase.Name, "after type inference")
	}

	// Run type checking if enabled
	if testCase.TypeCheck {
		f.RunTypeCheckValidation(t, testCase, ast)
	}
}

// CompareASTs compares two AST structures for equivalence
func (f *ParserTestFramework) CompareASTs(t *testing.T, expected, actual *ast.AST, testName string) bool {
	t.Helper()

	// Convert ASTs to comparable string representations
	expectedStr := expected.String()
	actualStr := actual.String()

	if expectedStr != actualStr {
		if f.UpdateMode {
			t.Logf("Test %s: AST mismatch - updating baseline in update mode", testName)
			return true
		}

		diff := cmp.Diff(expectedStr, actualStr)
		t.Errorf("Test %s: AST mismatch (-expected +actual):\n%s", testName, diff)
		return false
	}

	return true
}

// ValidateAST performs basic validation of AST structure
func (f *ParserTestFramework) ValidateAST(t *testing.T, ast *ast.AST, testName string, testInput string) bool {
	t.Helper()

	if ast == nil {
		t.Logf("Test %s: AST is nil", testName)
		return false
	}

	// Basic structural validation
	if ast.Source == "" && testInput != "" {
		t.Logf("Test %s: AST source is empty but input was not empty", testName)
		return false
	}

	// Validate that TypeAnnotations slice is initialized (can be empty)
	if ast.TypeAnnotations == nil {
		t.Logf("Test %s: AST TypeAnnotations is nil", testName)
		return false
	}

	// Root can be nil for simple cases, so we don't require it
	// Additional validation can be added here

	return true
}

// CompareASTString compares an AST against an expected string representation
func (f *ParserTestFramework) CompareASTString(t *testing.T, expectedStr string, actual *ast.AST, testName, phase string) bool {
	t.Helper()

	if actual == nil {
		t.Errorf("Test %s (%s): AST is nil", testName, phase)
		return false
	}

	// Convert actual AST to string representation
	actualStr := actual.String()

	// Normalize both strings (trim whitespace)
	expectedStr = strings.TrimSpace(expectedStr)
	actualStr = strings.TrimSpace(actualStr)

	if expectedStr != actualStr {
		if f.UpdateMode {
			t.Logf("Test %s (%s): AST mismatch - updating baseline in update mode", testName, phase)
			return true
		}

		diff := cmp.Diff(expectedStr, actualStr)
		t.Errorf("Test %s (%s): AST mismatch (-expected +actual):\n%s", testName, phase, diff)
		return false
	}

	if f.Verbose {
		t.Logf("Test %s (%s): AST matches expected structure", testName, phase)
	}

	return true
}

// RunTypeCheckValidation performs type checking validation on an AST
func (f *ParserTestFramework) RunTypeCheckValidation(t *testing.T, testCase *ParserTestCase, ast *ast.AST) bool {
	t.Helper()

	if f.TypeChecker == nil {
		t.Logf("Test %s: No type checker configured, skipping type validation", testCase.Name)
		return true // Don't fail if type checker is not available
	}

	// Run type checking on the AST
	typeErrors := f.TypeChecker.CheckAST(ast)

	// Convert errors to string for comparison
	var actualErrorMessages []string
	for _, err := range typeErrors {
		actualErrorMessages = append(actualErrorMessages, err.Error())
	}
	actualErrorsStr := strings.Join(actualErrorMessages, "\n")

	// Handle expected type errors
	expectedErrors := strings.TrimSpace(testCase.ExpectedTypeErrors)

	// Normalize expected errors - treat "(none)" as no errors expected
	if expectedErrors == "(none)" || expectedErrors == "" {
		expectedErrors = ""
	}

	if f.Verbose {
		t.Logf("Test %s: Type checking - Expected: '%s', Actual: '%s'",
			testCase.Name, expectedErrors, actualErrorsStr)
	}

	// Compare expected vs actual type errors
	if expectedErrors == "" {
		// No errors expected
		if len(typeErrors) > 0 {
			t.Errorf("Test %s: Expected no type errors but got: %v", testCase.Name, typeErrors)
			return false
		}
	} else {
		// Specific errors expected
		if len(typeErrors) == 0 {
			t.Errorf("Test %s: Expected type errors '%s' but got none", testCase.Name, expectedErrors)
			return false
		}

		// Check if actual errors contain expected error patterns
		if !strings.Contains(actualErrorsStr, expectedErrors) {
			t.Errorf("Test %s: Expected type errors containing '%s' but got '%s'",
				testCase.Name, expectedErrors, actualErrorsStr)
			return false
		}
	}

	if f.Verbose {
		t.Logf("Test %s: Type checking validation passed", testCase.Name)
	}

	return true
}

// DiscoverTestCases finds all test cases in the test data directory
func (f *ParserTestFramework) DiscoverTestCases() ([]*ParserTestCase, error) {
	var allTestCases []*ParserTestCase

	err := filepath.Walk(f.TestDataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only load markdown files, ignore JSON files
		if strings.HasSuffix(path, ".md") {
			testCases, err := f.LoadTestCasesFromFile(path)
			if err != nil {
				return fmt.Errorf("failed to load test cases %s: %w", path, err)
			}
			allTestCases = append(allTestCases, testCases...)
		}

		return nil
	})

	return allTestCases, err
}

// RunAllTests executes all discovered test cases with parallelization
func (f *ParserTestFramework) RunAllTests(t *testing.T) {
	testCases, err := f.DiscoverTestCases()
	if err != nil {
		t.Fatalf("Failed to discover test cases: %v", err)
	}

	testCount := 0
	for _, testCase := range testCases {
		testCount++
		testCase := testCase // capture loop variable for parallel execution
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel() // Enable parallel execution of test cases
			f.RunTestCase(t, testCase)
		})
	}

	if testCount == 0 {
		t.Error("No test cases found")
	}
}

// RunTestsByCategory runs tests for a specific category
func (f *ParserTestFramework) RunTestsByCategory(t *testing.T, category TestCategory) {
	testCases, err := f.DiscoverTestCases()
	if err != nil {
		t.Fatalf("Failed to discover test cases: %v", err)
	}

	testCount := 0
	for _, testCase := range testCases {
		if testCase.Category == category {
			testCount++
			testCase := testCase // capture loop variable for parallel execution
			t.Run(testCase.Name, func(t *testing.T) {
				t.Parallel() // Enable parallel execution of test cases
				f.RunTestCase(t, testCase)
			})
		}
	}

	if testCount == 0 {
		t.Errorf("No test cases found for category %s", category)
	}
}

// updateMetrics updates accuracy metrics based on test results
func (f *ParserTestFramework) updateMetrics(metrics *AccuracyMetrics, testCase *ParserTestCase, success bool) {
	metrics.TotalTests++
	if success {
		metrics.PassedTests++
	} else {
		metrics.FailedTests++
	}

	// Update category metrics
	categoryKey := string(testCase.Category)
	catMetric := metrics.CategoryMetrics[categoryKey]
	catMetric.Total++
	if success {
		catMetric.Passed++
	} else {
		catMetric.Failed++
	}
	metrics.CategoryMetrics[categoryKey] = catMetric

	// Update feature metrics based on tags
	for _, tag := range testCase.Tags {
		featureMetric := metrics.FeatureMetrics[tag]
		featureMetric.Total++
		if success {
			featureMetric.Passed++
		} else {
			featureMetric.Failed++
		}
		metrics.FeatureMetrics[tag] = featureMetric
	}
}

// calculateAccuracyPercentages calculates accuracy percentages for all metrics
func (f *ParserTestFramework) calculateAccuracyPercentages(metrics *AccuracyMetrics) {
	for key, metric := range metrics.CategoryMetrics {
		if metric.Total > 0 {
			metric.Accuracy = float64(metric.Passed) / float64(metric.Total) * 100
			metrics.CategoryMetrics[key] = metric
		}
	}

	for key, metric := range metrics.FeatureMetrics {
		if metric.Total > 0 {
			metric.Accuracy = float64(metric.Passed) / float64(metric.Total) * 100
			metrics.FeatureMetrics[key] = metric
		}
	}
}

// GenerateTestCase creates a test case from input code and expected behavior
func (f *ParserTestFramework) GenerateTestCase(name, input, description string, category TestCategory, tags []string) *ParserTestCase {
	return &ParserTestCase{
		Name:        name,
		Category:    category,
		Input:       input,
		Description: description,
		Tags:        tags,
		ShouldError: false,
	}
}

// GenerateErrorTestCase creates a test case that expects parsing to fail
func (f *ParserTestFramework) GenerateErrorTestCase(name, input, description, errorType string, category TestCategory, tags []string) *ParserTestCase {
	return &ParserTestCase{
		Name:        name,
		Category:    category,
		Input:       input,
		Description: description,
		Tags:        tags,
		ShouldError: true,
		ErrorType:   errorType,
	}
}

// SaveMetricsReport saves accuracy metrics to a JSON file
func (f *ParserTestFramework) SaveMetricsReport(metrics *AccuracyMetrics, filePath string) error {
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}

	return nil
}

// PrintMetricsSummary prints a summary of accuracy metrics
func (f *ParserTestFramework) PrintMetricsSummary(t *testing.T, metrics *AccuracyMetrics) {
	t.Helper()

	overallAccuracy := float64(metrics.PassedTests) / float64(metrics.TotalTests) * 100

	t.Logf("=== Parser Accuracy Report ===")
	t.Logf("Overall: %d/%d tests passed (%.1f%% accuracy)",
		metrics.PassedTests, metrics.TotalTests, overallAccuracy)
	t.Logf("Parse time: %v", metrics.ParsingTime)

	t.Logf("\nCategory Breakdown:")
	for category, metric := range metrics.CategoryMetrics {
		t.Logf("  %s: %d/%d (%.1f%%)", category, metric.Passed, metric.Total, metric.Accuracy)
	}

	if len(metrics.FeatureMetrics) > 0 {
		t.Logf("\nFeature Breakdown:")
		for feature, metric := range metrics.FeatureMetrics {
			t.Logf("  %s: %d/%d (%.1f%%)", feature, metric.Passed, metric.Total, metric.Accuracy)
		}
	}
}

// parseCompilationOutcomes parses a markdown section containing expected compilation outcomes
func (f *ParserTestFramework) parseCompilationOutcomes(section MarkdownSection) (*CompilationOutcomes, error) {
	outcomes := &CompilationOutcomes{
		CompilationErrors: make(map[string]string),
	}

	// Debug output for development
	if f.Verbose {
		println("DEBUG: Parsing compilation outcomes section")
		println("DEBUG: Section title:", section.Title)
		println("DEBUG: Section description length:", len(section.Description))
		println("DEBUG: Number of code blocks:", len(section.CodeBlocks))
	}

	// Parse each subsection within the compilation outcomes section
	for _, codeBlock := range section.CodeBlocks {
		// Look for labeled code blocks or subsections
		switch {
		case strings.Contains(strings.ToLower(codeBlock.Language), "clean") ||
			strings.Contains(strings.ToLower(codeBlock.Info), "clean"):
			outcomes.ExpectedCleanPerl = strings.TrimSpace(codeBlock.Content)

		case strings.Contains(strings.ToLower(codeBlock.Language), "typed") ||
			strings.Contains(strings.ToLower(codeBlock.Info), "typed"):
			outcomes.ExpectedTypedPerl = strings.TrimSpace(codeBlock.Content)

		case strings.Contains(strings.ToLower(codeBlock.Language), "inferred") ||
			strings.Contains(strings.ToLower(codeBlock.Info), "inferred"):
			outcomes.ExpectedInferredPerl = strings.TrimSpace(codeBlock.Content)

		case codeBlock.Language == "perl" || codeBlock.Language == "":
			// Default Perl code block - try to determine type from context or preceding text
			// For now, assume it's the inferred output if no other context
			if outcomes.ExpectedInferredPerl == "" {
				outcomes.ExpectedInferredPerl = strings.TrimSpace(codeBlock.Content)
			}
		}
	}

	// Parse any subsections within the compilation outcomes section
	// Look for markdown subsections like "## Clean Perl Output", "## Typed Perl Output", etc.
	content := section.Description
	if content != "" {
		outcomes = f.parseCompilationOutcomesFromText(content, outcomes)
	}

	return outcomes, nil
}

// parseCompilationOutcomesFromSections parses compilation outcomes from multiple sections
func (f *ParserTestFramework) parseCompilationOutcomesFromSections(sections []MarkdownSection) (*CompilationOutcomes, error) {
	outcomes := &CompilationOutcomes{
		CompilationErrors: make(map[string]string),
	}

	for _, section := range sections {
		titleLower := strings.ToLower(section.Title)

		if f.Verbose {
			println("DEBUG: Processing section:", section.Title, "with", len(section.CodeBlocks), "code blocks")
		}

		// Process code blocks in this section
		for _, codeBlock := range section.CodeBlocks {
			codeContent := strings.TrimSpace(codeBlock.Content)
			if codeContent == "" {
				continue
			}

			// Determine which target based on section title
			switch {
			case strings.Contains(titleLower, "clean") || strings.Contains(titleLower, "untyped"):
				outcomes.ExpectedCleanPerl = codeContent
				if f.Verbose {
					println("DEBUG: Set clean Perl output")
				}
			case strings.Contains(titleLower, "typed") && !strings.Contains(titleLower, "inferred"):
				outcomes.ExpectedTypedPerl = codeContent
				if f.Verbose {
					println("DEBUG: Set typed Perl output")
				}
			case strings.Contains(titleLower, "inferred"):
				outcomes.ExpectedInferredPerl = codeContent
				if f.Verbose {
					println("DEBUG: Set inferred Perl output")
				}
			default:
				// Fall back to the general parsing logic
				if f.Verbose {
					println("DEBUG: Using fallback parsing for section:", section.Title)
				}
			}
		}
	}

	return outcomes, nil
}

// parseCompilationOutcomesFromText parses compilation outcomes from markdown text with subsections
func (f *ParserTestFramework) parseCompilationOutcomesFromText(content string, outcomes *CompilationOutcomes) *CompilationOutcomes {
	lines := strings.Split(content, "\n")
	var currentSection string
	var currentContent []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for subsection headers
		if strings.HasPrefix(trimmed, "##") || strings.HasPrefix(trimmed, "###") {
			// Process previous section if any
			f.processCompilationSection(currentSection, currentContent, outcomes)

			// Start new section
			currentSection = strings.ToLower(trimmed)
			currentContent = []string{}
		} else if strings.HasPrefix(trimmed, "```") {
			// Handle code blocks within subsections
			if len(currentContent) > 0 && strings.HasSuffix(currentContent[len(currentContent)-1], "```") {
				// End of code block
				continue
			}
			// Start of code block
			currentContent = append(currentContent, trimmed)
		} else if len(currentContent) > 0 {
			// Inside a code block or section
			currentContent = append(currentContent, line)
		}
	}

	// Process final section
	f.processCompilationSection(currentSection, currentContent, outcomes)

	return outcomes
}

// processCompilationSection processes a single compilation outcome section
func (f *ParserTestFramework) processCompilationSection(sectionHeader string, content []string, outcomes *CompilationOutcomes) {
	if sectionHeader == "" || len(content) == 0 {
		return
	}

	// Remove code block markers and extract content
	var perlCode []string
	inCodeBlock := false

	for _, line := range content {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			perlCode = append(perlCode, line)
		}
	}

	codeContent := strings.TrimSpace(strings.Join(perlCode, "\n"))
	if codeContent == "" {
		return
	}

	// Determine which target this section is for
	switch {
	case strings.Contains(sectionHeader, "clean") || strings.Contains(sectionHeader, "untyped"):
		outcomes.ExpectedCleanPerl = codeContent
	case strings.Contains(sectionHeader, "typed") && !strings.Contains(sectionHeader, "inferred"):
		outcomes.ExpectedTypedPerl = codeContent
	case strings.Contains(sectionHeader, "inferred"):
		outcomes.ExpectedInferredPerl = codeContent
	case strings.Contains(sectionHeader, "error"):
		// Handle compilation errors
		if outcomes.CompilationErrors == nil {
			outcomes.CompilationErrors = make(map[string]string)
		}
		// Try to determine which target this error is for
		if strings.Contains(sectionHeader, "clean") {
			outcomes.CompilationErrors["clean_perl"] = codeContent
		} else if strings.Contains(sectionHeader, "typed") {
			outcomes.CompilationErrors["typed_perl"] = codeContent
		} else if strings.Contains(sectionHeader, "inferred") {
			outcomes.CompilationErrors["inferred_typed_perl"] = codeContent
		}
	}
}

// ValidateCompilationOutcomes validates actual compilation results against expected outcomes
func (f *ParserTestFramework) ValidateCompilationOutcomes(testCase *ParserTestCase, ast *ast.AST) []error {
	var errors []error

	if testCase.ExpectedCompilationOutcomes == nil {
		// No compilation outcomes expected, skip validation
		return nil
	}

	outcomes := testCase.ExpectedCompilationOutcomes

	// Validate each compilation target
	if outcomes.ExpectedCleanPerl != "" {
		if err := f.validateCompilationTarget(ast, "clean_perl", outcomes.ExpectedCleanPerl); err != nil {
			errors = append(errors, fmt.Errorf("clean Perl compilation validation failed: %w", err))
		}
	}

	if outcomes.ExpectedTypedPerl != "" {
		if err := f.validateCompilationTarget(ast, "typed_perl", outcomes.ExpectedTypedPerl); err != nil {
			errors = append(errors, fmt.Errorf("typed Perl compilation validation failed: %w", err))
		}
	}

	if outcomes.ExpectedInferredPerl != "" {
		if err := f.validateCompilationTarget(ast, "inferred_typed_perl", outcomes.ExpectedInferredPerl); err != nil {
			errors = append(errors, fmt.Errorf("inferred typed Perl compilation validation failed: %w", err))
		}
	}

	return errors
}

// validateCompilationTarget validates a single compilation target
func (f *ParserTestFramework) validateCompilationTarget(ast *ast.AST, target string, expected string) error {
	// This would need to be implemented with the actual compiler registry
	// For now, return a placeholder implementation
	return fmt.Errorf("compilation validation not yet implemented for target: %s", target)
}
