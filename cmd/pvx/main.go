// ABOUTME: Standalone entry point for PVX (Perl Version eXecutor)
// ABOUTME: Runs Perl scripts in isolated environments

package main

import (
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/pvx"
)

func init() {
	cli.GlobalRegistry.Register(cli.ComponentPVX, pvx.NewCommand)
}

func main() {
	component := cli.DetectComponent()
	rootCmd := cli.CreateRootCommand(component)
	cli.Execute(rootCmd)
}
