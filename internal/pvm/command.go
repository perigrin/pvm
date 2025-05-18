// ABOUTME: PVM-specific commands and functionality
// ABOUTME: Implements commands for Perl version management

package pvm

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/perl"
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
		newImportCommand(),
		newRehashCommand(),
		newResolveCommand(),
		
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
						if i < completeChars {
							progressBar += "="
						} else if i == completeChars {
							progressBar += ">"
						} else {
							progressBar += " "
						}
					}
					progressBar += "]"
					
					// Clear line and show progress
					fmt.Printf("\r%s %.1f%%                    ", 
						progressBar, progress * 100)
					
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
			
			// TODO: Register the installed version in a future prompt
			
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
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Use command not yet implemented")
		},
	}
}

func newGlobalCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "global [version]",
		Short: "Set the global Perl version",
		Long:  "Set the default Perl version for all shells",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Global command not yet implemented")
		},
	}
}

func newLocalCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "local [version]",
		Short: "Set the local version for a directory",
		Long:  "Set the Perl version for the current directory and subdirectories",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Local command not yet implemented")
		},
	}
}

func newVersionsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "versions",
		Short: "List installed versions",
		Long:  "List all installed Perl versions",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Versions command not yet implemented")
		},
	}
}

func newAvailableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "available",
		Short: "List available Perl versions",
		Long:  "List all Perl versions available for installation",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Available command not yet implemented")
		},
	}
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
						if i < completeChars {
							progressBar += "="
						} else if i == completeChars {
							progressBar += ">"
						} else {
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
			cmd.Println("Exec command not yet implemented")
		},
	}
}

func newUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall [version]",
		Short: "Remove a Perl version",
		Long:  "Uninstall a specific version of Perl",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Uninstall command not yet implemented")
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
	return &cobra.Command{
		Use:   "rehash",
		Short: "Rebuild shim executables",
		Long:  "Rebuild shim executables for all installed Perl versions",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Rehash command not yet implemented")
		},
	}
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

// newSymlinksCommand is implemented in symlinks.go

// newConfigCommand is implemented in config.go

// newSystemCommand creates a command for showing system Perl info
// This is now moved to perl.go as newPerlSystemCommand()

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
						if i < completeChars {
							progressBar += "="
						} else if i == completeChars {
							progressBar += ">"
						} else {
							progressBar += " "
						}
					}
					progressBar += "]"
					
					// Clear line and show progress
					fmt.Printf("\r%s %.1f%%                    ", 
						progressBar, progress * 100)
					
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
