// ABOUTME: PVM-specific commands and functionality
// ABOUTME: Implements commands for Perl version management

package pvm

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/current"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/psc"
	"tamarou.com/pvm/internal/pvi"
	"tamarou.com/pvm/internal/pvx"
	"tamarou.com/pvm/internal/tool"
	"tamarou.com/pvm/internal/tool/install"
	"tamarou.com/pvm/internal/tool/shim"
	"tamarou.com/pvm/internal/version"
	"tamarou.com/pvm/internal/xdg"
)

// init sets up package integrations
func init() {
	// Set up version checking integration between current and perl packages
	current.SetVersionChecker(perl.IsVersionInstalled)
}

// logCompileDetails logs build details with appropriate level based on content
func logCompileDetails(ui *ui.Output, details string) {
	switch {
	case strings.Contains(details, "ERROR") || strings.Contains(details, "error:"):
		ui.Error("%s", details)
	case strings.Contains(details, "WARNING") || strings.Contains(details, "warning:"):
		ui.Warning("%s", details)
	default:
		ui.Info("%s", details)
	}
}

// NewCommand creates a new PVM command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pvm",
		Short: "Perl Version Manager",
		Long:  "Manages Perl installations and versions",
	}

	// Add PVM-specific commands
	cmd.AddCommand(
		newInstallCommand(),
		newUseCommand(),
		newShUseCommand(),
		newShEnvActivateCommand(),
		newDetectVersionCommand(),
		newShellInitCommand(),
		newGlobalCommand(),
		newLocalCommand(),
		newCurrentCommand(), // Show current Perl version
		newVersionsCommand(),
		newListCommand(), // Alias for versions command for compatibility
		newAvailableCommand(),
		newDownloadCommand(),
		newUpdateCommand(),      // Self-updater functionality
		newAutoUpdateCommand(),  // Auto-update configuration and management
		NewBuildCommand(),       // Unified build system with PSC integration
		newBuildPerlCommand(),   // Build Perl from source (split from old build command)
		newInstallPerlCommand(), // Install Perl from build directories
		newRunCommand(),         // New unified run command (incorporates PVX)
		newModuleCommand(),      // New unified module command (incorporates PVI)
		newProjectCommand(),     // New project management command
		newDevCommand(),         // Development environment command
		newTestCommand(),        // Test execution command
		newExecCommand(),
		newUninstallCommand(),
		newImportSystemCommand(),
		newImportCommand(),
		newRehashCommand(),
		newResolveCommand(),
		newInitCommand(),
		newShellCommand(),
		newMCPCommand(),
		newEnvCommand(),
		createPerformanceCommand(),
		newToolCommand(),

		// Enhanced help system
		cli.CreateHelpCommand(), // Context-aware help with workflow suggestions

		// Shell aliases for convenience
		newPVXCommand(), // Shell alias to run command
		newPSCCommand(), // Shell alias to build command

		// These are implemented in their own files
		newSymlinksCommand(), // from symlinks.go
		newConfigCommand(),   // from config.go
		newPerlCommand(),     // from perl.go
		newVersionCommand(),  // from version.go

		// Hidden subcommands for help purposes only
		newHelpOnlyPVICommand(),
		newHelpOnlyPVXCommand(),
		newHelpOnlyPSCCommand(),
	)

	return cmd
}

// Placeholder commands, to be implemented later

func newInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [version]",
		Short: "Install a Perl version",
		Long:  "Download and install a specific version of Perl",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]

			// Get flags (same as build command)
			sourceFile, err := cmd.Flags().GetString("source")
			if err != nil {
				return err
			}

			installDir, err := cmd.Flags().GetString("prefix")
			if err != nil {
				return err
			}

			buildJobs, err := cmd.Flags().GetInt("jobs")
			if err != nil {
				return err
			}

			runTests, err := cmd.Flags().GetBool("test")
			if err != nil {
				return err
			}

			skipBuild, err := cmd.Flags().GetBool("skip-build")
			if err != nil {
				return err
			}

			// Get binary installation flags
			binaryOnly, err := cmd.Flags().GetBool("binary-only")
			if err != nil {
				return err
			}

			preferBinary, err := cmd.Flags().GetBool("prefer-binary")
			if err != nil {
				return err
			}

			forceSource, err := cmd.Flags().GetBool("force-source")
			if err != nil {
				return err
			}

			// Validate mutually exclusive flags
			if binaryOnly && forceSource {
				return fmt.Errorf("--binary-only and --force-source are mutually exclusive")
			}

			if skipBuild {
				ui := cli.GetUI(cmd)
				ui.Warning("Skip-build specified but no import functionality implemented yet.")
				ui.Info("This will be implemented in a future version.")
				return nil
			}

			// Check if binary installation is requested and available
			if !forceSource && (binaryOnly || preferBinary) {
				// Check if binary is available for this version and platform
				available, err := perl.CheckBinaryAvailability(version, "")

				// Handle error case first
				if err != nil {
					if binaryOnly {
						return fmt.Errorf("failed to check binary availability: %w", err)
					}
					// For prefer-binary, continue to source installation
					cmd.Printf("Warning: Failed to check binary availability, falling back to source: %v\n", err)
				} else {
					// No error, determine strategy based on availability and preference
					switch {
					case available:
						// Binary is available, attempt binary installation
						cmd.Printf("Installing Perl %s from pre-compiled binary...\n", version)

						// Create binary installation options
						binaryOptions := &perl.BinaryInstallOptions{
							Version:    version,
							Platform:   "", // Use default platform
							InstallDir: installDir,
							ProgressCallback: func(total, transferred int64, done bool) {
								// Simple progress reporting for binary download
								if total > 0 {
									percentage := float64(transferred) / float64(total) * 100
									width := 40
									completeChars := int(float64(width) * float64(transferred) / float64(total))

									progressBar := "["
									for i := 0; i < width; i++ {
										switch {
										case i < completeChars:
											progressBar += "="
										case i == completeChars:
											progressBar += ">"
										default:
											progressBar += " "
										}
									}
									progressBar += "]"

									fmt.Printf("\r%s %.1f%% (%d/%d bytes)                    ",
										progressBar, percentage, transferred, total)

									if done {
										fmt.Println()
									}
								}
							},
							Context: cmd.Context(),
						}

						// Attempt binary installation
						result, err := perl.InstallFromBinary(binaryOptions)
						if err != nil {
							if binaryOnly {
								return fmt.Errorf("binary installation failed: %w", err)
							}
							// For prefer-binary, fall back to source
							cmd.Printf("Binary installation failed, falling back to source: %v\n", err)
						} else {
							// Binary installation succeeded
							cmd.Printf("\nBinary installation completed successfully!\n")
							cmd.Printf("Perl %s installed at: %s\n", result.Version, result.InstallPath)
							cmd.Printf("Total installation time: %s\n", result.Duration.Round(time.Second))
							if result.FromCache {
								cmd.Println("Note: Installation was completed using cached binary")
							}
							return nil
						}
					case binaryOnly:
						return fmt.Errorf("binary for Perl %s is not available for your platform", version)
					default:
						// prefer-binary but not available, fall back to source
						cmd.Printf("Binary for Perl %s not available, falling back to source installation\n", version)
					}
				}
			}

			// Build Perl using our build functionality
			ui := cli.GetUI(cmd)
			ui.Info("Installing Perl %s...", version)

			// Create progress callback to display build progress
			var currentStage perl.BuildProgressStage
			progressCallback := func(stage perl.BuildProgressStage, details string, progress float64) {
				// Only print stage transition once
				if stage != currentStage {
					ui.Header(stage.String())
					currentStage = stage
				}

				// Print progress details
				if details != "" {
					// For compile and test stages, we get lots of output
					// Only print lines with errors or warnings, or important milestones
					if stage == perl.StageCompile || stage == perl.StageTest {
						if strings.Contains(details, "ERROR") ||
							strings.Contains(details, "WARNING") ||
							strings.Contains(details, "warning:") ||
							strings.Contains(details, "error:") ||
							strings.Contains(details, "Done") ||
							strings.Contains(details, "All tests successful") {
							logCompileDetails(ui, details)
						}
					} else {
						// For other stages, print all details
						ui.Info("%s", details)
					}
				}

				// If we have numeric progress, show a progress bar
				if progress > 0 && stage < perl.StageDone {
					width := 40
					completeChars := int(float64(width) * progress)

					// Format progress bar
					progressBar := "["
					for i := 0; i < width; i++ {
						switch {
						case i < completeChars:
							progressBar += "="
						case i == completeChars:
							progressBar += ">"
						default:
							progressBar += " "
						}
					}
					progressBar += "]"

					// Clear line and show progress
					ui := cli.GetUI(cmd)
					ui.Printf("\r%s %.1f%%                    ",
						progressBar, progress*100)

					if progress >= 1.0 {
						ui.Println()
					}
				}
			}

			// Create build options
			options := &perl.BuildOptions{
				Version:          version,
				SourceFile:       sourceFile,
				InstallDir:       installDir,
				BuildJobs:        buildJobs,
				RunTests:         runTests,
				CleanupBuildDir:  true, // Always clean up for install command
				ProgressCallback: progressCallback,
				Context:          cmd.Context(),
			}

			// Start the build
			result, err := perl.BuildPerl(options)
			if err != nil {
				ui.Error("Installation failed: %v", err)
				return err
			}

			// Show build results
			ui.Success("Installation completed successfully!")
			ui.Info("Perl %s installed at: %s", result.Version, result.InstallPath)
			ui.Info("Total installation time: %s", result.Duration.Round(time.Second))

			// The registration is now handled automatically in the BuildPerl function

			return nil
		},
	}

	// Add flags
	cmd.Flags().String("source", "", "Source archive file path (default: download or use cached)")
	cmd.Flags().String("prefix", "", "Installation directory (default: XDG_DATA_HOME/pvm/versions/<version>)")
	cmd.Flags().Int("jobs", 0, "Number of parallel build jobs (default: number of CPU cores)")
	cmd.Flags().Bool("test", false, "Run Perl tests after building")
	cmd.Flags().Bool("skip-build", false, "Skip build and import from existing installation")

	// Binary installation flags
	cmd.Flags().Bool("binary-only", false, "Install only from pre-compiled binary (fail if not available)")
	cmd.Flags().BoolP("prefer-binary", "B", false, "Try binary first, fallback to source if binary unavailable")
	cmd.Flags().Bool("force-source", false, "Force source compilation (skip binary check)")

	return cmd
}

func newUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use [version]",
		Short: "Use a specific version in the current shell",
		Long:  "Temporarily use a specific Perl version in the current shell session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("The 'pvm use' command requires shell integration to work properly.")
			cmd.Println("Please ensure you have run 'eval \"$(pvm init)\"' in your shell.")
			cmd.Println("The shell integration provides a 'pvm' function that handles version switching.")
			return nil
		},
	}
}

func newShUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "sh-use [version]",
		Short:  "Generate shell code to use a specific Perl version",
		Long:   "Outputs shell commands to set environment variables for a specific Perl version",
		Args:   cobra.ExactArgs(1),
		Hidden: true, // Hide from help output as this is internal
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			return perl.GenerateShellUse(version)
		},
	}
}

func newShEnvActivateCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "sh-env-activate [name]",
		Short:  "Generate shell code to activate a named environment",
		Long:   "Outputs shell commands to activate a named isolation environment",
		Args:   cobra.ExactArgs(1),
		Hidden: true, // Hide from help output as this is internal
		RunE: func(cmd *cobra.Command, args []string) error {
			envName := args[0]
			return generateShellEnvActivate(envName)
		},
	}
}

func newDetectVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "detect-version",
		Short: "Find .perl-version file in current directory tree",
		Long:  "Search for .perl-version file starting from current directory and walking up the directory tree",
		RunE: func(cmd *cobra.Command, args []string) error {
			return detectVersionFile(cmd)
		},
	}
}

func newShellInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "shell-init [shell]",
		Short: "Generate shell integration code",
		Long:  "Generate shell integration code for auto-switching and shell functions (alias for 'pvm init')",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var shell perl.ShellType
			var err error

			if len(args) > 0 {
				// Parse shell type from argument
				switch args[0] {
				case "bash":
					shell = perl.ShellBash
				case "zsh":
					shell = perl.ShellZsh
				case "fish":
					shell = perl.ShellFish
				case "powershell":
					shell = perl.ShellPowerShell
				case "cmd":
					shell = perl.ShellCmd
				default:
					return fmt.Errorf("unsupported shell: %s", args[0])
				}
			} else {
				// Detect shell type automatically
				shell, err = perl.DetectShell()
				if err != nil {
					return err
				}
			}

			// Get shell script for the specified/detected shell
			script, err := perl.GetCurrentShellScript(shell)
			if err != nil {
				return err
			}

			// Print the script to stdout (for eval)
			fmt.Print(script)
			return nil
		},
	}
}

func newGlobalCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "global [version]",
		Short: "Set the global Perl version",
		Long:  "Set the default Perl version for all shells",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]

			// Set global version
			err := perl.SetGlobalVersion(version)
			if err != nil {
				return err
			}

			// Success message
			ui := cli.GetUI(cmd)
			ui.Success("Global Perl version set to %s", version)
			ui.Info("This is now the default version when no other version is specified")

			return nil
		},
	}
}

func newLocalCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "local [version]",
		Short: "Set the local version for a directory",
		Long:  "Set the Perl version for the current directory and subdirectories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]

			// Set local version
			err := perl.SetLocalVersion(version)
			if err != nil {
				return err
			}

			// Get current directory for the message
			dir, err := os.Getwd()
			if err != nil {
				dir = "current directory"
			}

			// Success message
			ui := cli.GetUI(cmd)
			ui.Success("Local Perl version set to %s for %s", version, dir)
			ui.Info("This version will be used in this directory and its subdirectories")
			ui.Info("Note: Shell integration must be set up for automatic switching to work")

			return nil
		},
	}
}

// newCurrentCommand creates a command to show the currently active Perl version
func newCurrentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show currently active Perl version",
		Long: `Show the currently active Perl version with source attribution.

This command shows which Perl version is currently being used and where
that setting comes from (e.g., .perl-version file, environment variable,
user configuration, etc.).

The version resolution follows this precedence order:
1. Explicitly specified version
2. Project-local .perl-version file
3. Project-local .pvm/pvm.toml
4. Environment variables (PLENV_VERSION, PERLBREW_PERL)
5. User-level configuration
6. System Perl

Examples:
  pvm current              # Show current version with source
  pvm current --bare       # Show only version (for scripting)
  pvm current --detailed   # Show comprehensive information
  pvm current --json       # Output in JSON format`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get flags
			bare, err := cmd.Flags().GetBool("bare")
			if err != nil {
				return err
			}

			detailed, err := cmd.Flags().GetBool("detailed")
			if err != nil {
				return err
			}

			jsonOutput, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			showPath, err := cmd.Flags().GetBool("path")
			if err != nil {
				return err
			}

			validate, err := cmd.Flags().GetBool("validate")
			if err != nil {
				return err
			}

			// Determine display options based on flags
			var options *current.DisplayOptions
			switch {
			case bare:
				options = current.BareDisplayOptions()
			case detailed:
				options = current.DetailedDisplayOptions()
			case jsonOutput:
				options = current.DefaultDisplayOptions()
				options.Format = current.FormatJSON
			default:
				options = current.DefaultDisplayOptions()
			}

			// Apply additional flag overrides
			if showPath {
				options.ShowPath = true
			}
			if validate {
				options.Validate = true
				options.ShowComparison = true
			}

			// Get current version information
			info, err := current.GetCurrentVersion()
			if err != nil {
				return fmt.Errorf("failed to get current version: %w", err)
			}

			// Format and display the output
			output, err := current.FormatCurrentVersion(info, options)
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			cmd.Print(output)

			// Add newline for non-bare output
			if !bare {
				cmd.Println()
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().Bool("bare", false, "Show only the version string (for scripting)")
	cmd.Flags().Bool("detailed", false, "Show comprehensive version information")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("path", false, "Include file paths in output")
	cmd.Flags().Bool("validate", false, "Validate current version and show warnings")

	return cmd
}

func newVersionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions",
		Short: "List installed versions",
		Long:  "List all installed Perl versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the installed versions
			installedVersions, err := perl.GetInstalledVersions()
			if err != nil {
				return err
			}

			ui := cli.GetUI(cmd)

			// Check if we have any installed versions
			if len(installedVersions) == 0 {
				ui.Warning("No versions installed. Use 'pvm install <version>' to install a version.")
				return nil
			}

			// Get flags
			showPath, err := cmd.Flags().GetBool("paths")
			if err != nil {
				return err
			}

			showSource, err := cmd.Flags().GetBool("source")
			if err != nil {
				return err
			}

			// Display installed versions
			ui.Header("Installed Perl versions:")
			for i, versionInfo := range installedVersions {
				// Add decoration for special versions
				decoration := ""
				if i == 0 {
					decoration = " (latest)"
				}

				// Basic output
				if !showPath && !showSource {
					// Add import source indicator for imported versions
					var sourceIndicator string
					switch versionInfo.Source {
					case "plenv":
						sourceIndicator = " (imported from plenv)"
					case "perlbrew":
						sourceIndicator = " (imported from perlbrew)"
					case "system":
						sourceIndicator = " (system)"
					}
					ui.Printf("  %s%s%s", versionInfo.Version, decoration, sourceIndicator)
					continue
				}

				// Detailed output
				ui.Printf("  %s%s", versionInfo.Version, decoration)

				if showPath {
					ui.Info("    Path: %s", versionInfo.InstallPath)
				}

				if showSource {
					ui.Info("    Source: %s", versionInfo.Source)
					ui.Info("    Installed: %s", versionInfo.InstallTime.Format("2006-01-02 15:04:05"))
				}

				// Add a separator between versions for detailed output
				if i < len(installedVersions)-1 && (showPath || showSource) {
					ui.Println("")
				}
			}

			// Add a hint about current/active version
			ui.Info("Note: Use 'pvm current' to show the currently active Perl version.")

			return nil
		},
	}

	// Add flags
	cmd.Flags().Bool("paths", false, "Show installation paths")
	cmd.Flags().Bool("source", false, "Show source and installation time")

	return cmd
}

func newListCommand() *cobra.Command {
	// Create an alias to the versions command for compatibility with tests
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed versions (alias for versions)",
		Long:  "List all installed Perl versions (alias for versions command)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the installed versions
			installedVersions, err := perl.GetInstalledVersions()
			if err != nil {
				return err
			}

			ui := cli.GetUI(cmd)

			// Check if we have any installed versions
			if len(installedVersions) == 0 {
				ui.Warning("No versions installed. Use 'pvm install <version>' to install a version.")
				return nil
			}

			// Get flags
			showPath, err := cmd.Flags().GetBool("paths")
			if err != nil {
				return err
			}

			showSource, err := cmd.Flags().GetBool("source")
			if err != nil {
				return err
			}

			// Display installed versions
			ui.Header("Installed Perl versions:")
			for i, versionInfo := range installedVersions {
				// Add decoration for special versions
				decoration := ""
				if i == 0 {
					decoration = " (latest)"
				}

				// Basic output
				if !showPath && !showSource {
					// Add import source indicator for imported versions
					var sourceIndicator string
					switch versionInfo.Source {
					case "plenv":
						sourceIndicator = " (imported from plenv)"
					case "perlbrew":
						sourceIndicator = " (imported from perlbrew)"
					case "system":
						sourceIndicator = " (system)"
					}
					ui.Printf("  %s%s%s", versionInfo.Version, decoration, sourceIndicator)
					continue
				}

				// Detailed output
				ui.Printf("  %s%s", versionInfo.Version, decoration)

				if showPath {
					ui.Info("    Path: %s", versionInfo.InstallPath)
				}

				if showSource {
					ui.Info("    Source: %s", versionInfo.Source)
					ui.Info("    Installed: %s", versionInfo.InstallTime.Format("2006-01-02 15:04:05"))
				}

				// Add a separator between versions for detailed output
				if i < len(installedVersions)-1 && (showPath || showSource) {
					ui.Println("")
				}
			}

			// Add a hint about current/active version
			ui.Info("Note: Use 'pvm current' to show the currently active Perl version.")

			return nil
		},
	}

	// Add flags
	cmd.Flags().Bool("paths", false, "Show installation paths")
	cmd.Flags().Bool("source", false, "Show source and installation time")

	return cmd
}

func newAvailableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "available",
		Short: "List available Perl versions",
		Long:  "List all Perl versions available for installation",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the currently installed versions
			installedVersions, err := perl.GetInstalledVersions()
			if err != nil {
				return err
			}

			// Create a map for fast lookup of installed versions
			installedMap := make(map[string]bool)
			for _, info := range installedVersions {
				installedMap[info.Version] = true
			}

			// Define common stable Perl versions (5.10.0 through latest)
			// These would ideally come from a remote API in the future
			stableVersions := []string{
				"5.38.0", "5.36.0", "5.34.1", "5.34.0",
				"5.32.1", "5.32.0", "5.30.3", "5.30.2", "5.30.1", "5.30.0",
				"5.28.3", "5.28.2", "5.28.1", "5.28.0", "5.26.3", "5.26.2", "5.26.1", "5.26.0",
				"5.24.4", "5.24.3", "5.24.2", "5.24.1", "5.24.0", "5.22.4", "5.22.3", "5.22.2", "5.22.1", "5.22.0",
				"5.20.3", "5.20.2", "5.20.1", "5.20.0", "5.18.4", "5.18.3", "5.18.2", "5.18.1", "5.18.0",
				"5.16.3", "5.16.2", "5.16.1", "5.16.0", "5.14.4", "5.14.3", "5.14.2", "5.14.1", "5.14.0",
				"5.12.5", "5.12.4", "5.12.3", "5.12.2", "5.12.1", "5.12.0", "5.10.1", "5.10.0",
			}

			// Current development version
			devVersions := []string{"5.39.0"}

			// Display header
			ui := cli.GetUI(cmd)
			ui.Header("Available Perl versions:")

			// Group versions by major.minor
			groupedVersions := make(map[string][]string)

			// Add stable versions to groups
			for _, version := range stableVersions {
				// Parse version to get major.minor
				parsedVersion, err := perl.ParseVersion(version)
				if err != nil {
					continue // Skip invalid versions
				}

				// Create group key (e.g., "5.38")
				groupKey := fmt.Sprintf("%d.%d", parsedVersion.Major, parsedVersion.Minor)

				// Add to group
				groupedVersions[groupKey] = append(groupedVersions[groupKey], version)
			}

			// Add development versions separately
			ui.SubHeader("Development versions:")
			for _, version := range devVersions {
				installed := ""
				if installedMap[version] {
					installed = " (installed)"
				}
				ui.Printf("  %s%s", version, installed)
			}

			// Display stable versions by group
			ui.SubHeader("Stable versions:")

			// Sort groups (we could use a proper version sorting here)
			// But for this simple listing, the natural string sort will work reasonably well
			groupKeys := make([]string, 0, len(groupedVersions))
			for key := range groupedVersions {
				groupKeys = append(groupKeys, key)
			}
			// Note: This would be better with a custom sort

			// Display groups
			for _, groupKey := range groupKeys {
				versions := groupedVersions[groupKey]

				// Skip empty groups
				if len(versions) == 0 {
					continue
				}

				// Display group versions
				ui.Info("  %s series:", groupKey)
				for _, version := range versions {
					installed := ""
					if installedMap[version] {
						installed = " (installed)"
					}
					ui.Printf("    %s%s", version, installed)
				}
			}

			ui.Info("Use 'pvm install <version>' to install a specific version.")
			return nil
		},
	}
	return cmd
}

func newDownloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download [version]",
		Short: "Download Perl source",
		Long:  "Download Perl source code archive for a specific version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]

			// Get mirror from flags or config
			mirror, err := cmd.Flags().GetString("mirror")
			if err != nil {
				return err
			}

			// Get skip-cache flag
			skipCache, err := cmd.Flags().GetBool("skip-cache")
			if err != nil {
				return err
			}

			// Get skip-checksum flag
			skipChecksum, err := cmd.Flags().GetBool("skip-checksum")
			if err != nil {
				return err
			}

			// Create progress bar display
			progressCallback := func(total, transferred int64, done bool) {
				// Only show progress for downloads with known size
				if total > 0 {
					percentage := float64(transferred) / float64(total) * 100

					// Create a simple progress bar
					width := 40
					completeChars := int(float64(width) * float64(transferred) / float64(total))

					// Format progress bar
					progressBar := "["
					for i := 0; i < width; i++ {
						switch {
						case i < completeChars:
							progressBar += "="
						case i == completeChars:
							progressBar += ">"
						default:
							progressBar += " "
						}
					}
					progressBar += "]"

					// Clear line and show progress
					ui := cli.GetUI(cmd)
					ui.Printf("\r%s %.1f%% (%d/%d bytes)                    ",
						progressBar, percentage, transferred, total)

					if done {
						ui.Println()
					}
				}
			}

			// Create download options
			options := &perl.DownloadOptions{
				Mirror:           mirror,
				Version:          version,
				ProgressCallback: progressCallback,
				SkipCache:        skipCache,
				SkipChecksum:     skipChecksum,
				Context:          cmd.Context(),
			}

			// Start the download
			ui := cli.GetUI(cmd)
			ui.Info("Downloading Perl %s...", version)

			if mirror != "" {
				ui.Info("Using mirror: %s", mirror)
			}

			// Perform the download
			result, err := perl.DownloadPerlSource(options)
			if err != nil {
				return err
			}

			// Show download results
			if result.FromCache {
				ui.Success("Loaded from cache: %s", result.Path)
			} else {
				ui.Success("Download complete: %s", result.Path)
				ui.Info("Download size: %d bytes", result.Size)
				ui.Info("Download time: %s", result.Duration.Round(time.Millisecond))
			}

			if result.Checksum != "" {
				ui.Info("SHA-256: %s", result.Checksum)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().String("mirror", "", "Mirror URL to use for downloading (default: "+perl.DefaultMirror+")")
	cmd.Flags().Bool("skip-cache", false, "Skip using cached downloads")
	cmd.Flags().Bool("skip-checksum", false, "Skip checksum validation")

	return cmd
}

func newExecCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "exec [version] [command]",
		Short: "Execute a command with a specific version",
		Long:  "Execute a command using a specific Perl version",
		Run: func(cmd *cobra.Command, args []string) {
			ui := cli.GetUI(cmd)
			if len(args) < 2 {
				ui.Error("Usage: pvm exec [version] [command]")
				return
			}

			version := args[0]
			command := args[1:]

			err := execCommand(cmd, version, command)
			if err != nil {
				ui.Error("Error: %v", err)
				os.Exit(1)
			}
		},
	}
}

func newUninstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall [version]",
		Short: "Remove a Perl version",
		Long:  "Uninstall a specific version of Perl",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]

			// Check if the version is installed
			versionInfo, err := perl.GetVersionInfo(version)
			if err != nil {
				return err
			}

			// Get force flag
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}

			// Get confirmation unless force flag is used
			ui := cli.GetUI(cmd)
			if !force {
				ui.Printf("Are you sure you want to uninstall Perl %s? [y/N] ", version)
				var response string
				_, _ = fmt.Scanln(&response)

				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					ui.Info("Uninstall cancelled.")
					return nil
				}
			}

			// If this is a system Perl, warn the user
			if versionInfo.Source == "system" {
				ui.Warning("Note: This is a system Perl installation.")
				ui.Info("The installation will be unregistered from PVM but the actual files will not be removed.")
			}

			// Perform the uninstallation
			ui.Info("Uninstalling Perl %s...", version)
			err = perl.UninstallVersion(version)
			if err != nil {
				return err
			}

			ui.Success("Perl %s has been uninstalled.", version)
			return nil
		},
	}

	// Add flags
	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

func newImportSystemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-system",
		Short: "Import system Perl installation",
		Long:  "Import the system Perl installation into PVM registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return importSystemPerl(cmd)
		},
	}
	cmd.Flags().Bool("force", false, "Force re-import even if already registered")
	return cmd
}

func newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-from [tool]",
		Short: "Import Perl installations from other tools",
		Long:  "Import Perl installations from other version managers (plenv, perlbrew)",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "plenv",
			Short: "Import from plenv",
			Long:  "Import Perl installations from plenv",
			RunE: func(cmd *cobra.Command, args []string) error {
				return importFromLegacyTool(cmd, perl.Plenv)
			},
		},
		&cobra.Command{
			Use:   "perlbrew",
			Short: "Import from perlbrew",
			Long:  "Import Perl installations from perlbrew",
			RunE: func(cmd *cobra.Command, args []string) error {
				return importFromLegacyTool(cmd, perl.Perlbrew)
			},
		},
	)

	return cmd
}

// importFromLegacyTool implements the logic for importing from a legacy tool
func importFromLegacyTool(cmd *cobra.Command, tool perl.LegacyToolType) error {
	ui := cli.GetUI(cmd)
	ui.Info("Detecting %s installations...", tool)

	installations, err := perl.ImportFromLegacyTool(tool)
	if err != nil {
		return err
	}

	ui.Success("Found %d %s installation(s):", len(installations), tool)
	for i, inst := range installations {
		defaultMark := ""
		if inst.IsDefault {
			defaultMark = " (default)"
		}
		ui.Printf("%d. %s%s at %s", i+1, inst.Version, defaultMark, inst.Path)
	}

	// In a real implementation, here we would:
	// 1. Ask the user which installations to import
	// 2. Create symlinks or copy the installations
	// 3. Register them in PVM's internal database
	// But for now, we'll just report what was found

	ui.Warning("Note: This command currently only detects installations.")
	ui.Info("Actual import functionality will be implemented in a future version.")

	// If it's perlbrew, also show aliases
	if tool == perl.Perlbrew {
		aliases, err := perl.GetPerlbrewAliases()
		if err == nil && len(aliases) > 0 {
			ui.SubHeader("Perlbrew aliases detected:")
			for alias, target := range aliases {
				ui.Printf("  %s -> %s", alias, target)
			}
		}
	}

	return nil
}

func newRehashCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rehash",
		Short: "Rebuild shim executables",
		Long:  "Rebuild shim executables for all installed Perl versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)
			ui.Info("Rebuilding shim executables...")

			// Call rehash function
			err := perl.Rehash()
			if err != nil {
				return err
			}

			// Check if PATH is configured correctly
			pathConfigured, shimDir, err := perl.CheckPath()
			if err != nil {
				return err
			}

			// Show success message
			ui.Success("Shim executables rebuilt successfully.")

			// Warn if PATH is not configured
			if !pathConfigured {
				ui.Warning("The shim directory is not in your PATH.")
				ui.Printf("To use pvm, add the following directory to your PATH:\n%s\n", shimDir)

				// Get the command to add to PATH
				pathCmd, err := perl.GetPathConfigCommand()
				if err == nil {
					ui.Info("You can do this with the following command:")
					ui.Printf("%s", pathCmd)
				}
			}

			return nil
		},
	}

	return cmd
}

func newResolveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve [version]",
		Short: "Resolve a Perl version",
		Long:  "Resolve a Perl version based on the version resolution algorithm",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)
			var explicitVersion string
			if len(args) > 0 {
				explicitVersion = args[0]
			}

			options := &perl.ResolutionOptions{
				ExplicitVersion: explicitVersion,
			}

			// Set up callback to print resolution process
			perl.OnVersionResolved = func(version *perl.ResolvedVersion) {
				path := version.Path
				if path == "" {
					path = "N/A"
				}
				ui.Success("Resolved version: %s", version.Version)
				ui.Info("Source: %s", version.Source)
				ui.Info("Path: %s", path)
			}

			resolved, err := perl.ResolveVersion(options)
			if err != nil {
				return err
			}

			// OnVersionResolved callback will print the details
			_ = resolved // Avoid unused variable warning
			return nil
		},
	}
}

// newPVXCommand creates a PVX command as a shell alias to 'pvm run'
func newPVXCommand() *cobra.Command {
	// Get the PVX command from the PVX package
	pvxCmd := pvx.NewCommand()

	// Customize as shell alias
	pvxCmd.Use = "pvx"
	pvxCmd.Short = "Perl Version eXecutor (shell alias for 'pvm run')"
	pvxCmd.Long = "Executes Perl code in isolated environments. This is a shell alias for 'pvm run' - you can use either 'pvx' or 'pvm run' interchangeably."

	// Add version subcommand
	pvxCmd.AddCommand(newComponentVersionCommand("pvx"))

	return pvxCmd
}

