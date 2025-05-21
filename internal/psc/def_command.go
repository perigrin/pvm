// ABOUTME: PSC type definition command implementation
// ABOUTME: Manages type definitions for static type checking

package psc

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

// newDefCommand creates a command to manage type definitions
func newDefCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "def",
		Short: "Manage type definitions",
		Long:  "Generate and manage type definitions for Perl modules",
	}

	// Add subcommands
	cmd.AddCommand(
		newDefListCommand(),
		newDefGenerateCommand(),
		newDefImportCommand(),
		newDefExportCommand(),
		newDefInstallCommand(),
	)

	return cmd
}

// newDefListCommand creates a command to list available type definitions
func newDefListCommand() *cobra.Command {
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
				fmt.Println("Use 'psc def generate' to generate type definitions for modules.")
				return nil
			}

			fmt.Println("Available type definitions:")
			for _, module := range modules {
				fmt.Printf("  %s\n", module)
			}

			fmt.Println("\nUse 'psc def generate [module]' to generate new type definitions.")
			fmt.Println("Use 'psc check [file]' to type-check Perl files.")

			return nil
		},
	}
}

// newDefGenerateCommand creates a command to generate type definitions
func newDefGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [module]",
		Short: "Generate type definitions",
		Long:  "Generate type definitions for a Perl module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]
			version, _ := cmd.Flags().GetString("version")
			outputFile, _ := cmd.Flags().GetString("output")
			saveTypeDef, _ := cmd.Flags().GetBool("save")

			// Create a minimal type definition
			typeDef := &typedef.TypeDefinition{
				Module:     moduleName,
				Version:    version,
				Generated:  time.Now(),
				Maintainer: "PSC type generator",
				Source:     "generated",
				Types:      []typedef.TypeInfo{},
				Packages:   []typedef.PackageInfo{},
				Subs:       []typedef.SubInfo{},
				Methods:    []typedef.MethodInfo{},
			}

			// Normally, we would analyze the module here using Perl's introspection
			// facilities. For now, we'll create a minimal type definition.
			// TODO: Implement actual module analysis and type extraction

			// Add some placeholder types if none provided
			if len(typeDef.Types) == 0 {
				typeDef.Types = append(typeDef.Types, typedef.TypeInfo{
					Name:        moduleName,
					Description: "Auto-generated type information for " + moduleName,
					Kind:        "class",
					Methods:     []typedef.MethodInfo{},
					Properties:  []typedef.PropInfo{},
				})
			}

			// Marshal to JSON with indentation
			data, err := json.MarshalIndent(typeDef, "", "  ")
			if err != nil {
				return errors.NewTypeError(
					"801",
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
						return errors.NewSystemError("001",
							"Failed to create directory", err).
							WithLocation(dir)
					}
				}

				// Write the file
				if err := os.WriteFile(outputFile, data, 0644); err != nil {
					return errors.NewSystemError("002",
						"Failed to write file", err).
						WithLocation(outputFile)
				}

				fmt.Printf("Generated type definition for %s to %s\n", moduleName, outputFile)
			} else {
				// Print to stdout
				fmt.Println(string(data))
			}

			// If save flag is set, save the type definition
			if saveTypeDef {
				storage, err := typedef.NewStorage()
				if err != nil {
					return err
				}

				if err := storage.Save(typeDef); err != nil {
					return err
				}

				fmt.Printf("Saved type definition for %s to type registry\n", moduleName)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("version", "v", "0.0.1", "Module version")
	cmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	cmd.Flags().BoolP("save", "s", false, "Save the type definition to the registry")

	return cmd
}

// newDefImportCommand creates a command to import type definitions from a file
func newDefImportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import [file]",
		Short: "Import type definitions",
		Long:  "Import type definitions from a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			// Check if the file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				return errors.NewUserInputError(cli.PrefixPSC, "001",
					"File not found", err).
					WithLocation(filePath)
			}

			// Read the file
			data, err := os.ReadFile(filePath)
			if err != nil {
				return errors.NewSystemError("003",
					"Failed to read file", err).
					WithLocation(filePath)
			}

			// Unmarshal the type definition
			var typeDef typedef.TypeDefinition
			if err := json.Unmarshal(data, &typeDef); err != nil {
				return errors.NewTypeError(
					"802",
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

			fmt.Printf("Imported type definition for %s\n", typeDef.Module)
			return nil
		},
	}
}

// newDefExportCommand creates a command to export type definitions to a file
func newDefExportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "export [module] [file]",
		Short: "Export type definitions",
		Long:  "Export type definitions to a file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]
			outputFile := args[1]

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

			// Marshal to JSON with indentation
			data, err := json.MarshalIndent(typeDef, "", "  ")
			if err != nil {
				return errors.NewTypeError(
					"803",
					fmt.Sprintf("Failed to marshal type definition for %s", moduleName),
					err,
				)
			}

			// Create the directory if it doesn't exist
			dir := filepath.Dir(outputFile)
			if dir != "." {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return errors.NewSystemError("004",
						"Failed to create directory", err).
						WithLocation(dir)
				}
			}

			// Write the file
			if err := os.WriteFile(outputFile, data, 0644); err != nil {
				return errors.NewSystemError("005",
					"Failed to write file", err).
					WithLocation(outputFile)
			}

			fmt.Printf("Exported type definition for %s to %s\n", moduleName, outputFile)
			return nil
		},
	}
}

// newDefInstallCommand creates a command to install type definitions for a CPAN module
func newDefInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [module]",
		Short: "Install type definitions",
		Long:  "Install type definitions for a CPAN module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]
			forceGenerate, _ := cmd.Flags().GetBool("force")

			// Create a new storage instance
			storage, err := typedef.NewStorage()
			if err != nil {
				return err
			}

			// Check if the type definition already exists
			_, err = storage.Load(moduleName)
			if err == nil && !forceGenerate {
				fmt.Printf("Type definition for %s already exists. Use --force to regenerate.\n", moduleName)
				return nil
			}

			// Generate and save the type definition
			typeDef := &typedef.TypeDefinition{
				Module:     moduleName,
				Version:    "0.0.1",
				Generated:  time.Now(),
				Maintainer: "PSC type generator",
				Source:     "installed",
				Types:      []typedef.TypeInfo{},
				Packages:   []typedef.PackageInfo{},
				Subs:       []typedef.SubInfo{},
				Methods:    []typedef.MethodInfo{},
			}

			// Normally, we would analyze the module here using Perl's introspection
			// For now, we'll create a minimal type definition like in the generate command
			// TODO: Same as generate - implement module analysis

			// Add some placeholder types
			typeDef.Types = append(typeDef.Types, typedef.TypeInfo{
				Name:        moduleName,
				Description: "Auto-generated type information for " + moduleName,
				Kind:        "class",
				Methods:     []typedef.MethodInfo{},
				Properties:  []typedef.PropInfo{},
			})

			// Save the type definition
			if err := storage.Save(typeDef); err != nil {
				return err
			}

			fmt.Printf("Installed type definition for %s\n", moduleName)

			// Add instructions for using the type definition
			fmt.Printf("\nYou can now use the type definition in your Perl code:\n\n")
			fmt.Printf("use %s : typed;\n\n", moduleName)
			fmt.Printf("Or check your code with:\n\n")
			fmt.Printf("psc check your_script.pl\n")

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolP("force", "f", false, "Force generation of type definition even if it already exists")

	return cmd
}
