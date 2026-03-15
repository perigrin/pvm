// ABOUTME: PVM Perl detection and management commands
// ABOUTME: Provides commands for detecting and managing Perl installations

package pvm

import (
	"archive/tar"
	"compress/gzip"
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
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/platform"
)

// newPerlCommand creates a new perl command
func newPerlCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perl",
		Short: "Manage Perl installations",
		Long:  "Detect, install, and manage Perl versions",
	}

	// Add subcommands
	cmd.AddCommand(
		newPerlSystemCommand(),
		newPerlBuildCommand(),
		newPerlTarballCommand(),
		newPerlExecCommand(),
		newPerlExecAllCommand(),
		newPerlDownloadCommand(),
		newPerlInstallCommand(),
		newPerlUploadCommand(),
		newPerlGlobalCommand(),
		newPerlLocalCommand(),       // Moved from main commands
		newPerlUseCommand(),         // Moved from main commands
		newPerlResolveCommand(),     // Moved from main commands
		newPerlVersionUtilCommand(), // Moved from top-level version-util
	)

	return cmd
}

// newPerlSystemCommand creates a command to show system Perl installations
func newPerlSystemCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "system",
		Short: "Show system Perl installations",
		Long:  "Detect and display information about system Perl installations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			perls, err := perl.DetectAllSystemPerls()
			if err != nil {
				ui.Warning("Detection warning: %v", err)
			}

			if len(perls) == 0 {
				ui.Info("No system Perl installations found.")
				return nil
			}

			// Create table data
			headers := []string{"PATH", "VERSION", "ARCHITECTURE", "PRIMARY"}
			rows := make([][]string, len(perls))
			for i, p := range perls {
				rows[i] = []string{
					p.Path,
					p.Version,
					p.Architecture,
					fmt.Sprintf("%v", p.IsPrimary),
				}
			}

			ui.Table(headers, rows)
			return nil
		},
	}
}

// newPerlBuildCommand creates a command to build Perl from source
func newPerlBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [version|URL]",
		Short: "Build Perl from source, URL, or version",
		Long:  "Build and install Perl from source code, direct URL, or official version release",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			versionOrURL := args[0]
			return buildPerlFromSource(cmd, versionOrURL)
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

	// Perl configuration flags
	cmd.Flags().Bool("relocatable", false, "Build relocatable Perl (enables -Duserelocatableinc)")
	cmd.Flags().Bool("shared-lib", true, "Build shared libperl (enables -Duseshrplib)")

	// Upload integration flags
	cmd.Flags().Bool("upload", false, "Upload built binary after successful build")
	cmd.Flags().StringArray("platforms", nil, "Build for multiple platforms (e.g., linux-amd64,darwin-arm64)")
	cmd.Flags().String("mirror", "", "Specific mirror to upload to (default: all configured mirrors)")
	cmd.Flags().String("github-token", "", "GitHub API token for upload authentication")
	cmd.Flags().String("github-repo", "", "GitHub repository for upload (format: owner/repo)")
	cmd.Flags().String("release-tag", "", "GitHub release tag (created if doesn't exist)")
	cmd.Flags().Bool("draft-release", false, "Create release as draft when uploading")
	cmd.Flags().Bool("prerelease", false, "Mark release as prerelease when uploading")

	return cmd
}

