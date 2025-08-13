// ABOUTME: Implements carton compatibility interface for PVM
// ABOUTME: Maps carton commands to equivalent workspace and module operations

package compat

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// CartonMapper implements carton compatibility
type CartonMapper struct {
	*BaseMapper
}

// NewCartonMapper creates a new carton compatibility mapper
func NewCartonMapper() *CartonMapper {
	return &CartonMapper{
		BaseMapper: NewBaseMapper("carton", "Carton dependency manager compatibility for PVM"),
	}
}

// MapCommand maps carton commands to PVM equivalents
func (m *CartonMapper) MapCommand(args []string) (string, []string, error) {
	if len(args) == 0 {
		// Default to install when no command given
		return "module", []string{"install", "--deps-only"}, nil
	}

	command := args[0]
	restArgs := args[1:]

	switch command {
	case "help", "-h", "--help":
		// Help is handled by the wrapper
		return "", nil, nil

	case "install":
		return m.mapInstall(restArgs)

	case "exec":
		return m.mapExec(restArgs)

	case "list":
		return "module", []string{"list"}, nil

	case "show":
		if len(restArgs) > 0 {
			return "module", append([]string{"info"}, restArgs[0]), nil
		}
		return "", nil, fmt.Errorf("carton show requires a module name")

	case "update":
		return m.mapUpdate(restArgs)

	case "tree":
		return "module", []string{"tree"}, nil

	case "check":
		// Check if dependencies are satisfied
		return "module", []string{"outdated"}, nil

	case "bundle":
		// Bundle dependencies for deployment
		return "module", []string{"bundle"}, nil

	case "fatpack":
		// Create a fatpacked script - not directly supported
		return "", nil, fmt.Errorf("carton fatpack is not supported yet")

	case "version":
		return "version", nil, nil

	default:
		return "", nil, fmt.Errorf("unknown carton command: %s", command)
	}
}

// mapInstall maps carton install to PVM
func (m *CartonMapper) mapInstall(args []string) (string, []string, error) {
	pvmArgs := []string{"install", "--deps-only"}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--deployment":
			pvmArgs = append(pvmArgs, "--production")

		case arg == "--cached":
			pvmArgs = append(pvmArgs, "--cache-only")

		case arg == "--path":
			if i+1 < len(args) {
				pvmArgs = append(pvmArgs, "--install-dir", args[i+1])
				i++
			}

		case strings.HasPrefix(arg, "--path="):
			path := strings.TrimPrefix(arg, "--path=")
			pvmArgs = append(pvmArgs, "--install-dir", path)

		case arg == "--without":
			if i+1 < len(args) {
				// Handle feature exclusion
				pvmArgs = append(pvmArgs, "--without", args[i+1])
				i++
			}

		case !strings.HasPrefix(arg, "-"):
			// Pass through non-flag arguments
			pvmArgs = append(pvmArgs, arg)
		}
	}

	return "module", pvmArgs, nil
}

// mapExec maps carton exec to PVM
func (m *CartonMapper) mapExec(args []string) (string, []string, error) {
	// Remove the -- separator if present
	execArgs := args
	if len(args) > 0 && args[0] == "--" {
		execArgs = args[1:]
	}

	if len(execArgs) == 0 {
		return "", nil, fmt.Errorf("carton exec requires a command")
	}

	// Use pvm run to execute in the workspace context
	return "run", execArgs, nil
}

// mapUpdate maps carton update to PVM
func (m *CartonMapper) mapUpdate(args []string) (string, []string, error) {
	if len(args) > 0 {
		// Update specific modules
		return "module", append([]string{"update"}, args...), nil
	}
	// Update all modules
	return "module", []string{"update"}, nil
}

// GetHelp returns carton-style help text
func (m *CartonMapper) GetHelp() string {
	return `carton - compatibility interface for PVM

Usage:
  carton <command> [options]

Commands:
  install [options]       Install dependencies from cpanfile
  exec -- <command>       Execute command in carton environment
  list                    List installed modules
  show <module>           Show module information
  update [modules...]     Update dependencies
  tree                    Show dependency tree
  check                   Check if dependencies are satisfied
  bundle                  Bundle dependencies
  version                 Show version

Install Options:
  --deployment           Production mode (frozen deps)
  --cached               Use cached modules only
  --path DIR             Installation directory
  --without FEATURES     Exclude features

Examples:
  carton install                 # Install from cpanfile
  carton install --deployment    # Production install
  carton exec -- perl script.pl  # Run with deps
  carton list                    # List modules
  carton update DBI              # Update specific module

This is a PVM compatibility interface for carton.
For full PVM workspace management, use: pvm workspace --help`
}

// NewCartonCommand creates the carton compatibility command
func NewCartonCommand() *cobra.Command {
	mapper := NewCartonMapper()
	return createCompatCommand(mapper)
}
