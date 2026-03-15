package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetString(t *testing.T) {
	// Create a test configuration
	config, err := ParseString(`
		[pvm]
		default_perl = "5.36.0"
		download_mirror = "https://example.com/perl"

		[pvx]
		isolation_level = "clean"

		[pm]
		preferred_installer = "cpanm"

		[psc]
		type_definitions_path = "$XDG_DATA_HOME/pvm/types"
	`)
	if err != nil {
		t.Fatalf("Failed to parse configuration: %v", err)
	}
	if config == nil {
		t.Fatalf("Configuration is nil")
	}

	// Calculate expected XDG fallback for type definitions path
	homeDir, _ := os.UserHomeDir()
	var expectedTypeDefsPath string
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		// If XDG_DATA_HOME is set, use it
		expectedTypeDefsPath = filepath.Join(xdgDataHome, "pvm", "types")
	} else {
		// If XDG_DATA_HOME is unset, use fallback: ~/.local/share
		expectedTypeDefsPath = filepath.Join(homeDir, ".local", "share", "pvm", "types")
	}

	// Test getting string values
	tests := []struct {
		section  string
		key      string
		expected string
	}{
		{"pvm", "default_perl", "5.36.0"},
		{"pvm", "download_mirror", "https://example.com/perl"},
		{"pvx", "isolation_level", "clean"},
		{"pm", "preferred_installer", "cpanm"},
		{"psc", "type_definitions_path", expectedTypeDefsPath},
		// Test non-existent key
		{"pvm", "non_existent", ""},
	}

	for _, test := range tests {
		value := config.GetString(test.section, test.key)
		if value != test.expected {
			t.Errorf("GetString(%s, %s) = %s, expected %s",
				test.section, test.key, value, test.expected)
		}
	}

	// Test environment variable expansion
	oldXDGDataHome := os.Getenv("XDG_DATA_HOME")
	_ = os.Setenv("XDG_DATA_HOME", filepath.FromSlash("/home/user/.local/share"))
	defer func() {
		if oldXDGDataHome == "" {
			os.Unsetenv("XDG_DATA_HOME")
		} else {
			os.Setenv("XDG_DATA_HOME", oldXDGDataHome)
		}
	}()

	value := config.GetString("psc", "type_definitions_path")
	if value != filepath.FromSlash("/home/user/.local/share/pvm/types") {
		t.Errorf("Environment variable not expanded correctly, got %s", value)
	}
}

func TestGetInt(t *testing.T) {
	// Create a test configuration
	config, err := ParseString(`
		[pvm]
		build_jobs = 8

		[pvx]
		timeout = 300
	`)
	if err != nil {
		t.Fatalf("Failed to parse configuration: %v", err)
	}

	// Test getting integer values
	tests := []struct {
		section  string
		key      string
		expected int
	}{
		{"pvm", "build_jobs", 8},
		{"pvx", "timeout", 300},
		// Test non-existent key
		{"pvm", "non_existent", 0},
	}

	for _, test := range tests {
		value := config.GetInt(test.section, test.key)
		if value != test.expected {
			t.Errorf("GetInt(%s, %s) = %d, expected %d",
				test.section, test.key, value, test.expected)
		}
	}
}

func TestGetBool(t *testing.T) {
	// Create a test configuration
	config, err := ParseString(`
		[pvm]
		run_tests = true

		[pvx]
		cache_modules = false
		cleanup_after = true

		[pm]
		test_during_install = true

		[psc]
		strict_mode = true
	`)
	if err != nil {
		t.Fatalf("Failed to parse configuration: %v", err)
	}

	// Test getting boolean values
	tests := []struct {
		section  string
		key      string
		expected bool
	}{
		{"pvm", "run_tests", true},
		{"pvx", "cache_modules", false},
		{"pvx", "cleanup_after", true},
		{"pm", "test_during_install", true},
		{"psc", "strict_mode", true},
		// Test non-existent key
		{"pvm", "non_existent", false},
	}

	for _, test := range tests {
		value := config.GetBool(test.section, test.key)
		if value != test.expected {
			t.Errorf("GetBool(%s, %s) = %v, expected %v",
				test.section, test.key, value, test.expected)
		}
	}
}

func TestGetStringSlice(t *testing.T) {
	// Create a test configuration
	config, err := ParseString(`
		[psc]
		watch_exclude = ["**/node_modules/**", "**/.git/**"]
	`)
	if err != nil {
		t.Fatalf("Failed to parse configuration: %v", err)
	}

	// Test getting string slice values
	watchExclude := config.GetStringSlice("psc", "watch_exclude")
	if len(watchExclude) != 2 {
		t.Errorf("Expected 2 watch_exclude entries, got %d", len(watchExclude))
	}

	if watchExclude[0] != "**/node_modules/**" {
		t.Errorf("Expected watch_exclude[0] = '**/node_modules/**', got '%s'", watchExclude[0])
	}

	if watchExclude[1] != "**/.git/**" {
		t.Errorf("Expected watch_exclude[1] = '**/.git/**', got '%s'", watchExclude[1])
	}

	// Test non-existent key
	nonExistent := config.GetStringSlice("psc", "non_existent")
	if nonExistent != nil {
		t.Errorf("Expected nil for non-existent key, got %v", nonExistent)
	}
}

