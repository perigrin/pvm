// ABOUTME: Core CLI framework for PVM Ecosystem
// ABOUTME: Provides base command structure used by all components

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/version"
)

var (
	// Verbose output flag
	Verbose bool
)

// NewRootCommand creates a new root command for a component
func NewRootCommand(name string, description string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: description,
		Long:  description,
	}

	// Add global flags
	cmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")
	
	// Add version command
	cmd.AddCommand(newVersionCommand(name))
	
	return cmd
}

// newVersionCommand creates a version command for the provided component
func newVersionCommand(component string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print detailed version information about this component",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.ComponentVersion(component))
		},
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}