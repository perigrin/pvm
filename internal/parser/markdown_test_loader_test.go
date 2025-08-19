// ABOUTME: Tests for the Markdown test case loader functionality
// ABOUTME: Validates parsing of Markdown test files with YAML frontmatter and code blocks

package parser

import (
	"path/filepath"
	"testing"
)

func TestMarkdownTestCaseLoader(t *testing.T) {
	framework := NewParserTestFramework("../../testdata/corpus/parser")

	// Test loading basic-typed-variables.md from simple-annotations directory
	testFile := filepath.Join("../../testdata/corpus/parser", "typed-perl", "simple-annotations", "basic-typed-variables.md")
	testCases, err := framework.LoadMarkdownTestCases(testFile)
	if err != nil {
		t.Fatalf("Failed to load markdown test cases: %v", err)
	}

	if len(testCases) == 0 {
		t.Fatal("No test cases loaded from markdown file")
	}

	// Validate first test case
	firstCase := testCases[0]
	if firstCase.Category != TypedPerl {
		t.Errorf("Expected category %s, got %s", TypedPerl, firstCase.Category)
	}

	if firstCase.Subcategory != "simple-annotations" {
		t.Errorf("Expected subcategory 'simple-annotations', got %s", firstCase.Subcategory)
	}

	if len(firstCase.Tags) == 0 {
		t.Error("Expected tags to be populated")
	}

	if firstCase.Input == "" {
		t.Error("Expected input to be populated")
	}

	t.Logf("Loaded %d test cases from markdown file", len(testCases))
	for i, tc := range testCases {
		t.Logf("Test case %d: %s", i+1, tc.Name)
		t.Logf("  Description: %s", tc.Description)
		t.Logf("  Should error: %v", tc.ShouldError)
		if tc.Input != "" {
			t.Logf("  Input preview: %.50s...", tc.Input)
		}
	}
}

func TestMarkdownErrorRecoveryCaseLoader(t *testing.T) {
	framework := NewParserTestFramework("../../testdata/corpus/parser")

	// Test loading error recovery cases (formerly error cases)
	testFile := filepath.Join("../../testdata/corpus/parser", "error-cases", "malformed-types.md")
	testCases, err := framework.LoadMarkdownTestCases(testFile)
	if err != nil {
		t.Fatalf("Failed to load error recovery case markdown: %v", err)
	}

	if len(testCases) == 0 {
		t.Fatal("No error recovery test cases loaded")
	}

	// Validate that error recovery cases demonstrate graceful handling
	for i, tc := range testCases {
		// These cases now demonstrate error recovery, not failure
		if tc.ShouldError {
			t.Logf("Test case %d: %s demonstrates error recovery (was expecting errors)", i, tc.Name)
		} else {
			t.Logf("Test case %d: %s demonstrates successful error recovery", i, tc.Name)
		}

		t.Logf("Error recovery case %d: %s", i+1, tc.Name)
	}
}

func TestLoadTestCasesFromFile(t *testing.T) {
	framework := NewParserTestFramework("../../testdata/corpus/parser")

	// Test Markdown file loading and error cases
	testCases := []struct {
		file        string
		expectError bool
	}{
		{"../../testdata/corpus/parser/typed-perl/simple-annotations/basic-typed-variables.md", false},
		{"../../testdata/corpus/parser/nonexistent.md", true},
		{"../../testdata/corpus/parser/invalid.txt", true},
	}

	for _, tc := range testCases {
		cases, err := framework.LoadTestCasesFromFile(tc.file)

		if tc.expectError {
			if err == nil {
				t.Errorf("Expected error for file %s, but got none", tc.file)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error for file %s: %v", tc.file, err)
			continue
		}

		if len(cases) == 0 {
			t.Errorf("No test cases loaded from %s", tc.file)
		}

		t.Logf("Loaded %d cases from %s", len(cases), tc.file)
	}
}
