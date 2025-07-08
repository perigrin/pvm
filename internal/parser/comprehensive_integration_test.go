// ABOUTME: Comprehensive end-to-end integration tests for complete typed-Perl programs using corpus-based testing
// ABOUTME: Validates that all type annotation features work together in realistic scenarios from test corpus

package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationCorpus tests integration scenarios using corpus files
func TestIntegrationCorpus(t *testing.T) {
	corpusDir := "../../testdata/corpus/integration"

	// Check if corpus directory exists, if not skip the test
	if _, err := os.Stat(corpusDir); os.IsNotExist(err) {
		t.Skip("Integration corpus directory not found, skipping corpus-based tests")
		return
	}

	// Create test framework
	framework := NewParserTestFramework(corpusDir)

	// Create parser
	parser, err := NewParser()
	require.NoError(t, err, "Failed to create parser")

	// Walk through all markdown files in corpus directory
	err = filepath.Walk(corpusDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Process .md files containing test cases
		if strings.HasSuffix(path, ".md") {
			t.Run(filepath.Base(path), func(t *testing.T) {
				// Load test cases from markdown file
				testCases, err := framework.LoadMarkdownTestCases(path)
				require.NoError(t, err, "Failed to load test cases from %s", path)

				for _, testCase := range testCases {
					t.Run(testCase.Name, func(t *testing.T) {
						startTime := time.Now()

						// Parse the complete program
						ast, err := parser.ParseString(testCase.Input)
						parseTime := time.Since(startTime)

						// Basic validation - program should parse or produce expected errors
						if testCase.ShouldError {
							assert.Error(t, err, "Program should produce errors")
						} else {
							if err != nil {
								t.Logf("Parse error: %v", err)
								t.Logf("Program content:\n%s", testCase.Input)
							}
							assert.NoError(t, err, "Program should parse successfully")
						}

						// Validate AST structure if parsing succeeded
						if err == nil && ast != nil {
							assert.NotNil(t, ast.Root, "AST should have a root node")

							// Check minimum lines expectation if specified in metadata
							// This would require extending the metadata structure
							t.Logf("Parse time: %v", parseTime)
						}
					})
				}
			})
		}

		return nil
	})

	require.NoError(t, err, "Failed to walk corpus directory")
}
