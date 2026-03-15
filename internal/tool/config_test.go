// ABOUTME: Tests for configuration file handling and custom tool mappings
// ABOUTME: Validates YAML/JSON parsing, merging, and configuration validation
package tool

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigLoader(t *testing.T) {
	loader := NewConfigLoader()

	assert.NotNil(t, loader)
	// systemConfigPath might be empty if no system config exists
	assert.NotEmpty(t, loader.userConfigPath)
}

func TestLoadConfig_EmptyConfig(t *testing.T) {
	loader := &ConfigLoader{
		systemConfigPath: "",
		userConfigPath:   "",
	}

	config, err := loader.LoadConfig()
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.NotNil(t, config.Tools)
	assert.Empty(t, config.Tools)
}

func TestLoadConfig_YAML(t *testing.T) {
	// Create temporary config file
	tempDir, err := os.MkdirTemp("", "pvm-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "tools.yaml")
	yamlContent := `
tools:
  my-tool:
    module: "App::MyTool"
    description: "My custom tool"
    category: "utility"
    executable: "my-tool"
  another-tool:
    module: "Another::Tool"
    description: "Another tool"
    category: "testing"
`

	err = os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	loader := &ConfigLoader{
		systemConfigPath: "",
		userConfigPath:   configFile,
	}

	config, err := loader.LoadConfig()
	require.NoError(t, err)

	assert.Len(t, config.Tools, 2)

	// Check first tool
	myTool, exists := config.Tools["my-tool"]
	assert.True(t, exists)
	assert.Equal(t, "App::MyTool", myTool.Module)
	assert.Equal(t, "My custom tool", myTool.Description)
	assert.Equal(t, "utility", myTool.Category)
	assert.Equal(t, "my-tool", myTool.Executable)

	// Check second tool
	anotherTool, exists := config.Tools["another-tool"]
	assert.True(t, exists)
	assert.Equal(t, "Another::Tool", anotherTool.Module)
	assert.Equal(t, "Another tool", anotherTool.Description)
	assert.Equal(t, "testing", anotherTool.Category)
}

func TestLoadConfig_JSON(t *testing.T) {
	// Create temporary config file
	tempDir, err := os.MkdirTemp("", "pvm-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "tools.json")
	jsonContent := `{
	"tools": {
		"json-tool": {
			"module": "App::JsonTool",
			"description": "JSON configured tool",
			"category": "utility",
			"executable": "json-tool"
		}
	}
}`

	err = os.WriteFile(configFile, []byte(jsonContent), 0644)
	require.NoError(t, err)

	loader := &ConfigLoader{
		systemConfigPath: "",
		userConfigPath:   configFile,
	}

	config, err := loader.LoadConfig()
	require.NoError(t, err)

	assert.Len(t, config.Tools, 1)

	jsonTool, exists := config.Tools["json-tool"]
	assert.True(t, exists)
	assert.Equal(t, "App::JsonTool", jsonTool.Module)
	assert.Equal(t, "JSON configured tool", jsonTool.Description)
	assert.Equal(t, "utility", jsonTool.Category)
	assert.Equal(t, "json-tool", jsonTool.Executable)
}

func TestLoadConfig_ConfigMerging(t *testing.T) {
	// Create temporary config files
	tempDir, err := os.MkdirTemp("", "pvm-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	systemConfigFile := filepath.Join(tempDir, "system.yaml")
	systemContent := `
tools:
  system-tool:
    module: "App::SystemTool"
    description: "System tool"
    category: "system"
  shared-tool:
    module: "App::SystemSharedTool"
    description: "System version"
    category: "system"
`

	userConfigFile := filepath.Join(tempDir, "user.yaml")
	userContent := `
tools:
  user-tool:
    module: "App::UserTool"
    description: "User tool"
    category: "user"
  shared-tool:
    module: "App::UserSharedTool"
    description: "User version"
    category: "user"
`

	err = os.WriteFile(systemConfigFile, []byte(systemContent), 0644)
	require.NoError(t, err)

	err = os.WriteFile(userConfigFile, []byte(userContent), 0644)
	require.NoError(t, err)

	loader := &ConfigLoader{
		systemConfigPath: systemConfigFile,
		userConfigPath:   userConfigFile,
	}

	config, err := loader.LoadConfig()
	require.NoError(t, err)

	assert.Len(t, config.Tools, 3)

	// System-only tool should be present
	systemTool, exists := config.Tools["system-tool"]
	assert.True(t, exists)
	assert.Equal(t, "App::SystemTool", systemTool.Module)

	// User-only tool should be present
	userTool, exists := config.Tools["user-tool"]
	assert.True(t, exists)
	assert.Equal(t, "App::UserTool", userTool.Module)

	// Shared tool should use user config (user overrides system)
	sharedTool, exists := config.Tools["shared-tool"]
	assert.True(t, exists)
	assert.Equal(t, "App::UserSharedTool", sharedTool.Module)
	assert.Equal(t, "User version", sharedTool.Description)
	assert.Equal(t, "user", sharedTool.Category)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Create temporary config file with invalid YAML
	tempDir, err := os.MkdirTemp("", "pvm-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "invalid.yaml")
	invalidContent := `
tools:
  my-tool:
    module: "App::MyTool"
    description: "Unclosed quote
    category: "utility"
`

	err = os.WriteFile(configFile, []byte(invalidContent), 0644)
	require.NoError(t, err)

	loader := &ConfigLoader{
		systemConfigPath: "",
		userConfigPath:   configFile,
	}

	_, err = loader.LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML config")
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Create temporary config file with invalid JSON
	tempDir, err := os.MkdirTemp("", "pvm-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "invalid.json")
	invalidContent := `{
	"tools": {
		"my-tool": {
			"module": "App::MyTool",
			"description": "Missing closing quote
		}
	}
}`

	err = os.WriteFile(configFile, []byte(invalidContent), 0644)
	require.NoError(t, err)

	loader := &ConfigLoader{
		systemConfigPath: "",
		userConfigPath:   configFile,
	}

	_, err = loader.LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON config")
}

func TestValidateConfig_Valid(t *testing.T) {
	loader := NewConfigLoader()

	config := &ToolConfig{
		Tools: map[string]ToolConfigEntry{
			"valid-tool": {
				Module:      "App::ValidTool",
				Description: "A valid tool",
				Category:    "utility",
				Executable:  "valid-tool",
			},
		},
	}

	err := loader.validateConfig(config)
	assert.NoError(t, err)
}

func TestValidateConfig_Invalid(t *testing.T) {
	loader := NewConfigLoader()

	testCases := []struct {
		name   string
		config *ToolConfig
		error  string
	}{
		{
			name: "empty tool name",
			config: &ToolConfig{
				Tools: map[string]ToolConfigEntry{
					"": {Module: "App::Test"},
				},
			},
			error: "tool name cannot be empty",
		},
		{
			name: "empty module name",
			config: &ToolConfig{
				Tools: map[string]ToolConfigEntry{
					"test": {Module: ""},
				},
			},
			error: "module name cannot be empty",
		},
		{
			name: "invalid tool name",
			config: &ToolConfig{
				Tools: map[string]ToolConfigEntry{
					"invalid/tool": {Module: "App::Test"},
				},
			},
			error: "invalid tool name",
		},
		{
			name: "invalid module name",
			config: &ToolConfig{
				Tools: map[string]ToolConfigEntry{
					"test": {Module: "invalid::module::123abc"},
				},
			},
			error: "invalid module name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := loader.validateConfig(tc.config)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.error)
		})
	}
}

func TestSaveUserConfig(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pvm-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "user-tools.yaml")

	loader := &ConfigLoader{
		systemConfigPath: "",
		userConfigPath:   configFile,
	}

	config := &ToolConfig{
		Tools: map[string]ToolConfigEntry{
			"saved-tool": {
				Module:      "App::SavedTool",
				Description: "A saved tool",
				Category:    "utility",
				Executable:  "saved-tool",
			},
		},
	}

	err = loader.SaveUserConfig(config)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configFile)
	assert.NoError(t, err)

	// Load the config back and verify it's correct
	loadedConfig, err := loader.LoadConfig()
	require.NoError(t, err)

	assert.Len(t, loadedConfig.Tools, 1)
	savedTool, exists := loadedConfig.Tools["saved-tool"]
	assert.True(t, exists)
	assert.Equal(t, "App::SavedTool", savedTool.Module)
	assert.Equal(t, "A saved tool", savedTool.Description)
}

func TestCreateDefaultUserConfig(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pvm-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "default-tools.yaml")

	loader := &ConfigLoader{
		systemConfigPath: "",
		userConfigPath:   configFile,
	}

	err = loader.CreateDefaultUserConfig()
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configFile)
	assert.NoError(t, err)

	// Load and verify default config
	config, err := loader.LoadConfig()
	require.NoError(t, err)

	assert.Len(t, config.Tools, 1)
	exampleTool, exists := config.Tools["example-tool"]
	assert.True(t, exists)
	assert.Equal(t, "App::ExampleTool", exampleTool.Module)
	assert.Equal(t, "Example tool configuration", exampleTool.Description)
}
