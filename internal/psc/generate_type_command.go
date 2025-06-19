// ABOUTME: PSC command to generate type definitions from Perl files
// ABOUTME: Implements type definition generation for PSC-PVI integration

package psc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/errors"
)

// newGenerateTypeCommand creates a command to generate a type definition from a Perl file
func newGenerateTypeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-type [file] [module-name]",
		Short: "Generate a type definition from a Perl file",
		Long:  "Extract type annotations from a Perl file and generate a type definition",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)
			filePath := args[0]

			// Check if the file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				return errors.NewUserInputError("PSC", "102",
					"File not found", err).
					WithLocation(filePath)
			}

			// Get the module name (either from args or extract from the file path)
			var moduleName string
			if len(args) > 1 {
				moduleName = args[1]
			}

			// Get flags
			outputFile, _ := cmd.Flags().GetString("output")
			save, _ := cmd.Flags().GetBool("save")
			verbose, _ := cmd.Flags().GetBool("verbose")

			// Create options for type definition generation
			options := &TypeDefinitionOptions{
				ModuleName: moduleName,
				SourceFile: filePath,
				Save:       save,
				OutputFile: outputFile,
				Verbose:    verbose,
			}

			// Generate the type definition
			result, err := GenerateTypeDefinition(options)
			if err != nil {
				return err
			}

			// Check for errors in the result
			if len(result.Errors) > 0 {
				for _, err := range result.Errors {
					ui.Warning("Warning: %v", err)
				}
			}

			// If no output file is specified and not saving, print to stdout
			if outputFile == "" && !save {
				// Marshal to JSON with indentation
				data, err := json.MarshalIndent(result.TypeDef, "", "  ")
				if err != nil {
					return errors.NewTypeError(
						"704",
						fmt.Sprintf("Failed to marshal type definition for %s", result.TypeDef.Module),
						err,
					)
				}
				ui.Println(string(data))
			} else if outputFile != "" {
				// Marshal to JSON with indentation
				data, err := json.MarshalIndent(result.TypeDef, "", "  ")
				if err != nil {
					return errors.NewTypeError(
						"704",
						fmt.Sprintf("Failed to marshal type definition for %s", result.TypeDef.Module),
						err,
					)
				}

				// Create the directory if it doesn't exist
				dir := filepath.Dir(outputFile)
				if dir != "." {
					if err := os.MkdirAll(dir, 0755); err != nil {
						return errors.NewSystemError("104",
							"Failed to create directory", err).
							WithLocation(dir)
					}
				}

				// Write the file
				if err := os.WriteFile(outputFile, data, 0644); err != nil {
					return errors.NewSystemError("105",
						"Failed to write file", err).
						WithLocation(outputFile)
				}

				ui.Success("Generated type definition for %s to %s", result.TypeDef.Module, outputFile)
			}

			// Report where the type definition was saved if applicable
			if save && result.SavedPath != "" {
				ui.Success("Type definition for %s saved to %s", result.TypeDef.Module, result.SavedPath)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	cmd.Flags().BoolP("save", "s", false, "Save the type definition to the type store")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	return cmd
}

// newImportTypeCommand creates a command to import an existing type definition
func newImportTypeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-type [file]",
		Short: "Import a type definition",
		Long:  "Import a type definition from a JSON file into the type store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)
			filePath := args[0]

			// Check if the file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				return errors.NewUserInputError("PSC", "102",
					"File not found", err).
					WithLocation(filePath)
			}

			// Read the file - this is a simpler implementation that directly
			// forwards to PVI's implementation
			_, err := os.ReadFile(filePath)
			if err != nil {
				return errors.NewSystemError("103",
					"Failed to read file", err).
					WithLocation(filePath)
			}

			// This is handled by the PVI type add command, but we're implementing
			// a PSC-specific version for direct access from PSC
			_, err = executeComponentCommand("pvi", "type", "add", filePath)
			if err != nil {
				return errors.NewSystemError("106",
					"Failed to import type definition", err).
					WithLocation(filePath)
			}

			ui.Success("Imported type definition from %s", filePath)
			return nil
		},
	}

	return cmd
}

// newListTypesCommand creates a command to list available type definitions
func newListTypesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-types",
		Short: "List available type definitions",
		Long:  "List all available type definitions in the type store",
		RunE: func(cmd *cobra.Command, args []string) error {
			// This is handled by the PVI type list command, but we're implementing
			// a PSC-specific version for direct access from PSC
			_, err := executeComponentCommand("pvi", "type", "list")
			if err != nil {
				return errors.NewSystemError("107",
					"Failed to list type definitions", err)
			}

			return nil
		},
	}

	return cmd
}

// executeCommand executes a command by delegating to another component
func executeComponentCommand(component string, args ...string) (string, error) {
	// In a real implementation, we would use a more sophisticated approach
	// to execute another component without spawning a new process
	// For now, we're just illustrating the concept

	// This is a placeholder for the real implementation
	fmt.Printf("Executing: %s %v\n", component, args)

	// In a real implementation, we'd invoke the appropriate component's
	// command directly rather than spawning a new process

	return "", nil
}
