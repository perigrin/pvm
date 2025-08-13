// ABOUTME: Provides compatibility interfaces for traditional Perl ecosystem tools
// ABOUTME: Maps cpanm, carton, perlbrew, and plenv commands to PVM equivalents

package compat

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
)

// CommandMapper defines the interface for mapping external tool commands to PVM
type CommandMapper interface {
	// MapCommand transforms external tool command to PVM equivalent
	MapCommand(args []string) (pvmCmd string, pvmArgs []string, err error)
	// GetHelp returns help text in the style of the original tool
	GetHelp() string
	// GetToolName returns the name of the tool being emulated
	GetToolName() string
}

// BaseMapper provides common functionality for all compatibility mappers
type BaseMapper struct {
	toolName    string
	description string
}

// NewBaseMapper creates a new base mapper
func NewBaseMapper(toolName, description string) *BaseMapper {
	return &BaseMapper{
		toolName:    toolName,
		description: description,
	}
}

// GetToolName returns the tool name
func (b *BaseMapper) GetToolName() string {
	return b.toolName
}

// createCompatCommand creates a cobra command that wraps a CommandMapper
func createCompatCommand(mapper CommandMapper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   mapper.GetToolName(),
		Short: fmt.Sprintf("%s compatibility interface for PVM", mapper.GetToolName()),
		Long:  fmt.Sprintf("%s compatibility interface for PVM\n\nUse '%s --help' for usage details.", mapper.GetToolName(), mapper.GetToolName()),
		// Disable default help and usage to match original tools
		DisableFlagParsing: true,
		// Allow any arguments to be passed through
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check for help flags first
			for _, arg := range args {
				if arg == "-h" || arg == "--help" || arg == "help" {
					fmt.Print(mapper.GetHelp())
					return nil
				}
			}

			// If no arguments provided, show help
			if len(args) == 0 {
				fmt.Print(mapper.GetHelp())
				return nil
			}

			// Map the command
			pvmCmd, pvmArgs, err := mapper.MapCommand(args)
			if err != nil {
				return fmt.Errorf("%s compatibility: %w", mapper.GetToolName(), err)
			}

			// Handle empty mapping (like help commands)
			if pvmCmd == "" {
				return nil
			}

			// Execute the PVM command
			return executePVMCommand(pvmCmd, pvmArgs)
		},
	}

	return cmd
}

// executePVMCommand executes a PVM command with the given arguments
func executePVMCommand(command string, args []string) error {
	// Get the appropriate component from the registry
	registry := cli.GlobalRegistry
	if registry == nil {
		return fmt.Errorf("CLI registry not initialized")
	}

	// Build the full command path
	fullArgs := append([]string{command}, args...)

	// Create a new root command for the appropriate component
	var rootCmd *cobra.Command

	// Determine which component to use based on the command
	// PM commands (from cpanm, carton module operations)
	pmCommands := []string{"install", "list", "search", "add", "sync", "update", "remove", "deps", "bundle", "type", "mirror", "outdated", "info"}
	isPMCommand := false
	for _, pmCmd := range pmCommands {
		if command == pmCmd {
			isPMCommand = true
			break
		}
	}

	switch {
	case isPMCommand:
		// Use PM component for module-related commands
		if pmCmd, exists := registry.Get("pm"); exists {
			rootCmd = pmCmd()
		}
	case command == "workspace" || strings.HasPrefix(command, "ws"):
		// Use PVM component for workspace commands
		if pvmCmd, exists := registry.Get("pvm"); exists {
			rootCmd = pvmCmd()
		}
	default:
		// Default to PVM component for version management commands
		if pvmCmd, exists := registry.Get("pvm"); exists {
			rootCmd = pvmCmd()
		}
	}

	if rootCmd == nil {
		return fmt.Errorf("failed to get command for: %s", command)
	}

	// Set the arguments and execute
	rootCmd.SetArgs(fullArgs)
	return rootCmd.Execute()
}

// parseVersionString handles different version string formats
// perlbrew uses: perl-5.38.0, 5.38.0, perl-5.38
// plenv uses: 5.38.0, 5.38
func parseVersionString(version string) string {
	// Remove 'perl-' prefix if present
	version = strings.TrimPrefix(version, "perl-")

	// Ensure we have a full version (add .0 if needed)
	parts := strings.Split(version, ".")
	if len(parts) == 2 {
		version += ".0"
	}

	return version
}

// transformModuleArgs transforms module-related arguments
func transformModuleArgs(args []string) ([]string, error) {
	var result []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-v" || arg == "--verbose":
			result = append(result, "--verbose")
		case arg == "-q" || arg == "--quiet":
			result = append(result, "--quiet")
		case arg == "-f" || arg == "--force":
			result = append(result, "--force")
		case arg == "-n" || arg == "--notest":
			result = append(result, "--no-test")
		case arg == "-L" || arg == "--local-lib-contained":
			if i+1 < len(args) {
				result = append(result, "--install-dir", args[i+1])
				i++
			}
		case strings.HasPrefix(arg, "--local-lib="):
			path := strings.TrimPrefix(arg, "--local-lib=")
			result = append(result, "--local-lib", path)
		case arg == "--installdeps":
			result = append(result, "--deps-only")
		case arg == "--showdeps":
			// This needs special handling - show deps instead of install
			return []string{"info", "--dependencies"}, nil
		case arg == "--self-upgrade":
			// Map to PVM self-update
			return []string{"self-update"}, nil
		case !strings.HasPrefix(arg, "-"):
			// Module name or path
			result = append(result, arg)
		default:
			// Pass through unknown arguments
			result = append(result, arg)
		}
	}

	return result, nil
}

// ComponentNameMapping maps tool names to PVM component names for registration
var ComponentNameMapping = map[string]string{
	"cpanm":    "cpanm",
	"carton":   "carton",
	"perlbrew": "perlbrew",
	"plenv":    "plenv",
}

// RegisterComponents registers all compatibility components with the CLI registry
func RegisterComponents() {
	registry := cli.GlobalRegistry
	if registry == nil {
		return
	}

	// Register cpanm
	registry.Register("cpanm", func() *cobra.Command {
		return NewCpanmCommand()
	})

	// Register carton
	registry.Register("carton", func() *cobra.Command {
		return NewCartonCommand()
	})

	// Register perlbrew
	registry.Register("perlbrew", func() *cobra.Command {
		return NewPerlbrewCommand()
	})

	// Register plenv
	registry.Register("plenv", func() *cobra.Command {
		return NewPlenvCommand()
	})
}

// IsCompatibilityCommand checks if the given command name is a compatibility command
func IsCompatibilityCommand(name string) bool {
	_, exists := ComponentNameMapping[name]
	return exists
}

// HandleCompatibilityMode checks if we're running in compatibility mode and handles it
func HandleCompatibilityMode() bool {
	// Check the binary name to see if we're running as a compatibility tool
	progName := os.Args[0]
	baseName := strings.TrimSuffix(filepath.Base(progName), ".exe")

	if IsCompatibilityCommand(baseName) {
		// We're running in compatibility mode
		registry := cli.GlobalRegistry
		if registry != nil {
			if cmdFunc, exists := registry.Get(baseName); exists {
				cmd := cmdFunc()
				cmd.SetArgs(os.Args[1:])
				if err := cmd.Execute(); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				return true
			}
		}
	}

	return false
}
