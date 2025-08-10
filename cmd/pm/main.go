// ABOUTME: Main entry point for PM (Perl Module)
// ABOUTME: Manages CPAN modules for installed Perl versions

package main

import (
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/pm"
)

func init() {
	// Register PM command
	cli.GlobalRegistry.Register(cli.ComponentPM, pm.NewCommand)
}

func main() {
	// Detect component and create root command
	component := cli.DetectComponent()
	rootCmd := cli.CreateRootCommand(component)

	// Execute the root command
	cli.Execute(rootCmd)
}
