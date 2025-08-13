// ABOUTME: Implements cpanm compatibility interface for PVM
// ABOUTME: Maps cpanm commands to equivalent pm module install operations

package compat

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/log"
)

// CpanmMapper implements cpanm compatibility
type CpanmMapper struct {
	*BaseMapper
}

// NewCpanmMapper creates a new cpanm compatibility mapper
func NewCpanmMapper() *CpanmMapper {
	return &CpanmMapper{
		BaseMapper: NewBaseMapper("cpanm", "App::cpanminus compatibility for PVM"),
	}
}

// MapCommand maps cpanm commands to PVM equivalents
func (m *CpanmMapper) MapCommand(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("no arguments provided")
	}

	// Parse cpanm arguments
	var modules []string
	var pvmArgs []string
	showDeps := false
	selfUpgrade := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-h" || arg == "--help":
			// Help is handled by the wrapper
			return "", nil, nil

		case arg == "-v" || arg == "--verbose":
			pvmArgs = append(pvmArgs, "--verbose")

		case arg == "-q" || arg == "--quiet":
			pvmArgs = append(pvmArgs, "--quiet")

		case arg == "-f" || arg == "--force":
			pvmArgs = append(pvmArgs, "--force")

		case arg == "-n" || arg == "--notest":
			pvmArgs = append(pvmArgs, "--no-test")

		case arg == "-S" || arg == "--sudo":
			// PVM handles permissions differently
			log.Warnf("--sudo flag ignored, PVM manages permissions automatically")

		case arg == "-L" || arg == "--local-lib-contained":
			if i+1 < len(args) {
				pvmArgs = append(pvmArgs, "--install-dir", args[i+1])
				i++
			}

		case strings.HasPrefix(arg, "-L="):
			path := strings.TrimPrefix(arg, "-L=")
			pvmArgs = append(pvmArgs, "--install-dir", path)

		case arg == "-l" || arg == "--local-lib":
			if i+1 < len(args) {
				pvmArgs = append(pvmArgs, "--local-lib", args[i+1])
				i++
			}

		case strings.HasPrefix(arg, "--local-lib="):
			path := strings.TrimPrefix(arg, "--local-lib=")
			pvmArgs = append(pvmArgs, "--local-lib", path)

		case arg == "--mirror":
			if i+1 < len(args) {
				pvmArgs = append(pvmArgs, "--mirror", args[i+1])
				i++
			}

		case strings.HasPrefix(arg, "--mirror="):
			mirror := strings.TrimPrefix(arg, "--mirror=")
			pvmArgs = append(pvmArgs, "--mirror", mirror)

		case arg == "--mirror-only":
			pvmArgs = append(pvmArgs, "--mirror-only")

		case arg == "--installdeps":
			pvmArgs = append(pvmArgs, "--deps-only")

		case arg == "--showdeps":
			showDeps = true

		case arg == "--reinstall":
			pvmArgs = append(pvmArgs, "--reinstall")

		case arg == "--self-upgrade":
			selfUpgrade = true

		case arg == "--info":
			// Change command to info instead of install
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				return "module", append([]string{"info"}, args[i+1]), nil
			}

		case arg == "--look":
			// Open module in shell - not directly supported
			return "", nil, fmt.Errorf("--look is not supported yet")

		case arg == "--uninstall" || arg == "-U":
			// Change to remove command
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				return "module", append([]string{"remove"}, args[i+1]), nil
			}

		case arg == "." || arg == "./":
			// Install from current directory - treat as module name
			modules = append(modules, arg)

		case !strings.HasPrefix(arg, "-"):
			// Module name or path
			modules = append(modules, arg)

		default:
			// Unknown flag - pass through with warning
			log.Warnf("Unknown cpanm flag: %s", arg)
			pvmArgs = append(pvmArgs, arg)
		}
	}

	// Handle special cases
	if selfUpgrade {
		return "self-update", nil, nil
	}

	if showDeps {
		if len(modules) > 0 {
			return "module", append([]string{"info", "--dependencies"}, modules[0]), nil
		}
		return "", nil, fmt.Errorf("--showdeps requires a module name")
	}

	// Default to install (PM command)
	finalArgs := make([]string, 0, len(pvmArgs)+len(modules))
	finalArgs = append(finalArgs, pvmArgs...)
	finalArgs = append(finalArgs, modules...)

	return "install", finalArgs, nil
}

// GetHelp returns cpanm-style help text
func (m *CpanmMapper) GetHelp() string {
	return `cpanm - compatibility interface for PVM

Usage:
  cpanm [options] Module [...]
  cpanm [options] URL
  cpanm [options] .

Options:
  -v, --verbose              Turn on verbose output
  -q, --quiet                Turn off all output
  -f, --force                Force install
  -n, --notest               Skip tests
  -S, --sudo                 Use sudo (ignored by PVM)
  -l, --local-lib PATH       Install to local::lib path
  -L, --local-lib-contained PATH  Install to self-contained directory
  --mirror URL               Use mirror URL
  --mirror-only              Use only configured mirrors
  --installdeps              Install dependencies only
  --showdeps                 Show dependencies
  --reinstall                Force reinstall
  --self-upgrade             Upgrade PVM
  --info                     Show module info
  --uninstall, -U            Uninstall module
  -h, --help                 Show this help

Examples:
  cpanm Moose                    # Install Moose
  cpanm -l ~/perl5 DBI           # Install to local::lib
  cpanm --installdeps .          # Install deps from current dir
  cpanm --showdeps Catalyst      # Show dependencies

This is a PVM compatibility interface for cpanm.
For full PVM module management, use: pvm module --help`
}

// NewCpanmCommand creates the cpanm compatibility command
func NewCpanmCommand() *cobra.Command {
	mapper := NewCpanmMapper()
	return createCompatCommand(mapper)
}
