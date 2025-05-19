// ABOUTME: Main entry point for PSC (Perl Script Compiler)
// ABOUTME: Provides static type checking for Perl code

package main

import (
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/psc"
)

func init() {
	// Register PSC command
	cli.GlobalRegistry.Register(cli.ComponentPSC, psc.NewCommand)
}

func main() {
	// Detect component and create root command
	component := cli.DetectComponent()
	rootCmd := cli.CreateRootCommand(component)

	// Execute the root command
	cli.Execute(rootCmd)
}
