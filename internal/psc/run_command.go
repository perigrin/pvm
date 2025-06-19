// ABOUTME: PSC run command implementation
// ABOUTME: Manages type checking and execution of Perl scripts

package psc

import (
	"fmt"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
)

// newRunCommand creates a command to type-check and execute a file
func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [file] [args...]",
		Short: "Type-check and execute a file",
		Long:  "Perform type checking and then execute the Perl file if no errors are found",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			if len(args) == 0 {
				return fmt.Errorf("expected a file to run")
			}

			// Get the file to run
			file := args[0]

			// Get the script arguments (if any)
			scriptArgs := []string{}
			if len(args) > 1 {
				scriptArgs = args[1:]
			}

			// Get flags
			verbose, _ := cmd.Flags().GetBool("verbose")
			skipCheck, _ := cmd.Flags().GetBool("skip-check")
			perl, _ := cmd.Flags().GetString("perl")
			flowSensitive, _ := cmd.Flags().GetBool("flow-sensitive")
			noFlowSensitive, _ := cmd.Flags().GetBool("no-flow-sensitive")
			skipFlowChecks, _ := cmd.Flags().GetBool("skip-flow-checks")

			// Environment variables are not used in simple strip-and-execute approach
			// This simplifies the command for basic use cases

			// Determine if flow-sensitive analysis should be enabled
			// If --no-flow-sensitive is specified, it overrides --flow-sensitive
			enableFlowSensitiveAnalysis := flowSensitive
			if noFlowSensitive {
				enableFlowSensitiveAnalysis = false
			}

			// If not skipping type check, perform type checking first
			if !skipCheck {
				// Perform type checking
				tc, err := parser.NewTypeCheck()
				if err != nil {
					return errors.NewTypeError(
						ErrTypeCheckFailed,
						"Failed to create type checker",
						err)
				}

				// Configure flow-sensitive analysis
				tc.EnableFlowSensitiveAnalysis = enableFlowSensitiveAnalysis
				tc.SkipFlowChecks = skipFlowChecks

				// Perform type checking
				checkResult, err := tc.CheckFile(file)
				if err != nil {
					return errors.NewTypeError(
						ErrTypeCheckFailed,
						fmt.Sprintf("Failed to check file: %s", file),
						err)
				}

				// Check if there were type errors
				if len(checkResult.Errors) > 0 {
					if verbose {
						ui.Error("Type checking failed for %s", file)
						for _, err := range checkResult.Errors {
							ui.Printf("  %s:%d:%d: %s\n", err.Path, err.Line, err.Column, err.Message)
						}
					} else {
						ui.Error("Found %d type errors in %s", len(checkResult.Errors), file)
					}

					return errors.NewTypeError(
						ErrTypeCheckFailed,
						"Type checking failed, aborting execution",
						nil)
				}

				if verbose {
					ui.Success("Type checking passed for %s", file)
				}
			}

			// Strip type annotations and execute
			output, err := StripAndExecute(file, scriptArgs, perl, verbose)
			if err != nil {
				return err
			}

			// Print the output
			ui.Printf("%s", output)
			return nil
		},
	}

	// Add command-specific flags
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().Bool("skip-check", false, "Skip type checking")
	cmd.Flags().StringP("perl", "p", "", "Use a specific Perl version")
	cmd.Flags().Bool("isolate", true, "Run in isolated environment")
	cmd.Flags().StringArrayP("module", "m", []string{}, "Additional modules to install")
	cmd.Flags().StringArrayP("env", "e", []string{}, "Environment variables (KEY=VALUE)")
	cmd.Flags().Bool("flow-sensitive", true, "Enable flow-sensitive analysis (default: true)")
	cmd.Flags().Bool("no-flow-sensitive", false, "Disable flow-sensitive analysis")
	cmd.Flags().Bool("no-cleanup", false, "Skip cleanup after execution")
	cmd.Flags().Bool("skip-flow-checks", false, "Skip flow-sensitive type checks but still perform refinements")

	return cmd
}
