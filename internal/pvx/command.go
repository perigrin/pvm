// ABOUTME: PVX-specific commands and functionality
// ABOUTME: Implements commands for Perl execution

package pvx

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/config"
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
			// Get UI instance for styled output
			ui := cli.GetUI(cmd)
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
			envName, _ := cmd.Flags().GetString("name")

			// Get filesystem isolation flags
			readOnlyPaths, _ := cmd.Flags().GetStringArray("ro-path")
			readWritePaths, _ := cmd.Flags().GetStringArray("rw-path")
			isolatedOutput, _ := cmd.Flags().GetBool("isolated-output")
			saveOutputDir, _ := cmd.Flags().GetString("save-output-dir")

			// Get environment variable isolation flags
			preserveEnv, _ := cmd.Flags().GetStringArray("preserve-env")
			clearEnv, _ := cmd.Flags().GetStringArray("clear-env")

			// Get module path configuration flags
			includeModulePaths, _ := cmd.Flags().GetStringArray("include-path")
			customModulePath, _ := cmd.Flags().GetString("module-path")
			requiredModules, _ := cmd.Flags().GetStringArray("require")
			autoInstall, _ := cmd.Flags().GetBool("auto-install")
			autoDetectDeps, _ := cmd.Flags().GetBool("auto-detect-deps")
			additionalDeps, _ := cmd.Flags().GetStringArray("with")

			// Check if we're executing code directly or running a script
			if executeCode == "" && len(args) == 0 {
				_ = cmd.Help() // Ignoring error as we're exiting anyway
				return
			}

			// Load configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				log.Warnf("Failed to load configuration: %v", err)
				// Continue with default values
			}

			// Create execution options with defaults from configuration if available
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
				EnvName:      envName,
			}

			// Fill in values from configuration if available and not overridden by command line flags
			if cfg != nil && cfg.PVX != nil {
				// Set default isolation level from config if not specified on command line
				if isolationLevel == "" && !isolated && cfg.PVX.IsolationLevel != "" {
					options.IsolationLevel = IsolationLevel(cfg.PVX.IsolationLevel)
					if options.Verbose {
						log.Infof("Using isolation level '%s' from configuration", options.IsolationLevel)
					}
				}

				// Apply cleanup setting from config if not specified on command line
				if !cmd.Flags().Changed("no-cleanup") {
					options.NoCleanup = !cfg.PVX.CleanupAfter
					if options.Verbose {
						if options.NoCleanup {
							log.Infof("Using 'no-cleanup' setting from configuration")
						} else {
							log.Infof("Using 'cleanup' setting from configuration")
						}
					}
				}

				// Apply filesystem isolation settings from config if not specified on command line
				if len(readOnlyPaths) == 0 && len(cfg.PVX.IsolationReadOnlyPaths) > 0 {
					options.ReadOnlyPaths = cfg.PVX.IsolationReadOnlyPaths
					if options.Verbose {
						log.Infof("Using read-only paths from configuration: %v", options.ReadOnlyPaths)
					}
				} else {
					options.ReadOnlyPaths = readOnlyPaths
				}

				if len(readWritePaths) == 0 && len(cfg.PVX.IsolationReadWritePaths) > 0 {
					options.ReadWritePaths = cfg.PVX.IsolationReadWritePaths
					if options.Verbose {
						log.Infof("Using read-write paths from configuration: %v", options.ReadWritePaths)
					}
				} else {
					options.ReadWritePaths = readWritePaths
				}

				// Apply isolated output settings from config if not specified on command line
				if !cmd.Flags().Changed("isolated-output") {
					options.IsolatedOutput = cfg.PVX.IsolatedOutput
					if options.Verbose && options.IsolatedOutput {
						log.Infof("Using isolated output setting from configuration")
					}
				} else {
					options.IsolatedOutput = isolatedOutput
				}

				// Apply save output directory setting from config if not specified on command line
				if saveOutputDir == "" && cfg.PVX.SaveOutputDir != "" {
					options.SaveOutputDir = cfg.PVX.SaveOutputDir
					if options.Verbose {
						log.Infof("Using save output directory from configuration: %s", options.SaveOutputDir)
					}
				} else {
					options.SaveOutputDir = saveOutputDir
				}

				// Apply environment variable settings from config if not specified on command line
				if len(preserveEnv) == 0 && len(cfg.PVX.PreserveEnvVars) > 0 {
					options.PreserveEnv = cfg.PVX.PreserveEnvVars
					if options.Verbose {
						log.Infof("Using preserved environment variables from configuration: %v", options.PreserveEnv)
					}
				} else {
					options.PreserveEnv = preserveEnv
				}

				// Always apply clearEnv from command line as it's more specific
				options.ClearEnv = clearEnv

				// Apply module path settings from config if not specified on command line
				if len(includeModulePaths) == 0 && len(cfg.PVX.AdditionalModulePaths) > 0 {
					options.AdditionalModulePaths = cfg.PVX.AdditionalModulePaths
					if options.Verbose {
						log.Infof("Using additional module paths from configuration: %v", options.AdditionalModulePaths)
					}
				} else {
					options.AdditionalModulePaths = includeModulePaths
				}

				if customModulePath == "" && cfg.PVX.CustomModulePath != "" {
					options.CustomModulePath = cfg.PVX.CustomModulePath
					if options.Verbose {
						log.Infof("Using custom module path from configuration: %s", options.CustomModulePath)
					}
				} else {
					options.CustomModulePath = customModulePath
				}

				// Set required modules and auto-install flag
				options.RequiredModules = requiredModules
				options.RequiredModules = append(options.RequiredModules, additionalDeps...)
				options.AutoInstallModules = autoInstall
				options.AutoDetectDependencies = autoDetectDeps
			} else {
				// If no config or PVX config, use command line flags directly
				options.ReadOnlyPaths = readOnlyPaths
				options.ReadWritePaths = readWritePaths
				options.IsolatedOutput = isolatedOutput
				options.SaveOutputDir = saveOutputDir
				options.PreserveEnv = preserveEnv
				options.ClearEnv = clearEnv
				options.AdditionalModulePaths = includeModulePaths
				options.CustomModulePath = customModulePath
				options.RequiredModules = requiredModules
				options.RequiredModules = append(options.RequiredModules, additionalDeps...)
				options.AutoInstallModules = autoInstall
				options.AutoDetectDependencies = autoDetectDeps
			}

			// Set isolation level if provided
			if isolationLevel != "" {
				options.IsolationLevel = IsolationLevel(isolationLevel)
				if options.Verbose {
					log.Infof("Using isolation level '%s' from command line", options.IsolationLevel)
				}
			} else if isolated {
				// Backward compatibility: if --isolated flag is used without --isolation,
				// default to low isolation
				options.IsolationLevel = IsolationLow
				if options.Verbose {
					log.Infof("Using isolation level 'low' due to --isolated flag (legacy mode)")
				}
			}

			// Validate that high isolation features are only used with high isolation level
			if options.IsolationLevel != IsolationHigh &&
				(len(options.ReadOnlyPaths) > 0 || len(options.ReadWritePaths) > 0 || options.IsolatedOutput) {
				log.Warnf("Read-only paths, read-write paths, and isolated output are only fully effective with high isolation")
			}

			// Handle no-install flag
			if noInstall {
				options.AutoInstallModules = false
			}

			var output string

			switch {
			case executeCode != "":
				// Execute Perl code directly
				log.Debugf("Executing Perl code: %s", executeCode)
				options.InlineCode = executeCode
				output, err = ExecuteInlineCode(options, ui)
			case isToolName(args[0]):
				// Execute a tool directly (like uvx)
				toolName := args[0]
				toolArgs := []string{}
				if len(args) > 1 {
					toolArgs = args[1:]
				}

				log.Debugf("Executing tool: %s", toolName)
				output, err = ExecuteTool(options, toolName, toolArgs, ui)
			default:
				// Execute a script file
				scriptPath := args[0]
				scriptArgs := []string{}
				if len(args) > 1 {
					scriptArgs = args[1:]
				}

				options.ScriptPath = scriptPath
				options.Args = scriptArgs

				log.Debugf("Executing Perl script: %s", scriptPath)
				output, err = ExecuteScript(options, ui)
			}

			// If using isolated output and saveOutputDir is specified, copy generated files
			if options.IsolatedOutput && options.SaveOutputDir != "" && err == nil {
				if options.Verbose {
					ui.Status(fmt.Sprintf("Saving output files to %s", options.SaveOutputDir))
				}
				savedFiles, saveErr := saveOutputFiles(options, options.SaveOutputDir)
				if saveErr != nil {
					ui.Warning("Failed to save output files: %v", saveErr)
				} else if len(savedFiles) > 0 && options.Verbose {
					ui.Success("Saved %d files to %s", len(savedFiles), options.SaveOutputDir)
				}
			} else if options.IsolatedOutput && options.SaveOutputDir == "" && options.Verbose {
				ui.Info("Isolated output is enabled but no save directory was specified. Output files will be discarded after execution.")
			}

			// Print output regardless of error (may contain diagnostic information)
			if output != "" {
				ui.Printf("%s", output)
			}

			// Handle execution error
			if err != nil {
				ui.Error("Execution failed: %v", err)
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

	// Filesystem isolation flags
	cmd.Flags().StringArray("ro-path", []string{}, "Add a read-only path for filesystem isolation (high isolation only, can be specified multiple times)")
	cmd.Flags().StringArray("rw-path", []string{}, "Add a read-write path for filesystem isolation (high isolation only, can be specified multiple times)")
	cmd.Flags().Bool("isolated-output", false, "Create a temporary directory for script output (high isolation only)")
	cmd.Flags().String("save-output-dir", "", "Save generated files from isolated output to this directory")

	// Environment variable isolation flags
	cmd.Flags().StringArray("preserve-env", []string{}, "Preserve specific environment variables in isolation (can be specified multiple times)")
	cmd.Flags().StringArray("clear-env", []string{}, "Remove specific environment variables from all isolation levels (can be specified multiple times)")

	// Module path configuration flags
	cmd.Flags().StringArray("include-path", []string{}, "Add additional module paths to PERL5LIB (can be specified multiple times)")
	cmd.Flags().String("module-path", "", "Specify a custom module installation path")
	cmd.Flags().StringArray("require", []string{}, "Require specific modules and install if missing (can be specified multiple times)")
	cmd.Flags().Bool("auto-install", false, "Automatically install required modules using PVI")
	cmd.Flags().Bool("auto-detect-deps", true, "Automatically detect dependencies from use/require statements (default: true)")
	cmd.Flags().StringArray("with", []string{}, "Add additional dependencies beyond auto-detected ones (can be specified multiple times)")

	// Named environment flags
	cmd.Flags().String("name", "", "Create a named persistent isolation environment")

	return cmd
}

// isToolName determines if the given argument is a tool name rather than a script file
func isToolName(arg string) bool {
	// If it has a file extension or contains path separators, it's likely a script
	if strings.Contains(arg, "/") || strings.Contains(arg, "\\") || strings.Contains(arg, ".") {
		return false
	}

	// Check if it's a known Perl tool name
	knownTools := []string{
		"perl", "cpan", "prove", "perldoc", "h2ph", "h2xs", "enc2xs", "xsubpp",
		"corelist", "cpanm", "plackup", "carton", "dzil", "perlcritic", "perltidy",
	}

	for _, tool := range knownTools {
		if arg == tool {
			return true
		}
	}

	// If it doesn't look like a file path and we don't recognize it,
	// assume it's a tool name (auto-discovery)
	return true
}
