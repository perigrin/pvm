// ABOUTME: PVM config command implementation
// ABOUTME: Provides commands for interacting with the configuration system

package pvm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/xdg"
)

// newConfigCommand creates a config command
func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage PVM configuration",
		Long:  "View and edit PVM configuration settings",
	}

	// Add subcommands
	cmd.AddCommand(
		newConfigShowCommand(),
		newConfigGetCommand(),
		newConfigSetCommand(),
		newConfigInitCommand(),
	)

	return cmd
}

// newConfigShowCommand creates a command to show the current configuration
func newConfigShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Show the effective configuration (merged from all sources)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Marshal the configuration to TOML
			data, err := toml.Marshal(cfg)
			if err != nil {
				return errors.NewConfigError("101",
					"Failed to marshal configuration", err)
			}

			// Print the configuration
			fmt.Println(string(data))
			return nil
		},
	}
}

// newConfigGetCommand creates a command to get a specific configuration value
func newConfigGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [section.key]",
		Short: "Get a configuration value",
		Long:  "Get a specific configuration value (e.g., pvm.default_perl)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse the section.key argument
			parts := strings.SplitN(args[0], ".", 2)
			if len(parts) != 2 {
				return errors.NewUserInputError(cli.PrefixPVM, "101",
					"Invalid format for section.key", nil).
					WithHint("Use format 'section.key', e.g., 'pvm.default_perl'")
			}

			section := parts[0]
			key := parts[1]

			// Load configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Get the value based on type
			var result interface{}
			switch {
			case isStringKey(section, key):
				result = cfg.GetString(section, key)
			case isIntKey(section, key):
				result = cfg.GetInt(section, key)
			case isBoolKey(section, key):
				result = cfg.GetBool(section, key)
			case isStringSliceKey(section, key):
				slice := cfg.GetStringSlice(section, key)
				result = fmt.Sprintf("%v", slice)
			case isStringMapKey(section, key):
				m := cfg.GetStringMap(section, key)
				result = fmt.Sprintf("%v", m)
			default:
				return errors.NewUserInputError(cli.PrefixPVM, "102",
					"Unknown configuration key", nil).
					WithHint("Use 'pvm config show' to see all available keys")
			}

			// Print the value
			fmt.Println(result)
			return nil
		},
	}
}

// newConfigSetCommand creates a command to set a configuration value
func newConfigSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [section.key] [value]",
		Short: "Set a configuration value",
		Long:  "Set a specific configuration value (e.g., pvm.default_perl 5.36.0)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse the section.key argument
			parts := strings.SplitN(args[0], ".", 2)
			if len(parts) != 2 {
				return errors.NewUserInputError(cli.PrefixPVM, "103",
					"Invalid format for section.key", nil).
					WithHint("Use format 'section.key', e.g., 'pvm.default_perl'")
			}

			section := parts[0]
			key := parts[1]
			value := args[1]

			// Determine which config file to update
			var configPath string
			var err error

			system, _ := cmd.Flags().GetBool("system")
			_, _ = cmd.Flags().GetBool("user") // Not currently used but keeping for clarity
			project, _ := cmd.Flags().GetBool("project")

			switch {
			case system:
				// System-wide configuration
				configPath = xdg.GetSystemConfigPath()
			case project:
				// Project-level configuration
				projectRoot := config.GetProjectRoot()
				if projectRoot == "" {
					return errors.NewUserInputError(cli.PrefixPVM, "104",
						"Not in a project directory", nil).
						WithHint("Use --user to update user configuration instead")
				}
				configPath = xdg.GetProjectConfigPath(projectRoot)
			default:
				// User-level configuration (default)
				dirs, err := xdg.GetDirs()
				if err != nil {
					return err
				}
				configPath = dirs.GetConfigFilePath()
			}

			// Check if file exists
			var cfg *config.Config
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				// Create new config with defaults
				cfg = config.NewDefaultConfig()
			} else {
				// Load existing config
				cfg, err = config.ParseFile(configPath)
				if err != nil {
					return err
				}
			}

			// Update the value based on type
			err = updateConfigValue(cfg, section, key, value)
			if err != nil {
				return err
			}

			// Save the configuration
			return config.SaveConfig(cfg, configPath)
		},
	}

	// Add flags
	cmd.Flags().Bool("system", false, "Update system-wide configuration")
	cmd.Flags().Bool("user", true, "Update user configuration (default)")
	cmd.Flags().Bool("project", false, "Update project configuration")

	return cmd
}

