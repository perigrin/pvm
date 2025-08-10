// ABOUTME: PVI type command implementation
// ABOUTME: Manages type definitions for Perl modules

package pm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
			ui := cli.GetUI(cmd)
			if len(modules) == 0 {
				ui.Info("No type definitions available.")
				return nil
			}

			ui.SubHeader("Available type definitions:")
			ui.List(modules)

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
				ui := cli.GetUI(cmd)
				ui.Println(string(data))

			case "summary":
				// Print a summary of the type definition
				ui := cli.GetUI(cmd)
				ui.SubHeader(fmt.Sprintf("Type Definition Summary: %s", typeDef.Module))

				// Create key-value pairs for display
				info := map[string]string{
					"Module":      typeDef.Module,
					"Version":     typeDef.Version,
					"Generated":   typeDef.Generated.Format(time.RFC3339),
					"Maintainer":  typeDef.Maintainer,
					"Source":      typeDef.Source,
					"Types":       fmt.Sprintf("%d", len(typeDef.Types)),
					"Packages":    fmt.Sprintf("%d", len(typeDef.Packages)),
					"Subroutines": fmt.Sprintf("%d", len(typeDef.Subs)),
					"Methods":     fmt.Sprintf("%d", len(typeDef.Methods)),
				}
				ui.KeyValue(info)

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

			ui := cli.GetUI(cmd)
			ui.Success("Added type definition for %s", typeDef.Module)
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

			ui := cli.GetUI(cmd)
			ui.Success("Removed type definition for %s", moduleName)
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

			// Create a module analyzer
			analyzer, err := NewModuleAnalyzer()
			if err != nil {
				return errors.NewSystemError("303",
					"Failed to create module analyzer", err)
			}

			// Try to find the module file
			modulePath := findModuleFile(moduleName)
			if modulePath == "" {
				return errors.NewUserInputError(cli.PrefixPVI, "304",
					fmt.Sprintf("Cannot locate module file for %s", moduleName), nil).
					WithHint("Specify the full path to the .pm or .pl file")
			}

			// Analyze the module to generate type definition
			typeDef, err := analyzer.AnalyzeModule(modulePath)
			if err != nil {
				return err
			}

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

				ui := cli.GetUI(cmd)
				ui.Success("Generated type definition for %s to %s", moduleName, outputFile)
			} else {
				// Print to stdout
				ui := cli.GetUI(cmd)
				ui.Println(string(data))
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

				ui := cli.GetUI(cmd)
				ui.Success("Saved type definition for %s", moduleName)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	cmd.Flags().BoolP("save", "s", false, "Save the type definition")

	return cmd
}

// findModuleFile attempts to locate a module file given a module name
func findModuleFile(moduleName string) string {
	// If it's already a file path, use it directly
	if strings.HasSuffix(moduleName, ".pm") || strings.HasSuffix(moduleName, ".pl") {
		if _, err := os.Stat(moduleName); err == nil {
			return moduleName
		}
	}

	// Convert module name to file path (e.g., Foo::Bar -> Foo/Bar.pm)
	pathParts := strings.Split(moduleName, "::")
	relativePath := strings.Join(pathParts, string(filepath.Separator)) + ".pm"

	// Search in common locations
	searchPaths := []string{
		"lib/" + relativePath, // Standard lib structure
		relativePath,          // Direct relative path
	}

	// Also check environment @INC equivalent locations
	if homeDir, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(homeDir, ".perl5", "lib", relativePath),
			filepath.Join(homeDir, "perl5", "lib", relativePath),
		)
	}

	// Common system perl paths (simplified)
	systemPaths := []string{
		"/usr/share/perl5/" + relativePath,
		"/usr/local/share/perl5/" + relativePath,
		"/opt/perl/lib/" + relativePath,
	}
	searchPaths = append(searchPaths, systemPaths...)

	// Search for the file
	for _, searchPath := range searchPaths {
		if _, err := os.Stat(searchPath); err == nil {
			return searchPath
		}
	}

	return "" // Not found
}
