// ABOUTME: Core CLI framework for PVM Ecosystem
// ABOUTME: Provides base command structure used by all components

package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/current"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/version"
)

var (
	// Verbose output flag
	Verbose bool

	// Debug mode flag
	Debug bool

	// Quiet mode flag
	Quiet bool

	// RawMarkdown flag to disable styled markdown rendering
	RawMarkdown bool

	// Global UI instance for all commands
	globalUI *ui.Output
)

// ResetGlobalState resets all global CLI state
// This is useful for testing to prevent state leakage between tests
func ResetGlobalState() {
	Verbose = false
	Debug = false
	Quiet = false
	RawMarkdown = false
	globalUI = nil
	ResetGlobalRegistry()
}

// NewRootCommand creates a new root command for a component
func NewRootCommand(name string, description string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: description,
		Long:  description,
	}

	// Add global flags
	cmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "Enable debug mode")
	cmd.PersistentFlags().BoolVarP(&Quiet, "quiet", "q", false, "Enable quiet mode")
	cmd.PersistentFlags().BoolVar(&RawMarkdown, "raw-markdown", false, "Disable styled markdown rendering, use plain text")

	// Add version command
	cmd.AddCommand(newVersionCommand(name))

	return cmd
}

// newVersionCommand creates a version command for the provided component
func newVersionCommand(component string) *cobra.Command {
	var bare bool
	var showCurrent bool
	var detailed bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version information about Perl or PVM",
		RunE: func(cmd *cobra.Command, args []string) error {
			// For PVM component, show current Perl version only when explicitly requested
			if component == "pvm" && showCurrent {
				return showCurrentPerlVersionWithFlags(cmd, component, bare)
			}
			ui := GetUI(cmd)

			switch {
			case detailed:
				// Show detailed version information with release notes
				return showDetailedVersionInfo(cmd, component, ui)
			case Verbose:
				// Show detailed version information in verbose mode
				ui.Header(fmt.Sprintf("%s Version Information", component))

				buildInfo := version.GetBuildInfo()
				ui.KeyValue(map[string]string{
					"Version":    buildInfo["version"],
					"Build Time": buildInfo["buildTime"],
					"Commit":     buildInfo["commitHash"],
					"Go Version": buildInfo["goVersion"],
					"OS/Arch":    fmt.Sprintf("%s/%s", buildInfo["os"], buildInfo["arch"]),
				})
			default:
				// Show simple version in normal mode
				ui.Println(version.ComponentVersion(component))
			}
			return nil
		},
	}

	// Add flags for PVM component
	if component == "pvm" {
		cmd.Flags().BoolVar(&bare, "bare", false, "Show only the version string (for scripting)")
		cmd.Flags().BoolVar(&showCurrent, "current", false, "Show currently active Perl version")
		cmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed version information with release notes")
		cmd.Long = `Show PVM version information.

By default, this command shows PVM's own version.
Use --current to show which Perl version is currently active.

Examples:
  pvm version             # Show PVM's version (default)
  pvm version --current   # Show active Perl version
  pvm version --bare      # Show only version string for scripting
  pvm version --detailed  # Show detailed version with release notes`
	}

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(rootCmd *cobra.Command) {
	// Setup command pre-run hook to configure logging and UI
	origPreRun := rootCmd.PersistentPreRun
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Initialize UI framework for this command execution
		setupUI(cmd)

		// Set log level based on flags
		if Verbose {
			log.SetGlobalLevel(log.LevelDebug)
		} else {
			log.SetGlobalLevel(log.LevelInfo)
		}

		// Set component from command
		log.SetGlobalComponent(cmd.Root().Use)

		// Call the original pre-run if it exists
		if origPreRun != nil {
			origPreRun(cmd, args)
		}
	}

	// Enable command suggestions for typos
	enableCommandSuggestions(rootCmd)

	// Use Fang Execute with pager support
	if err := FangExecuteWithPager(context.Background(), rootCmd); err != nil {
		// Check if it's an unknown command error and provide suggestions
		if unknownCommandError, ok := err.(*UnknownCommandError); ok {
			handleUnknownCommand(rootCmd, unknownCommandError)
			os.Exit(1)
		}

		// Log the error using our structured logging
		errors.LogError(err)
		os.Exit(1)
	}
}

