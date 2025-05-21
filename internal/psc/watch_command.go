// ABOUTME: PSC watch command implementation
// ABOUTME: Provides file watching for continuous type checking

package psc

import (
	"github.com/spf13/cobra"
)

// newWatchCommand creates a command to watch files for type errors
func newWatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch [file|dir]",
		Short: "Watch files and report type errors on change",
		Long:  "Continuously monitor files for changes and perform type checking",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement file watching functionality
			return nil
		},
	}

	cmd.Flags().StringArrayP("exclude", "e", []string{}, "Patterns to exclude from watching")

	return cmd
}