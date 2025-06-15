// ABOUTME: PVM-specific commands and functionality
// ABOUTME: Implements commands for Perl version management

package pvm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/psc"
	"tamarou.com/pvm/internal/pvi"
	"tamarou.com/pvm/internal/pvx"
	"tamarou.com/pvm/internal/xdg"
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
		newListCommand(), // Alias for versions command for compatibility
		newAvailableCommand(),
		newDownloadCommand(),
		newBuildCommand(),  // Now incorporates PSC functionality
		newRunCommand(),    // New unified run command (incorporates PVX)
		newModuleCommand(), // New unified module command (incorporates PVI)
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
		newToolCommand(),

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
			fmt.Println("Shim executables rebuilt successfully.")

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

// newPVXCommand creates a PVX command as a shell alias to 'pvm run'
func newPVXCommand() *cobra.Command {
	// Get the PVX command from the PVX package
	pvxCmd := pvx.NewCommand()

	// Customize as shell alias
	pvxCmd.Use = "pvx"
	pvxCmd.Short = "Perl Version eXecutor (shell alias for 'pvm run')"
	pvxCmd.Long = "Executes Perl code in isolated environments. This is a shell alias for 'pvm run' - you can use either 'pvx' or 'pvm run' interchangeably."

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
	cmd.Printf("Building Perl %s from source...\n", version)

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
}

