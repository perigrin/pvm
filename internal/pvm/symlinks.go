// ABOUTME: Symlinks command for PVM
// ABOUTME: Creates or verifies symlinks for the PVM Ecosystem

package pvm

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
)

// newSymlinksCommand creates a new symlinks command
func newSymlinksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "symlinks",
		Short: "Manage entry point symlinks",
		Long:  "Create or verify symlinks for the different entry points (pvm, pvx, pvi, psc)",
	}

	// Add subcommands
	cmd.AddCommand(
		newSymlinksCreateCommand(),
		newSymlinksVerifyCommand(),
	)

	return cmd
}

// newSymlinksCreateCommand creates a command for creating symlinks
func newSymlinksCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create symlinks for all entry points",
		Long:  "Create symlinks for all entry points (pvm, pvx, pvi, psc) in the same directory as the current binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the path of the current binary
			binaryPath, err := os.Executable()
			if err != nil {
				return cli.NewError(cli.PrefixSYS, cli.CategorySystem, "001",
					"Failed to get executable path", err)
			}

			// Create symlinks
			symlinks, err := cli.CreateSymlinks(binaryPath)
			if err != nil {
				return cli.NewError(cli.PrefixSYS, cli.CategorySystem, "002",
					"Failed to create symlinks", err)
			}

			// Print results
			ui := cli.GetUI(cmd)
			ui.Success("Created symlinks:")
			for component, path := range symlinks {
				ui.Info("  %s -> %s", component, path)
			}

			return nil
		},
	}
}

// newSymlinksVerifyCommand creates a command for verifying symlinks
func newSymlinksVerifyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify symlinks for all entry points",
		Long:  "Check if symlinks for all entry points (pvm, pvx, pvi, psc) exist",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the path of the current binary
			binaryPath, err := os.Executable()
			if err != nil {
				return cli.NewError(cli.PrefixSYS, cli.CategorySystem, "001",
					"Failed to get executable path", err)
			}

			// Get the directory of the binary
			dir := filepath.Dir(binaryPath)

			// Verify symlinks
			status := cli.VerifySymlinks(binaryPath)

			// Print results
			ui := cli.GetUI(cmd)
			ui.Info("Symlinks status in %s:", dir)
			ui.Println()

			allExists := true
			for component, exists := range status {
				status := "✅ Exists"
				if !exists {
					status = "❌ Missing"
					allExists = false
				}
				ui.Info("  %s: %s", component, status)
			}

			if !allExists {
				ui.Println()
				ui.Info("To create missing symlinks, run: pvm symlinks create")
			}

			return nil
		},
	}
}