// newConfigInitCommand creates a command to initialize a configuration file
func newConfigInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a configuration file",
		Long:  "Create a new configuration file with default values",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine which config to initialize
			system, _ := cmd.Flags().GetBool("system")
			project, _ := cmd.Flags().GetBool("project")

			switch {
			case system:
				// System-wide configuration
				configPath := xdg.GetSystemConfigPath()
				return initializeConfig(configPath)
			case project:
				// Project-level configuration
				cwd, err := os.Getwd()
				if err != nil {
					return errors.NewSystemError("101",
						"Failed to get current directory", err)
				}
				return config.InitProjectConfig(cwd)
			default:
				// User-level configuration (default)
				return config.InitUserConfig()
			}
		},
	}

	// Add flags
	cmd.Flags().Bool("system", false, "Create system-wide configuration")
	cmd.Flags().Bool("user", true, "Create user configuration (default)")
	cmd.Flags().Bool("project", false, "Create project configuration")

	return cmd
}

// Helper functions

// initializeConfig initializes a configuration file at the specified path
func initializeConfig(configPath string) error {
	// Check if file exists
	if _, err := os.Stat(configPath); err == nil {
		return errors.NewUserInputError(cli.PrefixPVM, "105",
			"Configuration file already exists", nil).
			WithLocation(configPath).
			WithHint("Use 'pvm config set' to update existing configuration")
	}

	// Create the directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return errors.NewSystemError("102",
			"Failed to create configuration directory", err).
			WithLocation(configDir)
	}

	// Create new config with defaults
	cfg := config.NewDefaultConfig()

	// Save the configuration
	err := config.SaveConfig(cfg, configPath)
	if err != nil {
		return err
	}

	fmt.Printf("Created configuration file: %s\n", configPath)
	return nil
}

// Helper functions to determine the type of a configuration key

func isStringKey(section, key string) bool {
	switch section {
	case "pvm":
		return key == "default_perl" ||
			key == "download_mirror" ||
			key == "patches_dir" ||
			key == "compiler"
	case "pvx":
		return key == "isolation_level" ||
			key == "max_memory" ||
			key == "save_output_dir" ||
			key == "custom_module_path"
	case "pvi":
		return key == "preferred_installer" ||
			key == "default_mirror"
	case "psc":
		return key == "type_definitions_path"
	}
	return false
}

func isIntKey(section, key string) bool {
	switch section {
	case "pvm":
		return key == "build_jobs"
	case "pvx":
		return key == "timeout"
	}
	return false
}

func isBoolKey(section, key string) bool {
	switch section {
	case "pvm":
		return key == "run_tests"
	case "pvx":
		return key == "cache_modules" ||
			key == "cleanup_after" ||
			key == "always_install_deps" ||
			key == "isolated_output"
	case "pvi":
		return key == "test_during_install" ||
			key == "cache_modules" ||
			key == "force_reinstall" ||
			key == "check_signatures"
	case "psc":
		return key == "strict_mode" ||
			key == "generate_missing_types" ||
			key == "check_before_run"
	}
	return false
}

func isStringSliceKey(section, key string) bool {
	switch section {
	case "psc":
		return key == "watch_exclude"
	case "pvx":
		return key == "isolation_ro_paths" ||
			key == "isolation_rw_paths" ||
			key == "preserve_env_vars" ||
			key == "additional_module_paths"
	}
	return false
}

func isStringMapKey(section, key string) bool {
	if section == "pvm" {
		return key == "version_aliases"
	}
	return false
}

// updateConfigValue updates a configuration value based on its type
func updateConfigValue(cfg *config.Config, section, key, value string) error {
	switch section {
	case "pvm":
		if cfg.PVM == nil {
			cfg.PVM = &config.PVMConfig{}
		}
		return updatePVMValue(cfg.PVM, key, value)
	case "pvx":
		if cfg.PVX == nil {
			cfg.PVX = &config.PVXConfig{}
		}
		return updatePVXValue(cfg.PVX, key, value)
	case "pvi":
		if cfg.PVI == nil {
			cfg.PVI = &config.PVIConfig{}
		}
		return updatePVIValue(cfg.PVI, key, value)
	case "psc":
		if cfg.PSC == nil {
			cfg.PSC = &config.PSCConfig{}
		}
		return updatePSCValue(cfg.PSC, key, value)
	default:
		return errors.NewUserInputError(cli.PrefixPVM, "106",
			"Unknown configuration section", nil).
			WithHint("Valid sections are: pvm, pvx, pvi, psc")
	}
}

// Update functions for each configuration section

func updatePVMValue(cfg *config.PVMConfig, key, value string) error {
	switch key {
	case "default_perl":
		cfg.DefaultPerl = value
	case "download_mirror":
		cfg.DownloadMirror = value
	case "patches_dir":
		cfg.PatchesDir = value
	case "compiler":
		cfg.Compiler = value
	case "build_jobs":
		var err error
		cfg.BuildJobs, err = parseInt(value)
		if err != nil {
			return err
		}
	case "run_tests":
		var err error
		cfg.RunTests, err = parseBool(value)
		if err != nil {
			return err
		}
	default:
		return errors.NewUserInputError(cli.PrefixPVM, "107",
			"Unknown configuration key", nil).
			WithHint("Use 'pvm config show' to see all available keys")
	}
	return nil
}

