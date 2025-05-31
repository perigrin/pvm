//go:build ignore

// ABOUTME: Script to convert JSON test files to Markdown format
// ABOUTME: Processes all test data directories and converts individual/array JSON files to Markdown

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type TestCategory string

const (
	UntypedPerl TestCategory = "untyped-perl"
	TypedPerl   TestCategory = "typed-perl"
	ErrorCases  TestCategory = "error-cases"
)

type ParserTestCase struct {
	Name               string       `json:"name"`
	Category           TestCategory `json:"category"`
	Subcategory        string       `json:"subcategory"`
	Input              string       `json:"input"`
	ShouldError        bool         `json:"should_error"`
	ErrorType          string       `json:"error_type,omitempty"`
	Description        string       `json:"description"`
	Tags               []string     `json:"tags"`
	ExpectedError      string       `json:"expected_error,omitempty"`
	ExpectedSuggestion string       `json:"expected_suggestion,omitempty"`
	Context            string       `json:"context,omitempty"`
}

type MarkdownTestMetadata struct {
	Category    TestCategory `yaml:"category"`
	Subcategory string       `yaml:"subcategory"`
	Tags        []string     `yaml:"tags"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <testdata-directory>\n", os.Args[0])
		os.Exit(1)
	}

	testDataDir := os.Args[1]

	fmt.Printf("Converting JSON test files to Markdown in %s\n", testDataDir)

	err := convertTestDirectory(testDataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Conversion completed successfully!")
}

func convertTestDirectory(testDataDir string) error {
	// Group test files by directory (subcategory)
	testsBySubcategory := make(map[string][]*ParserTestCase)

	err := filepath.Walk(testDataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		fmt.Printf("Processing %s\n", path)

		testCases, err := loadJSONTestCases(path)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", path, err)
		}

		for _, testCase := range testCases {
			// Determine subcategory from file path if not set
			if testCase.Subcategory == "" {
				testCase.Subcategory = extractSubcategoryFromPath(path, testDataDir)
			}

			// Determine category from path if not set
			if testCase.Category == "" {
				testCase.Category = extractCategoryFromPath(path, testDataDir)
			}

			key := fmt.Sprintf("%s/%s", testCase.Category, testCase.Subcategory)
			testsBySubcategory[key] = append(testsBySubcategory[key], testCase)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Write markdown files for each subcategory
	for _, testCases := range testsBySubcategory {
		if len(testCases) == 0 {
			continue
		}

		// Use first test case to determine metadata
		firstCase := testCases[0]

		// Create output path - place markdown at the correct level
		// If we're processing a subdirectory, go up to the parent
		parentDir := filepath.Dir(testDataDir)
		categoryDir := filepath.Join(parentDir, string(firstCase.Category))

		// If parent doesn't exist or we're already at testdata level, use current structure
		if _, err := os.Stat(categoryDir); os.IsNotExist(err) {
			categoryDir = filepath.Join(testDataDir, string(firstCase.Category))
		}

		outputPath := filepath.Join(categoryDir, firstCase.Subcategory+".md")

		fmt.Printf("Writing %s with %d test cases\n", outputPath, len(testCases))

		err := writeMarkdownFile(outputPath, testCases)
		if err != nil {
			return fmt.Errorf("failed to write %s: %w", outputPath, err)
		}
	}

	return nil
}

func loadJSONTestCases(filePath string) ([]*ParserTestCase, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Try single test case first
	var testCase ParserTestCase
	err = json.Unmarshal(data, &testCase)
	if err == nil {
		return []*ParserTestCase{&testCase}, nil
	}

	// Try array format
	var testCases []ParserTestCase
	err = json.Unmarshal(data, &testCases)
	if err != nil {
		// Try nested format (like malformed_types.json)
		var nested struct {
			Name        string           `json:"name"`
			Description string           `json:"description"`
			TestCases   []ParserTestCase `json:"test_cases"`
		}
		err = json.Unmarshal(data, &nested)
		if err != nil {
			return nil, fmt.Errorf("failed to parse as single, array, or nested format: %w", err)
		}
		testCases = nested.TestCases
	}

	// Convert to pointer slice
	var result []*ParserTestCase
	for i := range testCases {
		result = append(result, &testCases[i])
	}

	return result, nil
}

func extractCategoryFromPath(filePath, testDataDir string) TestCategory {
	relPath, _ := filepath.Rel(testDataDir, filePath)
	parts := strings.Split(relPath, string(filepath.Separator))

	if len(parts) > 0 {
		switch parts[0] {
		case "typed-perl":
			return TypedPerl
		case "untyped-perl":
			return UntypedPerl
		case "error-cases":
			return ErrorCases
		}
	}

	return TypedPerl // default
}

func extractSubcategoryFromPath(filePath, testDataDir string) string {
	relPath, _ := filepath.Rel(testDataDir, filePath)
	parts := strings.Split(relPath, string(filepath.Separator))

	if len(parts) >= 2 {
		return parts[1]
	}

	// Extract from filename if no subdirectory
	fileName := filepath.Base(filePath)
	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return fileName
}

func writeMarkdownFile(outputPath string, testCases []*ParserTestCase) error {
	if len(testCases) == 0 {
		return fmt.Errorf("no test cases to write")
	}

	// Create directory if needed
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return err
	}

	var sb strings.Builder

	// Write YAML frontmatter
	firstCase := testCases[0]
	metadata := MarkdownTestMetadata{
		Category:    firstCase.Category,
		Subcategory: firstCase.Subcategory,
		Tags:        getAllTags(testCases),
	}

	yamlData, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}

	sb.WriteString("---\n")
	sb.WriteString(string(yamlData))
	sb.WriteString("---\n\n")

	// Sort test cases by name for consistency
	sort.Slice(testCases, func(i, j int) bool {
		return testCases[i].Name < testCases[j].Name
	})

	// Write test cases
	for i, testCase := range testCases {
		// Convert name to title
		title := convertNameToTitle(testCase.Name)

		// Header level based on position
		headerLevel := "##"
		if i == 0 {
			headerLevel = "#"
		}

		sb.WriteString(fmt.Sprintf("%s %s\n\n", headerLevel, title))

		// Description
		if testCase.Description != "" {
			sb.WriteString(testCase.Description)
			sb.WriteString("\n\n")
		}

		// Error case comments
		if testCase.ShouldError {
			sb.WriteString("<!-- should_error: true -->\n")

			if testCase.ErrorType != "" {
				sb.WriteString(fmt.Sprintf("<!-- expected_error: %s -->\n", testCase.ErrorType))
			}
			if testCase.ExpectedError != "" {
				sb.WriteString(fmt.Sprintf("<!-- expected_error: %s -->\n", testCase.ExpectedError))
			}
			if testCase.ExpectedSuggestion != "" {
				sb.WriteString(fmt.Sprintf("<!-- expected_suggestion: %s -->\n", testCase.ExpectedSuggestion))
			}
			if testCase.Context != "" {
				sb.WriteString(fmt.Sprintf("<!-- context: %s -->\n", testCase.Context))
			}

			sb.WriteString("\n")
		}

		// Code block
		sb.WriteString("```perl\n")
		sb.WriteString(testCase.Input)
		sb.WriteString("\n```\n")

		if i < len(testCases)-1 {
			sb.WriteString("\n")
		}
	}

	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

func getAllTags(testCases []*ParserTestCase) []string {
	tagSet := make(map[string]bool)

	for _, testCase := range testCases {
		for _, tag := range testCase.Tags {
			tagSet[tag] = true
		}
	}

	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	sort.Strings(tags)
	return tags
}

func convertNameToTitle(name string) string {
	// Convert snake_case to Title Case
	words := strings.Split(name, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}