// newPerlTarballCommand creates a command to create tarballs from existing Perl installations
func newPerlTarballCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tarball [version]",
		Short: "Create tarball from existing Perl installation",
		Long: `Create a compressed tarball (.tar.gz) from an existing Perl installation.
This command addresses Build Perl Binaries workflow issues by providing a reliable,
platform-consistent way to create tarballs using Go's native archive handling.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			return createPerlTarball(cmd, version)
		},
	}

	// Tarball creation flags
	cmd.Flags().String("output", "", "Output tarball path (default: perl-<version>-<platform>.tar.gz)")
	cmd.Flags().Int("compression-level", 6, "Gzip compression level (1-9, default: 6)")
	cmd.Flags().StringArray("exclude", []string{"*.log", "*.tmp", ".pvm-*"}, "File patterns to exclude from tarball")
	cmd.Flags().Bool("verify", true, "Verify Perl installation before creating tarball")

	return cmd
}

// createPerlTarball creates a tarball from an existing Perl installation
func createPerlTarball(cmd *cobra.Command, version string) error {
	ui := cli.GetUI(cmd)

	// Validate that the Perl version is installed
	if verify, _ := cmd.Flags().GetBool("verify"); verify {
		installed, err := perl.IsVersionInstalled(version)
		if err != nil {
			return fmt.Errorf("failed to check if version is installed: %w", err)
		}
		if !installed {
			return fmt.Errorf("Perl version %s is not installed", version)
		}
	}

	// Get the installation directory
	installDir := filepath.Join(platform.DataDir(), "pvm", "versions", version)

	// Verify the installation directory exists and has a perl binary
	perlBinary := filepath.Join(installDir, "bin", "perl")
	if _, err := os.Stat(perlBinary); os.IsNotExist(err) {
		return fmt.Errorf("perl binary not found at expected location: %s", perlBinary)
	}

	// Get output path
	outputPath, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}
	if outputPath == "" {
		platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
		outputPath = fmt.Sprintf("perl-%s-%s.tar.gz", version, platform)
	}

	// Get compression level
	compressionLevel, err := cmd.Flags().GetInt("compression-level")
	if err != nil {
		return err
	}
	if compressionLevel < 1 || compressionLevel > 9 {
		return fmt.Errorf("compression level must be between 1 and 9, got %d", compressionLevel)
	}

	// Get exclusion patterns
	excludePatterns, err := cmd.Flags().GetStringArray("exclude")
	if err != nil {
		return err
	}

	ui.Status("Creating tarball from Perl installation...")
	ui.Info("Version: %s", version)
	ui.Info("Source directory: %s", installDir)
	ui.Info("Output path: %s", outputPath)

	// Create the enhanced tarball
	err = createEnhancedTarGzArchive(installDir, outputPath, compressionLevel, excludePatterns, ui)
	if err != nil {
		return fmt.Errorf("failed to create tarball: %w", err)
	}

	// Get file info for success message
	if fileInfo, err := os.Stat(outputPath); err == nil {
		ui.Success("Successfully created tarball: %s (%.2f MB)", outputPath, float64(fileInfo.Size())/(1024*1024))
	} else {
		ui.Success("Successfully created tarball: %s", outputPath)
	}

	return nil
}

// createEnhancedTarGzArchive creates a tar.gz archive with enhanced features for addressing workflow issues
func createEnhancedTarGzArchive(sourceDir, outputPath string, compressionLevel int, excludePatterns []string, ui *ui.Output) error {
	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer file.Close()

	// Create gzip writer with specified compression level
	gzipWriter, err := gzip.NewWriterLevel(file, compressionLevel)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	fileCount := 0

	// Walk the source directory
	err = filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			// Handle file access errors gracefully
			ui.Warning("Skipping file due to access error: %s (%v)", filePath, err)
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Check exclusion patterns
		for _, pattern := range excludePatterns {
			if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
				ui.Info("Excluding: %s (matches pattern: %s)", relPath, pattern)
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Create tar header - use a snapshot of file info to avoid "file changed as we read it" errors
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Set the name to the relative path
		header.Name = relPath

		// For regular files, ensure we have the size at the time of header creation
		if info.Mode().IsRegular() {
			// Re-stat the file to get current size to minimize timing issues
			if currentInfo, statErr := os.Stat(filePath); statErr == nil {
				header.Size = currentInfo.Size()
			}
		}

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content for regular files
		if info.Mode().IsRegular() {
			file, err := os.Open(filePath)
			if err != nil {
				ui.Warning("Skipping file due to open error: %s (%v)", relPath, err)
				return nil
			}
			defer file.Close()

			// Copy file content with error handling
			_, err = io.Copy(tarWriter, file)
			if err != nil {
				ui.Warning("Error copying file content: %s (%v)", relPath, err)
				return nil
			}
		}

		fileCount++
		if fileCount%100 == 0 {
			ui.Status("Processed files...")
		}

		return nil
	})
	if err != nil {
		return err
	}

	ui.Status("Successfully archived files")
	return nil
}

// newPerlExecCommand creates a command to execute commands with specific Perl versions
func newPerlExecCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "exec [version] [command]",
		Short: "Execute a command with a specific version",
		Long:  "Execute a command using a specific Perl version",
		Run: func(cmd *cobra.Command, args []string) {
			ui := cli.GetUI(cmd)
			if len(args) < 1 {
				ui.Error("Usage: pvm perl exec [version] [command]")
				return
			}

			var version string
			var command []string

			if len(args) == 1 {
				// Only command provided, use current version
				command = args
			} else {
				// Check if first arg looks like a version or a command
				firstArg := args[0]
				if isLikelyVersion(firstArg) {
					// First arg is a version
					version = firstArg
					command = args[1:]
				} else {
					// First arg is probably a command, use current version
					command = args
				}
			}

			err := execCommand(cmd, version, command)
			if err != nil {
				ui.Error("Error: %v", err)
				os.Exit(1)
			}
		},
	}
}

// isLikelyVersion checks if a string looks like a Perl version
func isLikelyVersion(arg string) bool {
	// Check for common version patterns
	// 5.38.0, 5.40.1, etc.
	if matched, _ := filepath.Match("[0-9]*.[0-9]*", arg); matched {
		return true
	}
	// Handle version aliases like @latest, @stable
	if strings.HasPrefix(arg, "@") {
		return true
	}
	// Handle special identifiers
	if arg == "system" || arg == "latest" {
		return true
	}
	// Common commands that are definitely NOT versions
	commonCommands := []string{"perl", "cpan", "prove", "perldoc", "cpanm"}
	for _, cmd := range commonCommands {
		if arg == cmd {
			return false
		}
	}
	return false
}

// execCommand executes a command with the specified Perl version
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

	execCmd := exec.Command(command[0], command[1:]...)
	execCmd.Env = env
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

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

// newPerlDownloadCommand creates a perl download command (wraps main download command)
func newPerlDownloadCommand() *cobra.Command {
	return newDownloadCommand()
}

// newPerlInstallCommand creates a perl install command (wraps main install-perl command)
func newPerlInstallCommand() *cobra.Command {
	cmd := newInstallPerlCommand()
	cmd.Use = "install [build-dir]"
	return cmd
}

// newPerlUploadCommand creates a perl upload command (wraps main upload-binary command)
func newPerlUploadCommand() *cobra.Command {
	cmd := newUploadBinaryCommand()
	cmd.Use = "upload [archive-path]"
	return cmd
}

// newPerlGlobalCommand creates a perl global command (wraps main global command)
func newPerlGlobalCommand() *cobra.Command {
	cmd := newGlobalCommand()
	cmd.Use = "global [version]"
	return cmd
}

// newPerlLocalCommand creates a perl local command (wraps main local command)
func newPerlLocalCommand() *cobra.Command {
	cmd := newLocalCommand()
	cmd.Use = "local [version]"
	return cmd
}

// newPerlUseCommand creates a perl use command (wraps main use command)
func newPerlUseCommand() *cobra.Command {
	cmd := newUseCommand()
	cmd.Use = "use [version]"
	return cmd
}

// newPerlResolveCommand creates a perl resolve command (wraps main resolve command)
func newPerlResolveCommand() *cobra.Command {
	cmd := newResolveCommand()
	cmd.Use = "resolve [version]"
	return cmd
}

// ExecAllResult represents the result of executing a command with a specific Perl version
type ExecAllResult struct {
	Version  string
	ExitCode int
	Output   string
	Error    string
	Duration time.Duration
}

// newPerlExecAllCommand creates a command to execute commands with all installed Perl versions
func newPerlExecAllCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec-all [command...]",
		Short: "Execute a command with all installed Perl versions",
		Long: `Execute a command using all installed Perl versions sequentially.

