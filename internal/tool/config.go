// ABOUTME: Configuration file handling for custom tool mappings
// ABOUTME: Supports YAML and JSON formats with validation and merging capabilities
package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"encoding/json"

	"gopkg.in/yaml.v3"
)

// ToolConfig represents the configuration structure for tool mappings
type ToolConfig struct {
	Tools map[string]ToolConfigEntry `yaml:"tools" json:"tools"`
}

// ToolConfigEntry represents a single tool configuration entry
type ToolConfigEntry struct {
	Module      string `yaml:"module" json:"module"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Category    string `yaml:"category,omitempty" json:"category,omitempty"`
	Executable  string `yaml:"executable,omitempty" json:"executable,omitempty"`
}

// ConfigLoader handles loading and merging tool configurations
type ConfigLoader struct {
	systemConfigPath string
	userConfigPath   string
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		systemConfigPath: getSystemConfigPath(),
		userConfigPath:   getUserConfigPath(),
	}
}

// LoadConfig loads and merges system and user configurations
func (cl *ConfigLoader) LoadConfig() (*ToolConfig, error) {
	config := &ToolConfig{
		Tools: make(map[string]ToolConfigEntry),
	}

	// Load system config first
	if err := cl.loadConfigFromPath(cl.systemConfigPath, config); err != nil {
		// System config is optional, log but don't fail
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load system config: %w", err)
		}
	}

	// Load user config (overrides system config)
	if err := cl.loadConfigFromPath(cl.userConfigPath, config); err != nil {
		// User config is optional
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load user config: %w", err)
		}
	}

	// Validate the merged configuration
	if err := cl.validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// loadConfigFromPath loads configuration from a specific file path
func (cl *ConfigLoader) loadConfigFromPath(path string, config *ToolConfig) error {
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Determine format by file extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return cl.loadYAMLConfig(data, config)
	case ".json":
		return cl.loadJSONConfig(data, config)
	default:
		// Try YAML first, then JSON
		if err := cl.loadYAMLConfig(data, config); err != nil {
			return cl.loadJSONConfig(data, config)
		}
		return nil
	}
}

// loadYAMLConfig loads YAML configuration
func (cl *ConfigLoader) loadYAMLConfig(data []byte, config *ToolConfig) error {
	var tempConfig ToolConfig
	if err := yaml.Unmarshal(data, &tempConfig); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Merge into existing config
	for tool, entry := range tempConfig.Tools {
		config.Tools[tool] = entry
	}

	return nil
}

// loadJSONConfig loads JSON configuration
func (cl *ConfigLoader) loadJSONConfig(data []byte, config *ToolConfig) error {
	var tempConfig ToolConfig
	if err := json.Unmarshal(data, &tempConfig); err != nil {
		return fmt.Errorf("failed to parse JSON config: %w", err)
	}

	// Merge into existing config
	for tool, entry := range tempConfig.Tools {
		config.Tools[tool] = entry
	}

	return nil
}

// validateConfig validates the configuration structure and content
func (cl *ConfigLoader) validateConfig(config *ToolConfig) error {
	for toolName, entry := range config.Tools {
		if err := cl.validateConfigEntry(toolName, entry); err != nil {
			return fmt.Errorf("invalid config entry for tool '%s': %w", toolName, err)
		}
	}
	return nil
}

// validateConfigEntry validates a single configuration entry
func (cl *ConfigLoader) validateConfigEntry(toolName string, entry ToolConfigEntry) error {
	if toolName == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if entry.Module == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	if !isValidToolName(toolName) {
		return fmt.Errorf("invalid tool name '%s'", toolName)
	}

	if !isValidModuleName(entry.Module) {
		return fmt.Errorf("invalid module name '%s'", entry.Module)
	}

	return nil
}

// SaveUserConfig saves configuration to user config file
func (cl *ConfigLoader) SaveUserConfig(config *ToolConfig) error {
	if cl.userConfigPath == "" {
		return fmt.Errorf("user config path not set")
	}

	// Ensure directory exists
	dir := filepath.Dir(cl.userConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(cl.userConfigPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getSystemConfigPath returns the system-wide configuration path
func getSystemConfigPath() string {
	// Try common system config locations
	paths := []string{
		"/etc/pvm/tools.yaml",
		"/usr/local/etc/pvm/tools.yaml",
		"/opt/pvm/tools.yaml",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// getUserConfigPath returns the user-specific configuration path
func getUserConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Use XDG config directory if available
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "pvm", "tools.yaml")
	}

	// Fallback to ~/.config/pvm/tools.yaml
	return filepath.Join(homeDir, ".config", "pvm", "tools.yaml")
}

// CreateDefaultUserConfig creates a default user configuration file
func (cl *ConfigLoader) CreateDefaultUserConfig() error {
	defaultConfig := &ToolConfig{
		Tools: map[string]ToolConfigEntry{
			"example-tool": {
				Module:      "App::ExampleTool",
				Description: "Example tool configuration",
				Category:    "example",
				Executable:  "example-tool",
			},
		},
	}

	return cl.SaveUserConfig(defaultConfig)
}