// isPerlVersion checks if a string looks like a Perl version
func isPerlVersion(s string) bool {
	// Simple heuristic: if it matches X.Y.Z format, it's likely a Perl version
	if strings.Count(s, ".") >= 1 {
		parts := strings.Split(s, ".")
		if len(parts) >= 2 {
			// Check if first part is numeric and looks like a major version
			if len(parts[0]) > 0 && parts[0][0] >= '0' && parts[0][0] <= '9' {
				return true
			}
		}
	}
	return false
}

// buildPerlFromSource handles building Perl from source
func buildPerlFromSource(cmd *cobra.Command, version string) error {
	if version == "" {
		return fmt.Errorf("version required for Perl source build")
	}

	// Get flags
	sourceFile, err := cmd.Flags().GetString("source")
	if err != nil {
		return err
	}

	installDir, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	outputDir, err := cmd.Flags().GetString("output-dir")
	if err != nil {
		return err
	}

	buildJobs, err := cmd.Flags().GetInt("jobs")
	if err != nil {
		return err
	}

	runTests, err := cmd.Flags().GetBool("test")
	if err != nil {
		return err
	}

	cleanupBuildDir, err := cmd.Flags().GetBool("cleanup")
	if err != nil {
		return err
	}

	configureOptions, err := cmd.Flags().GetStringArray("configure-options")
	if err != nil {
		return err
	}

	buildOnly, err := cmd.Flags().GetBool("build-only")
	if err != nil {
		return err
	}

	// Create progress callback to display build progress
	ui := cli.GetUI(cmd)
	var currentStage perl.BuildProgressStage
	progressCallback := func(stage perl.BuildProgressStage, details string, progress float64) {
		// Only print stage transition once
		if stage != currentStage {
			ui.Header(stage.String())
			currentStage = stage
		}

		// Print progress details
		if details != "" {
			// For compile and test stages, we get lots of output
			// Only print lines with errors or warnings, or important milestones
			if stage == perl.StageCompile || stage == perl.StageTest {
				if strings.Contains(details, "ERROR") ||
					strings.Contains(details, "WARNING") ||
					strings.Contains(details, "warning:") ||
					strings.Contains(details, "error:") ||
					strings.Contains(details, "Done") ||
					strings.Contains(details, "All tests successful") {
					logCompileDetails(ui, details)
				}
			} else {
				// For other stages, print all details
				ui.Info("%s", details)
			}
		}

		// If we have numeric progress, show a progress bar
		if progress > 0 && stage < perl.StageDone {
			width := 40
			completeChars := int(float64(width) * progress)

			// Format progress bar
			progressBar := "["
			for i := 0; i < width; i++ {
				switch {
				case i < completeChars:
					progressBar += "="
				case i == completeChars:
					progressBar += ">"
				default:
					progressBar += " "
				}
			}
			progressBar += "]"

			// Clear line and show progress
			ui := cli.GetUI(cmd)
			ui.Printf("\r%s %.1f%%                    ",
				progressBar, progress*100)

			if progress >= 1.0 {
				ui.Println()
			}
		}
	}

	// Determine final installation directory
	// If --output-dir is specified, it takes precedence over --prefix
	finalInstallDir := installDir
	if outputDir != "" {
		finalInstallDir = outputDir
	}

	// Create build options
	options := &perl.BuildOptions{
		Version:          version,
		SourceFile:       sourceFile,
		InstallDir:       finalInstallDir,
		BuildJobs:        buildJobs,
		RunTests:         runTests,
		CleanupBuildDir:  cleanupBuildDir,
		ConfigureOptions: configureOptions,
		BuildOnly:        buildOnly,
		ProgressCallback: progressCallback,
		Context:          cmd.Context(),
	}

	// Print build information
	ui.Info("Building Perl %s from source...", version)

	if sourceFile != "" {
		ui.Info("Using source file: %s", sourceFile)
	}

	if outputDir != "" {
		ui.Info("Output directory: %s", outputDir)
	} else if installDir != "" {
		ui.Info("Installation directory: %s", installDir)
	}

	if buildJobs > 0 {
		ui.Info("Using %d parallel build jobs", buildJobs)
	} else {
		ui.Info("Using default number of build jobs")
	}

	if runTests {
		ui.Info("Will run tests after building")
	}

	if buildOnly {
		ui.Info("Build-only mode: Perl will be built without installation")
	}

	if cleanupBuildDir {
		if buildOnly {
			ui.Info("Will clean up build directory after build")
		} else {
			ui.Info("Will clean up build directory after installation")
		}
	}

	// Start the build
	result, err := perl.BuildPerl(options)
	if err != nil {
		ui.Error("Build failed: %v", err)
		return err
	}

	// Show build results
	if buildOnly {
		ui.Success("Build completed successfully!")
		ui.Info("Perl %s built at: %s", result.Version, result.InstallPath)
	} else {
		ui.Success("Build completed successfully!")
		ui.Info("Perl %s installed at: %s", result.Version, result.InstallPath)
	}
	ui.Info("Total build time: %s", result.Duration.Round(time.Second))

	// Show timing for each stage
	ui.SubHeader("Build stage timing:")
	for stage, duration := range result.Stages {
		ui.Printf("  %-12s: %s", stage.String(), duration.Round(time.Second))
	}

	return nil
}

// installPerlFromBuild installs Perl from a build directory
func installPerlFromBuild(cmd *cobra.Command, args []string) error {
	// Get flags
	buildDir, err := cmd.Flags().GetString("from-build")
	if err != nil {
		return err
	}

	versionOverride, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	// Determine source directory
	var sourceDir string
	switch {
	case buildDir != "":
		sourceDir = buildDir
	case len(args) > 0:
		sourceDir = args[0]
	default:
		return fmt.Errorf("no build directory specified - use --from-build flag or provide directory as argument")
	}

	// Get UI for progress reporting
	ui := cli.GetUI(cmd)

	ui.Info("Installing Perl from build directory: %s", sourceDir)

	// Validate the build directory contains a complete Perl installation
	ui.Info("Validating build directory...")
	valid, warnings, err := perl.ValidateBinaryInstallation(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to validate build directory: %w", err)
	}

	if !valid {
		return fmt.Errorf("build directory does not contain a valid Perl installation")
	}

	// Print any warnings
	for _, warning := range warnings {
		ui.Warning("%s", warning)
	}

	// Detect version from the Perl executable
	perlPath := filepath.Join(sourceDir, "bin", "perl")
	if runtime.GOOS == "windows" {
		perlPath = filepath.Join(sourceDir, "bin", "perl.exe")
	}

	detectedVersion := versionOverride
	if detectedVersion == "" {
		ui.Info("Detecting Perl version...")
		cmd := exec.Command(perlPath, "-e", "print $^V")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to detect Perl version: %w", err)
		}

		versionStr := strings.TrimSpace(string(output))
		// Convert from v5.38.0 format to 5.38.0
		if strings.HasPrefix(versionStr, "v") {
			versionStr = versionStr[1:]
		}
		detectedVersion = versionStr
	}

	ui.Info("Detected Perl version: %s", detectedVersion)

	// Check if version already exists
	if !force {
		installed, err := perl.IsVersionInstalled(detectedVersion)
		if err != nil {
			return fmt.Errorf("failed to check if version is installed: %w", err)
		}
		if installed {
			return fmt.Errorf("version %s is already installed - use --force to override", detectedVersion)
		}
	}

	// Determine installation directory
	dirs, err := xdg.GetDirs()
	if err != nil {
		return fmt.Errorf("failed to get XDG directories: %w", err)
	}

	installDir := filepath.Join(dirs.DataDir, "versions", detectedVersion)

	ui.Info("Installing to: %s", installDir)

	// Create installation directory
	err = os.MkdirAll(installDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Copy the build directory to the installation directory
	ui.Info("Copying Perl installation...")
	err = copyDirectory(sourceDir, installDir)
	if err != nil {
		return fmt.Errorf("failed to copy Perl installation: %w", err)
	}

	// Register the version in PVM registry
	ui.Info("Registering version with PVM...")
	versionInfo := perl.VersionInfo{
		Version:     detectedVersion,
		InstallPath: installDir,
		InstallTime: time.Now(),
		Source:      "install-perl",
	}

	err = perl.RegisterVersion(versionInfo)
	if err != nil {
		// Clean up on registration failure
		os.RemoveAll(installDir)
		return fmt.Errorf("failed to register version: %w", err)
	}

	ui.Success("Successfully installed Perl %s", detectedVersion)
	ui.Info("Installation path: %s", installDir)

	return nil
}

// copyDirectory recursively copies a directory
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// Copy file
			return copyFile(path, dstPath, info.Mode())
		}
	})
}

// copyFile copies a single file with permissions
func copyFile(src, dst string, mode os.FileMode) error {
	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(dst)
	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		return err
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Set permissions
	return os.Chmod(dst, mode)
}

// buildProject functionality moved to internal/pvm/build.go

// newSymlinksCommand is implemented in symlinks.go

// newConfigCommand is implemented in config.go

// newSystemCommand creates a command for showing system Perl info
// This is now moved to perl.go as newPerlSystemCommand()

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize shell integration",
		Long:  "Generate shell integration script for the current shell",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Detect shell type
			shell, err := perl.DetectShell()
			if err != nil {
				return err
			}

			// Check if --generate flag is provided
			generate, err := cmd.Flags().GetBool("generate")
			if err != nil {
				return err
			}

			if generate {
				// Generate and create shell initialization scripts
				err = perl.CreateShellInitScripts()
				if err != nil {
					return err
				}
				cmd.Println("Shell initialization scripts generated successfully.")
				return nil
			}

			// Check if we should perform automatic import
			// Only do this if we have no existing versions (first run)
			if perl.ShouldAutoImport() {
				// Perform automatic import of legacy tools
				results, err := perl.AutoImportLegacyVersions()
				if err == nil && results.TotalImported > 0 {
					// Print import results to stderr so it doesn't interfere with shell eval
					perl.PrintAutoImportResults(results)
				}
			}

			// Get shell script for the detected shell
			script, err := perl.GetCurrentShellScript(shell)
			if err != nil {
				return err
			}

			// Print the script to stdout (for eval)
			ui := cli.GetUI(cmd)
			ui.Printf("%s", script)
			return nil
		},
	}

	// Add flags
	cmd.Flags().Bool("generate", false, "Generate shell initialization scripts instead of printing them")

	return cmd
}

func newShellCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Shell integration utilities",
		Long:  "Commands for shell integration and management",
	}

	// Add subcommands
	cmd.AddCommand(
		&cobra.Command{
			Use:   "init",
			Short: "Initialize shell integration",
			Long:  "Generate shell integration script for the current shell",
			RunE: func(cmd *cobra.Command, args []string) error {
				// Generate and create shell initialization scripts
				err := perl.CreateShellInitScripts()
				if err != nil {
					return err
				}
				ui := cli.GetUI(cmd)
				ui.Success("Shell integration initialized")
				return nil
			},
		},
		&cobra.Command{
			Use:   "setup",
			Short: "Setup shell integration",
			Long:  "Show instructions for setting up shell integration",
			RunE: func(cmd *cobra.Command, args []string) error {
				// Detect shell type
				shell, err := perl.DetectShell()
				if err != nil {
					return err
				}

				// Check if shell is already initialized
				initialized, instruction, err := perl.CheckShellInit(shell)
				if err != nil {
					// If error checking initialization, just show instructions
					cmd.Println("Could not check if shell is initialized.")
					cmd.Println("Here are the instructions for setting up shell integration:")
					cmd.Println(perl.GetShellInitInstructions(shell))
					return nil
				}

				// Show appropriate message based on initialization status
				if initialized {
					cmd.Println("Shell integration is already set up.")
				} else {
					cmd.Println("Shell integration is not set up.")
					cmd.Println("Here's how to set it up:")
					cmd.Println(instruction)
					cmd.Println("\nDetailed instructions:")
					cmd.Println(perl.GetShellInitInstructions(shell))
				}

				return nil
			},
		},
	)

	return cmd
}

// newBuildPerlCommand creates a command for building Perl from source (moved from old build command)
func newBuildPerlCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build-perl [version]",
		Short: "Build Perl from source",
		Long:  "Build and install Perl from source code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			return buildPerlFromSource(cmd, version)
		},
	}

	// Perl source build flags
	cmd.Flags().String("source", "", "Source archive file path (default: download or use cached)")
	cmd.Flags().String("prefix", "", "Installation directory (default: XDG_DATA_HOME/pvm/versions/<version>)")
	cmd.Flags().String("output-dir", "", "Build output directory (default: uses prefix for installation)")
	cmd.Flags().Int("jobs", 0, "Number of parallel build jobs (default: number of CPU cores)")
	cmd.Flags().Bool("test", false, "Run Perl tests after building")
	cmd.Flags().Bool("cleanup", true, "Clean up build directory after installation")
	cmd.Flags().Bool("build-only", false, "Build Perl without installing (creates relocatable build in output directory)")
	cmd.Flags().StringArray("configure-options", nil, "Additional options to pass to Configure (can be specified multiple times)")

	return cmd
}

// newInstallPerlCommand creates a new install-perl command
func newInstallPerlCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install-perl",
		Short: "Install Perl from build directory",
		Long:  "Install Perl from a previously built directory into PVM's version management system",
		RunE:  installPerlFromBuild,
	}

	// Installation source flags
	cmd.Flags().String("from-build", "", "Install from build directory")
	cmd.Flags().String("version", "", "Override version detection (optional)")
	cmd.Flags().Bool("force", false, "Force installation even if version already exists")

	return cmd
}

// newModuleCommand creates a unified module management command
func newModuleCommand() *cobra.Command {
	// Get the PVI command from the PVI package
	pviCmd := pvi.NewCommand()

	// Customize for integration with PVM
	pviCmd.Use = "module"
	pviCmd.Short = "Module management"
	pviCmd.Long = "Manage CPAN modules and dependencies"

	return pviCmd
}

// newRunCommand creates a unified run/execution command
func newRunCommand() *cobra.Command {
	// Get the PVX command from the PVX package
	pvxCmd := pvx.NewCommand()

	// Customize for integration with PVM
	pvxCmd.Use = "run"
	pvxCmd.Short = "Execute Perl code"
	pvxCmd.Long = "Execute Perl code in isolated environments"

	return pvxCmd
}

// newPSCCommand creates a PSC command as a shell alias to 'pvm build'
func newPSCCommand() *cobra.Command {
	// Get the PSC command from the PSC package
	pscCmd := psc.NewCommand()

	// Customize as shell alias
	pscCmd.Use = "psc"
	pscCmd.Short = "Perl Script Compiler (shell alias for 'pvm build')"
	pscCmd.Long = "Provides static type checking for Perl code with type annotations. This is a shell alias for 'pvm build' - you can use either 'psc' or 'pvm build' interchangeably."

	// Add version subcommand
	pscCmd.AddCommand(newComponentVersionCommand("psc"))

	return pscCmd
}

// execCommand implements the exec command functionality
func execCommand(cmd *cobra.Command, version string, command []string) error {
	// Resolve the version to get the Perl path
	options := &perl.ResolutionOptions{
		ExplicitVersion: version,
	}

	resolved, err := perl.ResolveVersion(options)
	if err != nil {
		return fmt.Errorf("failed to resolve Perl version %s: %w", version, err)
	}

	// Check if the resolved version has a path
	if resolved.Path == "" {
		return fmt.Errorf("no path found for Perl version %s", resolved.Version)
	}

	// Check if the Perl binary exists
	if _, err := os.Stat(resolved.Path); os.IsNotExist(err) {
		return fmt.Errorf("perl version %s not found at %s", resolved.Version, resolved.Path)
	}

	// Create a new environment with the Perl bin directory in PATH
	env := os.Environ()
	perlBinDir := strings.TrimSuffix(resolved.Path, "/perl")
	if perlBinDir == resolved.Path {
		// Handle case where path doesn't end with /perl
		perlBinDir = resolved.Path[:strings.LastIndex(resolved.Path, "/")]
	}

	// Update PATH to include the Perl bin directory at the beginning
	for i, envVar := range env {
		if strings.HasPrefix(envVar, "PATH=") {
			currentPath := envVar[5:] // Remove "PATH=" prefix
			env[i] = fmt.Sprintf("PATH=%s:%s", perlBinDir, currentPath)
			break
		}
	}

	// Create the command to execute
	// If the first argument is "perl", replace it with the resolved Perl path
	if command[0] == "perl" {
		command[0] = resolved.Path
	}

	// Execute the command
	execCmd := exec.Command(command[0], command[1:]...)
	execCmd.Env = env
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// Run the command and wait for completion
	err = execCmd.Run()
	if err != nil {
		// Check if it's an exit error to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// importSystemPerl imports the system Perl installation into PVM registry
func importSystemPerl(cmd *cobra.Command) error {
	// Detect the system Perl
	systemPerl, err := perl.DetectSystemPerl()
	if err != nil {
		return fmt.Errorf("failed to detect system Perl: %w", err)
	}

	cmd.Printf("Found system Perl %s at %s\n", systemPerl.Version, systemPerl.Path)

	// Check force flag
	force, _ := cmd.Flags().GetBool("force")

	// Check if this version is already registered
	installed, err := perl.IsVersionInstalled(systemPerl.Version)
	if err != nil {
		return fmt.Errorf("failed to check if version is installed: %w", err)
	}
	if installed && !force {
		cmd.Printf("System Perl %s is already registered with PVM\n", systemPerl.Version)
		return nil
	}

	// If forcing and already installed, uninstall first
	if installed && force {
		cmd.Printf("Force re-importing system Perl %s\n", systemPerl.Version)
		err = perl.UninstallVersion(systemPerl.Version)
		if err != nil {
			return fmt.Errorf("failed to uninstall existing registration: %w", err)
		}
	}

	// Register the system Perl using the corrected install path logic
	// For system perl, InstallPath should be the directory containing the perl executable
	installPath := filepath.Dir(systemPerl.Path)
	versionInfo := perl.VersionInfo{
		Version:     systemPerl.Version,
		InstallPath: installPath,
		InstallTime: time.Now(),
		Source:      "system",
	}

	err = perl.RegisterVersion(versionInfo)
	if err != nil {
		return fmt.Errorf("failed to register system Perl: %w", err)
	}

	cmd.Printf("Successfully imported system Perl %s\n", systemPerl.Version)
	return nil
}

// newEnvCommand creates environment management commands
func newEnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage isolation environments",
		Long:  "Commands for managing named isolation environments",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List all named environments",
			Long:  "List all named isolation environments",
			RunE: func(cmd *cobra.Command, args []string) error {
				return listEnvironments(cmd)
			},
		},
		&cobra.Command{
			Use:   "activate [name]",
			Short: "Activate a named environment in current shell",
			Long:  "Activate a named isolation environment in the current shell session",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.Println("The 'pvm env activate' command requires shell integration to work properly.")
				cmd.Println("Please ensure you have run 'eval \"$(pvm init)\"' in your shell.")
				cmd.Println("The shell integration provides a 'pvm' function that handles environment activation.")
				return nil
			},
		},
		&cobra.Command{
			Use:   "remove [name]",
			Short: "Remove a named environment",
			Long:  "Remove a named isolation environment and all its files",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return removeEnvironment(cmd, args[0])
			},
		},
		&cobra.Command{
			Use:   "info [name]",
			Short: "Show environment information",
			Long:  "Show detailed information about a named environment",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return showEnvironmentInfo(cmd, args[0])
			},
		},
	)

	return cmd
}

// listEnvironments lists all named isolation environments
func listEnvironments(cmd *cobra.Command) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	envDir := filepath.Join(dirs.DataDir, "environments")

	// Check if environments directory exists
	if _, err := os.Stat(envDir); os.IsNotExist(err) {
		cmd.Println("No environments found.")
		return nil
	}

	// Read environments directory
	entries, err := os.ReadDir(envDir)
	if err != nil {
		return fmt.Errorf("failed to read environments directory: %w", err)
	}

	if len(entries) == 0 {
		cmd.Println("No environments found.")
		return nil
	}

	cmd.Println("Named isolation environments:")
	for _, entry := range entries {
		if entry.IsDir() {
			envPath := filepath.Join(envDir, entry.Name())
			info, err := os.Stat(envPath)
			if err == nil {
				cmd.Printf("  %s (created: %s)\n", entry.Name(), info.ModTime().Format("2006-01-02 15:04:05"))
			} else {
				cmd.Printf("  %s\n", entry.Name())
			}
		}
	}

	cmd.Printf("\nTo activate an environment, add its bin directory to your PATH:\n")
	cmd.Printf("export PATH=\"%s/<env-name>/bin:$PATH\"\n", envDir)

	return nil
}

// activateEnvironment outputs shell command to activate an environment
func activateEnvironment(cmd *cobra.Command, envName string) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	envPath := filepath.Join(dirs.DataDir, "environments", envName)

	// Check if environment exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return fmt.Errorf("environment '%s' does not exist", envName)
	}

	binDir := filepath.Join(envPath, "bin")

	// Output the activation command
	fmt.Printf("export PATH=\"%s:$PATH\"\n", binDir)
	fmt.Printf("# Environment '%s' activated. Use 'deactivate' or start a new shell to deactivate.\n", envName)

	return nil
}

// detectVersionFile finds a .perl-version file by walking up the directory tree
func detectVersionFile(cmd *cobra.Command) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	versionFile, err := findVersionFile(wd)
	if err != nil {
		return err
	}

	if versionFile == "" {
		return fmt.Errorf("no .perl-version file found in current directory tree")
	}

	cmd.Println(versionFile)
	return nil
}

