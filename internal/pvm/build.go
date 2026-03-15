// ABOUTME: Build command stub for PVM
// ABOUTME: Returns "not yet available" since the build system requires type-system components

package pvm

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewBuildCommand creates a stub build command that reports it is unavailable
func NewBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [target]",
		Short: "Build Perl projects with type checking and compilation (not yet available)",
		Long:  `The build command requires type-system components that are not available in this build.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("build command is not yet available in this build")
		},
	}

	return cmd
}
