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
		newCheckCommand(),
		newStripCommand(),
		newRunCommand(),
		newWatchCommand(),
		newDefCommand(),
	)

	return cmd
}

// Placeholder commands, to be implemented later

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
			checker, err := parser.NewTypeChecker()
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
					return checkSingleFile(checker, filePath, verbose)
				})
			} else {
				// Check a single file
				return checkSingleFile(checker, path, verbose)
			}
		},
	}

	// Add command-specific flags
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	return cmd
}

// checkSingleFile checks a single Perl file for type errors
func checkSingleFile(checker *parser.TypeChecker, path string, verbose bool) error {
	if verbose {
		fmt.Printf("Checking file: %s\n", path)
	}

	// Type check the file
	result, err := checker.CheckFile(path)
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

func newStripCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "strip [file] [output]",
		Short: "Strip type annotations from a file",
		Long:  "Remove type annotations from a Perl file for compatibility",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("expected a file to strip")
			}

			// Get the input file
			inputFile := args[0]

			// Get the output file (default to stdout if not provided)
			outputFile := ""
			if len(args) > 1 {
				outputFile = args[1]
			}

			// Strip type annotations
			result, err := parser.StripAnnotations(inputFile)
			if err != nil {
				return fmt.Errorf("failed to strip type annotations from %s: %v", inputFile, err)
			}

			// Write the result to output file or stdout
			if outputFile != "" {
				err = os.WriteFile(outputFile, []byte(result), 0644)
				if err != nil {
					return fmt.Errorf("failed to write output file %s: %v", outputFile, err)
				}

				fmt.Printf("Stripped type annotations from %s and wrote result to %s\n", inputFile, outputFile)
			} else {
				fmt.Println(result)
			}

			return nil
		},
	}
}

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [file] [args...]",
		Short: "Type-check and execute a file",
		Long:  "Perform type checking and then execute the Perl file if no errors are found",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("expected a file to run")
			}

			// Get the file to run
			file := args[0]

			// Get the script arguments (if any)
			scriptArgs := []string{}
			if len(args) > 1 {
				scriptArgs = args[1:]
			}

			// Get flags
			verbose, _ := cmd.Flags().GetBool("verbose")
			skipCheck, _ := cmd.Flags().GetBool("skip-check")
			perl, _ := cmd.Flags().GetString("perl")

			// Type check the file first (unless skipped)
			if !skipCheck {
				// Create a type checker
				checker, err := parser.NewTypeChecker()
				if err != nil {
					return fmt.Errorf("failed to create type checker: %v", err)
				}

				// Type check the file
				result, err := checker.CheckFile(file)
				if err != nil {
					return fmt.Errorf("failed to check file %s: %v", file, err)
				}

				// If there are errors, report them and exit
				if len(result.Errors) > 0 {
					fmt.Printf("Found %d type errors in %s\n", len(result.Errors), file)

					for _, errInfo := range result.Errors {
						fmt.Printf("  %s\n", errInfo.Error())
					}

					return fmt.Errorf("type checking failed, aborting execution")
				}

				if verbose {
					fmt.Printf("Type checking passed for %s\n", file)
				}
			} else if verbose {
				fmt.Printf("Skipping type checking for %s\n", file)
			}

			// Execute the file using PVX
			// For now, we'll use a simple implementation that just prints a message
			// In a real implementation, we would use the PVX executor

			if verbose {
				fmt.Printf("Executing %s with arguments: %v\n", file, scriptArgs)
				if perl != "" {
					fmt.Printf("Using Perl version: %s\n", perl)
				}
			}

			// TODO: Implement actual execution using PVX
			fmt.Printf("This is a placeholder for executing %s - actual execution not yet implemented\n", file)

			return nil
		},
	}

	// Add command-specific flags
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().Bool("skip-check", false, "Skip type checking")
	cmd.Flags().StringP("perl", "p", "", "Use a specific Perl version")

	return cmd
}

func newWatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch [file|dir]",
		Short: "Watch files and report type errors on change",
		Long:  "Continuously monitor files for changes and perform type checking",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("expected a file or directory to watch")
			}

			// Get the path to watch
			path := args[0]

			// Create a type checker
			checker, err := parser.NewTypeChecker()
			if err != nil {
				return fmt.Errorf("failed to create type checker: %v", err)
			}

			// Check if the path exists
			_, err = os.Stat(path)
			if err != nil {
				return fmt.Errorf("failed to stat path: %v", err)
			}

			fmt.Printf("Watching %s for changes...\n", path)
			fmt.Println("(This is a placeholder implementation. Press Ctrl+C to exit.)")

			// In a real implementation, we would:
			// 1. Set up a file watcher using something like fsnotify
			// 2. Watch for file changes
			// 3. Re-check files when they change

			// For now, we'll just do an initial check and then wait for user input

			// Do an initial check
			if info, err := os.Stat(path); err == nil {
				if info.IsDir() {
					// Check all Perl files in the directory
					fmt.Printf("Checking all Perl files in directory: %s\n", path)

					filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
						if err != nil {
							fmt.Printf("Error accessing %s: %v\n", filePath, err)
							return nil
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
						result, err := checker.CheckFile(filePath)
						if err != nil {
							fmt.Printf("Error checking %s: %v\n", filePath, err)
							return nil
						}

						// Report errors
						if len(result.Errors) > 0 {
							fmt.Printf("Found %d type errors in %s\n", len(result.Errors), filePath)

							for _, errInfo := range result.Errors {
								fmt.Printf("  %s\n", errInfo.Error())
							}
						} else {
							fmt.Printf("No type errors found in %s\n", filePath)
						}

						return nil
					})
				} else {
					// Check a single file
					result, err := checker.CheckFile(path)
					if err != nil {
						fmt.Printf("Error checking %s: %v\n", path, err)
					} else if len(result.Errors) > 0 {
						fmt.Printf("Found %d type errors in %s\n", len(result.Errors), path)

						for _, errInfo := range result.Errors {
							fmt.Printf("  %s\n", errInfo.Error())
						}
					} else {
						fmt.Printf("No type errors found in %s\n", path)
					}
				}
			}

			// Wait for user input (simulating watching for changes)
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()

			return nil
		},
	}

	// Add command-specific flags
	cmd.Flags().StringArray("exclude", []string{}, "Patterns to exclude from watching")

	return cmd
}

// The newDefCommand implementation is in def_command.go
