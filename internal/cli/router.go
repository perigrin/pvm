// ABOUTME: Command router for multiple entry points
// ABOUTME: Routes to the appropriate command set based on binary name

package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/version"
)

// Component names
const (
	ComponentPVM = "pvm"
	ComponentPVX = "pvx"
	ComponentPVI = "pvi"
	ComponentPSC = "psc"
)

// Component descriptions
const (
	DescriptionPVM = "Perl Version Manager - Manages Perl installations and versions"
	DescriptionPVX = "Perl Version eXecutor - Executes modules/scripts in isolated environments"
	DescriptionPVI = "Perl Version Installer - Manages CPAN modules for installed Perl versions"
	DescriptionPSC = "Perl Script Compiler - Provides static type checking for Perl code"
)

// DetectComponent detects which component to run based on the binary name
func DetectComponent() string {
	// Get the executable name
	executable := filepath.Base(os.Args[0])
	
	// Remove extension on Windows
	executable = strings.TrimSuffix(executable, filepath.Ext(executable))
	
	// Check for known components
	switch executable {
	case ComponentPVM:
		return ComponentPVM
	case ComponentPVX:
		return ComponentPVX
	case ComponentPVI:
		return ComponentPVI
	case ComponentPSC:
		return ComponentPSC
	default:
		// Default to PVM if not recognized
		return ComponentPVM
	}
}

// GetComponentDescription returns the description for a component
func GetComponentDescription(component string) string {
	switch component {
	case ComponentPVM:
		return DescriptionPVM
	case ComponentPVX:
		return DescriptionPVX
	case ComponentPVI:
		return DescriptionPVI
	case ComponentPSC:
		return DescriptionPSC
	default:
		return "Unknown component"
	}
}

// CreateRootCommand creates the appropriate root command based on the component
func CreateRootCommand(component string) *cobra.Command {
	description := GetComponentDescription(component)
	rootCmd := NewRootCommand(component, description)
	
	// Add component-specific initialization here
	rootCmd.Long = rootCmd.Long + "\n\nVersion: " + version.GetVersion()
	
	// Look up registered commands for this component
	provider, exists := GlobalRegistry.Get(component)
	if exists {
		componentCmd := provider()
		for _, cmd := range componentCmd.Commands() {
			rootCmd.AddCommand(cmd)
		}
	}
	
	return rootCmd
}