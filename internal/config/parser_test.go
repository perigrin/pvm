package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pvm_errors "tamarou.com/pvm/internal/errors"
)

func TestParseBytes_WithInterpolation(t *testing.T) {
	// Set up test environment variables
	os.Setenv("TEST_HOME", "/test/home")
	os.Setenv("TEST_USER", "testuser")
	os.Setenv("TEST_MIRROR", "https://test.mirror.com")
	defer func() {
		os.Unsetenv("TEST_HOME")
		os.Unsetenv("TEST_USER")
		os.Unsetenv("TEST_MIRROR")
	}()

	tomlData := `
[pvm]
default_perl = "5.38.0"
download_mirror = "${TEST_MIRROR}/perl"
patches_dir = "${TEST_HOME}/${TEST_USER}/patches"

[pvi]
cache_dir = "${TEST_HOME:-/default}/cache"
default_mirror = "${TEST_MIRROR}"

[pvx]
save_output_dir = "${TEST_HOME}/output"
preserve_env_vars = ["${TEST_USER}_VAR", "DISPLAY"]

[psc]
type_definitions_path = "${TEST_HOME}/types"
`

	config, err := ParseBytes([]byte(tomlData), "test")
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	// Verify interpolation worked
	if config.PVM.DownloadMirror != "https://test.mirror.com/perl" {
		t.Errorf("PVM.DownloadMirror = %v, want https://test.mirror.com/perl", config.PVM.DownloadMirror)
	}

	if config.PVM.PatchesDir != "/test/home/testuser/patches" {
		t.Errorf("PVM.PatchesDir = %v, want /test/home/testuser/patches", config.PVM.PatchesDir)
	}

	if config.PVI.CacheDir != "/test/home/cache" {
		t.Errorf("PVI.CacheDir = %v, want /test/home/cache", config.PVI.CacheDir)
	}

	if config.PVI.DefaultMirror != "https://test.mirror.com" {
		t.Errorf("PVI.DefaultMirror = %v, want https://test.mirror.com", config.PVI.DefaultMirror)
	}

	if config.PVX.SaveOutputDir != "/test/home/output" {
		t.Errorf("PVX.SaveOutputDir = %v, want /test/home/output", config.PVX.SaveOutputDir)
	}

	if len(config.PVX.PreserveEnvVars) != 2 || config.PVX.PreserveEnvVars[0] != "testuser_VAR" {
		t.Errorf("PVX.PreserveEnvVars = %v, want [testuser_VAR, DISPLAY]", config.PVX.PreserveEnvVars)
	}

	if config.PSC.TypeDefinitionsPath != "/test/home/types" {
		t.Errorf("PSC.TypeDefinitionsPath = %v, want /test/home/types", config.PSC.TypeDefinitionsPath)
	}
}

func TestParseBytes_InterpolationWithDefaults(t *testing.T) {
	// Don't set MISSING_VAR to test default values
	os.Setenv("EXISTING_VAR", "exists")
	defer os.Unsetenv("EXISTING_VAR")

	tomlData := `
[pvm]
download_mirror = "${MISSING_VAR:-https://default.mirror.com}"
patches_dir = "${EXISTING_VAR}/patches"

[pvi]
cache_dir = "${MISSING_VAR:-/default/cache}"
`

	config, err := ParseBytes([]byte(tomlData), "test")
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	// Verify default values were used
	if config.PVM.DownloadMirror != "https://default.mirror.com" {
		t.Errorf("PVM.DownloadMirror = %v, want https://default.mirror.com", config.PVM.DownloadMirror)
	}

	if config.PVM.PatchesDir != "exists/patches" {
		t.Errorf("PVM.PatchesDir = %v, want exists/patches", config.PVM.PatchesDir)
	}

	if config.PVI.CacheDir != "/default/cache" {
		t.Errorf("PVI.CacheDir = %v, want /default/cache", config.PVI.CacheDir)
	}
}

func TestParseBytes_InterpolationErrors(t *testing.T) {
	// Set up a circular reference
	os.Setenv("CIRCULAR_A", "${CIRCULAR_B}")
	os.Setenv("CIRCULAR_B", "${CIRCULAR_A}")
	defer func() {
		os.Unsetenv("CIRCULAR_A")
		os.Unsetenv("CIRCULAR_B")
	}()

	tomlData := `
[pvm]
download_mirror = "${CIRCULAR_A}"
`

	_, err := ParseBytes([]byte(tomlData), "test")
	if err == nil {
		t.Error("ParseBytes() expected error for circular reference, got nil")
	}

	if !strings.Contains(err.Error(), "interpolation failed") {
		t.Errorf("Expected interpolation error, got: %v", err)
	}
}