// UnknownCommandError represents an unknown command error with the command name
type UnknownCommandError struct {
	Command string
	Message string
}

func (e *UnknownCommandError) Error() string {
	return e.Message
}

// enableCommandSuggestions sets up command suggestion handling
func enableCommandSuggestions(rootCmd *cobra.Command) {
	// Set up cobra's suggestion settings
	rootCmd.SuggestionsMinimumDistance = 2
	rootCmd.SilenceUsage = true // Don't show usage on errors

	// Override the unknown command handling, but preserve existing Run/RunE functions
	origRunE := rootCmd.RunE
	origRun := rootCmd.Run
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if origRunE != nil {
			return origRunE(cmd, args)
		}
		if origRun != nil {
			origRun(cmd, args)
			return nil
		}
		return cmd.Help()
	}
}

// handleUnknownCommand provides enhanced error handling for unknown commands
func handleUnknownCommand(rootCmd *cobra.Command, err *UnknownCommandError) {
	// Get UI instance for error formatting
	ui := GetUI(rootCmd)
	ui.SetWriter(os.Stderr)

	ui.Error("unknown command '%s'", err.Command)
	ui.Println("")

	// Get available commands
	availableCommands := []string{}
	for _, cmd := range rootCmd.Commands() {
		if !cmd.Hidden {
			availableCommands = append(availableCommands, cmd.Name())
		}
	}

	// Get suggestions using our enhanced suggestion system
	suggestions := SuggestCommand(err.Command, availableCommands)

	if len(suggestions) > 0 {
		ui.Info("Did you mean?")
		for _, suggestion := range suggestions {
			ui.Printf("  pvm %s\n", suggestion)
		}
		ui.Println("")
	}

	ui.Info("Run 'pvm help' for usage information")
	ui.Info("Run 'pvm help workflows' for common workflows")
}

// setupUI initializes the UI framework for the given command
func setupUI(cmd *cobra.Command) {
	// Use command's output writer if available, otherwise default to os.Stdout
	var writer io.Writer = os.Stdout
	if cmd.OutOrStdout() != nil {
		writer = cmd.OutOrStdout()
	}

	// Parse raw-markdown flag - try inherited flags first, then global variable
	rawMarkdown := RawMarkdown
	if cmd.InheritedFlags().Lookup("raw-markdown") != nil {
		if flagValue, err := cmd.InheritedFlags().GetBool("raw-markdown"); err == nil {
			rawMarkdown = flagValue
		}
	}

	// Create UI context based on command flags and environment
	ctx := &ui.UIContext{
		Writer:      writer,
		ErrorWriter: os.Stderr, // Errors should always go to stderr
		ColorMode:   detectColorMode(),
		Quiet:       Quiet,
		Verbose:     Verbose,
		Interactive: isTerminal(os.Stdout),
		RawMarkdown: rawMarkdown,
	}

	// Create and store the UI instance
	globalUI = ui.NewOutput(ctx)
}

// GetUI returns the UI instance for the given command
// If no UI instance exists, creates a default one
func GetUI(cmd *cobra.Command) *ui.Output {
	if globalUI == nil {
		setupUI(cmd)
	}
	return globalUI
}

// FangExecuteWithPager wraps Fang's Execute function with pager support for help output
func FangExecuteWithPager(ctx context.Context, rootCmd *cobra.Command) error {
	// Get UI styles for Fang integration
	ui := GetUI(rootCmd)
	styles := ui.Styles()

	// Detect terminal background and create adapted Fang color scheme.
	// Only query the terminal when attached to a tty to avoid a 2-second
	// timeout in non-responsive environments (screen, some containers).
	isDark := true
	if isTerminal(os.Stdin) || isTerminal(os.Stdout) {
		isDark = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	}
	colorScheme := styles.FangColorScheme(isDark)

	// Check if we're showing basic help (no arguments) and use our hybrid help system instead of Fang's
	args := os.Args[1:]
	if isBasicHelpCommand(args) {
		ShowHybridHelp(rootCmd, args)
		return nil
	}

	// Use standard Fang execution
	err := fang.Execute(ctx, rootCmd,
		fang.WithTheme(colorScheme),
		fang.WithVersion(version.GetVersion()),
	)

	return err
}

