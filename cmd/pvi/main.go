// ABOUTME: Main entry point for PVI (Perl Version Installer)
// ABOUTME: Manages CPAN modules for installed Perl versions

package main

import (
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/pvi"
)

func init() {
	// Register PVI command
	cli.GlobalRegistry.Register(cli.ComponentPVI, pvi.NewCommand)
}

func main() {
	// Detect component and create root command
	component := cli.DetectComponent()
	rootCmd := cli.CreateRootCommand(component)

	// Execute the root command
	cli.Execute(rootCmd)
}