This command runs the specified command with each installed Perl version
and provides a summary of the results.

Examples:
  pvm perl exec-all prove t/basic.t
  pvm perl exec-all perl -MData::Dumper -e 'print $Data::Dumper::VERSION'
  pvm perl exec-all perl my-script.pl`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no command specified")
			}
			return execAllCommand(cmd, args)
		},
		DisableFlagParsing: true, // Pass all arguments through to the executed command
	}
	return cmd
}

// execAllCommand executes a command with all installed Perl versions
func execAllCommand(cmd *cobra.Command, args []string) error {
	ui := cli.GetUI(cmd)

	// Get all installed versions
	installedVersions, err := perl.GetInstalledVersions()
	if err != nil {
		return fmt.Errorf("failed to get installed versions: %w", err)
	}

	if len(installedVersions) == 0 {
		ui.Warning("No Perl versions are installed.")
		return nil
	}

	ui.Info("Running command with %d installed Perl versions...\n", len(installedVersions))

	// Execute command with each version
	var results []ExecAllResult
	for _, versionInfo := range installedVersions {
		result := executeWithVersion(versionInfo.Version, args, ui)
		results = append(results, result)
	}

	// Display summary
	displayExecAllSummary(results, ui)

	// Determine overall exit code
	return determineExitCode(results)
}

// executeWithVersion executes a command with a specific Perl version
func executeWithVersion(version string, args []string, ui *ui.Output) ExecAllResult {
	start := time.Now()

	// Display version header
	ui.Info("=== Running with Perl %s ===", version)

	// Resolve version
	options := &perl.ResolutionOptions{
		ExplicitVersion: version,
	}

	resolved, err := perl.ResolveVersion(options)
	if err != nil {
		ui.Error("Failed to resolve version %s: %v", version, err)
		return ExecAllResult{
			Version:  version,
			ExitCode: 1,
			Error:    fmt.Sprintf("Failed to resolve version: %v", err),
			Duration: time.Since(start),
		}
	}

	// Check if Perl binary exists
	if resolved.Path == "" {
		err := fmt.Errorf("no path found for Perl version %s", version)
		ui.Error("No path found for Perl version %s", version)
		return ExecAllResult{
			Version:  version,
			ExitCode: 1,
			Error:    err.Error(),
			Duration: time.Since(start),
		}
	}

	if _, err := os.Stat(resolved.Path); os.IsNotExist(err) {
		ui.Error("Perl version %s not found at %s", version, resolved.Path)
		return ExecAllResult{
			Version:  version,
			ExitCode: 1,
			Error:    fmt.Sprintf("Perl not found at %s", resolved.Path),
			Duration: time.Since(start),
		}
	}

	// Prepare environment
	env := os.Environ()
	perlBinDir := strings.TrimSuffix(resolved.Path, "/perl")
	if perlBinDir == resolved.Path {
		perlBinDir = resolved.Path[:strings.LastIndex(resolved.Path, "/")]
	}

	// Update PATH
	for i, envVar := range env {
		if strings.HasPrefix(envVar, "PATH=") {
			currentPath := envVar[5:]
			env[i] = fmt.Sprintf("PATH=%s:%s", perlBinDir, currentPath)
			break
		}
	}

	// Prepare command
	command := make([]string, len(args))
	copy(command, args)
	if command[0] == "perl" {
		command[0] = resolved.Path
	}

	// Execute command
	execCmd := exec.Command(command[0], command[1:]...)
	execCmd.Env = env

	output, err := execCmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			} else {
				exitCode = 1
			}
		} else {
			exitCode = 1
		}
	}

	// Display output
	if len(output) > 0 {
		fmt.Print(string(output))
	}

	// Add blank line for readability unless this is the last version
	fmt.Println()

	return ExecAllResult{
		Version:  version,
		ExitCode: exitCode,
		Output:   string(output),
		Duration: time.Since(start),
	}
}

// displayExecAllSummary displays a summary of all execution results
func displayExecAllSummary(results []ExecAllResult, ui *ui.Output) {
	ui.Info("=== Summary ===")

	successCount := 0
	for _, result := range results {
		if result.ExitCode == 0 {
			ui.Success("✓ %s: SUCCESS", result.Version)
			successCount++
		} else {
			ui.Error("✗ %s: FAILED (exit code %d)", result.Version, result.ExitCode)
		}
	}

	// Overall result
	if successCount == len(results) {
		ui.Success("All versions completed successfully.")
	} else {
		failCount := len(results) - successCount
		ui.Error("%d of %d versions failed.", failCount, len(results))
	}
}

// determineExitCode determines the overall exit code based on results
func determineExitCode(results []ExecAllResult) error {
	for _, result := range results {
		if result.ExitCode != 0 {
			os.Exit(1)
		}
	}
	return nil
}
