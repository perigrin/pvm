// ABOUTME: Configuration generation command implementation
// ABOUTME: Provides functionality to generate configurations from templates and profiles

package pvm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/xdg"
)

// newConfigGenerateCommand creates a config generate command
func newConfigGenerateCommand() *cobra.Command {
	var (
		templateName   string
		profileName    string
		outputFile     string
		variables      []string
		listItems      bool
		forceOverwrite bool
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate configuration from templates or profiles",
		Long: `Generate configuration files from templates or profiles.

This command allows you to create new configuration files based on:
- Templates: Reusable configuration patterns with variables
- Profiles: Environment-specific configurations with inheritance

Examples:
  # Generate from template
  pvm config generate --template basic --output pvm.toml

  # Generate from profile
  pvm config generate --profile development --output pvm.toml

  # Generate with custom variables
  pvm config generate --template basic --var "perl_version=5.38.0" --var "environment=prod"

  # List available templates and profiles
  pvm config generate --list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get XDG directories
			dirs, err := xdg.GetDirs()
			if err != nil {
				return fmt.Errorf("failed to get XDG directories: %w", err)
			}

			// Setup managers
			templateManager := config.NewTemplateManager(filepath.Join(dirs.ConfigHome, "pvm", "templates"))
			profileManager := config.NewProfileManager(filepath.Join(dirs.ConfigHome, "pvm", "profiles"), templateManager)

			// Load templates and profiles
			if err := templateManager.LoadTemplates(); err != nil {
				return fmt.Errorf("failed to load templates: %w", err)
			}

			if err := profileManager.LoadProfiles(); err != nil {
				return fmt.Errorf("failed to load profiles: %w", err)
			}

			// Handle list command
			if listItems {
				return listTemplatesAndProfiles(templateManager, profileManager)
			}

			// Validate arguments
			if templateName == "" && profileName == "" {
				return fmt.Errorf("either --template or --profile must be specified")
			}

			if templateName != "" && profileName != "" {
				return fmt.Errorf("cannot specify both --template and --profile")
			}

			// Set default output if not specified
			if outputFile == "" {
				outputFile = "pvm.toml"
			}

			// Check if output file exists
			if !forceOverwrite {
				if _, err := os.Stat(outputFile); err == nil {
					return fmt.Errorf("output file '%s' already exists, use --force to overwrite", outputFile)
				}
			}

			// Parse variables
			parsedVariables, err := parseVariables(variables)
			if err != nil {
				return fmt.Errorf("failed to parse variables: %w", err)
			}

			// Generate configuration
			var generatedConfig *config.Config

			if templateName != "" {
				generatedConfig, err = templateManager.RenderTemplate(templateName, parsedVariables)
				if err != nil {
					return fmt.Errorf("failed to render template '%s': %w", templateName, err)
				}
				fmt.Printf("Generated configuration from template '%s'\n", templateName)
			} else {
				generatedConfig, err = profileManager.ResolveProfile(profileName, parsedVariables)
				if err != nil {
					return fmt.Errorf("failed to resolve profile '%s': %w", profileName, err)
				}
				fmt.Printf("Generated configuration from profile '%s'\n", profileName)
			}

			// Save to output file
			if err := config.SaveToFile(generatedConfig, outputFile); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			fmt.Printf("Configuration saved to '%s'\n", outputFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&templateName, "template", "", "Template name to use for generation")
	cmd.Flags().StringVar(&profileName, "profile", "", "Profile name to use for generation")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: pvm.toml)")
	cmd.Flags().StringArrayVar(&variables, "var", []string{}, "Template variables in key=value format")
	cmd.Flags().BoolVar(&listItems, "list", false, "List available templates and profiles")
	cmd.Flags().BoolVar(&forceOverwrite, "force", false, "Overwrite existing output file")

	return cmd
}

// listTemplatesAndProfiles lists available templates and profiles
func listTemplatesAndProfiles(templateManager *config.TemplateManager, profileManager *config.ProfileManager) error {
	fmt.Println("Available Templates:")
	templates := templateManager.ListTemplates()
	if len(templates) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, name := range templates {
			template, _ := templateManager.GetTemplate(name)
			fmt.Printf("  %s", name)
			if template.Description != "" {
				fmt.Printf(" - %s", template.Description)
			}
			fmt.Println()
		}
	}

	fmt.Println("\nAvailable Profiles:")
	profiles := profileManager.ListProfiles()
	if len(profiles) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, name := range profiles {
			profile, _ := profileManager.GetProfile(name)
			fmt.Printf("  %s (%s)", name, profile.Environment)
			if profile.Description != "" {
				fmt.Printf(" - %s", profile.Description)
			}
			fmt.Println()
		}
	}

	fmt.Println("\nEnvironments:")
	environments := profileManager.GetEnvironments()
	if len(environments) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, env := range environments {
			envProfiles := profileManager.ListProfilesByEnvironment(env)
			fmt.Printf("  %s: %v\n", env, envProfiles)
		}
	}

	return nil
}

// parseVariables parses variable strings into a map
func parseVariables(variableStrings []string) (map[string]string, error) {
	variables := make(map[string]string)

	for _, varStr := range variableStrings {
		parts := strings.SplitN(varStr, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid variable format '%s', expected key=value", varStr)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("empty variable key in '%s'", varStr)
		}

		variables[key] = value
	}

	return variables, nil
}
