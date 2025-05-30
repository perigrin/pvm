// ABOUTME: PSC strip command implementation
// ABOUTME: Provides functionality to strip type annotations from Perl code

package psc

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/compiler"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
)

// newStripCommand creates a command to strip type annotations from a file
func newStripCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "strip [file] [output]",
		Short: "Strip type annotations from a file",
		Long: `Remove type annotations from a Perl file to create clean, compatible Perl code.

The strip command safely removes type annotations while preserving all original
program logic and comments. The resulting code will run on any standard Perl
interpreter without modification.

This is useful for:
• Publishing modules to CPAN
• Running on systems without PSC
• Backwards compatibility with older Perl versions
• Code sharing with teams not using typed Perl

Output:
  If no output file is specified, writes to stdout
  If output file is specified, writes cleaned code to that file

Examples:
  psc strip typed.pl                    # Print to stdout
  psc strip typed.pl clean.pl           # Write to file
  psc strip lib/MyModule.pm lib/clean/  # Process to directory`,
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

			// Parse the file using parser pool for thread safety
			ast, err := parser.PooledParserFunc(func(p parser.Parser) (*parser.AST, error) {
				return p.ParseFile(inputFile)
			})
			if err != nil {
				return errors.NewTypeError(
					ErrIntegrationFailed,
					fmt.Sprintf("Failed to parse file %s", inputFile),
					err).WithLocation(inputFile)
			}

			// Use compiler to strip type annotations
			registry := compiler.NewCompilerRegistry()
			astAdapter := compiler.NewParserASTAdapter(ast)
			result, err := registry.Compile(astAdapter, compiler.TargetCleanPerl)
			if err != nil {
				return errors.NewTypeError(
					ErrIntegrationFailed,
					fmt.Sprintf("Failed to strip type annotations from %s", inputFile),
					err).WithLocation(inputFile)
			}

			// Write the result to output file or stdout
			if outputFile != "" {
				err = os.WriteFile(outputFile, []byte(result), 0644)
				if err != nil {
					return errors.NewTypeError(
						ErrIntegrationFailed,
						fmt.Sprintf("Failed to write output file %s", outputFile),
						err).WithLocation(outputFile)
				}

				fmt.Printf("Stripped type annotations from %s and wrote result to %s\n", inputFile, outputFile)
			} else {
				fmt.Println(result)
			}

			return nil
		},
	}
}
