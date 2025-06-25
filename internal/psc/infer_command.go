// ABOUTME: PSC infer command for type inference and annotation generation
// ABOUTME: Generates typed Perl code with inferred type annotations from untyped source

package psc

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/compiler"
	"tamarou.com/pvm/internal/inference"
	"tamarou.com/pvm/internal/parser"
)

// newInferCommand creates a command to infer types and generate annotated Perl code
func newInferCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "infer [file]",
		Short: "Infer types and generate annotated Perl code",
		Long: `Analyze Perl code and generate type-annotated versions based on type inference.

The infer command uses static analysis to determine variable types and generates
clean Perl code with appropriate type annotations. This helps in:

• Adding type annotations to existing untyped code
• Improving code documentation and maintainability
• Catching potential type errors through explicit annotations
• Gradual migration to typed Perl development

Confidence levels:
• High (90%+): Direct inline type annotations
• Medium (70-89%): Inline with confidence comments
• Low (50-69%): Comment-only type hints
• Very Low (<50%): No annotations (optional flag to include)

Output formats:
• inline    - Clean inline type annotations (default)
• verbose   - Detailed annotations with confidence info
• compact   - Minimal, space-efficient annotations
• comments  - Type information in comments only

Examples:
  psc infer script.pl                           # Infer and output to stdout
  psc infer --output=typed.pl script.pl         # Write to file
  psc infer --style=verbose --confidence=0.8 script.pl
  psc infer --include-uncertain script.pl       # Include low-confidence types`,
		Args: cobra.ExactArgs(1),
		RunE: runInferCommand,
	}

	// Command-line flags
	cmd.Flags().String("output", "", "Output file (default: stdout)")
	cmd.Flags().String("style", "inline", "Annotation style: inline, verbose, compact, comments")
	cmd.Flags().Float64("confidence", 0.7, "Minimum confidence threshold (0.0-1.0)")
	cmd.Flags().Bool("include-uncertain", false, "Include low-confidence type annotations")
	cmd.Flags().Bool("preserve-comments", true, "Preserve original code comments")
	cmd.Flags().Bool("preserve-formatting", true, "Preserve original code formatting")
	cmd.Flags().Bool("progress", false, "Show progress during inference")
	cmd.Flags().Bool("verbose", false, "Enable verbose output with confidence details")

	return cmd
}

// runInferCommand executes the type inference command
func runInferCommand(cmd *cobra.Command, args []string) error {
	ui := cli.GetUI(cmd)
	inputFile := args[0]

	// Get command flags
	outputFile, _ := cmd.Flags().GetString("output")
	style, _ := cmd.Flags().GetString("style")
	confidence, _ := cmd.Flags().GetFloat64("confidence")
	includeUncertain, _ := cmd.Flags().GetBool("include-uncertain")
	preserveComments, _ := cmd.Flags().GetBool("preserve-comments")
	preserveFormatting, _ := cmd.Flags().GetBool("preserve-formatting")
	showProgress, _ := cmd.Flags().GetBool("progress")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Validate inputs
	if confidence < 0.0 || confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0, got %f", confidence)
	}

	validStyles := map[string]bool{
		"inline": true, "verbose": true, "compact": true, "comments": true,
	}
	if !validStyles[style] {
		return fmt.Errorf("invalid style '%s', must be one of: inline, verbose, compact, comments", style)
	}

	// Check input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}

	if showProgress {
		ui.Info("Starting type inference for %s", inputFile)
	}

	// Parse the input file
	if showProgress {
		ui.Info("Parsing source code...")
	}

	parser, err := parser.NewParser()
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	ast, err := parser.ParseFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", inputFile, err)
	}

	if showProgress {
		ui.Info("Performing type inference...")
	}

	// Create type inference engine
	inferenceOptions := inference.InferenceOptions{
		EnableFlowAnalysis:        true,
		MinConfidenceThreshold:    confidence,
		EnableVariablePropagation: true,
	}
	engine := inference.NewTypeInferenceEngineWithOptions(inferenceOptions)

	// Perform type inference
	inferredAST, err := engine.InferTypes(ast)
	if err != nil {
		return fmt.Errorf("type inference failed: %w", err)
	}

	// Report inference statistics
	if showProgress || verbose {
		allTypeInfo := inferredAST.GetAllTypeInfo()
		totalInferences := len(allTypeInfo)
		highConfidence := 0
		mediumConfidence := 0
		lowConfidence := 0

		for _, typeInfo := range allTypeInfo {
			switch {
			case typeInfo.Confidence >= 0.9:
				highConfidence++
			case typeInfo.Confidence >= 0.7:
				mediumConfidence++
			default:
				lowConfidence++
			}
		}

		ui.Info("Inference complete: %d types inferred", totalInferences)
		if verbose {
			ui.Info("  High confidence (90%%+): %d", highConfidence)
			ui.Info("  Medium confidence (70-89%%): %d", mediumConfidence)
			ui.Info("  Low confidence (<70%%): %d", lowConfidence)
		}
	}

	if showProgress {
		ui.Info("Generating annotated code...")
	}

	// Set up compiler options
	var annotationStyle compiler.FormattingStyle
	switch style {
	case "inline":
		annotationStyle = compiler.StyleInline
	case "verbose":
		annotationStyle = compiler.StyleVerbose
	case "compact":
		annotationStyle = compiler.StyleCompact
	case "comments":
		annotationStyle = compiler.StyleCommentOnly
	}

	compilerOptions := compiler.InferredCompilerOptions{
		ConfidenceThreshold:   confidence,
		AnnotationStyle:       annotationStyle,
		PreserveComments:      preserveComments,
		PreserveFormatting:    preserveFormatting,
		IncludeUncertainTypes: includeUncertain,
		VerboseOutput:         verbose,
	}

	// Create compiler and generate code
	inferredCompiler := compiler.NewInferredTypedPerlCompilerWithOptions(compilerOptions)

	// Use the compiler to generate annotated code - we need to call a special method for InferredAST
	result, err := inferredCompiler.CompileInferred(inferredAST)
	if err != nil {
		return fmt.Errorf("code generation failed: %w", err)
	}

	// Handle output
	if outputFile != "" {
		// Ensure output directory exists
		if dir := filepath.Dir(outputFile); dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		}

		// Write to file
		if err := os.WriteFile(outputFile, []byte(result), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		if showProgress {
			ui.Success("Annotated code written to %s", outputFile)
		}
	} else {
		// Write to stdout
		fmt.Print(result)
	}

	// Report any inference errors or warnings
	inferenceErrors := engine.GetInferenceErrors()
	if len(inferenceErrors) > 0 && verbose {
		ui.Warning("Type inference encountered %d issues:", len(inferenceErrors))
		for _, inferErr := range inferenceErrors {
			ui.Warning("  %s: %s", inferErr.NodeID, inferErr.Message)
		}
	}

	return nil
}
