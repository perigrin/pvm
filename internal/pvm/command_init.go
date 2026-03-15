// ABOUTME: This file contains the implementation of the `pvm init` command for shell integration setup.
// ABOUTME: It handles shell detection, registry rebuilding, and generates shell integration scripts.

package pvm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/xdg"
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

			// Check if registry needs rebuilding using comprehensive detection
			if needsRegistryRebuild() {
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

// needsRegistryRebuild checks if the registry needs rebuilding
// Rebuilds if registry is missing, corrupted, empty, or contains invalid entries
func needsRegistryRebuild() bool {
	// Load and validate registry
	registry, err := perl.LoadRegistry()
	if err != nil {
		return true // Registry file corrupted or missing
	}

	// Get XDG directories for filesystem validation
	dirs, err := xdg.GetDirs()
	if err != nil {
		return true // If we can't get dirs, assume rebuild needed
	}

	versionsDir := filepath.Join(dirs.DataDir, "versions")

	// If registry is empty, check if versions exist on filesystem
	if len(registry.Versions) == 0 {
		if entries, err := os.ReadDir(versionsDir); err == nil && len(entries) > 0 {
			return true // Registry is empty but versions directory contains installations
		}
		return false // Both registry and filesystem are empty
	}

	// Registry has entries - validate that they correspond to actual installations
	validEntries := 0
	for _, versionInfo := range registry.Versions {
		// Skip system perl (doesn't need to be in versions directory)
		if versionInfo.Source == "system" {
			validEntries++
			continue
		}

		// For PVM installations, verify the path exists and looks reasonable
		if versionInfo.Source == "pvm" || versionInfo.Source == "plenv" {
			// Check if this looks like a test path (corrupted entry)
			if strings.Contains(versionInfo.InstallPath, "/tmp/") ||
				strings.Contains(versionInfo.InstallPath, "pvm-shim-test") ||
				strings.Contains(versionInfo.InstallPath, "/var/folders/") {
				continue // Skip corrupted test entries
			}

			// Verify the installation path exists
			if _, err := os.Stat(versionInfo.InstallPath); err == nil {
				validEntries++
			}
		}
	}

	// If we have no valid entries but filesystem has installations, rebuild
	if validEntries == 0 {
		if entries, err := os.ReadDir(versionsDir); err == nil && len(entries) > 0 {
			return true
		}
	}

	return false // Registry has valid entries
}