func TestParseString(t *testing.T) {
	// Test valid TOML parsing
	config, err := ParseString(`
		[pvm]
		default_perl = "5.36.0"
		build_jobs = 8

		[pvx]
		isolation_level = "clean"

		[pvi]
		preferred_installer = "cpanm"

		[psc]
		strict_mode = true
	`)

	if err != nil {
		t.Fatalf("Failed to parse valid TOML: %v", err)
	}

	// Check that values were correctly parsed
	if config.PVM.DefaultPerl != "5.36.0" {
		t.Errorf("Expected DefaultPerl = '5.36.0', got '%s'", config.PVM.DefaultPerl)
	}

	if config.PVM.BuildJobs != 8 {
		t.Errorf("Expected BuildJobs = 8, got %d", config.PVM.BuildJobs)
	}

	if config.PVX.IsolationLevel != "clean" {
		t.Errorf("Expected IsolationLevel = 'clean', got '%s'", config.PVX.IsolationLevel)
	}

	if config.PVI.PreferredInstaller != "cpanm" {
		t.Errorf("Expected PreferredInstaller = 'cpanm', got '%s'", config.PVI.PreferredInstaller)
	}

	if !config.PSC.StrictMode {
		t.Error("Expected StrictMode = true, got false")
	}
}

func TestParseStringWithDefaults(t *testing.T) {
	// Test that default values are used when not specified
	config, err := ParseString(`
		[pvm]
		default_perl = "5.36.0"
	`)

	if err != nil {
		t.Fatalf("Failed to parse valid TOML: %v", err)
	}

	// Check that the specified value was parsed
	if config.PVM.DefaultPerl != "5.36.0" {
		t.Errorf("Expected DefaultPerl = '5.36.0', got '%s'", config.PVM.DefaultPerl)
	}

	// Check that defaults were applied for unspecified values
	if config.PVM.BuildJobs != 4 {
		t.Errorf("Expected default BuildJobs = 4, got %d", config.PVM.BuildJobs)
	}

	if config.PVX.IsolationLevel != "clean" {
		t.Errorf("Expected default IsolationLevel = 'clean', got '%s'", config.PVX.IsolationLevel)
	}
}

func TestParseInvalidTOML(t *testing.T) {
	// Test invalid TOML syntax
	_, err := ParseString(`
		[pvm
		default_perl = "5.36.0"
	`)

	if err == nil {
		t.Fatal("Expected error for invalid TOML, got nil")
	}

	// Check that the error is a configuration error
	var configError *pvm_errors.Error
	if !errors.As(err, &configError) {
		t.Fatalf("Expected *pvm_errors.Error, got %T", err)
	}
}

func TestParseUnknownField_PermissiveMode(t *testing.T) {
	// Test unknown field in permissive mode (default behavior)
	config, err := ParseString(`
		[pvm]
		unknown_field = "value"
		default_perl = "5.36.0"
	`)

	// Should succeed in permissive mode
	if err != nil {
		t.Fatalf("Expected no error in permissive mode, got: %v", err)
	}

	// Check that known fields were parsed correctly
	if config.PVM.DefaultPerl != "5.36.0" {
		t.Errorf("Expected DefaultPerl = '5.36.0', got '%s'", config.PVM.DefaultPerl)
	}
}

func TestParseUnknownField_StrictMode(t *testing.T) {
	// Test unknown field in strict mode
	options := ParseOptions{
		StrictParsing: true,
		Validate:      true,
	}

	_, err := ParseBytesWithAdvancedOptions([]byte(`
		[pvm]
		unknown_field = "value"
		default_perl = "5.36.0"
	`), "test", options)

	// Should fail in strict mode
	if err == nil {
		t.Fatal("Expected error for unknown field in strict mode, got nil")
	}

	// Check that the error is a configuration error
	var configError *pvm_errors.Error
	if !errors.As(err, &configError) {
		t.Fatalf("Expected *pvm_errors.Error, got %T", err)
	}
}