// findVersionFile walks up the directory tree looking for .perl-version file
func findVersionFile(startDir string) (string, error) {
	dir := startDir

	for {
		versionFile := filepath.Join(dir, ".perl-version")
		if _, err := os.Stat(versionFile); err == nil {
			return versionFile, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", nil
}

// generateShellEnvActivate outputs shell commands to activate a named environment
func generateShellEnvActivate(envName string) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	envPath := filepath.Join(dirs.DataDir, "environments", envName)

	// Check if environment exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return fmt.Errorf("environment '%s' does not exist", envName)
	}

	binDir := filepath.Join(envPath, "bin")

	// Detect shell type from environment
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	// Extract shell name from full path
	shellName := filepath.Base(shell)

	// Generate shell-specific environment commands
	switch shellName {
	case "fish":
		fmt.Printf("set -gx PATH %s $PATH\n", binDir)
		fmt.Printf("echo \"Environment '%s' activated\"\n", envName)
	default: // bash, zsh, sh
		fmt.Printf("export PATH=\"%s:$PATH\"\n", binDir)
		fmt.Printf("echo \"Environment '%s' activated\"\n", envName)
	}

	return nil
}

// removeEnvironment removes a named isolation environment
func removeEnvironment(cmd *cobra.Command, envName string) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	envPath := filepath.Join(dirs.DataDir, "environments", envName)

	// Check if environment exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return fmt.Errorf("environment '%s' does not exist", envName)
	}

	// Confirm removal
	cmd.Printf("Are you sure you want to remove environment '%s'? [y/N] ", envName)
	var response string
	_, _ = fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		cmd.Println("Environment removal cancelled.")
		return nil
	}

	// Remove the environment directory
	err = os.RemoveAll(envPath)
	if err != nil {
		return fmt.Errorf("failed to remove environment '%s': %w", envName, err)
	}

	cmd.Printf("Environment '%s' has been removed.\n", envName)
	return nil
}

// showEnvironmentInfo shows detailed information about an environment
func showEnvironmentInfo(cmd *cobra.Command, envName string) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	envPath := filepath.Join(dirs.DataDir, "environments", envName)

	// Check if environment exists
	info, err := os.Stat(envPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("environment '%s' does not exist", envName)
	}

	cmd.Printf("Environment: %s\n", envName)
	cmd.Printf("Path: %s\n", envPath)
	cmd.Printf("Created: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))

	// Check for bin directory and shims
	binDir := filepath.Join(envPath, "bin")
	if binInfo, err := os.Stat(binDir); err == nil && binInfo.IsDir() {
		cmd.Printf("Bin directory: %s\n", binDir)

		// List shims
		shims, err := os.ReadDir(binDir)
		if err == nil && len(shims) > 0 {
			cmd.Printf("Available commands:\n")
			for _, shim := range shims {
				if !shim.IsDir() {
					cmd.Printf("  %s\n", shim.Name())
				}
			}
		}
	}

	// Check for lib directory
	libDir := filepath.Join(envPath, "lib")
	if libInfo, err := os.Stat(libDir); err == nil && libInfo.IsDir() {
		cmd.Printf("Library directory: %s\n", libDir)
	}

	cmd.Printf("\nTo activate this environment:\n")
	cmd.Printf("export PATH=\"%s:$PATH\"\n", binDir)

	return nil
}

// newToolCommand creates tool management commands
func newToolCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool",
		Short: "Manage tool installations",
		Long:  "Commands for installing, running, and managing Perl tools globally and per-project",
	}

	// Add --global flag to the parent command
	cmd.PersistentFlags().Bool("global", false, "Operate on global tools instead of project tools")

	installCmd := &cobra.Command{
		Use:   "install [tool[@version]]",
		Short: "Install a tool",
		Long:  "Install a Perl tool (module) and make it available as a command. Use --global for system-wide installation.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			return installTool(cmd, args[0], global)
		},
	}

	runCmd := &cobra.Command{
		Use:   "run [tool] [args...]",
		Short: "Run a tool",
		Long:  "Run a Perl tool without installing it permanently. Automatically detects global tools.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTool(cmd, args[0], args[1:])
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed tools",
		Long:  "List all installed tools and their versions. Use --global to show only global tools.",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			return listTools(cmd, global)
		},
	}

	upgradeCmd := &cobra.Command{
		Use:   "upgrade [tool]",
		Short: "Upgrade a tool",
		Long:  "Upgrade an installed tool to the latest version. Use --global for global tools.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			return upgradeTool(cmd, args[0], global)
		},
	}

	uninstallCmd := &cobra.Command{
		Use:   "uninstall [tool]",
		Short: "Uninstall a tool",
		Long:  "Remove an installed tool. Use --global for global tools.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			return uninstallTool(cmd, args[0], global)
		},
	}

	// Add shim management commands
	shimCmd := &cobra.Command{
		Use:   "shim",
		Short: "Manage PATH shims for tools",
		Long:  "Commands for managing PATH shims that make global tools available as direct commands",
	}

	shimInstallCmd := &cobra.Command{
		Use:   "install",
		Short: "Install shell integration for shims",
		Long:  "Add shim directory to your shell's PATH configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			shell, _ := cmd.Flags().GetString("shell")
			position, _ := cmd.Flags().GetString("position")
			return installShimIntegration(cmd, shell, position, force)
		},
	}
	shimInstallCmd.Flags().Bool("force", false, "Force reinstallation even if already installed")
	shimInstallCmd.Flags().String("shell", "", "Target shell (auto-detected if not specified)")
	shimInstallCmd.Flags().String("position", "after-system", "Position in PATH: first, last, after-system")

	shimRemoveCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove shell integration for shims",
		Long:  "Remove shim directory from your shell's PATH configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			shell, _ := cmd.Flags().GetString("shell")
			return removeShimIntegration(cmd, shell)
		},
	}
	shimRemoveCmd.Flags().String("shell", "", "Target shell (auto-detected if not specified)")

	shimListCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed shims",
		Long:  "Show all tools that have shims available in PATH",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listShims(cmd)
		},
	}

	shimStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show shim integration status",
		Long:  "Display information about shim PATH integration and conflicts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return shimStatus(cmd)
		},
	}

	shimCmd.AddCommand(shimInstallCmd, shimRemoveCmd, shimListCmd, shimStatusCmd)

	cmd.AddCommand(installCmd, runCmd, listCmd, upgradeCmd, uninstallCmd, shimCmd)

	return cmd
}

// installTool installs a tool and creates a shim for it
func installTool(cmd *cobra.Command, toolSpec string, global bool) error {
	// Parse tool specification (tool@version or just tool)
	parts := strings.Split(toolSpec, "@")
	toolName := parts[0]
	var version string
	if len(parts) > 1 {
		version = parts[1]
	}

	if global {
		// Use global tool installation infrastructure
		cmd.Printf("Installing global tool '%s'", toolName)
		if version != "" {
			cmd.Printf(" (version %s)", version)
		}
		cmd.Println("...")

		// Initialize tool mapping to resolve tool name to module
		mapping := tool.NewToolMapping()

		// Resolve tool to module name
		resolution, err := mapping.ResolveToolToModule(toolName)
		if err != nil {
			return fmt.Errorf("failed to resolve tool '%s' to module: %w", toolName, err)
		}

		// Create installation options
		options := &install.InstallOptions{
			ToolName:          toolName,
			ModuleName:        resolution.ModuleName,
			VersionConstraint: version,
			Force:             false,
			Verbose:           true,
			Context:           cmd.Context(),
		}

		// Use global installer
		installer, err := install.NewToolInstaller()
		if err != nil {
			return fmt.Errorf("failed to create installer: %w", err)
		}

		result, err := installer.InstallTool(options)
		if err != nil {
			return fmt.Errorf("failed to install global tool '%s': %w", toolName, err)
		}

		cmd.Printf("Successfully installed global tool '%s' -> %s\n", toolName, result.InstallPath)

		// Create shim for the tool
		shimManager, err := shim.NewManager()
		if err != nil {
			cmd.Printf("Warning: Failed to create shim manager: %v\n", err)
		} else {
			toolInfo := &tool.ToolInfo{
				Name:        toolName,
				Module:      resolution.ModuleName,
				Version:     result.Version,
				Description: resolution.Description,
				InstallDate: time.Now(),
			}

			if err := shimManager.CreateShim(toolName, toolInfo); err != nil {
				cmd.Printf("Warning: Failed to create shim for '%s': %v\n", toolName, err)
			} else {
				cmd.Printf("Created shim in PATH for command '%s'\n", toolName)
			}
		}

		return nil
	}

	// Legacy project-local tool installation (keep existing behavior for backward compatibility)
	cmd.Printf("Installing project tool '%s'", toolName)
	if version != "" {
		cmd.Printf(" (version %s)", version)
	}
	cmd.Println("...")

	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	// Create tools directory
	toolsDir := filepath.Join(dirs.DataDir, "tools")
	if err := os.MkdirAll(toolsDir, 0755); err != nil {
		return fmt.Errorf("failed to create tools directory: %w", err)
	}

	// Create tool-specific environment
	toolEnvDir := filepath.Join(toolsDir, toolName)
	if version != "" {
		toolEnvDir = filepath.Join(toolsDir, fmt.Sprintf("%s-%s", toolName, version))
	}

	// Use PVX to install the tool in an isolated environment
	options := &pvx.ExecutionOptions{
		PerlVersion:    "", // Use default
		IsolationLevel: pvx.IsolationLocal,
		IsolationDir:   toolEnvDir,
		EnvName:        fmt.Sprintf("tool-%s", toolName),
		NoCleanup:      true,
	}

	// Create installation script
	var installScript string
	if version != "" {
		installScript = fmt.Sprintf(`
use App::cpanminus;
exec { 'cpanm' } 'cpanm', '%s@%s';
`, toolName, version)
	} else {
		installScript = fmt.Sprintf(`
use App::cpanminus;
exec { 'cpanm' } 'cpanm', '%s';
`, toolName)
	}

	// Create a temporary script file
	tmpDir := os.TempDir()
	scriptPath := filepath.Join(tmpDir, fmt.Sprintf("install-%s-%d.pl", toolName, os.Getpid()))

	err = os.WriteFile(scriptPath, []byte(installScript), 0644)
	if err != nil {
		return fmt.Errorf("failed to create installation script: %w", err)
	}
	defer os.Remove(scriptPath)

	// Set the script path
	options.ScriptPath = scriptPath

	// Execute installation
	_, err = pvx.ExecuteScript(options)
	if err != nil {
		return fmt.Errorf("failed to install tool '%s': %w", toolName, err)
	}

	// Create shim for the tool
	err = createToolShim(toolName, toolEnvDir)
	if err != nil {
		return fmt.Errorf("failed to create shim for tool '%s': %w", toolName, err)
	}

	cmd.Printf("Tool '%s' installed successfully\n", toolName)
	cmd.Printf("Shim created in tools directory\n")

	return nil
}

