// ABOUTME: PVM version-related commands
// ABOUTME: Provides commands for working with Perl versions

package pvm

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/perl"
)

// newPerlVersionUtilCommand creates a command for version utilities
func newPerlVersionUtilCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Perl version utilities",
		Long:  "Parse, validate, and compare Perl versions",
	}

	// Add subcommands
	cmd.AddCommand(
		newVersionParseCommand(),
		newVersionCompareCommand(),
		newVersionSatisfiesCommand(),
		newVersionAliasCommand(),
	)

	return cmd
}

// newVersionParseCommand creates a command to parse a version string
func newVersionParseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "parse [version]",
		Short: "Parse a version string",
		Long:  "Parse a version string into its components",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			version, err := perl.ParseVersion(args[0])
			if err != nil {
				return err
			}

			ui.KeyValue(map[string]string{
				"Version":     version.String(),
				"Major":       fmt.Sprintf("%d", version.Major),
				"Minor":       fmt.Sprintf("%d", version.Minor),
				"Patch":       fmt.Sprintf("%d", version.Patch),
				"Development": fmt.Sprintf("%t", version.Dev),
				"Stable":      fmt.Sprintf("%t", version.IsStable()),
			})

			return nil
		},
	}
}

// newVersionCompareCommand creates a command to compare two versions
func newVersionCompareCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "compare [version1] [version2]",
		Short: "Compare two versions",
		Long:  "Compare two versions and determine their relationship",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			v1, err := perl.ParseVersion(args[0])
			if err != nil {
				return errors.NewVersionError("001",
					fmt.Sprintf("Invalid first version: %s", args[0]), err)
			}

			v2, err := perl.ParseVersion(args[1])
			if err != nil {
				return errors.NewVersionError("002",
					fmt.Sprintf("Invalid second version: %s", args[1]), err)
			}

			result := v1.Compare(v2)
			var relationship string
			switch {
			case result < 0:
				relationship = "less than"
			case result > 0:
				relationship = "greater than"
			default:
				relationship = "equal to"
			}

			ui.Info("%s is %s %s", v1.String(), relationship, v2.String())
			return nil
		},
	}
}

// newVersionSatisfiesCommand creates a command to check if a version satisfies constraints
func newVersionSatisfiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "satisfies [version] [constraints]",
		Short: "Check if a version satisfies constraints",
		Long:  "Check if a version meets the specified version constraints",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			version, err := perl.ParseVersion(args[0])
			if err != nil {
				return errors.NewVersionError("003",
					fmt.Sprintf("Invalid version: %s", args[0]), err)
			}

			constraints, err := perl.ParseConstraintSet(args[1])
			if err != nil {
				return errors.NewVersionError("004",
					fmt.Sprintf("Invalid constraint set: %s", args[1]), err)
			}

			satisfies := constraints.Satisfies(version)
			if satisfies {
				ui.Success("%s satisfies %s", version.String(), constraints.String())
			} else {
				ui.Error("%s does NOT satisfy %s", version.String(), constraints.String())
			}

			return nil
		},
	}
}

// newVersionAliasCommand creates a command to resolve version aliases
func newVersionAliasCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "alias [alias]",
		Short: "Resolve a version alias",
		Long:  "Resolve a version alias to an actual version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			// Load configuration to get aliases
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Get version aliases
			versionAliases := map[string]string{}
			if cfg.PVM != nil && cfg.PVM.VersionAliases != nil {
				versionAliases = cfg.PVM.VersionAliases
			}

			// Ensure alias starts with @
			alias := args[0]
			if !strings.HasPrefix(alias, "@") {
				alias = "@" + alias
			}

			version, err := perl.ResolveVersionAlias(alias, versionAliases)
			if err != nil {
				return err
			}

			ui.KeyValue(map[string]string{
				"Alias":   alias,
				"Version": version,
			})
			return nil
		},
	}
}
