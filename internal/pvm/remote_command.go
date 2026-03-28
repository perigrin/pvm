// ABOUTME: PVM remote subcommand implementation (add, list, remove)
// ABOUTME: Manages fork remote sources stored in the user's PVM configuration

package pvm

import (
	"fmt"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/perl"
)

// newRemoteCommand creates the parent 'pvm remote' command
func newRemoteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remote",
		Short: "Manage fork remote sources",
		Long:  "Add, list, and remove custom Perl fork remote sources",
	}

	cmd.AddCommand(
		newRemoteAddCommand(),
		newRemoteListCommand(),
		newRemoteRemoveCommand(),
	)

	return cmd
}

// newRemoteAddCommand creates 'pvm remote add <name> <url>'
func newRemoteAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add a fork remote source",
		Long: `Add a named fork remote source to the user configuration.

The name must match [a-z0-9][a-z0-9-]* and cannot be 'origin' (which is reserved
for the official Perl source). The URL is not validated at add time.

Examples:
  pvm remote add mycompany https://github.com/mycompany/perl.git
  pvm remote add staging   https://git.internal/perl-fork.git`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			url := args[1]

			ui := cli.GetUI(cmd)

			// Guard the reserved name early so the error message matches spec.
			if name == "origin" {
				return fmt.Errorf("cannot add 'origin' — it is reserved")
			}

			// Load user-level config (via effective config — we modify and resave).
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			if cfg.PVM == nil {
				cfg.PVM = &config.PVMConfig{}
			}

			// Build a RemoteList from the existing config entries so we can use
			// the CRUD logic that already lives in perl.RemoteList.
			rl := perl.NewRemoteList()
			for _, rc := range cfg.PVM.Remotes {
				// Populate without validation — these are already-stored entries.
				existing := perl.Remote{Name: rc.Name, URL: rc.URL, Type: rc.Type}
				_ = rl.Add(existing)
			}

			// Add the new remote — this performs validation and duplicate checks.
			newRemote := perl.Remote{Name: name, URL: url}
			if err := rl.Add(newRemote); err != nil {
				return err
			}

			// Sync back to config.
			all := rl.All()
			cfg.PVM.Remotes = make([]config.PVMRemoteConfig, len(all))
			for i, r := range all {
				cfg.PVM.Remotes[i] = config.PVMRemoteConfig{
					Name: r.Name,
					URL:  r.URL,
					Type: r.Type,
				}
			}

			if err := config.SaveUserConfig(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			ui.Success("Added remote '%s' -> %s", name, url)
			return nil
		},
	}
}

// newRemoteListCommand creates 'pvm remote list'
func newRemoteListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured fork remote sources",
		Long:  "Display all configured fork remote sources. 'origin' (the official Perl source) is always present.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// origin is always shown first.
			ui.Printf("origin\t(official Perl source)\n")

			if cfg.PVM == nil || len(cfg.PVM.Remotes) == 0 {
				return nil
			}

			for _, r := range cfg.PVM.Remotes {
				remoteType := r.Type
				if remoteType == "" {
					remoteType = "git"
				}
				ui.Printf("%s\t%s\t[%s]\n", r.Name, r.URL, remoteType)
			}

			return nil
		},
	}
}

// newRemoteRemoveCommand creates 'pvm remote remove <name>'
func newRemoteRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a fork remote source",
		Long:  "Remove a named fork remote source from the user configuration.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			ui := cli.GetUI(cmd)

			// Guard the reserved name early so the error message matches spec.
			if name == "origin" {
				return fmt.Errorf("cannot remove 'origin'")
			}

			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			if cfg.PVM == nil {
				cfg.PVM = &config.PVMConfig{}
			}

			// Build RemoteList from config entries.
			rl := perl.NewRemoteList()
			for _, rc := range cfg.PVM.Remotes {
				existing := perl.Remote{Name: rc.Name, URL: rc.URL, Type: rc.Type}
				_ = rl.Add(existing)
			}

			if err := rl.Remove(name); err != nil {
				return err
			}

			// Sync back to config.
			all := rl.All()
			cfg.PVM.Remotes = make([]config.PVMRemoteConfig, len(all))
			for i, r := range all {
				cfg.PVM.Remotes[i] = config.PVMRemoteConfig{
					Name: r.Name,
					URL:  r.URL,
					Type: r.Type,
				}
			}

			if err := config.SaveUserConfig(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			ui.Success("Removed remote '%s'", name)
			return nil
		},
	}
}
