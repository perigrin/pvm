// ABOUTME: PVI type command implementation
// ABOUTME: Manages type definitions for Perl modules

package pvi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/typedef"
)

// createTypeCommands creates subcommands for type definitions
func createTypeCommands(cmd *cobra.Command) {
	// Add subcommands
	cmd.AddCommand(
		newTypeListCommand(),
		newTypeGetCommand(),
		newTypeAddCommand(),
		newTypeRemoveCommand(),
		newTypeGenerateCommand(),
	)
}

// newTypeListCommand creates a command to list type definitions
func newTypeListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available type definitions",
		Long:  "List all available type definitions for Perl modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create a new storage instance
			storage, err := typedef.NewStorage()
			if err != nil {
				return err
			}

			// Get the list of type definitions
			modules, err := storage.List()
			if err != nil {
				return err
			}

			// Print the list of modules
			if len(modules) == 0 {
				fmt.Println("No type definitions available.")
				return nil
			}

			fmt.Println("Available type definitions:")
			for _, module := range modules {
				fmt.Printf("  %s\n", module)
			}

			return nil
		},
	}
}

// newTypeGetCommand creates a command to get type definition information
func newTypeGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [module]",
		Short: "Get type definition for a module",
		Long:  "Get detailed type definition information for a Perl module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]

			// Create a new storage instance
			storage, err := typedef.NewStorage()
			if err != nil {
				return err
			}

			// Load the type definition
			typeDef, err := storage.Load(moduleName)
			if err != nil {
				return err
			}

			// Get output format
			format, _ := cmd.Flags().GetString("format")
			switch format {
			case "json":
				// Marshal to JSON with indentation
				data, err := json.MarshalIndent(typeDef, "", "  ")
				if err != nil {
					return errors.NewTypeError(
						"704",
						fmt.Sprintf("Failed to marshal type definition for %s", moduleName),
						err,
					)
				}
				fmt.Println(string(data))

			case "summary":
				// Print a summary of the type definition
				fmt.Printf("Module: %s\n", typeDef.Module)
				fmt.Printf("Version: %s\n", typeDef.Version)
				fmt.Printf("Generated: %s\n", typeDef.Generated.Format(time.RFC3339))
				fmt.Printf("Maintainer: %s\n", typeDef.Maintainer)
				fmt.Printf("Source: %s\n", typeDef.Source)
				fmt.Printf("Types: %d\n", len(typeDef.Types))
				fmt.Printf("Packages: %d\n", len(typeDef.Packages))
				fmt.Printf("Subroutines: %d\n", len(typeDef.Subs))
				fmt.Printf("Methods: %d\n", len(typeDef.Methods))

			default:
				return errors.NewUserInputError(cli.PrefixPVI, "101",
					"Invalid format", nil).
					WithHint("Valid formats are: json, summary")
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("format", "f", "summary", "Output format (json, summary)")

	return cmd
}

// newTypeAddCommand creates a command to add a type definition
func newTypeAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [file]",
		Short: "Add a type definition",
		Long:  "Add a type definition from a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			// Check if the file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				return errors.NewUserInputError(cli.PrefixPVI, "102",
					"File not found", err).
					WithLocation(filePath)
			}

			// Read the file
			data, err := os.ReadFile(filePath)
			if err != nil {
				return errors.NewSystemError("103",
					"Failed to read file", err).
					WithLocation(filePath)
			}

			// Unmarshal the type definition
			var typeDef typedef.TypeDefinition
			if err := json.Unmarshal(data, &typeDef); err != nil {
				return errors.NewTypeError(
					"704",
					"Failed to parse type definition",
					err,
				).WithLocation(filePath)
			}

			// Create a new storage instance
			storage, err := typedef.NewStorage()
			if err != nil {
				return err
			}

			// Save the type definition
			if err := storage.Save(&typeDef); err != nil {
				return err
			}

			fmt.Printf("Added type definition for %s\n", typeDef.Module)
			return nil
		},
	}

	return cmd
}

// newTypeRemoveCommand creates a command to remove a type definition
func newTypeRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [module]",
		Short: "Remove a type definition",
		Long:  "Remove a type definition for a Perl module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]

			// Create a new storage instance
			storage, err := typedef.NewStorage()
			if err != nil {
				return err
			}

			// Delete the type definition
			if err := storage.Delete(moduleName); err != nil {
				return err
			}

			fmt.Printf("Removed type definition for %s\n", moduleName)
			return nil
		},
	}
}

// newTypeGenerateCommand creates a command to generate a type definition
func newTypeGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [module]",
		Short: "Generate a type definition",
		Long:  "Generate a type definition for a Perl module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]
			outputFile, _ := cmd.Flags().GetString("output")
			autoSave, _ := cmd.Flags().GetBool("save")

			// Create a minimal type definition
			typeDef := &typedef.TypeDefinition{
				Module:     moduleName,
				Version:    "0.0.1", // This would be determined from the actual module
				Generated:  time.Now(),
				Maintainer: "PVI type generator",
				Source:     "generated",
				Types:      []typedef.TypeInfo{},
				Packages:   []typedef.PackageInfo{},
				Subs:       []typedef.SubInfo{},
				Methods:    []typedef.MethodInfo{},
			}

			// TODO: In a future implementation, we would actually analyze the module
			// to generate a more complete and accurate type definition.
			// For now, we'll just create a placeholder.

			// Marshal to JSON with indentation
			data, err := json.MarshalIndent(typeDef, "", "  ")
			if err != nil {
				return errors.NewTypeError(
					"704",
					fmt.Sprintf("Failed to marshal type definition for %s", moduleName),
					err,
				)
			}

			// If an output file is specified, write to it
			if outputFile != "" {
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

				fmt.Printf("Generated type definition for %s to %s\n", moduleName, outputFile)
			} else {
				// Print to stdout
				fmt.Println(string(data))
			}

			// If auto-save is enabled, save the type definition
			if autoSave {
				storage, err := typedef.NewStorage()
				if err != nil {
					return err
				}

				if err := storage.Save(typeDef); err != nil {
					return err
				}

				fmt.Printf("Saved type definition for %s\n", moduleName)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	cmd.Flags().BoolP("save", "s", false, "Save the type definition")

	return cmd
}