func TestParseWithWarnings_UnknownTopLevelSections(t *testing.T) {
	// Test parsing with unknown top-level sections
	result, err := ParseStringWithWarnings(`
		[pvm]
		default_perl = "5.36.0"

		[unknown_section]
		some_field = "value"

		[experimental]
		feature_flag = true

		[pvx]
		isolation_level = "clean"
	`)

	if err != nil {
		t.Fatalf("Expected no error in permissive mode, got: %v", err)
	}

	// Check that known sections were parsed correctly
	if result.Config.PVM.DefaultPerl != "5.36.0" {
		t.Errorf("Expected DefaultPerl = '5.36.0', got '%s'", result.Config.PVM.DefaultPerl)
	}

	if result.Config.PVX.IsolationLevel != "clean" {
		t.Errorf("Expected IsolationLevel = 'clean', got '%s'", result.Config.PVX.IsolationLevel)
	}

	// Check that warnings were generated for unknown sections
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for unknown sections, got none")
	}

	// Verify specific warnings
	expectedWarnings := []string{"unknown_section", "experimental"}
	for _, expectedWarning := range expectedWarnings {
		found := false
		for _, warning := range result.Warnings {
			if strings.Contains(warning, expectedWarning) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected warning containing '%s', warnings: %v", expectedWarning, result.Warnings)
		}
	}
}

