// ABOUTME: Implements perlbrew compatibility interface for PVM
// ABOUTME: Maps perlbrew commands to equivalent version management operations

package compat

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/log"
)

// PerlbrewMapper implements perlbrew compatibility
type PerlbrewMapper struct {
	*BaseMapper
}

// NewPerlbrewMapper creates a new perlbrew compatibility mapper
func NewPerlbrewMapper() *PerlbrewMapper {
	return &PerlbrewMapper{
		BaseMapper: NewBaseMapper("perlbrew", "Perlbrew version manager compatibility for PVM"),
	}
}

// MapCommand maps perlbrew commands to PVM equivalents
func (m *PerlbrewMapper) MapCommand(args []string) (string, []string, error) {
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
		// perlbrew init - setup environment
		return "init", nil, nil

	case "install":
		return m.mapInstall(restArgs)

	case "list":
		return "list", nil, nil

	case "available", "available-perls":
		return "available", nil, nil

	case "use":
		return m.mapUse(restArgs)

	case "switch":
		return m.mapSwitch(restArgs)

	case "off":
		// Switch to system perl
		return "use", []string{"system"}, nil

	case "current":
		return "current", nil, nil

	case "exec":
		return m.mapExec(restArgs)

	case "uninstall":
		return m.mapUninstall(restArgs)

	case "self-upgrade":
		return "self-update", nil, nil

	case "info":
		if len(restArgs) > 0 {
			version := parseVersionString(restArgs[0])
			return "info", []string{version}, nil
		}
		return "info", nil, nil

	case "lib":
		return m.mapLib(restArgs)

	case "clone":
		return m.mapClone(restArgs)

	case "clean":
		// Clean up installation files
		return "clean", nil, nil

	case "version":
		return "version", nil, nil

	case "env":
		// Show environment info
		return "env", nil, nil

	case "alias":
		return m.mapAlias(restArgs)

	default:
		return "", nil, fmt.Errorf("unknown perlbrew command: %s", command)
	}
}

// mapInstall maps perlbrew install to PVM install
func (m *PerlbrewMapper) mapInstall(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("perlbrew install requires a perl version")
	}

	var pvmArgs []string
	var version string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--notest" || arg == "-n":
			pvmArgs = append(pvmArgs, "--no-test")

		case arg == "--force" || arg == "-f":
			pvmArgs = append(pvmArgs, "--force")

		case arg == "--verbose" || arg == "-v":
			pvmArgs = append(pvmArgs, "--verbose")

		case arg == "--quiet" || arg == "-q":
			pvmArgs = append(pvmArgs, "--quiet")

		case arg == "--as":
			if i+1 < len(args) {
				pvmArgs = append(pvmArgs, "--alias", args[i+1])
				i++
			}

		case arg == "--both":
			// Install both threaded and non-threaded
			pvmArgs = append(pvmArgs, "--threaded", "--non-threaded")

		case arg == "--thread":
			pvmArgs = append(pvmArgs, "--threaded")

		case arg == "--multi":
			pvmArgs = append(pvmArgs, "--multi")

		case arg == "--64int":
			pvmArgs = append(pvmArgs, "--64int")

		case arg == "--64all":
			pvmArgs = append(pvmArgs, "--64all")

		case arg == "--ld":
			pvmArgs = append(pvmArgs, "--longdouble")

		case strings.HasPrefix(arg, "-D"):
			// Configure options
			pvmArgs = append(pvmArgs, arg)

		case strings.HasPrefix(arg, "-A"):
			// Configure options
			pvmArgs = append(pvmArgs, arg)

		case !strings.HasPrefix(arg, "-"):
			// Version specification
			version = parseVersionString(arg)

		default:
			log.Warnf("Unknown perlbrew install flag: %s", arg)
		}
	}

	if version == "" {
		return "", nil, fmt.Errorf("no perl version specified")
	}

	finalArgs := append([]string{version}, pvmArgs...)
	return "install", finalArgs, nil
}

// mapUse maps perlbrew use to PVM shell
func (m *PerlbrewMapper) mapUse(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("perlbrew use requires a perl version")
	}

	version := parseVersionString(args[0])
	return "shell", []string{version}, nil
}

// mapSwitch maps perlbrew switch to PVM use
func (m *PerlbrewMapper) mapSwitch(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("perlbrew switch requires a perl version")
	}

	version := parseVersionString(args[0])
	return "use", []string{version}, nil
}

