// ABOUTME: Core CLI framework for PVM Ecosystem
// ABOUTME: Provides base command structure used by all components

package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/version"
)

var (
	// Verbose output flag
	Verbose bool

	// Debug mode flag
	Debug bool

	// Global UI instance for all commands
	globalUI *ui.Output
)

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

	// Add version command
	cmd.AddCommand(newVersionCommand(name))

	return cmd
}

// newVersionCommand creates a version command for the provided component
func newVersionCommand(component string) *cobra.Command {
	var showCurrent bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print detailed version information about this component",
		RunE: func(cmd *cobra.Command, args []string) error {
			if showCurrent {
				return showCurrentPerlVersion(cmd, component)
			}
			ui := GetUI(cmd)

			if Verbose {
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
			} else {
				// Show simple version in normal mode
				ui.Println(version.ComponentVersion(component))
			}
			return nil
		},
	}

	// Add --current flag for PVM component
	if component == "pvm" {
		cmd.Flags().BoolVar(&showCurrent, "current", false, "Show currently active Perl version")
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

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
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

	// Override the unknown command handling
	origRunE := rootCmd.RunE
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if origRunE != nil {
			return origRunE(cmd, args)
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

	// Create UI context based on command flags and environment
	ctx := &ui.UIContext{
		Writer:      writer,
		ErrorWriter: os.Stderr, // Errors should always go to stderr
		ColorMode:   ui.ColorAuto,
		Quiet:       false, // TODO: Add quiet flag when needed
		Verbose:     Verbose,
		Interactive: true, // TODO: Detect TTY
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

// showCurrentPerlVersion displays the currently active Perl version
func showCurrentPerlVersion(cmd *cobra.Command, component string) error {
	ui := GetUI(cmd)
	
	// Only show current version for PVM component
	if component != "pvm" {
		ui.Println(version.ComponentVersion(component))
		return nil
	}

	// Try to resolve the current version using the standard resolution algorithm
	resolvedVersion, err := perl.ResolveVersion(nil)
	if err != nil {
		return fmt.Errorf("failed to determine current Perl version: %w", err)
	}

	if resolvedVersion == nil {
		ui.Println("No Perl version is currently active")
		return nil
	}

	ui.Printf("%s (from %s)\n", resolvedVersion.Version, resolvedVersion.Source)
	return nil
}
