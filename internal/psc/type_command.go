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
					fileErrors, fileAnnotations, err := checkTypeInFile(tc, filePath, verbose, showWarnings, reportMode)
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
				fileErrors, fileAnnotations, err := checkTypeInFile(tc, path, verbose, showWarnings, reportMode)
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

	return cmd
}

// checkTypeInFile checks a file for type errors and returns the count of errors and annotations
func checkTypeInFile(tc *parser.TypeCheck, path string, verbose, showWarnings bool, reportMode string) (int, int, error) {
	if verbose {
		fmt.Printf("Checking file: %s\n", path)
	}

	// Type check the file
	result, err := tc.CheckFile(path)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to check file %s: %v", path, err)
	}

	// Report type annotations if verbose
	if verbose && len(result.TypeAnnotations) > 0 {
		fmt.Printf("Found %d type annotations in %s\n", len(result.TypeAnnotations), path)

		for _, annotation := range result.TypeAnnotations {
			fmt.Printf("  %s:%d:%d: %s has type %s\n",
				filepath.Base(path),
				annotation.Pos.Line,
				annotation.Pos.Column,
				annotation.AnnotatedItem,
				annotation.TypeExpression.String())
		}
	}

	// Report errors based on format
	errorCount := len(result.Errors)
	if errorCount > 0 {
		switch reportMode {
		case "json":
			// Simple JSON error reporting
			fmt.Printf("{\"file\":\"%s\",\"errors\":%d,\"details\":[\n", path, errorCount)
			for i, errInfo := range result.Errors {
				fmt.Printf("  {\"line\":%d,\"column\":%d,\"message\":\"%s\"}",
					errInfo.Line, errInfo.Column, escapeJSON(errInfo.Message))
				if i < errorCount-1 {
					fmt.Println(",")
				} else {
					fmt.Println("")
				}
			}
			fmt.Println("]}")

		default: // text format
			fmt.Printf("Found %d type errors in %s\n", errorCount, path)
			for _, errInfo := range result.Errors {
				fmt.Printf("  %s\n", errInfo.Error())
			}
		}
	} else if verbose {
		fmt.Printf("No type errors found in %s\n", path)
	}

	return errorCount, len(result.TypeAnnotations), nil
}

// escapeJSON escapes quotes and other special characters for JSON strings
func escapeJSON(s string) string {
	// Replace special characters with escape sequences
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}