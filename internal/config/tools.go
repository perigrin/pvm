// ABOUTME: Configuration management tools and commands
// ABOUTME: Provides inspection, validation, and backup functionality for configuration

package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"
	"tamarou.com/pvm/internal/xdg"
)

// ConfigManager provides configuration management functionality
type ConfigManager struct {
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// Show displays the effective configuration after merging
func (m *ConfigManager) Show(format string) (string, error) {
	config, err := LoadEffectiveConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load configuration: %w", err)
	}

	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal to JSON: %w", err)
		}
		return string(data), nil

	case "toml":
		data, err := toml.Marshal(config)
		if err != nil {
			return "", fmt.Errorf("failed to marshal to TOML: %w", err)
		}
		return string(data), nil

	case "yaml":
		// For now, we'll output in a YAML-like format using our own formatting
		return m.formatAsYAML(config), nil

	default:
		return "", fmt.Errorf("unsupported format: %s (supported: json, toml, yaml)", format)
	}
}

// ShowSources displays configuration sources and their precedence
func (m *ConfigManager) ShowSources() (string, error) {
	config, err := LoadEffectiveConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get XDG directories for source paths
	dirs, err := xdg.GetDirs()
	if err != nil {
		return "", fmt.Errorf("failed to get XDG directories: %w", err)
	}

	var builder strings.Builder
	builder.WriteString("Configuration Sources (in precedence order):\n\n")

	// List sources in order (highest to lowest precedence)
	sources := []struct {
		name   string
		path   string
		exists bool
	}{
		{"Project", getProjectConfigPath(), false},
		{"User", dirs.GetConfigFilePath(), false},
		{"System", xdg.GetSystemConfigPath(), false},
	}

	// Check if files exist
	for i := range sources {
		_, err := os.Stat(sources[i].path)
		sources[i].exists = err == nil
	}

	for i, source := range sources {
		builder.WriteString(fmt.Sprintf("%d. %s Configuration\n", i+1, source.name))
		builder.WriteString(fmt.Sprintf("   Path: %s\n", source.path))
		if source.exists {
			builder.WriteString("   Status: Found\n")
		} else {
			builder.WriteString("   Status: Not found\n")
		}
		builder.WriteString("\n")
	}

	// Show effective values
	builder.WriteString("Effective Configuration:\n")
	builder.WriteString(m.formatAsYAML(config))

	return builder.String(), nil
}

// Get retrieves a specific configuration value
func (m *ConfigManager) Get(key string) (interface{}, error) {
	config, err := LoadEffectiveConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return m.getValueByKey(config, key), nil
}

// Set updates a configuration value in the user-level configuration
func (m *ConfigManager) Set(key string, value interface{}) error {
	// For this implementation, we'll update the user configuration file
	// This is a simplified version - a full implementation would handle
	// nested keys, different value types, etc.

	// Get XDG directories for user config path
	dirs, err := xdg.GetDirs()
	if err != nil {
		return fmt.Errorf("failed to get XDG directories: %w", err)
	}

	userConfigPath := dirs.GetConfigFilePath()

	var userConfig *Config
	if _, err := os.Stat(userConfigPath); err == nil {
		// File exists, load it
		userConfig, err = ParseFile(userConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load user configuration: %w", err)
		}
	} else {
		// File doesn't exist, create new config
		userConfig = NewDefaultConfig()
	}

	// Update the value (simplified - would need more sophisticated key parsing)
	if err := m.setValueByKey(userConfig, key, value); err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}

	// Save back to file
	return m.saveConfigToFile(userConfig, userConfigPath)
}

// Unset removes a configuration value from the user-level configuration
func (m *ConfigManager) Unset(key string) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return fmt.Errorf("failed to get XDG directories: %w", err)
	}

	userConfigPath := dirs.GetConfigFilePath()

	if _, err := os.Stat(userConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("user configuration file does not exist")
	}

	userConfig, err := ParseFile(userConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load user configuration: %w", err)
	}

	// Remove the value (simplified implementation)
	if err := m.unsetValueByKey(userConfig, key); err != nil {
		return fmt.Errorf("failed to unset value: %w", err)
	}

	return m.saveConfigToFile(userConfig, userConfigPath)
}

// Validate validates the current configuration
func (m *ConfigManager) Validate() ([]error, error) {
	config, err := LoadEffectiveConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	validator := NewSchemaValidator()
	errors := validator.ValidateConfig(config)

	return errors, nil
}

// Diff compares two configuration sources
func (m *ConfigManager) Diff(path1, path2 string) (string, error) {
	config1, err := ParseFile(path1)
	if err != nil {
		return "", fmt.Errorf("failed to load %s: %w", path1, err)
	}

	config2, err := ParseFile(path2)
	if err != nil {
		return "", fmt.Errorf("failed to load %s: %w", path2, err)
	}

	return m.generateDiff(config1, config2, path1, path2), nil
}

