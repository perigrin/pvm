// ABOUTME: This file contains the implementation of the `pvm init` command for shell integration setup.
// ABOUTME: It handles shell detection, registry rebuilding, and generates shell integration scripts.

package pvm

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/perl"
)

// newInitCommand creates a command to initialize shell integration
func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [shell]",
		Short: "Initialize shell integration",
		Long:  "Generate shell integration script for the specified or current shell.\n\nSupported shells: bash, zsh, fish, powershell",
		RunE: func(cmd *cobra.Command, args []string) error {
			var shell perl.ShellType
			var err error
			
			// Use explicit shell argument if provided, otherwise auto-detect
			if len(args) > 0 {
				shellName := args[0]
				switch shellName {
				case "bash":
					shell = perl.ShellBash
				case "zsh":
					shell = perl.ShellZsh
				case "fish":
					shell = perl.ShellFish
				case "powershell", "pwsh":
					shell = perl.ShellPowerShell
				default:
					return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shellName)
				}
			} else {
				// Detect shell type automatically
				shell, err = perl.DetectShell()
				if err != nil {
					return err
				}
			}

			// Check if registry exists, rebuild if missing or empty
			registry, err := perl.LoadRegistry()
			if err != nil || len(registry.Versions) == 0 {
				// Registry is missing, corrupted, or empty - try to rebuild it
				if rebuildErr := perl.RebuildRegistry(); rebuildErr == nil {
					fmt.Fprintf(os.Stderr, "Registry rebuilt from existing installations\n")
				}
			}

			// Check if we should perform automatic import
			// Only do this if we have no existing versions (first run)
			if perl.ShouldAutoImport() {
				// Perform automatic import of legacy tools
				results, err := perl.AutoImportLegacyVersions()
				if err == nil && results.TotalImported > 0 {
					// Print import results to stderr so it doesn't interfere with shell eval
					perl.PrintAutoImportResults(results)
				}
			}

			// Get shell script for the detected shell
			script, err := perl.GetCurrentShellScript(shell)
			if err != nil {
				return err
			}

			// Print the script to stdout (for eval)
			ui := cli.GetUI(cmd)
			ui.Printf("%s", script)
			return nil
		},
	}

	// No flags needed for shell integration output

	return cmd
}