// ABOUTME: Command router for multiple entry points
// ABOUTME: Routes to the appropriate command set based on binary name

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/version"
)

// Component names
const (
	ComponentPVM      = "pvm"
	ComponentPVX      = "pvx"
	ComponentPM       = "pm"
	ComponentPSC      = "psc"
	ComponentCpanm    = "cpanm"
	ComponentCarton   = "carton"
	ComponentPerlbrew = "perlbrew"
	ComponentPlenv    = "plenv"
)

// Component descriptions
const (
	DescriptionPVM      = "Perl Version Manager - Manages Perl installations and versions"
	DescriptionPVX      = "Perl Version eXecutor - Executes modules/scripts in isolated environments"
	DescriptionPM       = "Perl Module - Manages CPAN modules for installed Perl versions"
	DescriptionPSC      = "Perl Script Compiler - Provides static type checking for Perl code"
	DescriptionCpanm    = "cpanm compatibility interface - Install CPAN modules using cpanm syntax"
	DescriptionCarton   = "carton compatibility interface - Manage dependencies using carton syntax"
	DescriptionPerlbrew = "perlbrew compatibility interface - Manage Perl versions using perlbrew syntax"
	DescriptionPlenv    = "plenv compatibility interface - Manage Perl versions using plenv syntax"
)

// DetectComponent detects which component to run based on the binary name
func DetectComponent() string {
	// Use the enhanced detection logic
	info := DetectInvocation()

	// If verbose flag is set, we'll print debug info
	// This has to be checked manually since flags aren't parsed yet
	for _, arg := range os.Args {
		if arg == "--debug" || arg == "-d" {
			PrintDebugInfo()
			break
		}
	}

	return info.Component
}

// GetComponentDescription returns the description for a component
func GetComponentDescription(component string) string {
	switch component {
	case ComponentPVM:
		return DescriptionPVM
	case ComponentPVX:
		return DescriptionPVX
	case ComponentPM:
		return DescriptionPM
	case ComponentPSC:
		return DescriptionPSC
	case ComponentCpanm:
		return DescriptionCpanm
	case ComponentCarton:
		return DescriptionCarton
	case ComponentPerlbrew:
		return DescriptionPerlbrew
	case ComponentPlenv:
		return DescriptionPlenv
	default:
		return "Unknown component"
	}
}

// CreateRootCommand creates the appropriate root command based on the component
func CreateRootCommand(component string) *cobra.Command {
	description := GetComponentDescription(component)

	// Special handling for PVX - it should be the root command, not a subcommand
	if component == ComponentPVX {
		provider, exists := GlobalRegistry.Get(component)
		if exists {
			// Return the PVX command directly as the root command
			pvxCmd := provider()
			// Add version information to the description
			pvxCmd.Long = pvxCmd.Long + "\n\nVersion: " + version.GetVersion()
			return pvxCmd
		}
	}

	// Special handling for compatibility components - they should be the root command
	if component == ComponentCpanm || component == ComponentCarton ||
		component == ComponentPerlbrew || component == ComponentPlenv {
		provider, exists := GlobalRegistry.Get(component)
		if exists {
			// Return the compatibility command directly as the root command
			compatCmd := provider()
			// Add version information to the description
			compatCmd.Long = compatCmd.Long + "\n\nVersion: " + version.GetVersion()
			return compatCmd
		}
	}

	rootCmd := NewRootCommand(component, description)

	// Add component-specific initialization here
	rootCmd.Long = rootCmd.Long + "\n\nVersion: " + version.GetVersion()

	// Special handling for PVM component to support standard version flags
	if component == ComponentPVM {
		// Add --version flag (without short form to avoid conflict with -v verbose)
		var showVersion bool
		rootCmd.Flags().BoolVar(&showVersion, "version", false, "Show PVM version")

		// Override the pre-run to handle version flag
		origPreRun := rootCmd.PreRun
		rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
			if showVersion {
				fmt.Println(version.GetVersion())
				os.Exit(0)
			}
			// Handle -v as version for root command only (no subcommands)
			if len(args) == 0 && Verbose {
				fmt.Println(version.GetVersion())
				os.Exit(0)
			}
			if origPreRun != nil {
				origPreRun(cmd, args)
			}
		}
	}

	// Look up registered commands for this component
	provider, exists := GlobalRegistry.Get(component)
	if exists {
		componentCmd := provider()
		for _, cmd := range componentCmd.Commands() {
			rootCmd.AddCommand(cmd)
		}
	}

	// Disable Cobra's automatic help flag and command to use our own
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	// Add our own help flag handling
	var showHelp bool
	rootCmd.Flags().BoolVarP(&showHelp, "help", "h", false, "Help for "+component)

	// Override pre-run to handle our help flag
	origPreRun := rootCmd.PreRun
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if showHelp {
			ShowHybridHelp(cmd, args)
			os.Exit(0)
		}
		if origPreRun != nil {
			origPreRun(cmd, args)
		}
	}

	// Override the help function to use our hybrid help system for consistent output
	// This ensures both 'pvm help' and 'pvm -h' produce the same hybrid output
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ShowHybridHelp(cmd, args)
	})

	return rootCmd
}
