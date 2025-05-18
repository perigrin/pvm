package config

import (
	"os"
	"testing"
)

func TestGetString(t *testing.T) {
	// Create a test configuration
	config, _ := ParseString(`
		[pvm]
		default_perl = "5.36.0"
		download_mirror = "https://example.com/perl"

		[pvx]
		isolation_level = "high"

		[pvi]
		preferred_installer = "cpanm"

		[psc]
		type_definitions_path = "$XDG_DATA_HOME/pvm/types"
	`)

	// Test getting string values
	tests := []struct {
		section  string
		key      string
		expected string
	}{
		{"pvm", "default_perl", "5.36.0"},
		{"pvm", "download_mirror", "https://example.com/perl"},
		{"pvx", "isolation_level", "high"},
		{"pvi", "preferred_installer", "cpanm"},
		{"psc", "type_definitions_path", "$XDG_DATA_HOME/pvm/types"},
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
	os.Setenv("XDG_DATA_HOME", "/home/user/.local/share")
	value := config.GetString("psc", "type_definitions_path")
	if value != "$XDG_DATA_HOME/pvm/types" {
		t.Errorf("Environment variable not expanded correctly, got %s", value)
	}
}

func TestGetInt(t *testing.T) {
	// Create a test configuration
	config, _ := ParseString(`
		[pvm]
		build_jobs = 8

		[pvx]
		timeout = 300
	`)

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
	config, _ := ParseString(`
		[pvm]
		run_tests = true

		[pvx]
		cache_modules = false
		cleanup_after = true

		[pvi]
		test_during_install = true

		[psc]
		strict_mode = true
	`)

	// Test getting boolean values
	tests := []struct {
		section  string
		key      string
		expected bool
	}{
		{"pvm", "run_tests", true},
		{"pvx", "cache_modules", false},
		{"pvx", "cleanup_after", true},
		{"pvi", "test_during_install", true},
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
	config, _ := ParseString(`
		[psc]
		watch_exclude = ["**/node_modules/**", "**/.git/**"]
	`)

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
	config, _ := ParseString(`
		[pvm]
		version_aliases = { latest = "5.38.0", stable = "5.36.0" }
	`)

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
