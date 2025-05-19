// ABOUTME: Main entry point for PVX (Perl Version eXecutor)
// ABOUTME: Executes Perl code in isolated environments

package main

import (
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/pvx"
)

func init() {
	// Register PVX command
	cli.GlobalRegistry.Register(cli.ComponentPVX, pvx.NewCommand)
}

func main() {
	// Detect component and create root command
	component := cli.DetectComponent()
	rootCmd := cli.CreateRootCommand(component)

	// Execute the root command
	cli.Execute(rootCmd)
}