// runTool runs a tool temporarily without installing it
func runTool(cmd *cobra.Command, toolName string, toolArgs []string) error {
	// Use tool detector to determine how to handle this tool
	detector := tool.NewDetector()

	// Check if this is a known tool or should be treated as a script
	mode, err := detector.DetectExecutionMode(append([]string{toolName}, toolArgs...))
	if err != nil {
		return fmt.Errorf("failed to detect execution mode for '%s': %w", toolName, err)
	}

	if mode.Mode == tool.ModeTool {
		// Initialize tool mapping to resolve tool name to module
		mapping := tool.NewToolMapping()

		// Try to resolve tool to module name
		module, err := mapping.ResolveToolToModule(toolName)
		if err == nil {
			cmd.Printf("Running tool '%s' (module: %s)...\n", toolName, module)
		} else {
			cmd.Printf("Running tool '%s'...\n", toolName)
		}
	}

	// Use PVX to execute the tool directly
	options := &pvx.ExecutionOptions{
		PerlVersion:    "", // Use default
		IsolationLevel: pvx.IsolationLocal,
		NoCleanup:      false, // Clean up for temporary runs
	}

	output, err := pvx.ExecuteTool(options, toolName, toolArgs)
	if err != nil {
		return fmt.Errorf("failed to run tool '%s': %w", toolName, err)
	}

	// Print output
	fmt.Print(output) // Keep as-is for tool output passthrough
	return nil
}

// listTools lists all installed tools
func listTools(cmd *cobra.Command, global bool) error {
	if global {
		// List global tools using the global tool infrastructure
		storage, err := install.NewToolStorage()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}
		tools, err := storage.ListTools()
		if err != nil {
			return fmt.Errorf("failed to list global tools: %w", err)
		}

		if len(tools) == 0 {
			cmd.Println("No global tools installed.")
			return nil
		}

		cmd.Println("Installed global tools:")
		for _, tool := range tools {
			cmd.Printf("  %s", tool.ToolName)
			if tool.Version != "" {
				cmd.Printf(" (%s)", tool.Version)
			}
			if !tool.InstallDate.IsZero() {
				cmd.Printf(" - installed: %s", tool.InstallDate.Format("2006-01-02 15:04:05"))
			}
			cmd.Println()
		}

		return nil
	}

	// List both global and project tools for unified view
	cmd.Println("Tool Inventory:")
	cmd.Println()

	// First, list global tools
	storage, err := install.NewToolStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	globalTools, err := storage.ListTools()
	if err == nil && len(globalTools) > 0 {
		cmd.Println("Global tools:")
		for _, tool := range globalTools {
			cmd.Printf("  %s", tool.ToolName)
			if tool.Version != "" {
				cmd.Printf(" (%s)", tool.Version)
			}
			cmd.Printf(" [global]")
			if !tool.InstallDate.IsZero() {
				cmd.Printf(" - %s", tool.InstallDate.Format("2006-01-02 15:04:05"))
			}
			cmd.Println()
		}
		cmd.Println()
	}

	// Then, list project/legacy tools
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	toolsDir := filepath.Join(dirs.DataDir, "tools")

	// Check if tools directory exists
	if _, err := os.Stat(toolsDir); os.IsNotExist(err) {
		if len(globalTools) == 0 {
			cmd.Println("No tools installed.")
		}
		return nil
	}

	// Read tools directory
	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		return fmt.Errorf("failed to read tools directory: %w", err)
	}

	if len(entries) > 0 {
		cmd.Println("Project tools:")
		for _, entry := range entries {
			if entry.IsDir() {
				toolPath := filepath.Join(toolsDir, entry.Name())
				info, err := os.Stat(toolPath)
				if err == nil {
					cmd.Printf("  %s [project] - %s\n", entry.Name(), info.ModTime().Format("2006-01-02 15:04:05"))
				} else {
					cmd.Printf("  %s [project]\n", entry.Name())
				}
			}
		}
	} else if len(globalTools) == 0 {
		cmd.Println("No tools installed.")
	}

	return nil
}

// upgradeTool upgrades an installed tool to the latest version
func upgradeTool(cmd *cobra.Command, toolName string, global bool) error {
	if global {
		// Use global tool infrastructure for upgrade
		cmd.Printf("Upgrading global tool '%s'...\n", toolName)

		// Initialize installer
		installer, err := install.NewToolInstaller()
		if err != nil {
			return fmt.Errorf("failed to create installer: %w", err)
		}

		// Initialize tool mapping to resolve tool name to module
		mapping := tool.NewToolMapping()
		resolution, err := mapping.ResolveToolToModule(toolName)
		if err != nil {
			return fmt.Errorf("failed to resolve tool '%s' to module: %w", toolName, err)
		}

		// Create upgrade options (force reinstall with latest version)
		options := &install.InstallOptions{
			ToolName:          toolName,
			ModuleName:        resolution.ModuleName,
			VersionConstraint: "", // Latest version
			Force:             true,
			Verbose:           true,
			Context:           cmd.Context(),
		}

		// Execute upgrade
		result, err := installer.InstallTool(options)
		if err != nil {
			return fmt.Errorf("failed to upgrade global tool '%s': %w", toolName, err)
		}

		cmd.Printf("Successfully upgraded global tool '%s' -> %s\n", toolName, result.InstallPath)
		return nil
	}

	// Legacy project tool upgrade
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	toolsDir := filepath.Join(dirs.DataDir, "tools")
	toolEnvDir := filepath.Join(toolsDir, toolName)

	// Check if tool is installed
	if _, err := os.Stat(toolEnvDir); os.IsNotExist(err) {
		return fmt.Errorf("project tool '%s' is not installed", toolName)
	}

	cmd.Printf("Upgrading project tool '%s'...\n", toolName)

	// Use PVX to upgrade the tool
	options := &pvx.ExecutionOptions{
		PerlVersion:    "", // Use default
		IsolationLevel: pvx.IsolationLocal,
		IsolationDir:   toolEnvDir,
		EnvName:        fmt.Sprintf("tool-%s", toolName),
		NoCleanup:      true,
	}

	// Create upgrade script
	upgradeScript := fmt.Sprintf(`
use App::cpanminus;
exec { 'cpanm' } 'cpanm', '--force', '%s';
`, toolName)

	// Create a temporary script file
	tmpDir := os.TempDir()
	scriptPath := filepath.Join(tmpDir, fmt.Sprintf("upgrade-%s-%d.pl", toolName, os.Getpid()))

	err = os.WriteFile(scriptPath, []byte(upgradeScript), 0644)
	if err != nil {
		return fmt.Errorf("failed to create upgrade script: %w", err)
	}
	defer os.Remove(scriptPath)

	// Set the script path
	options.ScriptPath = scriptPath

	// Execute upgrade
	_, err = pvx.ExecuteScript(options)
	if err != nil {
		return fmt.Errorf("failed to upgrade tool '%s': %w", toolName, err)
	}

	cmd.Printf("Project tool '%s' upgraded successfully\n", toolName)
	return nil
}

// uninstallTool removes an installed tool
func uninstallTool(cmd *cobra.Command, toolName string, global bool) error {
	if global {
		// Use global tool infrastructure for uninstall
		// Confirm removal
		cmd.Printf("Are you sure you want to uninstall global tool '%s'? [y/N] ", toolName)
		var response string
		_, _ = fmt.Scanln(&response)

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			cmd.Println("Global tool uninstall cancelled.")
			return nil
		}

		// Initialize storage and remove the global tool
		storage, err := install.NewToolStorage()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		// Remove the global tool
		err = storage.RemoveTool(toolName)
		if err != nil {
			return fmt.Errorf("failed to remove global tool '%s': %w", toolName, err)
		}

		// Remove shim for the tool
		shimManager, err := shim.NewManager()
		if err != nil {
			cmd.Printf("Warning: Failed to create shim manager: %v\n", err)
		} else {
			if err := shimManager.RemoveShim(toolName); err != nil {
				cmd.Printf("Warning: Failed to remove shim for '%s': %v\n", toolName, err)
			} else {
				cmd.Printf("Removed shim for command '%s'\n", toolName)
			}
		}

		cmd.Printf("Global tool '%s' has been uninstalled.\n", toolName)
		return nil
	}

	// Legacy project tool uninstall
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	toolsDir := filepath.Join(dirs.DataDir, "tools")
	toolEnvDir := filepath.Join(toolsDir, toolName)

	// Check if tool is installed
	if _, err := os.Stat(toolEnvDir); os.IsNotExist(err) {
		return fmt.Errorf("project tool '%s' is not installed", toolName)
	}

	// Confirm removal
	cmd.Printf("Are you sure you want to uninstall project tool '%s'? [y/N] ", toolName)
	var response string
	_, _ = fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		cmd.Println("Project tool uninstall cancelled.")
		return nil
	}

	// Remove the tool directory
	err = os.RemoveAll(toolEnvDir)
	if err != nil {
		return fmt.Errorf("failed to remove project tool '%s': %w", toolName, err)
	}

	// Remove shim if it exists
	removeToolShim(toolName)

	cmd.Printf("Project tool '%s' has been uninstalled.\n", toolName)
	return nil
}

// createToolShim creates a shim for a tool in the tools directory
func createToolShim(toolName, toolEnvDir string) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	// Create shims directory in tools
	shimsDir := filepath.Join(dirs.DataDir, "tools", "bin")
	if err := os.MkdirAll(shimsDir, 0755); err != nil {
		return fmt.Errorf("failed to create shims directory: %w", err)
	}

	shimPath := filepath.Join(shimsDir, toolName)

	// Create shim script that uses the tool's environment
	shimContent := fmt.Sprintf(`#!/bin/bash
# Tool shim for %s
export PATH="%s/bin:$PATH"
export PERL5LIB="%s/lib/perl5:$PERL5LIB"
exec "%s" "$@"
`, toolName, toolEnvDir, toolEnvDir, toolName)

	err = os.WriteFile(shimPath, []byte(shimContent), 0755)
	if err != nil {
		return fmt.Errorf("failed to create shim: %w", err)
	}

	return nil
}

// removeToolShim removes a tool's shim
func removeToolShim(toolName string) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	shimPath := filepath.Join(dirs.DataDir, "tools", "bin", toolName)
	if _, err := os.Stat(shimPath); err == nil {
		return os.Remove(shimPath)
	}
	return nil
}

// newHelpOnlyPVICommand creates a hidden PVI subcommand for help purposes only
func newHelpOnlyPVICommand() *cobra.Command {
	pviCmd := pvi.NewCommand()
	pviCmd.Use = "pvi"
	pviCmd.Hidden = true
	pviCmd.Short = "Perl Version Installer (shell alias for 'pvm module')"
	pviCmd.Long = "Manages CPAN modules for installed Perl versions. This is a shell alias for 'pvm module' - you can use either 'pvi' or 'pvm module' interchangeably."

	// Add custom help function with Fang styling
	setupStyledHelp(pviCmd, "PVI")

	// Add version subcommand
	pviCmd.AddCommand(newComponentVersionCommand("pvi"))

	return pviCmd
}

// newHelpOnlyPVXCommand creates a hidden PVX subcommand for help purposes only
func newHelpOnlyPVXCommand() *cobra.Command {
	pvxCmd := pvx.NewCommand()
	pvxCmd.Use = "pvx"
	pvxCmd.Hidden = true
	pvxCmd.Short = "Perl Version eXecutor (shell alias for 'pvm run')"
	pvxCmd.Long = "Executes Perl code in isolated environments. This is a shell alias for 'pvm run' - you can use either 'pvx' or 'pvm run' interchangeably."

	// Add custom help function with Fang styling
	setupStyledHelp(pvxCmd, "PVX")

	// Add version subcommand
	pvxCmd.AddCommand(newComponentVersionCommand("pvx"))

	return pvxCmd
}

