// ABOUTME: PSC type checking command implementation
// ABOUTME: Provides static type checking for Perl code files

package psc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
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
			showInferred, _ := cmd.Flags().GetBool("show-inferred")
			dumpAST, _ := cmd.Flags().GetBool("dump-ast")

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
						files, errors, err := checkDirectory(arg, strict, verbose, showInferred)
						if err != nil {
							return err
						}
						totalFiles += files
						totalErrors += errors
					} else {
						fmt.Printf("Skipping directory %s (use --recursive to check directories)\n", arg)
					}
				} else {
					errors, err := checkFile(arg, strict, verbose, showInferred, dumpAST)
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
	cmd.Flags().BoolP("show-inferred", "i", false, "Show inferred types")
	cmd.Flags().Bool("dump-ast", false, "Dump AST structure for debugging")

	return cmd
}

// checkFile performs type checking on a single file
func checkFile(filePath string, strict, verbose, showInferred, dumpAST bool) (int, error) {
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

	// If dumping AST, parse and dump the AST structure
	if dumpAST {
		fmt.Printf("=== AST DUMP for %s ===\n", filePath)
		err := dumpASTStructure(filePath)
		if err != nil {
			return 0, errors.NewSystemError("005",
				"Failed to dump AST", err).
				WithLocation(filePath)
		}
		fmt.Printf("=== END AST DUMP ===\n\n")
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

	// Print errors with enhanced formatting
	if len(result.Errors) > 0 {
		formatter := NewErrorFormatter()
		if !verbose {
			// Less context in non-verbose mode
			formatter.SetContextLines(1)
		}
		fmt.Print(formatter.FormatErrors(result.Errors))
	}

	if verbose {
		fmt.Printf("Found %d type annotations in %s\n", len(result.TypeAnnotations), filePath)
		for i, annotation := range result.TypeAnnotations {
			fmt.Printf("  [%d] %s: %s (kind: %d)\n", i+1, annotation.AnnotatedItem, annotation.TypeExpression.String(), annotation.Kind)
		}
	}

	// Show inferred types if requested
	if showInferred && len(result.RefinedTypes) > 0 {
		fmt.Printf("\nInferred types in %s:\n", filePath)
		for varName, inferredType := range result.RefinedTypes {
			fmt.Printf("  %s: %s\n", varName, inferredType)
		}
	}

	if len(result.Errors) == 0 && verbose {
		fmt.Printf("✓ %s: No type errors found\n", filePath)
	}

	return len(result.Errors), nil
}

// checkDirectory recursively checks all Perl files in a directory
func checkDirectory(dirPath string, strict, verbose, showInferred bool) (int, int, error) {
	totalFiles := 0
	totalErrors := 0

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && isPerlFileCheck(path) {
			errors, err := checkFile(path, strict, verbose, showInferred, false) // Never dump AST in directory mode
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

// dumpASTStructure parses a file and dumps its AST structure for debugging
func dumpASTStructure(filePath string) error {
	// Parse the file using the parser
	astTree, err := parser.PooledParserFunc(func(p parser.Parser) (*parser.AST, error) {
		return p.ParseFile(filePath)
	})
	if err != nil {
		return fmt.Errorf("failed to parse file: %v", err)
	}

	if astTree == nil || astTree.Root == nil {
		fmt.Println("No AST root node found")
		return nil
	}

	// Dump the AST structure
	fmt.Printf("Root node: %s\n", astTree.Root.Type())
	dumpNode(astTree.Root, "", 0)

	return nil
}

// dumpNode recursively dumps AST node information
func dumpNode(node ast.Node, prefix string, depth int) {
	if node == nil {
		return
	}

	// Limit depth to avoid excessive output
	if depth > 10 {
		fmt.Printf("%s... (max depth reached)\n", prefix)
		return
	}

	// Get node information
	nodeType := node.Type()
	start := node.Start()
	end := node.End()
	text := strings.TrimSpace(node.Text())

	// Truncate long text for readability
	if len(text) > 50 {
		text = text[:47] + "..."
	}

	// Replace newlines and tabs for single line display
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "\t", "\\t")

	// Display node information
	fmt.Printf("%s├─ %s [%d:%d-%d:%d]", prefix, nodeType, start.Line, start.Column, end.Line, end.Column)
	if text != "" {
		fmt.Printf(" %q", text)
	}
	fmt.Println()

	// Recursively dump children
	children := node.Children()
	childPrefix := prefix + "│  "
	lastChildPrefix := prefix + "   "

	for i, child := range children {
		isLast := i == len(children)-1
		if isLast {
			fmt.Printf("%s└─ ", prefix)
			dumpNode(child, lastChildPrefix, depth+1)
		} else {
			dumpNode(child, childPrefix, depth+1)
		}
	}
}