// mapExec maps perlbrew exec to PVM exec
func (m *PerlbrewMapper) mapExec(args []string) (string, []string, error) {
	// Remove the -- separator if present
	execArgs := args
	if len(args) > 0 && args[0] == "--" {
		execArgs = args[1:]
	}

	if len(execArgs) == 0 {
		return "", nil, fmt.Errorf("perlbrew exec requires a command")
	}

	return "exec", execArgs, nil
}

// mapUninstall maps perlbrew uninstall to PVM uninstall
func (m *PerlbrewMapper) mapUninstall(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("perlbrew uninstall requires a perl version")
	}

	version := parseVersionString(args[0])
	return "uninstall", []string{version}, nil
}

// mapLib maps perlbrew lib commands
func (m *PerlbrewMapper) mapLib(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "lib", []string{"list"}, nil
	}

	subcommand := args[0]
	restArgs := args[1:]

	switch subcommand {
	case "create":
		if len(restArgs) > 0 {
			return "lib", append([]string{"create"}, restArgs...), nil
		}
		return "", nil, fmt.Errorf("perlbrew lib create requires a library name")

	case "delete":
		if len(restArgs) > 0 {
			return "lib", append([]string{"delete"}, restArgs...), nil
		}
		return "", nil, fmt.Errorf("perlbrew lib delete requires a library name")

	case "list":
		return "lib", []string{"list"}, nil

	default:
		return "", nil, fmt.Errorf("unknown perlbrew lib command: %s", subcommand)
	}
}

// mapClone maps perlbrew clone
func (m *PerlbrewMapper) mapClone(args []string) (string, []string, error) {
	if len(args) < 2 {
		return "", nil, fmt.Errorf("perlbrew clone requires source and target versions")
	}

	source := parseVersionString(args[0])
	target := parseVersionString(args[1])

	return "clone", []string{source, target}, nil
}

// mapAlias maps perlbrew alias commands
func (m *PerlbrewMapper) mapAlias(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "alias", []string{"list"}, nil
	}

	subcommand := args[0]
	restArgs := args[1:]

	switch subcommand {
	case "create":
		if len(restArgs) >= 2 {
			version := parseVersionString(restArgs[0])
			alias := restArgs[1]
			return "alias", []string{"create", version, alias}, nil
		}
		return "", nil, fmt.Errorf("perlbrew alias create requires version and alias name")

	case "delete":
		if len(restArgs) > 0 {
			return "alias", append([]string{"delete"}, restArgs...), nil
		}
		return "", nil, fmt.Errorf("perlbrew alias delete requires an alias name")

	case "list":
		return "alias", []string{"list"}, nil

	default:
		return "", nil, fmt.Errorf("unknown perlbrew alias command: %s", subcommand)
	}
}

// GetHelp returns perlbrew-style help text
func (m *PerlbrewMapper) GetHelp() string {
	return `perlbrew - compatibility interface for PVM

Usage:
  perlbrew <command> [options]

Commands:
  init                     Initialize perlbrew environment
  install <perl>           Install a perl version
  list                     List installed perls
  available                List available perls
  use <perl>               Use perl for this shell
  switch <perl>            Switch to perl permanently
  off                      Switch to system perl
  current                  Show current perl
  exec [-- command]        Execute command with perl
  uninstall <perl>         Uninstall perl version
  info [perl]              Show perl information
  version                  Show perlbrew version
  self-upgrade            Upgrade perlbrew

Library Management:
  lib                     List libraries
  lib create <name>       Create a library
  lib delete <name>       Delete a library

Advanced:
  clone <from> <to>       Clone perl installation
  alias create <perl> <alias>  Create alias
  alias delete <alias>    Delete alias
  clean                   Clean up installation files

Install Options:
  -n, --notest           Skip tests
  -f, --force            Force installation
  -v, --verbose          Verbose output
  -q, --quiet            Quiet mode
  --as <name>            Install as alias
  --thread               Build with threading
  --multi                Build with multiplicity
  --64int                Build with 64-bit integers

Examples:
  perlbrew install perl-5.38.0    # Install Perl 5.38.0
  perlbrew switch perl-5.38.0     # Switch to 5.38.0
  perlbrew exec -- prove t/       # Run tests

This is a PVM compatibility interface for perlbrew.
For full PVM version management, use: pvm --help`
}

// NewPerlbrewCommand creates the perlbrew compatibility command
func NewPerlbrewCommand() *cobra.Command {
	mapper := NewPerlbrewMapper()
	return createCompatCommand(mapper)
}
