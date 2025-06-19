// ABOUTME: PVM config command implementation
// ABOUTME: Provides commands for interacting with the configuration system

package pvm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		newConfigValidateCommand(),
		newConfigDiffCommand(),
		newConfigUnsetCommand(),
		newConfigSourcesCommand(),
		newConfigBackupCommand(),
		newConfigRestoreCommand(),
		newConfigListBackupsCommand(),
		newConfigGenerateCommand(),
	)

	return cmd
}

// newConfigShowCommand creates a command to show the current configuration
func newConfigShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Show the effective configuration (merged from all sources)",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")

			// Use enhanced config manager
			manager := config.NewConfigManager()

			output, err := manager.Show(format)
			if err != nil {
				return errors.NewConfigError("101",
					"Failed to display configuration", err)
			}

			ui := cli.GetUI(cmd)
			ui.Printf("%s", output)
			return nil
		},
	}

	// Add format flag
	cmd.Flags().StringP("format", "f", "toml", "Output format (toml, json, yaml)")

	return cmd
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
			ui := cli.GetUI(cmd)
			ui.Printf("%v\n", result)
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
			return config.SaveToFile(cfg, configPath)
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
	err := config.SaveToFile(cfg, configPath)
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
	case "pvm.binary":
		return key == "default_install_method" ||
			key == "timeout" ||
			key == "bandwidth_limit"
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
	case "pvm.binary":
		return key == "cache_retention_days" ||
			key == "max_cache_size" ||
			key == "max_retries"
	case "pvx":
		return key == "timeout"
	}
	return false
}

func isBoolKey(section, key string) bool {
	switch section {
	case "pvm":
		return key == "run_tests"
	case "pvm.binary":
		return key == "verify_checksums" ||
			key == "parallel_downloads"
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
	case "pvm.binary":
		return key == "binary_mirrors"
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
	case "pvm.binary":
		if cfg.PVM == nil {
			cfg.PVM = &config.PVMConfig{}
		}
		if cfg.PVM.Binary == nil {
			cfg.PVM.Binary = &config.PVMBinaryConfig{}
		}
		return updatePVMBinaryValue(cfg.PVM.Binary, key, value)
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

// updatePVMBinaryValue updates a binary configuration value based on its type
func updatePVMBinaryValue(cfg *config.PVMBinaryConfig, key, value string) error {
	switch key {
	case "default_install_method":
		cfg.DefaultInstallMethod = value
	case "binary_mirrors":
		cfg.BinaryMirrors = parseStringSlice(value)
	case "cache_retention_days":
		var err error
		cfg.CacheRetentionDays, err = parseInt(value)
		if err != nil {
			return err
		}
	case "max_cache_size":
		var err error
		cfg.MaxCacheSize, err = parseInt(value)
		if err != nil {
			return err
		}
	case "verify_checksums":
		var err error
		cfg.VerifyChecksums, err = parseBool(value)
		if err != nil {
			return err
		}
	case "parallel_downloads":
		var err error
		cfg.ParallelDownloads, err = parseBool(value)
		if err != nil {
			return err
		}
	case "max_retries":
		var err error
		cfg.MaxRetries, err = parseInt(value)
		if err != nil {
			return err
		}
	case "timeout":
		cfg.Timeout = value
	case "bandwidth_limit":
		cfg.BandwidthLimit = value
	default:
		return errors.NewUserInputError(cli.PrefixPVM, "113",
			"Unknown binary configuration key", nil).
			WithHint("Use 'pvm config show' to see all available keys")
	}
	return nil
}

// newConfigValidateCommand creates a command to validate configuration
func newConfigValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		Long:  "Validate the current configuration for errors and warnings",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Create schema validator
			validator := config.NewSchemaValidator()

			// Validate with schema
			validationErrors := validator.ValidateConfig(cfg)

			ui := cli.GetUI(cmd)
			if len(validationErrors) == 0 {
				ui.Success("Configuration is valid ✓")
				return nil
			}

			// Report validation errors
			ui.Error("Configuration validation failed with %d errors:", len(validationErrors))
			for i, err := range validationErrors {
				ui.Error("  %d. %s", i+1, err.Error())
			}

			return errors.NewConfigError("200", "Configuration validation failed", nil)
		},
	}
}