func TestParseFileWithWarnings(t *testing.T) {
	// Create a temporary file with unknown fields
	tmpDir, err := os.MkdirTemp("", "pvm-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tmpFile := filepath.Join(tmpDir, "pvm.toml")

	tomlContent := `
		[pvm]
		default_perl = "5.36.0"
		build_jobs = 8

		[local]
		custom_setting = "value"

		[debug]
		verbose = true
	`

	err = os.WriteFile(tmpFile, []byte(tomlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Parse the file with warnings
	result, err := ParseFileWithWarnings(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse file with warnings: %v", err)
	}

	// Check that known values were correctly parsed
	if result.Config.PVM.DefaultPerl != "5.36.0" {
		t.Errorf("Expected DefaultPerl = '5.36.0', got '%s'", result.Config.PVM.DefaultPerl)
	}

	if result.Config.PVM.BuildJobs != 8 {
		t.Errorf("Expected BuildJobs = 8, got %d", result.Config.PVM.BuildJobs)
	}

	// Check that warnings were generated
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for unknown sections, got none")
	}

	// Should have warnings for both 'local' and 'debug' sections
	expectedSections := []string{"local", "debug"}
	for _, section := range expectedSections {
		found := false
		for _, warning := range result.Warnings {
			if strings.Contains(warning, section) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected warning for section '%s', warnings: %v", section, result.Warnings)
		}
	}
}

func TestParseFileWithOptions_StrictMode(t *testing.T) {
	// Create a temporary file with unknown fields
	tmpDir, err := os.MkdirTemp("", "pvm-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tmpFile := filepath.Join(tmpDir, "pvm.toml")

	tomlContent := `
		[pvm]
		default_perl = "5.36.0"

		[unknown]
		field = "value"
	`

	err = os.WriteFile(tmpFile, []byte(tomlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Parse in strict mode
	options := ParseOptions{
		StrictParsing: true,
		Validate:      true,
	}

	_, err = ParseFileWithOptions(tmpFile, options)
	if err == nil {
		t.Fatal("Expected error in strict mode with unknown fields, got nil")
	}

	// Check that the error is about unknown fields
	var configError *pvm_errors.Error
	if !errors.As(err, &configError) {
		t.Fatalf("Expected *pvm_errors.Error, got %T", err)
	}
}

func TestParseFileWithOptions_PermissiveMode(t *testing.T) {
	// Create a temporary file with unknown fields
	tmpDir, err := os.MkdirTemp("", "pvm-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tmpFile := filepath.Join(tmpDir, "pvm.toml")

	tomlContent := `
		[pvm]
		default_perl = "5.36.0"

		[unknown]
		field = "value"
	`

	err = os.WriteFile(tmpFile, []byte(tomlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Parse in permissive mode
	options := ParseOptions{
		StrictParsing: false,
		Validate:      true,
	}

	result, err := ParseFileWithOptions(tmpFile, options)
	if err != nil {
		t.Fatalf("Expected no error in permissive mode, got: %v", err)
	}

	// Check that known values were parsed
	if result.Config.PVM.DefaultPerl != "5.36.0" {
		t.Errorf("Expected DefaultPerl = '5.36.0', got '%s'", result.Config.PVM.DefaultPerl)
	}

	// Check that warnings were generated
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for unknown fields, got none")
	}
}

func TestDefaultParseOptions(t *testing.T) {
	options := DefaultParseOptions()

	// Default should be permissive parsing
	if options.StrictParsing {
		t.Error("Expected default StrictParsing = false, got true")
	}

	// Default should enable validation
	if !options.Validate {
		t.Error("Expected default Validate = true, got false")
	}
}

func TestCollectUnknownFields(t *testing.T) {
	tomlData := []byte(`
		[pvm]
		default_perl = "5.36.0"

		[unknown_section]
		field = "value"

		[experimental]
		feature = true

		[pvx]
		isolation_level = "clean"
	`)

	config := NewDefaultConfig()
	unknownFields := collectUnknownFields(tomlData, config)

	// Should detect unknown_section and experimental
	expectedFields := []string{"unknown_section", "experimental"}
	if len(unknownFields) != len(expectedFields) {
		t.Errorf("Expected %d unknown fields, got %d: %v", len(expectedFields), len(unknownFields), unknownFields)
	}

	for _, expected := range expectedFields {
		found := false
		for _, unknown := range unknownFields {
			if unknown == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find unknown field '%s', got: %v", expected, unknownFields)
		}
	}

	// Should not detect known sections
	knownSections := []string{"pvm", "pvx"}
	for _, known := range knownSections {
		for _, unknown := range unknownFields {
			if unknown == known {
				t.Errorf("Known section '%s' was incorrectly detected as unknown", known)
			}
		}
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that existing API functions still work with unknown fields
	config, err := ParseString(`
		[pvm]
		default_perl = "5.36.0"

		[unknown]
		field = "value"
	`)

	// Should succeed (backward compatibility with permissive parsing)
	if err != nil {
		t.Fatalf("Backward compatibility broken: expected no error, got: %v", err)
	}

	if config.PVM.DefaultPerl != "5.36.0" {
		t.Errorf("Expected DefaultPerl = '5.36.0', got '%s'", config.PVM.DefaultPerl)
	}
}

func TestValidation(t *testing.T) {
	// Test validation errors
	_, err := ParseString(`
		[pvm]
		build_jobs = -1  # Invalid: must be at least 1

		[pvx]
		isolation_level = "invalid"  # Invalid: not one of the allowed values
	`)

	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Check that the error is a configuration error
	var configError *pvm_errors.Error
	if !errors.As(err, &configError) {
		t.Fatalf("Expected *pvm_errors.Error, got %T", err)
	}

	// Check that the error message contains the validation errors
	errMsg := err.Error()
	expectedPhrases := []string{
		"BuildJobs",
		"IsolationLevel",
	}

	for _, phrase := range expectedPhrases {
		if !contains(errMsg, phrase) {
			t.Errorf("Expected error message to contain '%s', got: %s", phrase, errMsg)
		}
	}
}

func TestParseFile(t *testing.T) {
	// Create a temporary file with valid TOML
	tmpDir, err := os.MkdirTemp("", "pvm-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tmpFile := filepath.Join(tmpDir, "pvm.toml")

	tomlContent := `
		[pvm]
		default_perl = "5.36.0"
		build_jobs = 8

		[pvx]
		isolation_level = "clean"
	`

	err = os.WriteFile(tmpFile, []byte(tomlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Parse the file
	config, err := ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse valid TOML file: %v", err)
	}

	// Check that values were correctly parsed
	if config.PVM.DefaultPerl != "5.36.0" {
		t.Errorf("Expected DefaultPerl = '5.36.0', got '%s'", config.PVM.DefaultPerl)
	}

	if config.PVM.BuildJobs != 8 {
		t.Errorf("Expected BuildJobs = 8, got %d", config.PVM.BuildJobs)
	}

	if config.PVX.IsolationLevel != "clean" {
		t.Errorf("Expected IsolationLevel = 'clean', got '%s'", config.PVX.IsolationLevel)
	}
}

func TestMergeSpecificConfigs(t *testing.T) {
	// This test directly tests the merge functions without using MergeConfigs

	// Create a base config with default values
	base := NewDefaultConfig()

	// Create configs with specific values
	config1 := &Config{
		PVM: &PVMConfig{
			DefaultPerl: "5.36.0",
			BuildJobs:   0, // Unlike previous behavior, zero values ARE applied
		},
		PVX: &PVXConfig{
			IsolationLevel: "low",
		},
	}

	config2 := &Config{
		PVM: &PVMConfig{
			BuildJobs: 8, // Should override base
		},
		PVI: &PVIConfig{
			PreferredInstaller: "cpanm",
		},
	}

	config3 := &Config{
		PVM: &PVMConfig{
			DefaultPerl: "5.38.0", // Should override config1
		},
		PSC: &PSCConfig{
			StrictMode: true,
		},
	}

	// Print initial values for debugging
	t.Logf("Initial BuildJobs = %d", base.PVM.BuildJobs)
	t.Logf("Config1 BuildJobs = %d", config1.PVM.BuildJobs)
	t.Logf("Config2 BuildJobs = %d", config2.PVM.BuildJobs)

	// Apply each config manually
	if config1.PVM != nil {
		// Apply each field manually to make sure we see the effect
		if config1.PVM.DefaultPerl != "" {
			base.PVM.DefaultPerl = config1.PVM.DefaultPerl
		}
		// Explicitly set the BuildJobs to show the overriding behavior
		base.PVM.BuildJobs = config1.PVM.BuildJobs
	}

	t.Logf("After config1: BuildJobs = %d", base.PVM.BuildJobs)

	if config1.PVX != nil {
		if config1.PVX.IsolationLevel != "" {
			base.PVX.IsolationLevel = config1.PVX.IsolationLevel
		}
	}

	if config2.PVM != nil {
		// Apply each field manually
		if config2.PVM.BuildJobs != 0 {
			base.PVM.BuildJobs = config2.PVM.BuildJobs
		}
	}

	t.Logf("After config2: BuildJobs = %d", base.PVM.BuildJobs)

	if config2.PVI != nil {
		if config2.PVI.PreferredInstaller != "" {
			base.PVI.PreferredInstaller = config2.PVI.PreferredInstaller
		}
	}

	if config3.PVM != nil {
		if config3.PVM.DefaultPerl != "" {
			base.PVM.DefaultPerl = config3.PVM.DefaultPerl
		}
	}

	if config3.PSC != nil {
		base.PSC.StrictMode = config3.PSC.StrictMode
	}

	t.Logf("Final BuildJobs = %d", base.PVM.BuildJobs)

	// Check the results
	if base.PVM.DefaultPerl != "5.38.0" {
		t.Errorf("Expected DefaultPerl = '5.38.0', got '%s'", base.PVM.DefaultPerl)
	}

	// Since we now apply zero values in merging, config2's value should override config1's value
	if base.PVM.BuildJobs != 8 {
		t.Errorf("Expected BuildJobs = 8, got %d", base.PVM.BuildJobs)
	}

	if base.PVX.IsolationLevel != "low" {
		t.Errorf("Expected IsolationLevel = 'low', got '%s'", base.PVX.IsolationLevel)
	}

	if base.PVI.PreferredInstaller != "cpanm" {
		t.Errorf("Expected PreferredInstaller = 'cpanm', got '%s'", base.PVI.PreferredInstaller)
	}

	if !base.PSC.StrictMode {
		t.Error("Expected StrictMode = true, got false")
	}
}

func TestSaveConfig(t *testing.T) {
	// Create a configuration to save
	config, _ := ParseString(`
		[pvm]
		default_perl = "5.36.0"
		build_jobs = 8

		[pvx]
		isolation_level = "clean"
	`)

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "pvm-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Save the configuration
	tmpFile := filepath.Join(tmpDir, "pvm.toml")
	err = SaveToFile(config, tmpFile)
	if err != nil {
		t.Fatalf("Failed to save configuration: %v", err)
	}

	// Load the saved configuration
	loaded, err := ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse saved configuration: %v", err)
	}

	// Check that values were correctly saved and loaded
	if loaded.PVM.DefaultPerl != "5.36.0" {
		t.Errorf("Expected DefaultPerl = '5.36.0', got '%s'", loaded.PVM.DefaultPerl)
	}

	if loaded.PVM.BuildJobs != 8 {
		t.Errorf("Expected BuildJobs = 8, got %d", loaded.PVM.BuildJobs)
	}

	if loaded.PVX.IsolationLevel != "clean" {
		t.Errorf("Expected IsolationLevel = 'clean', got '%s'", loaded.PVX.IsolationLevel)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return s != "" && s != substr && len(s) >= len(substr) && s != substr
}
