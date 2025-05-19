// ABOUTME: PVX-specific commands and functionality
// ABOUTME: Implements commands for Perl execution

package pvx

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/log"
)

// osExit allows for test mocking of os.Exit
var osExit = os.Exit

// NewCommand creates a new PVX command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pvx [options] script.pl [args...]",
		Short: "Perl Version eXecutor",
		Long:  "Executes Perl code in isolated environments",
		Run: func(cmd *cobra.Command, args []string) {
			// Get flags
			perlVersion, _ := cmd.Flags().GetString("perl")
			rootDir, _ := cmd.Flags().GetString("root")
			typeCheck, _ := cmd.Flags().GetBool("type-check")
			verbose, _ := cmd.Flags().GetBool("verbose")
			executeCode, _ := cmd.Flags().GetString("execute")
			forceVersion, _ := cmd.Flags().GetBool("force")
			noInstall, _ := cmd.Flags().GetBool("no-install")
			isolated, _ := cmd.Flags().GetBool("isolated")
			isolationDir, _ := cmd.Flags().GetString("isolation-dir")
			isolationLevel, _ := cmd.Flags().GetString("isolation")
			noCleanup, _ := cmd.Flags().GetBool("no-cleanup")

			// Check if we're executing code directly or running a script
			if executeCode == "" && len(args) == 0 {
				cmd.Help()
				return
			}

			// Create execution options
			options := &ExecutionOptions{
				Args:         args,
				PerlVersion:  perlVersion,
				RootDir:      rootDir,
				TypeCheck:    typeCheck,
				Verbose:      verbose,
				ForceVersion: forceVersion,
				Env:          make(map[string]string),
				Isolated:     isolated,
				IsolationDir: isolationDir,
				NoCleanup:    noCleanup,
			}

			// Set isolation level if provided
			if isolationLevel != "" {
				options.IsolationLevel = IsolationLevel(isolationLevel)
			} else if isolated {
				// Backward compatibility: if --isolated flag is used without --isolation,
				// default to low isolation
				options.IsolationLevel = IsolationLow
			}

			// Set the no-install environment variable if needed
			if noInstall {
				options.Env["PVM_NO_INSTALL"] = "1"
			}

			var output string
			var err error

			if executeCode != "" {
				// Execute Perl code directly
				log.Debugf("Executing Perl code: %s", executeCode)
				options.InlineCode = executeCode
				output, err = ExecuteInlineCode(options)
			} else {
				// Execute a script file
				scriptPath := args[0]
				scriptArgs := []string{}
				if len(args) > 1 {
					scriptArgs = args[1:]
				}

				options.ScriptPath = scriptPath
				options.Args = scriptArgs

				log.Debugf("Executing Perl script: %s", scriptPath)
				output, err = ExecuteScript(options)
			}

			// Print output regardless of error (may contain diagnostic information)
			if output != "" {
				fmt.Print(output)
			}

			// Handle execution error
			if err != nil {
				log.Errorf("Execution failed: %v", err)
				osExit(1)
			}
		},
	}

	// Add PVX-specific flags
	cmd.Flags().Bool("no-install", false, "Don't install missing modules")
	cmd.Flags().StringP("perl", "p", "", "Use a specific Perl version")
	cmd.Flags().StringP("execute", "e", "", "Execute Perl code directly")
	cmd.Flags().String("root", "", "Set environment root directory")
	cmd.Flags().Bool("type-check", false, "Enable type checking before execution")
	cmd.Flags().BoolP("verbose", "v", false, "Show additional output")
	cmd.Flags().BoolP("force", "f", false, "Force using specified Perl version")
	cmd.Flags().BoolP("isolated", "i", false, "Create an isolated environment for the script (deprecated, use --isolation=low instead)")
	cmd.Flags().String("isolation-dir", "", "Specify the directory to use for isolation (default: auto-generated temp dir)")
	cmd.Flags().String("isolation", "", "Set isolation level: none, low, medium, high")
	cmd.Flags().Bool("no-cleanup", false, "Keep isolation directory after execution (default: cleanup)")

	return cmd
}
