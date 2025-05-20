// ABOUTME: PSC type checking command implementation
// ABOUTME: Provides type checking for Perl files

package psc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/parser"
)

// newCheckTypeCommand creates an enhanced command to check a file or directory for type errors
func newCheckTypeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [file|dir]",
		Short: "Check a file or directory for type errors",
		Long:  "Analyze Perl code for type errors without executing it",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("expected a file or directory to check")
			}

			// Get the path to check
			path := args[0]
			verbose, _ := cmd.Flags().GetBool("verbose")
			showWarnings, _ := cmd.Flags().GetBool("warnings")
			reportMode, _ := cmd.Flags().GetString("format")

			// Create a type checker
			tc, err := parser.NewTypeCheck()
			if err != nil {
				return fmt.Errorf("failed to create type checker: %v", err)
			}

			// Handle flow-sensitive analysis flags
			flowSensitive, _ := cmd.Flags().GetBool("flow-sensitive")
			noFlowSensitive, _ := cmd.Flags().GetBool("no-flow-sensitive")
			showRefinements, _ := cmd.Flags().GetBool("show-refinements")
			skipFlowChecks, _ := cmd.Flags().GetBool("skip-flow-checks")
			flowPatterns, _ := cmd.Flags().GetStringArray("flow-pattern")

			// If --no-flow-sensitive is specified, it overrides --flow-sensitive
			if noFlowSensitive {
				tc.EnableFlowSensitiveAnalysis = false
			} else {
				tc.EnableFlowSensitiveAnalysis = flowSensitive
			}

			// Configure flow-sensitive analysis options
			if tc.EnableFlowSensitiveAnalysis {
				// Pass additional flow-sensitive options to type checker
				tc.SkipFlowChecks = skipFlowChecks
				tc.FlowPatterns = flowPatterns

				if len(flowPatterns) > 0 && verbose {
					fmt.Printf("Using additional flow-sensitive patterns: %v\n", flowPatterns)
				}

				if skipFlowChecks && verbose {
					fmt.Println("Flow-sensitive type checks will be skipped, but refinements will be performed")
				}

				if showRefinements && verbose {
					fmt.Println("Flow-sensitive refinements will be shown in the output")
				}
			} else if verbose {
				fmt.Println("Flow-sensitive analysis is disabled")
			}

			// Check if it's a directory or a file
			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("failed to stat path: %v", err)
			}

			var totalErrors int
			var totalFiles int
			var totalAnnotations int

			if info.IsDir() {
				// Check all Perl files in the directory
				if verbose {
					fmt.Printf("Checking all Perl files in directory: %s\n", path)
				}

				excludePatterns, _ := cmd.Flags().GetStringArray("exclude")

				err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					// Skip directories
					if info.IsDir() {
						return nil
					}

					// Only check Perl files
					if !strings.HasSuffix(filePath, ".pl") && !strings.HasSuffix(filePath, ".pm") {
						return nil
					}

					// Check for excluded patterns
					for _, pattern := range excludePatterns {
						if match, _ := filepath.Match(pattern, filepath.Base(filePath)); match {
							if verbose {
								fmt.Printf("Skipping excluded file: %s\n", filePath)
							}
							return nil
						}
					}

					// Check the file
					fileErrors, fileAnnotations, err := checkTypeInFile(tc, filePath, verbose, showWarnings, showRefinements, reportMode)
					if err != nil {
						fmt.Printf("Error checking %s: %v\n", filePath, err)
						return nil
					}

					totalErrors += fileErrors
					totalAnnotations += fileAnnotations
					totalFiles++

					return nil
				})

				if err != nil {
					return fmt.Errorf("error walking directory: %v", err)
				}

				if verbose || totalErrors > 0 {
					fmt.Printf("\nSummary: Checked %d files with %d type annotations, found %d errors\n",
						totalFiles, totalAnnotations, totalErrors)
				}

				if totalErrors > 0 {
					return fmt.Errorf("type checking failed with %d errors", totalErrors)
				}

			} else {
				// Check a single file
				fileErrors, fileAnnotations, err := checkTypeInFile(tc, path, verbose, showWarnings, showRefinements, reportMode)
				if err != nil {
					return err
				}

				totalErrors = fileErrors
				totalAnnotations = fileAnnotations

				if totalErrors > 0 {
					return fmt.Errorf("type checking failed with %d errors", totalErrors)
				}
			}

			return nil
		},
	}

	// Add command-specific flags
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolP("warnings", "w", false, "Show warnings as well as errors")
	cmd.Flags().StringP("format", "f", "text", "Report format (text, json)")
	cmd.Flags().StringArrayP("exclude", "e", []string{}, "Patterns to exclude (e.g., 'test_*.pl')")
	cmd.Flags().Bool("flow-sensitive", true, "Enable flow-sensitive analysis (default: true)")
	cmd.Flags().Bool("no-flow-sensitive", false, "Disable flow-sensitive analysis")
	cmd.Flags().Bool("show-refinements", false, "Show type refinements from flow-sensitive analysis")
	cmd.Flags().Bool("skip-flow-checks", false, "Skip flow-sensitive type checks but still perform refinements")
	cmd.Flags().StringArrayP("flow-pattern", "p", []string{}, "Additional flow-sensitive patterns to recognize (e.g., 'isa_check')")

	return cmd
}