// newConfigDiffCommand creates a command to show configuration differences
func newConfigDiffCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff [config-file]",
		Short: "Show configuration differences",
		Long:  "Compare effective configuration with a specific configuration file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load effective configuration
			effectiveConfig, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			var compareConfig *config.Config

			if len(args) == 1 {
				// Compare with specified file
				compareConfig, err = config.ParseFile(args[0])
				if err != nil {
					return err
				}
			} else {
				// Compare with default configuration
				compareConfig = config.NewDefaultConfig()
			}

			// Detect conflicts/differences
			detector := config.NewConflictDetector()
			conflicts := detector.DetectConflicts(compareConfig, effectiveConfig)

			ui := cli.GetUI(cmd)
			if len(conflicts) == 0 {
				ui.Info("No differences found")
				return nil
			}

			ui.Info("Found %d differences:", len(conflicts))
			for _, conflict := range conflicts {
				ui.Info("  %s: %v → %v (%s)",
					conflict.Path,
					conflict.TargetValue,
					conflict.SourceValue,
					conflict.Resolution)
			}

			return nil
		},
	}

	return cmd
}

// newConfigUnsetCommand creates a command to unset configuration values
func newConfigUnsetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset [section.key]",
		Short: "Unset a configuration value",
		Long:  "Remove a specific configuration value (revert to default)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse the section.key argument
			parts := strings.SplitN(args[0], ".", 2)
			if len(parts) != 2 {
				return errors.NewUserInputError(cli.PrefixPVM, "201",
					"Invalid format for section.key", nil).
					WithHint("Use format 'section.key', e.g., 'pvm.default_perl'")
			}

			section := parts[0]
			key := parts[1]

			// Determine which config file to update
			var configPath string
			var err error

			system, _ := cmd.Flags().GetBool("system")
			project, _ := cmd.Flags().GetBool("project")

			switch {
			case system:
				configPath = xdg.GetSystemConfigPath()
			case project:
				projectRoot := config.GetProjectRoot()
				if projectRoot == "" {
					return errors.NewUserInputError(cli.PrefixPVM, "202",
						"Not in a project directory", nil).
						WithHint("Use --user to update user configuration instead")
				}
				configPath = xdg.GetProjectConfigPath(projectRoot)
			default:
				dirs, err := xdg.GetDirs()
				if err != nil {
					return err
				}
				configPath = dirs.GetConfigFilePath()
			}

			// Load existing config or return error if it doesn't exist
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				return errors.NewUserInputError(cli.PrefixPVM, "203",
					"Configuration file does not exist", nil).
					WithLocation(configPath).
					WithHint("Use 'pvm config init' to create a configuration file first")
			}

			cfg, err := config.ParseFile(configPath)
			if err != nil {
				return err
			}

			// Unset the value by setting it to the default value
			defaultCfg := config.NewDefaultConfig()
			err = unsetConfigValue(cfg, defaultCfg, section, key)
			if err != nil {
				return err
			}

			// Save the configuration
			return config.SaveToFile(cfg, configPath)
		},
	}

	// Add flags
	cmd.Flags().Bool("system", false, "Update system-wide configuration")
	cmd.Flags().Bool("user", true, "Update user configuration (default)")
	cmd.Flags().Bool("project", false, "Update project configuration")

	return cmd
}

// unsetConfigValue resets a configuration value to its default
func unsetConfigValue(cfg, defaultCfg *config.Config, section, key string) error {
	switch section {
	case "pvm":
		if cfg.PVM == nil {
			return nil // Already unset
		}
		return unsetPVMValue(cfg.PVM, defaultCfg.PVM, key)
	case "pvx":
		if cfg.PVX == nil {
			return nil // Already unset
		}
		return unsetPVXValue(cfg.PVX, defaultCfg.PVX, key)
	case "pvi":
		if cfg.PVI == nil {
			return nil // Already unset
		}
		return unsetPVIValue(cfg.PVI, defaultCfg.PVI, key)
	case "psc":
		if cfg.PSC == nil {
			return nil // Already unset
		}
		return unsetPSCValue(cfg.PSC, defaultCfg.PSC, key)
	default:
		return errors.NewUserInputError(cli.PrefixPVM, "204",
			"Unknown configuration section", nil).
			WithHint("Valid sections are: pvm, pvx, pvi, psc")
	}
}

func unsetPVMValue(cfg, defaultCfg *config.PVMConfig, key string) error {
	switch key {
	case "default_perl":
		cfg.DefaultPerl = defaultCfg.DefaultPerl
	case "download_mirror":
		cfg.DownloadMirror = defaultCfg.DownloadMirror
	case "patches_dir":
		cfg.PatchesDir = defaultCfg.PatchesDir
	case "compiler":
		cfg.Compiler = defaultCfg.Compiler
	case "build_jobs":
		cfg.BuildJobs = defaultCfg.BuildJobs
	case "run_tests":
		cfg.RunTests = defaultCfg.RunTests
	default:
		return errors.NewUserInputError(cli.PrefixPVM, "205",
			"Unknown configuration key", nil).
			WithHint("Use 'pvm config show' to see all available keys")
	}
	return nil
}

