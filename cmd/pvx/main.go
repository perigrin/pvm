// ABOUTME: Main entry point for PVX (Perl Version eXecutor)
// ABOUTME: Executes Perl code in isolated environments

package main

import (
	"os"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/pvx"
	"tamarou.com/pvm/internal/version"
)

func main() {
	// Create a PVX command directly
	pvxCmd := pvx.NewCommand()
	
	// Add global flags (these would normally be added by the CLI framework)
	pvxCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug mode")
	
	// Add version information
	pvxCmd.Long = pvxCmd.Long + "\n\nVersion: " + version.GetVersion()
	
	// Add version command
	pvxCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("PVX version:", version.GetVersion())
		},
	})
	
	// Execute the command
	if err := pvxCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
