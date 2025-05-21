// ABOUTME: PSC-specific commands and functionality
// ABOUTME: Implements commands for Perl type checking

package psc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/parser"
)

// NewCommand creates a new PSC command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "psc",
		Short: "Perl Script Compiler",
		Long:  "Provides static type checking for Perl code",
	}

	// Add PSC-specific commands
	cmd.AddCommand(
		newCheckTypeCommand(), // Use the enhanced type checking command
		newStripCommand(),
		newRunCommand(),
		newWatchCommand(),
		newDefCommand(),
		// Add new type definition commands for PSC-PVI integration
		newGenerateTypeCommand(),
		newImportTypeCommand(),
		newListTypesCommand(),
	)

	return cmd
}

// Legacy command - kept for backwards compatibility but delegates to the new implementation
func newCheckCommand() *cobra.Command {
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

			if info.IsDir() {
				// Check all Perl files in the directory
				if verbose {
					fmt.Printf("Checking all Perl files in directory: %s\n", path)
				}

				return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
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

					// Check the file
					return checkSingleFile(tc, filePath, verbose)
				})
			} else {
				// Check a single file
				return checkSingleFile(tc, path, verbose)
			}
		},
	}

	// Add command-specific flags
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	return cmd
}

// checkSingleFile checks a single Perl file for type errors
func checkSingleFile(tc *parser.TypeCheck, path string, verbose bool) error {
	if verbose {
		fmt.Printf("Checking file: %s\n", path)
	}

	// Type check the file
	result, err := tc.CheckFile(path)
	if err != nil {
		return fmt.Errorf("failed to check file %s: %v", path, err)
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

	// Report errors
	if len(result.Errors) > 0 {
		fmt.Printf("Found %d type errors in %s\n", len(result.Errors), path)

		for _, errInfo := range result.Errors {
			fmt.Printf("  %s\n", errInfo.Error())
		}

		return fmt.Errorf("type checking failed")
	}

	if verbose {
		fmt.Printf("No type errors found in %s\n", path)
	}

	return nil
}

// newCheckTypeCommand is defined in a separate file to avoid conflicts

// performCheck checks a file or directory for type errors
func performCheck(tc *parser.TypeCheck, path string, verbose, showWarnings, showRefinements bool,
	reportMode string, excludePatterns []string, clearScreen bool) error {

	// Clear the screen if requested
	if clearScreen {
		// In a real implementation, we would use the appropriate ANSI escape sequence
		// or platform-specific command to clear the screen
		fmt.Println("\n\n--- New Check Results ---")
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

	} else {
		// Check a single file
		fileErrors, fileAnnotations, err := checkTypeInFile(tc, path, verbose, showWarnings, showRefinements, reportMode)
		if err != nil {
			return err
		}

		totalErrors = fileErrors
		totalAnnotations = fileAnnotations
	}

	// Return success if no errors, otherwise return an error with the count
	if totalErrors > 0 {
		return fmt.Errorf("type checking found %d errors", totalErrors)
	}

	return nil
}

// checkTypeInFile checks a file for type errors and returns the count of errors and annotations
func checkTypeInFile(tc *parser.TypeCheck, path string, verbose, showWarnings, showRefinements bool, reportMode string) (int, int, error) {
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

	// Report refined types if requested and available
	if (verbose || showRefinements) && result.FlowSensitiveEnabled && len(result.RefinedTypes) > 0 {
		fmt.Printf("Flow-sensitive analysis refined %d types in %s\n", len(result.RefinedTypes), path)

		for varName, refinedType := range result.RefinedTypes {
			fmt.Printf("  %s refined to %s\n", varName, refinedType)
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

// The newWatchCommand implementation is in watch_command.go
// The newDefCommand implementation is in def_command.go
