// ABOUTME: Main entry point for PVM (Perl Version Manager)
// ABOUTME: Handles Perl installation, version switching, and management

package main

import (
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/psc"
	"tamarou.com/pvm/internal/pvi"
	"tamarou.com/pvm/internal/pvm"
	"tamarou.com/pvm/internal/pvx"
)

func init() {
	// Register all component commands for symlink support
	cli.GlobalRegistry.Register(cli.ComponentPVM, pvm.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPSC, psc.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPVI, pvi.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPVX, pvx.NewCommand)
}

func main() {
	// Detect component and create root command
	component := cli.DetectComponent()
	rootCmd := cli.CreateRootCommand(component)

	// Execute the root command
	cli.Execute(rootCmd)
}