func TestGetStringMap(t *testing.T) {
	// Create a test configuration
	config, err := ParseString(`
		[pvm]
		version_aliases = { latest = "5.38.0", stable = "5.36.0" }
	`)
	if err != nil {
		t.Fatalf("Failed to parse configuration: %v", err)
	}

	// Test getting string map values
	versionAliases := config.GetStringMap("pvm", "version_aliases")
	if len(versionAliases) != 2 {
		t.Errorf("Expected 2 version_aliases entries, got %d", len(versionAliases))
	}

	if versionAliases["latest"] != "5.38.0" {
		t.Errorf("Expected version_aliases['latest'] = '5.38.0', got '%s'", versionAliases["latest"])
	}

	if versionAliases["stable"] != "5.36.0" {
		t.Errorf("Expected version_aliases['stable'] = '5.36.0', got '%s'", versionAliases["stable"])
	}

	// Test non-existent key
	nonExistent := config.GetStringMap("pvm", "non_existent")
	if nonExistent != nil {
		t.Errorf("Expected nil for non-existent key, got %v", nonExistent)
	}
}

// TestExpandEnvironmentVariables_XDGFallbacks tests comprehensive XDG fallback behavior
func TestExpandEnvironmentVariables_XDGFallbacks(t *testing.T) {
	t.Run("XDG_Fallbacks_Comprehensive", func(t *testing.T) {
		// Test cases for all XDG variables
		testCases := []struct {
			envVar       string
			inputPath    string
			expectedPath string
		}{
			{
				envVar:       "XDG_CACHE_HOME",
				inputPath:    "$XDG_CACHE_HOME/pvm",
				expectedPath: ".cache/pvm",
			},
			{
				envVar:       "XDG_DATA_HOME",
				inputPath:    "$XDG_DATA_HOME/pvm",
				expectedPath: ".local/share/pvm",
			},
			{
				envVar:       "XDG_CONFIG_HOME",
				inputPath:    "$XDG_CONFIG_HOME/pvm",
				expectedPath: ".config/pvm",
			},
			{
				envVar:       "XDG_STATE_HOME",
				inputPath:    "$XDG_STATE_HOME/pvm",
				expectedPath: ".local/state/pvm",
			},
		}

		homeDir, _ := os.UserHomeDir()

		for _, tc := range testCases {
			t.Run("Fallback_"+tc.envVar, func(t *testing.T) {
				// Ensure the environment variable is unset
				originalValue := os.Getenv(tc.envVar)
				_ = os.Unsetenv(tc.envVar)
				defer func() {
					if originalValue != "" {
						_ = os.Setenv(tc.envVar, originalValue)
					}
				}()

				// Test expansion
				result := expandEnvironmentVariables(tc.inputPath)
				expected := filepath.Join(homeDir, tc.expectedPath)

				if result != expected {
					t.Errorf("expandEnvironmentVariables(%s) = %s, expected XDG fallback %s", tc.inputPath, result, expected)
				}

				if strings.Contains(result, "$") {
					t.Errorf("Result %s should not contain literal environment variables", result)
				}
			})
		}
	})

	t.Run("XDG_Fallbacks_With_Braces", func(t *testing.T) {
		// Test XDG fallbacks with ${VAR} format
		_ = os.Unsetenv("XDG_DATA_HOME")
		defer func() {
			// Restore if it was set
			if val := os.Getenv("XDG_DATA_HOME"); val != "" {
				_ = os.Setenv("XDG_DATA_HOME", val)
			}
		}()

		homeDir, _ := os.UserHomeDir()
		result := expandEnvironmentVariables("${XDG_DATA_HOME}/pvm/types")
		expected := filepath.Join(homeDir, ".local", "share", "pvm", "types")

		if result != expected {
			t.Errorf("expandEnvironmentVariables(${XDG_DATA_HOME}/pvm/types) = %s, expected XDG fallback %s", result, expected)
		}

		if strings.Contains(result, "$") {
			t.Errorf("Result %s should not contain literal environment variables", result)
		}
	})

	t.Run("Non_XDG_Variable_Unset", func(t *testing.T) {
		// Test that non-XDG variables still return literal when unset
		_ = os.Unsetenv("NON_XDG_VAR")

		result := expandEnvironmentVariables("$NON_XDG_VAR/test")
		expected := "$NON_XDG_VAR/test"

		if result != expected {
			t.Errorf("expandEnvironmentVariables($NON_XDG_VAR/test) = %s, expected literal %s", result, expected)
		}
	})

	t.Run("Mixed_XDG_And_Regular_Variables", func(t *testing.T) {
		// Test mixing XDG fallbacks with regular environment variables
		_ = os.Unsetenv("XDG_CACHE_HOME")
		_ = os.Setenv("TEST_PREFIX", "custom")
		defer func() {
			_ = os.Unsetenv("TEST_PREFIX")
		}()

		homeDir, _ := os.UserHomeDir()
		result := expandEnvironmentVariables("$TEST_PREFIX/$XDG_CACHE_HOME/pvm")
		// When XDG_CACHE_HOME is unset, it expands to absolute path ~/.cache
		// filepath.Clean normalizes the double slash between "custom" and the absolute fallback
		expected := filepath.Clean("custom" + "/" + filepath.Join(homeDir, ".cache") + "/pvm")

		if result != expected {
			t.Errorf("expandEnvironmentVariables($TEST_PREFIX/$XDG_CACHE_HOME/pvm) = %s, expected %s", result, expected)
		}
	})
}
