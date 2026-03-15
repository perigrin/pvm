// ABOUTME: Main entry point for PVM (Perl Version Manager)
// ABOUTME: Handles Perl installation, version switching, and management

package main

import (
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/compat"
	"tamarou.com/pvm/internal/pm"
	"tamarou.com/pvm/internal/pvm"
	"tamarou.com/pvm/internal/pvx"
	"tamarou.com/pvm/internal/templates"
)

func init() {
	pvm.GlobalTemplatesFS = templates.FS
	cli.GlobalRegistry.Register(cli.ComponentPVM, pvm.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPVX, pvx.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPM, pm.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentCpanm, compat.NewCpanmCommand)
	cli.GlobalRegistry.Register(cli.ComponentCarton, compat.NewCartonCommand)
	cli.GlobalRegistry.Register(cli.ComponentPerlbrew, compat.NewPerlbrewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPlenv, compat.NewPlenvCommand)
}

func main() {
	component := cli.DetectComponent()
	rootCmd := cli.CreateRootCommand(component)
	cli.Execute(rootCmd)
}
