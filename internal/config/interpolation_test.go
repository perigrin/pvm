// ABOUTME: Tests for environment variable interpolation functionality
// ABOUTME: Covers all interpolation scenarios including edge cases and error conditions

package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestInterpolationEngine_InterpolateString(t *testing.T) {
	engine := NewInterpolationEngine()

	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple variable substitution",
			input:    "${HOME}/config",
			envVars:  map[string]string{"HOME": "/users/test"},
			expected: "/users/test/config",
		},
		{
			name:     "variable with default value - env exists",
			input:    "${USER:-default}/config",
			envVars:  map[string]string{"USER": "testuser"},
			expected: "testuser/config",
		},
		{
			name:     "variable with default value - env missing",
			input:    "${MISSING:-default}/config",
			envVars:  map[string]string{},
			expected: "default/config",
		},
		{
			name:     "variable with empty default",
			input:    "${MISSING:-}/config",
			envVars:  map[string]string{},
			expected: "/config",
		},
		{
			name:     "multiple variables",
			input:    "${HOME}/${USER}/config",
			envVars:  map[string]string{"HOME": "/users", "USER": "test"},
			expected: "/users/test/config",
		},
		{
			name:     "no variables",
			input:    "/static/path/config",
			envVars:  map[string]string{},
			expected: "/static/path/config",
		},
		{
			name:     "empty variable",
			input:    "${EMPTY}/config",
			envVars:  map[string]string{"EMPTY": ""},
			expected: "/config",
		},
		{
			name:     "variable not found, no default",
			input:    "${NOTFOUND}/config",
			envVars:  map[string]string{},
			expected: "/config",
		},
		{
			name:     "recursive interpolation",
			input:    "${PATH_VAR}",
			envVars:  map[string]string{"PATH_VAR": "${BASE}/subdir", "BASE": "/root"},
			expected: "/root/subdir",
		},
		{
			name:     "complex default with variables",
			input:    "${CONFIG_DIR:-${HOME}/.config}/app",
			envVars:  map[string]string{"HOME": "/users/test"},
			expected: "/users/test/.config/app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			result, err := engine.InterpolateString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("InterpolateString() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("InterpolateString() error = %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("InterpolateString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestInterpolationEngine_CycleDetection(t *testing.T) {
	engine := NewInterpolationEngine()

	tests := []struct {
		name    string
		input   string
		envVars map[string]string
		wantErr bool
	}{
		{
			name:    "direct cycle",
			input:   "${VAR_A}",
			envVars: map[string]string{"VAR_A": "${VAR_A}"},
			wantErr: true,
		},
		{
			name:    "indirect cycle",
			input:   "${VAR_A}",
			envVars: map[string]string{"VAR_A": "${VAR_B}", "VAR_B": "${VAR_A}"},
			wantErr: true,
		},
		{
			name:    "three-way cycle",
			input:   "${VAR_A}",
			envVars: map[string]string{"VAR_A": "${VAR_B}", "VAR_B": "${VAR_C}", "VAR_C": "${VAR_A}"},
			wantErr: true,
		},
		{
			name:    "no cycle - reuse variable",
			input:   "${BASE}/${BASE}",
			envVars: map[string]string{"BASE": "test"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			_, err := engine.InterpolateString(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("InterpolateString() expected error for cycle detection, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("InterpolateString() unexpected error = %v", err)
			}
		})
	}
}

func TestInterpolationEngine_InterpolateConfig(t *testing.T) {
	engine := NewInterpolationEngine()

	// Set up test environment variables
	os.Setenv("TEST_HOME", "/test/home")
	os.Setenv("TEST_USER", "testuser")
	os.Setenv("TEST_PORT", "3001")
	defer func() {
		os.Unsetenv("TEST_HOME")
		os.Unsetenv("TEST_USER")
		os.Unsetenv("TEST_PORT")
	}()

	config := &Config{
		PVM: &PVMConfig{
			DefaultPerl:    "5.38.0",
			DownloadMirror: "${TEST_HOME}/mirror",
			PatchesDir:     "${TEST_HOME}/${TEST_USER}/patches",
		},
		PVX: &PVXConfig{
			SaveOutputDir:   "${TEST_HOME}/output",
			PreserveEnvVars: []string{"${TEST_USER}_VAR", "DISPLAY"},
		},
		PVI: &PVIConfig{
			CacheDir: "${TEST_HOME:-/default}/.cache",
		},
		PSC: &PSCConfig{
			TypeDefinitionsPath: "${TEST_HOME}/types",
		},
		MCP: &MCPConfig{
			Host: "localhost",
			Port: 3000, // This won't be interpolated as it's an int
		},
	}

	result, err := engine.InterpolateConfig(config)
	if err != nil {
		t.Fatalf("InterpolateConfig() error = %v", err)
	}

	// Verify PVM interpolation
	if result.PVM.DownloadMirror != "/test/home/mirror" {
		t.Errorf("PVM.DownloadMirror = %v, want /test/home/mirror", result.PVM.DownloadMirror)
	}
	if result.PVM.PatchesDir != "/test/home/testuser/patches" {
		t.Errorf("PVM.PatchesDir = %v, want /test/home/testuser/patches", result.PVM.PatchesDir)
	}

	// Verify PVX interpolation
	if result.PVX.SaveOutputDir != "/test/home/output" {
		t.Errorf("PVX.SaveOutputDir = %v, want /test/home/output", result.PVX.SaveOutputDir)
	}
	if len(result.PVX.PreserveEnvVars) != 2 || result.PVX.PreserveEnvVars[0] != "testuser_VAR" {
		t.Errorf("PVX.PreserveEnvVars = %v, want [testuser_VAR, DISPLAY]", result.PVX.PreserveEnvVars)
	}

	// Verify PVI interpolation
	if result.PVI.CacheDir != "/test/home/.cache" {
		t.Errorf("PVI.CacheDir = %v, want /test/home/.cache", result.PVI.CacheDir)
	}

	// Verify PSC interpolation
	if result.PSC.TypeDefinitionsPath != "/test/home/types" {
		t.Errorf("PSC.TypeDefinitionsPath = %v, want /test/home/types", result.PSC.TypeDefinitionsPath)
	}

	// Verify MCP values remain unchanged
	if result.MCP.Port != 3000 {
		t.Errorf("MCP.Port = %v, want 3000", result.MCP.Port)
	}
	if result.MCP.Host != "localhost" {
		t.Errorf("MCP.Host = %v, want localhost", result.MCP.Host)
	}
}

func TestInterpolationEngine_TypeConversion(t *testing.T) {
	engine := NewInterpolationEngine()

	tests := []struct {
		name     string
		envVars  map[string]string
		config   *Config
		validate func(*Config, *testing.T)
	}{
		{
			name: "string interpolation in various config sections",
			envVars: map[string]string{
				"CONFIG_BASE": "/custom/config",
				"CACHE_BASE":  "/custom/cache",
			},
			config: &Config{
				PVM: &PVMConfig{
					DownloadMirror: "${CONFIG_BASE}/mirror",
				},
				PVI: &PVIConfig{
					CacheDir: "${CACHE_BASE}/pvi",
				},
			},
			validate: func(result *Config, t *testing.T) {
				if result.PVM.DownloadMirror != "/custom/config/mirror" {
					t.Errorf("Expected /custom/config/mirror, got %v", result.PVM.DownloadMirror)
				}
				if result.PVI.CacheDir != "/custom/cache/pvi" {
					t.Errorf("Expected /custom/cache/pvi, got %v", result.PVI.CacheDir)
				}
			},
		},
		{
			name: "map interpolation",
			envVars: map[string]string{
				"LATEST_VERSION": "5.38.0",
				"STABLE_VERSION": "5.36.0",
			},
			config: &Config{
				PVM: &PVMConfig{
					VersionAliases: map[string]string{
						"latest": "${LATEST_VERSION}",
						"stable": "${STABLE_VERSION}",
					},
				},
			},
			validate: func(result *Config, t *testing.T) {
				if result.PVM.VersionAliases["latest"] != "5.38.0" {
					t.Errorf("Expected 5.38.0, got %v", result.PVM.VersionAliases["latest"])
				}
				if result.PVM.VersionAliases["stable"] != "5.36.0" {
					t.Errorf("Expected 5.36.0, got %v", result.PVM.VersionAliases["stable"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			result, err := engine.InterpolateConfig(tt.config)
			if err != nil {
				t.Fatalf("InterpolateConfig() error = %v", err)
			}

			tt.validate(result, t)
		})
	}
}

func TestInterpolationEngine_SensitiveVariables(t *testing.T) {
	engine := NewInterpolationEngine()

	tests := []struct {
		name        string
		varName     string
		isSensitive bool
	}{
		{"password variable", "DB_PASSWORD", true},
		{"secret variable", "API_SECRET", true},
		{"key variable", "PRIVATE_KEY", true},
		{"token variable", "AUTH_TOKEN", true},
		{"api key variable", "OPENAI_API_KEY", true},
		{"normal variable", "HOME", false},
		{"user variable", "USER", false},
		{"path variable", "PATH", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.IsSensitiveVariable(tt.varName)
			if result != tt.isSensitive {
				t.Errorf("IsSensitiveVariable(%v) = %v, want %v", tt.varName, result, tt.isSensitive)
			}
		})
	}
}

func TestInterpolationEngine_MaskSensitiveValue(t *testing.T) {
	engine := NewInterpolationEngine()

	tests := []struct {
		name     string
		varName  string
		value    string
		expected string
	}{
		{
			name:     "mask password",
			varName:  "DB_PASSWORD",
			value:    "supersecret123",
			expected: "su***23",
		},
		{
			name:     "mask short secret",
			varName:  "API_SECRET",
			value:    "short",
			expected: "***",
		},
		{
			name:     "don't mask normal variable",
			varName:  "USER",
			value:    "testuser",
			expected: "testuser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.MaskSensitiveValue(tt.varName, tt.value)
			if result != tt.expected {
				t.Errorf("MaskSensitiveValue(%v, %v) = %v, want %v", tt.varName, tt.value, result, tt.expected)
			}
		})
	}
}

func TestInterpolationEngine_ValidationAfterInterpolation(t *testing.T) {
	engine := NewInterpolationEngine()

	// Set up environment variables that will result in invalid configuration
	os.Setenv("INVALID_PORT", "99999")
	os.Setenv("INVALID_JOBS", "-1")
	defer func() {
		os.Unsetenv("INVALID_PORT")
		os.Unsetenv("INVALID_JOBS")
	}()

	config := &Config{
		PVM: &PVMConfig{
			BuildJobs:      1, // This will remain valid
			DownloadMirror: "https://example.com",
		},
		MCP: &MCPConfig{
			Port:                  99999, // This is invalid but won't be interpolated since it's an int
			Host:                  "localhost",
			EmbeddingProvider:     "openai",
			ValidationCacheSize:   "50MB",
			EmbeddingCacheSize:    "100MB",
			GenerationMemorySize:  50,
			MaxConcurrentRequests: 10,
			RequestTimeout:        30 * time.Second,
		},
	}

	result, err := engine.InterpolateConfig(config)
	if err != nil {
		t.Fatalf("InterpolateConfig() error = %v", err)
	}

	// Validate the interpolated configuration
	err = engine.ValidateInterpolatedConfig(result)
	if err == nil {
		t.Error("Expected validation error for invalid port, got nil")
	}

	if !strings.Contains(err.Error(), "Port must be between 1 and 65535") {
		t.Errorf("Expected port validation error, got: %v", err)
	}
}

func TestInterpolationEngine_ComplexScenarios(t *testing.T) {
	engine := NewInterpolationEngine()

	// Test nested interpolation with multiple levels
	os.Setenv("BASE_DIR", "/opt")
	os.Setenv("APP_NAME", "pvm")
	os.Setenv("VERSION", "1.0")
	os.Setenv("FULL_PATH", "${BASE_DIR}/${APP_NAME}/${VERSION}")
	defer func() {
		os.Unsetenv("BASE_DIR")
		os.Unsetenv("APP_NAME")
		os.Unsetenv("VERSION")
		os.Unsetenv("FULL_PATH")
	}()

	config := &Config{
		PSC: &PSCConfig{
			TypeDefinitionsPath: "${FULL_PATH}/types",
		},
		PVI: &PVIConfig{
			CacheDir: "${BASE_DIR:-/default}/${APP_NAME:-unknown}/cache",
		},
	}

	result, err := engine.InterpolateConfig(config)
	if err != nil {
		t.Fatalf("InterpolateConfig() error = %v", err)
	}

	expectedPath := "/opt/pvm/1.0/types"
	if result.PSC.TypeDefinitionsPath != expectedPath {
		t.Errorf("Expected %v, got %v", expectedPath, result.PSC.TypeDefinitionsPath)
	}

	expectedCache := "/opt/pvm/cache"
	if result.PVI.CacheDir != expectedCache {
		t.Errorf("Expected %v, got %v", expectedCache, result.PVI.CacheDir)
	}
}

func TestInterpolationEngine_EdgeCases(t *testing.T) {
	engine := NewInterpolationEngine()

	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "malformed variable (no closing brace)",
			input:    "${UNCLOSED/path",
			envVars:  map[string]string{},
			expected: "${UNCLOSED/path", // Should remain unchanged
		},
		{
			name:     "empty variable name",
			input:    "${}/path",
			envVars:  map[string]string{},
			expected: "/path",
		},
		{
			name:     "whitespace in variable name",
			input:    "${ SPACE }/path",
			envVars:  map[string]string{"SPACE": "test"},
			expected: "test/path",
		},
		{
			name:     "default with colon but no dash",
			input:    "${VAR:nodash}/path",
			envVars:  map[string]string{},
			expected: "/path", // Treated as missing variable
		},
		{
			name:     "nested braces in default",
			input:    "${VAR:-${NESTED}}/path",
			envVars:  map[string]string{"NESTED": "default"},
			expected: "default/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			result, err := engine.InterpolateString(tt.input)
			if err != nil {
				t.Errorf("InterpolateString() unexpected error = %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("InterpolateString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestInterpolationEngine_MaxDepthProtection(t *testing.T) {
	engine := NewInterpolationEngine()
	engine.maxDepth = 3 // Set a low max depth for testing

	// Create a deep chain of variables
	os.Setenv("VAR1", "${VAR2}")
	os.Setenv("VAR2", "${VAR3}")
	os.Setenv("VAR3", "${VAR4}")
	os.Setenv("VAR4", "${VAR5}")
	os.Setenv("VAR5", "final")
	defer func() {
		for i := 1; i <= 5; i++ {
			os.Unsetenv("VAR" + string(rune('0'+i)))
		}
	}()

	_, err := engine.InterpolateString("${VAR1}")
	if err == nil {
		t.Error("Expected error for exceeding max depth, got nil")
	}

	if !strings.Contains(err.Error(), "depth exceeded") {
		t.Errorf("Expected depth exceeded error, got: %v", err)
	}
}
