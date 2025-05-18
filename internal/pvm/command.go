// ABOUTME: PVM-specific commands and functionality
// ABOUTME: Implements commands for Perl version management

package pvm

import (
	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/perl"
)

// NewCommand creates a new PVM command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pvm",
		Short: "Perl Version Manager",
		Long:  "Manages Perl installations and versions",
	}

	// Add PVM-specific commands
	cmd.AddCommand(
		newInstallCommand(),
		newUseCommand(),
		newGlobalCommand(),
		newLocalCommand(),
		newVersionsCommand(),
		newAvailableCommand(),
		newExecCommand(),
		newUninstallCommand(),
		newImportCommand(),
		newRehashCommand(),
		newSymlinksCommand(),
		newConfigCommand(),
		newPerlCommand(),
		newVersionUtilCommand(),
	)

	return cmd
}

// Placeholder commands, to be implemented later

func newInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install [version]",
		Short: "Install a Perl version",
		Long:  "Download and install a specific version of Perl",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Install command not yet implemented")
		},
	}
}

func newUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use [version]",
		Short: "Use a specific version in the current shell",
		Long:  "Temporarily use a specific Perl version in the current shell session",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Use command not yet implemented")
		},
	}
}

func newGlobalCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "global [version]",
		Short: "Set the global Perl version",
		Long:  "Set the default Perl version for all shells",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Global command not yet implemented")
		},
	}
}

func newLocalCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "local [version]",
		Short: "Set the local version for a directory",
		Long:  "Set the Perl version for the current directory and subdirectories",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Local command not yet implemented")
		},
	}
}

func newVersionsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "versions",
		Short: "List installed versions",
		Long:  "List all installed Perl versions",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Versions command not yet implemented")
		},
	}
}

func newAvailableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "available",
		Short: "List available Perl versions",
		Long:  "List all Perl versions available for installation",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Available command not yet implemented")
		},
	}
}

func newExecCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "exec [version] [command]",
		Short: "Execute a command with a specific version",
		Long:  "Execute a command using a specific Perl version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Exec command not yet implemented")
		},
	}
}

func newUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall [version]",
		Short: "Remove a Perl version",
		Long:  "Uninstall a specific version of Perl",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Uninstall command not yet implemented")
		},
	}
}

func newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-from [tool]",
		Short: "Import Perl installations from other tools",
		Long:  "Import Perl installations from other version managers (plenv, perlbrew)",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "plenv",
			Short: "Import from plenv",
			Long:  "Import Perl installations from plenv",
			RunE: func(cmd *cobra.Command, args []string) error {
				return importFromLegacyTool(cmd, perl.Plenv)
			},
		},
		&cobra.Command{
			Use:   "perlbrew",
			Short: "Import from perlbrew",
			Long:  "Import Perl installations from perlbrew",
			RunE: func(cmd *cobra.Command, args []string) error {
				return importFromLegacyTool(cmd, perl.Perlbrew)
			},
		},
	)

	return cmd
}

// importFromLegacyTool implements the logic for importing from a legacy tool
func importFromLegacyTool(cmd *cobra.Command, tool perl.LegacyToolType) error {
	cmd.Printf("Detecting %s installations...\n", tool)

	installations, err := perl.ImportFromLegacyTool(tool)
	if err != nil {
		return err
	}

	cmd.Printf("Found %d %s installation(s):\n", len(installations), tool)
	for i, inst := range installations {
		defaultMark := ""
		if inst.IsDefault {
			defaultMark = " (default)"
		}
		cmd.Printf("%d. %s%s at %s\n", i+1, inst.Version, defaultMark, inst.Path)
	}

	// In a real implementation, here we would:
	// 1. Ask the user which installations to import
	// 2. Create symlinks or copy the installations
	// 3. Register them in PVM's internal database
	// But for now, we'll just report what was found

	cmd.Println("\nNote: This command currently only detects installations.")
	cmd.Println("Actual import functionality will be implemented in a future version.")

	// If it's perlbrew, also show aliases
	if tool == perl.Perlbrew {
		aliases, err := perl.GetPerlbrewAliases()
		if err == nil && len(aliases) > 0 {
			cmd.Println("\nPerlbrew aliases detected:")
			for alias, target := range aliases {
				cmd.Printf("  %s -> %s\n", alias, target)
			}
		}
	}

	return nil
}

func newRehashCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rehash",
		Short: "Rebuild shim executables",
		Long:  "Rebuild shim executables for all installed Perl versions",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Rehash command not yet implemented")
		},
	}
}
