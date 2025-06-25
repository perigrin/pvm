// ABOUTME: PSC parse command for generating test cases and AST inspection
// ABOUTME: Provides parsing with multiple output formats including test generation

package psc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/parser"
)

// newParseCommand creates a command to parse Perl files with various output formats
func newParseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse [file]",
		Short: "Parse Perl files and output in various formats",
		Long: `Parse Perl files and output the results in different formats.

This command is useful for:
• Generating test cases for the parser test suite
• Inspecting AST structures for debugging
• Creating baseline expectations for regression testing
• Converting existing Perl code to test format

Output formats:
• test     - Generate markdown test case with AST baselines
• ast      - Pretty-printed AST structure
• json     - JSON representation of the AST
• summary  - Brief parsing summary

Examples:
  psc parse --format=test script.pl      # Generate test case
  psc parse --format=ast script.pl       # Show AST structure
  psc parse --format=json script.pl      # JSON output
  psc parse script.pl                     # Default summary format`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)
			filePath := args[0]

			format, _ := cmd.Flags().GetString("format")
			output, _ := cmd.Flags().GetString("output")

			return runParseCommand(ui, filePath, format, output)
		},
	}

	// Add flags
	cmd.Flags().StringP("format", "f", "summary", "Output format (test, ast, json, summary)")
	cmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")

	return cmd
}

// runParseCommand executes the parse command with the specified format
func runParseCommand(ui *ui.Output, filePath, format, outputPath string) error {
	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Parse the file
	parser, err := parser.NewParser()
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	ast, err := parser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Generate output based on format
	var content string
	switch format {
	case "test":
		content, err = generateTestMarkdown(filePath, ast)
	case "ast":
		content = generateASTOutput(ast)
	case "json":
		content, err = generateJSONOutput(ast)
	case "summary":
		content = generateSummaryOutput(ast)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to generate %s output: %w", format, err)
	}

	// Output to file or stdout
	if outputPath != "" {
		err = os.WriteFile(outputPath, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		ui.Success("Output written to %s", outputPath)
	} else {
		fmt.Print(content)
	}

	return nil
}

// generateTestMarkdown creates a markdown test case from the parsed file
func generateTestMarkdown(filePath string, ast *ast.AST) (string, error) {
	// Read the original source code
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %w", err)
	}

	// Extract base filename for the test case
	baseName := filepath.Base(filePath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// Convert to title case
	title := strings.ReplaceAll(baseName, "_", " ")
	title = strings.ReplaceAll(title, "-", " ")
	title = strings.Title(title)

	// Determine category based on code analysis
	category := "untyped-perl"
	if len(ast.TypeAnnotations) > 0 {
		category = "typed-perl"
	}

	// Generate tags based on AST content
	tags := generateTags(ast)

	// Get AST representations
	astBeforeInfer := ast.String()
	astAfterInfer := astBeforeInfer // TODO: Add type inference when available

	// Determine if type checking should be enabled
	typeCheck := len(ast.TypeAnnotations) > 0

	// Build the markdown content
	var content strings.Builder

	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("category: %s\n", category))
	content.WriteString("subcategory: generated\n")
	content.WriteString("tags:\n")
	for _, tag := range tags {
		content.WriteString(fmt.Sprintf("    - %s\n", tag))
	}
	if typeCheck {
		content.WriteString("type_check: true\n")
	}
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# %s\n\n", title))
	content.WriteString(fmt.Sprintf("Generated test case from %s\n\n", filePath))

	content.WriteString("```perl\n")
	content.WriteString(string(sourceCode))
	content.WriteString("\n```\n\n")

	// Add Expected AST section
	content.WriteString("# Expected AST\n\n")
	content.WriteString("## Before Type Inference\n\n")
	content.WriteString("```\n")
	content.WriteString(astBeforeInfer)
	content.WriteString("\n```\n\n")

	content.WriteString("## After Type Inference\n\n")
	content.WriteString("```\n")
	content.WriteString(astAfterInfer)
	content.WriteString("\n```\n\n")

	// Add Expected Type Errors section if type checking is enabled
	if typeCheck {
		content.WriteString("# Expected Type Errors\n\n")
		content.WriteString("```\n")
		content.WriteString("(none)")
		content.WriteString("\n```\n")
	}

	return content.String(), nil
}

// generateASTOutput creates a pretty-printed AST representation
func generateASTOutput(ast *ast.AST) string {
	if ast == nil {
		return "AST is nil\n"
	}
	return ast.String()
}

// generateJSONOutput creates a JSON representation of the AST
func generateJSONOutput(ast *ast.AST) (string, error) {
	// TODO: Implement JSON serialization of AST
	return "JSON output not yet implemented", nil
}

// generateSummaryOutput creates a brief summary of the parse results
func generateSummaryOutput(ast *ast.AST) string {
	if ast == nil {
		return "Parse failed - AST is nil\n"
	}

	var summary strings.Builder
	summary.WriteString("Parse Summary:\n")
	summary.WriteString(fmt.Sprintf("  Source length: %d characters\n", len(ast.Source)))
	summary.WriteString(fmt.Sprintf("  Type annotations: %d\n", len(ast.TypeAnnotations)))
	summary.WriteString(fmt.Sprintf("  Parse errors: %d\n", len(ast.Errors)))

	if len(ast.Errors) > 0 {
		summary.WriteString("\nErrors:\n")
		for i, err := range ast.Errors {
			summary.WriteString(fmt.Sprintf("  %d: %s\n", i+1, err.Error()))
		}
	}

	return summary.String()
}

// generateTags creates appropriate tags based on AST analysis
func generateTags(ast *ast.AST) []string {
	var tags []string

	if len(ast.TypeAnnotations) > 0 {
		tags = append(tags, "typed-variables")
	}

	// Add more tag analysis based on AST content
	// This could be enhanced to detect specific language features

	if len(tags) == 0 {
		tags = append(tags, "basic")
	}

	return tags
}