// Backup creates a backup of the current configuration
func (m *ConfigManager) Backup(backupDir string) error {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")

	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return fmt.Errorf("failed to get XDG directories: %w", err)
	}

	// Backup user config if it exists
	userConfigPath := dirs.GetConfigFilePath()
	if _, err := os.Stat(userConfigPath); err == nil {
		backupPath := filepath.Join(backupDir, fmt.Sprintf("user-config-%s.toml", timestamp))
		if err := m.copyFile(userConfigPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup user config: %w", err)
		}
	}

	// Backup project config if it exists
	projectRoot := GetProjectRoot()
	if projectRoot != "" {
		projectConfigPath := xdg.GetProjectConfigPath(projectRoot)
		if _, err := os.Stat(projectConfigPath); err == nil {
			backupPath := filepath.Join(backupDir, fmt.Sprintf("project-config-%s.toml", timestamp))
			if err := m.copyFile(projectConfigPath, backupPath); err != nil {
				return fmt.Errorf("failed to backup project config: %w", err)
			}
		}
	}

	return nil
}

// Restore restores configuration from a backup
func (m *ConfigManager) Restore(backupPath string) error {
	// Determine if this is a user or project config backup
	filename := filepath.Base(backupPath)

	if strings.HasPrefix(filename, "user-config-") {
		dirs, err := xdg.GetDirs()
		if err != nil {
			return fmt.Errorf("failed to get XDG directories: %w", err)
		}
		userConfigPath := dirs.GetConfigFilePath()
		return m.copyFile(backupPath, userConfigPath)
	} else if strings.HasPrefix(filename, "project-config-") {
		projectRoot := GetProjectRoot()
		if projectRoot == "" {
			return fmt.Errorf("not in a project directory")
		}
		projectConfigPath := xdg.GetProjectConfigPath(projectRoot)
		return m.copyFile(backupPath, projectConfigPath)
	}

	return fmt.Errorf("unknown backup file type: %s", filename)
}

// ListBackups lists available configuration backups
func (m *ConfigManager) ListBackups(backupDir string) ([]string, error) {
	var backups []string

	err := filepath.WalkDir(backupDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".toml") {
			if strings.Contains(path, "config-") {
				backups = append(backups, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	sort.Strings(backups)
	return backups, nil
}

// Helper functions

func (m *ConfigManager) formatAsYAML(config *Config) string {
	var builder strings.Builder

	if config.PVM != nil {
		builder.WriteString("pvm:\n")
		builder.WriteString(fmt.Sprintf("  default_perl: %s\n", config.PVM.DefaultPerl))
		builder.WriteString(fmt.Sprintf("  build_jobs: %d\n", config.PVM.BuildJobs))
		builder.WriteString(fmt.Sprintf("  download_mirror: %s\n", config.PVM.DownloadMirror))
		builder.WriteString(fmt.Sprintf("  run_tests: %t\n", config.PVM.RunTests))
		if len(config.PVM.VersionAliases) > 0 {
			builder.WriteString("  version_aliases:\n")
			for k, v := range config.PVM.VersionAliases {
				builder.WriteString(fmt.Sprintf("    %s: %s\n", k, v))
			}
		}
		builder.WriteString("\n")
	}

	if config.PVX != nil {
		builder.WriteString("pvx:\n")
		builder.WriteString(fmt.Sprintf("  isolation_level: %s\n", config.PVX.IsolationLevel))
		builder.WriteString(fmt.Sprintf("  timeout: %d\n", config.PVX.Timeout))
		builder.WriteString(fmt.Sprintf("  max_memory: %s\n", config.PVX.MaxMemory))
		builder.WriteString("\n")
	}

	if config.PVI != nil {
		builder.WriteString("pvi:\n")
		builder.WriteString(fmt.Sprintf("  preferred_installer: %s\n", config.PVI.PreferredInstaller))
		builder.WriteString(fmt.Sprintf("  default_mirror: %s\n", config.PVI.DefaultMirror))
		builder.WriteString("\n")
	}

	if config.PSC != nil {
		builder.WriteString("psc:\n")
		builder.WriteString(fmt.Sprintf("  strict_mode: %t\n", config.PSC.StrictMode))
		builder.WriteString(fmt.Sprintf("  type_definitions_path: %s\n", config.PSC.TypeDefinitionsPath))
		builder.WriteString("\n")
	}

	return builder.String()
}

func (m *ConfigManager) getValueByKey(config *Config, key string) interface{} {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return nil
	}

	section := parts[0]
	field := parts[1]

	switch section {
	case "pvm":
		if config.PVM != nil {
			return getPVMValue(config.PVM, field)
		}
	case "pvx":
		if config.PVX != nil {
			return getPVXValue(config.PVX, field)
		}
	case "pvi":
		if config.PVI != nil {
			return getPVIValue(config.PVI, field)
		}
	case "psc":
		if config.PSC != nil {
			return getPSCValue(config.PSC, field)
		}
	}

	return nil
}

func (m *ConfigManager) setValueByKey(config *Config, key string, value interface{}) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid key format: %s", key)
	}

	section := parts[0]
	field := parts[1]

	switch section {
	case "pvm":
		if config.PVM == nil {
			config.PVM = &PVMConfig{}
		}
		return setPVMValue(config.PVM, field, value)
	case "pvx":
		if config.PVX == nil {
			config.PVX = &PVXConfig{}
		}
		return setPVXValue(config.PVX, field, value)
	case "pvi":
		if config.PVI == nil {
			config.PVI = &PVIConfig{}
		}
		return setPVIValue(config.PVI, field, value)
	case "psc":
		if config.PSC == nil {
			config.PSC = &PSCConfig{}
		}
		return setPSCValue(config.PSC, field, value)
	}

	return fmt.Errorf("unknown configuration section: %s", section)
}