func updatePVXValue(cfg *config.PVXConfig, key, value string) error {
	switch key {
	case "isolation_level":
		cfg.IsolationLevel = value
	case "max_memory":
		cfg.MaxMemory = value
	case "timeout":
		var err error
		cfg.Timeout, err = parseInt(value)
		if err != nil {
			return err
		}
	case "cache_modules":
		var err error
		cfg.CacheModules, err = parseBool(value)
		if err != nil {
			return err
		}
	case "cleanup_after":
		var err error
		cfg.CleanupAfter, err = parseBool(value)
		if err != nil {
			return err
		}
	case "always_install_deps":
		var err error
		cfg.AlwaysInstallDeps, err = parseBool(value)
		if err != nil {
			return err
		}
	case "isolation_ro_paths":
		// Parse the comma-separated list of paths
		cfg.IsolationReadOnlyPaths = parseStringSlice(value)
	case "isolation_rw_paths":
		// Parse the comma-separated list of paths
		cfg.IsolationReadWritePaths = parseStringSlice(value)
	case "isolated_output":
		var err error
		cfg.IsolatedOutput, err = parseBool(value)
		if err != nil {
			return err
		}
	case "save_output_dir":
		cfg.SaveOutputDir = value
	case "preserve_env_vars":
		// Parse the comma-separated list of environment variables
		cfg.PreserveEnvVars = parseStringSlice(value)
	case "additional_module_paths":
		// Parse the comma-separated list of module paths
		cfg.AdditionalModulePaths = parseStringSlice(value)
	case "custom_module_path":
		cfg.CustomModulePath = value
	default:
		return errors.NewUserInputError(cli.PrefixPVM, "108",
			"Unknown configuration key", nil).
			WithHint("Use 'pvm config show' to see all available keys")
	}
	return nil
}

func updatePVIValue(cfg *config.PVIConfig, key, value string) error {
	switch key {
	case "preferred_installer":
		cfg.PreferredInstaller = value
	case "default_mirror":
		cfg.DefaultMirror = value
	case "test_during_install":
		var err error
		cfg.TestDuringInstall, err = parseBool(value)
		if err != nil {
			return err
		}
	case "cache_modules":
		var err error
		cfg.CacheModules, err = parseBool(value)
		if err != nil {
			return err
		}
	case "force_reinstall":
		var err error
		cfg.ForceReinstall, err = parseBool(value)
		if err != nil {
			return err
		}
	case "check_signatures":
		var err error
		cfg.CheckSignatures, err = parseBool(value)
		if err != nil {
			return err
		}
	default:
		return errors.NewUserInputError(cli.PrefixPVM, "109",
			"Unknown configuration key", nil).
			WithHint("Use 'pvm config show' to see all available keys")
	}
	return nil
}

func updatePSCValue(cfg *config.PSCConfig, key, value string) error {
	switch key {
	case "type_definitions_path":
		cfg.TypeDefinitionsPath = value
	case "strict_mode":
		var err error
		cfg.StrictMode, err = parseBool(value)
		if err != nil {
			return err
		}
	case "generate_missing_types":
		var err error
		cfg.GenerateMissingTypes, err = parseBool(value)
		if err != nil {
			return err
		}
	case "check_before_run":
		var err error
		cfg.CheckBeforeRun, err = parseBool(value)
		if err != nil {
			return err
		}
	default:
		return errors.NewUserInputError(cli.PrefixPVM, "110",
			"Unknown configuration key", nil).
			WithHint("Use 'pvm config show' to see all available keys")
	}
	return nil
}

// Helper functions for parsing values

func parseInt(value string) (int, error) {
	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return 0, errors.NewUserInputError(cli.PrefixPVM, "111",
			"Invalid integer value", err).
			WithHint("Value must be a number")
	}
	return result, nil
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "true", "yes", "y", "1", "on":
		return true, nil
	case "false", "no", "n", "0", "off":
		return false, nil
	default:
		return false, errors.NewUserInputError(cli.PrefixPVM, "112",
			"Invalid boolean value", nil).
			WithHint("Valid values are: true/false, yes/no, y/n, 1/0, on/off")
	}
}

func parseStringSlice(value string) []string {
	// Split the comma-separated string into an array
	items := strings.Split(value, ",")

	// Trim whitespace from each item
	for i, item := range items {
		items[i] = strings.TrimSpace(item)
	}

	// Filter out empty items
	var filteredItems []string
	for _, item := range items {
		if item != "" {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems
}
