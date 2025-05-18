// ABOUTME: PVM Perl detection and management commands
// ABOUTME: Provides commands for detecting and managing Perl installations

package pvm

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
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
			perls, err := perl.DetectAllSystemPerls()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Warning:", err)
			}

			if len(perls) == 0 {
				fmt.Println("No system Perl installations found.")
				return nil
			}

			// Create a tabwriter for formatting
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PATH\tVERSION\tARCHITECTURE\tPRIMARY\t")
			for _, p := range perls {
				fmt.Fprintf(w, "%s\t%s\t%s\t%v\t\n", p.Path, p.Version, p.Architecture, p.IsPrimary)
			}
			w.Flush()

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
			// Detect system Perl
			systemPerl, err := perl.DetectSystemPerl()
			if err != nil {
				return err
			}

			cmd.Printf("Detected system Perl %s at %s\n", systemPerl.Version, systemPerl.Path)

			// Check if this version is already registered
			installed, err := perl.IsVersionInstalled(systemPerl.Version)
			if err != nil {
				return err
			}

			if installed {
				cmd.Printf("System Perl %s is already registered with PVM.\n", systemPerl.Version)
				return nil
			}

			// Import the system Perl
			cmd.Printf("Importing system Perl into PVM registry...\n")
			err = perl.ImportSystemPerl()
			if err != nil {
				return err
			}

			cmd.Printf("Successfully imported system Perl %s.\n", systemPerl.Version)
			cmd.Println("You can now use this version with PVM commands.")

			return nil
		},
	}
}