// isHelpCommand checks if the command being executed is a help command
func isHelpCommand(args []string) bool {
	if len(args) == 0 {
		return false
	}

	// Check for help flags
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}

	// Check for help command
	if args[0] == "help" {
		return true
	}

	return false
}

// isBasicHelpCommand checks if the command is basic help (no specific topic)
func isBasicHelpCommand(args []string) bool {
	if len(args) == 0 {
		return false
	}

	// Check for help flags at the root level (no other commands)
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		return true
	}

	// Check for help command without arguments
	if args[0] == "help" && len(args) == 1 {
		return true
	}

	return false
}

// shouldEnablePager checks if pager should be enabled based on environment and content size
func shouldEnablePager() bool {
	// Don't use pager if output is being redirected
	if !isTerminal(os.Stdout) {
		return false
	}

	// Check if PAGER is explicitly disabled
	if os.Getenv("PAGER") == "cat" || os.Getenv("NO_PAGER") != "" {
		return false
	}

	return true
}

// shouldEnablePagerForOutput checks if pager should be enabled based on content size vs terminal size
func shouldEnablePagerForOutput(contentLines int) bool {
	// Don't use pager if output is being redirected
	if !isTerminal(os.Stdout) {
		return false
	}

	// Check if PAGER is explicitly disabled
	if os.Getenv("PAGER") == "cat" || os.Getenv("NO_PAGER") != "" {
		return false
	}

	// Get terminal size
	terminalHeight := getTerminalHeight()
	if terminalHeight <= 0 {
		// Can't determine terminal size, default to no pager for short content
		return contentLines > 25 // Conservative fallback
	}

	// Use pager if content would overflow terminal (leave some buffer for prompt)
	return contentLines > (terminalHeight - 3)
}

// isTerminal checks if the given file is a terminal
func isTerminal(f *os.File) bool {
	fileInfo, err := f.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// detectColorMode determines the color mode based on environment variables
// and terminal capabilities. Respects the NO_COLOR (https://no-color.org/)
// and CLICOLOR/CLICOLOR_FORCE conventions.
func detectColorMode() ui.ColorMode {
	// NO_COLOR takes precedence when set (any value, including empty)
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return ui.ColorNever
	}
	// CLICOLOR_FORCE=1 forces color even when not a TTY
	if os.Getenv("CLICOLOR_FORCE") == "1" {
		return ui.ColorAlways
	}
	// CLICOLOR=0 disables color
	if os.Getenv("CLICOLOR") == "0" {
		return ui.ColorNever
	}
	return ui.ColorAuto
}

// getTerminalHeightFromEnv gets terminal height from environment variables
func getTerminalHeightFromEnv() int {
	if lines := os.Getenv("LINES"); lines != "" {
		if h, err := strconv.Atoi(lines); err == nil && h > 0 {
			return h
		}
	}

	// Try tput as last resort
	if cmd := exec.Command("tput", "lines"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			if h, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil && h > 0 {
				return h
			}
		}
	}

	return 0 // Unknown terminal size
}

// executeWithPagerIfNeeded pre-renders output and uses pager only if content overflows terminal
func executeWithPagerIfNeeded(ctx context.Context, rootCmd *cobra.Command, colorScheme fang.ColorScheme) error {
	// First, capture the output to determine if we need a pager
	tmpFile, err := os.CreateTemp("", "pvm-help-check-*")
	if err != nil {
		// Fall back to standard execution if temp file creation fails
		return fang.Execute(ctx, rootCmd,
			fang.WithTheme(colorScheme),
			fang.WithVersion(version.GetVersion()),
		)
	}
	defer os.Remove(tmpFile.Name())

	// Temporarily redirect stdout to temp file to capture output
	originalStdout := os.Stdout
	os.Stdout = tmpFile

	// Execute with Fang to capture the output
	err = fang.Execute(ctx, rootCmd,
		fang.WithTheme(colorScheme),
		fang.WithVersion(version.GetVersion()),
	)

	// Restore stdout
	os.Stdout = originalStdout
	tmpFile.Close()

	if err != nil {
		return err
	}

	// Read the captured output and count lines
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	lineCount := len(lines)

	// Decide whether to use pager based on content size vs terminal size
	if shouldEnablePagerForOutput(lineCount) {
		return executeWithPagerFromContent(ctx, string(content), colorScheme)
	}

	// Content fits in terminal, output directly
	fmt.Print(string(content))
	return nil
}