// buildProject handles building Perl projects (type checking, compilation)
func buildProject(cmd *cobra.Command, target string) error {
	// Get flags
	checkOnly, err := cmd.Flags().GetBool("check-only")
	if err != nil {
		return err
	}

	inline, err := cmd.Flags().GetBool("inline")
	if err != nil {
		return err
	}

	watch, err := cmd.Flags().GetBool("watch")
	if err != nil {
		return err
	}

	outputDir, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	clean, err := cmd.Flags().GetBool("clean")
	if err != nil {
		return err
	}

	// Determine what to build
	if target == "" {
		target = "." // Current directory
	}

	cmd.Printf("Building Perl project in %s...\n", target)

	// For now, delegate to PSC for project builds
	// This is a simplified implementation that will be enhanced in later steps
	pscCmd := psc.NewCommand()

	if checkOnly {
		// Run type checking only
		cmd.Println("Running type check...")

		// Find a check subcommand
		var checkCmd *cobra.Command
		for _, subCmd := range pscCmd.Commands() {
			if subCmd.Use == "check" {
				checkCmd = subCmd
				break
			}
		}

		if checkCmd != nil {
			// Set up args for PSC check command
			if target == "." {
				// Check all files in current directory
				checkCmd.SetArgs([]string{"."})
			} else {
				checkCmd.SetArgs([]string{target})
			}

			return checkCmd.Execute()
		} else {
			return fmt.Errorf("type checking not available")
		}
	}

	if watch {
		// Find a watch subcommand
		var watchCmd *cobra.Command
		for _, subCmd := range pscCmd.Commands() {
			if subCmd.Use == "watch" {
				watchCmd = subCmd
				break
			}
		}

		if watchCmd != nil {
			cmd.Println("Starting watch mode...")
			if target == "." {
				watchCmd.SetArgs([]string{"."})
			} else {
				watchCmd.SetArgs([]string{target})
			}
			return watchCmd.Execute()
		} else {
			return fmt.Errorf("watch mode not available")
		}
	}

	// Default: run type check and compilation
	cmd.Println("Running type check and compilation...")

	// This is a placeholder for future build system functionality
	// For now, just run type checking
	var checkCmd *cobra.Command
	for _, subCmd := range pscCmd.Commands() {
		if subCmd.Use == "check" {
			checkCmd = subCmd
			break
		}
	}

	if checkCmd != nil {
		if target == "." {
			checkCmd.SetArgs([]string{"."})
		} else {
			checkCmd.SetArgs([]string{target})
		}

		err := checkCmd.Execute()
		if err != nil {
			return err
		}

		cmd.Println("Type check passed!")

		if inline {
			cmd.Println("Note: .pmc generation will be implemented in a future version")
		}

		if outputDir != "" {
			cmd.Printf("Note: Custom output directory '%s' will be supported in a future version\n", outputDir)
		}

		if clean {
			cmd.Println("Note: Clean build functionality will be implemented in a future version")
		}

		cmd.Println("Project build completed!")
		return nil
	}

	return fmt.Errorf("project build functionality not yet fully implemented")
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
				// Generate and create shell initialization scripts
				err := perl.CreateShellInitScripts()
				if err != nil {
					return err
				}
				fmt.Println("Shell integration initialized")
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

func newBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [target]",
		Short: "Build Perl projects or Perl from source",
		Long:  "Build Perl projects (type checking, compilation) or build Perl from source code",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine build mode: project build (default) or Perl source build
			var target string
			if len(args) > 0 {
				target = args[0]
			}

			// Check if this is a Perl version (for source build) or project build
			isSourceBuild, err := cmd.Flags().GetBool("from-source")
			if err != nil {
				return err
			}

			if isSourceBuild || (target != "" && isPerlVersion(target)) {
				// Perl source build
				return buildPerlFromSource(cmd, target)
			} else {
				// Project build (type checking, compilation)
				return buildProject(cmd, target)
			}
		},
	}

	// Add flags for both project and Perl source builds
	cmd.Flags().Bool("from-source", false, "Build Perl from source (use when target is a Perl version)")

	// Project build flags
	cmd.Flags().Bool("check-only", false, "Only run type checking, don't compile")
	cmd.Flags().Bool("inline", false, "Generate .pmc files for development")
	cmd.Flags().Bool("watch", false, "Watch for changes and rebuild")
	cmd.Flags().String("output", "", "Output directory for build artifacts")
	cmd.Flags().Bool("clean", false, "Clean build artifacts before building")

	// Perl source build flags
	cmd.Flags().String("source", "", "Source archive file path (default: download or use cached)")
	cmd.Flags().String("prefix", "", "Installation directory (default: XDG_DATA_HOME/pvm/versions/<version>)")
	cmd.Flags().Int("jobs", 0, "Number of parallel build jobs (default: number of CPU cores)")
	cmd.Flags().Bool("test", false, "Run Perl tests after building")
	cmd.Flags().Bool("cleanup", true, "Clean up build directory after installation")
	cmd.Flags().StringArray("configure-options", nil, "Additional options to pass to Configure (can be specified multiple times)")

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
			Short: "Generate activation command for an environment",
			Long:  "Output shell command to activate a named environment",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return activateEnvironment(cmd, args[0])
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
		Long:  "Commands for installing, running, and managing Perl tools",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "install [tool[@version]]",
			Short: "Install a tool",
			Long:  "Install a Perl tool (module) and make it available as a command",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return installTool(cmd, args[0])
			},
		},
		&cobra.Command{
			Use:   "run [tool] [args...]",
			Short: "Run a tool",
			Long:  "Run a Perl tool without installing it permanently",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runTool(cmd, args[0], args[1:])
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List installed tools",
			Long:  "List all installed tools and their versions",
			RunE: func(cmd *cobra.Command, args []string) error {
				return listTools(cmd)
			},
		},
		&cobra.Command{
			Use:   "upgrade [tool]",
			Short: "Upgrade a tool",
			Long:  "Upgrade an installed tool to the latest version",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return upgradeTool(cmd, args[0])
			},
		},
		&cobra.Command{
			Use:   "uninstall [tool]",
			Short: "Uninstall a tool",
			Long:  "Remove an installed tool",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return uninstallTool(cmd, args[0])
			},
		},
	)

	return cmd
}

// installTool installs a tool and creates a shim for it
func installTool(cmd *cobra.Command, toolSpec string) error {
	// Parse tool specification (tool@version or just tool)
	parts := strings.Split(toolSpec, "@")
	toolName := parts[0]
	var version string
	if len(parts) > 1 {
		version = parts[1]
	}

	cmd.Printf("Installing tool '%s'", toolName)
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
		IsolationLevel: pvx.IsolationLow,
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
	// Use PVX to execute the tool directly
	options := &pvx.ExecutionOptions{
		PerlVersion:    "", // Use default
		IsolationLevel: pvx.IsolationLow,
		NoCleanup:      false, // Clean up for temporary runs
	}

	output, err := pvx.ExecuteTool(options, toolName, toolArgs)
	if err != nil {
		return fmt.Errorf("failed to run tool '%s': %w", toolName, err)
	}

	// Print output
	fmt.Print(output)
	return nil
}

