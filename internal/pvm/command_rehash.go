// ABOUTME: This file contains the implementation of the `pvm rehash` command for updating shell PATH.
// ABOUTME: It manually triggers the shell PATH update logic that normally happens during directory changes.

package pvm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tamarou.com/pvm/internal/current"
	"tamarou.com/pvm/internal/xdg"
)

// newRehashCommand creates a command to manually update the shell PATH
func newRehashCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rehash",
		Short: "Update shell PATH to reflect current Perl version",
		Long: `Update the shell PATH to reflect the currently active Perl version.

This command manually triggers the same PATH update logic that normally happens
automatically when you change directories. Use this after:

- Manually editing .perl-version files
- Setting PVM_PERL_VERSION or other environment variables
- Changing global PVM configuration
- Installing or uninstalling Perl versions

The command will:
1. Clean existing PVM perl directories from PATH
2. Determine the currently active Perl version
3. Add the appropriate perl bin directory to PATH
4. Export the updated PATH

Examples:
  pvm rehash              # Update PATH for current version
  pvm rehash --dry-run    # Show what the new PATH would be
  pvm rehash --verbose    # Show detailed PATH update process`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}

			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			// Get current PATH
			originalPath := os.Getenv("PATH")
			if verbose {
				fmt.Printf("Original PATH: %s\n", originalPath)
			}

			// Clean PATH by removing PVM perl directories
			cleanPath, err := cleanPVMFromPath(originalPath, verbose)
			if err != nil {
				return fmt.Errorf("failed to clean PATH: %w", err)
			}

			if verbose {
				fmt.Printf("Cleaned PATH: %s\n", cleanPath)
			}

			// Get current version
			info, err := current.GetCurrentVersion()
			if err != nil {
				return fmt.Errorf("failed to get current version: %w", err)
			}

			if verbose {
				fmt.Printf("Current version: %s (source: %s)\n", info.Version, info.Source)
			}

			// Build new PATH
			newPath := cleanPath
			if info.Version != "" && info.Version != "system" {
				// Get XDG directories
				dirs, err := xdg.GetDirs()
				if err != nil {
					return fmt.Errorf("failed to get XDG directories: %w", err)
				}

				// Add version bin path
				versionBinPath := filepath.Join(dirs.DataDir, "versions", info.Version, "bin")
				if stat, err := os.Stat(versionBinPath); err == nil && stat.IsDir() {
					newPath = versionBinPath + string(os.PathListSeparator) + cleanPath
					if verbose {
						fmt.Printf("Added version bin path: %s\n", versionBinPath)
					}
				} else if verbose {
					fmt.Printf("Version bin path not found: %s\n", versionBinPath)
				}
			}

			if dryRun {
				// Just show what the new PATH would be
				fmt.Println(newPath)
				return nil
			}

			// Set the new PATH
			os.Setenv("PATH", newPath)

			if verbose {
				fmt.Printf("Updated PATH: %s\n", newPath)
				fmt.Println("Shell PATH updated successfully")
			} else {
				fmt.Printf("Updated shell PATH for Perl %s\n", info.Version)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().Bool("dry-run", false, "Show what the new PATH would be without setting it")
	cmd.Flags().Bool("verbose", false, "Show detailed information about PATH updates")

	return cmd
}

// cleanPVMFromPath removes PVM-managed perl directories from PATH
func cleanPVMFromPath(path string, verbose bool) (string, error) {
	if path == "" {
		return "", nil
	}

	// Get XDG directories to identify PVM paths
	dirs, err := xdg.GetDirs()
	if err != nil {
		return "", err
	}

	pvmVersionsPath := filepath.Join(dirs.DataDir, "versions")

	// Split PATH and filter out PVM directories
	pathEntries := strings.Split(path, string(os.PathListSeparator))
	var cleanEntries []string

	for _, entry := range pathEntries {
		// Skip empty entries
		if entry == "" {
			continue
		}

		// Check if this is a PVM perl bin directory
		if strings.Contains(entry, pvmVersionsPath) && strings.HasSuffix(entry, "/bin") {
			if verbose {
				fmt.Printf("Removing PVM path: %s\n", entry)
			}
			continue
		}

		cleanEntries = append(cleanEntries, entry)
	}

	return strings.Join(cleanEntries, string(os.PathListSeparator)), nil
}
