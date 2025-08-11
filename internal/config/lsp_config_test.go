package config

import (
	"testing"
	"time"
)

func TestPSCLSPConfig_Validate(t *testing.T) {
	tests := []struct {
		name          string
		config        *PSCLSPConfig
		expectedError string
	}{
		{
			name: "valid minimal config",
			config: &PSCLSPConfig{
				LogLevel:     "info",
				DefaultMode:  "stdio",
				TCPPort:      9999,
				MaxCacheSize: 1000,
			},
			expectedError: "",
		},
		{
			name: "invalid log level",
			config: &PSCLSPConfig{
				LogLevel: "invalid",
			},
			expectedError: "LSP log_level must be one of: debug, info, warn, error",
		},
		{
			name: "invalid default mode",
			config: &PSCLSPConfig{
				DefaultMode: "invalid",
			},
			expectedError: "LSP default_mode must be one of: stdio, tcp",
		},
		{
			name: "invalid TCP port - too low",
			config: &PSCLSPConfig{
				TCPPort: 0,
			},
			expectedError: "",
		},
		{
			name: "invalid TCP port - too high",
			config: &PSCLSPConfig{
				TCPPort: 70000,
			},
			expectedError: "LSP tcp_port must be between 1 and 65535",
		},
		{
			name: "negative max cache size",
			config: &PSCLSPConfig{
				MaxCacheSize: -1,
			},
			expectedError: "LSP max_cache_size cannot be negative",
		},
		{
			name: "negative request timeout",
			config: &PSCLSPConfig{
				RequestTimeout: -1 * time.Second,
			},
			expectedError: "LSP request_timeout cannot be negative",
		},
		{
			name: "negative diagnostics delay",
			config: &PSCLSPConfig{
				DiagnosticsDelay: -1 * time.Millisecond,
			},
			expectedError: "LSP diagnostics_delay cannot be negative",
		},
		{
			name: "log file with null bytes",
			config: &PSCLSPConfig{
				LogFile: "test\x00file.log",
			},
			expectedError: "LSP log_file cannot contain null bytes or control characters",
		},
		{
			name: "empty exclude patterns",
			config: &PSCLSPConfig{
				ExcludePatterns: []string{"valid", "", "also_valid"},
			},
			expectedError: "LSP exclude_patterns cannot contain empty patterns",
		},
		{
			name: "empty include directories",
			config: &PSCLSPConfig{
				IncludeDirectories: []string{"lib", "", "script"},
			},
			expectedError: "LSP include_directories cannot contain empty directories",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errors := test.config.Validate()

			if test.expectedError == "" {
				if len(errors) > 0 {
					t.Errorf("Expected no validation errors, but got: %v", errors)
				}
			} else {
				found := false
				for _, err := range errors {
					if err.Error() == test.expectedError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error '%s', but got: %v", test.expectedError, errors)
				}
			}
		})
	}
}