// executeWithPagerFromContent pipes pre-rendered content through a pager
func executeWithPagerFromContent(ctx context.Context, content string, colorScheme fang.ColorScheme) error {
	// Get pager command
	pagerCmd := getPagerCommand()
	if pagerCmd == "" {
		// No pager available, output content directly
		fmt.Print(content)
		return nil
	}

	// Create pager process
	pager := exec.CommandContext(ctx, "sh", "-c", pagerCmd)
	pager.Stdout = os.Stdout
	pager.Stderr = os.Stderr

	// Create pipe to pager
	pagerInput, err := pager.StdinPipe()
	if err != nil {
		// Fall back to direct output if pager setup fails
		fmt.Print(content)
		return nil
	}

	// Start pager
	if err := pager.Start(); err != nil {
		pagerInput.Close()
		// Fall back to direct output
		fmt.Print(content)
		return nil
	}

	// Send content to pager
	_, err = pagerInput.Write([]byte(content))
	pagerInput.Close()

	// Wait for pager to finish
	pagerErr := pager.Wait()
	if pagerErr != nil {
		// If pager failed but we already sent content, that's ok
		// This handles cases where user exits pager early (SIGPIPE)
	}

	return err
}

// executeWithPager executes the command and pipes help output through a pager
func executeWithPager(ctx context.Context, rootCmd *cobra.Command, colorScheme fang.ColorScheme) error {
	// Get pager command
	pagerCmd := getPagerCommand()
	if pagerCmd == "" {
		// No pager available, use standard execution
		return fang.Execute(ctx, rootCmd,
			fang.WithTheme(colorScheme),
			fang.WithVersion(version.ComponentVersion(rootCmd.Use)),
		)
	}

	// Create pager process
	pager := exec.CommandContext(ctx, "sh", "-c", pagerCmd)
	pager.Stdout = os.Stdout
	pager.Stderr = os.Stderr

	// Create pipe to pager
	pagerInput, err := pager.StdinPipe()
	if err != nil {
		// Fall back to standard execution if pager setup fails
		return fang.Execute(ctx, rootCmd,
			fang.WithTheme(colorScheme),
			fang.WithVersion(version.ComponentVersion(rootCmd.Use)),
		)
	}

	// Start pager
	if err := pager.Start(); err != nil {
		pagerInput.Close()
		// Fall back to standard execution
		return fang.Execute(ctx, rootCmd,
			fang.WithTheme(colorScheme),
			fang.WithVersion(version.ComponentVersion(rootCmd.Use)),
		)
	}

	// Create a temporary file to capture stdout
	tmpFile, err := os.CreateTemp("", "pvm-help-*")
	if err != nil {
		pagerInput.Close()
		return fang.Execute(ctx, rootCmd,
			fang.WithTheme(colorScheme),
			fang.WithVersion(version.GetVersion()),
		)
	}
	defer os.Remove(tmpFile.Name())

	// Temporarily redirect stdout to temp file
	originalStdout := os.Stdout
	os.Stdout = tmpFile

	// Execute with Fang
	err = fang.Execute(context.WithValue(ctx, "pager", true), rootCmd,
		fang.WithTheme(colorScheme),
		fang.WithVersion(version.GetVersion()),
	)

	// Restore stdout
	os.Stdout = originalStdout
	tmpFile.Close()

	// Copy temp file content to pager if command executed successfully
	if err == nil {
		tmpFile, readErr := os.Open(tmpFile.Name())
		if readErr == nil {
			_, copyErr := io.Copy(pagerInput, tmpFile)
			tmpFile.Close()
			if copyErr != nil {
				// If copy failed, fall back to standard output
				pagerInput.Close()
				pager.Wait()
				return fang.Execute(ctx, rootCmd,
					fang.WithTheme(colorScheme),
					fang.WithVersion(version.GetVersion()),
				)
			}
		}
	}

	pagerInput.Close()

	// Wait for pager to finish
	pagerErr := pager.Wait()
	if pagerErr != nil && err == nil {
		// If command succeeded but pager failed, don't treat it as an error
		// This handles cases where user exits pager early (SIGPIPE)
	}

	return err
}