func unsetPVXValue(cfg, defaultCfg *config.PVXConfig, key string) error {
	switch key {
	case "isolation_level":
		cfg.IsolationLevel = defaultCfg.IsolationLevel
	case "max_memory":
		cfg.MaxMemory = defaultCfg.MaxMemory
	case "timeout":
		cfg.Timeout = defaultCfg.Timeout
	case "cache_modules":
		cfg.CacheModules = defaultCfg.CacheModules
	case "cleanup_after":
		cfg.CleanupAfter = defaultCfg.CleanupAfter
	case "always_install_deps":
		cfg.AlwaysInstallDeps = defaultCfg.AlwaysInstallDeps
	case "isolated_output":
		cfg.IsolatedOutput = defaultCfg.IsolatedOutput
	case "save_output_dir":
		cfg.SaveOutputDir = defaultCfg.SaveOutputDir
	case "custom_module_path":
		cfg.CustomModulePath = defaultCfg.CustomModulePath
	case "isolation_ro_paths":
		cfg.IsolationReadOnlyPaths = make([]string, len(defaultCfg.IsolationReadOnlyPaths))
		copy(cfg.IsolationReadOnlyPaths, defaultCfg.IsolationReadOnlyPaths)
	case "isolation_rw_paths":
		cfg.IsolationReadWritePaths = make([]string, len(defaultCfg.IsolationReadWritePaths))
		copy(cfg.IsolationReadWritePaths, defaultCfg.IsolationReadWritePaths)
	case "preserve_env_vars":
		cfg.PreserveEnvVars = make([]string, len(defaultCfg.PreserveEnvVars))
		copy(cfg.PreserveEnvVars, defaultCfg.PreserveEnvVars)
	case "additional_module_paths":
		cfg.AdditionalModulePaths = make([]string, len(defaultCfg.AdditionalModulePaths))
		copy(cfg.AdditionalModulePaths, defaultCfg.AdditionalModulePaths)
	default:
		return errors.NewUserInputError(cli.PrefixPVM, "206",
			"Unknown configuration key", nil).
			WithHint("Use 'pvm config show' to see all available keys")
	}
	return nil
}

func unsetPVIValue(cfg, defaultCfg *config.PVIConfig, key string) error {
	switch key {
	case "preferred_installer":
		cfg.PreferredInstaller = defaultCfg.PreferredInstaller
	case "default_mirror":
		cfg.DefaultMirror = defaultCfg.DefaultMirror
	case "test_during_install":
		cfg.TestDuringInstall = defaultCfg.TestDuringInstall
	case "cache_modules":
		cfg.CacheModules = defaultCfg.CacheModules
	case "force_reinstall":
		cfg.ForceReinstall = defaultCfg.ForceReinstall
	case "check_signatures":
		cfg.CheckSignatures = defaultCfg.CheckSignatures
	default:
		return errors.NewUserInputError(cli.PrefixPVM, "207",
			"Unknown configuration key", nil).
			WithHint("Use 'pvm config show' to see all available keys")
	}
	return nil
}

func unsetPSCValue(cfg, defaultCfg *config.PSCConfig, key string) error {
	switch key {
	case "type_definitions_path":
		cfg.TypeDefinitionsPath = defaultCfg.TypeDefinitionsPath
	case "strict_mode":
		cfg.StrictMode = defaultCfg.StrictMode
	case "generate_missing_types":
		cfg.GenerateMissingTypes = defaultCfg.GenerateMissingTypes
	case "check_before_run":
		cfg.CheckBeforeRun = defaultCfg.CheckBeforeRun
	default:
		return errors.NewUserInputError(cli.PrefixPVM, "208",
			"Unknown configuration key", nil).
			WithHint("Use 'pvm config show' to see all available keys")
	}
	return nil
}

// newConfigSourcesCommand creates a command to show configuration sources
func newConfigSourcesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sources",
		Short: "Show configuration sources",
		Long:  "Display configuration sources and their precedence order",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := config.NewConfigManager()

			output, err := manager.ShowSources()
			if err != nil {
				return errors.NewConfigError("301",
					"Failed to display configuration sources", err)
			}

			ui := cli.GetUI(cmd)
			ui.Printf("%s", strings.TrimSpace(output))
			return nil
		},
	}
}

