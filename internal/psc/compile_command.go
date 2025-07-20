// ABOUTME: PSC compile command for converting between different Perl variants
// ABOUTME: Supports compilation to standard Perl, typed Perl, and inferred annotations

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
• standard        - Strip type annotations for standard Perl compatibility
• clean           - Deprecated alias for 'standard'
• typed           - Preserve existing annotations and infer missing types (default for .pl files)

The standard target is perfect for publishing to CPAN or running on systems
without PSC. The typed target preserves existing annotations and infers
missing types through static analysis.

Output options:
• Single file compilation with --output flag
• Directory compilation with automatic target detection
• In-place compilation with --in-place flag (use with caution)

Inference options (for typed target):
• --style: Annotation style (inline, verbose, compact, comments)
• --disable-inference: Skip type inference and only preserve existing annotations

Examples:
  psc compile --target=standard typed.pl            # Strip to stdout
  psc compile --target=standard --output=standard.pl typed.pl
  psc compile --target=typed script.pl              # Add inferred types
  psc compile --target=typed --output=final.pl draft.pl`,
		Args: cobra.ExactArgs(1),
		RunE: runCompileCommand,
	}

	// Core compilation flags
	cmd.Flags().String("target", "typed", "Compilation target: standard, clean (deprecated), typed")
	cmd.Flags().String("output", "", "Output file (default: stdout)")
	cmd.Flags().Bool("in-place", false, "Modify files in-place (dangerous)")
	cmd.Flags().Bool("preserve-comments", true, "Preserve original code comments")
	cmd.Flags().Bool("preserve-formatting", true, "Preserve original code formatting")

	// Inference flags (used with typed target)
	cmd.Flags().String("style", "inline", "Annotation style for inferred types: inline, verbose, compact, comments")
	cmd.Flags().Bool("disable-inference", false, "Skip type inference and only preserve existing annotations")

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
	disableInference, _ := cmd.Flags().GetBool("disable-inference")
	showProgress, _ := cmd.Flags().GetBool("progress")
	verbose, _ := cmd.Flags().GetBool("verbose")
	strict, _ := cmd.Flags().GetBool("strict")

	// Validate inputs
	validTargets := map[string]bool{
		"standard": true, "clean": true, "typed": true,
	}
	if !validTargets[target] {
		return fmt.Errorf("invalid target '%s', must be one of: standard, clean (deprecated), typed", target)
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
	case "standard", "clean":
		if showProgress {
			targetName := "standard"
			if target == "clean" {
				targetName = "clean (deprecated, use 'standard')"
			}
			ui.Info("Compiling to %s Perl (stripping type annotations)...", targetName)
		}
		result, err = registry.CompileWithOptions(ast, compiler.TargetStandardPerl, options)

	case "typed":
		if disableInference {
			if showProgress {
				ui.Info("Compiling to typed Perl (preserving existing annotations only)...")
			}
			result, err = registry.CompileWithOptions(ast, compiler.TargetTypedPerl, options)
		} else {
			if showProgress {
				ui.Info("Compiling to typed Perl (preserving annotations and inferring missing types)...")
				ui.Info("Performing type inference...")
			}

			// For typed target with inference, we need to do type inference first
			inferenceOptions := inference.InferenceOptions{
				EnableFlowAnalysis:        true,
				EnableVariablePropagation: true,
			}
			engine := inference.NewTypeInferenceEngineWithOptions(inferenceOptions)

			inferredAST, inferErr := engine.InferTypes(ast)
			if inferErr != nil {
				return fmt.Errorf("type inference failed: %w", inferErr)
			}

			// Report inference statistics
			if showProgress || verbose {
				allTypeInfo := inferredAST.GetAllTypeInfo()
				totalInferences := len(allTypeInfo)
				ui.Info("Type inference complete: %d types inferred", totalInferences)
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
				AnnotationStyle:    annotationStyle,
				PreserveComments:   preserveComments,
				PreserveFormatting: preserveFormatting,
				VerboseOutput:      verbose,
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
