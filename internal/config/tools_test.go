// ABOUTME: Tests for configuration management tools and enhanced functionality
// ABOUTME: Validates enhanced configuration features and environment variable handling

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigManager_Show(t *testing.T) {
	// Create temporary directory for test config
	tempDir := t.TempDir()

	// Set up environment to use temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	manager := NewConfigManager()

	tests := []struct {
		name     string
		format   string
		wantErr  bool
		contains string
	}{
		{
			name:     "TOML format",
			format:   "toml",
			wantErr:  false,
			contains: "[pvm]",
		},
		{
			name:     "JSON format",
			format:   "json",
			wantErr:  false,
			contains: "\"pvm\":",
		},
		{
			name:     "YAML format",
			format:   "yaml",
			wantErr:  false,
			contains: "pvm:",
		},
		{
			name:    "Invalid format",
			format:  "xml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manager.Show(tt.format)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, result, tt.contains)
			}
		})
	}
}

func TestConfigManager_GetSet(t *testing.T) {
	// Create temporary directory for test config
	tempDir := t.TempDir()

	// Set up environment to use temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	manager := NewConfigManager()

	// Test setting a value
	err := manager.Set("pvm.default_perl", "5.38.0")
	assert.NoError(t, err)

	// Test getting the value
	value, err := manager.Get("pvm.default_perl")
	assert.NoError(t, err)
	assert.Equal(t, "5.38.0", value)

	// Test unsetting the value
	err = manager.Unset("pvm.default_perl")
	assert.NoError(t, err)
}

func TestConfigManager_Backup(t *testing.T) {
	// Create temporary directories
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backups")

	// Set up environment to use temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	manager := NewConfigManager()

	// Test backup (should work even if no config files exist)
	err := manager.Backup(backupDir)
	assert.NoError(t, err)

	// Check that backup directory was created
	assert.DirExists(t, backupDir)
}