func TestPSCConfig_ValidateLSP(t *testing.T) {
	// Test that PSCConfig properly validates its LSP field
	config := &PSCConfig{
		TypeDefinitionsPath: "/path/to/types",
		LSP: &PSCLSPConfig{
			LogLevel: "invalid",
		},
	}

	errors := config.Validate()
	expectedError := "LSP log_level must be one of: debug, info, warn, error"
	found := false
	for _, err := range errors {
		if err.Error() == expectedError {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected PSC validation to include LSP validation error: %s, but got: %v", expectedError, errors)
	}
}

func TestMergePSCLSPConfig(t *testing.T) {
	// Test the merge function for PSC LSP configuration
	target := &PSCLSPConfig{
		LogFile:            "/original/log.txt",
		LogLevel:           "info",
		DefaultMode:        "stdio",
		TCPPort:            9999,
		Verbose:            false,
		EnableHover:        true,
		MaxCacheSize:       1000,
		RequestTimeout:     5 * time.Second,
		ExcludePatterns:    []string{"**/original/**"},
		IncludeDirectories: []string{"lib"},
	}

	source := &PSCLSPConfig{
		LogFile:            "/new/log.txt",
		LogLevel:           "debug",
		DefaultMode:        "tcp",
		TCPPort:            8080,
		Verbose:            true,
		EnableHover:        false, // Zero value - will override target's true value
		EnableCompletion:   true,
		MaxCacheSize:       2000,
		RequestTimeout:     10 * time.Second,
		ExcludePatterns:    []string{"**/new/**", "**/temp/**"},
		IncludeDirectories: []string{"lib", "script"},
	}

	mergePSCLSPConfig(target, source)

	// Check string fields (non-empty source overrides target)
	if target.LogFile != "/new/log.txt" {
		t.Errorf("LogFile = %v, want /new/log.txt", target.LogFile)
	}
	if target.LogLevel != "debug" {
		t.Errorf("LogLevel = %v, want debug", target.LogLevel)
	}
	if target.DefaultMode != "tcp" {
		t.Errorf("DefaultMode = %v, want tcp", target.DefaultMode)
	}

	// Check integer fields (always merge)
	if target.TCPPort != 8080 {
		t.Errorf("TCPPort = %v, want 8080", target.TCPPort)
	}
	if target.MaxCacheSize != 2000 {
		t.Errorf("MaxCacheSize = %v, want 2000", target.MaxCacheSize)
	}

	// Check duration fields (always merge)
	if target.RequestTimeout != 10*time.Second {
		t.Errorf("RequestTimeout = %v, want 10s", target.RequestTimeout)
	}

	// Check boolean fields (always merge - source takes precedence even if false)
	if target.Verbose != true {
		t.Errorf("Verbose = %v, want true", target.Verbose)
	}
	if target.EnableHover != false {
		t.Errorf("EnableHover = %v, want false (source value)", target.EnableHover)
	}
	if target.EnableCompletion != true {
		t.Errorf("EnableCompletion = %v, want true", target.EnableCompletion)
	}

	// Check arrays (replaced entirely)
	if len(target.ExcludePatterns) != 2 {
		t.Errorf("ExcludePatterns length = %v, want 2", len(target.ExcludePatterns))
	}
	if target.ExcludePatterns[0] != "**/new/**" || target.ExcludePatterns[1] != "**/temp/**" {
		t.Errorf("ExcludePatterns = %v, want [**/new/** **/temp/**]", target.ExcludePatterns)
	}
	if len(target.IncludeDirectories) != 2 {
		t.Errorf("IncludeDirectories length = %v, want 2", len(target.IncludeDirectories))
	}
	if target.IncludeDirectories[0] != "lib" || target.IncludeDirectories[1] != "script" {
		t.Errorf("IncludeDirectories = %v, want [lib script]", target.IncludeDirectories)
	}
}

func TestMergePSCConfigWithLSP(t *testing.T) {
	// Test that PSC config merge properly handles LSP subsection
	target := &PSCConfig{
		TypeDefinitionsPath: "/path/to/types",
		StrictMode:          false,
		LSP: &PSCLSPConfig{
			LogLevel:    "info",
			DefaultMode: "stdio",
		},
	}

	source := &PSCConfig{
		StrictMode: true,
		LSP: &PSCLSPConfig{
			LogLevel:    "debug",
			DefaultMode: "tcp",
			TCPPort:     8080,
		},
	}

	mergePSCConfig(target, source)

	// Check that general PSC fields were merged
	if target.StrictMode != true {
		t.Errorf("StrictMode = %v, want true", target.StrictMode)
	}

	// Check that LSP subsection was merged properly
	if target.LSP == nil {
		t.Fatal("LSP config should not be nil after merge")
	}
	if target.LSP.LogLevel != "debug" {
		t.Errorf("LSP LogLevel = %v, want debug", target.LSP.LogLevel)
	}
	if target.LSP.DefaultMode != "tcp" {
		t.Errorf("LSP DefaultMode = %v, want tcp", target.LSP.DefaultMode)
	}
	if target.LSP.TCPPort != 8080 {
		t.Errorf("LSP TCPPort = %v, want 8080", target.LSP.TCPPort)
	}
}

func TestMergePSCConfigWithNilLSP(t *testing.T) {
	// Test that PSC config merge creates LSP config when target is nil
	target := &PSCConfig{
		TypeDefinitionsPath: "/path/to/types",
		LSP:                 nil,
	}

	source := &PSCConfig{
		LSP: &PSCLSPConfig{
			LogLevel:    "debug",
			DefaultMode: "tcp",
			TCPPort:     8080,
		},
	}

	mergePSCConfig(target, source)

	// Check that LSP config was created
	if target.LSP == nil {
		t.Fatal("LSP config should have been created during merge")
	}
	if target.LSP.LogLevel != "debug" {
		t.Errorf("LSP LogLevel = %v, want debug", target.LSP.LogLevel)
	}
	if target.LSP.DefaultMode != "tcp" {
		t.Errorf("LSP DefaultMode = %v, want tcp", target.LSP.DefaultMode)
	}
}

func TestNewDefaultConfig_LSP(t *testing.T) {
	// Test that default configuration includes reasonable LSP defaults
	config := NewDefaultConfig()

	if config.PSC == nil {
		t.Fatal("PSC config should not be nil in default config")
	}
	if config.PSC.LSP == nil {
		t.Fatal("LSP config should not be nil in default PSC config")
	}

	lsp := config.PSC.LSP

	// Check defaults are reasonable
	if lsp.LogLevel != "info" {
		t.Errorf("Default LogLevel = %v, want info", lsp.LogLevel)
	}
	if lsp.DefaultMode != "stdio" {
		t.Errorf("Default DefaultMode = %v, want stdio", lsp.DefaultMode)
	}
	if lsp.TCPPort != 9999 {
		t.Errorf("Default TCPPort = %v, want 9999", lsp.TCPPort)
	}
	if !lsp.EnableHover {
		t.Error("Default EnableHover should be true")
	}
	if !lsp.EnableCompletion {
		t.Error("Default EnableCompletion should be true")
	}
	if lsp.MaxCacheSize != 1000 {
		t.Errorf("Default MaxCacheSize = %v, want 1000", lsp.MaxCacheSize)
	}
	if lsp.RequestTimeout != 5*time.Second {
		t.Errorf("Default RequestTimeout = %v, want 5s", lsp.RequestTimeout)
	}
}
