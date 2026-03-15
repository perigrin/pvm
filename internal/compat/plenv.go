// ABOUTME: Implements plenv compatibility interface for PVM
// ABOUTME: Maps plenv commands to equivalent version management operations

package compat

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/log"
)

// PlenvMapper implements plenv compatibility
type PlenvMapper struct {
	*BaseMapper
}

// NewPlenvMapper creates a new plenv compatibility mapper
func NewPlenvMapper() *PlenvMapper {
	return &PlenvMapper{
		BaseMapper: NewBaseMapper("plenv", "plenv version manager compatibility for PVM"),
	}
}

// MapCommand maps plenv commands to PVM equivalents
func (m *PlenvMapper) MapCommand(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("no command provided")
	}

	command := args[0]
	restArgs := args[1:]

	switch command {
	case "help", "-h", "--help":
		// Help is handled by the wrapper
		return "", nil, nil

	case "init":
		// plenv init - setup environment
		return "init", nil, nil

	case "install":
		return m.mapInstall(restArgs)

	case "uninstall":
		return m.mapUninstall(restArgs)

	case "versions":
		return "list", nil, nil

	case "version":
		return "current", nil, nil

	case "global":
		return m.mapGlobal(restArgs)

	case "local":
		return m.mapLocal(restArgs)

	case "shell":
		return m.mapShell(restArgs)

	case "rehash":
		// Rehash shims - PVM handles this automatically
		log.Infof("plenv rehash not needed, PVM handles shims automatically")
		return "", nil, nil

	case "which":
		return m.mapWhich(restArgs)

	case "whence":
		return m.mapWhence(restArgs)

	case "exec":
		return m.mapExec(restArgs)

	case "shims":
		// List shims
		return "shims", nil, nil

	case "prefix":
		return m.mapPrefix(restArgs)

	case "root":
		// Show plenv root directory
		return "root", nil, nil

	case "version-name":
		return "current", []string{"--name-only"}, nil

	case "version-origin":
		return "current", []string{"--origin"}, nil

	case "commands":
		// List available commands
		return "help", nil, nil

	default:
		return "", nil, fmt.Errorf("unknown plenv command: %s", command)
	}
}

// mapInstall maps plenv install to PVM install
func (m *PlenvMapper) mapInstall(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("plenv install requires arguments")
	}

	// Check for special flags
	if args[0] == "--list" || args[0] == "-l" {
		return "available", nil, nil
	}

	if args[0] == "--skip-existing" || args[0] == "-s" {
		if len(args) < 2 {
			return "", nil, fmt.Errorf("plenv install --skip-existing requires a version")
		}
		version := parseVersionString(args[1])
		return "install", []string{version, "--skip-existing"}, nil
	}

	var pvmArgs []string
	var version string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--verbose" || arg == "-v":
			pvmArgs = append(pvmArgs, "--verbose")

		case arg == "--keep" || arg == "-k":
			pvmArgs = append(pvmArgs, "--keep-build")

		case arg == "--patch":
			pvmArgs = append(pvmArgs, "--patch")

		case arg == "--debug":
			pvmArgs = append(pvmArgs, "--debug")

		case !strings.HasPrefix(arg, "-"):
			// Version specification
			version = parseVersionString(arg)

		default:
			log.Warnf("Unknown plenv install flag: %s", arg)
		}
	}

	if version == "" {
		return "", nil, fmt.Errorf("no perl version specified")
	}

	finalArgs := append([]string{version}, pvmArgs...)
	return "install", finalArgs, nil
}

// mapUninstall maps plenv uninstall to PVM uninstall
func (m *PlenvMapper) mapUninstall(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("plenv uninstall requires a perl version")
	}

	version := parseVersionString(args[0])
	var pvmArgs []string

	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--force" || arg == "-f":
			pvmArgs = append(pvmArgs, "--force")
		default:
			pvmArgs = append(pvmArgs, arg)
		}
	}

	finalArgs := append([]string{version}, pvmArgs...)
	return "uninstall", finalArgs, nil
}

// mapGlobal maps plenv global to PVM use --global
func (m *PlenvMapper) mapGlobal(args []string) (string, []string, error) {
	if len(args) == 0 {
		// Show current global version
		return "current", []string{"--global"}, nil
	}

	version := parseVersionString(args[0])
	return "use", []string{version, "--global"}, nil
}

// mapLocal maps plenv local to PVM use --local
func (m *PlenvMapper) mapLocal(args []string) (string, []string, error) {
	if len(args) == 0 {
		// Show current local version
		return "current", []string{"--local"}, nil
	}

	if args[0] == "--unset" {
		return "use", []string{"--unset-local"}, nil
	}

	version := parseVersionString(args[0])
	return "use", []string{version, "--local"}, nil
}

// mapShell maps plenv shell to PVM shell
func (m *PlenvMapper) mapShell(args []string) (string, []string, error) {
	if len(args) == 0 {
		// Show current shell version
		return "current", []string{"--shell"}, nil
	}

	if args[0] == "--unset" {
		return "shell", []string{"--unset"}, nil
	}

	version := parseVersionString(args[0])
	return "shell", []string{version}, nil
}

// mapWhich maps plenv which to PVM which
func (m *PlenvMapper) mapWhich(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("plenv which requires a command name")
	}

	return "which", args, nil
}

// mapWhence maps plenv whence to PVM whence
func (m *PlenvMapper) mapWhence(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("plenv whence requires a command name")
	}

	// whence shows which versions provide a command
	return "whence", args, nil
}

// mapExec maps plenv exec to PVM exec
func (m *PlenvMapper) mapExec(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("plenv exec requires a command")
	}

	return "exec", args, nil
}

// mapPrefix maps plenv prefix to PVM prefix
func (m *PlenvMapper) mapPrefix(args []string) (string, []string, error) {
	if len(args) == 0 {
		// Show current version prefix
		return "prefix", nil, nil
	}

	version := parseVersionString(args[0])
	return "prefix", []string{version}, nil
}

// GetHelp returns plenv-style help text
func (m *PlenvMapper) GetHelp() string {
	return `plenv - compatibility interface for PVM

Usage:
  plenv <command> [<args>]

Commands:
  init                    Initialize plenv environment
  install <version>       Install a perl version
  install --list          List available versions
  uninstall <version>     Uninstall a perl version
  versions                List installed versions
  version                 Show current version
  global [<version>]      Set or show global version
  local [<version>]       Set or show local version
  shell [<version>]       Set or show shell version
  rehash                  Refresh shims (automatic in PVM)
  which <command>         Display path to executable
  whence <command>        List versions providing command
  exec <command> [args]   Execute command with plenv
  shims                   List shims
  prefix [<version>]      Show installation prefix
  root                    Show plenv root directory
  version-name            Show version name only
  version-origin          Show how version was selected
  commands                List available commands

Install Options:
  -l, --list             List available versions
  -s, --skip-existing    Skip if version already exists
  -k, --keep             Keep build directory
  -v, --verbose          Verbose output
  --patch                Apply patches
  --debug                Debug build

Local/Global Options:
  --unset                Unset local/shell version

Examples:
  plenv install 5.38.0         # Install Perl 5.38.0
  plenv global 5.38.0          # Set global version
  plenv local 5.36.0           # Set local version
  plenv exec cpan Module::Name # Execute with current version

This is a PVM compatibility interface for plenv.
For full PVM version management, use: pvm --help`
}

// NewPlenvCommand creates the plenv compatibility command
func NewPlenvCommand() *cobra.Command {
	mapper := NewPlenvMapper()
	return createCompatCommand(mapper)
}
