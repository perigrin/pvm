// ABOUTME: Core CLI framework for PVM Ecosystem
// ABOUTME: Provides base command structure used by all components

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
	var showPVM bool
	var bare bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version information about Perl or PVM",
		RunE: func(cmd *cobra.Command, args []string) error {
			// For PVM component, default behavior is to show Perl version
			if component == "pvm" && !showPVM {
				return showCurrentPerlVersionWithFlags(cmd, component, bare)
			}
			fmt.Println(version.ComponentVersion(component))
			return nil
		},
	}

	// Add flags for PVM component
	if component == "pvm" {
		cmd.Flags().BoolVar(&showPVM, "pvm", false, "Show PVM version instead of active Perl version")
		cmd.Flags().BoolVar(&bare, "bare", false, "Show only the version string (for scripting)")
		cmd.Long = `Show the currently active Perl version.

By default, this command shows which Perl version is currently active.
Use --pvm to show PVM's own version instead.

Examples:
  pvm version         # Show active Perl version (default)
  pvm version --pvm   # Show PVM's version
  pvm version --bare  # Show only version string for scripting`
	}

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(rootCmd *cobra.Command) {
	// Setup command pre-run hook to configure logging
	origPreRun := rootCmd.PersistentPreRun
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
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
	fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n\n", err.Command)

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
		fmt.Fprintf(os.Stderr, "Did you mean?\n")
		for _, suggestion := range suggestions {
			fmt.Fprintf(os.Stderr, "  pvm %s\n", suggestion)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	fmt.Fprintf(os.Stderr, "Run 'pvm help' for usage information\n")
	fmt.Fprintf(os.Stderr, "Run 'pvm help workflows' for common workflows\n")
}

// showCurrentPerlVersion displays the currently active Perl version (legacy function)
func showCurrentPerlVersion(cmd *cobra.Command, component string) error {
	return showCurrentPerlVersionWithFlags(cmd, component, false)
}

// showCurrentPerlVersionWithFlags displays the currently active Perl version with flag support
func showCurrentPerlVersionWithFlags(cmd *cobra.Command, component string, bare bool) error {
	// Only show current version for PVM component
	if component != "pvm" {
		fmt.Println(version.ComponentVersion(component))
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

	cmd.Print(output)

	// Add newline for non-bare output (bare output doesn't include newline)
	if !bare && info.IsAvailable {
		cmd.Println()
	}

	return nil
}