func TestConfigManager_ListBackups(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewConfigManager()

	// Test with empty directory
	backups, err := manager.ListBackups(tempDir)
	assert.NoError(t, err)
	assert.Empty(t, backups)

	// Create a mock backup file
	backupFile := filepath.Join(tempDir, "user-config-20240101-120000.toml")
	err = os.WriteFile(backupFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Test listing backups
	backups, err = manager.ListBackups(tempDir)
	assert.NoError(t, err)
	assert.Len(t, backups, 1)
	assert.Contains(t, backups[0], "user-config-20240101-120000.toml")
}

func TestEnvironmentVariableExpansion(t *testing.T) {
	// Set up test USER environment variable
	oldUser := os.Getenv("USER")
	os.Setenv("USER", "perigrin")
	defer func() {
		if oldUser == "" {
			os.Unsetenv("USER")
		} else {
			os.Setenv("USER", oldUser)
		}
	}()

	tests := []struct {
		name     string
		input    string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "Simple variable",
			input:    "$HOME/test",
			envVar:   "HOME",
			envValue: "/home/user",
			expected: "/home/user/test",
		},
		{
			name:     "Braced variable",
			input:    "${XDG_CONFIG_HOME}/pvm",
			envVar:   "XDG_CONFIG_HOME",
			envValue: "/home/user/.config",
			expected: "/home/user/.config/pvm",
		},
		{
			name:     "Multiple variables",
			input:    "$HOME/${USER}/config",
			envVar:   "HOME",
			envValue: "/home/test",
			expected: "/home/test/perigrin/config", // HOME expanded to test value, USER expanded to current env value
		},
		{
			name:     "No variables",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "Undefined variable",
			input:    "$UNDEFINED_VAR/test",
			expected: "$UNDEFINED_VAR/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envVar != "" && tt.envValue != "" {
				oldValue := os.Getenv(tt.envVar)
				os.Setenv(tt.envVar, tt.envValue)
				defer func() {
					if oldValue == "" {
						os.Unsetenv(tt.envVar)
					} else {
						os.Setenv(tt.envVar, oldValue)
					}
				}()
			}

			result := expandEnvironmentVariables(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdvancedMerger(t *testing.T) {
	merger := NewAdvancedMerger(nil) // Use default config

	// Create test configurations
	config1 := &Config{
		PVM: &PVMConfig{
			DefaultPerl:    "5.36.0",
			BuildJobs:      4,
			VersionAliases: map[string]string{"stable": "5.36.0"},
		},
	}

	config2 := &Config{
		PVM: &PVMConfig{
			DefaultPerl:    "5.38.0",
			RunTests:       true,
			VersionAliases: map[string]string{"latest": "5.38.0"},
		},
	}

	// Test merging
	result := merger.MergeConfigs(config1, config2)

	// Check that later config overrides earlier config
	assert.Equal(t, "5.38.0", result.PVM.DefaultPerl)
	assert.Equal(t, 4, result.PVM.BuildJobs) // From config1
	assert.True(t, result.PVM.RunTests)      // From config2

	// Check map merging
	assert.Contains(t, result.PVM.VersionAliases, "stable")
	assert.Contains(t, result.PVM.VersionAliases, "latest")
}

func TestConflictDetector(t *testing.T) {
	detector := NewConflictDetector()

	config1 := &Config{
		PVM: &PVMConfig{
			DefaultPerl: "5.36.0",
			BuildJobs:   4,
		},
	}

	config2 := &Config{
		PVM: &PVMConfig{
			DefaultPerl: "5.38.0",
			BuildJobs:   8,
		},
	}

	conflicts := detector.DetectConflicts(config1, config2)

	// Should detect conflicts in both fields
	assert.Len(t, conflicts, 2)

	// Check that conflicts are detected correctly
	found := make(map[string]bool)
	for _, conflict := range conflicts {
		found[conflict.Path] = true
	}

	assert.True(t, found["pvm.default_perl"])
	assert.True(t, found["pvm.build_jobs"])
}

func TestConfigAccessors(t *testing.T) {
	config := &Config{
		PVM: &PVMConfig{
			DefaultPerl: "5.38.0",
			BuildJobs:   4,
			RunTests:    true,
		},
	}

	// Test HasSection
	assert.True(t, config.HasSection("pvm"))
	assert.False(t, config.HasSection("nonexistent"))

	// Test HasKey
	assert.True(t, config.HasKey("pvm", "default_perl"))
	assert.False(t, config.HasKey("pvm", "nonexistent_key"))

	// Test GetStringWithDefault
	assert.Equal(t, "5.38.0", config.GetStringWithDefault("pvm", "default_perl", "default"))
	assert.Equal(t, "default", config.GetStringWithDefault("pvm", "nonexistent", "default"))

	// Test GetIntWithDefault
	assert.Equal(t, 4, config.GetIntWithDefault("pvm", "build_jobs", 1))
	assert.Equal(t, 1, config.GetIntWithDefault("pvm", "nonexistent", 1))
}

func TestConfigManagerValidation(t *testing.T) {
	// Create temporary directory for test config
	tempDir := t.TempDir()

	// Set up environment to use temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	manager := NewConfigManager()

	// Test validation (should not error on default config)
	errors, err := manager.Validate()
	assert.NoError(t, err)

	// Default config should be valid
	assert.Empty(t, errors)
}

func TestConfigFormattingYAML(t *testing.T) {
	manager := NewConfigManager()
	config := NewDefaultConfig()

	yaml := manager.formatAsYAML(config)

	// Check that YAML contains expected sections
	assert.Contains(t, yaml, "pvm:")
	assert.Contains(t, yaml, "pvx:")
	assert.Contains(t, yaml, "pvi:")
	assert.Contains(t, yaml, "psc:")

	// Check that it contains some expected values
	assert.Contains(t, yaml, "default_perl:")
	assert.Contains(t, yaml, "isolation_level:")
}

func TestMergeStringSlice(t *testing.T) {
	merger := NewAdvancedMerger(&MergerConfig{
		ArrayStrategy: MergeAppend,
	})

	target := []string{"item1", "item2"}
	source := []string{"item2", "item3"}

	result := merger.mergeStringSlice(target, source)

	// Should append unique items
	assert.Contains(t, result, "item1")
	assert.Contains(t, result, "item2")
	assert.Contains(t, result, "item3")
	assert.Len(t, result, 3) // No duplicates
}

func TestMergeStringSliceReplace(t *testing.T) {
	merger := NewAdvancedMerger(&MergerConfig{
		ArrayStrategy: MergeReplace,
	})

	target := []string{"item1", "item2"}
	source := []string{"item3", "item4"}

	result := merger.mergeStringSlice(target, source)

	// Should replace entirely
	assert.Equal(t, source, result)
}
