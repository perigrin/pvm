// ABOUTME: PSC format command that uses the transformation pipeline for code formatting
// ABOUTME: Demonstrates the new pipeline architecture by providing a code formatter capability

package psc

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/compiler"
	"tamarou.com/pvm/internal/compiler/pipeline"
)

// FormatOptions contains options for the format command
type FormatOptions struct {
	InPlace      bool   // Whether to modify files in place
	Output       string // Output file (if not in place)
	IndentSize   int    // Number of spaces for indentation
	UseTabs      bool   // Whether to use tabs instead of spaces
	PreserveType bool   // Whether to preserve type annotations
	Preset       string // Pipeline preset to use
	Verbose      bool   // Whether to show verbose output
}

// formatOptions holds the global format options
var formatOptions = FormatOptions{}

// newFormatCommand creates the format command
func newFormatCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "format [flags] <file>",
		Short: "Format Perl code using transformation pipelines",
		Long: `Format Perl code with consistent indentation and whitespace.

The format command uses the new transformation pipeline architecture to provide
composable code formatting. You can choose different presets or customize the
formatting behavior.

Examples:
  psc format script.pl                    # Format and print to stdout
  psc format -i script.pl                 # Format in place
  psc format -o clean.pl script.pl        # Format to specific output file
  psc format --preset=formatter script.pl # Use specific formatting preset
  psc format --tabs script.pl             # Use tabs for indentation
  psc format --preserve-types script.pl   # Keep type annotations while formatting`,
		Args: cobra.ExactArgs(1),
		RunE: runFormatCommand,
	}

	// Add flags
	cmd.Flags().BoolVarP(&formatOptions.InPlace, "in-place", "i", false, "modify files in place")
	cmd.Flags().StringVarP(&formatOptions.Output, "output", "o", "", "output file (default: stdout)")
	cmd.Flags().IntVar(&formatOptions.IndentSize, "indent", 4, "number of spaces for indentation")
	cmd.Flags().BoolVar(&formatOptions.UseTabs, "tabs", false, "use tabs instead of spaces")
	cmd.Flags().BoolVar(&formatOptions.PreserveType, "preserve-types", false, "preserve type annotations")
	cmd.Flags().StringVar(&formatOptions.Preset, "preset", "", "formatting preset (formatter, typed_formatter, minimal)")
	cmd.Flags().BoolVarP(&formatOptions.Verbose, "verbose", "v", false, "show verbose output")

	return cmd
}

// runFormatCommand executes the format command
func runFormatCommand(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Read the input file
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", inputFile, err)
	}

	// Create the appropriate pipeline
	formatterPipeline, err := createFormatterPipeline()
	if err != nil {
		return fmt.Errorf("failed to create formatter pipeline: %w", err)
	}

	// Create a pipeline compiler
	pipelineCompiler := compiler.NewPipelineCompiler("formatter", formatterPipeline)

	// Create CST-based AST
	cstAST, err := compiler.NewCSTBasedAST(inputFile, string(content))
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", inputFile, err)
	}

	// Format the code
	if formatOptions.Verbose {
		// Get detailed transformation steps
		result, err := pipelineCompiler.GetTransformationSteps(cstAST)
		if err != nil {
			return fmt.Errorf("failed to format file %s: %w", inputFile, err)
		}

		// Show transformation details
		showTransformationDetails(result)

		// Use the formatted content
		err = writeOutput(result.Content, inputFile)
		if err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		// Simple compilation
		formattedCode, err := pipelineCompiler.Compile(cstAST)
		if err != nil {
			return fmt.Errorf("failed to format file %s: %w", inputFile, err)
		}

		err = writeOutput(formattedCode, inputFile)
		if err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	return nil
}

// createFormatterPipeline creates the appropriate formatter pipeline based on options
func createFormatterPipeline() (pipeline.TransformationPipeline, error) {
	// If a preset is specified, use it
	if formatOptions.Preset != "" {
		presets := pipeline.GetAllPresets()
		for _, preset := range presets {
			if preset.Name == formatOptions.Preset {
				return preset.Pipeline, nil
			}
		}
		return nil, fmt.Errorf("unknown preset: %s", formatOptions.Preset)
	}

	// Build custom pipeline based on options
	builder := pipeline.NewPipelineBuilder()

	// Add type handling
	if formatOptions.PreserveType {
		builder = builder.WithTypePreservation()
	}

	// Always add whitespace normalization
	builder = builder.WithWhitespaceNormalization()

	// Add indentation normalization
	if formatOptions.UseTabs {
		builder = builder.WithTabIndentation()
	} else {
		builder = builder.WithSpaceIndentation(formatOptions.IndentSize)
	}

	// Set pipeline options
	options := pipeline.DefaultPipelineOptions()
	options.EnableOptimizations = true
	options.Debug = formatOptions.Verbose
	builder = builder.WithOptions(options)

	return builder.Build(), nil
}

// writeOutput writes the formatted code to the appropriate destination
func writeOutput(formattedCode, inputFile string) error {
	var writer io.Writer

	if formatOptions.InPlace {
		// Write back to the same file
		file, err := os.Create(inputFile)
		if err != nil {
			return fmt.Errorf("failed to open file for writing: %w", err)
		}
		defer file.Close()
		writer = file
	} else if formatOptions.Output != "" {
		// Write to specified output file
		// Create directory if needed
		dir := filepath.Dir(formatOptions.Output)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		file, err := os.Create(formatOptions.Output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = file
	} else {
		// Write to stdout
		writer = os.Stdout
	}

	_, err := writer.Write([]byte(formattedCode))
	return err
}

// showTransformationDetails displays detailed information about the transformation pipeline
func showTransformationDetails(result *pipeline.TransformationResult) {
	fmt.Fprintf(os.Stderr, "Transformation Pipeline Results:\n")
	fmt.Fprintf(os.Stderr, "=====================================\n")
	fmt.Fprintf(os.Stderr, "Total Duration: %.2fms\n", float64(result.Metrics.TotalDuration)/1000000.0)
	fmt.Fprintf(os.Stderr, "Total Transformers: %d\n", len(result.Transformations))
	fmt.Fprintf(os.Stderr, "Skipped Transformers: %d\n", result.Metrics.SkippedTransformers)
	fmt.Fprintf(os.Stderr, "Nodes Processed: %d\n", result.Metrics.TotalMetrics.NodesProcessed)
	fmt.Fprintf(os.Stderr, "Bytes Processed: %d\n", result.Metrics.TotalMetrics.BytesProcessed)
	fmt.Fprintf(os.Stderr, "\n")

	fmt.Fprintf(os.Stderr, "Transformation Steps:\n")
	fmt.Fprintf(os.Stderr, "---------------------\n")
	for i, step := range result.Transformations {
		status := "SKIPPED"
		if step.Modified {
			status = "MODIFIED"
		} else if step.Duration > 0 {
			status = "NO CHANGE"
		}

		fmt.Fprintf(os.Stderr, "%d. %-25s [%s] %.2fms\n",
			i+1, step.Name, status, float64(step.Duration)/1000000.0)

		if formatOptions.Verbose && step.Description != "" {
			// Wrap description text
			wrapped := wrapText(step.Description, 50)
			for _, line := range wrapped {
				fmt.Fprintf(os.Stderr, "   %s\n", line)
			}
		}
	}
	fmt.Fprintf(os.Stderr, "\n")
}

// wrapText wraps text to specified width
func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 <= width {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