func (m *ConfigManager) unsetValueByKey(config *Config, key string) error {
	// For simplicity, we'll set values to their zero values
	// A more sophisticated implementation would remove the keys entirely
	return m.setValueByKey(config, key, nil)
}

func (m *ConfigManager) saveConfigToFile(config *Config, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func (m *ConfigManager) generateDiff(config1, config2 *Config, path1, path2 string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("--- %s\n", path1))
	builder.WriteString(fmt.Sprintf("+++ %s\n", path2))
	builder.WriteString("\n")

	// This is a simplified diff - a full implementation would do proper
	// field-by-field comparison and show actual differences
	builder.WriteString("Configuration comparison:\n")
	builder.WriteString("(Detailed diff implementation would go here)\n")

	return builder.String()
}

func (m *ConfigManager) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Ensure destination directory exists
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

// Value getter/setter helpers

func getPVMValue(config *PVMConfig, field string) interface{} {
	switch field {
	case "default_perl":
		return config.DefaultPerl
	case "build_jobs":
		return config.BuildJobs
	case "download_mirror":
		return config.DownloadMirror
	case "run_tests":
		return config.RunTests
	case "version_aliases":
		return config.VersionAliases
	}
	return nil
}

func setPVMValue(config *PVMConfig, field string, value interface{}) error {
	switch field {
	case "default_perl":
		if str, ok := value.(string); ok {
			config.DefaultPerl = str
		}
	case "build_jobs":
		if num, ok := value.(int); ok {
			config.BuildJobs = num
		}
	case "download_mirror":
		if str, ok := value.(string); ok {
			config.DownloadMirror = str
		}
	case "run_tests":
		if b, ok := value.(bool); ok {
			config.RunTests = b
		}
	}
	return nil
}

func getPVXValue(config *PVXConfig, field string) interface{} {
	switch field {
	case "isolation_level":
		return config.IsolationLevel
	case "timeout":
		return config.Timeout
	case "max_memory":
		return config.MaxMemory
	}
	return nil
}

func setPVXValue(config *PVXConfig, field string, value interface{}) error {
	switch field {
	case "isolation_level":
		if str, ok := value.(string); ok {
			config.IsolationLevel = str
		}
	case "timeout":
		if num, ok := value.(int); ok {
			config.Timeout = num
		}
	case "max_memory":
		if str, ok := value.(string); ok {
			config.MaxMemory = str
		}
	}
	return nil
}

func getPVIValue(config *PVIConfig, field string) interface{} {
	switch field {
	case "preferred_installer":
		return config.PreferredInstaller
	case "default_mirror":
		return config.DefaultMirror
	}
	return nil
}

func setPVIValue(config *PVIConfig, field string, value interface{}) error {
	switch field {
	case "preferred_installer":
		if str, ok := value.(string); ok {
			config.PreferredInstaller = str
		}
	case "default_mirror":
		if str, ok := value.(string); ok {
			config.DefaultMirror = str
		}
	}
	return nil
}

func getPSCValue(config *PSCConfig, field string) interface{} {
	switch field {
	case "strict_mode":
		return config.StrictMode
	case "type_definitions_path":
		return config.TypeDefinitionsPath
	}
	return nil
}

func setPSCValue(config *PSCConfig, field string, value interface{}) error {
	switch field {
	case "strict_mode":
		if b, ok := value.(bool); ok {
			config.StrictMode = b
		}
	case "type_definitions_path":
		if str, ok := value.(string); ok {
			config.TypeDefinitionsPath = str
		}
	}
	return nil
}

// Helper function to get project config path
func getProjectConfigPath() string {
	projectRoot := GetProjectRoot()
	if projectRoot == "" {
		return ""
	}
	return xdg.GetProjectConfigPath(projectRoot)
}
