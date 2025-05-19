// ABOUTME: Main entry point for PVM (Perl Version Manager)
// ABOUTME: Handles Perl installation, version switching, and management

package main

import (
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/pvm"
)

func init() {
	// Register PVM command
	cli.GlobalRegistry.Register(cli.ComponentPVM, pvm.NewCommand)
}

func main() {
	// Detect component and create root command
	component := cli.DetectComponent()
	rootCmd := cli.CreateRootCommand(component)

	// Execute the root command
	cli.Execute(rootCmd)
}
