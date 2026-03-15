// ABOUTME: Standalone entry point for PM (Perl Module manager)
// ABOUTME: Manages CPAN modules for installed Perl versions

package main

import (
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/pm"
)

func init() {
	cli.GlobalRegistry.Register(cli.ComponentPM, pm.NewCommand)
}

func main() {
	component := cli.DetectComponent()
	rootCmd := cli.CreateRootCommand(component)
	cli.Execute(rootCmd)
}
