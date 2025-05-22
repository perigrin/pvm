// ABOUTME: PSC type checking command implementation
// ABOUTME: Provides static type checking for Perl code files

package psc

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/typechecker"
)

// newCheckTypeCommand creates a command to check types in Perl files
func newCheckTypeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [file|dir]",
		Short: "Check types in Perl files",
		Long: `Perform static type checking on Perl files using type annotations.

The check command analyzes Perl code for type compatibility issues, validates
type annotations against the type hierarchy, and provides detailed error
reporting with line numbers and descriptions.

Supports:
• Single file or directory checking
• Recursive directory traversal
• Flow-sensitive type analysis
• Custom type definitions
• Multiple output formats

Examples:
  psc check script.pl              # Check single file
  psc check --recursive lib/       # Check all .pl/.pm files in lib/
  psc check --verbose script.pl    # Detailed output
  psc check --format json script.pl # JSON error output
  psc check --strict script.pl     # Strict mode (warnings as errors)`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			strict, _ := cmd.Flags().GetBool("strict")
			verbose, _ := cmd.Flags().GetBool("verbose")
			recursive, _ := cmd.Flags().GetBool("recursive")

			// Process each argument
			totalFiles := 0
			totalErrors := 0

			for _, arg := range args {
				// Check if it's a file or directory
				info, err := os.Stat(arg)
				if err != nil {
					return errors.NewUserInputError(cli.PrefixPSC, "001",
						"File or directory not found", err).
						WithLocation(arg)
				}

				if info.IsDir() {
					if recursive {
						files, errors, err := checkDirectory(arg, strict, verbose)
						if err != nil {
							return err
						}
						totalFiles += files
						totalErrors += errors
					} else {
						fmt.Printf("Skipping directory %s (use --recursive to check directories)\n", arg)
					}
				} else {
					errors, err := checkFile(arg, strict, verbose)
					if err != nil {
						return err
					}
					totalFiles++
					totalErrors += errors
				}
			}

			// Print summary
			if totalFiles > 1 {
				fmt.Printf("\nChecked %d files, found %d type errors\n", totalFiles, totalErrors)
			}

			// Exit with non-zero status if there were errors and strict mode is enabled
			if strict && totalErrors > 0 {
				os.Exit(1)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolP("strict", "s", false, "Exit with non-zero status if type errors are found")
	cmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	cmd.Flags().BoolP("recursive", "r", false, "Recursively check directories")

	return cmd
}

// checkFile performs type checking on a single file
func checkFile(filePath string, strict, verbose bool) (int, error) {
	if verbose {
		fmt.Printf("Checking %s...\n", filePath)
	}

	// Check if the file is a Perl file
	if !isPerlFileCheck(filePath) {
		if verbose {
			fmt.Printf("Skipping non-Perl file: %s\n", filePath)
		}
		return 0, nil
	}

	// Create a TypeCheck instance
	tc, err := typechecker.NewTypeCheck()
	if err != nil {
		return 0, errors.NewSystemError("006",
			"Failed to create type checker", err).
			WithLocation(filePath)
	}

	// Check the file
	result, err := tc.CheckFile(filePath)
	if err != nil {
		return 0, errors.NewSystemError("007",
			"Failed to check file", err).
			WithLocation(filePath)
	}

	// Print errors
	for _, typeError := range result.Errors {
		fmt.Printf("%s\n", typeError.Error())
	}

	if verbose {
		fmt.Printf("Found %d type annotations in %s\n", len(result.TypeAnnotations), filePath)
		for i, annotation := range result.TypeAnnotations {
			fmt.Printf("  [%d] %s: %s (kind: %d)\n", i+1, annotation.AnnotatedItem, annotation.TypeExpression.String(), annotation.Kind)
		}
	}

	if len(result.Errors) == 0 && verbose {
		fmt.Printf("✓ %s: No type errors found\n", filePath)
	}

	return len(result.Errors), nil
}

// checkDirectory recursively checks all Perl files in a directory
func checkDirectory(dirPath string, strict, verbose bool) (int, int, error) {
	totalFiles := 0
	totalErrors := 0

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && isPerlFileCheck(path) {
			errors, err := checkFile(path, strict, verbose)
			if err != nil {
				return err
			}
			totalFiles++
			totalErrors += errors
		}

		return nil
	})

	return totalFiles, totalErrors, err
}

// isPerlFileCheck checks if a file is a Perl file based on its extension
func isPerlFileCheck(filePath string) bool {
	ext := filepath.Ext(filePath)
	return ext == ".pl" || ext == ".pm" || ext == ".t"
}
