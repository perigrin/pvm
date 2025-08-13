// ABOUTME: PSC run command implementation
// ABOUTME: Manages type checking and execution of Perl scripts

package psc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/pvx"
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
			isolate, _ := cmd.Flags().GetBool("isolate")
			modules, _ := cmd.Flags().GetStringArray("module")
			envVars, _ := cmd.Flags().GetStringArray("env")
			noCleanup, _ := cmd.Flags().GetBool("no-cleanup")

			// Determine if flow-sensitive analysis should be enabled
			// If --no-flow-sensitive is specified, it overrides --flow-sensitive
			enableFlowSensitiveAnalysis := flowSensitive
			if noFlowSensitive {
				enableFlowSensitiveAnalysis = false
			}

			// Parse environment variables
			envMap := make(map[string]string)
			for _, envVar := range envVars {
				if parts := strings.SplitN(envVar, "=", 2); len(parts) == 2 {
					envMap[parts[0]] = parts[1]
				}
			}

			// Set isolation level based on flags
			isolationLevel := pvx.IsolationGlobal
			if isolate {
				isolationLevel = pvx.IsolationLocal
			}

			// Create PVX execution options with PSC-specific configuration
			options := &pvx.ExecutionOptions{
				ScriptPath:             file,
				Args:                   scriptArgs,
				PerlVersion:            perl,
				Env:                    envMap,
				Verbose:                verbose,
				TypeCheck:              true, // Always enable type checking for PSC
				SkipTypeCheck:          skipCheck,
				FlowSensitive:          enableFlowSensitiveAnalysis,
				SkipFlowChecks:         skipFlowChecks,
				IsolationLevel:         isolationLevel,
				NoCleanup:              noCleanup,
				RequiredModules:        modules,
				AutoInstallModules:     len(modules) > 0,
				AutoDetectDependencies: true, // Enable automatic dependency detection
			}

			// Execute via PVX with PSC-specific options
			output, err := pvx.ExecuteScript(options, ui)
			if err != nil {
				// Convert PVX errors to PSC-specific error types if needed
				if execErr, ok := err.(*errors.Error); ok && execErr.Code() == pvx.ErrExecutionFailed {
					return errors.NewTypeError(
						ErrExecutionFailed,
						fmt.Sprintf("Script execution failed: %s", execErr.Error()),
						execErr)
				}
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