// newHelpOnlyPSCCommand creates a hidden PSC subcommand for help purposes only
func newHelpOnlyPSCCommand() *cobra.Command {
	pscCmd := psc.NewCommand()
	pscCmd.Use = "psc"
	pscCmd.Hidden = true
	pscCmd.Short = "Perl Script Compiler (shell alias for 'pvm build')"
	pscCmd.Long = "Provides static type checking for Perl code with type annotations. This is a shell alias for 'pvm build' - you can use either 'psc' or 'pvm build' interchangeably."

	// Add custom help function with Fang styling
	setupStyledHelp(pscCmd, "PSC")

	// Add version subcommand
	pscCmd.AddCommand(newComponentVersionCommand("psc"))

	return pscCmd
}

// newComponentVersionCommand creates a version command for a specific component
func newComponentVersionCommand(component string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  fmt.Sprintf("Print detailed version information about the %s component", component),
		Run: func(cmd *cobra.Command, args []string) {
			ui := cli.GetUI(cmd)

			if cli.Verbose {
				// Show detailed version information in verbose mode
				ui.Header(fmt.Sprintf("%s Version Information", strings.ToUpper(component)))

				buildInfo := version.GetBuildInfo()
				ui.KeyValue(map[string]string{
					"Version":    buildInfo["version"],
					"Build Time": buildInfo["buildTime"],
					"Commit":     buildInfo["commitHash"],
					"Go Version": buildInfo["goVersion"],
					"OS/Arch":    fmt.Sprintf("%s/%s", buildInfo["os"], buildInfo["arch"]),
				})
			} else {
				// Show simple version in normal mode
				ui.Println(version.ComponentVersion(component))
			}
		},
	}
}

// setupStyledHelp configures a command to use Fang-styled help output
func setupStyledHelp(cmd *cobra.Command, componentName string) {
	// Override the help function to use styled output
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ui := cli.GetUI(cmd)

		// Help should always be shown, even in quiet mode
		originalQuiet := ui.Context().Quiet
		ui.SetQuiet(false)
		defer ui.SetQuiet(originalQuiet)

		// Header with component name
		ui.Header(fmt.Sprintf("%s - %s", componentName, cmd.Short))
		ui.Println("")

		// Usage section
		if cmd.Use != "" {
			ui.SubHeader("Usage:")
			ui.Printf("  %s\n", cmd.Use)
			ui.Println("")
		}

		// Description section
		if cmd.Long != "" {
			ui.SubHeader("Description:")
			ui.Printf("  %s\n", cmd.Long)
			ui.Println("")
		}

		// Commands section
		commands := cmd.Commands()
		if len(commands) > 0 {
			ui.SubHeader("Available Commands:")
			for _, subCmd := range commands {
				if !subCmd.Hidden {
					ui.Printf("  %-15s %s\n", subCmd.Name(), subCmd.Short)
				}
			}
			ui.Println("")
		}

		// Flags section
		if cmd.HasAvailableFlags() {
			ui.SubHeader("Flags:")
			ui.Printf("%s", cmd.Flags().FlagUsages())
			ui.Println("")
		}

		// Global flags section
		if cmd.HasAvailableInheritedFlags() {
			ui.SubHeader("Global Flags:")
			ui.Printf("%s", cmd.InheritedFlags().FlagUsages())
			ui.Println("")
		}

		// Footer with additional help
		ui.Info("Use \"%s [command] --help\" for more information about a command.", cmd.CommandPath())
	})
}

// newUpdateCommand creates the update command for PVM self-updater
func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update PVM to the latest version",
		Long:  "Update PVM to the latest version from GitHub releases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeUpdateCommand(cmd, args)
		},
	}

	// Add flags
	cmd.Flags().String("version", "", "Update to a specific version")
	cmd.Flags().Bool("check", false, "Check for updates without installing")
	cmd.Flags().Bool("force", false, "Force update even if already up to date")
	cmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")
	cmd.Flags().Bool("no-backup", false, "Skip creating backup before update")
	cmd.Flags().Bool("no-rollback", false, "Disable automatic rollback on failure")
	cmd.Flags().Bool("prerelease", false, "Include pre-release versions")
	cmd.Flags().String("token", "", "GitHub token for higher API rate limits")
	cmd.Flags().Bool("ignore-install-method", false, "Force self-update regardless of installation method")

	return cmd
}

// newAutoUpdateCommand creates the auto-update command
func newAutoUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auto-update [enable|disable|config|check|status] [args...]",
		Short: "Manage automatic update checking",
		Long: `Configure and manage automatic update checking for PVM.

Subcommands:
  enable                 Enable automatic update checking
  disable                Disable automatic update checking
  config <key> <value>   Configure auto-update settings
  check                  Check for updates now
  status                 Show current auto-update status

Configuration keys:
  channel                Update channel (stable, beta, alpha, nightly, developer)
  interval               Check interval (e.g., 24h, 12h, 1h30m)
  quiet                  Quiet mode (true/false)
  auto-install           Auto-install updates (true/false)
  repository             GitHub repository (owner/repo)
  install-schedule       Installation schedule (days hour minute)

Examples:
  pvm auto-update enable
  pvm auto-update config channel beta
  pvm auto-update config interval 12h
  pvm auto-update config install-schedule "sun,wed" 2 30
  pvm auto-update check`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeAutoUpdateCommand(cmd, args)
		},
	}

	// Add flags
	cmd.Flags().String("token", "", "GitHub token for higher API rate limits")

	return cmd
}

// Shim management functions

// installShimIntegration installs shell integration for shims
func installShimIntegration(cmd *cobra.Command, shell, position string, force bool) error {
	shimManager, err := shim.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create shim manager: %w", err)
	}

	shimDir := shimManager.GetShimDirectory()
	integrator := shim.NewShellIntegrator(shimDir)

	// Auto-detect shell if not provided
	if shell == "" {
		shell = shim.DetectShellType()
		cmd.Printf("Auto-detected shell: %s\n", shell)
	}

	// Parse position
	var pos shim.PathPosition
	switch position {
	case "first":
		pos = shim.PathPositionFirst
	case "last":
		pos = shim.PathPositionLast
	case "after-system":
		pos = shim.PathPositionAfterSystem
	default:
		return fmt.Errorf("invalid position '%s': must be first, last, or after-system", position)
	}

	// Check if already installed
	if !force {
		installed, err := integrator.IsShellIntegrationInstalled(shell)
		if err != nil {
			return fmt.Errorf("failed to check installation status: %w", err)
		}
		if installed {
			cmd.Printf("Shell integration for %s is already installed. Use --force to reinstall.\n", shell)
			return nil
		}
	}

	// Install integration
	if err := integrator.InstallShellIntegration(shell, pos); err != nil {
		return fmt.Errorf("failed to install shell integration: %w", err)
	}

	cmd.Printf("Successfully installed shell integration for %s\n", shell)
	cmd.Printf("Shim directory: %s\n", shimDir)
	cmd.Printf("Restart your shell or run: source ~/.<shell>rc\n")

	return nil
}

// removeShimIntegration removes shell integration for shims
func removeShimIntegration(cmd *cobra.Command, shell string) error {
	shimManager, err := shim.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create shim manager: %w", err)
	}

	shimDir := shimManager.GetShimDirectory()
	integrator := shim.NewShellIntegrator(shimDir)

	// Auto-detect shell if not provided
	if shell == "" {
		shell = shim.DetectShellType()
		cmd.Printf("Auto-detected shell: %s\n", shell)
	}

	// Check if installed
	installed, err := integrator.IsShellIntegrationInstalled(shell)
	if err != nil {
		return fmt.Errorf("failed to check installation status: %w", err)
	}
	if !installed {
		cmd.Printf("Shell integration for %s is not installed.\n", shell)
		return nil
	}

	// Remove integration
	if err := integrator.RemoveShellIntegration(shell); err != nil {
		return fmt.Errorf("failed to remove shell integration: %w", err)
	}

	cmd.Printf("Successfully removed shell integration for %s\n", shell)
	cmd.Printf("Restart your shell for changes to take effect.\n")

	return nil
}

// listShims lists all installed shims
func listShims(cmd *cobra.Command) error {
	shimManager, err := shim.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create shim manager: %w", err)
	}

	shims, err := shimManager.ListShims()
	if err != nil {
		return fmt.Errorf("failed to list shims: %w", err)
	}

	if len(shims) == 0 {
		cmd.Println("No shims installed.")
		return nil
	}

	cmd.Printf("Installed shims (%d):\n", len(shims))
	for _, shimName := range shims {
		cmd.Printf("  %s\n", shimName)
	}

	shimDir := shimManager.GetShimDirectory()
	cmd.Printf("\nShim directory: %s\n", shimDir)

	return nil
}

// shimStatus shows the status of shim integration
func shimStatus(cmd *cobra.Command) error {
	shimManager, err := shim.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create shim manager: %w", err)
	}

	shimDir := shimManager.GetShimDirectory()
	pathManager := shim.NewPathManager(shimDir)
	integrator := shim.NewShellIntegrator(shimDir)

	cmd.Printf("Shim Status Report\n")
	cmd.Printf("==================\n\n")

	cmd.Printf("Shim directory: %s\n", shimDir)

	// Check if shim directory is in PATH
	inPath := pathManager.IsInPath()
	cmd.Printf("In PATH: %t\n", inPath)

	if inPath {
		precedence, err := pathManager.GetPrecedence()
		if err == nil {
			cmd.Printf("PATH precedence: %d (lower is higher priority)\n", precedence)
		}
	}

	// Check shell integration status
	shell := shim.DetectShellType()
	cmd.Printf("Detected shell: %s\n", shell)

	installed, err := integrator.IsShellIntegrationInstalled(shell)
	if err != nil {
		cmd.Printf("Shell integration: error checking (%v)\n", err)
	} else {
		cmd.Printf("Shell integration: %t\n", installed)
	}

	// List shims
	shims, err := shimManager.ListShims()
	if err != nil {
		cmd.Printf("Shims: error listing (%v)\n", err)
	} else {
		cmd.Printf("Shims installed: %d\n", len(shims))
	}

	// Check for conflicts
	if len(shims) > 0 {
		conflicts, err := pathManager.FindConflicts(shims)
		switch {
		case err != nil:
			cmd.Printf("PATH conflicts: error checking (%v)\n", err)
		case len(conflicts) > 0:
			cmd.Printf("\nPATH Conflicts:\n")
			for toolName, conflictPaths := range conflicts {
				cmd.Printf("  %s: conflicts with %d other executables\n", toolName, len(conflictPaths))
				for _, conflictPath := range conflictPaths {
					cmd.Printf("    %s\n", conflictPath)
				}
			}
		default:
			cmd.Printf("PATH conflicts: none\n")
		}
	}

	return nil
}
