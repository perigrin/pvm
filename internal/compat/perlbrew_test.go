// ABOUTME: Unit tests for perlbrew compatibility interface
// ABOUTME: Tests command mapping and argument transformation for perlbrew

package compat

import (
	"strings"
	"testing"
)

func TestPerlbrewMapper_MapCommand(t *testing.T) {
	mapper := NewPerlbrewMapper()

	tests := []struct {
		name     string
		args     []string
		wantCmd  string
		wantArgs []string
		wantErr  bool
	}{
		{
			name:     "Init command",
			args:     []string{"init"},
			wantCmd:  "init",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Install perl version",
			args:     []string{"install", "perl-5.38.0"},
			wantCmd:  "install",
			wantArgs: []string{"5.38.0"},
			wantErr:  false,
		},
		{
			name:     "Install with notest flag",
			args:     []string{"install", "--notest", "perl-5.38.0"},
			wantCmd:  "install",
			wantArgs: []string{"5.38.0", "--no-test"},
			wantErr:  false,
		},
		{
			name:     "Install with force and verbose",
			args:     []string{"install", "-f", "-v", "5.38.0"},
			wantCmd:  "install",
			wantArgs: []string{"5.38.0", "--force", "--verbose"},
			wantErr:  false,
		},
		{
			name:     "Install with alias",
			args:     []string{"install", "--as", "stable", "perl-5.38.0"},
			wantCmd:  "install",
			wantArgs: []string{"5.38.0", "--alias", "stable"},
			wantErr:  false,
		},
		{
			name:     "Install with thread options",
			args:     []string{"install", "--thread", "--multi", "5.38.0"},
			wantCmd:  "install",
			wantArgs: []string{"5.38.0", "--threaded", "--multi"},
			wantErr:  false,
		},
		{
			name:     "List installed versions",
			args:     []string{"list"},
			wantCmd:  "list",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "List available versions",
			args:     []string{"available"},
			wantCmd:  "available",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Use perl version (shell)",
			args:     []string{"use", "perl-5.38.0"},
			wantCmd:  "shell",
			wantArgs: []string{"5.38.0"},
			wantErr:  false,
		},
		{
			name:     "Switch perl version (permanent)",
			args:     []string{"switch", "5.38.0"},
			wantCmd:  "use",
			wantArgs: []string{"5.38.0"},
			wantErr:  false,
		},
		{
			name:     "Switch off (system perl)",
			args:     []string{"off"},
			wantCmd:  "use",
			wantArgs: []string{"system"},
			wantErr:  false,
		},
		{
			name:     "Current version",
			args:     []string{"current"},
			wantCmd:  "current",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Execute command",
			args:     []string{"exec", "--", "prove", "t/"},
			wantCmd:  "exec",
			wantArgs: []string{"prove", "t/"},
			wantErr:  false,
		},
		{
			name:     "Execute without double dash",
			args:     []string{"exec", "perl", "-v"},
			wantCmd:  "exec",
			wantArgs: []string{"perl", "-v"},
			wantErr:  false,
		},
		{
			name:     "Uninstall version",
			args:     []string{"uninstall", "perl-5.36.0"},
			wantCmd:  "uninstall",
			wantArgs: []string{"5.36.0"},
			wantErr:  false,
		},
		{
			name:     "Self upgrade",
			args:     []string{"self-upgrade"},
			wantCmd:  "self-update",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Info command",
			args:     []string{"info", "perl-5.38.0"},
			wantCmd:  "info",
			wantArgs: []string{"5.38.0"},
			wantErr:  false,
		},
		{
			name:     "Info without version",
			args:     []string{"info"},
			wantCmd:  "info",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Lib list",
			args:     []string{"lib"},
			wantCmd:  "lib",
			wantArgs: []string{"list"},
			wantErr:  false,
		},
		{
			name:     "Lib create",
			args:     []string{"lib", "create", "mylib"},
			wantCmd:  "lib",
			wantArgs: []string{"create", "mylib"},
			wantErr:  false,
		},
		{
			name:     "Lib delete",
			args:     []string{"lib", "delete", "mylib"},
			wantCmd:  "lib",
			wantArgs: []string{"delete", "mylib"},
			wantErr:  false,
		},
		{
			name:     "Clone installation",
			args:     []string{"clone", "perl-5.38.0", "perl-5.38.0-custom"},
			wantCmd:  "clone",
			wantArgs: []string{"5.38.0", "5.38.0-custom"},
			wantErr:  false,
		},
		{
			name:     "Clean command",
			args:     []string{"clean"},
			wantCmd:  "clean",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Version command",
			args:     []string{"version"},
			wantCmd:  "version",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Env command",
			args:     []string{"env"},
			wantCmd:  "env",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Alias list",
			args:     []string{"alias"},
			wantCmd:  "alias",
			wantArgs: []string{"list"},
			wantErr:  false,
		},
		{
			name:     "Alias create",
			args:     []string{"alias", "create", "5.38.0", "stable"},
			wantCmd:  "alias",
			wantArgs: []string{"create", "5.38.0", "stable"},
			wantErr:  false,
		},
		{
			name:     "Help command",
			args:     []string{"help"},
			wantCmd:  "",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:    "No command provided",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Install without version",
			args:    []string{"install"},
			wantErr: true,
		},
		{
			name:    "Use without version",
			args:    []string{"use"},
			wantErr: true,
		},
		{
			name:    "Switch without version",
			args:    []string{"switch"},
			wantErr: true,
		},
		{
			name:    "Uninstall without version",
			args:    []string{"uninstall"},
			wantErr: true,
		},
		{
			name:    "Exec without command",
			args:    []string{"exec"},
			wantErr: true,
		},
		{
			name:    "Lib create without name",
			args:    []string{"lib", "create"},
			wantErr: true,
		},
		{
			name:    "Clone without enough args",
			args:    []string{"clone", "5.38.0"},
			wantErr: true,
		},
		{
			name:    "Unknown command",
			args:    []string{"unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs, err := mapper.MapCommand(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("MapCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return // Expected error, test passed
			}

			if gotCmd != tt.wantCmd {
				t.Errorf("MapCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}

			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("MapCommand() gotArgs length = %v, want %v", len(gotArgs), len(tt.wantArgs))
				return
			}

			for i, arg := range gotArgs {
				if arg != tt.wantArgs[i] {
					t.Errorf("MapCommand() gotArgs[%d] = %v, want %v", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestPerlbrewMapper_GetHelp(t *testing.T) {
	mapper := NewPerlbrewMapper()
	help := mapper.GetHelp()

	expectedStrings := []string{
		"perlbrew",
		"Usage:",
		"Commands:",
		"Examples:",
		"install",
		"switch",
		"use",
		"list",
		"available",
		"lib",
		"--notest",
		"PVM compatibility interface",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(help, expected) {
			t.Errorf("GetHelp() missing expected string: %s", expected)
		}
	}
}

func TestPerlbrewMapper_GetToolName(t *testing.T) {
	mapper := NewPerlbrewMapper()
	if mapper.GetToolName() != "perlbrew" {
		t.Errorf("GetToolName() = %v, want %v", mapper.GetToolName(), "perlbrew")
	}
}

func TestParseVersionString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Full version with perl prefix",
			input:    "perl-5.38.0",
			expected: "5.38.0",
		},
		{
			name:     "Full version without prefix",
			input:    "5.38.0",
			expected: "5.38.0",
		},
		{
			name:     "Two part version with prefix",
			input:    "perl-5.38",
			expected: "5.38.0",
		},
		{
			name:     "Two part version without prefix",
			input:    "5.38",
			expected: "5.38.0",
		},
		{
			name:     "Single part version",
			input:    "5",
			expected: "5",
		},
		{
			name:     "Complex version",
			input:    "perl-5.38.2",
			expected: "5.38.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersionString(tt.input)
			if result != tt.expected {
				t.Errorf("parseVersionString(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}
