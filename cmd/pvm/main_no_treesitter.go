//go:build no_treesitter

// ABOUTME: Main entry point for PVM without tree-sitter dependencies (Windows builds)
// ABOUTME: Excludes PSC functionality that requires tree-sitter/CGO

package main

import (
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/pvi"
	"tamarou.com/pvm/internal/pvm"
	"tamarou.com/pvm/internal/pvx"
)

func init() {
	// Register component commands for symlink support (excluding PSC)
	cli.GlobalRegistry.Register(cli.ComponentPVM, pvm.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPVI, pvi.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPVX, pvx.NewCommand)

	// PSC is not available in builds without tree-sitter support
	// This affects Windows builds until purego tree-sitter migration is complete
}

func main() {
	// Detect component and create root command
	component := cli.DetectComponent()
	rootCmd := cli.CreateRootCommand(component)

	// Execute the root command
	cli.Execute(rootCmd)
}
