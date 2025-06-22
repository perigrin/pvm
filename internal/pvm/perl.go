// ABOUTME: PVM Perl detection and management commands
// ABOUTME: Provides commands for detecting and managing Perl installations

package pvm

import (
	"fmt"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/perl"
)

// newPerlCommand creates a new perl command
func newPerlCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perl",
		Short: "Manage Perl installations",
		Long:  "Detect, install, and manage Perl versions",
	}

	// Add subcommands
	cmd.AddCommand(
		newPerlSystemCommand(),
		newPerlImportSystemCommand(),
		// Additional commands like install, list, etc. will be added later
	)

	return cmd
}

// newPerlSystemCommand creates a command to show system Perl installations
func newPerlSystemCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "system",
		Short: "Show system Perl installations",
		Long:  "Detect and display information about system Perl installations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			perls, err := perl.DetectAllSystemPerls()
			if err != nil {
				ui.Warning("Detection warning: %v", err)
			}

			if len(perls) == 0 {
				ui.Info("No system Perl installations found.")
				return nil
			}

			// Create table data
			headers := []string{"PATH", "VERSION", "ARCHITECTURE", "PRIMARY"}
			rows := make([][]string, len(perls))
			for i, p := range perls {
				rows[i] = []string{
					p.Path,
					p.Version,
					p.Architecture,
					fmt.Sprintf("%v", p.IsPrimary),
				}
			}

			ui.Table(headers, rows)
			return nil
		},
	}
}

// newPerlImportSystemCommand creates a command for importing system Perl
func newPerlImportSystemCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import-system",
		Short: "Import system Perl",
		Long:  "Register the system Perl in PVM's version registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			// Detect system Perl
			systemPerl, err := perl.DetectSystemPerl()
			if err != nil {
				return err
			}

			ui.Info("Detected system Perl %s at %s", systemPerl.Version, systemPerl.Path)

			// Check if this version is already registered
			installed, err := perl.IsVersionInstalled(systemPerl.Version)
			if err != nil {
				return err
			}

			if installed {
				ui.Warning("System Perl %s is already registered with PVM.", systemPerl.Version)
				return nil
			}

			// Import the system Perl
			ui.Status("Importing system Perl into PVM registry...")
			err = perl.ImportSystemPerl()
			if err != nil {
				return err
			}

			ui.Success("Successfully imported system Perl %s.", systemPerl.Version)
			ui.Info("You can now use this version with PVM commands.")

			return nil
		},
	}
}