// listTools lists all installed tools
func listTools(cmd *cobra.Command) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	toolsDir := filepath.Join(dirs.DataDir, "tools")

	// Check if tools directory exists
	if _, err := os.Stat(toolsDir); os.IsNotExist(err) {
		cmd.Println("No tools installed.")
		return nil
	}

	// Read tools directory
	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		return fmt.Errorf("failed to read tools directory: %w", err)
	}

	if len(entries) == 0 {
		cmd.Println("No tools installed.")
		return nil
	}

	cmd.Println("Installed tools:")
	for _, entry := range entries {
		if entry.IsDir() {
			toolPath := filepath.Join(toolsDir, entry.Name())
			info, err := os.Stat(toolPath)
			if err == nil {
				cmd.Printf("  %s (installed: %s)\n", entry.Name(), info.ModTime().Format("2006-01-02 15:04:05"))
			} else {
				cmd.Printf("  %s\n", entry.Name())
			}
		}
	}

	return nil
}

// upgradeTool upgrades an installed tool to the latest version
func upgradeTool(cmd *cobra.Command, toolName string) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	toolsDir := filepath.Join(dirs.DataDir, "tools")
	toolEnvDir := filepath.Join(toolsDir, toolName)

	// Check if tool is installed
	if _, err := os.Stat(toolEnvDir); os.IsNotExist(err) {
		return fmt.Errorf("tool '%s' is not installed", toolName)
	}

	cmd.Printf("Upgrading tool '%s'...\n", toolName)

	// Use PVX to upgrade the tool
	options := &pvx.ExecutionOptions{
		PerlVersion:    "", // Use default
		IsolationLevel: pvx.IsolationLow,
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

	cmd.Printf("Tool '%s' upgraded successfully\n", toolName)
	return nil
}

// uninstallTool removes an installed tool
func uninstallTool(cmd *cobra.Command, toolName string) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return err
	}

	toolsDir := filepath.Join(dirs.DataDir, "tools")
	toolEnvDir := filepath.Join(toolsDir, toolName)

	// Check if tool is installed
	if _, err := os.Stat(toolEnvDir); os.IsNotExist(err) {
		return fmt.Errorf("tool '%s' is not installed", toolName)
	}

	// Confirm removal
	cmd.Printf("Are you sure you want to uninstall tool '%s'? [y/N] ", toolName)
	var response string
	_, _ = fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		cmd.Println("Tool uninstall cancelled.")
		return nil
	}

	// Remove the tool directory
	err = os.RemoveAll(toolEnvDir)
	if err != nil {
		return fmt.Errorf("failed to remove tool '%s': %w", toolName, err)
	}

	// Remove shim if it exists
	removeToolShim(toolName)

	cmd.Printf("Tool '%s' has been uninstalled.\n", toolName)
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
	return pviCmd
}

// newHelpOnlyPVXCommand creates a hidden PVX subcommand for help purposes only
func newHelpOnlyPVXCommand() *cobra.Command {
	pvxCmd := pvx.NewCommand()
	pvxCmd.Use = "pvx"
	pvxCmd.Hidden = true
	pvxCmd.Short = "Perl Version eXecutor (shell alias for 'pvm run')"
	pvxCmd.Long = "Executes Perl code in isolated environments. This is a shell alias for 'pvm run' - you can use either 'pvx' or 'pvm run' interchangeably."
	return pvxCmd
}

// newHelpOnlyPSCCommand creates a hidden PSC subcommand for help purposes only
func newHelpOnlyPSCCommand() *cobra.Command {
	pscCmd := psc.NewCommand()
	pscCmd.Use = "psc"
	pscCmd.Hidden = true
	pscCmd.Short = "Perl Script Compiler (shell alias for 'pvm build')"
	pscCmd.Long = "Provides static type checking for Perl code with type annotations. This is a shell alias for 'pvm build' - you can use either 'psc' or 'pvm build' interchangeably."
	return pscCmd
}
