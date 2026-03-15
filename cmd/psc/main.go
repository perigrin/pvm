// ABOUTME: Standalone entry point for PSC (Perl Structural Checker)
// ABOUTME: Provides Perl parsing, analysis, and LSP server as a standalone binary

package main

import (
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/psc"
)

func init() {
	cli.GlobalRegistry.Register(cli.ComponentPSC, psc.NewCommand)
}

func main() {
	component := cli.DetectComponent()
	rootCmd := cli.CreateRootCommand(component)
	cli.Execute(rootCmd)
}