// getPagerCommand returns the pager command to use
func getPagerCommand() string {
	// Check PAGER environment variable first
	if pager := os.Getenv("PAGER"); pager != "" && pager != "cat" {
		return pager
	}

	// Try common pagers in order of preference
	pagers := []string{"less -R", "more", "cat"}

	for _, pager := range pagers {
		// Extract command name (first word) to check if it exists
		cmdName := pager
		if idx := strings.IndexByte(pager, ' '); idx != -1 {
			cmdName = pager[:idx]
		}

		if _, err := exec.LookPath(cmdName); err == nil {
			return pager
		}
	}

	return ""
}

// PrintWithPager outputs content using pager if it would overflow terminal
func PrintWithPager(content string) {
	lines := strings.Split(content, "\n")
	lineCount := len(lines)

	if shouldEnablePagerForOutput(lineCount) {
		// Use pager for long content
		pagerCmd := getPagerCommand()
		if pagerCmd != "" {
			pager := exec.Command("sh", "-c", pagerCmd)
			pager.Stdout = os.Stdout
			pager.Stderr = os.Stderr

			pagerInput, err := pager.StdinPipe()
			if err == nil {
				if pager.Start() == nil {
					pagerInput.Write([]byte(content))
					pagerInput.Close()
					pager.Wait()
					return
				}
			}
		}
	}

	// Fall back to direct output
	fmt.Print(content)
}

// showCurrentPerlVersion displays the currently active Perl version (legacy function)
func showCurrentPerlVersion(cmd *cobra.Command, component string) error {
	return showCurrentPerlVersionWithFlags(cmd, component, false)
}

// showCurrentPerlVersionWithFlags displays the currently active Perl version with flag support
func showCurrentPerlVersionWithFlags(cmd *cobra.Command, component string, bare bool) error {
	ui := GetUI(cmd)

	// Only show current version for PVM component
	if component != "pvm" {
		ui.Println(version.ComponentVersion(component))
		return nil
	}

	// Use the current package for consistent formatting
	info, err := current.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Format output using the current package for consistency
	options := current.DefaultDisplayOptions()
	if bare {
		options = current.BareDisplayOptions()
	}

	output, err := current.FormatCurrentVersion(info, options)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	ui.Printf("%s", output)

	// Add newline for non-bare output (bare output doesn't include newline)
	if !bare && info.IsAvailable {
		ui.Println()
	}

	return nil
}

// showDetailedVersionInfo displays detailed version information with release notes
func showDetailedVersionInfo(cmd *cobra.Command, component string, ui *ui.Output) error {
	// Show basic version information
	ui.Header(fmt.Sprintf("%s Detailed Version Information", component))

	buildInfo := version.GetBuildInfo()
	ui.KeyValue(map[string]string{
		"Version":    buildInfo["version"],
		"Build Time": buildInfo["buildTime"],
		"Commit":     buildInfo["commitHash"],
		"Go Version": buildInfo["goVersion"],
		"OS/Arch":    fmt.Sprintf("%s/%s", buildInfo["os"], buildInfo["arch"]),
	})

	// Only show release notes for PVM component
	if component != "pvm" {
		return nil
	}

	// Get release notes for the current version
	client := version.NewGitHubClient()
	releaseInfo, err := client.GetReleaseByTag("perigrin", "pvm", buildInfo["version"])
	if err != nil {
		ui.Warning("Could not fetch release notes: %v", err)
		return nil
	}

	// Display release notes with glow formatting
	if releaseInfo.Body != "" {
		ui.SubHeader("Release Notes")
		ui.GlowMarkdown(releaseInfo.Body)
	} else {
		ui.Info("No release notes available for version %s", buildInfo["version"])
	}

	return nil
}
