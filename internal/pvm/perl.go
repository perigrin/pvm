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