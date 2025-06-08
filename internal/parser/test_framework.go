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
	"gopkg.in/yaml.v3"
	"tamarou.com/pvm/internal/ast"
)

// TestCategory represents different categories of parser tests
type TestCategory string

const (
	UntypedPerl TestCategory = "untyped-perl"
	TypedPerl   TestCategory = "typed-perl"
	ErrorCases  TestCategory = "error-cases"
)

// ParserTestCase represents a single test case for parser validation
type ParserTestCase struct {
	Name        string       `json:"name"`
	Category    TestCategory `json:"category"`
	Subcategory string       `json:"subcategory"`
	Input       string       `json:"input"`
	ExpectedAST *ast.AST     `json:"expected_ast,omitempty"`
	ShouldError bool         `json:"should_error"`
	ErrorType   string       `json:"error_type,omitempty"`
	Description string       `json:"description"`
	Tags        []string     `json:"tags"`
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
	Parser      interface {
		ParseString(string) (*ast.AST, error)
		ParseFile(string) (*ast.AST, error)
	}
}

// NewParserTestFramework creates a new parser testing framework
func NewParserTestFramework(testDataDir string) *ParserTestFramework {
	// Initialize parser - use default parser if creation fails
	parser, err := NewParser()
	if err != nil {
		// Log error but don't fail - tests will handle missing parser
		fmt.Fprintf(os.Stderr, "Warning: Failed to create parser for test framework: %v\n", err)
	}

	return &ParserTestFramework{
		TestDataDir: testDataDir,
		UpdateMode:  os.Getenv("UPDATE_BASELINES") == "1",
		Verbose:     os.Getenv("VERBOSE_TESTS") == "1",
		Parser:      parser,
	}
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

	for _, section := range sections {
		testCase, err := f.parseMarkdownSection(section, metadata, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse section: %w", err)
		}
		if testCase != nil {
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
	// Skip sections without Perl code blocks
	var perlCode string
	for _, block := range section.CodeBlocks {
		if block.Language == "perl" || block.Language == "" {
			perlCode = block.Content
			break
		}
	}

	if perlCode == "" {
		return nil, nil // Skip sections without Perl code
	}

	// Generate test case name from title and file
	name := f.generateTestCaseName(section.Title, filePath)

	// Parse error expectations from comments
	shouldError := section.Comments["should_error"] == "true"
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

// RunTestCase executes a single test case and returns the result
func (f *ParserTestFramework) RunTestCase(t *testing.T, testCase *ParserTestCase) bool {
	t.Helper()

	if f.Parser == nil {
		t.Errorf("No parser configured for test framework")
		return false
	}

	startTime := time.Now()
	ast, err := f.Parser.ParseString(testCase.Input)
	parseTime := time.Since(startTime)

	if testCase.ShouldError {
		if err == nil {
			t.Errorf("Test %s: Expected error but parsing succeeded", testCase.Name)
			return false
		}
		if testCase.ErrorType != "" && !strings.Contains(err.Error(), testCase.ErrorType) {
			t.Errorf("Test %s: Expected error type '%s' but got: %v",
				testCase.Name, testCase.ErrorType, err)
			return false
		}
		if f.Verbose {
			t.Logf("Test %s: Successfully caught expected error: %v", testCase.Name, err)
		}
		return true
	}

	if err != nil {
		t.Errorf("Test %s: Unexpected parsing error: %v", testCase.Name, err)
		return false
	}

	if ast == nil {
		t.Errorf("Test %s: Parser returned nil AST", testCase.Name)
		return false
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
		if !f.CompareASTs(t, testCase.ExpectedAST, ast, testCase.Name) {
			return false
		}
	}

	// Validate AST structure is reasonable
	if !f.ValidateAST(t, ast, testCase.Name, testCase.Input) {
		return false
	}

	return true
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

// DiscoverTestCases finds all test cases in the test data directory
func (f *ParserTestFramework) DiscoverTestCases() ([]*ParserTestCase, error) {
	var allTestCases []*ParserTestCase

	err := filepath.Walk(f.TestDataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".md") {
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

// RunAllTests executes all discovered test cases and returns accuracy metrics
func (f *ParserTestFramework) RunAllTests(t *testing.T) *AccuracyMetrics {
	testCases, err := f.DiscoverTestCases()
	if err != nil {
		t.Fatalf("Failed to discover test cases: %v", err)
	}

	metrics := &AccuracyMetrics{
		CategoryMetrics: make(map[string]Metric),
		FeatureMetrics:  make(map[string]Metric),
	}

	startTime := time.Now()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			success := f.RunTestCase(t, testCase)
			f.updateMetrics(metrics, testCase, success)
		})
	}

	metrics.ParsingTime = time.Since(startTime)
	f.calculateAccuracyPercentages(metrics)

	return metrics
}

// RunTestsByCategory runs tests for a specific category
func (f *ParserTestFramework) RunTestsByCategory(t *testing.T, category TestCategory) *AccuracyMetrics {
	testCases, err := f.DiscoverTestCases()
	if err != nil {
		t.Fatalf("Failed to discover test cases: %v", err)
	}

	metrics := &AccuracyMetrics{
		CategoryMetrics: make(map[string]Metric),
		FeatureMetrics:  make(map[string]Metric),
	}

	for _, testCase := range testCases {
		if testCase.Category == category {
			t.Run(testCase.Name, func(t *testing.T) {
				success := f.RunTestCase(t, testCase)
				f.updateMetrics(metrics, testCase, success)
			})
		}
	}

	f.calculateAccuracyPercentages(metrics)
	return metrics
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
