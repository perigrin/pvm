// ABOUTME: PVM-specific commands and functionality
// ABOUTME: Implements commands for Perl version management

package pvm

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/psc"
	"tamarou.com/pvm/internal/pvx"
)

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
		newGlobalCommand(),
		newLocalCommand(),
		newVersionsCommand(),
		newAvailableCommand(),
		newDownloadCommand(),
		newBuildCommand(),
		newExecCommand(),
		newUninstallCommand(),
		newImportSystemCommand(),
		newImportCommand(),
		newRehashCommand(),
		newResolveCommand(),
		newInitCommand(),
		newShellCommand(),
		newPVXCommand(),
		newPSCCommand(),
		newMCPCommand(),

		// These are implemented in their own files
		newSymlinksCommand(), // from symlinks.go
		newConfigCommand(),   // from config.go
		newPerlCommand(),     // from perl.go
		newVersionCommand(),  // from version.go
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

			if skipBuild {
				cmd.Println("Skip-build specified but no import functionality implemented yet.")
				cmd.Println("This will be implemented in a future version.")
				return nil
			}

			// Build Perl using our build functionality
			cmd.Printf("Installing Perl %s...\n", version)

			// Create progress callback to display build progress
			var currentStage perl.BuildProgressStage
			progressCallback := func(stage perl.BuildProgressStage, details string, progress float64) {
				// Only print stage transition once
				if stage != currentStage {
					cmd.Printf("\n=== %s ===\n", stage.String())
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
							cmd.Println(details)
						}
					} else {
						// For other stages, print all details
						cmd.Println(details)
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
					fmt.Printf("\r%s %.1f%%                    ",
						progressBar, progress*100)

					if progress >= 1.0 {
						fmt.Println()
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
				cmd.Printf("\nInstallation failed: %v\n", err)
				return err
			}

			// Show build results
			cmd.Printf("\nInstallation completed successfully!\n")
			cmd.Printf("Perl %s installed at: %s\n", result.Version, result.InstallPath)
			cmd.Printf("Total installation time: %s\n", result.Duration.Round(time.Second))

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

	return cmd
}

func newUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use [version]",
		Short: "Use a specific version in the current shell",
		Long:  "Temporarily use a specific Perl version in the current shell session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]

			// Switch version using shell integration
			err := perl.SwitchVersion(version, "shell")
			if err != nil {
				return err
			}

			// Success message
			cmd.Printf("Using Perl %s in current shell\n", version)
			cmd.Println("Note: This command only works properly when PVM's shell integration is set up")
			cmd.Println("If you haven't set up shell integration, run 'pvm init' to get started")

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
			cmd.Printf("Global Perl version set to %s\n", version)
			cmd.Println("This is now the default version when no other version is specified")

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
			cmd.Printf("Local Perl version set to %s for %s\n", version, dir)
			cmd.Println("This version will be used in this directory and its subdirectories")
			cmd.Println("Note: Shell integration must be set up for automatic switching to work")

			return nil
		},
	}
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

			// Check if we have any installed versions
			if len(installedVersions) == 0 {
				cmd.Println("No versions installed. Use 'pvm install <version>' to install a version.")
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
			cmd.Println("Installed Perl versions:")
			for i, versionInfo := range installedVersions {
				// Add decoration for special versions
				decoration := ""
				if i == 0 {
					decoration = " (latest)"
				}

				// Basic output
				if !showPath && !showSource {
					cmd.Printf("  %s%s\n", versionInfo.Version, decoration)
					continue
				}

				// Detailed output
				cmd.Printf("  %s%s\n", versionInfo.Version, decoration)

				if showPath {
					cmd.Printf("    Path: %s\n", versionInfo.InstallPath)
				}

				if showSource {
					cmd.Printf("    Source: %s\n", versionInfo.Source)
					cmd.Printf("    Installed: %s\n", versionInfo.InstallTime.Format("2006-01-02 15:04:05"))
				}

				// Add a separator between versions for detailed output
				if i < len(installedVersions)-1 && (showPath || showSource) {
					cmd.Println()
				}
			}

			// Add a hint about current/active version
			cmd.Println("\nNote: Use 'pvm current' to show the currently active Perl version.")

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
			cmd.Println("Available Perl versions:")

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
			cmd.Println("\nDevelopment versions:")
			for _, version := range devVersions {
				installed := ""
				if installedMap[version] {
					installed = " (installed)"
				}
				cmd.Printf("  %s%s\n", version, installed)
			}

			// Display stable versions by group
			cmd.Println("\nStable versions:")

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
				cmd.Printf("  %s series:\n", groupKey)
				for _, version := range versions {
					installed := ""
					if installedMap[version] {
						installed = " (installed)"
					}
					cmd.Printf("    %s%s\n", version, installed)
				}
			}

			cmd.Println("\nUse 'pvm install <version>' to install a specific version.")
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
					fmt.Printf("\r%s %.1f%% (%d/%d bytes)                    ",
						progressBar, percentage, transferred, total)

					if done {
						fmt.Println()
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
			cmd.Printf("Downloading Perl %s...\n", version)

			if mirror != "" {
				cmd.Printf("Using mirror: %s\n", mirror)
			}

			// Perform the download
			result, err := perl.DownloadPerlSource(options)
			if err != nil {
				return err
			}

			// Show download results
			if result.FromCache {
				cmd.Printf("Loaded from cache: %s\n", result.Path)
			} else {
				cmd.Printf("Download complete: %s\n", result.Path)
				cmd.Printf("Download size: %d bytes\n", result.Size)
				cmd.Printf("Download time: %s\n", result.Duration.Round(time.Millisecond))
			}

			if result.Checksum != "" {
				cmd.Printf("SHA-256: %s\n", result.Checksum)
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
			if len(args) < 2 {
				cmd.PrintErrln("Usage: pvm exec [version] [command]")
				return
			}

			version := args[0]
			command := args[1:]

			err := execCommand(cmd, version, command)
			if err != nil {
				cmd.PrintErrln("Error:", err)
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
			if !force {
				cmd.Printf("Are you sure you want to uninstall Perl %s? [y/N] ", version)
				var response string
				_, _ = fmt.Scanln(&response)

				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					cmd.Println("Uninstall cancelled.")
					return nil
				}
			}

			// If this is a system Perl, warn the user
			if versionInfo.Source == "system" {
				cmd.Println("Note: This is a system Perl installation.")
				cmd.Println("The installation will be unregistered from PVM but the actual files will not be removed.")
			}

			// Perform the uninstallation
			cmd.Printf("Uninstalling Perl %s...\n", version)
			err = perl.UninstallVersion(version)
			if err != nil {
				return err
			}

			cmd.Printf("Perl %s has been uninstalled.\n", version)
			return nil
		},
	}

	// Add flags
	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

func newImportSystemCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import-system",
		Short: "Import system Perl installation",
		Long:  "Import the system Perl installation into PVM registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return importSystemPerl(cmd)
		},
	}
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
	cmd.Printf("Detecting %s installations...\n", tool)

	installations, err := perl.ImportFromLegacyTool(tool)
	if err != nil {
		return err
	}

	cmd.Printf("Found %d %s installation(s):\n", len(installations), tool)
	for i, inst := range installations {
		defaultMark := ""
		if inst.IsDefault {
			defaultMark = " (default)"
		}
		cmd.Printf("%d. %s%s at %s\n", i+1, inst.Version, defaultMark, inst.Path)
	}

	// In a real implementation, here we would:
	// 1. Ask the user which installations to import
	// 2. Create symlinks or copy the installations
	// 3. Register them in PVM's internal database
	// But for now, we'll just report what was found

	cmd.Println("\nNote: This command currently only detects installations.")
	cmd.Println("Actual import functionality will be implemented in a future version.")

	// If it's perlbrew, also show aliases
	if tool == perl.Perlbrew {
		aliases, err := perl.GetPerlbrewAliases()
		if err == nil && len(aliases) > 0 {
			cmd.Println("\nPerlbrew aliases detected:")
			for alias, target := range aliases {
				cmd.Printf("  %s -> %s\n", alias, target)
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
			cmd.Println("Rebuilding shim executables...")

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
			cmd.Println("Shim executables rebuilt successfully.")

			// Warn if PATH is not configured
			if !pathConfigured {
				cmd.Println("\nWarning: The shim directory is not in your PATH.")
				cmd.Printf("To use pvm, add the following directory to your PATH:\n%s\n", shimDir)

				// Get the command to add to PATH
				pathCmd, err := perl.GetPathConfigCommand()
				if err == nil {
					cmd.Printf("\nYou can do this with the following command:\n%s\n", pathCmd)
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
				cmd.Printf("Resolved version: %s\n", version.Version)
				cmd.Printf("Source: %s\n", version.Source)
				cmd.Printf("Path: %s\n", path)
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

// newPVXCommand creates a command for the PVX subcommand
func newPVXCommand() *cobra.Command {
	// Import the PVX command directly from the pvx package
	pvxCommand := &cobra.Command{
		Use:   "pvx",
		Short: "Perl Version eXecutor",
		Long:  "Executes Perl code in isolated environments",
	}

	// Instead of reimplementing the PVX command here, we'll get a fresh command
	// from the pvx package and copy all its relevant properties
	originalCmd := pvx.NewCommand()

	// Copy the Run function
	pvxCommand.Run = originalCmd.Run

	// Copy all flags
	originalCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		pvxCommand.Flags().AddFlag(flag)
	})

	return pvxCommand
}

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

			// Get shell script for the detected shell
			script, err := perl.GetCurrentShellScript(shell)
			if err != nil {
				return err
			}

			// Print the script to stdout (for eval)
			fmt.Print(script)
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
				// Redirect to init command
				return newInitCommand().RunE(cmd, args)
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

func newBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [version]",
		Short: "Build Perl from source",
		Long:  "Build and install a specific version of Perl from source code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]

			// Get flags
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

			cleanupBuildDir, err := cmd.Flags().GetBool("cleanup")
			if err != nil {
				return err
			}

			// Create a slice for configure options
			configureOptions, err := cmd.Flags().GetStringArray("configure-options")
			if err != nil {
				return err
			}

			// Create progress callback to display build progress
			var currentStage perl.BuildProgressStage
			progressCallback := func(stage perl.BuildProgressStage, details string, progress float64) {
				// Only print stage transition once
				if stage != currentStage {
					cmd.Printf("\n=== %s ===\n", stage.String())
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
							cmd.Println(details)
						}
					} else {
						// For other stages, print all details
						cmd.Println(details)
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
					fmt.Printf("\r%s %.1f%%                    ",
						progressBar, progress*100)

					if progress >= 1.0 {
						fmt.Println()
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
				CleanupBuildDir:  cleanupBuildDir,
				ConfigureOptions: configureOptions,
				ProgressCallback: progressCallback,
				Context:          cmd.Context(),
			}

			// Print build information
			cmd.Printf("Building Perl %s...\n", version)

			if sourceFile != "" {
				cmd.Printf("Using source file: %s\n", sourceFile)
			}

			if installDir != "" {
				cmd.Printf("Installation directory: %s\n", installDir)
			}

			if buildJobs > 0 {
				cmd.Printf("Using %d parallel build jobs\n", buildJobs)
			} else {
				cmd.Printf("Using default number of build jobs\n")
			}

			if runTests {
				cmd.Println("Will run tests after building")
			}

			if cleanupBuildDir {
				cmd.Println("Will clean up build directory after installation")
			}

			// Start the build
			result, err := perl.BuildPerl(options)
			if err != nil {
				cmd.Printf("\nBuild failed: %v\n", err)
				return err
			}

			// Show build results
			cmd.Printf("\nBuild completed successfully!\n")
			cmd.Printf("Perl %s installed at: %s\n", result.Version, result.InstallPath)
			cmd.Printf("Total build time: %s\n", result.Duration.Round(time.Second))

			// Show timing for each stage
			cmd.Println("\nBuild stage timing:")
			for stage, duration := range result.Stages {
				cmd.Printf("  %-12s: %s\n", stage.String(), duration.Round(time.Second))
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().String("source", "", "Source archive file path (default: download or use cached)")
	cmd.Flags().String("prefix", "", "Installation directory (default: XDG_DATA_HOME/pvm/versions/<version>)")
	cmd.Flags().Int("jobs", 0, "Number of parallel build jobs (default: number of CPU cores)")
	cmd.Flags().Bool("test", false, "Run Perl tests after building")
	cmd.Flags().Bool("cleanup", true, "Clean up build directory after installation")
	cmd.Flags().StringArray("configure-options", nil, "Additional options to pass to Configure (can be specified multiple times)")

	return cmd
}

// newPSCCommand creates a PSC command that delegates to the PSC package
func newPSCCommand() *cobra.Command {
	// Get the PSC command from the PSC package
	pscCmd := psc.NewCommand()

	// Customize for integration with PVM
	pscCmd.Use = "psc"
	pscCmd.Short = "Perl Script Compiler (Type Checking)"
	pscCmd.Long = "Provides static type checking for Perl code with type annotations"

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
		return fmt.Errorf("Perl version %s not found at %s", resolved.Version, resolved.Path)
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

	// Check if this version is already registered
	installed, err := perl.IsVersionInstalled(systemPerl.Version)
	if err != nil {
		return fmt.Errorf("failed to check if version is installed: %w", err)
	}
	if installed {
		cmd.Printf("System Perl %s is already registered with PVM\n", systemPerl.Version)
		return nil
	}

	// Register the system Perl by creating a VersionInfo entry
	versionInfo := perl.VersionInfo{
		Version:     systemPerl.Version,
		InstallPath: systemPerl.Path,
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