// newConfigBackupCommand creates a command to backup configuration
func newConfigBackupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup [backup-directory]",
		Short: "Backup configuration files",
		Long:  "Create a backup of current configuration files",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupDir := "~/.config/pvm/backups" // Default backup directory
			if len(args) > 0 {
				backupDir = args[0]
			}

			// Expand tilde
			if strings.HasPrefix(backupDir, "~/") {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return errors.NewSystemError("301", "Failed to get home directory", err)
				}
				backupDir = filepath.Join(homeDir, backupDir[2:])
			}

			manager := config.NewConfigManager()

			err := manager.Backup(backupDir)
			if err != nil {
				return errors.NewConfigError("302",
					"Failed to backup configuration", err)
			}

			ui := cli.GetUI(cmd)
			ui.Success("Configuration backed up to: %s", backupDir)
			return nil
		},
	}

	return cmd
}

// newConfigRestoreCommand creates a command to restore configuration from backup
func newConfigRestoreCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restore [backup-file]",
		Short: "Restore configuration from backup",
		Long:  "Restore configuration from a backup file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupFile := args[0]

			manager := config.NewConfigManager()

			err := manager.Restore(backupFile)
			if err != nil {
				return errors.NewConfigError("303",
					"Failed to restore configuration", err)
			}

			ui := cli.GetUI(cmd)
			ui.Success("Configuration restored from: %s", backupFile)
			return nil
		},
	}
}

// newConfigListBackupsCommand creates a command to list available backups
func newConfigListBackupsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-backups [backup-directory]",
		Short: "List available configuration backups",
		Long:  "List all available configuration backup files",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupDir := "~/.config/pvm/backups" // Default backup directory
			if len(args) > 0 {
				backupDir = args[0]
			}

			// Expand tilde
			if strings.HasPrefix(backupDir, "~/") {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return errors.NewSystemError("302", "Failed to get home directory", err)
				}
				backupDir = filepath.Join(homeDir, backupDir[2:])
			}

			manager := config.NewConfigManager()

			backups, err := manager.ListBackups(backupDir)
			if err != nil {
				return errors.NewConfigError("304",
					"Failed to list backups", err)
			}

			ui := cli.GetUI(cmd)
			if len(backups) == 0 {
				ui.Info("No backups found")
				return nil
			}

			ui.Info("Available backups in %s:", backupDir)
			for _, backup := range backups {
				ui.Info("  %s", backup)
			}

			return nil
		},
	}

	return cmd
}

// newConfigGenerateCommand creates a command to generate configuration templates
func newConfigGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [template]",
		Short: "Generate configuration templates",
		Long:  "Generate configuration templates for common use cases",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			template := "default"
			if len(args) > 0 {
				template = args[0]
			}

			var cfg *config.Config
			switch template {
			case "default":
				cfg = config.NewDefaultConfig()
			case "minimal":
				cfg = &config.Config{
					PVM: &config.PVMConfig{
						DefaultPerl:    "5.40.2",
						BuildJobs:      4,
						DownloadMirror: "https://www.cpan.org/src/5.0",
						RunTests:       true,
						Update: &config.PVMUpdateConfig{
							Repository:           "perigrin/pvm-dev",
							Channel:              "stable",
							BackupEnabled:        true,
							AutoRollbackEnabled:  true,
							NotificationsEnabled: true,
							MaxRetries:           3,
							Timeout:              "5m",
						},
					},
				}
			case "development":
				cfg = config.NewDefaultConfig()
				cfg.PVM.Update.Channel = "developer"
				cfg.PVM.Update.CheckPrerelease = true
				cfg.PVM.Update.AutoUpdateEnabled = true
				cfg.PVM.Update.AutoUpdateInterval = "6h"
			default:
				return errors.NewUserInputError(cli.PrefixPVM, "301",
					"Unknown template", nil).
					WithHint("Available templates: default, minimal, development")
			}

			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// For templates, we'll just output the default TOML format for now
			// since we need the formatting functions that don't exist yet
			if format != "toml" {
				return errors.NewUserInputError(cli.PrefixPVM, "306",
					"Only TOML format is supported for templates currently", nil).
					WithHint("Use --format toml or omit the flag")
			}

			// Save to temporary file and read back the TOML
			tempFile, err := os.CreateTemp("", "pvm-config-template-*.toml")
			if err != nil {
				return errors.NewSystemError("303", "Failed to create temp file", err)
			}
			defer os.Remove(tempFile.Name())

			// Save the generated config to temp file
			if err := config.SaveToFile(cfg, tempFile.Name()); err != nil {
				return errors.NewConfigError("304", "Failed to save template", err)
			}
			tempFile.Close()

			// Read back the TOML content
			content, err := os.ReadFile(tempFile.Name())
			if err != nil {
				return errors.NewConfigError("305", "Failed to read generated config", err)
			}

			fmt.Print(string(content))
			return nil
		},
	}

	// Add format flag
	cmd.Flags().StringP("format", "f", "toml", "Output format (toml, json, yaml)")

	return cmd
}
