// ABOUTME: PSC run command implementation
// ABOUTME: Manages type checking and execution of Perl scripts

package psc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/errors"
)

// newRunCommand creates a command to type-check and execute a file
func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [file] [args...]",
		Short: "Type-check and execute a file",
		Long:  "Perform type checking and then execute the Perl file if no errors are found",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			isolate, _ := cmd.Flags().GetBool("isolate")
			modules, _ := cmd.Flags().GetStringArray("module")
			envVars, _ := cmd.Flags().GetStringArray("env")
			flowSensitive, _ := cmd.Flags().GetBool("flow-sensitive")
			noFlowSensitive, _ := cmd.Flags().GetBool("no-flow-sensitive")
			skipCleanup, _ := cmd.Flags().GetBool("no-cleanup")
			skipFlowChecks, _ := cmd.Flags().GetBool("skip-flow-checks")

			// Create environment variables map
			envMap := make(map[string]string)
			for _, env := range envVars {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					envMap[parts[0]] = parts[1]
				}
			}

			// Determine if flow-sensitive analysis should be enabled
			// If --no-flow-sensitive is specified, it overrides --flow-sensitive
			enableFlowSensitiveAnalysis := flowSensitive
			if noFlowSensitive {
				enableFlowSensitiveAnalysis = false
			}

			// Configure execution options
			options := &TypeCheckedExecutionOptions{
				ScriptPath:                  file,
				Args:                        scriptArgs,
				PerlVersion:                 perl,
				SkipTypeCheck:               skipCheck,
				DisableIsolation:            !isolate,
				EnableFlowSensitiveAnalysis: enableFlowSensitiveAnalysis,
				SkipFlowChecks:              skipFlowChecks,
				EnvironmentVars:             envMap,
				RequiredModules:             modules,
				Verbose:                     verbose,
				NoCleanup:                   skipCleanup,
			}

			// Execute the script with type checking
			result, err := ExecuteWithTypeChecking(options)
			if err != nil {
				// If it's a type checking error, we show the errors
				if result != nil && !result.TypeCheckPassed {
					// We've already shown the errors in the verbose mode in ExecuteWithTypeChecking
					if !verbose {
						// Just show the error count
						fmt.Printf("Found %d type errors in %s\n", len(result.TypeCheckErrors), file)
					}

					// Return the error
					return errors.NewTypeError(
						ErrTypeCheckFailed,
						"Type checking failed, aborting execution",
						nil)
				}

				// For other errors, return them directly
				return err
			}

			// If we made it here, the script executed successfully
			// The output has already been printed by ExecuteWithTypeChecking
			fmt.Println(result.Output)
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
