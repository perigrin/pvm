// ABOUTME: Development environment command stub for PVM
// ABOUTME: Returns "not yet available" since the dev environment requires type-system components

package pvm

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newDevCommand creates a stub dev command that reports it is unavailable
func newDevCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Start development environment (not yet available)",
		Long:  `The dev command requires type-system components that are not available in this build.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("dev command is not yet available in this build")
		},
	}

	return cmd
}
