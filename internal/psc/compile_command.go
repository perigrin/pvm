// ABOUTME: PSC compile command for converting between different Perl variants
// ABOUTME: Supports compilation to clean Perl, typed Perl, and inferred annotations

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

// newCompileCommand creates a command to compile Perl code to different targets
func newCompileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compile [file]",
		Short: "Compile Perl code to different target formats",
		Long: `Compile Perl code to different target formats for various use cases.

The compile command supports multiple compilation targets:

Compilation targets:
• clean           - Strip type annotations for standard Perl compatibility
• typed           - Preserve all type annotations (default for .pl files)
• inferred        - Add inferred type annotations to untyped code

The clean target is perfect for publishing to CPAN or running on systems
without PSC. The typed target preserves existing annotations. The inferred
target adds type annotations based on static analysis.

Output options:
• Single file compilation with --output flag
• Directory compilation with automatic target detection
• In-place compilation with --in-place flag (use with caution)

Confidence and style options (for inferred target):
• --confidence: Minimum confidence threshold for annotations
• --style: Annotation style (inline, verbose, compact, comments)
• --include-uncertain: Include low-confidence type hints

Examples:
  psc compile --target=clean typed.pl               # Strip to stdout
  psc compile --target=clean --output=clean.pl typed.pl
  psc compile --target=inferred --confidence=0.8 script.pl
  psc compile --target=typed --output=final.pl draft.pl`,
		Args: cobra.ExactArgs(1),
		RunE: runCompileCommand,
	}

	// Core compilation flags
	cmd.Flags().String("target", "typed", "Compilation target: clean, typed, inferred")
	cmd.Flags().String("output", "", "Output file (default: stdout)")
	cmd.Flags().Bool("in-place", false, "Modify files in-place (dangerous)")
	cmd.Flags().Bool("preserve-comments", true, "Preserve original code comments")
	cmd.Flags().Bool("preserve-formatting", true, "Preserve original code formatting")

	// Inference-specific flags (only used with inferred target)
	cmd.Flags().String("style", "inline", "Annotation style for inferred target: inline, verbose, compact, comments")
	cmd.Flags().Float64("confidence", 0.7, "Minimum confidence threshold for inferred annotations (0.0-1.0)")
	cmd.Flags().Bool("include-uncertain", false, "Include low-confidence annotations in inferred target")

	// Progress and debugging flags
	cmd.Flags().Bool("progress", false, "Show compilation progress")
	cmd.Flags().Bool("verbose", false, "Enable verbose output with detailed information")
	cmd.Flags().Bool("strict", false, "Enable strict compilation mode with enhanced validation")

	return cmd
}

// runCompileCommand executes the compilation command
func runCompileCommand(cmd *cobra.Command, args []string) error {
	ui := cli.GetUI(cmd)
	inputFile := args[0]

	// Get command flags
	target, _ := cmd.Flags().GetString("target")
	outputFile, _ := cmd.Flags().GetString("output")
	inPlace, _ := cmd.Flags().GetBool("in-place")
	preserveComments, _ := cmd.Flags().GetBool("preserve-comments")
	preserveFormatting, _ := cmd.Flags().GetBool("preserve-formatting")
	style, _ := cmd.Flags().GetString("style")
	confidence, _ := cmd.Flags().GetFloat64("confidence")
	includeUncertain, _ := cmd.Flags().GetBool("include-uncertain")
	showProgress, _ := cmd.Flags().GetBool("progress")
	verbose, _ := cmd.Flags().GetBool("verbose")
	strict, _ := cmd.Flags().GetBool("strict")

	// Validate inputs
	validTargets := map[string]bool{
		"clean": true, "typed": true, "inferred": true,
	}
	if !validTargets[target] {
		return fmt.Errorf("invalid target '%s', must be one of: clean, typed, inferred", target)
	}

	if confidence < 0.0 || confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0, got %f", confidence)
	}

	validStyles := map[string]bool{
		"inline": true, "verbose": true, "compact": true, "comments": true,
	}
	if !validStyles[style] {
		return fmt.Errorf("invalid style '%s', must be one of: inline, verbose, compact, comments", style)
	}

	if inPlace && outputFile != "" {
		return fmt.Errorf("cannot use both --in-place and --output flags together")
	}

	// Check input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}

	if showProgress {
		ui.Info("Starting compilation of %s (target: %s)", inputFile, target)
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

	// Create compiler registry and set up options
	registry := compiler.NewCompilerRegistry()
	options := &compiler.CompilerOptions{
		PreserveComments:   preserveComments,
		PreserveFormatting: preserveFormatting,
		StrictMode:         strict,
	}

	var result string

	// Handle different compilation targets
	switch target {
	case "clean":
		if showProgress {
			ui.Info("Compiling to clean Perl (stripping type annotations)...")
		}
		result, err = registry.CompileWithOptions(ast, compiler.TargetCleanPerl, options)

	case "typed":
		if showProgress {
			ui.Info("Compiling to typed Perl (preserving annotations)...")
		}
		result, err = registry.CompileWithOptions(ast, compiler.TargetTypedPerl, options)

	case "inferred":
		if showProgress {
			ui.Info("Performing type inference...")
		}

		// For inferred target, we need to do type inference first
		inferenceOptions := inference.InferenceOptions{
			EnableFlowAnalysis:        true,
			MinConfidenceThreshold:    confidence,
			EnableVariablePropagation: true,
		}
		engine := inference.NewTypeInferenceEngineWithOptions(inferenceOptions)

		inferredAST, err := engine.InferTypes(ast)
		if err != nil {
			return fmt.Errorf("type inference failed: %w", err)
		}

		// Report inference statistics
		if showProgress || verbose {
			allTypeInfo := inferredAST.GetAllTypeInfo()
			totalInferences := len(allTypeInfo)
			ui.Info("Type inference complete: %d types inferred", totalInferences)

			if verbose {
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

				ui.Info("  High confidence (90%%+): %d", highConfidence)
				ui.Info("  Medium confidence (70-89%%): %d", mediumConfidence)
				ui.Info("  Low confidence (<70%%): %d", lowConfidence)
			}
		}

		if showProgress {
			ui.Info("Generating annotated code...")
		}

		// Set up inferred compiler options
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

		inferredCompiler := compiler.NewInferredTypedPerlCompilerWithOptions(compilerOptions)
		result, err = inferredCompiler.CompileInferred(inferredAST)

		// Report inference errors if verbose
		if verbose {
			inferenceErrors := engine.GetInferenceErrors()
			if len(inferenceErrors) > 0 {
				ui.Warning("Type inference encountered %d issues:", len(inferenceErrors))
				for _, inferErr := range inferenceErrors {
					ui.Warning("  %s: %s", inferErr.NodeID, inferErr.Message)
				}
			}
		}
	}

	if err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	// Determine output destination
	var finalOutputFile string
	if inPlace {
		finalOutputFile = inputFile
	} else if outputFile != "" {
		finalOutputFile = outputFile
	}

	// Handle output
	if finalOutputFile != "" {
		// Ensure output directory exists
		if dir := filepath.Dir(finalOutputFile); dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		}

		// Write to file
		if err := os.WriteFile(finalOutputFile, []byte(result), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		if showProgress {
			if inPlace {
				ui.Success("File updated in-place: %s", finalOutputFile)
			} else {
				ui.Success("Compiled code written to %s", finalOutputFile)
			}
		}
	} else {
		// Write to stdout
		fmt.Fprint(cmd.OutOrStdout(), result)
	}

	if showProgress {
		ui.Success("Compilation complete (target: %s)", target)
	}

	return nil
}
